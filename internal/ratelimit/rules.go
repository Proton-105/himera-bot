package ratelimit

import (
	"errors"
	"time"

	"github.com/Proton-105/himera-bot/pkg/config"
)

// Rules encapsulates configured rate limits and helper methods.
type Rules struct {
	config config.RateLimitConfig
}

// NewRules constructs rate limiting rules from configuration settings.
func NewRules(cfg config.RateLimitConfig) *Rules {
	return &Rules{config: cfg}
}

// IsWhitelisted returns true if the userID bypasses rate limits.
func (r *Rules) IsWhitelisted(userID int64) bool {
	for _, id := range r.config.Whitelist {
		if id == userID {
			return true
		}
	}
	return false
}

// GetCommandLimit returns the limit and window for a specific command.
func (r *Rules) GetCommandLimit(command string) (int, time.Duration, error) {
	switch command {
	case "buy":
		return parseRule(r.config.Commands.Buy)
	case "sell":
		return parseRule(r.config.Commands.Sell)
	case "portfolio":
		return parseRule(r.config.Commands.Portfolio)
	default:
		return 0, 0, errors.New("unsupported command")
	}
}

// GetGlobalLimit returns the global rate limiting rule.
func (r *Rules) GetGlobalLimit() (int, time.Duration, error) {
	return parseRule(r.config.Global)
}

// GetPerUserLimit returns the per-user rate limiting rule.
func (r *Rules) GetPerUserLimit() (int, time.Duration, error) {
	return parseRule(r.config.PerUser)
}

func parseRule(rule config.RateLimitRule) (int, time.Duration, error) {
	if rule.Window == "" {
		return rule.Limit, 0, errors.New("window duration is not set")
	}
	window, err := time.ParseDuration(rule.Window)
	if err != nil {
		return 0, 0, err
	}
	return rule.Limit, window, nil
}
