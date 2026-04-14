package database

import (
	"fmt"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/models"
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
