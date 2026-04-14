package platform

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// In-memory state tracking for Order metrics, Idempotency, and External Signals per zone
type ZoneSignalState struct {
	mu                 sync.Mutex
	RecentOrders       []time.Time
	BaselineOrders     float64
	LastBaselineUpdate time.Time
	ActiveSignals      map[string]bool
	LastResetAt        time.Time // Added for simulator sync
	TotalOrdersEver    uint64    // Added for data-driven warm-up
}

// Global engine state
var (
	zoneSignals = make(map[uint]*ZoneSignalState)
	engineMu    sync.Mutex

	// Idempotency cache
	processedOrderIds = make(map[string]time.Time)
	idCacheMu         sync.Mutex
)

// ResetEngineForTests allows the testing suite to natively flush caches before independent scenarios
func ResetEngineForTests() {
	engineMu.Lock()
	zoneSignals = make(map[uint]*ZoneSignalState)
	engineMu.Unlock()
	
	idCacheMu.Lock()
	processedOrderIds = make(map[string]time.Time)
	idCacheMu.Unlock()
}

// getOrCreateZoneState returns a concurrency safe reference to a ZoneSignalState
func getOrCreateZoneState(zoneID uint) *ZoneSignalState {
	engineMu.Lock()
	defer engineMu.Unlock()
	if state, exists := zoneSignals[zoneID]; exists {
		return state
	}
	newState := &ZoneSignalState{
		BaselineOrders:     20.0, // Start with a sensible default of 20 orders per 10min
		LastBaselineUpdate: time.Now(),
		ActiveSignals:      make(map[string]bool),
	}
	zoneSignals[zoneID] = newState
	return newState
}

// checkAndCacheOrderId provides idempotency with a TTL to prevent memory leaks
func checkAndCacheOrderId(orderID string) bool {
	idCacheMu.Lock()
	defer idCacheMu.Unlock()

	now := time.Now()
	// Periodic cleanup of keys older than 30 minutes
	for k, timestamp := range processedOrderIds {
		if now.Sub(timestamp) > 30*time.Minute {
			delete(processedOrderIds, k)
		}
	}

	if _, exists := processedOrderIds[orderID]; exists {
		return false // Already processed
	}
	processedOrderIds[orderID] = now
	return true
}

// ----- 1. Order tracking per zone & Idempotency -----
// CheckAndTrackOrder processes a webhook. Returns true if it was processed, false if skipped by idempotency check.
func CheckAndTrackOrder(orderID string, zoneID uint, isCompleted bool) bool {
	if !checkAndCacheOrderId(orderID) {
		return false
	}

	state := getOrCreateZoneState(zoneID)
	state.mu.Lock()
	if isCompleted {
		state.RecentOrders = append(state.RecentOrders, time.Now())
	}
	// For cancellation, we don't add to recent orders, effectively lowering the count
	state.mu.Unlock()

	// evaluate synchronously
	evaluateDisruption(zoneID, state)

	return true
}

// ----- 3. External signal flag -----
// SetExternalSignal allows setting specific typed signals (weather, aqi, system)
func SetExternalSignal(zoneID uint, signalType string, isActive bool) {
	state := getOrCreateZoneState(zoneID)
	state.mu.Lock()
	if isActive {
		state.ActiveSignals[signalType] = true
	} else {
		delete(state.ActiveSignals, signalType)
	}
	state.mu.Unlock()
	evaluateDisruption(zoneID, state)
}

// calculateZoneStats provides a single-source-of-truth for health metrics
func calculateZoneStats(state *ZoneSignalState) (int, float64, float64, string) {
	now := time.Now()
	
	// 1. Data Availability Guardrail (Warm-up)
	// System remains "Healthy" until it has seen at least 3 orders total (Reduced for demo)
	if state.TotalOrdersEver <= 3 {
		return len(state.RecentOrders), state.BaselineOrders, 0.0, "healthy"
	}

	// 2. Window Filtering (30s window to capture drop orders better)
	var currentWindow []time.Time
	for _, t := range state.RecentOrders {
		if now.Sub(t) <= 30*time.Second {
			currentWindow = append(currentWindow, t)
		}
	}
	
	current := len(currentWindow)
	var orderDrop float64
	if state.BaselineOrders > 0 {
		orderDrop = (state.BaselineOrders - float64(current)) / state.BaselineOrders
	}

	// 3. Strict Clamping Guardrail (Clean Math)
	if orderDrop < 0 {
		orderDrop = 0.0
	} else if orderDrop > 1.0 {
		orderDrop = 1.0
	}

	status := "healthy"
	if orderDrop > 0.30 && len(state.ActiveSignals) > 0 {
		status = "disrupted"
	} else if orderDrop > 0.30 {
		status = "anomalous_demand"
	} else if len(state.ActiveSignals) > 0 {
		status = "monitoring"
	}

	return current, state.BaselineOrders, orderDrop, status
}

// ----- 2. Drop calculation & 4. Disruption creation -----
func evaluateDisruption(zoneID uint, state *ZoneSignalState) {
	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()
	
	// Increment total orders (capped for safety)
	if state.TotalOrdersEver < 1000 {
		state.TotalOrdersEver++
	}

	current, _, _, status := calculateZoneStats(state)

	// Persist the filtered window to the state (Keeping 60s for historical context)
	var persistentWindow []time.Time
	for _, t := range state.RecentOrders {
		if now.Sub(t) <= 60*time.Second {
			persistentWindow = append(persistentWindow, t)
		}
	}
	state.RecentOrders = persistentWindow

	// Adaptive Baseline Guardrail (Growing fast, decaying slow, floor at 5.0)
	newBaseline := state.BaselineOrders * 0.95
	if float64(current) > newBaseline {
		newBaseline = float64(current)
	}
	if newBaseline < 5.0 {
		newBaseline = 5.0
	}
	state.BaselineOrders = newBaseline
	state.LastBaselineUpdate = now

	// Recalculate drop after baseline update for logging
	orderDrop := 0.0
	if state.BaselineOrders > 0 {
		orderDrop = (state.BaselineOrders - float64(current)) / state.BaselineOrders
	}
	// Clamp again
	if orderDrop < 0 { orderDrop = 0 }
	if orderDrop > 1 { orderDrop = 1 }

    // Print to logs so we can see the exact math!
    log.Printf("[DECISION ENGINE] Zone %d | Output: %s | Window: %d / %.0f (Drop: %.1f%%)", zoneID, strings.ToUpper(status), current, state.BaselineOrders, orderDrop*100)

	// 3. Multi-Signal Validation
	hasExternalSignals := len(state.ActiveSignals) > 0

	if orderDrop > 0.40 { // Lower threshold for demo - auto-trigger even without signal
		createDisruptionRecord(zoneID, orderDrop, state.ActiveSignals)
	} else if orderDrop > 0.30 && hasExternalSignals {
		createDisruptionRecord(zoneID, orderDrop, state.ActiveSignals)
	}
}

func createDisruptionRecord(zoneID uint, orderDrop float64, signals map[string]bool) {
	if !hasDB() {
		return
	}

	// DEMO MODE: Reduced lock to 60 seconds (prevents double clicking UI)
	var existing int64
	cooldown := time.Now().Add(-60 * time.Second)
	platformDB.Model(&models.Disruption{}).Where("zone_id = ? AND created_at > ?", zoneID, cooldown).Count(&existing)
	if existing > 0 {
		log.Printf("[DISRUPTION] SKIPPED zone=%d reason=cooldown_active existing=%d (wait 60s)", zoneID, existing)
		return // Still active
	}

	var severity string
	if orderDrop >= 0.50 {
		severity = "HIGH"
	} else if orderDrop >= 0.40 {
		severity = "MEDIUM"
	} else {
		severity = "LOW"
	}

	// Confidence logic: base target drop + weight of multiple signals
	confidence := orderDrop + (float64(len(signals)) * 0.10)
	if confidence > 1.0 {
		confidence = 1.0
	}

	now := time.Now()
	// Build trigger string format: "weather + demand_drop"
	triggerStr := "demand_drop"
	for s := range signals {
		triggerStr = s + " + " + triggerStr
	}

	disruption := models.Disruption{
		ZoneID:          zoneID,
		Type:            triggerStr,
		Severity:        severity,
		Confidence:      confidence, // Use calculated confidence here

		Status:          "confirmed",
		StartTime:       &now,
		SignalTimestamp: &now, // THIS was missing, crashing the Postgres INSERT!
		ConfirmedAt:     &now,
	}

	if err := platformDB.Create(&disruption).Error; err != nil {
		log.Printf("Failed to create disruption: %v", err)
		return
	}

	// DYNAMIC PREMIUM: Bump the risk rating of the zone to increase future premiums
	_ = platformDB.Exec("UPDATE zones SET risk_rating = LEAST(risk_rating + 0.25, 1.0) WHERE id = ?", zoneID).Error
	log.Printf("[DYNAMIC PREMIUM] Zone %d risk_rating increased by 0.25", zoneID)

	if platformCoreOps != nil {
		if result, err := platformCoreOps.AutoProcessDisruption(disruption.ID, now.UTC()); err != nil {
			log.Printf("Failed to auto-process disruption %d: %v", disruption.ID, err)
		} else {
			log.Printf("[AUTOMATION] disruption=%s notified=%d claims=%d queued=%d processed=%d succeeded=%d failed=%d status=%s",
				result.DisruptionID,
				result.WorkersNotified,
				result.ClaimsGenerated,
				result.PayoutsQueued,
				result.PayoutsProcessed,
				result.PayoutsSucceeded,
				result.PayoutsFailed,
				result.Status,
			)
		}
	}
}

// ----- 5. API Exposure -----

// GetZoneHealth
func GetZoneHealth(c *gin.Context) {
	engineMu.Lock()
	zonesCopy := make([]uint, 0, len(zoneSignals))
	for z := range zoneSignals {
		zonesCopy = append(zonesCopy, z)
	}
	engineMu.Unlock()

	results := make([]map[string]interface{}, 0, len(zonesCopy))

	for _, zoneID := range zonesCopy {
		state := zoneSignals[zoneID]
		state.mu.Lock()

		current, baseline, drop, status := calculateZoneStats(state)

		results = append(results, map[string]interface{}{
			"zone_id":         zoneID,
			"order_drop":      drop,
			"current_orders":  current,
			"baseline_orders": baseline,
			"active_signals":  state.ActiveSignals,
			"status":          status,
			"last_reset_at":   state.LastResetAt.Unix(),
		})

		state.mu.Unlock()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"meta": gin.H{"timestamp": time.Now().UTC().Format(time.RFC3339)},
	})
}

// GetDisruptions
func GetDisruptions(c *gin.Context) {
	if !hasDB() {
		c.JSON(200, gin.H{"data": []interface{}{}})
		return
	}

	type disruptionRow struct {
		models.Disruption
		ClaimsGenerated   int64   `gorm:"column:claims_generated"`
		ClaimsInReview    int64   `gorm:"column:claims_in_review"`
		PayoutsProcessed  int64   `gorm:"column:payouts_processed"`
		PayoutAmountTotal float64 `gorm:"column:payout_amount_total"`
	}

	var records []disruptionRow
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	_ = platformDB.Table("disruptions d").
		Select(`
			d.*,
			COUNT(DISTINCT c.id) AS claims_generated,
			COUNT(DISTINCT CASE WHEN c.status = 'manual_review' THEN c.id END) AS claims_in_review,
			COUNT(DISTINCT CASE WHEN p.status IN ('processed', 'credited', 'completed') THEN p.id END) AS payouts_processed,
			COALESCE(SUM(CASE WHEN p.status IN ('processed', 'credited', 'completed') THEN p.amount ELSE 0 END), 0) AS payout_amount_total
		`).
		Joins("LEFT JOIN claims c ON c.disruption_id = d.id").
		Joins("LEFT JOIN payouts p ON p.claim_id = c.id").
		Where("d.created_at > ?", oneHourAgo).
		Group("d.id").
		Order("d.created_at desc").
		Limit(20).
		Scan(&records).Error

	results := make([]map[string]interface{}, 0, len(records))
	for _, r := range records {
		
		// Parse triggers back into the standardized `signals` array for the API contract
		signalsArr := make([]map[string]interface{}, 0)
		triggerParts := strings.Split(r.Type, " + ")
		for _, part := range triggerParts {
			var val float64
			if part == "demand_drop" {
				val = r.Confidence // The stored 'Confidence' field is holding the exact drop ratio
			} else {
				val = 1.0 // External signals like weather trigger at 1.0 boolean strength generally
			}
			signalsArr = append(signalsArr, map[string]interface{}{
				"source": part,
				"value":  val,
			})
		}
		
		// Recalculate the official confidence metric (drop severity + signal weight)
		officialConfidence := r.Confidence + (float64(len(triggerParts)-1) * 0.1)
		if officialConfidence > 1.0 {
			officialConfidence = 1.0
		}

		results = append(results, map[string]interface{}{
			"disruption_id": fmt.Sprintf("dis_%d", r.ID),
			"zone_id":       fmt.Sprintf("zone_%d", r.ZoneID), 
			"type":          r.Type, 
			"severity":      strings.ToLower(r.Severity),
			"confidence":    officialConfidence,
			"status":        "confirmed",
			"claims_generated": r.ClaimsGenerated,
			"claims_in_review": r.ClaimsInReview,
			"payouts_processed": r.PayoutsProcessed,
			"payout_amount_total": r.PayoutAmountTotal,
			"automation_status": func() string {
				switch {
				case r.ClaimsInReview > 0:
					return "manual_review"
				case r.PayoutsProcessed > 0:
					return "paid"
				case r.ClaimsGenerated > 0:
					return "queued"
				default:
					return "detected"
				}
			}(),
			"signals":       signalsArr,
			"started_at":    r.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"meta": gin.H{"timestamp": time.Now().UTC().Format(time.RFC3339)},
	})
}

// TriggerDemoDisruption
func TriggerDemoDisruption(c *gin.Context) {
	var req struct {
		ZoneID         uint   `json:"zone_id"`
		ForceOrderDrop bool   `json:"force_order_drop"`
		ExternalSignal string `json:"external_signal"` // e.g., "weather", "aqi", "system"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state := getOrCreateZoneState(req.ZoneID)
	state.mu.Lock()
	// Set the Reset Marker
	state.LastResetAt = time.Now()
	
	if req.ForceOrderDrop {
		// Force the engine into a post-warm-up demand crash so the UI can
		// move from healthy -> anomalous -> disrupted when external signals land.
		state.BaselineOrders = 100.0
		state.RecentOrders = []time.Time{}
		state.TotalOrdersEver = 32
		state.LastResetAt = time.Now()
	} else {
		// Normal volume reset
		state.BaselineOrders = 20.0
		state.RecentOrders = []time.Time{}
		state.TotalOrdersEver = 0
		state.LastResetAt = time.Now()
	}
	
	if req.ExternalSignal != "" {
		state.ActiveSignals[req.ExternalSignal] = true
	} else {
		// Clear signals when passing empty string to reset
		state.ActiveSignals = make(map[string]bool)
	}
	state.mu.Unlock()

	evaluateDisruption(req.ZoneID, state)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message":          "demo_disruption_evaluated",
			"zone_id":          req.ZoneID,
			"force_order_drop": req.ForceOrderDrop,
			"external_signal":  req.ExternalSignal,
		},
		"meta": gin.H{"timestamp": time.Now().UTC().Format(time.RFC3339)},
	})
}

// ExternalSignalWebhook handles incoming third-party signals like weather alerts
func ExternalSignalWebhook(c *gin.Context) {
	var req struct {
		ZoneID uint   `json:"zone_id"`
		Source string `json:"source"`
		Status string `json:"status"` // "active" or "resolved"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	isActive := strings.ToLower(req.Status) == "active"
	
	if req.Source == "all_signals" {
	    // Clear all signals. True/false doesn't matter, we just clear everything.
		state := getOrCreateZoneState(req.ZoneID)
		state.mu.Lock()
		state.ActiveSignals = make(map[string]bool)
		state.mu.Unlock()
		evaluateDisruption(req.ZoneID, state)
	} else {
	    SetExternalSignal(req.ZoneID, req.Source, isActive)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "external_signal_received",
			"zone_id": req.ZoneID,
			"source":  req.Source,
			"active":  isActive,
		},
		"meta": gin.H{"timestamp": time.Now().UTC().Format(time.RFC3339)},
	})
}
