package insurer

import (
	"strconv"
	"github.com/gin-gonic/gin"
)

// GetClaims returns insurer claims pipeline view.
func (h *InsurerHandler) GetClaims(c *gin.Context) {
	status := c.Query("status")
	fraudVerdict := c.Query("fraud_verdict")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	offset := (page - 1) * limit

	data, total, _ := h.Service.GetClaims(status, fraudVerdict, offset, limit)
	SendPaginated(c, data, page, limit, int(total))
}

// GetFraudQueue returns claims flagged for manual review.
func (h *InsurerHandler) GetFraudQueue(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	offset := (page - 1) * limit

	data, total, _ := h.Service.GetFraudQueue(offset, limit)
	SendPaginated(c, data, page, limit, int(total))
}
