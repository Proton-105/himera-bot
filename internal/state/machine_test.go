package state

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

var errStorageFailure = errors.New("storage error")

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) GetState(ctx context.Context, userID int64) (*UserState, error) {
	args := m.Called(ctx, userID)
	state, _ := args.Get(0).(*UserState)
	return state, args.Error(1)
}

func (m *mockStorage) SetState(ctx context.Context, userID int64, state *UserState) error {
	args := m.Called(ctx, userID, state)
	return args.Error(0)
}

func (m *mockStorage) ClearState(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestStateMachine_TransitionTo(t *testing.T) {
	ctx := context.Background()
	userID := int64(42)
	log := testLogger()

	testCases := []struct {
		name        string
		setupMocks  func(ms *mockStorage)
		newState    State
		expectedErr error
	}{
		{
			name: "successful transition",
			setupMocks: func(ms *mockStorage) {
				ms.On("GetState", mock.Anything, userID).
					Return(&UserState{CurrentState: StateIdle}, nil).Once()
				ms.On("SetState", mock.Anything, userID, mock.MatchedBy(func(state *UserState) bool {
					return state.CurrentState == StateBuyingSearch
				})).Return(nil).Once()
			},
			newState:    StateBuyingSearch,
			expectedErr: nil,
		},
		{
			name: "invalid transition",
			setupMocks: func(ms *mockStorage) {
				ms.On("GetState", mock.Anything, userID).
					Return(&UserState{CurrentState: StateIdle}, nil).Once()
			},
			newState:    StateBuyingConfirm,
			expectedErr: ErrInvalidTransition,
		},
		{
			name: "new user transition",
			setupMocks: func(ms *mockStorage) {
				ms.On("GetState", mock.Anything, userID).
					Return((*UserState)(nil), ErrStateNotFound).Once()
				ms.On("SetState", mock.Anything, userID, mock.MatchedBy(func(state *UserState) bool {
					return state.CurrentState == StateBuyingSearch
				})).Return(nil).Once()
			},
			newState:    StateBuyingSearch,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ms := &mockStorage{}
			tc.setupMocks(ms)

			fsm := NewStateMachine(ms, log, nil)
			err := fsm.TransitionTo(ctx, userID, tc.newState)

			if tc.expectedErr != nil {
				if err == nil || err != tc.expectedErr {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			ms.AssertExpectations(t)
		})
	}
}

func TestStateMachine_GetState(t *testing.T) {
	ctx := context.Background()
	userID := int64(7)
	log := testLogger()

	testCases := []struct {
		name        string
		setupMocks  func(ms *mockStorage)
		expectState *UserState
		expectErr   error
	}{
		{
			name: "state found",
			setupMocks: func(ms *mockStorage) {
				ms.On("GetState", mock.Anything, userID).
					Return(&UserState{UserID: userID, CurrentState: StateBuyingAmount}, nil).Once()
			},
			expectState: &UserState{UserID: userID, CurrentState: StateBuyingAmount},
			expectErr:   nil,
		},
		{
			name: "state not found",
			setupMocks: func(ms *mockStorage) {
				ms.On("GetState", mock.Anything, userID).
					Return((*UserState)(nil), ErrStateNotFound).Once()
			},
			expectState: nil,
			expectErr:   ErrStateNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ms := &mockStorage{}
			tc.setupMocks(ms)
			fsm := NewStateMachine(ms, log, nil)

			state, err := fsm.GetState(ctx, userID)

			if tc.expectErr != nil {
				if err == nil || err != tc.expectErr {
					t.Fatalf("expected error %v, got %v", tc.expectErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if tc.expectState != nil && state != nil {
				if tc.expectState.UserID != state.UserID || tc.expectState.CurrentState != state.CurrentState {
					t.Fatalf("unexpected state: %+v", state)
				}
			} else if tc.expectState != state {
				t.Fatalf("expected state %+v, got %+v", tc.expectState, state)
			}

			ms.AssertExpectations(t)
		})
	}
}

func TestStateMachine_SetState(t *testing.T) {
	ctx := context.Background()
	userID := int64(11)
	log := testLogger()

	testCases := []struct {
		name       string
		setupMocks func(ms *mockStorage)
		expectErr  error
	}{
		{
			name: "set state success",
			setupMocks: func(ms *mockStorage) {
				ms.On("SetState", mock.Anything, userID, mock.MatchedBy(func(userState *UserState) bool {
					return userState.CurrentState == StateBuyingConfirm
				})).Return(nil).Once()
			},
			expectErr: nil,
		},
		{
			name: "set state error",
			setupMocks: func(ms *mockStorage) {
				ms.On("SetState", mock.Anything, userID, mock.Anything).
					Return(errStorageFailure).Once()
			},
			expectErr: errStorageFailure,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ms := &mockStorage{}
			tc.setupMocks(ms)

			fsm := NewStateMachine(ms, log, nil)
			err := fsm.SetState(ctx, userID, StateBuyingConfirm, nil)

			if tc.expectErr != nil {
				if err == nil || err != tc.expectErr {
					t.Fatalf("expected error %v, got %v", tc.expectErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			ms.AssertExpectations(t)
		})
	}
}

func TestStateMachine_ClearState(t *testing.T) {
	ctx := context.Background()
	userID := int64(13)
	log := testLogger()

	testCases := []struct {
		name       string
		setupMocks func(ms *mockStorage)
		expectErr  error
	}{
		{
			name: "clear state success",
			setupMocks: func(ms *mockStorage) {
				ms.On("ClearState", mock.Anything, userID).
					Return(nil).Once()
			},
			expectErr: nil,
		},
		{
			name: "clear state error",
			setupMocks: func(ms *mockStorage) {
				ms.On("ClearState", mock.Anything, userID).
					Return(errStorageFailure).Once()
			},
			expectErr: errStorageFailure,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ms := &mockStorage{}
			tc.setupMocks(ms)

			fsm := NewStateMachine(ms, log, nil)
			err := fsm.ClearState(ctx, userID)

			if tc.expectErr != nil {
				if err == nil || err != tc.expectErr {
					t.Fatalf("expected error %v, got %v", tc.expectErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			ms.AssertExpectations(t)
		})
	}
}

func TestStateMachine_Lock(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	t.Cleanup(cleanup)

	storage := newInMemoryStorage(100 * time.Millisecond)
	fsm := NewStateMachine(storage, testLogger(), client)

	ctx := context.Background()
	userID := int64(77)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- fsm.SetState(ctx, userID, StateBuyingSearch, nil)
		}()
	}

	wg.Wait()
	close(errCh)

	var success, locked int
	for err := range errCh {
		if err == nil {
			success++
			continue
		}

		if errors.Is(err, ErrStateLocked) {
			locked++
			continue
		}

		t.Fatalf("unexpected error: %v", err)
	}

	if success != 1 {
		t.Fatalf("expected 1 successful transition, got %d", success)
	}
	if locked != 1 {
		t.Fatalf("expected 1 locked transition, got %d", locked)
	}
}

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		_ = client.Close()
		mr.Close()
	}

	return client, cleanup
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type inMemoryStorage struct {
	mu     sync.Mutex
	states map[int64]*UserState
	delay  time.Duration
}

func newInMemoryStorage(delay time.Duration) *inMemoryStorage {
	return &inMemoryStorage{
		states: make(map[int64]*UserState),
		delay:  delay,
	}
}

func (s *inMemoryStorage) GetState(ctx context.Context, userID int64) (*UserState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return nil, ErrStateNotFound
	}

	return cloneState(state), nil
}

func (s *inMemoryStorage) SetState(ctx context.Context, userID int64, state *UserState) error {
	time.Sleep(s.delay)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[userID] = cloneState(state)
	return nil
}

func (s *inMemoryStorage) ClearState(ctx context.Context, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, userID)
	return nil
}

func cloneState(state *UserState) *UserState {
	if state == nil {
		return nil
	}

	copyState := *state
	if state.Context != nil {
		ctxCopy := make(map[string]interface{}, len(state.Context))
		for k, v := range state.Context {
			ctxCopy[k] = v
		}
		copyState.Context = ctxCopy
	}
	return &copyState
}
