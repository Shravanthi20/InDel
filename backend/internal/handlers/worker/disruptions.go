package worker

import "github.com/gin-gonic/gin"

// GetDisruptions returns active disruptions in zone
func GetDisruptions(c *gin.Context) {
	// GET /api/disruptions
	c.JSON(200, gin.H{"disruptions": []interface{}{}})
}
