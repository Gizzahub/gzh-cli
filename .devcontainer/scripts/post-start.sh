#!/bin/bash

# Post-start script - runs every time the container starts
# This script handles startup tasks and service initialization

set -e

echo "🚀 Starting GZH Manager development services..."

# Ensure we're in the workspace directory
cd /workspace

# Create runtime directories
mkdir -p /workspace/tmp
mkdir -p /workspace/.devcontainer/logs

# Log startup
echo "$(date): Development container started" >> /workspace/.devcontainer/logs/startup.log

# Check if this is a fresh start or restart
if [[ -f "/workspace/.devcontainer/logs/startup.log" ]]; then
    START_COUNT=$(grep -c "Development container started" /workspace/.devcontainer/logs/startup.log || echo "1")
    if [[ $START_COUNT -gt 1 ]]; then
        echo "🔄 Container restart detected (start #$START_COUNT)"
    else
        echo "🎆 Fresh container start"
    fi
fi

# Verify Git configuration
echo "📝 Verifying Git configuration..."
if [[ -f "/home/vscode/.gitconfig" ]]; then
    git config --global --add safe.directory /workspace
    echo "✅ Git configuration verified"
else
    echo "⚠️ Git configuration not found - using defaults"
fi

# Check Go environment
echo "🟢 Checking Go environment..."
if go version >/dev/null 2>&1; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    echo "✅ Go $GO_VERSION is available"

    # Verify GOPATH and modules
    if [[ -f "go.mod" ]]; then
        echo "✅ Go module detected"
    else
        echo "⚠️ No go.mod found"
    fi
else
    echo "❌ Go is not available"
fi

# Check Node.js environment
echo "🟡 Checking Node.js environment..."
if node --version >/dev/null 2>&1; then
    NODE_VERSION=$(node --version)
    echo "✅ Node.js $NODE_VERSION is available"

    # Check npm
    if npm --version >/dev/null 2>&1; then
        NPM_VERSION=$(npm --version)
        echo "✅ npm $NPM_VERSION is available"
    fi
else
    echo "❌ Node.js is not available"
fi

# Check Python environment
echo "🔵 Checking Python environment..."
if python3 --version >/dev/null 2>&1; then
    PYTHON_VERSION=$(python3 --version | cut -d' ' -f2)
    echo "✅ Python $PYTHON_VERSION is available"

    # Check pip
    if pip3 --version >/dev/null 2>&1; then
        PIP_VERSION=$(pip3 --version | cut -d' ' -f2)
        echo "✅ pip $PIP_VERSION is available"
    fi
else
    echo "❌ Python is not available"
fi

# Check Docker environment
echo "🐳 Checking Docker environment..."
if docker --version >/dev/null 2>&1; then
    DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | sed 's/,//')
    echo "✅ Docker $DOCKER_VERSION is available"

    # Check if Docker daemon is accessible
    if docker info >/dev/null 2>&1; then
        echo "✅ Docker daemon is accessible"
    else
        echo "⚠️ Docker daemon is not accessible (this is normal in some setups)"
    fi
else
    echo "❌ Docker CLI is not available"
fi

# Check development tools
echo "🔧 Checking development tools..."

# Check make
if make --version >/dev/null 2>&1; then
    echo "✅ make is available"
else
    echo "❌ make is not available"
fi

# Check golangci-lint
if golangci-lint --version >/dev/null 2>&1; then
    GOLANGCI_VERSION=$(golangci-lint --version | head -n1 | cut -d' ' -f4)
    echo "✅ golangci-lint $GOLANGCI_VERSION is available"
else
    echo "⚠️ golangci-lint is not available"
fi

# Check pre-commit
if pre-commit --version >/dev/null 2>&1; then
    PRECOMMIT_VERSION=$(pre-commit --version | cut -d' ' -f2)
    echo "✅ pre-commit $PRECOMMIT_VERSION is available"
else
    echo "⚠️ pre-commit is not available"
fi

# Check GitHub CLI
if gh --version >/dev/null 2>&1; then
    GH_VERSION=$(gh --version | head -n1 | cut -d' ' -f3)
    echo "✅ GitHub CLI $GH_VERSION is available"
else
    echo "⚠️ GitHub CLI is not available"
fi

# Refresh Go module cache if needed
if [[ -f "go.mod" ]] && [[ -n "$(find . -name 'go.mod' -newer /workspace/.devcontainer/logs/startup.log 2>/dev/null || echo 'refresh')" ]]; then
    echo "🟢 Refreshing Go module cache..."
    go mod download || echo "⚠️ Failed to download Go modules"
fi

# Start background services if needed
echo "🔄 Starting background services..."

# Example: Start a file watcher for configuration changes
# (This would be customized based on specific needs)
# fswatch -o . | while read; do echo "File change detected"; done &

# Log service status
echo "$(date): Background services started" >> /workspace/.devcontainer/logs/startup.log

# Display helpful information
echo ""
echo "🎆 Development environment ready!"
echo ""
echo "📍 Current status:"
echo "   - Workspace: /workspace"
echo "   - Git safe directory: configured"
echo "   - Go modules: $(if [[ -f 'go.mod' ]]; then echo 'detected'; else echo 'not found'; fi)"
echo "   - Node.js project: $(if [[ -f 'package.json' ]] || [[ -d 'web' ]] || [[ -d 'bindings/nodejs' ]]; then echo 'detected'; else echo 'not found'; fi)"
echo "   - Python project: $(if [[ -d 'bindings/python' ]]; then echo 'detected'; else echo 'not found'; fi)"
echo ""
echo "🚀 Quick commands:"
echo "   make bootstrap  # Install dependencies"
echo "   make build     # Build project"
echo "   make test      # Run tests"
echo "   make lint      # Check code quality"
echo "   ./gz --help    # Show CLI help (after build)"
echo ""
echo "📝 Logs location: /workspace/.devcontainer/logs/"
echo ""
