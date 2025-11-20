package util

import (
	"fmt"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
)

func ToRepositorySummary(repo *models.Repositorie) models.RepositorySummary {
	var orgSummary *models.OrganisationSummary
	if repo.Organisation != nil {
		orgSummary = &models.OrganisationSummary{
			Uri:   repo.Organisation.Uri,
			Label: repo.Organisation.Label,
			Links: &models.Links{
				Self: &models.Link{Href: fmt.Sprintf("/v1/repositories?organisation=%s", repo.Organisation.Uri)},
			},
		}
	}
	return models.RepositorySummary{
		Id:             repo.Id,
		Name:           repo.Name,
		Description:    repo.Description,
		RepositorieUri: repo.RepositorieUri,
		PublicCodeUrl:  repo.PublicCodeUrl,
		CreatedAt:      repo.CreatedAt,
		UpdatedAt:      repo.UpdatedAt,
		Organisation:   orgSummary,
		Links: &models.Links{
			Self: &models.Link{Href: fmt.Sprintf("/v1/repositories/%s", repo.Id)},
		},
	}
}

func ToRepositorieDetail(repo *models.Repositorie) *models.RepositorieDetail {
	detail := &models.RepositorieDetail{
		RepositorySummary: ToRepositorySummary(repo),
	}
	detail.Links = nil
	return detail
}
