package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// BootstrapServiceEnv loads a service-specific .env file for local runs and
// validates required variables. In production, it uses process env vars only.
func BootstrapServiceEnv(service string) error {
	loadServiceEnv(service)
	if err := ValidateServiceEnv(service); err != nil {
		return err
	}
	return nil
}

func loadServiceEnv(service string) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_ENV")), "production") {
		log.Printf("[%s] production env detected, skipping .env file loading", service)
		return
	}

	if explicit := strings.TrimSpace(os.Getenv("ENV_FILE")); explicit != "" {
		if _, err := os.Stat(explicit); err == nil {
			if err := godotenv.Load(explicit); err == nil {
				log.Printf("[%s] loaded env file: %s", service, explicit)
				return
			}
		}
	}

	candidates := []string{
		filepath.Join("cmd", service, ".env.local"),
		filepath.Join("cmd", service, ".env"),
	}

	if cwd, err := os.Getwd(); err == nil {
		if filepath.Base(cwd) == service {
			candidates = append(candidates, ".env.local", ".env")
		}
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			if err := godotenv.Load(candidate); err == nil {
				log.Printf("[%s] loaded env file: %s", service, candidate)
				return
			}
		}
	}

	log.Printf("[%s] no local .env file loaded, using process environment", service)
}

func ValidateServiceEnv(service string) error {
	missing := make([]string, 0)
	invalid := make([]string, 0)

	require := func(key string) {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	requireAny := func(keys ...string) {
		for _, key := range keys {
			if strings.TrimSpace(os.Getenv(key)) != "" {
				return
			}
		}
		missing = append(missing, strings.Join(keys, "|"))
	}

	require("PORT")
	require("INDEL_ENV")
	require("INDEL_ALLOWED_ORIGINS")
	require("REDIS_URL")
	requireAny("DB_URL", "DATABASE_URL", "DB_HOST")

	switch service {
	case "worker-gateway":
		require("ML_SERVICE_URL")
		require("CORE_URL")
	case "insurer-gateway", "platform-gateway":
		require("CORE_URL")
	}

	if portRaw := strings.TrimSpace(os.Getenv("PORT")); portRaw != "" {
		port, err := strconv.Atoi(portRaw)
		if err != nil || port <= 0 || port > 65535 {
			invalid = append(invalid, "PORT must be a valid TCP port (1-65535)")
		}
	}

	validateURL := func(key string) {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			return
		}
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			invalid = append(invalid, fmt.Sprintf("%s must be a valid absolute URL", key))
		}
	}
	validateURL("CORE_URL")
	validateURL("ML_SERVICE_URL")

	validatePositiveInt := func(key string) {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			return
		}
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			invalid = append(invalid, fmt.Sprintf("%s must be a non-negative integer", key))
		}
	}
	validatePositiveInt("ML_TIMEOUT_MS")
	validatePositiveInt("ML_RETRY_COUNT")

	if len(missing) > 0 {
		return fmt.Errorf("[%s] missing required environment variables: %s", service, strings.Join(missing, ", "))
	}
	if len(invalid) > 0 {
		return fmt.Errorf("[%s] invalid environment configuration: %s", service, strings.Join(invalid, "; "))
	}
	return nil
}
