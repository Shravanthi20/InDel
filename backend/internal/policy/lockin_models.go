package policy

import (
	"time"
)

// Policy represents a policy record for lock-in tests
// Extend as needed for your schema

type Policy struct {
	ID              string
	WorkerID        string
	Status          string
	LockInStartTime *time.Time
	LockInEndTime   *time.Time
}

// Disruption represents a disruption record for tests

type Disruption struct {
	ID        string
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Zone      string
}

// ActivePolicy represents an active policy row

type ActivePolicy struct {
	ID       string
	PolicyID string
	WorkerID string
}
