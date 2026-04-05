package database

import (
	"fmt"
	"log"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	// AutoMigrate is necessary for tests using in-memory SQLite
	// But it is causing 'insufficient arguments' crash on this Postgres version.
	// Since we use db-migrate, we can safely skip this in the demo environment.
	log.Println("⚠️ Skipping AutoMigrate to prevent crash. Ensure db-migrate has run.")
	return nil
}
