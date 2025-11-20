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
	listFunc     func(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repositorie, models.Pagination, error)
	searchFunc   func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repositorie, models.Pagination, error)
	retrieveFunc func(ctx context.Context, id string) (*models.Repositorie, error)
	getOrgFunc   func(ctx context.Context) ([]models.Organisation, int, error)
	saveOrgFunc  func(org *models.Organisation) error
}

func (s *serviceStubRepo) GetRepositories(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repositorie, models.Pagination, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, page, perPage, organisation, ids)
	}
	return nil, models.Pagination{}, nil
}

func (s *serviceStubRepo) SearchRepositories(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repositorie, models.Pagination, error) {
	if s.searchFunc != nil {
		return s.searchFunc(ctx, page, perPage, organisation, query)
	}
	return []models.Repositorie{}, models.Pagination{}, nil
}

func (s *serviceStubRepo) GetRepositorieByID(ctx context.Context, id string) (*models.Repositorie, error) {
	if s.retrieveFunc != nil {
		return s.retrieveFunc(ctx, id)
	}
	return nil, nil
}

func (s *serviceStubRepo) SaveRepositorie(ctx context.Context, repository *models.Repositorie) error {
	return nil
}

func (s *serviceStubRepo) SaveOrganisatie(org *models.Organisation) error {
	if s.saveOrgFunc != nil {
		return s.saveOrgFunc(org)
	}
	return nil
}

func (s *serviceStubRepo) AllRepositories(ctx context.Context) ([]models.Repositorie, error) {
	return nil, nil
}

func (s *serviceStubRepo) GetOrganisations(ctx context.Context) ([]models.Organisation, int, error) {
	if s.getOrgFunc != nil {
		return s.getOrgFunc(ctx)
	}
	return nil, 0, nil
}

func (s *serviceStubRepo) FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error) {
	return nil, nil
}

func TestListRepositories_HandlerSetsHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		listFunc: func(ctx context.Context, page, perPage int, organisation *string, ids *string) ([]models.Repositorie, models.Pagination, error) {
			org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
			return []models.Repositorie{
				{Id: "repo-1", Name: "Repo One", Organisation: org},
			}, models.Pagination{TotalRecords: 1, TotalPages: 1, CurrentPage: 1, RecordsPerPage: 10}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoriesService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/repositories", nil)
	ctx.Request = req
	ctx.Params = gin.Params{}

	result, err := ctrl.ListRepositories(ctx, &models.ListRepositoriesParams{})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "1", w.Header().Get("Total-Count"))
}

func TestRetrieveRepositorie_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{}
	ctrl := handler.NewOSSController(services.NewRepositoriesService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/missing", nil)
	ctx.Request = req

	resp, err := ctrl.RetrieveRepositorie(ctx, &models.RepositorieParams{Id: "missing"})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestSearchRepositories_UsesService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		searchFunc: func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repositorie, models.Pagination, error) {
			return []models.Repositorie{{Id: "repo-2", Organisation: &models.Organisation{Uri: "org-1"}}}, models.Pagination{TotalRecords: 1}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoriesService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/_search?q=repo", nil)
	ctx.Request = req

	results, err := ctrl.SearchRepositories(ctx, &models.ListRepositoriesSearchParams{Query: "repo"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-2", results[0].Id)
}

func TestGetAllOrganisations_SetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		getOrgFunc: func(ctx context.Context) ([]models.Organisation, int, error) {
			return []models.Organisation{{Uri: "org-1", Label: "Org 1"}}, 1, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoriesService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/organisations", nil)
	ctx.Request = req

	orgs, err := ctrl.GetAllOrganisations(ctx)
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	assert.Equal(t, "1", w.Header().Get("Total-Count"))
}

func TestCreateOrganisation_DelegatesToService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		saveOrgFunc: func(org *models.Organisation) error {
			return nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoriesService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/v1/organisations", nil)
	ctx.Request = req

	org := &models.Organisation{Uri: "https://example.org", Label: "Example"}
	resp, err := ctrl.CreateOrganisation(ctx, org)
	require.NoError(t, err)
	assert.Equal(t, "Example", resp.Label)
}
