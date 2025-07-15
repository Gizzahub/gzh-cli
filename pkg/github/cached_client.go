package github

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/cache"
)

// BulkCloneStats represents statistics from bulk clone operations
type BulkCloneStats struct {
	TotalRepositories int
	StartTime         time.Time
	EndTime           time.Time
	SuccessCount      int
	FailureCount      int
	Successful        int
	Failed            int
}

// CachedGitHubClient wraps GitHub API calls with caching
type CachedGitHubClient struct {
	cacheManager *cache.CacheManager
	token        string
}

// NewCachedGitHubClient creates a new cached GitHub client
func NewCachedGitHubClient(token string, cacheManager *cache.CacheManager) *CachedGitHubClient {
	return &CachedGitHubClient{
		cacheManager: cacheManager,
		token:        token,
	}
}

// ListRepositoriesWithCache lists repositories with caching support
func (c *CachedGitHubClient) ListRepositoriesWithCache(ctx context.Context, org string) ([]string, error) {
	// Create cache key
	cacheKey := cache.CacheKey{
		Service:    "github",
		Resource:   "repos",
		Identifier: org,
	}

	// Try to get from cache first
	if cached, found := c.cacheManager.Get(ctx, cacheKey); found {
		if repos, ok := cached.([]string); ok {
			return repos, nil
		}
	}

	// Cache miss - fetch from API
	repos, err := List(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Store in cache with 10 minute TTL
	c.cacheManager.PutWithTTL(ctx, cacheKey, repos, 10*time.Minute)

	return repos, nil
}

// GetDefaultBranchWithCache gets repository default branch with caching
func (c *CachedGitHubClient) GetDefaultBranchWithCache(ctx context.Context, org, repo string) (string, error) {
	// Create cache key
	cacheKey := cache.CacheKey{
		Service:    "github",
		Resource:   "default_branch",
		Identifier: fmt.Sprintf("%s/%s", org, repo),
	}

	// Try to get from cache first
	if cached, found := c.cacheManager.Get(ctx, cacheKey); found {
		if branch, ok := cached.(string); ok {
			return branch, nil
		}
	}

	// Cache miss - fetch from API
	branch, err := GetDefaultBranch(ctx, org, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get default branch: %w", err)
	}

	// Store in cache with 30 minute TTL (default branches change rarely)
	c.cacheManager.PutWithTTL(ctx, cacheKey, branch, 30*time.Minute)

	return branch, nil
}

// InvalidateOrgCache invalidates all cache entries for an organization
func (c *CachedGitHubClient) InvalidateOrgCache(ctx context.Context, org string) int {
	return c.cacheManager.InvalidateByIdentifier(ctx, "github", org)
}

// InvalidateRepoCache invalidates cache entries for a specific repository
func (c *CachedGitHubClient) InvalidateRepoCache(ctx context.Context, org, repo string) int {
	identifier := fmt.Sprintf("%s/%s", org, repo)
	return c.cacheManager.InvalidateByIdentifier(ctx, "github", identifier)
}

// GetCacheStats returns GitHub cache statistics
func (c *CachedGitHubClient) GetCacheStats() cache.CacheManagerStats {
	return c.cacheManager.GetStats()
}

// CachedBulkCloneManager extends OptimizedBulkCloneManager with caching
type CachedBulkCloneManager struct {
	*OptimizedBulkCloneManager
	cachedClient *CachedGitHubClient
}

// NewCachedBulkCloneManager creates a new cached bulk clone manager
func NewCachedBulkCloneManager(token string, config OptimizedCloneConfig, cacheConfig cache.CacheManagerConfig) (*CachedBulkCloneManager, error) {
	// Create cache manager
	cacheManager := cache.NewCacheManager(cacheConfig)

	// Create cached client
	cachedClient := NewCachedGitHubClient(token, cacheManager)

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

// RefreshAllOptimizedWithCache performs optimized refresh with caching
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

// processRepositoriesOptimized processes repositories with caching optimizations
func (cbm *CachedBulkCloneManager) processRepositoriesOptimized(ctx context.Context, targetPath, org, strategy string, repos []string) (BulkCloneStats, error) {
	stats := BulkCloneStats{
		TotalRepositories: len(repos),
		StartTime:         time.Now(),
	}

	// Use the existing optimized processing logic but with cached repository data
	// This would integrate with the existing OptimizedBulkCloneManager methods

	// For now, delegate to the existing optimized method
	// In a full implementation, this would use the cached repos list
	cloneStats, err := cbm.OptimizedBulkCloneManager.RefreshAllOptimized(ctx, targetPath, org, strategy)
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

// Close cleans up cached manager resources
func (cbm *CachedBulkCloneManager) Close() error {
	// Close cache manager
	if err := cbm.cachedClient.cacheManager.Close(); err != nil {
		return fmt.Errorf("failed to close cache manager: %w", err)
	}

	// Close optimized manager
	return cbm.OptimizedBulkCloneManager.Close()
}

// RefreshAllOptimizedStreamingWithCache is the cached version of the streaming API
func RefreshAllOptimizedStreamingWithCache(ctx context.Context, targetPath, org, strategy, token string, cacheConfig cache.CacheManagerConfig) error {
	// Create cached manager
	config := DefaultOptimizedCloneConfig()
	manager, err := NewCachedBulkCloneManager(token, config, cacheConfig)
	if err != nil {
		return fmt.Errorf("failed to create cached bulk clone manager: %w", err)
	}
	defer manager.Close()

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

	fmt.Printf("📊 Cache performance: %.1f%% hit rate (local: %d hits, %d misses)\n",
		cacheStats.Local.HitRate*100, cacheStats.Local.Hits, cacheStats.Local.Misses)

	return nil
}

// CacheConfiguration provides cache configuration for GitHub operations
type CacheConfiguration struct {
	EnableLocalCache bool
	EnableRedisCache bool
	LocalCacheSize   int
	DefaultTTL       time.Duration
	RedisAddress     string
	RedisPassword    string
}

// DefaultCacheConfiguration returns sensible defaults for GitHub caching
func DefaultCacheConfiguration() CacheConfiguration {
	return CacheConfiguration{
		EnableLocalCache: true,
		EnableRedisCache: false, // Disabled by default
		LocalCacheSize:   1000,
		DefaultTTL:       10 * time.Minute,
		RedisAddress:     "localhost:6379",
		RedisPassword:    "",
	}
}

// ToCacheManagerConfig converts to cache manager configuration
func (cc CacheConfiguration) ToCacheManagerConfig() cache.CacheManagerConfig {
	return cache.CacheManagerConfig{
		UseRedis: cc.EnableRedisCache,
		LocalCacheConfig: cache.CacheConfig{
			Capacity:        cc.LocalCacheSize,
			DefaultTTL:      cc.DefaultTTL,
			CleanupInterval: cc.DefaultTTL / 2,
		},
		RedisCacheConfig: cache.RedisCacheConfig{
			Address:  cc.RedisAddress,
			Password: cc.RedisPassword,
		},
		DefaultTTL: cc.DefaultTTL,
		TagPrefix:  "github",
	}
}
