package state

import "time"

// State represents a finite-state machine state.
type State string

const (
	// StateIdle indicates that the bot is waiting for the next user command.
	StateIdle State = "idle"
	// StateBuyingSearch indicates that the user is searching for a token to buy.
	StateBuyingSearch State = "buying_search"
	// StateBuyingAmount indicates that the user is entering the purchase amount.
	StateBuyingAmount State = "buying_amount"
	// StateBuyingConfirm indicates that the user is confirming the purchase.
	StateBuyingConfirm State = "buying_confirm"
	// StateError indicates that the bot is in an error state and requires recovery.
	StateError State = "error"
)

// UserState captures the current FSM state for a Telegram user.
type UserState struct {
	UserID       int64                  `json:"user_id"`
	CurrentState State                  `json:"current_state"`
	Context      map[string]interface{} `json:"context"`
	UpdatedAt    time.Time              `json:"updated_at"`
}
