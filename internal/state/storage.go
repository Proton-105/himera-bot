// Package state manages user state and state machine data for the bot.
package state

import "context"

// Storage defines the persistence contract for user FSM state.
type Storage interface {
	// GetState returns the current state for the specified user.
	GetState(ctx context.Context, userID int64) (*UserState, error)
	// SetState saves the provided state for the specified user.
	SetState(ctx context.Context, userID int64, state *UserState) error
	// ClearState removes the state for the specified user.
	ClearState(ctx context.Context, userID int64) error
}
