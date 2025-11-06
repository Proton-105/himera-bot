// Package repository implements Redis-backed storage for bot state.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Proton-105/himera-bot/internal/state"
	appredis "github.com/Proton-105/himera-bot/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
)

const userStateKeyPattern = "user_state:%d"

// StateRepository persists user state in Redis.
type StateRepository struct {
	client *appredis.Client
}

// NewStateRepository creates a Redis-backed implementation of state.Storage.
func NewStateRepository(client *appredis.Client) *StateRepository {
	return &StateRepository{client: client}
}

// GetState retrieves user's state from Redis.
func (r *StateRepository) GetState(ctx context.Context, userID int64) (*state.UserState, error) {
	key := fmt.Sprintf(userStateKeyPattern, userID)

	value, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get state from redis: %w", err)
	}

	var userState state.UserState
	if err := json.Unmarshal([]byte(value), &userState); err != nil {
		return nil, fmt.Errorf("unmarshal user state: %w", err)
	}

	return &userState, nil
}

// SetState stores user's state in Redis with a TTL.
func (r *StateRepository) SetState(ctx context.Context, userID int64, userState *state.UserState, ttl time.Duration) error {
	key := fmt.Sprintf(userStateKeyPattern, userID)

	payload, err := json.Marshal(userState)
	if err != nil {
		return fmt.Errorf("marshal user state: %w", err)
	}

	if err := r.client.Set(ctx, key, payload, ttl); err != nil {
		return fmt.Errorf("set state to redis: %w", err)
	}

	return nil
}
