<!-- 🚫 AI_MODIFY_PROHIBITED -->

<!-- This file should not be modified by AI agents -->

# Interactive Debugging Shell Specification

## Overview

The `shell` command provides an interactive debugging shell (REPL) for real-time system inspection, dynamic configuration changes, and live troubleshooting. It offers a command-line interface for developers and administrators to interact with the system state, monitor performance, and execute debugging operations.

## Commands

### Core Shell Command

- `gz shell` - Start interactive debugging shell (REPL)

### Interactive Shell (`gz shell`)

**Purpose**: Start an interactive debugging shell for real-time system inspection

**Features**:

- Real-time system state inspection
- Dynamic configuration changes
- Live debugging and troubleshooting
- Interactive plugin execution
- Memory and performance monitoring
- Command history and auto-completion
- Timeout protection
- Signal handling for graceful shutdown

**Usage**:

```bash
gz shell                                    # Start interactive shell
gz shell --timeout 30m                     # Auto-exit after 30 minutes
gz shell --quiet                           # Quiet mode - minimal output
gz shell --no-history                      # Disable command history
```

**Parameters**:

- `--timeout` (default: 0): Auto-exit timeout (0 = no timeout)
- `--quiet` (default: false): Quiet mode - minimal output
- `--no-history` (default: false): Disable command history

## Built-in Shell Commands

### System Information Commands

#### help

**Purpose**: Show available commands or detailed help for specific commands

**Usage**:

```
gz> help                    # Show all available commands
gz> help status             # Show help for specific command
```

#### status

**Purpose**: Show comprehensive system status

**Features**:

- Health status reporting
- System uptime tracking
- Memory usage monitoring
- Goroutine count
- Version information

**Usage**:

```
gz> status                  # Show system status in table format
gz> status --json           # Show system status as JSON
```

#### memory

**Purpose**: Display detailed memory usage information

**Features**:

- Current memory allocation
- Total allocation tracking
- System memory usage
- Garbage collection statistics
- Heap object count
- Optional garbage collection trigger

**Usage**:

```
gz> memory                  # Show memory usage
gz> memory --json           # Show memory usage as JSON
gz> memory --gc             # Run garbage collection first
```

#### metrics

**Purpose**: Display system performance metrics

**Features**:

- CPU usage monitoring
- Memory usage tracking
- Disk usage statistics
- Network statistics
- Load average information
- Watch mode for continuous monitoring

**Usage**:

```
gz> metrics                 # Show current metrics
gz> metrics --json          # Show metrics as JSON
gz> metrics --watch         # Continuous monitoring mode
```

### Configuration Commands

#### config

**Purpose**: Show and modify system configuration

**Features**:

- Configuration viewing
- Dynamic configuration updates
- Configuration validation

**Usage**:

```
gz> config list             # List all configuration
gz> config get <key>        # Get specific configuration value
gz> config set <key> <value> # Set configuration value
```

**Note**: Configuration management integration is not yet fully implemented in the shell.

### Debugging Commands

#### trace

**Purpose**: Control execution tracing

**Features**:

- Start/stop execution tracing
- Trace status monitoring
- Integration with Go's trace package

**Usage**:

```
gz> trace start             # Start execution tracing
gz> trace stop              # Stop execution tracing
gz> trace status            # Show trace status
```

#### profile

**Purpose**: Control performance profiling within shell

**Features**:

- Start/stop profiling sessions
- Profiling status monitoring
- Integration with main profile command

**Usage**:

```
gz> profile start           # Start performance profiling
gz> profile stop            # Stop performance profiling
gz> profile status          # Show profiling status
```

### Plugin Management Commands

#### plugins

**Purpose**: List and manage plugins

**Features**:

- Plugin enumeration
- Plugin execution
- Plugin status monitoring

**Usage**:

```
gz> plugins list            # List available plugins
gz> plugins exec <n> <method> # Execute plugin method
```

**Note**: Plugin functionality has been disabled in the current version.

### Utility Commands

#### history

**Purpose**: Manage command history

**Features**:

- Display command history
- Clear command history
- Limit history display count

**Usage**:

```
gz> history                 # Show all command history
gz> history --count 10      # Show last 10 commands
gz> history --clear         # Clear command history
```

#### clear

**Purpose**: Clear the terminal screen

**Usage**:

```
gz> clear                   # Clear screen
```

#### context

**Purpose**: Show shell execution context

**Features**:

- Shell startup information
- Session uptime
- Command execution statistics
- Shell variables (future feature)

**Usage**:

```
gz> context                 # Show context in table format
gz> context --json          # Show context as JSON
```

#### logs

**Purpose**: Display recent system logs

**Features**:

- Recent log viewing
- Log level filtering
- Configurable log count

**Usage**:

```
gz> logs                    # Show recent logs
gz> logs --count 20         # Show last 20 log entries
gz> logs --level error      # Filter by log level
```

**Note**: Log system integration is not yet fully implemented.

### Session Control Commands

#### exit / quit

**Purpose**: Exit the shell gracefully

**Usage**:

```
gz> exit                    # Exit shell
gz> quit                    # Alternative exit command
```

## Shell Features

### Command History

**Features**:

- Automatic command history tracking
- Duplicate command filtering
- History size limitation (100 commands)
- History persistence during session
- Optional history disabling

**History Management**:

- Commands are automatically added to history
- Consecutive duplicate commands are filtered out
- History is limited to the last 100 commands
- History can be viewed, cleared, or disabled entirely

### Auto-completion

**Features**:

- Command name completion
- Parameter completion for specific commands
- Context-aware suggestions

**Supported Completions**:

- help: Complete with available command names
- plugins: Complete with plugin actions (list, exec)

### Signal Handling

**Features**:

- Graceful shutdown on SIGINT (Ctrl+C)
- SIGTERM handling for clean termination
- Context cancellation propagation

### Timeout Protection

**Features**:

- Configurable auto-exit timeout
- Background timeout monitoring
- Graceful shutdown on timeout

### Error Handling

**Features**:

- Command error capture and display
- Non-fatal error recovery
- Detailed error messages

## Integration

### GZH Client Integration

The shell integrates with the main GZH client for:

- System health monitoring
- Metrics collection
- Configuration access
- Plugin management

### Runtime Integration

Direct integration with Go runtime for:

- Memory statistics collection
- Goroutine monitoring
- Garbage collection control
- Performance profiling

## Shell Context

### Session Information

The shell maintains context information including:

- Session start time
- Total session uptime
- Number of commands executed
- Last executed command
- Shell variables (future feature)

### State Management

- Session state is maintained in memory
- Command history persists during session
- Configuration changes are applied immediately
- Context is reset on shell restart

## Examples

### Basic Shell Usage

```bash
# Start shell and explore system
gz shell

gz> help                    # List available commands
gz> status                  # Check system health
gz> memory                  # View memory usage
gz> metrics                 # View system metrics
gz> exit                    # Exit shell
```

### Performance Monitoring

```bash
gz shell

gz> memory --gc             # Run GC and check memory
gz> metrics --watch         # Start continuous monitoring
# Press Ctrl+C to stop watching
gz> profile start           # Start profiling
gz> profile status          # Check profiling status
gz> profile stop            # Stop profiling
```

### Debugging Session

```bash
gz shell --timeout 1h       # Start with 1-hour timeout

gz> status --json           # Get detailed status
gz> trace start             # Start execution tracing
# Perform operations to trace
gz> trace stop              # Stop tracing
gz> history                 # Review commands executed
gz> context --json          # Get session context
```

### Configuration Management

```bash
gz shell

gz> config list             # View current configuration
gz> config get log.level    # Get specific setting
gz> config set log.level debug # Change log level
gz> logs --level debug      # View debug logs
```

### Command History Management

```bash
gz shell --no-history       # Start without history

# Or with history enabled:
gz shell

gz> help
gz> status
gz> memory
gz> history                 # View command history
gz> history --count 5       # Show last 5 commands
gz> history --clear         # Clear history
```

## Error Handling

### Common Errors

- **Command not found**: Unknown command entered
- **Invalid parameters**: Incorrect command parameters
- **Client connection errors**: GZH client unavailable
- **Permission errors**: Insufficient system access
- **Resource errors**: Memory or system resource issues

### Error Recovery

- **Command errors**: Display error and continue shell session
- **Client errors**: Attempt reconnection or graceful degradation
- **System errors**: Log errors and continue operation where possible
- **Fatal errors**: Exit shell with appropriate error code

## Security Considerations

### Access Control

- Shell access should be restricted to authorized users
- System-level operations require appropriate permissions
- Configuration changes are limited by user permissions

### Audit Trail

- All shell commands are logged for audit purposes
- Session start/end times are recorded
- Configuration changes are tracked

### Resource Protection

- Memory usage monitoring prevents resource exhaustion
- Timeout protection prevents runaway sessions
- Signal handling ensures clean shutdown

## Performance Considerations

### Resource Usage

- Minimal memory footprint for shell operations
- Efficient command parsing and execution
- Optimized metric collection and display

### Scalability

- Single-user shell sessions
- Lightweight operation suitable for production use
- Non-blocking operations where possible

## Future Enhancements

### Planned Features

- **Enhanced plugin system**: Full plugin management and execution
- **Configuration hot-reload**: Dynamic configuration updates
- **Log integration**: Real-time log viewing and filtering
- **Script execution**: Batch command execution from files
- **Remote shell**: Network-accessible shell interface
- **Enhanced completion**: Smarter auto-completion system

### Extensibility

- **Custom commands**: Plugin-based command extensions
- **Output formatters**: Additional output format support
- **Integration hooks**: External system integration points
- **Scripting support**: Shell script execution capabilities

## Best Practices

### Usage Guidelines

- Use shell for interactive debugging and monitoring
- Limit session duration for security
- Regular monitoring of resource usage
- Document configuration changes made through shell

### Security

- Restrict shell access to authorized personnel
- Use timeout protection in production environments
- Monitor shell usage through audit logs
- Regular review of shell command history

### Performance

- Use watch mode judiciously for continuous monitoring
- Clear command history periodically
- Monitor memory usage during extended sessions
- Exit shell when not actively in use
