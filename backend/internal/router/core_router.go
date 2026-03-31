package router

import (
	"github.com/Shravanthi20/InDel/backend/internal/handlers/core"
	"github.com/gin-gonic/gin"
)

// SetupCoreRoutes sets up core internal routes.
func SetupCoreRoutes(router *gin.Engine) {
	internal := router.Group("/internal/v1")
	internal.POST("/claims/:claim_id/payout", core.QueueClaimPayout)
}
