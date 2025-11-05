// Package config provides configuration loading and validation utilities.
package config

import (
	"fmt"
	"os"
	"strings"

	validator "github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Load reads configuration from YAML files and environment variables, validates it, and returns the resulting Config.
func Load() (*Config, *viper.Viper, error) {
	if err := godotenv.Load(".env.local", ".env"); err != nil {
		// ignore missing env files in Phase 0
		_ = err
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	v := viper.New()
	v.SetConfigFile(fmt.Sprintf("./configs/%s.yaml", env))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.AppEnv = env

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return nil, nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, v, nil
}
