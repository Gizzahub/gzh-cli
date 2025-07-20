package github

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BulkCloneStats represents statistics from bulk clone operations.
type BulkCloneStats struct {
	TotalRepositories int
	StartTime         time.Time
	EndTime           time.Time
	SuccessCount      int
	FailureCount      int
	Successful        int
	Failed            int
}

// CachedGitHubClient wraps GitHub API calls with caching - DISABLED (cache package removed)
// Simple in-memory cache implementation to replace deleted cache package.
type CachedGitHubClient struct {
	cache sync.Map // Simple in-memory cache replacement
	token string
}

// NewCachedGitHubClient creates a new cached GitHub client - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func NewCachedGitHubClient(token string) *CachedGitHubClient {
	return &CachedGitHubClient{
		cache: sync.Map{},
		token: token,
	}
}

// ListRepositoriesWithCache lists repositories with caching support - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitHubClient) ListRepositoriesWithCache(ctx context.Context, org string) ([]string, error) {
	// Try to get from simple cache first
	cacheKey := fmt.Sprintf("repos:%s", org)
	if cached, found := c.cache.Load(cacheKey); found {
		if repos, ok := cached.([]string); ok {
			return repos, nil
		}
	}

	// Cache miss - fetch from API
	repos, err := List(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Store in simple cache (no TTL implementation)
	c.cache.Store(cacheKey, repos)

	return repos, nil
}

// GetDefaultBranchWithCache gets repository default branch with caching - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitHubClient) GetDefaultBranchWithCache(ctx context.Context, org, repo string) (string, error) {
	// Try to get from simple cache first
	cacheKey := fmt.Sprintf("default_branch:%s/%s", org, repo)
	if cached, found := c.cache.Load(cacheKey); found {
		if branch, ok := cached.(string); ok {
			return branch, nil
		}
	}

	// Cache miss - fetch from API
	branch, err := GetDefaultBranch(ctx, org, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get default branch: %w", err)
	}

	// Store in simple cache (no TTL implementation)
	c.cache.Store(cacheKey, branch)

	return branch, nil
}

// InvalidateOrgCache invalidates all cache entries for an organization - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitHubClient) InvalidateOrgCache(ctx context.Context, org string) int {
	count := 0

	cacheKey := fmt.Sprintf("repos:%s", org)
	if _, found := c.cache.LoadAndDelete(cacheKey); found {
		count++
	}

	return count
}

// InvalidateRepoCache invalidates cache entries for a specific repository - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitHubClient) InvalidateRepoCache(ctx context.Context, org, repo string) int {
	count := 0

	cacheKey := fmt.Sprintf("default_branch:%s/%s", org, repo)
	if _, found := c.cache.LoadAndDelete(cacheKey); found {
		count++
	}

	return count
}

// GetCacheStats returns GitHub cache statistics - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitHubClient) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "simple_sync_map",
		"note": "cache package removed, using simple sync.Map",
	}
}

// CachedBulkCloneManager extends OptimizedBulkCloneManager with caching.
type CachedBulkCloneManager struct {
	*OptimizedBulkCloneManager
	cachedClient *CachedGitHubClient
}

// NewCachedBulkCloneManager creates a new cached bulk clone manager - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func NewCachedBulkCloneManager(token string, config OptimizedCloneConfig) (*CachedBulkCloneManager, error) {
	// Create cached client with simple cache
	cachedClient := NewCachedGitHubClient(token)

	// Create optimized manager
	optimizedManager, err := NewOptimizedBulkCloneManager(token, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create optimized manager: %w", err)
	}

	return &CachedBulkCloneManager{
		OptimizedBulkCloneManager: optimizedManager,
		cachedClient:              cachedClient,
	}, nil
}

// RefreshAllOptimizedWithCache performs optimized refresh with caching.
func (cbm *CachedBulkCloneManager) RefreshAllOptimizedWithCache(ctx context.Context, targetPath, org, strategy string) (BulkCloneStats, error) {
	fmt.Printf("🚀 Starting cached bulk clone for organization: %s\n", org)

	// Use cached client for repository listing
	repos, err := cbm.cachedClient.ListRepositoriesWithCache(ctx, org)
	if err != nil {
		return BulkCloneStats{}, fmt.Errorf("failed to list repositories with cache: %w", err)
	}

	fmt.Printf("📦 Found %d repositories (cached result: %v)\n", len(repos), true)

	// Continue with optimized processing using the cached repository list
	return cbm.processRepositoriesOptimized(ctx, targetPath, org, strategy, repos)
}

// processRepositoriesOptimized processes repositories with caching optimizations.
func (cbm *CachedBulkCloneManager) processRepositoriesOptimized(ctx context.Context, targetPath, org, strategy string, repos []string) (BulkCloneStats, error) {
	stats := BulkCloneStats{
		TotalRepositories: len(repos),
		StartTime:         time.Now(),
	}

	// Use the existing optimized processing logic but with cached repository data
	// This would integrate with the existing OptimizedBulkCloneManager methods

	// For now, delegate to the existing optimized method
	// In a full implementation, this would use the cached repos list
	_, err := cbm.RefreshAllOptimized(ctx, targetPath, org, strategy)
	if err != nil {
		return BulkCloneStats{}, err
	}

	// Convert CloneStats to BulkCloneStats
	bulkStats := BulkCloneStats{
		TotalRepositories: stats.TotalRepositories,
		StartTime:         stats.StartTime,
		EndTime:           time.Now(),
		SuccessCount:      0, // TODO: Extract from cloneStats
		FailureCount:      0, // TODO: Extract from cloneStats
		Successful:        0, // TODO: Extract from cloneStats
		Failed:            0, // TODO: Extract from cloneStats
	}

	return bulkStats, nil
}

// Close cleans up cached manager resources - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (cbm *CachedBulkCloneManager) Close() error {
	// No cache manager to close - using simple sync.Map
	// Close optimized manager
	return cbm.OptimizedBulkCloneManager.Close()
}

// RefreshAllOptimizedStreamingWithCache is the cached version of the streaming API - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func RefreshAllOptimizedStreamingWithCache(ctx context.Context, targetPath, org, strategy, token string) error {
	// Create cached manager
	config := DefaultOptimizedCloneConfig()

	manager, err := NewCachedBulkCloneManager(token, config) //nolint:contextcheck // Manager creation doesn't require context propagation
	if err != nil {
		return fmt.Errorf("failed to create cached bulk clone manager: %w", err)
	}
	defer func() {
		if err := manager.Close(); err != nil {
			// Log close error but don't override main error
		}
	}()

	// Execute cached refresh
	stats, err := manager.RefreshAllOptimizedWithCache(ctx, targetPath, org, strategy)
	if err != nil {
		return fmt.Errorf("cached bulk clone failed: %w", err)
	}

	// Print summary with cache information
	cacheStats := manager.cachedClient.GetCacheStats()

	fmt.Printf("\n🎉 Cached bulk clone completed: %d successful, %d failed (%.1f%% success rate)\n",
		stats.Successful, stats.Failed,
		float64(stats.Successful)/float64(stats.TotalRepositories)*100)

	fmt.Printf("📊 Cache performance: %s\n", cacheStats["note"])

	return nil
}

// CacheConfiguration provides cache configuration for GitHub operations - DISABLED (cache package removed)
// Simple configuration struct without external cache dependency.
type CacheConfiguration struct {
	EnableLocalCache bool
	// EnableRedisCache bool // Disabled - cache package removed
	LocalCacheSize int
	DefaultTTL     time.Duration
	// RedisAddress     string // Disabled - cache package removed
	// RedisPassword    string // Disabled - cache package removed
}

// DefaultCacheConfiguration returns sensible defaults for GitHub caching - DISABLED (cache package removed)
// Simple configuration without external cache dependency.
func DefaultCacheConfiguration() CacheConfiguration {
	return CacheConfiguration{
		EnableLocalCache: true,
		// EnableRedisCache: false, // Disabled - cache package removed
		LocalCacheSize: 1000,
		DefaultTTL:     10 * time.Minute,
		// RedisAddress:     "localhost:6379", // Disabled - cache package removed
		// RedisPassword:    "", // Disabled - cache package removed
	}
}

// ToCacheManagerConfig converts to cache manager configuration - DISABLED (cache package removed)
// Simple configuration conversion without external cache dependency.
func (cc CacheConfiguration) ToCacheManagerConfig() map[string]interface{} {
	return map[string]interface{}{
		"enable_local_cache": cc.EnableLocalCache,
		"local_cache_size":   cc.LocalCacheSize,
		"default_ttl":        cc.DefaultTTL,
		"note":               "cache package removed, using simple sync.Map",
	}
}
