package razorpay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RazorpayClient handles Razorpay payments
type RazorpayClient struct {
	APIKey     string
	APISecret  string
	MockMode   bool
	HTTPClient *http.Client
	BaseURL    string
}

// CreatePayoutRequest represents Razorpay payout creation request
type CreatePayoutRequest struct {
	Account_ID       string `json:"account_id,omitempty"`
	Amount           int64  `json:"amount"`
	Currency         string `json:"currency"`
	Mode             string `json:"mode"`
	Purpose          string `json:"purpose"`
	Recipient_ID     string `json:"-"`
	Fund_Account_ID  string `json:"fund_account_id,omitempty"`
	Queue_If_Low_Bal bool   `json:"queue_if_low_bal,omitempty"`
}

// CreatePayoutResponse represents Razorpay payout response
type CreatePayoutResponse struct {
	ID               string `json:"id"`
	Amount           int64  `json:"amount"`
	Status           string `json:"status"`
	FundAccountID    string `json:"fund_account_id"`
	Error            string `json:"error,omitempty"`
	ErrorCode        string `json:"error_code,omitempty"`
	ErrorReason      string `json:"error_reason,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// NewRazorpayClient creates a new Razorpay client
func NewRazorpayClient(apiKey, apiSecret string) *RazorpayClient {
	mockMode := apiKey == "" || apiSecret == "" || strings.Contains(strings.ToLower(apiKey), "test")

	client := &RazorpayClient{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		MockMode:   mockMode,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    "https://api.razorpay.com",
	}

	// Use test/mock URL for mock mode
	if mockMode {
		client.BaseURL = "https://api.razorpay.com"
	}

	return client
}

// CreatePayout creates a payout in Razorpay
func (r *RazorpayClient) CreatePayout(workerID uint, amount float64, UPI string) (string, error) {
	// Mock mode - return mock ID
	if r.MockMode {
		return r.GenerateMockPayoutID(workerID), nil
	}

	// Convert amount to paise (1 rupee = 100 paise)
	amountInPaise := int64(amount * 100)

	// Create request
	req := CreatePayoutRequest{
		Amount:           amountInPaise,
		Currency:         "INR",
		Mode:             "UPI",
		Purpose:          "Worker Payout",
		Fund_Account_ID:  UPI,
		Queue_If_Low_Bal: true,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payout request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", r.BaseURL+"/v1/payouts", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	// Add Basic Auth header
	httpReq.SetBasicAuth(r.APIKey, r.APISecret)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := r.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to create payout: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var payoutResp CreatePayoutResponse
	if err := json.Unmarshal(bodyBytes, &payoutResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		errMsg := fmt.Sprintf("Razorpay API error: %s", payoutResp.ErrorDescription)
		if payoutResp.Error != "" {
			errMsg = fmt.Sprintf("Razorpay API error: %s (%s): %s", payoutResp.ErrorCode, payoutResp.Error, payoutResp.ErrorDescription)
		}
		return "", fmt.Errorf("%s", errMsg)
	}

	if payoutResp.ID == "" {
		return "", fmt.Errorf("razorpay returned empty payout ID")
	}

	return payoutResp.ID, nil
}

// CheckPayoutStatus checks the status of a payout
func (r *RazorpayClient) CheckPayoutStatus(payoutID string) (string, error) {
	if r.MockMode {
		return "processed", nil
	}

	httpReq, err := http.NewRequest("GET", r.BaseURL+"/v1/payouts/"+payoutID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.SetBasicAuth(r.APIKey, r.APISecret)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := r.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to check payout status: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var statusResp CreatePayoutResponse
	if err := json.Unmarshal(bodyBytes, &statusResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return statusResp.Status, nil
}

// CreateFundAccount creates a fund account for UPI
func (r *RazorpayClient) CreateFundAccount(UPI string) (string, error) {
	if r.MockMode {
		return "fa_" + strings.ReplaceAll(UPI, "@", "_"), nil
	}

	// Implementation for creating fund accounts if needed
	return "", fmt.Errorf("not implemented yet")
}

// GenerateMockPayoutID generates a mock Razorpay payout ID for testing
func (r *RazorpayClient) GenerateMockPayoutID(workerID uint) string {
	return fmt.Sprintf("rzp_mock_%d_%d", time.Now().Unix(), workerID)
}

// IsTransientError determines if the error is transient and retry-eligible
func IsTransientError(errMsg string) bool {
	transientErrors := []string{
		"TIMEOUT",
		"CONNECTION",
		"GATEWAY_ERROR",
		"SERVER_ERROR",
		"TEMPORARY_FAILURE",
		"429",
		"503",
		"502",
		"504",
	}

	upperMsg := strings.ToUpper(errMsg)
	for _, errType := range transientErrors {
		if strings.Contains(upperMsg, errType) {
			return true
		}
	}
	return false
}
