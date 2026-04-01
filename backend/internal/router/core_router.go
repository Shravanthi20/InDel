package router

import (
	"github.com/Shravanthi20/InDel/backend/internal/handlers/core"
	"github.com/gin-gonic/gin"
)

// SetupCoreRoutes sets up core internal routes.
func SetupCoreRoutes(router *gin.Engine) {
	internal := router.Group("/api/v1/internal")
	internal.POST("/policy/weekly-cycle/run", core.RunWeeklyCycle)
	internal.POST("/claims/generate-for-disruption/:disruption_id", core.GenerateClaimsForDisruption)
	internal.POST("/payouts/queue/:claim_id", core.QueueClaimPayout)
	internal.POST("/payouts/process", core.ProcessPayouts)
	internal.GET("/payouts/reconciliation", core.GetPayoutReconciliation)
	internal.POST("/data/synthetic/generate", core.GenerateSyntheticData)

	legacy := router.Group("/internal/v1")
	legacy.POST("/claims/:claim_id/payout", core.QueueClaimPayout)
}
