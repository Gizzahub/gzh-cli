# Debug Scripts

This directory contains debugging scripts for the GZH Manager Go project. These scripts provide convenient ways to debug various aspects of the application using the Delve debugger.

## Available Scripts

### 1. `debug-cli.sh` - CLI Application Debugging

**Purpose**: Debug the main GZH CLI application with any command and arguments.

**Usage**:
```bash
# Debug with default help command
./scripts/debug/debug-cli.sh

# Debug specific commands
./scripts/debug/debug-cli.sh bulk-clone --config samples/bulk-clone-simple.yaml --dry-run
./scripts/debug/debug-cli.sh monitoring start --port 8080
./scripts/debug/debug-cli.sh serve --port 8080 --static-dir web/build
./scripts/debug/debug-cli.sh ide monitor --interval 5s
```

**Features**:
- Builds with debug symbols automatically
- Sets up proper environment variables
- Starts headless Delve server on port 2345
- Provides web UI at http://127.0.0.1:2345
- Shows helpful debug command reference
- Automatic cleanup on exit

### 2. `debug-test.sh` - Test Debugging

**Purpose**: Debug Go tests in specific packages or functions.

**Usage**:
```bash
# Debug all tests in a package
./scripts/debug/debug-test.sh ./cmd/bulk-clone

# Debug specific test function
./scripts/debug/debug-test.sh ./cmd/bulk-clone TestBulkClone

# Debug with pattern matching
./scripts/debug/debug-test.sh ./pkg/github "Test.*Config"

# Debug integration tests
./scripts/debug/debug-test.sh ./test/integration/docker
```

**Features**:
- Package validation and existence checking
- Test filtering with regex patterns
- Starts headless Delve server on port 2346
- Provides web UI at http://127.0.0.1:2346
- Verbose test output for better debugging
- Proper test environment setup

### 3. `debug-attach.sh` - Process Attachment

**Purpose**: Attach the debugger to an already running GZH process.

**Usage**:
```bash
# Auto-detect and attach to GZH processes
./scripts/debug/debug-attach.sh

# Attach to specific process ID
./scripts/debug/debug-attach.sh 12345
```

**Features**:
- Automatic GZH process discovery
- Interactive process selection for multiple processes
- Process validation and error handling
- Starts headless Delve server on port 2347
- Provides web UI at http://127.0.0.1:2347
- Non-invasive attachment (process continues after detach)

## Common Debugging Workflows

### 1. Debugging Command Execution

```bash
# Start debugging session
./scripts/debug/debug-cli.sh bulk-clone --config samples/bulk-clone-simple.yaml

# In another terminal, connect to debugger
dlv connect 127.0.0.1:2345

# Set breakpoints and step through
(dlv) b main.main
(dlv) c
(dlv) n
(dlv) s
```

### 2. Debugging Failed Tests

```bash
# Debug failing test
./scripts/debug/debug-test.sh ./cmd/bulk-clone TestConfigValidation

# Connect and debug
dlv connect 127.0.0.1:2346

# Set test-specific breakpoints
(dlv) b bulk_clone_test.go:TestConfigValidation
(dlv) c
```

### 3. Debugging Running Services

```bash
# Start service in background
./gz monitoring start --port 8080 &

# Attach debugger
./scripts/debug/debug-attach.sh

# Debug live service
dlv connect 127.0.0.1:2347
```

## Web UI Access

All debug scripts provide web UI access for easier debugging:

- **CLI Debugging**: http://127.0.0.1:2345
- **Test Debugging**: http://127.0.0.1:2346  
- **Process Attachment**: http://127.0.0.1:2347

### Web UI Features

- **Source Code View**: Browse and set breakpoints
- **Variable Inspection**: View local and global variables
- **Stack Traces**: Navigate call stack
- **Goroutine View**: Monitor concurrent execution
- **Step Controls**: Step through code execution
- **Console**: Execute debugger commands

## Environment Setup

All scripts automatically set up the debugging environment:

```bash
export GZH_DEV_MODE=true
export GO111MODULE=on
export GORACE="halt_on_error=1"
```

## Delve Commands Reference

### Basic Commands

| Command | Description | Example |
|---------|-------------|--------|
| `c` | Continue execution | `c` |
| `n` | Next line (step over) | `n` |
| `s` | Step into function | `s` |
| `so` | Step out of function | `so` |
| `r` | Restart program | `r` |
| `q` | Quit debugger | `q` |

### Breakpoints

| Command | Description | Example |
|---------|-------------|--------|
| `b <location>` | Set breakpoint | `b main.main`, `b main.go:42` |
| `bp` | List breakpoints | `bp` |
| `clear <id>` | Clear breakpoint | `clear 1` |
| `clearall` | Clear all breakpoints | `clearall` |
| `on <id> <cmd>` | Execute command on breakpoint | `on 1 p myVar` |

### Variable Inspection

| Command | Description | Example |
|---------|-------------|--------|
| `p <var>` | Print variable | `p myVar` |
| `pp <var>` | Pretty print | `pp complexStruct` |
| `locals` | Show local variables | `locals` |
| `args` | Show function arguments | `args` |
| `vars` | Show package variables | `vars` |
| `whatis <var>` | Show variable type | `whatis myVar` |

### Navigation

| Command | Description | Example |
|---------|-------------|--------|
| `bt` | Stack trace | `bt` |
| `up` | Move up stack frame | `up` |
| `down` | Move down stack frame | `down` |
| `frame <n>` | Jump to frame | `frame 2` |
| `list` | Show source code | `list`, `list main.main` |
| `disassemble` | Show assembly | `disassemble main.main` |

### Goroutines

| Command | Description | Example |
|---------|-------------|--------|
| `goroutines` | List all goroutines | `goroutines` |
| `goroutine <id>` | Switch to goroutine | `goroutine 5` |
| `goroutine <id> bt` | Goroutine stack trace | `goroutine 5 bt` |

## Troubleshooting

### Common Issues

#### 1. "Permission denied" when attaching

**Linux**: Use sudo or add ptrace capability
```bash
sudo ./scripts/debug/debug-attach.sh
# OR
sudo setcap cap_sys_ptrace+ep $(which dlv)
```

**macOS**: Code sign Delve or disable SIP
```bash
codesign -s - -f --entitlements=debug.entitlements $(which dlv)
```

#### 2. "Connection refused" on web UI

**Solution**: Check if debugger is running and ports are available
```bash
# Check if debugger is running
ps aux | grep dlv

# Check port availability
netstat -tlnp | grep :2345

# Use different port if needed
dlv debug --listen=:2348 main.go
```

#### 3. "Breakpoints not hit"

**Solution**: Ensure debug symbols are present
```bash
# Check if built with debug symbols
go build -gcflags="-N -l" main.go
file gz  # Should show "not stripped"

# Verify breakpoint location
(dlv) b main.go:42
(dlv) bp  # List breakpoints
```

#### 4. "Source code not found"

**Solution**: Check source path mapping
```bash
# Verify working directory
(dlv) pwd

# Set source path if needed
(dlv) config substitute-path add /old/path /new/path
```

### Performance Issues

#### Slow debugging

1. **Limit variable inspection**:
   ```bash
   (dlv) config max-string-len 50
   (dlv) config max-array-values 10
   ```

2. **Use conditional breakpoints**:
   ```bash
   (dlv) b main.go:42 if myVar > 100
   ```

3. **Disable logging**:
   ```bash
   dlv debug --log=false main.go
   ```

## Script Customization

### Environment Variables

Customize script behavior with environment variables:

```bash
# Change debug ports
export DEBUG_CLI_PORT=3345
export DEBUG_TEST_PORT=3346
export DEBUG_ATTACH_PORT=3347

# Custom delve flags
export DLV_FLAGS="--log=false --check-go-version=false"

# Custom build flags
export BUILD_FLAGS="-gcflags='-N -l -dwarf=false'"
```

### Custom Scripts

Create custom debug scripts for specific scenarios:

```bash
#!/bin/bash
# scripts/debug/debug-monitoring.sh

./scripts/debug/debug-cli.sh monitoring start \
  --port 8080 \
  --metrics-port 9090 \
  --log-level debug
```

## Integration with IDEs

### VS Code

Use debug scripts with VS Code tasks:

```json
{
  "label": "Debug CLI with Script",
  "type": "shell",
  "command": "./scripts/debug/debug-cli.sh",
  "args": ["${input:debugArgs}"],
  "group": "build"
}
```

### GoLand/IntelliJ

Create external tools for debug scripts:

1. **Settings** → **Tools** → **External Tools**
2. **Add new tool**:
   - **Name**: Debug CLI
   - **Program**: `$ProjectFileDir$/scripts/debug/debug-cli.sh`
   - **Arguments**: `$Prompt$`
   - **Working directory**: `$ProjectFileDir$`

### Terminal Integration

Add aliases to your shell configuration:

```bash
# ~/.bashrc or ~/.zshrc
alias gzh-debug='./scripts/debug/debug-cli.sh'
alias gzh-debug-test='./scripts/debug/debug-test.sh'
alias gzh-debug-attach='./scripts/debug/debug-attach.sh'
```

## Best Practices

1. **Start with logs**: Check application logs before debugging
2. **Use meaningful breakpoints**: Set breakpoints at decision points
3. **Prepare test data**: Have sample configs ready
4. **Document findings**: Keep notes of debugging sessions
5. **Clean up**: Stop debug servers when done
6. **Use version control**: Commit working debug configurations

## Resources

- [Delve Documentation](https://github.com/go-delve/delve/tree/master/Documentation)
- [GZH Manager Debugging Guide](../docs/debugging-guide.md)
- [Go Debugging Tutorial](https://golang.org/doc/gdb)
- [Delve Command Reference](https://github.com/go-delve/delve/blob/master/Documentation/cli/README.md)

---

**Note**: These scripts are specifically designed for the GZH Manager Go project. Modify paths and configurations as needed for your specific debugging requirements.
