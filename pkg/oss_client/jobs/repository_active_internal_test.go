package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type activeJobRepoStub struct {
	allErr  error
	saveErr error
	all     []models.Repository
	saved   []models.Repository
}

func (s *activeJobRepoStub) AllRepositorys(_ context.Context) ([]models.Repository, error) {
	return s.all, s.allErr
}

func (s *activeJobRepoStub) SaveRepository(_ context.Context, r *models.Repository) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = append(s.saved, *r)
	return nil
}

func (s *activeJobRepoStub) GetRepositorys(_ context.Context, _, _ int, _ *models.RepositoryFiltersParams) ([]models.Repository, models.Pagination, error) {
	return nil, models.Pagination{}, nil
}

func (s *activeJobRepoStub) GetRepositoryByID(_ context.Context, _ string) (*models.Repository, error) {
	return nil, nil
}

func (s *activeJobRepoStub) SearchRepositorys(_ context.Context, _, _ int, _ *string, _ string) ([]models.Repository, models.Pagination, error) {
	return nil, models.Pagination{}, nil
}

func (s *activeJobRepoStub) SaveOrganisatie(_ *models.Organisation) error {
	return nil
}

func (s *activeJobRepoStub) GetOrganisations(_ context.Context, _, _ int) ([]models.Organisation, models.Pagination, error) {
	return nil, models.Pagination{}, nil
}

func (s *activeJobRepoStub) FindOrganisationByURI(_ context.Context, _ string) (*models.Organisation, error) {
	return nil, nil
}

func (s *activeJobRepoStub) GetGitOrganisations(_ context.Context, _, _ int, _ *string) ([]models.GitOrganisatie, models.Pagination, error) {
	return nil, models.Pagination{}, nil
}

func (s *activeJobRepoStub) FindGitOrganisationByURL(_ context.Context, _ string) (*models.GitOrganisatie, error) {
	return nil, nil
}

func (s *activeJobRepoStub) SaveGitOrganisatie(_ context.Context, _ *models.GitOrganisatie) error {
	return nil
}

func (s *activeJobRepoStub) GetRepositoryFilterCounts(_ context.Context, _ *models.RepositoryFiltersParams) (*models.RepositoryFilterCounts, error) {
	return &models.RepositoryFilterCounts{}, nil
}

func TestNextRunAtSameDayBeforeHour(t *testing.T) {
	now := time.Date(2024, 5, 1, 12, 30, 0, 0, time.Local)
	assert.Equal(t, time.Date(2024, 5, 1, 13, 0, 0, 0, time.Local), nextRunAt(now, 13))
}

func TestNextRunAtNextDayAfterHour(t *testing.T) {
	now := time.Date(2024, 5, 1, 13, 30, 0, 0, time.Local)
	assert.Equal(t, time.Date(2024, 5, 2, 13, 0, 0, 0, time.Local), nextRunAt(now, 13))
}

func TestRefreshRepositoryActiveFlagsUpdatesOnlyChangedRepositories(t *testing.T) {
	cutoff := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)
	repo := &activeJobRepoStub{
		all: []models.Repository{
			{Id: "recent-inactive", LastCrawledAt: cutoff.Add(time.Minute), Active: false},
			{Id: "old-active", LastCrawledAt: cutoff.Add(-time.Minute), Active: true},
			{Id: "recent-active", LastCrawledAt: cutoff, Active: true},
			{Id: "old-inactive", LastCrawledAt: cutoff.Add(-time.Hour), Active: false},
		},
	}
	job := &RepositoryActiveJob{repo: repo, staleAfter: time.Hour}

	err := job.refreshRepositoryActiveFlags(context.Background(), cutoff)
	require.NoError(t, err)
	require.Len(t, repo.saved, 2)
	assert.Equal(t, "recent-inactive", repo.saved[0].Id)
	assert.True(t, repo.saved[0].Active)
	assert.Equal(t, "old-active", repo.saved[1].Id)
	assert.False(t, repo.saved[1].Active)
}

func TestRefreshRepositoryActiveFlagsPropagatesErrors(t *testing.T) {
	expected := errors.New("database unavailable")
	repo := &activeJobRepoStub{allErr: expected}
	job := &RepositoryActiveJob{repo: repo}

	err := job.refreshRepositoryActiveFlags(context.Background(), time.Now())
	assert.ErrorIs(t, err, expected)

	repo = &activeJobRepoStub{
		saveErr: expected,
		all:     []models.Repository{{Id: "repo", LastCrawledAt: time.Now(), Active: false}},
	}
	job = &RepositoryActiveJob{repo: repo}

	err = job.refreshRepositoryActiveFlags(context.Background(), time.Now().Add(-time.Hour))
	assert.ErrorIs(t, err, expected)
}
