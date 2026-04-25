package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type Config struct {
	DatabaseURL            string
	DBPreferSimpleProtocol bool
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPassword             string
	DBName                 string
	DBSSLMode              string
	KafkaBrokers           string
	JWTSecret              string
	FirebaseKey            string
	RazorpayKey            string
	RazorpaySecret         string
	InDelEnv               string
	LogLevel               string
	PremiumMLURL           string
	FraudMLURL             string
	FraudServiceURL        string
	ForecastMLURL          string
	MLServiceURL           string
	MLTimeoutMS            int
	MLRetryCount           int
	RedisURL               string
	// Policy lock-in configuration
	PolicyLockInHours           int // POLICY_LOCKIN_HOURS (default 48)
	DisruptionBlockLookaheadHrs int // DISRUPTION_BLOCK_LOOKAHEAD_HOURS (default 12)
}

func Load() *Config {
	cfg := &Config{}
	
	// Parse and validate all configuration
	if err := parseAndValidateConfig(cfg); err != nil {
		log.Fatalf("CONFIGURATION ERROR: %v", err)
	}
	
	return cfg
}

// parseAndValidateConfig handles all configuration parsing and validation
func parseAndValidateConfig(cfg *Config) error {
	var errors []string
	
	// Database configuration
	cfg.DatabaseURL = strings.TrimSpace(os.Getenv("DB_URL"))
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	
	cfg.DBPreferSimpleProtocol = strings.EqualFold(envOrDefault("DB_PREFER_SIMPLE_PROTOCOL", "false"), "true")
	cfg.DBHost = envOrDefault("DB_HOST", "127.0.0.1")
	cfg.DBPort = envOrDefault("DB_PORT", "5432")
	cfg.DBUser = envOrDefault("DB_USER", "indel")
	cfg.DBPassword = envOrDefault("DB_PASSWORD", "password")
	cfg.DBName = envOrDefault("DB_NAME", "indel")
	cfg.DBSSLMode = os.Getenv("DB_SSLMODE")
	
	// Production safety checks for database
	if strings.EqualFold(cfg.InDelEnv, "production") {
		if cfg.DatabaseURL == "" {
			if strings.Contains(cfg.DBHost, "localhost") || strings.Contains(cfg.DBHost, "127.0.0.1") {
				errors = append(errors, "localhost database connections not allowed in production")
			}
		} else {
			if strings.Contains(cfg.DatabaseURL, "localhost") || strings.Contains(cfg.DatabaseURL, "127.0.0.1") {
				errors = append(errors, "localhost database connections not allowed in production")
			}
		}
		
		// Require SSL in production
		if cfg.DBSSLMode == "" || cfg.DBSSLMode == "disable" {
			errors = append(errors, "DB_SSLMODE must be set to 'require' or 'verify-full' in production")
		}
		
		// Check for default passwords
		if cfg.DBPassword == "password" || cfg.DBPassword == "indel" {
			errors = append(errors, "default database password not allowed in production")
		}
	}
	
	// Validate database configuration
	if cfg.DatabaseURL == "" && (cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBPassword == "") {
		errors = append(errors, "Either DB_URL/DATABASE_URL or DB_HOST, DB_PORT, DB_USER, DB_PASSWORD must be provided")
	}
	
	// Validate port numbers
	if port, err := strconv.Atoi(cfg.DBPort); err != nil || port <= 0 || port > 65535 {
		errors = append(errors, "DB_PORT must be a valid TCP port (1-65535)")
	}
	
	// Kafka configuration
	cfg.KafkaBrokers = strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	
	// Security configuration
	cfg.JWTSecret = envOrDefault("JWT_SECRET", "indel-dev-secret")
	if len(cfg.JWTSecret) < 16 {
		errors = append(errors, "JWT_SECRET must be at least 16 characters long")
	}
	
	// Production safety checks for JWT secret
	if strings.EqualFold(cfg.InDelEnv, "production") {
		if cfg.JWTSecret == "indel-dev-secret" || cfg.JWTSecret == "indel-dev" || cfg.JWTSecret == "dev-secret" {
			errors = append(errors, "default JWT secret not allowed in production")
		}
	}
	
	cfg.FirebaseKey = os.Getenv("FIREBASE_PROJECT_ID")
	cfg.RazorpayKey = os.Getenv("RAZORPAY_KEY_ID")
	cfg.RazorpaySecret = os.Getenv("RAZORPAY_KEY_SECRET")
	
	// Environment configuration
	cfg.InDelEnv = envOrDefault("INDEL_ENV", "development")
	validEnvs := []string{"development", "staging", "production", "test"}
	envValid := false
	for _, env := range validEnvs {
		if strings.EqualFold(cfg.InDelEnv, env) {
			envValid = true
			break
		}
	}
	if !envValid {
		errors = append(errors, fmt.Sprintf("INDEL_ENV must be one of: %v", validEnvs))
	}
	
	cfg.LogLevel = envOrDefault("LOG_LEVEL", "info")
	validLogLevels := []string{"debug", "info", "warn", "error"}
	logLevelValid := false
	for _, level := range validLogLevels {
		if strings.EqualFold(cfg.LogLevel, level) {
			logLevelValid = true
			break
		}
	}
	if !logLevelValid {
		errors = append(errors, fmt.Sprintf("LOG_LEVEL must be one of: %v", validLogLevels))
	}
	
	// ML Service configuration
	mlServiceURL := strings.TrimSpace(os.Getenv("ML_SERVICE_URL"))
	if mlServiceURL == "" {
		mlServiceURL = strings.TrimSpace(os.Getenv("PREMIUM_ML_URL"))
	}
	if mlServiceURL == "" {
		mlServiceURL = strings.TrimSpace(os.Getenv("FRAUD_ML_URL"))
	}
	if mlServiceURL == "" {
		mlServiceURL = strings.TrimSpace(os.Getenv("FORECAST_ML_URL"))
	}
	if mlServiceURL == "" {
		// In production, ML service URL must be explicitly set
		if strings.EqualFold(cfg.InDelEnv, "production") {
			errors = append(errors, "ML_SERVICE_URL must be explicitly set in production")
		} else {
			// Only use localhost fallback in non-production environments
			mlServiceURL = "http://localhost:5000"
		}
	}
	
	// Validate ML service URL
	if err := validateURL(mlServiceURL); err != nil {
		errors = append(errors, fmt.Sprintf("ML service URL invalid: %v", err))
	}
	
	cfg.PremiumMLURL = os.Getenv("PREMIUM_ML_URL")
	cfg.FraudMLURL = os.Getenv("FRAUD_ML_URL")
	cfg.FraudServiceURL = os.Getenv("FRAUD_SERVICE_URL")
	cfg.ForecastMLURL = os.Getenv("FORECAST_ML_URL")
	cfg.MLServiceURL = mlServiceURL
	
	// ML timeout and retry configuration
	mlTimeoutMS := 3000
	if v := strings.TrimSpace(os.Getenv("ML_TIMEOUT_MS")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 30000 {
			mlTimeoutMS = parsed
		} else {
			errors = append(errors, "ML_TIMEOUT_MS must be between 1 and 30000 milliseconds")
		}
	}
	
	mlRetryCount := 2
	if v := strings.TrimSpace(os.Getenv("ML_RETRY_COUNT")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 && parsed <= 10 {
			mlRetryCount = parsed
		} else {
			errors = append(errors, "ML_RETRY_COUNT must be between 0 and 10")
		}
	}
	
	cfg.MLTimeoutMS = mlTimeoutMS
	cfg.MLRetryCount = mlRetryCount
	
	// Redis configuration
	cfg.RedisURL = envOrDefault("REDIS_URL", "localhost:6379")
	
	// Production safety checks for Redis
	if strings.EqualFold(cfg.InDelEnv, "production") {
		if strings.Contains(cfg.RedisURL, "localhost") || strings.Contains(cfg.RedisURL, "127.0.0.1") {
			errors = append(errors, "localhost Redis connections not allowed in production")
		}
	}
	
	// Policy lock-in configuration
	lockInHours := 48
	if v := strings.TrimSpace(os.Getenv("POLICY_LOCKIN_HOURS")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 168 { // Max 1 week
			lockInHours = parsed
		} else {
			errors = append(errors, "POLICY_LOCKIN_HOURS must be between 1 and 168 hours")
		}
	}
	
	lookaheadHrs := 12
	if v := strings.TrimSpace(os.Getenv("DISRUPTION_BLOCK_LOOKAHEAD_HOURS")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 168 { // Max 1 week
			lookaheadHrs = parsed
		} else {
			errors = append(errors, "DISRUPTION_BLOCK_LOOKAHEAD_HOURS must be between 1 and 168 hours")
		}
	}
	
	cfg.PolicyLockInHours = lockInHours
	cfg.DisruptionBlockLookaheadHrs = lookaheadHrs
	
	// Fail fast if there are configuration errors
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n- %s", strings.Join(errors, "\n- "))
	}
	
	// Log successful configuration load (excluding secrets)
	log.Printf("[CONFIG] Successfully loaded configuration for environment: %s", cfg.InDelEnv)
	log.Printf("[CONFIG] ML Service: %s (timeout: %dms, retries: %d)", cfg.MLServiceURL, cfg.MLTimeoutMS, cfg.MLRetryCount)
	log.Printf("[CONFIG] Policy Lock-in: %d hours, Disruption Lookahead: %d hours", cfg.PolicyLockInHours, cfg.DisruptionBlockLookaheadHrs)
	
	return nil
}

// validateURL validates that a URL string is properly formatted
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	// Basic URL format check
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	
	// Check for localhost in production (deployment safety)
	if strings.EqualFold(os.Getenv("INDEL_ENV"), "production") {
		if strings.Contains(urlStr, "localhost") || strings.Contains(urlStr, "127.0.0.1") {
			return fmt.Errorf("localhost URLs not allowed in production")
		}
	}
	
	return nil
}
