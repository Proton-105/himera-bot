package usercache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/Proton-105/himera-bot/internal/domain"
)

// Cache provides Redis-backed caching for user profiles.
type Cache struct {
	client *redis.Client
}

// NewCache constructs a user cache backed by the provided Redis client.
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// Get fetches a cached user profile if it exists.
func (c *Cache) Get(ctx context.Context, userID int64) (*domain.User, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}

	data, err := c.client.Get(ctx, cacheKey(userID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get cached user: %w", err)
	}

	var user domain.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("decode cached user: %w", err)
	}

	return &user, nil
}

// Set stores the user profile in cache for the provided TTL.
func (c *Cache) Set(ctx context.Context, userID int64, user *domain.User, ttl time.Duration) error {
	if c == nil || c.client == nil || user == nil {
		return nil
	}

	payload, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("encode user for cache: %w", err)
	}

	if err := c.client.Set(ctx, cacheKey(userID), payload, ttl).Err(); err != nil {
		return fmt.Errorf("set cached user: %w", err)
	}

	return nil
}

// Invalidate removes the cached profile entry if it exists.
func (c *Cache) Invalidate(ctx context.Context, userID int64) error {
	if c == nil || c.client == nil {
		return nil
	}

	if err := c.client.Del(ctx, cacheKey(userID)).Err(); err != nil {
		return fmt.Errorf("delete cached user: %w", err)
	}

	return nil
}

func cacheKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}
