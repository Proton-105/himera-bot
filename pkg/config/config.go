package config

import "os"

// Config holds runtime configuration for the Himera trading bot.
type Config struct {
	Env      string
	LogLevel string
	HTTPPort string
}

// Load reads configuration from environment variables and applies defaults where needed.
func Load() *Config {
	cfg := &Config{
		Env:      getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}

	return defaultValue
}
