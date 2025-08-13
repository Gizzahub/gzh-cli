// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package builders

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/Gizzahub/gzh-manager-go/internal/env"
	"github.com/Gizzahub/gzh-manager-go/pkg/github"
)

// Sentinel errors for mock implementations.
var (
	ErrMockNotConfigured = errors.New("mock function not configured")
)

// MockLoggerBuilder provides a fluent interface for building test loggers.
type MockLoggerBuilder struct {
	logger *MockLogger
}

// MockLogger implements the Logger interface for testing.
type MockLogger struct {
	DebugCalls []LogCall
	InfoCalls  []LogCall
	WarnCalls  []LogCall
	ErrorCalls []LogCall
}

// LogCall represents a call to a logging method.
type LogCall struct {
	Message string
	Args    []interface{}
}

// NewMockLoggerBuilder creates a new MockLoggerBuilder.
func NewMockLoggerBuilder() *MockLoggerBuilder {
	return &MockLoggerBuilder{
		logger: &MockLogger{
			DebugCalls: []LogCall{},
			InfoCalls:  []LogCall{},
			WarnCalls:  []LogCall{},
			ErrorCalls: []LogCall{},
		},
	}
}

// Build returns the constructed mock logger.
func (b *MockLoggerBuilder) Build() *MockLogger {
	return b.logger
}

// Debug logs a debug message.
func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.DebugCalls = append(m.DebugCalls, LogCall{Message: msg, Args: args})
}

// Info logs an info message.
func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.InfoCalls = append(m.InfoCalls, LogCall{Message: msg, Args: args})
}

// Warn logs a warning message.
func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.WarnCalls = append(m.WarnCalls, LogCall{Message: msg, Args: args})
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.ErrorCalls = append(m.ErrorCalls, LogCall{Message: msg, Args: args})
}

// MockHTTPClientBuilder provides a fluent interface for building test HTTP clients.
type MockHTTPClientBuilder struct {
	client *MockHTTPClient
}

// MockHTTPClient implements the HTTPClient interface for testing.
type MockHTTPClient struct {
	GetResponses  map[string]*http.Response
	GetErrors     map[string]error
	PostResponses map[string]*http.Response
	PostErrors    map[string]error
	GetCalls      []string
	PostCalls     []PostCall
}

// PostCall represents a call to the Post method.
type PostCall struct {
	URL         string
	ContentType string
	Body        string
}

// NewMockHTTPClientBuilder creates a new MockHTTPClientBuilder.
func NewMockHTTPClientBuilder() *MockHTTPClientBuilder {
	return &MockHTTPClientBuilder{
		client: &MockHTTPClient{
			GetResponses:  make(map[string]*http.Response),
			GetErrors:     make(map[string]error),
			PostResponses: make(map[string]*http.Response),
			PostErrors:    make(map[string]error),
			GetCalls:      []string{},
			PostCalls:     []PostCall{},
		},
	}
}

// WithGetResponse configures a GET response for a specific URL.
func (b *MockHTTPClientBuilder) WithGetResponse(url string, response *http.Response) *MockHTTPClientBuilder {
	b.client.GetResponses[url] = response
	return b
}

// WithGetError configures a GET error for a specific URL.
func (b *MockHTTPClientBuilder) WithGetError(url string, err error) *MockHTTPClientBuilder {
	b.client.GetErrors[url] = err
	return b
}

// WithPostResponse configures a POST response for a specific URL.
func (b *MockHTTPClientBuilder) WithPostResponse(url string, response *http.Response) *MockHTTPClientBuilder {
	b.client.PostResponses[url] = response
	return b
}

// WithPostError configures a POST error for a specific URL.
func (b *MockHTTPClientBuilder) WithPostError(url string, err error) *MockHTTPClientBuilder {
	b.client.PostErrors[url] = err
	return b
}

// Build returns the constructed mock HTTP client.
func (b *MockHTTPClientBuilder) Build() *MockHTTPClient {
	return b.client
}

// Get performs a mock HTTP GET request.
func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	m.GetCalls = append(m.GetCalls, url)

	if err, exists := m.GetErrors[url]; exists {
		return nil, err
	}

	if response, exists := m.GetResponses[url]; exists {
		return response, nil
	}

	return nil, ErrMockNotConfigured
}

// Post performs a mock HTTP POST request.
func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	bodyStr := ""

	if body != nil {
		bodyBytes, err := io.ReadAll(body)
		if err == nil {
			bodyStr = string(bodyBytes)
		}
	}

	m.PostCalls = append(m.PostCalls, PostCall{
		URL:         url,
		ContentType: contentType,
		Body:        bodyStr,
	})

	if err, exists := m.PostErrors[url]; exists {
		return nil, err
	}

	if response, exists := m.PostResponses[url]; exists {
		return response, nil
	}

	return nil, ErrMockNotConfigured
}

// MockGitHubProviderFactoryBuilder provides a fluent interface for building GitHub provider factories.
type MockGitHubProviderFactoryBuilder struct {
	factory *MockGitHubProviderFactory
}

// MockGitHubProviderFactory implements the GitHubProviderFactory interface for testing.
type MockGitHubProviderFactory struct {
	CreateClonerFunc        func(ctx context.Context, token string) (github.GitHubCloner, error)
	CreateClonerWithEnvFunc func(ctx context.Context, token string, environment env.Environment) (github.GitHubCloner, error)
	CreateChangeLoggerFunc  func(ctx context.Context, changelog *github.ChangeLog, options *github.LoggerOptions) (*github.ChangeLogger, error)
	ProviderName            string
}

// NewMockGitHubProviderFactoryBuilder creates a new MockGitHubProviderFactoryBuilder.
func NewMockGitHubProviderFactoryBuilder() *MockGitHubProviderFactoryBuilder {
	return &MockGitHubProviderFactoryBuilder{
		factory: &MockGitHubProviderFactory{
			ProviderName: "github",
		},
	}
}

// WithCreateClonerFunc sets the CreateCloner function.
func (b *MockGitHubProviderFactoryBuilder) WithCreateClonerFunc(fn func(ctx context.Context, token string) (github.GitHubCloner, error)) *MockGitHubProviderFactoryBuilder {
	b.factory.CreateClonerFunc = fn
	return b
}

// WithCreateClonerWithEnvFunc sets the CreateClonerWithEnv function.
func (b *MockGitHubProviderFactoryBuilder) WithCreateClonerWithEnvFunc(fn func(ctx context.Context, token string, environment env.Environment) (github.GitHubCloner, error)) *MockGitHubProviderFactoryBuilder {
	b.factory.CreateClonerWithEnvFunc = fn
	return b
}

// WithProviderName sets the provider name.
func (b *MockGitHubProviderFactoryBuilder) WithProviderName(name string) *MockGitHubProviderFactoryBuilder {
	b.factory.ProviderName = name
	return b
}

// Build returns the constructed mock GitHub provider factory.
func (b *MockGitHubProviderFactoryBuilder) Build() *MockGitHubProviderFactory {
	return b.factory
}

// CreateCloner creates a mock GitHub cloner.
func (m *MockGitHubProviderFactory) CreateCloner(ctx context.Context, token string) (github.GitHubCloner, error) {
	if m.CreateClonerFunc != nil {
		return m.CreateClonerFunc(ctx, token)
	}

	return nil, ErrMockNotConfigured
}

// CreateClonerWithEnv creates a mock GitHub cloner with environment.
func (m *MockGitHubProviderFactory) CreateClonerWithEnv(ctx context.Context, token string, environment env.Environment) (github.GitHubCloner, error) {
	if m.CreateClonerWithEnvFunc != nil {
		return m.CreateClonerWithEnvFunc(ctx, token, environment)
	}

	return nil, ErrMockNotConfigured
}

// CreateChangeLogger creates a mock change logger.
func (m *MockGitHubProviderFactory) CreateChangeLogger(ctx context.Context, changelog *github.ChangeLog, options *github.LoggerOptions) (*github.ChangeLogger, error) {
	if m.CreateChangeLoggerFunc != nil {
		return m.CreateChangeLoggerFunc(ctx, changelog, options)
	}

	return nil, ErrMockNotConfigured
}

// GetProviderName returns the mock provider name.
func (m *MockGitHubProviderFactory) GetProviderName() string {
	return m.ProviderName
}
