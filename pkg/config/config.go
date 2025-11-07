package config

import (
	"fmt"
	"strings"
	"time"

	redisclient "github.com/Proton-105/himera-bot/pkg/redis"
)

// Config aggregates application configuration settings.
type Config struct {
	AppEnv    string          `mapstructure:"app_env" yaml:"-" validate:"required"`
	Server    ServerConfig    `mapstructure:"server" yaml:"server" validate:"required"`
	Bot       BotConfig       `mapstructure:"bot" yaml:"bot" validate:"required"`
	Database  DatabaseConfig  `mapstructure:"database" yaml:"database" validate:"required"`
	Redis     RedisConfig     `mapstructure:"redis" yaml:"redis" validate:"required"`
	API       APIConfig       `mapstructure:"api" yaml:"api" validate:"required"`
	Logger    LoggerConfig    `mapstructure:"logging" yaml:"logging" validate:"required"`
	Sentry    SentryConfig    `mapstructure:"sentry" yaml:"sentry" validate:"required"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit" yaml:"ratelimit"`
}

// String returns a masked representation of the configuration.
func (c Config) String() string {
	return fmt.Sprintf(
		"Config{AppEnv:%s, Server:%s, Bot:%s, Database:%s, Redis:%s, API:%s, Logger:%s, Sentry:%s, RateLimit:%s}",
		c.AppEnv,
		c.Server.String(),
		c.Bot.String(),
		c.Database.String(),
		c.Redis.String(),
		c.API.String(),
		c.Logger.String(),
		fmt.Sprintf("Sentry{DSN:%s, Enabled:%t}", maskSecret(c.Sentry.DSN), c.Sentry.Enabled),
		c.RateLimit.String(),
	)
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Port         string        `mapstructure:"port" yaml:"port" validate:"required"`
	MetricsPort  string        `mapstructure:"metrics_port" yaml:"metrics_port" validate:"required"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" validate:"required"`
}

func (s ServerConfig) String() string {
	return fmt.Sprintf("Server{Port:%s, MetricsPort:%s, ReadTimeout:%s, WriteTimeout:%s}", s.Port, s.MetricsPort, s.ReadTimeout, s.WriteTimeout)
}

// BotConfig contains bot-related settings.
type BotConfig struct {
	Token      string        `mapstructure:"token" yaml:"token" validate:"required"`
	Timeout    time.Duration `mapstructure:"timeout" yaml:"timeout" validate:"required"`
	Mode       string        `mapstructure:"mode" yaml:"mode" validate:"required"`
	WebhookURL string        `mapstructure:"webhook_url" yaml:"webhook_url"`
}

func (b BotConfig) String() string {
	return fmt.Sprintf(
		"Bot{Token:%s, Timeout:%s, Mode:%s, WebhookURL:%s}",
		maskSecret(b.Token),
		b.Timeout,
		b.Mode,
		b.WebhookURL,
	)
}

// DatabaseConfig contains PostgreSQL configuration.
type DatabaseConfig struct {
	Host     string `mapstructure:"host" yaml:"host" validate:"required"`
	Port     string `mapstructure:"port" yaml:"port" validate:"required"`
	User     string `mapstructure:"user" yaml:"user" validate:"required"`
	Password string `mapstructure:"password" yaml:"password" validate:"required"`
	Name     string `mapstructure:"name" yaml:"name" validate:"required"`
	SSLMode  string `mapstructure:"ssl_mode" yaml:"ssl_mode" validate:"required"`
}

func (d DatabaseConfig) String() string {
	return fmt.Sprintf("Database{Host:%s, Port:%s, User:%s, Password:%s, Name:%s, SSLMode:%s}", d.Host, d.Port, d.User, maskSecret(d.Password), d.Name, d.SSLMode)
}

// DSN returns the formatted Postgres connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

// RedisConfig contains Redis connection settings.
type RedisConfig struct {
	Host            string        `mapstructure:"host" yaml:"host" validate:"required"`
	Port            string        `mapstructure:"port" yaml:"port" validate:"required"`
	Password        string        `mapstructure:"password" yaml:"password"`
	DB              int           `mapstructure:"db" yaml:"db"`
	PoolSize        int           `mapstructure:"pool_size" yaml:"pool_size" validate:"required"`
	MinIdleConns    int           `mapstructure:"min_idle_conns" yaml:"min_idle_conns" validate:"required"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout" yaml:"pool_timeout" validate:"required"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout" validate:"required"`
	MaxRetries      int           `mapstructure:"max_retries" yaml:"max_retries" validate:"required"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff" yaml:"min_retry_backoff" validate:"required"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff" yaml:"max_retry_backoff" validate:"required"`
}

func (r RedisConfig) String() string {
	return fmt.Sprintf("Redis{Host:%s, Port:%s, Password:%s, DB:%d}", r.Host, r.Port, maskSecret(r.Password), r.DB)
}

// ToClientConfig converts to the redis client configuration.
func (r RedisConfig) ToClientConfig() redisclient.Config {
	return redisclient.Config{
		Addr:            fmt.Sprintf("%s:%s", r.Host, r.Port),
		Password:        r.Password,
		DB:              r.DB,
		PoolSize:        r.PoolSize,
		MinIdleConns:    r.MinIdleConns,
		PoolTimeout:     r.PoolTimeout,
		IdleTimeout:     r.IdleTimeout,
		MaxRetries:      r.MaxRetries,
		MinRetryBackoff: r.MinRetryBackoff,
		MaxRetryBackoff: r.MaxRetryBackoff,
	}
}

// APIConfig contains external API endpoints and timeouts.
type APIConfig struct {
	DexScreenerURL string        `mapstructure:"dex_screener_url" yaml:"dex_screener_url" validate:"required,url"`
	CoinGeckoURL   string        `mapstructure:"coin_gecko_url" yaml:"coin_gecko_url" validate:"required,url"`
	Timeout        time.Duration `mapstructure:"timeout" yaml:"timeout" validate:"required"`
}

func (a APIConfig) String() string {
	return fmt.Sprintf("API{DexScreenerURL:%s, CoinGeckoURL:%s, Timeout:%s}", a.DexScreenerURL, a.CoinGeckoURL, a.Timeout)
}

// LoggerConfig contains logging settings.
type LoggerConfig struct {
	Level  string `mapstructure:"level" yaml:"level" validate:"required"`
	Format string `mapstructure:"format" yaml:"format" validate:"required"`
}

func (l LoggerConfig) String() string {
	return fmt.Sprintf("Logger{Level:%s, Format:%s}", l.Level, l.Format)
}

// SentryConfig holds Sentry integration settings.
type SentryConfig struct {
	DSN     string `mapstructure:"dsn" yaml:"dsn"`
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
}

// RateLimitRule represents a single rate limit entry.
type RateLimitRule struct {
	Limit  int    `mapstructure:"limit" yaml:"limit"`
	Window string `mapstructure:"window" yaml:"window"`
}

// RateLimitCommandsConfig groups per-command limits.
type RateLimitCommandsConfig struct {
	Buy       RateLimitRule `mapstructure:"buy" yaml:"buy"`
	Sell      RateLimitRule `mapstructure:"sell" yaml:"sell"`
	Portfolio RateLimitRule `mapstructure:"portfolio" yaml:"portfolio"`
}

// RateLimitConfig aggregates rate-limiter settings.
type RateLimitConfig struct {
	Global    RateLimitRule           `mapstructure:"global" yaml:"global"`
	PerUser   RateLimitRule           `mapstructure:"per_user" yaml:"per_user"`
	Commands  RateLimitCommandsConfig `mapstructure:"commands" yaml:"commands"`
	Whitelist []int64                 `mapstructure:"whitelist" yaml:"whitelist"`
}

// String returns a compact string representation of rate limit settings.
func (r RateLimitConfig) String() string {
	return fmt.Sprintf(
		"RateLimit{Global:{Limit:%d Window:%s}, PerUser:{Limit:%d Window:%s}, Commands:{Buy:%d/%s Sell:%d/%s Portfolio:%d/%s}, Whitelist:%d}",
		r.Global.Limit, r.Global.Window,
		r.PerUser.Limit, r.PerUser.Window,
		r.Commands.Buy.Limit, r.Commands.Buy.Window,
		r.Commands.Sell.Limit, r.Commands.Sell.Window,
		r.Commands.Portfolio.Limit, r.Commands.Portfolio.Window,
		len(r.Whitelist),
	)
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 2 {
		return strings.Repeat("*", len(value))
	}
	return fmt.Sprintf("%s***%s", string(value[0]), string(value[len(value)-1]))
}
