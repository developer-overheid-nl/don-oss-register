package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScansRepository_LatestScanPerRepository(t *testing.T) {
	db := setupDB(t)
	require.NoError(t, db.AutoMigrate(&models.RepositoryScan{}))
	repo := repositories.NewScansRepository(db)
	ctx := context.Background()
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	key := "ort:1"

	for _, scan := range []models.RepositoryScan{
		{ID: "a-old", RepositoryID: "a", Status: "completed", ScannedAt: base, Summary: "{}", IdempotencyKey: &key},
		{ID: "a-new", RepositoryID: "a", Status: "failed", ScannedAt: base.Add(time.Hour), Summary: "{}"},
		{ID: "b-only", RepositoryID: "b", Status: "completed", ScannedAt: base, Summary: "{}"},
	} {
		require.NoError(t, repo.SaveScan(ctx, &scan))
	}
	duplicate := models.RepositoryScan{ID: "dup", RepositoryID: "a", Status: "completed", IdempotencyKey: &key, Summary: "{}"}
	require.Error(t, repo.SaveScan(ctx, &duplicate), "idempotency key must be unique")

	latest, err := repo.LatestScans(ctx, []string{"a", "b", "c"})
	require.NoError(t, err)
	require.Len(t, latest, 2)
	assert.Equal(t, "a-new", latest["a"].ID)
	assert.Equal(t, "b-only", latest["b"].ID)

	empty, err := repo.LatestScans(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, empty)

	found, err := repo.FindScanByIdempotencyKey(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "a-old", found.ID)

	missing, err := repo.FindScanByIdempotencyKey(ctx, "ort:missing")
	require.NoError(t, err)
	assert.Nil(t, missing)
}
