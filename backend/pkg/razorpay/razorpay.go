package razorpay

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	RazorpayBaseURL = "https://api.razorpay.com/v1"
	SandboxBaseURL  = "https://api.razorpay.com/v1/test_mode_true"
)

// RazorpayClient handles Razorpay payments
type RazorpayClient struct {
	APIKey     string
	APISecret  string
	IsTestMode bool
	HTTPClient *http.Client
}

// NewRazorpayClient creates a new Razorpay client
func NewRazorpayClient(apiKey, apiSecret string, isTestMode bool) *RazorpayClient {
	return &RazorpayClient{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		IsTestMode: isTestMode,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// PayoutRequest represents a payout request to Razorpay
type PayoutRequest struct {
	AccountNumber string                 `json:"account_number"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Mode          string                 `json:"mode"`
	Purpose       string                 `json:"purpose"`
	Recipient     map[string]interface{} `json:"recipient,omitempty"`
	Reference     string                 `json:"reference_id"`
}

// PayoutResponse represents response from Razorpay
type PayoutResponse struct {
	ID            string `json:"id"`
	Entity        string `json:"entity"`
	Fund          string `json:"fund_account_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Mode          string `json:"mode"`
	Purpose       string `json:"purpose"`
	Status        string `json:"status"`
	FeeType       string `json:"fee_type"`
	Fee           int64  `json:"fee"`
	Tax           int64  `json:"tax"`
	UtrNumber     string `json:"utr"`
	Reference1    string `json:"reference_id"`
	Reference2    string `json:"reference_id2"`
	Narration     string `json:"narration"`
	StatusCode    string `json:"status_code"`
	FailureReason string `json:"failure_reason"`
	CreatedAt     int64  `json:"created_at"`
}

// ErrorResponse represents error response from Razorpay
type ErrorResponse struct {
	Error struct {
		Code        string `json:"code"`
		Description string `json:"description"`
		Source      string `json:"source"`
		Reason      string `json:"reason"`
		Field       string `json:"field"`
	} `json:"error"`
}

// CreatePayout creates a payout via Razorpay API
func (r *RazorpayClient) CreatePayout(workerID uint, amount float64, upi string) (string, error) {
	if r.APIKey == "" || r.APISecret == "" {
		return fmt.Sprintf("rzp_mock_%d", workerID), nil // Mock mode when no credentials
	}

	// Convert amount to paise (Razorpay uses smallest currency unit)
	amountInPaise := int64(amount * 100)

	payload := map[string]interface{}{
		"account_number": "0112220061746180", // Test account number
		"amount":         amountInPaise,
		"currency":       "INR",
		"mode":           "UPI",
		"purpose":        "payout",
		"reference_id":   fmt.Sprintf("pay_wkr_%d_%d", workerID, time.Now().Unix()),
		"recipients": map[string]interface{}{
			"upi": upi,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", RazorpayBaseURL+"/payouts", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	auth := base64.StdEncoding.EncodeToString([]byte(r.APIKey + ":" + r.APISecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Entity-Id", "0112220061746180")

	// Execute request
	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return "", fmt.Errorf("payout failed with status %d: %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("payout failed: %s - %s", errResp.Error.Code, errResp.Error.Description)
	}

	var payoutResp PayoutResponse
	if err := json.Unmarshal(body, &payoutResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if payoutResp.ID == "" {
		return "", fmt.Errorf("invalid response: no payout id")
	}

	return payoutResp.ID, nil
}

// CheckPayoutStatus checks the status of a payout
func (r *RazorpayClient) CheckPayoutStatus(payoutID string) (string, error) {
	if r.APIKey == "" || r.APISecret == "" {
		// Mock response
		if payoutID == "rzp_mock_failed" {
			return "failed", nil
		}
		return "processed", nil
	}

	req, err := http.NewRequest("GET", RazorpayBaseURL+"/payouts/"+payoutID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(r.APIKey + ":" + r.APISecret))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var payoutResp PayoutResponse
	if err := json.Unmarshal(body, &payoutResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Map Razorpay status to internal status
	switch payoutResp.Status {
	case "processed":
		return "processed", nil
	case "pending":
		return "processing", nil
	case "queued":
		return "queued", nil
	case "failed", "rejected":
		return "failed", nil
	case "reversed":
		return "failed", nil
	default:
		return "unknown", nil
	}
}

// CreateFundAccount creates a fund account for UPI payout
func (r *RazorpayClient) CreateFundAccount(upi string) (string, error) {
	if r.APIKey == "" || r.APISecret == "" {
		return fmt.Sprintf("fa_mock_%s", upi), nil
	}

	payload := map[string]interface{}{
		"account_id":   "0112220061746180",
		"account_type": "vpa",
		"vpa": map[string]string{
			"address": upi,
		},
		"batch_id": "",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", RazorpayBaseURL+"/fund_accounts", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(r.APIKey + ":" + r.APISecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("fund account creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var fundResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &fundResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return fundResp.ID, nil
}
