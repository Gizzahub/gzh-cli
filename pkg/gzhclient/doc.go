// Package gzhclient provides a programmatic client interface for GZH Manager.
//
// The gzhclient package offers a comprehensive API for integrating GZH Manager
// functionality into Go applications. It provides high-level abstractions
// for bulk repository operations, plugin management, system monitoring,
// and event handling.
//
// # Quick Start
//
// Create a client with default configuration:
//
//	client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// # Features
//
// The client provides access to the following GZH Manager features:
//
//   - Bulk repository cloning from GitHub, GitLab, Gitea, and Gogs
//   - Plugin system for extending functionality
//   - System monitoring and metrics collection
//   - Event subscription and handling
//   - Development environment management
//   - Configuration generation and validation
//
// # Bulk Repository Operations
//
// Clone repositories from multiple platforms:
//
//	req := gzhclient.BulkCloneRequest{
//		Platforms: []gzhclient.PlatformConfig{
//			{
//				Type:          "github",
//				Token:         "your-github-token",
//				Organizations: []string{"your-org"},
//			},
//		},
//		OutputDir:   "./repos",
//		Concurrency: 5,
//		Strategy:    "reset",
//	}
//
//	result, err := client.BulkClone(context.Background(), req)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Cloned %d repositories successfully\n", result.SuccessCount)
//
// # Plugin Management
//
// List and execute plugins:
//
//	plugins, err := client.ListPlugins()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, plugin := range plugins {
//		fmt.Printf("Plugin: %s v%s\n", plugin.Name, plugin.Version)
//	}
//
//	// Execute a plugin method
//	result, err := client.ExecutePlugin(context.Background(), gzhclient.PluginExecuteRequest{
//		PluginName: "example-plugin",
//		Method:     "process",
//		Args: map[string]interface{}{
//			"input": "test data",
//		},
//	})
//
// # Configuration
//
// The client can be configured with various options:
//
//	config := gzhclient.ClientConfig{
//		Timeout:       60 * time.Second,
//		RetryCount:    5,
//		EnablePlugins: true,
//		PluginDir:     "/path/to/plugins",
//		LogLevel:      "debug",
//		Features: gzhclient.FeatureFlags{
//			BulkClone:  true,
//			DevEnv:     true,
//			NetEnv:     true,
//			Monitoring: true,
//			Plugins:    true,
//		},
//	}
//
//	client, err := gzhclient.NewClient(config)
//
// # Error Handling
//
// All methods return structured errors that can be type-asserted:
//
//	_, err := client.BulkClone(context.Background(), req)
//	if err != nil {
//		if apiErr, ok := err.(*gzhclient.APIError); ok {
//			fmt.Printf("API Error %s: %s\n", apiErr.Code, apiErr.Message)
//		} else {
//			fmt.Printf("Error: %v\n", err)
//		}
//	}
//
// # Thread Safety
//
// The client is safe for concurrent use by multiple goroutines.
// Each operation creates its own context and can be cancelled independently.
//
// # Platform-Specific Clients
//
// Access platform-specific functionality:
//
//	// GitHub operations
//	githubClient := client.GitHubClient("your-token")
//
//	// GitLab operations
//	gitlabClient := client.GitLabClient("https://gitlab.com", "your-token")
//
//	// Gitea operations
//	giteaClient := client.GiteaClient("https://git.example.com", "your-token")
//
// # Health Monitoring
//
// Monitor client and component health:
//
//	health := client.Health()
//	if health.Overall != gzhclient.StatusHealthy {
//		fmt.Printf("Client health: %s\n", health.Overall)
//		for component, status := range health.Components {
//			fmt.Printf("  %s: %s - %s\n", component, status.Status, status.Message)
//		}
//	}
//
// For more examples and detailed documentation, see the examples directory
// and the individual method documentation.
package gzhclient
