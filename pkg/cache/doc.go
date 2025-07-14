// Package cache provides caching implementations and utilities
// for the GZH Manager system.
//
// This package implements various caching strategies to improve
// performance, reduce external API calls, and provide efficient
// data storage and retrieval mechanisms.
//
// Key Components:
//
// LRU Cache:
//   - Least Recently Used eviction policy
//   - Configurable size limits
//   - Thread-safe operations
//   - TTL-based expiration
//
// Redis Cache:
//   - Distributed caching with Redis
//   - Cluster support and failover
//   - Pub/Sub for cache invalidation
//   - Persistence and durability options
//
// Cache Integration:
//   - Multi-level caching hierarchy
//   - Cache-aside and write-through patterns
//   - Automatic cache warming
//   - Cache coherence and consistency
//
// Features:
//   - Configurable eviction policies
//   - Cache hit/miss metrics
//   - Memory usage monitoring
//   - Cache performance analytics
//   - Automatic cache optimization
//
// Cache Types:
//   - In-memory cache for fast access
//   - Distributed cache for scalability
//   - Persistent cache for durability
//   - Hybrid cache for flexibility
//
// Example usage:
//
//	cache := cache.NewLRUCache(1000)
//	cache.Set("key", value, ttl)
//	value, found := cache.Get("key")
//
//	redisCache := cache.NewRedisCache(config)
//	err := redisCache.Set(ctx, "key", value, ttl)
//	value, err := redisCache.Get(ctx, "key")
//
// The package provides flexible caching solutions that improve
// application performance and reduce resource consumption.
package cache
