package insurer

import "github.com/gin-gonic/gin"

// GetOverview returns KPI overview
func GetOverview(c *gin.Context) {
	// GET /api/overview
	c.JSON(200, gin.H{"pool_health": "good", "loss_ratio": 0.45})
}
