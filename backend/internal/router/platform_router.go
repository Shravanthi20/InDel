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
	v1.GET("/zone-paths", platform.GetZonePaths)
	v1.POST("/demo/trigger-disruption", platform.TriggerDemoDisruption)
	v1.POST("/webhooks/order/assigned", platform.OrderAssignedWebhook)
	v1.POST("/webhooks/order/completed", platform.OrderCompletedWebhook)
	v1.POST("/webhooks/order/cancelled", platform.OrderCancelledWebhook)
	v1.POST("/webhooks/external-signal", platform.ExternalSignalWebhook)
	v1.GET("/zones/health", platform.GetZoneHealth)
	v1.GET("/disruptions", platform.GetDisruptions)

	// Demo endpoint mapped under /demo as per spec, though sometimes requested under /platform
	router.POST("/api/v1/demo/trigger-disruption", platform.TriggerDemoDisruption)
}
