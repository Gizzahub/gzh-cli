package plugins

import (
	"context"
	"fmt"
	"time"
)

// ExamplePlugin demonstrates how to implement the Plugin interface
type ExamplePlugin struct {
	api    PluginAPI
	logger Logger
	config PluginConfig
}

// NewExamplePlugin creates a new example plugin instance
// This is the function that will be looked up in plugin .so files
func NewExamplePlugin() Plugin {
	return &ExamplePlugin{}
}

// GetMetadata returns plugin metadata
func (p *ExamplePlugin) GetMetadata() PluginMetadata {
	return PluginMetadata{
		Name:        "example-plugin",
		Version:     "1.0.0",
		Description: "An example plugin demonstrating the plugin API",
		Author:      "GZH Manager Team",
		License:     "MIT",
		Tags:        []string{"example", "demo"},
		Capabilities: []string{
			"file_operations",
			"network_requests",
			"configuration",
		},
		Requirements: PluginRequirements{
			MinGZVersion: "1.0.0",
			Permissions: []string{
				"file.read",
				"file.write",
				"network.http",
			},
			SupportedOS: []string{"linux", "darwin", "windows"},
		},
		ConfigSchema: map[string]interface{}{
			"properties": map[string]interface{}{
				"greeting": map[string]interface{}{
					"type":        "string",
					"description": "Custom greeting message",
					"default":     "Hello",
				},
				"max_items": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum items to process",
					"default":     10,
				},
			},
		},
	}
}

// Initialize sets up the plugin
func (p *ExamplePlugin) Initialize(ctx context.Context, config PluginConfig) error {
	p.config = config

	// Get API from context (this would be injected by the plugin manager)
	if api, ok := ctx.Value("plugin_api").(PluginAPI); ok {
		p.api = api
		p.logger = api.GetLogger("example-plugin")
	} else {
		return fmt.Errorf("plugin API not available in context")
	}

	p.logger.Info("Example plugin initialized", map[string]interface{}{
		"config_keys": len(config.Settings),
	})

	// Subscribe to system events
	if p.api != nil {
		p.api.SubscribeToEvent("system.status", p.handleSystemStatus)
		p.api.SubscribeToEvent("task.completed", p.handleTaskCompleted)
	}

	return nil
}

// Execute performs the plugin's main operation
func (p *ExamplePlugin) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	p.logger.Info("Plugin execution started", map[string]interface{}{
		"args_count": len(args),
	})

	// Handle different methods/operations
	method, ok := args["__method"].(string)
	if !ok {
		method = "default"
	}

	switch method {
	case "greet":
		return p.greet(args)
	case "process_file":
		return p.processFile(args)
	case "fetch_data":
		return p.fetchData(args)
	case "get_status":
		return p.getStatus()
	default:
		return p.defaultOperation(args)
	}
}

// greet demonstrates basic string processing
func (p *ExamplePlugin) greet(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		name = "World"
	}

	greeting := "Hello"
	if customGreeting, exists := p.config.Settings["greeting"]; exists {
		if greetingStr, ok := customGreeting.(string); ok {
			greeting = greetingStr
		}
	}

	result := fmt.Sprintf("%s, %s!", greeting, name)

	p.logger.Info("Generated greeting", map[string]interface{}{
		"name":     name,
		"greeting": greeting,
		"result":   result,
	})

	return map[string]interface{}{
		"message":   result,
		"timestamp": time.Now(),
	}, nil
}

// processFile demonstrates file operations
func (p *ExamplePlugin) processFile(args map[string]interface{}) (interface{}, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path argument required")
	}

	// Read file using plugin API
	data, err := p.api.ReadFile(filePath)
	if err != nil {
		p.logger.Error("Failed to read file", err, map[string]interface{}{
			"file_path": filePath,
		})
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Process the file content (example: count lines)
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}

	result := map[string]interface{}{
		"file_path":    filePath,
		"size_bytes":   len(data),
		"line_count":   lines,
		"processed_at": time.Now(),
	}

	p.logger.Info("File processed", map[string]interface{}{
		"file_path": filePath,
		"size":      len(data),
		"lines":     lines,
	})

	return result, nil
}

// fetchData demonstrates HTTP requests
func (p *ExamplePlugin) fetchData(args map[string]interface{}) (interface{}, error) {
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url argument required")
	}

	// Make HTTP request using plugin API
	resp, err := p.api.HTTPRequest("GET", url, nil, nil)
	if err != nil {
		p.logger.Error("HTTP request failed", err, map[string]interface{}{
			"url": url,
		})
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	result := map[string]interface{}{
		"url":         url,
		"status_code": resp.StatusCode,
		"body_size":   len(resp.Body),
		"headers":     resp.Headers,
		"fetched_at":  time.Now(),
	}

	// Include body if small enough
	if len(resp.Body) < 1024 {
		result["body"] = string(resp.Body)
	}

	p.logger.Info("Data fetched", map[string]interface{}{
		"url":         url,
		"status_code": resp.StatusCode,
		"body_size":   len(resp.Body),
	})

	return result, nil
}

// getStatus returns plugin status information
func (p *ExamplePlugin) getStatus() (interface{}, error) {
	hostInfo := p.api.GetHostInfo()

	return map[string]interface{}{
		"plugin": map[string]interface{}{
			"name":    "example-plugin",
			"version": "1.0.0",
			"status":  "running",
		},
		"host": map[string]interface{}{
			"gz_version":   hostInfo.GZVersion,
			"os":           hostInfo.OS,
			"architecture": hostInfo.Architecture,
		},
		"config": map[string]interface{}{
			"settings_count": len(p.config.Settings),
		},
		"timestamp": time.Now(),
	}, nil
}

// defaultOperation is the fallback operation
func (p *ExamplePlugin) defaultOperation(args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"message":   "Example plugin executed successfully",
		"args":      args,
		"timestamp": time.Now(),
	}, nil
}

// Cleanup performs cleanup operations
func (p *ExamplePlugin) Cleanup(ctx context.Context) error {
	p.logger.Info("Plugin cleanup started")

	// Perform any necessary cleanup
	// - Close files
	// - Release resources
	// - Save state

	p.logger.Info("Plugin cleanup completed")
	return nil
}

// HealthCheck verifies plugin health
func (p *ExamplePlugin) HealthCheck(ctx context.Context) error {
	// Perform health checks
	// - Verify dependencies
	// - Check resource usage
	// - Validate configuration

	if p.api == nil {
		return fmt.Errorf("plugin API not available")
	}

	if p.logger == nil {
		return fmt.Errorf("logger not available")
	}

	p.logger.Debug("Health check passed")
	return nil
}

// Event handlers

// handleSystemStatus handles system status events
func (p *ExamplePlugin) handleSystemStatus(event Event) error {
	p.logger.Info("Received system status event", map[string]interface{}{
		"event_data": event.Data,
	})

	// React to system status changes
	if status, ok := event.Data["status"].(string); ok {
		switch status {
		case "high_load":
			p.logger.Warn("System under high load, reducing activity")
		case "low_memory":
			p.logger.Warn("System low on memory, performing cleanup")
		}
	}

	return nil
}

// handleTaskCompleted handles task completion events
func (p *ExamplePlugin) handleTaskCompleted(event Event) error {
	if taskName, ok := event.Data["task_name"].(string); ok {
		p.logger.Info("Task completed", map[string]interface{}{
			"task_name": taskName,
		})

		// Emit our own event
		p.api.EmitEvent(Event{
			Type:      "plugin.task_acknowledged",
			Source:    "example-plugin",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"acknowledged_task": taskName,
				"plugin_name":       "example-plugin",
			},
		})
	}

	return nil
}
