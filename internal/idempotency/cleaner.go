package idempotency

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cleaner struct {
	client   *redis.Client
	log      *slog.Logger
	interval time.Duration
}

func NewCleaner(client *redis.Client, log *slog.Logger, interval time.Duration) *Cleaner {
	if log == nil {
		log = slog.Default()
	}

	return &Cleaner{
		client:   client,
		log:      log,
		interval: interval,
	}
}

func (c *Cleaner) Run(ctx context.Context) {
	if c == nil || c.client == nil {
		return
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.cleanup(ctx)
		}
	}
}

func (c *Cleaner) cleanup(ctx context.Context) {
	var (
		cursor uint64
		err    error
	)

	for {
		var keys []string
		keys, cursor, err = c.client.Scan(ctx, cursor, "idempotency:*", 100).Result()
		if err != nil {
			c.log.Error("idempotency cleaner scan failed", slog.Any("error", err))
			return
		}

		for _, key := range keys {
			ttl, err := c.client.TTL(ctx, key).Result()
			if err != nil {
				c.log.Warn("failed to get key ttl", slog.String("key", key), slog.Any("error", err))
				continue
			}

			if ttl < 0 || ttl > 25*time.Hour {
				if err := c.client.Del(ctx, key).Err(); err != nil {
					c.log.Warn("failed to delete stale idempotency key", slog.String("key", key), slog.Any("error", err))
				}
			}
		}

		if cursor == 0 {
			break
		}
	}
}
