package worker

import "github.com/gin-gonic/gin"

// DemoReset resets all in-memory demo state.
func DemoReset(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec("DELETE FROM notifications WHERE worker_id = ?", workerIDUint).Error
			_ = workerDB.Exec("DELETE FROM auth_tokens WHERE user_id = ?", workerIDUint).Error
			_ = workerDB.Exec("UPDATE orders SET status='assigned', accepted_at=NULL, picked_up_at=NULL, delivered_at=NULL, updated_at=CURRENT_TIMESTAMP WHERE worker_id = ?", workerIDUint).Error
		}
	}
	store.reset()
	c.JSON(200, gin.H{"message": "demo_reset", "time": nowISO()})
}

// DemoTriggerDisruption creates a disruption notification.
func DemoTriggerDisruption(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	disruptionType := bodyString(body, "disruption_type", "heavy_rain")
	zone := bodyString(body, "zone", "Tambaram, Chennai")
	msg := disruptionType + " detected in " + zone + ". You are protected."

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec("INSERT INTO notifications (worker_id, type, message) VALUES (?, ?, ?)", workerIDUint, "disruption_alert", msg).Error
		}
	}

	store.mu.Lock()
	store.data.Notifications = append([]map[string]any{{
		"id":         nextID("ntf", len(store.data.Notifications)),
		"type":       "disruption_alert",
		"title":      "Disruption detected",
		"body":       msg,
		"created_at": nowISO(),
		"read":       false,
	}}, store.data.Notifications...)
	store.mu.Unlock()

	c.JSON(200, gin.H{
		"message":         "disruption_triggered",
		"disruption_type": disruptionType,
		"zone":            zone,
		"time":            nowISO(),
	})
}

// DemoSimulateOrders appends assigned orders for demo.
func DemoSimulateOrders(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)
	count := bodyInt(body, "count", 3)
	if count <= 0 {
		count = 1
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			var zoneID uint = 1
			_ = workerDB.Raw("SELECT zone_id FROM worker_profiles WHERE worker_id = ? LIMIT 1", workerIDUint).Scan(&zoneID).Error
			for i := 0; i < count; i++ {
				_ = workerDB.Exec(
					"INSERT INTO orders (worker_id, zone_id, order_value, status, pickup_area, drop_area, distance_km, updated_at) VALUES (?, ?, ?, 'assigned', ?, ?, ?, CURRENT_TIMESTAMP)",
					workerIDUint, zoneID, 55+i*8, "Tambaram", "Camp Road", 2.5+float64(i)*0.4,
				).Error
			}
		}
	}

	store.mu.Lock()
	base := len(store.data.Orders)
	for i := 0; i < count; i++ {
		store.data.Orders = append(store.data.Orders, map[string]any{
			"order_id":    nextID("ord", base+i),
			"pickup_area": "Tambaram",
			"drop_area":   "Camp Road",
			"distance_km": 2.5 + float64(i)*0.4,
			"earning_inr": 55 + i*8,
			"status":      "assigned",
			"assigned_at": nowISO(),
		})
	}
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "orders_simulated", "count": count})
}
