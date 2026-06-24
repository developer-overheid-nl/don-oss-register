package typesense

import (
	"context"
	"fmt"
	"strings"

	httpclient "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/httpclient"
	util "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	commontypesense "github.com/developer-overheid-nl/don-register-common/typesense"
)

const (
	defaultCollection    = "oss-register"
	defaultDetailBaseURL = "https://oss.developer.overheid.nl/repositories"
	defaultLanguage      = "nl"
	defaultItemPriority  = 1
)

// ErrDisabled is returned when Typesense configuration is missing.
var ErrDisabled = commontypesense.ErrDisabled

type config = commontypesense.Config

func loadConfigFromEnv() config {
	return commontypesense.LoadConfigFromEnv(commontypesense.Defaults{
		Collection:    defaultCollection,
		DetailBaseURL: defaultDetailBaseURL,
		Language:      defaultLanguage,
		ItemPriority:  defaultItemPriority,
		DefaultTags:   []string{"oss-register", "repository"},
	})
}

// Enabled reports whether Typesense indexing is active based on env vars.
func Enabled() bool {
	return loadConfigFromEnv().Enabled()
}

// PublishRepository pushes the provided repository to Typesense for full-text search.
func PublishRepository(ctx context.Context, repository *models.Repository) (err error) {
	if repository == nil {
		return fmt.Errorf("typesense: repository is nil")
	}

	cfg := loadConfigFromEnv()
	if !cfg.Enabled() {
		return ErrDisabled
	}

	return commontypesense.UpsertDocument(ctx, httpclient.HTTPClient, cfg, buildDocument(cfg, repository))
}

func buildDocument(cfg config, repository *models.Repository) map[string]any {
	doc := commontypesense.BaseDocument(cfg, repository.Id)

	if title := repositoryTitle(repository); title != "" {
		doc["hierarchy.lvl0"] = title
	}
	if org := repositoryOrganisationLabel(repository); org != "" {
		doc["hierarchy.lvl1"] = org
	}
	if repository.PublicCode != nil {
		if softwareType := strings.TrimSpace(repository.PublicCode.SoftwareType); softwareType != "" {
			doc["hierarchy.lvl2"] = labelWithCode(softwareType, models.SoftwareTypeLabels)
		}
		if developmentStatus := strings.TrimSpace(repository.PublicCode.DevelopmentStatus); developmentStatus != "" {
			doc["hierarchy.lvl3"] = labelWithCode(developmentStatus, models.DevelopmentStatusLabels)
		}
		if repository.PublicCode.Legal != nil {
			if license := strings.TrimSpace(repository.PublicCode.Legal.License); license != "" {
				doc["hierarchy.lvl4"] = license
			}
		}
	}

	if _, ok := doc["hierarchy.lvl2"]; !ok {
		doc["hierarchy.lvl2"] = "Repository"
		if forkType := util.DetectRepositoryForkType(repository); forkType != "" {
			doc["hierarchy.lvl3"] = forkTypeLabel(forkType)
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

	if title := repositoryTitle(repository); title != "" {
		parts = append(parts, fmt.Sprintf("Naam: %s", title))
	}
	if shortDesc := strings.TrimSpace(repository.ShortDescription); shortDesc != "" {
		parts = append(parts, fmt.Sprintf("Beschrijving: %s", shortDesc))
	}
	if longDesc := strings.TrimSpace(repository.LongDescription); longDesc != "" && longDesc != strings.TrimSpace(repository.ShortDescription) {
		parts = append(parts, longDesc)
	}
	if repoURL := strings.TrimSpace(repository.Url); repoURL != "" {
		parts = append(parts, fmt.Sprintf("Repository URL: %s", repoURL))
	}
	if publicCodeURL := strings.TrimSpace(repository.PublicCodeUrl); publicCodeURL != "" {
		parts = append(parts, fmt.Sprintf("Publiccode URL: %s", publicCodeURL))
	}
	if org := repositoryOrganisationLabel(repository); org != "" {
		parts = append(parts, fmt.Sprintf("Organisatie: %s", org))
	}

	forkType := util.DetectRepositoryForkType(repository)
	if forkType != "" {
		parts = append(parts, fmt.Sprintf("Repositorytype: %s", forkTypeLabel(forkType)))
	}

	pc := repository.PublicCode
	if pc == nil {
		if len(parts) == 0 {
			return repositoryTitle(repository)
		}
		return strings.Join(parts, "\n\n")
	}

	if softwareType := strings.TrimSpace(pc.SoftwareType); softwareType != "" {
		parts = append(parts, fmt.Sprintf("Softwaretype: %s", labelWithCode(softwareType, models.SoftwareTypeLabels)))
	}
	if developmentStatus := strings.TrimSpace(pc.DevelopmentStatus); developmentStatus != "" {
		parts = append(parts, fmt.Sprintf("Ontwikkelstatus: %s", labelWithCode(developmentStatus, models.DevelopmentStatusLabels)))
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
			parts = append(parts, fmt.Sprintf("Onderhoud: %s", labelWithCode(maintenanceType, models.MaintenanceTypeLabels)))
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
	out := make([]string, 0, len(cfg.DefaultTags)+10)

	for _, tag := range cfg.DefaultTags {
		out = appendUnique(out, tag, seen)
	}

	if repositoryID := strings.TrimSpace(repository.Id); repositoryID != "" {
		out = appendUnique(out, fmt.Sprintf("repository-id:%s", repositoryID), seen)
	}

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
	return commontypesense.AppendUnique(tags, value, seen)
}

func labelWithCode(value string, labels map[string][2]string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if meta, ok := labels[value]; ok {
		return fmt.Sprintf("%s (%s)", meta[0], value)
	}
	return value
}

func forkTypeLabel(forkType models.RepositoryForkType) string {
	switch forkType {
	case models.RepositoryForkTypeTechnicalFork:
		return "Technische fork"
	case models.RepositoryForkTypeVariantFork:
		return "Variant fork"
	case models.RepositoryForkTypeGitFork:
		return "Git fork"
	case models.RepositoryForkTypeURLMistake:
		return "URL komt niet overeen met publiccode.yml"
	default:
		return string(forkType)
	}
}
