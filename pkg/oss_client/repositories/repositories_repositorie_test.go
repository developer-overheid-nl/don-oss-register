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
	require.Len(t, results, 3)
	assert.Equal(t, 3, pagination.TotalRecords)

	results, pagination, err = repo.GetRepositorys(ctx, 1, 10, &models.RepositoryFiltersParams{
		Organisation: &org1.Uri,
	})
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, 3, pagination.TotalRecords)
	ids = make([]string, len(results))
	for i, repo := range results {
		ids[i] = repo.Id
	}
	assert.ElementsMatch(t, []string{"repo-1", "repo-2", "repo-5"}, ids)

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

func TestRepositoriesRepository_GetRepositoriesLastActivityAfterFilter(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	recent := time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)
	old := time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)
	repositoriesToSave := []*models.Repository{
		{Id: "repo-1", Name: "Repo One", OrganisationID: &org.Uri, LastActivityAt: recent, Active: true},
		{Id: "repo-2", Name: "Repo Two", OrganisationID: &org.Uri, LastActivityAt: old, Active: true},
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

func TestRepositoriesRepository_SearchRepositories(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	save := func(id, name, shortDesc, longDesc string, active bool) {
		require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
			Id:               id,
			Name:             name,
			ShortDescription: shortDesc,
			LongDescription:  longDesc,
			OrganisationID:   &org.Uri,
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

func TestRepositoriesRepository_SaveRepositoryPersistsForkFlag(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	fork := &models.Repository{
		Id:     "repo-fork",
		Name:   "Signalen frontend fork",
		Url:    "https://github.com/Signalen/frontend",
		IsFork: true,
		Active: true,
	}
	require.NoError(t, repo.SaveRepository(ctx, fork))

	all, err := repo.AllRepositorys(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, "https://github.com/Signalen/frontend", all[0].Url)
	assert.True(t, all[0].IsFork)
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
