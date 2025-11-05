// Package state manages user state and state machine data for the bot.
package state

import (
	"context"
	"time"
)

// UserState describes current user operation context.
type UserState struct {
	CurrentAction string
	Payload       []byte
}

// Storage defines contract for user state persistence.
type Storage interface {
	GetState(ctx context.Context, userID int64) (*UserState, error)
	SetState(ctx context.Context, userID int64, state *UserState, ttl time.Duration) error
}
