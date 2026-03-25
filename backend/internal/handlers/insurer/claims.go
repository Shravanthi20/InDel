package insurer

import "github.com/gin-gonic/gin"

// GetClaims returns insurer claims pipeline view.
func GetClaims(c *gin.Context) {
	if hasDB() {
		type row struct {
			ClaimID      uint    `gorm:"column:claim_id"`
			Status       string  `gorm:"column:status"`
			City         string  `gorm:"column:city"`
			Zone         string  `gorm:"column:zone"`
			ClaimAmount  float64 `gorm:"column:claim_amount"`
			FraudVerdict string  `gorm:"column:fraud_verdict"`
			CreatedAt    string  `gorm:"column:created_at"`
		}

		rows := make([]row, 0)
		_ = insurerDB.Raw(`
			SELECT c.id AS claim_id,
			       c.status,
			       z.city,
			       z.name AS zone,
			       c.claim_amount,
			       COALESCE(c.fraud_verdict, 'pending') AS fraud_verdict,
			       c.created_at::text AS created_at
			FROM claims c
			JOIN disruptions d ON d.id = c.disruption_id
			JOIN zones z ON z.id = d.zone_id
			ORDER BY c.created_at DESC
			LIMIT 100
		`).Scan(&rows).Error

		claims := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			claims = append(claims, gin.H{
				"claim_id":      r.ClaimID,
				"status":        r.Status,
				"city":          r.City,
				"zone":          r.Zone,
				"claim_amount":  int(r.ClaimAmount),
				"fraud_verdict": r.FraudVerdict,
				"created_at":    r.CreatedAt,
			})
		}

		c.JSON(200, gin.H{"claims": claims})
		return
	}

	c.JSON(200, gin.H{"claims": []gin.H{}})
}

// GetFraudQueue returns claims flagged for manual review.
func GetFraudQueue(c *gin.Context) {
	if hasDB() {
		type row struct {
			ClaimID      uint   `gorm:"column:claim_id"`
			FinalVerdict string `gorm:"column:final_verdict"`
			Violations   string `gorm:"column:violations"`
			CreatedAt    string `gorm:"column:created_at"`
		}

		rows := make([]row, 0)
		_ = insurerDB.Raw(`
			SELECT c.id AS claim_id,
			       COALESCE(cfs.final_verdict, 'pending') AS final_verdict,
			       COALESCE(cfs.rule_violations::text, '[]') AS violations,
			       c.created_at::text AS created_at
			FROM claims c
			LEFT JOIN claim_fraud_scores cfs ON cfs.claim_id = c.id
			WHERE COALESCE(cfs.final_verdict, 'pending') IN ('flagged', 'manual_review', 'pending')
			ORDER BY c.created_at DESC
			LIMIT 100
		`).Scan(&rows).Error

		queue := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			queue = append(queue, gin.H{
				"claim_id":      r.ClaimID,
				"final_verdict": r.FinalVerdict,
				"violations":    r.Violations,
				"created_at":    r.CreatedAt,
			})
		}

		c.JSON(200, gin.H{"fraud_queue": queue})
		return
	}

	c.JSON(200, gin.H{"fraud_queue": []gin.H{{
		"claim_id":      1,
		"final_verdict": "pending",
		"violations":    "[]",
		"created_at":    "mock",
	}}})
}
