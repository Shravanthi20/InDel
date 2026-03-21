package worker

import "github.com/gin-gonic/gin"

// GetPayouts returns payout history
func GetPayouts(c *gin.Context) {
	// GET /api/payouts/history
	c.JSON(200, gin.H{"payouts": []interface{}{}})
}

// GetWallet returns wallet balance
func GetWallet(c *gin.Context) {
	// GET /api/payouts/wallet
	c.JSON(200, gin.H{"balance": 5000})
}

// ConfirmPayout confirms payout
func ConfirmPayout(c *gin.Context) {
	// POST /api/payouts/:id/confirm
	c.JSON(200, gin.H{"status": "confirmed"})
}
