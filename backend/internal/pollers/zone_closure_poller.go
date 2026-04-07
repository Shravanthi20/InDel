package pollers

import (
	"log"
	"math/rand"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

// ZoneClosurePoller checks for curfews, strikes, and official zone restrictions.
// Trigger 5: Zone Closure — polls mock gov/traffic alert API every 60 minutes.
// In production, integrate with: traffic.gov.in, state police APIs, or city alert feeds.
type ZoneClosurePoller struct {
	DB *gorm.DB
}

type zoneAlert struct {
	ZoneName string
	City     string
	Type     string // "curfew" | "strike" | "zone_closure"
	Severity string
	Active   bool
}

func (p *ZoneClosurePoller) Start() {
	go func() {
		p.poll()
		ticker := time.NewTicker(60 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			p.poll()
		}
	}()
}

func (p *ZoneClosurePoller) poll() {
	if p.DB == nil {
		return
	}

	var zones []models.Zone
	if err := p.DB.Find(&zones).Error; err != nil {
		log.Printf("[ZoneClosurePoller] DB error: %v", err)
		return
	}

	// Fetch from mock/gov API
	alerts := p.fetchAlerts(zones)

	now := time.Now().UTC()
	for _, alert := range alerts {
		if !alert.Active {
			continue
		}

		// Find matching zone in DB
		var zone models.Zone
		err := p.DB.Where("name = ? AND city = ?", alert.ZoneName, alert.City).First(&zone).Error
		if err != nil {
			continue
		}

		p.fireDisruption(zone, alert, now)
	}
}

// fetchAlerts simulates a government/traffic alert API.
// In production: replace with actual API calls to traffic.gov.in or state police endpoints.
func (p *ZoneClosurePoller) fetchAlerts(zones []models.Zone) []zoneAlert {
	// Probabilistic simulation: 2% chance of a curfew/strike per zone per hour
	// This provides realistic rare-event behaviour for the demo
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	alerts := make([]zoneAlert, 0)

	alertTypes := []string{"curfew", "strike", "zone_closure"}
	for _, zone := range zones {
		if rng.Float64() < 0.02 { // 2% chance
			alerts = append(alerts, zoneAlert{
				ZoneName: zone.Name,
				City:     zone.City,
				Type:     alertTypes[rng.Intn(len(alertTypes))],
				Severity: "high",
				Active:   true,
			})
		}
	}
	return alerts
}

func (p *ZoneClosurePoller) fireDisruption(zone models.Zone, alert zoneAlert, now time.Time) {
	var existing models.Disruption
	err := p.DB.Where(
		"zone_id = ? AND type = ? AND created_at >= ?",
		zone.ID, alert.Type, now.Add(-4*time.Hour),
	).First(&existing).Error
	if err == nil {
		return
	}

	signalTime := now
	confirmedAt := now.Add(5 * time.Minute) // gov alerts confirm faster

	disruption := models.Disruption{
		ZoneID:          zone.ID,
		Type:            alert.Type,
		Severity:        alert.Severity,
		Confidence:      0.95, // gov alerts = high confidence
		Status:          "confirmed",
		SignalTimestamp: &signalTime,
		ConfirmedAt:     &confirmedAt,
		StartTime:       &signalTime,
	}

	if err := p.DB.Create(&disruption).Error; err != nil {
		log.Printf("[ZoneClosurePoller] Failed to create disruption: %v", err)
		return
	}

	log.Printf("[ZoneClosurePoller] ✅ %s in %s, %s", alert.Type, zone.Name, zone.City)
}
