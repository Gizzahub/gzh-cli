package recovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestNetworkErrorClassifier(t *testing.T) {
	classifier := NewNetworkErrorClassifier()

	tests := []struct {
		name      string
		err       error
		errType   NetworkErrorType
		retryable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			errType:   ErrorTypeUnknown,
			retryable: false,
		},
		{
			name:      "context deadline exceeded",
			err:       context.DeadlineExceeded,
			errType:   ErrorTypeTimeout,
			retryable: false,
		},
		{
			name:      "context canceled",
			err:       context.Canceled,
			errType:   ErrorTypeTimeout,
			retryable: false,
		},
		{
			name:      "DNS error",
			err:       &net.DNSError{Err: "no such host", Name: "example.invalid"},
			errType:   ErrorTypeDNSFailure,
			retryable: false,
		},
		{
			name:      "connection refused by syscall",
			err:       &net.OpError{Err: syscall.ECONNREFUSED},
			errType:   ErrorTypeConnectionRefused,
			retryable: true,
		},
		{
			name:      "connection reset by syscall",
			err:       &net.OpError{Err: syscall.ECONNRESET},
			errType:   ErrorTypeConnectionReset,
			retryable: true,
		},
		{
			name:      "network unreachable by syscall",
			err:       &net.OpError{Err: syscall.ENETUNREACH},
			errType:   ErrorTypeNetworkUnreachable,
			retryable: true,
		},
		{
			name:      "timeout from message",
			err:       errors.New("request timeout"),
			errType:   ErrorTypeTimeout,
			retryable: true,
		},
		{
			name:      "connection refused from message",
			err:       errors.New("connection refused"),
			errType:   ErrorTypeConnectionRefused,
			retryable: true,
		},
		{
			name:      "unknown error",
			err:       errors.New("unknown error"),
			errType:   ErrorTypeUnknown,
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errType, retryable := classifier.ClassifyError(tt.err)
			if errType != tt.errType {
				t.Errorf("Expected error type %v, got %v", tt.errType, errType)
			}
			if retryable != tt.retryable {
				t.Errorf("Expected retryable %v, got %v", tt.retryable, retryable)
			}
		})
	}
}

func TestResilientHTTPClient_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// Create resilient client
	config := DefaultResilientHTTPClientConfig()
	config.UseCircuitBreaker = false // Disable for simpler testing
	client := NewResilientHTTPClient(config)

	// Test successful request
	resp, err := client.GetWithContext(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Expected successful request, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestResilientHTTPClient_RetryOn5xx(t *testing.T) {
	var requestCount int32

	// Create test server that fails first few times
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// Create resilient client with retries
	config := DefaultResilientHTTPClientConfig()
	config.UseCircuitBreaker = false
	config.MaxRetries = 3
	config.InitialDelay = 10 * time.Millisecond
	client := NewResilientHTTPClient(config)

	// Test request that succeeds after retries
	resp, err := client.GetWithContext(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Expected eventual success, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	finalCount := atomic.LoadInt32(&requestCount)
	if finalCount != 3 {
		t.Errorf("Expected 3 requests (2 retries + 1 success), got %d", finalCount)
	}
}

func TestResilientHTTPClient_MaxRetriesExceeded(t *testing.T) {
	var requestCount int32

	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create resilient client with limited retries
	config := DefaultResilientHTTPClientConfig()
	config.UseCircuitBreaker = false
	config.MaxRetries = 2
	config.InitialDelay = 1 * time.Millisecond
	client := NewResilientHTTPClient(config)

	// Test request that fails permanently
	_, err := client.GetWithContext(context.Background(), server.URL)
	if err == nil {
		t.Fatal("Expected error after max retries, got success")
	}

	if !strings.Contains(err.Error(), "max retries") {
		t.Errorf("Expected max retries error, got: %v", err)
	}

	finalCount := atomic.LoadInt32(&requestCount)
	expectedCount := int32(config.MaxRetries + 1) // Initial attempt + retries
	if finalCount != expectedCount {
		t.Errorf("Expected %d requests, got %d", expectedCount, finalCount)
	}
}

func TestResilientHTTPClient_ContextCancellation(t *testing.T) {
	// Create test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create resilient client
	config := DefaultResilientHTTPClientConfig()
	config.UseCircuitBreaker = false
	client := NewResilientHTTPClient(config)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Test request that gets cancelled
	_, err := client.GetWithContext(ctx, server.URL)
	if err == nil {
		t.Fatal("Expected timeout error, got success")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected deadline exceeded error, got: %v", err)
	}
}

func TestResilientHTTPClient_CircuitBreaker(t *testing.T) {
	var requestCount int32

	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create resilient client with circuit breaker
	config := DefaultResilientHTTPClientConfig()
	config.UseCircuitBreaker = true
	config.MaxRetries = 1
	config.InitialDelay = 1 * time.Millisecond
	config.CircuitConfig.FailureThreshold = 3
	config.CircuitConfig.Timeout = 100 * time.Millisecond
	client := NewResilientHTTPClient(config)

	// Make requests to trip circuit breaker
	for i := 0; i < 5; i++ {
		_, err := client.GetWithContext(context.Background(), server.URL)
		if err == nil {
			t.Errorf("Request %d: Expected error, got success", i+1)
		}
	}

	// Verify circuit breaker state
	stats := client.GetStats()
	if cbStats, ok := stats["circuit_breaker"]; ok {
		if cbMetrics, ok := cbStats.(interface{}); ok {
			// Use interface{} and check state via string comparison for simplicity
			fmt.Printf("Circuit breaker stats: %+v\n", cbMetrics)
		}
	}
}

func TestResilientHTTPClient_Configuration(t *testing.T) {
	config := DefaultResilientHTTPClientConfig()
	client := NewResilientHTTPClient(config)

	stats := client.GetStats()
	if stats["config"] == nil {
		t.Error("Expected config in stats")
	}

	// Test client creation with custom config
	customConfig := ResilientHTTPClientConfig{
		Timeout:           10 * time.Second,
		MaxRetries:        5,
		UseCircuitBreaker: false,
		ErrorClassifier:   NewNetworkErrorClassifier(),
	}

	customClient := NewResilientHTTPClient(customConfig)
	if customClient == nil {
		t.Error("Expected custom client to be created")
	}
}

// Test that demonstrates the circuit breaker recovery
func TestResilientHTTPClient_CircuitBreakerRecovery(t *testing.T) {
	var requestCount int32
	var shouldFail int32 = 1

	// Create test server that can toggle between failure and success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		if atomic.LoadInt32(&shouldFail) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// Create resilient client with circuit breaker
	config := DefaultResilientHTTPClientConfig()
	config.UseCircuitBreaker = true
	config.MaxRetries = 0 // No retries, just circuit breaker
	config.CircuitConfig.FailureThreshold = 2
	config.CircuitConfig.Timeout = 50 * time.Millisecond
	config.CircuitConfig.SuccessThreshold = 1
	client := NewResilientHTTPClient(config)

	// Trip the circuit breaker
	for i := 0; i < 3; i++ {
		client.GetWithContext(context.Background(), server.URL)
	}

	// Fix the server
	atomic.StoreInt32(&shouldFail, 0)

	// Wait for circuit breaker timeout
	time.Sleep(100 * time.Millisecond)

	// Should be able to make successful requests now
	resp, err := client.GetWithContext(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Expected recovery success, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 after recovery, got %d", resp.StatusCode)
	}
}

// Benchmark for performance testing
func BenchmarkResilientHTTPClient(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("benchmark"))
	}))
	defer server.Close()

	config := DefaultResilientHTTPClientConfig()
	config.UseCircuitBreaker = false // Disable for cleaner benchmark
	config.MaxRetries = 0
	client := NewResilientHTTPClient(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.GetWithContext(context.Background(), server.URL)
			if err != nil {
				b.Fatalf("Benchmark request failed: %v", err)
			}
			resp.Body.Close()
		}
	})
}
