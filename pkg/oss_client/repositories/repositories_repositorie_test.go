package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Organisation{}, &models.Repository{}, &models.GitOrganisatie{}))
	return db
}

func TestRepositoriesRepository_SaveAndRetrieve(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	repository := &models.Repository{
		Id:               "repo-1",
		Name:             "Repo One",
		ShortDescription: "Repository description",
		LongDescription:  "Repository description",
		OrganisationID:   &org.Uri,
		Url:              "https://example.org/repos/repo-1",
		PublicCodeUrl:    "https://publiccode.net/repo-1",
		Active:           true,
	}
	require.NoError(t, repo.SaveRepository(ctx, repository))

	got, err := repo.GetRepositoryByID(ctx, "repo-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Repo One", got.Name)
	require.NotNil(t, got.Organisation)
	assert.Equal(t, "Org 1", got.Organisation.Label)
}

func TestRepositoriesRepository_SaveRepositoryUpdatesExistingByURL(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	createdAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
		Id:             "repo-original",
		Name:           "Original",
		OrganisationID: &org.Uri,
		Url:            "https://example.org/repo",
		CreatedAt:      createdAt,
		Active:         true,
	}))

	replacement := &models.Repository{
		Name: "Replacement",
		Url:  "https://example.org/repo",
	}
	require.NoError(t, repo.SaveRepository(ctx, replacement))

	assert.Equal(t, "repo-original", replacement.Id)
	assert.Equal(t, createdAt, replacement.CreatedAt)
	assert.Equal(t, &org.Uri, replacement.OrganisationID)

	got, err := repo.GetRepositoryByID(ctx, "repo-original")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Replacement", got.Name)
}

func TestRepositoriesRepository_GetRepositoriesOrganisationFilter(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org1 := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	org2 := &models.Organisation{Uri: "org-2", Label: "Org 2"}
	require.NoError(t, repo.SaveOrganisatie(org1))
	require.NoError(t, repo.SaveOrganisatie(org2))

	repositoriesToSave := []*models.Repository{
		{
			Id:             "repo-1",
			Name:           "Repo One",
			OrganisationID: &org1.Uri,
			PublicCodeUrl:  "https://publiccode.net/repo-1",
			PublicCode: &models.PublicCode{
				SoftwareType:      "library",
				DevelopmentStatus: "stable",
			},
			Active: true,
		},
		{
			Id:             "repo-2",
			Name:           "Repo Two",
			OrganisationID: &org1.Uri,
			PublicCodeUrl:  "https://publiccode.net/repo-2",
			PublicCode: &models.PublicCode{
				SoftwareType: "standalone/mobile",
				Legal:        &models.PublicCodeLegal{License: "EUPL-1.2"},
			},
			Active: true,
		},
		{Id: "repo-3", Name: "Repo Three", OrganisationID: &org2.Uri, Active: true},
		{Id: "repo-4", Name: "Repo Four", OrganisationID: &org1.Uri, Active: false},
		{Id: "repo-5", Name: "Repo Five", OrganisationID: &org1.Uri, Active: true},
	}
	for _, r := range repositoriesToSave {
		require.NoError(t, repo.SaveRepository(ctx, r))
	}
	require.NoError(t, db.Exec("UPDATE repositories SET active = NULL WHERE id = ?", "repo-2").Error)

	publicCodeOnly := true
	results, pagination, err := repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Organisation: &org1.Uri,
		PublicCode:   &publicCodeOnly,
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 2, pagination.TotalRecords)
	ids := make([]string, len(results))
	for i, repo := range results {
		ids[i] = repo.Id
	}
	assert.ElementsMatch(t, []string{"repo-1", "repo-2"}, ids)

	publicCodeDisabled := false
	results, pagination, err = repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Organisation: &org1.Uri,
		PublicCode:   &publicCodeDisabled,
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, pagination.TotalRecords)
	assert.Equal(t, "repo-5", results[0].Id)

	includeArchived := true
	results, pagination, err = repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Organisation: &org1.Uri,
		PublicCode:   &publicCodeDisabled,
		Archived:     &includeArchived,
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 2, pagination.TotalRecords)
	ids = make([]string, len(results))
	for i, repo := range results {
		ids[i] = repo.Id
	}
	assert.ElementsMatch(t, []string{"repo-4", "repo-5"}, ids)

	results, pagination, err = repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Organisation: &org1.Uri,
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 2, pagination.TotalRecords)
	ids = make([]string, len(results))
	for i, repo := range results {
		ids[i] = repo.Id
	}
	assert.ElementsMatch(t, []string{"repo-1", "repo-2"}, ids)

	results, pagination, err = repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Organisation: &org1.Uri,
		SoftwareType: []string{"library"},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, pagination.TotalRecords)
	assert.Equal(t, "repo-1", results[0].Id)

	results, pagination, err = repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Organisation: &org1.Uri,
		License:      []string{"EUPL-1.2"},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, pagination.TotalRecords)
	assert.Equal(t, "repo-2", results[0].Id)
}

func TestRepositoriesRepository_GetRepositoriesArchivedFilter(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
		Id:             "active-repo",
		Name:           "Active Repo",
		OrganisationID: &org.Uri,
		PublicCodeUrl:  "https://example.org/active/publiccode.yml",
		Active:         true,
	}))
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
		Id:             "archived-repo",
		Name:           "Archived Repo",
		OrganisationID: &org.Uri,
		PublicCodeUrl:  "https://example.org/archived/publiccode.yml",
		Active:         false,
	}))
	require.NoError(t, db.Exec("UPDATE repositories SET active = NULL WHERE id = ?", "active-repo").Error)

	results, pagination, err := repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, pagination.TotalRecords)
	assert.Equal(t, "active-repo", results[0].Id)

	includeArchived := true
	results, pagination, err = repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{Archived: &includeArchived})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 2, pagination.TotalRecords)
	ids := make([]string, len(results))
	for i, repo := range results {
		ids[i] = repo.Id
	}
	assert.ElementsMatch(t, []string{"active-repo", "archived-repo"}, ids)
}

func TestRepositoriesRepository_GetRepositoriesCombinesQueryAndFilters(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	publicCodeOnly := true
	repositoriesToSave := []*models.Repository{
		{
			Id:               "repo-1",
			Name:             "Open Forms",
			ShortDescription: "Digital form handling",
			OrganisationID:   &org.Uri,
			PublicCodeUrl:    "https://publiccode.net/repo-1",
			PublicCode:       &models.PublicCode{SoftwareType: "library"},
			Active:           true,
		},
		{
			Id:               "repo-2",
			Name:             "Open Forms Docs",
			ShortDescription: "Documentation for form handling",
			OrganisationID:   &org.Uri,
			PublicCode:       &models.PublicCode{SoftwareType: "library"},
			Active:           true,
		},
		{
			Id:               "repo-3",
			Name:             "Catalogus",
			ShortDescription: "Digital catalogue",
			OrganisationID:   &org.Uri,
			PublicCodeUrl:    "https://publiccode.net/repo-3",
			PublicCode:       &models.PublicCode{SoftwareType: "library"},
			Active:           true,
		},
	}
	for _, r := range repositoriesToSave {
		require.NoError(t, repo.SaveRepository(ctx, r))
	}

	results, pagination, err := repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Query:        "forms",
		PublicCode:   &publicCodeOnly,
		SoftwareType: []string{"library"},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-1", results[0].Id)
	assert.Equal(t, 1, pagination.TotalRecords)
}

func TestRepositoriesRepository_GetRepositoriesPaginatesFilteredResults(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	for i, name := range []string{"Alpha", "Beta", "Gamma"} {
		require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
			Id:            string(rune('a' + i)),
			Name:          name,
			PublicCodeUrl: "https://example.org/" + name + "/publiccode.yml",
			Active:        true,
		}))
	}

	results, pagination, err := repo.GetRepositorys(ctx, 2, 2, &models.RepositoryFiltersParams{})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Gamma", results[0].Name)
	assert.Equal(t, 3, pagination.TotalRecords)
	assert.Equal(t, 2, pagination.TotalPages)
	require.NotNil(t, pagination.Previous)
	assert.Equal(t, 1, *pagination.Previous)
	assert.Nil(t, pagination.Next)

	results, pagination, err = repo.GetRepositorys(ctx, 3, 2, &models.RepositoryFiltersParams{})
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, 3, pagination.TotalRecords)
}

func TestRepositoriesRepository_GetRepositoriesLastActivityAfterFilter(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	recent := time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)
	old := time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)
	repositoriesToSave := []*models.Repository{
		{Id: "repo-1", Name: "Repo One", OrganisationID: &org.Uri, PublicCodeUrl: "https://example.org/repo-1/publiccode.yml", LastActivityAt: recent, Active: true},
		{Id: "repo-2", Name: "Repo Two", OrganisationID: &org.Uri, PublicCodeUrl: "https://example.org/repo-2/publiccode.yml", LastActivityAt: old, Active: true},
	}
	for _, r := range repositoriesToSave {
		require.NoError(t, repo.SaveRepository(ctx, r))
	}

	date := "2024-01-01"
	results, pagination, err := repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Organisation:      &org.Uri,
		LastActivityAfter: &date,
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, pagination.TotalRecords)
	assert.Equal(t, "repo-1", results[0].Id)
}

func TestRepositoriesRepository_GetRepositoriesInvalidLastActivityAfter(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)

	date := "01-01-2024"
	_, _, err := repo.GetRepositorys(context.Background(), 1, 10, &models.RepositoryFiltersParams{LastActivityAfter: &date})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid lastActivityAfter format")
}

func TestRepositoriesRepository_SearchRepositoriesByText(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	save := func(id, name, shortDesc, longDesc string, active bool) {
		require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
			Id:               id,
			Name:             name,
			ShortDescription: shortDesc,
			LongDescription:  longDesc,
			Active:           active,
		}))
	}
	save("repo-1", "Account API", "", "", true)
	save("repo-2", "User Portal", "", "", true)
	save("repo-3", "Account API Legacy", "", "", false)
	save("repo-4", "Account API v2", "", "", true)
	save("repo-5", "Payments", "Account payment service", "Account billing integration", true)
	require.NoError(t, db.Exec("UPDATE repositories SET active = NULL WHERE id = ?", "repo-4").Error)

	results, pagination, err := repo.SearchRepositorys(ctx, 1, 10, nil, "account")
	require.NoError(t, err)
	require.Len(t, results, 3)
	ids := make([]string, len(results))
	for i, repo := range results {
		ids[i] = repo.Id
	}
	assert.ElementsMatch(t, []string{"repo-1", "repo-4", "repo-5"}, ids)
	assert.Equal(t, 3, pagination.TotalRecords)
}

func TestRepositoriesRepository_SearchRepositoriesBlankQueryReturnsEmptyPage(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)

	results, pagination, err := repo.SearchRepositorys(context.Background(), 0, 0, nil, "   ")
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, 1, pagination.CurrentPage)
	assert.Equal(t, 20, pagination.RecordsPerPage)
}

func TestRepositoriesRepository_SearchRepositoriesOrganisationFilterAndPagination(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org1 := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	org2 := &models.Organisation{Uri: "org-2", Label: "Org 2"}
	require.NoError(t, repo.SaveOrganisatie(org1))
	require.NoError(t, repo.SaveOrganisatie(org2))

	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{Id: "repo-1", Name: "Account Alpha", OrganisationID: &org1.Uri, Active: true}))
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{Id: "repo-2", Name: "Account Beta", OrganisationID: &org1.Uri, Active: true}))
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{Id: "repo-3", Name: "Account Other", OrganisationID: &org2.Uri, Active: true}))

	results, pagination, err := repo.SearchRepositorys(ctx, 1, 1, &org1.Uri, "account")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 2, pagination.TotalRecords)
	assert.Equal(t, 2, pagination.TotalPages)
	require.NotNil(t, pagination.Next)
	assert.Equal(t, 2, *pagination.Next)
	assert.Nil(t, pagination.Previous)

	results, pagination, err = repo.SearchRepositorys(ctx, 2, 1, &org1.Uri, "account")
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.NotNil(t, pagination.Previous)
	assert.Equal(t, 1, *pagination.Previous)
}

func TestRepositoriesRepository_SaveRepositoryPersistsForkFlag(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	fork := &models.Repository{
		Id:     "repo-fork",
		Name:   "Signalen frontend fork",
		Url:    "https://git.example.org/custom/frontend",
		IsFork: true,
		Active: true,
	}
	require.NoError(t, repo.SaveRepository(ctx, fork))

	all, err := repo.AllRepositorys(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, "https://git.example.org/custom/frontend", all[0].Url)
	assert.True(t, all[0].IsFork)
}

func TestRepositoriesRepository_SaveRepositoryPersistsForkBasedOnURLs(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	fork := &models.Repository{
		Id:              "repo-based-on",
		Name:            "Variant fork",
		Url:             "https://github.com/example/variant",
		ForkBasedOnURLs: []string{"https://github.com/example/upstream"},
		Active:          true,
	}
	require.NoError(t, repo.SaveRepository(ctx, fork))

	all, err := repo.AllRepositorys(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, []string{"https://github.com/example/upstream"}, all[0].ForkBasedOnURLs)
}

func TestRepositoriesRepository_FindOrganisationByURI(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	found, err := repo.FindOrganisationByURI(context.Background(), "org-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Org 1", found.Label)
}

func TestRepositoriesRepository_FindOrganisationByURIMissingReturnsNil(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)

	found, err := repo.FindOrganisationByURI(context.Background(), "missing")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestRepositoriesRepository_GetOrganisationsOrdersAndPaginates(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)

	require.NoError(t, repo.SaveOrganisatie(&models.Organisation{Uri: "org-b", Label: "Beta"}))
	require.NoError(t, repo.SaveOrganisatie(&models.Organisation{Uri: "org-a", Label: "Alpha"}))
	require.NoError(t, repo.SaveOrganisatie(&models.Organisation{Uri: "org-c", Label: "Gamma"}))

	results, pagination, err := repo.GetOrganisations(context.Background(), 1, 2)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, []string{"Alpha", "Beta"}, []string{results[0].Label, results[1].Label})
	assert.Equal(t, 3, pagination.TotalRecords)
	assert.Equal(t, 2, pagination.TotalPages)
	require.NotNil(t, pagination.Next)
	assert.Equal(t, 2, *pagination.Next)
}

func TestRepositoriesRepository_GitOrganisationsCRUD(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org1 := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	org2 := &models.Organisation{Uri: "org-2", Label: "Org 2"}
	require.NoError(t, repo.SaveOrganisatie(org1))
	require.NoError(t, repo.SaveOrganisatie(org2))
	require.NoError(t, repo.SaveGitOrganisatie(ctx, &models.GitOrganisatie{Id: "git-1", OrganisationID: &org1.Uri, Url: "https://github.com/org-1"}))
	require.NoError(t, repo.SaveGitOrganisatie(ctx, &models.GitOrganisatie{Id: "git-2", OrganisationID: &org2.Uri, Url: "https://github.com/org-2"}))

	found, err := repo.FindGitOrganisationByURL(ctx, "https://github.com/org-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "git-1", found.Id)
	require.NotNil(t, found.Organisation)
	assert.Equal(t, "Org 1", found.Organisation.Label)

	missing, err := repo.FindGitOrganisationByURL(ctx, "https://github.com/missing")
	require.NoError(t, err)
	assert.Nil(t, missing)

	results, pagination, err := repo.GetGitOrganisations(ctx, 1, 10, &org2.Uri)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "git-2", results[0].Id)
	assert.Equal(t, 1, pagination.TotalRecords)
}

func TestRepositoriesRepository_GetRepositoryFilterCountsAppliesCrossFilters(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org1 := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	org2 := &models.Organisation{Uri: "org-2", Label: "Org 2"}
	require.NoError(t, repo.SaveOrganisatie(org1))
	require.NoError(t, repo.SaveOrganisatie(org2))

	recent := time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)
	old := time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
		Id:             "repo-1",
		Name:           "Repo One",
		OrganisationID: &org1.Uri,
		PublicCodeUrl:  "https://example.org/repo-1/publiccode.yml",
		LastActivityAt: recent,
		Active:         true,
		PublicCode: &models.PublicCode{
			SoftwareType:      "library",
			DevelopmentStatus: "stable",
			Platforms:         []string{"web", "linux"},
			Legal:             &models.PublicCodeLegal{License: "EUPL-1.2"},
			Maintenance:       &models.PublicCodeMaintenance{Type: "internal"},
			Localisation:      &models.PublicCodeLocalisation{AvailableLanguages: []string{"nl", "en"}},
		},
	}))
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
		Id:             "repo-2",
		Name:           "Repo Two",
		OrganisationID: &org1.Uri,
		PublicCodeUrl:  "https://example.org/repo-2/publiccode.yml",
		LastActivityAt: old,
		Active:         true,
		PublicCode: &models.PublicCode{
			SoftwareType:      "standalone/web",
			DevelopmentStatus: "beta",
			Platforms:         []string{"web"},
			Legal:             &models.PublicCodeLegal{License: "MIT"},
			Maintenance:       &models.PublicCodeMaintenance{Type: "community"},
			Localisation:      &models.PublicCodeLocalisation{AvailableLanguages: []string{"nl"}},
		},
	}))
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
		Id:             "repo-3",
		Name:           "Repo Three",
		OrganisationID: &org2.Uri,
		LastActivityAt: recent,
		Active:         true,
	}))
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
		Id:             "repo-4",
		Name:           "Inactive",
		OrganisationID: &org1.Uri,
		PublicCodeUrl:  "https://example.org/inactive/publiccode.yml",
		LastActivityAt: recent,
		Active:         false,
	}))

	date := "2024-01-01"
	counts, err := repo.GetRepositoryFilterCounts(ctx, &models.RepositoryFiltersParams{
		Organisation:      &org1.Uri,
		LastActivityAfter: &date,
	})
	require.NoError(t, err)

	assert.Equal(t, 1, counts.PublicCode)
	require.NotNil(t, counts.LastActivityAfter)
	assert.Equal(t, 1, *counts.LastActivityAfter)
	assert.Equal(t, []models.FilterCount{{Value: "library", Count: 1}}, counts.SoftwareType)
	assert.Equal(t, []models.FilterCount{{Value: "stable", Count: 1}}, counts.DevelopmentStatus)
	assert.Equal(t, []models.FilterCount{{Value: "internal", Count: 1}}, counts.MaintenanceType)
	assert.Equal(t, []models.FilterCount{{Value: "EUPL-1.2", Count: 1}}, counts.License)
	assert.Equal(t, 1, counts.Archived)

	platformCounts := map[string]int{}
	for _, fc := range counts.Platforms {
		platformCounts[fc.Value] = fc.Count
	}
	assert.Equal(t, map[string]int{"web": 1, "linux": 1}, platformCounts)

	languageCounts := map[string]int{}
	for _, fc := range counts.AvailableLanguages {
		languageCounts[fc.Value] = fc.Count
	}
	assert.Equal(t, map[string]int{"nl": 1, "en": 1}, languageCounts)
	orgCounts := map[string]int{}
	for _, fc := range counts.Organisation {
		orgCounts[fc.Value] = fc.Count
	}
	assert.Equal(t, map[string]int{"org-1": 1}, orgCounts)
}
