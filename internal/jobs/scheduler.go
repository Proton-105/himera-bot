package jobs

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

type Scheduler interface {
	RegisterTasks() error
	Run()
	Shutdown()
}

type scheduler struct {
	asynqScheduler *asynq.Scheduler
	log            *slog.Logger
}

func NewScheduler(redisOpt asynq.RedisConnOpt, log *slog.Logger) Scheduler {
	return &scheduler{
		asynqScheduler: asynq.NewScheduler(redisOpt, nil),
		log:            log,
	}
}

func (s *scheduler) RegisterTasks() error {
	task, err := NewPriceUpdateTask([]string{"ALL"})
	if err != nil {
		return err
	}

	if _, err := s.asynqScheduler.Register("*/30 * * * *", task); err != nil {
		return err
	}

	if s.log != nil {
		s.log.InfoContext(context.Background(), "scheduler: registered price update task")
	}

	return nil
}

func (s *scheduler) Run() {
	if s.log != nil {
		s.log.InfoContext(context.Background(), "scheduler: starting")
	}

	go func() {
		if err := s.asynqScheduler.Run(); err != nil && s.log != nil {
			s.log.ErrorContext(context.Background(), "scheduler: run failed", "error", err)
		}
	}()
}

func (s *scheduler) Shutdown() {
	if s.log != nil {
		s.log.InfoContext(context.Background(), "scheduler: shutting down")
	}

	s.asynqScheduler.Shutdown()
}
