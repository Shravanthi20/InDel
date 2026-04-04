package worker

import (
	"fmt"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
)

func inferPlanFromPremium(premium int) (string, string, int, int) {
	switch {
	case premium >= 12 && premium <= 18:
		return "plan-starter", "Seed", 10, 15
	case premium >= 19 && premium <= 26:
		return "plan-growth", "Scale", 15, 20
	case premium >= 27 && premium <= 35:
		return "plan-premium", "Soar", 20, 25
	default:
		return "", "", 0, 0
	}
}

// GetPolicy returns active policy
func GetPolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			var p models.Policy
			err := workerDB.Where("worker_id = ?", workerIDUint).Order("id DESC").First(&p).Error
			if err == nil {
				now := time.Now().UTC()
				paymentState, _ := getOrBootstrapPaymentSchedule(workerIDUint, now)
				if strings.EqualFold(paymentState.CoverageStatus, "Expired") {
					_ = workerDB.Exec("UPDATE policies SET status='expired', updated_at=CURRENT_TIMESTAMP WHERE id = ?", p.ID).Error
					p.Status = "expired"
				}
				planID, planName, rangeStart, rangeEnd := inferPlanFromPremium(int(p.PremiumAmount))
				planStatus := "selected"
				if p.Status == "skipped" {
					planStatus = "skipped"
					planID = ""
					planName = ""
					rangeStart = 0
					rangeEnd = 0
				}

				// Get worker's zone
				workerZone := "Unknown"
				type zoneRow struct {
					ZoneName string `gorm:"column:zone_name"`
					City     string `gorm:"column:city"`
					State    string `gorm:"column:state"`
				}
				var zr zoneRow
				if err := workerDB.Raw(`
					SELECT z.name AS zone_name, z.city, z.state
					FROM zones z
					INNER JOIN worker_profiles wp ON wp.zone_id = z.id
					WHERE wp.worker_id = ?
					LIMIT 1
				`, workerIDUint).Scan(&zr).Error; err == nil && zr.City != "" {
					workerZone = fmt.Sprintf("%s, %s", zr.City, zr.State)
				}

				// Calculate next due date (7 days after last payment, or "N/A" if no payment yet)
				nextDueDate := "N/A"
				if paymentState.LastPaymentRecorded != nil {
					nextDue := paymentState.LastPaymentRecorded.AddDate(0, 0, 7)
					nextDueDate = nextDue.Format("2006-01-02")
				}

				policy := gin.H{
					"policy_id":               fmt.Sprintf("pol-%03d", p.ID),
					"status":                  p.Status,
					"plan_status":             planStatus,
					"weekly_premium_inr":      int(p.PremiumAmount),
					"coverage_ratio":          0.8,
					"zone":                    workerZone,
					"next_due_date":           nextDueDate,
					"payment_status":          paymentState.PaymentStatus,
					"days_since_last_payment": paymentState.DaysSinceLastPay,
					"next_payment_enabled":    paymentState.NextPaymentEnabled,
					"coverage_status":         paymentState.CoverageStatus,
					"plan_id":                 planID,
					"plan_name":               planName,
					"range_start":             rangeStart,
					"range_end":               rangeEnd,
					"shap_breakdown": []gin.H{
						{"feature": "rain_risk", "impact": 0.42},
						{"feature": "order_drop_volatility", "impact": 0.31},
						{"feature": "historical_disruptions", "impact": 0.27},
					},
				}
				if paymentState.LastPaymentRecorded != nil {
					policy["last_payment_timestamp"] = paymentState.LastPaymentRecorded.UTC().Format(time.RFC3339)
				}
				c.JSON(200, gin.H{"policy": policy})
				return
			}
		}
	}

	store.mu.Lock()
	policy := store.data.Policy
	state := paymentStateFromInMemoryPolicy(policy, time.Now().UTC())
	applyPaymentStateToPolicy(policy, state)
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		if strings.EqualFold(state.CoverageStatus, "Expired") {
			profile["coverage_status"] = "expired"
		} else {
			profile["coverage_status"] = "active"
		}
	}
	store.mu.Unlock()

	c.JSON(200, gin.H{"policy": policy})
}

// EnrollPolicy enrolls in coverage
func EnrollPolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			policy := models.Policy{WorkerID: workerIDUint, Status: "active", PremiumAmount: 22}
			if err := workerDB.Create(&policy).Error; err == nil {
				c.JSON(200, gin.H{"message": "policy_enrolled", "policy": gin.H{
					"policy_id":          fmt.Sprintf("pol-%03d", policy.ID),
					"status":             policy.Status,
					"weekly_premium_inr": int(policy.PremiumAmount),
					"coverage_ratio":     0.8,
				}})
				return
			}
		}
	}

	store.mu.Lock()
	store.data.Policy["status"] = "active"
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		profile["coverage_status"] = "active"
		profile["enrolled"] = true
	}
	policy := store.data.Policy
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "policy_enrolled", "policy": policy})
}

// PausePolicy pauses coverage
func PausePolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec("UPDATE policies SET status='paused', updated_at=CURRENT_TIMESTAMP WHERE worker_id = ?", workerIDUint).Error
			c.JSON(200, gin.H{"message": "policy_paused", "policy": gin.H{"status": "paused"}})
			return
		}
	}

	store.mu.Lock()
	store.data.Policy["status"] = "paused"
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		profile["coverage_status"] = "paused"
	}
	policy := store.data.Policy
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "policy_paused", "policy": policy})
}

// CancelPolicy cancels coverage
func CancelPolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec("UPDATE policies SET status='cancelled', updated_at=CURRENT_TIMESTAMP WHERE worker_id = ?", workerIDUint).Error
			c.JSON(200, gin.H{"message": "policy_cancelled", "policy": gin.H{"status": "cancelled"}})
			return
		}
	}

	store.mu.Lock()
	store.data.Policy["status"] = "cancelled"
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		profile["coverage_status"] = "inactive"
		profile["enrolled"] = false
	}
	policy := store.data.Policy
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "policy_cancelled", "policy": policy})
}
