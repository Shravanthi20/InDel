package config

import (
	"os"
)

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	KafkaBrokers   string
	JWTSecret      string
	FirebaseKey    string
	RazorpayKey    string
	RazorpaySecret string
	InDelEnv       string
	LogLevel       string
	PremiumMLURL   string
	FraudMLURL     string
	ForecastMLURL  string
}

func Load() *Config {
	return &Config{
		DBHost:         envOrDefault("DB_HOST", "127.0.0.1"),
		DBPort:         envOrDefault("DB_PORT", "5432"),
		DBUser:         envOrDefault("DB_USER", "indel"),
		DBPassword:     envOrDefault("DB_PASSWORD", "password"),
		DBName:         envOrDefault("DB_NAME", "indel"),
		KafkaBrokers:   os.Getenv("KAFKA_BROKERS"),
		JWTSecret:      envOrDefault("JWT_SECRET", "indel-dev-secret"),
		FirebaseKey:    os.Getenv("FIREBASE_PROJECT_ID"),
		RazorpayKey:    os.Getenv("RAZORPAY_KEY_ID"),
		RazorpaySecret: os.Getenv("RAZORPAY_KEY_SECRET"),
		InDelEnv:       envOrDefault("INDEL_ENV", "development"),
		LogLevel:       envOrDefault("LOG_LEVEL", "info"),
		PremiumMLURL:   os.Getenv("PREMIUM_ML_URL"),
		FraudMLURL:     os.Getenv("FRAUD_ML_URL"),
		ForecastMLURL:  os.Getenv("FORECAST_ML_URL"),
	}
}
