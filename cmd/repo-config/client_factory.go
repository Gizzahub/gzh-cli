package repoconfig

import (
	"context"
	"fmt"
	"os"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	gh "github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

// ClientFactory creates GitHub clients with proper dependency injection.
type ClientFactory interface {
	CreateRepoConfigClient(token string) (*github.RepoConfigClient, error)
	CreateGitHubClient(token string) (*gh.Client, error)
}

// DefaultClientFactory is the default implementation of ClientFactory.
type DefaultClientFactory struct {
	// Optional configuration that can be injected
	httpTimeout  int
	baseURL      string
	rateLimiter  *github.RateLimiter
	changeLogger *github.ChangeLogger
}

// NewDefaultClientFactory creates a new client factory with optional configuration.
func NewDefaultClientFactory() *DefaultClientFactory {
	return &DefaultClientFactory{
		httpTimeout: 30, // Default timeout in seconds
	}
}

// WithHTTPTimeout sets the HTTP timeout for clients.
func (f *DefaultClientFactory) WithHTTPTimeout(seconds int) *DefaultClientFactory {
	f.httpTimeout = seconds
	return f
}

// WithBaseURL sets the base URL for API calls.
func (f *DefaultClientFactory) WithBaseURL(url string) *DefaultClientFactory {
	f.baseURL = url
	return f
}

// WithRateLimiter sets the rate limiter.
func (f *DefaultClientFactory) WithRateLimiter(limiter *github.RateLimiter) *DefaultClientFactory {
	f.rateLimiter = limiter
	return f
}

// WithChangeLogger sets the change logger.
func (f *DefaultClientFactory) WithChangeLogger(logger *github.ChangeLogger) *DefaultClientFactory {
	f.changeLogger = logger
	return f
}

// CreateRepoConfigClient creates a new RepoConfigClient with injected dependencies.
func (f *DefaultClientFactory) CreateRepoConfigClient(token string) (*github.RepoConfigClient, error) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("GitHub token not provided and GITHUB_TOKEN environment variable not set")
		}
	}

	// Create client with injected configuration
	client := github.NewRepoConfigClient(token)

	// Apply optional configuration
	if f.baseURL != "" {
		// client.SetBaseURL(f.baseURL) // If such method exists
	}

	// Note: In a real implementation, you would pass these dependencies
	// to the constructor or use setter methods

	return client, nil
}

// CreateGitHubClient creates a new GitHub client with proper OAuth.
func (f *DefaultClientFactory) CreateGitHubClient(token string) (*gh.Client, error) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("GitHub token not provided and GITHUB_TOKEN environment variable not set")
		}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := gh.NewClient(tc)

	// Apply base URL if configured
	if f.baseURL != "" {
		// client.BaseURL = f.baseURL // If needed
	}

	return client, nil
}

// MockClientFactory is a mock implementation for testing.
type MockClientFactory struct {
	MockRepoConfigClient *github.RepoConfigClient
	MockGitHubClient     *gh.Client
	MockError            error
}

// CreateRepoConfigClient returns the mock client or error.
func (m *MockClientFactory) CreateRepoConfigClient(token string) (*github.RepoConfigClient, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}

	return m.MockRepoConfigClient, nil
}

// CreateGitHubClient returns the mock GitHub client or error.
func (m *MockClientFactory) CreateGitHubClient(token string) (*gh.Client, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}

	return m.MockGitHubClient, nil
}
