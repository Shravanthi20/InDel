package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/events"
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/logging"
	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

// PolicyActivationWorker polls the database for LOCKED policies whose lock-in
// window has expired and promotes them to ACTIVE status.
//
// Design:
//   - Polls every PollInterval (default 60s)
//   - Uses atomic UPDATE WHERE status='locked' to be idempotent across instances
//   - Publishes policy.activated Kafka event on each promotion
//   - Writes a PolicyAuditLog entry for fraud traceability
type PolicyActivationWorker struct {
	DB            *gorm.DB
	KafkaProducer *kafka.Producer
	PollInterval  time.Duration
}

// NewPolicyActivationWorker creates a worker with sensible defaults.
func NewPolicyActivationWorker(db *gorm.DB, producer *kafka.Producer) *PolicyActivationWorker {
	return &PolicyActivationWorker{
		DB:            db,
		KafkaProducer: producer,
		PollInterval:  60 * time.Second,
	}
}

// Start begins the activation polling loop. It blocks until ctx is cancelled.
func (w *PolicyActivationWorker) Start(ctx context.Context) {
	logger := logging.WithContext(map[string]interface{}{
		"service": "policy-activation-worker",
		"poll_interval": w.PollInterval.String(),
	})

	logger.Info("PolicyActivationWorker started")

	// Run once immediately, then on interval
	w.runActivationCycle(ctx)

	ticker := time.NewTicker(w.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("PolicyActivationWorker shutting down")
			return
		case <-ticker.C:
			w.runActivationCycle(ctx)
		}
	}
}

// runActivationCycle finds all expired LOCKED policies and activates them.
func (w *PolicyActivationWorker) runActivationCycle(ctx context.Context) {
	logger := logging.WithContext(logging.ExtractContextFromContext(ctx))
	logger = logger.WithContext(map[string]interface{}{
		"service": "policy-activation-worker",
		"operation": "activation_cycle",
	})

	if w.DB == nil {
		logger.Error("Database is nil, cannot run activation cycle", fmt.Errorf("database unavailable"))
		return
	}

	now := time.Now().UTC()

	// Find all policies that are locked and past their lock-in end time.
	type lockedPolicy struct {
		ID       uint `gorm:"column:id"`
		WorkerID uint `gorm:"column:worker_id"`
	}

	var candidates []lockedPolicy
	if err := w.DB.Raw(`
		SELECT id, worker_id
		FROM policies
		WHERE status = 'locked'
		  AND lock_in_end_time IS NOT NULL
		  AND lock_in_end_time <= ?
		ORDER BY lock_in_end_time ASC
	`, now).Scan(&candidates).Error; err != nil {
		logger.Error("Error querying locked policies", err)
		return
	}

	if len(candidates) == 0 {
		logger.Debug("No policies ready for activation")
		return
	}

	logger.Info("Found policies ready for activation", map[string]interface{}{
		"candidate_count": len(candidates),
	})

	for _, candidate := range candidates {
		// Check if context is cancelled before processing each candidate
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, stopping activation cycle")
			return
		default:
			policyLogger := logger.WithContext(map[string]interface{}{
				"policy_id": candidate.ID,
				"worker_id": candidate.WorkerID,
			})
			w.activatePolicy(ctx, candidate.ID, candidate.WorkerID, now, policyLogger)
		}
	}
}

// activatePolicy promotes a single policy from LOCKED to ACTIVE.
// It is idempotent: if another instance already activated it, the UPDATE
// affects 0 rows and we safely skip.
func (w *PolicyActivationWorker) activatePolicy(ctx context.Context, policyID, workerID uint, now time.Time, logger *logging.ContextLogger) {
	// Atomic transition: only update if still 'locked'
	result := w.DB.Exec(
		"UPDATE policies SET status = 'active', updated_at = ? WHERE id = ? AND status = 'locked'",
		now, policyID,
	)
	if result.Error != nil {
		logger.Error("Failed to activate policy", result.Error)
		return
	}
	if result.RowsAffected == 0 {
		// Already activated by another goroutine/instance — skip silently
		logger.Debug("Policy already activated by another instance", map[string]interface{}{
			"rows_affected": result.RowsAffected,
		})
		return
	}

	logger.Info("Policy successfully activated", map[string]interface{}{
		"activated_at": now.Format(time.RFC3339),
	})

	// Fetch zone for active_policies upsert
	var zoneRow struct {
		ZoneID   uint   `gorm:"column:zone_id"`
		ZoneName string `gorm:"column:zone_name"`
	}
	if err := w.DB.Raw(`
		SELECT wp.zone_id, COALESCE(z.name, 'Unknown') AS zone_name
		FROM worker_profiles wp
		LEFT JOIN zones z ON z.id = wp.zone_id
		WHERE wp.worker_id = ?
		LIMIT 1
	`, workerID).Scan(&zoneRow).Error; err != nil {
		logger.Error("ERROR fetching zone for policy", err, map[string]interface{}{
			"policy_id": policyID,
			"worker_id": workerID,
		})
		return
	}
	if zoneRow.ZoneID == 0 {
		logger.Error("ERROR invalid zone (zone_id=0)", fmt.Errorf("zone_id is 0"), map[string]interface{}{
			"policy_id": policyID,
			"worker_id": workerID,
		})
		return
	}

	// Upsert into active_policies
	if err := w.DB.Exec(`
		INSERT INTO active_policies (user_id, policy_id, zone, started_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			policy_id  = EXCLUDED.policy_id,
			zone       = EXCLUDED.zone,
			updated_at = EXCLUDED.updated_at
	`, workerID, policyID, zoneRow.ZoneName, now, now).Error; err != nil {
		logger.Error("Failed to upsert active_policies", err)
		// Continue with audit log and Kafka event even if active_policies fails
	} else {
		logger.Debug("Successfully upserted active_policies", map[string]interface{}{
			"zone_name": zoneRow.ZoneName,
		})
	}

	// Write audit log
	audit := models.PolicyAuditLog{
		PolicyID:   policyID,
		WorkerID:   workerID,
		Action:     "policy_activated",
		FromStatus: models.PolicyStatusLocked,
		ToStatus:   models.PolicyStatusActive,
		Reason:     fmt.Sprintf("Lock-in period expired at %s", now.Format(time.RFC3339)),
		CreatedAt:  now,
	}
	if err := w.DB.Create(&audit).Error; err != nil {
		logger.Error("Failed to write audit log", err)
		// Continue with Kafka event even if audit log fails
	} else {
		logger.Debug("Successfully wrote audit log")
	}

	// Publish policy.activated Kafka event
	w.publishActivated(ctx, policyID, workerID, now, logger)
}

// publishActivated emits a policy.activated Kafka event.
func (w *PolicyActivationWorker) publishActivated(ctx context.Context, policyID, workerID uint, activatedAt time.Time, logger *logging.ContextLogger) {
	if w.KafkaProducer == nil {
		logger.Warn("Kafka producer is nil, skipping event publication")
		return
	}

	evt := events.PolicyActivatedEvent{
		EventType:   "policy.activated",
		PolicyID:    policyID,
		WorkerID:    workerID,
		ActivatedAt: activatedAt,
		Timestamp:   activatedAt,
	}
	b, err := json.Marshal(evt)
	if err != nil {
		logger.Error("Failed to marshal policy.activated event", err)
		return
	}

	key := fmt.Sprintf("policy-%d", policyID)
	if err := w.KafkaProducer.Publish(kafka.TopicPolicyActivated, key, b); err != nil {
		logger.Error("Failed to publish policy.activated event", err)
		// This is a non-critical error - policy activation succeeded even if Kafka fails
	} else {
		logger.Info("Successfully published policy.activated event", map[string]interface{}{
			"topic": kafka.TopicPolicyActivated,
			"key": key,
		})
	}
}
