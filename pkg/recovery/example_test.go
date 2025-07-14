package recovery_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/gizzahub/gzh-manager-go/pkg/gitlab"
	"github.com/gizzahub/gzh-manager-go/pkg/recovery"
)

// ExampleResilientHTTPClient demonstrates basic usage of the resilient HTTP client
func ExampleResilientHTTPClient() {
	// Create a resilient HTTP client with default configuration
	client := recovery.NewGenericClient()
	defer client.Close()

	// Use the client for HTTP requests
	resp, err := client.GetWithContext(context.Background(), "https://httpbin.org/status/200")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %s\n", resp.Status)
	// Output: Response status: 200 OK
}

// ExampleResilientHTTPClient_withRetry demonstrates retry behavior
func ExampleResilientHTTPClient_withRetry() {
	// Create a client optimized for quick operations
	client := recovery.NewQuickClient()
	defer client.Close()

	// This request will be retried automatically if it fails
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.GetWithContext(ctx, "https://httpbin.org/status/500")
	if err != nil {
		log.Printf("Request failed after retries: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %s\n", resp.Status)
}

// ExampleResilientGitHubClient demonstrates GitHub API usage with resilience
func ExampleResilientGitHubClient() {
	// Create a resilient GitHub client
	client := github.NewResilientGitHubClient("your-github-token")
	defer client.Close()

	ctx := context.Background()

	// Get rate limit information
	rateLimit, err := client.GetRateLimit(ctx)
	if err != nil {
		log.Printf("Failed to get rate limit: %v", err)
		return
	}

	fmt.Printf("Rate limit: %d/%d remaining\n", rateLimit.Remaining, rateLimit.Limit)

	// Check if we're rate limited
	if rateLimit.IsRateLimited() {
		fmt.Printf("Rate limited! Reset in: %v\n", rateLimit.TimeUntilReset())
		return
	}

	// List repositories for an organization
	repos, err := client.ListRepositories(ctx, "golang")
	if err != nil {
		log.Printf("Failed to list repositories: %v", err)
		return
	}

	fmt.Printf("Found %d repositories\n", len(repos))

	// Get default branch for a specific repository
	if len(repos) > 0 {
		branch, err := client.GetDefaultBranch(ctx, "golang", repos[0])
		if err != nil {
			log.Printf("Failed to get default branch: %v", err)
			return
		}
		fmt.Printf("Default branch for %s: %s\n", repos[0], branch)
	}
}

// ExampleResilientGitLabClient demonstrates GitLab API usage with resilience
func ExampleResilientGitLabClient() {
	// Create a resilient GitLab client
	client := gitlab.NewResilientGitLabClient("https://gitlab.com", "your-gitlab-token")
	defer client.Close()

	ctx := context.Background()

	// List accessible groups
	groups, err := client.ListGroups(ctx)
	if err != nil {
		log.Printf("Failed to list groups: %v", err)
		return
	}

	fmt.Printf("Found %d groups\n", len(groups))

	// List projects for the first group
	if len(groups) > 0 {
		groupID := fmt.Sprintf("%d", groups[0].ID)
		projects, err := client.ListGroupProjects(ctx, groupID)
		if err != nil {
			log.Printf("Failed to list projects: %v", err)
			return
		}

		fmt.Printf("Found %d projects in group %s\n", len(projects), groups[0].Name)

		// Get detailed information for the first project
		if len(projects) > 0 {
			projectID := fmt.Sprintf("%d", projects[0].ID)
			project, err := client.GetProject(ctx, projectID)
			if err != nil {
				log.Printf("Failed to get project details: %v", err)
				return
			}

			fmt.Printf("Project: %s, Default branch: %s\n", project.Name, project.DefaultBranch)
		}
	}
}

// ExampleHTTPClientFactory demonstrates using the factory for different scenarios
func ExampleHTTPClientFactory() {
	factory := recovery.NewHTTPClientFactory()

	// Create different types of clients for different scenarios
	githubClient := factory.CreateGitHubClient()
	gitlabClient := factory.CreateGitLabClient()
	giteaClient := factory.CreateGiteaClient()
	quickClient := factory.CreateQuickClient()
	longRunningClient := factory.CreateLongRunningClient()

	defer githubClient.Close()
	defer gitlabClient.Close()
	defer giteaClient.Close()
	defer quickClient.Close()
	defer longRunningClient.Close()

	fmt.Println("Created specialized HTTP clients for different services")

	// Each client is optimized for its specific use case:
	// - GitHub client: Higher failure threshold, longer timeouts
	// - GitLab client: Moderate settings for GitLab.com
	// - Gitea client: Lower timeouts for self-hosted instances
	// - Quick client: Fast-fail, minimal retries
	// - Long-running client: Extended timeouts, more retries
}

// ExampleNetworkErrorClassifier demonstrates error classification
func ExampleNetworkErrorClassifier() {
	classifier := recovery.NewNetworkErrorClassifier()

	// Simulate different types of network errors
	errors := []error{
		fmt.Errorf("connection refused"),
		fmt.Errorf("timeout"),
		fmt.Errorf("network unreachable"),
		fmt.Errorf("unknown error"),
	}

	for _, err := range errors {
		errorType, retryable := classifier.ClassifyError(err)
		fmt.Printf("Error: %v -> Type: %v, Retryable: %v\n", err, errorType, retryable)
	}
}

// ExampleCustomConfiguration demonstrates creating clients with custom settings
func ExampleCustomConfiguration() {
	// Create custom configuration for a specific use case
	config := recovery.DefaultResilientHTTPClientConfig()
	config.MaxRetries = 5
	config.InitialDelay = 2 * time.Second
	config.MaxDelay = 30 * time.Second
	config.UseCircuitBreaker = true
	config.CircuitConfig.FailureThreshold = 10

	// Create a client with the custom configuration
	client := recovery.NewResilientHTTPClient(config)
	defer client.Close()

	fmt.Println("Created HTTP client with custom configuration")

	// Use the client
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.GetWithContext(ctx, "https://httpbin.org/delay/2")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Request completed: %s\n", resp.Status)
}

// ExampleCircuitBreakerIntegration demonstrates circuit breaker behavior
func ExampleCircuitBreakerIntegration() {
	// Create a client with aggressive circuit breaker settings for demonstration
	config := recovery.DefaultResilientHTTPClientConfig()
	config.CircuitConfig.FailureThreshold = 2
	config.CircuitConfig.Timeout = 5 * time.Second
	config.MaxRetries = 1

	client := recovery.NewResilientHTTPClient(config)
	defer client.Close()

	ctx := context.Background()

	// Make several requests that will fail to trigger circuit breaker
	for i := 0; i < 5; i++ {
		_, err := client.GetWithContext(ctx, "https://httpbin.org/status/500")
		if err != nil {
			fmt.Printf("Request %d failed: %v\n", i+1, err)
		}

		// Check circuit breaker state
		stats := client.GetStats()
		if cbStats, ok := stats["circuit_breaker"]; ok {
			if cbMap, ok := cbStats.(map[string]interface{}); ok {
				if state, ok := cbMap["state"]; ok {
					fmt.Printf("Circuit breaker state: %v\n", state)
				}
			}
		}
	}

	// Wait for circuit breaker to potentially reset
	time.Sleep(6 * time.Second)

	// Try one more request
	_, err := client.GetWithContext(ctx, "https://httpbin.org/status/200")
	if err != nil {
		fmt.Printf("Final request failed: %v\n", err)
	} else {
		fmt.Println("Circuit breaker recovered, request succeeded")
	}
}
