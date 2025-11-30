package database

import (
	"fmt"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect connects to the database, optionally resets the schema, and performs migrations.
func Connect(connStr string, schema string, resetSchema bool) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connStr))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if resetSchema {
		resetSQL := fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE; CREATE SCHEMA %q;", schema, schema)
		if err := db.Exec(resetSQL).Error; err != nil {
			return nil, fmt.Errorf("failed to reset schema %s: %w", schema, err)
		}
	}

	if err := db.AutoMigrate(
		&models.Repository{},
		&models.Organisation{},
		&models.GitOrganisatie{},
		&models.CodeHosting{},
	); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}
