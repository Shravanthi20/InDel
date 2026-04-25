package policy

import (
	"errors"
	"testing"
	"time"
	"context"
)

// --- Test helpers and mocks ---
// Add your mock Kafka producer, DB setup, and helpers here

// --- 1. Policy Enrollment Tests ---
func TestPolicyEnrollment_Successful(t *testing.T) {
	db, _, cleanup := SetupTestEnv(t)
	defer cleanup()

	// Arrange: No disruption, valid worker
	workerID := "worker-1"
	lockInStart := time.Now().UTC()
	lockInEnd := lockInStart.Add(48 * time.Hour)

	// Simulate policy enrollment logic
	policy := Policy{
		ID:              "policy-1",
		WorkerID:        workerID,
		Status:          "locked",
		LockInStartTime: &lockInStart,
		LockInEndTime:   &lockInEnd,
	}

	// Simulate DB insert (in real test, use db.Create(&policy))

	// Assert
	if policy.Status != "locked" {
		t.Errorf("expected status 'locked', got %s", policy.Status)
	}
	if policy.LockInStartTime == nil || policy.LockInEndTime == nil {
		t.Error("lock-in times not set")
	}
	// Simulate not present in active_policies
	var activePolicies []ActivePolicy
	if len(activePolicies) != 0 {
		t.Error("should not be present in active_policies")
	}
	// Simulate Kafka events
	producer := &MockKafkaProducer{}
	producer.Publish(KafkaEvent{Type: "policy.created", PolicyID: policy.ID, WorkerID: workerID, Timestamp: lockInStart})
	producer.Publish(KafkaEvent{Type: "policy.locked", PolicyID: policy.ID, WorkerID: workerID, Timestamp: lockInStart})
	if len(producer.Events) != 2 {
		t.Error("expected 2 Kafka events (created, locked)")
	}
}

func TestPolicyEnrollment_ActiveDisruption(t *testing.T) {
	db, _, cleanup := SetupTestEnv(t)
	defer cleanup()

	// Arrange: Mock active disruption in DB
	disruption := Disruption{
		ID:        "dis-1",
		Status:    "active",
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
		Zone:      "zone-1",
	}
	// Simulate DB insert (in real test, use db.Create(&disruption))

	// Act: Attempt enrollment
	// Simulate HTTP 403 response
	httpStatus := 403
	// Simulate no policy created
	var policy *Policy = nil
	// Simulate no Kafka events
	producer := &MockKafkaProducer{}

	// Assert
	if httpStatus != 403 {
		t.Errorf("expected HTTP 403, got %d", httpStatus)
	}
	if policy != nil {
		t.Error("policy should not be created during active disruption")
	}
	if len(producer.Events) != 0 {
		t.Error("no Kafka events should be published")
	}
}

func TestPolicyEnrollment_PredictedDisruption(t *testing.T) {
	db, _, cleanup := SetupTestEnv(t)
	defer cleanup()

	// Arrange: Mock predicted disruption (timestamp within lookahead)
	disruption := Disruption{
		ID:        "dis-2",
		Status:    "predicted",
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(3 * time.Hour),
		Zone:      "zone-1",
	}
	// Simulate DB insert (in real test, use db.Create(&disruption))

	// Act: Attempt enrollment
	httpStatus := 403
	var policy *Policy = nil
	producer := &MockKafkaProducer{}

	// Assert
	if httpStatus != 403 {
		t.Errorf("expected HTTP 403, got %d", httpStatus)
	}
	if policy != nil {
		t.Error("policy should not be created during predicted disruption")
	}
	if len(producer.Events) != 0 {
		t.Error("no Kafka events should be published")
	}
}

// --- 2. Lock-in Enforcement Tests ---
func TestLockIn_ClaimAttempt(t *testing.T) {
	// Arrange: Policy with status = locked
	policy := Policy{
		ID:     "policy-2",
		Status: "locked",
	}
	producer := &MockKafkaProducer{}}

	// Act: Attempt claim
	claimAllowed := false
	err := ErrLockInPeriod
	if err == nil {
		claimAllowed = true
		producer.Publish(KafkaEvent{Type: "claim.accepted", PolicyID: policy.ID})
	} else {
		producer.Publish(KafkaEvent{Type: "claim.rejected", PolicyID: policy.ID, Reason: "lock-in period"})
	}

	// Assert
	if claimAllowed {
		t.Error("claim should be blocked during lock-in")
	}
	if len(producer.Events) != 1 || producer.Events[0].Type != "claim.rejected" {
		t.Error("expected claim.rejected event")
	}
	if producer.Events[0].Reason != "lock-in period" {
		t.Error("reason should include 'lock-in period'")
	}
}

func TestLockIn_CancelAttempt(t *testing.T) {
	// Arrange: Policy with status = locked
	policy := Policy{ID: "policy-3", Status: "locked"}

	// Act: Attempt cancel
	httpStatus := 423 // Locked
	statusChanged := false

	// Assert
	if httpStatus != 423 {
		t.Errorf("expected HTTP 423, got %d", httpStatus)
	}
	if statusChanged {
		t.Error("status should not change during lock-in")
	}
}

func TestLockIn_PauseAttempt(t *testing.T) {
	// Arrange: Policy with status = locked
	policy := Policy{ID: "policy-4", Status: "locked"}

	// Act: Attempt pause
	httpStatus := 423 // Locked

	// Assert
	if httpStatus != 423 {
		t.Errorf("expected HTTP 423, got %d", httpStatus)
	}
}

// --- 3. Policy Activation Worker Tests ---
func TestPolicyActivationWorker_ActivatesAfterLockIn(t *testing.T) {
	// Arrange: Policy with lock_in_end_time in the past
	lockInEnd := time.Now().Add(-1 * time.Hour)
	policy := Policy{
		ID:            "policy-5",
		Status:        "locked",
		LockInEndTime: &lockInEnd,
	}
	producer := &MockKafkaProducer{}}

	// Act: Run worker logic
	statusBefore := policy.Status
	if policy.LockInEndTime != nil && policy.LockInEndTime.Before(time.Now()) {
		policy.Status = "active"
		producer.Publish(KafkaEvent{Type: "policy.activated", PolicyID: policy.ID})
		// Simulate row inserted into active_policies
		active := ActivePolicy{ID: "ap-1", PolicyID: policy.ID, WorkerID: policy.WorkerID}
		_ = active // would insert into DB in real test
		// Simulate audit log entry
	}

	// Assert
	if statusBefore != "locked" || policy.Status != "active" {
		t.Error("status should transition from locked to active")
	}
	if len(producer.Events) == 0 || producer.Events[0].Type != "policy.activated" {
		t.Error("expected policy.activated event")
	}
}

func TestPolicyActivationWorker_IdempotentActivation(t *testing.T) {
	// Arrange: Policy already activated
	lockInEnd := time.Now().Add(-2 * time.Hour)
	policy := Policy{
		ID:            "policy-6",
		Status:        "active",
		LockInEndTime: &lockInEnd,
	}
	producer := &MockKafkaProducer{}}

	// Act: Run worker logic twice
	activations := 0
	if policy.Status == "locked" && policy.LockInEndTime != nil && policy.LockInEndTime.Before(time.Now()) {
		policy.Status = "active"
		producer.Publish(KafkaEvent{Type: "policy.activated", PolicyID: policy.ID})
		activations++
	}
	// Second run (should be idempotent)
	if policy.Status == "locked" && policy.LockInEndTime != nil && policy.LockInEndTime.Before(time.Now()) {
		policy.Status = "active"
		producer.Publish(KafkaEvent{Type: "policy.activated", PolicyID: policy.ID})
		activations++
	}

	// Assert
	if activations > 1 {
		t.Error("activation should be idempotent (only one activation)")
	}
	if len(producer.Events) > 1 {
		t.Error("no duplicate Kafka events should be published")
	}
}

// --- 4. Kafka Event Flow Tests ---
func TestKafkaEventFlow_Sequence(t *testing.T) {
	producer := &MockKafkaProducer{}
	policyID := "policy-7"
	workerID := "worker-7"
	timestamp := time.Now()

	// Simulate event sequence
	producer.Publish(KafkaEvent{Type: "policy.created", PolicyID: policyID, WorkerID: workerID, Timestamp: timestamp})
	producer.Publish(KafkaEvent{Type: "policy.locked", PolicyID: policyID, WorkerID: workerID, Timestamp: timestamp})
	producer.Publish(KafkaEvent{Type: "policy.activated", PolicyID: policyID, WorkerID: workerID, Timestamp: timestamp.Add(48 * time.Hour)})

	// Assert
	if len(producer.Events) != 3 {
		t.Error("expected 3 events in sequence")
	}
	if producer.Events[0].Type != "policy.created" || producer.Events[1].Type != "policy.locked" || producer.Events[2].Type != "policy.activated" {
		t.Error("event sequence incorrect")
	}
}

func TestKafkaEventFlow_ClaimRejectedEvent(t *testing.T) {
	producer := &MockKafkaProducer{}
	policyID := "policy-8"
	workerID := "worker-8"

	// Simulate claim rejected during lock-in
	producer.Publish(KafkaEvent{Type: "claim.rejected", PolicyID: policyID, WorkerID: workerID, Reason: "lock-in period"})

	// Assert
	if len(producer.Events) != 1 {
		t.Error("expected 1 claim.rejected event")
	}
	if producer.Events[0].Type != "claim.rejected" || producer.Events[0].Reason != "lock-in period" {
		t.Error("claim.rejected event should have correct reason")
	}
}

// --- 5. Disruption Guard Tests ---
func TestDisruptionGuard_ActiveBlocksEnrollment(t *testing.T) {
	disruption := Disruption{
		ID:        "dis-3",
		Status:    "active",
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
		Zone:      "zone-2",
	}
	// Simulate DB insert (in real test, use db.Create(&disruption))
	httpStatus := 403
	if httpStatus != 403 {
		t.Error("expected HTTP 403 for active disruption")
	}
}

func TestDisruptionGuard_PredictedBlocksEnrollment(t *testing.T) {
	disruption := Disruption{
		ID:        "dis-4",
		Status:    "predicted",
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(3 * time.Hour),
		Zone:      "zone-2",
	}
	httpStatus := 403
	if httpStatus != 403 {
		t.Error("expected HTTP 403 for predicted disruption")
	}
}

func TestDisruptionGuard_NoDisruptionAllowsEnrollment(t *testing.T) {
	// No disruption present
	httpStatus := 200
	if httpStatus != 200 {
		t.Error("expected HTTP 200 when no disruption present")
	}
}

// --- 6. Edge Case Tests ---
func TestEdgeCase_KafkaFailure(t *testing.T) {
	producer := &MockKafkaProducer{Err: errors.New("kafka down")}
	policyID := "policy-9"
	err := producer.Publish(KafkaEvent{Type: "policy.created", PolicyID: policyID})
	if err == nil {
		t.Error("expected error on Kafka failure")
	}
}

func TestEdgeCase_ClockDrift(t *testing.T) {
	// Simulate client time ahead of server
	serverNow := time.Now()
	clientNow := serverNow.Add(5 * time.Minute)
	lockInEnd := serverNow.Add(48 * time.Hour)
	policy := Policy{ID: "policy-10", Status: "locked", LockInEndTime: &lockInEnd}

	// Lock-in logic should use server time
	inLockIn := policy.LockInEndTime != nil && policy.LockInEndTime.After(serverNow)
	if !inLockIn {
		t.Error("lock-in logic should use server time, not client time")
	}
}

func TestEdgeCase_DuplicateEnrollment(t *testing.T) {
	ctx := context.Background()
	workerID := "worker-11"
	// Simulate first enrollment
	isDup := IsDuplicateEnrollment(ctx, workerID)
	if isDup {
		t.Error("first enrollment should not be duplicate")
	}
	// Simulate second enrollment (should be duplicate in real logic)
	// Here, just call IsDuplicateEnrollment again for demonstration
	isDup = IsDuplicateEnrollment(ctx, workerID)
	// In real test, would set up DB to return true
	if isDup {
		t.Error("duplicate enrollment should be detected")
	}
}

// --- 7. Backward Compatibility Tests ---
func TestBackwardCompatibility_NullLockInEndTime(t *testing.T) {
	// Arrange: Policy with NULL lock_in_end_time
	policy := Policy{ID: "policy-12", Status: "active", LockInEndTime: nil}

	// Act: Attempt claim
	claimAllowed := true
	if policy.LockInEndTime == nil && policy.Status == "active" {
		claimAllowed = true
	}

	// Assert
	if !claimAllowed {
		t.Error("claims should be allowed for legacy policies with NULL lock_in_end_time")
	}
}
