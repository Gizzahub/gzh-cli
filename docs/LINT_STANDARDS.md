# Lint Standards and Code Quality Guidelines

## Overview

This document outlines the lint standards and code quality guidelines established after the comprehensive lint fixing session that reduced errors from 58 to 13 (78% reduction).

## Lint Configuration

The project uses `golangci-lint` with the configuration in `.golangci.yml`. Key linters enabled:

### High Priority (Security & Correctness)
- **errcheck**: Ensures all errors are properly handled
- **gosec**: Identifies security vulnerabilities
- **noctx**: Ensures context is used for cancellable operations

### Medium Priority (Code Quality)
- **gocritic**: Provides code style suggestions
- **revive**: Enforces Go best practices
- **godot**: Ensures comments end with periods
- **staticcheck**: Advanced static analysis

### Style & Convention
- **tagliatelle**: Enforces snake_case for JSON struct tags

## Established Standards

### 1. Error Handling
```go
// Always handle errors explicitly
homeDir, err := os.UserHomeDir()
if err != nil {
    // Provide fallback or return error
    homeDir = "/tmp"
}
```

### 2. File Permissions
```go
// Use secure file permissions
os.OpenFile(path, flags, 0o600)  // Files: 600
os.MkdirAll(path, 0o750)         // Directories: 750
```

### 3. Context Usage
```go
// Always use context for exec commands
cmd := exec.CommandContext(ctx, "command", "args")
```

### 4. JSON Tag Naming
```go
// Use snake_case for JSON tags
type Config struct {
    ServerURL  string `json:"server_url"`  // ✓
    APIKey     string `json:"api_key"`     // ✓
    // Not: serverUrl, apiKey
}
```

### 5. Comment Standards
```go
// Package comments are required
// Package pm provides package manager commands.
package pm

// Exported entities need proper comments
// ErrorCodeInvalidConfig indicates invalid configuration data.
const ErrorCodeInvalidConfig = "INVALID_CONFIG"

// All comments should end with periods.
```

### 6. Code Complexity
- Functions should maintain cognitive complexity < 20
- Cyclomatic complexity should be < 15
- Break complex functions into smaller, focused functions

## Remaining Acceptable Issues

The following low-priority issues are acceptable:

### 1. Unused Parameters (unparam)
Context parameters may be unused but kept for:
- Interface consistency
- Future extensibility
- Framework patterns (e.g., Cobra commands)

```go
// Acceptable: ctx unused but part of interface
func newCommand(ctx context.Context) *cobra.Command {
    // ctx might not be used but keeps interface consistent
}
```

### 2. Complexity Issues
Some functions exceed complexity thresholds but are acceptable if:
- They handle multiple related cases
- Breaking them would reduce readability
- They're well-tested and documented

## Maintenance Guidelines

### Pre-commit Checks
Run before committing:
```bash
make fmt        # Format code
make lint       # Check for lint errors
make test       # Run tests
```

### Adding New Code
1. Follow the established patterns in existing code
2. Ensure all exported entities have comments
3. Use secure file permissions
4. Handle all errors appropriately
5. Use context for cancellable operations

### Continuous Improvement
- Regularly run `make lint` to catch new issues
- Update `.golangci.yml` as needed
- Document any exceptions to standards
- Refactor complex functions when possible

## Tools and Commands

```bash
# Format code
make fmt

# Run all linters
make lint

# Run specific linter
golangci-lint run --no-config --disable-all --enable=errcheck ./...

# Clean and rebuild
make clean && make build
```

## Summary

These standards ensure:
- **Security**: Proper permissions and secure coding practices
- **Reliability**: Comprehensive error handling
- **Maintainability**: Consistent code style and documentation
- **Quality**: Low complexity and high readability

By following these guidelines, the codebase maintains its improved quality and remains easy to work with.