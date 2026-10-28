package services_test

import (
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/services"
	"github.com/stretchr/testify/assert"
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
