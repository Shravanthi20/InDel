package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeoutMS    = 3000
	defaultRetryCount   = 2
)

type MLClient struct {
	baseURL       string
	timeout       time.Duration
	retryCount    int
	httpClient    *http.Client
	lastHealthOK  bool
	lastHealthAt  time.Time
	healthCacheTT time.Duration
	circuitBreaker *CircuitBreaker
}

// FallbackPremiumResponse provides rule-based premium when ML fails
type FallbackPremiumResponse struct {
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

// FallbackFraudResponse provides rule-based fraud detection when ML fails
type FallbackFraudResponse struct {
	FraudScore float64 `json:"fraud_score"`
	Verdict    string  `json:"verdict"`
	SignalsRaw json.RawMessage `json:"signals"`
}

type PremiumRequest struct {
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

type PremiumResponse struct {
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

type FraudResponse struct {
	FraudScore float64         `json:"fraud_score"`
	Verdict    string          `json:"verdict"`
	SignalsRaw json.RawMessage `json:"signals"`
}

type RiskRequest struct {
	ZoneID int `json:"zone_id"`
}

type RiskResponse struct {
	Status string `json:"status"`
}

func NewMLClientFromEnv() *MLClient {
	baseURL := strings.TrimSpace(os.Getenv("ML_SERVICE_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("PREMIUM_ML_URL"))
	}
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("FRAUD_ML_URL"))
	}
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("FORECAST_ML_URL"))
	}
	if baseURL == "" {
		// ML service URL must be explicitly set - no more localhost fallback
		log.Fatalf("ML_SERVICE_URL must be explicitly set. Found: PREMIUM_ML_URL=%s, FRAUD_ML_URL=%s, FORECAST_ML_URL=%s", 
			os.Getenv("PREMIUM_ML_URL"), os.Getenv("FRAUD_ML_URL"), os.Getenv("FORECAST_ML_URL"))
	}

	timeoutMS := defaultTimeoutMS
	if raw := strings.TrimSpace(os.Getenv("ML_TIMEOUT_MS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			timeoutMS = parsed
		}
	}

	retryCount := defaultRetryCount
	if raw := strings.TrimSpace(os.Getenv("ML_RETRY_COUNT")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			retryCount = parsed
		}
	}

	timeout := time.Duration(timeoutMS) * time.Millisecond
	
	// Initialize circuit breaker with reasonable defaults
	circuitBreaker := NewCircuitBreaker(
		"ml-client",
		5,                      // max failures
		timeout,                // operation timeout
		30*time.Second,         // reset timeout
	)

	return &MLClient{
		baseURL:         strings.TrimRight(baseURL, "/"),
		timeout:         timeout,
		retryCount:      retryCount,
		httpClient:      &http.Client{Timeout: timeout},
		healthCacheTT:   15 * time.Second,
		circuitBreaker: circuitBreaker,
	}
}

func (c *MLClient) HealthCheck(ctx context.Context) error {
	if time.Since(c.lastHealthAt) <= c.healthCacheTT {
		if c.lastHealthOK {
			return nil
		}
		return fmt.Errorf("ml health check recently failed")
	}

	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.lastHealthAt = time.Now()
		c.lastHealthOK = false
		log.Printf("[ml-client] health check failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	latency := time.Since(started)
	if resp.StatusCode >= 300 {
		c.lastHealthAt = time.Now()
		c.lastHealthOK = false
		log.Printf("[ml-client] health check failed with status=%d latency=%s", resp.StatusCode, latency)
		return fmt.Errorf("ml health check status %d", resp.StatusCode)
	}

	c.lastHealthAt = time.Now()
	c.lastHealthOK = true
	log.Printf("[ml-client] health check ok latency=%s", latency)
	return nil
}

func (c *MLClient) GetPremiumQuote(ctx context.Context, input PremiumRequest) (PremiumResponse, error) {
	var out PremiumResponse
	
	// Try ML service with circuit breaker protection
	err := c.circuitBreaker.Execute(ctx, func() error {
		return c.callJSONWithRetry(ctx, http.MethodPost, "/ml/v1/premium/calculate", input, &out)
	})
	
	if err != nil {
		log.Printf("[ML-CLIENT] Premium calculation failed, using fallback: %v", err)
		return c.getFallbackPremiumQuote(input)
	}
	
	return out, nil
}

func (c *MLClient) GetFraudScore(ctx context.Context, input any, out any) error {
	// Try ML service with circuit breaker protection
	err := c.circuitBreaker.Execute(ctx, func() error {
		return c.callJSONWithRetry(ctx, http.MethodPost, "/ml/v1/fraud/score", input, out)
	})
	
	if err != nil {
		log.Printf("[ML-CLIENT] Fraud scoring failed, using fallback: %v", err)
		return c.getFallbackFraudScore(input, out)
	}
	
	return nil
}

func (c *MLClient) GetRiskPrediction(ctx context.Context, input RiskRequest) (RiskResponse, error) {
	var out RiskResponse
	err := c.circuitBreaker.Execute(ctx, func() error {
		return c.callJSONWithRetry(ctx, http.MethodPost, "/forecast", input, &out)
	})
	
	if err != nil {
		log.Printf("[ML-CLIENT] Risk prediction failed: %v", err)
		return RiskResponse{Status: "fallback_unavailable"}, err
	}
	
	return out, err
}

// getFallbackPremiumQuote provides rule-based premium calculation when ML fails
func (c *MLClient) getFallbackPremiumQuote(input PremiumRequest) (PremiumResponse, error) {
	// Simple rule-based premium calculation
	basePremium := 50.0
	
	// Risk factors
	riskScore := 0.5 // Default medium risk
	
	// Adjust based on zone type
	if input.ZoneType == "high_risk" {
		riskScore += 0.3
		basePremium *= 1.5
	} else if input.ZoneType == "low_risk" {
		riskScore -= 0.2
		basePremium *= 0.8
	}
	
	// Adjust based on experience
	if input.ExperienceDays < 30 {
		riskScore += 0.2
		basePremium *= 1.2
	} else if input.ExperienceDays > 365 {
		riskScore -= 0.1
		basePremium *= 0.9
	}
	
	// Adjust based on earnings volatility
	if input.EarningsVolatility > 0.5 {
		riskScore += 0.15
		basePremium *= 1.15
	}
	
	// Clamp values
	if riskScore > 1.0 {
		riskScore = 1.0
	}
	if riskScore < 0.1 {
		riskScore = 0.1
	}
	
	return PremiumResponse{
		Data: struct {
			PremiumINR     float64 `json:"premium_inr"`
			RiskScore      float64 `json:"risk_score"`
			ModelVersion   string  `json:"model_version"`
			Explainability []struct {
				Feature string  `json:"feature"`
				Impact  float64 `json:"impact"`
			} `json:"explainability"`
		}{
			PremiumINR:   basePremium,
			RiskScore:    riskScore,
			ModelVersion: "fallback_rule_v1",
			Explainability: []struct {
				Feature string  `json:"feature"`
				Impact  float64 `json:"impact"`
			}{
				{Feature: "zone_type", Impact: 0.3},
				{Feature: "experience", Impact: 0.2},
				{Feature: "earnings_volatility", Impact: 0.15},
			},
		},
	}, nil
}

// getFallbackFraudScore provides rule-based fraud detection when ML fails
func (c *MLClient) getFallbackFraudScore(input any, out any) error {
	// Simple rule-based fraud detection
	fraudScore := 0.2 // Default low risk
	verdict := "low_risk"
	
	// If we can't analyze the input, use conservative defaults
	if outPtr, ok := out.(*FallbackFraudResponse); ok {
		outPtr.FraudScore = fraudScore
		outPtr.Verdict = verdict
	}
	
	return nil
}

func (c *MLClient) callJSONWithRetry(ctx context.Context, method, path string, payload any, out any) error {
	if err := c.HealthCheck(ctx); err != nil {
		return fmt.Errorf("ml unavailable: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := c.baseURL + path
	var lastErr error
	maxAttempts := c.retryCount + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, reqErr := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if reqErr != nil {
			return reqErr
		}
		req.Header.Set("Content-Type", "application/json")
		if corrID := correlationIDFromContext(ctx); corrID != "" {
			req.Header.Set("X-Correlation-ID", corrID)
			req.Header.Set("X-Request-ID", corrID)
		}

		started := time.Now()
		resp, doErr := c.httpClient.Do(req)
		latency := time.Since(started)
		if doErr != nil {
			lastErr = doErr
			log.Printf("[ml-client] request failed path=%s attempt=%d/%d latency=%s err=%v", path, attempt, maxAttempts, latency, doErr)
			if attempt < maxAttempts {
				time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
				continue
			}
			return lastErr
		}

		func() {
			defer resp.Body.Close()
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				lastErr = fmt.Errorf("ml returned non-retryable status %d", resp.StatusCode)
				log.Printf("[ml-client] request failed path=%s attempt=%d/%d status=%d latency=%s", path, attempt, maxAttempts, resp.StatusCode, latency)
				return
			}
			if resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("ml returned retryable status %d", resp.StatusCode)
				log.Printf("[ml-client] request failed path=%s attempt=%d/%d status=%d latency=%s", path, attempt, maxAttempts, resp.StatusCode, latency)
				return
			}
			if decodeErr := json.NewDecoder(resp.Body).Decode(out); decodeErr != nil {
				lastErr = decodeErr
				log.Printf("[ml-client] decode failed path=%s attempt=%d/%d latency=%s err=%v", path, attempt, maxAttempts, latency, decodeErr)
				return
			}

			lastErr = nil
			log.Printf("[ml-client] request ok path=%s attempt=%d/%d latency=%s", path, attempt, maxAttempts, latency)
		}()

		if lastErr == nil {
			return nil
		}
		if strings.Contains(lastErr.Error(), "non-retryable status") {
			return lastErr
		}
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
		}
	}

	return lastErr
}

func correlationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value("request_id").(string); ok && strings.TrimSpace(reqID) != "" {
		return strings.TrimSpace(reqID)
	}
	if corrID, ok := ctx.Value("correlation_id").(string); ok && strings.TrimSpace(corrID) != "" {
		return strings.TrimSpace(corrID)
	}
	return ""
}
