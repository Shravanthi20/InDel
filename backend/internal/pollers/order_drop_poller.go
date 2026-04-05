package pollers

import (
	"log"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

// OrderDropPoller checks internal DB every 15 minutes for zone-level order drop.
// Trigger 4: Platform Order Drop — >40% decline vs 4-week zone average.
type OrderDropPoller struct {
	DB *gorm.DB
}

func (p *OrderDropPoller) Start() {
	go func() {
		p.poll()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			p.poll()
		}
	}()
}

func (p *OrderDropPoller) poll() {
	if p.DB == nil {
		return
	}

	var zones []models.Zone
	if err := p.DB.Find(&zones).Error; err != nil {
		log.Printf("[OrderDropPoller] DB error: %v", err)
		return
	}

	now := time.Now().UTC()
	windowStart := now.Add(-1 * time.Hour)  // last 1-hour window
	historicStart := now.AddDate(0, 0, -28)  // 4 weeks ago

	for _, zone := range zones {
		type orderRow struct {
			Recent   float64 `gorm:"column:recent"`
			Historic float64 `gorm:"column:historic"`
		}
		var row orderRow

		err := p.DB.Raw(`
			SELECT
				COALESCE(
					(SELECT COUNT(*) FROM orders
					 WHERE zone_id = ? AND created_at >= ?), 0
				) AS recent,
				COALESCE(
					(SELECT COUNT(*) / 4.0 FROM orders
					 WHERE zone_id = ? AND created_at >= ? AND created_at < ?), 1
				) AS historic
		`, zone.ID, windowStart,
			zone.ID, historicStart, windowStart,
		).Scan(&row).Error

		if err != nil {
			continue
		}

		if row.Historic < 1 {
			continue // not enough history
		}

		dropRatio := 1.0 - (row.Recent / row.Historic)

		// Trigger 4: >40% drop vs historical average
		if dropRatio >= 0.40 {
			p.fireDisruption(zone, dropRatio, now)
		}
	}
}

func (p *OrderDropPoller) fireDisruption(zone models.Zone, dropRatio float64, now time.Time) {
	var existing models.Disruption
	err := p.DB.Where(
		"zone_id = ? AND type = ? AND created_at >= ?",
		zone.ID, "demand_drop", now.Add(-10*time.Second),
	).First(&existing).Error
	if err == nil {
		return
	}

	confidence := 0.70 + dropRatio*0.25
	if confidence > 0.99 {
		confidence = 0.99
	}
	severity := "medium"
	if dropRatio >= 0.60 {
		severity = "high"
	}

	signalTime := now
	confirmedAt := now.Add(15 * time.Minute)

	disruption := models.Disruption{
		ZoneID:          zone.ID,
		Type:            "demand_drop",
		Severity:        severity,
		Confidence:      confidence,
		Status:          "confirmed",
		SignalTimestamp: &signalTime,
		ConfirmedAt:     &confirmedAt,
		StartTime:       &signalTime,
	}

	if err := p.DB.Create(&disruption).Error; err != nil {
		log.Printf("[OrderDropPoller] Failed to create disruption: %v", err)
		return
	}

	log.Printf("[OrderDropPoller] ✅ Order drop in %s (drop=%.0f%%)", zone.Name, dropRatio*100)
}
