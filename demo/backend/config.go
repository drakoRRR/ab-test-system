package main

import (
	"errors"
	"os"
)

type Config struct {
	APIKey        string
	ServiceURL    string
	Port          string
	ExperimentKey string
	FlagKey       string
	StaticDir     string
}

func loadConfig() (Config, error) {
	cfg := Config{
		APIKey:        os.Getenv("SDK_API_KEY"),
		ServiceURL:    envOrDefault("SDK_SERVICE_URL", "http://localhost:8080/api/v1"),
		Port:          envOrDefault("PORT", "8081"),
		ExperimentKey: envOrDefault("EXPERIMENT_KEY", "checkout-btn"),
		FlagKey:       envOrDefault("FLAG_KEY", "new-checkout"),
		StaticDir:     envOrDefault("STATIC_DIR", "../static"),
	}
	if cfg.APIKey == "" {
		return Config{}, errors.New("SDK_API_KEY is required")
	}
	return cfg, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
