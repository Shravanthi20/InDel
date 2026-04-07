package pollers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

// AQIPoller polls OpenAQ (free, no key needed) every 30 minutes per zone.
// Trigger 3: Severe Pollution — AQI > 300 (hazardous).
type AQIPoller struct {
	DB *gorm.DB
}

type openAQResponse struct {
	Results []struct {
		Measurements []struct {
			Parameter string  `json:"parameter"`
			Value     float64 `json:"value"`
		} `json:"measurements"`
	} `json:"results"`
}

func (p *AQIPoller) Start() {
	go func() {
		p.poll()
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			p.poll()
		}
	}()
}

func (p *AQIPoller) poll() {
	if p.DB == nil {
		return
	}

	var zones []models.Zone
	if err := p.DB.Find(&zones).Error; err != nil {
		log.Printf("[AQIPoller] DB error: %v", err)
		return
	}

	for _, zone := range zones {
		aqi, err := p.fetchAQI(zone)
		if err != nil {
			// Fallback: simulate AQI based on zone risk
			aqi = 80 + zone.RiskRating*140
			log.Printf("[AQIPoller] Using mock AQI=%.0f for zone %s", aqi, zone.Name)
		}

		// Trigger 3: Hazardous AQI
		if aqi > 300 {
			p.fireDisruption(zone, aqi)
		}
	}
}

func (p *AQIPoller) fetchAQI(zone models.Zone) (float64, error) {
	// OpenAQ v2 — free, no auth required
	city := zone.City
	if city == "" {
		city = zone.Name
	}
	url := fmt.Sprintf(
		"https://api.openaq.org/v2/latest?city=%s&parameter=pm25&limit=1&order_by=lastUpdated&sort=desc",
		city,
	)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("OpenAQ returned %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var result openAQResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	if len(result.Results) == 0 || len(result.Results[0].Measurements) == 0 {
		return 0, fmt.Errorf("no AQI data")
	}
	// Convert PM2.5 µg/m³ to AQI (simplified US EPA breakpoints)
	pm25 := result.Results[0].Measurements[0].Value
	return pm25ToAQI(pm25), nil
}

// pm25ToAQI converts PM2.5 µg/m³ to US EPA AQI scale
func pm25ToAQI(pm25 float64) float64 {
	switch {
	case pm25 <= 12.0:
		return linearAQI(pm25, 0, 50, 0, 12.0)
	case pm25 <= 35.4:
		return linearAQI(pm25, 51, 100, 12.1, 35.4)
	case pm25 <= 55.4:
		return linearAQI(pm25, 101, 150, 35.5, 55.4)
	case pm25 <= 150.4:
		return linearAQI(pm25, 151, 200, 55.5, 150.4)
	case pm25 <= 250.4:
		return linearAQI(pm25, 201, 300, 150.5, 250.4)
	default:
		return linearAQI(pm25, 301, 500, 250.5, 500.4)
	}
}

func linearAQI(cp, iLo, iHi, bpLo, bpHi float64) float64 {
	return ((iHi-iLo)/(bpHi-bpLo))*(cp-bpLo) + iLo
}

func (p *AQIPoller) fireDisruption(zone models.Zone, aqi float64) {
	now := time.Now().UTC()

	var existing models.Disruption
	err := p.DB.Where(
		"zone_id = ? AND type = ? AND created_at >= ?",
		zone.ID, "severe_pollution", now.Add(-2*time.Hour),
	).First(&existing).Error
	if err == nil {
		return
	}

	confidence := 0.78 + (aqi-300)/1000
	if confidence > 0.99 {
		confidence = 0.99
	}
	signalTime := now
	confirmedAt := now.Add(15 * time.Minute)

	disruption := models.Disruption{
		ZoneID:          zone.ID,
		Type:            "severe_pollution",
		Severity:        "high",
		Confidence:      confidence,
		Status:          "confirmed",
		SignalTimestamp: &signalTime,
		ConfirmedAt:     &confirmedAt,
		StartTime:       &signalTime,
	}

	if err := p.DB.Create(&disruption).Error; err != nil {
		log.Printf("[AQIPoller] Failed to create disruption: %v", err)
		return
	}

	log.Printf("[AQIPoller] ✅ Severe pollution in %s (AQI=%.0f)", zone.Name, aqi)
}
