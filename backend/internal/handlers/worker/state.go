package worker

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type appData struct {
	PhoneToOTP      map[string]string
	TokenToWorkerID map[string]string
	WorkerProfiles  map[string]map[string]any
	Policy          map[string]any
	Earnings        map[string]any
	Claims          []map[string]any
	Wallet          map[string]any
	Payouts         []map[string]any
	Orders          []map[string]any
	Notifications   []map[string]any
}

type stateStore struct {
	mu   sync.RWMutex
	data *appData
}

var store = &stateStore{data: newDefaultAppData()}

func newDefaultAppData() *appData {
	return &appData{
		PhoneToOTP:      map[string]string{"+919999999999": "123456"},
		TokenToWorkerID: map[string]string{"mock-jwt-token": "worker-001"},
		WorkerProfiles: map[string]map[string]any{
			"worker-001": {
				"worker_id":        "worker-001",
				"name":             "Gayathri Worker",
				"phone":            "+919999999999",
				"zone_level":       "a",
				"zone_name":        "Tambaram",
				"zone":             "Tambaram, Chennai",
				"vehicle_type":     "bike",
				"upi_id":           "gayathri@upi",
				"coverage_status":  "active",
				"enrolled":         true,
				"online":           true,
				"orders_completed": 0,
			},
		},
		Policy: map[string]any{
			"policy_id":          "pol-001",
			"status":             "active",
			"weekly_premium_inr": 22,
			"coverage_ratio":     0.8,
			"zone":               "Tambaram, Chennai",
			"next_due_date":      "2026-03-30",
			"shap_breakdown": []map[string]any{
				{"feature": "rain_risk", "impact": 0.42},
				{"feature": "order_drop_volatility", "impact": 0.31},
				{"feature": "historical_disruptions", "impact": 0.27},
			},
		},
		Earnings: map[string]any{
			"currency":           "INR",
			"this_week_actual":   3120,
			"this_week_baseline": 4080,
			"protected_income":   3264,
			"history": []map[string]any{
				{"week": "2026-W08", "actual": 3520, "baseline": 3980},
				{"week": "2026-W09", "actual": 3410, "baseline": 4010},
				{"week": "2026-W10", "actual": 3290, "baseline": 4050},
				{"week": "2026-W11", "actual": 3120, "baseline": 4080},
			},
		},
		Claims: []map[string]any{
			{
				"claim_id":          "clm-001",
				"status":            "approved",
				"zone":              "Tambaram, Chennai",
				"disruption_type":   "heavy_rain",
				"disruption_window": map[string]any{"start": "2026-03-18T11:00:00Z", "end": "2026-03-18T16:00:00Z"},
				"income_loss":       870,
				"payout_amount":     696,
				"fraud_verdict":     "clear",
				"created_at":        "2026-03-18T16:20:00Z",
			},
		},
		Wallet: map[string]any{
			"currency":           "INR",
			"available_balance":  1580,
			"last_payout_amount": 696,
			"last_payout_at":     "2026-03-19T09:10:00Z",
		},
		Payouts: []map[string]any{
			{
				"payout_id":    "pay-001",
				"claim_id":     "clm-001",
				"amount":       696,
				"method":       "upi",
				"status":       "processed",
				"processed_at": "2026-03-19T09:10:00Z",
			},
		},
		Orders: []map[string]any{
			{
				"order_id":    "ord-001",
				"pickup_area": "Tambaram West",
				"drop_area":   "Selaiyur",
				"distance_km": 3.8,
				"earning_inr": 78,
				"status":      "assigned",
				"assigned_at": "2026-03-23T11:10:00Z",
			},
			{
				"order_id":    "ord-002",
				"pickup_area": "Chromepet",
				"drop_area":   "Pallavaram",
				"distance_km": 2.9,
				"earning_inr": 62,
				"status":      "assigned",
				"assigned_at": "2026-03-23T11:15:00Z",
			},
		},
		Notifications: []map[string]any{
			{
				"id":         "ntf-001",
				"type":       "disruption_alert",
				"title":      "Heavy rain detected",
				"body":       "Heavy rain detected in Tambaram. You are protected.",
				"created_at": "2026-03-23T10:00:00Z",
				"read":       false,
			},
			{
				"id":         "ntf-002",
				"type":       "payout_credited",
				"title":      "Payout credited",
				"body":       "Rs 696 credited to your wallet for claim clm-001.",
				"created_at": "2026-03-23T10:30:00Z",
				"read":       false,
			},
		},
	}
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func (s *stateStore) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = newDefaultAppData()
}

func nextID(prefix string, current int) string {
	return fmt.Sprintf("%s-%03d", prefix, current+1)
}

func parseBody(c *gin.Context) map[string]any {
	var body map[string]any
	_ = c.ShouldBindJSON(&body)
	if body == nil {
		return map[string]any{}
	}
	return body
}

func bodyString(body map[string]any, key string, fallback string) string {
	v, ok := body[key]
	if !ok || v == nil {
		return fallback
	}
	if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
		return s
	}
	return fallback
}

func bodyInt(body map[string]any, key string, fallback int) int {
	v, ok := body[key]
	if !ok || v == nil {
		return fallback
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return fallback
	}
}

func requireAuth(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_or_invalid_bearer_token"})
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

	if hasDB() {
		type tokenRow struct {
			UserID uint `gorm:"column:user_id"`
		}
		var row tokenRow
		err := workerDB.Raw(
			"SELECT user_id FROM auth_tokens WHERE token = ? AND expires_at > CURRENT_TIMESTAMP LIMIT 1",
			token,
		).Scan(&row).Error
		if err == nil && row.UserID != 0 {
			return fmt.Sprintf("%d", row.UserID), true
		}
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	workerID, ok := store.data.TokenToWorkerID[token]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unknown_token"})
		return "", false
	}
	return workerID, true
}
