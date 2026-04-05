package worker

import (
	"fmt"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/gin-gonic/gin"
)

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
				quote, _ := services.QuotePremium(workerDB, workerIDUint, time.Now().UTC())
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

				dueDate := p.CreatedAt.AddDate(0, 0, 7).Format("2006-01-02")
				var lastPaymentDate time.Time
				if err := workerDB.Table("premium_payments").Select("payment_date").Where("policy_id = ?", p.ID).Order("payment_date DESC").Limit(1).Scan(&lastPaymentDate).Error; err == nil && !lastPaymentDate.IsZero() {
					// If they made a payment, due date is exactly 7 days after their last payment
					dueDate = lastPaymentDate.AddDate(0, 0, 7).Format("2006-01-02")
				} else if p.Status == "active" {
					// Fallback if no payment recorded but active
					if p.CreatedAt.AddDate(0, 0, 7).Before(time.Now()) {
						dueDate = time.Now().AddDate(0, 0, 7).Format("2006-01-02")
					}
				}

				policy := gin.H{
					"policy_id":          fmt.Sprintf("pol-%03d", p.ID),
					"status":             p.Status,
					"weekly_premium_inr": premiumAmount,
					"coverage_ratio":     coverageRatio,
					"zone":               "Tambaram, Chennai",
					"next_due_date":      dueDate,
					"risk_score":         riskScore,
					"pricing_source":     source,
					"model_version":      modelVersion,
					"shap_breakdown":     breakdown,
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

// EnrollPolicy enrolls in coverage
func EnrollPolicy(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			premiumAmount := 22.0
			if quote, err := services.QuotePremium(workerDB, workerIDUint, time.Now().UTC()); err == nil && quote != nil && quote.WeeklyPremiumINR > 0 {
				premiumAmount = quote.WeeklyPremiumINR
			}
			policy := models.Policy{WorkerID: workerIDUint, Status: "active", PremiumAmount: premiumAmount}
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
