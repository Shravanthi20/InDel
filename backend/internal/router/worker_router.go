package router

import (
	"github.com/Shravanthi20/InDel/backend/internal/handlers/worker"
	"github.com/gin-gonic/gin"
)

// SetupWorkerRoutes sets up worker gateway routes
func SetupWorkerRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")

	// Public auth endpoints.
	v1.POST("/auth/register", worker.Register)
	v1.POST("/auth/login", worker.Login)
	v1.POST("/auth/otp/send", worker.SendOTP)
	v1.POST("/auth/otp/verify", worker.VerifyOTP)

	// Worker protected surface.
	v1.POST("/worker/onboard", worker.Onboard)
	v1.GET("/worker/profile", worker.GetProfile)
	v1.PUT("/worker/profile", worker.UpdateProfile)

	v1.GET("/worker/policy", worker.GetPolicy)
	v1.POST("/worker/policy/enroll", worker.EnrollPolicy)
	v1.PUT("/worker/policy/pause", worker.PausePolicy)
	v1.PUT("/worker/policy/cancel", worker.CancelPolicy)
	v1.GET("/worker/policy/premium", worker.GetPremium)
	v1.POST("/worker/policy/premium/pay", worker.PayPremium)

	v1.GET("/worker/plans", worker.GetPlans)
	v1.POST("/worker/plans/select", worker.SelectPlan)
	v1.POST("/worker/plans/skip", worker.SkipPlan)

	v1.GET("/worker/earnings", worker.GetEarnings)
	v1.GET("/worker/earnings/history", worker.GetEarningsHistory)
	v1.GET("/worker/earnings/baseline", worker.GetEarningsBaseline)

	v1.GET("/worker/claims", worker.GetClaims)
	v1.GET("/worker/claims/:claim_id", worker.GetClaimDetail)
	v1.GET("/worker/wallet", worker.GetWallet)
	v1.GET("/worker/payouts", worker.GetPayouts)

	v1.GET("/worker/orders", worker.GetOrders)
	v1.GET("/worker/orders/assigned", worker.GetAssignedOrders)
	v1.GET("/worker/batches", worker.GetAvailableBatches)
	v1.GET("/worker/batches/assigned", worker.GetAssignedBatches)
	v1.PUT("/worker/orders/:order_id/accept", worker.AcceptOrder)
	v1.PUT("/worker/orders/:order_id/picked-up", worker.PickedUpOrder)
	v1.PUT("/worker/orders/:order_id/delivered", worker.DeliverOrder)

	v1.GET("/worker/notifications", worker.GetNotifications)
	v1.PUT("/worker/notifications/preferences", worker.SetNotificationPreferences)
	v1.POST("/worker/notifications/fcm-token", worker.RegisterFCMToken)

	// Demo control endpoints.
	v1.POST("/demo/trigger-disruption", worker.DemoTriggerDisruption)
	v1.POST("/demo/settle-earnings", worker.DemoSettleEarnings)
	v1.POST("/demo/reset-zone", worker.DemoResetZone)
	v1.POST("/demo/reset", worker.DemoReset)
	v1.POST("/demo/simulate-orders", worker.DemoSimulateOrders)
	v1.POST("/demo/orders/publisher/initiate", worker.InitiateOrderPublisher)
	v1.POST("/demo/orders/publisher/ack", worker.AckOrderPublisher)
	v1.GET("/demo/orders/publisher/status", worker.GetOrderPublisherStatus)
	v1.POST("/demo/orders/ingest", worker.IngestDemoOrder)
	v1.GET("/demo/orders/search", worker.SearchDemoOrders)
	v1.GET("/demo/orders/available", worker.GetAvailableOrders)
	v1.GET("/demo/deliveries", worker.GetDeliveries)
}
