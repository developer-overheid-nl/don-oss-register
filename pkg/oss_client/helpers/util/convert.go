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
		Name:             repo.Name,
		ShortDescription: repo.ShortDescription,
		RepositoryUrl:    repo.RepositoryUrl,
		PublicCodeUrl:    repo.PublicCodeUrl,
		CreatedAt:        repo.CreatedAt,
		UpdatedAt:        repo.UpdatedAt,
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

func ToRepository(repo *models.PostRepository) *models.Repository {
	r := &models.Repository{
		Id:               uuid.NewString(),
		Name:             strings.TrimSpace(stringValue(repo.Name)),
		ShortDescription: strings.TrimSpace(stringValue(repo.Description)),
		LongDescription:  strings.TrimSpace(stringValue(repo.Description)),
		RepositoryUrl:    strings.TrimSpace(stringValue(repo.RepositoryUrl)),
		PublicCodeUrl:    "",
		UpdatedAt:        repo.UpdatedAt,
		CreatedAt:        repo.CreatedAt,
		Active:           repo.Active,
	}

	publicCodeRaw := strings.TrimSpace(stringValue(repo.PubliccodeYmlUrl))
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
			r.PublicCodeUrl = publicCodeRaw
		}

		url, name, shortDesc, longDesc := parsePublicCodeYAML(content)
		if url != "" {
			r.PublicCodeUrl = url
		}
		if name != "" {
			r.Name = name
		}
		if shortDesc != "" {
			r.ShortDescription = shortDesc
		}
		if longDesc != "" {
			r.LongDescription = longDesc
		}
	}

	return r
}

func ToGitOrganisatieSummary(gitOrg *models.GitOrganisatie) models.GitOrganisatieSummary {
	return models.GitOrganisatieSummary{
		Id:           gitOrg.Id,
		Organisation: gitOrg.Organisation,
		CodeHosting:  gitOrg.CodeHosting,
	}
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
