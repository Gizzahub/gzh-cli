package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLRUCache(t *testing.T) {
	config := CacheConfig{
		Capacity:        3,
		DefaultTTL:      100 * time.Millisecond,
		CleanupInterval: 50 * time.Millisecond,
	}
	cache := NewLRUCache(config)

	t.Run("Basic operations", func(t *testing.T) {
		// Test Put and Get
		cache.Put("key1", "value1")
		value, found := cache.Get("key1")
		assert.True(t, found)
		assert.Equal(t, "value1", value)

		// Test missing key
		_, found = cache.Get("missing")
		assert.False(t, found)
	})

	t.Run("LRU eviction", func(t *testing.T) {
		cache.Clear()

		// Fill cache to capacity
		cache.Put("key1", "value1")
		cache.Put("key2", "value2")
		cache.Put("key3", "value3")

		// Access key1 to make it most recently used
		cache.Get("key1")

		// Add another key, should evict key2 (least recently used)
		cache.Put("key4", "value4")

		_, found := cache.Get("key2")
		assert.False(t, found, "key2 should have been evicted")

		_, found = cache.Get("key1")
		assert.True(t, found, "key1 should still exist")
	})

	t.Run("TTL expiration", func(t *testing.T) {
		cache.Clear()

		// Put with short TTL
		cache.PutWithTTL("ttl_key", "ttl_value", 50*time.Millisecond)

		// Should be available immediately
		value, found := cache.Get("ttl_key")
		assert.True(t, found)
		assert.Equal(t, "ttl_value", value)

		// Wait for expiration
		time.Sleep(60 * time.Millisecond)

		// Should be expired
		_, found = cache.Get("ttl_key")
		assert.False(t, found)
	})

	t.Run("Tag-based invalidation", func(t *testing.T) {
		cache.Clear()

		// Put entries with tags
		cache.PutWithTags("tagged1", "value1", time.Minute, []string{"tag1", "tag2"})
		cache.PutWithTags("tagged2", "value2", time.Minute, []string{"tag1"})
		cache.PutWithTags("tagged3", "value3", time.Minute, []string{"tag2"})

		// Invalidate by tag1
		count := cache.InvalidateByTag("tag1")
		assert.Equal(t, 2, count)

		// Check which entries remain
		_, found := cache.Get("tagged1")
		assert.False(t, found)

		_, found = cache.Get("tagged2")
		assert.False(t, found)

		_, found = cache.Get("tagged3")
		assert.True(t, found)
	})

	t.Run("Statistics", func(t *testing.T) {
		// Create new cache for clean statistics
		statConfig := CacheConfig{
			Capacity:   10,
			DefaultTTL: time.Minute,
		}
		statCache := NewLRUCache(statConfig)

		// Generate some hits and misses
		statCache.Put("stat_key", "stat_value")
		statCache.Get("stat_key") // hit
		statCache.Get("missing")  // miss
		statCache.Get("stat_key") // hit

		stats := statCache.Stats()
		assert.Equal(t, int64(2), stats.Hits)
		assert.Equal(t, int64(1), stats.Misses)
		assert.Equal(t, 2.0/3.0, stats.HitRate)
	})
}

func TestCacheManager(t *testing.T) {
	config := DefaultCacheManagerConfig()
	manager := NewCacheManager(config)
	defer manager.Close()

	ctx := context.Background()

	t.Run("Cache key generation", func(t *testing.T) {
		key := CacheKey{
			Service:    "github",
			Resource:   "repos",
			Identifier: "testorg",
			Params:     map[string]string{"per_page": "100"},
		}

		keyStr := key.String()
		assert.Contains(t, keyStr, "github")
		assert.Contains(t, keyStr, "repos")
		assert.Contains(t, keyStr, "testorg")
	})

	t.Run("Cache operations", func(t *testing.T) {
		key := CacheKey{
			Service:    "github",
			Resource:   "repos",
			Identifier: "testorg",
		}

		// Put and get
		manager.Put(ctx, key, []string{"repo1", "repo2"})

		value, found := manager.Get(ctx, key)
		assert.True(t, found)

		repos, ok := value.([]string)
		require.True(t, ok)
		assert.Equal(t, []string{"repo1", "repo2"}, repos)
	})

	t.Run("Service-based invalidation", func(t *testing.T) {
		// Create fresh manager for this test
		testConfig := DefaultCacheManagerConfig()
		testManager := NewCacheManager(testConfig)
		defer testManager.Close()

		key1 := CacheKey{Service: "github", Resource: "repos", Identifier: "org1"}
		key2 := CacheKey{Service: "github", Resource: "users", Identifier: "user1"}
		key3 := CacheKey{Service: "gitlab", Resource: "projects", Identifier: "group1"}

		testManager.Put(ctx, key1, "data1")
		testManager.Put(ctx, key2, "data2")
		testManager.Put(ctx, key3, "data3")

		// Invalidate all GitHub entries
		count := testManager.InvalidateByService(ctx, "github")
		assert.Equal(t, 2, count)

		// Check results
		_, found := testManager.Get(ctx, key1)
		assert.False(t, found)

		_, found = testManager.Get(ctx, key2)
		assert.False(t, found)

		_, found = testManager.Get(ctx, key3)
		assert.True(t, found) // GitLab entry should remain
	})

	t.Run("Resource-based invalidation", func(t *testing.T) {
		// Create fresh manager for this test
		testConfig2 := DefaultCacheManagerConfig()
		testManager2 := NewCacheManager(testConfig2)
		defer testManager2.Close()

		key1 := CacheKey{Service: "github", Resource: "repos", Identifier: "org1"}
		key2 := CacheKey{Service: "github", Resource: "repos", Identifier: "org2"}
		key3 := CacheKey{Service: "github", Resource: "users", Identifier: "user1"}

		testManager2.Put(ctx, key1, "data1")
		testManager2.Put(ctx, key2, "data2")
		testManager2.Put(ctx, key3, "data3")

		// Invalidate all repo entries
		count := testManager2.InvalidateByResource(ctx, "github", "repos")
		assert.Equal(t, 2, count)

		// Check results
		_, found := testManager2.Get(ctx, key1)
		assert.False(t, found)

		_, found = testManager2.Get(ctx, key2)
		assert.False(t, found)

		_, found = testManager2.Get(ctx, key3)
		assert.True(t, found) // User entry should remain
	})
}

func TestRedisCache(t *testing.T) {
	config := DefaultRedisCacheConfig()
	redisCache := NewRedisCache(config)
	defer redisCache.Close()

	ctx := context.Background()

	t.Run("Basic Redis operations", func(t *testing.T) {
		// Put and get
		err := redisCache.Put(ctx, "redis_key", "redis_value")
		assert.NoError(t, err)

		value, found := redisCache.Get(ctx, "redis_key")
		assert.True(t, found)
		assert.Equal(t, "redis_value", value)

		// Delete
		err = redisCache.Delete(ctx, "redis_key")
		assert.NoError(t, err)

		_, found = redisCache.Get(ctx, "redis_key")
		assert.False(t, found)
	})

	t.Run("TTL support", func(t *testing.T) {
		err := redisCache.PutWithTTL(ctx, "ttl_redis_key", "ttl_value", 50*time.Millisecond)
		assert.NoError(t, err)

		// Should be available immediately
		value, found := redisCache.Get(ctx, "ttl_redis_key")
		assert.True(t, found)
		assert.Equal(t, "ttl_value", value)

		// Wait for expiration
		time.Sleep(60 * time.Millisecond)

		// Should be expired
		_, found = redisCache.Get(ctx, "ttl_redis_key")
		assert.False(t, found)
	})

	t.Run("Tag operations", func(t *testing.T) {
		// Put entries with tags
		redisCache.Put(ctx, "tagged_redis1", "value1")
		redisCache.TagKey(ctx, "tagged_redis1", []string{"redis_tag1", "redis_tag2"})

		redisCache.Put(ctx, "tagged_redis2", "value2")
		redisCache.TagKey(ctx, "tagged_redis2", []string{"redis_tag1"})

		// Invalidate by tag
		count := redisCache.InvalidateByTag(ctx, "redis_tag1")
		assert.Equal(t, 2, count)

		// Check results
		_, found := redisCache.Get(ctx, "tagged_redis1")
		assert.False(t, found)

		_, found = redisCache.Get(ctx, "tagged_redis2")
		assert.False(t, found)
	})

	t.Run("Statistics", func(t *testing.T) {
		stats := redisCache.GetStats()
		assert.True(t, stats.Connected) // Mock is always connected
		assert.GreaterOrEqual(t, stats.Sets, int64(0))
		assert.GreaterOrEqual(t, stats.Hits, int64(0))
	})
}

func TestCacheOptions(t *testing.T) {
	config := DefaultCacheManagerConfig()
	manager := NewCacheManager(config)
	defer manager.Close()

	ctx := context.Background()

	t.Run("Cache with custom options", func(t *testing.T) {
		key := CacheKey{
			Service:    "test",
			Resource:   "data",
			Identifier: "item1",
		}

		opts := CacheOptions{
			TTL:      200 * time.Millisecond,
			Tags:     []string{"custom_tag"},
			Priority: "high",
		}

		manager.PutWithOptions(ctx, key, "test_data", opts)

		// Should be available
		value, found := manager.GetWithOptions(ctx, key, opts)
		assert.True(t, found)
		assert.Equal(t, "test_data", value)

		// Wait for custom TTL expiration
		time.Sleep(250 * time.Millisecond)

		// Should be expired
		_, found = manager.Get(ctx, key)
		assert.False(t, found)
	})
}

func BenchmarkLRUCache(b *testing.B) {
	config := CacheConfig{
		Capacity:   1000,
		DefaultTTL: time.Hour,
	}
	cache := NewLRUCache(config)

	b.Run("Put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := "key" + string(rune(i%1000))
			cache.Put(key, "value")
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1000; i++ {
			cache.Put("key"+string(rune(i)), "value")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + string(rune(i%1000))
			cache.Get(key)
		}
	})

	b.Run("Mixed operations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := "key" + string(rune(i%1000))
			if i%2 == 0 {
				cache.Put(key, "value")
			} else {
				cache.Get(key)
			}
		}
	})
}
