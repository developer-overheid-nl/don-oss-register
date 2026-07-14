package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/handler"
	httpclient "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/httpclient"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serviceStubRepo struct {
	listFunc            func(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error)
	retrieveFunc        func(ctx context.Context, id string) (*models.Repository, error)
	searchFunc          func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error)
	saveRepositoryFunc  func(ctx context.Context, repository *models.Repository) error
	getOrgFunc          func(ctx context.Context, page, perPage int) ([]models.Organisation, models.Pagination, error)
	gitOrgListFunc      func(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error)
	saveOrgFunc         func(org *models.Organisation) error
	findOrgFunc         func(ctx context.Context, uri string) (*models.Organisation, error)
	findGitOrgByURLFunc func(ctx context.Context, url string) (*models.GitOrganisatie, error)
	saveGitOrgFunc      func(ctx context.Context, gitOrg *models.GitOrganisatie) error
	filterCountsFunc    func(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error)
}

func (s *serviceStubRepo) GetRepositorys(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, page, perPage, p)
	}
	return nil, models.Pagination{}, nil
}

func (s *serviceStubRepo) GetRepositoryByID(ctx context.Context, id string) (*models.Repository, error) {
	if s.retrieveFunc != nil {
		return s.retrieveFunc(ctx, id)
	}
	return nil, nil
}

func (s *serviceStubRepo) SearchRepositorys(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error) {
	if s.searchFunc != nil {
		return s.searchFunc(ctx, page, perPage, organisation, query)
	}
	return []models.Repository{}, models.Pagination{}, nil
}

func (s *serviceStubRepo) SaveRepository(ctx context.Context, repository *models.Repository) error {
	if s.saveRepositoryFunc != nil {
		return s.saveRepositoryFunc(ctx, repository)
	}
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

func (s *serviceStubRepo) GetOrganisations(ctx context.Context, page, perPage int) ([]models.Organisation, models.Pagination, error) {
	if s.getOrgFunc != nil {
		return s.getOrgFunc(ctx, page, perPage)
	}
	return nil, models.Pagination{}, nil
}

func (s *serviceStubRepo) GetGitOrganisations(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error) {
	if s.gitOrgListFunc != nil {
		return s.gitOrgListFunc(ctx, page, perPage, organisation)
	}
	return nil, models.Pagination{}, nil
}

func (s *serviceStubRepo) FindOrganisationByURI(ctx context.Context, uri string) (*models.Organisation, error) {
	if s.findOrgFunc != nil {
		return s.findOrgFunc(ctx, uri)
	}
	return nil, nil
}

func (s *serviceStubRepo) FindGitOrganisationByURL(ctx context.Context, url string) (*models.GitOrganisatie, error) {
	if s.findGitOrgByURLFunc != nil {
		return s.findGitOrgByURLFunc(ctx, url)
	}
	return nil, nil
}

func (s *serviceStubRepo) SaveGitOrganisatie(ctx context.Context, gitOrg *models.GitOrganisatie) error {
	if s.saveGitOrgFunc != nil {
		return s.saveGitOrgFunc(ctx, gitOrg)
	}
	return nil
}

func (s *serviceStubRepo) GetRepositoryFilterCounts(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
	if s.filterCountsFunc != nil {
		return s.filterCountsFunc(ctx, p)
	}
	return &models.RepositoryFilterCounts{}, nil
}

func TestListRepositorys_HandlerSetsHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		listFunc: func(ctx context.Context, page, perPage int, p *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error) {
			org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
			return []models.Repository{
				{Id: "repo-1", Name: "Repo One", Organisation: org},
			}, models.Pagination{TotalRecords: 1, TotalPages: 1, CurrentPage: 1, RecordsPerPage: 10}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/repositories", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/missing", nil)
	ctx.Request = req

	resp, err := ctrl.RetrieveRepository(ctx, &models.RepositoryParams{Id: "missing"})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestSearchRepositorys_UsesService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		searchFunc: func(ctx context.Context, page, perPage int, organisation *string, query string) ([]models.Repository, models.Pagination, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, perPage)
			assert.Nil(t, organisation)
			assert.Equal(t, "repo", query)
			return []models.Repository{{Id: "repo-2", Organisation: &models.Organisation{Uri: "org-1"}}}, models.Pagination{TotalRecords: 1}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/_search?q=repo", nil)
	ctx.Request = req

	results, err := ctrl.SearchRepositorys(ctx, &models.ListRepositorysSearchParams{Query: "repo"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-2", results[0].Id)
}

func TestListOrganisations_SetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		getOrgFunc: func(ctx context.Context, page, perPage int) ([]models.Organisation, models.Pagination, error) {
			return []models.Organisation{{Uri: "org-1", Label: "Org 1"}}, models.Pagination{
				TotalRecords:   1,
				TotalPages:     1,
				CurrentPage:    1,
				RecordsPerPage: 10,
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
}

func TestCreateOrganisation_DelegatesToService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tooiURI := "https://identifier.overheid.nl/tooi/id/oorg/oorg10111"
	tooiLabel := "KOOP"
	tooiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/ld+json")
		_ = json.NewEncoder(w).Encode([]httpclient.TooIGraph{{
			Graph: []httpclient.TooIObject{{
				ID: tooiURI,
				Label: []struct {
					Value    string `json:"@value"`
					Language string `json:"@language"`
				}{{Value: tooiLabel, Language: "nl"}},
			}},
		}})
	}))
	defer tooiServer.Close()

	orig := httpclient.HTTPClient
	defer func() { httpclient.HTTPClient = orig }()
	httpclient.HTTPClient = &http.Client{
		Transport: rewriteHostTransport(tooiServer.URL),
	}

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

	org := &models.Organisation{Uri: tooiURI, Label: "Wrong label from request"}
	resp, err := ctrl.CreateOrganisation(ctx, org)
	require.NoError(t, err)
	assert.Equal(t, tooiLabel, resp.Label)
}

func rewriteHostTransport(targetBase string) http.RoundTripper {
	return &rewriteTransport{
		base:   http.DefaultTransport,
		target: targetBase,
	}
}

type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u, _ := url.Parse(t.target)
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	return t.base.RoundTrip(req)
}

func TestCreateRepository_DelegatesToService(t *testing.T) {
	t.Setenv("ENABLE_TYPESENSE", "false")
	gin.SetMode(gin.TestMode)
	org := &models.Organisation{Uri: "https://example.org/org", Label: "Example"}
	var saved *models.Repository
	repo := &serviceStubRepo{
		findOrgFunc: func(ctx context.Context, uri string) (*models.Organisation, error) {
			return org, nil
		},
		saveRepositoryFunc: func(ctx context.Context, repository *models.Repository) error {
			saved = repository
			return nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/repositories", nil)

	url := "https://example.org/repo"
	name := "Repo"
	resp, err := ctrl.CreateRepository(ctx, &models.RepositoryInput{
		Url:             &url,
		OrganisationUri: &org.Uri,
		Name:            &name,
	})
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, "Repo", resp.Name)
	assert.Equal(t, &org.Uri, saved.OrganisationID)
}

func TestListGitOrganisations_SetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		gitOrgListFunc: func(ctx context.Context, page, perPage int, organisation *string) ([]models.GitOrganisatie, models.Pagination, error) {
			return []models.GitOrganisatie{{Id: "git-1", Url: "https://github.com/example"}}, models.Pagination{
				TotalRecords:   1,
				TotalPages:     1,
				CurrentPage:    1,
				RecordsPerPage: 20,
			}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/git-organisations", nil)

	results, err := ctrl.ListGitOrganisations(ctx, &models.ListGitOrganisationsParams{})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "1", w.Header().Get("Total-Count"))
}

func TestCreateGitOrganisation_DelegatesToService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	org := &models.Organisation{Uri: "https://example.org/org", Label: "Example"}
	repo := &serviceStubRepo{
		findOrgFunc: func(ctx context.Context, uri string) (*models.Organisation, error) {
			return org, nil
		},
		saveGitOrgFunc: func(ctx context.Context, gitOrg *models.GitOrganisatie) error {
			return nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/git-organisations", nil)

	resp, err := ctrl.CreateGitOrganisation(ctx, &models.GitOrganisationInput{
		Url:             "https://github.com/example",
		OrganisationUri: org.Uri,
	})
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/example", resp.Url)
	assert.Equal(t, &org.Uri, resp.OrganisationID)
}

func TestUpdateRepository_DelegatesToService(t *testing.T) {
	t.Setenv("ENABLE_TYPESENSE", "false")
	gin.SetMode(gin.TestMode)
	existing := &models.Repository{Id: "repo-1", Url: "https://example.org/old", Name: "Old", Active: true}
	var saved *models.Repository
	repo := &serviceStubRepo{
		retrieveFunc: func(ctx context.Context, id string) (*models.Repository, error) {
			return existing, nil
		},
		saveRepositoryFunc: func(ctx context.Context, repository *models.Repository) error {
			saved = repository
			return nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/v1/repositories/repo-1", nil)

	url := "https://example.org/new"
	name := "New"
	resp, err := ctrl.UpdateRepository(ctx, &models.UpdateRepositoryRequest{
		RepositoryParams: models.RepositoryParams{Id: "repo-1"},
		RepositoryInput:  models.RepositoryInput{Url: &url, Name: &name},
	})
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, "repo-1", saved.Id)
	assert.Equal(t, "New", resp.Name)
}

func TestListRepositoryFilters_DelegatesToService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &serviceStubRepo{
		filterCountsFunc: func(ctx context.Context, p *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
			return &models.RepositoryFilterCounts{PublicCode: 3}, nil
		},
	}
	ctrl := handler.NewOSSController(services.NewRepositoryService(repo))

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/repositories/filters", nil)

	groups, err := ctrl.ListRepositoryFilters(ctx, &models.RepositoryFiltersParams{})
	require.NoError(t, err)
	require.NotEmpty(t, groups)
	assert.Equal(t, "publiccode", groups[0].Key)
	require.NotNil(t, groups[0].Count)
	assert.Equal(t, 3, *groups[0].Count)
}
