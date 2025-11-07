package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	stateKeyPattern     = "user:state:*"
	stateScanBatchCount = 100
)

// Cleaner removes expired FSM states from Redis on a schedule.
type Cleaner struct {
	redisClient *redis.Client
	storage     Storage
	log         *slog.Logger
	ttl         time.Duration
	interval    time.Duration
}

// NewCleaner constructs a Cleaner instance.
func NewCleaner(redisClient *redis.Client, storage Storage, log *slog.Logger, ttl, interval time.Duration) *Cleaner {
	if log == nil {
		log = slog.Default()
	}

	return &Cleaner{
		redisClient: redisClient,
		storage:     storage,
		log:         log,
		ttl:         ttl,
		interval:    interval,
	}
}

// Run starts the cleanup loop until the context is cancelled.
func (c *Cleaner) Run(ctx context.Context) {
	if c == nil || c.redisClient == nil || c.storage == nil {
		return
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if c.log != nil {
				reason := ctx.Err()
				if reason != nil {
					c.log.Info("state cleaner stopped", slog.String("reason", reason.Error()))
				} else {
					c.log.Info("state cleaner stopped")
				}
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

	var cursor uint64
	for {
		keys, nextCursor, err := c.redisClient.Scan(ctx, cursor, stateKeyPattern, stateScanBatchCount).Result()
		if err != nil {
			if c.log != nil {
				c.log.Error("state cleaner scan failed", slog.Any("error", err))
			}
			return
		}

		for _, key := range keys {
			userID, err := extractUserID(key)
			if err != nil {
				if c.log != nil {
					c.log.Warn("state cleaner unable to parse user id", slog.String("key", key), slog.Any("error", err))
				}
				continue
			}

			state, err := c.storage.GetState(ctx, userID)
			if err != nil {
				if !errors.Is(err, ErrStateNotFound) && c.log != nil {
					c.log.Error("state cleaner failed to load state", slog.Int64("user_id", userID), slog.Any("error", err))
				}
				continue
			}

			if state == nil {
				continue
			}

			if time.Since(state.UpdatedAt) > c.ttl {
				if err := c.storage.ClearState(ctx, userID); err != nil {
					if c.log != nil {
						c.log.Error("state cleaner failed to clear state", slog.Int64("user_id", userID), slog.Any("error", err))
					}
					continue
				}
				if c.log != nil {
					c.log.Info("state session cleared", slog.Int64("user_id", userID))
				}
			}
		}

		if ctx.Err() != nil || nextCursor == 0 {
			return
		}
		cursor = nextCursor
	}
}

func extractUserID(key string) (int64, error) {
	segments := strings.Split(key, ":")
	if len(segments) == 0 {
		return 0, fmt.Errorf("invalid key format: %s", key)
	}

	return strconv.ParseInt(segments[len(segments)-1], 10, 64)
}
