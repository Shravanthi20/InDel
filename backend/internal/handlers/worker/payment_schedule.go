package worker

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	weeklyPaymentCycle = 7 * 24 * time.Hour
	expiryDelay        = 7 * 24 * time.Hour
)

type paymentScheduleState struct {
	PaymentStatus       string
	DaysSinceLastPay    int
	NextPaymentEnabled  bool
	CoverageStatus      string
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
		LastPaymentRecorded: &lastPayment,
	}

	if elapsed >= weeklyPaymentCycle {
		state.PaymentStatus = "Eligible"
		state.NextPaymentEnabled = true
	}
	if elapsed >= weeklyPaymentCycle+expiryDelay {
		state.PaymentStatus = "Expired"
		state.NextPaymentEnabled = true
		state.CoverageStatus = "Expired"
	}

	return state
}

func ensureWorkerPaymentsTable() {
	if !hasDB() {
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
		return paymentScheduleState{
			PaymentStatus:      "Eligible",
			DaysSinceLastPay:   0,
			NextPaymentEnabled: true,
			CoverageStatus:     "Expired",
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

func parseLastPaymentFromPolicy(policy map[string]any) (time.Time, bool) {
	raw, exists := policy["last_payment_timestamp"]
	if !exists || raw == nil {
		return time.Time{}, false
	}

	switch v := raw.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return time.Time{}, false
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, true
		}
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t, true
		}
	case time.Time:
		return v, true
	}
	return time.Time{}, false
}

func paymentStateFromInMemoryPolicy(policy map[string]any, now time.Time) paymentScheduleState {
	lastPayment, ok := parseLastPaymentFromPolicy(policy)
	if !ok {
		return paymentScheduleState{
			PaymentStatus:      "Eligible",
			DaysSinceLastPay:   0,
			NextPaymentEnabled: true,
			CoverageStatus:     "Expired",
		}
	}

	return evaluatePaymentSchedule(lastPayment, now)
}

func applyPaymentStateToPolicy(policy map[string]any, state paymentScheduleState) {
	policy["payment_status"] = state.PaymentStatus
	policy["days_since_last_payment"] = state.DaysSinceLastPay
	policy["next_payment_enabled"] = state.NextPaymentEnabled
	policy["coverage_status"] = state.CoverageStatus
	if state.LastPaymentRecorded != nil {
		policy["last_payment_timestamp"] = state.LastPaymentRecorded.UTC().Format(time.RFC3339)
	}
}

func paymentLockError(state paymentScheduleState) string {
	return fmt.Sprintf("payment_locked_until_weekly_cycle_complete(days_since_last_payment=%d)", state.DaysSinceLastPay)
}
