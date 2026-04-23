package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/handlers/platform"
	"github.com/Shravanthi20/InDel/backend/internal/middleware"
	routerpkg "github.com/Shravanthi20/InDel/backend/internal/router"
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
