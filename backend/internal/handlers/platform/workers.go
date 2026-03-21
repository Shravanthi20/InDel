package platform

import "github.com/gin-gonic/gin"

// GetWorkers returns worker list
func GetWorkers(c *gin.Context) {
	// GET /api/workers
	c.JSON(200, gin.H{"workers": []interface{}{}})
}
