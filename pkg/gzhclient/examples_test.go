package gzhclient_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/gzhclient"
)

// Example_basicUsage demonstrates basic client usage
func Example_basicUsage() {
	// Create client with default configuration
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Check client health
	health := client.Health()
	fmt.Printf("Client status: %s\n", health.Overall)

	// Output: Client status: healthy
}

// Example_bulkClone demonstrates bulk repository cloning
func Example_bulkClone() {
	config := gzhclient.DefaultConfig()
	config.Timeout = 5 * time.Minute

	client, err := gzhclient.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

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
	fmt.Printf("Duration: %v\n", result.Duration)

	// Output: Total repositories: 15
	// Successfully cloned: 14
	// Failed: 1
	// Duration: 2m30s
}

// Example_pluginManagement demonstrates plugin operations
func Example_pluginManagement() {
	config := gzhclient.DefaultConfig()
	config.EnablePlugins = true
	config.PluginDir = "./plugins"

	client, err := gzhclient.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// List available plugins
	plugins, err := client.ListPlugins()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded plugins: %d\n", len(plugins))
	for _, plugin := range plugins {
		fmt.Printf("- %s v%s: %s\n", plugin.Name, plugin.Version, plugin.Description)
	}

	// Execute a plugin method
	if len(plugins) > 0 {
		result, err := client.ExecutePlugin(context.Background(), gzhclient.PluginExecuteRequest{
			PluginName: plugins[0].Name,
			Method:     "info",
			Args:       map[string]interface{}{},
			Timeout:    30 * time.Second,
		})
		if err != nil {
			log.Printf("Plugin execution failed: %v", err)
		} else {
			fmt.Printf("Plugin result: %v\n", result.Result)
		}
	}

	// Output: Loaded plugins: 2
	// - security-scanner v1.0.0: Scans repositories for security vulnerabilities
	// - code-formatter v0.5.1: Formats code according to style guidelines
	// Plugin result: map[status:ok version:1.0.0]
}

// Example_platformSpecificClients demonstrates platform-specific client usage
func Example_platformSpecificClients() {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// GitHub operations
	githubClient := client.GitHubClient("ghp_xxxxxxxxxxxxxxxxxxxx")
	fmt.Printf("GitHub client created: %T\n", githubClient)

	// GitLab operations
	gitlabClient := client.GitLabClient("https://gitlab.com", "glpat-xxxxxxxxxxxxxxxxxxxx")
	fmt.Printf("GitLab client created: %T\n", gitlabClient)

	// Gitea operations
	giteaClient := client.GiteaClient("https://git.example.com", "your-gitea-token")
	fmt.Printf("Gitea client created: %T\n", giteaClient)

	// Output: GitHub client created: *github.Client
	// GitLab client created: *gitlab.Client
	// Gitea client created: *gitea.Client
}

// Example_systemMonitoring demonstrates system metrics collection
func Example_systemMonitoring() {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

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
	// Memory Total: 16 GB
	// Disk Usage: 65.2%
	// System Uptime: 24h0m0s
}

// Example_configurationOptions demonstrates various configuration options
func Example_configurationOptions() {
	// Custom configuration
	config := gzhclient.ClientConfig{
		Timeout:       60 * time.Second,
		RetryCount:    5,
		EnablePlugins: true,
		PluginDir:     "/opt/gzh-plugins",
		LogLevel:      "debug",
		LogFile:       "/var/log/gzh-client.log",
		Features: gzhclient.FeatureFlags{
			BulkClone:  true,
			DevEnv:     true,
			NetEnv:     false, // Disable network environment features
			Monitoring: true,
			Plugins:    true,
		},
	}

	client, err := gzhclient.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Get current configuration
	currentConfig := client.GetConfig()
	fmt.Printf("Timeout: %v\n", currentConfig.Timeout)
	fmt.Printf("Plugin directory: %s\n", currentConfig.PluginDir)
	fmt.Printf("Network features enabled: %t\n", currentConfig.Features.NetEnv)

	// Update configuration
	newConfig := currentConfig
	newConfig.Timeout = 120 * time.Second
	if err := client.UpdateConfig(newConfig); err != nil {
		log.Printf("Failed to update config: %v", err)
	}

	// Output: Timeout: 1m0s
	// Plugin directory: /opt/gzh-plugins
	// Network features enabled: false
}

// Example_errorHandling demonstrates proper error handling
func Example_errorHandling() {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Attempt to execute a non-existent plugin
	_, err = client.ExecutePlugin(context.Background(), gzhclient.PluginExecuteRequest{
		PluginName: "non-existent-plugin",
		Method:     "test",
		Args:       map[string]interface{}{},
	})
	if err != nil {
		// Check for specific error types
		if apiErr, ok := err.(*gzhclient.APIError); ok {
			fmt.Printf("API Error: %s - %s\n", apiErr.Code, apiErr.Message)
		} else {
			fmt.Printf("General error: %v\n", err)
		}
	}

	// Output: General error: plugin manager not initialized
}

// Example_contextCancellation demonstrates context-based cancellation
func Example_contextCancellation() {
	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// This operation will be cancelled if it takes longer than 10 seconds
	req := gzhclient.BulkCloneRequest{
		Platforms: []gzhclient.PlatformConfig{
			{
				Type:          "github",
				Organizations: []string{"large-organization"},
			},
		},
		OutputDir:   "./repos",
		Concurrency: 1,
	}

	_, err = client.BulkClone(ctx, req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Operation cancelled due to timeout")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	// Output: Operation cancelled due to timeout
}
