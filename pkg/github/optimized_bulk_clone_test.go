//nolint:testpackage // White-box testing needed for internal function access
package github

import (
	"testing"
	"time"
)

func TestParseMemorySize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"100B", 100, false},
		{"1KB", 1024, false},
		{"1MB", 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"1TB", 1024 * 1024 * 1024 * 1024, false},
		{"500MB", 500 * 1024 * 1024, false},
		{"2.5GB", int64(2.5 * 1024 * 1024 * 1024), false},
		{"", 0, true},
		{"invalid", 0, true},
		{"100XB", 0, true},
	}

	for _, test := range tests {
		result, err := parseMemorySize(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			}

			if result != test.expected {
				t.Errorf("For input %s, expected %d but got %d", test.input, test.expected, result)
			}
		}
	}
}

func TestMemoryStats(t *testing.T) {
	stats := GetMemoryStats()
	if stats == nil {
		t.Fatal("GetMemoryStats returned nil")
	}

	if stats.Alloc == 0 {
		t.Error("Expected non-zero allocated memory")
	}

	if stats.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	// Test string representation
	statsStr := stats.String()
	if statsStr == "" {
		t.Error("Expected non-empty string representation")
	}

	// Test efficiency metrics
	efficiency := stats.MemoryEfficiency()
	if len(efficiency) == 0 {
		t.Error("Expected efficiency metrics")
	}
}

func TestMemoryPressure(t *testing.T) {
	// Test with different memory limits
	pressure := GetMemoryPressure(1024 * 1024 * 1024) // 1GB

	validPressures := []MemoryPressureLevel{
		MemoryPressureLow,
		MemoryPressureMedium,
		MemoryPressureHigh,
		MemoryPressureCritical,
	}

	found := false

	for _, valid := range validPressures {
		if pressure == valid {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Invalid memory pressure level: %v", pressure)
	}

	// Test string representation
	pressureStr := pressure.String()
	if pressureStr == "" || pressureStr == "Unknown" {
		t.Errorf("Invalid pressure string: %s", pressureStr)
	}
}

func TestMemoryWatcher(t *testing.T) {
	maxMemory := int64(1024 * 1024 * 1024) // 1GB
	threshold := 0.8
	checkInterval := 10 * time.Millisecond

	watcher := NewMemoryWatcher(maxMemory, threshold, checkInterval)
	if watcher == nil {
		t.Fatal("NewMemoryWatcher returned nil")
	}

	pressureEvents := 0

	watcher.SetPressureHandler(func(level MemoryPressureLevel) {
		pressureEvents++
	})

	watcher.Start()
	time.Sleep(50 * time.Millisecond) // Let it run for a bit
	watcher.Stop()

	// Note: We can't guarantee pressure events will occur in this test
	// as it depends on actual memory usage, but we can test that the
	// watcher doesn't crash and can be started/stopped
}

func TestOptimizeMemoryUsage(t *testing.T) {
	// Get stats before optimization
	beforeStats := GetMemoryStats()

	// Perform optimization
	optimizationStats := OptimizeMemoryUsage()

	// Get stats after optimization
	afterStats := GetMemoryStats()

	if optimizationStats == nil {
		t.Fatal("OptimizeMemoryUsage returned nil")
	}

	// After optimization, we should have more GC cycles
	if afterStats.NumGC <= beforeStats.NumGC {
		t.Error("Expected more GC cycles after optimization")
	}

	// The optimization stats should have a timestamp
	if optimizationStats.Timestamp.IsZero() {
		t.Error("Expected optimization stats to have a timestamp")
	}
}

func TestDefaultOptimizedCloneConfig(t *testing.T) {
	config := DefaultOptimizedCloneConfig()

	if config.MaxMemoryUsage <= 0 {
		t.Error("Expected positive max memory usage")
	}

	if config.MemoryThreshold <= 0 || config.MemoryThreshold >= 1 {
		t.Error("Expected memory threshold between 0 and 1")
	}

	if config.GCInterval <= 0 {
		t.Error("Expected positive GC interval")
	}

	if config.BatchSize <= 0 {
		t.Error("Expected positive batch size")
	}

	if config.PrefetchSize <= 0 {
		t.Error("Expected positive prefetch size")
	}

	if config.WorkerPoolConfig.CloneWorkers <= 0 {
		t.Error("Expected positive number of clone workers")
	}
}

// Benchmark memory optimization.
func BenchmarkOptimizeMemoryUsage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		OptimizeMemoryUsage()
	}
}

// Benchmark memory stats collection.
func BenchmarkGetMemoryStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetMemoryStats()
	}
}

// Test streaming client creation (without actual API calls).
func TestNewStreamingClient(t *testing.T) {
	config := DefaultStreamingConfig()
	client := NewStreamingClient("test-token", config)

	if client == nil {
		t.Fatal("NewStreamingClient returned nil")
	}

	if client.token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", client.token)
	}

	// Test closing the client
	err := client.Close()
	if err != nil {
		t.Errorf("Unexpected error closing client: %v", err)
	}
}

// Test optimized bulk clone manager creation.
func TestNewOptimizedBulkCloneManager(t *testing.T) {
	config := DefaultOptimizedCloneConfig()

	manager, err := NewOptimizedBulkCloneManager("test-token", config)
	if err != nil {
		t.Fatalf("Unexpected error creating manager: %v", err)
	}

	if manager == nil {
		t.Fatal("NewOptimizedBulkCloneManager returned nil")
	}

	// Test closing the manager
	err = manager.Close()
	if err != nil {
		t.Errorf("Unexpected error closing manager: %v", err)
	}
}

// Test memory monitor.
func TestMemoryMonitor(t *testing.T) {
	config := DefaultOptimizedCloneConfig()

	manager, err := NewOptimizedBulkCloneManager("test-token", config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	// Test memory usage tracking
	usage := manager.getCurrentMemoryUsage()
	if usage < 0 {
		t.Error("Expected non-negative memory usage")
	}

	// Test memory optimization
	err = manager.checkAndOptimizeMemory()
	if err != nil {
		t.Errorf("Unexpected error in memory optimization: %v", err)
	}
}

// Test repository stream processing (mock).
func TestRepositoryStreamProcessing(t *testing.T) {
	// This would test the streaming functionality with a mock server
	// For now, we'll test the data structures and basic functionality
	repo := &Repository{
		ID:            12345,
		Name:          "test-repo",
		FullName:      "test-org/test-repo",
		DefaultBranch: "main",
		Private:       false,
	}

	if repo.ID != 12345 {
		t.Error("Repository ID not set correctly")
	}

	if repo.Name != "test-repo" {
		t.Error("Repository name not set correctly")
	}

	stream := RepositoryStream{
		Repository: repo,
		Error:      nil,
		Metadata: StreamMetadata{
			Page:        1,
			ProcessedAt: time.Now(),
			MemoryUsage: 1024,
		},
	}

	if stream.Repository.ID != 12345 {
		t.Error("Stream repository ID not correct")
	}

	if stream.Metadata.Page != 1 {
		t.Error("Stream metadata page not correct")
	}
}
