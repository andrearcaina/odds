package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr            string
	DraftKingsURL   string
	RefreshInterval time.Duration
	HTTPTimeout     time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Addr:            getEnv("PORT", ":8080"),
		DraftKingsURL:   getEnv("DRAFTKINGS_URL", ""),
		RefreshInterval: durationEnv("REFRESH_INTERVAL", 15*time.Second),
		HTTPTimeout:     durationEnv("HTTP_TIMEOUT", 10*time.Second),
	}

	if cfg.DraftKingsURL == "" {
		return Config{}, fmt.Errorf("DRAFTKINGS_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		if key == "PORT" && value[0] != ':' {
			return ":" + value
		}

		return value
	}

	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(value)

	if err != nil || seconds < 1 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}
