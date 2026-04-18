package events

import "time"

// ClaimDisruptionEvent is the Kafka event schema for claim disruptions
// Topic: claim.disruption.created, claim.disruption.updated
//
type ClaimDisruptionEvent struct {
	EventType      string                 `json:"event_type"`
	ClaimID        string                 `json:"claim_id"`
	DisruptionType string                 `json:"disruption_type"`
	Timestamp      time.Time              `json:"timestamp"`
	Severity       string                 `json:"severity"`
	Metadata       map[string]interface{} `json:"metadata"`
}
