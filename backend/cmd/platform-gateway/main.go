package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/handlers/platform"
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/middleware"
	routerpkg "github.com/Shravanthi20/InDel/backend/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.BootstrapServiceEnv("platform-gateway"); err != nil {
		log.Fatalf("Platform Gateway env validation failed: %v", err)
	}

	// Create Gin router
	router := gin.Default()
	router.Use(middleware.CORS())

	// Optional DB integration for platform webhooks.
	cfg := config.Load()
	if _, err := database.InitRedis(cfg); err != nil {
		log.Printf("Redis unavailable: %v", err)
	}
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("Platform Gateway DB unavailable, using fallback mode: %v", err)
	} else {
		platform.SetDB(db)
		log.Println("Platform Gateway connected to PostgreSQL")
	}

	// Initialize Kafka producer (graceful degradation if unavailable)
	var kafkaProducer *kafka.Producer
	if cfg.KafkaBrokers != "" {
		var kafkaErr error
		kafkaProducer, kafkaErr = kafka.NewProducer(cfg.KafkaBrokers)
		if kafkaErr != nil {
			log.Printf("[KAFKA] Platform Gateway Kafka producer unavailable: %v", kafkaErr)
		} else {
			log.Printf("[KAFKA] Platform Gateway Kafka producer connected to %s", cfg.KafkaBrokers)
			defer kafkaProducer.Close()
		}
	}

	// Wire Kafka producer into platform handler
	platform.SetKafkaProducer(kafkaProducer)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "platform-gateway"})
	})
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "service": "platform-gateway", "time": "mock"})
	})
	router.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up", "environment": os.Getenv("INDEL_ENV")})
	})

	// API routes
	routerpkg.SetupPlatformRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("PLATFORM_GATEWAY_PORT")
	}
	if port == "" {
		port = "8003"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Platform Gateway listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
