package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/events"
	"github.com/Shravanthi20/InDel/backend/internal/kafka"
	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// lockInHours returns the configured lock-in duration (default 48h).
func lockInHours() int {
	if v := os.Getenv("POLICY_LOCKIN_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 48
}

// disruptionLookaheadHours returns the lookahead window for disruption checks (default 12h).
func disruptionLookaheadHours() int {
	if v := os.Getenv("DISRUPTION_BLOCK_LOOKAHEAD_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 12
}

// GetPolicy returns active policy
func GetPolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			var p models.Policy
			err := workerDB.Where("worker_id = ?", workerIDUint).Order("id DESC").First(&p).Error
			if err == nil {
				now := time.Now().UTC()
				quote, _ := services.QuotePremium(workerDB, workerIDUint, now)
				zoneSummary := getWorkerZoneSummary(workerIDUint)
				premiumAmount := int(p.PremiumAmount)
				source := "stored_policy"
				riskScore := 0.0
				modelVersion := "fallback_rule_v2"
				var breakdown []gin.H

				if quote != nil {
					premiumAmount = int(quote.WeeklyPremiumINR)
					source = quote.Source
					riskScore = quote.RiskScore
					modelVersion = quote.ModelVersion
					breakdown = make([]gin.H, 0, len(quote.Explainability))
					for _, item := range quote.Explainability {
						breakdown = append(breakdown, gin.H{"feature": item.Feature, "impact": item.Impact})
					}
				} else {
					// Fallback static breakdown for historical UI context
					breakdown = []gin.H{
						{"feature": "rain_risk", "impact": 0.42},
						{"feature": "order_drop_volatility", "impact": 0.31},
						{"feature": "historical_disruptions", "impact": 0.27},
					}
				}
				// Dynamic calculations for "Real Data"
				coverageRatio := 0.85
				if riskScore > 0.7 {
					coverageRatio = 0.75
				} else if riskScore < 0.3 {
					coverageRatio = 0.95
				}

				paymentState, stateErr := getOrBootstrapPaymentSchedule(workerIDUint, now)
				if stateErr == nil {
					syncPolicyStatusWithPaymentState(workerIDUint, paymentState)
				}

				dueDateBase := p.CreatedAt
				var lastPaymentDate time.Time
				if err := workerDB.Table("premium_payments").
					Select("payment_date").
					Where("worker_id = ? AND status = 'completed'", workerIDUint).
					Order("payment_date DESC").
					Limit(1).
					Scan(&lastPaymentDate).Error; err == nil && !lastPaymentDate.IsZero() {
					dueDateBase = lastPaymentDate
				}
				if stateErr == nil && paymentState.LastPaymentRecorded != nil && !paymentState.LastPaymentRecorded.IsZero() {
					dueDateBase = *paymentState.LastPaymentRecorded
				}
				dueDate := dueDateBase.AddDate(0, 0, 7).Format("2006-01-02")
				if p.Status == "active" && dueDateBase.IsZero() {
					dueDate = time.Now().AddDate(0, 0, 7).Format("2006-01-02")
				}

				effectiveStatus := p.Status
				if stateErr == nil && paymentState.CoverageStatus == "Deactivated" {
					effectiveStatus = "cancelled"
				}

				requiredAmount := 0
				if stateErr == nil {
					if paymentState.PaymentStatus == "Eligible" {
						requiredAmount = premiumAmount + paymentState.LateFeeINR
					} else if paymentState.CoverageStatus == "NeedsActivation" {
						requiredAmount = premiumAmount * initialMultiplier
					}
					paymentState.RequiredAmountINR = requiredAmount
				}

				zoneLabel := "Tambaram, Chennai"
				if zoneSummary.ZoneName != "" && zoneSummary.City != "" {
					zoneLabel = formatZoneDisplay(zoneSummary.ZoneName, zoneSummary.City)
				} else if zoneSummary.ZoneName != "" {
					zoneLabel = zoneSummary.ZoneName
				}

				policy := gin.H{
					"policy_id":          fmt.Sprintf("pol-%03d", p.ID),
					"plan_id":            p.PlanID,
					"status":             effectiveStatus,
					"weekly_premium_inr": premiumAmount,
					"coverage_ratio":     coverageRatio,
					"zone":               zoneLabel,
					"next_due_date":      dueDate,
					"risk_score":         riskScore,
					"pricing_source":     source,
					"model_version":      modelVersion,
					"shap_breakdown":     breakdown,
					"plan_info": gin.H{
						"initial_payment_rule": "first_payment_is_double_weekly_premium",
						"weekly_cycle_days":    7,
						"grace_period_days":    2,
						"late_fee_rule":        "rs_1_per_day_during_grace",
						"termination_rule":     "deactivate_after_7_plus_2_days_without_payment",
					},
				}
				// Append lock-in fields if applicable
				if p.Status == models.PolicyStatusLocked && p.LockInEndTime != nil {
					policy["lock_in_end_time"] = p.LockInEndTime.Format(time.RFC3339)
					policy["lock_in_hours_remaining"] = int(time.Until(*p.LockInEndTime).Hours())
					policy["message"] = fmt.Sprintf(
						"Your policy is in lock-in period. Claims and cancellations are available after %s.",
						p.LockInEndTime.Format("02 Jan 2006, 15:04 UTC"),
					)
				}
				if stateErr == nil {
					applyPaymentStateToPolicy(policy, paymentState)
				}
				c.JSON(200, gin.H{"policy": policy})
				return
			}
		}
	}

	store.mu.RLock()
	policy := store.data.Policy
	store.mu.RUnlock()

	c.JSON(200, gin.H{"policy": policy})
}

// EnrollPolicy enrolls in coverage with lock-in period enforcement.
func EnrollPolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			now := time.Now().UTC()

			// --- Step 1: Fetch worker zone ---
			var zoneInfo struct {
				ZoneID   uint   `gorm:"column:zone_id"`
				ZoneName string `gorm:"column:zone_name"`
			}
			_ = workerDB.Raw(`
				SELECT wp.zone_id, COALESCE(z.name, 'Unknown') AS zone_name
				FROM worker_profiles wp
				LEFT JOIN zones z ON z.id = wp.zone_id
				WHERE wp.worker_id = ?
				LIMIT 1
			`, workerIDUint).Scan(&zoneInfo)

			// --- Step 2: Disruption guard ---
			if zoneInfo.ZoneID > 0 {
				lookahead := disruptionLookaheadHours()
				guard, guardErr := services.IsDisruptionActiveOrPredicted(workerDB, zoneInfo.ZoneID, lookahead)
				if guardErr != nil {
					log.Printf("[POLICY] Disruption guard error for workerID=%d: %v", workerIDUint, guardErr)
					// Fail open — don't block purchase on guard error
				} else if guard.Blocked {
					reason := "Policy purchase restricted due to active/predicted disruption"
					if guard.Reason == "predicted_disruption" {
						reason = fmt.Sprintf(
							"Policy purchase restricted: a %s disruption is predicted in your zone within %d hours",
							guard.Type, disruptionLookaheadHours(),
						)
					}
					log.Printf("[POLICY] Purchase blocked for workerID=%d zone=%d reason=%s", workerIDUint, zoneInfo.ZoneID, guard.Reason)

					// Write audit log
					writeAuditLog(0, workerIDUint, "purchase_blocked_disruption", "", "", reason)

					// Publish purchase_blocked Kafka event
					publishPurchaseBlockedEvent(workerIDUint, zoneInfo.ZoneID, guard)

					c.JSON(http.StatusForbidden, gin.H{
						"error":         "purchase_blocked",
						"reason":        reason,
						"disruption_id": guard.DisruptionID,
					})
					return
				}
			}

			// --- Step 3: Compute premium ---
			premiumAmount := 22.0
			if quote, err := services.QuotePremium(workerDB, workerIDUint, now); err == nil && quote != nil && quote.WeeklyPremiumINR > 0 {
				premiumAmount = quote.WeeklyPremiumINR
			}

			// --- Step 4: Check for existing locked/active policy ---
			var existingPolicy models.Policy
			existsErr := workerDB.Where("worker_id = ? AND status IN ?", workerIDUint,
				[]string{models.PolicyStatusLocked, models.PolicyStatusActive}).
				Order("id DESC").First(&existingPolicy).Error
			if existsErr == nil {
				// Already has an active or locked policy — re-lock the current one
				if existingPolicy.Status == models.PolicyStatusActive {
					// Re-enrollment: start a new lock-in
					lockEnd := now.Add(time.Duration(lockInHours()) * time.Hour)
					if err := workerDB.Model(&existingPolicy).Updates(map[string]interface{}{
						"status":            models.PolicyStatusLocked,
						"premium_amount":    premiumAmount,
						"lock_in_start_time": now,
						"lock_in_end_time":  lockEnd,
						"updated_at":        now,
					}).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_renew_policy"})
						return
					}
					// Remove from active_policies during re-lock
					_ = workerDB.Exec("DELETE FROM active_policies WHERE user_id = ?", workerIDUint).Error
					publishPolicyEvents(workerIDUint, zoneInfo.ZoneID, existingPolicy.ID, premiumAmount, now, lockEnd)
					writeAuditLog(existingPolicy.ID, workerIDUint, "policy_locked", models.PolicyStatusActive, models.PolicyStatusLocked, "policy renewed — lock-in started")
					c.JSON(200, buildLockedPolicyResponse(existingPolicy.ID, existingPolicy.PlanID, premiumAmount, lockEnd))
					return
				}
				// Already locked — return current state
				c.JSON(200, buildLockedPolicyResponse(existingPolicy.ID, existingPolicy.PlanID, existingPolicy.PremiumAmount, *existingPolicy.LockInEndTime))
				return
			}

			// --- Step 5: Create new policy in LOCKED state ---
			lockStart := now
			lockEnd := now.Add(time.Duration(lockInHours()) * time.Hour)
			idempotencyKey := fmt.Sprintf("policy_enroll_%d_%d", workerIDUint, now.Unix())

			newPolicy := models.Policy{
				WorkerID:        workerIDUint,
				Status:          models.PolicyStatusLocked,
				PremiumAmount:   premiumAmount,
				LockInStartTime: &lockStart,
				LockInEndTime:   &lockEnd,
				IdempotencyKey:  idempotencyKey,
			}
			if err := workerDB.Create(&newPolicy).Error; err != nil {
				log.Printf("[POLICY] Failed to create policy for workerID=%d: %v", workerIDUint, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_policy"})
				return
			}

			log.Printf("[POLICY] ✅ New policy created policyID=%d workerID=%d status=locked lock_in_end=%s",
				newPolicy.ID, workerIDUint, lockEnd.Format(time.RFC3339))

			// Write audit log
			writeAuditLog(newPolicy.ID, workerIDUint, "policy_locked", "", models.PolicyStatusLocked, "new policy enrollment — lock-in started")

			// Publish Kafka events
			publishPolicyEvents(workerIDUint, zoneInfo.ZoneID, newPolicy.ID, premiumAmount, lockStart, lockEnd)

			c.JSON(200, buildLockedPolicyResponse(newPolicy.ID, newPolicy.PlanID, premiumAmount, lockEnd))
			return
		}
	}

	// In-memory fallback (no DB)
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

// PausePolicy pauses coverage — blocked during lock-in period
func PausePolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			// Check if policy is in lock-in period
			var p models.Policy
			if err := workerDB.Where("worker_id = ?", workerIDUint).Order("id DESC").First(&p).Error; err == nil {
				if p.IsInLockIn(time.Now().UTC()) {
					c.JSON(http.StatusLocked, gin.H{
						"error":             "policy_in_lockin",
						"reason":            "Policy cannot be paused during lock-in period",
						"lock_in_end_time":  p.LockInEndTime.Format(time.RFC3339),
					})
					return
				}
			}
			_ = workerDB.Exec("UPDATE policies SET status='paused', updated_at=CURRENT_TIMESTAMP WHERE worker_id = ?", workerIDUint).Error
			// Remove from active_policies when pausing
			_ = workerDB.Exec("DELETE FROM active_policies WHERE user_id = ?", workerIDUint).Error
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

// CancelPolicy cancels coverage — blocked during lock-in period
func CancelPolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			// Check if policy is in lock-in period
			var p models.Policy
			if err := workerDB.Where("worker_id = ?", workerIDUint).Order("id DESC").First(&p).Error; err == nil {
				if p.IsInLockIn(time.Now().UTC()) {
					c.JSON(http.StatusLocked, gin.H{
						"error":            "policy_in_lockin",
						"reason":           "Policy cannot be cancelled during lock-in period",
						"lock_in_end_time": p.LockInEndTime.Format(time.RFC3339),
					})
					return
				}
			}
			_ = workerDB.Exec("UPDATE policies SET status='cancelled', updated_at=CURRENT_TIMESTAMP WHERE worker_id = ?", workerIDUint).Error
			// Remove from active_policies
			_ = workerDB.Exec("DELETE FROM active_policies WHERE user_id = ?", workerIDUint).Error
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

// --- helpers ---

func buildLockedPolicyResponse(policyID uint, planID string, premiumAmount float64, lockEnd time.Time) gin.H {
	remaining := time.Until(lockEnd)
	hoursRemaining := int(remaining.Hours())
	if hoursRemaining < 0 {
		hoursRemaining = 0
	}
	return gin.H{
		"message": "policy_enrolled",
		"policy": gin.H{
			"policy_id":               fmt.Sprintf("pol-%03d", policyID),
			"plan_id":                 planID,
			"status":                  models.PolicyStatusLocked,
			"weekly_premium_inr":      int(premiumAmount),
			"coverage_ratio":          0.8,
			"lock_in_end_time":        lockEnd.Format(time.RFC3339),
			"lock_in_hours_remaining": hoursRemaining,
			"message": fmt.Sprintf(
				"Policy is in lock-in period. Claims, upgrades, and cancellations will be available after %s.",
				lockEnd.Format("02 Jan 2006, 15:04 UTC"),
			),
		},
	}
}

// publishPolicyEvents emits policy.created and policy.locked Kafka events.
func publishPolicyEvents(workerID, zoneID, policyID uint, premiumAmount float64, lockStart, lockEnd time.Time) {
	producer := getKafkaProducer()
	if producer == nil {
		return
	}

	now := time.Now().UTC()
	key := fmt.Sprintf("policy-%d", policyID)

	created := events.PolicyCreatedEvent{
		EventType:      "policy.created",
		PolicyID:       policyID,
		WorkerID:       workerID,
		ZoneID:         zoneID,
		PremiumAmount:  premiumAmount,
		LockInStart:    lockStart,
		LockInEnd:      lockEnd,
		IdempotencyKey: fmt.Sprintf("policy_enroll_%d_%d", workerID, lockStart.Unix()),
		Timestamp:      now,
	}
	if b, err := json.Marshal(created); err == nil {
		_ = producer.Publish(kafka.TopicPolicyCreated, key, b)
		log.Printf("[KAFKA] policy.created published policyID=%d workerID=%d", policyID, workerID)
	}

	locked := events.PolicyLockedEvent{
		EventType:   "policy.locked",
		PolicyID:    policyID,
		WorkerID:    workerID,
		LockInEnd:   lockEnd,
		LockInHours: lockInHours(),
		Timestamp:   now,
	}
	if b, err := json.Marshal(locked); err == nil {
		_ = producer.Publish(kafka.TopicPolicyLocked, key, b)
		log.Printf("[KAFKA] policy.locked published policyID=%d workerID=%d lockEnd=%s", policyID, workerID, lockEnd.Format(time.RFC3339))
	}
}

// publishPurchaseBlockedEvent emits a policy.purchase_blocked event.
func publishPurchaseBlockedEvent(workerID, zoneID uint, guard services.DisruptionGuardResult) {
	producer := getKafkaProducer()
	if producer == nil {
		return
	}

	evt := events.PurchaseBlockedEvent{
		EventType:      "policy.purchase_blocked",
		WorkerID:       workerID,
		ZoneID:         zoneID,
		DisruptionID:   guard.DisruptionID,
		Reason:         guard.Reason,
		DisruptionType: guard.Type,
		Timestamp:      time.Now().UTC(),
	}
	if b, err := json.Marshal(evt); err == nil {
		key := fmt.Sprintf("worker-%d", workerID)
		_ = producer.Publish(kafka.TopicPurchaseBlocked, key, b)
		log.Printf("[KAFKA] policy.purchase_blocked published workerID=%d zoneID=%d reason=%s", workerID, zoneID, guard.Reason)
	}
}

// writeAuditLog writes a PolicyAuditLog entry (best effort — failures are logged, not fatal).
func writeAuditLog(policyID, workerID uint, action, fromStatus, toStatus, reason string) {
	if workerDB == nil {
		return
	}
	entry := models.PolicyAuditLog{
		PolicyID:   policyID,
		WorkerID:   workerID,
		Action:     action,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		Reason:     reason,
		CreatedAt:  time.Now().UTC(),
	}
	if err := workerDB.Create(&entry).Error; err != nil {
		log.Printf("[POLICY] WARN: audit log write failed (action=%s workerID=%d): %v", action, workerID, err)
	}
}

// getKafkaProducer returns the Kafka producer wired into the worker gateway (if any).
// The producer is stored as a package-level variable set during startup.
var workerKafkaProducer *kafka.Producer

// SetKafkaProducer wires a Kafka producer into the worker handler package.
func SetKafkaProducer(p *kafka.Producer) {
	workerKafkaProducer = p
}

func getKafkaProducer() *kafka.Producer {
	return workerKafkaProducer
}
