package worker

import (
	"fmt"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type planConfig struct {
	PlanID        string
	PlanName      string
	RangeStart    int
	RangeEnd      int
	PremiumMinINR int
	PremiumMaxINR int
	CoverageRatio float64
	MaxPayoutINR  int
	Description   string
}

func planUpgradeFee(planConfigs map[string]planConfig, currentPlanID string, selectedPlanID string) int {
	currentPlanID = strings.TrimSpace(currentPlanID)
	selectedPlanID = strings.TrimSpace(selectedPlanID)
	if currentPlanID == "" || currentPlanID == selectedPlanID {
		return 0
	}

	currentPlan, currentExists := planConfigs[currentPlanID]
	selectedPlan, selectedExists := planConfigs[selectedPlanID]
	if !currentExists || !selectedExists {
		return 0
	}

	if selectedPlan.MaxPayoutINR > currentPlan.MaxPayoutINR {
		return 5
	}
	return 0
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func applyDynamicPlanPremiums(planConfigs map[string]planConfig, mlPremium int, zoneLevel string) map[string]planConfig {
	normalized := strings.ToUpper(strings.TrimSpace(zoneLevel))
	zoneOffset := 0
	switch normalized {
	case "A":
		mlPremium = clampInt(int(float64(mlPremium)*0.96), 20, 50)
		zoneOffset = -2
	case "B":
		mlPremium = clampInt(mlPremium, 20, 50)
	case "C":
		mlPremium = clampInt(int(float64(mlPremium)*1.08), 20, 50)
		zoneOffset = 4
	default: // B or unknown
		mlPremium = clampInt(mlPremium, 20, 50)
	}

	mlPremium = clampInt(mlPremium+zoneOffset, 20, 50)

	seedMid := clampInt(mlPremium-6, 20, 50)
	scaleMid := clampInt(mlPremium, 20, 50)
	soarMid := clampInt(mlPremium+6, 20, 50)

	seedMin := clampInt(seedMid-3, 20, 50)
	seedMax := clampInt(seedMid+3, seedMin, 50)

	scaleMin := clampInt(scaleMid-3, seedMax+1, 50)
	scaleMax := clampInt(scaleMid+3, scaleMin, 50)

	soarMin := clampInt(soarMid-3, scaleMax+1, 50)
	soarMax := 50
	if soarMin > soarMax {
		soarMin = soarMax
	}

	if p, ok := planConfigs["plan-starter"]; ok {
		p.PremiumMinINR = seedMin
		p.PremiumMaxINR = seedMax
		planConfigs[p.PlanID] = p
	}
	if p, ok := planConfigs["plan-growth"]; ok {
		p.PremiumMinINR = scaleMin
		p.PremiumMaxINR = scaleMax
		planConfigs[p.PlanID] = p
	}
	if p, ok := planConfigs["plan-premium"]; ok {
		p.PremiumMinINR = soarMin
		p.PremiumMaxINR = soarMax
		planConfigs[p.PlanID] = p
	}

	return planConfigs
}

func getPlanConfigs() map[string]planConfig {
	return map[string]planConfig{
		"plan-starter": {
			PlanID:        "plan-starter",
			PlanName:      "Seed",
			RangeStart:    10,
			RangeEnd:      15,
			PremiumMinINR: 12,
			PremiumMaxINR: 18,
			CoverageRatio: 0.80,
			MaxPayoutINR:  600,
			Description:   "Perfect for part-time delivery workers. Covers disruptions up to Rs.600/week.",
		},
		"plan-growth": {
			PlanID:        "plan-growth",
			PlanName:      "Scale",
			RangeStart:    15,
			RangeEnd:      20,
			PremiumMinINR: 19,
			PremiumMaxINR: 26,
			CoverageRatio: 0.85,
			MaxPayoutINR:  800,
			Description:   "For active delivery workers. Enhanced coverage up to Rs.800/week.",
		},
		"plan-premium": {
			PlanID:        "plan-premium",
			PlanName:      "Soar",
			RangeStart:    20,
			RangeEnd:      25,
			PremiumMinINR: 27,
			PremiumMaxINR: 35,
			CoverageRatio: 0.90,
			MaxPayoutINR:  1000,
			Description:   "For high-volume workers. Maximum protection up to Rs.1000/week.",
		},
	}
}

func premiumForRange(plan planConfig, expectedDeliveries int) int {
	if expectedDeliveries < plan.RangeStart {
		expectedDeliveries = plan.RangeStart
	}
	if expectedDeliveries > plan.RangeEnd {
		expectedDeliveries = plan.RangeEnd
	}

	span := plan.RangeEnd - plan.RangeStart
	if span <= 0 {
		return plan.PremiumMinINR
	}

	progress := expectedDeliveries - plan.RangeStart
	return plan.PremiumMinINR + ((plan.PremiumMaxINR-plan.PremiumMinINR)*progress)/span
}

func bodyBool(body map[string]any, key string, fallback bool) bool {
	v, ok := body[key]
	if !ok || v == nil {
		return fallback
	}
	b, ok := v.(bool)
	if !ok {
		return fallback
	}
	return b
}

func getPremiumProfileFromDB(workerID string) map[string]any {
	if !HasDB() {
		return nil
	}

	workerIDUint, err := parseWorkerID(workerID)
	if err != nil {
		return nil
	}

	type row struct {
		ZoneID    uint    `gorm:"column:zone_id"`
		ZoneLevel string  `gorm:"column:zone_level"`
		ZoneCity  string  `gorm:"column:zone_city"`
		ZoneState string  `gorm:"column:zone_state"`
		Vehicle   string  `gorm:"column:vehicle_type"`
		Baseline  float64 `gorm:"column:baseline_amount"`
	}

	var r row
	if qErr := workerDB.Raw(`
		SELECT
			COALESCE(wp.zone_id, 0) AS zone_id,
			COALESCE(z.level, '') AS zone_level,
			COALESCE(z.city, '') AS zone_city,
			COALESCE(z.state, '') AS zone_state,
			COALESCE(wp.vehicle_type, 'two_wheeler') AS vehicle_type,
			COALESCE(eb.baseline_amount, 4200) AS baseline_amount
		FROM worker_profiles wp
		LEFT JOIN zones z ON z.id = wp.zone_id
		LEFT JOIN earnings_baseline eb ON eb.worker_id = wp.worker_id
		WHERE wp.worker_id = ?
		LIMIT 1
	`, workerIDUint).Scan(&r).Error; qErr != nil {
		return nil
	}

	return map[string]any{
		"zone_id":         float64(r.ZoneID),
		"zone_level":      r.ZoneLevel,
		"city":            r.ZoneCity,
		"state":           r.ZoneState,
		"vehicle_type":    r.Vehicle,
		"baseline_amount": r.Baseline,
	}
}

func enrichPremiumProfileWithZoneGeo(profile map[string]any, zoneID uint) {
	if profile == nil || !HasDB() || zoneID == 0 {
		return
	}

	type zoneRow struct {
		City  string `gorm:"column:city"`
		State string `gorm:"column:state"`
		Level string `gorm:"column:level"`
	}

	var z zoneRow
	if err := workerDB.Raw("SELECT COALESCE(city,'') AS city, COALESCE(state,'') AS state, COALESCE(level,'') AS level FROM zones WHERE id = ? LIMIT 1", zoneID).Scan(&z).Error; err != nil {
		return
	}

	if strings.TrimSpace(z.City) != "" {
		profile["city"] = z.City
	}
	if strings.TrimSpace(z.State) != "" {
		profile["state"] = z.State
	}
	if strings.TrimSpace(z.Level) != "" {
		profile["zone_level"] = z.Level
	}
}

func getPremiumEstimate(workerID string, profile map[string]any) (int, string) {
	if !HasDB() {
		return 35, "fallback"
	}

	workerIDUint, err := parseWorkerID(workerID)
	if err != nil {
		return 35, "fallback"
	}

	quote, qErr := services.QuotePremium(workerDB, workerIDUint, time.Now().UTC())
	if qErr != nil || quote == nil {
		return 35, "fallback"
	}

	premium := int(quote.WeeklyPremiumINR)
	if premium <= 0 {
		premium = 35
	}
	return premium, quote.Source
}

// GetPlans returns available delivery plans based on order ranges
func GetPlans(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	planConfigs := getPlanConfigs()

	store.mu.RLock()
	profile := store.data.WorkerProfiles[workerID]
	store.mu.RUnlock()
	if profile == nil {
		profile = getPremiumProfileFromDB(workerID)
	}
	if profile != nil {
		zoneLevel := bodyString(profile, "zone_level", "")
		if zoneIDRaw, ok := profile["zone_id"]; ok {
			if zoneIDFloat, ok := zoneIDRaw.(float64); ok {
				enrichPremiumProfileWithZoneGeo(profile, uint(zoneIDFloat))
			}
		}
		mlPremium, _ := getPremiumEstimate(workerID, profile)
		planConfigs = applyDynamicPlanPremiums(planConfigs, mlPremium, zoneLevel)
	}

	plans := make([]gin.H, 0, len(planConfigs))
	for _, p := range planConfigs {
		plans = append(plans, gin.H{
			"plan_id":                p.PlanID,
			"plan_name":              p.PlanName,
			"range_start":            p.RangeStart,
			"range_end":              p.RangeEnd,
			"weekly_premium_min_inr": p.PremiumMinINR,
			"weekly_premium_max_inr": p.PremiumMaxINR,
			"weekly_premium_inr":     p.PremiumMinINR,
			"coverage_ratio":         p.CoverageRatio,
			"max_payout_inr":         p.MaxPayoutINR,
			"description":            p.Description,
		})
	}

	c.JSON(200, gin.H{"plans": plans})
}

// SelectPlan enrolls worker in a selected plan and creates policy
func SelectPlan(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	body := parseBody(c)
	planID := bodyString(body, "plan_id", "")
	expectedDeliveries := bodyInt(body, "expected_deliveries", 0)
	paymentAmount := bodyInt(body, "payment_amount_inr", 0)
	paymentConfirmed := bodyBool(body, "payment_confirmed", false)

	if planID == "" {
		c.JSON(400, gin.H{"error": "plan_id is required"})
		return
	}

	planConfigs := getPlanConfigs()
	plan, exists := planConfigs[planID]
	if !exists {
		c.JSON(400, gin.H{"error": "invalid_plan_id"})
		return
	}

	if expectedDeliveries == 0 {
		expectedDeliveries = plan.RangeStart
	}
	if expectedDeliveries < plan.RangeStart || expectedDeliveries > plan.RangeEnd {
		c.JSON(400, gin.H{
			"error": "expected_deliveries_out_of_range",
			"range": gin.H{"start": plan.RangeStart, "end": plan.RangeEnd},
		})
		return
	}

	if !paymentConfirmed {
		c.JSON(400, gin.H{"error": "payment_confirmation_required"})
		return
	}

	premiumAmount := premiumForRange(plan, expectedDeliveries)
	upgradeFee := 0
	requiredPayment := premiumAmount
	isInitialActivation := false

	if HasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			var currentPolicy models.Policy
			if err := workerDB.Where("worker_id = ?", workerIDUint).Order("id DESC").First(&currentPolicy).Error; err == nil {
				upgradeFee = planUpgradeFee(planConfigs, currentPolicy.PlanID, planID)

				var completedPayments int64
				_ = workerDB.Raw("SELECT COUNT(*) FROM premium_payments WHERE worker_id = ? AND status = 'completed'", workerIDUint).Scan(&completedPayments).Error
				if completedPayments == 0 || strings.EqualFold(currentPolicy.Status, "cancelled") || strings.EqualFold(currentPolicy.Status, "skipped") {
					isInitialActivation = true
				}
			} else {
				isInitialActivation = true
			}
			requiredPayment = premiumAmount + upgradeFee
			if isInitialActivation {
				requiredPayment = (premiumAmount * initialMultiplier) + upgradeFee
			}
		}
	}

	if paymentAmount < requiredPayment {
		c.JSON(400, gin.H{
			"error":                      "insufficient_payment_amount",
			"required_amount_inr":        requiredPayment,
			"upgrade_fee_inr":            upgradeFee,
			"is_initial_activation":      isInitialActivation,
			"initial_payment_multiplier": initialMultiplier,
		})
		return
	}

	if HasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			now := time.Now().UTC()
			policy := models.Policy{
				WorkerID:      workerIDUint,
				PlanID:        planID,
				Status:        "active",
				PremiumAmount: float64(premiumAmount),
			}

			var currentPolicy models.Policy
			currentErr := workerDB.Where("worker_id = ?", workerIDUint).Order("id DESC").First(&currentPolicy).Error
			if currentErr == nil {
				if err := workerDB.Model(&models.Policy{}).Where("id = ?", currentPolicy.ID).Updates(map[string]interface{}{
					"plan_id":        planID,
					"status":         "active",
					"premium_amount": float64(premiumAmount),
					"updated_at":     now,
				}).Error; err == nil {
					_ = workerDB.Exec(
						"INSERT INTO premium_payments (worker_id, policy_id, amount, status, payment_date) VALUES (?, ?, ?, 'completed', CURRENT_TIMESTAMP)",
						workerIDUint, currentPolicy.ID, paymentAmount,
					).Error
					_ = upsertPaymentSchedule(workerIDUint, now, false, "Active")

					c.JSON(200, gin.H{
						"message": "plan_selected_successfully",
						"plan": gin.H{
							"plan_id":                    planID,
							"plan_name":                  plan.PlanName,
							"range_start":                plan.RangeStart,
							"range_end":                  plan.RangeEnd,
							"selected_deliveries":        expectedDeliveries,
							"weekly_premium_inr":         premiumAmount,
							"coverage_ratio":             plan.CoverageRatio,
							"max_payout_inr":             plan.MaxPayoutINR,
							"upgrade_fee_inr":            upgradeFee,
							"required_payment_inr":       requiredPayment,
							"is_initial_activation":      isInitialActivation,
							"initial_payment_multiplier": initialMultiplier,
						},
						"policy": gin.H{
							"policy_id":                  fmt.Sprintf("pol-%03d", currentPolicy.ID),
							"plan_id":                    planID,
							"status":                     "active",
							"weekly_premium_inr":         premiumAmount,
							"coverage_ratio":             plan.CoverageRatio,
							"payment_amount_inr":         paymentAmount,
							"payment_status":             "Locked",
							"days_since_last_payment":    0,
							"next_payment_enabled":       false,
							"coverage_status":            "Active",
							"initial_payment_multiplier": initialMultiplier,
							"last_payment_timestamp":     now.Format(time.RFC3339),
						},
					})
					return
				}
			}

			if err := workerDB.Create(&policy).Error; err == nil {
				_ = workerDB.Exec(
					"INSERT INTO premium_payments (worker_id, policy_id, amount, status, payment_date) VALUES (?, ?, ?, 'completed', CURRENT_TIMESTAMP)",
					workerIDUint, policy.ID, paymentAmount,
				).Error
				_ = upsertPaymentSchedule(workerIDUint, now, false, "Active")

				c.JSON(200, gin.H{
					"message": "plan_selected_successfully",
					"plan": gin.H{
						"plan_id":                    planID,
						"plan_name":                  plan.PlanName,
						"range_start":                plan.RangeStart,
						"range_end":                  plan.RangeEnd,
						"selected_deliveries":        expectedDeliveries,
						"weekly_premium_inr":         premiumAmount,
						"coverage_ratio":             plan.CoverageRatio,
						"max_payout_inr":             plan.MaxPayoutINR,
						"upgrade_fee_inr":            upgradeFee,
						"required_payment_inr":       requiredPayment,
						"is_initial_activation":      isInitialActivation,
						"initial_payment_multiplier": initialMultiplier,
					},
					"policy": gin.H{
						"policy_id":                  fmt.Sprintf("pol-%03d", policy.ID),
						"plan_id":                    planID,
						"status":                     policy.Status,
						"weekly_premium_inr":         int(policy.PremiumAmount),
						"coverage_ratio":             plan.CoverageRatio,
						"payment_amount_inr":         paymentAmount,
						"payment_status":             "Locked",
						"days_since_last_payment":    0,
						"next_payment_enabled":       false,
						"coverage_status":            "Active",
						"initial_payment_multiplier": initialMultiplier,
						"last_payment_timestamp":     now.Format(time.RFC3339),
					},
				})
				return
			}
		}
	}

	// In-memory store
	store.mu.Lock()
	defer store.mu.Unlock()

	// Update or create policy in in-memory store
	policy := gin.H{
		"plan_id":                    planID,
		"plan_name":                  plan.PlanName,
		"plan_status":                "selected",
		"range_start":                plan.RangeStart,
		"range_end":                  plan.RangeEnd,
		"selected_deliveries":        expectedDeliveries,
		"weekly_premium_inr":         premiumAmount,
		"coverage_ratio":             plan.CoverageRatio,
		"max_payout_inr":             plan.MaxPayoutINR,
		"status":                     "active",
		"payment_amount_inr":         paymentAmount,
		"payment_status":             "Locked",
		"days_since_last_payment":    0,
		"next_payment_enabled":       false,
		"coverage_status":            "Active",
		"last_payment_timestamp":     nowISO(),
		"created_at":                 nowISO(),
		"upgrade_fee_inr":            upgradeFee,
		"required_payment_inr":       requiredPayment,
		"is_initial_activation":      isInitialActivation,
		"initial_payment_multiplier": initialMultiplier,
	}

	store.data.Policy = policy

	// Update worker profile to reflect selected plan
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		profile["plan_id"] = planID
		profile["coverage_status"] = "active"
		profile["enrolled"] = true
	}

	c.JSON(200, gin.H{
		"message": "plan_selected_successfully",
		"plan":    policy,
		"policy": gin.H{
			"plan_id":                 planID,
			"status":                  "active",
			"weekly_premium_inr":      premiumAmount,
			"coverage_ratio":          plan.CoverageRatio,
			"payment_amount_inr":      paymentAmount,
			"payment_status":          "Locked",
			"days_since_last_payment": 0,
			"next_payment_enabled":    false,
			"coverage_status":         "Active",
			"last_payment_timestamp":  policy["last_payment_timestamp"],
		},
	})
}

// SkipPlan marks plan selection as skipped so worker can start later.
func SkipPlan(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			policy := models.Policy{
				WorkerID:      workerIDUint,
				Status:        "skipped",
				PremiumAmount: 0,
			}
			if err := workerDB.Create(&policy).Error; err == nil {
				c.JSON(200, gin.H{
					"message": "plan_skipped",
					"policy": gin.H{
						"policy_id":          fmt.Sprintf("pol-%03d", policy.ID),
						"status":             "skipped",
						"plan_status":        "skipped",
						"weekly_premium_inr": 0,
					},
				})
				return
			}
		}
	}

	store.mu.Lock()
	store.data.Policy = gin.H{
		"status":             "skipped",
		"plan_status":        "skipped",
		"weekly_premium_inr": 0,
		"coverage_ratio":     0,
		"zone":               "",
		"next_due_date":      "",
		"shap_breakdown":     []gin.H{},
		"created_at":         nowISO(),
	}
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		profile["coverage_status"] = "inactive"
		profile["enrolled"] = false
	}
	store.mu.Unlock()

	c.JSON(200, gin.H{
		"message": "plan_skipped",
		"policy": gin.H{
			"status":             "skipped",
			"plan_status":        "skipped",
			"weekly_premium_inr": 0,
		},
	})
}
