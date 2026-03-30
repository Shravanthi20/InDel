package insurer

import "github.com/gin-gonic/gin"

// GetOverview returns KPI overview
func (h *InsurerHandler) GetOverview(c *gin.Context) {
	data, poolHealth, _ := h.Service.GetOverview()
	
	SendSuccess(c, gin.H{
		"active_workers":      data.ActiveWorkers,
		"pending_claims":      data.PendingClaims,
		"approved_claims":     data.ApprovedClaims,
		"loss_ratio":          data.LossRatio,
		"reserve_utilization": data.ReserveUtilization,
		"reserve":             data.Reserve,
		"pool_health":         poolHealth,
	})
}
