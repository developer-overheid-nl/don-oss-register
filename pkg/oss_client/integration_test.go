package oss_client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	service *services.RepositoriesService
	client  *http.Client
}

func newIntegrationEnv(t *testing.T) *integrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Organisation{}, &models.Repositorie{}))

	repo := repositories.NewRepositoriesRepository(db)
	svc := services.NewRepositoriesService(repo)
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
	defer resp.Body.Close()
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

	repoModel := &models.Repositorie{
		Id:             "repo-1",
		Name:           "Integration Repo",
		Description:    "Integratietest repository",
		OrganisationID: &org.Uri,
		RepositorieUri: "https://example.org/repos/repo-1",
		PublicCodeUrl:  "https://publiccode.net/repo-1",
	}
	require.NoError(t, env.repo.SaveRepositorie(ctx, repoModel))

	t.Run("list repositories", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "test-version", resp.Header.Get("API-Version"))
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "repo-1", body[0].Id)
		require.NotNil(t, body[0].Organisation)
	})

	t.Run("retrieve repository", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories/repo-1")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "test-version", resp.Header.Get("API-Version"))

		body := decodeBody[models.RepositorieDetail](t, resp)
		require.Equal(t, "repo-1", body.Id)
		require.NotNil(t, body.Organisation)
	})

	t.Run("search repositories", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/repositories/_search?q=Integration")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.RepositorySummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, "repo-1", body[0].Id)
	})

	t.Run("list organisations", func(t *testing.T) {
		resp := env.doRequest(t, http.MethodGet, "/v1/organisations")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "1", resp.Header.Get("Total-Count"))

		body := decodeBody[[]models.OrganisationSummary](t, resp)
		require.Len(t, body, 1)
		require.Equal(t, org.Uri, body[0].Uri)
		require.NotNil(t, body[0].Links)
	})

	t.Run("create organisation validation", func(t *testing.T) {
		resp := env.doJSONRequest(t, http.MethodPost, "/v1/organisations", map[string]string{
			"uri":   "notaurl",
			"label": "Test",
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
