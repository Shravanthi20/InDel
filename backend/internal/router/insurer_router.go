package router

import (
	"github.com/Shravanthi20/InDel/backend/internal/handlers/insurer"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupInsurerRoutes sets up insurer gateway routes.
func SetupInsurerRoutes(router *gin.Engine, svc *services.InsurerService) {
	h := insurer.NewInsurerHandler(svc)
	v1 := router.Group("/api/v1/insurer")
	v1.GET("/overview", h.GetOverview)
	v1.GET("/loss-ratio", h.GetLossRatio)
	v1.GET("/claims", h.GetClaims)
	v1.GET("/claims/fraud-queue", h.GetFraudQueue)
	v1.GET("/claims/:id", h.GetClaimDetail)
	v1.POST("/claims/:id/review", h.ReviewClaim)
	v1.GET("/forecast", h.GetForecast)
	v1.GET("/pool/health", h.GetPoolHealth)
}
