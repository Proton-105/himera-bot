package ratelimit

import (
	"context"
	"errors"
	"time"
)

// Result captures the outcome of a rate-limit evaluation.
type Result struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

// Limiter describes a rate-limiting strategy interface.
type Limiter interface {
	Check(ctx context.Context, key string, limit int, window time.Duration) (*Result, error)
}

// ErrLimitExceeded indicates the rate limit has been reached for the key.
var ErrLimitExceeded = errors.New("rate limit exceeded")
