package clients

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	name               string
	maxFailures        int
	timeout            time.Duration
	resetTimeout       time.Duration
	state              CircuitBreakerState
	failures           int
	lastFailureTime    time.Time
	mu                 sync.RWMutex
	halfOpenMaxCalls   int
	halfOpenCalls      int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, timeout, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		maxFailures:      maxFailures,
		timeout:          timeout,
		resetTimeout:     resetTimeout,
		state:            StateClosed,
		halfOpenMaxCalls: 3, // Allow 3 test calls in half-open state
	}
}

// Execute runs the operation with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should attempt to reset
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.resetTimeout {
		cb.state = StateHalfOpen
		cb.halfOpenCalls = 0
		log.Printf("[CIRCUIT-BREAKER] %s transitioning from OPEN to HALF-OPEN", cb.name)
	}

	// Reject calls if circuit is open
	if cb.state == StateOpen {
		return fmt.Errorf("circuit breaker %s is OPEN", cb.name)
	}

	// Limit calls in half-open state
	if cb.state == StateHalfOpen && cb.halfOpenCalls >= cb.halfOpenMaxCalls {
		return fmt.Errorf("circuit breaker %s HALF-OPEN call limit exceeded", cb.name)
	}

	// Execute the operation with timeout
	if cb.state == StateHalfOpen {
		cb.halfOpenCalls++
	}

	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()

	select {
	case err := <-done:
		if err != nil {
			cb.onFailure()
			return fmt.Errorf("operation failed: %w", err)
		}
		cb.onSuccess()
		return nil
	case <-time.After(cb.timeout):
		cb.onFailure()
		return fmt.Errorf("operation timeout after %v", cb.timeout)
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	}
}

// onSuccess handles successful operations
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		log.Printf("[CIRCUIT-BREAKER] %s transitioning from HALF-OPEN to CLOSED", cb.name)
	}
}

// onFailure handles failed operations
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		log.Printf("[CIRCUIT-BREAKER] %s transitioning from HALF-OPEN to OPEN", cb.name)
	} else if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		log.Printf("[CIRCUIT-BREAKER] %s transitioning from CLOSED to OPEN after %d failures", cb.name, cb.failures)
	}
}

// GetState returns the current state for monitoring
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return map[string]interface{}{
		"name":        cb.name,
		"state":       cb.state,
		"failures":    cb.failures,
		"last_failure": cb.lastFailureTime,
	}
}
