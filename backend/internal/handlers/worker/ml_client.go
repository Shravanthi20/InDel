package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// MLPremiumRequest mirrors the premium service request schema
type MLPremiumRequest struct {
	WorkerID              string  `json:"worker_id"`
	ZoneID                string  `json:"zone_id"`
	City                  string  `json:"city"`
	State                 string  `json:"state"`
	ZoneType              string  `json:"zone_type"`
	VehicleType           string  `json:"vehicle_type"`
	Season                string  `json:"season"`
	ExperienceDays        int     `json:"experience_days"`
	AvgDailyOrders        float64 `json:"avg_daily_orders"`
	AvgDailyEarnings      float64 `json:"avg_daily_earnings"`
	ActiveHoursPerDay     float64 `json:"active_hours_per_day"`
	RainfallMm            float64 `json:"rainfall_mm"`
	AQI                   float64 `json:"aqi"`
	Temperature           float64 `json:"temperature"`
	Humidity              float64 `json:"humidity"`
	OrderVolatility       float64 `json:"order_volatility"`
	EarningsVolatility    float64 `json:"earnings_volatility"`
	RecentDisruptionRate  float64 `json:"recent_disruption_rate"`
}

// MLExplainabilityFactor represents a single SHAP factor
type MLExplainabilityFactor struct {
	Feature string  `json:"feature"`
	Impact  float64 `json:"impact"`
}

// MLPremiumData contains the premium prediction result
type MLPremiumData struct {
	WorkerID           string                     `json:"worker_id"`
	PremiumInr         float64                    `json:"premium_inr"`
	RiskScore          float64                    `json:"risk_score"`
	Explainability     []MLExplainabilityFactor   `json:"explainability"`
	ModelVersion       string                     `json:"model_version"`
}

// MLPremiumResponse is the complete response from the ML service
type MLPremiumResponse struct {
	Data MLPremiumData `json:"data"`
	Meta struct {
		RequestID string `json:"request_id"`
		Timestamp string `json:"timestamp"`
	} `json:"meta"`
}

var mlClient *http.Client

func init() {
	mlClient = &http.Client{
		Timeout: 5 * time.Second,
	}
}

// getPremiumFromML calls the premium pricing ML service
func getPremiumFromML(req MLPremiumRequest) (*MLPremiumData, error) {
	mlURL := os.Getenv("PREMIUM_ML_URL")
	if mlURL == "" {
		mlURL = "http://premium-ml:8000/ml/v1/premium/calculate"
	}
	
	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Printf("[ML] Failed to marshal request: %v", err)
		return nil, err
	}

	// Make HTTP call
	httpReq, err := http.NewRequest("POST", mlURL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("[ML] Failed to create request: %v", err)
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := mlClient.Do(httpReq)
	if err != nil {
		log.Printf("[ML] Failed to call premium service: %v (URL: %s)", err, mlURL)
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ML] Failed to read response: %v", err)
		return nil, err
	}

	// Check HTTP status
	if resp.StatusCode != 200 {
		log.Printf("[ML] Premium service returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("premium service returned status %d", resp.StatusCode)
	}

	// Unmarshal response
	var mlResp MLPremiumResponse
	err = json.Unmarshal(body, &mlResp)
	if err != nil {
		log.Printf("[ML] Failed to unmarshal response: %v", err)
		return nil, err
	}

	return &mlResp.Data, nil
}

// buildMLPremiumRequest assembles a request for the ML service from worker data
func buildMLPremiumRequest(workerID string, profile map[string]interface{}) MLPremiumRequest {
	// Extract fields from profile
	zoneID := mapGetString(profile, "zone_id", "")
	zoneLevel := strings.ToUpper(strings.TrimSpace(mapGetString(profile, "zone_level", "A")))
	city := mapGetString(profile, "city", "Chennai")
	vehicleType := mapGetString(profile, "vehicle_type", "two_wheeler")
	state := mapGetString(profile, "state", "Tamil Nadu")

	// Infer zone_type from zone_level
	zoneType := "urban"
	switch zoneLevel {
	case "A":
		zoneType = "urban"
	case "B":
		zoneType = "tier2"
	case "C":
		zoneType = "coastal"
	default:
		if zoneLevel == "coastal" {
			zoneType = "coastal"
		} else if zoneLevel == "pollution_heavy" {
			zoneType = "pollution_heavy"
		} else if zoneLevel == "tier2" {
			zoneType = "tier2"
		} else if zoneLevel == "dry" {
			zoneType = "dry"
		}
	}

	city = normalizeCity(city)
	zoneType = normalizeZoneType(zoneType)
	vehicleType = normalizeVehicleType(vehicleType)
	state = normalizeState(state, city)

	if zoneID == "" || !strings.HasPrefix(zoneID, "zone_") {
		zoneID = defaultZoneID(city, zoneType)
	}

	// Set reasonable defaults for worker/environment metrics
	// In production, these would be queried from the database (earnings, order history, etc.)
	experienceDays := mapGetInt(profile, "experience_days", 500)
	avgDailyOrders := mapGetFloat(profile, "avg_daily_orders", 18.0)
	avgDailyEarnings := mapGetFloat(profile, "avg_daily_earnings", 1000.0)
	activeHoursPerDay := mapGetFloat(profile, "active_hours_per_day", 8.5)

	// Geo defaults from selected zone/order traces. For B/C use midpoint of from/to.
	zoneLat := mapGetFloat(profile, "zone_lat", 0)
	zoneLon := mapGetFloat(profile, "zone_lon", 0)
	fromLat := mapGetFloat(profile, "from_lat", 0)
	fromLon := mapGetFloat(profile, "from_lon", 0)
	toLat := mapGetFloat(profile, "to_lat", 0)
	toLon := mapGetFloat(profile, "to_lon", 0)
	if (zoneLevel == "B" || zoneLevel == "C") && fromLat != 0 && toLat != 0 {
		zoneLat = (fromLat + toLat) / 2
		zoneLon = (fromLon + toLon) / 2
	} else if zoneLevel == "A" && fromLat != 0 {
		zoneLat = fromLat
		zoneLon = fromLon
	}

	if zoneLat == 0 || zoneLon == 0 {
		derivedLat, derivedLon := deriveZoneCoordinates(
			zoneLevel,
			city,
			mapGetString(profile, "zone_name", ""),
		)
		if derivedLat != 0 && derivedLon != 0 {
			zoneLat = derivedLat
			zoneLon = derivedLon
		}
	}

	// Environmental defaults (adjusted by zone level and geo context)
	rainfallMm := 2.0
	aqi := 80.0
	temperature := 28.0
	humidity := 65.0

	// Seasonal adjustments based on month (demo)
	month := time.Now().Month()
	season := "Summer"
	if month >= 6 && month <= 9 {
		season = "Monsoon"
		rainfallMm = 15.0
		humidity = 80.0
	} else if month >= 10 && month <= 2 {
		season = "Winter"
		temperature = 20.0
		humidity = 55.0
	}

	// Risk metrics (calibrated by zone level)
	orderVolatility := 0.20
	earningsVolatility := 0.18
	recentDisruptionRate := 0.05

	if zoneLevel == "B" {
		orderVolatility = 0.24
		earningsVolatility = 0.22
		recentDisruptionRate = 0.08
	} else if zoneLevel == "C" {
		orderVolatility = 0.28
		earningsVolatility = 0.26
		recentDisruptionRate = 0.11
		aqi += 10
		rainfallMm += 3
	}

	if zoneLat != 0 {
		temperature += (zoneLat - 12.9) * 0.15
	}
	if zoneLon != 0 {
		humidity += (zoneLon - 80.2) * 0.05
	}
	temperature = clampFloat(temperature, 18, 45)
	humidity = clampFloat(humidity, 30, 95)
	aqi = clampFloat(aqi, 40, 350)
	rainfallMm = clampFloat(rainfallMm, 0, 50)

	return MLPremiumRequest{
		WorkerID:             workerID,
		ZoneID:               zoneID,
		City:                 city,
		State:                state,
		ZoneType:             zoneType,
		VehicleType:          vehicleType,
		Season:               season,
		ExperienceDays:       experienceDays,
		AvgDailyOrders:       avgDailyOrders,
		AvgDailyEarnings:     avgDailyEarnings,
		ActiveHoursPerDay:    activeHoursPerDay,
		RainfallMm:           rainfallMm,
		AQI:                  aqi,
		Temperature:          temperature,
		Humidity:             humidity,
		OrderVolatility:      orderVolatility,
		EarningsVolatility:   earningsVolatility,
		RecentDisruptionRate: recentDisruptionRate,
	}
}

func normalizeCity(city string) string {
	allowed := map[string]string{
		"chennai":    "Chennai",
		"bengaluru":  "Bengaluru",
		"mumbai":     "Mumbai",
		"delhi":      "Delhi",
		"hyderabad":  "Hyderabad",
		"pune":       "Pune",
		"lucknow":    "Lucknow",
		"jaipur":     "Jaipur",
		"coimbatore": "Coimbatore",
		"indore":     "Indore",
		"kolkata":    "Kolkata",
		"ahmedabad":  "Ahmedabad",
	}
	if v, ok := allowed[strings.ToLower(strings.TrimSpace(city))]; ok {
		return v
	}
	return "Chennai"
}

func normalizeZoneType(zoneType string) string {
	v := strings.ToLower(strings.TrimSpace(zoneType))
	switch v {
	case "urban", "coastal", "pollution_heavy", "tier2", "dry":
		return v
	default:
		return "urban"
	}
}

func normalizeVehicleType(vehicleType string) string {
	v := strings.ToLower(strings.TrimSpace(vehicleType))
	switch v {
	case "bike", "two_wheeler", "scooter", "motorbike":
		return "two_wheeler"
	case "car", "four_wheeler":
		return "car"
	default:
		return "two_wheeler"
	}
}

func normalizeState(state, city string) string {
	if strings.TrimSpace(state) != "" {
		return state
	}
	cityState := map[string]string{
		"Chennai":    "Tamil Nadu",
		"Bengaluru":  "Karnataka",
		"Mumbai":     "Maharashtra",
		"Delhi":      "Delhi",
		"Hyderabad":  "Telangana",
		"Pune":       "Maharashtra",
		"Lucknow":    "Uttar Pradesh",
		"Jaipur":     "Rajasthan",
		"Coimbatore": "Tamil Nadu",
		"Indore":     "Madhya Pradesh",
		"Kolkata":    "West Bengal",
		"Ahmedabad":  "Gujarat",
	}
	if v, ok := cityState[city]; ok {
		return v
	}
	return "Tamil Nadu"
}

func defaultZoneID(city, zoneType string) string {
	cityZone := map[string]string{
		"Chennai":    "zone_chennai_urban",
		"Bengaluru":  "zone_bengaluru_urban",
		"Mumbai":     "zone_mumbai_coastal",
		"Delhi":      "zone_delhi_pollution_heavy",
		"Hyderabad":  "zone_hyderabad_urban",
		"Pune":       "zone_pune_urban",
		"Lucknow":    "zone_lucknow_tier2",
		"Jaipur":     "zone_jaipur_dry",
		"Coimbatore": "zone_coimbatore_tier2",
		"Indore":     "zone_indore_tier2",
		"Kolkata":    "zone_kolkata_pollution_heavy",
		"Ahmedabad":  "zone_ahmedabad_dry",
	}
	if zoneType == "coastal" {
		if city == "Mumbai" {
			return "zone_mumbai_coastal"
		}
		if city == "Chennai" {
			return "zone_chennai_urban"
		}
	}
	if v, ok := cityZone[city]; ok {
		return v
	}
	return "zone_chennai_urban"
}

func deriveZoneCoordinates(zoneLevel, city, zoneName string) (float64, float64) {
	cityCoords := map[string][2]float64{
		"chennai":    {13.0827, 80.2707},
		"bangalore":  {12.9716, 77.5946},
		"bengaluru":  {12.9716, 77.5946},
		"mumbai":     {19.0760, 72.8777},
		"hyderabad":  {17.3850, 78.4867},
		"delhi":      {28.6139, 77.2090},
		"coimbatore": {11.0168, 76.9558},
		"madurai":    {9.9252, 78.1198},
		"trichy":     {10.7905, 78.7047},
		"salem":      {11.6643, 78.1460},
		"tambaram":   {12.9249, 80.1000},
		"velachery":  {12.9791, 80.2209},
		"chromepet":  {12.9516, 80.1462},
		"selaiyur":   {12.9061, 80.1427},
	}

	normLevel := strings.ToUpper(strings.TrimSpace(zoneLevel))
	normCity := strings.ToLower(strings.TrimSpace(city))
	normZoneName := strings.TrimSpace(zoneName)

	if normLevel == "A" {
		if coords, ok := cityCoords[normCity]; ok {
			return coords[0], coords[1]
		}
		if coords, ok := cityCoords[strings.ToLower(normZoneName)]; ok {
			return coords[0], coords[1]
		}
		return 0, 0
	}

	fromCity, toCity := splitZonePathCities(normZoneName)
	if fromCity == "" && toCity == "" {
		if coords, ok := cityCoords[normCity]; ok {
			return coords[0], coords[1]
		}
		return 0, 0
	}

	from, fromOk := cityCoords[strings.ToLower(fromCity)]
	to, toOk := cityCoords[strings.ToLower(toCity)]
	if fromOk && toOk {
		return (from[0] + to[0]) / 2, (from[1] + to[1]) / 2
	}
	if fromOk {
		return from[0], from[1]
	}
	if toOk {
		return to[0], to[1]
	}
	return 0, 0
}

func splitZonePathCities(zoneName string) (string, string) {
	raw := strings.TrimSpace(zoneName)
	if raw == "" {
		return "", ""
	}

	clean := raw
	if idx := strings.Index(clean, "("); idx >= 0 {
		clean = strings.TrimSpace(clean[:idx])
	}
	if strings.Contains(clean, " to ") {
		parts := strings.SplitN(clean, " to ", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	if strings.Contains(clean, "-") {
		parts := strings.SplitN(clean, "-", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return clean, ""
}

// Helper functions for map access with defaults
func mapGetString(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultVal
}

func mapGetInt(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return int(f)
		} else if i, ok := val.(int); ok {
			return i
		}
	}
	return defaultVal
}

func mapGetFloat(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultVal
}

func clampFloat(value, min, max float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return min
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
