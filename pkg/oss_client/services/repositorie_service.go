package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	util "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/util"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
)

// RepositoriesService implementeert RepositoriesServicer met de benodigde repository
type RepositoriesService struct {
	repo repositories.RepositoriesRepository
}

// NewRepositoriesService Constructor-functie
func NewRepositoriesService(repo repositories.RepositoriesRepository) *RepositoriesService {
	return &RepositoriesService{
		repo: repo,
	}
}

func (s *RepositoriesService) ListRepositories(ctx context.Context, p *models.ListRepositoriesParams) ([]models.RepositorySummary, models.Pagination, error) {
	idFilter := p.FilterIDs()
	repositories, pagination, err := s.repo.GetRepositories(ctx, p.Page, p.PerPage, p.Organisation, idFilter)
	if err != nil {
		return nil, models.Pagination{}, err
	}

	dtos := make([]models.RepositorySummary, len(repositories))
	for i, repository := range repositories {
		dtos[i] = util.ToRepositorySummary(&repository)
	}

	return dtos, pagination, nil
}

func (s *RepositoriesService) RetrieveRepositorie(ctx context.Context, id string) (*models.RepositorieDetail, error) {
	api, err := s.repo.GetRepositorieByID(ctx, id)
	if err != nil || api == nil {
		return nil, err
	}
	detail := util.ToRepositorieDetail(api)
	return detail, nil
}

func (s *RepositoriesService) SearchRepositories(ctx context.Context, p *models.ListRepositoriesSearchParams) ([]models.RepositorySummary, models.Pagination, error) {
	trimmed := strings.TrimSpace(p.Query)
	if trimmed == "" {
		return []models.RepositorySummary{}, models.Pagination{}, nil
	}
	repositories, pagination, err := s.repo.SearchRepositories(ctx, p.Page, p.PerPage, p.Organisation, trimmed)
	if err != nil {
		return nil, models.Pagination{}, err
	}
	results := make([]models.RepositorySummary, len(repositories))
	for i := range repositories {
		results[i] = util.ToRepositorySummary(&repositories[i])
	}
	return results, pagination, nil
}

func (s *RepositoriesService) CreateRepositorie(ctx context.Context, requestBody models.PostRepositorie) (*models.Repositorie, error) {
	//todo crawler aanroepen en data opslaan
	return nil, nil
}

func (s *RepositoriesService) GetAllOrganisations(ctx context.Context) ([]models.Organisation, int, error) {
	return s.repo.GetOrganisations(ctx)
}

// CreateOrganisation validates and stores a new organisation
func (s *RepositoriesService) CreateOrganisation(ctx context.Context, org *models.Organisation) (*models.Organisation, error) {
	if _, err := url.ParseRequestURI(org.Uri); err != nil {
		return nil, problem.NewBadRequest(org.Uri, fmt.Sprintf("foutieve uri: %v", err),
			problem.InvalidParam{Name: "uri", Reason: "Moet een geldige URL zijn"})
	}
	if strings.TrimSpace(org.Label) == "" {
		return nil, problem.NewBadRequest(org.Uri, "label is verplicht",
			problem.InvalidParam{Name: "label", Reason: "label is verplicht"})
	}
	if err := s.repo.SaveOrganisatie(org); err != nil {
		return nil, err
	}
	return org, nil
}
