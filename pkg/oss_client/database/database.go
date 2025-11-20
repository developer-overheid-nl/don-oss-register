package database

import (
	"fmt"

	"github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/models"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(connStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connStr))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Repositorie{},
		&models.Organisation{},
	); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}
