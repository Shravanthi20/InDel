package insurer

import (
	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// GetClaimDetail returns full claim details.
func (h *InsurerHandler) GetClaimDetail(c *gin.Context) {
	claimID := c.Param("id")
	data, err := h.Service.GetClaimDetail(claimID)
	if err != nil {
		SendError(c, 404, "NOT_FOUND", "claim not found", "claim_id")
		return
	}
	SendSuccess(c, data)
}

// ReviewClaim processes manual decision.
func (h *InsurerHandler) ReviewClaim(c *gin.Context) {
	claimID := c.Param("id")
	var req models.ClaimAction
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, 400, "VALIDATION_ERROR", err.Error(), "")
		return
	}

	err := h.Service.ReviewClaim(claimID, req)
	if err != nil {
		SendError(c, 500, "INTERNAL_ERROR", "failed to review claim", "")
		return
	}

	SendSuccess(c, gin.H{"status": "success", "claim_id": claimID})
}
