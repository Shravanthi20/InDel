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
	v1.GET("/zone-levels", platform.GetZoneLevels)
	v1.GET("/zone-paths", platform.GetZonePaths)
	v1.GET("/zones/health", platform.GetZoneHealth)
	v1.GET("/disruptions", platform.GetDisruptions)

	controls := v1.Group("")
	controls.Use(platform.RequirePlatformOperatorRole())
	controls.POST("/demo/add-batches", platform.AddBatches)
	controls.POST("/demo/trigger-disruption", platform.TriggerDemoDisruption)

	webhooks := v1.Group("/webhooks")
	webhooks.Use(platform.RequirePlatformWebhookAuth())
	webhooks.POST("/order/assigned", platform.OrderAssignedWebhook)
	webhooks.POST("/order/completed", platform.OrderCompletedWebhook)
	webhooks.POST("/order/cancelled", platform.OrderCancelledWebhook)
	webhooks.POST("/external-signal", platform.ExternalSignalWebhook)
}
