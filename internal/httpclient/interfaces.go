// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package httpclient

import (
	"context"
	"io"
	"time"
)

// HTTPRequest represents an HTTP request.
type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    io.Reader         `json:"-"`
	Timeout time.Duration     `json:"timeout"`
}

// HTTPResponse represents an HTTP response.
type HTTPResponse struct {
	StatusCode int               `json:"statusCode"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Size       int64             `json:"size"`
	Duration   time.Duration     `json:"duration"`
}

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	// Basic HTTP methods
	Get(ctx context.Context, url string) (*HTTPResponse, error)
	Post(ctx context.Context, url string, body io.Reader) (*HTTPResponse, error)
	Put(ctx context.Context, url string, body io.Reader) (*HTTPResponse, error)
	Patch(ctx context.Context, url string, body io.Reader) (*HTTPResponse, error)
	Delete(ctx context.Context, url string) (*HTTPResponse, error)
	Head(ctx context.Context, url string) (*HTTPResponse, error)
	Options(ctx context.Context, url string) (*HTTPResponse, error)

	// Advanced request method
	Do(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error)

	// Configuration
	SetTimeout(timeout time.Duration)
	SetUserAgent(userAgent string)
	SetBaseURL(baseURL string)
	AddDefaultHeader(key, value string)
	RemoveDefaultHeader(key string)

	// Authentication
	SetBearerToken(token string)
	SetBasicAuth(username, password string)
	SetAPIKey(key, value string)

	// Request/Response middleware
	AddRequestMiddleware(middleware RequestMiddleware)
	AddResponseMiddleware(middleware ResponseMiddleware)
}

// RequestMiddleware defines the interface for request middleware.
type RequestMiddleware interface {
	ProcessRequest(ctx context.Context, req *HTTPRequest) (*HTTPRequest, error)
}

// ResponseMiddleware defines the interface for response middleware.
type ResponseMiddleware interface {
	ProcessResponse(ctx context.Context, req *HTTPRequest, resp *HTTPResponse) (*HTTPResponse, error)
}

// RetryPolicy defines the interface for retry logic.
type RetryPolicy interface {
	// Check if request should be retried
	ShouldRetry(ctx context.Context, req *HTTPRequest, resp *HTTPResponse, err error, attempt int) bool

	// Get delay before next retry
	GetRetryDelay(ctx context.Context, attempt int) time.Duration

	// Get maximum retry attempts
	MaxRetries() int
}

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	// Check if request is allowed
	Allow(ctx context.Context) bool

	// Wait until request is allowed
	Wait(ctx context.Context) error

	// Get rate limit information
	GetLimitInfo() *RateLimitInfo

	// Reset rate limiter
	Reset()
}

// RateLimitInfo represents rate limit information.
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	Reset     time.Time     `json:"reset"`
	Window    time.Duration `json:"window"`
}

// CachePolicy defines the interface for HTTP caching.
type CachePolicy interface {
	// Check if response can be cached
	ShouldCache(ctx context.Context, req *HTTPRequest, resp *HTTPResponse) bool

	// Get cached response
	GetCached(ctx context.Context, req *HTTPRequest) (*HTTPResponse, bool)

	// Store response in cache
	Store(ctx context.Context, req *HTTPRequest, resp *HTTPResponse) error

	// Invalidate cache entries
	Invalidate(ctx context.Context, pattern string) error

	// Get cache statistics
	GetStats() *CacheStats
}

// CacheStats represents cache statistics.
type CacheStats struct {
	Hits      int64 `json:"hits"`
	Misses    int64 `json:"misses"`
	Stores    int64 `json:"stores"`
	Evictions int64 `json:"evictions"`
	Size      int64 `json:"size"`
	MaxSize   int64 `json:"maxSize"`
}

// RequestLogger defines the interface for request logging.
type RequestLogger interface {
	// Log request
	LogRequest(ctx context.Context, req *HTTPRequest) error

	// Log response
	LogResponse(ctx context.Context, req *HTTPRequest, resp *HTTPResponse, err error) error

	// Set log level
	SetLogLevel(level LogLevel)

	// Get request logs
	GetLogs(ctx context.Context, filters LogFilters) ([]LogEntry, error)
}

// LogLevel represents logging level.
type LogLevel int

const (
	// LogLevelNone represents no logging.
	LogLevelNone LogLevel = iota
	// LogLevelError represents error logging.
	LogLevelError
	// LogLevelWarn represents warning logging.
	LogLevelWarn
	// LogLevelInfo represents info logging.
	LogLevelInfo
	// LogLevelDebug represents debug logging.
	LogLevelDebug
)

// LogFilters represents filters for log queries.
type LogFilters struct {
	Method      string        `json:"method,omitempty"`
	URL         string        `json:"url,omitempty"`
	StatusCode  int           `json:"statusCode,omitempty"`
	StartTime   time.Time     `json:"startTime,omitempty"`
	EndTime     time.Time     `json:"endTime,omitempty"`
	MinDuration time.Duration `json:"minDuration,omitempty"`
	MaxDuration time.Duration `json:"maxDuration,omitempty"`
}

// LogEntry represents a logged HTTP request/response.
type LogEntry struct {
	ID           string        `json:"id"`
	Timestamp    time.Time     `json:"timestamp"`
	Method       string        `json:"method"`
	URL          string        `json:"url"`
	StatusCode   int           `json:"statusCode"`
	Duration     time.Duration `json:"duration"`
	RequestSize  int64         `json:"requestSize"`
	ResponseSize int64         `json:"responseSize"`
	Error        string        `json:"error,omitempty"`
	UserAgent    string        `json:"userAgent"`
	RemoteAddr   string        `json:"remoteAddr"`
}

// MockClient defines the interface for HTTP client mocking.
type MockClient interface {
	HTTPClient

	// Mock management
	AddMock(mock *Mock) error
	RemoveMock(id string) error
	ClearMocks() error
	ListMocks() []*Mock

	// Recording
	StartRecording() error
	StopRecording() error
	GetRecordings() []*Recording
	SaveRecordings(path string) error
	LoadRecordings(path string) error
}

// Mock represents a mocked HTTP request/response.
type Mock struct {
	ID         string            `json:"id"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	URLPattern string            `json:"urlPattern,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	StatusCode int               `json:"statusCode"`
	Response   []byte            `json:"response"`
	Delay      time.Duration     `json:"delay,omitempty"`
	Times      int               `json:"times,omitempty"` // number of times to match, 0 = unlimited
	Condition  MockCondition     `json:"condition,omitempty"`
}

// MockCondition defines conditions for mock matching.
type MockCondition interface {
	Matches(ctx context.Context, req *HTTPRequest) bool
}

// Recording represents a recorded HTTP request/response.
type Recording struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Request   *HTTPRequest      `json:"request"`
	Response  *HTTPResponse     `json:"response"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
}

// MetricsCollector defines the interface for collecting HTTP metrics.
type MetricsCollector interface {
	// Record request metrics
	RecordRequest(ctx context.Context, req *HTTPRequest, resp *HTTPResponse, duration time.Duration, err error)

	// Get metrics
	GetMetrics() *HTTPMetrics

	// Reset metrics
	Reset()

	// Export metrics in different formats
	ExportPrometheus() ([]byte, error)
	ExportJSON() ([]byte, error)
}

// HTTPMetrics represents collected HTTP metrics.
type HTTPMetrics struct {
	TotalRequests      int64            `json:"totalRequests"`
	SuccessfulRequests int64            `json:"successfulRequests"`
	FailedRequests     int64            `json:"failedRequests"`
	TotalDuration      time.Duration    `json:"totalDuration"`
	AverageDuration    time.Duration    `json:"averageDuration"`
	MinDuration        time.Duration    `json:"minDuration"`
	MaxDuration        time.Duration    `json:"maxDuration"`
	StatusCodeCounts   map[int]int64    `json:"statusCodeCounts"`
	MethodCounts       map[string]int64 `json:"methodCounts"`
	ErrorCounts        map[string]int64 `json:"errorCounts"`
	ResponseSizes      *SizeStats       `json:"responseSizes"`
	RequestSizes       *SizeStats       `json:"requestSizes"`
	TopEndpoints       []EndpointStat   `json:"topEndpoints"`
}

// SizeStats represents size statistics.
type SizeStats struct {
	Total   int64 `json:"total"`
	Average int64 `json:"average"`
	Min     int64 `json:"min"`
	Max     int64 `json:"max"`
	P50     int64 `json:"p50"`
	P95     int64 `json:"p95"`
	P99     int64 `json:"p99"`
}

// EndpointStat represents statistics for an endpoint.
type EndpointStat struct {
	URL             string        `json:"url"`
	Count           int64         `json:"count"`
	AverageDuration time.Duration `json:"averageDuration"`
	ErrorRate       float64       `json:"errorRate"`
}

// HTTPService provides a unified interface for all HTTP operations.
type HTTPService interface {
	HTTPClient
	RetryPolicy
	RateLimiter
	CachePolicy
	RequestLogger
	MetricsCollector
}
