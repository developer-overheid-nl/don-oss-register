package oss_client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	oss_client "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/handler"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type integrationEnv struct {
	server  *httptest.Server
	repo    repositories.RepositoriesRepository
	service *services.RepositoryService
	client  *http.Client
}

func newIntegrationEnv(t *testing.T) *integrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Organisation{}, &models.Repository{}, &models.GitOrganisatie{}))

	repo := repositories.NewRepositoriesRepository(db)
	svc := services.NewRepositoryService(repo)
	controller := handler.NewOSSController(svc)
	router := oss_client.NewRouter("test-version", controller)

	server := httptest.NewServer(router)
	t.Cleanup(func() { server.Close() })

	return &integrationEnv{
		server:  server,
		repo:    repo,
		service: svc,
		client:  &http.Client{Timeout: 2 * time.Second},
	}
}

func (e *integrationEnv) doRequest(t *testing.T, method, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, e.server.URL+path, nil)
	require.NoError(t, err)
	resp, err := e.client.Do(req)
	require.NoError(t, err)
	return resp
}

func (e *integrationEnv) doJSONRequest(t *testing.T, method, path string, payload any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if payload != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(payload))
	}
	req, err := http.NewRequest(method, e.server.URL+path, &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	}()
	var out T
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func TestRepositoriesEndpoints(t *testing.T) {
	env := newIntegrationEnv(t)
	ctx := context.Background()

	org, err := env.service.CreateOrganisation(ctx, &models.Organisation{
		Uri:   "https://example.org/organisations/integration",
		Label: "Integration Org",
	})
	require.NoError(t, err)

	repoModel := &models.Repository{
		Id:               "repo-1",
		Name:             "Integration Repo",
		ShortDescription: "Integratietest repository",
		LongDescription:  "Integratietest repository",
		OrganisationID:   &org.Uri,
		Url:              "https://example.org/repos/repo-1",
		IsFork:           true,
		PublicCodeUrl:    "https://publiccode.net/repo-1",
		LastActivityAt:   time.Date(2024, 5, 10, 12, 0, 0, 0, time.UTC),
		Active:           true,
	}
	require.NoError(t, env.repo.SaveRepository(ctx, repoModel))

	repoWithoutPublicCode := &models.Repository{
		Id:               "repo-without-publiccode",
		Name:             "Repository zonder publiccode",
		ShortDescription: "Geen publiccode.yml",
		OrganisationID:   &org.Uri,
		Url:              "https://example.org/repos/repo-without-publiccode",
		LastActivityAt:   time.Date(2024, 5, 11, 12, 0, 0, 0, time.UTC),
		Active:           true,
	}
	require.NoError(t, env.repo.SaveRepository(ctx, repoWithoutPublicCode))

	archivedRepo := &models.Repository{
		Id:               "archived-repo",
		Name:             "Archived Repo",
		ShortDescription: "Archived repository",
		OrganisationID:   &org.Uri,
		Url:              "https://example.org/repos/archived-repo",
		PublicCodeUrl:    "https://publiccode.net/archived-repo",
		LastActivityAt:   time.Date(2024, 5, 12, 12, 0, 0, 0, time.UTC),
		Active:           true,
		Archived:         true,
	}
	require.NoError(t, env.repo.SaveRepository(ctx, archivedRepo))

	t.Run("list repositories", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "test-version", resp.Header.Get("API-Version"))
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "repo-1", body[0].Id)
		require.Equal(t, repoModel.Url, body[0].Url)
		require.Equal(t, models.RepositoryForkTypeGitFork, body[0].ForkType)
		require.NotNil(t, body[0].Organisation)
		require.Equal(t, repoModel.LastActivityAt, body[0].LastActivityAt)
	})

	t.Run("list repositories supports archived true", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories?archived=true")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "archived-repo", body[0].Id)
		require.True(t, body[0].Archived)
	})

	t.Run("list repositories combines publiccode and archived filters", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories?publiccode=true&archived=true")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "archived-repo", body[0].Id)
		require.True(t, body[0].Archived)
	})

	t.Run("retrieve repository", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories/repo-1")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "test-version", resp.Header.Get("API-Version"))

		body := decodeBody[models.RepositoryDetail](t, resp)
		require.Equal(t, "repo-1", body.Id)
		require.Equal(t, models.RepositoryForkTypeGitFork, body.ForkType)
		require.NotNil(t, body.Organisation)
		require.Equal(t, repoModel.LastActivityAt, body.LastActivityAt)
	})

	t.Run("search repositories", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories?q=Integration")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "repo-1", body[0].Id)
	})

	t.Run("list repositories supports publiccode false", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories?publiccode=false")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "repo-without-publiccode", body[0].Id)
	})

	t.Run("list repositories rejects repeated publiccode values", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories?publiccode=true&publiccode=false")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	})

	t.Run("list repositories rejects repeated archived values", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories?archived=true&archived=false")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	})

	t.Run("repository filters preserve publiccode false", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories/filters?publiccode=false")
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body := decodeBody[[]models.FilterGroup](t, resp)
		require.NotEmpty(t, body)
		require.Equal(t, "publiccode", body[0].Key)
		require.Equal(t, false, body[0].Value)
	})

	t.Run("repository filters preserve archived true", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories/filters?archived=true")
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body := decodeBody[[]models.FilterGroup](t, resp)
		require.NotEmpty(t, body)
		var archivedGroup *models.FilterGroup
		for i := range body {
			if body[i].Key == "archived" {
				archivedGroup = &body[i]
			}
		}
		require.NotNil(t, archivedGroup)
		require.Equal(t, true, archivedGroup.Value)
	})

	t.Run("repository filters reject repeated publiccode values", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories/filters?publiccode=true&publiccode=false")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	})

	t.Run("repository filters reject repeated archived values", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories/filters?archived=true&archived=false")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	})

	t.Run("legacy search path remains available", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories/_search?q=Integration")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "repo-1", body[0].Id)
	})

	t.Run("list repositories supports filter endpoint query params", func(t *testing.T) {
		repoWithPublicCode := &models.Repository{
			Id:             "repo-2",
			Name:           "Library Repo",
			OrganisationID: &org.Uri,
			Url:            "https://example.org/repos/repo-2",
			PublicCodeUrl:  "https://publiccode.net/repo-2",
			PublicCode: &models.PublicCode{
				SoftwareType:      "library",
				DevelopmentStatus: "stable",
			},
			Active: true,
		}
		repoWithoutMatch := &models.Repository{
			Id:             "repo-3",
			Name:           "Other Repo",
			OrganisationID: &org.Uri,
			Url:            "https://example.org/repos/repo-3",
			PublicCodeUrl:  "https://publiccode.net/repo-3",
			PublicCode: &models.PublicCode{
				SoftwareType:      "library",
				DevelopmentStatus: "stable",
			},
			Active: true,
		}
		require.NoError(t, env.repo.SaveRepository(ctx, repoWithPublicCode))
		require.NoError(t, env.repo.SaveRepository(ctx, repoWithoutMatch))

		resp := env.doRequest(t, http.MethodGet, "/v1/repositories?q=Library&softwareType=library&developmentStatus=stable")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "repo-2", body[0].Id)
	})

	t.Run("list organisations", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/organisations")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.OrganisationSummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, org.Uri, body[0].Uri)
	})

	t.Run("openapi marks legacy search path deprecated", func(t *testing.T) {
		data, err := os.ReadFile("../../api/openapi.json")
		require.NoError(t, err)

		var spec map[string]any
		require.NoError(t, json.Unmarshal(data, &spec))
		paths := spec["paths"].(map[string]any)
		searchPath := paths["/repositories/_search"].(map[string]any)
		searchGet := searchPath["get"].(map[string]any)
		require.Equal(t, true, searchGet["deprecated"])
		require.Equal(t, "searchRepositories", searchGet["operationId"])
		params := searchGet["parameters"].([]any)
		qParam := params[len(params)-1].(map[string]any)
		require.Equal(t, "q", qParam["name"])
		require.Equal(t, true, qParam["required"])
	})

	t.Run("create organisation validation", func(t *testing.T) {
		resp := env.doJSONRequest(t, http.MethodPost, "/v1/organisations", map[string]string{
			"uri":   "notaurl",
			"label": "Test",
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("method not allowed returns 405 with RFC7807 envelope", func(t *testing.T) {
		// Send a PATCH request to an existing route that only supports GET
		resp := env.doRequest(t, http.MethodPatch, "/v1/repositories")
		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		require.Equal(t, "test-version", resp.Header.Get("API-Version"))
		require.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))

		// Decode and verify RFC7807 problem response
		type problemResponse struct {
			Status int    `json:"status"`
			Title  string `json:"title"`
		}
		body := decodeBody[problemResponse](t, resp)
		require.Equal(t, http.StatusMethodNotAllowed, body.Status)
		require.Equal(t, "Method not allowed", body.Title)
	})
}
