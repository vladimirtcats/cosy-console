package config

import (
	"errors"
	"os"
)

type Config struct {
	AppPort     string
	DatabaseURL string
	RedisURL    string
}

func New() (Config, error) {
	databaseURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}

	redisURL, err := requireEnv("REDIS_URL")
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppPort:     envOr("APP_PORT", "8080"),
		DatabaseURL: databaseURL,
		RedisURL:    redisURL,
	}, nil
}

func requireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", errors.New("ENV " + key + " is required")
	}

	return value, nil
}

func envOr(key string, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	return fallback
}
