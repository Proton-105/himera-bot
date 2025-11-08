package jobs

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// Worker provides APIs to register handlers and control the background worker lifecycle.
type Worker interface {
	RegisterHandler(taskType string, handler asynq.Handler)
	Run() error
	Shutdown()
}

type worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	log    *slog.Logger
}

var _ Worker = (*worker)(nil)

// NewWorker constructs a Worker backed by an asynq.Server instance.
func NewWorker(redisOpt asynq.RedisConnOpt, queues map[string]int, log *slog.Logger) Worker {
	server := asynq.NewServer(redisOpt, asynq.Config{
		Queues:         queues,
		Concurrency:    10,
		RetryDelayFunc: asynq.DefaultRetryDelayFunc,
	})

	mux := asynq.NewServeMux()

	return &worker{
		server: server,
		mux:    mux,
		log:    log,
	}
}

// RegisterHandler wires a task type to the provided handler.
func (w *worker) RegisterHandler(taskType string, handler asynq.Handler) {
	w.mux.Handle(taskType, handler)
}

// Run starts the underlying asynq server to process tasks.
func (w *worker) Run() error {
	if w.log != nil {
		w.log.InfoContext(context.Background(), "jobs worker: starting processing loop")
	}

	return w.server.Run(w.mux)
}

// Shutdown gracefully stops the worker.
func (w *worker) Shutdown() {
	if w.log != nil {
		w.log.InfoContext(context.Background(), "jobs worker: shutting down")
	}

	w.server.Shutdown()
}
