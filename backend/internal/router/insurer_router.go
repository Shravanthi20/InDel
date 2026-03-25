package router

import (
	"github.com/Shravanthi20/InDel/backend/internal/handlers/insurer"
	"github.com/gin-gonic/gin"
)

// SetupInsurerRoutes sets up insurer gateway routes.
func SetupInsurerRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/insurer")
	v1.GET("/overview", insurer.GetOverview)
	v1.GET("/loss-ratio", insurer.GetLossRatio)
	v1.GET("/claims", insurer.GetClaims)
	v1.GET("/claims/fraud-queue", insurer.GetFraudQueue)
	v1.GET("/forecast", insurer.GetForecast)
	v1.GET("/pool/health", insurer.GetPoolHealth)
}
