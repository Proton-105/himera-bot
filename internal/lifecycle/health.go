package lifecycle

import (
	"context"
	"log/slog"
)

// HealthChecker exposes liveness and readiness probes.
type HealthChecker interface {
	Liveness(ctx context.Context) error
	Readiness(ctx context.Context) error
}

// Probes is a stub implementation of HealthChecker.
type Probes struct {
	log *slog.Logger
}

// NewProbes creates a new Probes instance.
func NewProbes(log *slog.Logger) *Probes {
	if log == nil {
		log = slog.Default()
	}
	return &Probes{log: log}
}

// Liveness currently always reports success.
func (p *Probes) Liveness(ctx context.Context) error {
	if p.log != nil {
		p.log.Debug("liveness probe called")
	}
	return nil
}

// Readiness currently always reports success.
func (p *Probes) Readiness(ctx context.Context) error {
	if p.log != nil {
		p.log.Debug("readiness probe called")
	}
	return nil
}
