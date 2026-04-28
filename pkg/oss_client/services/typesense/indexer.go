package typesense

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	httpclient "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/httpclient"
	util "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
)

const (
	defaultCollection    = "oss-register"
	defaultDetailBaseURL = "https://oss.developer.overheid.nl/repositories"
	defaultLanguage      = "nl"
	defaultItemPriority  = 1
)

// ErrDisabled is returned when Typesense configuration is missing.
var ErrDisabled = errors.New("typesense indexing disabled: missing endpoint, api key or collection name")

type config struct {
	endpoint       string
	apiKey         string
	collection     string
	detailBaseURL  string
	language       string
	itemPriority   int
	defaultTags    []string
	featureEnabled bool
}

func loadConfigFromEnv() config {
	endpoint := strings.TrimSpace(os.Getenv("TYPESENSE_ENDPOINT"))
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("TYPESENSE_BASE_URL"))
	}

	apiKey := strings.TrimSpace(os.Getenv("TYPESENSE_API_KEY"))
	collection := strings.TrimSpace(os.Getenv("TYPESENSE_COLLECTION"))
	if collection == "" {
		collection = defaultCollection
	}

	detailBase := strings.TrimSpace(os.Getenv("TYPESENSE_DETAIL_BASE_URL"))
	if detailBase == "" {
		detailBase = defaultDetailBaseURL
	}

	language := strings.TrimSpace(os.Getenv("TYPESENSE_LANGUAGE"))
	if language == "" {
		language = defaultLanguage
	}

	itemPriority := defaultItemPriority
	if raw := strings.TrimSpace(os.Getenv("TYPESENSE_ITEM_PRIORITY")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			itemPriority = v
		}
	}

	return config{
		endpoint:       endpoint,
		apiKey:         apiKey,
		collection:     collection,
		detailBaseURL:  detailBase,
		language:       language,
		itemPriority:   itemPriority,
		defaultTags:    parseDefaultTags(),
		featureEnabled: isFeatureEnabled(),
	}
}

func (c config) enabled() bool {
	return c.featureEnabled && c.endpoint != "" && c.apiKey != "" && c.collection != ""
}

func isFeatureEnabled() bool {
	raw := strings.TrimSpace(os.Getenv("ENABLE_TYPESENSE"))
	if raw == "" {
		return true
	}
	switch strings.ToLower(raw) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// Enabled reports whether Typesense indexing is active based on env vars.
func Enabled() bool {
	return loadConfigFromEnv().enabled()
}

func parseDefaultTags() []string {
	raw := os.Getenv("TYPESENSE_DEFAULT_TAGS")
	if strings.TrimSpace(raw) == "" {
		return []string{"oss-register", "repository"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return []string{"oss-register", "repository"}
	}
	return out
}

// PublishRepository pushes the provided repository to Typesense for full-text search.
func PublishRepository(ctx context.Context, repository *models.Repository) (err error) {
	if repository == nil {
		return fmt.Errorf("typesense: repository is nil")
	}

	cfg := loadConfigFromEnv()
	if !cfg.enabled() {
		return ErrDisabled
	}

	payload, err := json.Marshal(buildDocument(cfg, repository))
	if err != nil {
		return fmt.Errorf("typesense: marshal payload: %w", err)
	}

	base := strings.TrimRight(cfg.endpoint, "/")
	target := fmt.Sprintf("%s/collections/%s/documents?action=upsert", base, url.PathEscape(cfg.collection))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("typesense: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TYPESENSE-API-KEY", cfg.apiKey)

	resp, err := httpclient.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("typesense: request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("typesense: close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if readErr != nil {
			return fmt.Errorf("typesense: read error response: %w", readErr)
		}
		return fmt.Errorf("typesense: indexing failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func buildDocument(cfg config, repository *models.Repository) map[string]any {
	doc := map[string]any{
		"type":          "doc",
		"language":      cfg.language,
		"item_priority": cfg.itemPriority,
	}

	if id := strings.TrimSpace(repository.Id); id != "" {
		doc["id"] = id
	}

	detailBase := strings.TrimRight(cfg.detailBaseURL, "/")
	if detailBase != "" && repository.Id != "" {
		detailURL := fmt.Sprintf("%s/%s", detailBase, repository.Id)
		doc["url"] = detailURL
		doc["url_without_anchor"] = detailURL
		doc["anchor"] = nil
	}

	if title := repositoryTitle(repository); title != "" {
		doc["hierarchy.lvl0"] = title
	}
	if org := repositoryOrganisationLabel(repository); org != "" {
		doc["hierarchy.lvl1"] = org
	}
	if repository.PublicCode != nil {
		if softwareType := strings.TrimSpace(repository.PublicCode.SoftwareType); softwareType != "" {
			doc["hierarchy.lvl2"] = softwareType
		}
		if developmentStatus := strings.TrimSpace(repository.PublicCode.DevelopmentStatus); developmentStatus != "" {
			doc["hierarchy.lvl3"] = developmentStatus
		}
		if repository.PublicCode.Legal != nil {
			if license := strings.TrimSpace(repository.PublicCode.Legal.License); license != "" {
				doc["hierarchy.lvl4"] = license
			}
		}
	}

	if _, ok := doc["hierarchy.lvl2"]; !ok {
		if repoURL := strings.TrimSpace(repository.Url); repoURL != "" {
			doc["hierarchy.lvl2"] = repoURL
		}
	}

	if content := buildContent(repository); content != "" {
		doc["content"] = content
	}

	if tags := buildTags(cfg, repository); len(tags) > 0 {
		doc["tags"] = tags
	}

	return doc
}

func repositoryTitle(repository *models.Repository) string {
	if name := strings.TrimSpace(repository.Name); name != "" {
		return name
	}
	if repository.PublicCode != nil {
		if name := strings.TrimSpace(repository.PublicCode.Name); name != "" {
			return name
		}
	}
	return strings.TrimSpace(repository.Url)
}

func repositoryOrganisationLabel(repository *models.Repository) string {
	if repository.Organisation != nil {
		if label := strings.TrimSpace(repository.Organisation.Label); label != "" {
			return label
		}
		if uri := strings.TrimSpace(repository.Organisation.Uri); uri != "" {
			return uri
		}
	}
	if repository.OrganisationID != nil {
		return strings.TrimSpace(*repository.OrganisationID)
	}
	return ""
}

func buildContent(repository *models.Repository) string {
	parts := make([]string, 0)

	if shortDesc := strings.TrimSpace(repository.ShortDescription); shortDesc != "" {
		parts = append(parts, shortDesc)
	}
	if longDesc := strings.TrimSpace(repository.LongDescription); longDesc != "" && longDesc != strings.TrimSpace(repository.ShortDescription) {
		parts = append(parts, longDesc)
	}
	if repoURL := strings.TrimSpace(repository.Url); repoURL != "" {
		parts = append(parts, fmt.Sprintf("Repository: %s", repoURL))
	}
	if publicCodeURL := strings.TrimSpace(repository.PublicCodeUrl); publicCodeURL != "" {
		parts = append(parts, fmt.Sprintf("Publiccode: %s", publicCodeURL))
	}
	if org := repositoryOrganisationLabel(repository); org != "" {
		parts = append(parts, fmt.Sprintf("Organisatie: %s", org))
	}

	forkType := util.DetectRepositoryForkType(repository)
	if forkType != "" {
		parts = append(parts, fmt.Sprintf("ForkType: %s", forkType))
	}

	pc := repository.PublicCode
	if pc == nil {
		if len(parts) == 0 {
			return repositoryTitle(repository)
		}
		return strings.Join(parts, "\n\n")
	}

	if softwareType := strings.TrimSpace(pc.SoftwareType); softwareType != "" {
		parts = append(parts, fmt.Sprintf("Softwaretype: %s", softwareType))
	}
	if developmentStatus := strings.TrimSpace(pc.DevelopmentStatus); developmentStatus != "" {
		parts = append(parts, fmt.Sprintf("Ontwikkelstatus: %s", developmentStatus))
	}
	if pc.Legal != nil {
		if license := strings.TrimSpace(pc.Legal.License); license != "" {
			parts = append(parts, fmt.Sprintf("Licentie: %s", license))
		}
	}
	if len(pc.Platforms) > 0 {
		parts = append(parts, fmt.Sprintf("Platforms: %s", strings.Join(pc.Platforms, ", ")))
	}
	if pc.Localisation != nil && len(pc.Localisation.AvailableLanguages) > 0 {
		parts = append(parts, fmt.Sprintf("Talen: %s", strings.Join(pc.Localisation.AvailableLanguages, ", ")))
	}
	if pc.Maintenance != nil {
		if maintenanceType := strings.TrimSpace(pc.Maintenance.Type); maintenanceType != "" {
			parts = append(parts, fmt.Sprintf("Onderhoud: %s", maintenanceType))
		}
		if len(pc.Maintenance.Contacts) > 0 {
			contacts := make([]string, 0, len(pc.Maintenance.Contacts))
			for _, contact := range pc.Maintenance.Contacts {
				if name := strings.TrimSpace(contact.Name); name != "" {
					contacts = append(contacts, name)
				}
			}
			if len(contacts) > 0 {
				parts = append(parts, fmt.Sprintf("Contacten: %s", strings.Join(contacts, ", ")))
			}
		}
	}
	if len(pc.FundedBy) > 0 {
		fundedBy := make([]string, 0, len(pc.FundedBy))
		for _, organisation := range pc.FundedBy {
			if name := strings.TrimSpace(organisation.Name); name != "" {
				fundedBy = append(fundedBy, name)
			}
		}
		if len(fundedBy) > 0 {
			parts = append(parts, fmt.Sprintf("Gefinancierd door: %s", strings.Join(fundedBy, ", ")))
		}
	}

	for _, desc := range pc.Description {
		if shortDesc := strings.TrimSpace(desc.ShortDescription); shortDesc != "" && shortDesc != strings.TrimSpace(repository.ShortDescription) {
			parts = append(parts, shortDesc)
		}
		if longDesc := strings.TrimSpace(desc.LongDescription); longDesc != "" && longDesc != strings.TrimSpace(repository.LongDescription) {
			parts = append(parts, longDesc)
		}
		if len(desc.Features) > 0 {
			parts = append(parts, fmt.Sprintf("Features: %s", strings.Join(desc.Features, ", ")))
		}
	}

	if len(parts) == 0 {
		return repositoryTitle(repository)
	}
	return strings.Join(parts, "\n\n")
}

func buildTags(cfg config, repository *models.Repository) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(cfg.defaultTags)+10)

	for _, tag := range cfg.defaultTags {
		out = appendUnique(out, tag, seen)
	}

	out = appendUnique(out, fmt.Sprintf("repository-id:%s", repository.Id), seen)

	if repository.Organisation != nil {
		out = appendUnique(out, repository.Organisation.Label, seen)
		out = appendUnique(out, repository.Organisation.Uri, seen)
	} else if repository.OrganisationID != nil {
		out = appendUnique(out, *repository.OrganisationID, seen)
	}

	forkType := util.DetectRepositoryForkType(repository)
	if forkType != "" {
		out = appendUnique(out, fmt.Sprintf("forkType:%s", forkType), seen)
	}

	pc := repository.PublicCode
	if pc == nil {
		return out
	}

	out = appendUnique(out, "publiccode", seen)

	if softwareType := strings.TrimSpace(pc.SoftwareType); softwareType != "" {
		out = appendUnique(out, fmt.Sprintf("softwareType:%s", softwareType), seen)
	}
	if developmentStatus := strings.TrimSpace(pc.DevelopmentStatus); developmentStatus != "" {
		out = appendUnique(out, fmt.Sprintf("developmentStatus:%s", developmentStatus), seen)
	}
	if pc.Legal != nil {
		if license := strings.TrimSpace(pc.Legal.License); license != "" {
			out = appendUnique(out, fmt.Sprintf("license:%s", license), seen)
		}
	}
	for _, platform := range pc.Platforms {
		out = appendUnique(out, fmt.Sprintf("platform:%s", platform), seen)
	}
	if pc.Localisation != nil {
		for _, language := range pc.Localisation.AvailableLanguages {
			out = appendUnique(out, fmt.Sprintf("language:%s", language), seen)
		}
	}

	return out
}

func appendUnique(tags []string, value string, seen map[string]struct{}) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return tags
	}
	if _, ok := seen[value]; ok {
		return tags
	}
	seen[value] = struct{}{}
	return append(tags, value)
}
