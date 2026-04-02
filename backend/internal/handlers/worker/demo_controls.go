package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// ZonePair represents a from-to city pair for order generation
type ZonePair struct {
	ID         uint
	FromCity   string
	ToCity     string
	FromState  string
	ToState    string
	Distance   float64
	DistanceKm float64
	FromLat    float64
	FromLon    float64
	ToLat      float64
	ToLon      float64
}

type zoneBPair struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	State      string  `json:"state"`
	DistanceKm float64 `json:"distance_km"`
	FromLat    float64 `json:"from_lat"`
	FromLon    float64 `json:"from_lon"`
	ToLat      float64 `json:"to_lat"`
	ToLon      float64 `json:"to_lon"`
}

type zoneCPair struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	FromState  string  `json:"from_state"`
	ToState    string  `json:"to_state"`
	DistanceKm float64 `json:"distance_km"`
	FromLat    float64 `json:"from_lat"`
	FromLon    float64 `json:"from_lon"`
	ToLat      float64 `json:"to_lat"`
	ToLon      float64 `json:"to_lon"`
}

type zoneIDRow struct {
	ID uint `gorm:"column:id"`
}

func readFirstExistingFile(paths []string) ([]byte, string, error) {
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, p, nil
		}
	}
	return nil, "", fmt.Errorf("none of the candidate files exist: %v", paths)
}

// loadZonePairs loads pairs directly from zone_b.json and zone_c.json.
func loadZonePairs() ([]ZonePair, error) {
	var pairs []ZonePair

	zoneBBytes, zoneBPath, err := readFirstExistingFile([]string{
		"/root/zone_b.json",
		"/app/zone_b.json",
		"../zone_b.json",
		"zone_b.json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load zone_b.json: %w", err)
	}

	zoneCBytes, zoneCPath, err := readFirstExistingFile([]string{
		"/root/zone_c.json",
		"/app/zone_c.json",
		"../zone_c.json",
		"zone_c.json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load zone_c.json: %w", err)
	}

	var bPairs []zoneBPair
	if err := json.Unmarshal(zoneBBytes, &bPairs); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", zoneBPath, err)
	}

	for _, p := range bPairs {
		if p.From == "" || p.To == "" {
			continue
		}
		pairs = append(pairs, ZonePair{
			ID:         uint(len(pairs) + 1),
			FromCity:   p.From,
			ToCity:     p.To,
			FromState:  p.State,
			ToState:    p.State,
			Distance:   p.DistanceKm,
			DistanceKm: p.DistanceKm,
			FromLat:    p.FromLat,
			FromLon:    p.FromLon,
			ToLat:      p.ToLat,
			ToLon:      p.ToLon,
		})
	}

	var cPairs []zoneCPair
	if err := json.Unmarshal(zoneCBytes, &cPairs); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", zoneCPath, err)
	}

	for _, p := range cPairs {
		if p.From == "" || p.To == "" {
			continue
		}
		pairs = append(pairs, ZonePair{
			ID:         uint(len(pairs) + 1),
			FromCity:   p.From,
			ToCity:     p.To,
			FromState:  p.FromState,
			ToState:    p.ToState,
			Distance:   p.DistanceKm,
			DistanceKm: p.DistanceKm,
			FromLat:    p.FromLat,
			FromLon:    p.FromLon,
			ToLat:      p.ToLat,
			ToLon:      p.ToLon,
		})
	}

	if len(pairs) == 0 {
		return nil, fmt.Errorf("no usable zone pairs found in %s and %s", zoneBPath, zoneCPath)
	}

	log.Printf("loadZonePairs: loaded %d pairs from %s and %s", len(pairs), zoneBPath, zoneCPath)

	return pairs, nil
}

// calculateDeliveryFee calculates delivery fee based on distance and zone type
func calculateDeliveryFee(distanceKm float64, isInterState bool) float64 {
	if isInterState {
		return distanceKm * 2.0 // Inter-state: 2x multiplier
	}
	return distanceKm * 1.2 // Intra-state: 1.2x multiplier
}

func loadZoneIDs() ([]uint, error) {
	if !hasDB() {
		return nil, fmt.Errorf("no database connection available")
	}

	rows := make([]zoneIDRow, 0)
	if err := workerDB.Raw("SELECT id FROM zones ORDER BY id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	zoneIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.ID != 0 {
			zoneIDs = append(zoneIDs, row.ID)
		}
	}
	if len(zoneIDs) == 0 {
		return nil, fmt.Errorf("no zones available in database")
	}
	return zoneIDs, nil
}

func seedDemoOrdersForZones(workerIDUint uint, zoneIDs []uint, count int) {
	if len(zoneIDs) == 0 {
		log.Println("seedDemoOrdersForZones: no zone ids available, falling back to worker zone")
		seedDemoOrdersForWorker(workerIDUint, count)
		return
	}

	if count <= 0 {
		count = len(zoneIDs) * 2
	}

	pairs, err := loadZonePairs()
	if err != nil || len(pairs) == 0 {
		if err != nil {
			log.Printf("seedDemoOrdersForZones: failed to load zone pairs: %v\n", err)
		}
		seedDemoOrdersWithFallback(workerIDUint, zoneIDs[0], count)
		return
	}

	now := time.Now()
	for i := 0; i < count; i++ {
		pair := pairs[i%len(pairs)]
		zoneID := zoneIDs[i%len(zoneIDs)]
		isInterState := pair.FromState != pair.ToState
		deliveryFee := calculateDeliveryFee(pair.DistanceKm, isInterState)
		orderValue := 50.0 + pair.DistanceKm*0.5
		zoneRoute := []string{"A"}
		if isInterState {
			zoneRoute = []string{"C", "B", "A"}
		} else if pair.DistanceKm > 30 {
			zoneRoute = []string{"B", "A"}
		}

		err := workerDB.Exec(`
			INSERT INTO orders (
				worker_id, zone_id, order_value, status,
				pickup_area, drop_area, distance_km, from_city, to_city,
				from_state, to_state, from_lat, from_lon, to_lat, to_lon,
				tip_inr, delivery_fee_inr, zone_route_path,
				created_at, updated_at
			) VALUES (?, ?, ?, 'assigned', ?, ?, ?, ?, ?,
				  ?, ?, ?, ?, ?, ?,
				  ?, ?, ?, ?, ?)`,
			workerIDUint, zoneID, orderValue,
			pair.FromCity, pair.ToCity, pair.DistanceKm,
			pair.FromCity, pair.ToCity,
			pair.FromState, pair.ToState,
			pair.FromLat, pair.FromLon, pair.ToLat, pair.ToLon,
			0.0, deliveryFee, encodeZonePath(zoneRoute),
			now, now,
		).Error

		if err != nil {
			log.Printf("seedDemoOrdersForZones: failed to insert order %d for zone %d: %v\n", i+1, zoneID, err)
			continue
		}
		log.Printf("seedDemoOrdersForZones: order %d -> zone %d %s → %s | %.1f km | ₹%.2f\n", i+1, zoneID, pair.FromCity, pair.ToCity, pair.DistanceKm, deliveryFee)
	}
	log.Printf("seedDemoOrdersForZones: successfully seeded %d orders across %d zones for worker %d\n", count, len(zoneIDs), workerIDUint)
}

// seedDemoOrdersForWorker creates realistic demo orders using zone pairs
func seedDemoOrdersForWorker(workerIDUint uint, count int) {
	if count <= 0 {
		count = 3 // Default to 3 orders
	}

	if !hasDB() {
		log.Println("seedDemoOrdersForWorker: No database connection available")
		return
	}

	// Get worker's zone_id
	var zoneID uint
	err := workerDB.Raw("SELECT zone_id FROM worker_profiles WHERE worker_id = ? LIMIT 1", workerIDUint).Scan(&zoneID).Error
	if err != nil {
		log.Printf("seedDemoOrdersForWorker: Failed to get worker zone_id: %v\n", err)
		return
	}

	if zoneID == 0 {
		log.Printf("seedDemoOrdersForWorker: Worker %d has no zone_id assigned\n", workerIDUint)
		return
	}

	seedDemoOrdersForZones(workerIDUint, []uint{zoneID}, count)
}

// seedDemoOrdersWithFallback creates demo orders using hardcoded areas (fallback)
func seedDemoOrdersWithFallback(workerIDUint, zoneID uint, count int) {
	now := time.Now()
	pickupAreas := []string{"Tambaram", "Camp Road", "Perungudi", "T Nagar"}
	dropAreas := []string{"Camp Road", "Perungudi", "T Nagar", "Nungambakkam"}

	for i := 0; i < count; i++ {
		pickupIdx := i % len(pickupAreas)
		dropIdx := (i + 1) % len(dropAreas)

		err := workerDB.Exec(`
			INSERT INTO orders (
				worker_id, zone_id, order_value, status, 
				pickup_area, drop_area, distance_km, 
				tip_inr, delivery_fee_inr, zone_route_path,
				created_at, updated_at
			) VALUES (?, ?, ?, 'assigned', ?, ?, ?, ?, ?, ?, ?, ?)`,
			workerIDUint, zoneID, 55+float64(i*8),
			pickupAreas[pickupIdx], dropAreas[dropIdx], 2.5+float64(i)*0.4,
			10.0, 40.0, `["A"]`,
			now, now,
		).Error

		if err != nil {
			log.Printf("seedDemoOrdersWithFallback: Failed to insert order %d: %v\n", i+1, err)
		}
	}
}

// DemoReset resets all in-memory demo state and reseeds orders.
func DemoReset(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	var resetLog []string
	resetLog = append(resetLog, fmt.Sprintf("DemoReset initiated for worker: %s", workerID))

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			// Clean up notifications
			result1 := workerDB.Exec("DELETE FROM notifications WHERE worker_id = ?", workerIDUint)
			if result1.Error != nil {
				log.Printf("DemoReset: Error deleting notifications: %v\n", result1.Error)
			} else {
				resetLog = append(resetLog, fmt.Sprintf("Deleted %d notifications", result1.RowsAffected))
			}

			// Clean up auth tokens
			result2 := workerDB.Exec("DELETE FROM auth_tokens WHERE user_id = ?", workerIDUint)
			if result2.Error != nil {
				log.Printf("DemoReset: Error deleting auth_tokens: %v\n", result2.Error)
			} else {
				resetLog = append(resetLog, fmt.Sprintf("Deleted %d auth_tokens", result2.RowsAffected))
			}

			// Reset existing orders
			result3 := workerDB.Exec("UPDATE orders SET status='assigned', accepted_at=NULL, picked_up_at=NULL, delivered_at=NULL, updated_at=CURRENT_TIMESTAMP WHERE worker_id = ?", workerIDUint)
			if result3.Error != nil {
				log.Printf("DemoReset: Error resetting orders: %v\n", result3.Error)
			} else {
				resetLog = append(resetLog, fmt.Sprintf("Reset %d existing orders to assigned", result3.RowsAffected))
			}

			zoneIDs, zoneErr := loadZoneIDs()
			if zoneErr != nil {
				log.Printf("DemoReset: Failed to load zones for seeding: %v\n", zoneErr)
				resetLog = append(resetLog, "Zone load failed, falling back to current worker zone")
				seedDemoOrdersForWorker(workerIDUint, 3)
			} else {
				log.Printf("DemoReset: About to seed demo orders across %d zones for worker %d\n", len(zoneIDs), workerIDUint)
				seedDemoOrdersForZones(workerIDUint, zoneIDs, len(zoneIDs)*2)
				resetLog = append(resetLog, fmt.Sprintf("Seeded %d new demo orders across %d zones", len(zoneIDs)*2, len(zoneIDs)))
			}
		}
	}

	// Reset in-memory store
	store.reset()
	resetLog = append(resetLog, "In-memory store reset")

	log.Println("DemoReset: " + fmt.Sprint(resetLog))
	c.JSON(200, gin.H{
		"message": "demo_reset",
		"time":    nowISO(),
		"details": resetLog,
	})
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
			if zoneIDs, zoneErr := loadZoneIDs(); zoneErr == nil {
				seedDemoOrdersForZones(workerIDUint, zoneIDs, count)
			} else {
				seedDemoOrdersForWorker(workerIDUint, count)
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

// DemoSettleEarnings settles demo earnings and triggers premium reminder.
func DemoSettleEarnings(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec(
				`UPDATE weekly_earnings_summary
				 SET claim_eligible = TRUE
				 WHERE worker_id = ?
				   AND week_start = date_trunc('week', CURRENT_DATE)::date
				   AND week_end = (date_trunc('week', CURRENT_DATE)::date + INTERVAL '6 day')::date`,
				workerIDUint,
			).Error
			_ = workerDB.Exec(
				"INSERT INTO notifications (worker_id, type, message) VALUES (?, 'premium_due', 'Weekly earnings settled. Pay premium to keep coverage active.')",
				workerIDUint,
			).Error
		}
	}

	store.mu.Lock()
	store.data.Notifications = append([]map[string]any{{
		"id":         nextID("ntf", len(store.data.Notifications)),
		"type":       "premium_due",
		"title":      "Weekly settlement complete",
		"body":       "Weekly earnings settled. Pay premium to keep coverage active.",
		"created_at": nowISO(),
		"read":       false,
	}}, store.data.Notifications...)
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "earnings_settled", "time": nowISO()})
}

// DemoResetZone resets disruption and claim state for demo replay.
func DemoResetZone(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec("DELETE FROM payouts WHERE worker_id = ?", workerIDUint).Error
			_ = workerDB.Exec("DELETE FROM claims WHERE worker_id = ?", workerIDUint).Error
			_ = workerDB.Exec("DELETE FROM notifications WHERE worker_id = ? AND type IN ('disruption_alert', 'payout_credited')", workerIDUint).Error
		}
	}

	store.mu.Lock()
	store.data.Claims = []map[string]any{}
	store.data.Payouts = []map[string]any{}
	store.data.Notifications = []map[string]any{}
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "zone_reset", "time": nowISO()})
}
