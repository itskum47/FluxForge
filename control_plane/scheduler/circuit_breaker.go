package scheduler

import (
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitHalfOpen                     // Testing recovery
	CircuitOpen                         // Rejecting new tasks
)

func (cs CircuitState) String() string {
	switch cs {
	case CircuitClosed:
		return "closed"
	case CircuitHalfOpen:
		return "half_open"
	case CircuitOpen:
		return "open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements backpressure protection for the scheduler.
type CircuitBreaker struct {
	state CircuitState
	mu    sync.RWMutex

	// Configuration
	queueThreshold      int           // Max queue depth before opening
	saturationThreshold float64       // Max worker saturation before opening
	cooldownPeriod      time.Duration // Time before half-open

	// State tracking
	openedAt  time.Time
	testCount int // Number of test requests in half-open state
	testLimit int // Number of successes needed to close
}

// NewCircuitBreaker creates a new circuit breaker with production defaults.
func NewCircuitBreaker(queueThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:               CircuitClosed,
		queueThreshold:      queueThreshold,
		saturationThreshold: 0.95, // 95% worker saturation
		cooldownPeriod:      30 * time.Second,
		testLimit:           5, // 5 successful requests to close
	}
}

// ShouldAdmit determines if a new task should be admitted.
// Returns true if task should be accepted, false if rejected.
func (cb *CircuitBreaker) ShouldAdmit(queueDepth int, workerSaturation float64) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should transition from Open -> HalfOpen
	if cb.state == CircuitOpen && time.Since(cb.openedAt) > cb.cooldownPeriod {
		cb.state = CircuitHalfOpen
		cb.testCount = 0
	}

	// In half-open state, allow limited test traffic
	if cb.state == CircuitHalfOpen {
		// Allow small sample of requests
		if cb.testCount < cb.testLimit {
			cb.testCount++
			return true
		}
		// If test limit reached and still healthy, close circuit
		if queueDepth < cb.queueThreshold/2 && workerSaturation < cb.saturationThreshold {
			cb.state = CircuitClosed
			return true
		}
		// Still overloaded, stay half-open
		return false
	}

	// Check if we should open the circuit
	if queueDepth > cb.queueThreshold || workerSaturation > cb.saturationThreshold {
		cb.state = CircuitOpen
		cb.openedAt = time.Now()
		return false
	}

	// Normal operation
	return cb.state == CircuitClosed
}

// RecordSuccess notifies the circuit breaker of a successful task completion.
// Used in half-open state to determine if circuit should close.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		// If we've had enough successful tests, close the circuit
		if cb.testCount >= cb.testLimit {
			cb.state = CircuitClosed
		}
	}
}

// RecordFailure notifies the circuit breaker of a task failure.
// Used in half-open state to determine if circuit should re-open.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		// Re-open circuit on failure during testing
		cb.state = CircuitOpen
		cb.openedAt = time.Now()
		cb.testCount = 0
	}
}

// GetState returns the current circuit state (thread-safe).
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
