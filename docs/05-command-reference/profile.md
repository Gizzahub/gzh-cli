# profile Command Reference

Performance profiling and analysis using Go's standard pprof tools.

## Synopsis

```bash
gz profile <action> [flags]
```

## Description

The `profile` command provides comprehensive performance profiling capabilities, helping identify bottlenecks and optimize application performance.

## Profile Types

- **CPU** - CPU usage profiling
- **Memory** - Heap and memory allocation profiling
- **Block** - Blocking operations profiling
- **Mutex** - Mutex contention profiling
- **Goroutine** - Goroutine analysis
- **Trace** - Execution trace analysis

## Actions

### `gz profile start`

Start profiling session.

```bash
gz profile start --type <type> [flags]
```

**Flags:**
- `--type` - Profile type: cpu, memory, block, mutex, goroutine (required)
- `--duration` - Profile duration (default: 30s)
- `--output` - Output file name
- `--cpu-rate` - CPU profile rate in Hz (default: 100)
- `--memory-rate` - Memory profile rate in bytes (default: 512KB)

**Examples:**
```bash
# Start CPU profiling
gz profile start --type cpu

# Memory profiling for 60 seconds
gz profile start --type memory --duration 60s

# High-frequency CPU profiling
gz profile start --type cpu --cpu-rate 1000
```

### `gz profile stop`

Stop active profiling session.

```bash
gz profile stop [flags]
```

**Flags:**
- `--type` - Specific profile type to stop
- `--output` - Custom output file name

### `gz profile analyze`

Analyze profile data.

```bash
gz profile analyze <profile-file> [flags]
```

**Arguments:**
- `profile-file` - Profile file to analyze (required)

**Flags:**
- `--format` - Output format: text, json, svg, pdf
- `--top` - Show top N functions (default: 10)
- `--output` - Output file name

**Examples:**
```bash
# Analyze CPU profile
gz profile analyze cpu-profile.pprof

# Generate flame graph
gz profile analyze cpu-profile.pprof --format svg --output flamegraph.svg

# Show top 20 functions
gz profile analyze cpu-profile.pprof --top 20
```

### `gz profile web`

Open profile in web interface.

```bash
gz profile web <profile-file> [flags]
```

**Flags:**
- `--port` - Web server port (default: 6060)
- `--host` - Web server host (default: localhost)
- `--auto-open` - Automatically open browser (default: true)

### `gz profile diff`

Compare two profiles.

```bash
gz profile diff <baseline> <current> [flags]
```

**Arguments:**
- `baseline` - Baseline profile file
- `current` - Current profile file

**Flags:**
- `--output` - Output format: text, json, html
- `--metric` - Comparison metric: cpu, memory, samples

### `gz profile run`

Profile a command execution.

```bash
gz profile run --type <type> -- <command> [args...]
```

**Flags:**
- `--type` - Profile type
- `--output` - Profile output file

**Examples:**
```bash
# Profile gz synclone command
gz profile run --type cpu -- gz synclone github --org myorg

# Profile memory usage
gz profile run --type memory -- gz quality run
```

### `gz profile list`

List available profiles.

```bash
gz profile list [flags]
```

**Flags:**
- `--type` - Filter by profile type
- `--recent` - Show only recent profiles
- `--output` - Output format: table, json

### `gz profile trace`

Execution trace profiling.

```bash
gz profile trace <action> [flags]
```

**Actions:**
- `start` - Start trace recording
- `stop` - Stop trace recording
- `analyze` - Analyze trace file

## Configuration

```yaml
version: "1.0"

defaults:
  duration: "30s"
  output_dir: "~/.local/share/gzh-manager/profiles"

cpu:
  rate: 100

memory:
  rate: 524288  # 512KB

web:
  port: 6060
  host: "localhost"
  auto_open: true

storage:
  max_profiles: 100
  retention_days: 30
  compress: true
```

## Examples

### Basic Profiling Workflow

```bash
# 1. Start CPU profiling
gz profile start --type cpu --duration 60s

# 2. Run your workload
gz synclone github --org large-org

# 3. Analyze results
gz profile web cpu-profile.pprof
```

### Performance Comparison

```bash
# Create baseline
gz profile run --type cpu --output baseline.pprof -- gz quality run

# Make optimizations...

# Compare performance
gz profile run --type cpu --output optimized.pprof -- gz quality run
gz profile diff baseline.pprof optimized.pprof
```

### Memory Analysis

```bash
# Memory profiling
gz profile start --type memory --duration 30s

# Generate heap analysis
gz profile analyze memory-profile.pprof --format text
```

## Related Commands

- [`gz synclone`](synclone.md) - Repository synchronization
- [`gz quality`](quality.md) - Code quality management

## See Also

- [Performance Profiling Guide](../03-core-features/performance-profiling.md)
