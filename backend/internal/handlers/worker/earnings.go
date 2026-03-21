package worker

import "github.com/gin-gonic/gin"

// GetEarnings returns weekly earnings
func GetEarnings(c *gin.Context) {
	// GET /api/earnings/summary
	c.JSON(200, gin.H{"this_week": 5000, "baseline": 5500})
}

// GetEarningsHistory returns monthly history
func GetEarningsHistory(c *gin.Context) {
	// GET /api/earnings/history
	c.JSON(200, gin.H{"earnings": []interface{}{}})
}
