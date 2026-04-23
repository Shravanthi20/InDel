package database

import (
	"fmt"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dsn string
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	}

	// PgBouncer/pooler connections (common on hosted Postgres) can fail with
	// statement-cache errors unless simple protocol is used.
	preferSimpleProtocol := cfg.DBPreferSimpleProtocol || shouldForceSimpleProtocol(cfg)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: preferSimpleProtocol,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func shouldForceSimpleProtocol(cfg *config.Config) bool {
	if strings.EqualFold(strings.TrimSpace(cfg.InDelEnv), "production") {
		return true
	}
	databaseURL := strings.ToLower(strings.TrimSpace(cfg.DatabaseURL))
	return strings.Contains(databaseURL, "pooler.") || strings.Contains(databaseURL, "pgbouncer")
}

func Migrate(db *gorm.DB) error {
	// In Docker/runtime, schema is managed by SQL migrations (db-migrate service).
	// Keep AutoMigrate for SQLite-based tests only.
	if db != nil && db.Dialector != nil && db.Dialector.Name() == "postgres" {
		return nil
	}

	return db.AutoMigrate(
		&models.User{},
		&models.WorkerProfile{},
		&models.Notification{},
		&models.Zone{},
		&models.Policy{},
		&models.ActivePolicy{},
		&models.Claim{},
		&models.EarningsRecord{},
		&models.Order{},
		&models.WeeklyPolicyCycle{},
		&models.PremiumPayment{},
		&models.EarningsBaseline{},
		&models.WeeklyEarningsSummary{},
		&models.Disruption{},
		&models.Batch{},
		&models.BatchOrder{},
		&models.Payout{},
		&models.PayoutAttempt{},
		&models.KafkaEventLog{},
		&models.SyntheticGenerationRun{},
		&models.ClaimFraudScore{},
		&models.ClaimAuditLog{},
	)
}
