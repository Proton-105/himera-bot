package errors

import (
	"errors"
	"sync"
	"time"
)

const (
	ErrorThreshold      = 0.5
	MinRequests         = 10
	TimeoutDuration     = 30 * time.Second
	HalfOpenMaxRequests = 3
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var (
	errCircuitOpen             = errors.New("circuit breaker is open")
	errHalfOpenTooManyRequests = errors.New("too many requests in half-open")
)

type CircuitBreaker struct {
	mu              sync.Mutex
	state           State
	failures        int
	successes       int
	requests        int
	lastFailureTime time.Time
}

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state: StateClosed,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	if fn == nil {
		return nil
	}

	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) >= TimeoutDuration {
			cb.transitionToHalfOpenLocked()
		} else {
			cb.mu.Unlock()
			return errCircuitOpen
		}
	}

	if cb.state == StateHalfOpen && cb.requests >= HalfOpenMaxRequests {
		cb.mu.Unlock()
		return errHalfOpenTooManyRequests
	}
	cb.mu.Unlock()

	callErr := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if callErr != nil {
		cb.failures++
		cb.requests++

		if cb.state == StateHalfOpen {
			cb.tripToOpenLocked()
		} else {
			cb.evaluateState()
		}

		return callErr
	}

	cb.successes++
	cb.requests++

	if cb.state == StateHalfOpen && cb.successes >= HalfOpenMaxRequests {
		cb.state = StateClosed
		cb.resetCountersLocked()
		return nil
	}

	return nil
}

func (cb *CircuitBreaker) evaluateState() {
	if cb.requests < MinRequests {
		return
	}

	errorRate := float64(cb.failures) / float64(cb.requests)
	if errorRate >= ErrorThreshold {
		cb.tripToOpenLocked()
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) resetCountersLocked() {
	cb.failures = 0
	cb.successes = 0
	cb.requests = 0
}

func (cb *CircuitBreaker) transitionToHalfOpenLocked() {
	cb.state = StateHalfOpen
	cb.resetCountersLocked()
}

func (cb *CircuitBreaker) tripToOpenLocked() {
	cb.state = StateOpen
	cb.lastFailureTime = time.Now()
	cb.resetCountersLocked()
}
