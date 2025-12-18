package database

import (
	"fmt"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect connects to the database, optionally resets the schema, and performs migrations.
func Connect(connStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connStr))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
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

	if err := migrateRepositoryLastCrawledAt(db); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.Repository{},
		&models.Organisation{},
		&models.GitOrganisatie{},
	); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

// migrateRepositoryLastCrawledAt renames the legacy updated_at column to last_crawled_at.
func migrateRepositoryLastCrawledAt(db *gorm.DB) error {
	m := db.Migrator()
	hasUpdated := m.HasColumn(&models.Repository{}, "updated_at")
	hasLastCrawled := m.HasColumn(&models.Repository{}, "last_crawled_at")

	if hasUpdated && !hasLastCrawled {
		if err := m.RenameColumn(&models.Repository{}, "updated_at", "last_crawled_at"); err != nil {
			return fmt.Errorf("failed to rename column updated_at to last_crawled_at: %w", err)
		}
	}

	return nil
}
