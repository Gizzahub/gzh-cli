# Debug Package API Documentation

## Overview

The `pkg/debug` package provides comprehensive debugging, logging, profiling, and tracing capabilities for the gzh-manager CLI tool. It implements RFC 5424 compliant structured logging with centralized log management, dynamic log level control, and performance profiling tools.

## Table of Contents

- [Core Types](#core-types)
- [Structured Logging](#structured-logging)
- [Centralized Integration](#centralized-integration)
- [Log Level Management](#log-level-management)
- [Performance Profiling](#performance-profiling)
- [Configuration](#configuration)
- [Examples](#examples)

## Core Types

### RFC5424Severity

```go
type RFC5424Severity int
```

RFC5424Severity represents syslog severity levels as defined in RFC 5424. These numeric values correspond to standard syslog severity levels, with lower numbers indicating higher severity.

**Constants:**
- `SeverityEmergency` (0): Emergency: system is unusable
- `SeverityAlert` (1): Alert: action must be taken immediately
- `SeverityCritical` (2): Critical: critical conditions
- `SeverityError` (3): Error: error conditions
- `SeverityWarning` (4): Warning: warning conditions
- `SeverityNotice` (5): Notice: normal but significant condition
- `SeverityInfo` (6): Informational: informational messages
- `SeverityDebug` (7): Debug: debug-level messages

### StructuredLogEntry

```go
type StructuredLogEntry struct {
    Timestamp time.Time       `json:"@timestamp"`
    Version   int             `json:"@version"`
    Level     string          `json:"level"`
    Severity  RFC5424Severity `json:"severity"`
    Hostname  string          `json:"hostname"`
    AppName   string          `json:"appname"`
    ProcID    string          `json:"procid"`
    MsgID     string          `json:"msgid,omitempty"`
    Message   string          `json:"message"`
    
    // Distributed Tracing Fields
    TraceID string `json:"trace_id,omitempty"`
    SpanID  string `json:"span_id,omitempty"`
    
    // Source Code Fields
    Caller struct {
        File     string `json:"file,omitempty"`
        Line     int    `json:"line,omitempty"`
        Function string `json:"function,omitempty"`
    } `json:"caller,omitempty"`
    
    // Structured Data
    Fields map[string]interface{} `json:"fields,omitempty"`
    
    // Performance Fields
    Duration  *time.Duration `json:"duration,omitempty"`
    Latency   *time.Duration `json:"latency,omitempty"`
    BytesRead *int64         `json:"bytes_read,omitempty"`
    BytesOut  *int64         `json:"bytes_out,omitempty"`
}
```

StructuredLogEntry represents a standardized log entry following RFC 5424. It provides a comprehensive structure for logging with support for distributed tracing, caller information, and performance metrics.

## Structured Logging

### StructuredLogger

```go
type StructuredLogger struct {
    // Contains filtered or unexported fields
}
```

StructuredLogger provides RFC 5424 compliant structured logging with advanced features including async logging, sampling, caller information, and distributed tracing integration.

**Key Features:**
- Multiple output formats (JSON, logfmt, console)
- Configurable log levels and module-specific levels
- Async logging with buffering for high-performance scenarios
- Sampling for high-volume logging with intelligent rate control
- OpenTelemetry integration for distributed tracing
- File rotation and compression

### Methods

#### NewStructuredLogger

```go
func NewStructuredLogger(config *StructuredLoggerConfig) (*StructuredLogger, error)
```

Creates a new structured logger with the provided configuration. If config is nil, it uses DefaultStructuredLoggerConfig().

#### Emergency

```go
func (sl *StructuredLogger) Emergency(ctx context.Context, msg string, fields ...map[string]interface{})
```

Logs an emergency message (severity 0).

#### Alert

```go
func (sl *StructuredLogger) Alert(ctx context.Context, msg string, fields ...map[string]interface{})
```

Logs an alert message (severity 1).

#### Critical

```go
func (sl *StructuredLogger) Critical(ctx context.Context, msg string, fields ...map[string]interface{})
```

Logs a critical message (severity 2).

#### ErrorLevel

```go
func (sl *StructuredLogger) ErrorLevel(ctx context.Context, msg string, fields ...map[string]interface{})
```

Logs an error message (severity 3).

#### Warning

```go
func (sl *StructuredLogger) Warning(ctx context.Context, msg string, fields ...map[string]interface{})
```

Logs a warning message (severity 4).

#### Notice

```go
func (sl *StructuredLogger) Notice(ctx context.Context, msg string, fields ...map[string]interface{})
```

Logs a notice message (severity 5).

#### InfoLevel

```go
func (sl *StructuredLogger) InfoLevel(ctx context.Context, msg string, fields ...map[string]interface{})
```

Logs an info message (severity 6).

#### DebugLevel

```go
func (sl *StructuredLogger) DebugLevel(ctx context.Context, msg string, fields ...map[string]interface{})
```

Logs a debug message (severity 7).

#### SetLevel

```go
func (sl *StructuredLogger) SetLevel(level RFC5424Severity)
```

Sets the logging level dynamically.

#### SetModuleLevel

```go
func (sl *StructuredLogger) SetModuleLevel(module string, level RFC5424Severity)
```

Sets the logging level for a specific module.

#### WithModule

```go
func (sl *StructuredLogger) WithModule(module string) *ModuleStructuredLogger
```

Returns a logger with module-specific configuration.

#### WithFields

```go
func (sl *StructuredLogger) WithFields(fields map[string]interface{}) *FieldStructuredLogger
```

Returns a logger with pre-set fields.

#### Close

```go
func (sl *StructuredLogger) Close() error
```

Closes the logger and flushes any remaining logs.

## Centralized Integration

### CentralizedLoggerBridge

```go
type CentralizedLoggerBridge struct {
    // Contains filtered or unexported fields
}
```

CentralizedLoggerBridge provides a seamless bridge between StructuredLogger and CentralizedLogger, enabling automatic forwarding of structured log entries to the centralized logging system.

**Key Features:**
- Asynchronous log forwarding with configurable buffer sizes
- Automatic conversion between structured and centralized log formats
- Error handling with fallback to structured logging
- Runtime enable/disable capabilities
- Performance statistics and metrics
- Graceful shutdown with buffer draining

### Methods

#### NewCentralizedLoggerBridge

```go
func NewCentralizedLoggerBridge(
    structuredLogger *StructuredLogger,
    centralizedLogger *logging.CentralizedLogger,
    config *CentralizedBridgeConfig,
) (*CentralizedLoggerBridge, error)
```

Creates a new bridge instance that connects a StructuredLogger with a CentralizedLogger.

#### ForwardLogEntry

```go
func (clb *CentralizedLoggerBridge) ForwardLogEntry(entry *StructuredLogEntry) error
```

Forwards a structured log entry to the centralized logging system.

#### EnableForwarding

```go
func (clb *CentralizedLoggerBridge) EnableForwarding()
```

Enables log forwarding to centralized system.

#### DisableForwarding

```go
func (clb *CentralizedLoggerBridge) DisableForwarding()
```

Disables log forwarding to centralized system.

#### GetStats

```go
func (clb *CentralizedLoggerBridge) GetStats() map[string]interface{}
```

Returns bridge statistics including buffer utilization and performance metrics.

#### Shutdown

```go
func (clb *CentralizedLoggerBridge) Shutdown() error
```

Gracefully shuts down the bridge.

### EnhancedStructuredLogger

```go
type EnhancedStructuredLogger struct {
    *StructuredLogger
    // Contains filtered or unexported fields
}
```

EnhancedStructuredLogger extends StructuredLogger with centralized logging integration.

#### NewEnhancedStructuredLogger

```go
func NewEnhancedStructuredLogger(
    config *StructuredLoggerConfig,
    centralizedLogger *logging.CentralizedLogger,
    bridgeConfig *CentralizedBridgeConfig,
) (*EnhancedStructuredLogger, error)
```

Creates a structured logger with centralized logging integration.

## Log Level Management

### LogLevelManager

```go
type LogLevelManager struct {
    // Contains filtered or unexported fields
}
```

LogLevelManager provides advanced log level management with dynamic rule evaluation, profile switching, and performance-aware sampling.

**Key Features:**
- Dynamic rule evaluation based on conditions
- Profile-based configuration management
- HTTP API for runtime control
- Signal-based configuration reloading
- Adaptive sampling based on system performance
- Performance metrics collection and analysis

### LogLevelRule

```go
type LogLevelRule struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Enabled     bool           `json:"enabled"`
    Conditions  []LogCondition `json:"conditions"`
    Actions     []LogAction    `json:"actions"`
    Priority    int            `json:"priority"`
    Created     time.Time      `json:"created"`
    LastApplied *time.Time     `json:"last_applied,omitempty"`
    ApplyCount  int64          `json:"apply_count"`
}
```

LogLevelRule represents a rule for conditional logging that allows dynamic adjustment of log levels based on various conditions.

### LogLevelProfile

```go
type LogLevelProfile struct {
    Name         string                     `json:"name"`
    Description  string                     `json:"description"`
    GlobalLevel  RFC5424Severity            `json:"global_level"`
    ModuleLevels map[string]RFC5424Severity `json:"module_levels"`
    Rules        []LogLevelRule             `json:"rules"`
    Sampling     SamplingConfig             `json:"sampling"`
    Created      time.Time                  `json:"created"`
}
```

LogLevelProfile represents a predefined set of log level configurations for different operational scenarios (development, testing, production).

## Performance Profiling

### Profiler

```go
type Profiler struct {
    // Contains filtered or unexported fields
}
```

Profiler provides comprehensive profiling capabilities for performance analysis, memory debugging, and concurrency analysis.

**Features:**
- Multiple profiling types (CPU, memory, goroutine, block, mutex)
- Continuous and on-demand profiling modes
- HTTP endpoint integration with Go's pprof package
- Configurable output formats and destinations
- Thread-safe operation for concurrent environments

### ProfilerConfig

```go
type ProfilerConfig struct {
    Enabled        bool          `json:"enabled"`
    CPUProfile     bool          `json:"cpu_profile"`
    MemoryProfile  bool          `json:"memory_profile"`
    GoroutineTrace bool          `json:"goroutine_trace"`
    BlockProfile   bool          `json:"block_profile"`
    MutexProfile   bool          `json:"mutex_profile"`
    OutputDir      string        `json:"output_dir"`
    Duration       time.Duration `json:"duration"`
    Interval       time.Duration `json:"interval"`
    HTTPEndpoint   string        `json:"http_endpoint"`
}
```

ProfilerConfig holds comprehensive configuration for the profiler, enabling fine-grained control over different types of profiling activities.

## Configuration

### StructuredLoggerConfig

```go
type StructuredLoggerConfig struct {
    // Basic Configuration
    Level       RFC5424Severity `json:"level"`
    Format      string          `json:"format"` // "json", "logfmt", "console"
    Output      string          `json:"output"` // "stdout", "stderr", or file path
    AppName     string          `json:"app_name"`
    Version     string          `json:"version"`
    Environment string          `json:"environment"`
    
    // Trace Configuration
    EnableTracing bool `json:"enable_tracing"`
    EnableCaller  bool `json:"enable_caller"`
    CallerSkip    int  `json:"caller_skip"`
    
    // Sampling Configuration
    EnableSampling  bool    `json:"enable_sampling"`
    SampleRate      float64 `json:"sample_rate"`
    SampleThreshold int     `json:"sample_threshold"`
    
    // Performance Configuration
    AsyncLogging  bool          `json:"async_logging"`
    BufferSize    int           `json:"buffer_size"`
    FlushInterval time.Duration `json:"flush_interval"`
    MaxFileSize   int64         `json:"max_file_size"`
    MaxBackups    int           `json:"max_backups"`
    Compress      bool          `json:"compress"`
    
    // Filter Configuration
    ModuleLevels map[string]RFC5424Severity `json:"module_levels,omitempty"`
    IgnoreFields []string                   `json:"ignore_fields,omitempty"`
}
```

### IntegratedLoggingConfig

```go
type IntegratedLoggingConfig struct {
    // Application settings
    AppName     string `json:"app_name" yaml:"app_name"`
    Environment string `json:"environment" yaml:"environment"`
    Version     string `json:"version" yaml:"version"`
    
    // Log levels and output
    Level     string `json:"level" yaml:"level"`
    Format    string `json:"format" yaml:"format"`
    Directory string `json:"directory" yaml:"directory"`
    
    // Structured logging specific
    StructuredConfig *StructuredLoggerConfig `json:"structured" yaml:"structured"`
    
    // Centralized logging specific
    CentralizedConfig *logging.CentralizedLoggingConfig `json:"centralized" yaml:"centralized"`
    
    // Integration settings
    EnableCentralizedForwarding bool                     `json:"enable_centralized_forwarding" yaml:"enable_centralized_forwarding"`
    BridgeConfig                *CentralizedBridgeConfig `json:"bridge" yaml:"bridge"`
    RemoteShipping              *RemoteShippingConfig    `json:"remote_shipping" yaml:"remote_shipping"`
}
```

## Examples

### Basic Structured Logging

```go
package main

import (
    "context"
    "log"
    
    "github.com/gizzahub/gzh-manager-go/pkg/debug"
)

func main() {
    config := debug.DefaultStructuredLoggerConfig()
    logger, err := debug.NewStructuredLogger(config)
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Close()
    
    ctx := context.Background()
    logger.InfoLevel(ctx, "Application started", map[string]interface{}{
        "version": "1.0.0",
        "env":     "production",
    })
}
```

### Integrated Logging with Remote Shipping

```go
package main

import (
    "context"
    "log"
    
    "github.com/gizzahub/gzh-manager-go/pkg/debug"
)

func main() {
    config := debug.DefaultIntegratedLoggingConfig()
    config.AddElasticsearchShipper("es", "http://localhost:9200", "app-logs")
    config.AddLokiShipper("loki", "http://localhost:3100", map[string]string{
        "service": "gzh-manager",
        "env":     "production",
    })
    
    setup, err := debug.NewIntegratedLoggingSetup(config)
    if err != nil {
        log.Fatal(err)
    }
    defer setup.Shutdown()
    
    logger := setup.GetLogger()
    ctx := context.Background()
    logger.InfoLevel(ctx, "Processing request", map[string]interface{}{
        "request_id": "req-123",
        "user_id":    "user-456",
    })
}
```

### Dynamic Log Level Control

```go
package main

import (
    "context"
    "log"
    
    "github.com/gizzahub/gzh-manager-go/pkg/debug"
)

func main() {
    config := debug.DefaultStructuredLoggerConfig()
    logger, err := debug.NewStructuredLogger(config)
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Close()
    
    manager, err := debug.NewLogLevelManager(debug.DefaultLogLevelManagerConfig(), logger)
    if err != nil {
        log.Fatal(err)
    }
    
    // Apply production profile
    manager.ApplyProfile("production")
    
    // Start HTTP API for runtime control
    manager.StartHTTPServer(8080)
    
    // Add custom rule
    rule := debug.LogLevelRule{
        ID:      "debug-module",
        Enabled: true,
        Conditions: []debug.LogCondition{{
            Field:    "module",
            Operator: "eq",
            Value:    "database",
        }},
        Actions: []debug.LogAction{{
            Type:  "set_level",
            Value: debug.SeverityDebug,
        }},
    }
    manager.AddRule(rule)
}
```

### Performance Profiling

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/gizzahub/gzh-manager-go/pkg/debug"
)

func main() {
    config := debug.DefaultProfilerConfig()
    config.CPUProfile = true
    config.MemoryProfile = true
    config.Duration = 30 * time.Second
    
    profiler := debug.NewProfiler(config)
    
    ctx := context.Background()
    if err := profiler.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer profiler.Stop()
    
    // Application logic here
    time.Sleep(config.Duration)
}
```

## Utility Functions

### ParseRFC5424Severity

```go
func ParseRFC5424Severity(level string) (RFC5424Severity, error)
```

Parses a string log level into RFC5424Severity. Accepts both full names and common abbreviations.

### DefaultStructuredLoggerConfig

```go
func DefaultStructuredLoggerConfig() *StructuredLoggerConfig
```

Returns a default structured logger configuration suitable for development environments.

### DefaultIntegratedLoggingConfig

```go
func DefaultIntegratedLoggingConfig() *IntegratedLoggingConfig
```

Returns a default integrated logging configuration.

### DefaultProfilerConfig

```go
func DefaultProfilerConfig() *ProfilerConfig
```

Returns a default profiler configuration suitable for development and debugging scenarios.

## Global Instances

### InitGlobalStructuredLogger

```go
func InitGlobalStructuredLogger(config *StructuredLoggerConfig) error
```

Initializes the global structured logger instance.

### GetGlobalStructuredLogger

```go
func GetGlobalStructuredLogger() *StructuredLogger
```

Returns the global structured logger instance.

### InitGlobalIntegratedLogging

```go
func InitGlobalIntegratedLogging(config *IntegratedLoggingConfig) error
```

Initializes the global integrated logging system.

### GetGlobalIntegratedLogging

```go
func GetGlobalIntegratedLogging() *LoggingSetup
```

Returns the global integrated logging system.