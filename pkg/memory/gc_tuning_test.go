package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGCTuner(t *testing.T) {
	config := DefaultGCConfig()
	config.ProfilingEnabled = false // Disable profiling for tests
	config.ForceGCInterval = 0      // Disable force GC for tests

	tuner := NewGCTuner(config)
	require.NotNil(t, tuner)

	t.Run("Start and Stop", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := tuner.Start(ctx)
		assert.NoError(t, err)

		// Give it a moment to start
		time.Sleep(100 * time.Millisecond)

		tuner.Stop()
	})

	t.Run("CreatePool", func(t *testing.T) {
		pool := tuner.CreatePool("test-pool",
			func() interface{} {
				return make([]byte, 0, 1024)
			},
			func(obj interface{}) {
				// Note: this doesn't actually reset the slice, it's just for testing
				if _, ok := obj.([]byte); ok {
					_ = ok // Reset operation (demo only)
				}
			})

		require.NotNil(t, pool)
		assert.Equal(t, "test-pool", pool.name)
	})

	t.Run("PoolOperations", func(t *testing.T) {
		pool := tuner.CreatePool("test-pool-ops",
			func() interface{} {
				return make([]int, 0, 10)
			},
			func(obj interface{}) {
				// Note: this doesn't actually reset the slice, it's just for testing
				if _, ok := obj.([]int); ok {
					_ = ok // Reset operation (demo only)
				}
			})

		// Test Get and Put
		obj1 := pool.Get()
		assert.NotNil(t, obj1)

		slice, ok := obj1.([]int)
		require.True(t, ok)
		assert.Equal(t, 0, len(slice))
		assert.Equal(t, 10, cap(slice))

		pool.Put(obj1)

		// Get again - should reuse the object
		obj2 := pool.Get()
		assert.NotNil(t, obj2)

		pool.Put(obj2)

		// Check stats
		stats := pool.GetStats()
		assert.Equal(t, int64(2), stats.Gets)
		assert.Equal(t, int64(2), stats.Puts)
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := tuner.GetStats()
		assert.NotNil(t, stats)
		assert.True(t, stats.LastUpdate.Before(time.Now()) || stats.LastUpdate.Equal(time.Time{}))
	})

	t.Run("ForceGC", func(t *testing.T) {
		initialStats := tuner.GetStats()

		tuner.ForceGC()

		finalStats := tuner.GetStats()
		assert.Equal(t, initialStats.ForceGCCount+1, finalStats.ForceGCCount)
	})

	t.Run("OptimizeForWorkload", func(t *testing.T) {
		// Test different workload optimizations
		workloads := []string{"low-latency", "high-throughput", "memory-constrained", "balanced"}

		for _, workload := range workloads {
			tuner.OptimizeForWorkload(workload)
			// Just verify it doesn't panic
		}
	})

	t.Run("ClearAllPools", func(t *testing.T) {
		// Create a pool and add some data
		pool := tuner.CreatePool("clear-test",
			func() interface{} {
				return make([]string, 0, 5)
			},
			nil)

		obj := pool.Get()
		pool.Put(obj)

		// Clear all pools
		tuner.ClearAllPools()

		// Pool should still work but with fresh objects
		newObj := pool.Get()
		assert.NotNil(t, newObj)
	})
}

func TestMemoryPool(t *testing.T) {
	pool := &MemoryPool{
		name: "test",
		newFunc: func() interface{} {
			return make([]int, 0, 10)
		},
		resetFunc: func(obj interface{}) {
			// Note: this doesn't actually reset the slice, it's just for testing
			if _, ok := obj.([]int); ok {
				_ = ok // Reset operation (demo only)
			}
		},
	}

	// Initialize the sync.Pool
	pool.pool.New = func() interface{} {
		pool.mu.Lock()
		pool.stats.News++
		pool.mu.Unlock()
		return pool.newFunc()
	}

	t.Run("BasicOperations", func(t *testing.T) {
		// Get object from pool
		obj := pool.Get()
		assert.NotNil(t, obj)

		slice, ok := obj.([]int)
		require.True(t, ok)
		assert.Equal(t, 0, len(slice))

		// Modify the slice
		slice = append(slice, 1, 2, 3)
		assert.Equal(t, 3, len(slice))

		// Put back to pool
		pool.Put(slice)

		// Get stats
		stats := pool.GetStats()
		assert.Equal(t, int64(1), stats.Gets)
		assert.Equal(t, int64(1), stats.Puts)
		assert.Equal(t, int64(1), stats.News)
		assert.Equal(t, int64(1), stats.Resets)
	})

	t.Run("HitRate", func(t *testing.T) {
		// Clear stats
		pool.stats = PoolStats{}

		// First get creates new object
		obj1 := pool.Get()
		pool.Put(obj1)

		// Second get should reuse
		obj2 := pool.Get()
		pool.Put(obj2)

		stats := pool.GetStats()
		assert.Equal(t, int64(2), stats.Gets)
		assert.True(t, stats.HitRate > 0) // Should have some hit rate
	})
}

func TestGCConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultGCConfig()

		assert.Equal(t, 100, config.GCPercent)
		assert.Equal(t, int64(0), config.MemoryLimit)
		assert.False(t, config.ProfilingEnabled)
		assert.True(t, config.ObjectPoolingEnabled)
		assert.Equal(t, 0.8, config.MemoryPressureThreshold)
	})
}

func BenchmarkMemoryPool(b *testing.B) {
	pool := &MemoryPool{
		name: "benchmark",
		newFunc: func() interface{} {
			return make([]byte, 0, 1024)
		},
		resetFunc: func(obj interface{}) {
			if _, ok := obj.([]byte); ok {
				// Note: slice = slice[:0] doesn't actually reset, it's just demo
				_ = ok
			}
		},
	}

	pool.pool.New = func() interface{} {
		return pool.newFunc()
	}

	b.Run("GetPut", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			obj := pool.Get()
			pool.Put(obj)
		}
	})

	b.Run("GetModifyPut", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			obj := pool.Get()
			slice := obj.([]byte)
			slice = append(slice, byte(i%256))
			pool.Put(slice)
		}
	})
}

func BenchmarkWithoutPool(b *testing.B) {
	b.Run("DirectAllocation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]byte, 0, 1024)
			slice = append(slice, byte(i%256))
			_ = slice
		}
	})
}
