package worker

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPayouts returns payout history
func GetPayouts(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	limit := 10
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			type row struct {
				PayoutID    uint    `gorm:"column:payout_id"`
				ClaimID     uint    `gorm:"column:claim_id"`
				Amount      float64 `gorm:"column:amount"`
				Status      string  `gorm:"column:status"`
				ProcessedAt string  `gorm:"column:processed_at"`
			}
			rows := make([]row, 0)
			_ = workerDB.Raw(
				"SELECT id as payout_id, claim_id, amount, status, created_at::text as processed_at FROM payouts WHERE worker_id = ? ORDER BY created_at DESC LIMIT ?",
				workerIDUint, limit,
			).Scan(&rows).Error

			resp := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				resp = append(resp, gin.H{
					"payout_id":    fmt.Sprintf("pay_%d", row.PayoutID),
					"claim_id":     fmt.Sprintf("clm_%d", row.ClaimID),
					"amount":       int(row.Amount),
					"method":       "upi",
					"status":       row.Status,
					"processed_at": row.ProcessedAt,
				})
			}
			c.JSON(200, gin.H{"payouts": resp})
			return
		}
	}

	store.mu.RLock()
	payouts := store.data.Payouts
	if limit > len(payouts) {
		limit = len(payouts)
	}
	resp := append([]map[string]any{}, payouts[:limit]...)
	store.mu.RUnlock()

	c.JSON(200, gin.H{"payouts": resp})
}

// GetWallet returns wallet balance
func GetWallet(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			var available float64 = 0
			var lastAmount float64 = 0
			var lastAt string
			_ = workerDB.Raw("SELECT COALESCE(SUM(amount),0) FROM payouts WHERE worker_id = ? AND status IN ('processed', 'credited', 'completed')", workerIDUint).Scan(&available).Error
			_ = workerDB.Raw("SELECT COALESCE(amount,0), COALESCE(created_at::text,'') FROM payouts WHERE worker_id = ? ORDER BY created_at DESC LIMIT 1", workerIDUint).Row().Scan(&lastAmount, &lastAt)

			c.JSON(200, gin.H{
				"currency":           "INR",
				"available_balance":  int(available),
				"last_payout_amount": int(lastAmount),
				"last_payout_at":     lastAt,
			})
			return
		}
	}

	store.mu.RLock()
	wallet := store.data.Wallet
	store.mu.RUnlock()

	c.JSON(200, wallet)
}

// ConfirmPayout confirms payout
func ConfirmPayout(c *gin.Context) {
	// POST /api/payouts/:id/confirm
	c.JSON(200, gin.H{"status": "confirmed"})
}
