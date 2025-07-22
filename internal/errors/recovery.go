// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package errors

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// ErrorRecovery provides advanced error handling and recovery capabilities.
type ErrorRecovery struct {
	logger       Logger
	maxRetries   int
	retryDelay   time.Duration
	errorCounts  map[string]int
	recoveryFunc func(error) error
}

// Logger interface for error recovery.
type Logger interface {
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// RecoveryConfig configures error recovery behavior.
type RecoveryConfig struct {
	MaxRetries       int
	RetryDelay       time.Duration
	CircuitThreshold int
	ResetTimeout     time.Duration
	Logger           Logger
	RecoveryFunc     func(error) error
}

// ErrorType represents different types of errors.
type ErrorType string

const (
	// ErrorTypeNetwork represents network-related errors.
	ErrorTypeNetwork ErrorType = "network"
	// ErrorTypeAuth represents authentication-related errors.
	ErrorTypeAuth ErrorType = "auth"
	// ErrorTypeValidation represents validation-related errors.
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeSystem represents system-related errors.
	ErrorTypeSystem ErrorType = "system"
	// ErrorTypeTimeout represents timeout-related errors.
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeRateLimit represents rate limit-related errors.
	ErrorTypeRateLimit ErrorType = "rate_limit"
	// ErrorTypeUnknown represents unknown error types.
	ErrorTypeUnknown ErrorType = "unknown"
)

// RecoverableError represents an error that can be recovered from.
type RecoverableError struct {
	Type       ErrorType
	Message    string
	Cause      error
	Retryable  bool
	Context    map[string]interface{}
	StackTrace string
	Timestamp  time.Time
}

func (e *RecoverableError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
}

func (e *RecoverableError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns whether the error can be retried.
func (e *RecoverableError) IsRetryable() bool {
	return e.Retryable
}

// NewErrorRecovery creates a new error recovery system.
func NewErrorRecovery(config RecoveryConfig) *ErrorRecovery {
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}

	return &ErrorRecovery{
		logger:       config.Logger,
		maxRetries:   config.MaxRetries,
		retryDelay:   config.RetryDelay,
		errorCounts:  make(map[string]int),
		recoveryFunc: config.RecoveryFunc,
	}
}

// NewRecoverableError creates a new recoverable error.
func NewRecoverableError(errType ErrorType, message string, cause error, retryable bool) *RecoverableError {
	return &RecoverableError{
		Type:       errType,
		Message:    message,
		Cause:      cause,
		Retryable:  retryable,
		Context:    make(map[string]interface{}),
		StackTrace: string(debug.Stack()),
		Timestamp:  time.Now(),
	}
}

// WithContext adds context to the error.
func (e *RecoverableError) WithContext(key string, value interface{}) *RecoverableError {
	e.Context[key] = value
	return e
}

// Execute runs a function with automatic retry and recovery.
func (er *ErrorRecovery) Execute(ctx context.Context, operation string, fn func() error) error {
	return er.ExecuteWithResult(ctx, operation, func() (interface{}, error) {
		return nil, fn()
	})
}

// ExecuteWithResult runs a function with return value and automatic retry.
func (er *ErrorRecovery) ExecuteWithResult(ctx context.Context, operation string, fn func() (interface{}, error)) error { //nolint:gocognit // Complex error recovery with multiple retry strategies, timeout handling, and circuit breaker logic
	var lastErr error

	for attempt := 0; attempt <= er.maxRetries; attempt++ {
		if attempt > 0 {
			er.logger.Info("Retrying operation", "operation", operation, "attempt", attempt, "max_retries", er.maxRetries)

			select {
			case <-ctx.Done():
				return fmt.Errorf("operation canceled: %w", ctx.Err())
			case <-time.After(er.calculateBackoff(attempt)):
				// Continue with retry
			}
		}

		// Execute the function
		_, err := fn()
		if err == nil {
			if attempt > 0 {
				er.logger.Info("Operation succeeded after retry", "operation", operation, "attempts", attempt+1)
			}

			return nil
		}

		lastErr = err

		// Check if error is recoverable
		recErr := &RecoverableError{}
		if errors.As(err, &recErr) {
			if !recErr.IsRetryable() {
				er.logger.Error("Non-retryable error encountered", "operation", operation, "error", err)
				return err
			}
		}

		// Log the error
		er.logger.Warn("Operation failed, will retry",
			"operation", operation,
			"attempt", attempt+1,
			"error", err,
			"max_retries", er.maxRetries)

		// Try to recover
		if er.recoveryFunc != nil {
			if recoveryErr := er.recoveryFunc(err); recoveryErr != nil {
				er.logger.Error("Recovery function failed", "error", recoveryErr)
			}
		}
	}

	er.logger.Error("Operation failed after all retries",
		"operation", operation,
		"attempts", er.maxRetries+1,
		"final_error", lastErr)

	return fmt.Errorf("operation failed after %d attempts: %w", er.maxRetries+1, lastErr)
}

// RecoverPanic recovers from panics and converts them to errors.
func (er *ErrorRecovery) RecoverPanic() {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		err := fmt.Errorf("panic recovered: %v\nStack trace:\n%s", r, stack)

		er.logger.Error("Panic recovered", "panic", r, "stack", stack)

		if er.recoveryFunc != nil {
			if recoveryErr := er.recoveryFunc(err); recoveryErr != nil {
				er.logger.Error("Recovery function failed after panic", "error", recoveryErr)
			}
		}
	}
}

// WithPanicRecovery wraps a function with panic recovery.
func (er *ErrorRecovery) WithPanicRecovery(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err = fmt.Errorf("panic recovered: %v\nStack trace:\n%s", r, stack)

			er.logger.Error("Panic recovered", "panic", r, "stack", stack)

			if er.recoveryFunc != nil {
				if recoveryErr := er.recoveryFunc(err); recoveryErr != nil {
					er.logger.Error("Recovery function failed after panic", "error", recoveryErr)
				}
			}
		}
	}()

	fn()

	return nil
}

// calculateBackoff calculates the backoff delay for retries.
func (er *ErrorRecovery) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff with jitter
	// Limit attempt to prevent overflow (max 30 to stay within safe bit shift range)
	if attempt > 30 {
		attempt = 30
	}
	// Safe bit shift: ensure attempt is within range [0, 30] to prevent integer overflow
	shiftAmount := uint(attempt) //nolint:gosec // G115: bounds-checked above, safe conversion
	if shiftAmount > 30 {
		shiftAmount = 30
	}
	delay := er.retryDelay * time.Duration(1<<shiftAmount) //nolint:gosec // Bounded by max 30, safe from overflow
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}

	return delay
}

// CircuitBreaker implements a simple circuit breaker pattern.
type CircuitBreaker struct {
	maxFailures int
	resetTime   time.Duration
	failures    int
	lastFailure time.Time
	state       CircuitState
	mu          sync.RWMutex
}

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	// StateClosed represents a closed circuit breaker state.
	StateClosed CircuitState = iota
	// StateOpen represents an open circuit breaker state.
	StateOpen
	// StateHalfOpen represents a half-open circuit breaker state.
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(maxFailures int, resetTime time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		resetTime:   resetTime,
		state:       StateClosed,
	}
}

// Execute executes a function through the circuit breaker.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()
	cb.recordResult(err)

	return err
}

// allowRequest checks if a request should be allowed.
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return time.Since(cb.lastFailure) > cb.resetTime
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an operation.
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
	} else {
		cb.failures = 0
		cb.state = StateClosed
	}
}

// GetState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// GetMemoryStats returns current memory statistics.
func GetMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_mb":        bToMb(m.Alloc),
		"total_alloc_mb":  bToMb(m.TotalAlloc),
		"sys_mb":          bToMb(m.Sys),
		"num_gc":          m.NumGC,
		"goroutines":      runtime.NumGoroutine(),
		"heap_objects":    m.HeapObjects,
		"stack_in_use_mb": bToMb(m.StackInuse),
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// HealthCheck represents a system health check.
type HealthCheck struct {
	Name        string
	Description string
	CheckFunc   func() error
	Timeout     time.Duration
}

// HealthMonitor monitors system health.
type HealthMonitor struct {
	checks []HealthCheck
	logger Logger
}

// NewHealthMonitor creates a new health monitor.
func NewHealthMonitor(logger Logger) *HealthMonitor {
	return &HealthMonitor{
		checks: make([]HealthCheck, 0),
		logger: logger,
	}
}

// AddCheck adds a health check.
func (hm *HealthMonitor) AddCheck(check HealthCheck) {
	hm.checks = append(hm.checks, check)
}

// RunChecks runs all health checks.
func (hm *HealthMonitor) RunChecks(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for _, check := range hm.checks {
		var (
			checkCtx = ctx
			cancel   context.CancelFunc
		)

		if check.Timeout > 0 {
			checkCtx, cancel = context.WithTimeout(ctx, check.Timeout)
		}

		err := func() error {
			defer func() {
				if r := recover(); r != nil {
					hm.logger.Error("Health check panicked", "check", check.Name, "panic", r)
				}
			}()
			// Use checkCtx for timeout handling if needed in the future
			select {
			case <-checkCtx.Done():
				return checkCtx.Err()
			default:
				return check.CheckFunc()
			}
		}()

		if cancel != nil {
			cancel()
		}

		results[check.Name] = err
		if err != nil {
			hm.logger.Warn("Health check failed", "check", check.Name, "error", err)
		} else {
			hm.logger.Debug("Health check passed", "check", check.Name)
		}
	}

	return results
}
