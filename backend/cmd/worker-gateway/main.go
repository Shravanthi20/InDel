package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/handlers/worker"
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

	// Initialize DB and seed minimal worker demo data if available.
	cfg := config.Load()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("Worker Gateway DB unavailable, using in-memory fallback: %v", err)
	} else {
		worker.SetDB(db)
		if seedErr := worker.EnsureDemoSeed(); seedErr != nil {
			log.Printf("Worker Gateway DB seed warning: %v", seedErr)
		} else {
			log.Println("Worker Gateway connected to PostgreSQL")
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "worker-gateway"})
	})
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "service": "worker-gateway", "time": "mock"})
	})
	router.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up", "environment": "mock"})
	})

	// API routes
	routerpkg.SetupWorkerRoutes(router)

	// Start server
	port := os.Getenv("WORKER_GATEWAY_PORT")
	if port == "" {
		port = "8001"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Worker Gateway listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
