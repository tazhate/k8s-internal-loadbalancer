package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu sync.RWMutex

	maxRequests      uint32
	interval         time.Duration
	timeout          time.Duration
	consecutiveFailures uint32

	state            State
	generation       uint64
	counts           *counts
	expiry           time.Time
}

type counts struct {
	requests         uint32
	totalSuccesses   uint32
	totalFailures    uint32
	consecutiveSuccesses uint32
	consecutiveFailures  uint32
}

// New creates a new CircuitBreaker
func New(maxRequests uint32, interval, timeout time.Duration, consecutiveFailures uint32) *CircuitBreaker {
	cb := &CircuitBreaker{
		maxRequests:      maxRequests,
		interval:         interval,
		timeout:          timeout,
		consecutiveFailures: consecutiveFailures,
	}
	cb.toNewGeneration(time.Now())
	return cb
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(req func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	err = req()
	cb.afterRequest(generation, err == nil)
	return err
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state.String()
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, fmt.Errorf("circuit breaker is open")
	} else if state == StateHalfOpen && cb.counts.requests >= cb.maxRequests {
		return generation, fmt.Errorf("circuit breaker is half-open and max requests reached")
	}

	cb.counts.requests++
	return generation, nil
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// currentState returns the current state without modifying it
func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	cb.counts.totalSuccesses++
	cb.counts.consecutiveSuccesses++
	cb.counts.consecutiveFailures = 0

	if state == StateHalfOpen {
		if cb.counts.consecutiveSuccesses >= cb.maxRequests {
			cb.setState(StateClosed, now)
		}
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	cb.counts.totalFailures++
	cb.counts.consecutiveFailures++
	cb.counts.consecutiveSuccesses = 0

	if state == StateClosed && cb.counts.consecutiveFailures >= cb.consecutiveFailures {
		cb.setState(StateOpen, now)
	} else if state == StateHalfOpen {
		cb.setState(StateOpen, now)
	}
}

// setState transitions to a new state
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	cb.state = state
	cb.toNewGeneration(now)
}

// toNewGeneration starts a new generation
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = &counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default:
		cb.expiry = zero
	}
}

// Stats returns the current statistics
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":                cb.state.String(),
		"requests":             cb.counts.requests,
		"total_successes":      cb.counts.totalSuccesses,
		"total_failures":       cb.counts.totalFailures,
		"consecutive_successes": cb.counts.consecutiveSuccesses,
		"consecutive_failures":  cb.counts.consecutiveFailures,
	}
}
