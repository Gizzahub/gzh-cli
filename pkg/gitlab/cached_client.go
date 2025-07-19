package gitlab

import (
	"context"
	"fmt"
	"sync"
)

// CachedGitLabClient wraps GitLab API calls with caching - DISABLED (cache package removed)
// Simple in-memory cache implementation to replace deleted cache package.
type CachedGitLabClient struct {
	cache           sync.Map // Simple in-memory cache replacement
	streamingClient *StreamingClient
	token           string
	baseURL         string
}

// NewCachedGitLabClient creates a new cached GitLab client - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func NewCachedGitLabClient(token, baseURL string) *CachedGitLabClient {
	streamingConfig := DefaultStreamingConfig()
	streamingClient := NewStreamingClient(token, baseURL, streamingConfig)

	return &CachedGitLabClient{
		cache:           sync.Map{},
		streamingClient: streamingClient,
		token:           token,
		baseURL:         baseURL,
	}
}

// ListGroupProjectsWithCache lists group projects with caching support - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitLabClient) ListGroupProjectsWithCache(ctx context.Context, groupID string) ([]*Project, error) {
	// Try to get from simple cache first
	cacheKey := fmt.Sprintf("projects:%s", groupID)
	if cached, found := c.cache.Load(cacheKey); found {
		if projects, ok := cached.([]*Project); ok {
			return projects, nil
		}
	}

	// Cache miss - fetch from streaming API
	projects, err := c.fetchProjectsFromStream(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	// Store in simple cache (no TTL implementation)
	c.cache.Store(cacheKey, projects)

	return projects, nil
}

// fetchProjectsFromStream collects all projects from the streaming API.
func (c *CachedGitLabClient) fetchProjectsFromStream(ctx context.Context, groupID string) ([]*Project, error) {
	config := DefaultStreamingConfig()

	projectChan, err := c.streamingClient.StreamGroupProjects(ctx, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to start streaming: %w", err)
	}

	var projects []*Project

	for projectStream := range projectChan {
		if projectStream.Error != nil {
			return nil, fmt.Errorf("streaming error: %w", projectStream.Error)
		}

		if projectStream.Project != nil {
			projects = append(projects, projectStream.Project)
		}
	}

	return projects, nil
}

// GetProjectWithCache gets a specific project with caching - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitLabClient) GetProjectWithCache(_ context.Context, projectID string) (*Project, error) {
	// Try to get from simple cache first
	cacheKey := fmt.Sprintf("project:%s", projectID)
	if cached, found := c.cache.Load(cacheKey); found {
		if project, ok := cached.(*Project); ok {
			return project, nil
		}
	}

	// Cache miss - would fetch from GitLab API
	// For now, return error as single project fetch is not implemented
	return nil, fmt.Errorf("single project fetch not implemented")
}

// InvalidateGroupCache invalidates all cache entries for a group - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitLabClient) InvalidateGroupCache(ctx context.Context, groupID string) int {
	count := 0

	cacheKey := fmt.Sprintf("projects:%s", groupID)
	if _, found := c.cache.LoadAndDelete(cacheKey); found {
		count++
	}

	return count
}

// InvalidateProjectCache invalidates cache entries for a specific project - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitLabClient) InvalidateProjectCache(ctx context.Context, projectID string) int {
	count := 0

	cacheKey := fmt.Sprintf("project:%s", projectID)
	if _, found := c.cache.LoadAndDelete(cacheKey); found {
		count++
	}

	return count
}

// GetCacheStats returns GitLab cache statistics - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitLabClient) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "simple_sync_map",
		"note": "cache package removed, using simple sync.Map",
	}
}

// Close cleans up cached client resources - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func (c *CachedGitLabClient) Close() error {
	// Close streaming client
	if err := c.streamingClient.Close(); err != nil {
		return fmt.Errorf("failed to close streaming client: %w", err)
	}

	// No cache manager to close - using simple sync.Map
	return nil
}

// CachedStreamingClient extends StreamingClient with caching capabilities - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
type CachedStreamingClient struct {
	*StreamingClient
	cache sync.Map // Simple in-memory cache replacement
}

// NewCachedStreamingClient creates a streaming client with caching - DISABLED (cache package removed)
// Simple implementation without external cache dependency.
func NewCachedStreamingClient(token, baseURL string, config StreamingConfig) *CachedStreamingClient {
	streamingClient := NewStreamingClient(token, baseURL, config)

	return &CachedStreamingClient{
		StreamingClient: streamingClient,
		cache:           sync.Map{},
	}
}

// StreamGroupProjectsWithCache streams projects with intelligent caching.
func (csc *CachedStreamingClient) StreamGroupProjectsWithCache(ctx context.Context, groupID string, config StreamingConfig) (<-chan ProjectStream, error) {
	resultChan := make(chan ProjectStream, config.BufferSize)

	// Check if we have cached results first
	cacheKey := fmt.Sprintf("stream_projects:%s", groupID)
	if cached, found := csc.cache.Load(cacheKey); found {
		if projects, ok := cached.([]ProjectStream); ok {
			go func() {
				defer close(resultChan)

				for _, project := range projects {
					select {
					case resultChan <- project:
					case <-ctx.Done():
						return
					}
				}
			}()

			return resultChan, nil
		}
	}

	// Cache miss - use streaming API
	// For now, delegate to regular streaming
	// In a full implementation, this would cache the results
	return csc.StreamGroupProjects(ctx, groupID, config)
	/*
			// Original cache logic - disabled
			// Check if we have cached results first
			// cacheKey := cache.CacheKey{
			//	Service:    "gitlab",
			//	Resource:   "projects",
			//	Identifier: groupID,
			//	Params: map[string]string{
					"page_size": fmt.Sprintf("%d", config.PageSize),
				},
			}

			go func() {
				defer close(resultChan)

				// Try cache first
				if cached, found := csc.cacheManager.Get(ctx, cacheKey); found {
					if projects, ok := cached.([]*Project); ok {
						// Stream cached results
						for _, project := range projects {
							select {
							case <-ctx.Done():
								return
							case resultChan <- ProjectStream{
								Project: project,
								Metadata: StreamMetadata{
									ProcessedAt: time.Now(),
									CacheHit:    true,
								},
							}:
							}
						}
						return
					}
				}

				// Cache miss - stream from API and cache results
				var allProjects []*Project

				apiChan, err := csc.StreamingClient.StreamGroupProjects(ctx, groupID, config)
				if err != nil {
					resultChan <- ProjectStream{
						Error: fmt.Errorf("failed to start API streaming: %w", err),
					}
					return
				}

				for projectStream := range apiChan {
					// Forward to result channel
					resultChan <- projectStream

					// Collect for caching
					if projectStream.Error == nil && projectStream.Project != nil {
						allProjects = append(allProjects, projectStream.Project)
					}
				}

				// Cache the complete result
				if len(allProjects) > 0 {
					csc.cacheManager.PutWithTTL(ctx, cacheKey, allProjects, 15*time.Minute)
				}
			}()

			return resultChan, nil
		}

		// CacheConfiguration provides cache configuration for GitLab operations
		type CacheConfiguration struct {
			EnableLocalCache bool
			EnableRedisCache bool
			LocalCacheSize   int
			DefaultTTL       time.Duration
			RedisAddress     string
			RedisPassword    string
			StreamingConfig  StreamingConfig
		}

		// DefaultGitLabCacheConfiguration returns sensible defaults for GitLab caching
		func DefaultGitLabCacheConfiguration() CacheConfiguration {
			return CacheConfiguration{
				EnableLocalCache: true,
				EnableRedisCache: false, // Disabled by default
				LocalCacheSize:   1000,
				DefaultTTL:       15 * time.Minute, // Longer TTL for GitLab
				RedisAddress:     "localhost:6379",
				RedisPassword:    "",
				StreamingConfig:  DefaultStreamingConfig(),
			}
		}

		// ToCacheManagerConfig converts to cache manager configuration - DISABLED (cache package removed)
		// Simple configuration conversion without external cache dependency
		func (cc CacheConfiguration) ToCacheManagerConfig() map[string]interface{} {
			return map[string]interface{}{
				"enable_local_cache": cc.EnableLocalCache,
				"local_cache_size":   cc.LocalCacheSize,
				"default_ttl":        cc.DefaultTTL,
				"note":               "cache package removed, using simple sync.Map",
			}
		}

		// RefreshAllOptimizedStreamingWithCache performs GitLab refresh with caching - DISABLED (cache package removed)
		// Simple implementation without external cache dependency
		func RefreshAllOptimizedStreamingWithCache(ctx context.Context, targetPath, groupID, strategy, token, baseURL string, cacheConfig CacheConfiguration) error {
			// Create cached client with simple cache
			cachedClient := NewCachedGitLabClient(token, baseURL)
			defer cachedClient.Close()

			// Fetch projects with caching
			projects, err := cachedClient.ListGroupProjectsWithCache(ctx, groupID)
			if err != nil {
				return fmt.Errorf("failed to list group projects with cache: %w", err)
			}

			fmt.Printf("🚀 Starting cached GitLab bulk clone for group: %s\n", groupID)
			fmt.Printf("📦 Found %d projects (using cache)\n", len(projects))

			// Process projects (this would integrate with existing GitLab bulk clone logic)
			// For now, just report success
			cacheStats := cachedClient.GetCacheStats()
			fmt.Printf("📊 Cache performance: %.1f%% hit rate (local: %d hits, %d misses)\n",
				cacheStats.Local.HitRate*100, cacheStats.Local.Hits, cacheStats.Local.Misses)

			fmt.Printf("✅ GitLab cached bulk clone completed successfully\n")

			return nil
		}

		// CacheInvalidationStrategy defines cache invalidation behaviors
		type CacheInvalidationStrategy struct {
			InvalidateOnError    bool
			InvalidateAfterClone bool
			InvalidateByTags     []string
		}

		// DefaultCacheInvalidationStrategy returns default invalidation strategy
		func DefaultCacheInvalidationStrategy() CacheInvalidationStrategy {
			return CacheInvalidationStrategy{
				InvalidateOnError:    false, // Don't invalidate on API errors
				InvalidateAfterClone: false, // Don't invalidate after successful clone
				InvalidateByTags:     []string{},
			}
		}

		// ApplyCacheInvalidation applies the invalidation strategy
		func (csc *CachedStreamingClient) ApplyCacheInvalidation(ctx context.Context, groupID string, strategy CacheInvalidationStrategy) {
			if len(strategy.InvalidateByTags) > 0 {
				for _, tag := range strategy.InvalidateByTags {
					csc.cacheManager.InvalidateByService(ctx, tag)
				}
			}

			if strategy.InvalidateAfterClone {
				csc.cacheManager.InvalidateByIdentifier(ctx, "gitlab", groupID)
			}
		}
	*/
}
