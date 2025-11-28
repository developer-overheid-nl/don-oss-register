package util

import (
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/google/uuid"
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
		Id:            repo.Id,
		Name:          repo.Name,
		Description:   repo.Description,
		RepositoryUrl: repo.RepositoryUrl,
		PublicCodeUrl: repo.PublicCodeUrl,
		CreatedAt:     repo.CreatedAt,
		UpdatedAt:     repo.UpdatedAt,
		Organisation:  orgSummary,
	}
}

func ToRepositoryDetail(repo *models.Repository) *models.RepositoryDetail {
	detail := &models.RepositoryDetail{
		RepositorySummary: ToRepositorySummary(repo),
	}
	return detail
}

func ToRepository(repo *models.PostRepository) *models.Repository {
	return &models.Repository{
		Id:            uuid.NewString(),
		Name:          stringValue(repo.Name),
		Description:   stringValue(repo.Description),
		RepositoryUrl: stringValue(repo.RepositoryUrl),
		PublicCodeUrl: stringValue(repo.PubliccodeYml),
		UpdatedAt:     repo.UpdatedAt,
		CreatedAt:     repo.CreatedAt,
		Active:        repo.Active,
	}
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
