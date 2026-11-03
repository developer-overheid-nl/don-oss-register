package oss_client_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const completedRunReport = `{
  "schema_version": 1,
  "status": "completed",
  "repository": "https://gitlab.com/example/project.git",
  "revision": "cd9f408ae99dcc00eeac08743c2de55f1d32c6a0",
  "config_digest": "sha256:abc",
  "ort_image": "ghcr.io/oss-review-toolkit/ort:92.4.0",
  "runner_version": "3e8ab3b",
  "started_at": "2026-10-20T12:42:52Z",
  "finished_at": "2026-10-20T12:46:30Z",
  "package_count": 72,
  "dependencies": {
    "direct_count": 13,
    "transitive_count": 59,
    "projects": [
      {"package_manager": "GoMod", "definition_file": "go.mod", "direct_count": 11, "transitive_count": 15},
      {"package_manager": "GoMod", "definition_file": "test/mtls/go.mod", "direct_count": 2, "transitive_count": 45}
    ]
  },
  "stages": [{"name": "checkout", "status": "completed"}],
  "vulnerabilities": [{"package_id": "Go::golang.org/x/net:0.57.0", "id": "GO-2026-6603", "summary": "HTTP/2 issue"}],
  "findings": [
    {"rule": "VULNERABILITY_IN_DEPENDENCY", "severity": "WARNING", "message": "The package 'Go::golang.org/x/net:0.57.0' has a vulnerability."},
    {"rule": "MISSING_LICENSE_FILE", "severity": "ERROR", "message": "The project's code repository does not contain the file 'LICENSE'."},
    {"rule": "MISSING_CHANGELOG_FILE", "severity": "WARNING", "message": "No changelog."}
  ],
  "rules": [
    {"rule": "VULNERABILITY_IN_DEPENDENCY", "severity": "WARNING", "scope": "dependency", "passed": false},
    {"rule": "MISSING_README_FILE", "severity": "ERROR", "scope": "repository", "passed": true},
    {"rule": "MISSING_LICENSE_FILE", "severity": "ERROR", "scope": "repository", "passed": false},
    {"rule": "MISSING_CHANGELOG_FILE", "severity": "WARNING", "scope": "repository", "passed": false}
  ]
}`

func seedScanRepository(t *testing.T, env *integrationEnv) {
	t.Helper()
	require.NoError(t, env.repo.SaveRepository(context.Background(), &models.Repository{
		Id: "repo-1", Name: "project", Url: "https://gitlab.com/example/project", Active: true,
		PublicCodeUrl: "https://gitlab.com/example/project/-/raw/main/publiccode.yml",
	}))
}

func scanSubmission(repositoryID, report string) map[string]any {
	return map[string]any{"schemaVersion": 1, "repositoryId": repositoryID, "scan": json.RawMessage(report)}
}

func TestSubmitScanExposesSummaryWithoutVulnerabilities(t *testing.T) {
	env := newIntegrationEnv(t)
	seedScanRepository(t, env)

	resp := env.doJSONRequestWithHeaders(t, http.MethodPost, "/v1/scans", scanSubmission("repo-1", completedRunReport),
		map[string]string{"Idempotency-Key": "ort:1"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	created := decodeBody[map[string]any](t, resp)
	assert.Equal(t, "completed", created["status"])

	// A retried delivery with the same key does not add a second scan.
	resp = env.doJSONRequestWithHeaders(t, http.MethodPost, "/v1/scans", scanSubmission("repo-1", completedRunReport),
		map[string]string{"Idempotency-Key": "ort:1"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	resp = env.doRequest(t, http.MethodGet, "/v1/repositories")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	list := decodeBody[[]map[string]any](t, resp)
	require.Len(t, list, 1)
	summary := list[0]["scanSummary"].(map[string]any)
	assert.Equal(t, "completed", summary["status"])
	assert.Equal(t, "2026-10-20T12:46:30Z", summary["scannedAt"])
	assert.Equal(t, float64(33), summary["repoScorePercentage"])
	assert.Equal(t, float64(72), summary["packageCount"])
	assert.Equal(t, float64(13), summary["directDependencyCount"])
	assert.Equal(t, float64(59), summary["transitiveDependencyCount"])
	assert.Equal(t, []any{"GoMod"}, summary["packageManagers"])
	assert.Equal(t, map[string]any{"error": float64(1), "warning": float64(1), "hint": float64(0)}, summary["findingCounts"])
	assert.NotContains(t, summary, "findings")

	resp = env.doRequest(t, http.MethodGet, "/v1/repositories/repo-1")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	raw := readBody(t, resp)
	assert.NotContains(t, raw, "GO-2026-6603")
	assert.NotContains(t, raw, "VULNERABILITY_IN_DEPENDENCY")
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &detail))
	scan := detail["scan"].(map[string]any)
	assert.Equal(t, "cd9f408ae99dcc00eeac08743c2de55f1d32c6a0", scan["revision"])
	assert.Equal(t, float64(33), scan["repoScorePercentage"])
	assert.Len(t, scan["projects"], 2)
	findings := scan["findings"].([]any)
	require.Len(t, findings, 3)
	assert.Equal(t, map[string]any{"rule": "MISSING_README_FILE", "severity": "ERROR", "passed": true}, withoutURL(findings[0]))
	assert.Equal(t, map[string]any{"rule": "MISSING_LICENSE_FILE", "severity": "ERROR", "passed": false,
		"message": "The project's code repository does not contain the file 'LICENSE'."}, withoutURL(findings[1]))
	assert.Equal(t, map[string]any{"ortImage": "ghcr.io/oss-review-toolkit/ort:92.4.0", "runnerVersion": "3e8ab3b", "configDigest": "sha256:abc"}, scan["tooling"])
}

func TestLatestScanIsShown(t *testing.T) {
	env := newIntegrationEnv(t)
	seedScanRepository(t, env)

	failed := `{"schema_version": 1, "status": "failed", "error": "analyze: time budget of 30m exceeded",
	  "started_at": "2026-10-27T12:00:00Z", "finished_at": "2026-10-27T12:30:00Z",
	  "stages": [{"name": "checkout", "status": "completed"}, {"name": "analyze", "status": "failed"}],
	  "findings": [], "vulnerabilities": []}`
	for key, report := range map[string]string{"ort:new": failed, "ort:old": completedRunReport} {
		resp := env.doJSONRequestWithHeaders(t, http.MethodPost, "/v1/scans", scanSubmission("repo-1", report),
			map[string]string{"Idempotency-Key": key})
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		_ = resp.Body.Close()
	}

	resp := env.doRequest(t, http.MethodGet, "/v1/repositories/repo-1")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	scan := decodeBody[map[string]any](t, resp)["scan"].(map[string]any)
	assert.Equal(t, "failed", scan["status"])
	assert.Equal(t, "analyze", scan["failedStage"])
	assert.Equal(t, "analyze: time budget of 30m exceeded", scan["error"])
	assert.NotContains(t, scan, "repoScorePercentage")
}

func TestSubmitScanValidatesInput(t *testing.T) {
	env := newIntegrationEnv(t)
	seedScanRepository(t, env)

	tests := map[string]struct {
		body   map[string]any
		status int
	}{
		"unknown repository":   {scanSubmission("missing", completedRunReport), http.StatusNotFound},
		"unsupported version":  {map[string]any{"schemaVersion": 2, "repositoryId": "repo-1", "scan": json.RawMessage(completedRunReport)}, http.StatusBadRequest},
		"missing scan":         {map[string]any{"schemaVersion": 1, "repositoryId": "repo-1"}, http.StatusBadRequest},
		"unknown scan status":  {scanSubmission("repo-1", `{"status": "running", "started_at": "2026-10-20T12:42:52Z"}`), http.StatusBadRequest},
		"scan is not a report": {scanSubmission("repo-1", `[1, 2]`), http.StatusBadRequest},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp := env.doJSONRequest(t, http.MethodPost, "/v1/scans", tt.body)
			assert.Equal(t, tt.status, resp.StatusCode)
			_ = resp.Body.Close()
		})
	}
}

func TestRepositoryWithoutScanHasNoSummary(t *testing.T) {
	env := newIntegrationEnv(t)
	seedScanRepository(t, env)

	resp := env.doRequest(t, http.MethodGet, "/v1/repositories")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	list := decodeBody[[]map[string]any](t, resp)
	require.Len(t, list, 1)
	assert.NotContains(t, list[0], "scanSummary")
}

func withoutURL(value any) map[string]any {
	finding := value.(map[string]any)
	delete(finding, "howToFixUrl")
	return finding
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(data)
}
