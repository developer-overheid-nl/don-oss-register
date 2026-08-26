package services

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	httpclient "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/httpclient"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishToTypesenseLogsIndexFailureAsError(t *testing.T) {
	var output bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("TYPESENSE_ENDPOINT", server.URL)
	t.Setenv("TYPESENSE_API_KEY", "secret")
	t.Setenv("TYPESENSE_COLLECTION", "oss-register")
	t.Setenv("ENABLE_TYPESENSE", "true")

	previousClient := httpclient.HTTPClient
	httpclient.HTTPClient = server.Client()
	t.Cleanup(func() { httpclient.HTTPClient = previousClient })

	service := &RepositoryService{}
	service.publishToTypesense(models.Repository{Id: "repo-failed", Name: "Failed repo"})

	var event map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &event))
	assert.Equal(t, "ERROR", event["level"])
	assert.Equal(t, "typesense", event["component"])
	assert.Equal(t, "index_repository", event["operation"])
	assert.Equal(t, "repo-failed", event["repository_id"])
	assert.NotEmpty(t, event["error"])
}
