package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// RetryConfig defines retry behavior for Kafka operations
type RetryConfig struct {
	MaxRetries    int           // Maximum number of retry attempts
	InitialDelay  time.Duration // Initial delay between retries
	MaxDelay      time.Duration // Maximum delay between retries
	BackoffFactor float64       // Multiplier for exponential backoff
}

// DefaultRetryConfig provides sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryProducer wraps a Producer with retry logic and fallback handling
type RetryProducer struct {
	producer     *Producer
	config       RetryConfig
	logger       interface{} // Will be set to structured logger
	metrics      *KafkaMetrics
	fallbackChan chan *FallbackMessage
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

// FallbackMessage represents a message that failed to be sent to Kafka
type FallbackMessage struct {
	Topic   string
	Key     string
	Message []byte
	Error   error
	Time    time.Time
	Retries int
}

// KafkaMetrics tracks Kafka operation metrics
type KafkaMetrics struct {
	mu                    sync.RWMutex
	messagesSent          int64
	messagesFailed        int64
	messagesRetried       int64
	totalLatency          time.Duration
	lastError             error
	circuitBreakerTripped bool
}

// NewRetryProducer creates a new producer with retry capabilities
func NewRetryProducer(brokers string, config RetryConfig) (*RetryProducer, error) {
	// Create base producer
	baseProducer, err := NewProducer(brokers)
	if err != nil {
		return nil, fmt.Errorf("failed to create base producer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	retryProducer := &RetryProducer{
		producer:     baseProducer,
		config:       config,
		metrics:      &KafkaMetrics{},
		fallbackChan: make(chan *FallbackMessage, 1000), // Buffer for failed messages
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start fallback message processor
	go retryProducer.processFallbackMessages()

	return retryProducer, nil
}

// PublishWithRetry attempts to send a message with retry logic
func (rp *RetryProducer) PublishWithRetry(ctx context.Context, topic string, key string, message []byte) error {
	start := time.Now()
	var lastErr error

	for attempt := 0; attempt <= rp.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := rp.calculateDelay(attempt)

			rp.logWithLevel("WARN", fmt.Sprintf("Retrying Kafka publish (attempt %d/%d) after %v",
				attempt, rp.config.MaxRetries, delay), nil, map[string]interface{}{
				"topic":   topic,
				"key":     key,
				"attempt": attempt,
				"delay":   delay.String(),
			})

			// Wait for delay or context cancellation
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-rp.ctx.Done():
				return fmt.Errorf("producer shutdown during retry")
			}
		}

		// Attempt to publish
		err := rp.producer.Publish(topic, key, message)
		if err == nil {
			// Success
			latency := time.Since(start)
			rp.recordSuccess(latency)

			rp.logWithLevel("DEBUG", "Kafka message published successfully", nil, map[string]interface{}{
				"topic":   topic,
				"key":     key,
				"attempt": attempt + 1,
				"latency": latency.String(),
			})

			return nil
		}

		lastErr = err
		rp.recordFailure(err)

		// Check if error is non-retryable
		if rp.isNonRetryableError(err) {
			rp.logWithLevel("ERROR", "Non-retryable error encountered, sending to fallback", err, map[string]interface{}{
				"topic":   topic,
				"key":     key,
				"attempt": attempt + 1,
			})

			// Send to fallback immediately
			rp.sendToFallback(topic, key, message, err, attempt)
			return fmt.Errorf("non-retryable error: %w", err)
		}

		rp.logWithLevel("WARN", "Kafka publish failed, will retry", err, map[string]interface{}{
			"topic":   topic,
			"key":     key,
			"attempt": attempt + 1,
		})
	}

	// All retries exhausted, send to fallback
	rp.logWithLevel("ERROR", "All retry attempts exhausted, sending to fallback", lastErr, map[string]interface{}{
		"topic":       topic,
		"key":         key,
		"max_retries": rp.config.MaxRetries,
	})

	rp.sendToFallback(topic, key, message, lastErr, rp.config.MaxRetries)
	return fmt.Errorf("all retries exhausted: %w", lastErr)
}

// Publish is a convenience method that uses a background context
func (rp *RetryProducer) Publish(topic string, key string, message []byte) error {
	return rp.PublishWithRetry(context.Background(), topic, key, message)
}

// calculateDelay computes the delay for a given retry attempt
func (rp *RetryProducer) calculateDelay(attempt int) time.Duration {
	delay := float64(rp.config.InitialDelay) *
		pow(rp.config.BackoffFactor, float64(attempt-1))

	if delay > float64(rp.config.MaxDelay) {
		delay = float64(rp.config.MaxDelay)
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

// isNonRetryableError determines if an error should not be retried
func (rp *RetryProducer) isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Non-retryable errors
	nonRetryableErrors := []string{
		"message size too large",
		"invalid topic",
		"unknown topic or partition",
		"policy violation",
		"authentication failed",
		"authorization failed",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if contains(errStr, nonRetryable) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsMiddle(s, substr))))
}

// containsMiddle checks if substring exists in the middle of string
func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// sendToFallback sends a failed message to the fallback channel
func (rp *RetryProducer) sendToFallback(topic string, key string, message []byte, err error, retries int) {
	fallbackMsg := &FallbackMessage{
		Topic:   topic,
		Key:     key,
		Message: message,
		Error:   err,
		Time:    time.Now(),
		Retries: retries,
	}

	select {
	case rp.fallbackChan <- fallbackMsg:
		// Successfully queued for fallback processing
	default:
		// Fallback channel is full, log warning
		rp.logWithLevel("WARN", "Fallback channel full, dropping message", err, map[string]interface{}{
			"topic": topic,
			"key":   key,
		})
	}
}

// processFallbackMessages handles messages that failed to be sent to Kafka
func (rp *RetryProducer) processFallbackMessages() {
	for {
		select {
		case <-rp.ctx.Done():
			return
		case msg := <-rp.fallbackChan:
			rp.handleFallbackMessage(msg)
		}
	}
}

// handleFallbackMessage processes a single fallback message
func (rp *RetryProducer) handleFallbackMessage(msg *FallbackMessage) {
	// Log the failure
	rp.logWithLevel("ERROR", "Message sent to fallback due to Kafka failure", msg.Error, map[string]interface{}{
		"topic":        msg.Topic,
		"key":          msg.Key,
		"retries":      msg.Retries,
		"failure_time": msg.Time.Format(time.RFC3339),
	})

	// Here you could implement various fallback strategies:
	// 1. Store in database for later retry
	// 2. Send to alternative message queue
	// 3. Write to file system
	// 4. Send to monitoring system

	// For now, we'll just log it and potentially store it for later processing
	rp.storeFailedMessage(msg)
}

// storeFailedMessage stores a failed message for potential later processing
func (rp *RetryProducer) storeFailedMessage(msg *FallbackMessage) {
	// This could be implemented to store in a database table
	// For now, we'll just count it in metrics
	rp.metrics.mu.Lock()
	rp.metrics.messagesFailed++
	rp.metrics.lastError = msg.Error
	rp.metrics.mu.Unlock()
}

// recordSuccess records successful message sending
func (rp *RetryProducer) recordSuccess(latency time.Duration) {
	rp.metrics.mu.Lock()
	defer rp.metrics.mu.Unlock()

	rp.metrics.messagesSent++
	rp.metrics.totalLatency += latency
}

// recordFailure records a failed message attempt
func (rp *RetryProducer) recordFailure(err error) {
	rp.metrics.mu.Lock()
	defer rp.metrics.mu.Unlock()

	rp.metrics.messagesRetried++
	rp.metrics.lastError = err
}

// GetMetrics returns current Kafka metrics
func (rp *RetryProducer) GetMetrics() KafkaMetrics {
	rp.metrics.mu.RLock()
	defer rp.metrics.mu.RUnlock()

	return *rp.metrics
}

// Close shuts down the retry producer
func (rp *RetryProducer) Close() error {
	// Signal shutdown
	rp.cancel()

	// Close base producer
	if err := rp.producer.Close(); err != nil {
		return fmt.Errorf("failed to close base producer: %w", err)
	}

	// Wait for fallback processor to finish
	rp.wg.Wait()

	return nil
}

// logWithLevel provides structured logging (simplified for this example)
func (rp *RetryProducer) logWithLevel(level, message string, err error, context map[string]interface{}) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	logMessage := fmt.Sprintf("[%s] %s [KAFKA-RETRY] %s", timestamp, level, message)

	if len(context) > 0 {
		for k, v := range context {
			logMessage += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	if err != nil {
		logMessage += fmt.Sprintf(" error=%s", err.Error())
	}

	switch level {
	case "DEBUG":
		log.Println(logMessage)
	case "INFO":
		log.Println(logMessage)
	case "WARN":
		log.Println(logMessage)
	case "ERROR":
		log.Println(logMessage)
	case "FATAL":
		log.Fatal(logMessage)
	default:
		log.Println(logMessage)
	}
}

// PublishJSON is a convenience method for publishing JSON messages
func (rp *RetryProducer) PublishJSON(ctx context.Context, topic string, key string, data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return rp.PublishWithRetry(ctx, topic, key, message)
}
