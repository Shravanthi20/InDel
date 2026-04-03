package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PremiumQuote struct {
	WeeklyPremiumINR float64              `json:"weekly_premium_inr"`
	Currency         string               `json:"currency"`
	RiskScore        float64              `json:"risk_score,omitempty"`
	ShapBreakdown    []map[string]float64 `json:"-"`
	Explainability   []PremiumExplainItem `json:"explainability,omitempty"`
	ModelVersion     string               `json:"model_version,omitempty"`
	Source           string               `json:"source"`
}

type PremiumExplainItem struct {
	Feature string  `json:"feature"`
	Impact  float64 `json:"impact"`
}

type premiumMLRequest struct {
	WorkerID             string  `json:"worker_id"`
	ZoneID               string  `json:"zone_id"`
	City                 string  `json:"city"`
	State                string  `json:"state"`
	ZoneType             string  `json:"zone_type"`
	VehicleType          string  `json:"vehicle_type"`
	Season               string  `json:"season"`
	ExperienceDays       int     `json:"experience_days"`
	AvgDailyOrders       float64 `json:"avg_daily_orders"`
	AvgDailyEarnings     float64 `json:"avg_daily_earnings"`
	ActiveHoursPerDay    float64 `json:"active_hours_per_day"`
	RainfallMM           float64 `json:"rainfall_mm"`
	AQI                  float64 `json:"aqi"`
	Temperature          float64 `json:"temperature"`
	Humidity             float64 `json:"humidity"`
	OrderVolatility      float64 `json:"order_volatility"`
	EarningsVolatility   float64 `json:"earnings_volatility"`
	RecentDisruptionRate float64 `json:"recent_disruption_rate"`
}

type premiumMLResponse struct {
	Data struct {
		PremiumINR     float64 `json:"premium_inr"`
		RiskScore      float64 `json:"risk_score"`
		ModelVersion   string  `json:"model_version"`
		Explainability []struct {
			Feature string  `json:"feature"`
			Impact  float64 `json:"impact"`
		} `json:"explainability"`
	} `json:"data"`
}

type PremiumContext struct {
	WorkerID         uint
	ZoneID           uint
	City             string
	State            string
	ZoneType         string
	VehicleType      string
	ExperienceDays   int
	AvgDailyOrders   float64
	AvgDailyEarnings float64
	ActiveHours      float64
	RainfallMM       float64
	AQI              float64
	Temperature      float64
	Humidity         float64
	OrderVolatility  float64
	EarningsVol      float64
	DisruptionRate   float64
}

func QuotePremium(db *gorm.DB, workerID uint, now time.Time) (*PremiumQuote, error) {
	context, fallback, err := loadPremiumContext(db, workerID, now)
	if err != nil {
		return nil, err
	}

	quote, err := requestPremiumQuote(context)
	if err == nil {
		return quote, nil
	}

	return fallback, nil
}

func QuotePremiumForContext(context PremiumContext) *PremiumQuote {
	return fallbackPremiumQuote(context)
}

func requestPremiumQuote(context PremiumContext) (*PremiumQuote, error) {
	baseURL := strings.TrimSpace(os.Getenv("PREMIUM_ML_URL"))
	if baseURL == "" {
		return nil, fmt.Errorf("premium ml url missing")
	}

	payload := premiumMLRequest{
		WorkerID:             fmt.Sprintf("%d", context.WorkerID),
		ZoneID:               fmt.Sprintf("%d", context.ZoneID),
		City:                 context.City,
		State:                context.State,
		ZoneType:             context.ZoneType,
		VehicleType:          strings.ToLower(strings.TrimSpace(context.VehicleType)),
		Season:               currentSeason(time.Now().UTC()),
		ExperienceDays:       context.ExperienceDays,
		AvgDailyOrders:       context.AvgDailyOrders,
		AvgDailyEarnings:     context.AvgDailyEarnings,
		ActiveHoursPerDay:    context.ActiveHours,
		RainfallMM:           context.RainfallMM,
		AQI:                  context.AQI,
		Temperature:          context.Temperature,
		Humidity:             context.Humidity,
		OrderVolatility:      context.OrderVolatility,
		EarningsVolatility:   context.EarningsVol,
		RecentDisruptionRate: context.DisruptionRate,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/ml/v1/premium/calculate"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("premium ml returned %d", resp.StatusCode)
	}

	var decoded premiumMLResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}

	explainability := make([]PremiumExplainItem, 0, len(decoded.Data.Explainability))
	shap := make([]map[string]float64, 0, len(decoded.Data.Explainability))
	for _, item := range decoded.Data.Explainability {
		explainability = append(explainability, PremiumExplainItem{
			Feature: item.Feature,
			Impact:  math.Round(item.Impact*1000) / 1000,
		})
		shap = append(shap, map[string]float64{
			"impact": math.Round(item.Impact*1000) / 1000,
		})
	}

	return &PremiumQuote{
		WeeklyPremiumINR: decoded.Data.PremiumINR,
		Currency:         "INR",
		RiskScore:        decoded.Data.RiskScore,
		Explainability:   explainability,
		ModelVersion:     decoded.Data.ModelVersion,
		Source:           "premium-ml",
		ShapBreakdown:    shap,
	}, nil
}

func loadPremiumContext(db *gorm.DB, workerID uint, now time.Time) (PremiumContext, *PremiumQuote, error) {
	context := PremiumContext{
		WorkerID:         workerID,
		ZoneID:           1,
		City:             "Chennai",
		State:            "Tamil Nadu",
		ZoneType:         "urban",
		VehicleType:      "two_wheeler",
		ExperienceDays:   90,
		AvgDailyOrders:   12,
		AvgDailyEarnings: 600,
		ActiveHours:      8,
		RainfallMM:       18,
		AQI:              95,
		Temperature:      31,
		Humidity:         68,
		OrderVolatility:  0.24,
		EarningsVol:      0.21,
		DisruptionRate:   0.12,
	}

	if db == nil {
		return context, fallbackPremiumQuote(context), nil
	}

	type row struct {
		WorkerID        uint      `gorm:"column:worker_id"`
		ZoneID          uint      `gorm:"column:zone_id"`
		City            string    `gorm:"column:city"`
		State           string    `gorm:"column:state"`
		ZoneLevel       string    `gorm:"column:zone_level"`
		ZoneRisk        float64   `gorm:"column:zone_risk"`
		VehicleType     string    `gorm:"column:vehicle_type"`
		CreatedAt       time.Time `gorm:"column:created_at"`
		BaselineAmount  float64   `gorm:"column:baseline_amount"`
		WeeklyEarnings  float64   `gorm:"column:weekly_earnings"`
		PaidOrders      float64   `gorm:"column:paid_orders"`
		ActiveHours     float64   `gorm:"column:active_hours"`
		DisruptionCount float64   `gorm:"column:disruption_count"`
	}

	var data row
	weekStart, _ := weekBounds(now.UTC())
	err := db.Table("worker_profiles wp").
		Select(`
			wp.worker_id,
			COALESCE(wp.zone_id, 1) AS zone_id,
			COALESCE(z.city, 'Chennai') AS city,
			COALESCE(z.state, 'Tamil Nadu') AS state,
			COALESCE(z.level, 'A') AS zone_level,
			COALESCE(z.risk_rating, 0.45) AS zone_risk,
			COALESCE(wp.vehicle_type, 'two_wheeler') AS vehicle_type,
			COALESCE(u.created_at, CURRENT_TIMESTAMP) AS created_at,
			COALESCE(eb.baseline_amount, 4200) AS baseline_amount,
			COALESCE(wes.total_earnings, 0) AS weekly_earnings,
			COALESCE((
				SELECT COUNT(*)
				FROM orders o
				WHERE o.worker_id = wp.worker_id
				  AND o.created_at >= ?
			), 0) AS paid_orders,
			COALESCE((
				SELECT AVG(er.hours_worked)
				FROM earnings_records er
				WHERE er.worker_id = wp.worker_id
				  AND er.date >= ?
			), 8) AS active_hours,
			COALESCE((
				SELECT COUNT(*)
				FROM disruptions d
				WHERE d.zone_id = wp.zone_id
				  AND d.created_at >= ?
			), 0) AS disruption_count
		`, weekStart, weekStart.AddDate(0, 0, -28), now.UTC().AddDate(0, 0, -30)).
		Joins("LEFT JOIN zones z ON z.id = wp.zone_id").
		Joins("LEFT JOIN users u ON u.id = wp.worker_id").
		Joins("LEFT JOIN earnings_baselines eb ON eb.worker_id = wp.worker_id").
		Joins("LEFT JOIN weekly_earnings_summaries wes ON wes.worker_id = wp.worker_id AND wes.week_start = ?", weekStart).
		Where("wp.worker_id = ?", workerID).
		Scan(&data).Error
	if err != nil {
		return context, fallbackPremiumQuote(context), err
	}

	if data.WorkerID == 0 {
		return context, fallbackPremiumQuote(context), nil
	}

	context.WorkerID = data.WorkerID
	context.ZoneID = data.ZoneID
	context.City = safeString(data.City, "Chennai")
	context.State = safeString(data.State, "Tamil Nadu")
	context.ZoneType = zoneTypeFromLevel(data.ZoneLevel)
	context.VehicleType = safeString(data.VehicleType, "two_wheeler")
	context.ExperienceDays = maxInt(7, int(now.UTC().Sub(data.CreatedAt).Hours()/24))
	context.AvgDailyOrders = maxFloat(3, data.PaidOrders/7)
	baselineDaily := data.BaselineAmount / 7
	weeklyDaily := data.WeeklyEarnings / 7
	context.AvgDailyEarnings = maxFloat(250, math.Max(baselineDaily, weeklyDaily))
	context.ActiveHours = maxFloat(4, data.ActiveHours)
	context.OrderVolatility = clamp(0.12+data.ZoneRisk*0.38, 0.05, 0.95)
	context.EarningsVol = clamp(0.08+math.Abs(data.WeeklyEarnings-data.BaselineAmount)/maxFloat(1, data.BaselineAmount), 0.05, 0.95)
	context.DisruptionRate = clamp(data.DisruptionCount/12, 0.02, 0.95)
	context.RainfallMM = 18 + data.ZoneRisk*35
	context.AQI = 80 + data.ZoneRisk*90
	context.Temperature = 28 + data.ZoneRisk*8
	context.Humidity = 58 + data.ZoneRisk*25

	return context, fallbackPremiumQuote(context), nil
}

func fallbackPremiumQuote(context PremiumContext) *PremiumQuote {
	vehicleFactor := 1.0
	switch strings.ToLower(strings.TrimSpace(context.VehicleType)) {
	case "bike":
		vehicleFactor = 1.08
	case "scooter":
		vehicleFactor = 1.04
	case "two_wheeler":
		vehicleFactor = 1.06
	}

	riskScore := clamp(
		(context.OrderVolatility*0.24)+
			(context.EarningsVol*0.22)+
			(context.DisruptionRate*0.2)+
			(clamp(context.RainfallMM/100, 0, 1)*0.12)+
			(clamp(context.AQI/300, 0, 1)*0.1)+
			(clamp(context.Temperature/45, 0, 1)*0.12),
		0.1,
		0.95,
	)

	base := context.AvgDailyEarnings * 0.0375
	premium := clamp(base*(0.72+riskScore)*vehicleFactor, 10, 40)

	explainability := []PremiumExplainItem{
		{Feature: "order_volatility", Impact: roundTo(context.OrderVolatility*0.34, 3)},
		{Feature: "recent_disruption_rate", Impact: roundTo(context.DisruptionRate*0.28, 3)},
		{Feature: "earnings_volatility", Impact: roundTo(context.EarningsVol*0.22, 3)},
		{Feature: "weather_risk", Impact: roundTo(clamp(context.RainfallMM/100, 0, 1)*0.16, 3)},
	}

	return &PremiumQuote{
		WeeklyPremiumINR: roundTo(premium, 2),
		Currency:         "INR",
		RiskScore:        roundTo(riskScore, 3),
		Explainability:   explainability,
		ModelVersion:     "fallback_rule_v2",
		Source:           "fallback",
	}
}

func currentSeason(now time.Time) string {
	switch now.Month() {
	case time.June, time.July, time.August, time.September:
		return "monsoon"
	case time.October, time.November, time.December, time.January:
		return "winter"
	default:
		return "summer"
	}
}

func zoneTypeFromLevel(level string) string {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "C", "D":
		return "dense_urban"
	case "B":
		return "urban"
	default:
		return "suburban"
	}
}

func safeString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func roundTo(value float64, precision int) float64 {
	scale := math.Pow(10, float64(precision))
	return math.Round(value*scale) / scale
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
