# Configuration Hot-Reloading

This document describes the configuration hot-reloading functionality implemented in gzh-manager-go.

## Overview

Configuration hot-reloading allows the application to automatically detect and reload configuration file changes without requiring a restart. This feature is built on top of the centralized configuration service and uses file system notifications for efficient watching.

## Architecture

### Core Components

1. **ConfigService Interface** (`internal/config/service.go`)
   - `WatchConfiguration(ctx, callback)` - Start watching for changes
   - `StopWatching()` - Stop watching for changes
   - `ReloadConfiguration(ctx)` - Manually reload configuration

2. **File System Watcher** (using `github.com/fsnotify/fsnotify`)
   - Monitors the configuration file for write events
   - Handles file system events in a separate goroutine
   - Graceful shutdown on context cancellation

3. **Configuration Validation**
   - Automatic validation after each reload
   - Startup validator integration for comprehensive checks
   - Error reporting and warning collection

### Hot-Reload Process

```
File Change → fsnotify Event → handleConfigChange() → ReloadConfiguration() → Validation → Callback
```

1. **File Change Detection**: fsnotify detects write events on the watched configuration file
2. **Debouncing**: Small delay (100ms) to avoid rapid multiple changes during editing
3. **Reload**: Configuration is reloaded from disk using the unified facade
4. **Validation**: New configuration is validated using startup validator
5. **Callback**: Registered callback functions are notified of the change

## Usage

### Programmatic Usage

```go
// Create configuration service with watching enabled
options := &ConfigServiceOptions{
    WatchEnabled:      true,
    ValidationEnabled: true,
}
service, err := NewConfigService(options)

// Load initial configuration
config, err := service.LoadConfiguration(ctx, "gzh.yaml")

// Set up change callback
callback := func(newConfig *config.UnifiedConfig) {
    fmt.Printf("Configuration changed! New provider: %s\n", newConfig.DefaultProvider)
    
    // Check validation results
    result := service.GetValidationResult()
    if !result.IsValid {
        fmt.Printf("Validation failed: %d errors\n", len(result.Errors))
    }
}

// Start watching
err = service.WatchConfiguration(ctx, callback)
defer service.StopWatching()
```

### Command Line Usage

The `gz config watch` command demonstrates hot-reloading functionality:

```bash
# Watch default configuration file
gz config watch

# Watch specific configuration file
gz config watch my-config.yaml

# Verbose output with detailed change information
gz config watch --verbose

# Custom status display interval
gz config watch --interval 10s
```

#### Watch Command Features

- **Real-time Watching**: Shows configuration changes as they happen
- **Validation Feedback**: Displays validation errors and warnings immediately
- **Status Updates**: Periodic status reports showing uptime and change count
- **Graceful Shutdown**: Handles Ctrl+C and system signals properly
- **Configuration Summary**: Shows current configuration details

## Implementation Details

### Error Handling

- **Temporary Invalid Configurations**: During editing, configurations may be temporarily invalid. The watcher continues watching and reports errors without stopping.
- **File System Errors**: Watcher errors are logged but don't stop the watching process
- **Validation Errors**: Invalid configurations are loaded but marked with validation errors

### Debouncing

The system includes a 100ms delay after file changes to handle:
- Multiple rapid writes during file editing
- Temporary file operations by editors
- Atomic file operations (rename/move)

### Thread Safety

- All configuration access is protected by read-write mutexes
- Callback execution is synchronized to prevent race conditions
- Graceful shutdown prevents resource leaks

### Performance Considerations

- **Efficient File Watching**: Uses native file system events (inotify on Linux)
- **Minimal Overhead**: Only validates and reloads when files actually change
- **Memory Management**: Proper cleanup of watchers and goroutines

## Testing

The hot-reloading functionality includes comprehensive tests:

### Unit Tests

```go
func TestConfigService_WatchConfiguration(t *testing.T)  // Basic watching functionality
func TestWatchConfigHotReloading(t *testing.T)          // Comprehensive hot-reload testing
```

### Test Scenarios

1. **Single Configuration Change**: Verify basic change detection and callback execution
2. **Multiple Rapid Changes**: Test debouncing and rapid change handling
3. **Invalid Configuration Handling**: Ensure validation errors are properly reported
4. **Graceful Shutdown**: Test context cancellation and cleanup

### Running Tests

```bash
# Run configuration service tests
go test ./internal/config -v -run TestConfigService_WatchConfiguration

# Run watch command tests  
go test ./cmd/config -v -run TestWatchConfigHotReloading

# Skip file watching tests in CI (optional)
CI=true go test ./internal/config -v
```

## Configuration Examples

### Basic Hot-Reload Setup

```yaml
# gzh.yaml
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "my-org"
        clone_dir: "~/repos/my-org"
        visibility: "all"
        strategy: "reset"
```

### Watching Multiple Changes

1. Start watcher: `gz config watch --verbose`
2. Edit configuration file to change default provider
3. Add new organizations or providers
4. Observe real-time feedback and validation

## Error Scenarios

### Common Issues and Solutions

1. **File Not Found**: Ensure configuration file path is correct
2. **Permission Denied**: Check file and directory permissions
3. **Validation Errors**: Fix configuration syntax and required fields
4. **High CPU Usage**: Check for file system loops or excessive changes

### Debugging

Enable verbose mode for detailed information:
```bash
gz config watch --verbose --interval 5s
```

This shows:
- Configuration loading details
- Validation results with specific errors
- Change timestamps and frequencies
- Current configuration summaries

## Future Enhancements

Potential improvements to the hot-reloading system:

1. **Configuration Backup**: Automatic backup before applying changes
2. **Rollback Mechanism**: Revert to last known good configuration on validation failure
3. **Change History**: Track configuration change history and diff display
4. **Remote Configuration**: Support for remote configuration sources
5. **Hot-Reload API**: HTTP endpoint for triggering manual reloads
6. **Configuration Templates**: Dynamic configuration generation from templates

## Security Considerations

- **File Access**: Only watch files the process has read access to
- **Path Validation**: Ensure watched paths are within expected directories
- **Environment Variables**: Sensitive tokens are not logged in watch output
- **Callback Security**: Validate callback functions to prevent code injection

## Conclusion

The configuration hot-reloading system provides a robust, efficient way to manage configuration changes without application restarts. It's built with production reliability in mind, including proper error handling, validation, and resource cleanup.

For development workflows, it enables rapid iteration on configuration without interrupting running processes. For production environments, it allows for dynamic configuration updates with immediate validation feedback.