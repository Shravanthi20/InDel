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
	if state.LateFeeINR != 1 {
		t.Fatalf("expected late fee 1 on day 8, got %d", state.LateFeeINR)
	}
}

func TestEvaluatePaymentScheduleExpired(t *testing.T) {
	now := time.Now().UTC()
	last := now.Add(-10 * 24 * time.Hour)

	state := evaluatePaymentSchedule(last, now)

	if state.PaymentStatus != "Deactivated" {
		t.Fatalf("expected Deactivated, got %s", state.PaymentStatus)
	}
	if state.NextPaymentEnabled {
		t.Fatalf("expected next payment disabled after deactivation")
	}
	if state.CoverageStatus != "Deactivated" {
		t.Fatalf("expected Deactivated coverage, got %s", state.CoverageStatus)
	}
}
