package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisStorage_SetAndGet(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	t.Cleanup(cleanup)

	log := testLogger()
	storage := NewRedisStorage(client, log)

	ctx := context.Background()
	userState := &UserState{
		UserID:       123,
		CurrentState: StateBuyingSearch,
		Context: map[string]interface{}{
			"foo": "bar",
		},
	}

	err := storage.SetState(ctx, userState.UserID, userState)
	assert.NoError(t, err)

	result, err := storage.GetState(ctx, userState.UserID)
	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, userState.UserID, result.UserID)
		assert.Equal(t, userState.CurrentState, result.CurrentState)
		assert.Equal(t, userState.Context, result.Context)
	}
}

func TestRedisStorage_GetNotFound(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	t.Cleanup(cleanup)

	log := testLogger()
	storage := NewRedisStorage(client, log)

	ctx := context.Background()

	state, err := storage.GetState(ctx, 999)
	assert.Nil(t, state)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrStateNotFound)
}

func TestRedisStorage_ClearState(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	t.Cleanup(cleanup)

	log := testLogger()
	storage := NewRedisStorage(client, log)

	ctx := context.Background()
	userState := &UserState{
		UserID:       456,
		CurrentState: StateBuyingAmount,
		Context:      map[string]interface{}{"amount": 10},
	}

	err := storage.SetState(ctx, userState.UserID, userState)
	assert.NoError(t, err)

	err = storage.ClearState(ctx, userState.UserID)
	assert.NoError(t, err)

	state, err := storage.GetState(ctx, userState.UserID)
	assert.Nil(t, state)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrStateNotFound)
}
