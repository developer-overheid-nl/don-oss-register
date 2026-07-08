package database

import (
	"fmt"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	commondatabase "github.com/developer-overheid-nl/don-register-common/database"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

// Connect connects to the database, optionally resets the schema, and performs migrations.
func Connect(connStr string) (*gorm.DB, error) {
	db, err := commondatabase.ConnectPostgres(connStr)
	if err != nil {
		return nil, err
	}

	// if resetSchema {
	// 	m := db.Migrator()
	// 	for _, table := range []interface{}{
	// 		&models.Repository{},
	// 		&models.GitOrganisatie{},
	// 		&models.Organisation{},
	// 	} {
	// 		if m.HasTable(table) {
	// 			if err := m.DropTable(table); err != nil {
	// 				return nil, fmt.Errorf("failed to reset table %T in schema %s: %w", table, schema, err)
	// 			}
	// 		}
	// 	}
	// }

	if err := migrateRepositoryTimestampColumns(db); err != nil {
		return nil, err
	}
	if err := migrateRepositorySchemaColumns(db); err != nil {
		return nil, err
	}

	// if err := db.AutoMigrate(
	// 	&models.Repository{},
	// 	&models.Organisation{},
	// 	&models.GitOrganisatie{},
	// ); err != nil {
	// 	return nil, fmt.Errorf("migration failed: %w", err)
	// }

	return db, nil
}

// migrateRepositorySchemaColumns adds repository columns introduced after the
// initial production schema was created.
func migrateRepositorySchemaColumns(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable(&models.Repository{}) {
		return nil
	}

	if !m.HasColumn(&models.Repository{}, "is_fork") {
		if err := m.AddColumn(&models.Repository{}, "IsFork"); err != nil {
			return fmt.Errorf("failed to add column is_fork: %w", err)
		}
	}
	if err := db.Model(&models.Repository{}).
		Where("is_fork IS NULL").
		Update("is_fork", false).Error; err != nil {
		return fmt.Errorf("failed to backfill column is_fork: %w", err)
	}

	if !m.HasColumn(&models.Repository{}, "fork_based_on_urls") {
		if err := m.AddColumn(&models.Repository{}, "ForkBasedOnURLs"); err != nil {
			return fmt.Errorf("failed to add column fork_based_on_urls: %w", err)
		}
	}
	if !m.HasColumn(&models.Repository{}, "archived") {
		if err := m.AddColumn(&models.Repository{}, "Archived"); err != nil {
			return fmt.Errorf("failed to add column archived: %w", err)
		}
	}
	if err := db.Model(&models.Repository{}).
		Where("archived IS NULL").
		Update("archived", false).Error; err != nil {
		return fmt.Errorf("failed to backfill column archived: %w", err)
	}

	return nil
}

// migrateRepositoryTimestampColumns renames legacy timestamp columns.
func migrateRepositoryTimestampColumns(db *gorm.DB) error {
	m := db.Migrator()
	hasUpdated := m.HasColumn(&models.Repository{}, "updated_at")
	hasLastCrawled := m.HasColumn(&models.Repository{}, "last_crawled_at")
	hasLastActivity := m.HasColumn(&models.Repository{}, "last_activity")
	hasLastActivityAt := m.HasColumn(&models.Repository{}, "last_activity_at")

	if hasUpdated && !hasLastCrawled {
		if err := m.RenameColumn(&models.Repository{}, "updated_at", "last_crawled_at"); err != nil {
			return fmt.Errorf("failed to rename column updated_at to last_crawled_at: %w", err)
		}
	}

	if hasLastActivity && !hasLastActivityAt {
		if err := m.RenameColumn(&models.Repository{}, "last_activity", "last_activity_at"); err != nil {
			return fmt.Errorf("failed to rename column last_activity to last_activity_at: %w", err)
		}
	}

	return nil
}
