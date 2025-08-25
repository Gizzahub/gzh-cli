# 📊 Performance Profiling Guide

The `gz profile` command provides comprehensive performance profiling capabilities using Go's standard pprof tools, helping developers identify bottlenecks and optimize their applications.

## 📋 Table of Contents

- [Overview](#overview)
- [Profile Types](#profile-types)
- [Command Reference](#command-reference)
- [Analysis Tools](#analysis-tools)
- [Best Practices](#best-practices)
- [Integration Examples](#integration-examples)

## 🎯 Overview

Performance profiling is essential for understanding how your application uses CPU, memory, and other resources. The `gz profile` command streamlines the profiling workflow by providing an easy-to-use interface over Go's powerful pprof ecosystem.

### Key Features

- **Standard Go pprof Integration** - Uses Go's built-in profiling tools
- **Multiple Profile Types** - CPU, memory, block, mutex, goroutine, and trace profiles
- **Interactive Analysis** - Web-based visualization and command-line tools
- **Production-Safe** - Minimal overhead profiling suitable for production use
- **Profile Management** - Store, compare, and analyze profiles over time
- **Real-time Monitoring** - Live profiling with continuous data collection

## 📈 Profile Types

### CPU Profiling

Analyze CPU usage and identify performance bottlenecks:

```bash
# Basic CPU profiling
gz profile start --type cpu

# CPU profiling with custom duration
gz profile start --type cpu --duration 30s

# High-frequency CPU sampling
gz profile start --type cpu --rate 1000
```

### Memory Profiling

Track memory allocation and usage patterns:

```bash
# Heap memory profiling
gz profile start --type memory

# Allocation profiling
gz profile start --type allocs

# Memory with garbage collection info
gz profile start --type memory --gc-info
```

### Concurrency Profiling

Analyze goroutines and synchronization:

```bash
# Goroutine profiling
gz profile start --type goroutine

# Mutex contention profiling
gz profile start --type mutex

# Block profiling
gz profile start --type block
```

### Trace Profiling

Detailed execution tracing:

```bash
# Execution trace
gz profile start --type trace

# Trace with custom duration
gz profile start --type trace --duration 10s
```

## 📖 Command Reference

### Start Profiling

Begin a profiling session:

```bash
# Start CPU profiling
gz profile start --type cpu

# Start memory profiling
gz profile start --type memory

# Start with custom duration
gz profile start --type cpu --duration 30s

# Start multiple profile types
gz profile start --type cpu,memory,goroutine

# Production-safe profiling
gz profile start --type cpu --safe-mode
```

### Stop Profiling

End the current profiling session:

```bash
# Stop all active profiles
gz profile stop

# Stop specific profile type
gz profile stop --type cpu

# Stop and save with custom name
gz profile stop --output my-profile.pprof

# Stop and analyze immediately
gz profile stop --analyze
```

### Server Mode

Run a profiling server for continuous monitoring:

```bash
# Start profiling server
gz profile server

# Custom port and address
gz profile server --port 6060 --bind 0.0.0.0

# Server with authentication
gz profile server --auth-token secret123

# Background server
gz profile server --daemon
```

### Profile Management

Manage saved profiles:

```bash
# List available profiles
gz profile list

# Show profile information
gz profile info cpu-profile.pprof

# Compare two profiles
gz profile diff profile1.pprof profile2.pprof

# Delete old profiles
gz profile cleanup --older-than 7d
```

## 🔍 Analysis Tools

### Interactive Analysis

```bash
# Analyze profile interactively
gz profile analyze cpu-profile.pprof

# Web-based analysis
gz profile web cpu-profile.pprof

# Command-line analysis
gz profile cli memory-profile.pprof
```

### Report Generation

```bash
# Generate text report
gz profile report cpu-profile.pprof

# Generate flame graph
gz profile flamegraph cpu-profile.pprof

# Generate call graph
gz profile callgraph memory-profile.pprof

# Export to SVG
gz profile svg cpu-profile.pprof --output cpu-usage.svg
```

### Performance Statistics

```bash
# Show performance stats
gz profile stats

# CPU usage summary
gz profile stats --type cpu

# Memory usage trends
gz profile stats --type memory --trend

# Export statistics
gz profile stats --output json > perf-stats.json
```

## ⚙️ Configuration

### Basic Configuration

Add profiling settings to your `~/.config/gzh-manager/gzh.yaml`:

```yaml
commands:
  profile:
    # Default profile type
    default_type: cpu

    # Default duration
    default_duration: "30s"

    # Output directory
    output_dir: "$HOME/.config/gzh-manager/profiles"

    # Server settings
    server:
      port: 6060
      bind: "localhost"
      enable_auth: false

    # Automatic cleanup
    cleanup:
      enabled: true
      retention_days: 14
      max_profiles: 50
```

### Advanced Configuration

```yaml
commands:
  profile:
    # Profile-specific settings
    cpu:
      sample_rate: 100  # Hz
      duration: "30s"

    memory:
      sample_rate: 512  # KB
      gc_before: true

    goroutine:
      debug_level: 1

    # Analysis settings
    analysis:
      web_port: 8080
      auto_open_browser: true

    # Integration settings
    integration:
      prometheus:
        enabled: true
        endpoint: "/metrics"

      grafana:
        enabled: false
        dashboard_url: "http://grafana:3000"
```

## 🎯 Best Practices

### Production Profiling

```bash
# Safe production profiling
gz profile start --type cpu --duration 15s --safe-mode

# Low-overhead memory profiling
gz profile start --type memory --rate 512KB

# Background profiling with alerts
gz profile server --daemon --alert-threshold 80%
```

### Development Workflow

```bash
# Profile during development
gz profile start --type cpu,memory &
# Run your application
gz profile stop --analyze

# Continuous profiling
gz profile server --auto-profile
```

### Performance Regression Detection

```bash
# Baseline profiling
gz profile start --type cpu --baseline

# Compare against baseline
gz profile diff baseline.pprof current.pprof

# Automated regression detection
gz profile regression-check --threshold 10%
```

## 🔧 Analysis Examples

### CPU Bottleneck Analysis

```bash
# Identify CPU hotspots
gz profile analyze cpu.pprof --focus hot

# Function-level analysis
gz profile top cpu.pprof --functions

# Source code view
gz profile list cpu.pprof "function_name"
```

### Memory Leak Detection

```bash
# Memory allocation analysis
gz profile analyze memory.pprof --inuse_space

# Find potential leaks
gz profile diff memory-start.pprof memory-end.pprof

# Allocation patterns
gz profile top memory.pprof --alloc_objects
```

### Concurrency Analysis

```bash
# Goroutine analysis
gz profile analyze goroutine.pprof

# Mutex contention
gz profile top mutex.pprof --contentions

# Blocking operations
gz profile analyze block.pprof --delay
```

## 🚀 Integration Examples

### With CI/CD Pipelines

```yaml
# GitHub Actions example
- name: Performance Profiling
  run: |
    gz profile start --type cpu --duration 60s &
    # Run performance tests
    gz profile stop --output perf-results.pprof
    gz profile report perf-results.pprof --format json > performance.json
```

### With Monitoring Systems

```bash
# Prometheus integration
gz profile server --prometheus-endpoint /metrics

# Custom metrics export
gz profile stats --format prometheus | curl -X POST \
  --data-binary @- http://pushgateway:9091/metrics/job/gzh-profiling
```

### With Load Testing

```bash
# Profile during load test
gz profile start --type cpu,memory &
# Run load test (wrk, ab, etc.)
LOAD_TEST_PID=$!
sleep 60
gz profile stop --output load-test-profile.pprof
kill $LOAD_TEST_PID
gz profile analyze load-test-profile.pprof
```

## 📊 Visualization Options

### Flame Graphs

```bash
# Generate flame graph
gz profile flamegraph cpu.pprof --output flame.svg

# Interactive flame graph
gz profile web cpu.pprof --flamegraph
```

### Call Graphs

```bash
# Generate call graph
gz profile callgraph memory.pprof --output calls.svg

# Focus on specific function
gz profile callgraph cpu.pprof --focus "main.*"
```

### Timeline Views

```bash
# Trace timeline
gz profile trace-view trace.pprof

# Memory timeline
gz profile timeline memory.pprof --metric alloc_space
```

## 🆘 Troubleshooting

### Common Issues

#### No Profile Data

```bash
# Check if profiling is enabled
gz profile status

# Verify profile generation
gz profile test --duration 5s

# Debug profiling setup
gz profile debug
```

#### Performance Impact

```bash
# Monitor profiling overhead
gz profile overhead

# Use safe mode for production
gz profile start --safe-mode

# Adjust sampling rates
gz profile start --type cpu --rate 10  # Lower rate
```

#### Analysis Problems

```bash
# Validate profile file
gz profile validate profile.pprof

# Re-analyze with different settings
gz profile analyze profile.pprof --nodecount 100

# Export raw data
gz profile raw profile.pprof --format json
```

## 📋 Output Formats

All profiling commands support multiple output formats:

```bash
# JSON output for automation
gz profile stats --output json

# CSV for spreadsheet analysis
gz profile report cpu.pprof --output csv

# HTML for web viewing
gz profile report memory.pprof --output html

# Plain text for console
gz profile top cpu.pprof --output text
```

## 🔗 Related Commands

### Integration with Other gzh-cli Features

```bash
# Profile during repository sync
gz synclone github --org myorg &
gz profile start --type cpu
# Wait for sync to complete
gz profile stop --output synclone-profile.pprof

# Profile IDE monitoring
gz ide monitor &
gz profile server --daemon
```

______________________________________________________________________

**Profile Types**: CPU, Memory, Goroutine, Mutex, Block, Trace
**Output Formats**: pprof, SVG, JSON, CSV, HTML
**Integration**: Go pprof, Prometheus, Grafana
**Production Safe**: Minimal overhead profiling
