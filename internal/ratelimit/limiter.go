package ratelimit

import (
	"context"
	"fmt"
	"time"

	appredis "github.com/himera-bot/trading-bot/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
)

// Limiter implements sliding window rate limiting on top of Redis Sorted Set.
type Limiter struct {
	client *appredis.Client
	limit  int
	window time.Duration
}

// NewLimiter configures limiter with provided redis client, limit and window.
func NewLimiter(client *appredis.Client, limit int, window time.Duration) *Limiter {
	return &Limiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Allow returns true when request is permitted within limit/window.
func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	if l.limit <= 0 {
		return false, fmt.Errorf("limiter limit must be positive")
	}

	now := time.Now()
	windowStart := now.Add(-l.window).UnixNano()
	score := float64(now.UnixNano())

	pipe := l.client.TxPipeline()

	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("(%d", windowStart))
	cardCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, goredis.Z{
		Score:  score,
		Member: score,
	})
	pipe.Expire(ctx, key, l.window)

	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("rate limiter exec: %w", err)
	}

	currentCount := cardCmd.Val()
	if currentCount >= int64(l.limit) {
		return false, nil
	}

	return true, nil
}
