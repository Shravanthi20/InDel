package platform

import "testing"

func TestApplyProgressivePayout_IncrementsOnlyOnRiskIncrease(t *testing.T) {
	ResetEngineForTests()
	defer ResetEngineForTests()

	first := applyProgressivePayout(11, ProgressivePayoutInputs{
		AQI:           200,
		Temperature:   35,
		Rain:          0,
		Traffic:       60,
		MaxPayoutDay:  1000,
		CoverageRatio: 0.5,
	})
	if first.TriggerStatus != "No payout" {
		t.Fatalf("expected no payout on the initial flat-risk reading, got %q", first.TriggerStatus)
	}
	if first.CurrentRiskScore != 0 {
		t.Fatalf("expected zero risk at baseline, got %.2f", first.CurrentRiskScore)
	}

	second := applyProgressivePayout(11, ProgressivePayoutInputs{
		AQI:           320,
		Temperature:   45,
		Rain:          15,
		Traffic:       95,
		MaxPayoutDay:  1000,
		CoverageRatio: 0.5,
	})
	if second.TriggerStatus != "Incremental payout" {
		t.Fatalf("expected incremental payout on risk increase, got %q", second.TriggerStatus)
	}
	if second.CurrentRiskScore <= first.CurrentRiskScore {
		t.Fatalf("expected risk to increase, got %.2f -> %.2f", first.CurrentRiskScore, second.CurrentRiskScore)
	}
	if second.FinalPayout != 500 {
		t.Fatalf("expected payout of 500, got %.2f", second.FinalPayout)
	}
	if second.TotalPayoutSoFar != 500 {
		t.Fatalf("expected cumulative payout of 500, got %.2f", second.TotalPayoutSoFar)
	}
	if second.IncrementalRisk != 1 {
		t.Fatalf("expected incremental risk of 1.00, got %.2f", second.IncrementalRisk)
	}
}

func TestApplyProgressivePayout_MaxCoverageReached(t *testing.T) {
	ResetEngineForTests()
	defer ResetEngineForTests()

	state := getOrCreateProgressivePayoutState(27)
	state.mu.Lock()
	state.LastRiskScore = 0.95
	state.LastPayout = 500
	state.mu.Unlock()

	result := applyProgressivePayout(27, ProgressivePayoutInputs{
		AQI:           320,
		Temperature:   45,
		Rain:          15,
		Traffic:       95,
		MaxPayoutDay:  1000,
		CoverageRatio: 0.5,
	})

	if result.TriggerStatus != "Max coverage reached" {
		t.Fatalf("expected max coverage reached, got %q", result.TriggerStatus)
	}
	if result.FinalPayout != 0 {
		t.Fatalf("expected no remaining payout, got %.2f", result.FinalPayout)
	}
	if result.TotalPayoutSoFar != 500 {
		t.Fatalf("expected cumulative payout to remain at 500, got %.2f", result.TotalPayoutSoFar)
	}
}
