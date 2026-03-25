package insurer

import "github.com/gin-gonic/gin"

// GetLossRatio returns loss ratio by zone/city
func GetLossRatio(c *gin.Context) {
	if hasDB() {
		type row struct {
			City     string  `gorm:"column:city"`
			Zone     string  `gorm:"column:zone"`
			Premiums float64 `gorm:"column:premiums"`
			Claims   float64 `gorm:"column:claims"`
		}

		rows := make([]row, 0)
		_ = insurerDB.Raw(`
			SELECT z.city,
			       z.name AS zone,
			       COALESCE(p.premiums, 0) AS premiums,
			       COALESCE(cl.claims, 0) AS claims
			FROM zones z
			LEFT JOIN (
				SELECT wp.zone_id, SUM(pp.amount) AS premiums
				FROM premium_payments pp
				JOIN worker_profiles wp ON wp.worker_id = pp.worker_id
				WHERE pp.status IN ('completed', 'captured', 'processed')
				GROUP BY wp.zone_id
			) p ON p.zone_id = z.id
			LEFT JOIN (
				SELECT d.zone_id, SUM(c.claim_amount) AS claims
				FROM claims c
				JOIN disruptions d ON d.id = c.disruption_id
				GROUP BY d.zone_id
			) cl ON cl.zone_id = z.id
			ORDER BY z.city, z.name
		`).Scan(&rows).Error

		zones := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			lossRatio := 0.0
			if r.Premiums > 0 {
				lossRatio = r.Claims / r.Premiums
			}
			zones = append(zones, gin.H{
				"city":       r.City,
				"zone":       r.Zone,
				"premiums":   int(r.Premiums),
				"claims":     int(r.Claims),
				"loss_ratio": lossRatio,
			})
		}

		c.JSON(200, gin.H{"zones": zones})
		return
	}

	c.JSON(200, gin.H{"zones": []gin.H{{
		"city":       "Chennai",
		"zone":       "Tambaram",
		"premiums":   2200,
		"claims":     980,
		"loss_ratio": 0.445,
	}}})
}
