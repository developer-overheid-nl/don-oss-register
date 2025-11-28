package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/handler"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serviceStubRepo struct {
	listFunc            func(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repository, models.Pagination, error)
	searchFunc          func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error)
	retrieveFunc        func(ctx context.Context, id string) (*models.Repository, error)
	getOrgFunc          func(ctx context.Context, page, perPage int, ids *string) ([]models.Organisation, models.Pagination, error)
	gitOrgListFunc      func(ctx context.Context, page, perPage int, ids *string) ([]models.GitOrganisatie, models.Pagination, error)
	saveOrgFunc         func(org *models.Organisation) error
	findGitOrgByOrgFunc func(ctx context.Context, organisationURI string) (*models.GitOrganisatie, error)
	saveGitOrgFunc      func(ctx context.Context, gitOrg *models.GitOrganisatie) error
	addCodeHostingFunc  func(ctx context.Context, gitOrgID, url string, isGroup *bool) (*models.CodeHosting, error)
}

func (s *serviceStubRepo) GetRepositorys(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repository, models.Pagination, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, page, perPage, organisation, ids)
	}
	return nil, models.Pagination{}, nil
}

func (s *serviceStubRepo) SearchRepositorys(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error) {
	if s.searchFunc != nil {
		return s.searchFunc(ctx, page, perPage, organisation, query)
	}
	return []models.Repository{}, models.Pagination{}, nil
}

func (s *serviceStubRepo) GetRepositoryByID(ctx context.Context, id string) (*models.Repository, error) {
	if s.retrieveFunc != nil {
		return s.retrieveFunc(ctx, id)
	}
	return nil, nil
}

func (s *serviceStubRepo) SaveRepository(ctx context.Context, repository *models.Repository) error {
	return nil
}

func (s *serviceStubRepo) SaveOrganisatie(org *models.Organisation) error {
	if s.saveOrgFunc != nil {
		return s.saveOrgFunc(org)
	}
	return nil
}

func (s *serviceStubRepo) AllRepositorys(ctx context.Context) ([]models.Repository, error) {
	return nil, nil
}

func (s *serviceStubRepo) GetOrganisations(ctx context.Context, page, perPage int, ids *string) ([]models.Organisation, models.Pagination, error) {
	if s.getOrgFunc != nil {
		return s.getOrgFunc(ctx, page, perPage, ids)
	}
	return nil, models.Pagination{}, nil
}

func (s *serviceStubRepo) GetGitOrganisations(ctx context.Context, page, perPage int, ids *string) ([]models.GitOrganisatie, models.Pagination, error) {
	if s.gitOrgListFunc != nil {
		return s.gitOrgListFunc(ctx, page, perPage, ids)
	}
	return nil, models.Pagination{}, nil
}

func (s *serviceStubRepo) FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error) {
	return nil, nil
}

func (s *serviceStubRepo) FindGitOrganisationByOrganisationURI(ctx context.Context, organisationURI string) (*models.GitOrganisatie, error) {
	if s.findGitOrgByOrgFunc != nil {
		return s.findGitOrgByOrgFunc(ctx, organisationURI)
	}
	return nil, nil
}

func (s *serviceStubRepo) SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error {
	if s.saveGitOrgFunc != nil {
		return s.saveGitOrgFunc(ctx, gitOrg)
	}
	return nil
}

func (s *serviceStubRepo) AddCodeHosting(ctx context.Context, gitOrgID, url string, isGroup *bool) (*models.CodeHosting, error) {
	if s.addCodeHostingFunc != nil {
		return s.addCodeHostingFunc(ctx, gitOrgID, url, isGroup)
	}
	return nil, nil
}

func TestListRepositorys_HandlerSetsHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		listFunc: func(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repository, models.Pagination, error) {
			org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
			return []models.Repository{
				{Id: "repo-1", Name: "Repo One", Organisation: org},
			}, models.Pagination{TotalRecords: 1, TotalPages: 1, CurrentPage: 1, RecordsPerPage: 10}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/Repositories", nil)
	ctx.Request = req
	ctx.Params = gin.Params{}

	result, err := ctrl.ListRepositorys(ctx, &models.ListRepositorysParams{})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "1", w.Header().Get("Total-Count"))
}

func TestRetrieveRepository_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/Repositories/missing", nil)
	ctx.Request = req

	resp, err := ctrl.RetrieveRepository(ctx, &models.RepositoryParams{Id: "missing"})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestSearchRepositorys_UsesService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		searchFunc: func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error) {
			return []models.Repository{{Id: "repo-2", Organisation: &models.Organisation{Uri: "org-1"}}}, models.Pagination{TotalRecords: 1}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/Repositories/_search?q=repo", nil)
	ctx.Request = req

	results, err := ctrl.SearchRepositorys(ctx, &models.ListRepositorysSearchParams{Query: "repo"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-2", results[0].Id)
}

func TestListOrganisations_SetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		getOrgFunc: func(ctx context.Context, page, perPage int, ids *string) ([]models.Organisation, models.Pagination, error) {
			return []models.Organisation{{Uri: "org-1", Label: "Org 1"}}, models.Pagination{
				TotalRecords:   1,
				TotalPages:     1,
				CurrentPage:    page,
				RecordsPerPage: perPage,
			}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/organisations", nil)
	ctx.Request = req

	orgs, err := ctrl.ListOrganisations(ctx, &models.ListOrganisationsParams{})
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	assert.Equal(t, "1", w.Header().Get("Total-Count"))
	assert.NotEmpty(t, w.Header().Get("Link"))
}

func TestCreateOrganisation_DelegatesToService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		saveOrgFunc: func(org *models.Organisation) error {
			return nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/v1/organisations", nil)
	ctx.Request = req

	org := &models.Organisation{Uri: "https://example.org", Label: "Example"}
	resp, err := ctrl.CreateOrganisation(ctx, org)
	require.NoError(t, err)
	assert.Equal(t, "Example", resp.Label)
}
