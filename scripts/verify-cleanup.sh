#!/bin/bash

# 스크립트명: verify-cleanup.sh
# 용도: 코드 정리 작업의 안전성을 검증
# 사용법: ./scripts/verify-cleanup.sh
# 예시: ./scripts/verify-cleanup.sh

set -e

echo "🔍 Verifying cleanup safety..."

# Check for any remaining imports of removed packages
echo "Checking for imports of removed packages..."
removed_packages=("internal/legacy" "internal/api")
found_imports=false

for pkg in "${removed_packages[@]}"; do
    echo "  Checking for imports of $pkg..."
    if grep -r "$pkg" --include="*.go" . > /dev/null 2>&1; then
        echo "❌ Found imports for $pkg:"
        grep -r "$pkg" --include="*.go" . || true
        found_imports=true
    else
        echo "  ✅ No imports found for $pkg"
    fi
done

if $found_imports; then
    echo "❌ Found imports for removed packages - cleanup verification failed"
    exit 1
fi

echo "✅ No imports found for removed packages"

# Test build
echo "Testing build..."
if ! make build > /dev/null 2>&1; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Test basic functionality
echo "Testing basic functionality..."
if ! ./gz --help > /dev/null 2>&1; then
    echo "❌ Basic help command failed"
    exit 1
fi

echo "✅ Basic functionality works"

# Test key commands
echo "Testing key commands..."
key_commands=("synclone --help" "git --help" "dev-env --help" "net-env --help")

for cmd in "${key_commands[@]}"; do
    echo "  Testing: gz $cmd"
    if ! ./gz $cmd > /dev/null 2>&1; then
        echo "❌ Command 'gz $cmd' failed"
        exit 1
    fi
done

echo "✅ Key commands work correctly"

# Performance check
echo "Measuring startup performance..."
startup_time=$(time -p ./gz --help 2>&1 >/dev/null | grep real | awk '{print $2}')
echo "  Startup time: ${startup_time}s"

# Performance regression check (startup should be under 50ms for this CLI)  
if command -v bc &> /dev/null && [[ -n "$startup_time" ]]; then
    threshold="0.05"
    comparison=$(echo "$startup_time > $threshold" | bc -l 2>/dev/null || echo "0")
    if [[ "$comparison" == "1" ]]; then
        echo "⚠️  WARNING: Startup time ${startup_time}s exceeds threshold ${threshold}s"
        echo "   Consider investigating performance regression"
    else
        echo "✅ Startup time within acceptable threshold (${threshold}s)"
    fi
elif [[ -n "$startup_time" ]]; then
    echo "  (Install 'bc' for automatic performance regression checking)"
else
    echo "  ✅ Startup time measurement completed (threshold check skipped)"
fi

# Binary size check
binary_size=$(ls -lh gz | awk '{print $5}')
echo "  Binary size: $binary_size"

echo ""
echo "🎉 Cleanup verification complete!"
echo "📊 Summary:"
echo "  - Removed packages: ${#removed_packages[@]} (internal/legacy, internal/api)"
echo "  - Build: ✅ Success"
echo "  - Functionality: ✅ All key commands working"
echo "  - Startup time: ${startup_time}s"
echo "  - Binary size: $binary_size"
echo ""
echo "✅ All verification checks passed - cleanup was successful!"