package insurer

import "github.com/gin-gonic/gin"

// GetLossRatio returns loss ratio by zone/city
func GetLossRatio(c *gin.Context) {
	// GET /api/loss-ratio
	c.JSON(200, gin.H{"zones": []interface{}{}})
}
