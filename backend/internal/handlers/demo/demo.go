package demo

import "github.com/gin-gonic/gin"

// TriggerDisruption triggers a demo disruption
func TriggerDisruption(c *gin.Context) {
	// POST /demo/trigger-disruption
	c.JSON(200, gin.H{"disruption_id": 1})
}

// SettleEarnings settles earnings for demo
func SettleEarnings(c *gin.Context) {
	// POST /demo/settle-earnings
	c.JSON(200, gin.H{"settled": true})
}

// ResetZone resets demo zone
func ResetZone(c *gin.Context) {
	// POST /demo/reset-zone
	c.JSON(200, gin.H{"reset": true})
}
