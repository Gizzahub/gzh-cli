#!/bin/bash

# Post-attach script - runs when VS Code attaches to the container
# This script handles user-specific setup and workspace configuration

set -e

echo "🔗 VS Code attached to GZH Manager development container"

# Ensure we're in the workspace directory
cd /workspace

# Log the attach event
echo "$(date): VS Code attached" >> /workspace/.devcontainer/logs/vscode.log

# Display welcome message
echo ""
echo "🚀 Welcome to the GZH Manager Development Environment!"
echo ""
echo "📋 Project Information:"
echo "   Name: GZH Manager Go"
echo "   Type: Multi-language CLI tool (Go, Node.js, Python, React)"
echo "   Version: $(git describe --tags --always 2>/dev/null || echo 'development')"
echo "   Branch: $(git branch --show-current 2>/dev/null || echo 'unknown')"
echo ""

# Check project status
echo "🔍 Project Status:"

# Check if project is built
if [[ -f "gz" ]]; then
    echo "✅ Binary built: gz"
    GZ_VERSION=$(./gz version 2>/dev/null | cut -d' ' -f3 || echo 'unknown')
    echo "   Version: $GZ_VERSION"
else
    echo "⚠️ Binary not built - run 'make build' to build"
fi

# Check dependencies
if [[ -f "go.sum" ]]; then
    echo "✅ Go dependencies downloaded"
else
    echo "⚠️ Go dependencies not downloaded - run 'make bootstrap'"
fi

# Check Node.js projects
if [[ -d "web/node_modules" ]]; then
    echo "✅ React dashboard dependencies installed"
else
    echo "⚠️ React dashboard dependencies not installed"
fi

if [[ -d "bindings/nodejs/node_modules" ]]; then
    echo "✅ Node.js binding dependencies installed"
else
    echo "⚠️ Node.js binding dependencies not installed"
fi

# Check Python environment
if [[ -d "bindings/python/venv" ]]; then
    echo "✅ Python virtual environment created"
else
    echo "⚠️ Python virtual environment not created"
fi

# Check pre-commit hooks
if [[ -f ".git/hooks/pre-commit" ]]; then
    echo "✅ Pre-commit hooks installed"
else
    echo "⚠️ Pre-commit hooks not installed - run 'pre-commit install'"
fi

# Display useful commands
echo ""
echo "🛠️ Available Make Targets:"
echo "   make bootstrap     # Install all build dependencies"
echo "   make build        # Build the gz binary"
echo "   make test         # Run all tests"
echo "   make lint         # Run linting and code quality checks"
echo "   make fmt          # Format all code"
echo "   make clean        # Clean build artifacts"
echo "   make dev-frontend # Start React development server"
echo "   make security     # Run security analysis"
echo "   make pre-commit   # Run pre-commit hooks"
echo ""
echo "🟢 Go-specific commands:"
echo "   go test ./...     # Run Go tests"
echo "   go mod tidy       # Clean up go.mod"
echo "   golangci-lint run # Run Go linter"
echo ""
echo "🟡 Node.js commands:"
echo "   cd web && npm start           # Start React dev server"
echo "   cd bindings/nodejs && npm test # Test Node.js bindings"
echo ""
echo "🔵 Python commands:"
echo "   cd bindings/python && source venv/bin/activate  # Activate Python env"
echo ""
echo "🐳 Docker commands:"
echo "   docker build -t gzh-manager .  # Build Docker image"
echo "   make test-docker              # Run Docker-based integration tests"
echo ""

# Check for common issues and provide guidance
echo "🔍 Health Check:"

# Check if Go module is tidy
if ! go mod tidy -diff >/dev/null 2>&1; then
    echo "⚠️ Go modules may be untidy - consider running 'go mod tidy'"
else
    echo "✅ Go modules are tidy"
fi

# Check for Git configuration
if ! git config user.name >/dev/null 2>&1 || ! git config user.email >/dev/null 2>&1; then
    echo "⚠️ Git user not configured - set with:"
    echo "     git config --global user.name 'Your Name'"
    echo "     git config --global user.email 'your.email@example.com'"
else
    GIT_USER=$(git config user.name)
    GIT_EMAIL=$(git config user.email)
    echo "✅ Git configured as: $GIT_USER <$GIT_EMAIL>"
fi

# Check for environment variables that might be needed
echo ""
echo "🔑 Environment Variables:"
if [[ -n "$GITHUB_TOKEN" ]]; then
    echo "✅ GITHUB_TOKEN is set"
else
    echo "⚠️ GITHUB_TOKEN not set (may be needed for GitHub operations)"
fi

if [[ -n "$GITLAB_TOKEN" ]]; then
    echo "✅ GITLAB_TOKEN is set"
else
    echo "ℹ️ GITLAB_TOKEN not set (optional, for GitLab operations)"
fi

# Display recent activity
echo ""
echo "📋 Recent Activity:"
if [[ -f "/workspace/.devcontainer/logs/startup.log" ]]; then
    echo "   Last startup: $(tail -n1 /workspace/.devcontainer/logs/startup.log | cut -d':' -f1-2)"
fi

if git log --oneline -1 >/dev/null 2>&1; then
    LAST_COMMIT=$(git log --oneline -1 | cut -c1-50)
    echo "   Last commit: $LAST_COMMIT"
fi

# Display port information
echo ""
echo "📍 Port Forwarding:"
echo "   8080: GZH API Server (when running)"
echo "   3000: React Development Server (when running)"
echo "   9090: Prometheus (when enabled)"
echo "   6060: Go pprof (when enabled)"
echo ""

# Quick start suggestion
if [[ ! -f "gz" ]]; then
    echo "🏁 Quick Start:"
    echo "   1. Run: make bootstrap"
    echo "   2. Run: make build"
    echo "   3. Run: ./gz --help"
    echo ""
fi

# Display logs location
echo "📝 Logs:"
echo "   Container logs: /workspace/.devcontainer/logs/"
echo "   VS Code logs: /workspace/.devcontainer/logs/vscode.log"
echo ""

echo "🎉 Happy coding! The development environment is ready."
echo ""
