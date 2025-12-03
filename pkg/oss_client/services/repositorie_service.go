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
	"github.com/google/uuid"
)

// RepositoryService implementeert RepositoriesServicer met de benodigde repository
type RepositoryService struct {
	repo repositories.RepositoriesRepository
}

// NewRepositoryService Constructor-functie
func NewRepositoryService(repo repositories.RepositoriesRepository) *RepositoryService {
	return &RepositoryService{
		repo: repo,
	}
}

func (s *RepositoryService) ListRepositorys(ctx context.Context, p *models.ListRepositorysParams) ([]models.RepositorySummary, models.Pagination, error) {
	idFilter := p.FilterIDs()
	repositories, pagination, err := s.repo.GetRepositorys(ctx, p.Page, p.PerPage, p.Organisation, idFilter)
	if err != nil {
		return nil, models.Pagination{}, err
	}

	dtos := make([]models.RepositorySummary, len(repositories))
	for i, repository := range repositories {
		dtos[i] = util.ToRepositorySummary(&repository)
	}

	return dtos, pagination, nil
}

func (s *RepositoryService) ListGitOrganisations(ctx context.Context, p *models.ListGitOrganisationsParams) ([]models.GitOrganisatieSummary, models.Pagination, error) {
	gitOrganisations, pagination, err := s.repo.GetGitOrganisations(ctx, p.Page, p.PerPage, p.Ids)
	if err != nil {
		return nil, models.Pagination{}, err
	}

	dtos := make([]models.GitOrganisatieSummary, len(gitOrganisations))
	for i, gitOrganisation := range gitOrganisations {
		dtos[i] = util.ToGitOrganisatieSummary(&gitOrganisation)
	}
	return dtos, pagination, nil
}

func (s *RepositoryService) CreateGitOrganisatie(ctx context.Context, requestBody models.PostGitOrganisatie) (*models.GitOrganisatie, error) {
	gitURL := strings.TrimSpace(requestBody.GitOrganisationUrl)
	orgURL := strings.TrimSpace(requestBody.OrganisationUrl)

	if _, err := url.ParseRequestURI(gitURL); err != nil {
		return nil, problem.NewBadRequest(gitURL, fmt.Sprintf("foutieve git url: %v", err),
			problem.InvalidParam{Name: "gitOrganisationUrl", Reason: "Moet een geldige URL zijn"})
	}
	if _, err := url.ParseRequestURI(orgURL); err != nil {
		return nil, problem.NewBadRequest(orgURL, fmt.Sprintf("foutieve organisation url: %v", err),
			problem.InvalidParam{Name: "organisationUrl", Reason: "Moet een geldige URL zijn"})
	}

	organisation, err := s.repo.FindOrganisationByURI(ctx, orgURL)
	if err != nil {
		return nil, err
	}
	if organisation == nil {
		return nil, problem.NewNotFound(orgURL, "Organisation not found")
	}

	existingByURL, err := s.repo.FindGitOrganisationByURL(ctx, gitURL)
	if err != nil {
		return nil, err
	}
	if existingByURL != nil {
		return existingByURL, nil
	}

	gitOrg := &models.GitOrganisatie{
		Id:             uuid.NewString(),
		OrganisationID: &organisation.Uri,
		Organisation:   organisation,
		Url:            gitURL,
	}
	if err := s.repo.SaveGitOrganisatie(ctx, gitOrg); err != nil {
		return nil, err
	}

	return gitOrg, nil
}

func (s *RepositoryService) RetrieveRepository(ctx context.Context, id string) (*models.RepositoryDetail, error) {
	api, err := s.repo.GetRepositoryByID(ctx, id)
	if err != nil || api == nil {
		return nil, err
	}
	detail := util.ToRepositoryDetail(api)
	return detail, nil
}

func (s *RepositoryService) SearchRepositorys(ctx context.Context, p *models.ListRepositorysSearchParams) ([]models.RepositorySummary, models.Pagination, error) {
	trimmed := strings.TrimSpace(p.Query)
	if trimmed == "" {
		return []models.RepositorySummary{}, models.Pagination{}, nil
	}
	repositories, pagination, err := s.repo.SearchRepositorys(ctx, p.Page, p.PerPage, p.Organisation, trimmed)
	if err != nil {
		return nil, models.Pagination{}, err
	}
	results := make([]models.RepositorySummary, len(repositories))
	for i := range repositories {
		results[i] = util.ToRepositorySummary(&repositories[i])
	}
	return results, pagination, nil
}

func (s *RepositoryService) CreateRepository(ctx context.Context, requestBody models.PostRepository) (*models.RepositoryDetail, error) {
	if (requestBody.PubliccodeYmlUrl == nil) && (requestBody.Description == nil && requestBody.Name == nil) {
		return nil, problem.NewBadRequest("repository", "name en description zijn verplicht zonder publiccodeYml",
			problem.InvalidParam{Name: "name", Reason: "is verplicht zonder publiccodeYml"},
			problem.InvalidParam{Name: "description", Reason: "is verplicht zonder publiccodeYml"},
		)
	}

	repo := util.ToRepository(&requestBody)
	if err := s.repo.SaveRepository(ctx, repo); err != nil {
		return nil, err
	}

	return util.ToRepositoryDetail(repo), nil
}

func (s *RepositoryService) ListOrganisations(ctx context.Context, p *models.ListOrganisationsParams) ([]models.OrganisationSummary, models.Pagination, error) {
	organisations, pagination, err := s.repo.GetOrganisations(ctx, p.Page, p.PerPage, p.FilterIDs())
	if err != nil {
		return nil, models.Pagination{}, err
	}

	orgSummaries := make([]models.OrganisationSummary, len(organisations))
	for i, org := range organisations {
		orgSummaries[i] = models.OrganisationSummary(org)
	}

	return orgSummaries, pagination, nil
}

// CreateOrganisation validates and stores a new organisation
func (s *RepositoryService) CreateOrganisation(ctx context.Context, org *models.Organisation) (*models.Organisation, error) {
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
