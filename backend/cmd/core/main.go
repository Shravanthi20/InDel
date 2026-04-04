package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shravanthi20/InDel/backend/internal/config"
	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/Shravanthi20/InDel/backend/internal/handlers/core"
	"github.com/Shravanthi20/InDel/backend/internal/pollers"
	routerpkg "github.com/Shravanthi20/InDel/backend/internal/router"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/Shravanthi20/InDel/backend/pkg/razorpay"
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

	core.SetDB(db)

	// ─── Initialize Razorpay Client ────────────────────────────────────
	razorpayAPIKey := os.Getenv("RAZORPAY_API_KEY")
	razorpayAPISecret := os.Getenv("RAZORPAY_API_SECRET")
	razorpayClient := razorpay.NewRazorpayClient(razorpayAPIKey, razorpayAPISecret)
	log.Printf("✅ Razorpay client initialized (Mock Mode: %v)", razorpayClient.MockMode)

	// ─── Start Automated Disruption Triggers ────────────────────────────
	coreSvc := services.NewCoreOpsService(db)
	coreSvc.SetRazorpayClient(razorpayClient)

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

	log.Println("✅ All 5 disruption triggers started")
	log.Println("✅ Disruption pipeline processor started")
	// ────────────────────────────────────────────────────────────────────

	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

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
