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
	require.NoError(t, db.AutoMigrate(&models.Organisation{}, &models.Repositorie{}))
	return db
}

func TestRepositoriesRepository_SaveAndRetrieve(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	repositorie := &models.Repositorie{
		Id:             "repo-1",
		Name:           "Repo One",
		Description:    "Repository description",
		OrganisationID: &org.Uri,
		RepositorieUri: "https://example.org/repos/repo-1",
		PublicCodeUrl:  "https://publiccode.net/repo-1",
	}
	require.NoError(t, repo.SaveRepositorie(ctx, repositorie))

	got, err := repo.GetRepositorieByID(ctx, "repo-1")
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

	repositoriesToSave := []*models.Repositorie{
		{Id: "repo-1", Name: "Repo One", OrganisationID: &org1.Uri},
		{Id: "repo-2", Name: "Repo Two", OrganisationID: &org1.Uri},
		{Id: "repo-3", Name: "Repo Three", OrganisationID: &org2.Uri},
	}
	for _, r := range repositoriesToSave {
		require.NoError(t, repo.SaveRepositorie(ctx, r))
	}

	results, pagination, err := repo.GetRepositories(ctx, 1, 10, &org1.Uri, nil)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 2, pagination.TotalRecords)

	ids := "repo-3"
	results, pagination, err = repo.GetRepositories(ctx, 1, 10, nil, &ids)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "repo-3", results[0].Id)
	assert.Equal(t, 1, pagination.TotalRecords)
}

func TestRepositoriesRepository_SearchRepositories(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewRepositoriesRepository(db)
	ctx := context.Background()

	org := &models.Organisation{Uri: "org-1", Label: "Org 1"}
	require.NoError(t, repo.SaveOrganisatie(org))

	save := func(id, name string) {
		require.NoError(t, repo.SaveRepositorie(ctx, &models.Repositorie{
			Id:             id,
			Name:           name,
			OrganisationID: &org.Uri,
		}))
	}
	save("repo-1", "Account API")
	save("repo-2", "User Portal")

	results, pagination, err := repo.SearchRepositories(ctx, 1, 10, nil, "account")
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
