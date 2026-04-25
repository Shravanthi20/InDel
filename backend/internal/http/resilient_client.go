package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RetryConfig defines retry behavior for HTTP requests
type RetryConfig struct {
	MaxRetries       int           // Maximum number of retry attempts
	InitialDelay     time.Duration // Initial delay between retries
	MaxDelay         time.Duration // Maximum delay between retries
	BackoffFactor    float64       // Multiplier for exponential backoff
	RetryableStatus  []int         // HTTP status codes that should be retried
	NonRetryableStatus []int       // HTTP status codes that should NOT be retried
}

// DefaultRetryConfig provides sensible defaults for HTTP requests
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		RetryableStatus: []int{
			http.StatusRequestTimeout,      // 408
			http.StatusTooManyRequests,     // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout,      // 504
		},
		NonRetryableStatus: []int{
			http.StatusBadRequest,           // 400
			http.StatusUnauthorized,         // 401
			http.StatusForbidden,            // 403
			http.StatusNotFound,             // 404
			http.StatusMethodNotAllowed,     // 405
			http.StatusNotAcceptable,        // 406
			http.StatusConflict,             // 409
			http.StatusGone,                 // 410
			http.StatusLengthRequired,       // 411
			http.StatusPreconditionFailed,   // 412
			http.StatusRequestEntityTooLarge, // 413
			http.StatusRequestURITooLong,   // 414
			http.StatusUnsupportedMediaType, // 415
			http.StatusRequestedRangeNotSatisfiable, // 416
			http.StatusExpectationFailed,   // 417
			http.StatusUnprocessableEntity, // 422
			http.StatusLocked,               // 423
			http.StatusFailedDependency,     // 424
			http.StatusUpgradeRequired,     // 426
			http.StatusPreconditionRequired, // 428
			http.StatusTooManyRequests,     // 429 (already in retryable, but also handled specially)
			http.StatusRequestHeaderFieldsTooLarge, // 431
		},
	}
}

// ResilientClient provides HTTP client with retry, timeout, and circuit breaker capabilities
type ResilientClient struct {
	httpClient    *http.Client
	retryConfig   RetryConfig
	defaultHeaders map[string]string
	baseURL       string
}

// NewResilientClient creates a new resilient HTTP client
func NewResilientClient(timeout time.Duration, retryConfig RetryConfig) *ResilientClient {
	if retryConfig.MaxRetries == 0 {
		retryConfig = DefaultRetryConfig()
	}

	return &ResilientClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		retryConfig:   retryConfig,
		defaultHeaders: make(map[string]string),
	}
}

// WithBaseURL sets the base URL for all requests
func (rc *ResilientClient) WithBaseURL(baseURL string) *ResilientClient {
	rc.baseURL = strings.TrimRight(baseURL, "/")
	return rc
}

// WithDefaultHeader sets a default header for all requests
func (rc *ResilientClient) WithDefaultHeader(key, value string) *ResilientClient {
	rc.defaultHeaders[key] = value
	return rc
}

// WithDefaultHeaders sets multiple default headers
func (rc *ResilientClient) WithDefaultHeaders(headers map[string]string) *ResilientClient {
	for k, v := range headers {
		rc.defaultHeaders[k] = v
	}
	return rc
}

// Do executes an HTTP request with retry logic
func (rc *ResilientClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Apply default headers
	for key, value := range rc.defaultHeaders {
		if req.Header.Get(key) == "" {
			req.Header.Set(key, value)
		}
	}

	// Apply correlation ID from context if available
	if corrID := getCorrelationID(ctx); corrID != "" {
		if req.Header.Get("X-Correlation-ID") == "" {
			req.Header.Set("X-Correlation-ID", corrID)
		}
		if req.Header.Get("X-Request-ID") == "" {
			req.Header.Set("X-Request-ID", corrID)
		}
	}

	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= rc.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := rc.calculateDelay(attempt)
			
			// Wait for delay or context cancellation
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during HTTP retry: %w", ctx.Err())
			}
		}

		// Clone request for each attempt
		retryReq, err := cloneRequest(req)
		if err != nil {
			return nil, fmt.Errorf("failed to clone request: %w", err)
		}

		// Execute request
		resp, err := rc.httpClient.Do(retryReq.WithContext(ctx))
		if err != nil {
			lastErr = err
			
			// Check if error is retryable (network errors, timeouts, etc.)
			if !rc.isRetryableNetworkError(err) {
				return nil, fmt.Errorf("non-retryable network error: %w", err)
			}
			
			// Log retry attempt
			fmt.Printf("[HTTP-RETRY] Network error (attempt %d/%d): %v\n", 
				attempt+1, rc.retryConfig.MaxRetries+1, err)
			continue
		}

		// Check if response status should be retried
		if rc.shouldRetryStatus(resp.StatusCode) {
			// Close response body before retry
			resp.Body.Close()
			
			lastResp = resp
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			
			fmt.Printf("[HTTP-RETRY] HTTP %d (attempt %d/%d)\n", 
				resp.StatusCode, attempt+1, rc.retryConfig.MaxRetries+1)
			continue
		}

		// Success or non-retryable status
		return resp, nil
	}

	// All retries exhausted
	if lastResp != nil {
		lastResp.Body.Close()
	}
	
	return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", 
		rc.retryConfig.MaxRetries+1, lastErr)
}

// Get executes a GET request
func (rc *ResilientClient) Get(ctx context.Context, url string) (*http.Response, error) {
	fullURL := rc.resolveURL(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	
	return rc.Do(ctx, req)
}

// Post executes a POST request
func (rc *ResilientClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	fullURL := rc.resolveURL(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	
	return rc.Do(ctx, req)
}

// PostJSON executes a POST request with JSON body
func (rc *ResilientClient) PostJSON(ctx context.Context, url string, data interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return rc.Post(ctx, url, "application/json", bytes.NewReader(jsonBody))
}

// Put executes a PUT request
func (rc *ResilientClient) Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	fullURL := rc.resolveURL(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create PUT request: %w", err)
	}
	
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	
	return rc.Do(ctx, req)
}

// PutJSON executes a PUT request with JSON body
func (rc *ResilientClient) PutJSON(ctx context.Context, url string, data interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return rc.Put(ctx, url, "application/json", bytes.NewReader(jsonBody))
}

// Delete executes a DELETE request
func (rc *ResilientClient) Delete(ctx context.Context, url string) (*http.Response, error) {
	fullURL := rc.resolveURL(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}
	
	return rc.Do(ctx, req)
}

// resolveURL resolves a URL against the base URL
func (rc *ResilientClient) resolveURL(url string) string {
	if rc.baseURL == "" {
		return url
	}
	
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	
	return fmt.Sprintf("%s/%s", rc.baseURL, strings.TrimLeft(url, "/"))
}

// calculateDelay computes the delay for a given retry attempt
func (rc *ResilientClient) calculateDelay(attempt int) time.Duration {
	delay := float64(rc.retryConfig.InitialDelay) * 
		pow(rc.retryConfig.BackoffFactor, float64(attempt-1))
	
	if delay > float64(rc.retryConfig.MaxDelay) {
		delay = float64(rc.retryConfig.MaxDelay)
	}
	
	return time.Duration(delay)
}

// pow calculates x^y (simple implementation for backoff calculation)
func pow(x, y float64) float64 {
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	return result
}

// shouldRetryStatus determines if an HTTP status code should be retried
func (rc *ResilientClient) shouldRetryStatus(status int) bool {
	// Check non-retryable status codes first (explicit exclusion)
	for _, nonRetryable := range rc.retryConfig.NonRetryableStatus {
		if status == nonRetryable {
			return false
		}
	}
	
	// Check retryable status codes
	for _, retryable := range rc.retryConfig.RetryableStatus {
		if status == retryable {
			return true
		}
	}
	
	// Default: don't retry other status codes
	return false
}

// isRetryableNetworkError determines if a network error should be retried
func (rc *ResilientClient) isRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	
	// Retryable network errors
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"network is unreachable",
		"no route to host",
		"connection timed out",
		"temporary failure",
		"resource temporarily unavailable",
		"connection lost",
		"server has gone away",
		"tls handshake",
		"certificate",
		"dns resolution failed",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryable) {
			return true
		}
	}

	// Check for specific error types
	if strings.Contains(errStr, "net/http: request canceled") {
		return false // Context cancellation is not retryable
	}
	
	return true // Default to retrying unknown network errors
}

// cloneRequest creates a deep copy of an HTTP request
func cloneRequest(req *http.Request) (*http.Request, error) {
	// Clone the request
	clone := req.Clone(req.Context())
	
	// Clone the body if it exists
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		
		// Restore original body
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		
		// Set cloned body
		clone.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	
	return clone, nil
}

// getCorrelationID extracts correlation ID from context
func getCorrelationID(ctx context.Context) string {
	if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
		return reqID
	}
	if corrID, ok := ctx.Value("correlation_id").(string); ok && corrID != "" {
		return corrID
	}
	return ""
}

// ValidateURL validates that a URL is properly formatted
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}
	
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	
	return nil
}
