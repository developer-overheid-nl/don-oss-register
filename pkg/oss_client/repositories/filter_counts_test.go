package repositories

import (
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/assert"
)

var orgID = "https://example.org"

func ptr(s string) *string { return &s }
func boolPtr(b bool) *bool { return &b }

func makeRepo(opts ...func(*models.Repository)) models.Repository {
	r := models.Repository{
		Id:             "repo-1",
		Active:         true,
		OrganisationID: &orgID,
		LastActivityAt: time.Now(),
	}
	for _, o := range opts {
		o(&r)
	}
	return r
}

func withPublicCodeUrl(url string) func(*models.Repository) {
	return func(r *models.Repository) { r.PublicCodeUrl = url }
}

func withPublicCode(pc *models.PublicCode) func(*models.Repository) {
	return func(r *models.Repository) { r.PublicCode = pc }
}

func withLastActivity(t time.Time) func(*models.Repository) {
	return func(r *models.Repository) { r.LastActivityAt = t }
}

func withOrg(uri string) func(*models.Repository) {
	return func(r *models.Repository) { r.OrganisationID = &uri }
}

// repoMatchesFilters

func TestRepoMatchesFilters_NoFilters(t *testing.T) {
	repo := makeRepo()
	assert.True(t, repoMatchesFilters(repo, &models.RepositoryFiltersParams{}, ""))
}

func TestRepoMatchesFilters_PublicCode_Match(t *testing.T) {
	repo := makeRepo(withPublicCodeUrl("https://example.org/publiccode.yml"))
	p := &models.RepositoryFiltersParams{PublicCode: boolPtr(true)}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_PublicCode_NoMatch(t *testing.T) {
	repo := makeRepo()
	p := &models.RepositoryFiltersParams{PublicCode: boolPtr(true)}
	assert.False(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_PublicCode_Excluded(t *testing.T) {
	repo := makeRepo()
	p := &models.RepositoryFiltersParams{PublicCode: boolPtr(true)}
	assert.True(t, repoMatchesFilters(repo, p, "publiccode"))
}

func TestRepoMatchesFilters_Organisation_Match(t *testing.T) {
	repo := makeRepo(withOrg("https://example.org"))
	p := &models.RepositoryFiltersParams{Organisation: ptr("https://example.org")}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_Organisation_NoMatch(t *testing.T) {
	repo := makeRepo(withOrg("https://other.org"))
	p := &models.RepositoryFiltersParams{Organisation: ptr("https://example.org")}
	assert.False(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_LastActivityAfter_Match(t *testing.T) {
	repo := makeRepo(withLastActivity(time.Now()))
	p := &models.RepositoryFiltersParams{LastActivityAfter: ptr("2020-01-01")}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_LastActivityAfter_NoMatch(t *testing.T) {
	repo := makeRepo(withLastActivity(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)))
	p := &models.RepositoryFiltersParams{LastActivityAfter: ptr("2020-01-01")}
	assert.False(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_SoftwareType_Match(t *testing.T) {
	repo := makeRepo(withPublicCode(&models.PublicCode{SoftwareType: "library"}))
	p := &models.RepositoryFiltersParams{SoftwareType: []string{"library"}}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_SoftwareType_NoMatch(t *testing.T) {
	repo := makeRepo(withPublicCode(&models.PublicCode{SoftwareType: "addon"}))
	p := &models.RepositoryFiltersParams{SoftwareType: []string{"library"}}
	assert.False(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_SoftwareType_NilPublicCode(t *testing.T) {
	repo := makeRepo()
	p := &models.RepositoryFiltersParams{SoftwareType: []string{"library"}}
	assert.False(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_Platforms_AllMatch(t *testing.T) {
	repo := makeRepo(withPublicCode(&models.PublicCode{Platforms: []string{"web", "linux"}}))
	p := &models.RepositoryFiltersParams{Platforms: []string{"web", "linux"}}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_Platforms_PartialMatch(t *testing.T) {
	repo := makeRepo(withPublicCode(&models.PublicCode{Platforms: []string{"web"}}))
	p := &models.RepositoryFiltersParams{Platforms: []string{"web", "linux"}}
	assert.False(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_AvailableLanguages_Match(t *testing.T) {
	repo := makeRepo(withPublicCode(&models.PublicCode{
		Localisation: &models.PublicCodeLocalisation{AvailableLanguages: []string{"nl", "en"}},
	}))
	p := &models.RepositoryFiltersParams{AvailableLanguages: []string{"nl"}}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_License_Match(t *testing.T) {
	repo := makeRepo(withPublicCode(&models.PublicCode{
		Legal: &models.PublicCodeLegal{License: "MIT"},
	}))
	p := &models.RepositoryFiltersParams{License: []string{"MIT"}}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_MaintenanceType_Match(t *testing.T) {
	repo := makeRepo(withPublicCode(&models.PublicCode{
		Maintenance: &models.PublicCodeMaintenance{Type: "internal"},
	}))
	p := &models.RepositoryFiltersParams{MaintenanceType: []string{"internal"}}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

func TestRepoMatchesFilters_MultipleFilters_AllApply(t *testing.T) {
	repo := makeRepo(
		withPublicCodeUrl("https://example.org/publiccode.yml"),
		withPublicCode(&models.PublicCode{
			SoftwareType: "library",
			Platforms:    []string{"web"},
		}),
	)
	p := &models.RepositoryFiltersParams{
		PublicCode:   boolPtr(true),
		SoftwareType: []string{"library"},
		Platforms:    []string{"web"},
	}
	assert.True(t, repoMatchesFilters(repo, p, ""))
}

// countByField

func TestCountByField_CountsCorrectly(t *testing.T) {
	repos := []models.Repository{
		makeRepo(withPublicCode(&models.PublicCode{SoftwareType: "library"})),
		makeRepo(withPublicCode(&models.PublicCode{SoftwareType: "library"})),
		makeRepo(withPublicCode(&models.PublicCode{SoftwareType: "addon"})),
		makeRepo(),
	}
	p := &models.RepositoryFiltersParams{}
	result := countByField(repos, p, "", func(r models.Repository) string {
		if r.PublicCode == nil {
			return ""
		}
		return r.PublicCode.SoftwareType
	})

	assert.Len(t, result, 2)
	assert.Equal(t, "library", result[0].Value)
	assert.Equal(t, 2, result[0].Count)
	assert.Equal(t, "addon", result[1].Value)
	assert.Equal(t, 1, result[1].Count)
}

func TestCountByField_RespectsOtherFilters(t *testing.T) {
	org1 := "https://org1.nl"
	org2 := "https://org2.nl"
	repos := []models.Repository{
		{OrganisationID: &org1, Active: true, PublicCode: &models.PublicCode{SoftwareType: "library"}},
		{OrganisationID: &org2, Active: true, PublicCode: &models.PublicCode{SoftwareType: "addon"}},
	}
	p := &models.RepositoryFiltersParams{Organisation: &org1}
	result := countByField(repos, p, "softwareType", func(r models.Repository) string {
		if r.PublicCode == nil {
			return ""
		}
		return r.PublicCode.SoftwareType
	})

	assert.Len(t, result, 1)
	assert.Equal(t, "library", result[0].Value)
}

// countByArrayField

func TestCountByArrayField_CountsEachValue(t *testing.T) {
	repos := []models.Repository{
		makeRepo(withPublicCode(&models.PublicCode{Platforms: []string{"web", "linux"}})),
		makeRepo(withPublicCode(&models.PublicCode{Platforms: []string{"web"}})),
		makeRepo(withPublicCode(&models.PublicCode{Platforms: []string{"linux"}})),
	}
	p := &models.RepositoryFiltersParams{}
	result := countByArrayField(repos, p, "", func(r models.Repository) []string {
		if r.PublicCode == nil {
			return nil
		}
		return r.PublicCode.Platforms
	})

	counts := make(map[string]int)
	for _, fc := range result {
		counts[fc.Value] = fc.Count
	}
	assert.Equal(t, 2, counts["web"])
	assert.Equal(t, 2, counts["linux"])
}

// countRepos

func TestCountRepos_Toggle(t *testing.T) {
	repos := []models.Repository{
		makeRepo(withPublicCodeUrl("https://example.org/publiccode.yml")),
		makeRepo(),
		makeRepo(withPublicCodeUrl("https://other.org/publiccode.yml")),
	}
	p := &models.RepositoryFiltersParams{}
	count := countRepos(repos, p, "publiccode", func(r models.Repository) bool {
		return r.PublicCodeUrl != ""
	})
	assert.Equal(t, 2, count)
}
