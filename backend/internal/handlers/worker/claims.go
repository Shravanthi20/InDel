package worker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClaims returns claim history
func GetClaims(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			type claimRow struct {
				ClaimID        uint    `gorm:"column:claim_id"`
				Status         string  `gorm:"column:status"`
				DisruptionType string  `gorm:"column:disruption_type"`
				Zone           string  `gorm:"column:zone"`
				IncomeLoss     float64 `gorm:"column:income_loss"`
				PayoutAmount   float64 `gorm:"column:payout_amount"`
				FraudVerdict   string  `gorm:"column:fraud_verdict"`
				CreatedAt      string  `gorm:"column:created_at"`
			}

			rows := make([]claimRow, 0)
			_ = workerDB.Raw(`
				SELECT c.id AS claim_id,
				       c.status,
				       d.type AS disruption_type,
				       COALESCE(z.name || ', ' || z.city, 'Unknown Zone') AS zone,
				       c.claim_amount AS income_loss,
				       COALESCE(p.amount, 0) AS payout_amount,
				       COALESCE(c.fraud_verdict, 'pending') AS fraud_verdict,
				       c.created_at::text AS created_at
				FROM claims c
				LEFT JOIN disruptions d ON d.id = c.disruption_id
				LEFT JOIN zones z ON z.id = d.zone_id
				LEFT JOIN payouts p ON p.claim_id = c.id
				WHERE c.worker_id = ?
				ORDER BY c.created_at DESC
			`, workerIDUint).Scan(&rows).Error

			claims := make([]gin.H, 0, len(rows))
			for _, row := range rows {
				claims = append(claims, gin.H{
					"claim_id":        fmt.Sprintf("clm-%03d", row.ClaimID),
					"status":          row.Status,
					"disruption_type": row.DisruptionType,
					"zone":            row.Zone,
					"income_loss":     int(row.IncomeLoss),
					"payout_amount":   int(row.PayoutAmount),
					"fraud_verdict":   row.FraudVerdict,
					"created_at":      row.CreatedAt,
				})
			}

			c.JSON(200, gin.H{"claims": claims})
			return
		}
	}

	store.mu.RLock()
	claims := store.data.Claims
	store.mu.RUnlock()

	c.JSON(200, gin.H{"claims": claims})
}

// GetClaimDetail returns claim details
func GetClaimDetail(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	claimID := c.Param("claim_id")

	if hasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			claimNumStr := strings.TrimPrefix(strings.TrimSpace(claimID), "clm-")
			claimNumID, parseClaimErr := strconv.ParseUint(claimNumStr, 10, 64)
			if parseClaimErr == nil {
				type detailRow struct {
					ClaimID        uint    `gorm:"column:claim_id"`
					Status         string  `gorm:"column:status"`
					DisruptionType string  `gorm:"column:disruption_type"`
					Zone           string  `gorm:"column:zone"`
					IncomeLoss     float64 `gorm:"column:income_loss"`
					PayoutAmount   float64 `gorm:"column:payout_amount"`
					FraudVerdict   string  `gorm:"column:fraud_verdict"`
					CreatedAt      string  `gorm:"column:created_at"`
					StartAt        string  `gorm:"column:start_at"`
					EndAt          string  `gorm:"column:end_at"`
				}
				var row detailRow
				err := workerDB.Raw(`
					SELECT c.id AS claim_id,
					       c.status,
					       d.type AS disruption_type,
					       COALESCE(z.name || ', ' || z.city, 'Unknown Zone') AS zone,
					       c.claim_amount AS income_loss,
					       COALESCE(p.amount, 0) AS payout_amount,
					       COALESCE(c.fraud_verdict, 'pending') AS fraud_verdict,
					       c.created_at::text AS created_at,
					       d.signal_timestamp::text AS start_at,
					       COALESCE(d.confirmed_at::text, d.signal_timestamp::text) AS end_at
					FROM claims c
					LEFT JOIN disruptions d ON d.id = c.disruption_id
					LEFT JOIN zones z ON z.id = d.zone_id
					LEFT JOIN payouts p ON p.claim_id = c.id
					WHERE c.worker_id = ? AND c.id = ?
				`, workerIDUint, claimNumID).Scan(&row).Error
				if err == nil && row.ClaimID != 0 {
					c.JSON(200, gin.H{
						"claim_id":          fmt.Sprintf("clm-%03d", row.ClaimID),
						"status":            row.Status,
						"zone":              row.Zone,
						"disruption_type":   row.DisruptionType,
						"disruption_window": gin.H{"start": row.StartAt, "end": row.EndAt},
						"income_loss":       int(row.IncomeLoss),
						"payout_amount":     int(row.PayoutAmount),
						"fraud_verdict":     row.FraudVerdict,
						"created_at":        row.CreatedAt,
					})
					return
				}
			}
		}
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	for _, claim := range store.data.Claims {
		if claim["claim_id"] == claimID {
			c.JSON(200, claim)
			return
		}
	}

	c.JSON(404, gin.H{"error": "claim_not_found"})
}
