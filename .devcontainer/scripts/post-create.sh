#!/bin/bash

# Post-create script - runs after container is created
# This script sets up the development environment inside the container

set -e

echo "🚀 Setting up GZH Manager development environment..."

# Ensure we're in the workspace directory
cd /workspace

# Create development directories
mkdir -p /workspace/tmp
mkdir -p /workspace/dist
mkdir -p /workspace/.devcontainer/logs

echo "✅ Development directories created"

# Install Go dependencies and tools
echo "🟢 Installing Go dependencies..."
go mod download
go mod tidy

# Install additional Go tools not in the base image
echo "🟢 Installing additional Go tools..."
go install github.com/gotesttools/gotestfmt/v2@latest || true
go install honnef.co/go/tools/cmd/staticcheck@latest || true
go install github.com/kisielk/errcheck@latest || true

echo "✅ Go environment setup complete"

# Setup Node.js environment
if [[ -d "bindings/nodejs" ]]; then
    echo "🟡 Setting up Node.js binding environment..."
    cd /workspace/bindings/nodejs
    
    # Install Node.js dependencies
    npm ci
    
    # Build TypeScript
    npm run build:ts || echo "⚠️ TypeScript build failed - this is normal on first setup"
    
    cd /workspace
    echo "✅ Node.js binding environment setup complete"
fi

# Setup Python environment
if [[ -d "bindings/python" ]]; then
    echo "🔵 Setting up Python binding environment..."
    cd /workspace/bindings/python
    
    # Create virtual environment
    python3 -m venv venv
    source venv/bin/activate
    
    # Install dependencies
    pip install --upgrade pip
    pip install -r requirements.txt || pip install -e . || echo "⚠️ Python setup incomplete - dependencies may need manual installation"
    
    deactivate
    cd /workspace
    echo "✅ Python binding environment setup complete"
fi

# Setup React web dashboard
if [[ -d "web" ]]; then
    echo "⚙️ Setting up React dashboard environment..."
    cd /workspace/web
    
    # Install React dependencies
    npm ci
    
    cd /workspace
    echo "✅ React dashboard environment setup complete"
fi

# Setup pre-commit hooks
echo "🔒 Setting up pre-commit hooks..."
if command -v pre-commit >/dev/null 2>&1; then
    pre-commit install --install-hooks || echo "⚠️ Pre-commit hook installation failed"
    pre-commit install --hook-type commit-msg || echo "⚠️ Commit message hook installation failed"
    pre-commit install --hook-type pre-push || echo "⚠️ Pre-push hook installation failed"
    echo "✅ Pre-commit hooks installed"
else
    echo "⚠️ Pre-commit not available - skipping hook installation"
fi

# Initialize Git hooks and configuration
echo "📝 Configuring Git environment..."

# Set up Git configuration for the container
git config --global --add safe.directory /workspace
git config --global init.defaultBranch main
git config --global pull.rebase false

# Create .gitconfig template if it doesn't exist
if [[ ! -f "/home/vscode/.gitconfig" ]]; then
    cat > /home/vscode/.gitconfig << 'EOF'
[user]
	name = Developer
	email = dev@localhost
[init]
	defaultBranch = main
[pull]
	rebase = false
[core]
	editor = code --wait
[diff]
	tool = vscode
[difftool "vscode"]
	cmd = code --wait --diff $LOCAL $REMOTE
[merge]
	tool = vscode
[mergetool "vscode"]
	cmd = code --wait $MERGED
EOF
fi

echo "✅ Git environment configured"

# Bootstrap the project
echo "🎆 Bootstrapping project dependencies..."
make bootstrap || echo "⚠️ Bootstrap failed - some dependencies may need manual installation"

# Run initial build
echo "🔨 Running initial build..."
make build || echo "⚠️ Initial build failed - this is normal if dependencies are missing"

# Create useful aliases and shortcuts
echo "🔧 Creating development shortcuts..."
cat >> /home/vscode/.zshrc << 'EOF'

# GZH Manager Development Shortcuts
alias gzh-build="cd /workspace && make build"
alias gzh-test="cd /workspace && make test"
alias gzh-lint="cd /workspace && make lint"
alias gzh-fmt="cd /workspace && make fmt"
alias gzh-run="cd /workspace && make run"
alias gzh-clean="cd /workspace && make clean"
alias gzh-bootstrap="cd /workspace && make bootstrap"

# React development
alias react-dev="cd /workspace/web && npm start"
alias react-build="cd /workspace/web && npm run build"

# Node.js binding development
alias node-build="cd /workspace/bindings/nodejs && npm run build"
alias node-test="cd /workspace/bindings/nodejs && npm test"

# Python binding development
alias py-activate="cd /workspace/bindings/python && source venv/bin/activate"
alias py-test="cd /workspace/bindings/python && source venv/bin/activate && python -m pytest"

# Docker shortcuts
alias docker-build="docker build -t gzh-manager ."
alias docker-run="docker run -it gzh-manager"

# Development helpers
alias logs="tail -f /workspace/.devcontainer/logs/*.log"
alias workspace="cd /workspace"

# Git shortcuts
alias gs="git status"
alias ga="git add"
alias gc="git commit"
alias gp="git push"
alias gl="git log --oneline -10"

EOF

echo "✅ Development shortcuts created"

# Create a development log
echo "$(date): Development container created and configured" > /workspace/.devcontainer/logs/setup.log

# Display completion message
echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "📝 Available commands:"
echo "   make bootstrap    # Install all build dependencies"
echo "   make build       # Build the gz binary"
echo "   make test        # Run all tests"
echo "   make lint        # Run code quality checks"
echo "   make fmt         # Format code"
echo "   make dev-frontend # Start React development server"
echo ""
echo "🚀 Quick start:"
echo "   1. Run: make bootstrap"
echo "   2. Run: make build"
echo "   3. Run: ./gz --help"
echo ""
echo "📍 Project ports:"
echo "   - 8080: GZH API Server"
echo "   - 3000: React Development Server"
echo "   - 9090: Prometheus (if enabled)"
echo "   - 6060: Go pprof (if enabled)"
echo ""