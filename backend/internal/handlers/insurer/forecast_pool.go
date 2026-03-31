package insurer

import "github.com/gin-gonic/gin"

// GetForecast returns 7-day disruption probability by zone.
func (h *InsurerHandler) GetForecast(c *gin.Context) {
	if h.Service.DB != nil {
		type row struct {
			City        string  `gorm:"column:city"`
			Zone        string  `gorm:"column:zone"`
			Date        string  `gorm:"column:forecast_date"`
			Probability float64 `gorm:"column:probability"`
		}

		rows := make([]row, 0)
		_ = h.Service.DB.Raw(`
			SELECT z.city,
			       z.name AS zone,
			       f.forecast_date::text AS forecast_date,
			       COALESCE(f.predicted_disruption_probability, 0) AS probability
			FROM forecast_model_outputs f
			JOIN zones z ON z.id = f.zone_id
			WHERE f.forecast_date >= CURRENT_DATE
			  AND f.forecast_date < CURRENT_DATE + INTERVAL '7 day'
			ORDER BY f.forecast_date ASC, z.city, z.name
		`).Scan(&rows).Error

		forecast := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			forecast = append(forecast, gin.H{
				"city":        r.City,
				"zone":        r.Zone,
				"date":        r.Date,
				"probability": r.Probability,
			})
		}

		c.JSON(200, gin.H{"forecast": forecast})
		return
	}

	c.JSON(200, gin.H{"forecast": []gin.H{{
		"city":        "Chennai",
		"zone":        "Tambaram",
		"date":        "2026-03-25",
		"probability": 0.37,
	}}})
}

// GetPoolHealth returns premiums vs payouts for insurer reserve tracking.
func (h *InsurerHandler) GetPoolHealth(c *gin.Context) {
	if h.Service.DB != nil {
		var weeklyPremiums float64
		var weeklyPayouts float64
		var pendingPayouts int64

		_ = h.Service.DB.Raw(`
			SELECT COALESCE(SUM(amount), 0)
			FROM premium_payments
			WHERE payment_date >= date_trunc('week', CURRENT_DATE)
			  AND payment_date < date_trunc('week', CURRENT_DATE) + INTERVAL '7 day'
			  AND status IN ('completed', 'captured', 'processed')
		`).Scan(&weeklyPremiums).Error

		_ = h.Service.DB.Raw(`
			SELECT COALESCE(SUM(amount), 0)
			FROM payouts
			WHERE created_at >= date_trunc('week', CURRENT_DATE)
			  AND created_at < date_trunc('week', CURRENT_DATE) + INTERVAL '7 day'
			  AND status IN ('processed', 'credited', 'completed')
		`).Scan(&weeklyPayouts).Error

		_ = h.Service.DB.Raw("SELECT COUNT(*) FROM payouts WHERE status IN ('queued', 'pending')").Scan(&pendingPayouts).Error

		c.JSON(200, gin.H{
			"week_premiums":   int(weeklyPremiums),
			"week_payouts":    int(weeklyPayouts),
			"net_pool":        int(weeklyPremiums - weeklyPayouts),
			"pending_payouts": pendingPayouts,
		})
		return
	}

	c.JSON(200, gin.H{
		"week_premiums":   2200,
		"week_payouts":    960,
		"net_pool":        1240,
		"pending_payouts": 0,
	})
}
