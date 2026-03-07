package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port            string
	DataFile        string
	AlertWebhookURL string
	RequestUA       string

	// Optional API authentication. If empty, the API is unauthenticated.
	APIKey string

	// Network timeouts.
	HTTPTimeoutSec    int
	WebhookTimeoutSec int
}

func Load() (Config, error) {
	cfg := Config{
		Port:              getEnv("PORT", "8080"),
		DataFile:          getEnv("DATA_FILE", "./data/checks.json"),
		AlertWebhookURL:   strings.TrimSpace(getEnv("ALERT_WEBHOOK_URL", "")),
		RequestUA:         getEnv("REQUEST_USER_AGENT", "go-healthcheck-bot/1.0"),
		APIKey:            strings.TrimSpace(getEnv("API_KEY", "")),
		HTTPTimeoutSec:    getEnvInt("HTTP_TIMEOUT_SEC", 10),
		WebhookTimeoutSec: getEnvInt("WEBHOOK_TIMEOUT_SEC", 5),
	}

	if _, err := strconv.Atoi(cfg.Port); err != nil {
		return Config{}, fmt.Errorf("PORT must be numeric: %w", err)
	}
	if cfg.HTTPTimeoutSec < 1 {
		return Config{}, fmt.Errorf("HTTP_TIMEOUT_SEC must be >= 1")
	}
	if cfg.WebhookTimeoutSec < 1 {
		return Config{}, fmt.Errorf("WEBHOOK_TIMEOUT_SEC must be >= 1")
	}

	return cfg, nil
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
