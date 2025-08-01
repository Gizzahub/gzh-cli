<!-- 🚫 AI_MODIFY_PROHIBITED -->
<!-- This file should not be modified by AI agents -->

# Performance Profiling and Benchmarking Specification

## Overview

The `profile` command provides comprehensive performance profiling and benchmarking capabilities for analyzing system performance, detecting bottlenecks, and comparing performance metrics over time. It integrates with Go's built-in profiling tools and provides additional benchmarking utilities.

## Commands

### Core Commands

- `gz profile start` - Start a new profiling session
- `gz profile stop` - Stop an active profiling session
- `gz profile list` - List active profiling sessions
- `gz profile server` - Start HTTP profiling server (pprof)
- `gz profile benchmark` - Run performance benchmarks
- `gz profile stats` - Show runtime performance statistics
- `gz profile compare` - Compare benchmark results

### Start Profiling Session (`gz profile start`)

**Purpose**: Initialize a new profiling session of the specified type

**Features**:
- CPU profiling with customizable duration
- Memory heap profiling
- Goroutine profiling
- Block profiling
- Mutex profiling
- Thread creation profiling
- Custom output directory support
- Automatic session ID generation

**Usage**:
```bash
gz profile start --type cpu --duration 30s          # CPU profiling for 30 seconds
gz profile start --type memory                      # Memory heap profiling
gz profile start --type goroutine --output-dir ./profiles  # Goroutine profiling with custom output
gz profile start --type block                       # Block profiling
gz profile start --type mutex                       # Mutex profiling
gz profile start --type threadcreate               # Thread creation profiling
```

**Parameters**:
- `--type` (required): Profile type (cpu, memory, goroutine, block, mutex, threadcreate)
- `--duration` (default: 30s): Duration for CPU profiling (ignored for other types)
- `--output-dir`: Custom output directory for profile files

### Stop Profiling Session (`gz profile stop`)

**Purpose**: Stop an active profiling session and save results

**Features**:
- Session ID validation
- Automatic profile file generation
- Results summary display

**Usage**:
```bash
gz profile stop --session cpu_1640995200           # Stop specific CPU profiling session
gz profile stop --session memory_1640995300        # Stop specific memory profiling session
```

**Parameters**:
- `--session` (required): Session ID to stop

### List Active Sessions (`gz profile list`)

**Purpose**: Display all currently active profiling sessions

**Features**:
- Session ID display
- Profile type identification
- Start time and duration tracking
- Real-time duration updates

**Usage**:
```bash
gz profile list                                    # Show all active sessions
```

### HTTP Profiling Server (`gz profile server`)

**Purpose**: Start an HTTP server exposing pprof endpoints for live profiling

**Features**:
- Standard pprof endpoints (/debug/pprof/*)
- CPU profile endpoint (/debug/pprof/profile)
- Heap profile endpoint (/debug/pprof/heap)
- Goroutine profile endpoint (/debug/pprof/goroutine)
- Block profile endpoint (/debug/pprof/block)
- Mutex profile endpoint (/debug/pprof/mutex)
- Runtime statistics endpoint (/debug/stats)
- Automatic periodic profiling option

**Usage**:
```bash
gz profile server --port 6060                      # Start server on port 6060
gz profile server --port 8080 --auto-profile       # Start with automatic profiling
```

**Parameters**:
- `--port` (default: 6060): HTTP server port
- `--auto-profile` (default: false): Enable automatic periodic profiling

**Endpoints**:
- `http://localhost:6060/debug/pprof/` - Profile index
- `http://localhost:6060/debug/pprof/profile` - CPU profile
- `http://localhost:6060/debug/pprof/heap` - Heap profile
- `http://localhost:6060/debug/pprof/goroutine` - Goroutine profile
- `http://localhost:6060/debug/pprof/block` - Block profile
- `http://localhost:6060/debug/pprof/mutex` - Mutex profile
- `http://localhost:6060/debug/stats` - Runtime statistics

### Performance Benchmarks (`gz profile benchmark`)

**Purpose**: Run built-in performance benchmarks for system operations

**Features**:
- Built-in benchmark suites
- Custom iteration counts
- Concurrency testing
- Memory profiling integration
- CPU profiling integration
- Warmup runs support
- Percentile calculations
- Operations per second metrics

**Usage**:
```bash
gz profile benchmark --name memory-allocation --iterations 1000     # Memory allocation benchmark
gz profile benchmark --name goroutine-creation --concurrency 4      # Concurrent goroutine benchmark
gz profile benchmark --name json-marshal --duration 10s --memory-profiling  # JSON marshaling with memory profiling
gz profile benchmark --name channel-operations --iterations 500     # Channel operations benchmark
gz profile benchmark --name string-operations --warmup 100          # String operations with warmup
```

**Parameters**:
- `--name` (default: memory-allocation): Benchmark name
- `--iterations` (default: 1000): Number of benchmark iterations
- `--concurrency` (default: 1): Number of concurrent goroutines
- `--duration` (default: 0): Benchmark duration (overrides iterations)
- `--warmup` (default: 100): Number of warmup runs
- `--memory-profiling` (default: true): Enable memory profiling
- `--cpu-profiling` (default: false): Enable CPU profiling

**Built-in Benchmarks**:
- `memory-allocation`: Memory allocation performance
- `goroutine-creation`: Goroutine creation performance
- `channel-operations`: Channel send/receive performance
- `json-marshal`: JSON marshaling performance
- `string-operations`: String concatenation and manipulation

### Runtime Statistics (`gz profile stats`)

**Purpose**: Display current runtime performance statistics

**Features**:
- Memory usage statistics
- Goroutine count
- Garbage collection statistics
- CGO call count
- Multiple output formats (table, json, yaml)

**Usage**:
```bash
gz profile stats                                   # Show statistics in table format
gz profile stats --format json                     # Show statistics as JSON
gz profile stats --format yaml                     # Show statistics as YAML
```

**Parameters**:
- `--format` (default: table): Output format (table, json, yaml)

### Benchmark Comparison (`gz profile compare`)

**Purpose**: Compare benchmark results to analyze performance differences

**Features**:
- Performance regression detection
- Improvement identification
- Side-by-side comparison
- Statistical analysis

**Usage**:
```bash
gz profile compare --baseline "v1.0" --current "v1.1"    # Compare versions
```

**Parameters**:
- `--baseline` (required): Baseline benchmark name
- `--current` (required): Current benchmark name

**Note**: This feature requires implementation of benchmark result storage and retrieval system.

## Configuration

### Profile Configuration

The profiling system can be configured through:
- Environment variables
- Configuration files
- Command-line flags

### Output Directory

Default profile output directory: `tmp/profiles`
Custom output directory can be specified with `--output-dir` flag.

### Session Management

- Session IDs are automatically generated with format: `{type}_{timestamp}`
- Sessions are tracked in memory during application lifetime
- Profile files are saved to disk for persistent storage

## Integration

### Go pprof Integration

The profile command integrates seamlessly with Go's built-in pprof package:
```bash
# After running profiling
go tool pprof tmp/profiles/cpu_profile.prof
go tool pprof tmp/profiles/heap_profile.prof
```

### HTTP Server Integration

The HTTP server can be used with standard pprof tools:
```bash
go tool pprof http://localhost:6060/debug/pprof/profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Examples

### Complete Profiling Workflow

```bash
# Start CPU profiling
gz profile start --type cpu --duration 60s

# In another terminal, generate load
gz synclone github --org myorg --dry-run

# Check active sessions
gz profile list

# Start HTTP server for live monitoring
gz profile server --port 6060 &

# Run benchmarks
gz profile benchmark --name memory-allocation --iterations 10000

# View runtime statistics
gz profile stats --format json

# Analyze results with pprof
go tool pprof tmp/profiles/cpu_*.prof
```

### Continuous Monitoring

```bash
# Start server with auto-profiling
gz profile server --port 6060 --auto-profile

# Monitor in browser
open http://localhost:6060/debug/pprof/
```

### Performance Testing

```bash
# Run comprehensive benchmark suite
gz profile benchmark --name memory-allocation --iterations 5000 --concurrency 2
gz profile benchmark --name goroutine-creation --iterations 1000 --memory-profiling
gz profile benchmark --name channel-operations --duration 30s --cpu-profiling
```

## Error Handling

### Common Errors

- **Session not found**: Invalid session ID provided to stop command
- **Port already in use**: HTTP server port conflicts
- **Permission denied**: Insufficient permissions to write profile files
- **Unknown benchmark**: Invalid benchmark name specified

### Recovery

- Profile sessions are automatically cleaned up on application exit
- Temporary files are stored in configurable output directory
- HTTP server gracefully shuts down on context cancellation
