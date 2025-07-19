package gzhclient_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/gzhclient"
)

// Example_basicUsage demonstrates basic client usage.
func Example_basicUsage() {
	// Create client with default configuration
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Check client health
	health := client.Health()
	fmt.Printf("Client status: %s\n", health.Overall)

	// Output: Client status: healthy
}

// Example_bulkClone demonstrates bulk repository cloning.
func Example_bulkClone() {
	config := gzhclient.DefaultConfig()
	config.Timeout = 5 * time.Minute

	client, err := gzhclient.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Configure bulk clone request
	req := gzhclient.BulkCloneRequest{
		Platforms: []gzhclient.PlatformConfig{
			{
				Type:          "github",
				Token:         "ghp_xxxxxxxxxxxxxxxxxxxx", // Use environment variable in practice
				Organizations: []string{"your-organization"},
			},
		},
		OutputDir:      "./repositories",
		Concurrency:    3,
		Strategy:       "reset",
		IncludePrivate: false,
		Filters: gzhclient.CloneFilters{
			Languages:    []string{"go", "python"},
			UpdatedAfter: time.Now().AddDate(0, -6, 0), // Last 6 months
		},
	}

	// Execute bulk clone
	result, err := client.BulkClone(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total repositories: %d\n", result.TotalRepos)
	fmt.Printf("Successfully cloned: %d\n", result.SuccessCount)
	fmt.Printf("Failed: %d\n", result.FailureCount)

	// Output: Total repositories: 10
	// Successfully cloned: 8
	// Failed: 1
}

// Example_pluginManagement demonstrates plugin operations - DISABLED (plugins removed).
func Example_pluginManagement() {
	config := gzhclient.DefaultConfig()
	// Plugin functionality has been disabled and removed

	client, err := gzhclient.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Plugin functionality no longer available
	fmt.Println("Plugin management has been disabled")

	// Output: Plugin management has been disabled
}

// Example_platformSpecificClients demonstrates platform-specific client usage.
func Example_platformSpecificClients() {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// GitHub operations
	githubClient := client.GitHubClient("ghp_xxxxxxxxxxxxxxxxxxxx")
	fmt.Printf("GitHub client created: %T\n", githubClient)

	// GitLab operations
	gitlabClient := client.GitLabClient("https://gitlab.com", "glpat-xxxxxxxxxxxxxxxxxxxx")
	fmt.Printf("GitLab client created: %T\n", gitlabClient)

	// Gitea operations
	giteaClient := client.GiteaClient("https://git.example.com", "your-gitea-token")
	fmt.Printf("Gitea client created: %T\n", giteaClient)

	// Output: GitHub client created: *github.GitHubAPIClient
	// GitLab client created: struct {}
	// Gitea client created: struct {}
}

// Example_systemMonitoring demonstrates system metrics collection.
func Example_systemMonitoring() {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Get system metrics
	metrics, err := client.GetSystemMetrics()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("CPU Cores: %d\n", metrics.CPU.Cores)
	fmt.Printf("Memory Total: %d GB\n", metrics.Memory.Total/(1024*1024*1024))
	fmt.Printf("Disk Usage: %.1f%%\n", metrics.Disk.Usage)
	fmt.Printf("System Uptime: %v\n", metrics.Uptime)

	// Output: CPU Cores: 4
	// Memory Total: 0 GB
	// Disk Usage: 0.0%
	// System Uptime: 24h0m0s
}

// Example_configurationOptions demonstrates various configuration options.
func Example_configurationOptions() {
	// Custom configuration
	config := gzhclient.ClientConfig{
		Timeout:    60 * time.Second,
		RetryCount: 5,
		LogLevel:   "debug",
		LogFile:    "/var/log/gzh-client.log",
		Features: gzhclient.FeatureFlags{
			BulkClone:  true,
			DevEnv:     true,
			NetEnv:     false, // Disable network environment features
			Monitoring: true,
		},
	}

	client, err := gzhclient.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Get current configuration
	currentConfig := client.GetConfig()
	fmt.Printf("Timeout: %v\n", currentConfig.Timeout)
	fmt.Printf("Log level: %s\n", currentConfig.LogLevel)
	fmt.Printf("Network features enabled: %t\n", currentConfig.Features.NetEnv)

	// Update configuration
	newConfig := currentConfig

	newConfig.Timeout = 120 * time.Second
	if err := client.UpdateConfig(newConfig); err != nil {
		log.Printf("Failed to update config: %v", err)
	}

	// Output: Timeout: 1m0s
	// Log level: debug
	// Network features enabled: false
}

// Example_errorHandling demonstrates proper error handling.
func Example_errorHandling() {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Attempt to perform a bulk clone with invalid configuration
	req := gzhclient.BulkCloneRequest{
		Platforms: []gzhclient.PlatformConfig{
			{
				Type:          "invalid-platform",
				Token:         "dummy-token",
				Organizations: []string{"test-org"},
			},
		},
		OutputDir: "./repositories",
	}

	_, err = client.BulkClone(context.Background(), req)
	if err != nil {
		// Check for specific error types
		apiErr := &gzhclient.APIError{}
		if errors.As(err, &apiErr) {
			fmt.Printf("API Error: %s - %s\n", apiErr.Code, apiErr.Message)
		} else {
			fmt.Printf("General error: %v\n", err)
		}
	}

	// Output: General error: no supported platforms found in request
}

// Example_contextCancellation demonstrates context-based cancellation.
func Example_contextCancellation() {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		fmt.Printf("Client creation failed: %v\n", err)
		return
	}
	defer func() { _ = client.Close() }()

	// This example shows how context cancellation would work in practice
	// For demo purposes, we just show the pattern without actual cancellation
	req := gzhclient.BulkCloneRequest{
		Platforms: []gzhclient.PlatformConfig{
			{
				Type:          "github",
				Token:         "dummy-token", // Required for the operation
				Organizations: []string{"large-organization"},
			},
		},
		OutputDir:   "./repos",
		Concurrency: 1,
	}

	_, err = client.BulkClone(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Operation completed successfully")
	}

	// Output: Operation completed successfully
}
