package platform

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type orderAssignedRequest struct {
	OrderID    string  `json:"order_id"`
	WorkerID   uint    `json:"worker_id"`
	ZoneID     uint    `json:"zone_id"`
	OrderValue float64 `json:"order_value"`
	PickupArea string  `json:"pickup_area"`
	DropArea   string  `json:"drop_area"`
	DistanceKM float64 `json:"distance_km"`
}

type orderCompletedRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

// GetZones returns active zones with risk rating.
func GetZones(c *gin.Context) {
	if hasDB() {
		type row struct {
			ZoneID      uint    `gorm:"column:zone_id"`
			Name        string  `gorm:"column:name"`
			City        string  `gorm:"column:city"`
			State       string  `gorm:"column:state"`
			RiskRating  float64 `gorm:"column:risk_rating"`
			WorkerCount int64   `gorm:"column:worker_count"`
		}

		rows := make([]row, 0)
		_ = platformDB.Raw(`
			SELECT z.id AS zone_id,
			       z.name,
			       z.city,
			       z.state,
			       z.risk_rating,
			       COUNT(wp.worker_id) AS worker_count
			FROM zones z
			LEFT JOIN worker_profiles wp ON wp.zone_id = z.id
			GROUP BY z.id, z.name, z.city, z.state, z.risk_rating
			ORDER BY z.city, z.name
		`).Scan(&rows).Error

		zones := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			zones = append(zones, gin.H{
				"zone_id":        r.ZoneID,
				"name":           r.Name,
				"city":           r.City,
				"state":          r.State,
				"risk_rating":    r.RiskRating,
				"active_workers": r.WorkerCount,
			})
		}

		c.JSON(200, gin.H{"zones": zones})
		return
	}

	c.JSON(200, gin.H{"zones": []gin.H{{
		"zone_id":        1,
		"name":           "Tambaram",
		"city":           "Chennai",
		"state":          "Tamil Nadu",
		"risk_rating":    0.62,
		"active_workers": 1,
	}}})
}

// OrderAssignedWebhook creates an assigned order.
func OrderAssignedWebhook(c *gin.Context) {
	var req orderAssignedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if req.WorkerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "worker_id_required"})
		return
	}

	if req.ZoneID == 0 {
		req.ZoneID = 1
	}
	if req.OrderValue <= 0 {
		req.OrderValue = 60
	}

	if hasDB() {
		var createdOrderID uint
		err := platformDB.Raw(
			`INSERT INTO orders (worker_id, zone_id, order_value, status, pickup_area, drop_area, distance_km, updated_at)
			 VALUES (?, ?, ?, 'assigned', ?, ?, ?, CURRENT_TIMESTAMP)
			 RETURNING id`,
			req.WorkerID, req.ZoneID, req.OrderValue, req.PickupArea, req.DropArea, req.DistanceKM,
		).Scan(&createdOrderID).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order_create_failed"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":   "order_assigned",
			"order_id":  createdOrderID,
			"worker_id": req.WorkerID,
			"status":    "assigned",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "order_assigned",
		"order_id":  req.OrderID,
		"worker_id": req.WorkerID,
		"status":    "assigned",
	})
}

// OrderCompletedWebhook marks order delivered and updates worker earnings.
func OrderCompletedWebhook(c *gin.Context) {
	var req orderCompletedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if strings.TrimSpace(req.OrderID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id_required"})
		return
	}

	orderRaw := strings.TrimPrefix(strings.TrimSpace(req.OrderID), "ord-")
	orderNum, err := strconv.ParseUint(orderRaw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_order_id"})
		return
	}

	if hasDB() {
		type orderRow struct {
			WorkerID   uint    `gorm:"column:worker_id"`
			OrderValue float64 `gorm:"column:order_value"`
		}
		var row orderRow
		_ = platformDB.Raw("SELECT worker_id, order_value FROM orders WHERE id = ?", orderNum).Scan(&row).Error
		if row.WorkerID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "order_not_found"})
			return
		}

		amount := req.Amount
		if amount <= 0 {
			amount = row.OrderValue
		}

		_ = platformDB.Exec(
			"UPDATE orders SET status='delivered', delivered_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id = ?",
			orderNum,
		).Error

		_ = platformDB.Exec(
			`INSERT INTO earnings_records (worker_id, date, hours_worked, amount_earned)
			 VALUES (?, CURRENT_DATE, 1, ?)
			 ON CONFLICT (worker_id, date)
			 DO UPDATE SET amount_earned = earnings_records.amount_earned + EXCLUDED.amount_earned`,
			row.WorkerID, amount,
		).Error

		result := platformDB.Exec(
			`UPDATE weekly_earnings_summary
			 SET total_earnings = total_earnings + ?
			 WHERE worker_id = ?
			   AND week_start = date_trunc('week', CURRENT_DATE)::date
			   AND week_end = (date_trunc('week', CURRENT_DATE)::date + INTERVAL '6 day')::date`,
			amount, row.WorkerID,
		)

		if result.RowsAffected == 0 {
			_ = platformDB.Exec(
				`INSERT INTO weekly_earnings_summary (worker_id, week_start, week_end, total_earnings, claim_eligible)
				 VALUES (?, date_trunc('week', CURRENT_DATE)::date, (date_trunc('week', CURRENT_DATE)::date + INTERVAL '6 day')::date, ?, FALSE)`,
				row.WorkerID, amount,
			).Error
		}

		c.JSON(200, gin.H{
			"message":   "order_completed",
			"order_id":  req.OrderID,
			"worker_id": row.WorkerID,
			"amount":    int(amount),
		})
		return
	}

	c.JSON(200, gin.H{
		"message":  "order_completed",
		"order_id": req.OrderID,
		"amount":   int(req.Amount),
	})
}
