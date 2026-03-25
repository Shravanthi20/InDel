package platform

import "github.com/gin-gonic/gin"

// GetWorkers returns worker list
func GetWorkers(c *gin.Context) {
	if hasDB() {
		type row struct {
			WorkerID uint   `gorm:"column:worker_id"`
			Name     string `gorm:"column:name"`
			Phone    string `gorm:"column:phone"`
			Zone     string `gorm:"column:zone"`
		}

		rows := make([]row, 0)
		_ = platformDB.Raw(`
			SELECT u.id AS worker_id,
			       wp.name,
			       u.phone,
			       COALESCE(z.name || ', ' || z.city, 'Unknown Zone') AS zone
			FROM users u
			JOIN worker_profiles wp ON wp.worker_id = u.id
			LEFT JOIN zones z ON z.id = wp.zone_id
			WHERE u.role = 'worker'
			ORDER BY u.id
		`).Scan(&rows).Error

		workers := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			workers = append(workers, gin.H{
				"worker_id": r.WorkerID,
				"name":      r.Name,
				"phone":     r.Phone,
				"zone":      r.Zone,
			})
		}

		c.JSON(200, gin.H{"workers": workers})
		return
	}

	c.JSON(200, gin.H{"workers": []gin.H{{
		"worker_id": 1,
		"name":      "Gayathri Worker",
		"phone":     "+919999999999",
		"zone":      "Tambaram, Chennai",
	}}})
}
