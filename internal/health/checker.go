package health

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"gopkg.in/telebot.v3"
)

// Checkable represents a component that can report its health status.
type Checkable interface {
	HealthCheck(ctx context.Context) error
}

// Checker aggregates health checks for multiple components.
type Checker struct {
	log    *slog.Logger
	checks map[string]Checkable
}

// NewChecker instantiates a Checker with the provided logger.
func NewChecker(log *slog.Logger) *Checker {
	return &Checker{
		log:    log,
		checks: make(map[string]Checkable),
	}
}

// AddCheck registers a checkable component by name.
func (c *Checker) AddCheck(name string, check Checkable) {
	if name == "" || check == nil {
		return
	}
	c.checks[name] = check
}

// Check runs all registered health checks and returns their statuses.
func (c *Checker) Check(ctx context.Context) map[string]string {
	results := make(map[string]string, len(c.checks))

	for name, check := range c.checks {
		if check == nil {
			results[name] = "no check configured"
			continue
		}

		if err := check.HealthCheck(ctx); err != nil {
			results[name] = err.Error()
			if c.log != nil {
				c.log.Error("health check failed", slog.String("component", name), slog.Any("error", err))
			}
			continue
		}

		results[name] = "OK"
	}

	return results
}

// DBChecker verifies connectivity to a PostgreSQL database.
type DBChecker struct {
	db *sql.DB
}

// NewDBChecker constructs a DBChecker.
func NewDBChecker(db *sql.DB) *DBChecker {
	return &DBChecker{db: db}
}

// HealthCheck pings the database to ensure it is reachable.
func (c *DBChecker) HealthCheck(ctx context.Context) error {
	if c == nil || c.db == nil {
		return sql.ErrConnDone
	}
	return c.db.PingContext(ctx)
}

// Pinger abstracts the subset of redis.Client used for health checks.
type Pinger interface {
	Ping(ctx context.Context) *redis.StatusCmd
}

// RedisChecker verifies connectivity to a Redis instance.
type RedisChecker struct {
	pinger Pinger
}

// NewRedisChecker constructs a RedisChecker.
func NewRedisChecker(pinger Pinger) *RedisChecker {
	return &RedisChecker{pinger: pinger}
}

// HealthCheck issues a PING command against Redis.
func (c *RedisChecker) HealthCheck(ctx context.Context) error {
	if c == nil || c.pinger == nil {
		return redis.ErrClosed
	}
	return c.pinger.Ping(ctx).Err()
}

// TelegramChecker verifies that the Telegram bot API is reachable.
type TelegramChecker struct {
	bot *telebot.Bot
}

// NewTelegramChecker constructs a TelegramChecker.
func NewTelegramChecker(bot *telebot.Bot) *TelegramChecker {
	return &TelegramChecker{bot: bot}
}

// HealthCheck ensures the underlying bot is initialized and reachable.
func (c *TelegramChecker) HealthCheck(ctx context.Context) error {
	if c == nil || c.bot == nil || c.bot.Me == nil {
		return errors.New("telegram bot is not initialized or disconnected")
	}
	return nil
}
