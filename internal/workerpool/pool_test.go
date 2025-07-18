package workerpool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_BasicFunctionality(t *testing.T) {
	config := Config{
		WorkerCount: 2,
		BufferSize:  4,
		Timeout:     time.Second,
	}

	pool := New[int](config)
	require.NotNil(t, pool)

	err := pool.Start()
	require.NoError(t, err)

	defer pool.Stop()

	// Submit a simple job
	processFn := func(ctx context.Context, data int) error {
		return nil
	}

	err = pool.Submit(1, processFn)
	assert.NoError(t, err)

	// Get result
	select {
	case result := <-pool.Results():
		assert.Equal(t, 1, result.Data)
		assert.NoError(t, result.Error)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for result")
	}
}

func TestPool_ConcurrentProcessing(t *testing.T) {
	config := Config{
		WorkerCount: 3,
		BufferSize:  10,
		Timeout:     time.Second,
	}

	pool := New[int](config)
	require.NotNil(t, pool)

	err := pool.Start()
	require.NoError(t, err)

	defer pool.Stop()

	var processedCount int64

	processFn := func(ctx context.Context, data int) error {
		time.Sleep(100 * time.Millisecond) // Simulate work
		atomic.AddInt64(&processedCount, 1)

		return nil
	}

	// Submit multiple jobs
	jobCount := 6
	for i := 0; i < jobCount; i++ {
		err = pool.Submit(i, processFn)
		require.NoError(t, err)
	}

	// Collect results
	results := make([]Result[int], 0, jobCount)
	for i := 0; i < jobCount; i++ {
		select {
		case result := <-pool.Results():
			results = append(results, result)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for results")
		}
	}

	assert.Len(t, results, jobCount)
	assert.Equal(t, int64(jobCount), atomic.LoadInt64(&processedCount))

	// Verify all jobs completed successfully
	for _, result := range results {
		assert.NoError(t, result.Error)
	}
}

func TestPool_ErrorHandling(t *testing.T) {
	config := Config{
		WorkerCount: 2,
		BufferSize:  4,
		Timeout:     time.Second,
	}

	pool := New[int](config)
	require.NotNil(t, pool)

	err := pool.Start()
	require.NoError(t, err)

	defer pool.Stop()

	// Submit jobs that will fail
	processFn := func(ctx context.Context, data int) error {
		if data%2 == 0 {
			return errors.New("simulated error")
		}

		return nil
	}

	jobCount := 4
	for i := 0; i < jobCount; i++ {
		err = pool.Submit(i, processFn)
		require.NoError(t, err)
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < jobCount; i++ {
		select {
		case result := <-pool.Results():
			if result.Error != nil {
				errorCount++
			} else {
				successCount++
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for results")
		}
	}

	assert.Equal(t, 2, successCount) // Odd numbers should succeed
	assert.Equal(t, 2, errorCount)   // Even numbers should fail
}

func TestPool_DoubleStart(t *testing.T) {
	config := DefaultConfig()
	pool := New[int](config)

	err := pool.Start()
	require.NoError(t, err)

	defer pool.Stop()

	// Second start should fail
	err = pool.Start()
	assert.Error(t, err)
}

func TestPool_SubmitWithoutStart(t *testing.T) {
	config := DefaultConfig()
	pool := New[int](config)

	processFn := func(ctx context.Context, data int) error {
		return nil
	}

	err := pool.Submit(1, processFn)
	assert.Error(t, err)
}

func TestProcessBatch(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	config := Config{
		WorkerCount: 2,
		BufferSize:  4,
		Timeout:     time.Second,
	}

	processFn := func(ctx context.Context, data int) error {
		time.Sleep(50 * time.Millisecond) // Simulate work
		return nil
	}

	ctx := context.Background()
	results, err := ProcessBatch(ctx, items, config, processFn)

	require.NoError(t, err)
	assert.Len(t, results, len(items))

	// All results should be successful
	for _, result := range results {
		assert.NoError(t, result.Error)
	}
}

func TestProcessBatch_WithContext(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	config := Config{
		WorkerCount: 2,
		BufferSize:  4,
		Timeout:     time.Second,
	}

	processFn := func(ctx context.Context, data int) error {
		time.Sleep(200 * time.Millisecond) // Simulate slow work
		return nil
	}

	// Context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	results, err := ProcessBatch(ctx, items, config, processFn)

	// Should get context cancellation error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")

	// May have some results if they completed before timeout
	assert.True(t, len(results) <= len(items))
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.WorkerCount > 0)
	assert.True(t, config.BufferSize > 0)
	assert.True(t, config.Timeout > 0)
}

func BenchmarkPool_Processing(b *testing.B) {
	config := Config{
		WorkerCount: 4,
		BufferSize:  10,
		Timeout:     time.Second,
	}

	pool := New[int](config)

	require.NoError(b, pool.Start())
	defer pool.Stop()

	processFn := func(ctx context.Context, data int) error {
		// Simulate some CPU work
		sum := 0
		for i := 0; i < 1000; i++ {
			sum += i
		}

		return nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := pool.Submit(1, processFn)
			if err != nil {
				b.Error(err)
				return
			}

			// Consume result
			select {
			case <-pool.Results():
			case <-time.After(time.Second):
				b.Error("timeout")
				return
			}
		}
	})
}
