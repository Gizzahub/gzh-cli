# Debugging Guide

This guide provides comprehensive debugging configurations and instructions for the GZH Manager Go project across different IDEs and debugging tools.

## Overview

The project includes debugging configurations for:

- **VS Code** - Complete debug configurations and tasks
- **JetBrains IDEs** (GoLand/IntelliJ IDEA) - Run configurations
- **Delve** - Go-specific debugging with command-line tools
- **Command Line** - Debug scripts for various scenarios

## VS Code Debugging

### Available Debug Configurations

Access via the Debug panel (`Ctrl+Shift+D`) or Command Palette (`F5`):

| Configuration | Purpose | Arguments |
| ----------------------------------- | ---------------------------------- | -------------------------------------------------------------------- |
| **Debug GZH CLI** | Main application help | `--help` |
| **Debug GZH Synclone** | Repository sync with sample config | `synclone --config examples/synclone/synclone-simple.yaml --dry-run` |
| **Debug GZH Config Validate** | Configuration validation | `synclone validate --config examples/synclone/synclone-example.yaml` |
| **Debug Current Go File** | Debug the currently open file | N/A |
| **Debug Go Test (Current Package)** | Test debugging | Test files in current directory |
| **Debug Go Test (Current File)** | Specific test function | Prompts for test name |
| **Attach to Running Process** | Attach to running process | Prompts for process selection |

### Quick Start

1. **Set breakpoints** by clicking in the gutter (left of line numbers)
1. **Select debug configuration** from the dropdown
1. **Press F5** or click the green play button
1. **Use debug controls**:
   - `F5` - Continue
   - `F10` - Step Over
   - `F11` - Step Into
   - `Shift+F11` - Step Out
   - `Ctrl+Shift+F5` - Restart
   - `Shift+F5` - Stop

### VS Code Tasks

Access via Command Palette (`Ctrl+Shift+P` → "Tasks: Run Task"):

#### Build Tasks

- **go: build** - Build the gz binary
- **go: clean** - Clean build artifacts
- **go: bootstrap** - Install dependencies
- **go: format** - Format code
- **go: lint** - Run linting
- **go: security scan** - Run security analysis

#### Test Tasks

- **go: test** - Run all tests
- **go: test (current package)** - Test current package
- **docker: test integration** - Run integration tests

#### Development Tasks

- **react: start dev server** - Start React development
- **react: build** - Build React app
- **docker: build image** - Build Docker image

### Environment Variables

All debug configurations include:

```json
{
  "env": {
    "GZH_DEV_MODE": "true",
    "GO111MODULE": "on"
  }
}
```

## JetBrains IDEs (GoLand/IntelliJ IDEA)

### Available Run Configurations

Pre-configured run configurations in `.idea/runConfigurations/`:

- **Debug GZH CLI** - Main application debugging
- **Debug GZH Bulk Clone** - Bulk clone functionality
- **Build and Test** - Makefile-based build and test
- **Go Tests** - All Go tests with proper environment

### Usage

1. **Open project** in GoLand/IntelliJ IDEA
1. **Select configuration** from the run configuration dropdown
1. **Set breakpoints** by clicking in the gutter
1. **Click the debug button** (bug icon) or press `Shift+F9`
1. **Use debug controls** in the debug panel

### Creating Custom Configurations

1. Go to **Run** → **Edit Configurations**
1. Click **+** → **Go Application**
1. Configure:
   - **Name**: Your configuration name
   - **Run kind**: Package
   - **Package**: `github.com/gizzahub/gzh-cli`
   - **Working directory**: Project root
   - **Program arguments**: Your command arguments
   - **Environment variables**: `GZH_DEV_MODE=true;GO111MODULE=on`

## Delve Debugger

### Configuration File

Delve configuration in `.delve/config.yml` includes:

- **Enhanced output**: Colors and detailed information
- **Useful aliases**: Shortcuts for common commands
- **Environment setup**: Development-specific variables
- **Performance settings**: Optimizations disabled for debugging

### Command Line Usage

#### Basic Debugging

```bash
# Debug main application
dlv debug main.go -- --help

# Debug with arguments
dlv debug main.go -- synclone --config examples/synclone/synclone-simple.yaml

# Debug tests
dlv test ./cmd/synclone

# Debug specific test
dlv test ./cmd/synclone -- -test.run TestSynclone
```

#### Remote Debugging

```bash
# Start headless debug server
dlv debug --listen=:2345 --headless --api-version=2 main.go -- --help

# Connect from another terminal
dlv connect :2345
```

#### Attach to Running Process

```bash
# Find process
ps aux | grep gz

# Attach to process
dlv attach <pid>
```

### Delve Commands

| Command | Description | Example |
| ---------------- | ----------------------- | ----------------------------- |
| `b <location>` | Set breakpoint | `b main.main`, `b main.go:42` |
| `c` | Continue | `c` |
| `n` | Next line | `n` |
| `s` | Step into | `s` |
| `so` | Step out | `so` |
| `p <var>` | Print variable | `p myVar` |
| `pp <var>` | Pretty print | `pp complexStruct` |
| `locals` | Show local variables | `locals` |
| `args` | Show function arguments | `args` |
| `vars` | Show package variables | `vars` |
| `bt` | Stack trace | `bt` |
| `goroutines` | List goroutines | `goroutines` |
| `goroutine <id>` | Switch to goroutine | `goroutine 5` |
| `list` | Show source code | `list`, `list main.main` |
| `edit <file>` | Open file in editor | `edit main.go` |
| `restart` | Restart program | `restart` |
| `quit` | Exit debugger | `quit` |

## Debug Scripts

Convenience scripts in `scripts/debug/`:

### 1. CLI Debugging

```bash
# Debug with default help command
./scripts/debug/debug-cli.sh

# Debug with specific command
./scripts/debug/debug-cli.sh synclone --config examples/synclone/synclone-simple.yaml

```

**Features:**

- Automatic debug symbol building
- Web UI on http://127.0.0.1:2345
- Helpful command reference
- Environment setup

### 2. Test Debugging

```bash
# Debug all tests in package
./scripts/debug/debug-test.sh ./cmd/synclone

# Debug specific test
./scripts/debug/debug-test.sh ./cmd/synclone TestSynclone

# Debug with pattern
./scripts/debug/debug-test.sh ./pkg/github "Test.*Config"
```

**Features:**

- Package validation
- Test filtering
- Web UI on http://127.0.0.1:2346
- Verbose test output

### 3. Process Attachment

```bash
# Auto-detect and attach
./scripts/debug/debug-attach.sh

# Attach to specific PID
./scripts/debug/debug-attach.sh 12345
```

**Features:**

- Process discovery
- Interactive selection
- Web UI on http://127.0.0.1:2347
- Non-invasive attachment

## Debugging Strategies

### 1. Application Flow Debugging

**Goal**: Understand command execution flow

```bash
# Set breakpoints at key locations
b main.main
b cmd.Execute
b cmd/root.go:Execute

# Step through initialization
s
n

# Inspect configuration
p config
pp viper.AllSettings()
```

### 2. API Debugging

**Goal**: Debug HTTP API endpoints

```bash
# Start API server in debug mode
./scripts/debug/debug-cli.sh serve --port 8080

# Set breakpoints on handlers
b pkg/github/webhook.go:HandleWebhook
b cmd/serve/serve.go:setupRoutes

# Make requests and debug
curl http://localhost:8080/api/health
```

### 3. Concurrency Debugging

**Goal**: Debug goroutines and race conditions

```bash
# Enable race detector
export GORACE="halt_on_error=1"
go run -race main.go

# In debugger, inspect goroutines
goroutines
goroutine 5
bt

# Check for data races
p runtime.NumGoroutine()
```

### 4. Configuration Debugging

**Goal**: Debug configuration loading and validation

```bash
# Debug config validation
./scripts/debug/debug-cli.sh bulk-clone validate --config examples/bulk-clone-example.yaml

# Set breakpoints in config package
b pkg/config/loader.go:LoadConfig
b pkg/config/validator.go:Validate

# Inspect configuration state
p config
pp configErrors
```

### 5. Test Debugging

**Goal**: Debug failing tests

```bash
# Debug specific test
./scripts/debug/debug-test.sh ./cmd/synclone TestConfigValidation

# Set breakpoints in test
b bulk_clone_test.go:TestConfigValidation
b bulk_clone_test.go:42

# Inspect test data
p testConfig
pp result
```

## Debugging Different Components

### Go CLI Application

```bash
# Main application
dlv debug main.go -- bulk-clone --help

# Specific commands
dlv debug main.go -- synclone --config examples/synclone/synclone-simple.yaml --dry-run
dlv debug main.go -- serve --port 8080
```

### React Dashboard

```bash
# Start React dev server with debugging
cd web
npm start

# Access React DevTools in browser
# Debug in Chrome DevTools (F12)
```

### Integration Tests

```bash
# Debug Docker integration tests
dlv test ./test/integration/docker

# Debug with testcontainers
export TESTCONTAINERS_RYUK_DISABLED=true
dlv test ./test/integration/testcontainers
```

## Troubleshooting

### Common Issues

#### 1. "could not launch process: fork/exec: no such file or directory"

**Solution**: Build with debug symbols first

```bash
go build -gcflags="-N -l" -o gz-debug main.go
dlv exec gz-debug -- --help
```

#### 2. "API server listening at: \[::\]:2345, but not accessible"

**Solution**: Check firewall and use specific interface

```bash
dlv debug --listen=127.0.0.1:2345 --headless main.go
```

#### 3. "breakpoint not hit in optimized code"

**Solution**: Disable optimizations

```bash
go build -gcflags="-N -l" main.go
```

#### 4. "goroutine stack exceeds limit"

**Solution**: Increase stack size

```bash
export GODEBUG=asyncpreemptoff=1
dlv debug main.go
```

#### 5. "permission denied when attaching to process"

**Solution**: Run with appropriate permissions

```bash
# Linux: Use sudo or add capability
sudo dlv attach <pid>

# macOS: Sign the binary or disable SIP
codesign -s - -f --entitlements=debug.entitlements dlv
```

### Performance Issues

#### Slow Debugging

1. **Disable unnecessary breakpoints**
1. **Use conditional breakpoints**:
   ```
   b main.go:42 if myVar > 100
   ```
1. **Limit variable inspection**:
   ```
   config max-string-len 50
   config max-array-values 10
   ```

#### Memory Issues

1. **Monitor memory usage**:
   ```
   p runtime.MemStats
   call runtime.GC()
   ```
1. **Use heap profiling**:
   ```
   import _ "net/http/pprof"
   go tool pprof http://localhost:6060/debug/pprof/heap
   ```

### IDE-Specific Issues

#### VS Code

1. **Go extension not working**:

   - Restart Go language server: `Ctrl+Shift+P` → "Go: Restart Language Server"
   - Check Go tools: `Ctrl+Shift+P` → "Go: Install/Update Tools"

1. **Debugger not stopping at breakpoints**:

   - Check `launch.json` configuration
   - Verify file paths are correct
   - Ensure build is not optimized

#### GoLand/IntelliJ

1. **Run configuration not found**:

   - Reimport project
   - Check `.idea/runConfigurations/` directory
   - Create new configuration manually

1. **Source code not showing**:

   - Check source path mappings
   - Verify module settings
   - Rebuild project

## Best Practices

### 1. Debugging Preparation

- **Use debug builds**: Always build with `-gcflags="-N -l"`
- **Set meaningful breakpoints**: Focus on key decision points
- **Prepare test data**: Use sample configurations and test cases
- **Document issues**: Keep notes of debugging sessions

### 2. Efficient Debugging

- **Start with logs**: Check application logs first
- **Use conditional breakpoints**: Avoid stopping on every iteration
- **Inspect variable state**: Use `p`, `pp`, and `locals` commands
- **Follow execution path**: Use step commands strategically

### 3. Collaborative Debugging

- **Share debug configurations**: Commit IDE configurations to repo
- **Use remote debugging**: Share debug sessions with team
- **Document findings**: Update code comments and documentation
- **Create regression tests**: Add tests for debugged issues

### 4. Production Debugging

- **Use observability tools**: Logs, metrics, and tracing
- **Enable pprof endpoints**: For runtime profiling
- **Use core dumps**: For post-mortem analysis
- **Avoid production debugging**: Use staging environments

## Advanced Debugging Techniques

### 1. Memory Debugging

```bash
# Heap profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Memory allocation profiling
go tool pprof http://localhost:6060/debug/pprof/allocs

# In delve
p runtime.MemStats
call runtime.GC()
```

### 2. CPU Profiling

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 3. Race Detection

```bash
# Build with race detector
go build -race main.go

# Run with race detection
export GORACE="halt_on_error=1"
./main
```

### 4. Tracing

```bash
# Execution tracing
go tool trace trace.out

# In code
import "runtime/trace"
trace.Start(os.Stderr)
defer trace.Stop()
```

## Resources

- [Delve Documentation](https://github.com/go-delve/delve/tree/master/Documentation)
- [VS Code Go Extension](https://github.com/golang/vscode-go)
- [GoLand Debugging Guide](https://www.jetbrains.com/help/go/debugging-code.html)
- [Go Debugging Best Practices](https://golang.org/doc/gdb)
- [pprof Documentation](https://golang.org/pkg/net/http/pprof/)

______________________________________________________________________

**Note**: This debugging setup is specifically configured for the GZH Manager Go project. Adjust configurations as needed for your specific debugging requirements.
