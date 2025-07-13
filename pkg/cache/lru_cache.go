package cache

import (
	"container/list"
	"sync"
	"time"
)

// LRUCache represents a thread-safe LRU cache with TTL support
type LRUCache struct {
	capacity  int
	cache     map[string]*cacheEntry
	evictList *list.List
	mu        sync.RWMutex

	// TTL support
	defaultTTL time.Duration

	// Statistics
	hits      int64
	misses    int64
	evictions int64
}

// cacheEntry represents a cache entry with value and metadata
type cacheEntry struct {
	key         string
	value       interface{}
	element     *list.Element
	createdAt   time.Time
	ttl         time.Duration
	accessCount int64
	tags        []string
}

// CacheConfig represents configuration for LRU cache
type CacheConfig struct {
	Capacity        int
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
}

// DefaultCacheConfig returns default configuration for LRU cache
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Capacity:        1000,
		DefaultTTL:      10 * time.Minute,
		CleanupInterval: 5 * time.Minute,
	}
}

// NewLRUCache creates a new LRU cache with the given capacity and TTL
func NewLRUCache(config CacheConfig) *LRUCache {
	cache := &LRUCache{
		capacity:   config.Capacity,
		cache:      make(map[string]*cacheEntry),
		evictList:  list.New(),
		defaultTTL: config.DefaultTTL,
	}

	// Start cleanup goroutine for expired entries
	if config.CleanupInterval > 0 {
		go cache.cleanupExpired(config.CleanupInterval)
	}

	return cache
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// Check if entry has expired
	if c.isExpired(entry) {
		c.mu.Lock()
		c.removeElement(entry)
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.evictList.MoveToFront(entry.element)
	entry.accessCount++
	c.hits++
	c.mu.Unlock()

	return entry.value, true
}

// Put adds a value to the cache
func (c *LRUCache) Put(key string, value interface{}) {
	c.PutWithTTL(key, value, c.defaultTTL)
}

// PutWithTTL adds a value to the cache with a specific TTL
func (c *LRUCache) PutWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if entry, exists := c.cache[key]; exists {
		// Update existing entry
		entry.value = value
		entry.createdAt = time.Now()
		entry.ttl = ttl
		c.evictList.MoveToFront(entry.element)
		return
	}

	// Create new entry
	entry := &cacheEntry{
		key:         key,
		value:       value,
		createdAt:   time.Now(),
		ttl:         ttl,
		accessCount: 0,
	}

	// Add to front of eviction list
	entry.element = c.evictList.PushFront(key)
	c.cache[key] = entry

	// Check if we need to evict
	if c.evictList.Len() > c.capacity {
		c.evictOldest()
	}
}

// PutWithTags adds a value to the cache with tags for tag-based invalidation
func (c *LRUCache) PutWithTags(key string, value interface{}, ttl time.Duration, tags []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if entry, exists := c.cache[key]; exists {
		// Update existing entry
		entry.value = value
		entry.createdAt = time.Now()
		entry.ttl = ttl
		entry.tags = tags
		c.evictList.MoveToFront(entry.element)
		return
	}

	// Create new entry
	entry := &cacheEntry{
		key:         key,
		value:       value,
		createdAt:   time.Now(),
		ttl:         ttl,
		accessCount: 0,
		tags:        tags,
	}

	// Add to front of eviction list
	entry.element = c.evictList.PushFront(key)
	c.cache[key] = entry

	// Check if we need to evict
	if c.evictList.Len() > c.capacity {
		c.evictOldest()
	}
}

// Delete removes a key from the cache
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.cache[key]; exists {
		c.removeElement(entry)
		return true
	}
	return false
}

// InvalidateByTag removes all entries with the specified tag
func (c *LRUCache) InvalidateByTag(tag string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var toRemove []*cacheEntry

	// Find all entries with the tag
	for _, entry := range c.cache {
		for _, entryTag := range entry.tags {
			if entryTag == tag {
				toRemove = append(toRemove, entry)
				break
			}
		}
	}

	// Remove them
	for _, entry := range toRemove {
		c.removeElement(entry)
	}

	return len(toRemove)
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	c.evictList.Init()
}

// Size returns the current number of items in the cache
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalRequests := c.hits + c.misses
	var hitRate float64
	if totalRequests > 0 {
		hitRate = float64(c.hits) / float64(totalRequests)
	}

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		HitRate:   hitRate,
		Size:      len(c.cache),
		Capacity:  c.capacity,
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	HitRate   float64 `json:"hit_rate"`
	Size      int     `json:"size"`
	Capacity  int     `json:"capacity"`
}

// Keys returns all keys in the cache (for debugging)
func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	return keys
}

// isExpired checks if a cache entry has expired
func (c *LRUCache) isExpired(entry *cacheEntry) bool {
	if entry.ttl <= 0 {
		return false // No expiration
	}
	return time.Since(entry.createdAt) > entry.ttl
}

// evictOldest removes the oldest entry from the cache
func (c *LRUCache) evictOldest() {
	oldest := c.evictList.Back()
	if oldest != nil {
		key := oldest.Value.(string)
		if entry, exists := c.cache[key]; exists {
			c.removeElement(entry)
			c.evictions++
		}
	}
}

// removeElement removes an entry from both the cache map and eviction list
func (c *LRUCache) removeElement(entry *cacheEntry) {
	c.evictList.Remove(entry.element)
	delete(c.cache, entry.key)
}

// cleanupExpired runs in a goroutine to periodically clean up expired entries
func (c *LRUCache) cleanupExpired(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.removeExpiredEntries()
	}
}

// removeExpiredEntries removes all expired entries from the cache
func (c *LRUCache) removeExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var toRemove []*cacheEntry

	// Find expired entries
	for _, entry := range c.cache {
		if c.isExpired(entry) {
			toRemove = append(toRemove, entry)
		}
	}

	// Remove them
	for _, entry := range toRemove {
		c.removeElement(entry)
		c.evictions++
	}
}
