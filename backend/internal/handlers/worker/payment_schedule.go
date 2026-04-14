package worker

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	weeklyPaymentCycle = 7 * 24 * time.Hour
	gracePeriodWindow  = 2 * 24 * time.Hour
	initialMultiplier  = 2
)

type paymentScheduleState struct {
	PaymentStatus       string
	DaysSinceLastPay    int
	NextPaymentEnabled  bool
	CoverageStatus      string
	LateFeeINR          int
	RequiredAmountINR   int
	GraceDaysRemaining  int
	BillingCycleDays    int
	GracePeriodDays     int
	InitialMultiplier   int
	LastPaymentRecorded *time.Time
}

var ensureWorkerPaymentsTableOnce sync.Once

func evaluatePaymentSchedule(lastPayment time.Time, now time.Time) paymentScheduleState {
	elapsed := now.Sub(lastPayment)
	daysSince := int(elapsed.Hours() / 24)
	if daysSince < 0 {
		daysSince = 0
	}

	state := paymentScheduleState{
		PaymentStatus:       "Locked",
		DaysSinceLastPay:    daysSince,
		NextPaymentEnabled:  false,
		CoverageStatus:      "Active",
		LateFeeINR:          0,
		RequiredAmountINR:   0,
		GraceDaysRemaining:  0,
		BillingCycleDays:    int(weeklyPaymentCycle.Hours() / 24),
		GracePeriodDays:     int(gracePeriodWindow.Hours() / 24),
		InitialMultiplier:   initialMultiplier,
		LastPaymentRecorded: &lastPayment,
	}

	if elapsed >= weeklyPaymentCycle && elapsed < weeklyPaymentCycle+gracePeriodWindow {
		state.PaymentStatus = "Eligible"
		state.NextPaymentEnabled = true
		daysLate := daysSince - state.BillingCycleDays
		if daysLate < 0 {
			daysLate = 0
		}
		if daysLate > state.GracePeriodDays {
			daysLate = state.GracePeriodDays
		}
		state.LateFeeINR = daysLate
		state.GraceDaysRemaining = state.GracePeriodDays - daysLate
		if state.GraceDaysRemaining < 0 {
			state.GraceDaysRemaining = 0
		}
	}
	if elapsed >= weeklyPaymentCycle+gracePeriodWindow {
		state.PaymentStatus = "Deactivated"
		state.NextPaymentEnabled = false
		state.CoverageStatus = "Deactivated"
		state.LateFeeINR = state.GracePeriodDays
		state.GraceDaysRemaining = 0
	}

	return state
}

func ensureWorkerPaymentsTable() {
	if !HasDB() {
		return
	}

	ensureWorkerPaymentsTableOnce.Do(func() {
		_ = workerDB.Exec(`
			CREATE TABLE IF NOT EXISTS worker_payments (
				worker_id INTEGER PRIMARY KEY REFERENCES users(id),
				last_payment_timestamp TIMESTAMP NOT NULL,
				next_payment_enabled BOOLEAN NOT NULL DEFAULT FALSE,
				coverage_status VARCHAR(20) NOT NULL DEFAULT 'Active',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`).Error
	})
}

func getOrBootstrapPaymentSchedule(workerID uint, now time.Time) (paymentScheduleState, error) {
	ensureWorkerPaymentsTable()

	type row struct {
		LastPaymentTimestamp time.Time `gorm:"column:last_payment_timestamp"`
		NextPaymentEnabled   bool      `gorm:"column:next_payment_enabled"`
		CoverageStatus       string    `gorm:"column:coverage_status"`
	}

	var r row
	err := workerDB.Raw(`
		SELECT last_payment_timestamp, next_payment_enabled, coverage_status
		FROM worker_payments
		WHERE worker_id = ?
		LIMIT 1
	`, workerID).Scan(&r).Error
	if err != nil {
		return paymentScheduleState{}, err
	}
	if !r.LastPaymentTimestamp.IsZero() {
		state := evaluatePaymentSchedule(r.LastPaymentTimestamp, now)
		if !strings.EqualFold(state.CoverageStatus, r.CoverageStatus) || state.NextPaymentEnabled != r.NextPaymentEnabled {
			_ = upsertPaymentSchedule(workerID, r.LastPaymentTimestamp, state.NextPaymentEnabled, state.CoverageStatus)
		}
		return state, nil
	}

	var fallback struct {
		PaymentDate time.Time `gorm:"column:payment_date"`
	}
	_ = workerDB.Raw(`
		SELECT payment_date
		FROM premium_payments
		WHERE worker_id = ? AND status = 'completed'
		ORDER BY payment_date DESC
		LIMIT 1
	`, workerID).Scan(&fallback).Error

	if fallback.PaymentDate.IsZero() {
		var policyFallback struct {
			Status    string    `gorm:"column:status"`
			CreatedAt time.Time `gorm:"column:created_at"`
		}
		_ = workerDB.Raw(`
			SELECT status, created_at
			FROM policies
			WHERE worker_id = ?
			ORDER BY id DESC
			LIMIT 1
		`, workerID).Scan(&policyFallback).Error

		if strings.EqualFold(policyFallback.Status, "active") && !policyFallback.CreatedAt.IsZero() {
			state := evaluatePaymentSchedule(policyFallback.CreatedAt, now)
			_ = upsertPaymentSchedule(workerID, policyFallback.CreatedAt, state.NextPaymentEnabled, state.CoverageStatus)
			return state, nil
		}

		return paymentScheduleState{
			PaymentStatus:      "Eligible",
			DaysSinceLastPay:   0,
			NextPaymentEnabled: true,
			CoverageStatus:     "NeedsActivation",
			BillingCycleDays:   int(weeklyPaymentCycle.Hours() / 24),
			GracePeriodDays:    int(gracePeriodWindow.Hours() / 24),
			InitialMultiplier:  initialMultiplier,
		}, nil
	}

	state := evaluatePaymentSchedule(fallback.PaymentDate, now)
	_ = upsertPaymentSchedule(workerID, fallback.PaymentDate, state.NextPaymentEnabled, state.CoverageStatus)
	return state, nil
}

func upsertPaymentSchedule(workerID uint, lastPayment time.Time, nextEnabled bool, coverageStatus string) error {
	ensureWorkerPaymentsTable()
	return workerDB.Exec(`
		INSERT INTO worker_payments (worker_id, last_payment_timestamp, next_payment_enabled, coverage_status, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (worker_id) DO UPDATE SET
			last_payment_timestamp = EXCLUDED.last_payment_timestamp,
			next_payment_enabled = EXCLUDED.next_payment_enabled,
			coverage_status = EXCLUDED.coverage_status,
			updated_at = CURRENT_TIMESTAMP
	`, workerID, lastPayment, nextEnabled, coverageStatus).Error
}

func applyPaymentStateToPolicy(policy map[string]any, state paymentScheduleState) {
	policy["payment_status"] = state.PaymentStatus
	policy["days_since_last_payment"] = state.DaysSinceLastPay
	policy["next_payment_enabled"] = state.NextPaymentEnabled
	policy["coverage_status"] = state.CoverageStatus
	policy["late_fee_inr"] = state.LateFeeINR
	policy["required_payment_inr"] = state.RequiredAmountINR
	policy["grace_days_remaining"] = state.GraceDaysRemaining
	policy["billing_cycle_days"] = state.BillingCycleDays
	policy["grace_period_days"] = state.GracePeriodDays
	policy["initial_payment_multiplier"] = state.InitialMultiplier
	if state.LastPaymentRecorded != nil {
		policy["last_payment_timestamp"] = state.LastPaymentRecorded.UTC().Format(time.RFC3339)
	}
}

func paymentLockError(state paymentScheduleState) string {
	return fmt.Sprintf("payment_locked_until_weekly_cycle_complete(days_since_last_payment=%d)", state.DaysSinceLastPay)
}

func syncPolicyStatusWithPaymentState(workerID uint, state paymentScheduleState) {
	if !HasDB() {
		return
	}

	if strings.EqualFold(state.CoverageStatus, "Deactivated") {
		_ = workerDB.Exec(
			"UPDATE policies SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP WHERE worker_id = ? AND status <> 'cancelled'",
			workerID,
		).Error
	}
}
