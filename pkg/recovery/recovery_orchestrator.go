package recovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/errors"
)

// RecoveryStrategy defines how to recover from failures
type RecoveryStrategy int

const (
	StrategyNone RecoveryStrategy = iota
	StrategyRetry
	StrategyFallback
	StrategyCircuitBreaker
	StrategyGradualRecovery
	StrategyAutoFix
)

func (rs RecoveryStrategy) String() string {
	switch rs {
	case StrategyNone:
		return "NONE"
	case StrategyRetry:
		return "RETRY"
	case StrategyFallback:
		return "FALLBACK"
	case StrategyCircuitBreaker:
		return "CIRCUIT_BREAKER"
	case StrategyGradualRecovery:
		return "GRADUAL_RECOVERY"
	case StrategyAutoFix:
		return "AUTO_FIX"
	default:
		return "UNKNOWN"
	}
}

// RecoveryPolicy defines how to handle different types of failures
type RecoveryPolicy struct {
	ErrorCodePattern string           `json:"error_code_pattern"`
	DomainPattern    string           `json:"domain_pattern"`
	Strategy         RecoveryStrategy `json:"strategy"`
	MaxAttempts      int              `json:"max_attempts"`
	BackoffStrategy  BackoffStrategy  `json:"backoff_strategy"`
	Timeout          time.Duration    `json:"timeout"`
	FallbackAction   string           `json:"fallback_action,omitempty"`
	AutoFixCommand   string           `json:"auto_fix_command,omitempty"`
	Priority         int              `json:"priority"` // Higher = more important
}

// BackoffStrategy defines different backoff strategies
type BackoffStrategy int

const (
	BackoffFixed BackoffStrategy = iota
	BackoffLinear
	BackoffExponential
	BackoffExponentialJitter
)

// RecoveryAttempt tracks a recovery attempt
type RecoveryAttempt struct {
	ID           string                 `json:"id"`
	ErrorCode    string                 `json:"error_code"`
	Strategy     RecoveryStrategy       `json:"strategy"`
	AttemptCount int                    `json:"attempt_count"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Success      bool                   `json:"success"`
	Error        error                  `json:"error,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// FallbackProvider defines interface for fallback mechanisms
type FallbackProvider interface {
	CanHandle(errorCode string) bool
	Execute(ctx context.Context, originalError error) error
	GetInfo() FallbackInfo
}

// FallbackInfo describes a fallback provider
type FallbackInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

// RecoveryOrchestrator manages automatic recovery across the system
type RecoveryOrchestrator struct {
	policies          []RecoveryPolicy
	circuitBreakers   map[string]*CircuitBreaker
	fallbackProviders []FallbackProvider
	attempts          map[string]*RecoveryAttempt
	solutionEngine    *errors.SolutionEngine
	metrics           RecoveryMetrics
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
}

// RecoveryMetrics tracks recovery statistics
type RecoveryMetrics struct {
	TotalAttempts        int64                    `json:"total_attempts"`
	SuccessfulRecoveries int64                    `json:"successful_recoveries"`
	FailedRecoveries     int64                    `json:"failed_recoveries"`
	StrategyUsage        map[string]int64         `json:"strategy_usage"`
	RecoveryTimes        map[string]time.Duration `json:"recovery_times"`
	CircuitBreakerTrips  int64                    `json:"circuit_breaker_trips"`
	LastUpdated          time.Time                `json:"last_updated"`
}

// NewRecoveryOrchestrator creates a new recovery orchestrator
func NewRecoveryOrchestrator() *RecoveryOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	ro := &RecoveryOrchestrator{
		policies:          make([]RecoveryPolicy, 0),
		circuitBreakers:   make(map[string]*CircuitBreaker),
		fallbackProviders: make([]FallbackProvider, 0),
		attempts:          make(map[string]*RecoveryAttempt),
		solutionEngine:    errors.NewSolutionEngine(),
		metrics: RecoveryMetrics{
			StrategyUsage: make(map[string]int64),
			RecoveryTimes: make(map[string]time.Duration),
			LastUpdated:   time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize default policies
	ro.initializeDefaultPolicies()

	return ro
}

// RecoverFromError attempts to recover from an error using configured strategies
func (ro *RecoveryOrchestrator) RecoverFromError(ctx context.Context, userErr *errors.UserError, operation func() error) error {
	// Find matching policy
	policy := ro.findPolicy(userErr)
	if policy == nil {
		return fmt.Errorf("no recovery policy found for error: %s", userErr.Code.String())
	}

	// Create recovery attempt
	attempt := &RecoveryAttempt{
		ID:           fmt.Sprintf("recovery_%d", time.Now().UnixNano()),
		ErrorCode:    userErr.Code.String(),
		Strategy:     policy.Strategy,
		AttemptCount: 0,
		StartTime:    time.Now(),
		Context:      make(map[string]interface{}),
	}

	ro.mu.Lock()
	ro.attempts[attempt.ID] = attempt
	ro.metrics.TotalAttempts++
	ro.metrics.StrategyUsage[policy.Strategy.String()]++
	ro.mu.Unlock()

	// Execute recovery strategy
	var err error
	switch policy.Strategy {
	case StrategyRetry:
		err = ro.retryWithBackoff(ctx, operation, policy, attempt)
	case StrategyFallback:
		err = ro.executeFallback(ctx, userErr, policy, attempt)
	case StrategyCircuitBreaker:
		err = ro.executeWithCircuitBreaker(ctx, userErr.Code.String(), operation, attempt)
	case StrategyGradualRecovery:
		err = ro.executeGradualRecovery(ctx, operation, policy, attempt)
	case StrategyAutoFix:
		err = ro.executeAutoFix(ctx, userErr, policy, attempt)
	default:
		err = operation()
	}

	// Update attempt
	now := time.Now()
	attempt.EndTime = &now
	attempt.Success = err == nil
	attempt.Error = err

	// Update metrics
	ro.mu.Lock()
	if err == nil {
		ro.metrics.SuccessfulRecoveries++
	} else {
		ro.metrics.FailedRecoveries++
	}
	duration := now.Sub(attempt.StartTime)
	ro.metrics.RecoveryTimes[policy.Strategy.String()] = duration
	ro.metrics.LastUpdated = time.Now()
	ro.mu.Unlock()

	return err
}

// retryWithBackoff implements retry logic with configurable backoff
func (ro *RecoveryOrchestrator) retryWithBackoff(ctx context.Context, operation func() error, policy *RecoveryPolicy, attempt *RecoveryAttempt) error {
	var lastErr error

	for attempt.AttemptCount < policy.MaxAttempts {
		attempt.AttemptCount++

		// Execute operation
		lastErr = operation()
		if lastErr == nil {
			return nil
		}

		// Don't delay on last attempt
		if attempt.AttemptCount >= policy.MaxAttempts {
			break
		}

		// Calculate backoff delay
		delay := ro.calculateBackoffDelay(attempt.AttemptCount, policy.BackoffStrategy)

		// Wait with context cancellation support
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// executeFallback tries fallback providers
func (ro *RecoveryOrchestrator) executeFallback(ctx context.Context, userErr *errors.UserError, policy *RecoveryPolicy, attempt *RecoveryAttempt) error {
	ro.mu.RLock()
	providers := ro.fallbackProviders
	ro.mu.RUnlock()

	// Try fallback providers in priority order
	for _, provider := range providers {
		if provider.CanHandle(userErr.Code.String()) {
			attempt.Context["fallback_provider"] = provider.GetInfo().Name
			return provider.Execute(ctx, userErr)
		}
	}

	return fmt.Errorf("no fallback provider available for error: %s", userErr.Code.String())
}

// executeWithCircuitBreaker wraps operation with circuit breaker
func (ro *RecoveryOrchestrator) executeWithCircuitBreaker(ctx context.Context, errorCode string, operation func() error, attempt *RecoveryAttempt) error {
	cb := ro.getOrCreateCircuitBreaker(errorCode)

	attempt.Context["circuit_breaker_state"] = cb.GetState().String()

	err := cb.Execute(ctx, operation)
	if cb.GetState() == StateOpen {
		ro.mu.Lock()
		ro.metrics.CircuitBreakerTrips++
		ro.mu.Unlock()
	}

	return err
}

// executeGradualRecovery implements gradual load increase after recovery
func (ro *RecoveryOrchestrator) executeGradualRecovery(ctx context.Context, operation func() error, policy *RecoveryPolicy, attempt *RecoveryAttempt) error {
	// Start with reduced load
	stages := []float64{0.1, 0.3, 0.5, 0.8, 1.0}

	for i, load := range stages {
		attempt.Context[fmt.Sprintf("stage_%d_load", i)] = load

		// Simulate reduced load (in real implementation, this would control actual load)
		time.Sleep(time.Duration(float64(time.Second) * (1.0 - load)))

		err := operation()
		if err != nil {
			return fmt.Errorf("gradual recovery failed at stage %d (load %.1f): %w", i+1, load, err)
		}

		// Wait between stages
		select {
		case <-time.After(2 * time.Second):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// executeAutoFix attempts to automatically fix the issue
func (ro *RecoveryOrchestrator) executeAutoFix(ctx context.Context, userErr *errors.UserError, policy *RecoveryPolicy, attempt *RecoveryAttempt) error {
	// Get solutions from solution engine
	solutions, err := ro.solutionEngine.GetSolutions(ctx, userErr)
	if err != nil {
		return fmt.Errorf("failed to get solutions: %w", err)
	}

	// Try automated solutions
	for _, solution := range solutions {
		if solution.Automated {
			attempt.Context["auto_fix_solution"] = solution.ID
			err := ro.solutionEngine.ApplySolution(ctx, solution)
			if err == nil {
				// Record successful fix
				ro.solutionEngine.RecordSolutionFeedback(solution.ID, 1.0, true)
				return nil
			}
		}
	}

	return fmt.Errorf("no automated solutions available for error: %s", userErr.Code.String())
}

// calculateBackoffDelay calculates delay based on backoff strategy
func (ro *RecoveryOrchestrator) calculateBackoffDelay(attempt int, strategy BackoffStrategy) time.Duration {
	baseDelay := time.Second

	switch strategy {
	case BackoffFixed:
		return baseDelay
	case BackoffLinear:
		return time.Duration(attempt) * baseDelay
	case BackoffExponential:
		return time.Duration(1<<uint(attempt-1)) * baseDelay
	case BackoffExponentialJitter:
		delay := time.Duration(1<<uint(attempt-1)) * baseDelay
		jitter := time.Duration(float64(delay) * 0.1)
		return delay + jitter
	default:
		return baseDelay
	}
}

// findPolicy finds the best matching policy for an error
func (ro *RecoveryOrchestrator) findPolicy(userErr *errors.UserError) *RecoveryPolicy {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	var bestMatch *RecoveryPolicy
	bestPriority := -1

	for i := range ro.policies {
		policy := &ro.policies[i]

		// Check error code pattern match
		if policy.ErrorCodePattern != "" {
			if matched, _ := matchPattern(policy.ErrorCodePattern, userErr.Code.String()); matched {
				if policy.Priority > bestPriority {
					bestMatch = policy
					bestPriority = policy.Priority
				}
			}
		}

		// Check domain pattern match
		if policy.DomainPattern != "" {
			if matched, _ := matchPattern(policy.DomainPattern, userErr.Code.Domain); matched {
				if policy.Priority > bestPriority {
					bestMatch = policy
					bestPriority = policy.Priority
				}
			}
		}
	}

	return bestMatch
}

// getOrCreateCircuitBreaker gets or creates a circuit breaker for an error code
func (ro *RecoveryOrchestrator) getOrCreateCircuitBreaker(errorCode string) *CircuitBreaker {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	if cb, exists := ro.circuitBreakers[errorCode]; exists {
		return cb
	}

	config := DefaultCircuitBreakerConfig(errorCode)
	cb := NewCircuitBreaker(config)

	// Set state change callback
	cb.SetStateChangeCallback(func(name string, from, to CircuitState) {
		if to == StateOpen {
			ro.mu.Lock()
			ro.metrics.CircuitBreakerTrips++
			ro.mu.Unlock()
		}
	})

	ro.circuitBreakers[errorCode] = cb
	return cb
}

// AddPolicy adds a recovery policy
func (ro *RecoveryOrchestrator) AddPolicy(policy RecoveryPolicy) {
	ro.mu.Lock()
	defer ro.mu.Unlock()
	ro.policies = append(ro.policies, policy)
}

// RegisterFallbackProvider registers a fallback provider
func (ro *RecoveryOrchestrator) RegisterFallbackProvider(provider FallbackProvider) {
	ro.mu.Lock()
	defer ro.mu.Unlock()
	ro.fallbackProviders = append(ro.fallbackProviders, provider)
}

// GetMetrics returns recovery metrics
func (ro *RecoveryOrchestrator) GetMetrics() RecoveryMetrics {
	ro.mu.RLock()
	defer ro.mu.RUnlock()
	return ro.metrics
}

// GetCircuitBreakerMetrics returns all circuit breaker metrics
func (ro *RecoveryOrchestrator) GetCircuitBreakerMetrics() map[string]CircuitBreakerMetrics {
	ro.mu.RLock()
	defer ro.mu.RUnlock()

	metrics := make(map[string]CircuitBreakerMetrics)
	for name, cb := range ro.circuitBreakers {
		metrics[name] = cb.GetMetrics()
	}
	return metrics
}

// Shutdown gracefully shuts down the recovery orchestrator
func (ro *RecoveryOrchestrator) Shutdown() {
	ro.cancel()
}

// initializeDefaultPolicies sets up default recovery policies
func (ro *RecoveryOrchestrator) initializeDefaultPolicies() {
	policies := []RecoveryPolicy{
		{
			ErrorCodePattern: "NETWORK_TIMEOUT_.*",
			Strategy:         StrategyRetry,
			MaxAttempts:      3,
			BackoffStrategy:  BackoffExponentialJitter,
			Timeout:          30 * time.Second,
			Priority:         10,
		},
		{
			ErrorCodePattern: "GITHUB_AUTH_.*",
			Strategy:         StrategyAutoFix,
			MaxAttempts:      1,
			BackoffStrategy:  BackoffFixed,
			Timeout:          10 * time.Second,
			Priority:         15,
		},
		{
			DomainPattern:   "api",
			Strategy:        StrategyCircuitBreaker,
			MaxAttempts:     5,
			BackoffStrategy: BackoffExponential,
			Timeout:         60 * time.Second,
			Priority:        8,
		},
		{
			ErrorCodePattern: "CONFIG_VALIDATION_.*",
			Strategy:         StrategyAutoFix,
			MaxAttempts:      2,
			BackoffStrategy:  BackoffFixed,
			Timeout:          5 * time.Second,
			Priority:         12,
		},
		{
			DomainPattern:   "file",
			Strategy:        StrategyFallback,
			MaxAttempts:     2,
			BackoffStrategy: BackoffLinear,
			Timeout:         15 * time.Second,
			Priority:        7,
		},
	}

	for _, policy := range policies {
		ro.AddPolicy(policy)
	}
}

// matchPattern is a simple pattern matcher (could be enhanced with regex)
func matchPattern(pattern, text string) (bool, error) {
	// For now, just check if pattern is a prefix or contains wildcard
	if pattern == "*" {
		return true, nil
	}

	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(text) >= len(prefix) && text[:len(prefix)] == prefix, nil
	}

	return pattern == text, nil
}
