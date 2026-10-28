package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildScanDetailWithoutRulesManifestListsOnlyViolations(t *testing.T) {
	received := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	detail := services.BuildScanDetail(models.RunReport{
		Status: "completed",
		Findings: []models.RunFinding{
			{Rule: "VULNERABILITY_IN_DEPENDENCY", Severity: "WARNING", Message: "vulnerable"},
			{Rule: "MISSING_SECURITY_FILE", Severity: "ERROR", Message: "missing"},
			{Rule: "CUSTOM_HINT", Severity: "HINT", Message: "hint"},
		},
	}, received)

	assert.Equal(t, received, detail.ScannedAt)
	assert.Nil(t, detail.RepoScorePercentage)
	assert.Nil(t, detail.DirectDependencyCount)
	assert.Equal(t, models.ScanFindingCounts{Error: 1, Hint: 1}, detail.FindingCounts)
	assert.Equal(t, []models.ScanFinding{
		{Rule: "MISSING_SECURITY_FILE", Severity: "ERROR", Message: "missing",
			HowToFixUrl: "https://developer.overheid.nl/kennisbank/open-source/standaarden/security"},
		{Rule: "CUSTOM_HINT", Severity: "HINT", Message: "hint"},
	}, detail.Findings)
}

func TestBuildScanDetailRoundsScore(t *testing.T) {
	rules := []models.RunRule{
		{Rule: "A", Severity: "ERROR", Scope: "repository", Passed: true},
		{Rule: "B", Severity: "ERROR", Scope: "repository", Passed: true},
		{Rule: "C", Severity: "WARNING", Scope: "repository", Passed: false},
		{Rule: "VULNERABILITY_IN_DEPENDENCY", Severity: "WARNING", Scope: "dependency", Passed: false},
	}
	detail := services.BuildScanDetail(models.RunReport{Status: "completed", Rules: rules}, time.Now())

	if assert.NotNil(t, detail.RepoScorePercentage) {
		assert.Equal(t, 67, *detail.RepoScorePercentage)
	}
	assert.Len(t, detail.Findings, 3)
	assert.Equal(t, models.ScanFindingCounts{Warning: 1}, detail.FindingCounts)
}

type fakeScans struct {
	saved  []models.RepositoryScan
	latest map[string]models.RepositoryScan
}

func (f *fakeScans) SaveScan(_ context.Context, scan *models.RepositoryScan) error {
	f.saved = append(f.saved, *scan)
	return nil
}

func (f *fakeScans) FindScanByIdempotencyKey(_ context.Context, key string) (*models.RepositoryScan, error) {
	for _, scan := range f.saved {
		if scan.IdempotencyKey != nil && *scan.IdempotencyKey == key {
			return &scan, nil
		}
	}
	return nil, nil
}

func (f *fakeScans) LatestScans(_ context.Context, ids []string) (map[string]models.RepositoryScan, error) {
	result := map[string]models.RepositoryScan{}
	for _, id := range ids {
		if scan, ok := f.latest[id]; ok {
			result[id] = scan
		}
	}
	return result, nil
}

func scanServiceWithRepository(scans *fakeScans) *services.RepositoryService {
	repo := &stubRepo{retrieveFunc: func(_ context.Context, id string) (*models.Repository, error) {
		if id != "repo-1" {
			return nil, nil
		}
		return &models.Repository{Id: "repo-1", Name: "Repo"}, nil
	}}
	return services.NewRepositoryService(repo).WithScans(scans)
}

func TestSubmitScanStoresOnceAndValidates(t *testing.T) {
	scans := &fakeScans{}
	svc := scanServiceWithRepository(scans)
	ctx := context.Background()
	report := json.RawMessage(`{"status": "completed", "revision": "abc", "finished_at": "2026-10-20T12:46:30Z",
	  "rules": [{"rule": "MISSING_README_FILE", "severity": "ERROR", "scope": "repository", "passed": true}]}`)

	detail, err := svc.SubmitScan(ctx, models.ScanSubmission{IdempotencyKey: "ort:1", SchemaVersion: 1, RepositoryID: "repo-1", Scan: report})
	require.NoError(t, err)
	assert.Equal(t, "completed", detail.Status)
	require.NotNil(t, detail.RepoScorePercentage)
	assert.Equal(t, 100, *detail.RepoScorePercentage)
	require.Len(t, scans.saved, 1)
	assert.Equal(t, "repo-1", scans.saved[0].RepositoryID)
	assert.Equal(t, string(report), scans.saved[0].Report)

	replay, err := svc.SubmitScan(ctx, models.ScanSubmission{IdempotencyKey: "ort:1", SchemaVersion: 1, RepositoryID: "repo-1", Scan: report})
	require.NoError(t, err)
	assert.Equal(t, detail.ScannedAt, replay.ScannedAt)
	assert.Len(t, scans.saved, 1, "a replayed delivery must not store a second scan")

	for name, tt := range map[string]struct {
		submission models.ScanSubmission
		status     int
	}{
		"schema version": {models.ScanSubmission{SchemaVersion: 2, RepositoryID: "repo-1", Scan: report}, http.StatusBadRequest},
		"not an object":  {models.ScanSubmission{SchemaVersion: 1, RepositoryID: "repo-1", Scan: json.RawMessage(`"x"`)}, http.StatusBadRequest},
		"unknown status": {models.ScanSubmission{SchemaVersion: 1, RepositoryID: "repo-1", Scan: json.RawMessage(`{"status": "running"}`)}, http.StatusBadRequest},
		"empty id":       {models.ScanSubmission{SchemaVersion: 1, RepositoryID: " ", Scan: report}, http.StatusBadRequest},
		"unknown repo":   {models.ScanSubmission{SchemaVersion: 1, RepositoryID: "repo-2", Scan: report}, http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := svc.SubmitScan(ctx, tt.submission)
			var problemErr problem.ProblemJSON
			require.ErrorAs(t, err, &problemErr)
			assert.Equal(t, tt.status, problemErr.Status)
		})
	}

	_, err = services.NewRepositoryService(&stubRepo{}).SubmitScan(ctx, models.ScanSubmission{SchemaVersion: 1, RepositoryID: "repo-1", Scan: report})
	require.Error(t, err, "scans must be enabled")
}

func TestLatestScanIsAttachedToListAndDetail(t *testing.T) {
	summary, err := json.Marshal(services.BuildScanDetail(models.RunReport{Status: "completed", Revision: "abc"}, time.Now()))
	require.NoError(t, err)
	scans := &fakeScans{latest: map[string]models.RepositoryScan{
		"repo-1": {ID: "scan-1", RepositoryID: "repo-1", Summary: string(summary)},
		"repo-2": {ID: "scan-2", RepositoryID: "repo-2", Summary: "not json"},
	}}
	repo := &stubRepo{
		listFunc: func(context.Context, int, int, *models.RepositoryFiltersParams, models.RepositorySort) ([]models.Repository, models.Pagination, error) {
			return []models.Repository{{Id: "repo-1"}, {Id: "repo-2"}, {Id: "repo-3"}}, models.Pagination{}, nil
		},
		retrieveFunc: func(_ context.Context, id string) (*models.Repository, error) {
			return &models.Repository{Id: id}, nil
		},
	}
	svc := services.NewRepositoryService(repo).WithScans(scans)
	ctx := context.Background()

	list, _, err := svc.ListRepositorys(ctx, &models.ListRepositorysParams{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, list, 3)
	require.NotNil(t, list[0].ScanSummary)
	assert.Equal(t, "completed", list[0].ScanSummary.Status)
	assert.Nil(t, list[1].ScanSummary, "an unreadable summary is skipped")
	assert.Nil(t, list[2].ScanSummary)

	detail, err := svc.RetrieveRepository(ctx, "repo-1")
	require.NoError(t, err)
	require.NotNil(t, detail.Scan)
	assert.Equal(t, "abc", detail.Scan.Revision)

	withoutScan, err := svc.RetrieveRepository(ctx, "repo-3")
	require.NoError(t, err)
	assert.Nil(t, withoutScan.Scan)

	_, err = svc.RetrieveRepository(ctx, "repo-2")
	require.Error(t, err, "an unreadable summary on the detail page is an error")
}

func TestBuildScanDetailMapsDependenciesAndFailedStage(t *testing.T) {
	finished := time.Date(2026, 10, 20, 12, 46, 30, 0, time.UTC)
	detail := services.BuildScanDetail(models.RunReport{
		Status: "incomplete", Error: "advise failed", FinishedAt: &finished,
		Stages: []models.RunStage{{Name: "analyze", Status: "completed"}, {Name: "advise", Status: "failed"}, {Name: "evaluate", Status: "failed"}},
		Dependencies: &models.RunDependencies{DirectCount: 3, TransitiveCount: 4, Projects: []models.RunProject{
			{PackageManager: "GoMod", DefinitionFile: "go.mod", DirectCount: 2, TransitiveCount: 4},
			{PackageManager: "NPM", DefinitionFile: "web/package.json", DirectCount: 1},
			{PackageManager: "GoMod", DefinitionFile: "tools/go.mod", DirectCount: 1},
		}},
	}, time.Now())

	assert.Equal(t, finished, detail.ScannedAt)
	assert.Equal(t, "advise", detail.FailedStage)
	assert.Equal(t, []string{"GoMod", "NPM"}, detail.PackageManagers)
	assert.Equal(t, 3, *detail.DirectDependencyCount)
	assert.Equal(t, 4, *detail.TransitiveDependencyCount)
	assert.Equal(t, models.ScanProject{PackageManager: "NPM", DefinitionFile: "web/package.json", DirectDependencyCount: 1}, detail.Projects[1])
}
