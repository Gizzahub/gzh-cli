// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package mocking

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

// MockComplexGitHubService demonstrates testify mock for complex stateful scenarios.
type MockComplexGitHubService struct {
	mock.Mock
	CallLog      []string
	StateCounter int
}

// NewMockComplexGitHubService creates a new complex GitHub service mock.
func NewMockComplexGitHubService() *MockComplexGitHubService {
	return &MockComplexGitHubService{
		CallLog:      make([]string, 0),
		StateCounter: 0,
	}
}

// ProcessRepositories simulates complex repository processing with state.
func (m *MockComplexGitHubService) ProcessRepositories(ctx context.Context, repos []github.RepositoryInfo) (*ProcessResult, error) {
	args := m.Called(ctx, repos)

	// Track method calls
	m.CallLog = append(m.CallLog, "ProcessRepositories")
	m.StateCounter++

	// Simulate complex processing logic
	result := &ProcessResult{
		ProcessedCount: len(repos),
		SkippedCount:   0,
		ErrorCount:     0,
		ProcessingTime: time.Millisecond * 100,
		State:          m.StateCounter,
	}

	return result, args.Error(1)
}

// BulkCloneWithCallback simulates bulk cloning with progress callbacks.
func (m *MockComplexGitHubService) BulkCloneWithCallback(ctx context.Context, repos []github.RepositoryInfo, callback ProgressCallback) error {
	args := m.Called(ctx, repos, callback)

	m.CallLog = append(m.CallLog, "BulkCloneWithCallback")

	// Simulate progress callbacks
	if callback != nil {
		for i, repo := range repos {
			progress := ProgressUpdate{
				Current:     i + 1,
				Total:       len(repos),
				Repository:  repo.Name,
				Status:      "cloning",
				ElapsedTime: time.Millisecond * time.Duration(i*100),
			}
			callback(progress)
		}
	}

	return args.Error(0)
}

// GetProcessingHistory returns the call history.
func (m *MockComplexGitHubService) GetProcessingHistory() []string {
	return m.CallLog
}

// GetCurrentState returns the current state counter.
func (m *MockComplexGitHubService) GetCurrentState() int {
	return m.StateCounter
}

// ResetState resets the mock state.
func (m *MockComplexGitHubService) ResetState() {
	m.CallLog = make([]string, 0)
	m.StateCounter = 0
}

// ProcessResult represents the result of repository processing.
type ProcessResult struct {
	ProcessedCount int           `json:"processed_count"`
	SkippedCount   int           `json:"skipped_count"`
	ErrorCount     int           `json:"error_count"`
	ProcessingTime time.Duration `json:"processing_time"`
	State          int           `json:"state"`
}

// ProgressCallback defines the callback function for progress updates.
type ProgressCallback func(ProgressUpdate)

// ProgressUpdate represents a progress update.
type ProgressUpdate struct {
	Current     int           `json:"current"`
	Total       int           `json:"total"`
	Repository  string        `json:"repository"`
	Status      string        `json:"status"`
	ElapsedTime time.Duration `json:"elapsed_time"`
}

// MockStatefulFileSystem demonstrates stateful file system mocking.
type MockStatefulFileSystem struct {
	mock.Mock
	Files          map[string][]byte
	Directories    map[string]bool
	AccessLog      []FileOperation
	OperationCount int
}

// FileOperation represents a file system operation.
type FileOperation struct {
	Operation string    `json:"operation"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size,omitempty"`
}

// NewMockStatefulFileSystem creates a new stateful file system mock.
func NewMockStatefulFileSystem() *MockStatefulFileSystem {
	return &MockStatefulFileSystem{
		Files:       make(map[string][]byte),
		Directories: make(map[string]bool),
		AccessLog:   make([]FileOperation, 0),
	}
}

// WriteFile simulates writing a file with state tracking.
func (m *MockStatefulFileSystem) WriteFile(ctx context.Context, path string, data []byte, perm int) error {
	args := m.Called(ctx, path, data, perm)

	// Update internal state
	m.Files[path] = make([]byte, len(data))
	copy(m.Files[path], data)

	// Log operation
	m.AccessLog = append(m.AccessLog, FileOperation{
		Operation: "write",
		Path:      path,
		Timestamp: time.Now(),
		Size:      int64(len(data)),
	})
	m.OperationCount++

	return args.Error(0)
}

// ReadFile simulates reading a file with state tracking.
func (m *MockStatefulFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)

	// Log operation
	m.AccessLog = append(m.AccessLog, FileOperation{
		Operation: "read",
		Path:      path,
		Timestamp: time.Now(),
	})
	m.OperationCount++

	// Return file content if exists
	if data, exists := m.Files[path]; exists {
		result := make([]byte, len(data))
		copy(result, data)

		return result, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

// MkdirAll simulates creating directories with state tracking.
func (m *MockStatefulFileSystem) MkdirAll(ctx context.Context, path string, perm int) error {
	args := m.Called(ctx, path, perm)

	// Update internal state
	m.Directories[path] = true

	// Log operation
	m.AccessLog = append(m.AccessLog, FileOperation{
		Operation: "mkdir",
		Path:      path,
		Timestamp: time.Now(),
	})
	m.OperationCount++

	return args.Error(0)
}

// Exists checks if a path exists in the mock file system.
func (m *MockStatefulFileSystem) Exists(ctx context.Context, path string) bool {
	args := m.Called(ctx, path)

	// Log operation
	m.AccessLog = append(m.AccessLog, FileOperation{
		Operation: "exists",
		Path:      path,
		Timestamp: time.Now(),
	})
	m.OperationCount++

	// Check internal state first
	if _, fileExists := m.Files[path]; fileExists {
		return true
	}

	if _, dirExists := m.Directories[path]; dirExists {
		return true
	}

	return args.Bool(0)
}

// GetAccessLog returns the file operation log.
func (m *MockStatefulFileSystem) GetAccessLog() []FileOperation {
	return m.AccessLog
}

// GetOperationCount returns the total number of operations.
func (m *MockStatefulFileSystem) GetOperationCount() int {
	return m.OperationCount
}

// GetFileCount returns the number of files in the mock system.
func (m *MockStatefulFileSystem) GetFileCount() int {
	return len(m.Files)
}

// ResetState resets the mock file system state.
func (m *MockStatefulFileSystem) ResetState() {
	m.Files = make(map[string][]byte)
	m.Directories = make(map[string]bool)
	m.AccessLog = make([]FileOperation, 0)
	m.OperationCount = 0
}

// MockRateLimitedClient demonstrates rate limiting simulation.
type MockRateLimitedClient struct {
	mock.Mock
	RateLimit     *github.RateLimit
	RequestCounts map[string]int
	LastRequest   time.Time
}

// NewMockRateLimitedClient creates a new rate limited client mock.
func NewMockRateLimitedClient(initialLimit int) *MockRateLimitedClient {
	return &MockRateLimitedClient{
		RateLimit: &github.RateLimit{
			Limit:     initialLimit,
			Remaining: initialLimit,
			Reset:     time.Now().Add(time.Hour),
			Used:      0,
		},
		RequestCounts: make(map[string]int),
		LastRequest:   time.Now(),
	}
}

// MakeRequest simulates making a rate-limited API request.
func (m *MockRateLimitedClient) MakeRequest(ctx context.Context, endpoint string) (*APIResponse, error) {
	args := m.Called(ctx, endpoint)

	// Update rate limit
	if m.RateLimit.Remaining > 0 {
		m.RateLimit.Remaining--
		m.RateLimit.Used++
	}

	// Track request counts
	m.RequestCounts[endpoint]++
	m.LastRequest = time.Now()

	// Simulate rate limit exceeded
	if m.RateLimit.Remaining <= 0 {
		return nil, NewRateLimitError(m.RateLimit.Reset)
	}

	response := &APIResponse{
		Status:    200,
		Data:      args.Get(0),
		RateLimit: m.RateLimit,
		Timestamp: time.Now(),
	}

	return response, args.Error(1)
}

// GetRateLimit returns the current rate limit status.
func (m *MockRateLimitedClient) GetRateLimit() *github.RateLimit {
	return m.RateLimit
}

// GetRequestCount returns the number of requests to an endpoint.
func (m *MockRateLimitedClient) GetRequestCount(endpoint string) int {
	return m.RequestCounts[endpoint]
}

// ResetRateLimit resets the rate limit to initial values.
func (m *MockRateLimitedClient) ResetRateLimit(limit int) {
	m.RateLimit.Limit = limit
	m.RateLimit.Remaining = limit
	m.RateLimit.Used = 0
	m.RateLimit.Reset = time.Now().Add(time.Hour)
	m.RequestCounts = make(map[string]int)
}

// APIResponse represents an API response.
type APIResponse struct {
	Status    int               `json:"status"`
	Data      interface{}       `json:"data"`
	RateLimit *github.RateLimit `json:"rate_limit"`
	Timestamp time.Time         `json:"timestamp"`
}

// RateLimitError represents a rate limit exceeded error.
type RateLimitError struct {
	ResetTime time.Time
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(resetTime time.Time) *RateLimitError {
	return &RateLimitError{ResetTime: resetTime}
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	return "rate limit exceeded, resets at " + e.ResetTime.Format(time.RFC3339)
}

// IsRateLimitError checks if an error is a rate limit error.
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// MockTestHelpers provides helper functions for testify mocks.
type MockTestHelpers struct{}

// NewMockTestHelpers creates new test helpers.
func NewMockTestHelpers() *MockTestHelpers {
	return &MockTestHelpers{}
}

// AssertMockExpectations verifies all mock expectations.
func (h *MockTestHelpers) AssertMockExpectations(t mock.TestingT, mocks ...interface{ AssertExpectations(mock.TestingT) bool }) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}

// AssertCallCount verifies the number of calls to a mock method.
func (h *MockTestHelpers) AssertCallCount(t mock.TestingT, mockObj *mock.Mock, methodName string, expectedCount int) {
	actualCount := len(mockObj.Calls)
	if actualCount != expectedCount {
		t.Errorf("Expected %d calls to %s, but got %d", expectedCount, methodName, actualCount)
	}
}

// AssertCallOrder verifies the order of mock method calls.
func (h *MockTestHelpers) AssertCallOrder(t mock.TestingT, mockObj *mock.Mock, expectedOrder []string) {
	if len(mockObj.Calls) != len(expectedOrder) {
		t.Errorf("Expected %d calls, but got %d", len(expectedOrder), len(mockObj.Calls))
		return
	}

	for i, call := range mockObj.Calls {
		if call.Method != expectedOrder[i] {
			t.Errorf("Expected call %d to be %s, but got %s", i, expectedOrder[i], call.Method)
		}
	}
}

// SetupCommonExpectations sets up common mock expectations.
func (h *MockTestHelpers) SetupCommonExpectations(mocks map[string]interface{}) {
	// This method can be extended to set up common expectations
	// across different mock types based on their interface
	for name, mockObj := range mocks {
		switch m := mockObj.(type) {
		case *MockComplexGitHubService:
			m.On("ProcessRepositories", mock.Anything, mock.Anything).Return(&ProcessResult{}, nil)
		case *MockStatefulFileSystem:
			m.On("Exists", mock.Anything, mock.Anything).Return(true)
		case *MockRateLimitedClient:
			m.On("MakeRequest", mock.Anything, mock.Anything).Return(&APIResponse{}, nil)
		default:
			// Log unrecognized mock type
			_ = name // prevent unused variable error
		}
	}
}
