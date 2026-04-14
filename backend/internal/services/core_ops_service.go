package services

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/Shravanthi20/InDel/backend/pkg/razorpay"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CoreOpsService struct {
	DB             *gorm.DB
	razorpayClient *razorpay.RazorpayClient
	producer       interface{} // Kafka producer
}

func NewCoreOpsService(db *gorm.DB) *CoreOpsService {
	return &CoreOpsService{DB: db}
}

// SetRazorpayClient sets the Razorpay client
func (s *CoreOpsService) SetRazorpayClient(client *razorpay.RazorpayClient) {
	s.razorpayClient = client
}

// SetProducer sets the Kafka producer
func (s *CoreOpsService) SetProducer(producer interface{}) {
	s.producer = producer
}

type WeeklyCycleResult struct {
	CycleID          string `json:"cycle_id"`
	WorkersEvaluated int    `json:"workers_evaluated"`
	PremiumsComputed int    `json:"premiums_computed"`
	PremiumFailures  int    `json:"premium_failures"`
	Status           string `json:"status"`
}

type GeneratedClaimsResult struct {
	DisruptionID      string `json:"disruption_id"`
	WorkersChecked    int    `json:"workers_checked"`
	ClaimsGenerated   int    `json:"claims_generated"`
	ClaimsSkipped     int    `json:"claims_skipped"`
	Status            string `json:"status"`
	GeneratedClaimIDs []uint `json:"-"`
}

type AutoProcessDisruptionResult struct {
	DisruptionID       string `json:"disruption_id"`
	WorkersNotified    int    `json:"workers_notified"`
	WorkersChecked     int    `json:"workers_checked"`
	ClaimsGenerated    int    `json:"claims_generated"`
	ClaimsSkipped      int    `json:"claims_skipped"`
	PayoutsQueued      int    `json:"payouts_queued"`
	PayoutsProcessed   int    `json:"payouts_processed"`
	PayoutsSucceeded   int    `json:"payouts_succeeded"`
	PayoutsFailed      int    `json:"payouts_failed"`
	ManualReviewClaims int    `json:"manual_review_claims"`
	Status             string `json:"status"`
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
	Seed               int                      `json:"seed"`
	Scenario           string                   `json:"scenario"`
	OutputDir          string                   `json:"output_dir"`
	WeeksHistory       int                      `json:"weeks_history,omitempty"`
	ClaimCount         int                      `json:"claim_count,omitempty"`
	DisruptionsPerZone int                      `json:"disruptions_per_zone,omitempty"`
	AccountOptions     []SyntheticAccountOption `json:"account_options,omitempty"`
}

type SyntheticAccountOption struct {
	Label          string  `json:"label,omitempty"`
	Count          int     `json:"count"`
	ZoneLevel      string  `json:"zone_level,omitempty"`
	VehicleType    string  `json:"vehicle_type,omitempty"`
	PremiumEnabled *bool   `json:"premium_enabled,omitempty"`
	WeeklyPremium  float64 `json:"weekly_premium,omitempty"`
	AQIZone        string  `json:"aqi_zone,omitempty"`
	BaselineMin    float64 `json:"baseline_min,omitempty"`
	BaselineMax    float64 `json:"baseline_max,omitempty"`
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

const syntheticEmailDomain = "synthetic.indel.local"

type syntheticWorkerSpec struct {
	Label          string
	ZoneID         uint
	VehicleType    string
	PremiumEnabled bool
	WeeklyPremium  float64
	AQIZone        string
	BaselineMin    float64
	BaselineMax    float64
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
		Select("p.worker_id, wp.zone_id, z.risk_rating, wp.vehicle_type, eb.baseline_amount AS baseline_amount").
		Joins("LEFT JOIN worker_profiles wp ON wp.worker_id = p.worker_id").
		Joins("LEFT JOIN zones z ON z.id = wp.zone_id").
		Joins("LEFT JOIN earnings_baseline eb ON eb.worker_id = p.worker_id").
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
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		quote, quoteErr := QuotePremium(s.DB, worker.WorkerID, now.UTC())
		premium := s.computePremium(worker.BaselineAmount, worker.RiskRating, worker.VehicleType)
		if quoteErr == nil && quote != nil && quote.WeeklyPremiumINR > 0 {
			premium = quote.WeeklyPremiumINR
		}
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
	return s.generateClaimsForDisruption(disruptionID, now)
}

func (s *CoreOpsService) generateClaimsForDisruption(disruptionID uint, now time.Time) (*GeneratedClaimsResult, error) {
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
	// Check for workers in the disrupted zone who have ACTIVE protection plans
	// This ensures claims are only generated for workers currently enrolled in the protection policy
	if err := s.DB.Table("worker_profiles wp").
		Select("wp.worker_id, COALESCE(eb.baseline_amount, 5000.0) AS baseline_amount, COALESCE(wes.total_earnings, 0) AS actual_earnings").
		Joins("INNER JOIN policies p ON p.worker_id = wp.worker_id AND p.status = 'active'").
		Joins("LEFT JOIN earnings_baseline eb ON eb.worker_id = wp.worker_id").
		Joins("LEFT JOIN weekly_earnings_summary wes ON wes.worker_id = wp.worker_id AND wes.week_start = ?", weekStart).
		Where("wp.zone_id = ?", disruption.ZoneID).
		Scan(&workers).Error; err != nil {
		return nil, err
	}

	generated := 0
	skipped := 0
	generatedClaimIDs := make([]uint, 0)

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

		// Compare the worker's baseline against observed weekly earnings so the
		// seeded demo fixtures still produce claims when the disruption impacts income.
		loss := math.Max(worker.BaselineAmount-worker.ActualEarnings, 0)

		if loss <= 0 {
			skipped++
			continue
		}

		status := "approved"
		fraudVerdict := "clear"
		fraudScore := 0.19 // Default low fraud score for legitimate disruptions

		// DEMO MODE: Disable manual review threshold (set to 100k)
		if loss > 100000 {
			status = "manual_review"
			fraudVerdict = "review"
			fraudScore = 0.66 // Higher fraud score for large claims
		}

		claim := models.Claim{DisruptionID: disruptionID, WorkerID: worker.WorkerID, ClaimAmount: round2(loss * 0.85), Status: status, FraudVerdict: fraudVerdict, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
		if err := s.DB.Create(&claim).Error; err != nil {
			return nil, err
		}

		score := models.ClaimFraudScore{ClaimID: claim.ID, IsolationForestScore: fraudScore, DbscanScore: fraudScore, FinalVerdict: fraudVerdict, RuleViolations: "[]"}
		if err := s.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "claim_id"}}, DoUpdates: clause.AssignmentColumns([]string{"isolation_forest_score", "dbscan_score", "final_verdict", "rule_violations"})}).Create(&score).Error; err != nil {
			return nil, err
		}

		payload, _ := json.Marshal(map[string]interface{}{"event_id": fmt.Sprintf("evt_claim_%d", claim.ID), "event_type": "claim.generated", "occurred_at": now.UTC().Format(time.RFC3339), "producer": "core-backend", "payload": map[string]interface{}{"claim_id": claim.ID, "disruption_id": disruptionID, "worker_id": worker.WorkerID, "amount": claim.ClaimAmount, "status": claim.Status}})
		_ = s.DB.Create(&models.KafkaEventLog{Topic: "indel.claims.generated", EventType: "claim.generated", PayloadJSON: string(payload)}).Error

		// Trigger the missing worker notification!
		_ = s.notifyClaimGenerated(worker.WorkerID, claim.ID, claim.ClaimAmount, now, disruption.Type)

		generated++
		generatedClaimIDs = append(generatedClaimIDs, claim.ID)
	}

	return &GeneratedClaimsResult{DisruptionID: fmt.Sprintf("dis_%d", disruptionID), WorkersChecked: len(workers), ClaimsGenerated: generated, ClaimsSkipped: skipped, Status: "completed", GeneratedClaimIDs: generatedClaimIDs}, nil
}

func (s *CoreOpsService) AutoProcessDisruption(disruptionID uint, now time.Time) (*AutoProcessDisruptionResult, error) {
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

	notified, err := s.notifyWorkersForDisruption(disruption, now)
	if err != nil {
		return nil, err
	}

	claimsResult, err := s.generateClaimsForDisruption(disruptionID, now)
	if err != nil {
		return nil, err
	}

	targetClaims, err := s.loadApprovedClaimsNeedingPayout(disruptionID)
	if err != nil {
		return nil, err
	}

	queued := 0
	for _, claim := range targetClaims {
		result, err := s.QueueClaimPayout(claim.ID)
		if err != nil {
			return nil, err
		}
		if result.Status == "queued" {
			queued++
		}
	}

	targetPayoutIDs, err := s.loadProcessablePayoutIDsForDisruption(disruptionID, now)
	if err != nil {
		return nil, err
	}

	processResult, err := s.processPayoutsByID(targetPayoutIDs, now)
	if err != nil {
		return nil, err
	}

	var manualReviewClaims int64
	if err := s.DB.Model(&models.Claim{}).Where("disruption_id = ? AND status = ?", disruptionID, "manual_review").Count(&manualReviewClaims).Error; err != nil {
		return nil, err
	}

	status := "completed"
	if processResult.Failed > 0 {
		status = "partial_failure"
	}

	return &AutoProcessDisruptionResult{
		DisruptionID:       fmt.Sprintf("dis_%d", disruptionID),
		WorkersNotified:    notified,
		WorkersChecked:     claimsResult.WorkersChecked,
		ClaimsGenerated:    claimsResult.ClaimsGenerated,
		ClaimsSkipped:      claimsResult.ClaimsSkipped,
		PayoutsQueued:      queued,
		PayoutsProcessed:   processResult.Processed,
		PayoutsSucceeded:   processResult.Succeeded,
		PayoutsFailed:      processResult.Failed,
		ManualReviewClaims: int(manualReviewClaims),
		Status:             status,
	}, nil
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

	var existing models.Payout
	if err := s.DB.Where("claim_id = ?", claim.ID).First(&existing).Error; err == nil {
		return &PayoutResult{
			PayoutID:       fmt.Sprintf("pay_%d", existing.ID),
			ClaimID:        fmt.Sprintf("clm_%d", existing.ClaimID),
			WorkerID:       fmt.Sprintf("wkr_%d", existing.WorkerID),
			AmountINR:      existing.Amount,
			Status:         existing.Status,
			IdempotencyKey: existing.IdempotencyKey,
			RetryCount:     existing.RetryCount,
		}, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Validate worker has an active policy before queueing a new payout.
	var policy models.Policy
	if err := s.DB.Where("worker_id = ? AND status = ?", claim.WorkerID, "active").First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("worker does not have an active policy - payout denied")
		}
		return nil, err
	}

	idempotencyKey := fmt.Sprintf("pay_clm_%d", claim.ID)
	payout := models.Payout{ClaimID: claim.ID, WorkerID: claim.WorkerID, Amount: round2(claim.ClaimAmount), Status: "queued", IdempotencyKey: idempotencyKey, RazorpayStatus: "queued"}

	if err := s.DB.Create(&payout).Error; err != nil {
		return nil, err
	}

	if err := s.DB.Model(&models.Claim{}).Where("id = ?", claim.ID).Updates(map[string]interface{}{"status": "queued_for_payout", "updated_at": time.Now().UTC()}).Error; err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(map[string]interface{}{"event_id": fmt.Sprintf("evt_payout_%d", payout.ID), "event_type": "payout.queued", "occurred_at": time.Now().UTC().Format(time.RFC3339), "producer": "core-backend", "payload": map[string]interface{}{"claim_id": payout.ClaimID, "worker_id": payout.WorkerID, "amount": payout.Amount}})
	_ = s.DB.Create(&models.KafkaEventLog{Topic: "indel.payouts.queued", EventType: "payout.queued", PayloadJSON: string(payload)}).Error

	return &PayoutResult{PayoutID: fmt.Sprintf("pay_%d", payout.ID), ClaimID: fmt.Sprintf("clm_%d", payout.ClaimID), WorkerID: fmt.Sprintf("wkr_%d", payout.WorkerID), AmountINR: payout.Amount, Status: payout.Status, IdempotencyKey: payout.IdempotencyKey, RetryCount: payout.RetryCount}, nil
}

func (s *CoreOpsService) ProcessQueuedPayouts(now time.Time) (*ProcessPayoutsResult, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("db unavailable")
	}

	var payoutIDs []uint
	if err := s.DB.Model(&models.Payout{}).
		Where("status IN ?", []string{"queued", "retry_pending"}).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now.UTC()).
		Order("id ASC").
		Pluck("id", &payoutIDs).Error; err != nil {
		return nil, err
	}

	return s.processPayoutsByID(payoutIDs, now)
}

func (s *CoreOpsService) processPayoutsByID(payoutIDs []uint, now time.Time) (*ProcessPayoutsResult, error) {
	if s.DB == nil {
		return nil, fmt.Errorf("db unavailable")
	}

	if len(payoutIDs) == 0 {
		return &ProcessPayoutsResult{}, nil
	}

	var payouts []models.Payout
	if err := s.DB.Where("id IN ?", payoutIDs).Where("status IN ?", []string{"queued", "retry_pending"}).Where("next_retry_at IS NULL OR next_retry_at <= ?", now.UTC()).Order("id ASC").Find(&payouts).Error; err != nil {
		return nil, err
	}

	result := &ProcessPayoutsResult{Processed: len(payouts)}
	mockSequence := 0

	for _, payout := range payouts {
		payout.RetryCount++
		attempt := models.PayoutAttempt{PayoutID: payout.ID, AttemptNo: payout.RetryCount, Status: "processing", CreatedAt: now.UTC()}

		isMockMode := s.razorpayClient == nil || s.razorpayClient.MockMode
		if isMockMode {
			if payout.Status == "retry_pending" {
				result.Succeeded++
				processedAt := now.UTC()
				attempt.Status = "succeeded"
				payout.Status = "processed"
				payout.LastError = ""
				payout.NextRetryAt = nil
				payout.ProcessedAt = &processedAt
				payout.RazorpayStatus = "processed"
				payout.RazorpayID = fmt.Sprintf("rzp_demo_mock_%d_%d", time.Now().Unix(), payout.WorkerID)

				if s.producer != nil {
					_ = s.notifyPayoutProcessed(payout.WorkerID, payout.ClaimID, payout.Amount, processedAt)
				}

				_ = s.DB.Model(&models.Claim{}).Where("id = ?", payout.ClaimID).Updates(map[string]interface{}{"status": "paid", "updated_at": processedAt}).Error

				s.DB.Create(&attempt)
				s.DB.Save(&payout)
				continue
			}

			mockSequence++
			if mockSequence%2 != 0 {
				result.Succeeded++
				processedAt := now.UTC()
				attempt.Status = "succeeded"
				payout.Status = "processed"
				payout.LastError = ""
				payout.NextRetryAt = nil
				payout.ProcessedAt = &processedAt
				payout.RazorpayStatus = "processed"
				payout.RazorpayID = fmt.Sprintf("rzp_demo_mock_%d_%d", time.Now().Unix(), payout.WorkerID)

				if s.producer != nil {
					_ = s.notifyPayoutProcessed(payout.WorkerID, payout.ClaimID, payout.Amount, processedAt)
				}

				_ = s.DB.Model(&models.Claim{}).Where("id = ?", payout.ClaimID).Updates(map[string]interface{}{"status": "paid", "updated_at": processedAt}).Error

				s.DB.Create(&attempt)
				s.DB.Save(&payout)
				continue
			}

			result.Failed++
			result.Retried++
			nextRetry := now.UTC().Add(5 * time.Minute)
			attempt.Status = "failed"
			attempt.Error = "transient_mock_failure"
			payout.Status = "retry_pending"
			payout.LastError = attempt.Error
			payout.NextRetryAt = &nextRetry
			payout.RazorpayStatus = "retry_pending"
			if err := s.DB.Create(&attempt).Error; err == nil {
				_ = s.DB.Save(&payout).Error
			}
			continue
		}

		// Fetch worker UPI
		var worker models.WorkerProfile
		if err := s.DB.Where("worker_id = ?", payout.WorkerID).First(&worker).Error; err != nil {
			result.Failed++
			attempt.Status = "failed"
			attempt.Error = "worker_not_found"
			payout.LastError = attempt.Error
			payout.Status = "failed"
			s.DB.Create(&attempt)
			s.DB.Save(&payout)
			continue
		}

		// Verify worker still has an active policy before processing payout
		var policy models.Policy
		if err := s.DB.Where("worker_id = ? AND status = ?", payout.WorkerID, "active").First(&policy).Error; err != nil {
			result.Failed++
			attempt.Status = "failed"
			attempt.Error = "no_active_policy"
			payout.LastError = attempt.Error
			payout.Status = "failed"
			s.DB.Create(&attempt)
			s.DB.Save(&payout)
			continue
		}

		if worker.UPIId == "" {
			result.Failed++
			attempt.Status = "failed"
			attempt.Error = "no_upi_found"
			payout.LastError = attempt.Error
			payout.Status = "failed"
			s.DB.Create(&attempt)
			s.DB.Save(&payout)
			continue
		}

		// Call Razorpay API with retry logic
		var razorpayID string
		var lastError string
		maxRetries := 5

		for attempt_count := 0; attempt_count < maxRetries; attempt_count++ {
			if s.razorpayClient == nil {
				lastError = "razorpay_client_not_available"
				break
			}

			pID, pErr := s.razorpayClient.CreatePayout(payout.WorkerID, payout.Amount, worker.UPIId)
			if pErr == nil {
				razorpayID = pID
				lastError = ""
				break
			}

			lastError = pErr.Error()

			// Check if error is transient
			if !razorpay.IsTransientError(lastError) {
				// Permanent error - don't retry
				break
			}

			if attempt_count < maxRetries-1 {
				// Back off and retry
				backoffDuration := time.Duration(attempt_count+1) * 100 * time.Millisecond
				time.Sleep(backoffDuration)
			}
		}

		// Update payout based on result
		if razorpayID != "" {
			result.Succeeded++
			processedAt := now.UTC()
			attempt.Status = "succeeded"
			payout.Status = "processed"
			payout.LastError = ""
			payout.NextRetryAt = nil
			payout.ProcessedAt = &processedAt
			payout.RazorpayStatus = "processed"
			payout.RazorpayID = razorpayID

			// Publish Kafka event
			if s.producer != nil {
				_ = s.notifyPayoutProcessed(payout.WorkerID, payout.ClaimID, payout.Amount, processedAt)
			}

			// Update claim status to paid
			_ = s.DB.Model(&models.Claim{}).Where("id = ?", payout.ClaimID).Updates(map[string]interface{}{"status": "paid", "updated_at": processedAt}).Error
		} else {
			// Check if we should retry
			if razorpay.IsTransientError(lastError) && payout.RetryCount < maxRetries {
				result.Retried++
				nextRetry := now.UTC().Add(time.Duration(payout.RetryCount) * 5 * time.Minute)
				attempt.Status = "failed"
				attempt.Error = lastError
				payout.Status = "retry_pending"
				payout.LastError = lastError
				payout.NextRetryAt = &nextRetry
				payout.RazorpayStatus = "retry_pending"
			} else {
				result.Failed++
				attempt.Status = "failed"
				attempt.Error = lastError
				payout.Status = "failed"
				payout.LastError = lastError
				payout.RazorpayStatus = "failed"
			}
		}

		if err := s.DB.Create(&attempt).Error; err != nil {
			continue // Log and continue instead of returning error
		}
		if err := s.DB.Save(&payout).Error; err != nil {
			continue // Log and continue instead of returning error
		}
	}

	return result, nil
}

func (s *CoreOpsService) loadApprovedClaimsNeedingPayout(disruptionID uint) ([]models.Claim, error) {
	type claimRow struct {
		models.Claim
		PayoutStatus string `gorm:"column:payout_status"`
	}

	rows := make([]claimRow, 0)
	if err := s.DB.Table("claims c").
		Select("c.*, COALESCE(p.status, '') AS payout_status").
		Joins("LEFT JOIN payouts p ON p.claim_id = c.id").
		Joins("LEFT JOIN policies pol ON pol.worker_id = c.worker_id AND pol.status = ?", "active").
		Where("c.disruption_id = ? AND c.status = ?", disruptionID, "approved").
		Where("pol.worker_id IS NOT NULL").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	claims := make([]models.Claim, 0, len(rows))
	for _, row := range rows {
		if row.PayoutStatus == "" || row.PayoutStatus == "queued" || row.PayoutStatus == "retry_pending" {
			claims = append(claims, row.Claim)
		}
	}

	return claims, nil
}

func (s *CoreOpsService) loadProcessablePayoutIDsForDisruption(disruptionID uint, now time.Time) ([]uint, error) {
	payoutIDs := make([]uint, 0)
	err := s.DB.Table("payouts p").
		Select("p.id").
		Joins("JOIN claims c ON c.id = p.claim_id").
		Joins("LEFT JOIN policies pol ON pol.worker_id = p.worker_id AND pol.status = ?", "active").
		Where("c.disruption_id = ?", disruptionID).
		Where("p.status IN ?", []string{"queued", "retry_pending"}).
		Where("p.next_retry_at IS NULL OR p.next_retry_at <= ?", now.UTC()).
		Where("pol.worker_id IS NOT NULL").
		Order("p.id ASC").
		Pluck("p.id", &payoutIDs).Error
	return payoutIDs, err
}

func (s *CoreOpsService) notifyWorkersForDisruption(disruption models.Disruption, now time.Time) (int, error) {
	type workerRow struct {
		WorkerID uint
		ZoneName string
	}

	workers := make([]workerRow, 0)
	if err := s.DB.Table("worker_profiles wp").
		Select("wp.worker_id, z.name AS zone_name").
		Joins("JOIN policies p ON p.worker_id = wp.worker_id AND p.status = ?", "active").
		Joins("JOIN zones z ON z.id = wp.zone_id").
		Where("wp.zone_id = ?", disruption.ZoneID).
		Scan(&workers).Error; err != nil {
		return 0, err
	}

	inserted := 0
	for _, worker := range workers {
		message := fmt.Sprintf("%s detected in %s. Automatic payout review has started for %s.", humanizeDisruptionType(disruption.Type), worker.ZoneName, fmt.Sprintf("dis_%d", disruption.ID))
		if err := s.DB.Exec(
			`INSERT INTO notifications (worker_id, type, message, created_at)
			 SELECT ?, ?, ?, ?
			 WHERE NOT EXISTS (
			   SELECT 1 FROM notifications
			   WHERE worker_id = ? AND type = ? AND message = ?
			 )`,
			worker.WorkerID, "disruption_alert", message, now.UTC(),
			worker.WorkerID, "disruption_alert", message,
		).Error; err != nil {
			return inserted, err
		}
		inserted++
	}

	return inserted, nil
}

func (s *CoreOpsService) notifyPayoutProcessed(workerID uint, claimID uint, amount float64, now time.Time) error {
	message := fmt.Sprintf("Automatic payout of Rs %.0f for claim clm_%d has been credited.", amount, claimID)
	return s.DB.Exec(
		`INSERT INTO notifications (worker_id, type, message, created_at)
		 SELECT ?, ?, ?, ?
		 WHERE NOT EXISTS (
		   SELECT 1 FROM notifications
		   WHERE worker_id = ? AND type = ? AND message = ?
		 )`,
		workerID, "payout_credited", message, now.UTC(),
		workerID, "payout_credited", message,
	).Error
}

func (s *CoreOpsService) notifyClaimGenerated(workerID uint, claimID uint, amount float64, now time.Time, disruptionType string) error {
	message := fmt.Sprintf("An automated claim of Rs %.2f has been generated for disruption %s. Review is underway.", amount, humanizeDisruptionType(disruptionType))
	return s.DB.Exec(
		`INSERT INTO notifications (worker_id, type, message, created_at)
		 SELECT ?, ?, ?, ?
		 WHERE NOT EXISTS (
		   SELECT 1 FROM notifications
		   WHERE worker_id = ? AND type = ? AND message = ?
		 )`,
		workerID, "claim_generated", message, now.UTC(),
		workerID, "claim_generated", message,
	).Error
}

func humanizeDisruptionType(raw string) string {
	parts := strings.Split(strings.TrimSpace(raw), " + ")
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" || candidate == "demand_drop" {
			continue
		}
		return strings.ToTitle(strings.ReplaceAll(candidate, "_", " "))
	}
	return "Disruption"
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
	if err := cleanupSyntheticData(s.DB); err != nil {
		return nil, err
	}

	zones := syntheticZones()
	for i := range zones {
		zones[i].RiskRating = syntheticZoneRisk(scenario, rng, i)
	}
	zoneRiskByID := map[uint]float64{}
	for i := range zones {
		var existing models.Zone
		err := s.DB.Where("name = ?", zones[i].Name).First(&existing).Error
		if err == nil {
			existing.City = zones[i].City
			existing.State = zones[i].State
			existing.Level = zones[i].Level
			existing.RiskRating = zones[i].RiskRating
			if err := s.DB.Save(&existing).Error; err != nil {
				return nil, err
			}
			zones[i].ID = existing.ID
			zoneRiskByID[existing.ID] = existing.RiskRating
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err := s.DB.Create(&zones[i]).Error; err != nil {
			return nil, err
		}
		zoneRiskByID[zones[i].ID] = zones[i].RiskRating
	}

	historyWeeks := req.WeeksHistory
	if historyWeeks <= 0 {
		historyWeeks = 8
	}

	claimTarget := req.ClaimCount
	if claimTarget <= 0 {
		claimTarget = 2000
	}

	users := make([]models.User, 0, 500)
	workerSpecs, err := buildSyntheticWorkerSpecs(req.AccountOptions, zones, zoneRiskByID, seed, rng)
	if err != nil {
		return nil, err
	}
	if len(workerSpecs) == 0 {
		return nil, fmt.Errorf("at least one synthetic account is required")
	}
	profiles := make([]models.WorkerProfile, 0, 500)
	baselines := make([]models.EarningsBaseline, 0, 500)
	policies := make([]models.Policy, 0, 500)
	weeklySummaries := make([]models.WeeklyEarningsSummary, 0, 4000)
	premiumPayments := make([]models.PremiumPayment, 0, 4000)

	weekStart, weekEnd := weekBounds(now.UTC())
	currentCycle := models.WeeklyPolicyCycle{CycleID: fmt.Sprintf("syn_cycle_%s", runID), WeekStart: weekStart, WeekEnd: weekEnd, Status: "seeded"}
	if err := s.DB.Create(&currentCycle).Error; err != nil {
		return nil, err
	}

	for i := 0; i < len(workerSpecs); i++ {
		serial := i + 1
		users = append(users, models.User{
			Phone: fmt.Sprintf("+9187%08d", (seed*1000+serial)%100000000),
			Email: fmt.Sprintf("synthetic+%s-%04d@%s", runID, serial, syntheticEmailDomain),
			Role:  "worker",
		})
	}

	if err := s.DB.Create(&users).Error; err != nil {
		return nil, err
	}

	for idx, user := range users {
		spec := workerSpecs[idx]
		zoneRisk := zoneRiskByID[spec.ZoneID]
		baseline := spec.BaselineMin + rng.Float64()*(spec.BaselineMax-spec.BaselineMin)
		if baseline <= 0 {
			baseline = 3200 + float64(rng.Intn(2200))
		}
		namePrefix := strings.TrimSpace(spec.Label)
		if namePrefix == "" {
			namePrefix = "Synthetic Worker"
		}
		policyStatus := "inactive"
		if spec.PremiumEnabled {
			policyStatus = "active"
		}
		policyPremium := spec.WeeklyPremium
		if policyPremium <= 0 {
			policyPremium = round2(18 + zoneRisk*8)
		}
		profiles = append(profiles, models.WorkerProfile{WorkerID: user.ID, Name: fmt.Sprintf("%s %03d", namePrefix, idx+1), ZoneID: spec.ZoneID, VehicleType: spec.VehicleType, UPIId: fmt.Sprintf("synthetic%03d@upi", idx+1), AQIZone: spec.AQIZone, TotalEarningsLifetime: baseline * 18})
		baselines = append(baselines, models.EarningsBaseline{WorkerID: user.ID, BaselineAmount: round2(baseline), LastUpdatedAt: now.UTC()})
		policies = append(policies, models.Policy{WorkerID: user.ID, Status: policyStatus, PremiumAmount: policyPremium, PolicyCycleID: currentCycle.ID})
		for weekOffset := 0; weekOffset < historyWeeks; weekOffset++ {
			historyStart := weekStart.AddDate(0, 0, -7*weekOffset)
			historyEnd := historyStart.AddDate(0, 0, 6)
			earnings := baseline * (0.82 + rng.Float64()*0.26)
			if scenario == "severe_disruption" && weekOffset == 0 && (idx+1)%5 == 0 {
				earnings *= 0.42
			}
			if scenario == "fraud_burst" && weekOffset == 0 && (idx+1)%11 == 0 {
				earnings *= 0.33
			}
			weeklySummaries = append(weeklySummaries, models.WeeklyEarningsSummary{WorkerID: user.ID, WeekStart: historyStart, WeekEnd: historyEnd, TotalEarnings: round2(earnings), ClaimEligible: weekOffset == 0})
			if spec.PremiumEnabled {
				amount := spec.WeeklyPremium
				if amount <= 0 {
					amount = round2(15 + zoneRisk*10)
				}
				premiumPayments = append(premiumPayments, models.PremiumPayment{WorkerID: user.ID, PolicyCycleID: currentCycle.ID, Amount: amount, Status: "completed", IdempotencyKey: fmt.Sprintf("seed_%s_%d_%d", runID, idx+1, weekOffset), Date: historyStart})
			}
		}
	}
	if err := s.DB.Create(&profiles).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Create(&baselines).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Create(&policies).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Create(&weeklySummaries).Error; err != nil {
		return nil, err
	}
	if len(premiumPayments) > 0 {
		if err := s.DB.Create(&premiumPayments).Error; err != nil {
			return nil, err
		}
	}

	disruptions := make([]models.Disruption, 0, len(zones)*2)
	disruptionsPerZone := req.DisruptionsPerZone
	if disruptionsPerZone <= 0 {
		disruptionsPerZone = 2
		if scenario == "severe_disruption" || scenario == "fraud_burst" {
			disruptionsPerZone = 3
		}
	}
	for _, zone := range zones {
		for idx := 0; idx < disruptionsPerZone; idx++ {
			start := now.UTC().Add(-time.Duration((idx+1)*12) * time.Hour)
			confirmed := start.Add(15 * time.Minute)
			disruptions = append(disruptions, models.Disruption{ZoneID: zone.ID, Type: "synthetic_" + disruptionTypeForScenario(scenario, idx), Severity: severityForScenario(scenario, rng), Confidence: round2(0.72 + rng.Float64()*0.24), Status: "confirmed", SignalTimestamp: &start, ConfirmedAt: &confirmed, StartTime: &start})
		}
	}
	if err := s.DB.Create(&disruptions).Error; err != nil {
		return nil, err
	}

	claims := make([]models.Claim, 0, claimTarget)
	scores := make([]models.ClaimFraudScore, 0, claimTarget)
	payouts := make([]models.Payout, 0, 1200)

	for claimNo := 0; claimNo < claimTarget; claimNo++ {
		worker := profiles[rng.Intn(len(profiles))]
		disruption := disruptions[rng.Intn(len(disruptions))]
		for disruption.ZoneID != worker.ZoneID {
			disruption = disruptions[rng.Intn(len(disruptions))]
		}
		isFlagged := syntheticFraudFlag(scenario, worker.WorkerID, rng)
		status := "approved"
		verdict := "clear"
		if claimNo%4 == 0 {
			status = "pending"
		}
		if isFlagged {
			status = "manual_review"
			verdict = "flagged"
		}
		claims = append(claims, models.Claim{DisruptionID: disruption.ID, WorkerID: worker.WorkerID, ClaimAmount: round2(280 + rng.Float64()*900), Status: status, FraudVerdict: verdict, CreatedAt: now.UTC().Add(-time.Duration(rng.Intn(240)) * time.Hour), UpdatedAt: now.UTC()})
	}
	if err := s.DB.Create(&claims).Error; err != nil {
		return nil, err
	}

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
		scores = append(scores, models.ClaimFraudScore{ClaimID: claim.ID, IsolationForestScore: round2(score), DbscanScore: round2(score), FinalVerdict: finalVerdict, RuleViolations: factors})
		if claim.Status == "approved" {
			processedAt := now.UTC().Add(-30 * time.Minute)
			payouts = append(payouts, models.Payout{ClaimID: claim.ID, WorkerID: claim.WorkerID, Amount: round2(claim.ClaimAmount * 0.9), Status: "processed", IdempotencyKey: fmt.Sprintf("pay_clm_%d", claim.ID), RetryCount: 1, RazorpayID: fmt.Sprintf("rzp_seed_%d", claim.ID), RazorpayStatus: "processed", ProcessedAt: &processedAt})
		}
	}
	if err := s.DB.Create(&scores).Error; err != nil {
		return nil, err
	}
	if len(payouts) > 0 {
		if err := s.DB.Create(&payouts).Error; err != nil {
			return nil, err
		}
	}

	sqlPath := filepath.Join(outputDir, "seed.sql")
	workersCSV := filepath.Join(outputDir, "workers.csv")
	claimsCSV := filepath.Join(outputDir, "claims.csv")
	payoutsCSV := filepath.Join(outputDir, "payouts.csv")
	if err := writeSyntheticSQL(sqlPath, zones, profiles, claims, payouts); err != nil {
		return nil, err
	}
	if err := writeWorkersCSV(workersCSV, profiles); err != nil {
		return nil, err
	}
	if err := writeClaimsCSV(claimsCSV, claims); err != nil {
		return nil, err
	}
	if err := writePayoutsCSV(payoutsCSV, payouts); err != nil {
		return nil, err
	}

	run := models.SyntheticGenerationRun{RunID: runID, Seed: seed, Scenario: scenario, OutputDir: outputDir, WorkersCreated: len(profiles), ZonesCreated: len(zones), DisruptionsCreated: len(disruptions), ClaimsCreated: len(claims), PayoutsCreated: len(payouts), Status: "completed"}
	if err := s.DB.Create(&run).Error; err != nil {
		return nil, err
	}

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

func cleanupSyntheticData(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	var syntheticUserIDs []uint
	if db.Migrator().HasTable("users") {
		if err := db.Table("users").Where("email LIKE ?", "%@"+syntheticEmailDomain).Pluck("id", &syntheticUserIDs).Error; err != nil {
			return err
		}
	}

	if len(syntheticUserIDs) > 0 {
		if err := deleteByWorkerIDs(db, syntheticUserIDs); err != nil {
			return err
		}
	}

	if db.Migrator().HasTable("disruption_signals") {
		if err := db.Exec("DELETE FROM disruption_signals WHERE disruption_id IN (SELECT id FROM disruptions WHERE type LIKE ?)", "synthetic_%").Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("disruption_eligibility") {
		if err := db.Exec("DELETE FROM disruption_eligibility WHERE disruption_id IN (SELECT id FROM disruptions WHERE type LIKE ?)", "synthetic_%").Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("disruptions") {
		if err := db.Exec("DELETE FROM disruptions WHERE type LIKE ?", "synthetic_%").Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("weekly_policy_cycles") {
		if err := db.Exec("DELETE FROM weekly_policy_cycles WHERE cycle_id LIKE ?", "syn_cycle_%").Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("synthetic_generation_runs") {
		if err := db.Exec("DELETE FROM synthetic_generation_runs WHERE run_id LIKE ?", "syn_%").Error; err != nil {
			return err
		}
	}

	return nil
}

func deleteByWorkerIDs(db *gorm.DB, workerIDs []uint) error {
	if len(workerIDs) == 0 {
		return nil
	}

	if db.Migrator().HasTable("payout_attempts") {
		if err := db.Exec("DELETE FROM payout_attempts WHERE payout_id IN (SELECT id FROM payouts WHERE worker_id IN ?)", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("claim_audit_logs") {
		if err := db.Exec("DELETE FROM claim_audit_logs WHERE claim_id IN (SELECT id FROM claims WHERE worker_id IN ?)", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("maintenance_check") {
		if err := db.Exec("DELETE FROM maintenance_check WHERE claim_id IN (SELECT id FROM claims WHERE worker_id IN ?)", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("claim_fraud_scores") {
		if err := db.Exec("DELETE FROM claim_fraud_scores WHERE claim_id IN (SELECT id FROM claims WHERE worker_id IN ?)", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("payouts") {
		if err := db.Exec("DELETE FROM payouts WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("claims") {
		if err := db.Exec("DELETE FROM claims WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("premium_payments") {
		if err := db.Exec("DELETE FROM premium_payments WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("weekly_earnings_summary") {
		if err := db.Exec("DELETE FROM weekly_earnings_summary WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("earnings_records") {
		if err := db.Exec("DELETE FROM earnings_records WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("earnings_baseline") {
		if err := db.Exec("DELETE FROM earnings_baseline WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("orders") {
		if err := db.Exec("DELETE FROM orders WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("policies") {
		if err := db.Exec("DELETE FROM policies WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("worker_zone_history") {
		if err := db.Exec("DELETE FROM worker_zone_history WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("worker_profiles") {
		if err := db.Exec("DELETE FROM worker_profiles WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("auth_tokens") {
		if err := db.Exec("DELETE FROM auth_tokens WHERE user_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("fcm_tokens") {
		if err := db.Exec("DELETE FROM fcm_tokens WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("notifications") {
		if err := db.Exec("DELETE FROM notifications WHERE worker_id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}
	if db.Migrator().HasTable("users") {
		if err := db.Exec("DELETE FROM users WHERE id IN ?", workerIDs).Error; err != nil {
			return err
		}
	}

	return nil
}

func buildSyntheticWorkerSpecs(options []SyntheticAccountOption, zones []models.Zone, zoneRiskByID map[uint]float64, seed int, rng *rand.Rand) ([]syntheticWorkerSpec, error) {
	defaultVehicles := []string{"two_wheeler", "bike", "scooter"}

	if len(options) == 0 {
		totalWorkers := 500
		workersPerZone := totalWorkers / len(zones)
		specs := make([]syntheticWorkerSpec, 0, totalWorkers)
		for zoneIdx, zone := range zones {
			assignments := workersPerZone
			if zoneIdx < totalWorkers%len(zones) {
				assignments++
			}
			for i := 0; i < assignments; i++ {
				specs = append(specs, syntheticWorkerSpec{
					Label:          "Synthetic Worker",
					ZoneID:         zone.ID,
					VehicleType:    defaultVehicles[(len(specs)+1)%len(defaultVehicles)],
					PremiumEnabled: true,
					WeeklyPremium:  0,
					AQIZone:        "medium",
					BaselineMin:    3200,
					BaselineMax:    5400,
				})
			}
		}
		return specs, nil
	}

	var specs []syntheticWorkerSpec
	for idx, option := range options {
		if option.Count <= 0 {
			continue
		}

		vehicle := strings.TrimSpace(option.VehicleType)
		if vehicle == "" {
			vehicle = defaultVehicles[(idx+seed)%len(defaultVehicles)]
		}
		aqi := strings.TrimSpace(option.AQIZone)
		if aqi == "" {
			aqi = "medium"
		}

		baselineMin := option.BaselineMin
		baselineMax := option.BaselineMax
		if baselineMin <= 0 {
			baselineMin = 3200
		}
		if baselineMax <= 0 {
			baselineMax = 5400
		}
		if baselineMax < baselineMin {
			baselineMax = baselineMin
		}

		premiumEnabled := true
		if option.PremiumEnabled != nil {
			premiumEnabled = *option.PremiumEnabled
		}

		zonePool := zones
		if lvl := strings.ToUpper(strings.TrimSpace(option.ZoneLevel)); lvl != "" {
			zonePool = make([]models.Zone, 0)
			for _, z := range zones {
				if strings.EqualFold(strings.TrimSpace(z.Level), lvl) {
					zonePool = append(zonePool, z)
				}
			}
			if len(zonePool) == 0 {
				return nil, fmt.Errorf("account_options[%d]: no zones found for zone_level=%s", idx, lvl)
			}
		}

		for j := 0; j < option.Count; j++ {
			zone := zonePool[rng.Intn(len(zonePool))]
			weeklyPremium := option.WeeklyPremium
			if weeklyPremium <= 0 {
				weeklyPremium = round2(18 + zoneRiskByID[zone.ID]*8)
			}

			specs = append(specs, syntheticWorkerSpec{
				Label:          strings.TrimSpace(option.Label),
				ZoneID:         zone.ID,
				VehicleType:    vehicle,
				PremiumEnabled: premiumEnabled,
				WeeklyPremium:  weeklyPremium,
				AQIZone:        aqi,
				BaselineMin:    baselineMin,
				BaselineMax:    baselineMax,
			})
		}
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("account_options did not produce any users; provide count > 0")
	}

	return specs, nil
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
		if rng.Float64() > 0.2 {
			return "high"
		}
		return "medium"
	case "fraud_burst":
		return "medium"
	default:
		if rng.Float64() > 0.7 {
			return "high"
		}
		return "medium"
	}
}

func disruptionTypeForScenario(scenario string, idx int) string {
	switch scenario {
	case "mild_disruption":
		return "order_drop"
	case "severe_disruption":
		if idx%2 == 0 {
			return "heavy_rain"
		}
		return "flood"
	case "fraud_burst":
		return "order_drop"
	default:
		return "heavy_rain"
	}
}

func syntheticFraudFlag(scenario string, workerID uint, rng *rand.Rand) bool {
	rate := 0.12
	if scenario == "fraud_burst" {
		rate = 0.18
	}
	if workerID%17 == 0 {
		return true
	}
	return rng.Float64() < rate
}

func writeSyntheticSQL(path string, zones []models.Zone, profiles []models.WorkerProfile, claims []models.Claim, payouts []models.Payout) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
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
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"worker_id", "name", "zone_id", "vehicle_type", "total_earnings_lifetime"}); err != nil {
		return err
	}
	for _, profile := range profiles {
		if err := w.Write([]string{strconv.Itoa(int(profile.WorkerID)), profile.Name, strconv.Itoa(int(profile.ZoneID)), profile.VehicleType, fmt.Sprintf("%.2f", profile.TotalEarningsLifetime)}); err != nil {
			return err
		}
	}
	return nil
}

func writeClaimsCSV(path string, claims []models.Claim) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"claim_id", "disruption_id", "worker_id", "claim_amount", "status", "fraud_verdict"}); err != nil {
		return err
	}
	for _, claim := range claims {
		if err := w.Write([]string{strconv.Itoa(int(claim.ID)), strconv.Itoa(int(claim.DisruptionID)), strconv.Itoa(int(claim.WorkerID)), fmt.Sprintf("%.2f", claim.ClaimAmount), claim.Status, claim.FraudVerdict}); err != nil {
			return err
		}
	}
	return nil
}

func writePayoutsCSV(path string, payouts []models.Payout) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"payout_id", "claim_id", "worker_id", "amount", "status", "retry_count"}); err != nil {
		return err
	}
	for _, payout := range payouts {
		if err := w.Write([]string{strconv.Itoa(int(payout.ID)), strconv.Itoa(int(payout.ClaimID)), strconv.Itoa(int(payout.WorkerID)), fmt.Sprintf("%.2f", payout.Amount), payout.Status, strconv.Itoa(payout.RetryCount)}); err != nil {
			return err
		}
	}
	return nil
}

func escapeSQL(value string) string { return strings.ReplaceAll(value, "'", "''") }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
