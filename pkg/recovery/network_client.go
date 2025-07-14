package recovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
)

// NetworkError types for classification and recovery strategies
type NetworkErrorType int

const (
	ErrorTypeUnknown NetworkErrorType = iota
	ErrorTypeTimeout
	ErrorTypeConnectionRefused
	ErrorTypeDNSFailure
	ErrorTypeNetworkUnreachable
	ErrorTypeConnectionReset
	ErrorTypeTemporary
	ErrorTypePermanent
)

// NetworkErrorClassifier determines the type and recovery strategy for network errors
type NetworkErrorClassifier struct {
	retryableErrors map[NetworkErrorType]bool
}

// NewNetworkErrorClassifier creates a new error classifier with default settings
func NewNetworkErrorClassifier() *NetworkErrorClassifier {
	return &NetworkErrorClassifier{
		retryableErrors: map[NetworkErrorType]bool{
			ErrorTypeTimeout:            true,
			ErrorTypeConnectionRefused:  true,
			ErrorTypeNetworkUnreachable: true,
			ErrorTypeConnectionReset:    true,
			ErrorTypeTemporary:          true,
			ErrorTypeDNSFailure:         false, // DNS failures are usually permanent
			ErrorTypePermanent:          false,
		},
	}
}

// ClassifyError analyzes a network error and returns its type and whether it's retryable
func (nec *NetworkErrorClassifier) ClassifyError(err error) (NetworkErrorType, bool) {
	if err == nil {
		return ErrorTypeUnknown, false
	}

	// Check for context timeout/cancellation
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrorTypeTimeout, false // Don't retry context cancellations
	}

	// Check for URL errors (wraps network errors)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return nec.ClassifyError(urlErr.Err)
	}

	// Check for net.Error interface
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ErrorTypeTimeout, nec.retryableErrors[ErrorTypeTimeout]
		}
		if netErr.Temporary() {
			return ErrorTypeTemporary, nec.retryableErrors[ErrorTypeTemporary]
		}
	}

	// Check for specific network error types
	if isConnectionRefused(err) {
		return ErrorTypeConnectionRefused, nec.retryableErrors[ErrorTypeConnectionRefused]
	}

	if isConnectionReset(err) {
		return ErrorTypeConnectionReset, nec.retryableErrors[ErrorTypeConnectionReset]
	}

	if isDNSError(err) {
		return ErrorTypeDNSFailure, nec.retryableErrors[ErrorTypeDNSFailure]
	}

	if isNetworkUnreachable(err) {
		return ErrorTypeNetworkUnreachable, nec.retryableErrors[ErrorTypeNetworkUnreachable]
	}

	// Check error message for common patterns
	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "timeout"):
		return ErrorTypeTimeout, nec.retryableErrors[ErrorTypeTimeout]
	case strings.Contains(errMsg, "connection refused"):
		return ErrorTypeConnectionRefused, nec.retryableErrors[ErrorTypeConnectionRefused]
	case strings.Contains(errMsg, "connection reset"):
		return ErrorTypeConnectionReset, nec.retryableErrors[ErrorTypeConnectionReset]
	case strings.Contains(errMsg, "no route to host"):
		return ErrorTypeNetworkUnreachable, nec.retryableErrors[ErrorTypeNetworkUnreachable]
	case strings.Contains(errMsg, "temporary failure"):
		return ErrorTypeTemporary, nec.retryableErrors[ErrorTypeTemporary]
	}

	return ErrorTypeUnknown, false
}

// Helper functions for error type detection
func isConnectionRefused(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var syscallErr *syscall.Errno
		if errors.As(opErr.Err, &syscallErr) {
			return *syscallErr == syscall.ECONNREFUSED
		}
	}
	return false
}

func isConnectionReset(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var syscallErr *syscall.Errno
		if errors.As(opErr.Err, &syscallErr) {
			return *syscallErr == syscall.ECONNRESET
		}
	}
	return false
}

func isDNSError(err error) bool {
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}

func isNetworkUnreachable(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var syscallErr *syscall.Errno
		if errors.As(opErr.Err, &syscallErr) {
			return *syscallErr == syscall.ENETUNREACH || *syscallErr == syscall.EHOSTUNREACH
		}
	}
	return false
}

// ResilientHTTPClientConfig configures the resilient HTTP client
type ResilientHTTPClientConfig struct {
	// Base HTTP client configuration
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration

	// Retry configuration
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64

	// Circuit breaker configuration
	UseCircuitBreaker bool
	CircuitConfig     CircuitBreakerConfig

	// Custom error classifier
	ErrorClassifier *NetworkErrorClassifier
}

// DefaultResilientHTTPClientConfig returns a configuration with sensible defaults
func DefaultResilientHTTPClientConfig() ResilientHTTPClientConfig {
	return ResilientHTTPClientConfig{
		Timeout:           30 * time.Second,
		MaxIdleConns:      100,
		IdleConnTimeout:   90 * time.Second,
		MaxRetries:        3,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffFactor:     2.0,
		UseCircuitBreaker: true,
		CircuitConfig:     DefaultCircuitBreakerConfig("http-client"),
		ErrorClassifier:   NewNetworkErrorClassifier(),
	}
}

// ResilientHTTPClient provides HTTP client with automatic retry and circuit breaker
type ResilientHTTPClient struct {
	client          *http.Client
	config          ResilientHTTPClientConfig
	circuitBreaker  *CircuitBreaker
	errorClassifier *NetworkErrorClassifier
}

// NewResilientHTTPClient creates a new resilient HTTP client
func NewResilientHTTPClient(config ResilientHTTPClientConfig) *ResilientHTTPClient {
	// Configure base HTTP client with timeouts
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 10,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	resilientClient := &ResilientHTTPClient{
		client:          client,
		config:          config,
		errorClassifier: config.ErrorClassifier,
	}

	// Initialize circuit breaker if enabled
	if config.UseCircuitBreaker {
		resilientClient.circuitBreaker = NewCircuitBreaker(config.CircuitConfig)
	}

	return resilientClient
}

// Do executes an HTTP request with retry and circuit breaker protection
func (rc *ResilientHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if rc.circuitBreaker != nil {
		// Use circuit breaker
		var resp *http.Response
		err := rc.circuitBreaker.Execute(req.Context(), func() error {
			var execErr error
			resp, execErr = rc.doWithRetry(req)
			return execErr
		})
		return resp, err
	}

	// Direct retry without circuit breaker
	return rc.doWithRetry(req)
}

// doWithRetry performs HTTP request with retry logic
func (rc *ResilientHTTPClient) doWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	delay := rc.config.InitialDelay

	for attempt := 0; attempt <= rc.config.MaxRetries; attempt++ {
		// Clone request for retry attempts
		reqClone := req.Clone(req.Context())

		resp, err := rc.client.Do(reqClone)
		if err == nil {
			// Check for HTTP-level errors that should be retried
			if rc.shouldRetryStatus(resp.StatusCode) {
				resp.Body.Close()
				lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			} else {
				return resp, nil
			}
		} else {
			lastErr = err
		}

		// Check if error is retryable
		errType, retryable := rc.errorClassifier.ClassifyError(lastErr)
		if !retryable || attempt == rc.config.MaxRetries {
			break
		}

		// Check for context cancellation
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * rc.config.BackoffFactor)
		if delay > rc.config.MaxDelay {
			delay = rc.config.MaxDelay
		}

		// Log retry attempt (you might want to use a proper logger here)
		fmt.Printf("Retrying HTTP request (attempt %d/%d) after %v due to %v error: %v\n",
			attempt+1, rc.config.MaxRetries, delay, errType, lastErr)
	}

	return nil, fmt.Errorf("max retries (%d) exceeded, last error: %w", rc.config.MaxRetries, lastErr)
}

// shouldRetryStatus determines if an HTTP status code should trigger a retry
func (rc *ResilientHTTPClient) shouldRetryStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// Get performs a GET request with resilience
func (rc *ResilientHTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return rc.Do(req)
}

// GetWithContext performs a GET request with context and resilience
func (rc *ResilientHTTPClient) GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return rc.Do(req)
}

// Post performs a POST request with resilience
func (rc *ResilientHTTPClient) Post(url, contentType string, body interface{}) (*http.Response, error) {
	// Implementation would depend on body type (io.Reader, []byte, etc.)
	// For now, we'll keep it simple
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return rc.Do(req)
}

// GetStats returns statistics about the client's performance
func (rc *ResilientHTTPClient) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"config": rc.config,
	}

	if rc.circuitBreaker != nil {
		stats["circuit_breaker"] = rc.circuitBreaker.GetMetrics()
	}

	return stats
}

// Close closes the underlying HTTP client connections
func (rc *ResilientHTTPClient) Close() {
	if transport, ok := rc.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}
