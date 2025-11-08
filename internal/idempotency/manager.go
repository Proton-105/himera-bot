package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"
)

var ErrRequestInProgress = errors.New("request with this key is already in progress")

type Operation func(ctx context.Context) (interface{}, error)

type Result struct {
	Response  interface{}
	FromCache bool
}

type Manager interface {
	Execute(
		ctx context.Context,
		key string,
		ttl time.Duration,
		fn Operation,
	) (*Result, error)
}

type manager struct {
	store Store
	log   *slog.Logger
}

func NewManager(store Store, log *slog.Logger) Manager {
	if log == nil {
		log = slog.Default()
	}

	return &manager{
		store: store,
		log:   log,
	}
}

func (m *manager) Execute(ctx context.Context, key string, ttl time.Duration, fn Operation) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if fn == nil {
		return nil, errors.New("operation fn cannot be nil")
	}

	for {
		locked, err := m.store.Lock(ctx, key, 5*time.Minute)
		if err != nil {
			return nil, err
		}

		if !locked {
			record, err := m.store.Get(ctx, key)
			if err != nil {
				return nil, err
			}

			if record == nil {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}

			switch record.Status {
			case StatusProcessing:
				return nil, ErrRequestInProgress
			case StatusCompleted:
				var response interface{}
				if len(record.Response) > 0 {
					if err := json.Unmarshal(record.Response, &response); err != nil {
						return nil, err
					}
				}
				return &Result{Response: response, FromCache: true}, nil
			default:
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}
		}

		defer m.store.ReleaseLock(ctx, key)

		result, err := fn(ctx)
		if err != nil {
			return nil, err
		}

		responseBytes, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		if err := m.store.Set(ctx, key, &Record{
			Status:   StatusCompleted,
			Response: responseBytes,
		}, ttl); err != nil {
			return nil, err
		}

		return &Result{
			Response:  result,
			FromCache: false,
		}, nil
	}
}
