package testlib

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// NetworkErrorSimulator simulates various network error conditions
// for testing synclone behavior under network failure scenarios.
type NetworkErrorSimulator struct {
	server     *httptest.Server
	errorRules []ErrorRule
	mu         sync.RWMutex
}

// ErrorRule defines when and how to simulate network errors
type ErrorRule struct {
	Pattern      string        // URL pattern to match
	ErrorType    ErrorType     // Type of error to simulate
	Probability  float64       // Probability of error (0.0 to 1.0)
	Delay        time.Duration // Delay before error/response
	ResponseCode int           // HTTP response code for HTTP errors
	Message      string        // Error message
	Count        int           // Number of times to apply this rule (0 = infinite)
	applied      int           // Internal counter
}

// ErrorType represents different types of network errors
type ErrorType int

const (
	ErrorTypeTimeout            ErrorType = iota // Connection timeout
	ErrorTypeRefused                             // Connection refused
	ErrorType404                                 // HTTP 404 Not Found
	ErrorType500                                 // HTTP 500 Internal Server Error
	ErrorType503                                 // HTTP 503 Service Unavailable
	ErrorTypeSlowResponse                        // Slow response (high latency)
	ErrorTypeIncompleteResponse                  // Connection drops mid-response
	ErrorTypeDNSFailure                          // DNS resolution failure
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeRefused:
		return "connection_refused"
	case ErrorType404:
		return "http_404"
	case ErrorType500:
		return "http_500"
	case ErrorType503:
		return "http_503"
	case ErrorTypeSlowResponse:
		return "slow_response"
	case ErrorTypeIncompleteResponse:
		return "incomplete_response"
	case ErrorTypeDNSFailure:
		return "dns_failure"
	default:
		return "unknown"
	}
}

// NewNetworkErrorSimulator creates a new NetworkErrorSimulator instance
func NewNetworkErrorSimulator() *NetworkErrorSimulator {
	return &NetworkErrorSimulator{
		errorRules: make([]ErrorRule, 0),
	}
}

// StartServer starts the mock server for simulating network conditions
func (nes *NetworkErrorSimulator) StartServer() string {
	nes.mu.Lock()
	defer nes.mu.Unlock()

	if nes.server != nil {
		return nes.server.URL
	}

	nes.server = httptest.NewServer(http.HandlerFunc(nes.handleRequest))
	return nes.server.URL
}

// StopServer stops the mock server
func (nes *NetworkErrorSimulator) StopServer() {
	nes.mu.Lock()
	defer nes.mu.Unlock()

	if nes.server != nil {
		nes.server.Close()
		nes.server = nil
	}
}

// AddErrorRule adds a rule for simulating network errors
func (nes *NetworkErrorSimulator) AddErrorRule(rule ErrorRule) {
	nes.mu.Lock()
	defer nes.mu.Unlock()
	nes.errorRules = append(nes.errorRules, rule)
}

// ClearErrorRules removes all error rules
func (nes *NetworkErrorSimulator) ClearErrorRules() {
	nes.mu.Lock()
	defer nes.mu.Unlock()
	nes.errorRules = make([]ErrorRule, 0)
}

// SimulateGitCloneTimeout simulates timeout during git clone operations
func (nes *NetworkErrorSimulator) SimulateGitCloneTimeout(ctx context.Context) *NetworkErrorSimulator {
	rule := ErrorRule{
		Pattern:     "/.*\\.git",
		ErrorType:   ErrorTypeTimeout,
		Probability: 1.0,
		Delay:       5 * time.Second,
		Message:     "Connection timeout during git clone",
	}
	nes.AddErrorRule(rule)
	return nes
}

// SimulateIntermittentConnection simulates intermittent connection issues
func (nes *NetworkErrorSimulator) SimulateIntermittentConnection(ctx context.Context, failureRate float64) *NetworkErrorSimulator {
	rules := []ErrorRule{
		{
			Pattern:     ".*",
			ErrorType:   ErrorTypeRefused,
			Probability: failureRate * 0.4,
			Message:     "Connection refused intermittently",
		},
		{
			Pattern:     ".*",
			ErrorType:   ErrorTypeTimeout,
			Probability: failureRate * 0.3,
			Delay:       3 * time.Second,
			Message:     "Intermittent timeout",
		},
		{
			Pattern:     ".*",
			ErrorType:   ErrorTypeSlowResponse,
			Probability: failureRate * 0.3,
			Delay:       10 * time.Second,
			Message:     "Slow response",
		},
	}

	for _, rule := range rules {
		nes.AddErrorRule(rule)
	}
	return nes
}

// SimulateHTTPErrors simulates various HTTP error responses
func (nes *NetworkErrorSimulator) SimulateHTTPErrors(ctx context.Context) *NetworkErrorSimulator {
	rules := []ErrorRule{
		{
			Pattern:      "/repo/.*",
			ErrorType:    ErrorType404,
			Probability:  0.2,
			ResponseCode: 404,
			Message:      "Repository not found",
		},
		{
			Pattern:      "/api/.*",
			ErrorType:    ErrorType500,
			Probability:  0.1,
			ResponseCode: 500,
			Message:      "Internal server error",
		},
		{
			Pattern:      "/.*",
			ErrorType:    ErrorType503,
			Probability:  0.05,
			ResponseCode: 503,
			Message:      "Service unavailable",
		},
	}

	for _, rule := range rules {
		nes.AddErrorRule(rule)
	}
	return nes
}

// CreateRecoveryScenario creates a scenario where network recovers after initial failures
func (nes *NetworkErrorSimulator) CreateRecoveryScenario(ctx context.Context, initialFailures int) *NetworkErrorSimulator {
	// Fail for the first N requests, then work normally
	rule := ErrorRule{
		Pattern:     ".*",
		ErrorType:   ErrorTypeRefused,
		Probability: 1.0,
		Count:       initialFailures,
		Message:     "Network unavailable initially",
	}
	nes.AddErrorRule(rule)
	return nes
}

// GetRequestStats returns statistics about handled requests
func (nes *NetworkErrorSimulator) GetRequestStats() RequestStats {
	nes.mu.RLock()
	defer nes.mu.RUnlock()

	stats := RequestStats{}
	for _, rule := range nes.errorRules {
		stats.TotalRules++
		stats.TotalApplications += rule.applied
	}

	return stats
}

// RequestStats contains statistics about simulated requests
type RequestStats struct {
	TotalRules        int
	TotalApplications int
}

// handleRequest handles incoming HTTP requests and applies error rules
func (nes *NetworkErrorSimulator) handleRequest(w http.ResponseWriter, r *http.Request) {
	nes.mu.Lock()
	defer nes.mu.Unlock()

	// Check each error rule
	for i := range nes.errorRules {
		rule := &nes.errorRules[i]

		// Check if rule applies to this request
		if !nes.matchesPattern(r.URL.Path, rule.Pattern) {
			continue
		}

		// Check if rule has been used up
		if rule.Count > 0 && rule.applied >= rule.Count {
			continue
		}

		// Check probability
		if rule.Probability < 1.0 && !nes.shouldApplyRule(rule.Probability) {
			continue
		}

		// Apply the rule
		rule.applied++
		nes.applyErrorRule(w, r, rule)
		return
	}

	// No error rule matched, return success response
	nes.handleSuccessResponse(w, r)
}

// matchesPattern checks if a URL path matches a pattern (simple implementation)
func (nes *NetworkErrorSimulator) matchesPattern(path, pattern string) bool {
	// Simple pattern matching - in a real implementation, would use regex
	if pattern == ".*" {
		return true
	}
	if pattern == path {
		return true
	}
	// Simple wildcard matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(path) >= len(prefix) && path[:len(prefix)] == prefix
	}
	return false
}

// shouldApplyRule determines if a rule should be applied based on probability
func (nes *NetworkErrorSimulator) shouldApplyRule(probability float64) bool {
	// Simple probability check - in real implementation would use proper random
	return time.Now().UnixNano()%100 < int64(probability*100)
}

// applyErrorRule applies the specified error rule to the request
func (nes *NetworkErrorSimulator) applyErrorRule(w http.ResponseWriter, r *http.Request, rule *ErrorRule) {
	// Apply delay if specified
	if rule.Delay > 0 {
		time.Sleep(rule.Delay)
	}

	switch rule.ErrorType {
	case ErrorTypeTimeout:
		// For timeout, we just hang and let the client timeout
		time.Sleep(30 * time.Second)

	case ErrorTypeRefused:
		// Close connection immediately
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}

	case ErrorType404, ErrorType500, ErrorType503:
		w.WriteHeader(rule.ResponseCode)
		fmt.Fprintf(w, `{"error": "%s"}`, rule.Message)

	case ErrorTypeSlowResponse:
		// Simulate slow response by adding delay
		time.Sleep(rule.Delay)
		nes.handleSuccessResponse(w, r)

	case ErrorTypeIncompleteResponse:
		// Send partial response then close connection
		w.WriteHeader(200)
		fmt.Fprint(w, `{"data": "partial`)
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}

	case ErrorTypeDNSFailure:
		// DNS failure is handled at a different level
		w.WriteHeader(503)
		fmt.Fprintf(w, `{"error": "DNS resolution failed"}`)
	}
}

// handleSuccessResponse handles a successful response
func (nes *NetworkErrorSimulator) handleSuccessResponse(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	// Return different content based on the path
	switch {
	case nes.matchesPattern(r.URL.Path, "/.*\\.git"):
		// Git repository response
		fmt.Fprint(w, `{"type": "git", "status": "ok"}`)
	case nes.matchesPattern(r.URL.Path, "/api/.*"):
		// API response
		fmt.Fprint(w, `{"api_version": "1.0", "status": "healthy"}`)
	default:
		// Generic success response
		fmt.Fprint(w, `{"status": "ok", "message": "Request successful"}`)
	}
}

// GitOperationSimulator simulates network issues during Git operations
type GitOperationSimulator struct {
	nes *NetworkErrorSimulator
}

// NewGitOperationSimulator creates a Git operation simulator
func NewGitOperationSimulator() *GitOperationSimulator {
	return &GitOperationSimulator{
		nes: NewNetworkErrorSimulator(),
	}
}

// SimulateCloneFailure simulates failures during git clone
func (gos *GitOperationSimulator) SimulateCloneFailure(ctx context.Context, failureType string) error {
	switch failureType {
	case "timeout":
		gos.nes.SimulateGitCloneTimeout(ctx)
	case "intermittent":
		gos.nes.SimulateIntermittentConnection(ctx, 0.7)
	case "http_errors":
		gos.nes.SimulateHTTPErrors(ctx)
	case "recovery":
		gos.nes.CreateRecoveryScenario(ctx, 3)
	default:
		return fmt.Errorf("unknown failure type: %s", failureType)
	}

	return nil
}

// GetServerURL returns the URL of the test server
func (gos *GitOperationSimulator) GetServerURL() string {
	return gos.nes.StartServer()
}

// Cleanup stops the test server
func (gos *GitOperationSimulator) Cleanup() {
	gos.nes.StopServer()
}
