package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Phone        string `gorm:"uniqueIndex"`
	Email        string
	PasswordHash string `gorm:"column:password_hash"`
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AuthToken struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	CreatedAt time.Time
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
	IsOnline              bool      `gorm:"column:is_online;default:true"`
	LastActiveAt          time.Time `gorm:"column:last_active_at"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

const WorkerStatusStalenessThreshold = 15 * time.Minute

func (wp WorkerProfile) IsStaleAt(now time.Time) bool {
	return IsWorkerStatusStale(wp.IsOnline, wp.LastActiveAt, now)
}

func (wp WorkerProfile) EffectiveOnlineStatus(now time.Time) bool {
	return EffectiveWorkerOnlineStatus(wp.IsOnline, wp.LastActiveAt, now)
}

func IsWorkerStatusStale(isOnline bool, lastActiveAt, now time.Time) bool {
	if !isOnline {
		return false
	}
	// If marked online but NEVER pinged (zero time), do not treat as stale
	// to avoid penalizing new workers or missing data.
	if lastActiveAt.IsZero() {
		return true
	}
	if now.IsZero() {
		now = time.Now()
	}
	return now.Sub(lastActiveAt) > WorkerStatusStalenessThreshold
}

func EffectiveWorkerOnlineStatus(isOnline bool, lastActiveAt, now time.Time) bool {
	return isOnline && !IsWorkerStatusStale(isOnline, lastActiveAt, now)
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

// Policy status constants
const (
	PolicyStatusLocked   = "locked"
	PolicyStatusActive   = "active"
	PolicyStatusInactive = "inactive"
	PolicyStatusPaused   = "paused"
	PolicyStatusCancelled = "cancelled"
	PolicyStatusExpired  = "expired"
)

type Policy struct {
	ID              uint       `gorm:"primaryKey"`
	WorkerID        uint
	PlanID          string
	Status          string     // locked | active | inactive | paused | cancelled | expired
	PremiumAmount   float64
	PolicyCycleID   uint
	LockInStartTime *time.Time `gorm:"column:lock_in_start_time"`
	LockInEndTime   *time.Time `gorm:"column:lock_in_end_time"`
	IdempotencyKey  string     `gorm:"column:idempotency_key"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// IsInLockIn returns true when the policy is currently in the lock-in window.
func (p *Policy) IsInLockIn(now time.Time) bool {
	if p.Status != PolicyStatusLocked {
		return false
	}
	if p.LockInEndTime == nil {
		return false
	}
	return now.Before(*p.LockInEndTime)
}

// PolicyAuditLog records every lock-in related state transition for fraud analysis.
type PolicyAuditLog struct {
	ID         uint      `gorm:"primaryKey"`
	PolicyID   uint      `gorm:"index"`
	WorkerID   uint
	Action     string    // policy_locked | policy_activated | claim_rejected_lockin | purchase_blocked_disruption
	FromStatus string
	ToStatus   string
	Reason     string
	Metadata   string    `gorm:"type:text"` // JSON
	CreatedAt  time.Time
}

func (PolicyAuditLog) TableName() string {
	return "policy_audit_logs"
}

type Claim struct {
	ID             uint `gorm:"primaryKey"`
	DisruptionID   uint
	WorkerID       uint
	ClaimAmount    float64
	Status         string
	FraudVerdict   string
	ManualReviewAt *time.Time `gorm:"column:manual_reviewed_at"`
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
	SourceNode      string // e.g., "Tambaram"
	DestinationNode string // e.g., "Pondicherry"
	CurrentNode     string // e.g., "Chennai" (updated as order moves)
	Route           string // e.g., "Tambaram,Chennai,Pondicherry" (comma-separated path)
	VehicleType     string // e.g., "bike", "van", "truck"
	VehicleCapacity int    // e.g., 15 (kg)
	AllowedZones    string // comma-separated zone IDs or names
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
	ID             uint `gorm:"primaryKey"`
	WorkerID       uint `gorm:"index"`
	PolicyCycleID  uint `gorm:"index"`
	Amount         float64
	Status         string
	IdempotencyKey string `gorm:"uniqueIndex"`
	Date           time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EarningsBaseline struct {
	WorkerID       uint `gorm:"primaryKey"`
	BaselineAmount float64
	LastUpdatedAt  time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (EarningsBaseline) TableName() string {
	return "earnings_baseline"
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

func (WeeklyEarningsSummary) TableName() string {
	return "weekly_earnings_summary"
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
	ProcessedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Payout struct {
	ID             uint `gorm:"primaryKey"`
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
	PayoutID  uint `gorm:"index"`
	AttemptNo int
	Status    string
	Error     string
	CreatedAt time.Time
}

type KafkaEventLog struct {
	ID          uint   `gorm:"primaryKey"`
	Topic       string `gorm:"index"`
	EventType   string
	PayloadJSON string `gorm:"type:text"`
	CreatedAt   time.Time
}

type SyntheticGenerationRun struct {
	ID                 uint   `gorm:"primaryKey"`
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

type Notification struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	WorkerID  uint       `gorm:"index" json:"worker_id"`
	Type      string     `gorm:"size:50" json:"type"`
	Message   string     `gorm:"type:text" json:"message"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}
