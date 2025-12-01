# Go Version Mismatch Troubleshooting

## Issue

When running `make build`, you may encounter errors like:

```
compile: version "go1.25.4" does not match go tool version "go1.25.1 X:nodwarf5"
```

## Root Cause

This error occurs when:

1. Go standard library was pre-compiled with a different Go version (e.g., go1.25.4)
1. Current `go` tool version is different (e.g., go1.25.1)
1. Build cache contains incompatible compiled objects

This is typically caused by:

- Upgrading/downgrading Go toolchain
- System-wide Go installation changes
- Multiple Go versions installed
- Shared build cache between different Go versions

## Solutions

### Option 1: Direct Build (Recommended)

Use `go build` directly instead of `make build`:

```bash
# Clean caches first
go clean -cache -modcache

# Build directly
go build -o gz ./cmd/gz

# Or with version flag
go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o gz ./cmd/gz
```

**Why this works**: Direct `go build` handles toolchain selection automatically.

### Option 2: Update Go to Match

Upgrade your Go installation to match the expected version:

```bash
# Check current version
go version

# If you see go1.25.1 but need go1.25.4:
# Download and install go1.25.4 from https://go.dev/dl/

# Or use your package manager
# For Arch/Manjaro:
sudo pacman -S go

# For Ubuntu/Debian:
sudo apt update && sudo apt install golang
```

### Option 3: Force Rebuild Standard Library

Rebuild Go standard library with current toolchain:

```bash
# Clean all caches
go clean -cache -modcache

# Rebuild standard library
go install std

# Try make build again
make build
```

### Option 4: Use Specific Toolchain

Specify toolchain version in go.mod:

```go
module github.com/Gizzahub/gzh-cli

go 1.24.0

toolchain go1.25.1  // Add this line
```

Then clean and rebuild:

```bash
go clean -cache -modcache
make build
```

## Verification

After applying a solution, verify the build works:

```bash
# Test direct build
go build -o /tmp/gz-test ./cmd/gz && /tmp/gz-test version

# Test make build (if using Makefile)
make build && ./gz version
```

## Prevention

### For Development

1. **Use single Go version**: Avoid having multiple Go versions installed
1. **Clean after upgrade**: Always run `go clean -cache -modcache` after Go upgrade/downgrade
1. **Use go modules**: Ensure `GO111MODULE=on` (default in Go 1.16+)

### For CI/CD

1. **Pin Go version** in CI configuration:

   ```yaml
   # .github/workflows/build.yml
   - uses: actions/setup-go@v5
     with:
       go-version: '1.24.0'  # Pin specific version
   ```

1. **Clean cache in CI**:

   ```yaml
   - name: Clean Go cache
     run: go clean -cache -modcache
   ```

## Related Issues

- Go issue: https://github.com/golang/go/issues/38705
- Toolchain selection: https://go.dev/doc/toolchain

## Quick Reference

| Command | Purpose |
|---------|---------|
| `go version` | Check current Go version |
| `go env GOVERSION` | Check Go version from env |
| `go clean -cache` | Clean build cache |
| `go clean -modcache` | Clean module cache |
| `go install std` | Rebuild standard library |
| `go build -o gz ./cmd/gz` | Direct build (bypasses Make) |

______________________________________________________________________

**Status**: Known Environment Issue
**Impact**: Blocks `make build` but not `go build`
**Workaround**: Use direct `go build` command
**Last Updated**: 2025-12-01
