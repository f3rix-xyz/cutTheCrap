package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	OpenRouterKey  string
	MaxConcurrent  int
	RequestTimeout time.Duration
	ChunkSize      int
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		OpenRouterKey:  getEnv("OPENROUTER_API_KEY", ""),
		MaxConcurrent:  getEnvAsInt("MAX_CONCURRENT", 10),
		RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 30*time.Second),
		ChunkSize:      getEnvAsInt("CHUNK_SIZE", 900),
	}
}

// getEnv retrieves environment variable or returns default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt retrieves environment variable as integer or returns default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsDuration retrieves environment variable as duration or returns default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
