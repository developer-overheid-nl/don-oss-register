package util

import (
	"io"
	"net/http"
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
				defer resp.Body.Close()
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
	URL         string                                    `yaml:"url"`
	Name        string                                    `yaml:"name"`
	Description map[string]publicCodeLocalizedDescription `yaml:"description"`
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
	return strings.TrimSpace(parsed.URL),
		strings.TrimSpace(parsed.Name),
		strings.TrimSpace(desc.ShortDescription),
		strings.TrimSpace(desc.LongDescription)
}

func selectDescription(descriptions map[string]publicCodeLocalizedDescription) publicCodeLocalizedDescription {
	if len(descriptions) == 0 {
		return publicCodeLocalizedDescription{}
	}
	if d, ok := descriptions["nl"]; ok {
		return d
	}
	if d, ok := descriptions["en"]; ok {
		return d
	}
	for _, d := range descriptions {
		return d
	}
	return publicCodeLocalizedDescription{}
}

func isLikelyURL(val string) bool {
	return strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://")
}
