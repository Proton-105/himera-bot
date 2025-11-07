package ratelimit

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	rateLimitChecksTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ratelimit_checks_total",
		Help: "Total number of rate limit checks by backend and result.",
	}, []string{"backend", "result"})

	rateLimitRejectedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ratelimit_rejected_total",
		Help: "Total number of rejected requests per backend.",
	}, []string{"backend"})

	rateLimitRedisErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ratelimit_redis_errors_total",
		Help: "Total number of Redis errors encountered by the limiter.",
	})
)

func init() {
	prometheus.MustRegister(rateLimitChecksTotal, rateLimitRejectedTotal, rateLimitRedisErrorsTotal)
}

// AdaptiveLimiter delegates to a primary (Redis) limiter and falls back to
// a stricter in-memory limiter when the primary fails.
type AdaptiveLimiter struct {
	primary  Limiter
	fallback Limiter
	log      *slog.Logger
}

// NewAdaptiveLimiter creates a limiter that adapts between Redis and in-memory backends.
func NewAdaptiveLimiter(primary, fallback Limiter, log *slog.Logger) Limiter {
	if log == nil {
		log = slog.Default()
	}

	return &AdaptiveLimiter{
		primary:  primary,
		fallback: fallback,
		log:      log,
	}
}

// Check evaluates the limit using the primary backend, falling back to memory on errors.
func (a *AdaptiveLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (*Result, error) {
	result, err := a.primary.Check(ctx, key, limit, window)
	if err == nil {
		rateLimitChecksTotal.WithLabelValues("redis", boolLabel(result.Allowed)).Inc()
		if !result.Allowed {
			rateLimitRejectedTotal.WithLabelValues("redis").Inc()
			return result, ErrLimitExceeded
		}
		return result, nil
	}

	rateLimitRedisErrorsTotal.Inc()
	if a.log != nil {
		a.log.Warn("redis limiter failed, falling back to in-memory", "key", key, "error", err)
	}

	fallbackLimit := limit / 2
	if fallbackLimit <= 0 {
		fallbackLimit = 1
	}

	fallbackResult, fallbackErr := a.fallback.Check(ctx, key, fallbackLimit, window)
	if fallbackErr != nil {
		return fallbackResult, fallbackErr
	}

	rateLimitChecksTotal.WithLabelValues("fallback", boolLabel(fallbackResult.Allowed)).Inc()
	if !fallbackResult.Allowed {
		rateLimitRejectedTotal.WithLabelValues("fallback").Inc()
		return fallbackResult, ErrLimitExceeded
	}

	return fallbackResult, nil
}

func boolLabel(value bool) string {
	if value {
		return "allowed"
	}
	return "rejected"
}
