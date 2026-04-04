package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/handlers/core"
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/pollers"
	routerpkg "github.com/Shravanthi20/InDel/backend/internal/router"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && os.Getenv("INDEL_ENV") != "production" {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// ─── Start Automated Disruption Triggers ────────────────────────────
	brokers := os.Getenv("KAFKA_BROKERS")
	var producer *kafka.Producer
	if brokers != "" {
		producer, _ = kafka.NewProducer(brokers)
	}

	// Initialize handlers with producer
	core.SetDBWithProducer(db, producer)

	coreSvc := services.NewCoreOpsService(db, producer)

	// Trigger 1 & 2: Heavy Rain + Extreme Heat (OpenWeatherMap, every 10 min)
	weatherPoller := &pollers.WeatherPoller{DB: db}
	weatherPoller.Start()

	// Trigger 3: Severe Pollution (OpenAQ, every 30 min)
	aqiPoller := &pollers.AQIPoller{DB: db}
	aqiPoller.Start()

	// Trigger 4: Platform Order Drop (internal DB, every 15 min)
	orderDropPoller := &pollers.OrderDropPoller{DB: db}
	orderDropPoller.Start()

	// Trigger 5: Zone Closure / Curfew / Strike (mock gov API, every 60 min)
	zoneClosurePoller := &pollers.ZoneClosurePoller{DB: db}
	zoneClosurePoller.Start()

	// Pipeline Processor: picks up confirmed disruptions → auto-generates claims + payouts
	disruptionProcessor := &pollers.DisruptionProcessor{DB: db, CoreSvc: coreSvc}
	disruptionProcessor.Start()

	// Payout Consumer: Async processing of queued payouts
	if brokers != "" {
		consumer, err := kafka.NewConsumer(brokers, "core-payout-group", []string{kafka.TopicPayoutsQueued})
		if err == nil {
			payoutConsumer := &pollers.PayoutConsumer{Consumer: consumer, CoreSvc: coreSvc}
			payoutConsumer.Start()
			log.Println("✅ Payout consumer started")
		} else {
			log.Printf("⚠️ Failed to start payout consumer: %v", err)
		}
	}

	log.Println("✅ All 5 disruption triggers started")
	log.Println("✅ Disruption pipeline processor started")
	// ────────────────────────────────────────────────────────────────────

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "triggers": "active"})
	})

	routerpkg.SetupCoreRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Core service listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
