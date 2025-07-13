package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// RedisCache interface for Redis-based caching
type RedisCache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Put(ctx context.Context, key string, value interface{}) error
	PutWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	InvalidateByTag(ctx context.Context, tag string) int
	TagKey(ctx context.Context, key string, tags []string) error
	GetStats() RedisCacheStats
	Close() error
}

// RedisCacheConfig configures Redis cache behavior
type RedisCacheConfig struct {
	Address      string
	Password     string
	DB           int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	KeyPrefix    string
	TagPrefix    string
}

// DefaultRedisCacheConfig returns default Redis configuration
func DefaultRedisCacheConfig() RedisCacheConfig {
	return RedisCacheConfig{
		Address:      "localhost:6379",
		Password:     "",
		DB:           0,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
		KeyPrefix:    "gzh:cache:",
		TagPrefix:    "gzh:tag:",
	}
}

// RedisCacheStats represents Redis cache statistics
type RedisCacheStats struct {
	Hits        int64         `json:"hits"`
	Misses      int64         `json:"misses"`
	Sets        int64         `json:"sets"`
	Deletes     int64         `json:"deletes"`
	Errors      int64         `json:"errors"`
	HitRate     float64       `json:"hit_rate"`
	AvgLatency  time.Duration `json:"avg_latency"`
	Connected   bool          `json:"connected"`
	MemoryUsage int64         `json:"memory_usage"`
}

// MockRedisCache provides a mock Redis implementation for testing/fallback
type MockRedisCache struct {
	data   map[string]CachedValue
	tags   map[string][]string // tag -> keys mapping
	stats  RedisCacheStats
	config RedisCacheConfig
}

// CachedValue represents a cached value with metadata
type CachedValue struct {
	Data      interface{} `json:"data"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
}

// NewRedisCache creates a new Redis cache instance
// For now, returns a mock implementation
func NewRedisCache(config RedisCacheConfig) RedisCache {
	return &MockRedisCache{
		data:   make(map[string]CachedValue),
		tags:   make(map[string][]string),
		config: config,
		stats: RedisCacheStats{
			Connected: true, // Mock is always "connected"
		},
	}
}

// Get retrieves a value from Redis cache
func (rc *MockRedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	fullKey := rc.config.KeyPrefix + key

	value, exists := rc.data[fullKey]
	if !exists {
		rc.stats.Misses++
		return nil, false
	}

	// Check expiration
	if !value.ExpiresAt.IsZero() && time.Now().After(value.ExpiresAt) {
		delete(rc.data, fullKey)
		rc.stats.Misses++
		return nil, false
	}

	rc.stats.Hits++
	return value.Data, true
}

// Put stores a value in Redis cache with default TTL
func (rc *MockRedisCache) Put(ctx context.Context, key string, value interface{}) error {
	return rc.PutWithTTL(ctx, key, value, 10*time.Minute) // Default TTL
}

// PutWithTTL stores a value with specific TTL
func (rc *MockRedisCache) PutWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := rc.config.KeyPrefix + key

	cached := CachedValue{
		Data:      value,
		CreatedAt: time.Now(),
	}

	if ttl > 0 {
		cached.ExpiresAt = time.Now().Add(ttl)
	}

	rc.data[fullKey] = cached
	rc.stats.Sets++
	return nil
}

// Delete removes a key from Redis cache
func (rc *MockRedisCache) Delete(ctx context.Context, key string) error {
	fullKey := rc.config.KeyPrefix + key

	if _, exists := rc.data[fullKey]; exists {
		delete(rc.data, fullKey)
		rc.stats.Deletes++
	}

	return nil
}

// InvalidateByTag removes all keys associated with a tag
func (rc *MockRedisCache) InvalidateByTag(ctx context.Context, tag string) int {
	tagKey := rc.config.TagPrefix + tag
	keys, exists := rc.tags[tagKey]
	if !exists {
		return 0
	}

	count := 0
	for _, key := range keys {
		if _, exists := rc.data[key]; exists {
			delete(rc.data, key)
			count++
		}
	}

	// Clear tag mapping
	delete(rc.tags, tagKey)
	rc.stats.Deletes += int64(count)

	return count
}

// TagKey associates a key with one or more tags
func (rc *MockRedisCache) TagKey(ctx context.Context, key string, tags []string) error {
	fullKey := rc.config.KeyPrefix + key

	for _, tag := range tags {
		tagKey := rc.config.TagPrefix + tag
		if rc.tags[tagKey] == nil {
			rc.tags[tagKey] = make([]string, 0)
		}

		// Add key to tag if not already present
		found := false
		for _, existingKey := range rc.tags[tagKey] {
			if existingKey == fullKey {
				found = true
				break
			}
		}

		if !found {
			rc.tags[tagKey] = append(rc.tags[tagKey], fullKey)
		}
	}

	return nil
}

// GetStats returns current Redis cache statistics
func (rc *MockRedisCache) GetStats() RedisCacheStats {
	stats := rc.stats

	// Calculate hit rate
	totalRequests := stats.Hits + stats.Misses
	if totalRequests > 0 {
		stats.HitRate = float64(stats.Hits) / float64(totalRequests)
	}

	// Estimate memory usage (simplified)
	stats.MemoryUsage = int64(len(rc.data) * 1024) // Rough estimate

	return stats
}

// Close cleans up Redis connections
func (rc *MockRedisCache) Close() error {
	// For mock implementation, just clear data
	rc.data = make(map[string]CachedValue)
	rc.tags = make(map[string][]string)
	return nil
}

// RealRedisCache would be the actual Redis implementation
// This is a placeholder for future Redis integration using go-redis or similar
type RealRedisCache struct {
	// client redis.Client // Would use go-redis or similar
	config RedisCacheConfig
	stats  RedisCacheStats
}

// NewRealRedisCache creates a real Redis cache (placeholder)
func NewRealRedisCache(config RedisCacheConfig) RedisCache {
	// This would initialize a real Redis client
	// For now, return mock implementation
	return NewRedisCache(config)
}

// Helper functions for Redis key management

// BuildRedisKey constructs a Redis key with proper prefixing
func BuildRedisKey(prefix, key string) string {
	return prefix + key
}

// ParseRedisKey extracts the original key from a prefixed Redis key
func ParseRedisKey(prefix, redisKey string) string {
	return strings.TrimPrefix(redisKey, prefix)
}

// SerializeValue converts a value to Redis-compatible format
func SerializeValue(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

// DeserializeValue converts Redis value back to original format
func DeserializeValue(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

// RedisCacheMiddleware provides caching middleware for HTTP-like patterns
type RedisCacheMiddleware struct {
	cache RedisCache
	ttl   time.Duration
}

// NewRedisCacheMiddleware creates caching middleware
func NewRedisCacheMiddleware(cache RedisCache, ttl time.Duration) *RedisCacheMiddleware {
	return &RedisCacheMiddleware{
		cache: cache,
		ttl:   ttl,
	}
}

// CacheWrapper wraps a function with caching
func (rcm *RedisCacheMiddleware) CacheWrapper(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := rcm.cache.Get(ctx, key); found {
		return value, nil
	}

	// Execute function and cache result
	result, err := fn()
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheErr := rcm.cache.PutWithTTL(ctx, key, result, rcm.ttl); cacheErr != nil {
		// Log cache error but don't fail the operation
		fmt.Printf("Failed to cache result: %v\n", cacheErr)
	}

	return result, nil
}

// Utility functions for cache key generation

// GenerateRepoListKey creates a cache key for repository lists
func GenerateRepoListKey(service, org string, params map[string]string) string {
	key := fmt.Sprintf("%s:repos:%s", service, org)
	if len(params) > 0 {
		// Add parameter hash for uniqueness
		paramStr, _ := json.Marshal(params)
		key += fmt.Sprintf(":%x", paramStr)
	}
	return key
}

// GenerateUserInfoKey creates a cache key for user information
func GenerateUserInfoKey(service, username string) string {
	return fmt.Sprintf("%s:user:%s", service, username)
}

// GenerateOrgInfoKey creates a cache key for organization information
func GenerateOrgInfoKey(service, org string) string {
	return fmt.Sprintf("%s:org:%s", service, org)
}
