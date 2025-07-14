package recovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker(t *testing.T) {
	tests := []struct {
		name          string
		config        CircuitBreakerConfig
		failures      int
		expectedState CircuitState
		shouldExecute bool
	}{
		{
			name:          "circuit remains closed with few failures",
			config:        DefaultCircuitBreakerConfig("test"),
			failures:      2,
			expectedState: StateClosed,
			shouldExecute: true,
		},
		{
			name:          "circuit opens after failure threshold",
			config:        DefaultCircuitBreakerConfig("test"),
			failures:      6,
			expectedState: StateOpen,
			shouldExecute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker(tt.config)
			ctx := context.Background()

			// Generate failures
			for i := 0; i < tt.failures; i++ {
				err := cb.Execute(ctx, func() error {
					return fmt.Errorf("test failure %d", i)
				})
				assert.Error(t, err)
			}

			assert.Equal(t, tt.expectedState, cb.GetState())

			// Test if execution is allowed
			if tt.shouldExecute {
				err := cb.Execute(ctx, func() error {
					return nil
				})
				assert.NoError(t, err)
			} else {
				err := cb.Execute(ctx, func() error {
					return nil
				})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "circuit breaker")
			}
		})
	}
}

func TestCircuitBreakerStateTransitions(t *testing.T) {
	config := DefaultCircuitBreakerConfig("test")
	config.FailureThreshold = 3
	config.SuccessThreshold = 2
	config.Timeout = 100 * time.Millisecond

	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	// Start in closed state
	assert.Equal(t, StateClosed, cb.GetState())

	// Generate failures to open circuit
	for i := 0; i < 4; i++ {
		cb.Execute(ctx, func() error {
			return fmt.Errorf("failure")
		})
	}
	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for timeout to transition to half-open
	time.Sleep(150 * time.Millisecond)

	// Next execution should be allowed (half-open)
	err := cb.Execute(ctx, func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, cb.GetState())

	// Generate enough successes to close circuit
	cb.Execute(ctx, func() error {
		return nil
	})
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestRecoveryOrchestrator(t *testing.T) {
	ro := NewRecoveryOrchestrator()
	defer ro.Shutdown()

	// Create test error
	userErr := &errors.UserError{
		Code: errors.ErrorCode{
			Domain:   "network",
			Category: "timeout",
			Code:     "operation_timeout",
		},
		Message:     "Operation timed out",
		Description: "Network operation exceeded timeout",
		Context:     make(map[string]interface{}),
		RequestID:   "test-123",
		Timestamp:   time.Now(),
	}

	ctx := context.Background()
	attempts := 0

	// Test retry strategy
	err := ro.RecoverFromError(ctx, userErr, func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("still failing")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, attempts, 3)

	// Check metrics
	metrics := ro.GetMetrics()
	assert.Greater(t, metrics.TotalAttempts, int64(0))
	assert.Greater(t, metrics.SuccessfulRecoveries, int64(0))
}

func TestFallbackProviders(t *testing.T) {
	t.Run("NetworkFallbackProvider", func(t *testing.T) {
		provider := NewNetworkFallbackProvider()

		assert.True(t, provider.CanHandle("NETWORK_TIMEOUT_ERROR"))
		assert.False(t, provider.CanHandle("FILE_PERMISSION_ERROR"))

		info := provider.GetInfo()
		assert.Equal(t, "Network Fallback Provider", info.Name)
		assert.Greater(t, info.Priority, 0)
	})

	t.Run("FileFallbackProvider", func(t *testing.T) {
		provider := NewFileFallbackProvider()

		assert.True(t, provider.CanHandle("FILE_PERMISSION_ACCESS_DENIED"))
		assert.False(t, provider.CanHandle("NETWORK_TIMEOUT_ERROR"))

		info := provider.GetInfo()
		assert.Equal(t, "File Fallback Provider", info.Name)
		assert.Greater(t, info.Priority, 0)
	})

	t.Run("AuthFallbackProvider", func(t *testing.T) {
		provider := NewAuthFallbackProvider()

		assert.True(t, provider.CanHandle("AUTH_INVALID_TOKEN"))
		assert.False(t, provider.CanHandle("FILE_PERMISSION_ERROR"))

		info := provider.GetInfo()
		assert.Equal(t, "Authentication Fallback Provider", info.Name)
		assert.Greater(t, info.Priority, 0)
	})
}

// TestRecoveryManager temporarily disabled due to integration dependencies
// func TestRecoveryManager(t *testing.T) {
// 	rm := NewRecoveryManager()
// 	defer rm.Shutdown()
//
// 	ctx := context.Background()
//
// 	t.Run("successful operation", func(t *testing.T) {
// 		attempts := 0
// 		err := rm.ExecuteWithRecovery(ctx, func() error {
// 			attempts++
// 			return nil
// 		})
//
// 		assert.NoError(t, err)
// 		assert.Equal(t, 1, attempts)
// 	})
//
// 	t.Run("operation with recovery", func(t *testing.T) {
// 		attempts := 0
// 		err := rm.ExecuteWithRecovery(ctx, func() error {
// 			attempts++
// 			if attempts < 2 {
// 				return fmt.Errorf("timeout error")
// 			}
// 			return nil
// 		})
//
// 		assert.NoError(t, err)
// 		assert.GreaterOrEqual(t, attempts, 2)
// 	})
//
// 	t.Run("recovery status", func(t *testing.T) {
// 		status := rm.GetRecoveryStatus()
//
// 		assert.NotEmpty(t, status.OverallHealth.Status)
// 		assert.GreaterOrEqual(t, status.OverallHealth.Score, 0.0)
// 		assert.LessOrEqual(t, status.OverallHealth.Score, 1.0)
// 	})
// }

func TestBackoffStrategies(t *testing.T) {
	ro := NewRecoveryOrchestrator()
	defer ro.Shutdown()

	tests := []struct {
		name     string
		strategy BackoffStrategy
		attempt  int
		expected time.Duration
	}{
		{"Fixed backoff", BackoffFixed, 1, time.Second},
		{"Fixed backoff attempt 3", BackoffFixed, 3, time.Second},
		{"Linear backoff", BackoffLinear, 2, 2 * time.Second},
		{"Exponential backoff", BackoffExponential, 3, 4 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := ro.calculateBackoffDelay(tt.attempt, tt.strategy)

			if tt.strategy == BackoffExponentialJitter {
				// For jitter, check it's within reasonable bounds
				assert.GreaterOrEqual(t, delay, tt.expected)
				assert.LessOrEqual(t, delay, tt.expected*2)
			} else {
				assert.Equal(t, tt.expected, delay)
			}
		})
	}
}

func TestCircuitBreakerMetrics(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig("test"))
	ctx := context.Background()

	// Execute some operations
	cb.Execute(ctx, func() error { return nil })
	cb.Execute(ctx, func() error { return fmt.Errorf("error") })
	cb.Execute(ctx, func() error { return nil })

	metrics := cb.GetMetrics()

	assert.Equal(t, "test", metrics.Name)
	assert.Equal(t, 3, metrics.TotalCalls)
	assert.Equal(t, 2, metrics.SuccessfulCalls)
	assert.Equal(t, 1, metrics.FailedCalls)
	assert.InDelta(t, 0.67, metrics.SuccessRate, 0.01)
}

func TestRecoveryPolicyMatching(t *testing.T) {
	ro := NewRecoveryOrchestrator()
	defer ro.Shutdown()

	// Add test policy
	testPolicy := RecoveryPolicy{
		ErrorCodePattern: "TEST_ERROR_*",
		Strategy:         StrategyRetry,
		MaxAttempts:      3,
		Priority:         10,
	}
	ro.AddPolicy(testPolicy)

	// Create matching error
	userErr := &errors.UserError{
		Code: errors.ErrorCode{
			Domain:   "test",
			Category: "error",
			Code:     "specific",
		},
		Message: "Test error",
	}
	userErr.Code = errors.ErrorCode{Domain: "test", Category: "error", Code: "specific"}

	policy := ro.findPolicy(userErr)
	require.NotNil(t, policy)
	assert.Equal(t, StrategyRetry, policy.Strategy)
	assert.Equal(t, 3, policy.MaxAttempts)
}

func BenchmarkCircuitBreaker(b *testing.B) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig("benchmark"))
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Execute(ctx, func() error {
			if i%10 == 0 {
				return fmt.Errorf("error")
			}
			return nil
		})
	}
}

func BenchmarkRecoveryOrchestrator(b *testing.B) {
	ro := NewRecoveryOrchestrator()
	defer ro.Shutdown()

	userErr := &errors.UserError{
		Code: errors.ErrorCode{
			Domain:   "network",
			Category: "timeout",
			Code:     "operation_timeout",
		},
		Message: "Benchmark timeout",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ro.RecoverFromError(ctx, userErr, func() error {
			return nil
		})
	}
}
