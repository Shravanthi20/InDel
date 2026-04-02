package services

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCoreOpsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:part4_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}
	return db
}

func seedCycleFixtures(t *testing.T, db *gorm.DB, includeBroken bool) {
	t.Helper()
	zone := models.Zone{Name: "Tambaram", City: "Chennai", State: "Tamil Nadu", RiskRating: 0.62}
	if err := db.Create(&zone).Error; err != nil {
		t.Fatalf("seed zone: %v", err)
	}

	for i := 1; i <= 2; i++ {
		user := models.User{ID: uint(i), Phone: fmt.Sprintf("+91990000000%d", i), Role: "worker"}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
		if err := db.Create(&models.WorkerProfile{WorkerID: uint(i), Name: "Worker", ZoneID: zone.ID, VehicleType: "two_wheeler", UPIId: "w@upi", AQIZone: "medium", TotalEarningsLifetime: 100000}).Error; err != nil {
			t.Fatalf("seed profile: %v", err)
		}
		if err := db.Create(&models.Policy{WorkerID: uint(i), Status: "active", PremiumAmount: 22}).Error; err != nil {
			t.Fatalf("seed policy: %v", err)
		}
		if !includeBroken || i == 1 {
			if err := db.Create(&models.EarningsBaseline{WorkerID: uint(i), BaselineAmount: 4200, LastUpdatedAt: time.Now().UTC()}).Error; err != nil {
				t.Fatalf("seed baseline: %v", err)
			}
		}
	}
}

func TestRunWeeklyCycleIdempotencyAndRecovery(t *testing.T) {
	db := setupCoreOpsTestDB(t)
	seedCycleFixtures(t, db, true)
	service := NewCoreOpsService(db)
	now := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)

	result1, err := service.RunWeeklyCycle(now)
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if result1.Status != "partial_failure" {
		t.Fatalf("expected partial_failure, got %s", result1.Status)
	}
	if result1.PremiumsComputed != 1 || result1.PremiumFailures != 1 {
		t.Fatalf("unexpected first run counts: %+v", result1)
	}

	result2, err := service.RunWeeklyCycle(now)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if result2.PremiumsComputed != 1 || result2.PremiumFailures != 1 {
		t.Fatalf("expected idempotent second run before fix, got %+v", result2)
	}

	if err := db.Create(&models.EarningsBaseline{WorkerID: 2, BaselineAmount: 3800, LastUpdatedAt: now}).Error; err != nil {
		t.Fatalf("insert missing baseline: %v", err)
	}

	result3, err := service.RunWeeklyCycle(now)
	if err != nil {
		t.Fatalf("recovery run failed: %v", err)
	}
	if result3.Status != "completed" {
		t.Fatalf("expected completed after recovery, got %s", result3.Status)
	}
	if result3.PremiumsComputed != 2 || result3.PremiumFailures != 0 {
		t.Fatalf("unexpected recovery counts: %+v", result3)
	}
}

func TestGenerateClaimsAndQueueProcessPayouts(t *testing.T) {
	db := setupCoreOpsTestDB(t)
	service := NewCoreOpsService(db)
	now := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	weekStart, weekEnd := weekBounds(now)

	zone := models.Zone{Name: "Rohini", City: "Delhi", State: "Delhi", RiskRating: 0.7}
	if err := db.Create(&zone).Error; err != nil { t.Fatal(err) }
	for _, workerID := range []uint{11, 12} {
		if err := db.Create(&models.User{ID: workerID, Phone: fmt.Sprintf("+919900000%d", workerID), Role: "worker"}).Error; err != nil { t.Fatal(err) }
		if err := db.Create(&models.WorkerProfile{WorkerID: workerID, Name: "Worker", ZoneID: zone.ID, VehicleType: "bike", UPIId: "w@upi", AQIZone: "medium", TotalEarningsLifetime: 100000}).Error; err != nil { t.Fatal(err) }
		if err := db.Create(&models.Policy{WorkerID: workerID, Status: "active", PremiumAmount: 22}).Error; err != nil { t.Fatal(err) }
		if err := db.Create(&models.EarningsBaseline{WorkerID: workerID, BaselineAmount: 4000, LastUpdatedAt: now}).Error; err != nil { t.Fatal(err) }
		if err := db.Create(&models.WeeklyEarningsSummary{WorkerID: workerID, WeekStart: weekStart, WeekEnd: weekEnd, TotalEarnings: 1200, ClaimEligible: true}).Error; err != nil { t.Fatal(err) }
	}

	start := now.Add(-2 * time.Hour)
	confirmed := start.Add(15 * time.Minute)
	disruption := models.Disruption{ZoneID: zone.ID, Type: "heavy_rain", Severity: "high", Confidence: 0.88, Status: "confirmed", StartTime: &start, ConfirmedAt: &confirmed}
	if err := db.Create(&disruption).Error; err != nil { t.Fatal(err) }

	claimResult, err := service.GenerateClaimsForDisruption(disruption.ID, now)
	if err != nil { t.Fatalf("generate claims failed: %v", err) }
	if claimResult.ClaimsGenerated != 2 { t.Fatalf("expected 2 generated claims, got %+v", claimResult) }

	var claims []models.Claim
	if err := db.Order("worker_id asc").Find(&claims).Error; err != nil { t.Fatal(err) }
	for _, claim := range claims {
		if _, err := service.QueueClaimPayout(claim.ID); err != nil {
			t.Fatalf("queue payout failed for claim %d: %v", claim.ID, err)
		}
	}

	process1, err := service.ProcessQueuedPayouts(now)
	if err != nil { t.Fatalf("process payouts failed: %v", err) }
	if process1.Processed != 2 || process1.Succeeded != 1 || process1.Failed != 1 || process1.Retried != 1 {
		t.Fatalf("unexpected process1 result: %+v", process1)
	}

	process2, err := service.ProcessQueuedPayouts(now.Add(10 * time.Minute))
	if err != nil { t.Fatalf("second payout processing failed: %v", err) }
	if process2.Succeeded != 1 || process2.Failed != 0 {
		t.Fatalf("unexpected process2 result: %+v", process2)
	}
}

func TestPayoutReconciliationMath(t *testing.T) {
	db := setupCoreOpsTestDB(t)
	service := NewCoreOpsService(db)
	now := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)

	claim1 := models.Claim{ID: 1, WorkerID: 1, ClaimAmount: 500, Status: "paid"}
	claim2 := models.Claim{ID: 2, WorkerID: 2, ClaimAmount: 300, Status: "queued_for_payout"}
	if err := db.Create(&claim1).Error; err != nil { t.Fatal(err) }
	if err := db.Create(&claim2).Error; err != nil { t.Fatal(err) }
	processedAt := now
	if err := db.Create(&models.Payout{ClaimID: 1, WorkerID: 1, Amount: 500, Status: "processed", IdempotencyKey: "pay1", ProcessedAt: &processedAt, CreatedAt: now}).Error; err != nil { t.Fatal(err) }
	if err := db.Create(&models.Payout{ClaimID: 2, WorkerID: 2, Amount: 300, Status: "retry_pending", IdempotencyKey: "pay2", CreatedAt: now}).Error; err != nil { t.Fatal(err) }

	result, err := service.GetPayoutReconciliation(now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil { t.Fatalf("reconciliation failed: %v", err) }
	if result.Counts["processed"] != 1 || result.Counts["retry_pending"] != 1 {
		t.Fatalf("unexpected counts: %+v", result.Counts)
	}
	if result.Totals["processed_amount"] != 500 || result.Totals["retry_amount"] != 300 {
		t.Fatalf("unexpected totals: %+v", result.Totals)
	}
	if !result.ReconciliationOK || result.MismatchCount != 0 {
		t.Fatalf("expected reconciliation ok, got %+v", result)
	}
}

func TestSyntheticGenerationOutputs(t *testing.T) {
	db := setupCoreOpsTestDB(t)
	service := NewCoreOpsService(db)
	outputDir := filepath.Join(t.TempDir(), "synthetic")
	now := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)

	result, err := service.GenerateSyntheticData(SyntheticGenerateRequest{Seed: 42, Scenario: "fraud_burst", OutputDir: outputDir}, now)
	if err != nil {
		t.Fatalf("synthetic generation failed: %v", err)
	}
	if result.Counts["workers"] != 500 || result.Counts["claims"] != 2000 {
		t.Fatalf("unexpected synthetic counts: %+v", result.Counts)
	}
	for _, key := range []string{"workers_csv", "claims_csv", "payouts_csv", "seed_sql"} {
		if _, err := os.Stat(result.Artifacts[key]); err != nil {
			t.Fatalf("expected artifact %s at %s: %v", key, result.Artifacts[key], err)
		}
	}
}
