#!/bin/bash
# Pre-commit hook script for maintaining code quality
# Install by running: ln -sf ../../scripts/pre-commit-lint.sh .git/hooks/pre-commit

set -e

echo "🔍 Running pre-commit checks..."

# 1. Format check
echo "📝 Checking code formatting..."
if ! make fmt-check > /dev/null 2>&1; then
    echo "❌ Code formatting issues detected. Running 'make fmt'..."
    make fmt
    echo "✅ Code formatted. Please review and stage the changes."
    exit 1
fi

# 2. Lint check
echo "🔎 Running lint checks..."
LINT_OUTPUT=$(make lint 2>&1 || true)
LINT_ERRORS=$(echo "$LINT_OUTPUT" | grep -E '^[^:]+:[0-9]+:[0-9]+:' | wc -l | tr -d ' ')

if [ "$LINT_ERRORS" -gt "13" ]; then
    echo "❌ Lint errors detected: $LINT_ERRORS errors (threshold: 13)"
    echo ""
    echo "High priority issues to fix:"
    echo "$LINT_OUTPUT" | grep -E "(errcheck|gosec|noctx):" | head -10
    echo ""
    echo "Run 'make lint' to see all issues."
    exit 1
fi

# 3. Build check
echo "🔨 Checking build..."
if ! make build > /dev/null 2>&1; then
    echo "❌ Build failed. Please fix compilation errors."
    exit 1
fi

echo "✅ All pre-commit checks passed!"
echo "   Lint errors: $LINT_ERRORS (within acceptable range)"
