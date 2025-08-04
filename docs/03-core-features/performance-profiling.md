# Performance Profiling Guide

The `gz profile` command provides comprehensive performance profiling capabilities using Go's standard pprof tools, helping developers identify bottlenecks and optimize their applications.

## Overview

Performance profiling is essential for understanding how your application uses CPU, memory, and other resources. The `gz profile` command streamlines the profiling workflow by providing an easy-to-use interface over Go's powerful pprof ecosystem.

## Key Features

- **Standard Go pprof Integration**: Uses Go's built-in profiling tools
- **Multiple Profile Types**: CPU, memory, block, mutex, goroutine, and trace profiles
- **Interactive Analysis**: Web-based visualization and command-line tools
- **Production-Safe**: Minimal overhead profiling suitable for production use
- **Profile Management**: Store, compare, and analyze profiles over time

## Command Reference

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
```

### Analyze Profiles

View and analyze collected profiles:

```bash
# List available profiles
gz profile list

# Analyze specific profile
gz profile analyze cpu-20250804-101523.pprof

# Open web interface
gz profile web cpu-profile.pprof

# Compare two profiles
gz profile diff baseline.pprof current.pprof

# Generate report
gz profile report --format text cpu-profile.pprof
```

### Live Profiling

Profile running gz commands:

```bash
# Profile another gz command
gz profile run -- gz synclone github --org myorg

# Profile with specific type
gz profile run --type memory -- gz quality run

# Profile external command
gz profile run --type cpu -- go test ./...
```

## Profile Types

### CPU Profiling

Identifies CPU-intensive operations:

```bash
# Basic CPU profile
gz profile start --type cpu

# High-frequency sampling
gz profile start --type cpu --cpu-rate 1000

# Profile specific duration
gz profile start --type cpu --duration 60s
```

### Memory Profiling

Tracks memory allocations and heap usage:

```bash
# Heap profile
gz profile start --type memory

# Allocation profile
gz profile start --type allocs

# Inuse objects
gz profile start --type inuse
```

### Block Profiling

Detects blocking operations:

```bash
# Enable block profiling
gz profile start --type block

# Set block profile rate
gz profile start --type block --block-rate 100
```

### Mutex Profiling

Identifies mutex contention:

```bash
# Enable mutex profiling
gz profile start --type mutex

# Set mutex profile fraction
gz profile start --type mutex --mutex-rate 100
```

### Goroutine Profiling

Analyzes goroutine usage:

```bash
# Snapshot current goroutines
gz profile snapshot --type goroutine

# Monitor goroutine growth
gz profile monitor --type goroutine --interval 1s
```

### Execution Trace

Records detailed execution traces:

```bash
# Start trace
gz profile trace start

# Stop and analyze
gz profile trace stop
gz profile trace analyze trace.out
```

## Configuration

### Configuration File

Create `~/.config/gzh-manager/profile.yaml`:

```yaml
profile:
  # Default settings
  defaults:
    duration: 30s
    output_dir: "~/.local/share/gzh-manager/profiles"

  # CPU profiling
  cpu:
    rate: 100  # Hz

  # Memory profiling
  memory:
    rate: 512 * 1024  # bytes

  # Block profiling
  block:
    rate: 1

  # Mutex profiling
  mutex:
    fraction: 100

  # Web UI settings
  web:
    port: 6060
    host: "localhost"
    auto_open: true

  # Storage settings
  storage:
    max_profiles: 100
    retention_days: 30
    compress: true
```

### Environment Variables

```bash
# Profile output directory
export GZ_PROFILE_DIR="/path/to/profiles"

# Default profile duration
export GZ_PROFILE_DURATION="60s"

# Web UI settings
export GZ_PROFILE_WEB_PORT="8080"
export GZ_PROFILE_WEB_HOST="0.0.0.0"

# Auto-enable profiling
export GZ_PROFILE_AUTO=true
export GZ_PROFILE_AUTO_TYPE="cpu,memory"
```

## Usage Examples

### Basic Workflow

```bash
# 1. Start profiling your application
$ gz profile start --type cpu
🎯 CPU profiling started
   Output: ~/.local/share/gzh-manager/profiles/cpu-20250804-101523.pprof

# 2. Run your workload
$ gz synclone github --org large-org
📦 Cloning 150 repositories...

# 3. Stop profiling
$ gz profile stop
✅ CPU profiling stopped
   Duration: 45.3s
   Samples: 4,532

# 4. Analyze results
$ gz profile web cpu-20250804-101523.pprof
🌐 Opening profile viewer at http://localhost:6060
```

### Advanced Analysis

```bash
# Top functions by CPU usage
$ gz profile top cpu-profile.pprof
Showing nodes accounting for 12.5s, 85% of 14.7s total

      flat  flat%   sum%        cum   cum%
     3.2s  21.8%  21.8%      4.5s  30.6%  runtime.mallocgc
     2.8s  19.0%  40.8%      2.8s  19.0%  runtime.memmove
     1.5s  10.2%  51.0%      3.2s  21.8%  encoding/json.(*Decoder).Decode

# Generate flame graph
$ gz profile flamegraph cpu-profile.pprof
📊 Generated: cpu-profile-flamegraph.svg

# Memory allocation tracking
$ gz profile analyze --type alloc memory-profile.pprof
Total allocations: 1.2GB
Live objects: 45MB
Allocation rate: 26.7 MB/s
```

### Production Profiling

```bash
# Low-overhead production profiling
$ gz profile start --type cpu --rate 10 --duration 5m

# Profile with labels
$ gz profile start --labels env=prod,service=api

# Continuous profiling
$ gz profile continuous --interval 5m --keep 24
```

### Comparing Profiles

```bash
# Compare before/after optimization
$ gz profile diff before.pprof after.pprof
📊 Performance Comparison:
   CPU usage: -35% ✅
   Memory allocs: -22% ✅
   Goroutines: +5 ⚠️

Top improvements:
1. parseJSON(): -2.1s (45% faster)
2. processData(): -1.5s (30% faster)
3. writeOutput(): -0.8s (25% faster)
```

## Integration with Other Commands

### Profile-Guided Optimization

```bash
# Profile quality checks
gz profile run --type cpu -- gz quality run

# Profile large sync operations
gz profile run --type memory -- gz synclone github --org kubernetes

# Profile with specific focus
gz profile run --type block -- gz net-env transition home
```

### Automated Performance Testing

```bash
# Performance regression testing
gz profile benchmark --baseline v1.0.0 --current HEAD

# CI/CD integration
gz profile ci --threshold cpu=10%,memory=5%
```

## Best Practices

### 1. Profile Selection

- **CPU**: For computation-heavy operations
- **Memory**: For allocation-heavy or memory leak detection
- **Block**: For I/O or synchronization issues
- **Mutex**: For lock contention problems

### 2. Sampling Rates

```bash
# Development (high detail)
gz profile start --type cpu --rate 100

# Production (low overhead)
gz profile start --type cpu --rate 10

# Memory profiling rates
gz profile start --type memory --alloc-rate 1  # Every allocation
gz profile start --type memory --alloc-rate 1000  # Every 1000th
```

### 3. Duration Guidelines

- **CPU profiles**: 30-60 seconds for most cases
- **Memory profiles**: Snapshot at peak usage
- **Block profiles**: During high-contention periods
- **Trace**: 5-10 seconds maximum (large files)

## Troubleshooting

### Common Issues

1. **"No profile data collected"**
   ```bash
   # Ensure profiling is enabled
   gz profile check

   # Verify profile type is supported
   gz profile types
   ```

2. **"Cannot open web interface"**
   ```bash
   # Check if port is in use
   gz profile web --port 8080

   # Use CLI analysis instead
   gz profile top profile.pprof
   ```

3. **"Profile file too large"**
   ```bash
   # Reduce sampling rate
   gz profile start --type cpu --rate 10

   # Shorter duration
   gz profile start --duration 10s
   ```

### Debug Mode

```bash
# Enable debug logging
gz profile --debug start --type cpu

# Verify profiling runtime settings
gz profile runtime

# Test profile collection
gz profile test --type all
```

## Advanced Features

### Custom Profiles

```bash
# Define custom profile
gz profile custom \
  --name "api-endpoints" \
  --type cpu,allocs \
  --filter "api/*"

# Use custom profile
gz profile start --profile api-endpoints
```

### Profile Aggregation

```bash
# Merge multiple profiles
gz profile merge -o combined.pprof *.pprof

# Average profiles
gz profile average -o baseline.pprof profiles/*.pprof
```

### Export Formats

```bash
# Export as text
gz profile export --format text profile.pprof

# Export as JSON
gz profile export --format json profile.pprof

# Export for external tools
gz profile export --format pprof --binary profile.pprof
```

## Performance Tips

1. **Start Simple**: Begin with CPU profiling to identify hot paths
2. **Profile Realistically**: Use production-like workloads
3. **Compare Profiles**: Always compare before/after changes
4. **Focus on Top Functions**: Optimize the top 20% first
5. **Consider Trade-offs**: CPU vs memory optimization
6. **Automate Testing**: Include profiling in CI/CD

## Related Documentation

- [Development Guide](../06-development/)
- [Debugging Guide](../06-development/debugging-guide.md)
- [Testing Strategy](../06-development/testing-strategy.md)
- [Architecture Overview](../02-architecture/overview.md)
