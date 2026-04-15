package worker

import (
	"fmt"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// GetPremium returns current premium
func GetPremium(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			quote, _ := services.QuotePremium(workerDB, workerIDUint, time.Now().UTC())
			amount := 65.0 // Mid-level default if no quote exists
			source := "fallback"
			riskScore := 0.0
			modelVersion := "fallback_rule_v2"
			var breakdown []gin.H

			if quote != nil {
				amount = quote.WeeklyPremiumINR
				source = quote.Source
				riskScore = quote.RiskScore
				modelVersion = quote.ModelVersion
				breakdown = make([]gin.H, 0, len(quote.Explainability))
				for _, item := range quote.Explainability {
					breakdown = append(breakdown, gin.H{"feature": item.Feature, "impact": item.Impact})
				}
				fmt.Printf("[PREMIUM] ML used for worker %v: amount=%.2f, model=%s\n", workerID, amount, modelVersion)
			} else {
				// Static breakdown if no quote available
				breakdown = []gin.H{
					{"feature": "rain_risk", "impact": 0.42},
					{"feature": "order_drop_volatility", "impact": 0.31},
					{"feature": "historical_disruptions", "impact": 0.27},
				}
				fmt.Printf("[PREMIUM] Fallback used for worker %v: amount=%.2f\n", workerID, amount)
			}
			c.JSON(200, gin.H{
				"weekly_premium_inr": int(amount),
				"currency":           "INR",
				"risk_score":         riskScore,
				"pricing_source":     source,
				"model_version":      modelVersion,
				"shap_breakdown":     breakdown,
			})
			return
		}
	}

	store.mu.RLock()
	resp := gin.H{
		"weekly_premium_inr": store.data.Policy["weekly_premium_inr"],
		"currency":           "INR",
		"shap_breakdown":     store.data.Policy["shap_breakdown"],
	}
	store.mu.RUnlock()

	c.JSON(200, resp)
}

// PayPremium makes premium payment
func PayPremium(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)

	store.mu.RLock()
	defaultAmount, _ := store.data.Policy["weekly_premium_inr"].(int)
	store.mu.RUnlock()

	amount := bodyInt(body, "amount", defaultAmount)
	if amount <= 0 {
		amount = defaultAmount
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			now := time.Now().UTC()
			checkoutID := fmt.Sprintf("chk_%d_%d", workerIDUint, now.Unix())

			state, scheduleErr := getOrBootstrapPaymentSchedule(workerIDUint, now)
			if scheduleErr == nil {
				syncPolicyStatusWithPaymentState(workerIDUint, state)
			}

			var policyID uint
			var currentPremium int
			var policyStatus string
			var p struct {
				ID            uint    `gorm:"column:id"`
				Status        string  `gorm:"column:status"`
				PremiumAmount float64 `gorm:"column:premium_amount"`
			}
			if err := workerDB.Raw("SELECT id, status, premium_amount FROM policies WHERE worker_id = ? ORDER BY id DESC LIMIT 1", workerIDUint).Scan(&p).Error; err == nil {
				policyID = p.ID
				policyStatus = p.Status
				currentPremium = int(p.PremiumAmount)
			}

			activationPayment := policyStatus != "" && policyStatus != "active"

			if !activationPayment && scheduleErr == nil && state.LastPaymentRecorded != nil && !state.LastPaymentRecorded.IsZero() && !state.NextPaymentEnabled {
				nextDue := state.LastPaymentRecorded.AddDate(0, 0, state.BillingCycleDays).Format("2006-01-02")
				c.JSON(409, gin.H{
					"error":           paymentLockError(state),
					"message":         "premium_already_paid_for_current_week",
					"next_due_date":   nextDue,
					"payment_status":  state.PaymentStatus,
					"coverage_status": state.CoverageStatus,
				})
				return
			}

			quote, _ := services.QuotePremium(workerDB, workerIDUint, now)
			basePremium := currentPremium
			if quote != nil && quote.WeeklyPremiumINR > 0 {
				basePremium = int(quote.WeeklyPremiumINR)
			}
			if basePremium <= 0 {
				basePremium = defaultAmount
			}

			lateFee := 0
			if scheduleErr == nil && !activationPayment {
				lateFee = state.LateFeeINR
			}
			requiredAmount := basePremium + lateFee
			if activationPayment {
				requiredAmount = basePremium * initialMultiplier
			}
			if amount < requiredAmount {
				c.JSON(400, gin.H{
					"error":               "insufficient_payment_amount",
					"required_amount_inr": requiredAmount,
					"base_premium_inr":    basePremium,
					"late_fee_inr":        lateFee,
					"is_activation":       activationPayment,
				})
				return
			}

			_ = workerDB.Exec(
				"INSERT INTO premium_payments (worker_id, policy_id, amount, status, payment_date) VALUES (?, (SELECT id FROM policies WHERE worker_id = ? ORDER BY id DESC LIMIT 1), ?, 'completed', CURRENT_TIMESTAMP)",
				workerIDUint, workerIDUint, amount,
			).Error

			nextQuote, _ := services.QuotePremium(workerDB, workerIDUint, now)
			nextPremium := basePremium
			if nextQuote != nil && nextQuote.WeeklyPremiumINR > 0 {
				nextPremium = int(nextQuote.WeeklyPremiumINR)
			}

			_ = workerDB.Exec(
				"UPDATE policies SET status = 'active', premium_amount = ?, updated_at = CURRENT_TIMESTAMP WHERE worker_id = ?",
				nextPremium, workerIDUint,
			).Error
			_ = workerDB.Exec(
				"INSERT INTO active_policies (user_id, policy_id, zone, started_at, updated_at) SELECT ?, p.id, COALESCE(z.name, ''), NOW(), NOW() FROM policies p LEFT JOIN worker_profiles wp ON wp.worker_id = p.worker_id LEFT JOIN zones z ON z.id = wp.zone_id WHERE p.worker_id = ? ORDER BY p.id DESC LIMIT 1 ON CONFLICT (user_id) DO UPDATE SET policy_id = EXCLUDED.policy_id, zone = EXCLUDED.zone, updated_at = NOW()",
				workerIDUint, workerIDUint,
			).Error
			_ = upsertPaymentSchedule(workerIDUint, now, false, "Active")
			_ = workerDB.Exec(
				"INSERT INTO notifications (worker_id, type, message, created_at) VALUES (?, 'premium_due', ?, CURRENT_TIMESTAMP)",
				workerIDUint, fmt.Sprintf("Premium payment of Rs %d completed. Coverage remains active.", amount),
			).Error
			c.JSON(200, gin.H{
				"message":               "payment_successful",
				"amount":                amount,
				"currency":              "INR",
				"payment_id":            fmt.Sprintf("db-payment-%d", workerIDUint),
				"checkout_id":           checkoutID,
				"payment_status":        "completed",
				"checkout_mode":         "platform_demo_checkout",
				"base_premium_inr":      basePremium,
				"late_fee_inr":          lateFee,
				"required_payment_inr":  requiredAmount,
				"is_activation":         activationPayment,
				"next_week_premium_inr": nextPremium,
				"next_due_date":         now.AddDate(0, 0, 7).Format("2006-01-02"),
				"grace_period_days":     2,
				"policy_id":             policyID,
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"message":        "payment_successful",
		"amount":         amount,
		"currency":       "INR",
		"payment_id":     "mock-payment-001",
		"checkout_id":    "chk_mock_001",
		"payment_status": "completed",
		"checkout_mode":  "mock_checkout",
	})
}
