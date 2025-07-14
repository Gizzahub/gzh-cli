package largescale

import (
	"context"
	"fmt"
	"testing"
)

func TestDefaultLargeScaleConfig(t *testing.T) {
	config := DefaultLargeScaleConfig()

	if config.MaxConcurrency <= 0 {
		t.Error("MaxConcurrency should be positive")
	}

	if config.BatchSize <= 0 {
		t.Error("BatchSize should be positive")
	}

	if config.MemoryThreshold <= 0 {
		t.Error("MemoryThreshold should be positive")
	}

	// Should not exceed reasonable limits
	if config.MaxConcurrency > 50 {
		t.Error("MaxConcurrency seems too high for default config")
	}
}

func TestLargeScaleManagerCreation(t *testing.T) {
	// Test with default config
	manager := NewLargeScaleManager(nil, nil)
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if manager.config == nil {
		t.Error("Config should be initialized")
	}

	if manager.rateLimiter == nil {
		t.Error("Rate limiter should be initialized")
	}

	// Test with custom config
	config := &LargeScaleConfig{
		MaxConcurrency:  10,
		BatchSize:       50,
		UseShallowClone: false,
	}

	manager2 := NewLargeScaleManager(config, nil)
	if manager2.config.MaxConcurrency != 10 {
		t.Error("Custom config not applied correctly")
	}
}

func TestProgressCallback(t *testing.T) {
	callbackCalled := false
	var lastProcessed, lastTotal int
	var lastMessage string

	callback := func(processed, total int, message string) {
		callbackCalled = true
		lastProcessed = processed
		lastTotal = total
		lastMessage = message
	}

	manager := NewLargeScaleManager(nil, callback)

	// Simulate progress update
	if manager.progressCallback != nil {
		manager.progressCallback(50, 100, "Test message")
	}

	if !callbackCalled {
		t.Error("Progress callback should have been called")
	}

	if lastProcessed != 50 {
		t.Errorf("Expected processed=50, got %d", lastProcessed)
	}

	if lastTotal != 100 {
		t.Errorf("Expected total=100, got %d", lastTotal)
	}

	if lastMessage != "Test message" {
		t.Errorf("Expected message='Test message', got '%s'", lastMessage)
	}
}

func TestRepositoryFiltering(t *testing.T) {
	config := DefaultLargeScaleConfig()
	manager := NewLargeScaleManager(config, nil)

	testCases := []struct {
		repo     LargeScaleRepository
		expected bool
		name     string
	}{
		{
			repo: LargeScaleRepository{
				Name:     "normal-repo",
				Archived: false,
				Size:     1000, // 1MB
			},
			expected: false,
			name:     "normal repository should not be skipped",
		},
		{
			repo: LargeScaleRepository{
				Name:     "archived-repo",
				Archived: true,
				Size:     1000,
			},
			expected: true,
			name:     "archived repository should be skipped",
		},
		{
			repo: LargeScaleRepository{
				Name:     "huge-repo",
				Archived: false,
				Size:     2000000, // 2GB
			},
			expected: true,
			name:     "very large repository should be skipped under memory pressure",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := manager.shouldSkipRepository(tc.repo)
			if result != tc.expected {
				t.Errorf("Expected %v for %s, got %v", tc.expected, tc.repo.Name, result)
			}
		})
	}
}

func TestConcurrencyCalculation(t *testing.T) {
	config := &LargeScaleConfig{
		MaxConcurrency:  20,
		MemoryThreshold: 100 * 1024 * 1024, // 100MB
	}
	manager := NewLargeScaleManager(config, nil)

	testCases := []struct {
		totalRepos  int
		name        string
		maxExpected int
	}{
		{
			totalRepos:  100,
			name:        "small operation",
			maxExpected: 20,
		},
		{
			totalRepos:  2000,
			name:        "large operation should reduce concurrency",
			maxExpected: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			concurrency := manager.calculateOptimalConcurrency(tc.totalRepos)

			if concurrency <= 0 {
				t.Error("Concurrency should be positive")
			}

			if concurrency > tc.maxExpected {
				t.Errorf("Concurrency %d exceeds expected maximum %d", concurrency, tc.maxExpected)
			}
		})
	}
}

func TestStatsTracking(t *testing.T) {
	manager := NewLargeScaleManager(nil, nil)

	// Initial stats should be zero
	stats := manager.GetStats()
	if stats.ProcessedRepos != 0 {
		t.Error("Initial processed repos should be 0")
	}

	// Update stats
	manager.updateStats(5, 1, 2)

	stats = manager.GetStats()
	if stats.ProcessedRepos != 5 {
		t.Errorf("Expected 5 processed repos, got %d", stats.ProcessedRepos)
	}

	if stats.FailedRepos != 1 {
		t.Errorf("Expected 1 failed repo, got %d", stats.FailedRepos)
	}

	if stats.SkippedRepos != 2 {
		t.Errorf("Expected 2 skipped repos, got %d", stats.SkippedRepos)
	}

	// Stats should be cumulative
	manager.updateStats(3, 0, 1)
	stats = manager.GetStats()

	if stats.ProcessedRepos != 8 {
		t.Errorf("Expected 8 total processed repos, got %d", stats.ProcessedRepos)
	}

	if stats.SkippedRepos != 3 {
		t.Errorf("Expected 3 total skipped repos, got %d", stats.SkippedRepos)
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test minInt function
	if minInt(5, 3) != 3 {
		t.Error("minInt(5, 3) should be 3")
	}

	if minInt(1, 10) != 1 {
		t.Error("minInt(1, 10) should be 1")
	}

	// Test maxInt function
	if maxInt(5, 3) != 5 {
		t.Error("maxInt(5, 3) should be 5")
	}

	if maxInt(1, 10) != 10 {
		t.Error("maxInt(1, 10) should be 10")
	}

	// Test containsNextLink function
	testCases := []struct {
		linkHeader string
		expected   bool
		name       string
	}{
		{
			linkHeader: `<https://api.github.com/orgs/org/repos?page=2>; rel="next"`,
			expected:   true,
			name:       "should detect next link",
		},
		{
			linkHeader: `<https://api.github.com/orgs/org/repos?page=1>; rel="prev"`,
			expected:   false,
			name:       "should not detect prev-only link",
		},
		{
			linkHeader: "",
			expected:   false,
			name:       "should handle empty link header",
		},
		{
			linkHeader: `<https://api.github.com/orgs/org/repos?page=5>; rel="last"`,
			expected:   true,
			name:       "should detect last link",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := containsNextLink(tc.linkHeader)
			if result != tc.expected {
				t.Errorf("Expected %v for '%s', got %v", tc.expected, tc.linkHeader, result)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	manager := NewLargeScaleManager(nil, nil)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// This should return context.Canceled error
	_, err := manager.ListAllRepositories(ctx, "test-org")
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestConfigurationValidation(t *testing.T) {
	testCases := []struct {
		config    *LargeScaleConfig
		name      string
		shouldFix bool
	}{
		{
			config: &LargeScaleConfig{
				MaxConcurrency: -1, // Invalid
				BatchSize:      100,
			},
			name:      "negative concurrency should be fixed",
			shouldFix: true,
		},
		{
			config: &LargeScaleConfig{
				MaxConcurrency: 10,
				BatchSize:      0, // Invalid
			},
			name:      "zero batch size should be fixed",
			shouldFix: true,
		},
		{
			config: &LargeScaleConfig{
				MaxConcurrency: 10,
				BatchSize:      100,
				MaxRetries:     -1, // Invalid
			},
			name:      "negative retries should be fixed",
			shouldFix: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewLargeScaleManager(tc.config, nil)

			// The manager should have valid configuration
			if manager.config.MaxConcurrency <= 0 {
				t.Error("Manager should fix invalid MaxConcurrency")
			}

			if manager.config.BatchSize <= 0 {
				t.Error("Manager should fix invalid BatchSize")
			}

			if manager.config.MaxRetries < 0 {
				t.Error("Manager should fix invalid MaxRetries")
			}
		})
	}
}

// Benchmark tests for performance validation

func BenchmarkRepositoryCreation(b *testing.B) {
	manager := NewLargeScaleManager(nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo := LargeScaleRepository{
			Name:          fmt.Sprintf("repo-%d", i),
			FullName:      fmt.Sprintf("org/repo-%d", i),
			HTMLURL:       fmt.Sprintf("https://github.com/org/repo-%d", i),
			CloneURL:      fmt.Sprintf("https://github.com/org/repo-%d.git", i),
			DefaultBranch: "main",
			Size:          1000,
		}

		// Simulate processing
		_ = manager.shouldSkipRepository(repo)
	}
}

func BenchmarkStatsUpdate(b *testing.B) {
	manager := NewLargeScaleManager(nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.updateStats(1, 0, 0)
	}
}

func BenchmarkConcurrencyCalculation(b *testing.B) {
	manager := NewLargeScaleManager(nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.calculateOptimalConcurrency(1000)
	}
}
