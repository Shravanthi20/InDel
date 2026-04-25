package policy

import (
	"context"
	"errors"
	"time"
)

// KafkaEvent is a struct for event payloads
// Add fields as needed for your event schema
// Example:
type KafkaEvent struct {
	Type      string
	PolicyID  string
	WorkerID  string
	Timestamp time.Time
	Reason    string
}

// MockKafkaProducer is a mock for Kafka event publishing
type MockKafkaProducer struct {
	Events []KafkaEvent
	Err    error
}

func (m *MockKafkaProducer) Publish(event KafkaEvent) error {
	if m.Err != nil {
		return m.Err
	}
	m.Events = append(m.Events, event)
	return nil
}

// Test helpers for DB setup, disruption, and policy insertion would go here
// e.g., InsertDisruption, InsertPolicy, etc.

// Example helper for idempotency key enforcement
func IsDuplicateEnrollment(ctx context.Context, workerID string) bool {
	// TODO: Implement DB check for idempotency key
	return false
}

// Example error for lock-in period
var ErrLockInPeriod = errors.New("lock-in period active")
