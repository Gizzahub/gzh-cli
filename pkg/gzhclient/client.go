package gzhclient

import (
	"context"
	"fmt"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/bulk-clone"
	"github.com/gizzahub/gzh-manager-go/pkg/gitea"
	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/gizzahub/gzh-manager-go/pkg/gitlab"
	"github.com/gizzahub/gzh-manager-go/pkg/plugins"
)

// Client provides programmatic access to GZH Manager functionality
type Client struct {
	config        ClientConfig
	pluginManager *plugins.Manager
	ctx           context.Context
	cancel        context.CancelFunc
}

// ClientConfig holds configuration for the GZH client
type ClientConfig struct {
	// Connection settings
	ServerURL  string        `yaml:"server_url,omitempty" json:"server_url,omitempty"`
	APIKey     string        `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
	RetryCount int           `yaml:"retry_count" json:"retry_count"`

	// Plugin settings
	PluginDir     string `yaml:"plugin_dir,omitempty" json:"plugin_dir,omitempty"`
	EnablePlugins bool   `yaml:"enable_plugins" json:"enable_plugins"`

	// Logging settings
	LogLevel string `yaml:"log_level" json:"log_level"`
	LogFile  string `yaml:"log_file,omitempty" json:"log_file,omitempty"`

	// Feature flags
	Features FeatureFlags `yaml:"features" json:"features"`
}

// FeatureFlags enables/disables specific features
type FeatureFlags struct {
	BulkClone  bool `yaml:"bulk_clone" json:"bulk_clone"`
	DevEnv     bool `yaml:"dev_env" json:"dev_env"`
	NetEnv     bool `yaml:"net_env" json:"net_env"`
	Monitoring bool `yaml:"monitoring" json:"monitoring"`
	Plugins    bool `yaml:"plugins" json:"plugins"`
}

// DefaultConfig returns a default client configuration
func DefaultConfig() ClientConfig {
	return ClientConfig{
		Timeout:       30 * time.Second,
		RetryCount:    3,
		EnablePlugins: true,
		LogLevel:      "info",
		Features: FeatureFlags{
			BulkClone:  true,
			DevEnv:     true,
			NetEnv:     true,
			Monitoring: true,
			Plugins:    true,
		},
	}
}

// NewClient creates a new GZH Manager client
func NewClient(config ClientConfig) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize plugin manager if enabled
	if config.EnablePlugins && config.Features.Plugins {
		if err := client.initializePluginManager(); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize plugin manager: %w", err)
		}
	}

	return client, nil
}

// Close cleanly shuts down the client
func (c *Client) Close() error {
	if c.pluginManager != nil {
		if err := c.pluginManager.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown plugin manager: %w", err)
		}
	}

	c.cancel()
	return nil
}

// GetConfig returns the current client configuration
func (c *Client) GetConfig() ClientConfig {
	return c.config
}

// UpdateConfig updates the client configuration
func (c *Client) UpdateConfig(config ClientConfig) error {
	c.config = config

	// Reinitialize plugin manager if settings changed
	if config.EnablePlugins && config.Features.Plugins && c.pluginManager == nil {
		return c.initializePluginManager()
	} else if (!config.EnablePlugins || !config.Features.Plugins) && c.pluginManager != nil {
		if err := c.pluginManager.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown plugin manager: %w", err)
		}
		c.pluginManager = nil
	}

	return nil
}

// Health checks the health of the client and its components
func (c *Client) Health() HealthStatus {
	status := HealthStatus{
		Overall:    StatusHealthy,
		Components: make(map[string]ComponentHealth),
		Timestamp:  time.Now(),
	}

	// Check plugin manager health
	if c.pluginManager != nil {
		pluginErrors := c.pluginManager.HealthCheck()
		if len(pluginErrors) > 0 {
			status.Components["plugins"] = ComponentHealth{
				Status:  StatusUnhealthy,
				Message: fmt.Sprintf("%d plugin(s) unhealthy", len(pluginErrors)),
				Details: pluginErrors,
			}
			status.Overall = StatusDegraded
		} else {
			status.Components["plugins"] = ComponentHealth{
				Status:  StatusHealthy,
				Message: "All plugins healthy",
			}
		}
	}

	// Add more component checks here
	status.Components["client"] = ComponentHealth{
		Status:  StatusHealthy,
		Message: "Client operational",
	}

	return status
}

// initializePluginManager sets up the plugin manager
func (c *Client) initializePluginManager() error {
	if c.config.PluginDir == "" {
		return fmt.Errorf("plugin directory not configured")
	}

	// Create plugin manager configuration
	managerConfig := plugins.ManagerConfig{
		PluginDir:           c.config.PluginDir,
		EnableSandbox:       true,
		LoadTimeout:         30 * time.Second,
		ExecuteTimeout:      c.config.Timeout,
		HealthCheckInterval: 60 * time.Second,
		DefaultLimits: plugins.ResourceLimits{
			MaxMemoryMB:      256,
			MaxCPUPercent:    25.0,
			MaxExecutionTime: 5 * time.Minute,
			MaxFileHandles:   100,
		},
	}

	// Create plugin API
	eventBus := plugins.NewEventBus()
	securityMgr := plugins.NewSecurityManager(true)
	hostInfo := plugins.HostInfo{
		GZVersion:    "1.0.0",   // This should come from build info
		OS:           "unknown", // This should be detected
		Architecture: "unknown", // This should be detected
		WorkingDir:   ".",
		ConfigDir:    "~/.config/gzh-manager",
		PluginDir:    c.config.PluginDir,
	}

	api := plugins.NewDefaultPluginAPI(eventBus, securityMgr, hostInfo)

	// Create plugin manager
	manager := plugins.NewManager(managerConfig, api)
	api.SetManager(manager)

	// Load plugins from directory
	if err := manager.LoadPluginsFromDirectory(c.config.PluginDir); err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	c.pluginManager = manager
	return nil
}

// BulkClone performs bulk repository cloning operation
func (c *Client) BulkClone(ctx context.Context, req BulkCloneRequest) (*BulkCloneResult, error) {
	// Create bulk clone configuration
	config := bulkclone.BulkCloneConfig{
		OutputDir:      req.OutputDir,
		Concurrency:    req.Concurrency,
		Strategy:       req.Strategy,
		IncludePrivate: req.IncludePrivate,
		Platforms:      make(map[string]bulkclone.PlatformConfig),
	}

	// Convert platform configurations
	for _, platform := range req.Platforms {
		platformConfig := bulkclone.PlatformConfig{
			Type:          platform.Type,
			URL:           platform.URL,
			Token:         platform.Token,
			Organizations: platform.Organizations,
			Users:         platform.Users,
		}
		config.Platforms[platform.Type] = platformConfig
	}

	// Create bulk clone facade
	facade := bulkclone.NewFacade(config)

	// Execute bulk clone operation
	result, err := facade.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("bulk clone operation failed: %w", err)
	}

	// Convert result to client response format
	response := &BulkCloneResult{
		TotalRepos:   result.TotalRepos,
		SuccessCount: result.SuccessCount,
		FailureCount: result.FailureCount,
		SkippedCount: result.SkippedCount,
		Duration:     result.Duration,
		Results:      make([]RepositoryCloneResult, len(result.Results)),
		Summary:      result.Summary,
	}

	for i, repo := range result.Results {
		response.Results[i] = RepositoryCloneResult{
			RepoName:  repo.Name,
			Platform:  repo.Platform,
			URL:       repo.URL,
			LocalPath: repo.LocalPath,
			Status:    repo.Status,
			Error:     repo.Error,
			Duration:  repo.Duration,
			Size:      repo.Size,
		}
	}

	return response, nil
}

// GitHubClient returns a GitHub-specific client
func (c *Client) GitHubClient(token string) *github.Client {
	return github.NewClient(token)
}

// GitLabClient returns a GitLab-specific client
func (c *Client) GitLabClient(baseURL, token string) *gitlab.Client {
	return gitlab.NewClient(baseURL, token)
}

// GiteaClient returns a Gitea-specific client
func (c *Client) GiteaClient(baseURL, token string) *gitea.Client {
	return gitea.NewClient(baseURL, token)
}

// ListPlugins returns information about loaded plugins
func (c *Client) ListPlugins() ([]PluginInfo, error) {
	if c.pluginManager == nil {
		return nil, fmt.Errorf("plugin manager not initialized")
	}

	plugins := c.pluginManager.ListPlugins()
	result := make([]PluginInfo, len(plugins))

	for i, plugin := range plugins {
		result[i] = PluginInfo{
			Name:         plugin.Name,
			Version:      plugin.Version,
			Description:  plugin.Description,
			Author:       plugin.Author,
			Status:       plugin.Status,
			Capabilities: plugin.Capabilities,
			LoadTime:     plugin.LoadTime,
			LastUsed:     plugin.LastUsed,
			CallCount:    plugin.CallCount,
			ErrorCount:   plugin.ErrorCount,
		}
	}

	return result, nil
}

// ExecutePlugin executes a plugin method with given arguments
func (c *Client) ExecutePlugin(ctx context.Context, req PluginExecuteRequest) (*PluginExecuteResult, error) {
	if c.pluginManager == nil {
		return nil, fmt.Errorf("plugin manager not initialized")
	}

	// Set default timeout if not specified
	timeout := req.Timeout
	if timeout == 0 {
		timeout = c.config.Timeout
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute plugin
	startTime := time.Now()
	result, err := c.pluginManager.ExecutePlugin(execCtx, req.PluginName, req.Method, req.Args)
	duration := time.Since(startTime)

	response := &PluginExecuteResult{
		PluginName: req.PluginName,
		Method:     req.Method,
		Result:     result,
		Duration:   duration,
		Timestamp:  startTime,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response, nil
}

// GetSystemMetrics returns current system metrics
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

// Subscribe creates an event subscription
func (c *Client) Subscribe(subscription EventSubscription) error {
	// This would integrate with event system
	// Implementation depends on event bus architecture
	return fmt.Errorf("event subscription not yet implemented")
}

// Unsubscribe removes an event subscription
func (c *Client) Unsubscribe(subscriptionID string) error {
	// This would integrate with event system
	return fmt.Errorf("event unsubscription not yet implemented")
}
