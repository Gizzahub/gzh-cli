package largescale

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewAdaptiveRateLimiter(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	if rl == nil {
		t.Fatal("Rate limiter should not be nil")
	}

	if rl.remaining <= 0 {
		t.Error("Initial remaining requests should be positive")
	}

	if rl.resetTime.Before(time.Now()) {
		t.Error("Reset time should be in the future")
	}
}

func TestRateLimiterWait(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// With plenty of remaining requests, should not delay
	ctx := context.Background()
	start := time.Now()

	err := rl.Wait(ctx)
	if err != nil {
		t.Fatalf("Wait should not error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > time.Millisecond*100 {
		t.Errorf("Wait took too long with plenty of requests: %v", elapsed)
	}
}

func TestRateLimiterContextCancellation(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// Set up rate limiter to require waiting
	rl.UpdateRemaining(0) // No remaining requests
	rl.UpdateResetTime(time.Now().Add(time.Hour))

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start waiting in goroutine
	errCh := make(chan error, 1)

	go func() {
		errCh <- rl.Wait(ctx)
	}()

	// Cancel after short delay
	time.Sleep(time.Millisecond * 10)
	cancel()

	// Should get cancellation error
	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Error("Wait should return quickly after context cancellation")
	}
}

func TestRateLimiterUpdateRemaining(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// Test updating remaining requests
	rl.UpdateRemaining(100)

	remaining, _, _ := rl.GetStatus()
	if remaining != 100 {
		t.Errorf("Expected remaining=100, got %d", remaining)
	}

	// Test adaptive behavior with low remaining
	rl.UpdateRemaining(50)

	// Should adapt rate limiting behavior
	remaining, _, _ = rl.GetStatus()
	if remaining != 50 {
		t.Errorf("Expected remaining=50, got %d", remaining)
	}
}

func TestRateLimiterUpdateResetTime(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	newResetTime := time.Now().Add(time.Hour * 2)
	rl.UpdateResetTime(newResetTime)

	_, resetTime, _ := rl.GetStatus()
	if !resetTime.Equal(newResetTime) {
		t.Errorf("Expected reset time %v, got %v", newResetTime, resetTime)
	}
}

func TestRateLimiterAdaptiveBehavior(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	testCases := []struct {
		remaining  int
		name       string
		shouldSlow bool
	}{
		{
			remaining:  5000,
			name:       "high remaining requests",
			shouldSlow: false,
		},
		{
			remaining:  100,
			name:       "low remaining requests",
			shouldSlow: true,
		},
		{
			remaining:  10,
			name:       "very low remaining requests",
			shouldSlow: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl.UpdateRemaining(tc.remaining)
			rl.UpdateResetTime(time.Now().Add(time.Hour))

			// Measure delay for multiple requests
			ctx := context.Background()
			start := time.Now()

			for i := 0; i < 3; i++ {
				err := rl.Wait(ctx)
				if err != nil {
					t.Fatalf("Wait failed: %v", err)
				}
			}

			elapsed := time.Since(start)

			if tc.shouldSlow && elapsed < time.Millisecond*100 {
				t.Errorf("Expected significant delay with %d remaining, got %v", tc.remaining, elapsed)
			}
		})
	}
}

func TestRateLimiterEstimateTimeToCompletion(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	testCases := []struct {
		remaining      int
		requestsNeeded int
		name           string
		expectLong     bool
	}{
		{
			remaining:      5000,
			requestsNeeded: 100,
			name:           "plenty of requests available",
			expectLong:     false,
		},
		{
			remaining:      50,
			requestsNeeded: 1000,
			name:           "need multiple cycles",
			expectLong:     true,
		},
		{
			remaining:      0,
			requestsNeeded: 10,
			name:           "no remaining requests",
			expectLong:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl.UpdateRemaining(tc.remaining)
			rl.UpdateResetTime(time.Now().Add(time.Hour))

			estimate := rl.EstimateTimeToCompletion(tc.requestsNeeded)

			if tc.expectLong && estimate < time.Minute {
				t.Errorf("Expected long estimate for %s, got %v", tc.name, estimate)
			}

			if !tc.expectLong && estimate > time.Minute*10 {
				t.Errorf("Expected short estimate for %s, got %v", tc.name, estimate)
			}
		})
	}
}

func TestRateLimiterConfiguration(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// Test setting configuration
	rl.SetConfiguration(20, 0.05, false)

	// Configuration should affect behavior
	// This is mostly testing that the function doesn't panic
	// and that the values are stored correctly

	ctx := context.Background()

	err := rl.Wait(ctx)
	if err != nil {
		t.Errorf("Wait should not error after configuration: %v", err)
	}
}

func TestRateLimiterReset(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// Modify state
	rl.UpdateRemaining(100)
	rl.UpdateResetTime(time.Now().Add(time.Minute))

	// Make some requests to populate history
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		rl.Wait(ctx)
	}

	// Reset should restore defaults
	rl.Reset()

	remaining, resetTime, _ := rl.GetStatus()

	if remaining != 5000 {
		t.Errorf("Expected remaining=5000 after reset, got %d", remaining)
	}

	if resetTime.Before(time.Now().Add(time.Minute * 30)) {
		t.Error("Reset time should be well in the future after reset")
	}
}

func TestRateLimiterRequestHistory(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	ctx := context.Background()

	// Make several requests quickly
	for i := 0; i < 10; i++ {
		err := rl.Wait(ctx)
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		time.Sleep(time.Millisecond * 10)
	}

	// The rate limiter should adapt to the request frequency
	// This is mostly testing that history tracking doesn't panic
	// and that it affects delay calculation

	_, _, delay := rl.GetStatus()

	// With recent requests, there should be some delay
	if delay < 0 {
		t.Error("Delay should not be negative")
	}
}

func TestRateLimiterMemoryEfficiency(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	ctx := context.Background()

	// Make many requests to test history cleanup
	for i := 0; i < 200; i++ {
		rl.Wait(ctx)

		if i%50 == 0 {
			time.Sleep(time.Millisecond) // Small delay to vary timestamps
		}
	}

	// History should not grow unboundedly
	// This is hard to test directly, but the test should not run out of memory

	_, _, delay := rl.GetStatus()
	if delay < 0 {
		t.Error("Delay should not be negative after many requests")
	}
}

func TestRateLimiterEdgeCases(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// Test with reset time in the past
	rl.UpdateResetTime(time.Now().Add(-time.Hour))
	rl.UpdateRemaining(0)

	ctx := context.Background()
	start := time.Now()

	err := rl.Wait(ctx)
	if err != nil {
		t.Fatalf("Wait should not error with past reset time: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > time.Millisecond*100 {
		t.Errorf("Should not delay when reset time is in past: %v", elapsed)
	}
}

// Benchmark tests

func BenchmarkRateLimiterWait(b *testing.B) {
	rl := NewAdaptiveRateLimiter()
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rl.Wait(ctx)
	}
}

func BenchmarkRateLimiterUpdateRemaining(b *testing.B) {
	rl := NewAdaptiveRateLimiter()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rl.UpdateRemaining(i % 5000)
	}
}

func BenchmarkRateLimiterEstimateCompletion(b *testing.B) {
	rl := NewAdaptiveRateLimiter()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rl.EstimateTimeToCompletion(1000)
	}
}
