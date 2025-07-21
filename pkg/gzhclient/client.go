// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gzhclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	bulkclone "github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

// Client provides programmatic access to GZH Manager functionality.
type Client struct {
	config ClientConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// ClientConfig holds configuration for the GZH client.
type ClientConfig struct {
	// Connection settings
	ServerURL  string        `yaml:"server_url,omitempty" json:"server_url,omitempty"`
	APIKey     string        `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
	RetryCount int           `yaml:"retry_count" json:"retry_count"`

	// Plugin settings (disabled - plugins package removed)
	// PluginDir     string `yaml:"plugin_dir,omitempty" json:"plugin_dir,omitempty"`
	// EnablePlugins bool   `yaml:"enable_plugins" json:"enable_plugins"`

	// Logging settings
	LogLevel string `yaml:"log_level" json:"log_level"`
	LogFile  string `yaml:"log_file,omitempty" json:"log_file,omitempty"`

	// Feature flags
	Features FeatureFlags `yaml:"features" json:"features"`
}

// FeatureFlags enables/disables specific features.
type FeatureFlags struct {
	BulkClone  bool `yaml:"bulk_clone" json:"bulk_clone"`
	DevEnv     bool `yaml:"dev_env" json:"dev_env"`
	NetEnv     bool `yaml:"net_env" json:"net_env"`
	Monitoring bool `yaml:"monitoring" json:"monitoring"`
	// Plugins    bool `yaml:"plugins" json:"plugins"` // Disabled - plugins package removed
}

// DefaultConfig returns a default client configuration.
func DefaultConfig() ClientConfig {
	return ClientConfig{
		Timeout:    30 * time.Second,
		RetryCount: 3,
		// EnablePlugins: true, // Disabled - plugins package removed
		LogLevel: "info",
		Features: FeatureFlags{
			BulkClone:  true,
			DevEnv:     true,
			NetEnv:     true,
			Monitoring: true,
			// Plugins:    true, // Disabled - plugins package removed
		},
	}
}

// NewClient creates a new GZH Manager client.
func NewClient(config ClientConfig) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Plugin manager disabled - plugins package removed
	// if config.EnablePlugins && config.Features.Plugins {
	//	if err := client.initializePluginManager(); err != nil {
	//		cancel()
	//		return nil, fmt.Errorf("failed to initialize plugin manager: %w", err)
	//	}
	// }

	return client, nil
}

// Close cleanly shuts down the client.
func (c *Client) Close() error {
	// Plugin manager disabled - plugins package removed
	// if c.pluginManager != nil {
	//	if err := c.pluginManager.Shutdown(); err != nil {
	//		return fmt.Errorf("failed to shutdown plugin manager: %w", err)
	//	}
	// }
	c.cancel()
	return nil
}

// GetConfig returns the current client configuration.
func (c *Client) GetConfig() ClientConfig {
	return c.config
}

// UpdateConfig updates the client configuration.
func (c *Client) UpdateConfig(config ClientConfig) error {
	c.config = config

	// Plugin manager disabled - plugins package removed
	// if config.EnablePlugins && config.Features.Plugins && c.pluginManager == nil {
	//	return c.initializePluginManager()
	// } else if (!config.EnablePlugins || !config.Features.Plugins) && c.pluginManager != nil {
	//	if err := c.pluginManager.Shutdown(); err != nil {
	//		return fmt.Errorf("failed to shutdown plugin manager: %w", err)
	//	}
	//	c.pluginManager = nil
	// }

	return nil
}

// Health checks the health of the client and its components.
func (c *Client) Health() HealthStatus {
	status := HealthStatus{
		Overall:    StatusHealthy,
		Components: make(map[string]ComponentHealth),
		Timestamp:  time.Now(),
	}

	// Plugin manager disabled - plugins package removed
	// if c.pluginManager != nil {
	//	pluginErrors := c.pluginManager.HealthCheck()
	//	if len(pluginErrors) > 0 {
	//		status.Components["plugins"] = ComponentHealth{
	//			Status:  StatusUnhealthy,
	//			Message: fmt.Sprintf("%d plugin(s) unhealthy", len(pluginErrors)),
	//			Details: pluginErrors,
	//		}
	//		status.Overall = StatusDegraded
	//	} else {
	//		status.Components["plugins"] = ComponentHealth{
	//			Status:  StatusHealthy,
	//			Message: "All plugins healthy",
	//		}
	//	}
	// }

	// Add more component checks here
	status.Components["client"] = ComponentHealth{
		Status:  StatusHealthy,
		Message: "Client operational",
	}

	return status
}

// initializePluginManager sets up the plugin manager - DISABLED (plugins package removed)
// func (c *Client) initializePluginManager() error {
//	if c.config.PluginDir == "" {
//		return fmt.Errorf("plugin directory not configured")
//	}
//
//	// Create plugin manager configuration
//	managerConfig := plugins.ManagerConfig{
//		PluginDir:           c.config.PluginDir,
//		EnableSandbox:       true,
//		LoadTimeout:         30 * time.Second,
//		ExecuteTimeout:      c.config.Timeout,
//		HealthCheckInterval: 60 * time.Second,
//		DefaultLimits: plugins.ResourceLimits{
//			MaxMemoryMB:      256,
//			MaxCPUPercent:    25.0,
//			MaxExecutionTime: 5 * time.Minute,
//			MaxFileHandles:   100,
//		},
//	}
//
//	// Create plugin API
//	eventBus := plugins.NewEventBus()
//	securityMgr := plugins.NewSecurityManager(true)
//	hostInfo := plugins.HostInfo{
//		GZVersion:    "1.0.0",   // This should come from build info
//		OS:           "unknown", // This should be detected
//		Architecture: "unknown", // This should be detected
//		WorkingDir:   ".",
//		ConfigDir:    "~/.config/gzh-manager",
//		PluginDir:    c.config.PluginDir,
//	}
//
//	api := plugins.NewDefaultPluginAPI(eventBus, securityMgr, hostInfo)
//
//	// Create plugin manager
//	manager := plugins.NewManager(managerConfig, api)
//	api.SetManager(manager)
//
//	// Load plugins from directory
//	if err := manager.LoadPluginsFromDirectory(c.config.PluginDir); err != nil {
//		return fmt.Errorf("failed to load plugins: %w", err)
//	}
//
//	c.pluginManager = manager
//	return nil
// }

// BulkClone performs bulk repository cloning operation.
func (c *Client) BulkClone(ctx context.Context, req BulkCloneRequest) (*BulkCloneResult, error) {
	// Create bulk clone manager with configuration manager and logger
	configManager := &configManagerImpl{}
	logger := &silentLoggerImpl{}

	manager := bulkclone.NewBulkCloneManager(configManager, logger)

	// Create organization clone request for GitHub platforms
	for _, platform := range req.Platforms {
		if platform.Type == "github" {
			for _, org := range platform.Organizations {
				orgRequest := &bulkclone.OrganizationCloneRequest{
					Provider:     platform.Type,
					Organization: org,
					TargetPath:   req.OutputDir,
					Strategy:     req.Strategy,
					Token:        platform.Token,
					Concurrency:  req.Concurrency,
					DryRun:       false,
				}

				result, err := manager.CloneOrganization(ctx, orgRequest)
				if err != nil {
					return nil, fmt.Errorf("bulk clone operation failed for org %s: %w", org, err)
				}

				// Convert result to client response format
				response := &BulkCloneResult{
					TotalRepos:   result.TotalRepositories,
					SuccessCount: result.ClonesSuccessful,
					FailureCount: result.ClonesFailed,
					SkippedCount: result.ClonesSkipped,
					Duration:     result.ExecutionTime,
					Results:      make([]RepositoryCloneResult, len(result.RepositoryResults)),
					Summary:      make(map[string]interface{}),
				}

				for i, repo := range result.RepositoryResults {
					response.Results[i] = RepositoryCloneResult{
						RepoName:  repo.Repository,
						Platform:  repo.Provider,
						URL:       "", // Not available in RepositoryResult
						LocalPath: repo.Path,
						Status:    getStatus(repo.Success),
						Error:     repo.Error,
						Duration:  repo.Duration,
						Size:      repo.SizeBytes,
					}
				}

				// Add statistics to summary
				if result.Statistics != nil {
					response.Summary["average_clone_time"] = result.Statistics.AverageCloneTime
					response.Summary["total_data_transferred"] = result.Statistics.TotalDataTransferred
					response.Summary["largest_repository"] = result.Statistics.LargestRepository
					response.Summary["errors_by_type"] = result.Statistics.ErrorsByType
				}

				return response, nil
			}
		}
	}

	return nil, fmt.Errorf("no supported platforms found in request")
}

// getStatus converts boolean success to string status.
func getStatus(success bool) string {
	if success {
		return "success"
	}

	return "failed"
}

// GitHubClient returns a GitHub-specific API client.
func (c *Client) GitHubClient(token string) github.APIClient {
	config := github.DefaultAPIClientConfig()
	config.Token = token

	httpClient := &httpClientWrapper{&http.Client{Timeout: c.config.Timeout}}
	logger := &silentLoggerImpl{}

	return github.NewAPIClient(config, httpClient, logger)
}

// GitLabClient returns a GitLab-specific client (placeholder).
func (c *Client) GitLabClient(baseURL, token string) interface{} {
	// GitLab client would be implemented here
	// For now, return a placeholder
	return struct{}{}
}

// GiteaClient returns a Gitea-specific client (placeholder).
func (c *Client) GiteaClient(baseURL, token string) interface{} {
	// Gitea client would be implemented here
	// For now, return a placeholder
	return struct{}{}
}

// ListPlugins returns information about loaded plugins - DISABLED (plugins package removed)
// func (c *Client) ListPlugins() ([]PluginInfo, error) {
//	if c.pluginManager == nil {
//		return nil, fmt.Errorf("plugin manager not initialized")
//	}
//
//	plugins := c.pluginManager.ListPlugins()
//	result := make([]PluginInfo, len(plugins))
//
//	for i, plugin := range plugins {
//		result[i] = PluginInfo{
//			Name:         plugin.Name,
//			Version:      plugin.Version,
//			Description:  plugin.Description,
//			Author:       plugin.Author,
//			Status:       plugin.Status,
//			Capabilities: plugin.Capabilities,
//			LoadTime:     plugin.LoadTime,
//			LastUsed:     plugin.LastUsed,
//			CallCount:    plugin.CallCount,
//			ErrorCount:   plugin.ErrorCount,
//		}
//	}
//
//	return result, nil
// }

// ExecutePlugin executes a plugin method with given arguments - DISABLED (plugins package removed)
// func (c *Client) ExecutePlugin(ctx context.Context, req PluginExecuteRequest) (*PluginExecuteResult, error) {
//	if c.pluginManager == nil {
//		return nil, fmt.Errorf("plugin manager not initialized")
//	}
//
//	// Set default timeout if not specified
//	timeout := req.Timeout
//	if timeout == 0 {
//		timeout = c.config.Timeout
//	}
//
//	// Create execution context with timeout
//	execCtx, cancel := context.WithTimeout(ctx, timeout)
//	defer cancel()
//
//	// Execute plugin
//	startTime := time.Now()
//	result, err := c.pluginManager.ExecutePlugin(execCtx, req.PluginName, req.Method, req.Args)
//	duration := time.Since(startTime)
//
//	response := &PluginExecuteResult{
//		PluginName: req.PluginName,
//		Method:     req.Method,
//		Result:     result,
//		Duration:   duration,
//		Timestamp:  startTime,
//	}
//
//	if err != nil {
//		response.Error = err.Error()
//	}
//
//	return response, nil
// }

// GetSystemMetrics returns current system metrics.
func (c *Client) GetSystemMetrics() (*SystemMetrics, error) {
	// This would integrate with system monitoring packages
	// For now, return a basic implementation
	return &SystemMetrics{
		Timestamp: time.Now(),
		CPU: CPUMetrics{
			Cores: 4, // This should be detected
		},
		Memory: MemoryMetrics{
			// This should be populated with actual system data
		},
		Disk: DiskMetrics{
			// This should be populated with actual disk data
		},
		Uptime: time.Hour * 24, // This should be actual uptime
	}, nil
}

// Subscribe creates an event subscription.
func (c *Client) Subscribe(subscription EventSubscription) error {
	// This would integrate with event system
	// Implementation depends on event bus architecture
	return fmt.Errorf("event subscription not yet implemented")
}

// Unsubscribe removes an event subscription.
func (c *Client) Unsubscribe(subscriptionID string) error {
	// This would integrate with event system
	return fmt.Errorf("event unsubscription not yet implemented")
}

// configManagerImpl implements bulkclone.ConfigurationManager interface.
type configManagerImpl struct{}

func (c *configManagerImpl) LoadConfiguration(ctx context.Context) (*bulkclone.BulkCloneConfig, error) {
	return bulkclone.LoadConfig("")
}

func (c *configManagerImpl) ValidateConfiguration(ctx context.Context, config *bulkclone.BulkCloneConfig) error {
	// Validation logic would go here
	return nil
}

// loggerImpl implements bulkclone.Logger interface.
type loggerImpl struct{}

func (l *loggerImpl) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

func (l *loggerImpl) Info(msg string, args ...interface{}) {
	// Format the message properly with key-value pairs
	formatted := msg

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			formatted += fmt.Sprintf(" %s=%v", args[i], args[i+1])
		}
	}

	fmt.Printf("[INFO] %s\n", formatted)
}

func (l *loggerImpl) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] "+msg+"\n", args...)
}

func (l *loggerImpl) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}

// silentLoggerImpl implements bulkclone.Logger interface but doesn't output anything.
type silentLoggerImpl struct{}

func (l *silentLoggerImpl) Debug(msg string, args ...interface{}) {}
func (l *silentLoggerImpl) Info(msg string, args ...interface{})  {}
func (l *silentLoggerImpl) Warn(msg string, args ...interface{})  {}
func (l *silentLoggerImpl) Error(msg string, args ...interface{}) {}

// httpClientWrapper wraps http.Client to implement github.HTTPClientInterface.
type httpClientWrapper struct {
	client *http.Client
}

func (h *httpClientWrapper) Do(req *http.Request) (*http.Response, error) {
	return h.client.Do(req)
}

func (h *httpClientWrapper) Get(url string) (*http.Response, error) {
	return h.client.Get(url)
}

func (h *httpClientWrapper) Post(url, contentType string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte

	var err error

	switch v := body.(type) {
	case []byte:
		bodyBytes = v
	case string:
		bodyBytes = []byte(v)
	case nil:
		bodyBytes = nil
	default:
		bodyBytes, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}

		contentType = "application/json"
	}

	return h.client.Post(url, contentType, bytes.NewReader(bodyBytes))
}
