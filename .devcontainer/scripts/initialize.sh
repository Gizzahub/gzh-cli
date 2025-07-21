#!/bin/bash

# Initialize script - runs before container creation
# This script prepares the host environment for the dev container

set -e

echo "🚀 Initializing GZH Manager development environment..."

# Check if required tools are available on host
command -v docker >/dev/null 2>&1 || {
    echo "❌ Docker is required but not installed. Please install Docker first."
    exit 1
}

# Check Docker daemon is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker daemon is not running. Please start Docker first."
    exit 1
fi

echo "✅ Docker is available and running"

# Check if we're in the correct directory
if [[ ! -f "go.mod" ]] || [[ ! -f "main.go" ]]; then
    echo "❌ This doesn't appear to be the GZH Manager project root."
    echo "   Please run this from the project root directory."
    exit 1
fi

echo "✅ Project structure validated"

# Create necessary directories if they don't exist
mkdir -p .devcontainer/logs
mkdir -p tmp
mkdir -p dist

echo "✅ Development directories created"

# Set proper permissions for scripts
chmod +x .devcontainer/scripts/*.sh 2>/dev/null || true

echo "✅ Script permissions set"

echo "🎉 Initialization complete! You can now open this project in VS Code Dev Containers."
echo "   Use: 'Remote-Containers: Open Folder in Container'"

echo ""
echo "📝 Quick start commands after container opens:"
echo "   make bootstrap    # Install all dependencies"
echo "   make build       # Build the gz binary"
echo "   make test        # Run tests"
echo "   make dev-frontend # Start React development server"
echo ""
