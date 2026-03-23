package worker

import (
	"fmt"

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
			var amount float64 = 22
			_ = workerDB.Raw("SELECT premium_amount FROM policies WHERE worker_id = ? ORDER BY id DESC LIMIT 1", workerIDUint).Scan(&amount).Error
			c.JSON(200, gin.H{
				"weekly_premium_inr": int(amount),
				"currency":           "INR",
				"shap_breakdown": []gin.H{
					{"feature": "rain_risk", "impact": 0.42},
					{"feature": "order_drop_volatility", "impact": 0.31},
					{"feature": "historical_disruptions", "impact": 0.27},
				},
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
			_ = workerDB.Exec(
				"INSERT INTO premium_payments (worker_id, policy_id, amount, status, payment_date) VALUES (?, (SELECT id FROM policies WHERE worker_id = ? ORDER BY id DESC LIMIT 1), ?, 'completed', CURRENT_TIMESTAMP)",
				workerIDUint, workerIDUint, amount,
			).Error
			c.JSON(200, gin.H{
				"message":    "payment_successful",
				"amount":     amount,
				"currency":   "INR",
				"payment_id": fmt.Sprintf("db-payment-%d", workerIDUint),
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"message":    "payment_successful",
		"amount":     amount,
		"currency":   "INR",
		"payment_id": "mock-payment-001",
	})
}
