# ADR-001: Architecture Simplification - Container and Profiling System Removal

## Status
Accepted and Implemented

## Date
2025-01-04

## Context
The gzh-manager-go CLI tool had accumulated over-engineered components that added unnecessary complexity for a command-line application:

1. **Dependency Injection Container System** (`internal/container/`)
   - Complex builder pattern with 1,188+ lines of code
   - Sophisticated validation and lifecycle management
   - More appropriate for long-running server applications

2. **Custom Profiling System** (`cmd/profile/` and related packages)
   - Custom HTTP server implementation
   - Complex abstraction layers
   - Duplicated functionality available in standard Go pprof

## Decision
Remove over-engineered components and replace with simpler, more appropriate alternatives:

### 1. Container System Removal
- **Removed**: Entire `internal/container/` package (1,188 lines)
- **Replaced with**: Direct constructor calls in command initialization
- **Files removed**:
  - `internal/container/builder.go`
  - `internal/container/container.go` 
  - `internal/container/container_test.go`
  - `internal/container/validation.go`

### 2. Profiling System Simplification
- **Removed**: Complex custom profiling HTTP server
- **Replaced with**: Standard Go pprof integration via `internal/simpleprof/`
- **New implementation**:
  - Direct use of `runtime/pprof` and `net/http/pprof`
  - Simplified API with essential profiling features
  - Compatible with standard Go tooling

### 3. Performance Monitoring
- **Added**: Automated performance benchmarking system
- **Scripts**:
  - `scripts/simple-benchmark.sh` - Quick performance checks
  - `scripts/benchmark-performance.sh` - Comprehensive benchmarking with baselines
- **Features**:
  - Startup time monitoring (50ms threshold)
  - Binary size tracking
  - Memory usage analysis
  - Regression detection

## Consequences

### Positive
- **Reduced complexity**: ~1,200+ lines of code removed
- **Improved maintainability**: Simpler, more direct code paths
- **Better performance monitoring**: Automated regression detection
- **Standard tooling compatibility**: Uses Go's built-in pprof tools
- **Reduced binary size**: Eliminated unnecessary abstractions
- **Faster development**: Less cognitive overhead for new contributors

### Negative
- **Lost flexibility**: No longer supports complex dependency graphs
- **Migration effort**: Required updating command initialization code
- **Documentation updates**: Need to reflect new simplified architecture

## Implementation Details

### Container Removal
```go
// Before (container-based)
syncCloneCmd := synclone.NewSyncCloneCmdWithContainer(container)

// After (direct constructor)
syncCloneCmd := synclone.NewSyncCloneCmd()
```

### Profiling Simplification
```go
// Before: Complex custom HTTP server with multiple abstractions
// After: Direct pprof integration
profiler := simpleprof.NewSimpleProfiler("tmp/profiles")
profiler.StartHTTPServer(port)
```

### Performance Verification
- All tests continue to pass
- Binary size maintained at ~33MB  
- Startup time remains under 10ms
- Memory usage patterns unchanged

## References
- Original analysis: `/tasks/refactoring/unused-code-analysis.md`
- Architecture evaluation: `/tasks/refactoring/architecture-simplification.md`
- Action plan: `/tasks/refactoring/cleanup-action-plan.md`
- Performance scripts: `scripts/simple-benchmark.sh`, `scripts/benchmark-performance.sh`

## Related Changes
- Updated `cmd/root.go` to use direct constructors
- Created `internal/simpleprof/` package for simplified profiling
- Added performance monitoring scripts
- Removed unused imports and dependencies