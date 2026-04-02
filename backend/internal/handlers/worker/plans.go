package worker

import (
	"fmt"

	"github.com/Shravanthi20/InDel/backend/internal/models"
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

func getPlanConfigs() map[string]planConfig {
	return map[string]planConfig{
		"plan-starter": {
			PlanID:        "plan-starter",
			PlanName:      "Range-01: Starter",
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
			PlanName:      "Range-02: Growth",
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
			PlanName:      "Range-03: Premium",
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

// GetPlans returns available delivery plans based on order ranges
func GetPlans(c *gin.Context) {
	planConfigs := getPlanConfigs()
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
	if paymentAmount < premiumAmount {
		c.JSON(400, gin.H{
			"error":               "insufficient_payment_amount",
			"required_amount_inr": premiumAmount,
		})
		return
	}

	if hasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			// Create or update policy
			policy := models.Policy{
				WorkerID:      workerIDUint,
				Status:        "active",
				PremiumAmount: float64(premiumAmount),
			}
			if err := workerDB.Create(&policy).Error; err == nil {
				_ = workerDB.Exec(
					"INSERT INTO premium_payments (worker_id, policy_id, amount, status, payment_date) VALUES (?, ?, ?, 'completed', CURRENT_TIMESTAMP)",
					workerIDUint, policy.ID, paymentAmount,
				).Error

				c.JSON(200, gin.H{
					"message": "plan_selected_successfully",
					"plan": gin.H{
						"plan_id":             planID,
						"plan_name":           plan.PlanName,
						"range_start":         plan.RangeStart,
						"range_end":           plan.RangeEnd,
						"selected_deliveries": expectedDeliveries,
						"weekly_premium_inr":  premiumAmount,
						"coverage_ratio":      plan.CoverageRatio,
						"max_payout_inr":      plan.MaxPayoutINR,
					},
					"policy": gin.H{
						"policy_id":          fmt.Sprintf("pol-%03d", policy.ID),
						"status":             policy.Status,
						"weekly_premium_inr": int(policy.PremiumAmount),
						"coverage_ratio":     plan.CoverageRatio,
						"payment_amount_inr": paymentAmount,
						"payment_status":     "completed",
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
		"plan_id":             planID,
		"plan_name":           plan.PlanName,
		"plan_status":         "selected",
		"range_start":         plan.RangeStart,
		"range_end":           plan.RangeEnd,
		"selected_deliveries": expectedDeliveries,
		"weekly_premium_inr":  premiumAmount,
		"coverage_ratio":      plan.CoverageRatio,
		"max_payout_inr":      plan.MaxPayoutINR,
		"status":              "active",
		"payment_amount_inr":  paymentAmount,
		"payment_status":      "completed",
		"created_at":          nowISO(),
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
			"status":             "active",
			"weekly_premium_inr": premiumAmount,
			"coverage_ratio":     plan.CoverageRatio,
			"payment_amount_inr": paymentAmount,
			"payment_status":     "completed",
		},
	})
}

// SkipPlan marks plan selection as skipped so worker can start later.
func SkipPlan(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
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
