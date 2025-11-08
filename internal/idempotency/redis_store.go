package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
)

type Record struct {
	Status   string
	Response []byte
}

type Store interface {
	Lock(ctx context.Context, key string, lockTTL time.Duration) (bool, error)
	Get(ctx context.Context, key string) (*Record, error)
	Set(ctx context.Context, key string, record *Record, ttl time.Duration) error
	ReleaseLock(ctx context.Context, key string) error
}

type RedisStore struct {
	client *redis.Client
	log    *slog.Logger
}

func NewRedisStore(client *redis.Client, log *slog.Logger) Store {
	if log == nil {
		log = slog.Default()
	}

	return &RedisStore{
		client: client,
		log:    log,
	}
}

func (s *RedisStore) Lock(ctx context.Context, key string, lockTTL time.Duration) (bool, error) {
	lockKey := lockKey(key)
	acquired, err := s.client.SetNX(ctx, lockKey, 1, lockTTL).Result()
	if err != nil {
		s.log.Error("failed to acquire idempotency lock", slog.String("key", key), slog.Any("error", err))
		return false, err
	}

	return acquired, nil
}

func (s *RedisStore) Get(ctx context.Context, key string) (*Record, error) {
	result, err := s.client.HGetAll(ctx, recordKey(key)).Result()
	if err != nil {
		s.log.Error("failed to fetch idempotency record", slog.String("key", key), slog.Any("error", err))
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	responseData := []byte{}
	if encoded, ok := result["response"]; ok && encoded != "" {
		if err := json.Unmarshal([]byte(encoded), &responseData); err != nil {
			s.log.Error("failed to decode idempotency response", slog.String("key", key), slog.Any("error", err))
			return nil, err
		}
	}

	return &Record{
		Status:   result["status"],
		Response: responseData,
	}, nil
}

func (s *RedisStore) Set(ctx context.Context, key string, record *Record, ttl time.Duration) error {
	if record == nil {
		return nil
	}

	responseJSON, err := json.Marshal(record.Response)
	if err != nil {
		s.log.Error("failed to encode idempotency response", slog.String("key", key), slog.Any("error", err))
		return err
	}

	args := map[string]interface{}{
		"status":   record.Status,
		"response": string(responseJSON),
	}

	if err := s.client.HSet(ctx, recordKey(key), args).Err(); err != nil {
		s.log.Error("failed to store idempotency record", slog.String("key", key), slog.Any("error", err))
		return err
	}

	if err := s.client.Expire(ctx, recordKey(key), ttl).Err(); err != nil {
		s.log.Error("failed to set idempotency ttl", slog.String("key", key), slog.Any("error", err))
		return err
	}

	return nil
}

func (s *RedisStore) ReleaseLock(ctx context.Context, key string) error {
	if err := s.client.Del(ctx, lockKey(key)).Err(); err != nil {
		s.log.Error("failed to release idempotency lock", slog.String("key", key), slog.Any("error", err))
		return err
	}

	return nil
}

func recordKey(key string) string {
	return fmt.Sprintf("idempotency:%s", key)
}

func lockKey(key string) string {
	return fmt.Sprintf("idempotency:%s:lock", key)
}
