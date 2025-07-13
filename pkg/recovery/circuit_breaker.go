package recovery

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the current state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig defines configuration for circuit breaker
type CircuitBreakerConfig struct {
	Name                  string        `json:"name"`
	FailureThreshold      int           `json:"failure_threshold"`        // Number of failures to trip breaker
	SuccessThreshold      int           `json:"success_threshold"`        // Consecutive successes to close from half-open
	Timeout               time.Duration `json:"timeout"`                  // How long to keep breaker open
	MaxConcurrentCalls    int           `json:"max_concurrent_calls"`     // Max concurrent calls in half-open state
	SlowCallThreshold     time.Duration `json:"slow_call_threshold"`      // Threshold for considering call slow
	SlowCallRateThreshold float64       `json:"slow_call_rate_threshold"` // Percentage of slow calls to trip
	MonitoringWindow      time.Duration `json:"monitoring_window"`        // Window for calculating rates
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:                  name,
		FailureThreshold:      5,
		SuccessThreshold:      3,
		Timeout:               30 * time.Second,
		MaxConcurrentCalls:    5,
		SlowCallThreshold:     5 * time.Second,
		SlowCallRateThreshold: 0.5,
		MonitoringWindow:      1 * time.Minute,
	}
}

// CallResult represents the result of a function call
type CallResult struct {
	Success  bool
	Duration time.Duration
	Error    error
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitState
	failureCount    int
	successCount    int
	calls           []CallResult
	lastFailure     time.Time
	lastStateChange time.Time
	concurrentCalls int64
	mu              sync.RWMutex
	onStateChange   func(name string, from, to CircuitState)
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config:          config,
		state:           StateClosed,
		calls:           make([]CallResult, 0),
		lastStateChange: time.Now(),
	}
}

// SetStateChangeCallback sets a callback for state changes
func (cb *CircuitBreaker) SetStateChangeCallback(callback func(name string, from, to CircuitState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = callback
}

// Execute executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can execute
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker %s is OPEN", cb.config.Name)
	}

	// Track concurrent calls for half-open state
	cb.mu.Lock()
	cb.concurrentCalls++
	cb.mu.Unlock()

	defer func() {
		cb.mu.Lock()
		cb.concurrentCalls--
		cb.mu.Unlock()
	}()

	// Execute with timeout
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record result
	result := CallResult{
		Success:  err == nil,
		Duration: duration,
		Error:    err,
	}

	cb.recordResult(result)

	return err
}

// canExecute determines if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastStateChange) >= cb.config.Timeout {
			cb.mu.RUnlock()
			cb.setState(StateHalfOpen)
			cb.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		// Allow limited concurrent calls
		return cb.concurrentCalls < int64(cb.config.MaxConcurrentCalls)
	default:
		return false
	}
}

// recordResult records the result of a function call and updates state if necessary
func (cb *CircuitBreaker) recordResult(result CallResult) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Add to call history
	cb.calls = append(cb.calls, result)

	// Clean old calls outside monitoring window
	cutoff := time.Now().Add(-cb.config.MonitoringWindow)
	for i, call := range cb.calls {
		if time.Now().Add(-call.Duration).After(cutoff) {
			cb.calls = cb.calls[i:]
			break
		}
	}

	if result.Success {
		cb.successCount++
		cb.failureCount = 0 // Reset failure count on success

		// In half-open state, check if we can close
		if cb.state == StateHalfOpen && cb.successCount >= cb.config.SuccessThreshold {
			cb.setState(StateClosed)
		}
	} else {
		cb.failureCount++
		cb.successCount = 0 // Reset success count on failure
		cb.lastFailure = time.Now()

		// Check if we should open the circuit
		if cb.shouldTrip() {
			cb.setState(StateOpen)
		}
	}
}

// shouldTrip determines if the circuit breaker should trip based on failure rate
func (cb *CircuitBreaker) shouldTrip() bool {
	// Don't trip if we're already open
	if cb.state == StateOpen {
		return false
	}

	// Check failure count threshold
	if cb.failureCount >= cb.config.FailureThreshold {
		return true
	}

	// Check slow call rate if we have enough data
	if len(cb.calls) < cb.config.FailureThreshold {
		return false
	}

	slowCalls := 0
	for _, call := range cb.calls {
		if call.Duration >= cb.config.SlowCallThreshold {
			slowCalls++
		}
	}

	slowCallRate := float64(slowCalls) / float64(len(cb.calls))
	return slowCallRate >= cb.config.SlowCallRateThreshold
}

// setState changes the circuit breaker state and triggers callback
func (cb *CircuitBreaker) setState(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters on state change
	if newState == StateClosed {
		cb.failureCount = 0
		cb.successCount = 0
	} else if newState == StateHalfOpen {
		cb.successCount = 0
	}

	// Trigger callback if set
	if cb.onStateChange != nil {
		go cb.onStateChange(cb.config.Name, oldState, newState)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	totalCalls := len(cb.calls)
	successfulCalls := 0
	slowCalls := 0
	totalDuration := time.Duration(0)

	for _, call := range cb.calls {
		if call.Success {
			successfulCalls++
		}
		if call.Duration >= cb.config.SlowCallThreshold {
			slowCalls++
		}
		totalDuration += call.Duration
	}

	var avgDuration time.Duration
	if totalCalls > 0 {
		avgDuration = totalDuration / time.Duration(totalCalls)
	}

	return CircuitBreakerMetrics{
		Name:            cb.config.Name,
		State:           cb.state,
		TotalCalls:      totalCalls,
		SuccessfulCalls: successfulCalls,
		FailedCalls:     totalCalls - successfulCalls,
		SlowCalls:       slowCalls,
		SuccessRate:     float64(successfulCalls) / float64(totalCalls),
		SlowCallRate:    float64(slowCalls) / float64(totalCalls),
		AverageDuration: avgDuration,
		FailureCount:    cb.failureCount,
		SuccessCount:    cb.successCount,
		LastFailure:     cb.lastFailure,
		LastStateChange: cb.lastStateChange,
		ConcurrentCalls: cb.concurrentCalls,
	}
}

// CircuitBreakerMetrics contains metrics for a circuit breaker
type CircuitBreakerMetrics struct {
	Name            string        `json:"name"`
	State           CircuitState  `json:"state"`
	TotalCalls      int           `json:"total_calls"`
	SuccessfulCalls int           `json:"successful_calls"`
	FailedCalls     int           `json:"failed_calls"`
	SlowCalls       int           `json:"slow_calls"`
	SuccessRate     float64       `json:"success_rate"`
	SlowCallRate    float64       `json:"slow_call_rate"`
	AverageDuration time.Duration `json:"average_duration"`
	FailureCount    int           `json:"failure_count"`
	SuccessCount    int           `json:"success_count"`
	LastFailure     time.Time     `json:"last_failure"`
	LastStateChange time.Time     `json:"last_state_change"`
	ConcurrentCalls int64         `json:"concurrent_calls"`
}

// Reset resets the circuit breaker to its initial state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed)
	cb.failureCount = 0
	cb.successCount = 0
	cb.calls = make([]CallResult, 0)
	cb.concurrentCalls = 0
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateOpen)
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateClosed)
}
