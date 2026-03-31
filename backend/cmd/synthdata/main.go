package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Simplified DB models for generation
type Zone struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
	City string
}

type WorkerProfile struct {
	WorkerID              uint    `gorm:"primaryKey"`
	ZoneID                uint
	TotalEarningsLifetime float64
}

type Policy struct {
	WorkerID      uint   `gorm:"primaryKey"`
	Status        string
	PremiumAmount float64
}

type PremiumPayment struct {
	ID       uint `gorm:"primaryKey"`
	WorkerID uint
	Amount   float64
	Status   string
	Date     time.Time
}

type Disruption struct {
	ID        uint `gorm:"primaryKey"`
	ZoneID    uint
	Severity  string
	StartTime time.Time
}

type Claim struct {
	ID           uint `gorm:"primaryKey"`
	DisruptionID uint
	WorkerID     uint
	ClaimAmount  float64
	Status       string
	FraudVerdict string
	CreatedAt    time.Time
}

type ClaimFraudScore struct {
	ClaimID        uint `gorm:"primaryKey"`
	Score          float64
	FinalVerdict   string
	RuleViolations string
}

// Generate deterministically seeds the DB
func main() {
	seed := flag.Int("seed", 42, "Fixed seed for reproducible metrics")
	scenarios := flag.String("scenarios", "all", "Scenarios to run")
	flag.Parse()

	log.Printf("Starting data synthesis (seed=%d, scenarios=%s)", *seed, *scenarios)

	// 1. Setup deterministic seed generator
	rand.Seed(int64(*seed)) // Fixed seed for reproducible metrics

	// 2. Setup SQLite DB (or postgres if URL provided)
	os.Remove("synthdata.db")
	db, err := gorm.Open(sqlite.Open("synthdata.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate schemas
	db.AutoMigrate(&Zone{}, &WorkerProfile{}, &Policy{}, &PremiumPayment{}, &Disruption{}, &Claim{}, &ClaimFraudScore{})

	// 3. Generate 8-12 zones
	zones := []Zone{
		{ID: 1, Name: "Tambaram", City: "Chennai"},
		{ID: 2, Name: "Adyar", City: "Chennai"},
		{ID: 3, Name: "Velachery", City: "Chennai"},
		{ID: 4, Name: "Koramangala", City: "Bengaluru"},
		{ID: 5, Name: "Indiranagar", City: "Bengaluru"},
		{ID: 6, Name: "Whitefield", City: "Bengaluru"},
		{ID: 7, Name: "Andheri", City: "Mumbai"},
		{ID: 8, Name: "Bandra", City: "Mumbai"},
		{ID: 9, Name: "Powai", City: "Mumbai"},
		{ID: 10, Name: "Gurgaon Sec 29", City: "Delhi NCR"},
	}
	db.Create(&zones)

	// 4. Generate 500 workers across the zones
	log.Println("Generating workers...")
	for i := 1; i <= 500; i++ {
		z := zones[rand.Intn(len(zones))]
		worker := WorkerProfile{WorkerID: uint(i), ZoneID: z.ID, TotalEarningsLifetime: float64(10000 + rand.Intn(50000))}
		db.Create(&worker)
		// Active policy
		db.Create(&Policy{WorkerID: uint(i), Status: "active", PremiumAmount: 150.0})

		// 8 weeks earnings baselines (approximated as weekly premium payments)
		for w := 0; w < 8; w++ {
			db.Create(&PremiumPayment{
				WorkerID: uint(i),
				Amount:   150.0,
				Status:   "completed",
				Date:     time.Now().AddDate(0, 0, -w*7),
			})
		}
	}

	// 5. Generate 2-3 disruptions per zone
	log.Println("Generating disruptions...")
	disruptionID := 1
	for _, z := range zones {
		numDisruptions := 2 + rand.Intn(2) // 2 or 3
		for d := 0; d < numDisruptions; d++ {
			severity := "mild"
			if rand.Float64() > 0.7 {
				severity = "severe" // Scenario preset logic injection point
			}
			
			db.Create(&Disruption{
				ID:        uint(disruptionID),
				ZoneID:    z.ID,
				Severity:  severity,
				StartTime: time.Now().AddDate(0, 0, -rand.Intn(30)),
			})
			disruptionID++
		}
	}

	// 6. Generate 2,000 claims with 10-15% flagged
	log.Println("Generating claims...")
	for i := 1; i <= 2000; i++ {
		dID := 1 + rand.Intn(disruptionID-1)
		wID := 1 + rand.Intn(500)
		
		isFraud := rand.Float64() < 0.12 // ~12% flagged
		verdict := "clear"
		status := "approved"
		score := rand.Float64() * 0.4
		violations := "[]"

		if isFraud {
			verdict = "flagged"
			status = "manual_review"
			score = 0.7 + rand.Float64()*0.3
			factors := []map[string]interface{}{
				{"name": "gps_mismatch", "impact": 0.24},
				{"name": "session_gap", "impact": 0.12},
			}
			v, _ := json.Marshal(factors)
			violations = string(v)
		} else {
			if rand.Float64() < 0.2 { status = "pending" }
		}

		claimAmount := float64(300 + rand.Intn(700))
		db.Create(&Claim{
			ID:           uint(i),
			DisruptionID: uint(dID),
			WorkerID:     uint(wID),
			ClaimAmount:  claimAmount,
			Status:       status,
			FraudVerdict: verdict,
			CreatedAt:    time.Now().AddDate(0, 0, -rand.Intn(14)),
		})

		db.Create(&ClaimFraudScore{
			ClaimID:        uint(i),
			Score:          score,
			FinalVerdict:   verdict,
			RuleViolations: violations,
		})
	}

	log.Println("Synthetic data generation complete!")
}
