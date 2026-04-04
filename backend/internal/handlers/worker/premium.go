package worker

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetPremium returns current premium (calling ML service with fallback)
func GetPremium(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	store.mu.RLock()
	profile := store.data.WorkerProfiles[workerID]
	store.mu.RUnlock()
	if profile == nil {
		profile = getPremiumProfileFromDB(workerID)
	}

	// Try to get ML-based premium with fallback to defaults
	premium, explainability := getPremiumEstimate(workerID, profile)
	now := time.Now().UTC()
	paymentState := paymentScheduleState{
		PaymentStatus:      "Eligible",
		DaysSinceLastPay:   0,
		NextPaymentEnabled: true,
		CoverageStatus:     "Expired",
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			if state, err := getOrBootstrapPaymentSchedule(workerIDUint, now); err == nil {
				paymentState = state
			}
		}
	} else {
		store.mu.RLock()
		paymentState = paymentStateFromInMemoryPolicy(store.data.Policy, now)
		store.mu.RUnlock()
	}

	resp := gin.H{
		"weekly_premium_inr":      premium,
		"currency":                "INR",
		"shap_breakdown":          explainability,
		"payment_status":          paymentState.PaymentStatus,
		"days_since_last_payment": paymentState.DaysSinceLastPay,
		"next_payment_enabled":    paymentState.NextPaymentEnabled,
		"coverage_status":         paymentState.CoverageStatus,
	}
	if paymentState.LastPaymentRecorded != nil {
		resp["last_payment_timestamp"] = paymentState.LastPaymentRecorded.UTC().Format(time.RFC3339)
	}

	c.JSON(200, resp)
}

func getPremiumProfileFromDB(workerID string) map[string]interface{} {
	if !hasDB() {
		return nil
	}
	workerIDUint, parseErr := parseWorkerID(workerID)
	if parseErr != nil {
		return nil
	}

	type row struct {
		ZoneID      uint    `gorm:"column:zone_id"`
		ZoneName    string  `gorm:"column:zone_name"`
		ZoneIDName  string  `gorm:"column:zone_id_name"`
		City        string  `gorm:"column:city"`
		State       string  `gorm:"column:state"`
		VehicleType string  `gorm:"column:vehicle_type"`
		Earnings    float64 `gorm:"column:earnings"`
	}

	var r row
	err := workerDB.Raw(`
		SELECT
			wp.zone_id AS zone_id,
			z.name AS zone_name,
			z.name AS zone_id_name,
			z.city AS city,
			z.state AS state,
			wp.vehicle_type AS vehicle_type,
			wp.total_earnings_lifetime AS earnings
		FROM worker_profiles wp
		LEFT JOIN zones z ON z.id = wp.zone_id
		WHERE wp.worker_id = ?
		LIMIT 1
	`, workerIDUint).Scan(&r).Error
	if err != nil {
		return nil
	}

	profile := map[string]interface{}{
		"zone_id":            firstNonEmpty(r.ZoneIDName, fmt.Sprintf("zone_%d", r.ZoneID)),
		"zone_level":         "A",
		"zone_name":          r.ZoneName,
		"city":               r.City,
		"state":              r.State,
		"vehicle_type":       r.VehicleType,
		"avg_daily_earnings": r.Earnings,
	}

	enrichPremiumProfileWithZoneGeo(profile, r.ZoneID)

	if profile["city"] == "" {
		profile["city"] = "Chennai"
	}
	if profile["state"] == "" {
		profile["state"] = "Tamil Nadu"
	}
	if profile["vehicle_type"] == "" {
		profile["vehicle_type"] = "two_wheeler"
	}

	return profile
}

func enrichPremiumProfileWithZoneGeo(profile map[string]interface{}, zoneID uint) {
	if !hasDB() || zoneID == 0 {
		return
	}

	type geoAgg struct {
		AvgFromLat float64 `gorm:"column:avg_from_lat"`
		AvgFromLon float64 `gorm:"column:avg_from_lon"`
		AvgToLat   float64 `gorm:"column:avg_to_lat"`
		AvgToLon   float64 `gorm:"column:avg_to_lon"`
		CountA     int64   `gorm:"column:count_a"`
		CountB     int64   `gorm:"column:count_b"`
		CountC     int64   `gorm:"column:count_c"`
	}

	var agg geoAgg
	_ = workerDB.Raw(`
		SELECT
			COALESCE(AVG(CASE WHEN o.from_lat <> 0 THEN o.from_lat END), 0) AS avg_from_lat,
			COALESCE(AVG(CASE WHEN o.from_lon <> 0 THEN o.from_lon END), 0) AS avg_from_lon,
			COALESCE(AVG(CASE WHEN o.to_lat <> 0 THEN o.to_lat END), 0) AS avg_to_lat,
			COALESCE(AVG(CASE WHEN o.to_lon <> 0 THEN o.to_lon END), 0) AS avg_to_lon,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(o.from_city, ''))) = LOWER(TRIM(COALESCE(o.to_city, ''))) THEN 1 ELSE 0 END) AS count_a,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(o.from_city, ''))) <> LOWER(TRIM(COALESCE(o.to_city, '')))
				AND LOWER(TRIM(COALESCE(o.from_state, ''))) = LOWER(TRIM(COALESCE(o.to_state, ''))) THEN 1 ELSE 0 END) AS count_b,
			SUM(CASE WHEN LOWER(TRIM(COALESCE(o.from_state, ''))) <> LOWER(TRIM(COALESCE(o.to_state, ''))) THEN 1 ELSE 0 END) AS count_c
		FROM orders o
		WHERE o.zone_id = ?
	`, zoneID).Scan(&agg).Error

	zoneLevel := "A"
	if agg.CountC > agg.CountB && agg.CountC > agg.CountA {
		zoneLevel = "C"
	} else if agg.CountB > agg.CountA {
		zoneLevel = "B"
	}

	zoneLat := agg.AvgFromLat
	zoneLon := agg.AvgFromLon
	if strings.EqualFold(zoneLevel, "B") || strings.EqualFold(zoneLevel, "C") {
		if agg.AvgFromLat != 0 && agg.AvgToLat != 0 {
			zoneLat = (agg.AvgFromLat + agg.AvgToLat) / 2
		} else if agg.AvgToLat != 0 {
			zoneLat = agg.AvgToLat
		}
		if agg.AvgFromLon != 0 && agg.AvgToLon != 0 {
			zoneLon = (agg.AvgFromLon + agg.AvgToLon) / 2
		} else if agg.AvgToLon != 0 {
			zoneLon = agg.AvgToLon
		}
	}

	profile["zone_level"] = zoneLevel
	profile["from_lat"] = agg.AvgFromLat
	profile["from_lon"] = agg.AvgFromLon
	profile["to_lat"] = agg.AvgToLat
	profile["to_lon"] = agg.AvgToLon
	if zoneLat != 0 {
		profile["zone_lat"] = zoneLat
	}
	if zoneLon != 0 {
		profile["zone_lon"] = zoneLon
	}
}

// getPremiumEstimate tries to get ML-based premium, falls back to defaults
func getPremiumEstimate(workerID string, profile map[string]interface{}) (int, []gin.H) {
	// Default explainability
	defaultExplainability := []gin.H{
		{"feature": "rain_risk", "impact": 0.42},
		{"feature": "order_drop_volatility", "impact": 0.31},
		{"feature": "historical_disruptions", "impact": 0.27},
	}

	if profile == nil {
		store.mu.RLock()
		defaultPremium, _ := store.data.Policy["weekly_premium_inr"].(int)
		store.mu.RUnlock()
		if defaultPremium == 0 {
			defaultPremium = 22
		}
		return defaultPremium, defaultExplainability
	}

	// Build ML request from profile
	mlReq := buildMLPremiumRequest(workerID, profile)

	// Call ML service
	mlResult, err := getPremiumFromML(mlReq)
	if err != nil {
		log.Printf("[Premium] ML service unavailable, using default: %v", err)
		store.mu.RLock()
		defaultPremium, _ := store.data.Policy["weekly_premium_inr"].(int)
		store.mu.RUnlock()
		if defaultPremium == 0 {
			defaultPremium = 22
		}
		return defaultPremium, defaultExplainability
	}

	// Convert SHAP explainability to gin.H format
	explainability := make([]gin.H, len(mlResult.Explainability))
	for i, factor := range mlResult.Explainability {
		explainability[i] = gin.H{
			"feature": factor.Feature,
			"impact":  factor.Impact,
		}
	}

	// Log ML result for monitoring
	log.Printf("[Premium] ML service returned premium: INR %.2f, risk: %.3f for worker %s",
		mlResult.PremiumInr, mlResult.RiskScore, workerID)

	premium := int(math.Round(mlResult.PremiumInr))
	if premium < 20 {
		premium = 20
	}
	if premium > 50 {
		premium = 50
	}

	return premium, explainability
}

// PayPremium makes premium payment
func PayPremium(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	now := time.Now().UTC()

	store.mu.RLock()
	defaultAmount, _ := store.data.Policy["weekly_premium_inr"].(int)
	store.mu.RUnlock()

	amount := bodyInt(body, "amount", defaultAmount)
	if amount <= 0 {
		amount = defaultAmount
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			state, err := getOrBootstrapPaymentSchedule(workerIDUint, now)
			if err == nil && state.PaymentStatus == "Locked" {
				c.JSON(409, gin.H{
					"error":                   "payment_locked",
					"message":                 paymentLockError(state),
					"payment_status":          state.PaymentStatus,
					"days_since_last_payment": state.DaysSinceLastPay,
					"next_payment_enabled":    state.NextPaymentEnabled,
					"coverage_status":         state.CoverageStatus,
				})
				return
			}

			_ = workerDB.Exec(
				"INSERT INTO premium_payments (worker_id, policy_id, amount, status, payment_date) VALUES (?, (SELECT id FROM policies WHERE worker_id = ? ORDER BY id DESC LIMIT 1), ?, 'completed', CURRENT_TIMESTAMP)",
				workerIDUint, workerIDUint, amount,
			).Error
			_ = upsertPaymentSchedule(workerIDUint, now, false, "Active")
			_ = workerDB.Exec("UPDATE policies SET status='active', updated_at=CURRENT_TIMESTAMP WHERE worker_id = ?", workerIDUint).Error
			c.JSON(200, gin.H{
				"message":                 "payment_successful",
				"amount":                  amount,
				"currency":                "INR",
				"payment_id":              fmt.Sprintf("db-payment-%d", workerIDUint),
				"payment_status":          "Locked",
				"days_since_last_payment": 0,
				"next_payment_enabled":    false,
				"coverage_status":         "Active",
				"last_payment_timestamp":  now.Format(time.RFC3339),
			})
			return
		}
	}

	store.mu.Lock()
	state := paymentStateFromInMemoryPolicy(store.data.Policy, now)
	if state.PaymentStatus == "Locked" {
		store.mu.Unlock()
		c.JSON(409, gin.H{
			"error":                   "payment_locked",
			"message":                 paymentLockError(state),
			"payment_status":          state.PaymentStatus,
			"days_since_last_payment": state.DaysSinceLastPay,
			"next_payment_enabled":    state.NextPaymentEnabled,
			"coverage_status":         state.CoverageStatus,
		})
		return
	}
	store.data.Policy["last_payment_timestamp"] = now.Format(time.RFC3339)
	store.data.Policy["next_payment_enabled"] = false
	store.data.Policy["coverage_status"] = "Active"
	store.data.Policy["payment_status"] = "Locked"
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		profile["coverage_status"] = "active"
	}
	store.mu.Unlock()

	c.JSON(200, gin.H{
		"message":                 "payment_successful",
		"amount":                  amount,
		"currency":                "INR",
		"payment_id":              "mock-payment-001",
		"payment_status":          "Locked",
		"days_since_last_payment": 0,
		"next_payment_enabled":    false,
		"coverage_status":         "Active",
		"last_payment_timestamp":  now.Format(time.RFC3339),
	})
}
