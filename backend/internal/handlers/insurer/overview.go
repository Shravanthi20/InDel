package insurer

import "github.com/gin-gonic/gin"

// GetOverview returns KPI overview
func GetOverview(c *gin.Context) {
	if hasDB() {
		var activeWorkers int64
		var pendingClaims int64
		var premiums float64
		var payouts float64

		_ = insurerDB.Raw("SELECT COUNT(DISTINCT worker_id) FROM policies WHERE status = 'active'").Scan(&activeWorkers).Error
		_ = insurerDB.Raw("SELECT COUNT(*) FROM claims WHERE status IN ('pending', 'manual_review')").Scan(&pendingClaims).Error
		_ = insurerDB.Raw("SELECT COALESCE(SUM(amount), 0) FROM premium_payments WHERE status IN ('completed', 'captured', 'processed')").Scan(&premiums).Error
		_ = insurerDB.Raw("SELECT COALESCE(SUM(amount), 0) FROM payouts WHERE status IN ('processed', 'credited', 'completed')").Scan(&payouts).Error

		lossRatio := 0.0
		if premiums > 0 {
			lossRatio = payouts / premiums
		}

		poolHealth := "healthy"
		if lossRatio > 0.8 {
			poolHealth = "watch"
		}
		if lossRatio > 1.0 {
			poolHealth = "critical"
		}

		c.JSON(200, gin.H{
			"active_workers": activeWorkers,
			"pending_claims": pendingClaims,
			"loss_ratio":     lossRatio,
			"reserve":        int(premiums - payouts),
			"pool_health":    poolHealth,
		})
		return
	}

	c.JSON(200, gin.H{
		"active_workers": 3,
		"pending_claims": 1,
		"loss_ratio":     0.45,
		"reserve":        1260,
		"pool_health":    "healthy",
	})
}
