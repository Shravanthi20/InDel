package worker

import "github.com/gin-gonic/gin"

// Onboard completes worker onboarding
func Onboard(c *gin.Context) {
	// POST /api/worker/onboard
	c.JSON(201, gin.H{"worker_id": 1})
}

// GetProfile returns worker profile
func GetProfile(c *gin.Context) {
	// GET /api/worker/profile
	c.JSON(200, gin.H{"name": "John", "zone": "Zone-A", "vehicle": "bike"})
}

// UpdateProfile updates worker profile
func UpdateProfile(c *gin.Context) {
	// PUT /api/worker/profile
	c.JSON(200, gin.H{"updated": true})
}
