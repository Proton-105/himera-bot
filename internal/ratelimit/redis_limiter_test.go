package ratelimit

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		_ = client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestRedisLimiter_AllowsWithinLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	t.Cleanup(cleanup)

	limiter := NewRedisLimiter(client, testLogger())
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		result, err := limiter.Check(ctx, "test:allows", 5, time.Minute)
		assert.NoError(t, err)
		assert.True(t, result.Allowed)
	}
}

func TestRedisLimiter_BlocksWhenExceeded(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	t.Cleanup(cleanup)

	limiter := NewRedisLimiter(client, testLogger())
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		result, err := limiter.Check(ctx, "test:blocks", 2, time.Minute)
		assert.NoError(t, err)
		if i < 2 {
			assert.True(t, result.Allowed)
		} else {
			assert.False(t, result.Allowed)
		}
	}
}

func TestRedisLimiter_SlidingWindow(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	t.Cleanup(cleanup)

	limiter := NewRedisLimiter(client, testLogger())
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		result, err := limiter.Check(ctx, "test:window", 2, time.Second)
		assert.NoError(t, err)
		assert.True(t, result.Allowed)
	}

	time.Sleep(1100 * time.Millisecond)

	result, err := limiter.Check(ctx, "test:window", 2, time.Second)
	assert.NoError(t, err)
	assert.True(t, result.Allowed)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
