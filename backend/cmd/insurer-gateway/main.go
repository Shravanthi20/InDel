package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/handlers/insurer"
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

	// Optional DB wiring for live aggregate metrics.
	cfg := config.Load()
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Printf("Insurer Gateway DB unavailable, using fallback responses: %v", err)
	} else {
		insurer.SetDB(db)
		log.Println("Insurer Gateway connected to PostgreSQL")
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "insurer-gateway"})
	})

	// API routes
	routerpkg.SetupInsurerRoutes(router)

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
