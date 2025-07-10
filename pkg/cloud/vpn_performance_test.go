package cloud

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVPNPerformanceOptimizer_ConnectionPooling tests connection pooling performance
func TestVPNPerformanceOptimizer_ConnectionPooling(t *testing.T) {
	config := &OptimizationConfig{
		EnableConnectionPooling: true,
		ConnectionPoolSize:      10,
		EnableMetrics:           true,
	}

	optimizer := NewVPNPerformanceOptimizer(config)
	pool := optimizer.connectionPool

	// Test connection creation and reuse
	connFactory := func() *VPNConnection {
		return &VPNConnection{
			Name:     "test-conn",
			Type:     "wireguard",
			Server:   "127.0.0.1",
			Priority: 100,
		}
	}

	// First retrieval should create new connection
	conn1 := pool.GetConnection("test-conn", connFactory)
	assert.NotNil(t, conn1)
	assert.Equal(t, 1, pool.created)
	assert.Equal(t, 0, pool.recycled)

	// Return to pool
	pool.ReturnConnection("test-conn")

	// Second retrieval should reuse connection
	conn2 := pool.GetConnection("test-conn", connFactory)
	assert.NotNil(t, conn2)
	assert.Equal(t, 1, pool.created)
	assert.Equal(t, 1, pool.recycled)

	// Verify it's the same connection
	assert.Equal(t, conn1.Connection, conn2.Connection)
	assert.Equal(t, 2, conn2.UseCount)
}

// TestVPNPerformanceOptimizer_CacheManager tests result caching
func TestVPNPerformanceOptimizer_CacheManager(t *testing.T) {
	cacheManager := NewCacheManager(1*time.Second, 100)

	// Test cache miss
	_, found := cacheManager.Get("key1")
	assert.False(t, found)

	// Test cache set and hit
	cacheManager.Set("key1", "value1")
	value, found := cacheManager.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Test cache expiry
	cacheManager.Set("key2", "value2")
	time.Sleep(1100 * time.Millisecond) // Wait for expiry
	_, found = cacheManager.Get("key2")
	assert.False(t, found)

	// Test hit rate calculation
	hitRate := cacheManager.GetHitRate()
	assert.Greater(t, hitRate, 0.0)
	assert.LessOrEqual(t, hitRate, 1.0)
}

// TestVPNPerformanceOptimizer_MemoryOptimization tests memory usage optimization
func TestVPNPerformanceOptimizer_MemoryOptimization(t *testing.T) {
	memOptimizer := NewMemoryOptimizer(100 * time.Millisecond)

	// Test object pool usage
	factory := func() interface{} {
		return &VPNConnection{}
	}

	// Get and put objects
	obj1 := memOptimizer.GetObject("vpn_conn", factory)
	assert.NotNil(t, obj1)

	memOptimizer.PutObject("vpn_conn", obj1)

	// Get again - should reuse from pool
	obj2 := memOptimizer.GetObject("vpn_conn", factory)
	assert.NotNil(t, obj2)

	// Test cleanup trigger
	memOptimizer.TriggerCleanup()
	assert.False(t, memOptimizer.lastCleanup.IsZero())
}

// TestVPNPerformanceOptimizer_BatchProcessing tests batch processing
func TestVPNPerformanceOptimizer_BatchProcessing(t *testing.T) {
	batchProcessor := NewBatchProcessor(2) // Smaller batch size for faster testing

	// Start batch processing
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		for {
			select {
			case batch := <-batchProcessor.processingCh:
				batchProcessor.processBatch(batch)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Add operations to batch
	resultCh := make(chan interface{}, 1)
	errorCh := make(chan error, 1)

	// Add first operation
	op1 := BatchOperation{
		Type:       "status_check",
		Data:       "test_data_1",
		ResultChan: resultCh,
		ErrorChan:  errorCh,
	}
	batchProcessor.AddOperation(op1)

	// Add second operation to trigger batch processing
	op2 := BatchOperation{
		Type:       "status_check",
		Data:       "test_data_2",
		ResultChan: make(chan interface{}, 1),
		ErrorChan:  make(chan error, 1),
	}
	batchProcessor.AddOperation(op2)

	// Wait for result
	select {
	case result := <-resultCh:
		assert.NotNil(t, result)
	case err := <-errorCh:
		t.Fatalf("Unexpected error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Batch processing timeout")
	}
}

// TestOptimizedVPNManager_Performance tests overall performance improvements
func TestOptimizedVPNManager_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create base manager
	baseManager := NewVPNManager()

	// Create optimizer with all features enabled
	config := &OptimizationConfig{
		EnableConnectionPooling:  true,
		ConnectionPoolSize:       20,
		EnableBatchProcessing:    true,
		BatchSize:                5,
		EnableMemoryOptimization: true,
		MemoryCleanupInterval:    1 * time.Second,
		EnableResultCaching:      true,
		CacheTTL:                 5 * time.Second,
		EnableMetrics:            true,
	}

	optimizer := NewVPNPerformanceOptimizer(config)
	optimizedManager := optimizer.OptimizeVPNManager(baseManager)

	// Add test connections
	const numConnections = 100
	for i := 0; i < numConnections; i++ {
		conn := &VPNConnection{
			Name:     fmt.Sprintf("perf-test-%d", i),
			Type:     "wireguard",
			Server:   "127.0.0.1",
			Priority: i,
		}
		err := optimizedManager.AddVPNConnection(conn)
		require.NoError(t, err)
	}

	// Benchmark status checks with caching
	start := time.Now()
	for i := 0; i < 50; i++ {
		status := optimizedManager.GetConnectionStatus()
		assert.Equal(t, numConnections, len(status))
	}
	cachedDuration := time.Since(start)

	// Benchmark status checks without caching (direct base manager)
	start = time.Now()
	for i := 0; i < 50; i++ {
		status := baseManager.GetConnectionStatus()
		assert.Equal(t, numConnections, len(status))
	}
	uncachedDuration := time.Since(start)

	// Optimized version should be faster due to caching
	assert.Less(t, cachedDuration, uncachedDuration)
	t.Logf("Cached duration: %v, Uncached duration: %v, Speed improvement: %.2fx",
		cachedDuration, uncachedDuration, float64(uncachedDuration)/float64(cachedDuration))

	// Check cache hit rate
	hitRate := optimizer.cacheManager.GetHitRate()
	assert.Greater(t, hitRate, 0.9) // Should have high hit rate
	t.Logf("Cache hit rate: %.2f%%", hitRate*100)

	// Check performance metrics
	metrics := optimizer.metrics.GetMetrics()
	t.Logf("Performance metrics: %+v", metrics)
}

// TestVPNPerformanceOptimizer_MemoryUsage tests memory usage optimization
func TestVPNPerformanceOptimizer_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	// Force garbage collection before test
	runtime.GC()
	runtime.GC()

	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Create optimizer with memory optimization
	config := &OptimizationConfig{
		EnableMemoryOptimization: true,
		MemoryCleanupInterval:    100 * time.Millisecond,
		EnableMetrics:            true,
	}

	optimizer := NewVPNPerformanceOptimizer(config)
	baseManager := NewVPNManager()
	optimizedManager := optimizer.OptimizeVPNManager(baseManager)

	// Create many connections using memory optimization
	const numConnections = 1000
	for i := 0; i < numConnections; i++ {
		conn := &VPNConnection{
			Name:     fmt.Sprintf("mem-test-%d", i),
			Type:     "openvpn",
			Server:   "127.0.0.1",
			Priority: i,
		}
		err := optimizedManager.AddVPNConnection(conn)
		require.NoError(t, err)

		// Trigger periodic cleanup
		if i%100 == 0 {
			optimizer.memoryOptimizer.TriggerCleanup()
		}
	}

	// Force garbage collection after test
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	// Verify memory usage is reasonable
	memUsed := memAfter.Alloc - memBefore.Alloc
	memPerConnection := memUsed / numConnections

	t.Logf("Memory used: %d bytes, Per connection: %d bytes", memUsed, memPerConnection)
	assert.Less(t, memPerConnection, uint64(10*1024)) // Should be less than 10KB per connection
}

// TestVPNPerformanceOptimizer_ConcurrentAccess tests concurrent access performance
func TestVPNPerformanceOptimizer_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent access test in short mode")
	}

	config := &OptimizationConfig{
		EnableConnectionPooling:  true,
		ConnectionPoolSize:       50,
		EnableBatchProcessing:    true,
		BatchSize:                10,
		EnableMemoryOptimization: true,
		MemoryCleanupInterval:    1 * time.Second,
		EnableResultCaching:      true,
		CacheTTL:                 5 * time.Second,
		EnableMetrics:            true,
	}

	optimizer := NewVPNPerformanceOptimizer(config)
	baseManager := NewVPNManager()
	optimizedManager := optimizer.OptimizeVPNManager(baseManager)

	// Add initial connections
	for i := 0; i < 50; i++ {
		conn := &VPNConnection{
			Name:     fmt.Sprintf("concurrent-test-%d", i),
			Type:     "wireguard",
			Server:   "127.0.0.1",
			Priority: i,
		}
		err := optimizedManager.AddVPNConnection(conn)
		require.NoError(t, err)
	}

	const numGoroutines = 20
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	errorCh := make(chan error, numGoroutines*operationsPerGoroutine)

	start := time.Now()

	// Start concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Mix of operations
				switch j % 3 {
				case 0:
					// Status check (should benefit from caching)
					status := optimizedManager.GetConnectionStatus()
					if len(status) == 0 {
						errorCh <- fmt.Errorf("empty status in goroutine %d", goroutineID)
					}
				case 1:
					// Get active connections
					active := optimizedManager.GetActiveConnections()
					_ = active // Use the result
				case 2:
					// Add/remove connection (should benefit from pooling)
					tempConn := &VPNConnection{
						Name:     fmt.Sprintf("temp-%d-%d", goroutineID, j),
						Type:     "wireguard",
						Server:   "127.0.0.1",
						Priority: j,
					}
					if err := optimizedManager.AddVPNConnection(tempConn); err != nil {
						errorCh <- fmt.Errorf("failed to add connection: %v", err)
					}
					if err := optimizedManager.RemoveVPNConnection(tempConn.Name); err != nil {
						errorCh <- fmt.Errorf("failed to remove connection: %v", err)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorCh)

	duration := time.Since(start)
	totalOperations := numGoroutines * operationsPerGoroutine

	// Check for errors
	var errorCount int
	for err := range errorCh {
		t.Errorf("Concurrent operation error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Concurrent operations test failed with %d errors", errorCount)
	}

	// Calculate performance metrics
	opsPerSecond := float64(totalOperations) / duration.Seconds()
	t.Logf("Concurrent operations completed: %d operations in %v (%.2f ops/sec)",
		totalOperations, duration, opsPerSecond)

	// Performance threshold - should be efficient
	assert.Greater(t, opsPerSecond, 5000.0) // Should handle at least 5000 ops/sec

	// Check cache performance
	hitRate := optimizer.cacheManager.GetHitRate()
	t.Logf("Cache hit rate: %.2f%%", hitRate*100)
	assert.Greater(t, hitRate, 0.8) // Should have high hit rate

	// Check final metrics
	metrics := optimizer.metrics.GetMetrics()
	t.Logf("Final performance metrics: %+v", metrics)
}

// BenchmarkVPNPerformanceOptimizer_StatusChecks benchmarks status check performance
func BenchmarkVPNPerformanceOptimizer_StatusChecks(b *testing.B) {
	config := &OptimizationConfig{
		EnableResultCaching: true,
		CacheTTL:            5 * time.Second,
		EnableMetrics:       true,
	}

	optimizer := NewVPNPerformanceOptimizer(config)
	baseManager := NewVPNManager()
	optimizedManager := optimizer.OptimizeVPNManager(baseManager)

	// Add test connections
	for i := 0; i < 100; i++ {
		conn := &VPNConnection{
			Name:     fmt.Sprintf("bench-%d", i),
			Type:     "wireguard",
			Server:   "127.0.0.1",
			Priority: i,
		}
		optimizedManager.AddVPNConnection(conn)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		status := optimizedManager.GetConnectionStatus()
		_ = status // Use the result
	}
}

// BenchmarkVPNPerformanceOptimizer_ConnectionPooling benchmarks connection pooling
func BenchmarkVPNPerformanceOptimizer_ConnectionPooling(b *testing.B) {
	optimizer := NewVPNPerformanceOptimizer(&OptimizationConfig{
		EnableConnectionPooling: true,
		ConnectionPoolSize:      100,
	})

	pool := optimizer.connectionPool
	connFactory := func() *VPNConnection {
		return &VPNConnection{
			Name:     "bench-conn",
			Type:     "wireguard",
			Server:   "127.0.0.1",
			Priority: 100,
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		conn := pool.GetConnection("bench-conn", connFactory)
		_ = conn // Use the result
		pool.ReturnConnection("bench-conn")
	}
}

// BenchmarkVPNPerformanceOptimizer_MemoryOptimization benchmarks memory optimization
func BenchmarkVPNPerformanceOptimizer_MemoryOptimization(b *testing.B) {
	memOptimizer := NewMemoryOptimizer(1 * time.Minute)

	factory := func() interface{} {
		return &VPNConnection{}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		obj := memOptimizer.GetObject("vpn_conn", factory)
		_ = obj // Use the result
		memOptimizer.PutObject("vpn_conn", obj)
	}
}

// BenchmarkVPNPerformanceOptimizer_CacheAccess benchmarks cache access performance
func BenchmarkVPNPerformanceOptimizer_CacheAccess(b *testing.B) {
	cacheManager := NewCacheManager(10*time.Minute, 1000)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		cacheManager.Set(key, value)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		value, found := cacheManager.Get(key)
		_ = value // Use the result
		_ = found // Use the result
	}
}

// TestVPNPerformanceOptimizer_ResourceCleanup tests resource cleanup
func TestVPNPerformanceOptimizer_ResourceCleanup(t *testing.T) {
	config := &OptimizationConfig{
		EnableMemoryOptimization: true,
		MemoryCleanupInterval:    100 * time.Millisecond,
		EnableResultCaching:      true,
		CacheTTL:                 200 * time.Millisecond,
	}

	optimizer := NewVPNPerformanceOptimizer(config)

	// Add some data to cache
	optimizer.cacheManager.Set("test-key", "test-value")

	// Wait for cache to expire
	time.Sleep(300 * time.Millisecond)

	// Check that expired data is cleaned up
	_, found := optimizer.cacheManager.Get("test-key")
	assert.False(t, found)

	// Test memory cleanup
	optimizer.memoryOptimizer.TriggerCleanup()
	assert.False(t, optimizer.memoryOptimizer.lastCleanup.IsZero())

	t.Log("Resource cleanup test completed successfully")
}
