package plugin

// Template definitions for plugin generation

const goModTemplate = `module {{.ModuleName}}

go {{.GoVersion}}

require (
	github.com/gzh-manager/gzh-manager-go v0.0.0-latest
)

replace github.com/gzh-manager/gzh-manager-go => ../../
`

const mainGoTemplate = `package main

import (
	"{{.ModuleName}}/pkg/{{.PluginName}}"
	gzhplugins "github.com/gzh-manager/gzh-manager-go/pkg/plugins"
)

// NewPlugin is the entry point for the plugin system
// This function must be exported for the plugin to be loadable
func NewPlugin() gzhplugins.Plugin {
	return {{.PluginName}}.New{{.PluginNameGo}}Plugin()
}
`

const pluginGoTemplate = `package {{.PluginName}}

import (
	"context"
	"fmt"
	"time"

	"github.com/gzh-manager/gzh-manager-go/pkg/plugins"
)

// {{.PluginNameGo}}Plugin implements the Plugin interface
type {{.PluginNameGo}}Plugin struct {
	api    plugins.PluginAPI
	logger plugins.Logger
	config plugins.PluginConfig
}

// New{{.PluginNameGo}}Plugin creates a new instance of the plugin
func New{{.PluginNameGo}}Plugin() plugins.Plugin {
	return &{{.PluginNameGo}}Plugin{}
}

// GetMetadata returns plugin metadata
func (p *{{.PluginNameGo}}Plugin) GetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "{{.PluginName}}",
		Version:     "1.0.0",
		Description: "{{.PluginNameTitle}} plugin for GZH Manager",
		Author:      "{{.Author}}",
		License:     "MIT",
		Tags:        []string{"{{.PluginType}}", "utility"},
		Capabilities: []string{
			"basic_operations",
		},
		Requirements: plugins.PluginRequirements{
			MinGZVersion: "1.0.0",
			Permissions: []string{
				"file.read",
				"file.write",
			},
			SupportedOS: []string{"linux", "darwin", "windows"},
		},
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"enabled": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable the plugin",
					"default":     true,
				},
				"log_level": map[string]interface{}{
					"type":        "string",
					"description": "Logging level",
					"enum":        []string{"debug", "info", "warn", "error"},
					"default":     "info",
				},
			},
		},
	}
}

// Initialize sets up the plugin
func (p *{{.PluginNameGo}}Plugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	p.config = config
	
	// Get API from context
	if api, ok := ctx.Value("plugin_api").(plugins.PluginAPI); ok {
		p.api = api
		p.logger = api.GetLogger("{{.PluginName}}")
	} else {
		return fmt.Errorf("plugin API not available in context")
	}
	
	p.logger.Info("{{.PluginNameTitle}} plugin initialized", map[string]interface{}{
		"version": "1.0.0",
		"config_keys": len(config.Settings),
	})
	
	// Subscribe to relevant events
	if p.api != nil {
		p.api.SubscribeToEvent("system.status", p.handleSystemStatus)
	}
	
	return nil
}

// Execute performs the plugin's main operation
func (p *{{.PluginNameGo}}Plugin) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	p.logger.Info("Plugin execution started", map[string]interface{}{
		"args_count": len(args),
	})
	
	// Handle different methods/operations
	method, ok := args["__method"].(string)
	if !ok {
		method = "default"
	}
	
	switch method {
	case "status":
		return p.getStatus()
	case "process":
		return p.processData(args)
	default:
		return p.defaultOperation(args)
	}
}

// getStatus returns plugin status information
func (p *{{.PluginNameGo}}Plugin) getStatus() (interface{}, error) {
	hostInfo := p.api.GetHostInfo()
	
	return map[string]interface{}{
		"plugin": map[string]interface{}{
			"name":    "{{.PluginName}}",
			"version": "1.0.0",
			"status":  "running",
		},
		"host": map[string]interface{}{
			"gz_version":   hostInfo.GZVersion,
			"os":           hostInfo.OS,
			"architecture": hostInfo.Architecture,
		},
		"config": map[string]interface{}{
			"enabled": p.getConfigBool("enabled", true),
			"log_level": p.getConfigString("log_level", "info"),
		},
		"timestamp": time.Now(),
	}, nil
}

// processData processes data based on plugin type
func (p *{{.PluginNameGo}}Plugin) processData(args map[string]interface{}) (interface{}, error) {
	data, ok := args["data"]
	if !ok {
		return nil, fmt.Errorf("data argument required")
	}
	
	p.logger.Info("Processing data", map[string]interface{}{
		"data_type": fmt.Sprintf("%T", data),
	})
	
	// Implement your plugin logic here
	result := map[string]interface{}{
		"input":        data,
		"processed_at": time.Now(),
		"status":       "success",
	}
	
	return result, nil
}

// defaultOperation is the fallback operation
func (p *{{.PluginNameGo}}Plugin) defaultOperation(args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"message":    "{{.PluginNameTitle}} plugin executed successfully",
		"plugin":     "{{.PluginName}}",
		"version":    "1.0.0",
		"args":       args,
		"timestamp":  time.Now(),
	}, nil
}

// Cleanup performs cleanup operations
func (p *{{.PluginNameGo}}Plugin) Cleanup(ctx context.Context) error {
	p.logger.Info("Plugin cleanup started")
	
	// Perform any necessary cleanup
	// - Close connections
	// - Release resources
	// - Save state
	
	p.logger.Info("Plugin cleanup completed")
	return nil
}

// HealthCheck verifies plugin health
func (p *{{.PluginNameGo}}Plugin) HealthCheck(ctx context.Context) error {
	// Perform health checks
	if p.api == nil {
		return fmt.Errorf("plugin API not available")
	}
	
	if p.logger == nil {
		return fmt.Errorf("logger not available")
	}
	
	enabled := p.getConfigBool("enabled", true)
	if !enabled {
		return fmt.Errorf("plugin is disabled")
	}
	
	p.logger.Debug("Health check passed")
	return nil
}

// Event handlers

// handleSystemStatus handles system status events
func (p *{{.PluginNameGo}}Plugin) handleSystemStatus(event plugins.Event) error {
	p.logger.Info("Received system status event", map[string]interface{}{
		"event_type": event.Type,
		"source":     event.Source,
	})
	
	// React to system status changes
	if status, ok := event.Data["status"].(string); ok {
		switch status {
		case "high_load":
			p.logger.Warn("System under high load")
		case "low_memory":
			p.logger.Warn("System low on memory")
		}
	}
	
	return nil
}

// Helper methods

// getConfigString gets a string configuration value with default
func (p *{{.PluginNameGo}}Plugin) getConfigString(key, defaultValue string) string {
	if value, exists := p.config.Settings[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// getConfigBool gets a boolean configuration value with default
func (p *{{.PluginNameGo}}Plugin) getConfigBool(key string, defaultValue bool) bool {
	if value, exists := p.config.Settings[key]; exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

// getConfigInt gets an integer configuration value with default
func (p *{{.PluginNameGo}}Plugin) getConfigInt(key string, defaultValue int) int {
	if value, exists := p.config.Settings[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return defaultValue
}
`

const readmeTemplate = `# {{.PluginNameTitle}} Plugin

{{.PluginNameTitle}} plugin for GZH Manager.

## Description

This plugin provides {{.PluginType}} functionality for GZH Manager.

## Installation

1. Build the plugin:
` + "```" + `bash
   gz plugin build
` + "```" + `

2. Install the plugin:
` + "```" + `bash
   cp {{.PluginName}}.so /path/to/gzh/plugins/
` + "```" + `

## Configuration

The plugin supports the following configuration options:

` + "```" + `yaml
# Plugin configuration
plugins:
  {{.PluginName}}:
    enabled: true
    log_level: info
` + "```" + `

### Configuration Options

- **enabled** (bool): Enable or disable the plugin (default: true)
- **log_level** (string): Logging level - debug, info, warn, error (default: info)

## Usage

### Basic Operation

` + "```" + `bash
gz plugin-execute {{.PluginName}} --method=process --data="example data"
` + "```" + `

### Status Check

` + "```" + `bash
gz plugin-execute {{.PluginName}} --method=status
` + "```" + `

## Development

### Prerequisites

- Go {{.GoVersion}} or later
- GZH Manager development environment

### Building

` + "```" + `bash
make build
` + "```" + `

### Testing

` + "```" + `bash
make test
` + "```" + `

### Development Mode

` + "```" + `bash
make dev
` + "```" + `

## API Reference

### Methods

#### ` + "`process`" + `
Processes input data according to plugin logic.

**Parameters:**
- ` + "`data`" + ` (any): Input data to process

**Returns:**
- Processed data with metadata

#### ` + "`status`" + `
Returns plugin status and configuration information.

**Returns:**
- Plugin status, host information, and current configuration

## Events

The plugin subscribes to the following events:

- ` + "`system.status`" + ` - System status changes

## Permissions

The plugin requires the following permissions:

- ` + "`file.read`" + ` - Read file access
- ` + "`file.write`" + ` - Write file access

## License

{{.Year}} {{.Author}}. Licensed under the MIT License.

## Support

For issues and feature requests, please visit the project repository.
`

const makefileTemplate = `.PHONY: build test clean dev install

PLUGIN_NAME={{.PluginName}}
OUTPUT_FILE=$(PLUGIN_NAME).so

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	go build -buildmode=plugin -o $(OUTPUT_FILE) .

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(OUTPUT_FILE)

dev: clean build test
	@echo "Development build complete"

install: build
	@echo "Installing plugin..."
	@if [ -n "$(GZH_PLUGIN_DIR)" ]; then ` + "\\" + `
		cp $(OUTPUT_FILE) $(GZH_PLUGIN_DIR)/; ` + "\\" + `
		echo "Plugin installed to $(GZH_PLUGIN_DIR)/$(OUTPUT_FILE)"; ` + "\\" + `
	else ` + "\\" + `
		echo "GZH_PLUGIN_DIR not set. Plugin built as $(OUTPUT_FILE)"; ` + "\\" + `
	fi

validate:
	@echo "Validating plugin..."
	gz plugin validate $(OUTPUT_FILE)

format:
	@echo "Formatting code..."
	go fmt ./...
	gofumpt -w .
	gci write .

lint:
	@echo "Running linter..."
	golangci-lint run

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

help:
	@echo "Available commands:"
	@echo "  build     - Build the plugin"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  dev       - Development build (clean + build + test)"
	@echo "  install   - Install plugin to GZH_PLUGIN_DIR"
	@echo "  validate  - Validate the built plugin"
	@echo "  format    - Format source code"
	@echo "  lint      - Run linter"
	@echo "  deps      - Download and tidy dependencies"
	@echo "  help      - Show this help"
`

const gitignoreTemplate = `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Plugin build artifacts
{{.PluginName}}.so

# Local configuration
config.local.yaml
*.local.yaml

# Logs
*.log
logs/

# Temporary files
tmp/
temp/
`

const pluginTestTemplate = `package test

import (
	"context"
	"testing"
	"time"

	"{{.ModuleName}}/pkg/{{.PluginName}}"
	"github.com/gzh-manager/gzh-manager-go/pkg/plugins"
)

func TestPlugin(t *testing.T) {
	plugin := {{.PluginName}}.New{{.PluginNameGo}}Plugin()
	
	// Test metadata
	metadata := plugin.GetMetadata()
	if metadata.Name != "{{.PluginName}}" {
		t.Errorf("Expected plugin name '{{.PluginName}}', got '%s'", metadata.Name)
	}
	
	if metadata.Version == "" {
		t.Error("Plugin version should not be empty")
	}
}

func TestPluginInitialization(t *testing.T) {
	plugin := {{.PluginName}}.New{{.PluginNameGo}}Plugin()
	
	// Mock plugin API
	mockAPI := &MockPluginAPI{}
	
	config := plugins.PluginConfig{
		Settings: map[string]interface{}{
			"enabled": true,
			"log_level": "info",
		},
	}
	
	ctx := context.WithValue(context.Background(), "plugin_api", mockAPI)
	
	err := plugin.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Plugin initialization failed: %v", err)
	}
}

func TestPluginExecution(t *testing.T) {
	plugin := {{.PluginName}}.New{{.PluginNameGo}}Plugin()
	
	// Initialize plugin
	mockAPI := &MockPluginAPI{}
	config := plugins.PluginConfig{
		Settings: map[string]interface{}{
			"enabled": true,
		},
	}
	
	ctx := context.WithValue(context.Background(), "plugin_api", mockAPI)
	err := plugin.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Plugin initialization failed: %v", err)
	}
	
	// Test default execution
	args := map[string]interface{}{
		"test": "data",
	}
	
	result, err := plugin.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Plugin execution failed: %v", err)
	}
	
	if result == nil {
		t.Error("Plugin execution should return a result")
	}
}

func TestPluginStatus(t *testing.T) {
	plugin := {{.PluginName}}.New{{.PluginNameGo}}Plugin()
	
	// Initialize plugin
	mockAPI := &MockPluginAPI{}
	config := plugins.PluginConfig{
		Settings: map[string]interface{}{
			"enabled": true,
		},
	}
	
	ctx := context.WithValue(context.Background(), "plugin_api", mockAPI)
	err := plugin.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Plugin initialization failed: %v", err)
	}
	
	// Test status method
	args := map[string]interface{}{
		"__method": "status",
	}
	
	result, err := plugin.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Status check failed: %v", err)
	}
	
	statusMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Status result should be a map")
		return
	}
	
	if statusMap["plugin"] == nil {
		t.Error("Status should include plugin information")
	}
}

func TestPluginHealthCheck(t *testing.T) {
	plugin := {{.PluginName}}.New{{.PluginNameGo}}Plugin()
	
	// Test health check before initialization (should fail)
	err := plugin.HealthCheck(context.Background())
	if err == nil {
		t.Error("Health check should fail before initialization")
	}
	
	// Initialize plugin
	mockAPI := &MockPluginAPI{}
	config := plugins.PluginConfig{
		Settings: map[string]interface{}{
			"enabled": true,
		},
	}
	
	ctx := context.WithValue(context.Background(), "plugin_api", mockAPI)
	err = plugin.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Plugin initialization failed: %v", err)
	}
	
	// Test health check after initialization (should pass)
	err = plugin.HealthCheck(context.Background())
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestPluginCleanup(t *testing.T) {
	plugin := {{.PluginName}}.New{{.PluginNameGo}}Plugin()
	
	// Initialize plugin
	mockAPI := &MockPluginAPI{}
	config := plugins.PluginConfig{
		Settings: map[string]interface{}{
			"enabled": true,
		},
	}
	
	ctx := context.WithValue(context.Background(), "plugin_api", mockAPI)
	err := plugin.Initialize(ctx, config)
	if err != nil {
		t.Fatalf("Plugin initialization failed: %v", err)
	}
	
	// Test cleanup
	err = plugin.Cleanup(context.Background())
	if err != nil {
		t.Errorf("Plugin cleanup failed: %v", err)
	}
}

// MockPluginAPI implements a mock version of the PluginAPI for testing
type MockPluginAPI struct{}

func (m *MockPluginAPI) GetLogger(pluginName string) plugins.Logger {
	return &MockLogger{}
}

func (m *MockPluginAPI) GetConfig(key string) (interface{}, error) {
	return nil, nil
}

func (m *MockPluginAPI) SetConfig(key string, value interface{}) error {
	return nil
}

func (m *MockPluginAPI) EmitEvent(event plugins.Event) error {
	return nil
}

func (m *MockPluginAPI) SubscribeToEvent(eventType string, handler plugins.EventHandler) error {
	return nil
}

func (m *MockPluginAPI) ReadFile(path string) ([]byte, error) {
	return []byte("mock file content"), nil
}

func (m *MockPluginAPI) WriteFile(path string, data []byte) error {
	return nil
}

func (m *MockPluginAPI) HTTPRequest(method, url string, headers map[string]string, body []byte) (*plugins.HTTPResponse, error) {
	return &plugins.HTTPResponse{
		StatusCode: 200,
		Headers:    make(map[string][]string),
		Body:       []byte("mock response"),
	}, nil
}

func (m *MockPluginAPI) CallPlugin(pluginName string, method string, args map[string]interface{}) (interface{}, error) {
	return "mock plugin response", nil
}

func (m *MockPluginAPI) GetHostInfo() plugins.HostInfo {
	return plugins.HostInfo{
		GZVersion:    "1.0.0",
		OS:           "linux",
		Architecture: "amd64",
		WorkingDir:   "/tmp",
		ConfigDir:    "/tmp/.config",
		PluginDir:    "/tmp/plugins",
	}
}

// MockLogger implements a mock logger for testing
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...map[string]interface{}) {}
func (m *MockLogger) Info(msg string, fields ...map[string]interface{})  {}
func (m *MockLogger) Warn(msg string, fields ...map[string]interface{})  {}
func (m *MockLogger) Error(msg string, err error, fields ...map[string]interface{}) {}
`

const developmentDocsTemplate = `# {{.PluginNameTitle}} Plugin Development Guide

This document provides guidance for developing and maintaining the {{.PluginNameTitle}} plugin.

## Architecture

The plugin follows the GZH Manager plugin architecture:

1. **Plugin Interface**: Implements the core Plugin interface
2. **Event Handling**: Subscribes to system events
3. **Configuration**: Uses structured configuration
4. **Logging**: Integrated logging system

## Project Structure

` + "```" + `
{{.PluginName}}/
├── main.go              # Plugin entry point
├── pkg/{{.PluginName}}/        # Plugin implementation
│   └── plugin.go        # Core plugin logic
├── test/                # Tests
│   └── plugin_test.go   # Plugin tests
├── docs/                # Documentation
├── examples/            # Usage examples
├── go.mod               # Go module definition
├── Makefile             # Build automation
└── README.md            # User documentation
` + "```" + `

## Development Workflow

### 1. Setup Development Environment

` + "```" + `bash
# Clone/navigate to plugin directory
cd {{.PluginName}}

# Install dependencies
make deps

# Format code
make format
` + "```" + `

### 2. Development Cycle

` + "```" + `bash
# Make changes to code

# Format and lint
make format
make lint

# Run tests
make test

# Build plugin
make build

# Validate plugin
make validate
` + "```" + `

### 3. Testing

The plugin includes comprehensive tests:

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test plugin lifecycle and API interactions
- **Mock Tests**: Use mock APIs for isolated testing

Run tests with:
` + "```" + `bash
make test
go test -v ./...
` + "```" + `

### 4. Building

Build the plugin for different scenarios:

` + "```" + `bash
# Development build
make dev

# Production build
make build

# Install to plugin directory
make install
` + "```" + `

## Plugin Implementation

### Core Interface Methods

1. **GetMetadata()**: Returns plugin information
2. **Initialize()**: Sets up the plugin
3. **Execute()**: Main plugin logic
4. **Cleanup()**: Resource cleanup
5. **HealthCheck()**: Validates plugin state

### Method Handling

The plugin supports multiple methods through the ` + "`__method`" + ` parameter:

- ` + "`default`" + `: Basic operation
- ` + "`status`" + `: Status information
- ` + "`process`" + `: Data processing

### Configuration

The plugin uses structured configuration:

` + "```" + `go
type Config struct {
    Enabled  bool   ` + "`yaml:\"enabled\"`" + `
    LogLevel string ` + "`yaml:\"log_level\"`" + `
}
` + "```" + `

### Event Handling

Subscribe to system events:

` + "```" + `go
p.api.SubscribeToEvent("system.status", p.handleSystemStatus)
` + "```" + `

## Best Practices

### Error Handling

- Always check for errors and handle them appropriately
- Use structured error messages with context
- Log errors with relevant information

### Logging

- Use appropriate log levels (debug, info, warn, error)
- Include structured fields for better log analysis
- Avoid logging sensitive information

### Resource Management

- Clean up resources in the Cleanup method
- Use context cancellation for long-running operations
- Implement proper timeout handling

### Testing

- Write tests for all public methods
- Use mock objects for external dependencies
- Test error conditions and edge cases

### Configuration

- Provide sensible defaults
- Validate configuration values
- Support configuration reloading when possible

## Debugging

### Enable Debug Logging

Set log level to debug in configuration:

` + "```" + `yaml
plugins:
  {{.PluginName}}:
    log_level: debug
` + "```" + `

### Use Plugin Validation

` + "```" + `bash
gz plugin validate --strict {{.PluginName}}.so
` + "```" + `

### Test Individual Methods

` + "```" + `bash
gz plugin-execute {{.PluginName}} --method=status
` + "```" + `

## Performance Considerations

- Minimize memory allocations in hot paths
- Use efficient data structures
- Implement proper caching when beneficial
- Monitor resource usage during development

## Security

- Validate all input parameters
- Follow principle of least privilege
- Use secure coding practices
- Be aware of plugin sandbox restrictions

## Deployment

### Building for Production

` + "```" + `bash
# Clean build
make clean
make build

# Validate
make validate
` + "```" + `

### Installation

` + "```" + `bash
# Set plugin directory
export GZH_PLUGIN_DIR=/path/to/plugins

# Install
make install
` + "```" + `

### Configuration

Place configuration in GZH Manager config file:

` + "```" + `yaml
plugins:
  {{.PluginName}}:
    enabled: true
    log_level: info
` + "```" + `

## Troubleshooting

### Common Issues

1. **Plugin not loading**: Check exports and interface implementation
2. **Configuration errors**: Validate YAML syntax and schema
3. **Permission errors**: Review required permissions
4. **Performance issues**: Profile memory and CPU usage

### Debug Commands

` + "```" + `bash
# Check plugin metadata
gz plugin validate {{.PluginName}}.so

# Test plugin execution
gz plugin-execute {{.PluginName}} --method=status

# View plugin logs
gz logs --plugin={{.PluginName}}
` + "```" + `

## Contributing

1. Follow the coding standards
2. Write comprehensive tests
3. Update documentation
4. Submit pull requests with clear descriptions

## Resources

- [GZH Manager Plugin API Reference](../../../docs/plugin-api.md)
- [Plugin Development Best Practices](../../../docs/plugin-best-practices.md)
- [Go Plugin Documentation](https://golang.org/pkg/plugin/)
`

const configExampleTemplate = `# {{.PluginNameTitle}} Plugin Configuration Example

# Basic configuration
plugins:
  {{.PluginName}}:
    # Enable or disable the plugin
    enabled: true
    
    # Logging level: debug, info, warn, error
    log_level: info

# Advanced configuration example
# plugins:
#   {{.PluginName}}:
#     enabled: true
#     log_level: debug
#     
#     # Custom settings (plugin-specific)
#     custom_setting: "value"
#     timeout: 30
#     retry_count: 3
`

const commandTemplate = `package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// commandCmd represents the command functionality for this plugin
var commandCmd = &cobra.Command{
	Use:   "{{.PluginName}}",
	Short: "{{.PluginNameTitle}} command",
	Long:  ` + "`{{.PluginNameTitle}} command provides functionality for the {{.PluginName}} plugin.`" + `,
	RunE:  runCommand,
}

var (
	configFile string
	verbose    bool
	dryRun     bool
)

func init() {
	commandCmd.Flags().StringVar(&configFile, "config", "", "Configuration file")
	commandCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	commandCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
}

func runCommand(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Println("Running {{.PluginName}} command...")
	}
	
	// Implement command logic here
	fmt.Printf("{{.PluginNameTitle}} command executed successfully\n")
	
	if dryRun {
		fmt.Println("Dry run mode - no changes made")
	}
	
	return nil
}

func main() {
	if err := commandCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
`

const serviceTemplate = `package internal

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// {{.PluginNameGo}}Service provides service functionality for the plugin
type {{.PluginNameGo}}Service struct {
	config  ServiceConfig
	running bool
	mu      sync.RWMutex
	stopCh  chan struct{}
}

// ServiceConfig holds service configuration
type ServiceConfig struct {
	Enabled      bool          ` + "`yaml:\"enabled\"`" + `
	Interval     time.Duration ` + "`yaml:\"interval\"`" + `
	MaxRetries   int           ` + "`yaml:\"max_retries\"`" + `
	Timeout      time.Duration ` + "`yaml:\"timeout\"`" + `
}

// NewService creates a new service instance
func NewService(config ServiceConfig) *{{.PluginNameGo}}Service {
	return &{{.PluginNameGo}}Service{
		config: config,
		stopCh: make(chan struct{}),
	}
}

// Start starts the service
func (s *{{.PluginNameGo}}Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return fmt.Errorf("service is already running")
	}
	
	if !s.config.Enabled {
		return fmt.Errorf("service is disabled")
	}
	
	s.running = true
	
	go s.run(ctx)
	
	return nil
}

// Stop stops the service
func (s *{{.PluginNameGo}}Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return nil
	}
	
	close(s.stopCh)
	s.running = false
	
	return nil
}

// IsRunning returns whether the service is running
func (s *{{.PluginNameGo}}Service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// run is the main service loop
func (s *{{.PluginNameGo}}Service) run(ctx context.Context) {
	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.performWork(ctx)
		}
	}
}

// performWork performs the main service work
func (s *{{.PluginNameGo}}Service) performWork(ctx context.Context) {
	workCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	
	// Implement service work here
	_ = workCtx
	
	// Example: process data, check status, etc.
}

// GetStatus returns service status information
func (s *{{.PluginNameGo}}Service) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return map[string]interface{}{
		"running":    s.running,
		"enabled":    s.config.Enabled,
		"interval":   s.config.Interval.String(),
		"timeout":    s.config.Timeout.String(),
		"max_retries": s.config.MaxRetries,
	}
}
`

const filterTemplate = `package internal

import (
	"fmt"
	"strings"
)

// {{.PluginNameGo}}Filter provides filtering functionality for the plugin
type {{.PluginNameGo}}Filter struct {
	config FilterConfig
}

// FilterConfig holds filter configuration
type FilterConfig struct {
	Enabled    bool     ` + "`yaml:\"enabled\"`" + `
	Rules      []string ` + "`yaml:\"rules\"`" + `
	Mode       string   ` + "`yaml:\"mode\"`" + ` // include, exclude
	CaseSensitive bool  ` + "`yaml:\"case_sensitive\"`" + `
}

// FilterResult represents the result of a filter operation
type FilterResult struct {
	Matched bool                   ` + "`json:\"matched\"`" + `
	Rule    string                 ` + "`json:\"rule,omitempty\"`" + `
	Data    map[string]interface{} ` + "`json:\"data,omitempty\"`" + `
}

// NewFilter creates a new filter instance
func NewFilter(config FilterConfig) *{{.PluginNameGo}}Filter {
	return &{{.PluginNameGo}}Filter{
		config: config,
	}
}

// Filter applies filter rules to input data
func (f *{{.PluginNameGo}}Filter) Filter(input interface{}) (*FilterResult, error) {
	if !f.config.Enabled {
		return &FilterResult{
			Matched: true,
			Data:    map[string]interface{}{"input": input},
		}, nil
	}
	
	inputStr := f.convertToString(input)
	
	for _, rule := range f.config.Rules {
		matched := f.matchRule(inputStr, rule)
		
		if matched {
			result := &FilterResult{
				Matched: f.shouldInclude(matched),
				Rule:    rule,
				Data:    map[string]interface{}{"input": input},
			}
			
			if result.Matched {
				result.Data["filter_rule"] = rule
			}
			
			return result, nil
		}
	}
	
	// No rules matched
	return &FilterResult{
		Matched: f.config.Mode == "exclude", // If exclude mode, no match means include
		Data:    map[string]interface{}{"input": input},
	}, nil
}

// FilterBatch applies filter to multiple inputs
func (f *{{.PluginNameGo}}Filter) FilterBatch(inputs []interface{}) ([]*FilterResult, error) {
	results := make([]*FilterResult, len(inputs))
	
	for i, input := range inputs {
		result, err := f.Filter(input)
		if err != nil {
			return nil, fmt.Errorf("failed to filter item %d: %w", i, err)
		}
		results[i] = result
	}
	
	return results, nil
}

// GetMatchedResults returns only the matched results from a batch
func (f *{{.PluginNameGo}}Filter) GetMatchedResults(results []*FilterResult) []*FilterResult {
	var matched []*FilterResult
	for _, result := range results {
		if result.Matched {
			matched = append(matched, result)
		}
	}
	return matched
}

// convertToString converts input to string for pattern matching
func (f *{{.PluginNameGo}}Filter) convertToString(input interface{}) string {
	switch v := input.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// matchRule checks if input matches a specific rule
func (f *{{.PluginNameGo}}Filter) matchRule(input, rule string) bool {
	if !f.config.CaseSensitive {
		input = strings.ToLower(input)
		rule = strings.ToLower(rule)
	}
	
	// Support different pattern types
	switch {
	case strings.HasPrefix(rule, "prefix:"):
		pattern := strings.TrimPrefix(rule, "prefix:")
		return strings.HasPrefix(input, pattern)
	case strings.HasPrefix(rule, "suffix:"):
		pattern := strings.TrimPrefix(rule, "suffix:")
		return strings.HasSuffix(input, pattern)
	case strings.HasPrefix(rule, "contains:"):
		pattern := strings.TrimPrefix(rule, "contains:")
		return strings.Contains(input, pattern)
	case strings.HasPrefix(rule, "exact:"):
		pattern := strings.TrimPrefix(rule, "exact:")
		return input == pattern
	default:
		// Default to contains match
		return strings.Contains(input, rule)
	}
}

// shouldInclude determines if a match should be included based on mode
func (f *{{.PluginNameGo}}Filter) shouldInclude(matched bool) bool {
	switch f.config.Mode {
	case "include":
		return matched
	case "exclude":
		return !matched
	default:
		return matched
	}
}

// GetConfig returns the current filter configuration
func (f *{{.PluginNameGo}}Filter) GetConfig() FilterConfig {
	return f.config
}

// UpdateConfig updates the filter configuration
func (f *{{.PluginNameGo}}Filter) UpdateConfig(config FilterConfig) {
	f.config = config
}

// GetStats returns filter statistics
func (f *{{.PluginNameGo}}Filter) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":        f.config.Enabled,
		"mode":          f.config.Mode,
		"rule_count":    len(f.config.Rules),
		"case_sensitive": f.config.CaseSensitive,
		"rules":         f.config.Rules,
	}
}
`
