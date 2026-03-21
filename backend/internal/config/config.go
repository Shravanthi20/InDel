package config

import (
	"os"
)

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
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		KafkaBrokers:   os.Getenv("KAFKA_BROKERS"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		FirebaseKey:    os.Getenv("FIREBASE_PROJECT_ID"),
		RazorpayKey:    os.Getenv("RAZORPAY_KEY_ID"),
		RazorpaySecret: os.Getenv("RAZORPAY_KEY_SECRET"),
		InDelEnv:       os.Getenv("INDEL_ENV"),
		LogLevel:       os.Getenv("LOG_LEVEL"),
		PremiumMLURL:   os.Getenv("PREMIUM_ML_URL"),
		FraudMLURL:     os.Getenv("FRAUD_ML_URL"),
		ForecastMLURL:  os.Getenv("FORECAST_ML_URL"),
	}
}
