package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/handlers/worker"
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/middleware"
	routerpkg "github.com/Shravanthi20/InDel/backend/internal/router"
	workers "github.com/Shravanthi20/InDel/backend/internal/workers"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.BootstrapServiceEnv("worker-gateway"); err != nil {
		log.Fatalf("Worker Gateway env validation failed: %v", err)
	}

	cfg := config.Load()

	// Create Gin router
	router := gin.Default()
	router.Use(middleware.CORS())

	// Initialize DB and seed minimal worker demo data if available.
	if _, err := database.InitRedis(cfg); err != nil {
		log.Printf("Redis unavailable: %v", err)
	}
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

	// Initialize Kafka producer (graceful degradation if unavailable)
	var kafkaProducer *kafka.Producer
	if cfg.KafkaBrokers != "" {
		var kafkaErr error
		kafkaProducer, kafkaErr = kafka.NewProducer(cfg.KafkaBrokers)
		if kafkaErr != nil {
			log.Printf("[KAFKA] Worker Gateway Kafka producer unavailable: %v", kafkaErr)
		} else {
			log.Printf("[KAFKA] Worker Gateway Kafka producer connected to %s", cfg.KafkaBrokers)
			defer kafkaProducer.Close()
		}
	}

	// Wire Kafka producer into policy handler
	worker.SetKafkaProducer(kafkaProducer)

	// Start PolicyActivationWorker (background goroutine)
	if db != nil {
		ctx := context.Background()
		activationWorker := workers.NewPolicyActivationWorker(db, kafkaProducer)

		// Allow shorter poll interval for demo/testing via env var
		if pollSec := os.Getenv("POLICY_ACTIVATION_POLL_SECONDS"); pollSec != "" {
			var secs int
			if _, scanErr := fmt.Sscan(pollSec, &secs); scanErr == nil && secs > 0 {
				activationWorker.PollInterval = time.Duration(secs) * time.Second
			}
		}

		go activationWorker.Start(ctx)
		log.Printf("[POLICY-WORKER] PolicyActivationWorker started (poll: %s, lock-in: %dh)",
			activationWorker.PollInterval, cfg.PolicyLockInHours)
	}

	// Health check endpoints
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
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("WORKER_GATEWAY_PORT")
	}
	if port == "" {
		port = "8001"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Worker Gateway listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
