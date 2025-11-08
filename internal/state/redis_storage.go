package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	userStateKeyPattern  = "user:state:%d"
	userStateScanPattern = "user:state:*"
)

// RedisStorage persists user FSM states in Redis.
type RedisStorage struct {
	client *redis.Client
	log    *slog.Logger
}

// NewRedisStorage initializes a Redis-backed Storage implementation.
func NewRedisStorage(client *redis.Client, log *slog.Logger) Storage {
	if log == nil {
		log = slog.Default()
	}

	return &RedisStorage{
		client: client,
		log:    log,
	}
}

// GetState returns the stored user state or ErrStateNotFound when absent.
func (s *RedisStorage) GetState(ctx context.Context, userID int64) (*UserState, error) {
	key := redisUserStateKey(userID)

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrStateNotFound
		}

		s.log.Error("failed to get state from redis", "user_id", userID, "error", err)
		return nil, err
	}

	var state UserState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		s.log.Error("failed to decode user state", "user_id", userID, "error", err)
		return nil, err
	}

	return &state, nil
}

// SetState saves the provided user state with a one-hour TTL.
func (s *RedisStorage) SetState(ctx context.Context, userID int64, state *UserState) error {
	state.UpdatedAt = time.Now().UTC()

	data, err := json.Marshal(state)
	if err != nil {
		s.log.Error("failed to encode user state", "user_id", userID, "error", err)
		return err
	}

	key := redisUserStateKey(userID)
	if err := s.client.Set(ctx, key, data, time.Hour).Err(); err != nil {
		s.log.Error("failed to save state in redis", "user_id", userID, "error", err)
		return err
	}

	return nil
}

// ClearState removes the stored state for the given user.
func (s *RedisStorage) ClearState(ctx context.Context, userID int64) error {
	key := redisUserStateKey(userID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		s.log.Error("failed to clear user state", "user_id", userID, "error", err)
		return err
	}

	return nil
}

// GetAllStates retrieves every stored user state by scanning Redis keys.
func (s *RedisStorage) GetAllStates(ctx context.Context) ([]*UserState, error) {
	var (
		cursor uint64
		result []*UserState
	)

	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, userStateScanPattern, 100).Result()
		if err != nil {
			s.log.Error("failed to scan user states", "error", err)
			return nil, err
		}

		for _, key := range keys {
			data, err := s.client.Get(ctx, key).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					continue
				}

				s.log.Error("failed to fetch user state", "key", key, "error", err)
				return nil, err
			}

			var userState UserState
			if err := json.Unmarshal([]byte(data), &userState); err != nil {
				s.log.Error("failed to decode user state", "key", key, "error", err)
				continue
			}

			copied := userState
			result = append(result, &copied)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return result, nil
}

func redisUserStateKey(userID int64) string {
	return fmt.Sprintf(userStateKeyPattern, userID)
}
