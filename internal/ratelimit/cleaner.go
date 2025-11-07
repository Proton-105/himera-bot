package ratelimit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cleaner periodically scans rate-limit keys and removes stale entries.
type Cleaner struct {
	redisClient *redis.Client
	log         *slog.Logger
	interval    time.Duration
}

// NewCleaner constructs a Cleaner instance.
func NewCleaner(client *redis.Client, log *slog.Logger, interval time.Duration) *Cleaner {
	if log == nil {
		log = slog.Default()
	}

	return &Cleaner{
		redisClient: client,
		log:         log,
		interval:    interval,
	}
}

// Run starts the cleaner loop until the context is cancelled.
func (c *Cleaner) Run(ctx context.Context) {
	if c.redisClient == nil || c.interval <= 0 {
		return
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if c.log != nil {
				c.log.Info("rate limit cleaner stopped", slog.String("reason", ctx.Err().Error()))
			}
			return
		case <-ticker.C:
			c.cleanup(ctx)
		}
	}
}

func (c *Cleaner) cleanup(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	const pattern = "ratelimit:*"
	const scanCount = 100

	cutoff := time.Now().Add(-5 * time.Minute).Unix()
	var cursor uint64
	cleaned := 0

	for {
		keys, nextCursor, err := c.redisClient.Scan(ctx, cursor, pattern, scanCount).Result()
		if err != nil {
			if c.log != nil {
				c.log.Error("rate limit scan failed", slog.Any("error", err))
			}
			return
		}

		for _, key := range keys {
			pipe := c.redisClient.TxPipeline()
			pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("(%d", cutoff))
			cardCmd := pipe.ZCard(ctx, key)
			if _, err := pipe.Exec(ctx); err != nil {
				if c.log != nil {
					c.log.Warn("cleanup pipeline failed", slog.String("key", key), slog.Any("error", err))
				}
				continue
			}

			count, err := cardCmd.Result()
			if err != nil {
				if c.log != nil {
					c.log.Warn("failed to read zset cardinality", slog.String("key", key), slog.Any("error", err))
				}
				continue
			}

			if count == 0 {
				if err := c.redisClient.Del(ctx, key).Err(); err != nil {
					if c.log != nil {
						c.log.Warn("failed to delete empty rate limit key", slog.String("key", key), slog.Any("error", err))
					}
					continue
				}
				cleaned++
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	if cleaned > 0 && c.log != nil {
		c.log.Info("rate limit keys cleaned", slog.Int("keys_removed", cleaned))
	}
}
