# Build Guide - gzh-cli

## Build & Development

```bash
# One-time setup
make bootstrap

# Before every commit (CRITICAL)
make fmt && make lint && make test

# Build & install
make build
make install
```

## Development Workflow

```bash
# Quick development cycle
make dev          # format + lint + test
make dev-fast     # format + test only

# Before every commit (CRITICAL)
make fmt && make lint && make test

# Watch mode (if available)
make watch
```

## Code Quality

```bash
# Format code
make fmt          # gofumpt + gci

# Lint code
make lint         # golangci-lint with auto-fix
make lint-all     # format + lint + pre-commit

# All quality checks
make quality      # fmt + lint + test
```

## Testing

```bash
# All tests
make test

# Specific package
go test ./cmd/{module} -v

# Specific test
go test ./cmd/git -run "TestCloneOrUpdate" -v

# Coverage
make cover

# Race detection
go test -race ./...
```

## Mocking

```bash
# Generate mocks
make generate-mocks

# Regenerate (clean + generate)
make regenerate-mocks
```

## Build Output

- **Binary name**: `gz`
- **Install location**: `$GOPATH/bin/gz`

## Troubleshooting

### Build Issues

```bash
make clean
make bootstrap
make build
```

### Lint Failures

```bash
make fmt        # Fix formatting
make lint       # Auto-fix linting issues
```

### Test Failures

```bash
# Run specific test with verbose
go test ./cmd/{module} -run "TestName" -v

# Check for race conditions
go test ./cmd/{module} -race
```

### Import Cycle

- **Cause**: Circular dependencies between packages
- **Fix**: Move shared types to `internal/` or `pkg/`

## Dependencies

### Adding new dependencies

```bash
# Add dependency
go get github.com/example/package

# Clean up
go mod tidy
```

**Prefer**:
- Standard library when possible
- Well-maintained third-party libraries
- Avoid heavy dependencies

## Local Development with External Libraries

### Using replace directives

```go
// go.mod
replace github.com/gizzahub/gzh-cli-git => ../gzh-cli-git
replace github.com/gizzahub/gzh-cli-quality => ../gzh-cli-quality
```

**IMPORTANT**: Remove replace directives before committing to main branch.

### Testing changes across repositories

1. Make changes in external library
2. Add replace directive in gzh-cli
3. Test integration
4. Commit library changes first
5. Update gzh-cli dependency version
6. Remove replace directive
7. Commit gzh-cli changes
