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
}

func Load() (Config, error) {
	cfg := Config{
		Port:            getEnv("PORT", "8080"),
		DataFile:        getEnv("DATA_FILE", "./data/checks.json"),
		AlertWebhookURL: strings.TrimSpace(getEnv("ALERT_WEBHOOK_URL", "")),
		RequestUA:       getEnv("REQUEST_USER_AGENT", "go-healthcheck-bot/1.0"),
	}

	if _, err := strconv.Atoi(cfg.Port); err != nil {
		return Config{}, fmt.Errorf("PORT must be numeric: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
