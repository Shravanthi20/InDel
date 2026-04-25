package workers

import (
	"context"
	"testing"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite DB with the minimal schema for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	// Create tables
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS policies (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			worker_id        INTEGER NOT NULL,
			plan_id          TEXT DEFAULT '',
			status           TEXT NOT NULL DEFAULT 'active',
			premium_amount   REAL NOT NULL DEFAULT 0,
			policy_cycle_id  INTEGER DEFAULT 0,
			lock_in_start_time DATETIME,
			lock_in_end_time   DATETIME,
			idempotency_key  TEXT,
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		t.Fatalf("failed to create policies table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS active_policies (
			user_id    INTEGER PRIMARY KEY,
			policy_id  INTEGER NOT NULL,
			zone       TEXT,
			started_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("failed to create active_policies table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS worker_profiles (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			worker_id INTEGER UNIQUE NOT NULL,
			zone_id   INTEGER DEFAULT 1
		)
	`).Error; err != nil {
		t.Fatalf("failed to create worker_profiles table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS zones (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT 'TestZone'
		)
	`).Error; err != nil {
		t.Fatalf("failed to create zones table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS policy_audit_logs (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			policy_id   INTEGER NOT NULL,
			worker_id   INTEGER NOT NULL,
			action      TEXT,
			from_status TEXT,
			to_status   TEXT,
			reason      TEXT,
			metadata    TEXT,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		t.Fatalf("failed to create policy_audit_logs table: %v", err)
	}

	// Seed zone and worker profile
	db.Exec("INSERT INTO zones (id, name) VALUES (1, 'TestZone')")
	db.Exec("INSERT INTO worker_profiles (worker_id, zone_id) VALUES (1, 1)")

	return db
}

// insertLockedPolicy creates a locked policy with the given lock_in_end_time.
func insertLockedPolicy(t *testing.T, db *gorm.DB, workerID uint, lockEnd time.Time) uint {
	t.Helper()
	lockStart := lockEnd.Add(-48 * time.Hour)
	p := models.Policy{
		WorkerID:        workerID,
		Status:          models.PolicyStatusLocked,
		PremiumAmount:   22,
		LockInStartTime: &lockStart,
		LockInEndTime:   &lockEnd,
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("failed to insert locked policy: %v", err)
	}
	return p.ID
}

// TestActivationWorker_ActivatesExpiredLockedPolicy verifies that a policy
// whose lock_in_end_time is in the past gets promoted to ACTIVE.
func TestActivationWorker_ActivatesExpiredLockedPolicy(t *testing.T) {
	db := setupTestDB(t)
	w := NewPolicyActivationWorker(db, nil) // nil producer = no Kafka

	// Insert a locked policy that expired 1 second ago
	pastEnd := time.Now().UTC().Add(-1 * time.Second)
	policyID := insertLockedPolicy(t, db, 1, pastEnd)

	// Run one activation cycle
	w.runActivationCycle()

	// Assert: policy status is now 'active'
	var p models.Policy
	if err := db.First(&p, policyID).Error; err != nil {
		t.Fatalf("policy not found: %v", err)
	}
	if p.Status != models.PolicyStatusActive {
		t.Errorf("expected status=%q, got %q", models.PolicyStatusActive, p.Status)
	}
}

// TestActivationWorker_DoesNotActivateFutureLockedPolicy verifies that a policy
// whose lock_in_end_time is still in the future is NOT activated.
func TestActivationWorker_DoesNotActivateFutureLockedPolicy(t *testing.T) {
	db := setupTestDB(t)
	w := NewPolicyActivationWorker(db, nil)

	// Lock-in ends in 24 hours
	futureEnd := time.Now().UTC().Add(24 * time.Hour)
	policyID := insertLockedPolicy(t, db, 2, futureEnd)

	w.runActivationCycle()

	var p models.Policy
	if err := db.First(&p, policyID).Error; err != nil {
		t.Fatalf("policy not found: %v", err)
	}
	if p.Status != models.PolicyStatusLocked {
		t.Errorf("expected status=%q (future lock-in), got %q", models.PolicyStatusLocked, p.Status)
	}
}

// TestActivationWorker_IdempotentActivation verifies that running the cycle
// twice on the same policy does not error or produce duplicate rows.
func TestActivationWorker_IdempotentActivation(t *testing.T) {
	db := setupTestDB(t)
	w := NewPolicyActivationWorker(db, nil)

	pastEnd := time.Now().UTC().Add(-1 * time.Second)
	policyID := insertLockedPolicy(t, db, 3, pastEnd)

	// Run twice
	w.runActivationCycle()
	w.runActivationCycle() // second run should be a no-op

	var p models.Policy
	if err := db.First(&p, policyID).Error; err != nil {
		t.Fatalf("policy not found: %v", err)
	}
	if p.Status != models.PolicyStatusActive {
		t.Errorf("expected status=%q, got %q", models.PolicyStatusActive, p.Status)
	}

	// Check active_policies has exactly 1 row for this worker
	var count int64
	db.Table("active_policies").Where("user_id = ?", 3).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row in active_policies, got %d", count)
	}
}

// TestActivationWorker_WritesAuditLog verifies that a PolicyAuditLog entry is
// created when a policy is activated.
func TestActivationWorker_WritesAuditLog(t *testing.T) {
	db := setupTestDB(t)
	w := NewPolicyActivationWorker(db, nil)

	pastEnd := time.Now().UTC().Add(-1 * time.Second)
	policyID := insertLockedPolicy(t, db, 4, pastEnd)

	w.runActivationCycle()

	var logEntry models.PolicyAuditLog
	if err := db.Where("policy_id = ? AND action = ?", policyID, "policy_activated").
		First(&logEntry).Error; err != nil {
		t.Errorf("expected audit log entry for policy_activated, got error: %v", err)
	}
	if logEntry.FromStatus != models.PolicyStatusLocked {
		t.Errorf("expected from_status=%q, got %q", models.PolicyStatusLocked, logEntry.FromStatus)
	}
	if logEntry.ToStatus != models.PolicyStatusActive {
		t.Errorf("expected to_status=%q, got %q", models.PolicyStatusActive, logEntry.ToStatus)
	}
}

// TestActivationWorker_ContextCancellation verifies that the worker
// exits cleanly when its context is cancelled.
func TestActivationWorker_ContextCancellation(t *testing.T) {
	db := setupTestDB(t)
	w := NewPolicyActivationWorker(db, nil)
	w.PollInterval = 100 * time.Millisecond // fast poll for test

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		w.Start(ctx)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK — worker exited
	case <-time.After(2 * time.Second):
		t.Error("worker did not exit after context cancellation")
	}
}
