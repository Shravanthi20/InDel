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

	// Create the service object early (even with nil fields) so routes can be registered
	svc := services.NewInsurerService(nil, nil)

	// Initialize DB, Redis and Kafka in background to allow immediate health check response
	go func() {
		cfg := config.Load()
		if _, err := database.InitRedis(cfg); err != nil {
			log.Printf("Background: Redis unavailable: %v", err)
		}
		db, err := database.InitDB(cfg)
		if err != nil {
			log.Printf("Background: Insurer Gateway DB unavailable: %v", err)
		} else {
			log.Println("Background: Insurer Gateway connected to PostgreSQL")
			svc.DB = db

			var kp *kafka.Producer
			kafkaBrokers := os.Getenv("KAFKA_BROKERS")
			if kafkaBrokers != "" {
				kp, _ = kafka.NewProducer(kafkaBrokers)
				svc.KafkaProducer = kp
			}
		}
	}()

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
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("INSURER_GATEWAY_PORT")
	}
	if port == "" {
		port = "8002"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Insurer Gateway listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
