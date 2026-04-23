package router

import (
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/handlers/insurer"
	"github.com/Shravanthi20/InDel/backend/internal/middleware"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupInsurerRoutes sets up insurer gateway routes.
func SetupInsurerRoutes(router *gin.Engine, svc *services.InsurerService) {
	h := insurer.NewInsurerHandler(svc)
	v1 := router.Group("/api/v1/insurer")
	v1.GET("/overview", middleware.RedisCache(15*time.Second), h.GetOverview)
	v1.GET("/loss-ratio", middleware.RedisCache(15*time.Second), h.GetLossRatio)
	v1.GET("/claims", middleware.RedisCache(15*time.Second), h.GetClaims)
	v1.GET("/claims/fraud-queue", h.GetFraudQueue)
	v1.GET("/claims/:id", h.GetClaimDetail)
	v1.POST("/claims/:id/review", h.ReviewClaim)
	v1.GET("/forecast", middleware.RedisCache(15*time.Second), h.GetForecast)
	v1.GET("/pool/health", middleware.RedisCache(15*time.Second), h.GetPoolHealth)
	v1.GET("/money-exchange", h.GetMoneyExchange)
	v1.GET("/users", middleware.RedisCache(15*time.Second), h.GetPlanUsers)
	v1.POST("/users/:id/plan/start", h.StartUserPlan)
	v1.POST("/users/:id/plan/end", h.EndUserPlan)
	v1.GET("/plan-stats", middleware.RedisCache(15*time.Second), h.GetPlanUsers)
	v1.GET("/ledger", h.GetLedger)
}
