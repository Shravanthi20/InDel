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
	// We avoid GORM AutoMigrate due to prior compatibility issues on this Postgres setup.
	// Keep these idempotent guards so production can recover from partial/legacy schemas.
	stmts := []string{
		"ALTER TABLE disruptions ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP",
		"ALTER TABLE payouts ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMP",
		"ALTER TABLE payouts ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP",
		"CREATE INDEX IF NOT EXISTS idx_disruptions_status_processed_at ON disruptions(status, processed_at)",
		"CREATE INDEX IF NOT EXISTS idx_payouts_status_next_retry_at ON payouts(status, next_retry_at)",
	}

	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	log.Println("✅ Schema compatibility checks applied")
	return nil
}
