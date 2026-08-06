//nolint:testpackage // White-box testing needed for internal function access
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

// TestRateLimiterAdaptiveBehavior는 잔량이 줄수록 간격이 벌어지는지 본다.
//
// 실제로 기다려 보는 대신 GetStatus가 알려주는 예상 간격으로 확인한다.
// 예전에는 경우마다 Wait를 세 번씩 불렀는데, 기본값에서 잔량이 10이면 한 번에
// 5분을 잔다(1시간 / 유효잔량 10 = 6분, maxBackoff에서 잘림). 세 경우를 다
// 돌면 15분이 넘어 시험이 시간초과로 죽었고, Go는 꾸러미 하나를 한 프로세스에서
// 돌리므로 같은 꾸러미의 나머지 시험들도 통째로 끌려 내려갔다.
//
// 간격을 실제로 지키는지는 TestRateLimiterRecordsRequestAfterWaiting이
// 짧은 간격으로 확인한다.
func TestRateLimiterAdaptiveBehavior(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// 기준 시각을 심는다. lastRequest가 zero value면 calculateDelay가 "요청한 지
	// 2000년 지났다"고 보고 무조건 0을 돌려주므로 아무 차이도 드러나지 않는다.
	// 기본값에서 첫 Wait는 곧바로 돌아온다.
	if err := rl.Wait(context.Background()); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	// reset까지 1시간을 두면 간격은 대략 1시간 / 유효잔량이 된다.
	//   5000 → 버퍼 10%, 유효 4500 → 800ms
	//    100 → 버퍼  8%, 유효   92 → 39s
	//     10 → 버퍼  5%, 유효   10 → 6분이지만 maxBackoff 5분에서 잘린다
	testCases := []struct {
		name      string
		remaining int
		minDelay  time.Duration
		maxDelay  time.Duration
	}{
		{
			name:      "high remaining requests",
			remaining: 5000,
			minDelay:  time.Millisecond * 500,
			maxDelay:  time.Second * 2,
		},
		{
			name:      "low remaining requests",
			remaining: 100,
			minDelay:  time.Second * 30,
			maxDelay:  time.Second * 60,
		},
		{
			name:      "very low remaining requests",
			remaining: 10,
			minDelay:  time.Minute * 4,
			maxDelay:  time.Minute * 5,
		},
	}

	var previous time.Duration

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl.UpdateRemaining(tc.remaining)
			rl.UpdateResetTime(time.Now().Add(time.Hour))

			_, _, estimated := rl.GetStatus()

			if estimated < tc.minDelay || estimated > tc.maxDelay {
				t.Errorf("잔량 %d의 예상 간격이 [%v, %v] 밖이다: %v",
					tc.remaining, tc.minDelay, tc.maxDelay, estimated)
			}

			// 잔량이 줄면 간격은 반드시 늘어야 한다. 이게 "적응"의 내용이다.
			if estimated <= previous {
				t.Errorf("잔량 %d의 간격 %v가 그 앞(%v)보다 넓지 않다",
					tc.remaining, estimated, previous)
			}

			previous = estimated
		})
	}
}

// TestRateLimiterRecordsRequestAfterWaiting은 기다린 뒤의 시각을 적는지 본다.
//
// 요청은 Wait가 돌아온 다음에 나가므로 그때 시각을 적어야 한다. 예전에는
// 기다리기 전 시각을 적어서, 다음 계산이 기다린 시간만큼을 "요청 이후 흐른
// 시간"으로 세고 간격을 0으로 만들었다. 즉 한 번 쉬고 그다음은 곧바로 나가는,
// 몰아치지 않으려고 두는 장치가 정확히 몰아치는 모양이었다.
func TestRateLimiterRecordsRequestAfterWaiting(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// reset까지 2초를 두면 계산상 간격은 아주 짧지만(2s / 4500), 최소 간격이
	// 100ms(1초 / 초당 10건)라 그 값이 쓰인다.
	rl.UpdateRemaining(5000)
	rl.UpdateResetTime(time.Now().Add(time.Second * 2))

	ctx := context.Background()
	start := time.Now()

	// 첫 번째는 lastRequest가 비어 있어 그냥 통과한다. 간격이 걸리는 것은
	// 두 번째와 세 번째라서 100ms씩 두 번, 200ms쯤 걸려야 한다.
	for range 3 {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("Wait failed: %v", err)
		}
	}

	elapsed := time.Since(start)
	if elapsed < time.Millisecond*180 {
		t.Errorf("세 번째 요청이 최소 간격을 지키지 않았다: %v (>=180ms이어야 한다)", elapsed)
	}
}

// TestRateLimiterDoesNotHoldLockWhileWaiting은 기다리는 동안 잠금을 놓는지 본다.
//
// 예전에는 defer로 Wait가 끝날 때까지 쥐고 있어서, 한 쪽이 5분을 기다리면
// 그동안 UpdateRemaining도 GetStatus도 같이 멈췄다. 응답 헤더로 새 잔량을
// 알려줄 수 없으니 낡은 값으로 계산한 대기를 끝까지 지키게 되고, 진행 상황을
// 보여주는 쪽도 얼어붙는다.
func TestRateLimiterDoesNotHoldLockWhileWaiting(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// 잔량이 없으면 reset까지 통째로 기다린다 -- 여기서는 1시간이다.
	rl.UpdateRemaining(0)
	rl.UpdateResetTime(time.Now().Add(time.Hour))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	waiting := make(chan error, 1)
	go func() { waiting <- rl.Wait(ctx) }()

	// Wait가 잠금을 잡았다 놓고 잠들 틈을 준다.
	time.Sleep(time.Millisecond * 20)

	done := make(chan struct{})

	go func() {
		rl.UpdateRemaining(5000)
		rl.GetStatus()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("기다리는 동안 잠금을 쥐고 있다 -- UpdateRemaining과 GetStatus가 막혔다")
	}

	cancel()

	if err := <-waiting; !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
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

	// Modify state.
	//
	// reset 시각을 지나간 때로 둔다. calculateDelay가 맨 앞에서 0을 돌려주므로
	// 아래 Wait들이 실제로 자지 않는다. 여기서 보려는 것은 Reset이 상태를
	// 되돌리는지지 간격을 지키는지가 아니다.
	rl.UpdateRemaining(100)
	rl.UpdateResetTime(time.Now().Add(-time.Minute))

	// Make some requests to populate history
	ctx := context.Background()
	for range 5 {
		if err := rl.Wait(ctx); err != nil {
			t.Logf("Warning: rate limiter wait failed: %v", err)
		}
	}

	if len(rl.requestHistory) != 5 {
		t.Errorf("Expected 5 recorded requests before reset, got %d", len(rl.requestHistory))
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

	if len(rl.requestHistory) != 0 {
		t.Errorf("Expected empty history after reset, got %d entries", len(rl.requestHistory))
	}
}

func TestRateLimiterRequestHistory(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// reset 시각을 지나간 때로 둬서 Wait가 자지 않게 한다. 여기서 보려는 것은
	// 요청 이력을 남기고 그 이력에서 빈도를 계산하는지다.
	rl.UpdateResetTime(time.Now().Add(-time.Hour))

	ctx := context.Background()

	// Make several requests quickly
	for range 10 {
		err := rl.Wait(ctx)
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		time.Sleep(time.Millisecond * 10)
	}

	if len(rl.requestHistory) != 10 {
		t.Errorf("Expected 10 recorded requests, got %d", len(rl.requestHistory))
	}

	// 10ms 간격으로 열 번 불렀으니 초당 빈도가 잡혀야 한다. 예전 확인은
	// "delay >= 0"이었는데 calculateDelay는 음수를 돌려줄 수 없어서 그 확인은
	// 늘 통과했다 -- 이력이 아예 비어 있어도 마찬가지였다.
	if freq := rl.calculateRecentFrequency(time.Now()); freq <= 0 {
		t.Errorf("Expected a positive request frequency, got %v", freq)
	}
}

func TestRateLimiterMemoryEfficiency(t *testing.T) {
	rl := NewAdaptiveRateLimiter()

	// reset 시각을 지나간 때로 둬서 Wait가 자지 않게 한다. 기본값 그대로면
	// 한 번에 800ms(1시간 / 유효잔량 4500)씩 자므로 200번이면 2분 40초다.
	rl.UpdateResetTime(time.Now().Add(-time.Hour))

	ctx := context.Background()

	// Make many requests to test history cleanup
	for i := range 200 {
		if err := rl.Wait(ctx); err != nil {
			t.Logf("Warning: rate limiter wait failed: %v", err)
		}

		if i%50 == 0 {
			time.Sleep(time.Millisecond) // Small delay to vary timestamps
		}
	}

	// 이력은 요청 수를 따라 자라면 안 된다. cleanHistory가 Wait 앞머리에서
	// 100개로 자르고 그 뒤에 이번 것 하나를 붙이므로 101이 상한이다.
	//
	// 예전 확인은 "delay >= 0"이었는데 calculateDelay는 음수를 돌려줄 수 없어서
	// 이력이 200개로 불어나도 그대로 통과했다.
	if got := len(rl.requestHistory); got > 101 {
		t.Errorf("History grew with request count: %d entries after 200 requests", got)
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

	// reset 시각을 지나간 때로 둬서 대기 없이 잠금과 이력 손질만 재게 한다.
	// 그러지 않으면 벤치마크가 재는 것은 코드가 아니라 time.Sleep이다.
	rl.UpdateResetTime(time.Now().Add(-time.Hour))

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Ignore errors in benchmark to not affect performance measurements
		_ = rl.Wait(ctx)
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
