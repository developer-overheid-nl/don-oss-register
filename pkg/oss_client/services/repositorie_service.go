package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

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
	repositories, pagination, err := s.repo.GetRepositorys(ctx, p.Page, p.PerPage, p.Organisation, p.PublicCode)
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
	gitOrganisations, pagination, err := s.repo.GetGitOrganisations(ctx, p.Page, p.PerPage, p.Organisation)
	if err != nil {
		return nil, models.Pagination{}, err
	}

	dtos := make([]models.GitOrganisatieSummary, len(gitOrganisations))
	for i, gitOrganisation := range gitOrganisations {
		dtos[i] = util.ToGitOrganisatieSummary(&gitOrganisation)
	}
	return dtos, pagination, nil
}

func (s *RepositoryService) CreateGitOrganisatie(ctx context.Context, requestBody models.GitOrganisationInput) (*models.GitOrganisatie, error) {
	gitURL := strings.TrimSpace(requestBody.Url)
	orgURL := strings.TrimSpace(requestBody.OrganisationUri)

	if _, err := url.ParseRequestURI(gitURL); err != nil {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("url", "url", "must be a valid URL"),
		)
	}
	if _, err := url.ParseRequestURI(orgURL); err != nil {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("organisationUri", "url", "must be a valid URL"),
		)
	}

	organisation, err := s.repo.FindOrganisationByURI(ctx, orgURL)
	if err != nil {
		return nil, err
	}
	if organisation == nil {
		return nil, problem.NewNotFound("Resource does not exist")
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
	if err := validateRepositoryID(id); err != nil {
		return nil, err
	}
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
		return nil, models.Pagination{}, problem.NewBadRequest("Invalid input",
			queryError("q", "required", "q is required"),
		)
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

func (s *RepositoryService) CreateRepository(ctx context.Context, requestBody models.RepositoryInput) (*models.RepositoryDetail, error) {
	repo := util.ApplyRepositoryInput(nil, &requestBody)
	repo.Active = true

	repoURL := strings.TrimSpace(repo.Url)
	if repoURL == "" {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("url", "required", "url is required"),
		)
	}
	if _, err := url.ParseRequestURI(repoURL); err != nil {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("url", "url", "must be a valid URL"),
		)
	}

	orgURL := trimPtr(requestBody.OrganisationUri)
	if orgURL == "" {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("organisationUri", "required", "organisationUri is required"),
		)
	}
	if _, err := url.ParseRequestURI(orgURL); err != nil {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("organisationUri", "url", "must be a valid URL"),
		)
	}

	org, err := s.repo.FindOrganisationByURI(ctx, orgURL)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, problem.NewNotFound("Resource does not exist")
	}

	repo.OrganisationID = &org.Uri
	repo.Organisation = org
	if err := s.repo.SaveRepository(ctx, repo); err != nil {
		return nil, err
	}

	return util.ToRepositoryDetail(repo), nil
}

func (s *RepositoryService) UpdateRepository(ctx context.Context, id string, requestBody models.RepositoryInput) (*models.RepositoryDetail, error) {
	if err := validateRepositoryID(id); err != nil {
		return nil, err
	}
	existing, err := s.repo.GetRepositoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, problem.NewNotFound("Resource does not exist")
	}

	updated := util.ApplyRepositoryInput(existing, &requestBody)
	updated.Id = id
	updated.Active = true

	repoURL := strings.TrimSpace(updated.Url)
	if repoURL == "" {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("url", "required", "url is required"),
		)
	}
	if _, err := url.ParseRequestURI(repoURL); err != nil {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("url", "url", "must be a valid URL"),
		)
	}

	orgURL := trimPtr(requestBody.OrganisationUri)
	if orgURL != "" {
		if _, err := url.ParseRequestURI(orgURL); err != nil {
			return nil, problem.NewBadRequest("Invalid input",
				bodyError("organisationUri", "url", "must be a valid URL"),
			)
		}

		org, err := s.repo.FindOrganisationByURI(ctx, orgURL)
		if err != nil {
			return nil, err
		}
		if org == nil {
			return nil, problem.NewNotFound("Resource does not exist")
		}
		updated.OrganisationID = &org.Uri
		updated.Organisation = org
	}

	if err := s.repo.SaveRepository(ctx, updated); err != nil {
		return nil, err
	}

	return util.ToRepositoryDetail(updated), nil
}

func (s *RepositoryService) ListOrganisations(ctx context.Context, p *models.ListOrganisationsParams) ([]models.OrganisationSummary, models.Pagination, error) {
	organisations, pagination, err := s.repo.GetOrganisations(ctx, p.Page, p.PerPage)
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
	org.Uri = strings.TrimSpace(org.Uri)
	org.Label = strings.TrimSpace(org.Label)

	if _, err := url.ParseRequestURI(org.Uri); err != nil {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("uri", "url", "must be a valid URL"),
		)
	}
	if org.Label == "" {
		return nil, problem.NewBadRequest("Invalid input",
			bodyError("label", "required", "label is required"),
		)
	}
	existing, err := s.repo.FindOrganisationByURI(ctx, org.Uri)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, problem.New(http.StatusConflict, "Organisation already exists",
			bodyError("uri", "conflict", "organisation already exists"),
		)
	}
	if err := s.repo.SaveOrganisatie(org); err != nil {
		return nil, err
	}
	return org, nil
}

func bodyError(field, code, detail string) problem.ErrorDetail {
	return problem.ErrorDetail{
		In:       "body",
		Location: fmt.Sprintf("#/%s", field),
		Code:     code,
		Detail:   detail,
	}
}

func pathError(field, code, detail string) problem.ErrorDetail {
	return problem.ErrorDetail{
		In:       "path",
		Location: fmt.Sprintf("#/%s", field),
		Code:     code,
		Detail:   detail,
	}
}

func queryError(field, code, detail string) problem.ErrorDetail {
	return problem.ErrorDetail{
		In:       "query",
		Location: fmt.Sprintf("#/%s", field),
		Code:     code,
		Detail:   detail,
	}
}

func trimPtr(val *string) string {
	if val == nil {
		return ""
	}
	return strings.TrimSpace(*val)
}

func validateRepositoryID(id string) error {
	if id == "" {
		return problem.NewBadRequest("Invalid input",
			pathError("id", "required", "id is required"),
		)
	}
	if strings.ContainsRune(id, 0) {
		return problem.NewBadRequest("Invalid input",
			pathError("id", "invalid", "id must not contain NUL bytes"),
		)
	}
	if !utf8.ValidString(id) {
		return problem.NewBadRequest("Invalid input",
			pathError("id", "invalid", "id must be valid UTF-8"),
		)
	}
	return nil
}
