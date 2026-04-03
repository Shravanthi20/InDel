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

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			quote, _ := services.QuotePremium(workerDB, workerIDUint, time.Now().UTC())
			amount := 22
			source := "fallback"
			riskScore := 0.0
			modelVersion := "fallback_rule_v2"
			breakdown := []gin.H{
				{"feature": "rain_risk", "impact": 0.42},
				{"feature": "order_drop_volatility", "impact": 0.31},
				{"feature": "historical_disruptions", "impact": 0.27},
			}
			if quote != nil {
				amount = int(quote.WeeklyPremiumINR)
				source = quote.Source
				riskScore = quote.RiskScore
				modelVersion = quote.ModelVersion
				breakdown = make([]gin.H, 0, len(quote.Explainability))
				for _, item := range quote.Explainability {
					breakdown = append(breakdown, gin.H{"feature": item.Feature, "impact": item.Impact})
				}
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

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			checkoutID := fmt.Sprintf("chk_%d_%d", workerIDUint, time.Now().UTC().Unix())
			_ = workerDB.Exec(
				"INSERT INTO premium_payments (worker_id, policy_id, amount, status, payment_date) VALUES (?, (SELECT id FROM policies WHERE worker_id = ? ORDER BY id DESC LIMIT 1), ?, 'completed', CURRENT_TIMESTAMP)",
				workerIDUint, workerIDUint, amount,
			).Error
			_ = workerDB.Exec(
				"UPDATE policies SET status = 'active', premium_amount = ?, updated_at = CURRENT_TIMESTAMP WHERE worker_id = ?",
				amount, workerIDUint,
			).Error
			_ = workerDB.Exec(
				"INSERT INTO notifications (worker_id, type, message, created_at) VALUES (?, 'premium_due', ?, CURRENT_TIMESTAMP)",
				workerIDUint, fmt.Sprintf("Premium payment of Rs %d completed. Coverage remains active.", amount),
			).Error
			c.JSON(200, gin.H{
				"message":       "payment_successful",
				"amount":        amount,
				"currency":      "INR",
				"payment_id":    fmt.Sprintf("db-payment-%d", workerIDUint),
				"checkout_id":   checkoutID,
				"payment_status":"completed",
				"checkout_mode": "platform_demo_checkout",
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"message":       "payment_successful",
		"amount":        amount,
		"currency":      "INR",
		"payment_id":    "mock-payment-001",
		"checkout_id":   "chk_mock_001",
		"payment_status":"completed",
		"checkout_mode": "mock_checkout",
	})
}
