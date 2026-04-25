package events

import "time"

// PolicyCreatedEvent is published when a new policy is created (status=locked).
// Topic: policy.created
type PolicyCreatedEvent struct {
	EventType      string    `json:"event_type"` // "policy.created"
	PolicyID       uint      `json:"policy_id"`
	WorkerID       uint      `json:"worker_id"`
	ZoneID         uint      `json:"zone_id"`
	PremiumAmount  float64   `json:"premium_amount"`
	LockInStart    time.Time `json:"lock_in_start"`
	LockInEnd      time.Time `json:"lock_in_end"`
	IdempotencyKey string    `json:"idempotency_key"`
	Timestamp      time.Time `json:"timestamp"`
}

// PolicyLockedEvent is published immediately after policy.created to confirm lock-in status.
// Topic: policy.locked
type PolicyLockedEvent struct {
	EventType   string    `json:"event_type"` // "policy.locked"
	PolicyID    uint      `json:"policy_id"`
	WorkerID    uint      `json:"worker_id"`
	LockInEnd   time.Time `json:"lock_in_end"`
	LockInHours int       `json:"lock_in_hours"`
	Timestamp   time.Time `json:"timestamp"`
}

// PolicyActivatedEvent is published when a LOCKED policy transitions to ACTIVE
// after the lock-in window expires.
// Topic: policy.activated
type PolicyActivatedEvent struct {
	EventType   string    `json:"event_type"` // "policy.activated"
	PolicyID    uint      `json:"policy_id"`
	WorkerID    uint      `json:"worker_id"`
	ActivatedAt time.Time `json:"activated_at"`
	Timestamp   time.Time `json:"timestamp"`
}

// ClaimRejectedEvent is published when a claim is denied because the policy
// is still in lock-in period.
// Topic: claim.rejected
type ClaimRejectedEvent struct {
	EventType  string    `json:"event_type"` // "claim.rejected"
	WorkerID   uint      `json:"worker_id"`
	PolicyID   uint      `json:"policy_id"`
	Reason     string    `json:"reason"`    // "policy_in_lockin_period"
	LockInEnd  time.Time `json:"lock_in_end"`
	Timestamp  time.Time `json:"timestamp"`
}

// PurchaseBlockedEvent is published when a policy purchase is blocked due to
// an active or predicted disruption in the worker's zone.
// Topic: policy.purchase_blocked
type PurchaseBlockedEvent struct {
	EventType       string    `json:"event_type"` // "policy.purchase_blocked"
	WorkerID        uint      `json:"worker_id"`
	ZoneID          uint      `json:"zone_id"`
	DisruptionID    uint      `json:"disruption_id,omitempty"`
	Reason          string    `json:"reason"` // "active_disruption" or "predicted_disruption"
	DisruptionType  string    `json:"disruption_type,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}

// DisruptionSignalEvent is published when the system detects or confirms a disruption,
// allowing downstream consumers (e.g. policy service) to block purchases.
// Topic: disruption.active | disruption.predicted
type DisruptionSignalEvent struct {
	EventType      string                 `json:"event_type"` // "disruption.active" or "disruption.predicted"
	DisruptionID   uint                   `json:"disruption_id"`
	ZoneID         uint                   `json:"zone_id"`
	ZoneName       string                 `json:"zone_name"`
	DisruptionType string                 `json:"disruption_type"`
	Severity       string                 `json:"severity"`
	PredictedAt    *time.Time             `json:"predicted_at,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
