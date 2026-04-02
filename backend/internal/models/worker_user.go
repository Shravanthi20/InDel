package models

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Phone     string `gorm:"unique"`
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WorkerProfile struct {
	ID                    uint
	WorkerID              uint
	Name                  string
	ZoneID                uint
	VehicleType           string
	UPIId                 string
	AQIZone               string
	TotalEarningsLifetime float64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type Zone struct {
	ID         uint
	Name       string
	Level      string // A, B, C, D, E
	City       string
	State      string
	RiskRating float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Policy struct {
	ID            uint
	WorkerID      uint
	Status        string
	PremiumAmount float64
	PolicyCycleID uint
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Claim struct {
	ID             uint
	DisruptionID   uint
	WorkerID       uint
	ClaimAmount    float64
	Status         string
	FraudVerdict   string
	ManualReviewAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EarningsRecord struct {
	ID           uint
	WorkerID     uint
	Date         time.Time
	HoursWorked  int
	AmountEarned float64
	CreatedAt    time.Time
}

type Order struct {
	ID              uint
	WorkerID        uint
	ZoneID          uint
	OrderValue      float64
	SourceNode      string // e.g., "Tambaram"
	DestinationNode string // e.g., "Pondicherry"
	CurrentNode     string // e.g., "Chennai" (updated as order moves)
	Route           string // e.g., "Tambaram,Chennai,Pondicherry" (comma-separated path)
	VehicleType     string // e.g., "bike", "van", "truck"
	VehicleCapacity int    // e.g., 15 (kg)
	AllowedZones    string // comma-separated zone IDs or names
	CreatedAt       time.Time
}
