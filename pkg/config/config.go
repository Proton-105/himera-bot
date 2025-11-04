package config

import (
	"fmt"
	"os"
)

// Config holds runtime configuration for the Himera trading bot.
type Config struct {
	Env      string
	LogLevel string
	HTTPPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load reads configuration from environment variables and applies defaults where needed.
func Load() *Config {
	cfg := &Config{
		Env:      getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getRequiredEnv("DB_USER"),
		DBPassword: getRequiredEnv("DB_PASSWORD"),
		DBName:     getRequiredEnv("DB_NAME"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	return cfg
}

// GetDBConnectionString returns PostgreSQL DSN based on config values.
func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}

	return defaultValue
}

func getRequiredEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}

	panic(fmt.Sprintf("environment variable %s is required", key))
}
