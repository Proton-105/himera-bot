package ratelimit

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type bucket struct {
	requests []time.Time
}

// MemoryLimiter is a simple in-memory fallback implementation of Limiter.
type MemoryLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	log     *slog.Logger
}

// NewMemoryLimiter returns an in-memory limiter implementation.
func NewMemoryLimiter(log *slog.Logger) Limiter {
	if log == nil {
		log = slog.Default()
	}

	return &MemoryLimiter{
		buckets: make(map[string]*bucket),
		log:     log,
	}
}

// Check enforces a sliding-window limit for the provided key.
func (m *MemoryLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (*Result, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	bkt := m.loadOrCreateBucket(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Bucket might have been removed between locks, ensure it still exists.
	if bkt == nil {
		bkt = m.ensureBucketLocked(key)
	}

	bkt.requests = keepRecent(bkt.requests, windowStart)
	count := len(bkt.requests)

	allowed := count < limit
	if allowed {
		bkt.requests = append(bkt.requests, now)
		count++
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	result := &Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   windowStart.Add(window),
	}

	if !allowed {
		return result, ErrLimitExceeded
	}

	return result, nil
}

// Cleanup removes buckets that have been inactive for more than maxAge.
func (m *MemoryLimiter) Cleanup(maxAge time.Duration) {
	if maxAge <= 0 {
		return
	}

	cutoff := time.Now().Add(-maxAge)

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, bkt := range m.buckets {
		if len(bkt.requests) == 0 {
			delete(m.buckets, key)
			continue
		}

		if bkt.requests[len(bkt.requests)-1].Before(cutoff) {
			delete(m.buckets, key)
		}
	}
}

func (m *MemoryLimiter) loadOrCreateBucket(key string) *bucket {
	m.mu.RLock()
	bkt := m.buckets[key]
	m.mu.RUnlock()

	if bkt != nil {
		return bkt
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bkt = m.buckets[key]; bkt == nil {
		bkt = &bucket{requests: make([]time.Time, 0, 8)}
		m.buckets[key] = bkt
	}

	return bkt
}

func (m *MemoryLimiter) ensureBucketLocked(key string) *bucket {
	if bkt, ok := m.buckets[key]; ok {
		return bkt
	}

	bkt := &bucket{requests: make([]time.Time, 0, 8)}
	m.buckets[key] = bkt
	return bkt
}

func keepRecent(reqs []time.Time, windowStart time.Time) []time.Time {
	firstIdx := 0
	for firstIdx < len(reqs) && reqs[firstIdx].Before(windowStart) {
		firstIdx++
	}

	if firstIdx == 0 {
		return reqs
	}

	if firstIdx >= len(reqs) {
		return reqs[:0]
	}

	copy(reqs, reqs[firstIdx:])
	return reqs[:len(reqs)-firstIdx]
}
