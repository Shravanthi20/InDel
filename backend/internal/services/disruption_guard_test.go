package services

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupGuardTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS disruptions (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			zone_id          INTEGER NOT NULL,
			type             TEXT DEFAULT 'heavy_rain',
			severity         TEXT DEFAULT 'high',
			confidence       REAL DEFAULT 0.9,
			status           TEXT DEFAULT 'pending',
			signal_timestamp DATETIME,
			confirmed_at     DATETIME,
			start_time       DATETIME,
			end_time         DATETIME,
			processed_at     DATETIME,
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		t.Fatalf("failed to create disruptions table: %v", err)
	}
	return db
}

// TestDisruptionGuard_NoDisruption verifies that an empty zone passes the guard.
func TestDisruptionGuard_NoDisruption(t *testing.T) {
	db := setupGuardTestDB(t)
	result, err := IsDisruptionActiveOrPredicted(db, 1, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Blocked {
		t.Errorf("expected Blocked=false for empty zone, got true")
	}
}

// TestDisruptionGuard_ActiveDisruptionBlocks verifies that an active disruption blocks purchase.
func TestDisruptionGuard_ActiveDisruptionBlocks(t *testing.T) {
	db := setupGuardTestDB(t)
	db.Exec(`INSERT INTO disruptions (zone_id, type, severity, status) VALUES (1, 'heavy_rain', 'high', 'active')`)

	result, err := IsDisruptionActiveOrPredicted(db, 1, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Blocked {
		t.Error("expected Blocked=true for active disruption, got false")
	}
	if result.Reason != "active_disruption" {
		t.Errorf("expected reason=%q, got %q", "active_disruption", result.Reason)
	}
}

// TestDisruptionGuard_ConfirmedDisruptionBlocks verifies that a confirmed disruption blocks purchase.
func TestDisruptionGuard_ConfirmedDisruptionBlocks(t *testing.T) {
	db := setupGuardTestDB(t)
	db.Exec(`INSERT INTO disruptions (zone_id, type, severity, status) VALUES (1, 'flood', 'medium', 'confirmed')`)

	result, err := IsDisruptionActiveOrPredicted(db, 1, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Blocked {
		t.Error("expected Blocked=true for confirmed disruption, got false")
	}
}

// TestDisruptionGuard_PredictedDisruptionWithinWindow blocks when signal is within lookahead.
func TestDisruptionGuard_PredictedDisruptionWithinWindow(t *testing.T) {
	db := setupGuardTestDB(t)

	// Signal in 6 hours — within 12h lookahead
	signal := time.Now().UTC().Add(6 * time.Hour)
	db.Exec(`INSERT INTO disruptions (zone_id, type, severity, status, signal_timestamp)
		VALUES (1, 'cyclone', 'high', 'predicted', ?)`, signal)

	result, err := IsDisruptionActiveOrPredicted(db, 1, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Blocked {
		t.Error("expected Blocked=true for predicted disruption within window, got false")
	}
	if result.Reason != "predicted_disruption" {
		t.Errorf("expected reason=%q, got %q", "predicted_disruption", result.Reason)
	}
}

// TestDisruptionGuard_PredictedDisruptionOutsideWindow allows when signal is beyond lookahead.
func TestDisruptionGuard_PredictedDisruptionOutsideWindow(t *testing.T) {
	db := setupGuardTestDB(t)

	// Signal in 24 hours — beyond 12h lookahead
	signal := time.Now().UTC().Add(24 * time.Hour)
	db.Exec(`INSERT INTO disruptions (zone_id, type, severity, status, signal_timestamp)
		VALUES (1, 'cyclone', 'high', 'predicted', ?)`, signal)

	result, err := IsDisruptionActiveOrPredicted(db, 1, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Blocked {
		t.Error("expected Blocked=false for predicted disruption outside window, got true")
	}
}

// TestDisruptionGuard_DifferentZoneDoesNotBlock verifies zone isolation.
func TestDisruptionGuard_DifferentZoneDoesNotBlock(t *testing.T) {
	db := setupGuardTestDB(t)
	// Active disruption in zone 99, checking zone 1
	db.Exec(`INSERT INTO disruptions (zone_id, type, severity, status) VALUES (99, 'heavy_rain', 'high', 'active')`)

	result, err := IsDisruptionActiveOrPredicted(db, 1, 12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Blocked {
		t.Error("expected Blocked=false for disruption in different zone, got true")
	}
}

// TestDisruptionGuard_NilDBFailOpen verifies that a nil DB does not block purchases.
func TestDisruptionGuard_NilDBFailOpen(t *testing.T) {
	result, err := IsDisruptionActiveOrPredicted(nil, 1, 12)
	if err != nil {
		t.Fatalf("unexpected error with nil DB: %v", err)
	}
	if result.Blocked {
		t.Error("expected Blocked=false (fail-open) with nil DB, got true")
	}
}
