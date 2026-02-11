package util

import (
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
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

		url, name, shortDesc, longDesc := parsePublicCodeYAML(content)
		if url != "" {
			target.Url = url
		}
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

type publicCodeYAML struct {
	URL         any `yaml:"url"`
	Name        any `yaml:"name"`
	Description any `yaml:"description"`
}

type publicCodeLocalizedDescription struct {
	ShortDescription string `yaml:"shortDescription"`
	LongDescription  string `yaml:"longDescription"`
}

func parsePublicCodeYAML(raw string) (url, name, shortDescription, longDescription string) {
	var parsed publicCodeYAML
	if err := yaml.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", "", "", ""
	}

	desc := selectDescription(parsed.Description)
	return extractPreferredString(parsed.URL),
		extractPreferredString(parsed.Name),
		strings.TrimSpace(desc.ShortDescription),
		strings.TrimSpace(desc.LongDescription)
}

func selectDescription(descriptions any) publicCodeLocalizedDescription {
	descriptionMap, ok := asStringAnyMap(descriptions)
	if !ok || len(descriptionMap) == 0 {
		if text, ok := descriptions.(string); ok {
			trimmed := strings.TrimSpace(text)
			return publicCodeLocalizedDescription{
				ShortDescription: trimmed,
				LongDescription:  trimmed,
			}
		}
		return publicCodeLocalizedDescription{}
	}

	// Support flat descriptions:
	// description:
	//   shortDescription: ...
	//   longDescription: ...
	if looksLikeDescriptionObject(descriptionMap) {
		return parseDescriptionObject(descriptionMap)
	}

	// Support localized descriptions:
	// description:
	//   nl: { shortDescription: ... }
	//   en: { shortDescription: ... }
	for _, key := range preferredLocaleKeys(descriptionMap) {
		if parsed := parseLocalizedDescription(descriptionMap[key]); hasDescription(parsed) {
			return parsed
		}
	}

	return publicCodeLocalizedDescription{}
}

func parseLocalizedDescription(raw any) publicCodeLocalizedDescription {
	if text, ok := raw.(string); ok {
		trimmed := strings.TrimSpace(text)
		return publicCodeLocalizedDescription{
			ShortDescription: trimmed,
			LongDescription:  trimmed,
		}
	}

	descMap, ok := asStringAnyMap(raw)
	if !ok {
		return publicCodeLocalizedDescription{}
	}

	if looksLikeDescriptionObject(descMap) {
		return parseDescriptionObject(descMap)
	}

	text := selectLocalizedString(descMap)
	if text == "" {
		return publicCodeLocalizedDescription{}
	}

	return publicCodeLocalizedDescription{
		ShortDescription: text,
		LongDescription:  text,
	}
}

func looksLikeDescriptionObject(values map[string]any) bool {
	_, hasShort := values["shortDescription"]
	_, hasLong := values["longDescription"]
	return hasShort || hasLong
}

func parseDescriptionObject(values map[string]any) publicCodeLocalizedDescription {
	short := extractPreferredString(values["shortDescription"])
	long := extractPreferredString(values["longDescription"])
	if short == "" {
		short = long
	}
	return publicCodeLocalizedDescription{
		ShortDescription: short,
		LongDescription:  long,
	}
}

func hasDescription(value publicCodeLocalizedDescription) bool {
	return strings.TrimSpace(value.ShortDescription) != "" || strings.TrimSpace(value.LongDescription) != ""
}

func extractPreferredString(raw any) string {
	switch typed := raw.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		values, ok := asStringAnyMap(raw)
		if !ok {
			return ""
		}
		return selectLocalizedString(values)
	}
}

func selectLocalizedString(values map[string]any) string {
	for _, key := range preferredLocaleKeys(values) {
		if text, ok := values[key].(string); ok {
			trimmed := strings.TrimSpace(text)
			if trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func preferredLocaleKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
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
	appendMatches(func(key string) bool { return strings.HasPrefix(key, "nl-") })
	appendMatches(func(key string) bool { return key == "en" })
	appendMatches(func(key string) bool { return strings.HasPrefix(key, "en-") })
	appendMatches(func(_ string) bool { return true })

	return ordered
}

func asStringAnyMap(raw any) (map[string]any, bool) {
	switch typed := raw.(type) {
	case map[string]any:
		return typed, true
	case map[interface{}]interface{}:
		converted := make(map[string]any, len(typed))
		for key, value := range typed {
			keyString, ok := key.(string)
			if !ok {
				continue
			}
			converted[keyString] = value
		}
		return converted, len(converted) > 0
	default:
		return nil, false
	}
}

func isLikelyURL(val string) bool {
	return strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://")
}
