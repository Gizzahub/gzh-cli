// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProviderRegistry manages provider instances with caching and lifecycle management.
type ProviderRegistry struct {
	mu       sync.RWMutex
	factory  *ProviderFactory
	cache    map[string]*CachedProvider
	config   RegistryConfig
	stopCh   chan struct{}
	stopOnce sync.Once
}

// CachedProvider wraps a provider instance with metadata.
type CachedProvider struct {
	Provider   GitProvider            `json:"-"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	CreatedAt  time.Time              `json:"created_at"`
	LastUsed   time.Time              `json:"last_used"`
	UsageCount int64                  `json:"usage_count"`
	IsHealthy  bool                   `json:"is_healthy"`
	LastCheck  time.Time              `json:"last_check"`
	LastError  string                 `json:"last_error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RegistryConfig represents configuration for the provider registry.
type RegistryConfig struct {
	EnableCaching       bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTimeout        time.Duration `json:"cache_timeout" yaml:"cache_timeout"`
	EnableHealthChecks  bool          `json:"enable_health_checks" yaml:"enable_health_checks"`
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	MaxCacheSize        int           `json:"max_cache_size" yaml:"max_cache_size"`
	EnableMetrics       bool          `json:"enable_metrics" yaml:"enable_metrics"`
	AutoCleanup         bool          `json:"auto_cleanup" yaml:"auto_cleanup"`
}

// NewProviderRegistry creates a new provider registry.
func NewProviderRegistry(factory *ProviderFactory, config RegistryConfig) *ProviderRegistry {
	if config.CacheTimeout == 0 {
		config.CacheTimeout = 30 * time.Minute
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 5 * time.Minute
	}
	if config.MaxCacheSize == 0 {
		config.MaxCacheSize = 100
	}

	registry := &ProviderRegistry{
		factory: factory,
		cache:   make(map[string]*CachedProvider),
		config:  config,
		stopCh:  make(chan struct{}),
	}

	// Start background tasks
	if config.EnableHealthChecks {
		go registry.healthCheckLoop()
	}
	if config.AutoCleanup {
		go registry.cleanupLoop()
	}

	return registry
}

// GetProvider retrieves a provider instance, creating it if necessary.
func (r *ProviderRegistry) GetProvider(name string) (GitProvider, error) {
	// Check cache first
	if r.config.EnableCaching {
		if cached := r.getCachedProvider(name); cached != nil {
			r.updateUsage(cached)
			return cached.Provider, nil
		}
	}

	// Create new provider instance
	provider, err := r.factory.CreateProvider(name)
	if err != nil {
		return nil, WrapError("registry", "get_provider", err)
	}

	// Cache the provider if caching is enabled
	if r.config.EnableCaching {
		config, _ := r.factory.GetConfig(name)
		providerType := ""
		if config != nil {
			providerType = config.Type
		}

		cached := &CachedProvider{
			Provider:   provider,
			Name:       name,
			Type:       providerType,
			CreatedAt:  time.Now(),
			LastUsed:   time.Now(),
			UsageCount: 1,
			IsHealthy:  true,
			LastCheck:  time.Now(),
			Metadata:   make(map[string]interface{}),
		}

		r.cacheProvider(name, cached)
	}

	return provider, nil
}

// GetProviderByType creates a provider instance by type with temporary configuration.
func (r *ProviderRegistry) GetProviderByType(providerType string, config *ProviderConfig) (GitProvider, error) {
	return r.factory.CreateProviderByType(providerType, config)
}

// ListProviders returns a list of available provider names.
func (r *ProviderRegistry) ListProviders() []string {
	return r.factory.ListProviders()
}

// ListCachedProviders returns information about cached providers.
func (r *ProviderRegistry) ListCachedProviders() []*CachedProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cached := make([]*CachedProvider, 0, len(r.cache))
	for _, provider := range r.cache {
		// Create a copy without the actual provider instance
		info := &CachedProvider{
			Name:       provider.Name,
			Type:       provider.Type,
			CreatedAt:  provider.CreatedAt,
			LastUsed:   provider.LastUsed,
			UsageCount: provider.UsageCount,
			IsHealthy:  provider.IsHealthy,
			LastCheck:  provider.LastCheck,
			LastError:  provider.LastError,
			Metadata:   provider.Metadata,
		}
		cached = append(cached, info)
	}

	return cached
}

// ExecuteAcrossProviders executes a function across all available providers.
func (r *ProviderRegistry) ExecuteAcrossProviders(ctx context.Context, fn func(string, GitProvider) error) error {
	providers := r.ListProviders()

	for _, name := range providers {
		provider, err := r.GetProvider(name)
		if err != nil {
			return fmt.Errorf("failed to get provider %s: %w", name, err)
		}

		if err := fn(name, provider); err != nil {
			return fmt.Errorf("execution failed for provider %s: %w", name, err)
		}
	}

	return nil
}

// ExecuteAcrossProvidersParallel executes a function across providers in parallel.
func (r *ProviderRegistry) ExecuteAcrossProvidersParallel(ctx context.Context, fn func(string, GitProvider) error) error {
	providers := r.ListProviders()

	type result struct {
		name string
		err  error
	}

	results := make(chan result, len(providers))

	for _, name := range providers {
		go func(providerName string) {
			provider, err := r.GetProvider(providerName)
			if err != nil {
				results <- result{providerName, fmt.Errorf("failed to get provider: %w", err)}
				return
			}

			err = fn(providerName, provider)
			results <- result{providerName, err}
		}(name)
	}

	// Collect results
	var errors []error
	for i := 0; i < len(providers); i++ {
		select {
		case res := <-results:
			if res.err != nil {
				errors = append(errors, fmt.Errorf("provider %s: %w", res.name, res.err))
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("execution failed for %d providers: %v", len(errors), errors)
	}

	return nil
}

// InvalidateCache removes a provider from cache.
func (r *ProviderRegistry) InvalidateCache(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cache, name)
}

// ClearCache removes all providers from cache.
func (r *ProviderRegistry) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]*CachedProvider)
}

// GetCacheStats returns cache statistics.
func (r *ProviderRegistry) GetCacheStats() CacheStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := CacheStats{
		TotalCached:    len(r.cache),
		HealthyCount:   0,
		UnhealthyCount: 0,
		TotalUsage:     0,
	}

	for _, cached := range r.cache {
		if cached.IsHealthy {
			stats.HealthyCount++
		} else {
			stats.UnhealthyCount++
		}
		stats.TotalUsage += cached.UsageCount
	}

	return stats
}

// HealthCheckAll performs health checks on all cached providers.
func (r *ProviderRegistry) HealthCheckAll(ctx context.Context) error {
	r.mu.RLock()
	providers := make([]*CachedProvider, 0, len(r.cache))
	for _, cached := range r.cache {
		providers = append(providers, cached)
	}
	r.mu.RUnlock()

	for _, cached := range providers {
		r.performHealthCheck(ctx, cached)
	}

	return nil
}

// Close shuts down the registry and stops background tasks.
func (r *ProviderRegistry) Close() error {
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
	return nil
}

// Private methods

func (r *ProviderRegistry) getCachedProvider(name string) *CachedProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cached, exists := r.cache[name]
	if !exists {
		return nil
	}

	// Check if cache has expired
	if time.Since(cached.CreatedAt) > r.config.CacheTimeout {
		go r.InvalidateCache(name) // Remove asynchronously
		return nil
	}

	return cached
}

func (r *ProviderRegistry) cacheProvider(name string, cached *CachedProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check cache size limit
	if len(r.cache) >= r.config.MaxCacheSize {
		r.evictOldestProvider()
	}

	r.cache[name] = cached
}

func (r *ProviderRegistry) updateUsage(cached *CachedProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cached.LastUsed = time.Now()
	cached.UsageCount++
}

func (r *ProviderRegistry) evictOldestProvider() {
	var oldestName string
	var oldestTime time.Time = time.Now()

	for name, cached := range r.cache {
		if cached.LastUsed.Before(oldestTime) {
			oldestTime = cached.LastUsed
			oldestName = name
		}
	}

	if oldestName != "" {
		delete(r.cache, oldestName)
	}
}

func (r *ProviderRegistry) healthCheckLoop() {
	ticker := time.NewTicker(r.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			r.HealthCheckAll(ctx)
			cancel()
		case <-r.stopCh:
			return
		}
	}
}

func (r *ProviderRegistry) cleanupLoop() {
	ticker := time.NewTicker(time.Hour) // Run cleanup every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.performCleanup()
		case <-r.stopCh:
			return
		}
	}
}

func (r *ProviderRegistry) performCleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	expiredProviders := make([]string, 0)

	for name, cached := range r.cache {
		// Remove providers that haven't been used for twice the cache timeout
		if now.Sub(cached.LastUsed) > r.config.CacheTimeout*2 {
			expiredProviders = append(expiredProviders, name)
		}
	}

	for _, name := range expiredProviders {
		delete(r.cache, name)
	}
}

func (r *ProviderRegistry) performHealthCheck(ctx context.Context, cached *CachedProvider) {
	if cached.Provider == nil {
		return
	}

	// Perform health check
	healthStatus, err := cached.Provider.HealthCheck(ctx)

	r.mu.Lock()
	defer r.mu.Unlock()

	cached.LastCheck = time.Now()

	if err != nil {
		cached.IsHealthy = false
		cached.LastError = err.Error()
	} else if healthStatus != nil {
		cached.IsHealthy = healthStatus.Status == HealthStatusHealthy
		if !cached.IsHealthy {
			cached.LastError = healthStatus.Message
		} else {
			cached.LastError = ""
		}
	}
}

// CacheStats represents cache statistics.
type CacheStats struct {
	TotalCached    int   `json:"total_cached"`
	HealthyCount   int   `json:"healthy_count"`
	UnhealthyCount int   `json:"unhealthy_count"`
	TotalUsage     int64 `json:"total_usage"`
}

// ProviderMetadata represents metadata about a provider.
type ProviderMetadata struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Capabilities []Capability           `json:"capabilities"`
	BaseURL      string                 `json:"base_url"`
	IsHealthy    bool                   `json:"is_healthy"`
	LastCheck    time.Time              `json:"last_check"`
	Metrics      *ProviderMetrics       `json:"metrics,omitempty"`
	Extra        map[string]interface{} `json:"extra,omitempty"`
}

// GetProviderMetadata returns metadata about a provider.
func (r *ProviderRegistry) GetProviderMetadata(name string) (*ProviderMetadata, error) {
	provider, err := r.GetProvider(name)
	if err != nil {
		return nil, err
	}

	metadata := &ProviderMetadata{
		Name:         name,
		Type:         provider.GetName(),
		Capabilities: provider.GetCapabilities(),
		BaseURL:      provider.GetBaseURL(),
		Extra:        make(map[string]interface{}),
	}

	// Get cached information if available
	if cached := r.getCachedProvider(name); cached != nil {
		metadata.IsHealthy = cached.IsHealthy
		metadata.LastCheck = cached.LastCheck
	}

	// Get metrics if available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if metrics, err := provider.GetMetrics(ctx); err == nil {
		metadata.Metrics = metrics
	}

	return metadata, nil
}

// GetAllProviderMetadata returns metadata for all available providers.
func (r *ProviderRegistry) GetAllProviderMetadata() ([]*ProviderMetadata, error) {
	providers := r.ListProviders()
	metadata := make([]*ProviderMetadata, 0, len(providers))

	for _, name := range providers {
		if meta, err := r.GetProviderMetadata(name); err == nil {
			metadata = append(metadata, meta)
		}
	}

	return metadata, nil
}
