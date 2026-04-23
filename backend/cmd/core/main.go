package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/pollers"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/handlers/core"
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

	// Load configuration
	cfg := config.Load()

	// Initialize Redis
	if _, err := database.InitRedis(cfg); err != nil {
		log.Printf("Redis unavailable: %v", err)
	}

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Start keep-alive poller to ping backend services every 5 minutes
	keepAlive := &pollers.KeepAlivePoller{
		ServiceURLs: resolveKeepAliveURLs(),
	}
	keepAlive.Start()

	core.SetDB(db)

	// Create Gin router
	router := gin.Default()
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	routerpkg.SetupCoreRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Core service listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func resolveKeepAliveURLs() []string {
	// Highest priority: explicit comma-separated list of full health URLs.
	if raw := strings.TrimSpace(os.Getenv("KEEPALIVE_URLS")); raw != "" {
		parts := strings.Split(raw, ",")
		urls := make([]string, 0, len(parts))
		for _, part := range parts {
			if url := strings.TrimSpace(part); url != "" {
				urls = append(urls, url)
			}
		}
		if len(urls) > 0 {
			return urls
		}
	}

	// Next: known public base URLs, then append /health.
	for _, key := range []string{"PUBLIC_CORE_URL", "RENDER_EXTERNAL_URL"} {
		if base := strings.TrimSpace(os.Getenv(key)); base != "" {
			return []string{strings.TrimRight(base, "/") + "/health"}
		}
	}

	// Last-resort fallback.
	return []string{"https://indel-backend.onrender.com/health"}
}
