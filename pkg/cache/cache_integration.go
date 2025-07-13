package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CacheKey represents a structured cache key
type CacheKey struct {
	Service    string // "github", "gitlab"
	Resource   string // "repos", "projects", "user"
	Identifier string // org name, user ID, etc.
	Params     map[string]string
}

// String generates a unique string representation of the cache key
func (k CacheKey) String() string {
	var parts []string
	parts = append(parts, k.Service, k.Resource, k.Identifier)

	if len(k.Params) > 0 {
		paramBytes, _ := json.Marshal(k.Params)
		hash := md5.Sum(paramBytes)
		parts = append(parts, fmt.Sprintf("%x", hash))
	}

	return strings.Join(parts, ":")
}

// CacheManager manages multiple cache backends
type CacheManager struct {
	localCache *LRUCache
	redisCache RedisCache
	config     CacheManagerConfig
}

// CacheManagerConfig configures cache behavior
type CacheManagerConfig struct {
	UseRedis         bool
	LocalCacheConfig CacheConfig
	RedisCacheConfig RedisCacheConfig
	DefaultTTL       time.Duration
	TagPrefix        string
}

// DefaultCacheManagerConfig returns default cache manager configuration
func DefaultCacheManagerConfig() CacheManagerConfig {
	return CacheManagerConfig{
		UseRedis:         false,
		LocalCacheConfig: DefaultCacheConfig(),
		RedisCacheConfig: DefaultRedisCacheConfig(),
		DefaultTTL:       10 * time.Minute,
		TagPrefix:        "gzh",
	}
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config CacheManagerConfig) *CacheManager {
	manager := &CacheManager{
		localCache: NewLRUCache(config.LocalCacheConfig),
		config:     config,
	}

	if config.UseRedis {
		manager.redisCache = NewRedisCache(config.RedisCacheConfig)
	}

	return manager
}

// Get retrieves a value from cache (local first, then Redis)
func (cm *CacheManager) Get(ctx context.Context, key CacheKey) (interface{}, bool) {
	keyStr := key.String()

	// Try local cache first
	if value, found := cm.localCache.Get(keyStr); found {
		return value, true
	}

	// Try Redis if enabled
	if cm.config.UseRedis && cm.redisCache != nil {
		if value, found := cm.redisCache.Get(ctx, keyStr); found {
			// Store in local cache for faster access
			cm.localCache.PutWithTTL(keyStr, value, cm.config.DefaultTTL)
			return value, true
		}
	}

	return nil, false
}

// Put stores a value in cache (both local and Redis if enabled)
func (cm *CacheManager) Put(ctx context.Context, key CacheKey, value interface{}) {
	cm.PutWithTTL(ctx, key, value, cm.config.DefaultTTL)
}

// PutWithTTL stores a value with specific TTL
func (cm *CacheManager) PutWithTTL(ctx context.Context, key CacheKey, value interface{}, ttl time.Duration) {
	keyStr := key.String()
	tags := cm.generateTags(key)

	// Store in local cache
	cm.localCache.PutWithTags(keyStr, value, ttl, tags)

	// Store in Redis if enabled
	if cm.config.UseRedis && cm.redisCache != nil {
		cm.redisCache.PutWithTTL(ctx, keyStr, value, ttl)
		cm.redisCache.TagKey(ctx, keyStr, tags)
	}
}

// InvalidateByService invalidates all cache entries for a service
func (cm *CacheManager) InvalidateByService(ctx context.Context, service string) int {
	tag := fmt.Sprintf("%s:service:%s", cm.config.TagPrefix, service)

	count := cm.localCache.InvalidateByTag(tag)

	if cm.config.UseRedis && cm.redisCache != nil {
		count += cm.redisCache.InvalidateByTag(ctx, tag)
	}

	return count
}

// InvalidateByResource invalidates cache entries for a specific resource
func (cm *CacheManager) InvalidateByResource(ctx context.Context, service, resource string) int {
	tag := fmt.Sprintf("%s:resource:%s:%s", cm.config.TagPrefix, service, resource)

	count := cm.localCache.InvalidateByTag(tag)

	if cm.config.UseRedis && cm.redisCache != nil {
		count += cm.redisCache.InvalidateByTag(ctx, tag)
	}

	return count
}

// InvalidateByIdentifier invalidates cache entries for a specific identifier
func (cm *CacheManager) InvalidateByIdentifier(ctx context.Context, service, identifier string) int {
	tag := fmt.Sprintf("%s:identifier:%s:%s", cm.config.TagPrefix, service, identifier)

	count := cm.localCache.InvalidateByTag(tag)

	if cm.config.UseRedis && cm.redisCache != nil {
		count += cm.redisCache.InvalidateByTag(ctx, tag)
	}

	return count
}

// generateTags creates cache tags for efficient invalidation
func (cm *CacheManager) generateTags(key CacheKey) []string {
	prefix := cm.config.TagPrefix
	return []string{
		fmt.Sprintf("%s:service:%s", prefix, key.Service),
		fmt.Sprintf("%s:resource:%s:%s", prefix, key.Service, key.Resource),
		fmt.Sprintf("%s:identifier:%s:%s", prefix, key.Service, key.Identifier),
	}
}

// GetStats returns combined cache statistics
func (cm *CacheManager) GetStats() CacheManagerStats {
	localStats := cm.localCache.Stats()

	stats := CacheManagerStats{
		Local: localStats,
		Redis: RedisCacheStats{}, // Default empty stats
	}

	if cm.config.UseRedis && cm.redisCache != nil {
		stats.Redis = cm.redisCache.GetStats()
	}

	return stats
}

// CacheManagerStats represents combined cache statistics
type CacheManagerStats struct {
	Local CacheStats      `json:"local"`
	Redis RedisCacheStats `json:"redis"`
}

// Close cleans up cache resources
func (cm *CacheManager) Close() error {
	if cm.config.UseRedis && cm.redisCache != nil {
		return cm.redisCache.Close()
	}
	return nil
}

// CacheOptions represents options for individual cache operations
type CacheOptions struct {
	TTL      time.Duration
	Tags     []string
	Priority string // "high", "medium", "low"
}

// GetWithOptions retrieves with custom options
func (cm *CacheManager) GetWithOptions(ctx context.Context, key CacheKey, opts CacheOptions) (interface{}, bool) {
	// For now, options don't affect retrieval, but could be used for metrics
	return cm.Get(ctx, key)
}

// PutWithOptions stores with custom options
func (cm *CacheManager) PutWithOptions(ctx context.Context, key CacheKey, value interface{}, opts CacheOptions) {
	ttl := opts.TTL
	if ttl == 0 {
		ttl = cm.config.DefaultTTL
	}

	keyStr := key.String()
	tags := cm.generateTags(key)

	// Add custom tags
	if len(opts.Tags) > 0 {
		tags = append(tags, opts.Tags...)
	}

	// Store with custom TTL and tags
	cm.localCache.PutWithTags(keyStr, value, ttl, tags)

	if cm.config.UseRedis && cm.redisCache != nil {
		cm.redisCache.PutWithTTL(ctx, keyStr, value, ttl)
		cm.redisCache.TagKey(ctx, keyStr, tags)
	}
}
