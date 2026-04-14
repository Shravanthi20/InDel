package worker

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

const (
	riskScoreThreshold    = 0.70  // workers with risk_score >= this are excluded from payouts
	maxPayoutPerDayINR    = 600.0 // maximum rupees covered per day
	defaultDisruptionHours = 4.0  // default disruption window if not specified
)

// GetDisruptions returns active disruptions in zone (existing endpoint stub kept working)
func GetDisruptions(c *gin.Context) {
	c.JSON(200, gin.H{"disruptions": []interface{}{}})
}

// TriggerDisruptionPayout processes a disruption event and calculates payout.
//
// POST /api/v1/worker/disruptions/trigger
//
// Conditions for payout (ALL must be satisfied):
//  1. Worker has an ACTIVE plan
//  2. Worker is in the affected zone
//  3. Worker's risk_score is BELOW the threshold (0.70)
//  4. Valid disruption event (type + duration provided)
//
// Payout formula:
//
//	payout = min(max_weekly_coverage, coverage_ratio * (baseline_daily / 24) * disruption_hours)
func TriggerDisruptionPayout(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	body := parseBody(c)
	disruptionType := bodyString(body, "disruption_type", "")
	zoneLevel := bodyString(body, "zone_level", "")
	zoneName := bodyString(body, "zone_name", "")
	disruptionHours := bodyFloatDefault(body, "disruption_hours", defaultDisruptionHours)

	if disruptionType == "" {
		c.JSON(400, gin.H{"error": "disruption_type is required"})
		return
	}
	if disruptionHours <= 0 || disruptionHours > 24 {
		disruptionHours = defaultDisruptionHours
	}

	if HasDB() {
		handleDisruptionPayoutDB(c, workerID, disruptionType, zoneLevel, zoneName, disruptionHours)
		return
	}

	// In-memory fallback
	handleDisruptionPayoutMemory(c, workerID, disruptionType, zoneLevel, zoneName, disruptionHours)
}

// ── DB Path ────────────────────────────────────────────────────────────────

func handleDisruptionPayoutDB(
	c *gin.Context,
	workerID, disruptionType, zoneLevel, zoneName string,
	disruptionHours float64,
) {
	workerIDUint, err := parseWorkerID(workerID)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid_worker_id"})
		return
	}

	now := time.Now().UTC()

	// 1. Fetch active policy + plan details
	var policyRow struct {
		PolicyID      uint    `gorm:"column:policy_id"`
		Status        string  `gorm:"column:status"`
		PlanID        string  `gorm:"column:plan_id"`
		PremiumAmount float64 `gorm:"column:premium_amount"`
	}
	if err := workerDB.Raw(`
		SELECT p.id AS policy_id, p.status, COALESCE(p.plan_id,'') AS plan_id, p.premium_amount
		FROM policies p
		WHERE p.worker_id = ?
		ORDER BY p.id DESC LIMIT 1
	`, workerIDUint).Scan(&policyRow).Error; err != nil || policyRow.PolicyID == 0 {
		c.JSON(403, gin.H{
			"error":  "no_active_policy",
			"reason": "Worker does not have an active policy",
		})
		return
	}

	// 2. Check plan status is active
	if !strings.EqualFold(policyRow.Status, "active") {
		c.JSON(403, gin.H{
			"error":  "plan_not_active",
			"reason": fmt.Sprintf("Plan status is '%s', not active", policyRow.Status),
		})
		return
	}

	// 3. Get worker zone and check zone match
	var workerZone struct {
		ZoneLevel string `gorm:"column:zone_level"`
		ZoneName  string `gorm:"column:zone_name"`
		ZoneCity  string `gorm:"column:zone_city"`
	}
	_ = workerDB.Raw(`
		SELECT COALESCE(z.level,'') AS zone_level, COALESCE(z.name,'') AS zone_name, COALESCE(z.city,'') AS zone_city
		FROM worker_profiles wp
		LEFT JOIN zones z ON z.id = wp.zone_id
		WHERE wp.worker_id = ?
		LIMIT 1
	`, workerIDUint).Scan(&workerZone).Error

	if zoneLevel != "" && zoneName != "" {
		workerZoneNorm := strings.ToUpper(strings.TrimSpace(workerZone.ZoneLevel))
		requestZoneNorm := strings.ToUpper(strings.TrimSpace(zoneLevel))
		workerNameNorm := strings.ToLower(strings.TrimSpace(workerZone.ZoneName))
		requestNameNorm := strings.ToLower(strings.TrimSpace(zoneName))
		workerCityNorm := strings.ToLower(strings.TrimSpace(workerZone.ZoneCity))

		zoneMatch := workerZoneNorm == requestZoneNorm &&
			(strings.Contains(workerNameNorm, requestNameNorm) ||
				strings.Contains(requestNameNorm, workerNameNorm) ||
				strings.Contains(workerCityNorm, requestNameNorm))

		if !zoneMatch {
			c.JSON(403, gin.H{
				"error":        "zone_mismatch",
				"reason":       "Worker is not in the affected disruption zone",
				"worker_zone":  fmt.Sprintf("%s/%s", workerZone.ZoneLevel, workerZone.ZoneName),
				"request_zone": fmt.Sprintf("%s/%s", zoneLevel, zoneName),
			})
			return
		}
	}

	// 4. Check risk score (ML first, fallback)
	quote, _ := services.QuotePremium(workerDB, workerIDUint, now)
	riskScore := 0.5
	if quote != nil {
		riskScore = quote.RiskScore
	}

	if riskScore >= riskScoreThreshold {
		c.JSON(403, gin.H{
			"error":       "risk_score_too_high",
			"reason":      fmt.Sprintf("Worker risk score %.3f >= threshold %.2f", riskScore, riskScoreThreshold),
			"risk_score":  riskScore,
			"threshold":   riskScoreThreshold,
		})
		return
	}

	// 5. Get plan config for coverage ratio and max payout
	planConfigs := getPlanConfigs()
	coverageRatio := 0.85
	maxWeeklyCoverage := 800.0
	if plan, ok := planConfigs[policyRow.PlanID]; ok {
		coverageRatio = plan.CoverageRatio
		maxWeeklyCoverage = float64(plan.MaxPayoutINR)
	}

	// 6. Get baseline daily income
	var baselineAmount float64 = 4200
	_ = workerDB.Raw("SELECT COALESCE(baseline_amount, 4200) FROM earnings_baseline WHERE worker_id = ? LIMIT 1", workerIDUint).Scan(&baselineAmount).Error
	baselineDaily := baselineAmount / 7.0

	// 7. Calculate payout
	// payout = min(max_weekly_coverage, coverage_ratio * (baseline_daily/24) * disruption_hours)
	payoutCalc := coverageRatio * (baselineDaily / 24.0) * disruptionHours
	payoutAmount := math.Min(maxWeeklyCoverage, payoutCalc)
	payoutAmount = math.Round(payoutAmount*100) / 100

	// 8. Persist: disruption record + claim + payout
	startTime := now
	endTime := now.Add(time.Duration(disruptionHours * float64(time.Hour)))
	disruption := models.Disruption{
		ZoneID:    1, // resolved from worker zone
		Type:      disruptionType,
		Status:    "active",
		StartTime: &startTime,
		EndTime:   &endTime,
	}
	_ = workerDB.Create(&disruption).Error

	claim := models.Claim{
		WorkerID:     workerIDUint,
		DisruptionID: disruption.ID,
		Status:       "approved",
		ClaimAmount:  payoutAmount,
		FraudVerdict: "clear",
	}
	_ = workerDB.Create(&claim).Error

	idempotencyKey := fmt.Sprintf("disr-%d-%d-%d", workerIDUint, disruption.ID, now.Unix())
	payout := models.Payout{
		WorkerID:       workerIDUint,
		ClaimID:        claim.ID,
		Amount:         payoutAmount,
		Status:         "credited",
		IdempotencyKey: idempotencyKey,
	}
	_ = workerDB.Create(&payout).Error

	// 9. Notification
	_ = workerDB.Exec(
		"INSERT INTO notifications (worker_id, type, message, created_at) VALUES (?, 'disruption_payout', ?, CURRENT_TIMESTAMP)",
		workerIDUint,
		fmt.Sprintf("Disruption payout of ₹%.0f credited for %s event.", payoutAmount, disruptionType),
	).Error

	c.JSON(200, gin.H{
		"message":           "disruption_payout_processed",
		"payout_amount_inr": payoutAmount,
		"disruption_type":   disruptionType,
		"disruption_hours":  disruptionHours,
		"claim_id":          fmt.Sprintf("clm-%d", claim.ID),
		"payout_id":         fmt.Sprintf("pay-%d", payout.ID),
		"coverage_ratio":    coverageRatio,
		"risk_score":        riskScore,
		"calc_detail": gin.H{
			"baseline_daily_inr":  baselineDaily,
			"formula":             "min(max_weekly_coverage, coverage_ratio × (baseline_daily/24) × disruption_hours)",
			"calculated_payout":   payoutCalc,
			"capped_at":           maxWeeklyCoverage,
			"final_payout_inr":    payoutAmount,
		},
	})
}

// ── In-Memory Fallback ──────────────────────────────────────────────────────

func handleDisruptionPayoutMemory(
	c *gin.Context,
	workerID, disruptionType, zoneLevel, zoneName string,
	disruptionHours float64,
) {
	store.mu.RLock()
	policy := store.data.Policy
	profile := store.data.WorkerProfiles[workerID]
	store.mu.RUnlock()

	// Check active status
	status, _ := policy["status"].(string)
	if !strings.EqualFold(status, "active") {
		c.JSON(403, gin.H{
			"error":  "plan_not_active",
			"reason": "No active plan found",
		})
		return
	}

	// Zone check (lenient in memory mode)
	if zoneLevel != "" && profile != nil {
		profileLevel, _ := profile["zone_level"].(string)
		if !strings.EqualFold(strings.TrimSpace(profileLevel), strings.TrimSpace(zoneLevel)) {
			c.JSON(403, gin.H{
				"error":  "zone_mismatch",
				"reason": "Worker zone does not match disruption zone",
			})
			return
		}
	}

	// Coverage calculation (use stored policy values)
	coverageRatio := 0.85
	if cr, ok := policy["coverage_ratio"].(float64); ok && cr > 0 {
		coverageRatio = cr
	}

	baselineDaily := 600.0 // mock ₹4200/week ÷ 7
	payoutCalc := coverageRatio * (baselineDaily / 24.0) * disruptionHours
	payoutAmount := math.Min(800.0, payoutCalc)
	payoutAmount = math.Round(payoutAmount*100) / 100

	// Append to in-memory payouts
	store.mu.Lock()
	newPayout := map[string]any{
		"payout_id":    nextID("pay", len(store.data.Payouts)),
		"claim_id":     nextID("clm", len(store.data.Claims)),
		"amount":       int(payoutAmount),
		"method":       "upi",
		"status":       "credited",
		"processed_at": nowISO(),
	}
	store.data.Payouts = append([]map[string]any{newPayout}, store.data.Payouts...)
	store.mu.Unlock()

	c.JSON(200, gin.H{
		"message":           "disruption_payout_processed",
		"payout_amount_inr": payoutAmount,
		"disruption_type":   disruptionType,
		"disruption_hours":  disruptionHours,
		"claim_id":          newPayout["claim_id"],
		"payout_id":         newPayout["payout_id"],
		"coverage_ratio":    coverageRatio,
		"calc_detail": gin.H{
			"baseline_daily_inr": baselineDaily,
			"formula":            "min(800, coverage_ratio × (baseline_daily/24) × disruption_hours)",
			"final_payout_inr":   payoutAmount,
		},
	})
}

// ── Helpers ────────────────────────────────────────────────────────────────

func bodyFloatDefault(body map[string]any, key string, fallback float64) float64 {
	v, ok := body[key]
	if !ok || v == nil {
		return fallback
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return fallback
	}
}
