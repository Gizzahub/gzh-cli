package recovery

import (
	"net/http"
	"time"
)

// HTTPClientFactory provides factory methods for creating resilient HTTP clients
type HTTPClientFactory struct {
	defaultConfig ResilientHTTPClientConfig
}

// NewHTTPClientFactory creates a factory with default configuration
func NewHTTPClientFactory() *HTTPClientFactory {
	return &HTTPClientFactory{
		defaultConfig: DefaultResilientHTTPClientConfig(),
	}
}

// NewHTTPClientFactoryWithConfig creates a factory with custom default configuration
func NewHTTPClientFactoryWithConfig(config ResilientHTTPClientConfig) *HTTPClientFactory {
	return &HTTPClientFactory{
		defaultConfig: config,
	}
}

// CreateGitHubClient creates an HTTP client optimized for GitHub API calls
func (f *HTTPClientFactory) CreateGitHubClient() *ResilientHTTPClient {
	config := f.defaultConfig

	// GitHub-specific optimizations
	config.Timeout = 45 * time.Second // GitHub can be slow sometimes
	config.MaxRetries = 3
	config.InitialDelay = 1 * time.Second
	config.MaxDelay = 10 * time.Second
	config.CircuitConfig.Name = "github-api"
	config.CircuitConfig.FailureThreshold = 5 // GitHub rate limiting is common
	config.CircuitConfig.Timeout = 60 * time.Second

	return NewResilientHTTPClient(config)
}

// CreateGitLabClient creates an HTTP client optimized for GitLab API calls
func (f *HTTPClientFactory) CreateGitLabClient() *ResilientHTTPClient {
	config := f.defaultConfig

	// GitLab-specific optimizations
	config.Timeout = 30 * time.Second
	config.MaxRetries = 3
	config.InitialDelay = 500 * time.Millisecond
	config.MaxDelay = 8 * time.Second
	config.CircuitConfig.Name = "gitlab-api"
	config.CircuitConfig.FailureThreshold = 4
	config.CircuitConfig.Timeout = 45 * time.Second

	return NewResilientHTTPClient(config)
}

// CreateGiteaClient creates an HTTP client optimized for Gitea API calls
func (f *HTTPClientFactory) CreateGiteaClient() *ResilientHTTPClient {
	config := f.defaultConfig

	// Gitea-specific optimizations (usually self-hosted, faster)
	config.Timeout = 20 * time.Second
	config.MaxRetries = 2
	config.InitialDelay = 300 * time.Millisecond
	config.MaxDelay = 5 * time.Second
	config.CircuitConfig.Name = "gitea-api"
	config.CircuitConfig.FailureThreshold = 3
	config.CircuitConfig.Timeout = 30 * time.Second

	return NewResilientHTTPClient(config)
}

// CreateGenericClient creates a standard resilient HTTP client
func (f *HTTPClientFactory) CreateGenericClient() *ResilientHTTPClient {
	return NewResilientHTTPClient(f.defaultConfig)
}

// CreateQuickClient creates a client optimized for fast, reliable endpoints
func (f *HTTPClientFactory) CreateQuickClient() *ResilientHTTPClient {
	config := f.defaultConfig

	// Quick client optimizations
	config.Timeout = 10 * time.Second
	config.MaxRetries = 1
	config.InitialDelay = 200 * time.Millisecond
	config.MaxDelay = 2 * time.Second
	config.UseCircuitBreaker = false // Fast-fail for quick clients

	return NewResilientHTTPClient(config)
}

// CreateLongRunningClient creates a client optimized for long-running operations
func (f *HTTPClientFactory) CreateLongRunningClient() *ResilientHTTPClient {
	config := f.defaultConfig

	// Long-running optimizations
	config.Timeout = 5 * time.Minute
	config.MaxRetries = 5
	config.InitialDelay = 2 * time.Second
	config.MaxDelay = 60 * time.Second
	config.CircuitConfig.Name = "long-running"
	config.CircuitConfig.FailureThreshold = 8
	config.CircuitConfig.Timeout = 2 * time.Minute

	return NewResilientHTTPClient(config)
}

// WrapHTTPClient wraps an existing http.Client with resilience features
func (f *HTTPClientFactory) WrapHTTPClient(client *http.Client) *ResilientHTTPClient {
	config := f.defaultConfig

	// Use existing client's timeout if available
	if client.Timeout > 0 {
		config.Timeout = client.Timeout
	}

	resilientClient := NewResilientHTTPClient(config)

	// Replace the base client while preserving resilience features
	resilientClient.client = client

	return resilientClient
}

// Global factory instance for convenience
var DefaultFactory = NewHTTPClientFactory()

// Convenience functions using the default factory

// NewGitHubClient creates a GitHub-optimized resilient HTTP client
func NewGitHubClient() *ResilientHTTPClient {
	return DefaultFactory.CreateGitHubClient()
}

// NewGitLabClient creates a GitLab-optimized resilient HTTP client
func NewGitLabClient() *ResilientHTTPClient {
	return DefaultFactory.CreateGitLabClient()
}

// NewGiteaClient creates a Gitea-optimized resilient HTTP client
func NewGiteaClient() *ResilientHTTPClient {
	return DefaultFactory.CreateGiteaClient()
}

// NewGenericClient creates a standard resilient HTTP client
func NewGenericClient() *ResilientHTTPClient {
	return DefaultFactory.CreateGenericClient()
}

// NewQuickClient creates a fast-fail resilient HTTP client
func NewQuickClient() *ResilientHTTPClient {
	return DefaultFactory.CreateQuickClient()
}

// NewLongRunningClient creates a resilient HTTP client for long operations
func NewLongRunningClient() *ResilientHTTPClient {
	return DefaultFactory.CreateLongRunningClient()
}
