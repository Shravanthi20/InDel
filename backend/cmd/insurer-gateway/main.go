package main

import (
	"fmt"
	"log"
	"os"

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

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "insurer-gateway"})
	})

	// API routes
	// GET /api/overview
	// GET /api/loss-ratio
	// GET /api/claims
	// TODO: Implement insurer gateway endpoints

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
