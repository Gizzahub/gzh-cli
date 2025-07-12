/*
Package plugins provides a comprehensive plugin system for the GZH Manager CLI tool.

The plugin system enables extending GZH Manager functionality through dynamically loaded
Go plugins (.so files) that implement the Plugin interface. It provides secure execution,
resource management, and inter-plugin communication capabilities.

# Architecture Overview

The plugin system consists of several key components:

1. Plugin Interface: Defines the contract that all plugins must implement
2. Plugin Manager: Handles plugin lifecycle (load, execute, unload)
3. Security Manager: Enforces sandboxing and resource limits
4. Event Bus: Enables communication between plugins and the host
5. Plugin API: Provides services to plugins from the host application

# Basic Usage

To create a plugin, implement the Plugin interface:

	type MyPlugin struct{}

	func NewMyPlugin() plugins.Plugin {
	    return &MyPlugin{}
	}

	func (p *MyPlugin) GetMetadata() plugins.PluginMetadata {
	    return plugins.PluginMetadata{
	        Name:        "my-plugin",
	        Version:     "1.0.0",
	        Description: "My custom plugin",
	        // ... other metadata
	    }
	}

	func (p *MyPlugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	    // Initialize plugin
	    return nil
	}

	func (p *MyPlugin) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	    // Perform plugin operation
	    return "Hello from plugin!", nil
	}

	func (p *MyPlugin) Cleanup(ctx context.Context) error {
	    // Clean up resources
	    return nil
	}

	func (p *MyPlugin) HealthCheck(ctx context.Context) error {
	    // Verify plugin health
	    return nil
	}

# Plugin Manager Usage

Load and execute plugins using the Plugin Manager:

	// Create plugin manager
	config := plugins.ManagerConfig{
	    PluginDir:      "/path/to/plugins",
	    EnableSandbox:  true,
	    LoadTimeout:    30 * time.Second,
	    ExecuteTimeout: 60 * time.Second,
	}

	api := plugins.NewDefaultPluginAPI(eventBus, securityMgr, hostInfo)
	manager := plugins.NewManager(config, api)

	// Load plugin
	err := manager.LoadPlugin("/path/to/plugin.so")
	if err != nil {
	    log.Fatal(err)
	}

	// Execute plugin
	result, err := manager.ExecutePlugin("my-plugin", map[string]interface{}{
	    "input": "test data",
	})

# Security Features

The plugin system provides comprehensive security features:

- Resource limits (memory, CPU, execution time)
- File system access controls
- Network access restrictions
- Environment variable access controls
- Sandboxed execution environments

Configure security through PluginConfig:

	config := plugins.PluginConfig{
	    Permissions: plugins.PermissionSet{
	        FileSystem: plugins.FileSystemPermissions{
	            ReadPaths:  []string{"/tmp", "/var/data"},
	            WritePaths: []string{"/tmp"},
	            DenyPaths:  []string{"/etc", "/proc"},
	        },
	        Network: plugins.NetworkPermissions{
	            AllowedHosts: []string{"api.example.com"},
	            BlockedHosts: []string{"internal.company.com"},
	            MaxRequests:  100,
	        },
	    },
	    Limits: plugins.ResourceLimits{
	        MaxMemoryMB:      256,
	        MaxCPUPercent:    25.0,
	        MaxExecutionTime: 5 * time.Minute,
	    },
	}

# Plugin API Services

Plugins can access host services through the PluginAPI interface:

	// File operations
	data, err := api.ReadFile("/path/to/file")
	err = api.WriteFile("/path/to/output", data)

	// HTTP requests
	resp, err := api.HTTPRequest("GET", "https://api.example.com", headers, body)

	// Configuration
	value, err := api.GetConfig("setting_name")
	err = api.SetConfig("setting_name", "new_value")

	// Events
	err = api.EmitEvent(plugins.Event{
	    Type: "custom.event",
	    Data: map[string]interface{}{"key": "value"},
	})

	err = api.SubscribeToEvent("system.status", handleSystemStatus)

	// Inter-plugin communication
	result, err := api.CallPlugin("other-plugin", "method_name", args)

	// Logging
	logger := api.GetLogger("my-plugin")
	logger.Info("Plugin operation completed")

# Event System

The event bus enables loose coupling between plugins and the host:

	// Subscribe to events
	api.SubscribeToEvent("file.changed", func(event plugins.Event) error {
	    filePath := event.Data["path"].(string)
	    log.Printf("File changed: %s", filePath)
	    return nil
	})

	// Emit custom events
	api.EmitEvent(plugins.Event{
	    Type:      "plugin.task_completed",
	    Source:    "my-plugin",
	    Timestamp: time.Now(),
	    Data: map[string]interface{}{
	        "task_id": "task-123",
	        "result":  "success",
	    },
	})

# Plugin Metadata

Rich metadata enables plugin discovery and validation:

	metadata := plugins.PluginMetadata{
	    Name:        "backup-plugin",
	    Version:     "2.1.0",
	    Description: "Automated backup solution",
	    Author:      "Company Team",
	    Homepage:    "https://github.com/company/backup-plugin",
	    License:     "MIT",
	    Tags:        []string{"backup", "storage", "automation"},

	    Capabilities: []string{
	        "file_operations",
	        "cloud_storage",
	        "encryption",
	    },

	    Requirements: plugins.PluginRequirements{
	        MinGZVersion:    "1.2.0",
	        Dependencies:    []string{"encryption-lib", "cloud-sdk"},
	        Permissions:     []string{"file.read", "file.write", "network.https"},
	        SupportedOS:     []string{"linux", "darwin"},
	        RequiredEnvVars: []string{"CLOUD_API_KEY"},
	    },

	    ConfigSchema: map[string]interface{}{
	        "type": "object",
	        "properties": map[string]interface{}{
	            "backup_path": map[string]interface{}{
	                "type":        "string",
	                "description": "Path to backup directory",
	                "required":    true,
	            },
	        },
	    },
	}

# Plugin Development Best Practices

1. Always implement proper error handling and logging
2. Validate input parameters in Execute method
3. Use context cancellation for long-running operations
4. Clean up resources in the Cleanup method
5. Implement meaningful health checks
6. Follow semantic versioning for plugin versions
7. Document plugin capabilities and configuration options
8. Test plugins thoroughly before deployment

# Compilation

To compile a plugin as a shared object:

	go build -buildmode=plugin -o myplugin.so myplugin.go

The plugin must export a NewPlugin function that returns a Plugin interface implementation.

# Integration with GZH Manager

Plugins integrate with GZH Manager through command registration:

	// In your plugin's Execute method
	if method == "register_command" {
	    return map[string]interface{}{
	        "commands": []map[string]interface{}{
	            {
	                "name":        "backup",
	                "description": "Create system backup",
	                "flags": []map[string]interface{}{
	                    {"name": "path", "type": "string", "required": true},
	                    {"name": "compress", "type": "bool", "default": false},
	                },
	            },
	        },
	    }, nil
	}

This package provides a robust foundation for building extensible CLI applications
with secure plugin capabilities.
*/
package plugins
