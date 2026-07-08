package database

import (
	"testing"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateRepositorySchemaColumnsAddsRepositoryStatusColumns(t *testing.T) {
	db := openLegacyRepositoryDB(t)

	require.NoError(t, migrateRepositorySchemaColumns(db))

	m := db.Migrator()
	require.True(t, m.HasColumn(&models.Repository{}, "is_fork"))
	require.True(t, m.HasColumn(&models.Repository{}, "fork_based_on_urls"))
	require.True(t, m.HasColumn(&models.Repository{}, "archived"))
}

func TestMigrateRepositorySchemaColumnsBackfillsForkFlag(t *testing.T) {
	db := openLegacyRepositoryDB(t)
	require.NoError(t, db.Exec("ALTER TABLE repositories ADD COLUMN is_fork boolean").Error)
	require.NoError(t, db.Exec("INSERT INTO repositories (id, name, is_fork) VALUES (?, ?, NULL)", "repo-1", "Repo 1").Error)

	require.NoError(t, migrateRepositorySchemaColumns(db))

	var isFork bool
	require.NoError(t, db.Raw("SELECT is_fork FROM repositories WHERE id = ?", "repo-1").Scan(&isFork).Error)
	require.False(t, isFork)
}

func TestMigrateRepositorySchemaColumnsBackfillsArchivedFlag(t *testing.T) {
	db := openLegacyRepositoryDB(t)
	require.NoError(t, db.Exec("ALTER TABLE repositories ADD COLUMN archived boolean").Error)
	require.NoError(t, db.Exec("INSERT INTO repositories (id, name, archived) VALUES (?, ?, NULL)", "repo-1", "Repo 1").Error)

	require.NoError(t, migrateRepositorySchemaColumns(db))

	var archived bool
	require.NoError(t, db.Raw("SELECT archived FROM repositories WHERE id = ?", "repo-1").Scan(&archived).Error)
	require.False(t, archived)
}

func TestMigrateRepositorySchemaColumnsSkipsMissingRepositoryTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, migrateRepositorySchemaColumns(db))
}

func openLegacyRepositoryDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE repositories (
			id text PRIMARY KEY,
			name text
		)
	`).Error)

	return db
}
