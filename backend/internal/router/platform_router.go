package router

import (
	"github.com/Shravanthi20/InDel/backend/internal/handlers/platform"
	"github.com/gin-gonic/gin"
)

// SetupPlatformRoutes sets up platform gateway routes.
func SetupPlatformRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/platform")
	v1.GET("/workers", platform.GetWorkers)
	v1.GET("/zones", platform.GetZones)
	v1.POST("/webhooks/order/assigned", platform.OrderAssignedWebhook)
	v1.POST("/webhooks/order/completed", platform.OrderCompletedWebhook)
}
