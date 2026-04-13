package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/middleware"
	routerpkg "github.com/Shravanthi20/InDel/backend/internal/router"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil && os.Getenv("INDEL_ENV") != "production" {
		log.Println("No .env file found, using environment variables")
	}

	// Create Gin router
	router := gin.Default()
	router.Use(middleware.CORS())

	// Optional DB wiring for live aggregate metrics.
	cfg := config.Load()
	var svc *services.InsurerService
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("Insurer Gateway DB unavailable, using fallback responses: %v", err)
		svc = services.NewInsurerService(nil, nil)
	} else {
		log.Println("Insurer Gateway connected to PostgreSQL")

		var kp *kafka.Producer
		kafkaBrokers := os.Getenv("KAFKA_BROKERS")
		if kafkaBrokers != "" {
			kp, _ = kafka.NewProducer(kafkaBrokers)
		}
		svc = services.NewInsurerService(db, kp)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "insurer-gateway"})
	})
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "service": "insurer-gateway", "time": "mock"})
	})
	router.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up", "environment": os.Getenv("INDEL_ENV")})
	})

	// API routes
	routerpkg.SetupInsurerRoutes(router, svc)

	// Start server
	port := os.Getenv("INSURER_GATEWAY_PORT")
	if port == "" {
		port = "8002"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Insurer Gateway listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
