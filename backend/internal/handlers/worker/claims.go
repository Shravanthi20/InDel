package worker

import "github.com/gin-gonic/gin"

// GetClaims returns claim history
func GetClaims(c *gin.Context) {
	// GET /api/claims
	c.JSON(200, gin.H{"claims": []interface{}{}})
}

// GetClaimDetail returns claim details
func GetClaimDetail(c *gin.Context) {
	// GET /api/claims/:id
	c.JSON(200, gin.H{"claim_id": 1, "amount": 5000, "status": "approved"})
}
