package worker

import (
	"fmt"

	"github.com/Shravanthi20/InDel/backend/internal/models"
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
				policy := gin.H{
					"policy_id":          fmt.Sprintf("pol-%03d", p.ID),
					"status":             p.Status,
					"weekly_premium_inr": int(p.PremiumAmount),
					"coverage_ratio":     0.8,
					"zone":               "Tambaram, Chennai",
					"next_due_date":      "2026-03-30",
					"shap_breakdown": []gin.H{
						{"feature": "rain_risk", "impact": 0.42},
						{"feature": "order_drop_volatility", "impact": 0.31},
						{"feature": "historical_disruptions", "impact": 0.27},
					},
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
