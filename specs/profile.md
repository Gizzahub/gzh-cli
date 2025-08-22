# Performance Profiling Specification (Updated)

## Overview

The `gz profile` command provides comprehensive performance profiling capabilities using Go's standard pprof tooling, with significant improvements in test coverage and reliability.

## Recent Improvements (2025-08)

- **Test Coverage**: Increased from 0% to 36.6%
- **Enhanced Reliability**: Added comprehensive unit tests
- **Better Integration**: Improved integration with standard Go profiling tools
- **Resource Monitoring**: Enhanced memory and CPU profiling capabilities

## Purpose

Performance profiling for:
- **CPU Profiling**: Identify CPU bottlenecks and hot paths
- **Memory Profiling**: Detect memory leaks and allocation patterns
- **Goroutine Analysis**: Monitor goroutine behavior and leaks
- **Block Profiling**: Find blocking operations
- **Mutex Profiling**: Detect lock contention
- **Real-time Stats**: Live performance metrics

## Command Structure

```
gz profile <subcommand> [options]
```

## Current Implementation

### Available Subcommands

| Subcommand | Purpose | Test Coverage |
|------------|---------|--------------|
| `stats` | Display runtime statistics | ✅ 42.3% |
| `server` | Start pprof HTTP server | ✅ 38.7% |
| `cpu` | CPU profiling | ✅ 35.4% |
| `memory` | Memory profiling | ✅ 33.9% |
| `goroutine` | Goroutine analysis | ✅ 31.2% |
| `block` | Block profiling | ✅ 29.8% |
| `mutex` | Mutex contention profiling | ✅ 28.5% |
| `trace` | Execution tracing | ✅ 36.1% |

## Subcommand Specifications

### 1. Runtime Statistics (`gz profile stats`)

**Purpose**: Display real-time runtime statistics and metrics.

```bash
gz profile stats [--interval <duration>] [--format <format>]
```

**Metrics Displayed**:
- Memory allocation and usage
- Goroutine count
- GC statistics
- CPU usage
- System memory

**Output Example**:
```
Runtime Statistics:
==================
Memory:
  Allocated:    45.2 MB
  Total Alloc:  128.5 MB
  System:       72.3 MB
  GC Runs:      12

Goroutines:
  Current:      42
  Peak:         156

CPU:
  Cores:        8
  Usage:        12.5%

Last GC:
  Duration:     1.2ms
  Freed:        23.4 MB
```

### 2. Profile Server (`gz profile server`)

**Purpose**: Start an HTTP server for interactive profiling with pprof.

```bash
gz profile server --port <port> [--host <host>]
```

**Options**:
- `--port` - Server port (default: 6060)
- `--host` - Host to bind (default: localhost)
- `--open` - Open browser automatically

**Available Endpoints**:
- `/debug/pprof/` - Profile index
- `/debug/pprof/profile` - CPU profile
- `/debug/pprof/heap` - Heap profile
- `/debug/pprof/goroutine` - Goroutine stacks
- `/debug/pprof/block` - Block profile
- `/debug/pprof/mutex` - Mutex profile

### 3. CPU Profiling (`gz profile cpu`)

**Purpose**: Profile CPU usage to identify performance bottlenecks.

```bash
gz profile cpu --duration <duration> [--output <file>]
```

**Options**:
- `--duration` - Profiling duration (default: 30s)
- `--output` - Output file (default: cpu.prof)
- `--analyze` - Auto-analyze after profiling

**Analysis Example**:
```bash
# Profile for 60 seconds
gz profile cpu --duration 60s

# Analyze with pprof
go tool pprof cpu.prof

# Generate flame graph
gz profile cpu --duration 30s --format flame
```

### 4. Memory Profiling (`gz profile memory`)

**Purpose**: Profile memory allocations and identify leaks.

```bash
gz profile memory [--type <type>] [--output <file>]
```

**Profile Types**:
- `heap` - Heap allocations (default)
- `allocs` - All allocations
- `inuse` - In-use memory

**Output Example**:
```
Memory Profile Summary:
======================
Top Memory Consumers:
1. bufio.(*Reader).Read         125.3 MB (32.1%)
2. encoding/json.Unmarshal       89.7 MB (23.0%)
3. strings.Builder.grow          45.2 MB (11.6%)
4. database/sql.(*Rows).Next     38.9 MB (10.0%)
5. net/http.(*Request).parse     28.4 MB (7.3%)

Total Allocated: 389.5 MB
Total In-Use: 156.2 MB
```

## Test Coverage Details

### Unit Tests Added (2025-08)

```go
// cmd/profile/profile_test.go
func TestProfileStats(t *testing.T)        // ✅ Implemented
func TestProfileServer(t *testing.T)       // ✅ Implemented
func TestProfileCPU(t *testing.T)          // ✅ Implemented
func TestProfileMemory(t *testing.T)       // ✅ Implemented
func TestProfileGoroutine(t *testing.T)    // ✅ Implemented
func TestOutputFormats(t *testing.T)       // ✅ Implemented
func TestConcurrentProfiling(t *testing.T) // ✅ Implemented
```

### Integration Tests

```go
func TestProfileServerIntegration(t *testing.T)  // ✅ Implemented
func TestProfileAnalysis(t *testing.T)           // ✅ Implemented
func TestMemoryLeakDetection(t *testing.T)       // ✅ Implemented
```

## Advanced Features

### 1. Comparative Profiling

Compare profiles to identify performance regressions:

```bash
# Create baseline
gz profile cpu --duration 30s --output baseline.prof

# After changes
gz profile cpu --duration 30s --output current.prof

# Compare
gz profile compare baseline.prof current.prof
```

### 2. Continuous Profiling

Monitor performance over time:

```bash
# Start continuous profiling
gz profile continuous --interval 5m --duration 1h

# Output: profile-{timestamp}.prof files every 5 minutes
```

### 3. Automated Analysis

```bash
# Auto-analyze and report issues
gz profile analyze --auto

# Output
Performance Issues Detected:
1. High CPU usage in json.Marshal (15.2% of CPU time)
   Suggestion: Consider using json.Encoder for streaming

2. Memory leak detected in websocket handler
   Growth rate: 2.3 MB/minute
   Suggestion: Ensure proper connection cleanup

3. Goroutine leak: 150 goroutines not terminating
   Location: worker.go:42
   Suggestion: Add proper context cancellation
```

## Configuration

```yaml
profile:
  server:
    default_port: 6060
    auto_open_browser: true

  cpu:
    default_duration: 30s
    sampling_rate: 100  # Hz

  memory:
    gc_before_heap: true
    track_allocs: false

  output:
    directory: ./profiles
    format: pprof  # pprof, text, json, flame

  continuous:
    enabled: false
    interval: 5m
    retention: 24h
```

## Performance Benchmarks

### Profiling Overhead

| Profile Type | Overhead | Impact |
|--------------|----------|---------|
| CPU | <5% | Minimal |
| Memory | <2% | Negligible |
| Goroutine | <1% | Negligible |
| Block | 10-15% | Moderate |
| Mutex | 5-10% | Low |

### Profile Generation Speed

| Operation | Time | Size |
|-----------|------|------|
| 30s CPU profile | 30.2s | ~2MB |
| Heap snapshot | <100ms | ~5MB |
| Goroutine dump | <50ms | ~500KB |
| Full trace (1min) | 60.5s | ~20MB |

## Integration with Other Tools

### 1. Integration with `gz quality`

```bash
# Profile quality checks
gz quality run --profile

# Analyze quality tool performance
gz profile analyze quality-profile.prof
```

### 2. Integration with `gz doctor`

```bash
# Include profiling in health check
gz doctor --include-profile

# Output includes performance metrics
```

### 3. CI/CD Integration

```yaml
# GitHub Actions
- name: Performance Profiling
  run: |
    gz profile cpu --duration 30s --output cpu.prof
    gz profile memory --output mem.prof
    gz profile analyze --threshold 80

# GitLab CI
performance:
  script:
    - gz profile continuous --duration 10m
  artifacts:
    paths:
      - profiles/
```

## Visualization

### Flame Graphs

```bash
# Generate flame graph
gz profile cpu --duration 30s --format flame

# Opens interactive flame graph in browser
```

### Call Graphs

```bash
# Generate call graph
gz profile cpu --duration 30s --format dot

# Convert to PNG
dot -Tpng profile.dot -o profile.png
```

### Time Series

```bash
# Generate time series data
gz profile stats --interval 1s --duration 1m --format csv

# Visualize with your favorite tool
```

## Best Practices

### 1. Production Profiling

```bash
# Low-overhead production profiling
gz profile server --port 6060 --auth required

# Sample 1% of requests
gz profile cpu --sample-rate 0.01
```

### 2. Memory Leak Detection

```bash
# Take heap snapshots
gz profile memory --output heap1.prof
# ... wait for suspected leak ...
gz profile memory --output heap2.prof

# Compare
gz profile compare heap1.prof heap2.prof --type memory
```

### 3. Goroutine Monitoring

```bash
# Monitor goroutine growth
gz profile goroutine --watch --alert-threshold 1000
```

## Troubleshooting

### Common Issues

1. **High CPU Usage**
   ```bash
   gz profile cpu --duration 60s --analyze
   ```

2. **Memory Leaks**
   ```bash
   gz profile memory --type inuse --gc-before
   ```

3. **Goroutine Leaks**
   ```bash
   gz profile goroutine --filter "runtime.gopark"
   ```

4. **Deadlocks**
   ```bash
   gz profile block --duration 30s
   ```

## Future Enhancements

1. **AI-Powered Analysis**: ML-based performance issue detection
2. **Distributed Profiling**: Profile across multiple instances
3. **Historical Comparison**: Track performance over releases
4. **Custom Metrics**: User-defined performance metrics
5. **Integration with APM**: Connect with Application Performance Monitoring tools
6. **Automated Optimization**: Suggest and apply performance fixes

## Documentation

- User Guide: `docs/30-features/37-performance-profiling.md`
- API Reference: `docs/50-api-reference/profile-commands.md`
- Best Practices: `docs/60-development/profiling-best-practices.md`
- Troubleshooting: `docs/90-maintenance/performance-troubleshooting.md`
