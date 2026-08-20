package jobs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/jobs"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepositoriesRepo struct {
	all   []models.Repository
	saved []*models.Repository
}

func (s *stubRepositoriesRepo) AllRepositorys(_ context.Context) ([]models.Repository, error) {
	return s.all, nil
}

func (s *stubRepositoriesRepo) SaveRepository(_ context.Context, r *models.Repository) error {
	s.saved = append(s.saved, r)
	return nil
}

func (s *stubRepositoriesRepo) GetRepositorys(_ context.Context, _, _ int, _ *models.RepositoryFiltersParams, _ models.RepositorySort) ([]models.Repository, models.Pagination, error) {
	return nil, models.Pagination{}, nil
}
func (s *stubRepositoriesRepo) SearchRepositorys(_ context.Context, _, _ int, _ *string, _ string) ([]models.Repository, models.Pagination, error) {
	return nil, models.Pagination{}, nil
}

func (s *stubRepositoriesRepo) GetRepositoryByID(_ context.Context, _ string) (*models.Repository, error) {
	return nil, nil
}

func (s *stubRepositoriesRepo) SaveOrganisatie(_ *models.Organisation) error {
	return nil
}

func (s *stubRepositoriesRepo) GetOrganisations(_ context.Context, _, _ int) ([]models.Organisation, models.Pagination, error) {
	return nil, models.Pagination{}, nil
}

func (s *stubRepositoriesRepo) GetGitOrganisations(_ context.Context, _, _ int, _ *string) ([]models.GitOrganisatie, models.Pagination, error) {
	return nil, models.Pagination{}, nil
}

func (s *stubRepositoriesRepo) FindOrganisationByURI(_ context.Context, _ string) (*models.Organisation, error) {
	return nil, nil
}

func (s *stubRepositoriesRepo) FindGitOrganisationByURL(_ context.Context, _ string) (*models.GitOrganisatie, error) {
	return nil, nil
}

func (s *stubRepositoriesRepo) SaveGitOrganisatie(_ context.Context, _ *models.GitOrganisatie) error {
	return nil
}

func (s *stubRepositoriesRepo) GetRepositoryFilterCounts(_ context.Context, _ *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
	return &models.RepositoryFilterCounts{}, nil
}

func TestNewRepositoryActiveJob_DefaultStaleAfter(t *testing.T) {
	t.Setenv(jobs.EnvCrawlStaleAfterHours, "")
	repo := &stubRepositoriesRepo{}
	job := jobs.NewRepositoryActiveJob(repo)
	require.NotNil(t, job)
	assert.Equal(t, jobs.DefaultRepositoryActiveStaleAfter, job.StaleAfter())
}

func TestNewRepositoryActiveJob_EnvOverride(t *testing.T) {
	t.Setenv(jobs.EnvCrawlStaleAfterHours, "168")
	repo := &stubRepositoriesRepo{}
	job := jobs.NewRepositoryActiveJob(repo)
	require.NotNil(t, job)
	assert.Equal(t, 168*time.Hour, job.StaleAfter())
}

func TestNewRepositoryActiveJob_InvalidEnvFallsBackToDefault(t *testing.T) {
	output := captureJobLogs(t)
	t.Setenv(jobs.EnvCrawlStaleAfterHours, "not-a-number")
	repo := &stubRepositoriesRepo{}
	job := jobs.NewRepositoryActiveJob(repo)
	require.NotNil(t, job)
	assert.Equal(t, jobs.DefaultRepositoryActiveStaleAfter, job.StaleAfter())

	event := decodeJobLog(t, output.Bytes())
	assert.Equal(t, "WARN", event["level"])
	assert.Equal(t, "repository_active", event["component"])
	assert.Equal(t, "configure_stale_after", event["operation"])
	assert.Equal(t, "not-a-number", event["configured_value"])
}

func TestNewRepositoryActiveJob_ZeroEnvFallsBackToDefault(t *testing.T) {
	t.Setenv(jobs.EnvCrawlStaleAfterHours, "0")
	repo := &stubRepositoriesRepo{}
	job := jobs.NewRepositoryActiveJob(repo)
	require.NotNil(t, job)
	assert.Equal(t, jobs.DefaultRepositoryActiveStaleAfter, job.StaleAfter())
}

func captureJobLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var output bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})
	return &output
}

func decodeJobLog(t *testing.T, raw []byte) map[string]any {
	t.Helper()

	var event map[string]any
	require.NoError(t, json.Unmarshal(raw, &event))
	return event
}
