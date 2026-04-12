package platform

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type orderAssignedRequest struct {
	OrderID         string  `json:"order_id"`
	WorkerID        uint    `json:"worker_id"`
	ZoneID          uint    `json:"zone_id"`
	OrderValue      float64 `json:"order_value"`
	PickupArea      string  `json:"pickup_area"`
	DropArea        string  `json:"drop_area"`
	DistanceKM      float64 `json:"distance_km"`
	VehicleType     string  `json:"vehicle_type"`
	VehicleCapacity int     `json:"vehicle_capacity"`
	AllowedZones    string  `json:"allowed_zones"`
}

type orderCompletedRequest struct {
	OrderID    string      `json:"order_id"`
	Amount     float64     `json:"amount"`
	EarningInr float64     `json:"earning_inr"`
	ZoneID     interface{} `json:"zone_id"`
}

type orderCancelledRequest struct {
	OrderID string      `json:"order_id"`
	ZoneID  interface{} `json:"zone_id"`
}

// GetZones returns active zones with risk rating.
func GetZones(c *gin.Context) {
	levelFilter := strings.ToUpper(strings.TrimSpace(c.Query("level")))
	if levelFilter == "ALL" {
		levelFilter = ""
	}
	if levelFilter != "" && levelFilter != "A" && levelFilter != "B" && levelFilter != "C" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_level", "message": "level must be one of A, B, C, or ALL"})
		return
	}

	// Static mapping for demo; replace with DB query if areas are in DB
	var zoneAreas = map[string][]string{
		"Tambaram":     {"Tambaram", "Velachery", "Pallikaranai"},
		"Selaiyur":     {"Selaiyur", "Chromepet", "Pallikaranai"},
		"Pallikaranai": {"Pallikaranai", "Velachery", "Tambaram"},
		"Chromepet":    {"Chromepet", "Selaiyur", "Tambaram"},
		"Velachery":    {"Velachery", "Tambaram", "Pallikaranai"},
		"Medavakkam":   {"Medavakkam", "Velachery", "Pallikaranai"},
		"Zone-A":       {"Whitefield", "Koramangala", "Indiranagar", "Bangalore City"},
		"Zone-B":       {"Koramangala", "Indiranagar", "Whitefield", "JP Nagar"},
		"Zone-C":       {"Bandra", "Andheri", "Dadar", "Marine Drive"},
		"Zone-D":       {"Connaught Place", "Nehru Place", "Noida", "Gurgaon"},
	}

	if hasDB() {
		type row struct {
			ZoneID      uint    `gorm:"column:zone_id"`
			Level       string  `gorm:"column:level"`
			Name        string  `gorm:"column:name"`
			City        string  `gorm:"column:city"`
			State       string  `gorm:"column:state"`
			RiskRating  float64 `gorm:"column:risk_rating"`
			WorkerCount int64   `gorm:"column:worker_count"`
		}
		rows := make([]row, 0)
		err := platformDB.Raw(`
			SELECT z.id AS zone_id,
				   COALESCE(z.level, '') AS level,
				   z.name,
				   z.city,
				   z.state,
				   z.risk_rating,
				   COUNT(wp.worker_id) AS worker_count
			FROM zones z
			LEFT JOIN worker_profiles wp ON wp.zone_id = z.id
			WHERE (? = '' OR UPPER(COALESCE(z.level, '')) = ?)
			GROUP BY z.id, z.level, z.name, z.city, z.state, z.risk_rating
			ORDER BY z.city, z.name
		`, levelFilter, levelFilter).Scan(&rows).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "zones_query_failed", "message": err.Error()})
			return
		}
		zones := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			areas := zoneAreas[r.Name]
			if len(areas) == 0 {
				areas = []string{r.Name, r.City}
			}
			zones = append(zones, gin.H{
				"zone_id":        r.ZoneID,
				"level":          r.Level,
				"name":           r.Name,
				"city":           r.City,
				"state":          r.State,
				"risk_rating":    r.RiskRating,
				"active_workers": r.WorkerCount,
				"areas":          areas,
			})
		}
		c.JSON(200, gin.H{"zones": zones})
		return
	}

	fallbackLevel := "B"
	if levelFilter != "" && levelFilter != fallbackLevel {
		c.JSON(200, gin.H{"zones": []gin.H{}})
		return
	}
	c.JSON(200, gin.H{"zones": []gin.H{{
		"zone_id":        1,
		"level":          fallbackLevel,
		"name":           "Tambaram",
		"city":           "Chennai",
		"state":          "Tamil Nadu",
		"risk_rating":    0.62,
		"active_workers": 1,
		"areas":          []string{"Tambaram", "Velachery", "Pallikaranai"},
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
			`INSERT INTO orders (worker_id, zone_id, order_value, status, pickup_area, drop_area, distance_km, vehicle_type, vehicle_capacity, allowed_zones, updated_at)
			 VALUES (?, ?, ?, 'assigned', ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			 RETURNING id`,
			req.WorkerID, req.ZoneID, req.OrderValue, req.PickupArea, req.DropArea, req.DistanceKM, req.VehicleType, req.VehicleCapacity, req.AllowedZones,
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
	fmt.Printf("\n📢 [WEBHOOK RECEIVED] OrderCompletedWebhook hit at %s\n", time.Now().Format(time.Kitchen))
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
	orderRaw = strings.TrimPrefix(orderRaw, "ord_")

	// HACKATHON DEMO: Robust 'fake' check to decouple from Part 4 DB
	isFakeOrder := strings.Contains(strings.ToLower(req.OrderID), "fake")

	if hasDB() && !isFakeOrder {
		orderNum, err := strconv.ParseUint(orderRaw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_order_id", "details": "not_numeric"})
			return
		}

		type orderRow struct {
			WorkerID   uint    `gorm:"column:worker_id"`
			OrderValue float64 `gorm:"column:order_value"`
		}
		var row orderRow
		// Use SCAN or take result to avoid crash on not found
		_ = platformDB.Raw("SELECT worker_id, order_value FROM orders WHERE id = ?", orderNum).Scan(&row).Error
		if row.WorkerID == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "order_not_found"})
			return
		}

		amount := req.Amount
		if amount <= 0 {
			amount = req.EarningInr
		}
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
			"data": gin.H{
				"message":   "order_completed",
				"order_id":  req.OrderID,
				"worker_id": row.WorkerID,
				"amount":    int(amount),
			},
			"meta": gin.H{"timestamp": "2026-03-30T10:00:00Z"},
		})

		// Order tracking per zone
		zoneParsed := uint(1) // Default to 1 if missing for hackathon scope
		switch v := req.ZoneID.(type) {
		case float64:
			zoneParsed = uint(v)
		case string:
			// parse string or map it
			zoneParsed = 1
		}
		CheckAndTrackOrder(req.OrderID, zoneParsed, true)

		return
	}

	CheckAndTrackOrder(req.OrderID, 1, true)

	c.JSON(200, gin.H{
		"data": gin.H{
			"message":  "order_completed",
			"order_id": req.OrderID,
			"amount":   int(req.Amount),
		},
		"meta": gin.H{"timestamp": "2026-03-30T10:00:00Z"},
	})
}

// OrderCancelledWebhook logs a dropped order for metric calculation
func OrderCancelledWebhook(c *gin.Context) {
	var req orderCancelledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if strings.TrimSpace(req.OrderID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id_required"})
		return
	}

	zoneParsed := uint(1)
	switch v := req.ZoneID.(type) {
	case float64:
		zoneParsed = uint(v)
	case string:
		zoneParsed = 1
	}

	CheckAndTrackOrder(req.OrderID, zoneParsed, false)

	c.JSON(200, gin.H{
		"data": gin.H{
			"message":  "order_cancelled",
			"order_id": req.OrderID,
		},
		"meta": gin.H{"timestamp": "2026-03-30T10:00:00Z"},
	})
}
