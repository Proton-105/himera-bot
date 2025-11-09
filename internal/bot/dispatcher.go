package bot

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	telebot "gopkg.in/telebot.v3"

	"github.com/Proton-105/himera-bot/internal/bot/handlers"
	"github.com/Proton-105/himera-bot/internal/state"
)

// Dispatcher routes incoming updates to state-specific handlers.
type Dispatcher struct {
	fsm           state.StateMachine
	stateHandlers map[state.State]handlers.Handler
	log           *slog.Logger
	mu            sync.RWMutex
}

// NewDispatcher creates a Dispatcher with an empty handlers registry.
func NewDispatcher(fsm state.StateMachine, log *slog.Logger) *Dispatcher {
	if log == nil {
		log = slog.Default()
	}

	return &Dispatcher{
		fsm:           fsm,
		stateHandlers: make(map[state.State]handlers.Handler),
		log:           log,
	}
}

// RegisterStateHandler registers a handler for the provided state.
func (d *Dispatcher) RegisterStateHandler(s state.State, h handlers.Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stateHandlers[s] = h
}

// Dispatch routes the update based on the user's current state.
func (d *Dispatcher) Dispatch(c telebot.Context) error {
	if c == nil || c.Sender() == nil {
		d.log.Warn("cannot dispatch without sender information")
		return nil
	}

	ctx := context.Background()
	userID := c.Sender().ID

	currentState := state.StateIdle
	userState, err := d.fsm.GetState(ctx, userID)
	if err != nil {
		if !errors.Is(err, state.ErrStateNotFound) {
			return err
		}
	} else if userState != nil {
		currentState = userState.CurrentState
	}

	handler := d.getHandler(currentState)
	if handler == nil {
		d.log.Info("no handler registered for state", "state", currentState, "user_id", userID)
		return nil
	}

	return handler(c)
}

func (d *Dispatcher) getHandler(s state.State) handlers.Handler {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.stateHandlers[s]
}
