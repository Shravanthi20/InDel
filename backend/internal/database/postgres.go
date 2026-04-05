package database

import (
	"fmt"
	"log"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.DatabaseURL
	if dsn == "" {
		sslMode := cfg.DBSSLMode
		if sslMode == "" {
			sslMode = defaultSSLMode(cfg.DBHost)
		}

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, sslMode)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func defaultSSLMode(host string) string {
	h := strings.ToLower(strings.TrimSpace(host))

	// Local and docker-compose hosts generally run without TLS.
	switch h {
	case "", "localhost", "127.0.0.1", "::1", "postgres":
		return "disable"
	default:
		return "require"
	}
}

func Migrate(db *gorm.DB) error {
	// AutoMigrate is necessary for tests using in-memory SQLite
	// But it is causing 'insufficient arguments' crash on this Postgres version.
	// Since we use db-migrate, we can safely skip this in the demo environment.
	log.Println("⚠️ Skipping AutoMigrate to prevent crash. Ensure db-migrate has run.")
	return nil
}
