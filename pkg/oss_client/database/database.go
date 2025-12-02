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
		m := db.Migrator()
		for _, table := range []interface{}{
			&models.CodeHosting{},
			&models.Repository{},
			&models.GitOrganisatie{},
			&models.Organisation{},
		} {
			if m.HasTable(table) {
				if err := m.DropTable(table); err != nil {
					return nil, fmt.Errorf("failed to reset table %T in schema %s: %w", table, schema, err)
				}
			}
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
