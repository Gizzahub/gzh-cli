# Cleanup Action Plan

## Overview

This document provides a **step-by-step action plan** for safely removing unused code and simplifying the architecture of `gzh-manager-go`. The plan prioritizes **safety, testing, and gradual migration**.

## Pre-Cleanup Checklist

### ✅ Before Starting
- [ ] Create backup branch: `git checkout -b cleanup-backup`
- [ ] Run full test suite: `make test`
- [ ] Verify build works: `make build`
- [ ] Run linting: `make lint`
- [ ] Document current binary size: `ls -lh gz`

### ✅ Safety Verification
```bash
# Verify no hidden imports
grep -r "internal/legacy" --include="*.go" .
grep -r "internal/analysis" --include="*.go" .
grep -r "internal/api" --include="*.go" .
grep -r "pkg/cloud" --include="*.go" .

# Check go.mod dependencies
go mod why github.com/gizzahub/gzh-manager-go/internal/legacy
go mod why github.com/gizzahub/gzh-manager-go/internal/analysis
go mod why github.com/gizzahub/gzh-manager-go/internal/api
go mod why github.com/gizzahub/gzh-manager-go/pkg/cloud
```

## Phase 1: Remove Completely Unused Packages

### Step 1.1: Remove `internal/legacy` ✅ COMPLETED
**Risk Level**: LOW - Zero dependencies found
**Status**: Successfully removed in commit c8c81af

```bash
# Create working branch
git checkout -b remove-legacy-package

# Remove the package
rm -rf internal/legacy/

# Test build
make build
make test

# Commit if successful
git add .
git commit -m "refactor(sonnet): remove unused internal/legacy package

- Removed legacy error handling package with zero dependencies
- Package was not imported anywhere in codebase
- Reduces codebase complexity without functional impact"
```

### Step 1.2: ~~Remove `internal/analysis`~~ ⚠️ **CORRECTION: KEEP**
**Risk Level**: NONE - Package is actually used
**Status**: **DO NOT REMOVE** - Used by `cmd/doctor/godoc.go`

**Evidence of Usage**:
```go
// cmd/doctor/godoc.go
import "github.com/gizzahub/gzh-manager-go/internal/analysis/godoc"
```

**Recommendation**: Keep this package as it's required for doctor functionality.

### Step 1.3: Remove `internal/api` ✅ COMPLETED
**Risk Level**: LOW - Zero dependencies found
**Status**: Successfully removed in commit c9d884b

```bash
# Create working branch
git checkout -b remove-api-package

# Remove the package
rm -rf internal/api/

# Test build
make build
make test

# Commit if successful
git add .
git commit -m "refactor(sonnet): remove unused internal/api package

- Removed API optimization components (batcher, deduplicator, rate limiter)
- No active imports found in codebase
- Web API functionality outside CLI tool scope"
```

### Step 1.4: ~~Remove `pkg/cloud`~~ ⚠️ **CORRECTION: KEEP**
**Risk Level**: NONE - Package is actually used
**Status**: **DO NOT REMOVE** - Used by net-env commands

**Evidence of Usage**:
```go
// cmd/net-env/cloud.go, cmd/net-env/vpn_hierarchy_cmd.go
import "github.com/gizzahub/gzh-manager-go/pkg/cloud"
```

**Recommendation**: Keep this package as it's required for network environment functionality.

## Phase 2: Architecture Simplification (Optional)

### Step 2.1: Evaluate `internal/container` Impact
**Risk Level**: MEDIUM - Currently used by multiple commands

```bash
# Create analysis branch
git checkout -b analyze-container-usage

# Find all usages
grep -r "internal/container" --include="*.go" .

# Measure startup time with container
time ./gz --help

# Create simple constructor alternative for one command
# (Implementation details in separate task)
```

### Step 2.2: Simplify `internal/profiling` (Optional)
**Risk Level**: MEDIUM - Used by profile/doctor commands

```bash
# Create analysis branch
git checkout -b simplify-profiling

# Consider replacing with standard Go pprof
# Only if profiling complexity is identified as bottleneck
```

## Phase 3: Post-Cleanup Verification

### Step 3.1: Comprehensive Testing
```bash
# Full test suite
make test

# Build verification
make build

# Lint check
make lint

# Binary size comparison
ls -lh gz

# Runtime verification
./gz --help
./gz synclone --help
./gz git --help
```

### Step 3.2: Performance Validation
```bash
# Startup time measurement
time ./gz --help

# Memory usage check
./gz doctor benchmark # if available

# Integration test
./gz synclone validate examples/synclone-simple.yaml
```

## Rollback Procedures

### If Issues Arise
```bash
# Quick rollback to previous state
git checkout develop
git branch -D cleanup-backup # only if confident

# Or restore specific package
git checkout HEAD~1 -- internal/legacy/
git add .
git commit -m "rollback: restore internal/legacy package"
```

### Validation After Rollback
```bash
make build
make test
./gz --help
```

## Risk Mitigation

### Low-Risk Removals (Phase 1)
- **Verification**: Zero imports confirmed via grep
- **Testing**: Automated test suite catches any hidden dependencies
- **Rollback**: Simple git revert if issues arise

### Medium-Risk Changes (Phase 2)
- **Gradual Approach**: One command at a time
- **A/B Testing**: Keep both old and new approaches temporarily
- **Performance Monitoring**: Measure before/after metrics

## Success Metrics

### Code Reduction
- **Target**: ~1,000 lines of unused code removed (revised from 2,000)
- **Measurement**: `cloc` before/after comparison

### Binary Size Impact
- **Baseline**: Current binary size with `ls -lh gz`
- **Target**: Measurable reduction after cleanup

### Performance Impact
- **Startup Time**: `time ./gz --help` comparison
- **Memory Usage**: Runtime memory profiling

## Timeline Estimate

| Phase | Duration | Risk Level | Dependencies |
|-------|----------|------------|--------------|
| Phase 1.1,1.3 | 2-4 hours | LOW | None (completed) |
| Phase 2.1 | 4-8 hours | MEDIUM | Analysis required |
| Phase 2.2 | 2-4 hours | MEDIUM | Phase 2.1 complete |
| Phase 3 | 1-2 hours | LOW | All previous phases |

## Automation Scripts

### Cleanup Verification Script
```bash
#!/bin/bash
# File: scripts/verify-cleanup.sh

echo "🔍 Verifying cleanup safety..."

# Check for any remaining imports
echo "Checking for imports..."
for pkg in "internal/legacy" "internal/api"; do
    if grep -r "$pkg" --include="*.go" . > /dev/null 2>&1; then
        echo "❌ Found imports for $pkg"
        exit 1
    fi
done

echo "✅ No imports found for removed packages"

# Test build
echo "Testing build..."
if ! make build > /dev/null 2>&1; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Test suite
echo "Running tests..."
if ! make test > /dev/null 2>&1; then
    echo "❌ Tests failed"
    exit 1
fi

echo "✅ All tests passed"
echo "🎉 Cleanup verification complete!"
```

## Next Steps After Cleanup

1. **Documentation Update**: Update CLAUDE.md to reflect simplified architecture
2. **Contributing Guide**: Update development setup instructions
3. **Performance Baseline**: Establish new performance metrics
4. **Code Review**: Team review of architectural changes

## Emergency Contacts

- **Rollback Authority**: Project maintainer
- **Testing Issues**: Run `make test` and report failures
- **Build Issues**: Verify `make build` works on clean checkout

---

**⚠️ Important**: Always test each phase thoroughly before proceeding to the next. The cleanup is designed to be **safe and reversible** at each step.
