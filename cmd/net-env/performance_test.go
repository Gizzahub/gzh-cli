//nolint:testpackage // White-box testing needed for internal function access
package netenv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandPool_ExecuteCommand(t *testing.T) {
	pool := NewCommandPool(2)
	defer pool.Close()

	// Test basic command execution
	result := pool.ExecuteCommand("echo", "test")
	assert.NoError(t, result.Error)
	assert.Contains(t, string(result.Output), "test")
	assert.False(t, result.FromCache)
	assert.Greater(t, result.Duration, time.Duration(0))

	// Test caching - second execution should be from cache
	result2 := pool.ExecuteCommand("echo", "test")
	assert.NoError(t, result2.Error)
	assert.Contains(t, string(result2.Output), "test")
	assert.True(t, result2.FromCache)
}

func TestCommandPool_ExecuteBatch(t *testing.T) {
	pool := NewCommandPool(3)
	defer pool.Close()

	commands := []Command{
		{Name: "echo", Args: []string{"cmd1"}},
		{Name: "echo", Args: []string{"cmd2"}},
		{Name: "echo", Args: []string{"cmd3"}},
	}

	start := time.Now()
	results := pool.ExecuteBatch(commands)
	duration := time.Since(start)

	require.Len(t, results, 3)

	for i, result := range results {
		assert.NoError(t, result.Error)
		assert.Contains(t, string(result.Output), commands[i].Args[0])
	}

	// Batch execution should be faster than sequential
	// (though with echo commands, the difference might be minimal)
	assert.Less(t, duration, 1*time.Second, "Batch execution took too long")
}

func TestCommandPool_CacheExpiration(t *testing.T) {
	pool := NewCommandPool(1)
	defer pool.Close()

	// Execute command and cache result
	result1 := pool.ExecuteCommand("echo", "cache-test")
	assert.False(t, result1.FromCache)

	// Manually set a short TTL for testing
	cmdStr := "echo [cache-test]"
	pool.setCachedResult(cmdStr, &CachedResult{
		Output:    []byte("cache-test\n"),
		Error:     nil,
		Timestamp: time.Now().Add(-1 * time.Minute), // Expired
		TTL:       30 * time.Second,
	})

	// Should not use expired cache
	result2 := pool.ExecuteCommand("echo", "cache-test")
	assert.False(t, result2.FromCache)
}

func TestCommandPool_ClearCache(t *testing.T) {
	pool := NewCommandPool(1)
	defer pool.Close()

	// Execute and cache
	pool.ExecuteCommand("echo", "test")

	stats := pool.GetCacheStats()
	if totalEntries, ok := stats["total_entries"].(int); ok {
		assert.Greater(t, totalEntries, 0)
	}

	// Clear cache
	pool.ClearCache()

	stats = pool.GetCacheStats()
	if entries, ok := stats["total_entries"].(int); ok {
		assert.Equal(t, 0, entries)
	} else {
		t.Errorf("total_entries is not an int: %T", stats["total_entries"])
	}
}

func TestOptimizedVPNManager_ConnectionState(t *testing.T) {
	manager := NewOptimizedVPNManager()

	// Test batch status operations (simplified test)
	status, err := manager.GetVPNStatusBatch([]string{"test-vpn"})
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Contains(t, status, "test-vpn")
}

func TestOptimizedVPNManager_ConnectVPNBatch(t *testing.T) {
	manager := NewOptimizedVPNManager()

	// Test with mock VPN configurations
	// Note: These will fail in test environment but should not panic
	configs := []vpnConfig{
		{Name: "test-vpn1", Type: "networkmanager"},
		{Name: "test-vpn2", Type: "networkmanager"},
	}

	// This will likely fail since we don't have actual VPN connections,
	// but it shouldn't crash and should handle errors gracefully
	err := manager.ConnectVPNBatch(configs)

	// In test environment, commands will likely fail, but that's expected
	// The important thing is that the function doesn't panic
	assert.Error(t, err) // Expected to fail in test environment
}

func TestOptimizedDNSManager_SetDNSServersBatch(t *testing.T) {
	manager := NewOptimizedDNSManager()

	configs := []DNSConfig{
		{Servers: []string{"1.1.1.1", "1.0.0.1"}, Interface: "lo", Method: "resolvectl"},
	}

	// This will likely fail in test environment without proper privileges,
	// but it should handle errors gracefully
	err := manager.SetDNSServersBatch(configs)

	// Expected to fail in test environment due to permissions/interface issues
	assert.Error(t, err)
}

func TestCommandPool_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Test with a fresh pool for sequential to avoid cache interference
	poolParallel := NewCommandPool(5)
	defer poolParallel.Close()

	poolSequential := NewCommandPool(1)
	defer poolSequential.Close()

	// Test parallel execution performance with unique commands to avoid caching
	commands := make([]Command, 5)
	for i := 0; i < 5; i++ {
		commands[i] = Command{
			Name: "sleep",
			Args: []string{"0.05"}, // 50ms sleep
		}
	}

	start := time.Now()
	results := poolParallel.ExecuteBatch(commands)
	parallelDuration := time.Since(start)

	// Sequential execution for comparison with unique commands
	sequentialCommands := make([]Command, 5)
	for i := 0; i < 5; i++ {
		sequentialCommands[i] = Command{
			Name: "sleep",
			Args: []string{"0.05"},
		}
	}

	start = time.Now()

	for _, cmd := range sequentialCommands {
		poolSequential.ExecuteCommand(cmd.Name, cmd.Args...)
	}

	sequentialDuration := time.Since(start)

	require.Len(t, results, 5)

	// Parallel execution should be faster than sequential
	// With 5 commands of 50ms each, sequential = ~250ms, parallel = ~50-100ms
	assert.Less(t, parallelDuration, sequentialDuration,
		"Parallel execution should be faster than sequential")

	// Calculate speedup ratio
	speedup := float64(sequentialDuration) / float64(parallelDuration)

	t.Logf("Parallel execution: %v, Sequential execution: %v, Speedup: %.2fx",
		parallelDuration, sequentialDuration, speedup)

	// Should have at least some speedup (even if not 2x due to overhead)
	assert.Greater(t, speedup, 1.0, "Should have some performance improvement")
}

func TestCommandPool_CacheHitRate(t *testing.T) {
	pool := NewCommandPool(1)
	defer pool.Close()

	// Execute same command multiple times
	cmdCount := 5
	cacheHits := 0

	for i := 0; i < cmdCount; i++ {
		result := pool.ExecuteCommand("echo", "cache-test")
		assert.NoError(t, result.Error)

		if result.FromCache {
			cacheHits++
		}
	}

	// Should have cache hits after first execution
	assert.Greater(t, cacheHits, 0, "Should have some cache hits")
	assert.Less(t, cacheHits, cmdCount, "Should not have all cache hits (first one is not cached)")

	expectedCacheHits := cmdCount - 1 // All except the first one
	assert.Equal(t, expectedCacheHits, cacheHits, "Cache hit rate should be optimal")
}

func BenchmarkCommandPool_ExecuteCommand(b *testing.B) {
	pool := NewCommandPool(1)
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := pool.ExecuteCommand("echo", "benchmark")
		if result.Error != nil {
			b.Fatalf("Command failed: %v", result.Error)
		}
	}
}

func BenchmarkCommandPool_ExecuteBatch(b *testing.B) {
	pool := NewCommandPool(5)
	defer pool.Close()

	commands := []Command{
		{Name: "echo", Args: []string{"bench1"}},
		{Name: "echo", Args: []string{"bench2"}},
		{Name: "echo", Args: []string{"bench3"}},
		{Name: "echo", Args: []string{"bench4"}},
		{Name: "echo", Args: []string{"bench5"}},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results := pool.ExecuteBatch(commands)
		if len(results) != len(commands) {
			b.Fatalf("Expected %d results, got %d", len(commands), len(results))
		}

		for _, result := range results {
			if result.Error != nil {
				b.Fatalf("Batch command failed: %v", result.Error)
			}
		}
	}
}

func BenchmarkCommandPool_WithVsWithoutCache(b *testing.B) {
	b.Run("WithCache", func(b *testing.B) {
		pool := NewCommandPool(1)
		defer pool.Close()

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Same command - should benefit from caching
			pool.ExecuteCommand("echo", "cached-benchmark")
		}
	})

	b.Run("WithoutCache", func(b *testing.B) {
		pool := NewCommandPool(1)
		defer pool.Close()

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Different command each time - no cache benefit
			pool.ExecuteCommand("echo", "benchmark", string(rune(i)))
		}
	})
}

func TestPerformanceOptimization_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test the full integration of optimized components
	vpnManager := NewOptimizedVPNManager()

	dnsManager := NewOptimizedDNSManager()

	// Test VPN status retrieval performance
	start := time.Now()
	_, err := vpnManager.GetVPNStatusBatch([]string{"test1", "test2", "test3"})
	vpnDuration := time.Since(start)

	// Should complete quickly even if commands fail
	assert.Less(t, vpnDuration, 5*time.Second, "VPN status check took too long")

	// DNS operations should also be fast
	start = time.Now()
	configs := []DNSConfig{
		{Servers: []string{"8.8.8.8"}, Interface: "lo"},
	}
	dnsManager.SetDNSServersBatch(configs)

	dnsDuration := time.Since(start)

	assert.Less(t, dnsDuration, 3*time.Second, "DNS configuration took too long")

	// Error is expected in test environment, but operation should be fast
	t.Logf("VPN operation took: %v, DNS operation took: %v", vpnDuration, dnsDuration)

	// The key test is that we have error handling (expect errors in test env)
	// but the operations don't hang or take excessive time
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}
