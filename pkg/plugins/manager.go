package plugins

import (
	"context"
	"fmt"
	"path/filepath"
	"plugin"
	"sync"
	"time"
)

// Manager handles plugin lifecycle and execution
type Manager struct {
	plugins     map[string]*PluginInstance
	pluginAPI   PluginAPI
	config      ManagerConfig
	eventBus    *EventBus
	securityMgr *SecurityManager
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// PluginInstance wraps a loaded plugin with runtime information
type PluginInstance struct {
	Plugin     Plugin
	Handle     *plugin.Plugin
	Metadata   PluginMetadata
	Config     PluginConfig
	LoadTime   time.Time
	LastUsed   time.Time
	CallCount  int64
	ErrorCount int64
	Status     PluginStatus
	mu         sync.RWMutex
}

// ManagerConfig configures the plugin manager
type ManagerConfig struct {
	PluginDir           string
	ConfigFile          string
	EnableSandbox       bool
	DefaultLimits       ResourceLimits
	LoadTimeout         time.Duration
	ExecuteTimeout      time.Duration
	HealthCheckInterval time.Duration
}

// NewManager creates a new plugin manager
func NewManager(config ManagerConfig, api PluginAPI) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		plugins:     make(map[string]*PluginInstance),
		pluginAPI:   api,
		config:      config,
		eventBus:    NewEventBus(),
		securityMgr: NewSecurityManager(config.EnableSandbox),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// LoadPlugin loads a plugin from a file
func (m *Manager) LoadPlugin(pluginPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load plugin file
	handle, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", pluginPath, err)
	}

	// Look for the required NewPlugin function
	sym, err := handle.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin %s missing NewPlugin function: %w", pluginPath, err)
	}

	// Cast to the expected function signature
	newPluginFunc, ok := sym.(func() Plugin)
	if !ok {
		return fmt.Errorf("plugin %s NewPlugin function has wrong signature", pluginPath)
	}

	// Create plugin instance
	pluginObj := newPluginFunc()
	metadata := pluginObj.GetMetadata()

	// Validate plugin
	if err := m.validatePlugin(metadata); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	// Check for conflicts
	if existing, exists := m.plugins[metadata.Name]; exists {
		return fmt.Errorf("plugin %s already loaded (version %s)",
			metadata.Name, existing.Metadata.Version)
	}

	// Create plugin instance
	instance := &PluginInstance{
		Plugin:   pluginObj,
		Handle:   handle,
		Metadata: metadata,
		LoadTime: time.Now(),
		Status:   PluginStatusLoading,
	}

	// Set up plugin configuration
	config, err := m.preparePluginConfig(metadata)
	if err != nil {
		return fmt.Errorf("failed to prepare config for %s: %w", metadata.Name, err)
	}
	instance.Config = config

	// Initialize plugin with timeout
	ctx, cancel := context.WithTimeout(m.ctx, m.config.LoadTimeout)
	defer cancel()

	if err := pluginObj.Initialize(ctx, config); err != nil {
		instance.Status = PluginStatusError
		return fmt.Errorf("failed to initialize plugin %s: %w", metadata.Name, err)
	}

	// Register plugin
	instance.Status = PluginStatusReady
	m.plugins[metadata.Name] = instance

	// Emit loaded event
	m.eventBus.Emit(Event{
		Type:      "plugin.loaded",
		Source:    "plugin_manager",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plugin_name": metadata.Name,
			"version":     metadata.Version,
		},
	})

	return nil
}

// UnloadPlugin removes a plugin from memory
func (m *Manager) UnloadPlugin(pluginName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	instance.Status = PluginStatusUnloading

	// Cleanup plugin
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	if err := instance.Plugin.Cleanup(ctx); err != nil {
		// Log error but continue unloading
		fmt.Printf("Warning: plugin %s cleanup failed: %v\n", pluginName, err)
	}

	// Remove from registry
	delete(m.plugins, pluginName)

	// Emit unloaded event
	m.eventBus.Emit(Event{
		Type:      "plugin.unloaded",
		Source:    "plugin_manager",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plugin_name": pluginName,
		},
	})

	return nil
}

// ExecutePlugin runs a plugin with the given arguments
func (m *Manager) ExecutePlugin(pluginName string, args map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	instance, exists := m.plugins[pluginName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.Status != PluginStatusReady {
		return nil, fmt.Errorf("plugin %s not ready (status: %s)", pluginName, instance.Status)
	}

	// Check resource limits and permissions
	if err := m.securityMgr.CheckExecution(instance, args); err != nil {
		instance.ErrorCount++
		return nil, fmt.Errorf("security check failed: %w", err)
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(m.ctx, m.config.ExecuteTimeout)
	defer cancel()

	instance.Status = PluginStatusRunning
	instance.CallCount++
	instance.LastUsed = time.Now()

	result, err := instance.Plugin.Execute(ctx, args)
	if err != nil {
		instance.ErrorCount++
		instance.Status = PluginStatusError

		// Emit error event
		m.eventBus.Emit(Event{
			Type:      "plugin.error",
			Source:    pluginName,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"error": err.Error(),
				"args":  args,
			},
		})

		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	instance.Status = PluginStatusReady

	// Emit execution event
	m.eventBus.Emit(Event{
		Type:      "plugin.executed",
		Source:    pluginName,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"call_count":  instance.CallCount,
			"result_type": fmt.Sprintf("%T", result),
		},
	})

	return result, nil
}

// ListPlugins returns information about loaded plugins
func (m *Manager) ListPlugins() []PluginMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]PluginMetadata, 0, len(m.plugins))
	for _, instance := range m.plugins {
		instance.mu.RLock()
		metadata := instance.Metadata
		metadata.Status = instance.Status
		metadata.LoadTime = instance.LoadTime
		metadata.LastUsed = instance.LastUsed
		instance.mu.RUnlock()
		plugins = append(plugins, metadata)
	}

	return plugins
}

// GetPlugin returns a specific plugin instance
func (m *Manager) GetPlugin(pluginName string) (*PluginInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return instance, nil
}

// LoadPluginsFromDirectory scans and loads all plugins from a directory
func (m *Manager) LoadPluginsFromDirectory(dir string) error {
	pattern := filepath.Join(dir, "*.so")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan plugin directory %s: %w", dir, err)
	}

	var errors []error
	for _, match := range matches {
		if err := m.LoadPlugin(match); err != nil {
			errors = append(errors, fmt.Errorf("failed to load %s: %w", match, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to load %d plugins: %v", len(errors), errors)
	}

	return nil
}

// HealthCheck runs health checks on all plugins
func (m *Manager) HealthCheck() map[string]error {
	m.mu.RLock()
	plugins := make([]*PluginInstance, 0, len(m.plugins))
	for _, instance := range m.plugins {
		plugins = append(plugins, instance)
	}
	m.mu.RUnlock()

	results := make(map[string]error)
	for _, instance := range plugins {
		instance.mu.RLock()
		name := instance.Metadata.Name
		status := instance.Status
		instance.mu.RUnlock()

		if status != PluginStatusReady {
			results[name] = fmt.Errorf("plugin not ready (status: %s)", status)
			continue
		}

		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		err := instance.Plugin.HealthCheck(ctx)
		cancel()

		if err != nil {
			results[name] = err
			instance.mu.Lock()
			instance.Status = PluginStatusError
			instance.ErrorCount++
			instance.mu.Unlock()
		}
	}

	return results
}

// Shutdown gracefully shuts down the plugin manager
func (m *Manager) Shutdown() error {
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error
	for name := range m.plugins {
		if err := m.UnloadPlugin(name); err != nil {
			errors = append(errors, fmt.Errorf("failed to unload %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// validatePlugin validates plugin metadata and requirements
func (m *Manager) validatePlugin(metadata PluginMetadata) error {
	if metadata.Name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if metadata.Version == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	// Additional validation can be added here
	return nil
}

// preparePluginConfig creates configuration for a plugin
func (m *Manager) preparePluginConfig(metadata PluginMetadata) (PluginConfig, error) {
	config := PluginConfig{
		Settings:    make(map[string]interface{}),
		Environment: make(map[string]string),
		Permissions: PermissionSet{
			FileSystem: FileSystemPermissions{
				ReadPaths:  []string{},
				WritePaths: []string{},
			},
			Network: NetworkPermissions{
				AllowedHosts: []string{},
				MaxRequests:  100,
			},
		},
		Limits: m.config.DefaultLimits,
	}

	// Load plugin-specific configuration from file
	// This would typically load from a config file

	return config, nil
}
