package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Shutdown coordinates graceful shutdown hooks in parallel.
type Shutdown struct {
	mu    sync.Mutex
	hooks []Hook
	log   *slog.Logger
}

// NewShutdown constructs a new Shutdown coordinator.
func NewShutdown(log *slog.Logger) *Shutdown {
	if log == nil {
		log = slog.Default()
	}

	return &Shutdown{log: log}
}

// Register adds a named shutdown hook.
func (s *Shutdown) Register(name string, fn func(context.Context) error) {
	if fn == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.hooks = append(s.hooks, Hook{Name: name, Fn: fn})
}

// Execute runs all registered hooks concurrently and waits for completion.
func (s *Shutdown) Execute(ctx context.Context) error {
	s.mu.Lock()
	hooks := append([]Hook(nil), s.hooks...)
	s.mu.Unlock()

	start := time.Now()
	if s.log != nil {
		s.log.Info("shutdown sequence started", slog.Int("hook_count", len(hooks)))
	}

	var wg sync.WaitGroup
	var errMu sync.Mutex
	errs := make([]string, 0)

	for _, hook := range hooks {
		h := hook
		if h.Fn == nil {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			hookCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			if s.log != nil {
				s.log.Info("running shutdown hook", slog.String("hook", h.Name))
			}

			if err := h.Fn(hookCtx); err != nil {
				if s.log != nil {
					s.log.Error("shutdown hook failed", slog.String("hook", h.Name), slog.Any("error", err))
				}
				errMu.Lock()
				errs = append(errs, fmt.Sprintf("%s: %v", h.Name, err))
				errMu.Unlock()
				return
			}

			if s.log != nil {
				s.log.Info("shutdown hook completed", slog.String("hook", h.Name))
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	if s.log != nil {
		s.log.Info("shutdown sequence finished", slog.Duration("elapsed", elapsed))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
