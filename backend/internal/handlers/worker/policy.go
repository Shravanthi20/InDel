package worker

import "github.com/gin-gonic/gin"

// GetPolicy returns active policy
func GetPolicy(c *gin.Context) {
	// GET /api/policy
	c.JSON(200, gin.H{"policy_id": 1, "status": "active", "premium": 300})
}

// EnrollPolicy enrolls in coverage
func EnrollPolicy(c *gin.Context) {
	// POST /api/policy/enroll
	c.JSON(201, gin.H{"policy_id": 1})
}

// PausePolicy pauses coverage
func PausePolicy(c *gin.Context) {
	// POST /api/policy/pause
	c.JSON(200, gin.H{"status": "paused"})
}

// CancelPolicy cancels coverage
func CancelPolicy(c *gin.Context) {
	// POST /api/policy/cancel
	c.JSON(200, gin.H{"status": "cancelled"})
}
