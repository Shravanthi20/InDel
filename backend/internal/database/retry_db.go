package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RetryConfig defines retry behavior for database operations
type RetryConfig struct {
	MaxRetries    int           // Maximum number of retry attempts
	InitialDelay  time.Duration // Initial delay between retries
	MaxDelay      time.Duration // Maximum delay between retries
	BackoffFactor float64       // Multiplier for exponential backoff
}

// DefaultRetryConfig provides sensible defaults for database operations
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  50 * time.Millisecond,
		MaxDelay:      2 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryDB wraps a GORM database instance with retry capabilities
type RetryDB struct {
	db     *gorm.DB
	config RetryConfig
}

// NewRetryDB creates a new database wrapper with retry capabilities
func NewRetryDB(db *gorm.DB, config RetryConfig) *RetryDB {
	return &RetryDB{
		db:     db,
		config: config,
	}
}

// WithRetry executes a database operation with retry logic for transient failures
func (rdb *RetryDB) WithRetry(ctx context.Context, operation func(*gorm.DB) error) error {
	var lastErr error

	for attempt := 0; attempt <= rdb.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := rdb.calculateDelay(attempt)
			
			// Wait for delay or context cancellation
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during database retry: %w", ctx.Err())
			}
		}

		// Attempt the operation
		err := operation(rdb.db)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is transient and retryable
		if !rdb.isRetryableError(err) {
			return fmt.Errorf("non-retryable database error: %w", err)
		}

		// Log retry attempt (in production, use structured logging)
		if attempt < rdb.config.MaxRetries {
			fmt.Printf("[DB-RETRY] Attempt %d/%d failed, retrying in %v: %v\n", 
				attempt+1, rdb.config.MaxRetries+1, rdb.calculateDelay(attempt+1), err)
		}
	}

	return fmt.Errorf("database operation failed after %d attempts: %w", rdb.config.MaxRetries+1, lastErr)
}

// WithTransaction executes multiple operations within a database transaction with retry
func (rdb *RetryDB) WithTransaction(ctx context.Context, operations func(*gorm.DB) error) error {
	return rdb.WithRetry(ctx, func(db *gorm.DB) error {
		return db.Transaction(func(tx *gorm.DB) error {
			return operations(tx)
		})
	})
}

// calculateDelay computes the delay for a given retry attempt
func (rdb *RetryDB) calculateDelay(attempt int) time.Duration {
	delay := float64(rdb.config.InitialDelay) * 
		pow(rdb.config.BackoffFactor, float64(attempt-1))
	
	if delay > float64(rdb.config.MaxDelay) {
		delay = float64(rdb.config.MaxDelay)
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

// isRetryableError determines if a database error should be retried
func (rdb *RetryDB) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	
	// Transient errors that should be retried
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"deadlock",
		"lock wait timeout",
		"connection lost",
		"server has gone away",
		"temporary failure",
		"resource temporarily unavailable",
		"network is unreachable",
		"no route to host",
		"connection timed out",
		"read-only transaction",
		"could not serialize access",
		"could not obtain lock",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	// Check for specific GORM error codes that are retryable
	// This would need to be expanded based on your database type
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

// GetDB returns the underlying GORM database instance
func (rdb *RetryDB) GetDB() *gorm.DB {
	return rdb.db
}

// ExecWithRetry executes a raw SQL query with retry logic
func (rdb *RetryDB) ExecWithRetry(ctx context.Context, sql string, values ...interface{}) error {
	return rdb.WithRetry(ctx, func(db *gorm.DB) error {
		return db.Exec(sql, values...).Error
	})
}

// FirstWithRetry executes a First query with retry logic
func (rdb *RetryDB) FirstWithRetry(ctx context.Context, dest interface{}, conds ...interface{}) error {
	return rdb.WithRetry(ctx, func(db *gorm.DB) error {
		return db.First(dest, conds...).Error
	})
}

// CreateWithRetry executes a Create operation with retry logic
func (rdb *RetryDB) CreateWithRetry(ctx context.Context, value interface{}) error {
	return rdb.WithRetry(ctx, func(db *gorm.DB) error {
		return db.Create(value).Error
	})
}

// UpdateWithRetry executes an Update operation with retry logic
func (rdb *RetryDB) UpdateWithRetry(ctx context.Context, model interface{}, updates interface{}) error {
	return rdb.WithRetry(ctx, func(db *gorm.DB) error {
		return db.Model(model).Updates(updates).Error
	})
}

// DeleteWithRetry executes a Delete operation with retry logic
func (rdb *RetryDB) DeleteWithRetry(ctx context.Context, value interface{}) error {
	return rdb.WithRetry(ctx, func(db *gorm.DB) error {
		return db.Delete(value).Error
	})
}

// RawWithRetry executes a raw SQL query with retry logic and scans the result
func (rdb *RetryDB) RawWithRetry(ctx context.Context, dest interface{}, sql string, values ...interface{}) error {
	return rdb.WithRetry(ctx, func(db *gorm.DB) error {
		return db.Raw(sql, values...).Scan(dest).Error
	})
}

// TransactionWithRetry executes a transaction with custom retry logic
// This is useful when you need to handle specific transaction errors differently
func (rdb *RetryDB) TransactionWithRetry(ctx context.Context, fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return rdb.WithRetry(ctx, func(db *gorm.DB) error {
		return db.Transaction(fn, opts...)
	})
}
