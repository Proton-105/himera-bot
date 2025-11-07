package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements Limiter using Redis sorted sets and a sliding window.
type RedisLimiter struct {
	client *redis.Client
	log    *slog.Logger
}

var _ Limiter = (*RedisLimiter)(nil)

// NewRedisLimiter creates a Redis-backed Limiter implementation.
func NewRedisLimiter(client *redis.Client, log *slog.Logger) Limiter {
	if log == nil {
		log = slog.Default()
	}

	return &RedisLimiter{
		client: client,
		log:    log,
	}
}

// Check evaluates the rate limit for a given key using a sliding window algorithm.
func (l *RedisLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (*Result, error) {
	if l.client == nil {
		return nil, errors.New("redis client is not configured for rate limiting")
	}

	if limit <= 0 {
		return &Result{Allowed: false, Remaining: 0, ResetAt: time.Now().Add(window)}, nil
	}

	now := time.Now()
	windowStart := now.Add(-window)
	redisKey := "ratelimit:" + key

	cutoff := float64(windowStart.UnixNano()) / float64(time.Millisecond)
	score := float64(now.UnixNano()) / float64(time.Millisecond)

	pipe := l.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "-inf", fmt.Sprintf("(%f", cutoff))
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  score,
		Member: uuid.NewString(),
	})
	countCmd := pipe.ZCard(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window*2)

	if _, err := pipe.Exec(ctx); err != nil {
		if l.log != nil {
			l.log.Error("rate limiter pipeline failed", slog.String("key", key), slog.Any("error", err))
		}
		return nil, err
	}

	count, err := countCmd.Result()
	if err != nil {
		if l.log != nil {
			l.log.Error("rate limiter failed to read count", slog.String("key", key), slog.Any("error", err))
		}
		return nil, err
	}

	allowed := count <= int64(limit)
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	result := &Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   windowStart.Add(window),
	}

	return result, nil
}
