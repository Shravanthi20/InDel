package worker

import (
	"fmt"
	"time"

	workerModels "github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// nowTime returns current UTC time for disruption records.
func nowTime() time.Time {
	return time.Now().UTC()
}

// workerCoreService returns a CoreOpsService backed by the worker DB.
func workerCoreService() *services.CoreOpsService {
	return services.NewCoreOpsService(workerDB, nil)
}


// DemoReset resets all in-memory demo state.
func DemoReset(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec("DELETE FROM notifications WHERE worker_id = ?", workerIDUint).Error
			_ = workerDB.Exec("DELETE FROM auth_tokens WHERE user_id = ?", workerIDUint).Error
			_ = workerDB.Exec("UPDATE orders SET status='assigned', accepted_at=NULL, picked_up_at=NULL, delivered_at=NULL, updated_at=CURRENT_TIMESTAMP WHERE worker_id = ?", workerIDUint).Error
		}
	}
	store.reset()
	c.JSON(200, gin.H{"message": "demo_reset", "time": nowISO()})
}

// DemoTriggerDisruption creates a disruption and runs the full pipeline:
// notify workers → generate claims → fraud check → queue payouts → process payouts.
func DemoTriggerDisruption(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	disruptionType := bodyString(body, "disruption_type", "heavy_rain")
	zone := bodyString(body, "zone", "Tambaram, Chennai")
	severity := bodyString(body, "severity", "high")

	// In-memory notification (fallback for no-DB mode)
	msg := disruptionType + " detected in " + zone + ". You are protected."
	store.mu.Lock()
	store.data.Notifications = append([]map[string]any{{
		"id":         nextID("ntf", len(store.data.Notifications)),
		"type":       "disruption_alert",
		"title":      "Disruption detected",
		"body":       msg,
		"created_at": nowISO(),
		"read":       false,
	}}, store.data.Notifications...)
	store.mu.Unlock()

	// Full pipeline if DB is connected
	if hasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr != nil {
			c.JSON(400, gin.H{"error": "invalid_worker_id"})
			return
		}

		// Find worker's zone
		var zoneID uint = 1
		_ = workerDB.Raw("SELECT zone_id FROM worker_profiles WHERE worker_id = ? LIMIT 1", workerIDUint).Scan(&zoneID).Error

		// 1. Create disruption record
		now := nowTime()
		confirmedAt := now.Add(1)
		disruption := workerModels.Disruption{
			ZoneID:          zoneID,
			Type:            disruptionType,
			Severity:        severity,
			Confidence:      0.91,
			Status:          "confirmed",
			SignalTimestamp: &now,
			ConfirmedAt:     &confirmedAt,
			StartTime:       &now,
		}
		if err := workerDB.Create(&disruption).Error; err != nil {
			c.JSON(500, gin.H{"error": "failed_to_create_disruption", "detail": err.Error()})
			return
		}

		// 2. Auto-process: notify → claims → payouts
		coreSvc := workerCoreService()
		result, err := coreSvc.AutoProcessDisruption(disruption.ID, now)
		if err != nil {
			c.JSON(500, gin.H{"error": "pipeline_failed", "detail": err.Error()})
			return
		}

		// Mark processed
		processed := nowTime()
		_ = workerDB.Model(&workerModels.Disruption{}).Where("id = ?", disruption.ID).Update("processed_at", processed).Error

		c.JSON(200, gin.H{
			"message":          "disruption_pipeline_complete",
			"disruption_id":    fmt.Sprintf("dis_%d", disruption.ID),
			"disruption_type":  disruptionType,
			"zone":             zone,
			"workers_notified": result.WorkersNotified,
			"claims_generated": result.ClaimsGenerated,
			"payouts_succeeded": result.PayoutsSucceeded,
			"manual_review":    result.ManualReviewClaims,
			"status":           result.Status,
			"time":             nowISO(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message":         "disruption_triggered",
		"disruption_type": disruptionType,
		"zone":            zone,
		"time":            nowISO(),
	})
}

// DemoSimulateOrders appends assigned orders for demo.
func DemoSimulateOrders(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	count := bodyInt(body, "count", 3)
	if count <= 0 {
		count = 1
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			var zoneID uint = 1
			_ = workerDB.Raw("SELECT zone_id FROM worker_profiles WHERE worker_id = ? LIMIT 1", workerIDUint).Scan(&zoneID).Error
			for i := 0; i < count; i++ {
				_ = workerDB.Exec(
					"INSERT INTO orders (worker_id, zone_id, order_value, status, pickup_area, drop_area, distance_km, updated_at) VALUES (?, ?, ?, 'assigned', ?, ?, ?, CURRENT_TIMESTAMP)",
					workerIDUint, zoneID, 55+i*8, "Tambaram", "Camp Road", 2.5+float64(i)*0.4,
				).Error
			}
		}
	}

	store.mu.Lock()
	base := len(store.data.Orders)
	for i := 0; i < count; i++ {
		store.data.Orders = append(store.data.Orders, map[string]any{
			"order_id":    nextID("ord", base+i),
			"pickup_area": "Tambaram",
			"drop_area":   "Camp Road",
			"distance_km": 2.5 + float64(i)*0.4,
			"earning_inr": 55 + i*8,
			"status":      "assigned",
			"assigned_at": nowISO(),
		})
	}
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "orders_simulated", "count": count})
}

// DemoSettleEarnings settles demo earnings and triggers premium reminder.
func DemoSettleEarnings(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec(
				`UPDATE weekly_earnings_summary
				 SET claim_eligible = TRUE
				 WHERE worker_id = ?
				   AND week_start = date_trunc('week', CURRENT_DATE)::date
				   AND week_end = (date_trunc('week', CURRENT_DATE)::date + INTERVAL '6 day')::date`,
				workerIDUint,
			).Error
			_ = workerDB.Exec(
				"INSERT INTO notifications (worker_id, type, message) VALUES (?, 'premium_due', 'Weekly earnings settled. Pay premium to keep coverage active.')",
				workerIDUint,
			).Error
		}
	}

	store.mu.Lock()
	store.data.Notifications = append([]map[string]any{{
		"id":         nextID("ntf", len(store.data.Notifications)),
		"type":       "premium_due",
		"title":      "Weekly settlement complete",
		"body":       "Weekly earnings settled. Pay premium to keep coverage active.",
		"created_at": nowISO(),
		"read":       false,
	}}, store.data.Notifications...)
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "earnings_settled", "time": nowISO()})
}

// DemoResetZone resets disruption and claim state for demo replay.
func DemoResetZone(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec("DELETE FROM payouts WHERE worker_id = ?", workerIDUint).Error
			_ = workerDB.Exec("DELETE FROM claims WHERE worker_id = ?", workerIDUint).Error
			_ = workerDB.Exec("DELETE FROM notifications WHERE worker_id = ? AND type IN ('disruption_alert', 'payout_credited')", workerIDUint).Error
		}
	}

	store.mu.Lock()
	store.data.Claims = []map[string]any{}
	store.data.Payouts = []map[string]any{}
	store.data.Notifications = []map[string]any{}
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "zone_reset", "time": nowISO()})
}
