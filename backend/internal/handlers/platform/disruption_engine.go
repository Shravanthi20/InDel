package platform

import (
	"fmt"
	"log"
	"math"
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

type ProgressivePayoutState struct {
	mu            sync.Mutex
	LastRiskScore float64
	LastPayout    float64
	LastUpdatedAt time.Time
}

type ProgressivePayoutInputs struct {
	AQI           float64
	Temperature   float64
	Rain          float64
	Traffic       float64
	MaxPayoutDay  float64
	CoverageRatio float64
}

type ProgressivePayoutResult struct {
	CurrentRiskScore  float64
	IncrementalRisk   float64
	NewPayout         float64
	FinalPayout       float64
	RemainingCoverage float64
	TotalPayoutSoFar  float64
	TriggerStatus     string
	AQIScore          float64
	TempScore         float64
	RainScore         float64
	TrafficScore      float64
}

// Global engine state
var (
	zoneSignals      = make(map[uint]*ZoneSignalState)
	zonePayoutStates = make(map[uint]*ProgressivePayoutState)
	engineMu         sync.Mutex

	// Idempotency cache
	processedOrderIds = make(map[string]time.Time)
	idCacheMu         sync.Mutex
)

// ResetEngineForTests allows the testing suite to natively flush caches before independent scenarios
func ResetEngineForTests() {
	engineMu.Lock()
	zoneSignals = make(map[uint]*ZoneSignalState)
	zonePayoutStates = make(map[uint]*ProgressivePayoutState)
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

func getOrCreateProgressivePayoutState(zoneID uint) *ProgressivePayoutState {
	engineMu.Lock()
	defer engineMu.Unlock()
	if state, exists := zonePayoutStates[zoneID]; exists {
		return state
	}
	newState := &ProgressivePayoutState{}
	zonePayoutStates[zoneID] = newState
	return newState
}

func resolveFloatInput(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

func clampRiskScore(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func calculateProgressiveRisk(inputs ProgressivePayoutInputs) ProgressivePayoutResult {
	aqiScore := clampRiskScore((inputs.AQI - 200) / 100)
	tempScore := clampRiskScore((inputs.Temperature - 35) / 10)
	rainScore := clampRiskScore(inputs.Rain / 15)
	trafficScore := clampRiskScore((inputs.Traffic - 60) / 25)

	riskScore := (0.30 * rainScore) + (0.25 * tempScore) + (0.25 * aqiScore) + (0.20 * trafficScore)
	if riskScore < 0 {
		riskScore = 0
	}
	if riskScore > 1 {
		riskScore = 1
	}

	maxCoverage := inputs.MaxPayoutDay * clampRiskScore(inputs.CoverageRatio)
	if maxCoverage < 0 {
		maxCoverage = 0
	}

	return ProgressivePayoutResult{
		CurrentRiskScore:  riskScore,
		RemainingCoverage: maxCoverage,
		AQIScore:          aqiScore,
		TempScore:         tempScore,
		RainScore:         rainScore,
		TrafficScore:      trafficScore,
	}
}

func applyProgressivePayout(zoneID uint, inputs ProgressivePayoutInputs) ProgressivePayoutResult {
	state := getOrCreateProgressivePayoutState(zoneID)
	state.mu.Lock()
	defer state.mu.Unlock()

	result := calculateProgressiveRisk(inputs)
	lastRisk := state.LastRiskScore
	lastPayout := state.LastPayout
	maxCoverage := result.RemainingCoverage
	remainingCoverage := maxCoverage - lastPayout
	if remainingCoverage < 0 {
		remainingCoverage = 0
	}

	result.RemainingCoverage = remainingCoverage
	result.TotalPayoutSoFar = lastPayout

	if result.CurrentRiskScore <= lastRisk {
		result.TriggerStatus = "No payout"
		state.LastUpdatedAt = time.Now().UTC()
		return result
	}

	result.IncrementalRisk = result.CurrentRiskScore - lastRisk
	result.NewPayout = maxCoverage * result.IncrementalRisk
	result.FinalPayout = math.Min(result.NewPayout, remainingCoverage)

	if remainingCoverage <= 0 || result.FinalPayout <= 0 {
		result.TriggerStatus = "Max coverage reached"
		state.LastRiskScore = result.CurrentRiskScore
		state.LastUpdatedAt = time.Now().UTC()
		return result
	}

	if result.FinalPayout < result.NewPayout {
		result.TriggerStatus = "Max coverage reached"
	} else {
		result.TriggerStatus = "Incremental payout"
	}

	state.LastRiskScore = result.CurrentRiskScore
	state.LastPayout = math.Round((lastPayout+result.FinalPayout)*100) / 100
	state.LastUpdatedAt = time.Now().UTC()
	result.TotalPayoutSoFar = state.LastPayout
	result.RemainingCoverage = math.Round((maxCoverage-state.LastPayout)*100) / 100
	if result.RemainingCoverage < 0 {
		result.RemainingCoverage = 0
	}

	return result
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
	// System remains "Healthy" until it has seen at least 11 orders total (Master-class logic)
	if state.TotalOrdersEver <= 10 {
		return len(state.RecentOrders), state.BaselineOrders, 0.0, "healthy"
	}

	// 2. Window Filtering (20s Real-time Window for UI)
	var currentWindow []time.Time
	for _, t := range state.RecentOrders {
		if now.Sub(t) <= 20*time.Second {
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
	if orderDrop < 0 {
		orderDrop = 0
	}
	if orderDrop > 1 {
		orderDrop = 1
	}

	// Print to logs so we can see the exact math!
	log.Printf("[DECISION ENGINE] Zone %d | Output: %s | Window: %d / %.0f (Drop: %.1f%%)", zoneID, strings.ToUpper(status), current, state.BaselineOrders, orderDrop*100)

	// 3. Multi-Signal Validation
	hasExternalSignals := len(state.ActiveSignals) > 0

	if orderDrop > 0.30 && hasExternalSignals {
		createDisruptionRecord(zoneID, orderDrop, state.ActiveSignals, false)
	}
}

func createDisruptionRecord(zoneID uint, orderDrop float64, signals map[string]bool, forceInsert bool) {
	if !hasDB() {
		return
	}

	if !forceInsert {
		// DUPLICATE PREVENTION: Check if a disruption exists in the last 10 minutes
		var existing int64
		tenMinsAgo := time.Now().Add(-10 * time.Minute)
		platformDB.Model(&models.Disruption{}).Where("zone_id = ? AND created_at > ?", zoneID, tenMinsAgo).Count(&existing)
		if existing > 0 {
			return // Still active
		}
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
		ZoneID:     zoneID,
		Type:       triggerStr,
		Severity:   severity,
		Confidence: confidence, // Use calculated confidence here

		Status:          "confirmed",
		StartTime:       &now,
		SignalTimestamp: &now, // THIS was missing, crashing the Postgres INSERT!
		ConfirmedAt:     &now,
	}

	if err := platformDB.Create(&disruption).Error; err != nil {
		log.Printf("Failed to create disruption: %v", err)
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

	var records []models.Disruption
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	if err := platformDB.Where("created_at > ?", oneHourAgo).Order("created_at desc").Limit(20).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "disruptions_fetch_failed"})
		return
	}

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
		ZoneID          uint     `json:"zone_id"`
		ForceOrderDrop  bool     `json:"force_order_drop"`
		ExternalSignal  string   `json:"external_signal"` // e.g., "weather", "aqi", "system"
		GenerateClaims  bool     `json:"generate_claims"`
		AQI             *float64 `json:"aqi"`
		Rain            *float64 `json:"rain"`
		Traffic         *float64 `json:"traffic"`
		Temperature     *float64 `json:"temperature"`
		MaxPayoutPerDay *float64 `json:"max_payout_per_day"`
		MaxPayoutINR    *float64 `json:"max_payout_inr"`
		CoverageRatio   *float64 `json:"coverage_ratio"`
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
		// Mock a 50% order drop
		state.BaselineOrders = 100.0
		state.TotalOrdersEver = 11 // bypass warm-up guard for explicit demo trigger
		state.RecentOrders = []time.Time{}
		state.LastResetAt = time.Now()
	} else {
		// Normal volume reset
		state.BaselineOrders = 20.0
		state.RecentOrders = []time.Time{}
		state.TotalOrdersEver = 11 // keep trigger evaluable for manual demo signals
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

	// Claims and notifications require an existing disruption record.
	// Explicit manual demo triggers should always create a fresh record for reliable DB write/fetch behavior.
	forcedDrop := 0.5
	if req.ForceOrderDrop {
		forcedDrop = 0.8
	}
	signals := map[string]bool{}
	if req.ExternalSignal != "" {
		signals[req.ExternalSignal] = true
	} else {
		signals["demo_manual_trigger"] = true
	}
	createDisruptionRecord(req.ZoneID, forcedDrop, signals, true)
	disruptionID := latestDisruptionIDForZone(req.ZoneID)

	maxPayoutPerDay := resolveFloatInput(req.MaxPayoutPerDay, resolveFloatInput(req.MaxPayoutINR, 2000))
	coverageRatio := resolveFloatInput(req.CoverageRatio, 0.8)
	if coverageRatio < 0 {
		coverageRatio = 0
	}
	if coverageRatio > 1 {
		coverageRatio = 1
	}

	payoutResult := applyProgressivePayout(req.ZoneID, ProgressivePayoutInputs{
		AQI:           resolveFloatInput(req.AQI, 190),
		Temperature:   resolveFloatInput(req.Temperature, 33),
		Rain:          resolveFloatInput(req.Rain, 3),
		Traffic:       resolveFloatInput(req.Traffic, 58),
		MaxPayoutDay:  maxPayoutPerDay,
		CoverageRatio: coverageRatio,
	})

	notificationsCreated := 0
	if req.ZoneID != 0 {
		notificationsCreated = createDisruptionNotificationsForZone(req.ZoneID, payoutResult.CurrentRiskScore, payoutResult.TotalPayoutSoFar, payoutResult.TriggerStatus)
	}

	claimsGenerated := 0
	if req.GenerateClaims && payoutResult.FinalPayout > 0 {
		if disruptionID != 0 {
			claimsGenerated = createClaimsForZoneDisruption(disruptionID, req.ZoneID, payoutResult.FinalPayout)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message":               "Demo disruption triggered evaluated",
			"disruption_sent":       true,
			"sent_at":               time.Now().UTC().Format(time.RFC3339),
			"zone_id":               req.ZoneID,
			"claims_generated":      claimsGenerated,
			"notifications_created": notificationsCreated,
			"current_risk_score":    payoutResult.CurrentRiskScore,
			"incremental_risk":      payoutResult.IncrementalRisk,
			"new_payout":            payoutResult.NewPayout,
			"final_payout":          payoutResult.FinalPayout,
			"total_payout_so_far":   payoutResult.TotalPayoutSoFar,
			"remaining_coverage":    payoutResult.RemainingCoverage,
			"trigger_status":        payoutResult.TriggerStatus,
		},
		"meta": gin.H{"timestamp": time.Now().UTC().Format(time.RFC3339)},
	})
}

func latestDisruptionIDForZone(zoneID uint) uint {
	if !hasDB() || zoneID == 0 {
		return 0
	}
	var disruption models.Disruption
	err := platformDB.Where("zone_id = ?", zoneID).Order("id desc").First(&disruption).Error
	if err != nil {
		return 0
	}
	return disruption.ID
}

func tempRiskMultiplier(temp float64) float64 {
	switch {
	case temp >= 40:
		return 1.0
	case temp >= 35:
		return 0.7
	case temp >= 30:
		return 0.4
	default:
		return 0.2
	}
}

func createClaimsAndNotificationsForZoneDisruption(disruptionID, zoneID uint, payoutAmount, currentRiskScore, totalPayoutSoFar float64, triggerStatus string) (int, int) {
	if !hasDB() || disruptionID == 0 || zoneID == 0 {
		return 0, 0
	}

	if payoutAmount <= 0 {
		return 0, 0
	}

	type workerRow struct {
		WorkerID uint `gorm:"column:worker_id"`
	}
	rows := make([]workerRow, 0)
	err := platformDB.Raw(`
		SELECT DISTINCT wp.worker_id
		FROM worker_profiles wp
		WHERE wp.zone_id = ?
	`, zoneID).Scan(&rows).Error
	if err != nil {
		log.Printf("createClaimsAndNotificationsForZoneDisruption: failed worker lookup: %v", err)
		return 0, 0
	}

	eligibleWorkers := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.WorkerID == 0 {
			continue
		}

		var existing int64
		_ = platformDB.Model(&models.Claim{}).
			Where("disruption_id = ? AND worker_id = ?", disruptionID, row.WorkerID).
			Count(&existing).Error
		if existing > 0 {
			continue
		}

		eligibleWorkers = append(eligibleWorkers, row.WorkerID)
	}

	if len(eligibleWorkers) == 0 {
		return 0, 0
	}

	baseShare := math.Round((payoutAmount/float64(len(eligibleWorkers)))*100) / 100
	claimsGenerated := 0
	allocated := 0.0
	for index, workerID := range eligibleWorkers {
		share := baseShare
		if index == len(eligibleWorkers)-1 {
			share = math.Round((payoutAmount-allocated)*100) / 100
		}
		if share <= 0 {
			continue
		}
		allocated += share

		claim := models.Claim{
			DisruptionID: disruptionID,
			WorkerID:     workerID,
			ClaimAmount:  share,
			Status:       "pending",
			FraudVerdict: "pending",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		if err := platformDB.Create(&claim).Error; err != nil {
			log.Printf("createClaimsAndNotificationsForZoneDisruption: failed claim create worker=%d err=%v", workerID, err)
			continue
		}

		if err := createClaimMadeNotification(workerID, claim.ID, share); err != nil {
			log.Printf("createClaimsAndNotificationsForZoneDisruption: failed notification worker=%d err=%v", workerID, err)
		}
		claimsGenerated++
	}

	return claimsGenerated, 0
}

func createDisruptionNotificationsForZone(zoneID uint, currentRiskScore, totalPayoutSoFar float64, triggerStatus string) int {
	if !hasDB() || zoneID == 0 {
		return 0
	}

	type workerRow struct {
		WorkerID uint `gorm:"column:worker_id"`
	}
	rows := make([]workerRow, 0)
	err := platformDB.Raw(`
		SELECT DISTINCT wp.worker_id
		FROM worker_profiles wp
		WHERE wp.zone_id = ?
	`, zoneID).Scan(&rows).Error
	if err != nil {
		log.Printf("createDisruptionNotificationsForZone: failed worker lookup: %v", err)
		return 0
	}

	created := 0
	for _, row := range rows {
		if row.WorkerID == 0 {
			continue
		}

		msg := fmt.Sprintf("Disruption detected in your zone. Risk %.2f. Total payout so far INR %.0f. Status: %s", currentRiskScore, totalPayoutSoFar, triggerStatus)
		if err := platformDB.Exec(
			"INSERT INTO notifications (worker_id, type, message, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)",
			row.WorkerID,
			"disruption_alert",
			msg,
		).Error; err != nil {
			log.Printf("createDisruptionNotificationsForZone: failed notification insert worker=%d err=%v", row.WorkerID, err)
			continue
		}
		created++
	}

	return created
}

func createClaimsForZoneDisruption(disruptionID, zoneID uint, payoutAmount float64) int {
	if !hasDB() || disruptionID == 0 || zoneID == 0 || payoutAmount <= 0 {
		return 0
	}

	type workerRow struct {
		WorkerID uint `gorm:"column:worker_id"`
	}
	rows := make([]workerRow, 0)
	err := platformDB.Raw(`
		SELECT DISTINCT wp.worker_id
		FROM worker_profiles wp
		WHERE wp.zone_id = ?
	`, zoneID).Scan(&rows).Error
	if err != nil {
		log.Printf("createClaimsForZoneDisruption: failed worker lookup: %v", err)
		return 0
	}

	eligibleWorkers := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.WorkerID == 0 {
			continue
		}

		var existing int64
		_ = platformDB.Model(&models.Claim{}).
			Where("disruption_id = ? AND worker_id = ?", disruptionID, row.WorkerID).
			Count(&existing).Error
		if existing > 0 {
			continue
		}

		eligibleWorkers = append(eligibleWorkers, row.WorkerID)
	}

	if len(eligibleWorkers) == 0 {
		return 0
	}

	baseShare := math.Round((payoutAmount/float64(len(eligibleWorkers)))*100) / 100
	claimsGenerated := 0
	allocated := 0.0
	for index, workerID := range eligibleWorkers {
		share := baseShare
		if index == len(eligibleWorkers)-1 {
			share = math.Round((payoutAmount-allocated)*100) / 100
		}
		if share <= 0 {
			continue
		}
		allocated += share

		claim := models.Claim{
			DisruptionID: disruptionID,
			WorkerID:     workerID,
			ClaimAmount:  share,
			Status:       "pending",
			FraudVerdict: "pending",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		if err := platformDB.Create(&claim).Error; err != nil {
			log.Printf("createClaimsForZoneDisruption: failed claim create worker=%d err=%v", workerID, err)
			continue
		}

		if err := createClaimMadeNotification(workerID, claim.ID, share); err != nil {
			log.Printf("createClaimsForZoneDisruption: failed notification worker=%d err=%v", workerID, err)
		}
		claimsGenerated++
	}

	return claimsGenerated
}

func createClaimMadeNotification(workerID, claimID uint, claimAmount float64) error {
	message := fmt.Sprintf("Claim is made for claim #%d. Amount INR %.0f. Time %s.", claimID, claimAmount, time.Now().UTC().Format(time.RFC3339))
	return platformDB.Exec(
		"INSERT INTO notifications (worker_id, type, message, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)",
		workerID,
		"claim_made",
		message,
	).Error
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
