package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	userLockKeyPattern = "user:lock:%d"
	lockTTL            = 5 * time.Second
)

var (
	// ErrInvalidTransition indicates that a requested FSM transition is not allowed.
	ErrInvalidTransition = errors.New("invalid state transition")
	// ErrStateNotFound indicates that a user state record does not exist.
	ErrStateNotFound = errors.New("user state not found")
	// ErrStateLocked indicates that a concurrent operation already holds the lock.
	ErrStateLocked = errors.New("state is locked, try again later")
)

var transitionRecorder = func(from, to string) {}

// RegisterTransitionRecorder allows external packages to observe FSM transitions.
func RegisterTransitionRecorder(recorder func(from, to string)) {
	if recorder == nil {
		transitionRecorder = func(string, string) {}
		return
	}

	transitionRecorder = recorder
}

// StateMachine describes the operations supported by the FSM controller.
type StateMachine interface {
	GetState(ctx context.Context, userID int64) (*UserState, error)
	SetState(ctx context.Context, userID int64, state State, contextData map[string]interface{}) error
	TransitionTo(ctx context.Context, userID int64, newState State) error
	ClearState(ctx context.Context, userID int64) error
	GetAllStates(ctx context.Context) ([]*UserState, error)
}

// machine is a concrete implementation of StateMachine backed by Storage and Redis locking.
type machine struct {
	storage     Storage
	log         *slog.Logger
	redisClient *redis.Client
}

// NewStateMachine creates a FSM controller using the provided storage backend and redis client for locking.
func NewStateMachine(storage Storage, log *slog.Logger, redisClient *redis.Client) StateMachine {
	if log == nil {
		log = slog.Default()
	}

	return &machine{
		storage:     storage,
		log:         log,
		redisClient: redisClient,
	}
}

// GetState proxies to the underlying storage implementation.
func (m *machine) GetState(ctx context.Context, userID int64) (*UserState, error) {
	return m.storage.GetState(ctx, userID)
}

// GetAllStates returns every persisted user state.
func (m *machine) GetAllStates(ctx context.Context) ([]*UserState, error) {
	return m.storage.GetAllStates(ctx)
}

// SetState composes a UserState and persists it via storage under a distributed lock.
func (m *machine) SetState(ctx context.Context, userID int64, state State, contextData map[string]interface{}) error {
	if err := m.lock(ctx, userID); err != nil {
		return err
	}
	defer m.unlock(ctx, userID)

	return m.saveState(ctx, userID, state, contextData)
}

// TransitionTo changes the state if the transition is allowed, guarded by a lock.
func (m *machine) TransitionTo(ctx context.Context, userID int64, newState State) error {
	if err := m.lock(ctx, userID); err != nil {
		return err
	}
	defer m.unlock(ctx, userID)

	current := StateIdle

	storedState, err := m.storage.GetState(ctx, userID)
	if err != nil {
		if !errors.Is(err, ErrStateNotFound) {
			return err
		}
	} else if storedState != nil {
		current = storedState.CurrentState
	}

	if !IsTransitionAllowed(current, newState) {
		if m.log != nil {
			m.log.Warn("invalid state transition", "user_id", userID, "from", current, "to", newState)
		}
		return ErrInvalidTransition
	}

	transitionRecorder(string(current), string(newState))

	return m.saveState(ctx, userID, newState, nil)
}

// ClearState removes the stored state via the backing storage while holding the lock.
func (m *machine) ClearState(ctx context.Context, userID int64) error {
	if err := m.lock(ctx, userID); err != nil {
		return err
	}
	defer m.unlock(ctx, userID)

	return m.storage.ClearState(ctx, userID)
}

func (m *machine) saveState(ctx context.Context, userID int64, state State, contextData map[string]interface{}) error {
	userState := &UserState{
		UserID:       userID,
		CurrentState: state,
		Context:      contextData,
	}

	return m.storage.SetState(ctx, userID, userState)
}

func (m *machine) lock(ctx context.Context, userID int64) error {
	if m.redisClient == nil {
		if m.log != nil {
			m.log.Warn("redis client not configured for state locks; skipping", "user_id", userID)
		}
		return nil
	}

	key := fmt.Sprintf(userLockKeyPattern, userID)
	acquired, err := m.redisClient.SetNX(ctx, key, 1, lockTTL).Result()
	if err != nil {
		if m.log != nil {
			m.log.Error("failed to acquire user state lock", "user_id", userID, "error", err)
		}
		return err
	}

	if !acquired {
		if m.log != nil {
			m.log.Warn("user state lock already held", "user_id", userID)
		}
		return ErrStateLocked
	}

	return nil
}

func (m *machine) unlock(ctx context.Context, userID int64) {
	if m.redisClient == nil {
		return
	}

	key := fmt.Sprintf(userLockKeyPattern, userID)
	if err := m.redisClient.Del(ctx, key).Err(); err != nil && m.log != nil {
		m.log.Error("failed to release user state lock", "user_id", userID, "error", err)
	}
}
