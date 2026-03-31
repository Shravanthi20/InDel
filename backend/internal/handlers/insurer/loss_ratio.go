package insurer

import "github.com/gin-gonic/gin"

// GetLossRatio returns loss ratio by zone/city
func (h *InsurerHandler) GetLossRatio(c *gin.Context) {
	zoneID := c.Query("zone_id")
	data, _ := h.Service.GetLossRatio(zoneID)
	SendSuccess(c, data)
}
