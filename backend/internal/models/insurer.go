package models

import "time"

// InsurerOverview represents the high-level dashboard metrics
type InsurerOverview struct {
	TotalWorkers       float64 `json:"total_workers"`
	ActiveWorkers      float64 `json:"active_workers"`
	PendingClaims      float64 `json:"pending_claims"`
	ApprovedClaims     float64 `json:"approved_claims"`
	LossRatio          float64 `json:"loss_ratio"`
	Reserve            float64 `json:"reserve"`
	ReserveUtilization float64 `json:"reserve_utilization"`
}

// LossRatio represents the loss-ratio grouped by a zone 
type LossRatio struct {
	ZoneID    uint    `json:"zone_id"`
	ZoneName  string  `json:"zone_name"`
	City      string  `json:"city"`
	Premiums  float64 `json:"premiums"`
	Claims    float64 `json:"claims"`
	LossRatio float64 `json:"loss_ratio"`
}

// ClaimListItem is a summary representation of a claim
type ClaimListItem struct {
	ClaimID      uint      `json:"claim_id"`
	DisruptionID uint      `json:"disruption_id"`
	WorkerID     uint      `json:"worker_id"`
	ZoneName     string    `json:"zone_name"`
	ClaimAmount  float64   `json:"claim_amount"`
	Status       string    `json:"status"`
	FraudVerdict string    `json:"fraud_verdict"`
	CreatedAt    time.Time `json:"created_at"`
}

// FraudFactor is a discrete reason for a fraud score
type FraudFactor struct {
	Name   string  `json:"name"`
	Impact float64 `json:"impact"`
}

// ClaimDetail is an expanded view of a claim including algorithmic scores
type ClaimDetail struct {
	ClaimID           string        `json:"claim_id"`
	WorkerID          string        `json:"worker_id"`
	ZoneID            string        `json:"zone_id"`
	DisruptionID      string        `json:"disruption_id"`
	LossAmount        float64       `json:"loss_amount"`
	RecommendedPayout float64       `json:"recommended_payout"`
	Status            string        `json:"status"`
	FraudVerdict      string        `json:"fraud_verdict"`
	FraudScore        float64       `json:"fraud_score"`
	Factors           []FraudFactor `json:"factors"`
	CreatedAt         string        `json:"created_at"`
}

// FraudQueueItem represents an item specifically sitting in the ML fraud queue
type FraudQueueItem struct {
	ClaimID      uint    `json:"claim_id"`
	WorkerID     uint    `json:"worker_id"`
	Status       string  `json:"status"`
	FraudVerdict string  `json:"fraud_verdict"`
	FraudScore   float64 `json:"fraud_score"`
	CreatedAt    string  `json:"created_at"`
}

// ClaimAction represents the user input from a manual review
type ClaimAction struct {
	Status       string  `json:"status" binding:"required"`
	FraudVerdict string  `json:"fraud_verdict" binding:"required"`
	Notes        string  `json:"notes"`
}

// --- DB Models representing relational concepts specifically ---

type ClaimFraudScore struct {
	ClaimID             uint      `gorm:"primaryKey;column:claim_id"`
	IsolationForestScore float64   `gorm:"column:isolation_forest_score"`
	DbscanScore          float64   `gorm:"column:dbscan_score"`
	FinalVerdict         string    `gorm:"column:final_verdict"`
	RuleViolations       string    `gorm:"column:rule_violations"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

type ClaimAuditLog struct {
	ID        uint      `gorm:"primaryKey"`
	ClaimID   uint
	Action    string
	Notes     string
	Reviewer  string
	CreatedAt time.Time
}
