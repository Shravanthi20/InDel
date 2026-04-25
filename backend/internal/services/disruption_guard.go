package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DisruptionGuardResult holds the outcome of an adversarial-selection check.
type DisruptionGuardResult struct {
	Blocked      bool
	Reason       string // "active_disruption" | "predicted_disruption"
	DisruptionID uint
	Type         string
	Severity     string
}

// IsDisruptionActiveOrPredicted checks whether the given zone has an active
// or soon-to-be-predicted disruption, which would block a new policy purchase.
//
// Checks performed:
//  1. status IN ('active','confirmed') → zone is currently disrupted
//  2. status IN ('pending','predicted') AND signal_timestamp between now and now+lookaheadHours
//     → a disruption signal has been filed and is expected to hit soon
//
// Returns (result, error). result.Blocked = true means purchase should be denied.
func IsDisruptionActiveOrPredicted(db *gorm.DB, zoneID uint, lookaheadHours int) (DisruptionGuardResult, error) {
	if db == nil {
		// No DB — fail open (don't block purchases if we can't check)
		return DisruptionGuardResult{Blocked: false}, nil
	}

	type row struct {
		ID       uint   `gorm:"column:id"`
		Status   string `gorm:"column:status"`
		Type     string `gorm:"column:type"`
		Severity string `gorm:"column:severity"`
	}

	now := time.Now().UTC()

	// 1. Check for actively confirmed disruptions in the zone
	var activeRow row
	err := db.Raw(`
		SELECT id, status, type, severity
		FROM disruptions
		WHERE zone_id = ?
		  AND status IN ('active', 'confirmed')
		ORDER BY id DESC
		LIMIT 1
	`, zoneID).Scan(&activeRow).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return DisruptionGuardResult{}, fmt.Errorf("disruption guard query failed: %w", err)
	}
	if activeRow.ID != 0 {
		return DisruptionGuardResult{
			Blocked:      true,
			Reason:       "active_disruption",
			DisruptionID: activeRow.ID,
			Type:         activeRow.Type,
			Severity:     activeRow.Severity,
		}, nil
	}

	// 2. Check for predicted/pending disruptions within the lookahead window
	if lookaheadHours <= 0 {
		lookaheadHours = 12
	}
	horizon := now.Add(time.Duration(lookaheadHours) * time.Hour)

	var predictedRow row
	err = db.Raw(`
		SELECT id, status, type, severity
		FROM disruptions
		WHERE zone_id = ?
		  AND status IN ('pending', 'predicted')
		  AND signal_timestamp IS NOT NULL
		  AND signal_timestamp >= ?
		  AND signal_timestamp <= ?
		ORDER BY signal_timestamp ASC
		LIMIT 1
	`, zoneID, now, horizon).Scan(&predictedRow).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return DisruptionGuardResult{}, fmt.Errorf("disruption prediction query failed: %w", err)
	}
	if predictedRow.ID != 0 {
		return DisruptionGuardResult{
			Blocked:      true,
			Reason:       "predicted_disruption",
			DisruptionID: predictedRow.ID,
			Type:         predictedRow.Type,
			Severity:     predictedRow.Severity,
		}, nil
	}

	return DisruptionGuardResult{Blocked: false}, nil
}
