package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Logger interface for dependency injection
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// MetricsCollector interface for dependency injection
type MetricsCollector interface {
	IncrementCounter(name string, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
	RecordGauge(name string, value float64, labels map[string]string)
}

// HTTPClientImpl implements the HTTPClient interface
type HTTPClientImpl struct {
	client           *http.Client
	logger           Logger
	metricsCollector MetricsCollector
	middleware       []Middleware
}

// HTTPClientConfig holds configuration for HTTP client
type HTTPClientConfig struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
	TLSHandshakeTimeout time.Duration
	UserAgent           string
	EnableMetrics       bool
}

// DefaultHTTPClientConfig returns default configuration
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout:             30 * time.Second,
		MaxIdleConns:        100,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		UserAgent:           "gzh-manager-go/1.0",
		EnableMetrics:       true,
	}
}

// NewHTTPClient creates a new HTTP client with dependencies
func NewHTTPClient(
	config *HTTPClientConfig,
	logger Logger,
	metricsCollector MetricsCollector,
) HTTPClient {
	if config == nil {
		config = DefaultHTTPClientConfig()
	}

	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		TLSHandshakeTimeout: config.TLSHandshakeTimeout,
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	return &HTTPClientImpl{
		client:           client,
		logger:           logger,
		metricsCollector: metricsCollector,
		middleware:       []Middleware{},
	}
}

// Do implements HTTPClient interface
func (c *HTTPClientImpl) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	c.logger.Debug("Making HTTP request", "method", req.Method, "url", req.URL.String())

	// Apply middleware
	for _, middleware := range c.middleware {
		req = middleware.ModifyRequest(ctx, req)
	}

	start := time.Now()
	resp, err := c.client.Do(req.WithContext(ctx))
	duration := time.Since(start)

	// Record metrics
	if c.metricsCollector != nil {
		labels := map[string]string{
			"method": req.Method,
			"host":   req.URL.Host,
		}
		if resp != nil {
			labels["status_code"] = fmt.Sprintf("%d", resp.StatusCode)
		}
		c.metricsCollector.RecordHistogram("http_request_duration", duration.Seconds(), labels)
		c.metricsCollector.IncrementCounter("http_requests_total", labels)
	}

	if err != nil {
		c.logger.Error("HTTP request failed", "method", req.Method, "url", req.URL.String(), "error", err)
		return nil, err
	}

	// Apply response middleware
	for _, middleware := range c.middleware {
		resp = middleware.ModifyResponse(ctx, resp)
	}

	return resp, nil
}

// Get implements HTTPClient interface
func (c *HTTPClientImpl) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req)
}

// Post implements HTTPClient interface
func (c *HTTPClientImpl) Post(ctx context.Context, url, contentType string, body interface{}) (*http.Response, error) {
	// Implementation would handle different body types
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return c.Do(ctx, req)
}

// Put implements HTTPClient interface
func (c *HTTPClientImpl) Put(ctx context.Context, url, contentType string, body interface{}) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return c.Do(ctx, req)
}

// Delete implements HTTPClient interface
func (c *HTTPClientImpl) Delete(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req)
}

// AddMiddleware implements HTTPClient interface
func (c *HTTPClientImpl) AddMiddleware(middleware Middleware) {
	c.middleware = append(c.middleware, middleware)
}

// RetryPolicyImpl implements the RetryPolicy interface
type RetryPolicyImpl struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
	logger     Logger
}

// RetryPolicyConfig holds configuration for retry policy
type RetryPolicyConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Backoff    string // "linear", "exponential", "constant"
}

// DefaultRetryPolicyConfig returns default configuration
func DefaultRetryPolicyConfig() *RetryPolicyConfig {
	return &RetryPolicyConfig{
		MaxRetries: 3,
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
		Backoff:    "exponential",
	}
}

// NewRetryPolicy creates a new retry policy with dependencies
func NewRetryPolicy(config *RetryPolicyConfig, logger Logger) RetryPolicy {
	if config == nil {
		config = DefaultRetryPolicyConfig()
	}

	return &RetryPolicyImpl{
		maxRetries: config.MaxRetries,
		baseDelay:  config.BaseDelay,
		maxDelay:   config.MaxDelay,
		logger:     logger,
	}
}

// ShouldRetry implements RetryPolicy interface
func (rp *RetryPolicyImpl) ShouldRetry(ctx context.Context, attempt int, err error) bool {
	if attempt >= rp.maxRetries {
		return false
	}

	// Implementation would check if error is retryable
	return true
}

// GetRetryDelay implements RetryPolicy interface
func (rp *RetryPolicyImpl) GetRetryDelay(ctx context.Context, attempt int) time.Duration {
	delay := rp.baseDelay * time.Duration(attempt)
	if delay > rp.maxDelay {
		delay = rp.maxDelay
	}
	return delay
}

// GetMaxRetries implements RetryPolicy interface
func (rp *RetryPolicyImpl) GetMaxRetries() int {
	return rp.maxRetries
}

// IsRetryableError implements RetryPolicy interface
func (rp *RetryPolicyImpl) IsRetryableError(err error) bool {
	// Implementation would determine if error is retryable
	return true
}

// RateLimiterImpl implements the RateLimiter interface
type RateLimiterImpl struct {
	tokens chan struct{}
	logger Logger
}

// RateLimiterConfig holds configuration for rate limiter
type RateLimiterConfig struct {
	RequestsPerSecond int
	BurstSize         int
}

// DefaultRateLimiterConfig returns default configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
	}
}

// NewRateLimiter creates a new rate limiter with dependencies
func NewRateLimiter(config *RateLimiterConfig, logger Logger) RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	tokens := make(chan struct{}, config.BurstSize)

	// Fill initial tokens
	for i := 0; i < config.BurstSize; i++ {
		tokens <- struct{}{}
	}

	// Start token refill goroutine
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(config.RequestsPerSecond))
		defer ticker.Stop()

		for range ticker.C {
			select {
			case tokens <- struct{}{}:
			default:
				// Bucket full, skip
			}
		}
	}()

	return &RateLimiterImpl{
		tokens: tokens,
		logger: logger,
	}
}

// Wait implements RateLimiter interface
func (rl *RateLimiterImpl) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Allow implements RateLimiter interface
func (rl *RateLimiterImpl) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// GetAvailableTokens implements RateLimiter interface
func (rl *RateLimiterImpl) GetAvailableTokens() int {
	return len(rl.tokens)
}

// CacheImpl implements the Cache interface
type CacheImpl struct {
	cache  map[string]*CacheEntry
	logger Logger
}

// CacheConfig holds configuration for cache
type CacheConfig struct {
	MaxSize int
	TTL     time.Duration
}

// DefaultCacheConfig returns default configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize: 1000,
		TTL:     5 * time.Minute,
	}
}

// NewCache creates a new cache with dependencies
func NewCache(config *CacheConfig, logger Logger) Cache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	return &CacheImpl{
		cache:  make(map[string]*CacheEntry),
		logger: logger,
	}
}

// Get implements Cache interface
func (c *CacheImpl) Get(ctx context.Context, key string) (*http.Response, bool) {
	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.CreatedAt) > entry.TTL {
		delete(c.cache, key)
		return nil, false
	}

	return entry.Response, true
}

// Set implements Cache interface
func (c *CacheImpl) Set(ctx context.Context, key string, response *http.Response, ttl time.Duration) {
	c.cache[key] = &CacheEntry{
		Response:  response,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}
}

// Delete implements Cache interface
func (c *CacheImpl) Delete(ctx context.Context, key string) {
	delete(c.cache, key)
}

// Clear implements Cache interface
func (c *CacheImpl) Clear(ctx context.Context) {
	c.cache = make(map[string]*CacheEntry)
}

// HTTPClientService implements the unified HTTP client service interface
type HTTPClientService struct {
	HTTPClient
	RetryPolicy
	RateLimiter
	Cache
}

// HTTPClientServiceConfig holds configuration for the HTTP client service
type HTTPClientServiceConfig struct {
	Client      *HTTPClientConfig
	Retry       *RetryPolicyConfig
	RateLimit   *RateLimiterConfig
	Cache       *CacheConfig
	EnableRetry bool
	EnableCache bool
}

// DefaultHTTPClientServiceConfig returns default configuration
func DefaultHTTPClientServiceConfig() *HTTPClientServiceConfig {
	return &HTTPClientServiceConfig{
		Client:      DefaultHTTPClientConfig(),
		Retry:       DefaultRetryPolicyConfig(),
		RateLimit:   DefaultRateLimiterConfig(),
		Cache:       DefaultCacheConfig(),
		EnableRetry: true,
		EnableCache: true,
	}
}

// NewHTTPClientService creates a new HTTP client service with all dependencies
func NewHTTPClientService(
	config *HTTPClientServiceConfig,
	logger Logger,
	metricsCollector MetricsCollector,
) HTTPClientService {
	if config == nil {
		config = DefaultHTTPClientServiceConfig()
	}

	httpClient := NewHTTPClient(config.Client, logger, metricsCollector)
	retryPolicy := NewRetryPolicy(config.Retry, logger)
	rateLimiter := NewRateLimiter(config.RateLimit, logger)
	cache := NewCache(config.Cache, logger)

	return &HTTPClientService{
		HTTPClient:  httpClient,
		RetryPolicy: retryPolicy,
		RateLimiter: rateLimiter,
		Cache:       cache,
	}
}

// ServiceDependencies holds all the dependencies needed for HTTP client services
type ServiceDependencies struct {
	Logger           Logger
	MetricsCollector MetricsCollector
}

// NewServiceDependencies creates a default set of service dependencies
func NewServiceDependencies(logger Logger, metricsCollector MetricsCollector) *ServiceDependencies {
	return &ServiceDependencies{
		Logger:           logger,
		MetricsCollector: metricsCollector,
	}
}
