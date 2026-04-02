package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	seed := flag.Int("seed", 42, "Fixed seed for reproducible metrics")
	scenario := flag.String("scenario", "normal_week", "Scenario preset: normal_week, mild_disruption, severe_disruption, fraud_burst")
	outputDir := flag.String("output-dir", filepath.Join("generated", "synthetic-cli"), "Directory for CSV and SQL artifacts")
	dbPath := flag.String("db", "synthdata.db", "SQLite output database path")
	flag.Parse()

	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open sqlite database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to migrate sqlite database: %v", err)
	}

	service := services.NewCoreOpsService(db)
	result, err := service.GenerateSyntheticData(services.SyntheticGenerateRequest{
		Seed:      *seed,
		Scenario:  *scenario,
		OutputDir: *outputDir,
	}, time.Now().UTC())
	if err != nil {
		log.Fatalf("synthetic generation failed: %v", err)
	}

	log.Printf("Synthetic generation complete: run_id=%s workers=%d claims=%d payouts=%d", result.RunID, result.Counts["workers"], result.Counts["claims"], result.Counts["payouts"])
	log.Printf("Artifacts written to %s", *outputDir)
}
