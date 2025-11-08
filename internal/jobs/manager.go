package jobs

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// Manager describes the minimal queue operations needed by the application.
type Manager interface {
	Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
	Close() error
}

type manager struct {
	client *asynq.Client
	log    *slog.Logger
}

// NewManager builds a Manager backed by an asynq client.
func NewManager(redisOpt asynq.RedisConnOpt, log *slog.Logger) Manager {
	client := asynq.NewClient(redisOpt)

	return &manager{
		client: client,
		log:    log,
	}
}

func (m *manager) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return m.client.EnqueueContext(ctx, task, opts...)
}

func (m *manager) Close() error {
	return m.client.Close()
}
