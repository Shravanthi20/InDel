package worker

import "github.com/gin-gonic/gin"

// GetPremium returns current premium
func GetPremium(c *gin.Context) {
	// GET /api/premium
	c.JSON(200, gin.H{"premium": 300, "due_date": "2026-03-28"})
}

// PayPremium makes premium payment
func PayPremium(c *gin.Context) {
	// POST /api/premium/pay
	c.JSON(200, gin.H{"payment_id": "pay_123", "status": "completed"})
}
