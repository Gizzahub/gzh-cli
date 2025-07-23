// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package httpclient

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/constants"
)

// SecureClientConfig defines configuration for secure HTTP clients.
type SecureClientConfig struct {
	// Timeout settings
	Timeout         time.Duration
	DialTimeout     time.Duration
	KeepAlive       time.Duration
	IdleConnTimeout time.Duration

	// TLS settings
	MinTLSVersion      uint16
	InsecureSkipVerify bool

	// Connection limits
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int

	// Retry settings
	RetryCount int
	RetryDelay time.Duration

	// Headers
	UserAgent      string
	DefaultHeaders map[string]string
}

// DefaultSecureClientConfig returns a secure default configuration.
func DefaultSecureClientConfig() *SecureClientConfig {
	return &SecureClientConfig{
		// Conservative timeouts to prevent resource exhaustion
		Timeout:         constants.DefaultHTTPTimeout,
		DialTimeout:     constants.DefaultDialTimeout,
		KeepAlive:       constants.DefaultKeepAlive,
		IdleConnTimeout: constants.DefaultIdleConnTimeout,

		// Secure TLS configuration
		MinTLSVersion:      tls.VersionTLS12,
		InsecureSkipVerify: false,

		// Reasonable connection limits
		MaxIdleConns:        constants.MaxIdleConnections,
		MaxIdleConnsPerHost: constants.MaxIdleConnectionsPerHost,
		MaxConnsPerHost:     constants.MaxConnectionsPerHost,

		// Retry configuration
		RetryCount: constants.DefaultMaxRetries,
		RetryDelay: constants.RetryDelay,

		// User agent identification
		UserAgent: "gzh-manager-go/1.0.0",
		DefaultHeaders: map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json",
		},
	}
}

// GitHubClientConfig returns optimized configuration for GitHub API.
func GitHubClientConfig() *SecureClientConfig {
	config := DefaultSecureClientConfig()
	config.UserAgent = "gzh-manager-go/1.0.0 (GitHub API Client)"
	config.MaxIdleConnsPerHost = constants.GitHubMaxIdleConnectionsPerHost
	config.Timeout = constants.LongHTTPTimeout // GitHub operations can be slow
	return config
}

// GitLabClientConfig returns optimized configuration for GitLab API.
func GitLabClientConfig() *SecureClientConfig {
	config := DefaultSecureClientConfig()
	config.UserAgent = "gzh-manager-go/1.0.0 (GitLab API Client)"
	config.MaxIdleConnsPerHost = constants.GitLabMaxIdleConnectionsPerHost
	config.Timeout = constants.MediumHTTPTimeout
	return config
}

// GiteaClientConfig returns optimized configuration for Gitea API.
func GiteaClientConfig() *SecureClientConfig {
	config := DefaultSecureClientConfig()
	config.UserAgent = "gzh-manager-go/1.0.0 (Gitea API Client)"
	config.MaxIdleConnsPerHost = constants.GiteaMaxIdleConnectionsPerHost
	config.Timeout = constants.DefaultHTTPTimeout
	return config
}

// SecureHTTPClientFactory creates secure HTTP clients.
type SecureHTTPClientFactory struct {
	config *SecureClientConfig
}

// NewSecureHTTPClientFactory creates a new secure HTTP client factory.
func NewSecureHTTPClientFactory(config *SecureClientConfig) *SecureHTTPClientFactory {
	if config == nil {
		config = DefaultSecureClientConfig()
	}
	return &SecureHTTPClientFactory{config: config}
}

// CreateClient creates a new secure HTTP client.
func (f *SecureHTTPClientFactory) CreateClient() *http.Client {
	// Create custom transport with security settings
	transport := &http.Transport{
		// Connection settings
		DialContext: (&net.Dialer{
			Timeout:   f.config.DialTimeout,
			KeepAlive: f.config.KeepAlive,
			DualStack: true,
		}).DialContext,

		// Connection pooling
		MaxIdleConns:        f.config.MaxIdleConns,
		MaxIdleConnsPerHost: f.config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     f.config.MaxConnsPerHost,
		IdleConnTimeout:     f.config.IdleConnTimeout,

		// TLS configuration
		TLSClientConfig: &tls.Config{
			MinVersion:         f.config.MinTLSVersion,
			InsecureSkipVerify: f.config.InsecureSkipVerify, //nolint:gosec // Configurable for development environments
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			},
		},

		// Security headers
		DisableCompression: false,
		ForceAttemptHTTP2:  true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   f.config.Timeout,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			// Limit redirects to prevent redirect loops
			if len(via) >= constants.MaxRedirectsAllowed {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return client
}

// CreateClientWithRoundTripper creates a client with custom round tripper.
func (f *SecureHTTPClientFactory) CreateClientWithRoundTripper(rt http.RoundTripper) *http.Client {
	client := f.CreateClient()
	client.Transport = rt
	return client
}

// SecureRoundTripper wraps http.RoundTripper with additional security features.
type SecureRoundTripper struct {
	base    http.RoundTripper
	config  *SecureClientConfig
	retries int
}

// NewSecureRoundTripper creates a new secure round tripper.
func NewSecureRoundTripper(base http.RoundTripper, config *SecureClientConfig) *SecureRoundTripper {
	return &SecureRoundTripper{
		base:    base,
		config:  config,
		retries: config.RetryCount,
	}
}

// RoundTrip implements http.RoundTripper with security enhancements.
func (rt *SecureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add security headers
	if rt.config.UserAgent != "" {
		req.Header.Set("User-Agent", rt.config.UserAgent)
	}

	// Add default headers
	for key, value := range rt.config.DefaultHeaders {
		if req.Header.Get(key) == "" {
			req.Header.Set(key, value)
		}
	}

	// Add security headers
	req.Header.Set("X-Content-Type-Options", "nosniff")
	req.Header.Set("X-Frame-Options", "DENY")
	req.Header.Set("X-XSS-Protection", "1; mode=block")

	// Execute request with retries
	var lastErr error
	for attempt := 0; attempt <= rt.retries; attempt++ {
		resp, err := rt.base.RoundTrip(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == rt.retries {
			break
		}

		// Wait before retry
		time.Sleep(rt.config.RetryDelay * time.Duration(attempt+1))
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", rt.retries, lastErr)
}

// ClientPool manages HTTP client instances for connection reusing.
type ClientPool struct {
	clients map[string]*http.Client
	factory *SecureHTTPClientFactory
}

// NewClientPool creates a new client pool.
func NewClientPool() *ClientPool {
	return &ClientPool{
		clients: make(map[string]*http.Client),
		factory: NewSecureHTTPClientFactory(DefaultSecureClientConfig()),
	}
}

// GetClient returns a cached client or creates a new one.
func (p *ClientPool) GetClient(clientType string) *http.Client {
	if client, exists := p.clients[clientType]; exists {
		return client
	}

	var config *SecureClientConfig
	switch clientType {
	case "github":
		config = GitHubClientConfig()
	case "gitlab":
		config = GitLabClientConfig()
	case "gitea":
		config = GiteaClientConfig()
	default:
		config = DefaultSecureClientConfig()
	}

	factory := NewSecureHTTPClientFactory(config)
	client := factory.CreateClient()
	p.clients[clientType] = client

	return client
}

// CloseIdleConnections closes idle connections for all clients.
func (p *ClientPool) CloseIdleConnections() {
	for _, client := range p.clients {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}

// Global client pool instance.
var globalClientPool = NewClientPool()

// GetGlobalClient returns a client from the global pool.
func GetGlobalClient(clientType string) *http.Client {
	return globalClientPool.GetClient(clientType)
}

// CloseGlobalIdleConnections closes idle connections in global pool.
func CloseGlobalIdleConnections() {
	globalClientPool.CloseIdleConnections()
}
