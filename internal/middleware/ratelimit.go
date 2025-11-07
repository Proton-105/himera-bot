package middleware

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Proton-105/himera-bot/internal/ratelimit"
	"gopkg.in/telebot.v3"
)

// RateLimitMiddleware enforces per-user rate limits for incoming Telegram updates.
type RateLimitMiddleware struct {
	limiter ratelimit.Limiter
	rules   *ratelimit.Rules
	log     *slog.Logger
}

// NewRateLimitMiddleware constructs a rate-limit middleware component.
func NewRateLimitMiddleware(limiter ratelimit.Limiter, rules *ratelimit.Rules, log *slog.Logger) *RateLimitMiddleware {
	if log == nil {
		log = slog.Default()
	}

	return &RateLimitMiddleware{
		limiter: limiter,
		rules:   rules,
		log:     log,
	}
}

// Handle returns a telebot middleware that enforces per-user rate limits.
func (m *RateLimitMiddleware) Handle(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if m.limiter == nil || m.rules == nil {
			return next(c)
		}

		sender := c.Sender()
		if sender == nil {
			return next(c)
		}

		userID := sender.ID
		if m.rules.IsWhitelisted(userID) {
			return next(c)
		}

		limit, window, err := m.rules.GetPerUserLimit()
		if err != nil {
			if m.log != nil {
				m.log.Error("failed to load per-user rate limit", slog.Int64("user_id", userID), slog.Any("error", err))
			}
			return next(c)
		}

		key := fmt.Sprintf("user:%d", userID)
		result, err := m.limiter.Check(context.Background(), key, limit, window)
		if err != nil {
			if m.log != nil {
				m.log.Warn("rate limiter error", slog.Int64("user_id", userID), slog.Any("error", err))
			}
			return next(c)
		}

		if !result.Allowed {
			if m.log != nil {
				m.log.Warn("rate limit exceeded", slog.Int64("user_id", userID))
			}
			return c.Send("Rate limit exceeded. Try again later.")
		}

		return next(c)
	}
}
