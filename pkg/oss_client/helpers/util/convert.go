package util

import (
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/google/uuid"
	publiccode "github.com/italia/publiccode-parser-go/v5"
)

func ToRepositorySummary(repo *models.Repository) models.RepositorySummary {
	var orgSummary *models.OrganisationSummary
	if repo.Organisation != nil {
		orgSummary = &models.OrganisationSummary{
			Uri:   repo.Organisation.Uri,
			Label: repo.Organisation.Label,
		}
	}
	return models.RepositorySummary{
		Id:               repo.Id,
		Url:              repo.Url,
		Name:             repo.Name,
		ShortDescription: repo.ShortDescription,
		PublicCodeUrl:    repo.PublicCodeUrl,
		CreatedAt:        repo.CreatedAt,
		LastCrawledAt:    repo.LastCrawledAt,
		LastActivityAt:   repo.LastActivityAt,
		Organisation:     orgSummary,
	}
}

func ToRepositoryDetail(repo *models.Repository) *models.RepositoryDetail {
	detail := &models.RepositoryDetail{
		RepositorySummary: ToRepositorySummary(repo),
		LongDescription:   repo.LongDescription,
	}
	return detail
}

func ToGitOrganisatieSummary(gitOrg *models.GitOrganisatie) models.GitOrganisatieSummary {
	return models.GitOrganisatieSummary{
		Id:           gitOrg.Id,
		Organisation: gitOrg.Organisation,
		Url:          gitOrg.Url,
	}
}

func ApplyRepositoryInput(target *models.Repository, input *models.RepositoryInput) *models.Repository {
	if target == nil {
		target = &models.Repository{
			Id: uuid.NewString(),
		}
	}

	if input == nil {
		return target
	}

	if input.Url != nil {
		target.Url = strings.TrimSpace(*input.Url)
	}
	if !input.CreatedAt.IsZero() {
		target.CreatedAt = input.CreatedAt
	}
	if !input.LastCrawledAt.IsZero() {
		target.LastCrawledAt = input.LastCrawledAt
	}
	if !input.LastActivityAt.IsZero() {
		target.LastActivityAt = input.LastActivityAt
	}
	if input.PublicCodeUrl != nil {
		target.PublicCodeUrl = strings.TrimSpace(*input.PublicCodeUrl)
	}
	if input.Name != nil {
		target.Name = strings.TrimSpace(*input.Name)
	}
	if input.ShortDescription != nil {
		target.ShortDescription = strings.TrimSpace(*input.ShortDescription)
		target.LongDescription = target.ShortDescription
	}

	publicCodeRaw := ""
	if input.PublicCodeUrl != nil {
		publicCodeRaw = strings.TrimSpace(*input.PublicCodeUrl)
	}

	if publicCodeRaw != "" {
		content := publicCodeRaw
		if isLikelyURL(publicCodeRaw) {
			if resp, err := http.Get(publicCodeRaw); err == nil && resp != nil {
				defer func() {
					if err := resp.Body.Close(); err != nil {
						_ = err
					}
				}()
				if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
					if body, err := io.ReadAll(resp.Body); err == nil {
						content = string(body)
					}
				}
			}
			target.PublicCodeUrl = publicCodeRaw
		}

		_, name, shortDesc, longDesc := parsePublicCodeYAML(content)
		if name != "" {
			target.Name = name
		}
		if shortDesc != "" {
			target.ShortDescription = shortDesc
			target.LongDescription = shortDesc
		}
		if longDesc != "" {
			target.LongDescription = longDesc
		}
	}

	return target
}

func parsePublicCodeYAML(raw string) (url, name, shortDescription, longDescription string) {
	parser, err := publiccode.NewParser(publiccode.ParserConfig{
		DisableExternalChecks: true,
	})
	if err != nil {
		log.Printf("publiccode parser initialization failed: %s, %v", url, err)
		return "", "", "", ""
	}

	parsed, parseErr := parser.ParseStream(strings.NewReader(strings.TrimPrefix(raw, "\ufeff")))
	if parsed == nil || hasValidationErrors(parseErr) {
		if parseErr != nil {
			log.Printf("publiccode parse validation failed: %v", parseErr)
		} else {
			log.Printf("publiccode parse validation failed: empty parse result")
		}
		return "", "", "", ""
	}

	v0, ok := asPublicCodeV0(parsed)
	if !ok {
		log.Printf("publiccode parse result is not version 0: %T", parsed)
		return "", "", "", ""
	}

	if v0.URL != nil {
		url = strings.TrimSpace(v0.URL.String())
	}

	name = strings.TrimSpace(v0.Name)

	desc := selectDescription(v0.Description)
	if name == "" && desc.LocalisedName != nil {
		name = strings.TrimSpace(*desc.LocalisedName)
	}

	shortDescription = strings.TrimSpace(desc.ShortDescription)
	longDescription = strings.TrimSpace(desc.LongDescription)
	if shortDescription == "" {
		log.Printf("publiccode description does not contain a short description for repository with url %q", url)
	}
	if longDescription == "" {
		log.Printf("publiccode description does not contain a long description for repository with url %q", url)
	}

	return url, name, shortDescription, longDescription
}

func hasValidationErrors(err error) bool {
	if err == nil {
		return false
	}

	results, ok := err.(publiccode.ValidationResults)
	if !ok {
		return true
	}

	for _, item := range results {
		if _, isError := item.(publiccode.ValidationError); isError {
			return true
		}
	}

	return false
}

func asPublicCodeV0(pc publiccode.PublicCode) (publiccode.PublicCodeV0, bool) {
	switch typed := pc.(type) {
	case publiccode.PublicCodeV0:
		return typed, true
	case *publiccode.PublicCodeV0:
		if typed == nil {
			return publiccode.PublicCodeV0{}, false
		}
		return *typed, true
	default:
		return publiccode.PublicCodeV0{}, false
	}
}

func selectDescription(descriptions map[string]publiccode.DescV0) publiccode.DescV0 {
	if len(descriptions) == 0 {
		return publiccode.DescV0{}
	}

	for _, key := range preferredLocaleKeys(descriptions) {
		value := descriptions[key]
		if strings.TrimSpace(value.ShortDescription) != "" || strings.TrimSpace(value.LongDescription) != "" {
			return value
		}
	}

	return publiccode.DescV0{}
}

func preferredLocaleKeys(descriptions map[string]publiccode.DescV0) []string {
	keys := make([]string, 0, len(descriptions))
	for key := range descriptions {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ordered := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	appendMatches := func(match func(key string) bool) {
		for _, key := range keys {
			lower := strings.ToLower(key)
			if !match(lower) {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			ordered = append(ordered, key)
		}
	}

	appendMatches(func(key string) bool { return key == "nl" })
	appendMatches(func(key string) bool { return key == "en" })
	appendMatches(func(_ string) bool { return true })

	return ordered
}

func isLikelyURL(val string) bool {
	return strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://")
}
