package worker

import (
	"testing"
	"time"
)

func TestEvaluatePaymentScheduleLocked(t *testing.T) {
	now := time.Now().UTC()
	last := now.Add(-3 * 24 * time.Hour)

	state := evaluatePaymentSchedule(last, now)

	if state.PaymentStatus != "Locked" {
		t.Fatalf("expected Locked, got %s", state.PaymentStatus)
	}
	if state.NextPaymentEnabled {
		t.Fatalf("expected next payment disabled during lock period")
	}
	if state.CoverageStatus != "Active" {
		t.Fatalf("expected Active coverage, got %s", state.CoverageStatus)
	}
}

func TestEvaluatePaymentScheduleEligible(t *testing.T) {
	now := time.Now().UTC()
	last := now.Add(-8 * 24 * time.Hour)

	state := evaluatePaymentSchedule(last, now)

	if state.PaymentStatus != "Eligible" {
		t.Fatalf("expected Eligible, got %s", state.PaymentStatus)
	}
	if !state.NextPaymentEnabled {
		t.Fatalf("expected next payment enabled in payment window")
	}
	if state.CoverageStatus != "Active" {
		t.Fatalf("expected Active coverage, got %s", state.CoverageStatus)
	}
}

func TestEvaluatePaymentScheduleExpired(t *testing.T) {
	now := time.Now().UTC()
	last := now.Add(-16 * 24 * time.Hour)

	state := evaluatePaymentSchedule(last, now)

	if state.PaymentStatus != "Expired" {
		t.Fatalf("expected Expired, got %s", state.PaymentStatus)
	}
	if !state.NextPaymentEnabled {
		t.Fatalf("expected next payment enabled for restart")
	}
	if state.CoverageStatus != "Expired" {
		t.Fatalf("expected Expired coverage, got %s", state.CoverageStatus)
	}
}
