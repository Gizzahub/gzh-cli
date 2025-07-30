# Debug Shell Usage (Internal)

## Overview

The debug shell is a hidden feature in GZH Manager that provides an interactive REPL (Read-Eval-Print Loop) for debugging and system inspection. This feature is intended for developers and advanced troubleshooting only.

## Activation Methods

The debug shell can be activated in three ways:

### 1. Environment Variable
```bash
GZH_DEBUG_SHELL=1 gz
```

### 2. Hidden Flag
```bash
gz --debug-shell
```

### 3. Direct Command (when enabled)
When the debug shell is enabled via environment variable or during development:
```bash
gz shell
```

## Available Commands

The debug shell provides the following commands:

- `help` - Show available commands
- `status` - Show system status
- `memory` - Show memory usage
- `config` - Show/modify configuration
- `plugins` - List and manage plugins
- `logs` - Show recent logs
- `metrics` - Show system metrics
- `trace` - Start/stop tracing
- `profile` - Start/stop profiling
- `history` - Show command history
- `clear` - Clear the screen
- `context` - Show shell context
- `exit`, `quit` - Exit the shell

## Security Considerations

### Production Builds
- The debug shell should NEVER be enabled in production builds
- Consider using build tags to exclude debug code from production binaries
- The shell command is hidden from normal help output

### Access Control
- Requires explicit activation via environment variable or flag
- No sensitive data should be exposed through the shell
- All debug operations should be logged for audit purposes

### Best Practices
1. Only use in development or controlled debugging environments
2. Never share debug shell output that may contain sensitive information
3. Disable debug features before deploying to production
4. Use time-limited sessions with the `--timeout` flag

## Development Usage

### Running with Debug Shell
```bash
# Start with environment variable
GZH_DEBUG_SHELL=1 go run main.go

# Start with flag
go run main.go --debug-shell

# With timeout
go run main.go --debug-shell --timeout 30m
```

### Debugging Examples

#### Check Memory Usage
```
gz> memory --json
{
  "allocated_mb": 12.5,
  "total_alloc_mb": 45.2,
  "sys_mb": 20.1,
  "num_gc": 5,
  "goroutines": 10,
  "heap_objects": 1234
}
```

#### Monitor System Metrics
```
gz> metrics --watch
[15:04:05] CPU: 2.5%, Memory: 125 MB, Uptime: 1h30m
```

#### View System Status
```
gz> status
System Status:
  Healthy: true
  Uptime: 1h30m
  Version: 1.0.0
  Memory: 12.50 MB
  Goroutines: 10
```

## Building with Debug Support

### Development Build
```bash
# Standard development build includes debug features
make build
```

### Production Build (without debug)
```bash
# Production build should exclude debug features
make build-prod
```

## Troubleshooting

### Shell Not Starting
1. Check if `GZH_DEBUG_SHELL` environment variable is set
2. Verify the `--debug-shell` flag is being passed correctly
3. Ensure the shell package is included in the build

### Commands Not Working
1. Use `help` to see available commands
2. Check command syntax with `help <command>`
3. Verify the GZH client is properly initialized

## Important Notes

- This is an internal debugging tool and should not be documented in user-facing documentation
- The shell interface may change between versions without notice
- Performance profiling and tracing features may impact system performance
- Always exit the shell properly to ensure cleanup of resources
