package errors

import (
	"context"
	"errors"
	"math"
	"time"
)

const (
	MaxRetries        = 3
	InitialBackoff    = 100 * time.Millisecond
	MaxBackoff        = 5 * time.Second
	BackoffMultiplier = 2.0
)

func WithRetry(ctx context.Context, fn func() error) error {
	if fn == nil {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	var err error
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = fn()
		if err == nil {
			return nil
		}

		if !IsRetryable(err) {
			return err
		}

		if attempt == MaxRetries {
			return err
		}

		backoff := calculateBackoffDuration(attempt + 1)
		time.Sleep(backoff)
	}

	return err
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var appErr *AppError
	if errors.As(err, &appErr) && appErr != nil {
		return appErr.Retryable
	}

	return false
}

func calculateBackoffDuration(attempt int) time.Duration {
	delay := float64(InitialBackoff) * math.Pow(BackoffMultiplier, float64(attempt))
	backoff := time.Duration(delay)
	if backoff > MaxBackoff {
		return MaxBackoff
	}

	return backoff
}
