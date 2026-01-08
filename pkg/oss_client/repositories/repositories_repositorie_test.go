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

func TestRepositoriesRepository_GetRepositoriesFilters(t *testing.T) {
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

	results, pagination, err := repo.GetRepositorys(ctx, 1, 10, &org1.Uri)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 2, pagination.TotalRecords)
	for _, repo := range results {
		assert.NotEqual(t, "repo-4", repo.Id)
	}
}

func TestRepositoriesRepository_SearchRepositories(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	save := func(id, name string) {
		require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
			Id:             id,
			Name:           name,
			OrganisationID: &org.Uri,
			Active:         true,
		}))
	}
	save("repo-1", "Account API")
	save("repo-2", "User Portal")
	require.NoError(t, repo.SaveRepository(ctx, &models.Repository{
		Id:             "repo-3",
		Name:           "Account API Legacy",
		OrganisationID: &org.Uri,
		Active:         false,
	}))

	results, pagination, err := repo.SearchRepositorys(ctx, 1, 10, nil, "account")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-1", results[0].Id)
	assert.Equal(t, 1, pagination.TotalRecords)
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
