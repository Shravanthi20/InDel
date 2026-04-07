package pollers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

// WeatherPoller polls OpenWeatherMap every 10 minutes per zone and fires
// disruption triggers for Heavy Rain (>25mm/h) and Extreme Heat (>42°C 11am-6pm).
type WeatherPoller struct {
	DB     *gorm.DB
	keyIdx int
}

type owmResponse struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Rain struct {
		OneH float64 `json:"1h"`
	} `json:"rain"`
	Name string `json:"name"`
}

func (p *WeatherPoller) Start() {
	apiKeyStr := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKeyStr == "" {
		log.Println("[WeatherPoller] No API key — using mock weather simulation")
	}
	apiKeys := strings.Split(apiKeyStr, ",")
	for i := range apiKeys {
		apiKeys[i] = strings.TrimSpace(apiKeys[i])
	}

	go func() {
		// Poll immediately on start, then every 10 minutes
		p.poll(apiKeys)
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			p.poll(apiKeys)
		}
	}()
}

func (p *WeatherPoller) poll(apiKeys []string) {
	if p.DB == nil {
		return
	}

	var zones []models.Zone
	if err := p.DB.Find(&zones).Error; err != nil {
		log.Printf("[WeatherPoller] DB error fetching zones: %v", err)
		return
	}

	for _, zone := range zones {
		// Rotate API key for each zone fetch
		var currentKey string
		if len(apiKeys) > 0 && apiKeys[0] != "" {
			currentKey = apiKeys[p.keyIdx%len(apiKeys)]
			p.keyIdx++
		}

		rain, temp, err := p.fetchWeather(currentKey, zone)
		if err != nil {
			log.Printf("[WeatherPoller] Error fetching weather for zone %s using key %s...: %v", zone.Name, maskKey(currentKey), err)
			rain, temp = mockWeather(zone)
		}
		p.evaluate(zone, rain, temp)
	}
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func (p *WeatherPoller) fetchWeather(apiKey string, zone models.Zone) (rainMM, tempC float64, err error) {
	if apiKey == "" {
		return 0, 0, fmt.Errorf("no api key")
	}
	query := fmt.Sprintf("%s,%s", zone.Name, zone.State)
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		query, apiKey,
	)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return 0, 0, fmt.Errorf("OWM returned %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var owm owmResponse
	if err := json.Unmarshal(body, &owm); err != nil {
		return 0, 0, err
	}
	return owm.Rain.OneH, owm.Main.Temp, nil
}

func mockWeather(zone models.Zone) (rainMM, tempC float64) {
	// Simulate realistic weather based on zone risk rating
	risk := zone.RiskRating
	rainMM = risk * 45  // High-risk zones get more rain
	tempC = 28 + risk*15
	return
}

func (p *WeatherPoller) evaluate(zone models.Zone, rainMM, tempC float64) {
	now := time.Now().UTC()

	// Trigger 1: Heavy Rain — >25mm/h
	if rainMM > 25 {
		p.fireDisruption(zone, "heavy_rain", "high", rainMM, now)
	}

	// Trigger 2: Extreme Heat — >42°C during working hours (11am–6pm IST = 5:30–12:30 UTC)
	hour := now.Add(5*time.Hour + 30*time.Minute).Hour()
	if tempC > 42 && hour >= 11 && hour <= 18 {
		p.fireDisruption(zone, "extreme_heat", "medium", tempC, now)
	}
}

func (p *WeatherPoller) fireDisruption(zone models.Zone, disruptionType, severity string, value float64, now time.Time) {
	// Deduplicate: skip if an active disruption of same type exists in zone within last 2 hours
	var existing models.Disruption
	err := p.DB.Where(
		"zone_id = ? AND type = ? AND created_at >= ?",
		zone.ID, disruptionType, now.Add(-2*time.Hour),
	).First(&existing).Error
	if err == nil {
		return // already have a recent disruption of this type
	}

	confidence := 0.82 + (value/100)*0.12
	if confidence > 0.99 {
		confidence = 0.99
	}
	signalTime := now
	confirmedAt := now.Add(15 * time.Minute)

	disruption := models.Disruption{
		ZoneID:         zone.ID,
		Type:           disruptionType,
		Severity:       severity,
		Confidence:     confidence,
		Status:         "confirmed",
		SignalTimestamp: &signalTime,
		ConfirmedAt:    &confirmedAt,
		StartTime:      &signalTime,
	}

	if err := p.DB.Create(&disruption).Error; err != nil {
		log.Printf("[WeatherPoller] Failed to create disruption for zone %s: %v", zone.Name, err)
		return
	}

	log.Printf("[WeatherPoller] ✅ Disruption created: %s in %s (confidence=%.2f)", disruptionType, zone.Name, confidence)
}
