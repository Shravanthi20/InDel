package claimeval

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/clients"
)

const fallbackFraudScore = 0.10

type fraudRequest struct {
	ClaimID                    int                `json:"claim_id"`
	WorkerID                   int                `json:"worker_id"`
	ZoneID                     int                `json:"zone_id"`
	ClaimAmount                float64            `json:"claim_amount"`
	BaselineEarnings           float64            `json:"baseline_earnings"`
	DisruptionType             string             `json:"disruption_type"`
	DisruptionHours            float64            `json:"disruption_hours"`
	GPSInZone                  bool               `json:"gps_in_zone"`
	DistanceFromZoneCenter     float64            `json:"distance_from_zone_center"`
	DeliveriesDuringDisruption int                `json:"deliveries_during_disruption"`
	ZoneAvgClaimAmount         float64            `json:"zone_avg_claim_amount"`
	ZoneRiskScore              float64            `json:"zone_risk_score"`
	WorkerHistory              fraudWorkerHistory `json:"worker_history"`
}

type fraudWorkerHistory struct {
	TotalClaimsLast8Weeks    int     `json:"total_claims_last_8_weeks"`
	ApprovedClaimsLast8Weeks int     `json:"approved_claims_last_8_weeks"`
	AvgClaimAmount           float64 `json:"avg_claim_amount"`
	EarningsVariance         float64 `json:"earnings_variance"`
	ZoneChangeCount          int     `json:"zone_change_count"`
	DaysActive               int     `json:"days_active"`
	DeliveryAttemptRate      float64 `json:"delivery_attempt_rate"`
}

type fraudResponse struct {
	FraudScore float64       `json:"fraud_score"`
	Verdict    string        `json:"verdict"`
	Signals    []FraudSignal `json:"signals"`
}

// FetchFraudScore integrates with the existing Python service and falls back
// to a low-risk score when the service is unavailable.
func FetchFraudScore(ctx context.Context, activity WorkerActivity) (float64, []FraudSignal, bool) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("FRAUD_ML_MOCK_ERROR")), "true") {
		return fallbackFraudScore, []FraudSignal{{
			Name:        "fraud_service_mock_error",
			Impact:      0.05,
			Description: "Fraud ML failure simulated for test mode.",
		}}, true
	}
	if raw := strings.TrimSpace(os.Getenv("FRAUD_ML_MOCK_SCORE")); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			return clamp(parsed, 0, 1), []FraudSignal{{
				Name:        "fraud_service_mock_score",
				Impact:      clamp(parsed, 0, 1),
				Description: "Fraud ML score simulated for test mode.",
			}}, false
		}
	}

	payload := buildFraudRequest(activity)
	client := clients.NewMLClientFromEnv()
	var decoded fraudResponse
	err := client.GetFraudScore(withCorrelation(ctx, activity.WorkerID), payload, &decoded)
	if err != nil {
		return fallbackFraudScore, []FraudSignal{{
			Name:        "fraud_service_timeout",
			Impact:      0.05,
			Description: "Fraud ML call failed. Falling back to low-risk score.",
		}}, true
	}
	return clamp(decoded.FraudScore, 0, 1), decoded.Signals, false
}

func withCorrelation(ctx context.Context, workerID uint) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if existing, ok := ctx.Value("request_id").(string); ok && strings.TrimSpace(existing) != "" {
		return ctx
	}
	return context.WithValue(ctx, "request_id", fmt.Sprintf("fraud-%d", workerID))
}

func buildFraudRequest(activity WorkerActivity) fraudRequest {
	loss := ComputeLoss(activity)
	expected := maxFloat(activity.EarningsExpected, loss)
	attemptRate := 0.0
	if activity.OrdersAttempted > 0 {
		attemptRate = float64(activity.OrdersCompleted) / float64(activity.OrdersAttempted)
	}

	daysActive := 30
	if !activity.ActiveBefore {
		daysActive = 3
	}

	zoneRisk := 0.45
	if activity.OrdersCompleted > 2 && loss > activity.EarningsExpected*0.5 {
		zoneRisk = 0.62
	}

	// The fraud service expects numeric zone metadata, but our contract remains
	// zone-string based. We keep the scoring anchored in economic behaviour and
	// use safe defaults for the extra fields.
	return fraudRequest{
		ClaimID:                    int(activity.WorkerID),
		WorkerID:                   int(activity.WorkerID),
		ZoneID:                     1,
		ClaimAmount:                round2(loss),
		BaselineEarnings:           round2(expected),
		DisruptionType:             "economic_loss",
		DisruptionHours:            clamp(activity.LoginDuration, 1, 12),
		GPSInZone:                  true,
		DistanceFromZoneCenter:     0.5,
		DeliveriesDuringDisruption: activity.OrdersCompleted,
		ZoneAvgClaimAmount:         round2(expected),
		ZoneRiskScore:              zoneRisk,
		WorkerHistory: fraudWorkerHistory{
			TotalClaimsLast8Weeks:    0,
			ApprovedClaimsLast8Weeks: 0,
			AvgClaimAmount:           round2(loss),
			EarningsVariance:         clamp(safeDiv(loss, maxFloat(expected, 1)), 0, 1),
			ZoneChangeCount:          0,
			DaysActive:               daysActive,
			DeliveryAttemptRate:      clamp(attemptRate, 0, 1),
		},
	}
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
