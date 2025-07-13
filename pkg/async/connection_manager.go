package async

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ConnectionConfig configures the connection manager
type ConnectionConfig struct {
	MaxIdleConns          int           `json:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `json:"max_idle_conns_per_host"`
	MaxConnsPerHost       int           `json:"max_conns_per_host"`
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout"`
	KeepAlive             time.Duration `json:"keep_alive"`
	TLSHandshakeTimeout   time.Duration `json:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout"`
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout"`
	RequestTimeout        time.Duration `json:"request_timeout"`
	RetryConfig           RetryConfig   `json:"retry_config"`
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries           int           `json:"max_retries"`
	BaseDelay            time.Duration `json:"base_delay"`
	MaxDelay             time.Duration `json:"max_delay"`
	BackoffFactor        float64       `json:"backoff_factor"`
	JitterFactor         float64       `json:"jitter_factor"`
	RetryableStatusCodes []int         `json:"retryable_status_codes"`
	RetryableErrors      []string      `json:"retryable_errors"`
}

// ConnectionStats tracks connection performance metrics
type ConnectionStats struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	RetryAttempts      int64         `json:"retry_attempts"`
	AverageLatency     time.Duration `json:"average_latency"`
	ConnectionReuses   int64         `json:"connection_reuses"`
	NewConnections     int64         `json:"new_connections"`
	ActiveConnections  int           `json:"active_connections"`
	IdleConnections    int           `json:"idle_connections"`
	DNSLookupTime      time.Duration `json:"dns_lookup_time"`
	TCPConnectTime     time.Duration `json:"tcp_connect_time"`
	TLSHandshakeTime   time.Duration `json:"tls_handshake_time"`
}

// ConnectionManager manages HTTP client connections with optimization
type ConnectionManager struct {
	config    ConnectionConfig
	client    *http.Client
	transport *http.Transport
	stats     ConnectionStats
	mu        sync.RWMutex
	retryFunc RetryDecisionFunc
}

// RetryDecisionFunc determines if a request should be retried
type RetryDecisionFunc func(req *http.Request, resp *http.Response, err error, attempt int) bool

// DefaultConnectionConfig returns sensible defaults for connection management
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       20,
		IdleConnTimeout:       90 * time.Second,
		KeepAlive:             30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		RequestTimeout:        60 * time.Second,
		RetryConfig: RetryConfig{
			MaxRetries:    3,
			BaseDelay:     100 * time.Millisecond,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
			JitterFactor:  0.1,
			RetryableStatusCodes: []int{
				http.StatusRequestTimeout,
				http.StatusTooManyRequests,
				http.StatusInternalServerError,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout,
			},
			RetryableErrors: []string{
				"connection reset",
				"connection refused",
				"timeout",
				"temporary failure",
				"network unreachable",
			},
		},
	}
}

// NewConnectionManager creates a new optimized connection manager
func NewConnectionManager(config ConnectionConfig) *ConnectionManager {
	cm := &ConnectionManager{
		config:    config,
		retryFunc: defaultRetryDecision,
	}

	cm.setupTransport()
	cm.setupClient()

	return cm
}

// setupTransport configures the HTTP transport with optimizations
func (cm *ConnectionManager) setupTransport() {
	// Custom dialer with keep-alive and timeout settings
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: cm.config.KeepAlive,
		DualStack: true, // Enable IPv4 and IPv6
	}

	cm.transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           cm.instrumentedDial(dialer.DialContext),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          cm.config.MaxIdleConns,
		MaxIdleConnsPerHost:   cm.config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cm.config.MaxConnsPerHost,
		IdleConnTimeout:       cm.config.IdleConnTimeout,
		TLSHandshakeTimeout:   cm.config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cm.config.ResponseHeaderTimeout,
		ExpectContinueTimeout: cm.config.ExpectContinueTimeout,

		// TLS configuration for security and performance
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},

		// Connection tracking
		DisableKeepAlives:      false,
		DisableCompression:     false,
		MaxResponseHeaderBytes: 4096,
	}
}

// setupClient configures the HTTP client
func (cm *ConnectionManager) setupClient() {
	cm.client = &http.Client{
		Transport: cm.transport,
		Timeout:   cm.config.RequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Limit redirects to prevent infinite loops
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

// instrumentedDial wraps the dial function to collect connection metrics
func (cm *ConnectionManager) instrumentedDial(dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()

		conn, err := dial(ctx, network, addr)
		duration := time.Since(start)

		cm.mu.Lock()
		if err == nil {
			cm.stats.NewConnections++
			cm.stats.TCPConnectTime = cm.updateAverage(cm.stats.TCPConnectTime, duration)
		}
		cm.mu.Unlock()

		return conn, err
	}
}

// DoWithRetry performs an HTTP request with intelligent retry logic
func (cm *ConnectionManager) DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	start := time.Now()
	cm.updateStats(true, false, false)

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= cm.config.RetryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff and jitter
			delay := cm.calculateRetryDelay(attempt)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			cm.mu.Lock()
			cm.stats.RetryAttempts++
			cm.mu.Unlock()
		}

		// Clone request for retry safety
		retryReq := cm.cloneRequest(req, ctx)

		resp, err = cm.client.Do(retryReq)

		// Check if we should retry
		if !cm.shouldRetry(req, resp, err, attempt) {
			break
		}

		// Close response body if it exists to prevent connection leaks
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	duration := time.Since(start)
	success := err == nil && resp != nil && resp.StatusCode < 400

	cm.updateStats(false, success, err != nil)
	cm.updateLatency(duration)

	return resp, err
}

// shouldRetry determines if a request should be retried
func (cm *ConnectionManager) shouldRetry(req *http.Request, resp *http.Response, err error, attempt int) bool {
	if attempt >= cm.config.RetryConfig.MaxRetries {
		return false
	}

	// Use custom retry function if available
	if cm.retryFunc != nil {
		return cm.retryFunc(req, resp, err, attempt)
	}

	return defaultRetryDecision(req, resp, err, attempt)
}

// defaultRetryDecision implements the default retry logic
func defaultRetryDecision(req *http.Request, resp *http.Response, err error, attempt int) bool {
	// Don't retry non-idempotent methods unless it's a connection error
	if req.Method != http.MethodGet && req.Method != http.MethodHead &&
		req.Method != http.MethodOptions && req.Method != http.MethodPut &&
		req.Method != http.MethodDelete {
		return err != nil && isConnectionError(err)
	}

	// Retry on connection errors
	if err != nil {
		return isConnectionError(err) || isTemporaryError(err)
	}

	// Retry on specific HTTP status codes
	if resp != nil {
		retryableCodes := []int{
			http.StatusRequestTimeout,
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		}

		for _, code := range retryableCodes {
			if resp.StatusCode == code {
				return true
			}
		}
	}

	return false
}

// calculateRetryDelay calculates the delay for the next retry with jitter
func (cm *ConnectionManager) calculateRetryDelay(attempt int) time.Duration {
	config := cm.config.RetryConfig

	// Exponential backoff
	delay := float64(config.BaseDelay) * math.Pow(config.BackoffFactor, float64(attempt-1))

	// Apply jitter to prevent thundering herd
	jitter := delay * config.JitterFactor * (rand.Float64() - 0.5) * 2
	delay += jitter

	// Cap at maximum delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	return time.Duration(delay)
}

// cloneRequest creates a copy of the request for retry safety
func (cm *ConnectionManager) cloneRequest(req *http.Request, ctx context.Context) *http.Request {
	clone := req.Clone(ctx)

	// If the request has a body, we need to handle it carefully
	if req.Body != nil && req.GetBody != nil {
		body, err := req.GetBody()
		if err == nil {
			clone.Body = body
		}
	}

	return clone
}

// SetRetryDecisionFunc sets a custom retry decision function
func (cm *ConnectionManager) SetRetryDecisionFunc(fn RetryDecisionFunc) {
	cm.retryFunc = fn
}

// GetClient returns the underlying HTTP client
func (cm *ConnectionManager) GetClient() *http.Client {
	return cm.client
}

// updateStats updates connection statistics
func (cm *ConnectionManager) updateStats(starting, success, failed bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if starting {
		cm.stats.TotalRequests++
		cm.stats.ActiveConnections++
	} else {
		cm.stats.ActiveConnections--
		if success {
			cm.stats.SuccessfulRequests++
		}
		if failed {
			cm.stats.FailedRequests++
		}
	}
}

// updateLatency updates average latency using exponential moving average
func (cm *ConnectionManager) updateLatency(duration time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.stats.AverageLatency == 0 {
		cm.stats.AverageLatency = duration
	} else {
		alpha := 0.1
		cm.stats.AverageLatency = time.Duration(
			alpha*float64(duration) + (1-alpha)*float64(cm.stats.AverageLatency),
		)
	}
}

// updateAverage updates an average duration
func (cm *ConnectionManager) updateAverage(current, new time.Duration) time.Duration {
	if current == 0 {
		return new
	}
	alpha := 0.1
	return time.Duration(alpha*float64(new) + (1-alpha)*float64(current))
}

// GetStats returns current connection statistics
func (cm *ConnectionManager) GetStats() ConnectionStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.stats
}

// PrintStats prints detailed connection statistics
func (cm *ConnectionManager) PrintStats() {
	stats := cm.GetStats()

	fmt.Printf("=== Connection Manager Statistics ===\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful: %d\n", stats.SuccessfulRequests)
	fmt.Printf("Failed: %d\n", stats.FailedRequests)
	fmt.Printf("Retry Attempts: %d\n", stats.RetryAttempts)
	fmt.Printf("Active Connections: %d\n", stats.ActiveConnections)
	fmt.Printf("New Connections: %d\n", stats.NewConnections)
	fmt.Printf("Connection Reuses: %d\n", stats.ConnectionReuses)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
	fmt.Printf("DNS Lookup Time: %v\n", stats.DNSLookupTime)
	fmt.Printf("TCP Connect Time: %v\n", stats.TCPConnectTime)
	fmt.Printf("TLS Handshake Time: %v\n", stats.TLSHandshakeTime)

	if stats.TotalRequests > 0 {
		successRate := float64(stats.SuccessfulRequests) / float64(stats.TotalRequests) * 100
		fmt.Printf("Success Rate: %.2f%%\n", successRate)

		if stats.NewConnections > 0 {
			reuseRatio := float64(stats.ConnectionReuses) / float64(stats.NewConnections) * 100
			fmt.Printf("Connection Reuse Ratio: %.2f%%\n", reuseRatio)
		}
	}
}

// Close gracefully shuts down the connection manager
func (cm *ConnectionManager) Close() error {
	if cm.transport != nil {
		cm.transport.CloseIdleConnections()
	}
	return nil
}

// Utility functions for error checking
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"no route to host",
		"network unreachable",
		"broken pipe",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errStr, connErr) {
			return true
		}
	}

	// Check for net.Error types
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}

	return false
}

func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary()
	}

	return false
}
