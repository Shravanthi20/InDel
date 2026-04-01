package models

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Phone     string `gorm:"uniqueIndex"`
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WorkerProfile struct {
	ID                    uint `gorm:"primaryKey"`
	WorkerID              uint `gorm:"uniqueIndex"`
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
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"uniqueIndex"`
	Level      string
	City       string
	State      string
	RiskRating float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Policy struct {
	ID            uint `gorm:"primaryKey"`
	WorkerID      uint
	Status        string
	PremiumAmount float64
	PolicyCycleID uint
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Claim struct {
	ID             uint `gorm:"primaryKey"`
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
	ID           uint      `gorm:"primaryKey"`
	WorkerID     uint      `gorm:"index:idx_worker_day,unique"`
	Date         time.Time `gorm:"index:idx_worker_day,unique"`
	HoursWorked  int
	AmountEarned float64
	CreatedAt    time.Time
}

type Order struct {
	ID              uint `gorm:"primaryKey"`
	WorkerID        uint
	ZoneID          uint
	OrderValue      float64
	Status          string
	PickupArea      string
	DropArea        string
	DistanceKM      float64
	DeliveredAt     *time.Time
	SourceNode      string
	DestinationNode string
	CurrentNode     string
	Route           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type WeeklyPolicyCycle struct {
	ID               uint      `gorm:"primaryKey"`
	CycleID          string    `gorm:"uniqueIndex"`
	WeekStart        time.Time `gorm:"index"`
	WeekEnd          time.Time
	WorkersEvaluated int
	PremiumsComputed int
	PremiumFailures  int
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PremiumPayment struct {
	ID             uint      `gorm:"primaryKey"`
	WorkerID       uint      `gorm:"index"`
	PolicyCycleID  uint      `gorm:"index"`
	Amount         float64
	Status         string
	IdempotencyKey string    `gorm:"uniqueIndex"`
	Date           time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EarningsBaseline struct {
	WorkerID       uint      `gorm:"primaryKey"`
	BaselineAmount float64
	LastUpdatedAt  time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type WeeklyEarningsSummary struct {
	ID            uint      `gorm:"primaryKey"`
	WorkerID      uint      `gorm:"index:idx_worker_week,unique"`
	WeekStart     time.Time `gorm:"index:idx_worker_week,unique"`
	WeekEnd       time.Time
	TotalEarnings float64
	ClaimEligible bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Disruption struct {
	ID              uint `gorm:"primaryKey"`
	ZoneID          uint `gorm:"index"`
	Type            string
	Severity        string
	Confidence      float64
	Status          string
	SignalTimestamp *time.Time
	ConfirmedAt     *time.Time
	StartTime       *time.Time
	EndTime         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Payout struct {
	ID            uint `gorm:"primaryKey"`
	ClaimID        uint `gorm:"uniqueIndex"`
	WorkerID       uint `gorm:"index"`
	Amount         float64
	Status         string
	IdempotencyKey string `gorm:"uniqueIndex"`
	RetryCount     int
	LastError      string
	NextRetryAt    *time.Time
	ProcessedAt    *time.Time
	RazorpayID     string
	RazorpayStatus string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PayoutAttempt struct {
	ID        uint `gorm:"primaryKey"`
	PayoutID   uint `gorm:"index"`
	AttemptNo int
	Status    string
	Error     string
	CreatedAt time.Time
}

type KafkaEventLog struct {
	ID          uint `gorm:"primaryKey"`
	Topic       string `gorm:"index"`
	EventType   string
	PayloadJSON string `gorm:"type:text"`
	CreatedAt   time.Time
}

type SyntheticGenerationRun struct {
	ID                 uint `gorm:"primaryKey"`
	RunID              string `gorm:"uniqueIndex"`
	Seed               int
	Scenario           string
	OutputDir          string
	WorkersCreated     int
	ZonesCreated       int
	DisruptionsCreated int
	ClaimsCreated      int
	PayoutsCreated     int
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
