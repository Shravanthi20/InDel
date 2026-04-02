package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CoreOpsService struct {
	DB *gorm.DB
}

func NewCoreOpsService(db *gorm.DB) *CoreOpsService {
	return &CoreOpsService{DB: db}
}

type WeeklyCycleResult struct {
	CycleID          string `json:"cycle_id"`
	WorkersEvaluated int    `json:"workers_evaluated"`
	PremiumsComputed int    `json:"premiums_computed"`
	PremiumFailures  int    `json:"premium_failures"`
	Status           string `json:"status"`
}

type GeneratedClaimsResult struct {
	DisruptionID    string `json:"disruption_id"`
	WorkersChecked  int    `json:"workers_checked"`
	ClaimsGenerated int    `json:"claims_generated"`
	ClaimsSkipped   int    `json:"claims_skipped"`
	Status          string `json:"status"`
}

type PayoutResult struct {
	PayoutID       string  `json:"payout_id"`
	ClaimID        string  `json:"claim_id"`
	WorkerID       string  `json:"worker_id"`
	AmountINR      float64 `json:"amount_inr"`
	Status         string  `json:"status"`
	IdempotencyKey string  `json:"idempotency_key"`
	RetryCount     int     `json:"retry_count,omitempty"`
}

type ProcessPayoutsResult struct {
	Processed int `json:"processed"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Retried   int `json:"retried"`
}

type ReconciliationResult struct {
	From             string             `json:"from"`
	To               string             `json:"to"`
	Totals           map[string]float64 `json:"totals"`
	Counts           map[string]int     `json:"counts"`
	ReconciliationOK bool               `json:"reconciliation_ok"`
	MismatchCount    int                `json:"mismatch_count"`
}

type SyntheticGenerateRequest struct {
	Seed      int    `json:"seed"`
	Scenario  string `json:"scenario"`
	OutputDir string `json:"output_dir"`
}

type SyntheticGenerateResult struct {
	RunID       string            `json:"run_id"`
	Seed        int               `json:"seed"`
	Scenario    string            `json:"scenario"`
	Status      string            `json:"status"`
	Counts      map[string]int    `json:"counts"`
	Artifacts   map[string]string `json:"artifacts"`
	Integration map[string]string `json:"integration,omitempty"`
}

func (s *CoreOpsService) RunWeeklyCycle(now time.Time) (*WeeklyCycleResult, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("db unavailable")
	}

	weekStart, weekEnd := weekBounds(now.UTC())
	cycleID := cycleIDForDate(weekStart)

	var cycle models.WeeklyPolicyCycle
	err := s.DB.Where("cycle_id = ?", cycleID).First(&cycle).Error
	if err == nil && cycle.Status == "completed" {
		return &WeeklyCycleResult{
			CycleID:          cycle.CycleID,
			WorkersEvaluated: cycle.WorkersEvaluated,
			PremiumsComputed: cycle.PremiumsComputed,
			PremiumFailures:  cycle.PremiumFailures,
			Status:           cycle.Status,
		}, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err == gorm.ErrRecordNotFound {
		cycle = models.WeeklyPolicyCycle{CycleID: cycleID, WeekStart: weekStart, WeekEnd: weekEnd, Status: "running"}
		if err := s.DB.Create(&cycle).Error; err != nil {
			return nil, err
		}
	} else {
		cycle.Status = "running"
		cycle.WeekStart = weekStart
		cycle.WeekEnd = weekEnd
		if err := s.DB.Save(&cycle).Error; err != nil {
			return nil, err
		}
	}

	type cycleWorker struct {
		WorkerID       uint
		ZoneID         uint
		RiskRating     float64
		VehicleType    string
		BaselineAmount float64
	}

	var workers []cycleWorker
	if err := s.DB.Table("policies p").
		Select("p.worker_id, wp.zone_id, z.risk_rating, wp.vehicle_type, eb.baseline_amount").
		Joins("LEFT JOIN worker_profiles wp ON wp.worker_id = p.worker_id").
		Joins("LEFT JOIN zones z ON z.id = wp.zone_id").
		Joins("LEFT JOIN earnings_baselines eb ON eb.worker_id = p.worker_id").
		Where("p.status = ?", "active").
		Scan(&workers).Error; err != nil {
		return nil, err
	}

	evaluated := len(workers)
	computed := 0
	failures := 0

	for _, worker := range workers {
		if worker.ZoneID == 0 || worker.BaselineAmount <= 0 {
			failures++
			continue
		}

		idempotencyKey := fmt.Sprintf("premium_%s_%d", cycleID, worker.WorkerID)
		var existing models.PremiumPayment
		err := s.DB.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error
		if err == nil {
			computed++
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		premium := s.computePremium(worker.BaselineAmount, worker.RiskRating, worker.VehicleType)
		payment := models.PremiumPayment{WorkerID: worker.WorkerID, PolicyCycleID: cycle.ID, Amount: premium, Status: "computed", IdempotencyKey: idempotencyKey, Date: weekStart}
		if err := s.DB.Create(&payment).Error; err != nil {
			failures++
			continue
		}

		if err := s.DB.Model(&models.Policy{}).Where("worker_id = ? AND status = ?", worker.WorkerID, "active").Updates(map[string]interface{}{"premium_amount": premium, "policy_cycle_id": cycle.ID, "updated_at": time.Now().UTC()}).Error; err != nil {
			failures++
			continue
		}

		computed++
	}

	status := "completed"
	if failures > 0 {
		status = "partial_failure"
	}

	cycle.WorkersEvaluated = evaluated
	cycle.PremiumsComputed = computed
	cycle.PremiumFailures = failures
	cycle.Status = status

	if err := s.DB.Save(&cycle).Error; err != nil {
		return nil, err
	}

	return &WeeklyCycleResult{CycleID: cycle.CycleID, WorkersEvaluated: cycle.WorkersEvaluated, PremiumsComputed: cycle.PremiumsComputed, PremiumFailures: cycle.PremiumFailures, Status: cycle.Status}, nil
}

func (s *CoreOpsService) GenerateClaimsForDisruption(disruptionID uint, now time.Time) (*GeneratedClaimsResult, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("db unavailable")
	}

	var disruption models.Disruption
	if err := s.DB.First(&disruption, disruptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("disruption not found")
		}
		return nil, err
	}

	type eligibleWorker struct {
		WorkerID       uint
		BaselineAmount float64
		ActualEarnings float64
	}

	weekStart, _ := weekBounds(now.UTC())
	var workers []eligibleWorker
	if err := s.DB.Table("worker_profiles wp").
		Select("wp.worker_id, eb.baseline_amount, COALESCE(wes.total_earnings, 0) AS actual_earnings").
		Joins("JOIN policies p ON p.worker_id = wp.worker_id AND p.status = ?", "active").
		Joins("LEFT JOIN earnings_baselines eb ON eb.worker_id = wp.worker_id").
		Joins("LEFT JOIN weekly_earnings_summaries wes ON wes.worker_id = wp.worker_id AND wes.week_start = ?", weekStart).
		Where("wp.zone_id = ?", disruption.ZoneID).
		Scan(&workers).Error; err != nil {
		return nil, err
	}

	generated := 0
	skipped := 0

	for _, worker := range workers {
		if worker.BaselineAmount <= 0 {
			skipped++
			continue
		}

		var existingCount int64
		if err := s.DB.Model(&models.Claim{}).Where("disruption_id = ? AND worker_id = ?", disruptionID, worker.WorkerID).Count(&existingCount).Error; err != nil {
			return nil, err
		}
		if existingCount > 0 {
			skipped++
			continue
		}

		loss := math.Max(worker.BaselineAmount-worker.ActualEarnings, 0)
		if loss == 0 {
			skipped++
			continue
		}

		status := "pending"
		if loss > 1200 {
			status = "manual_review"
		}

		claim := models.Claim{DisruptionID: disruptionID, WorkerID: worker.WorkerID, ClaimAmount: round2(loss * 0.85), Status: status, FraudVerdict: "pending", CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
		if err := s.DB.Create(&claim).Error; err != nil {
			return nil, err
		}

		score := models.ClaimFraudScore{ClaimID: claim.ID, Score: 0.19, FinalVerdict: "clear", RuleViolations: "[]"}
		if status == "manual_review" {
			score.Score = 0.66
			score.FinalVerdict = "review"
		}
		if err := s.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "claim_id"}}, DoUpdates: clause.AssignmentColumns([]string{"score", "final_verdict", "rule_violations", "updated_at"})}).Create(&score).Error; err != nil {
			return nil, err
		}

		payload, _ := json.Marshal(map[string]interface{}{"event_id": fmt.Sprintf("evt_claim_%d", claim.ID), "event_type": "claim.generated", "occurred_at": now.UTC().Format(time.RFC3339), "producer": "core-backend", "payload": map[string]interface{}{"claim_id": claim.ID, "disruption_id": disruptionID, "worker_id": worker.WorkerID, "amount": claim.ClaimAmount}})
		_ = s.DB.Create(&models.KafkaEventLog{Topic: "indel.claims.generated", EventType: "claim.generated", PayloadJSON: string(payload)}).Error

		generated++
	}

	return &GeneratedClaimsResult{DisruptionID: fmt.Sprintf("dis_%d", disruptionID), WorkersChecked: len(workers), ClaimsGenerated: generated, ClaimsSkipped: skipped, Status: "completed"}, nil
}

func (s *CoreOpsService) QueueClaimPayout(claimID uint) (*PayoutResult, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("db unavailable")
	}

	var claim models.Claim
	if err := s.DB.First(&claim, claimID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("claim not found")
		}
		return nil, err
	}

	idempotencyKey := fmt.Sprintf("pay_clm_%d", claim.ID)
	payout := models.Payout{ClaimID: claim.ID, WorkerID: claim.WorkerID, Amount: round2(claim.ClaimAmount), Status: "queued", IdempotencyKey: idempotencyKey, RazorpayStatus: "queued"}

	if err := s.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "claim_id"}}, DoUpdates: clause.Assignments(map[string]interface{}{"amount": payout.Amount, "status": "queued", "idempotency_key": idempotencyKey, "updated_at": time.Now().UTC()})}).Create(&payout).Error; err != nil {
		return nil, err
	}

	var persisted models.Payout
	if err := s.DB.Where("claim_id = ?", claim.ID).First(&persisted).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Model(&models.Claim{}).Where("id = ?", claim.ID).Updates(map[string]interface{}{"status": "queued_for_payout", "updated_at": time.Now().UTC()}).Error; err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(map[string]interface{}{"event_id": fmt.Sprintf("evt_payout_%d", persisted.ID), "event_type": "payout.queued", "occurred_at": time.Now().UTC().Format(time.RFC3339), "producer": "core-backend", "payload": map[string]interface{}{"claim_id": persisted.ClaimID, "worker_id": persisted.WorkerID, "amount": persisted.Amount}})
	_ = s.DB.Create(&models.KafkaEventLog{Topic: "indel.payouts.queued", EventType: "payout.queued", PayloadJSON: string(payload)}).Error

	return &PayoutResult{PayoutID: fmt.Sprintf("pay_%d", persisted.ID), ClaimID: fmt.Sprintf("clm_%d", persisted.ClaimID), WorkerID: fmt.Sprintf("wkr_%d", persisted.WorkerID), AmountINR: persisted.Amount, Status: persisted.Status, IdempotencyKey: persisted.IdempotencyKey, RetryCount: persisted.RetryCount}, nil
}

func (s *CoreOpsService) ProcessQueuedPayouts(now time.Time) (*ProcessPayoutsResult, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("db unavailable")
	}

	var payouts []models.Payout
	if err := s.DB.Where("status IN ?", []string{"queued", "retry_pending"}).Where("next_retry_at IS NULL OR next_retry_at <= ?", now.UTC()).Order("id ASC").Find(&payouts).Error; err != nil {
		return nil, err
	}

	result := &ProcessPayoutsResult{Processed: len(payouts)}

	for _, payout := range payouts {
		payout.RetryCount++
		attempt := models.PayoutAttempt{PayoutID: payout.ID, AttemptNo: payout.RetryCount, Status: "processing", CreatedAt: now.UTC()}

		if shouldFailPayout(payout) {
			result.Failed++
			result.Retried++
			nextRetry := now.UTC().Add(time.Duration(payout.RetryCount) * 5 * time.Minute)
			attempt.Status = "failed"
			attempt.Error = "transient_gateway_failure"
			payout.Status = "retry_pending"
			payout.LastError = attempt.Error
			payout.NextRetryAt = &nextRetry
			payout.RazorpayStatus = "retry_pending"
		} else {
			result.Succeeded++
			processedAt := now.UTC()
			attempt.Status = "succeeded"
			payout.Status = "processed"
			payout.LastError = ""
			payout.NextRetryAt = nil
			payout.ProcessedAt = &processedAt
			payout.RazorpayStatus = "processed"
			payout.RazorpayID = fmt.Sprintf("rzp_mock_%d", payout.ID)
			_ = s.DB.Model(&models.Claim{}).Where("id = ?", payout.ClaimID).Updates(map[string]interface{}{"status": "paid", "updated_at": processedAt}).Error
		}

		if err := s.DB.Create(&attempt).Error; err != nil {
			return nil, err
		}
		if err := s.DB.Save(&payout).Error; err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *CoreOpsService) GetPayoutReconciliation(from, to time.Time) (*ReconciliationResult, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("db unavailable")
	}

	var payouts []models.Payout
	if err := s.DB.Where("created_at >= ? AND created_at <= ?", from.UTC(), to.UTC()).Find(&payouts).Error; err != nil {
		return nil, err
	}

	counts := map[string]int{"queued": 0, "processed": 0, "retry_pending": 0, "failed": 0}
	totals := map[string]float64{"queued_amount": 0, "processed_amount": 0, "retry_amount": 0}
	mismatchCount := 0

	for _, payout := range payouts {
		switch payout.Status {
		case "queued":
			counts["queued"]++
			totals["queued_amount"] += payout.Amount
		case "processed":
			counts["processed"]++
			totals["processed_amount"] += payout.Amount
		case "retry_pending":
			counts["retry_pending"]++
			totals["retry_amount"] += payout.Amount
		default:
			counts["failed"]++
		}

		var claim models.Claim
		if err := s.DB.First(&claim, payout.ClaimID).Error; err != nil || (payout.Status == "processed" && claim.Status != "paid") {
			mismatchCount++
		}
	}

	return &ReconciliationResult{From: from.UTC().Format(time.RFC3339), To: to.UTC().Format(time.RFC3339), Totals: totals, Counts: counts, ReconciliationOK: mismatchCount == 0, MismatchCount: mismatchCount}, nil
}

func (s *CoreOpsService) GenerateSyntheticData(req SyntheticGenerateRequest, now time.Time) (*SyntheticGenerateResult, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("db unavailable")
	}

	seed := req.Seed
	if seed == 0 {
		seed = 42
	}
	scenario := strings.TrimSpace(req.Scenario)
	if scenario == "" {
		scenario = "normal_week"
	}

	runID := fmt.Sprintf("syn_%d_%d", seed, now.UTC().Unix())
	outputDir := req.OutputDir
	if strings.TrimSpace(outputDir) == "" {
		outputDir = filepath.Join("generated", "synthetic", runID)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(int64(seed)))
	if err := truncateSyntheticTables(s.DB); err != nil {
		return nil, err
	}

	zones := syntheticZones()
	for i := range zones {
		zones[i].RiskRating = syntheticZoneRisk(scenario, rng, i)
	}
	if err := s.DB.Create(&zones).Error; err != nil {
		return nil, err
	}

	users := make([]models.User, 0, 500)
	profiles := make([]models.WorkerProfile, 0, 500)
	baselines := make([]models.EarningsBaseline, 0, 500)
	policies := make([]models.Policy, 0, 500)
	weeklySummaries := make([]models.WeeklyEarningsSummary, 0, 4000)
	premiumPayments := make([]models.PremiumPayment, 0, 4000)
	workersPerZone := 500 / len(zones)

	weekStart, weekEnd := weekBounds(now.UTC())
	currentCycle := models.WeeklyPolicyCycle{CycleID: cycleIDForDate(weekStart), WeekStart: weekStart, WeekEnd: weekEnd, Status: "seeded"}
	if err := s.DB.Create(&currentCycle).Error; err != nil {
		return nil, err
	}

	workerID := 1
	for zoneIdx, zone := range zones {
		assignments := workersPerZone
		if zoneIdx < 500%len(zones) {
			assignments++
		}
		for j := 0; j < assignments; j++ {
			users = append(users, models.User{ID: uint(workerID), Phone: fmt.Sprintf("+9199%08d", workerID), Role: "worker"})
			baseline := 3200 + float64(rng.Intn(2200))
			profiles = append(profiles, models.WorkerProfile{WorkerID: uint(workerID), Name: fmt.Sprintf("Worker %03d", workerID), ZoneID: zone.ID, VehicleType: []string{"two_wheeler", "bike", "scooter"}[workerID%3], UPIId: fmt.Sprintf("worker%03d@upi", workerID), AQIZone: "medium", TotalEarningsLifetime: baseline * 18})
			baselines = append(baselines, models.EarningsBaseline{WorkerID: uint(workerID), BaselineAmount: round2(baseline), LastUpdatedAt: now.UTC()})
			policies = append(policies, models.Policy{WorkerID: uint(workerID), Status: "active", PremiumAmount: round2(18 + zone.RiskRating*8), PolicyCycleID: currentCycle.ID})
			for weekOffset := 0; weekOffset < 8; weekOffset++ {
				historyStart := weekStart.AddDate(0, 0, -7*weekOffset)
				historyEnd := historyStart.AddDate(0, 0, 6)
				earnings := baseline * (0.82 + rng.Float64()*0.26)
				if scenario == "severe_disruption" && weekOffset == 0 && workerID%5 == 0 {
					earnings *= 0.42
				}
				if scenario == "fraud_burst" && weekOffset == 0 && workerID%11 == 0 {
					earnings *= 0.33
				}
				weeklySummaries = append(weeklySummaries, models.WeeklyEarningsSummary{WorkerID: uint(workerID), WeekStart: historyStart, WeekEnd: historyEnd, TotalEarnings: round2(earnings), ClaimEligible: weekOffset == 0})
				premiumPayments = append(premiumPayments, models.PremiumPayment{WorkerID: uint(workerID), PolicyCycleID: currentCycle.ID, Amount: round2(15 + zone.RiskRating*10), Status: "completed", IdempotencyKey: fmt.Sprintf("seed_%d_%d", workerID, weekOffset), Date: historyStart})
			}
			workerID++
		}
	}

	if err := s.DB.Create(&users).Error; err != nil { return nil, err }
	if err := s.DB.Create(&profiles).Error; err != nil { return nil, err }
	if err := s.DB.Create(&baselines).Error; err != nil { return nil, err }
	if err := s.DB.Create(&policies).Error; err != nil { return nil, err }
	if err := s.DB.Create(&weeklySummaries).Error; err != nil { return nil, err }
	if err := s.DB.Create(&premiumPayments).Error; err != nil { return nil, err }

	disruptions := make([]models.Disruption, 0, len(zones)*2)
	for _, zone := range zones {
		count := 2
		if scenario == "severe_disruption" || scenario == "fraud_burst" {
			count = 3
		}
		for idx := 0; idx < count; idx++ {
			start := now.UTC().Add(-time.Duration((idx+1)*12) * time.Hour)
			confirmed := start.Add(15 * time.Minute)
			disruptions = append(disruptions, models.Disruption{ZoneID: zone.ID, Type: disruptionTypeForScenario(scenario, idx), Severity: severityForScenario(scenario, rng), Confidence: round2(0.72 + rng.Float64()*0.24), Status: "confirmed", SignalTimestamp: &start, ConfirmedAt: &confirmed, StartTime: &start})
		}
	}
	if err := s.DB.Create(&disruptions).Error; err != nil { return nil, err }

	claims := make([]models.Claim, 0, 2000)
	scores := make([]models.ClaimFraudScore, 0, 2000)
	payouts := make([]models.Payout, 0, 1200)

	for claimNo := 0; claimNo < 2000; claimNo++ {
		worker := profiles[rng.Intn(len(profiles))]
		disruption := disruptions[rng.Intn(len(disruptions))]
		for disruption.ZoneID != worker.ZoneID {
			disruption = disruptions[rng.Intn(len(disruptions))]
		}
		isFlagged := syntheticFraudFlag(scenario, worker.WorkerID, rng)
		status := "approved"
		verdict := "clear"
		if claimNo%4 == 0 { status = "pending" }
		if isFlagged {
			status = "manual_review"
			verdict = "flagged"
		}
		claims = append(claims, models.Claim{DisruptionID: disruption.ID, WorkerID: worker.WorkerID, ClaimAmount: round2(280 + rng.Float64()*900), Status: status, FraudVerdict: verdict, CreatedAt: now.UTC().Add(-time.Duration(rng.Intn(240)) * time.Hour), UpdatedAt: now.UTC()})
	}
	if err := s.DB.Create(&claims).Error; err != nil { return nil, err }

	for _, claim := range claims {
		finalVerdict := "clear"
		score := 0.18 + rng.Float64()*0.21
		factors := "[]"
		if claim.FraudVerdict == "flagged" {
			finalVerdict = "flagged"
			score = 0.78 + rng.Float64()*0.16
			payload, _ := json.Marshal([]map[string]interface{}{{"name": "gps_mismatch", "impact": 0.24}, {"name": "session_gap", "impact": 0.12}})
			factors = string(payload)
		}
		scores = append(scores, models.ClaimFraudScore{ClaimID: claim.ID, Score: round2(score), FinalVerdict: finalVerdict, RuleViolations: factors})
		if claim.Status == "approved" {
			processedAt := now.UTC().Add(-30 * time.Minute)
			payouts = append(payouts, models.Payout{ClaimID: claim.ID, WorkerID: claim.WorkerID, Amount: round2(claim.ClaimAmount * 0.9), Status: "processed", IdempotencyKey: fmt.Sprintf("pay_clm_%d", claim.ID), RetryCount: 1, RazorpayID: fmt.Sprintf("rzp_seed_%d", claim.ID), RazorpayStatus: "processed", ProcessedAt: &processedAt})
		}
	}
	if err := s.DB.Create(&scores).Error; err != nil { return nil, err }
	if len(payouts) > 0 {
		if err := s.DB.Create(&payouts).Error; err != nil { return nil, err }
	}

	sqlPath := filepath.Join(outputDir, "seed.sql")
	workersCSV := filepath.Join(outputDir, "workers.csv")
	claimsCSV := filepath.Join(outputDir, "claims.csv")
	payoutsCSV := filepath.Join(outputDir, "payouts.csv")
	if err := writeSyntheticSQL(sqlPath, zones, profiles, claims, payouts); err != nil { return nil, err }
	if err := writeWorkersCSV(workersCSV, profiles); err != nil { return nil, err }
	if err := writeClaimsCSV(claimsCSV, claims); err != nil { return nil, err }
	if err := writePayoutsCSV(payoutsCSV, payouts); err != nil { return nil, err }

	run := models.SyntheticGenerationRun{RunID: runID, Seed: seed, Scenario: scenario, OutputDir: outputDir, WorkersCreated: len(profiles), ZonesCreated: len(zones), DisruptionsCreated: len(disruptions), ClaimsCreated: len(claims), PayoutsCreated: len(payouts), Status: "completed"}
	if err := s.DB.Create(&run).Error; err != nil { return nil, err }

	return &SyntheticGenerateResult{RunID: runID, Seed: seed, Scenario: scenario, Status: "completed", Counts: map[string]int{"workers": len(profiles), "zones": len(zones), "disruptions": len(disruptions), "claims": len(claims), "payouts": len(payouts)}, Artifacts: map[string]string{"workers_csv": workersCSV, "claims_csv": claimsCSV, "payouts_csv": payoutsCSV, "seed_sql": sqlPath}, Integration: map[string]string{"premium_service": "fallback rule-based pricing active until Part 3 premium service is connected", "fraud_service": "synthetic fraud verdicts seeded deterministically until Part 3 fraud service is connected", "forecast_service": "not required for Part 4 execution path; reserve forecasting remains an integration point"}}, nil
}

func (s *CoreOpsService) computePremium(baselineAmount, riskRating float64, vehicleType string) float64 {
	vehicleFactor := 1.0
	switch strings.ToLower(strings.TrimSpace(vehicleType)) {
	case "bike":
		vehicleFactor = 1.08
	case "scooter":
		vehicleFactor = 1.04
	case "two_wheeler":
		vehicleFactor = 1.06
	}
	base := baselineAmount * 0.0052
	riskAdjusted := base * (1 + riskRating) * vehicleFactor
	return round2(math.Max(10, math.Min(riskAdjusted, 35)))
}

func shouldFailPayout(payout models.Payout) bool {
	return payout.WorkerID%11 == 0 && payout.RetryCount == 1
}

func weekBounds(now time.Time) (time.Time, time.Time) {
	normalized := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	offset := (int(normalized.Weekday()) + 6) % 7
	start := normalized.AddDate(0, 0, -offset)
	end := start.AddDate(0, 0, 6)
	return start, end
}

func cycleIDForDate(weekStart time.Time) string {
	year, week := weekStart.ISOWeek()
	return fmt.Sprintf("cyc_%d_w%02d", year, week)
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func truncateSyntheticTables(db *gorm.DB) error {
	tableNames := []string{"payout_attempts", "payouts", "claim_fraud_scores", "claim_audit_logs", "claims", "disruptions", "premium_payments", "weekly_earnings_summaries", "earnings_baselines", "policies", "worker_profiles", "users", "weekly_policy_cycles", "synthetic_generation_runs", "zones"}
	for _, name := range tableNames {
		if err := db.Exec("DELETE FROM " + name).Error; err != nil {
			return err
		}
	}
	return nil
}

func syntheticZones() []models.Zone {
	return []models.Zone{{Name: "Tambaram", City: "Chennai", State: "Tamil Nadu", Level: "A"}, {Name: "Adyar", City: "Chennai", State: "Tamil Nadu", Level: "A"}, {Name: "Velachery", City: "Chennai", State: "Tamil Nadu", Level: "A"}, {Name: "Koramangala", City: "Bengaluru", State: "Karnataka", Level: "B"}, {Name: "Indiranagar", City: "Bengaluru", State: "Karnataka", Level: "B"}, {Name: "Whitefield", City: "Bengaluru", State: "Karnataka", Level: "B"}, {Name: "Andheri", City: "Mumbai", State: "Maharashtra", Level: "C"}, {Name: "Bandra", City: "Mumbai", State: "Maharashtra", Level: "C"}, {Name: "Powai", City: "Mumbai", State: "Maharashtra", Level: "C"}, {Name: "Rohini", City: "Delhi", State: "Delhi", Level: "B"}}
}

func syntheticZoneRisk(scenario string, rng *rand.Rand, idx int) float64 {
	base := 0.28 + float64(idx%4)*0.1 + rng.Float64()*0.08
	switch scenario {
	case "mild_disruption":
		base += 0.08
	case "severe_disruption":
		base += 0.18
	case "fraud_burst":
		base += 0.12
	}
	return round2(math.Min(base, 0.92))
}

func severityForScenario(scenario string, rng *rand.Rand) string {
	switch scenario {
	case "severe_disruption":
		if rng.Float64() > 0.2 { return "high" }
		return "medium"
	case "fraud_burst":
		return "medium"
	default:
		if rng.Float64() > 0.7 { return "high" }
		return "medium"
	}
}

func disruptionTypeForScenario(scenario string, idx int) string {
	switch scenario {
	case "mild_disruption":
		return "order_drop"
	case "severe_disruption":
		if idx%2 == 0 { return "heavy_rain" }
		return "flood"
	case "fraud_burst":
		return "order_drop"
	default:
		return "heavy_rain"
	}
}

func syntheticFraudFlag(scenario string, workerID uint, rng *rand.Rand) bool {
	rate := 0.12
	if scenario == "fraud_burst" { rate = 0.18 }
	if workerID%17 == 0 { return true }
	return rng.Float64() < rate
}

func writeSyntheticSQL(path string, zones []models.Zone, profiles []models.WorkerProfile, claims []models.Claim, payouts []models.Payout) error {
	f, err := os.Create(path)
	if err != nil { return err }
	defer f.Close()
	lines := []string{"-- deterministic synthetic seed output"}
	for _, zone := range zones {
		lines = append(lines, fmt.Sprintf("INSERT INTO zones (id, name, city, state, level, risk_rating) VALUES (%d, '%s', '%s', '%s', '%s', %.2f);", zone.ID, escapeSQL(zone.Name), escapeSQL(zone.City), escapeSQL(zone.State), escapeSQL(zone.Level), zone.RiskRating))
	}
	for _, profile := range profiles[:min(25, len(profiles))] {
		lines = append(lines, fmt.Sprintf("INSERT INTO worker_profiles (worker_id, name, zone_id, vehicle_type, upi_id, aqi_zone, total_earnings_lifetime) VALUES (%d, '%s', %d, '%s', '%s', '%s', %.2f);", profile.WorkerID, escapeSQL(profile.Name), profile.ZoneID, escapeSQL(profile.VehicleType), escapeSQL(profile.UPIId), escapeSQL(profile.AQIZone), profile.TotalEarningsLifetime))
	}
	for _, claim := range claims[:min(50, len(claims))] {
		lines = append(lines, fmt.Sprintf("INSERT INTO claims (id, disruption_id, worker_id, claim_amount, status, fraud_verdict) VALUES (%d, %d, %d, %.2f, '%s', '%s');", claim.ID, claim.DisruptionID, claim.WorkerID, claim.ClaimAmount, escapeSQL(claim.Status), escapeSQL(claim.FraudVerdict)))
	}
	for _, payout := range payouts[:min(25, len(payouts))] {
		lines = append(lines, fmt.Sprintf("INSERT INTO payouts (claim_id, worker_id, amount, status, idempotency_key) VALUES (%d, %d, %.2f, '%s', '%s');", payout.ClaimID, payout.WorkerID, payout.Amount, escapeSQL(payout.Status), escapeSQL(payout.IdempotencyKey)))
	}
	_, err = f.WriteString(strings.Join(lines, "\n"))
	return err
}

func writeWorkersCSV(path string, profiles []models.WorkerProfile) error {
	f, err := os.Create(path)
	if err != nil { return err }
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"worker_id", "name", "zone_id", "vehicle_type", "total_earnings_lifetime"}); err != nil { return err }
	for _, profile := range profiles {
		if err := w.Write([]string{strconv.Itoa(int(profile.WorkerID)), profile.Name, strconv.Itoa(int(profile.ZoneID)), profile.VehicleType, fmt.Sprintf("%.2f", profile.TotalEarningsLifetime)}); err != nil { return err }
	}
	return nil
}

func writeClaimsCSV(path string, claims []models.Claim) error {
	f, err := os.Create(path)
	if err != nil { return err }
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"claim_id", "disruption_id", "worker_id", "claim_amount", "status", "fraud_verdict"}); err != nil { return err }
	for _, claim := range claims {
		if err := w.Write([]string{strconv.Itoa(int(claim.ID)), strconv.Itoa(int(claim.DisruptionID)), strconv.Itoa(int(claim.WorkerID)), fmt.Sprintf("%.2f", claim.ClaimAmount), claim.Status, claim.FraudVerdict}); err != nil { return err }
	}
	return nil
}

func writePayoutsCSV(path string, payouts []models.Payout) error {
	f, err := os.Create(path)
	if err != nil { return err }
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"payout_id", "claim_id", "worker_id", "amount", "status", "retry_count"}); err != nil { return err }
	for _, payout := range payouts {
		if err := w.Write([]string{strconv.Itoa(int(payout.ID)), strconv.Itoa(int(payout.ClaimID)), strconv.Itoa(int(payout.WorkerID)), fmt.Sprintf("%.2f", payout.Amount), payout.Status, strconv.Itoa(payout.RetryCount)}); err != nil { return err }
	}
	return nil
}

func escapeSQL(value string) string { return strings.ReplaceAll(value, "'", "''") }

func min(a, b int) int {
	if a < b { return a }
	return b
}
