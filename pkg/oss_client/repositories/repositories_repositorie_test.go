package repositories_test

import (
	"context"
	"testing"

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
		{Id: "repo-1", Name: "Repo One", OrganisationID: &org1.Uri, Active: true},
		{Id: "repo-2", Name: "Repo Two", OrganisationID: &org1.Uri, Active: true},
		{Id: "repo-3", Name: "Repo Three", OrganisationID: &org2.Uri, Active: true},
		{Id: "repo-4", Name: "Repo Four", OrganisationID: &org1.Uri, Active: false},
	}
	for _, r := range repositoriesToSave {
		require.NoError(t, repo.SaveRepository(ctx, r))
	}
	require.NoError(t, db.Exec("UPDATE repositories SET active = NULL WHERE id = ?", "repo-2").Error)

	results, pagination, err := repo.GetRepositorys(ctx, 1, 10, &org1.Uri)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 2, pagination.TotalRecords)
	ids := make([]string, len(results))
	for i, repo := range results {
		ids[i] = repo.Id
	}
	assert.ElementsMatch(t, []string{"repo-1", "repo-2"}, ids)
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
