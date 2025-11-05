// Package redis provides a type-safe Redis client wrapper for the application.
package redis

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Config defines connection parameters for initializing the Redis client.
type Config struct {
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	DB              int           `yaml:"db"`
	PoolSize        int           `yaml:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns"`
	PoolTimeout     time.Duration `yaml:"pool_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	MaxRetries      int           `yaml:"max_retries"`
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`
}

// Client wraps the go-redis client to expose typed helper methods.
type Client struct {
	*redis.Client
}

// New creates a Redis client configured with cfg and verifies the connection with Ping.
func New(ctx context.Context, cfg Config) (*Client, error) {
	opts := &redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return &Client{rdb}, nil
}

// Get retrieves a value for the provided key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Set stores a value under key with the specified TTL.
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Client.Set(ctx, key, value, ttl).Err()
}

// Delete removes the specified key.
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

// TxPipeline creates a transactional pipeline.
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.Client.TxPipeline()
}

// Close shuts down the Redis client.
func (c *Client) Close() error {
	return c.Client.Close()
}
