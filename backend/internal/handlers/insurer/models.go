package insurer

import (
	"encoding/json"
	"time"
)

type ClaimFraudScore struct {
	ID             uint            `gorm:"primaryKey"`
	ClaimID        uint            
	Score          float64         
	FinalVerdict   string          
	RuleViolations json.RawMessage `gorm:"type:jsonb"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type FraudFactor struct {
	Name   string  `json:"name"`
	Impact float64 `json:"impact"`
}

type ClaimAuditLog struct {
	ID        uint   `gorm:"primaryKey"`
	ClaimID   uint
	Action    string // "review", "approve", "deny"
	Notes     string
	Reviewer  string
	CreatedAt time.Time
}

// Event structure for domain event buses
type DomainEvent struct {
	EventID    string      `json:"event_id"`
	EventType  string      `json:"event_type"`
	OccurredAt string      `json:"occurred_at"`
	Producer   string      `json:"producer"`
	Payload    interface{} `json:"payload"`
}
