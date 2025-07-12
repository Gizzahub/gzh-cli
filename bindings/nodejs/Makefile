# Makefile for GZH Manager Node.js Bindings

.PHONY: all build clean test lint format install dev prepare publish help

# Default target
all: build

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	npm install

# Build the Go shared library
build-go:
	@echo "🔨 Building Go shared library..."
	cd ../.. && CGO_ENABLED=1 go build -buildmode=c-shared -o bindings/nodejs/libgzh_node.so bindings/nodejs/libgzh.go
	@if [ "$(shell uname)" = "Darwin" ]; then \
		cd ../.. && CGO_ENABLED=1 go build -buildmode=c-shared -o bindings/nodejs/libgzh_node.dylib bindings/nodejs/libgzh.go; \
	fi

# Build the native Node.js addon
build-native: build-go
	@echo "🔨 Building native addon..."
	npm run build:native

# Build TypeScript
build-ts:
	@echo "🔨 Building TypeScript..."
	npm run build:ts

# Build everything
build: install build-native build-ts
	@echo "✅ Build completed successfully"

# Development build (faster, no optimizations)
dev: install
	@echo "🔨 Development build..."
	npm run build:native
	npm run build:ts
	@echo "✅ Development build completed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	npm run clean
	rm -rf node_modules
	rm -f libgzh_node.so libgzh_node.dylib libgzh_node.h
	rm -rf build/

# Run tests
test: build
	@echo "🧪 Running tests..."
	npm test

# Run tests in watch mode
test-watch:
	@echo "🧪 Running tests in watch mode..."
	npm run test:watch

# Lint code
lint:
	@echo "🔍 Linting code..."
	npm run lint

# Format code
format:
	@echo "✨ Formatting code..."
	npm run format

# Prepare for publishing (build + test + lint)
prepare: clean build lint test
	@echo "📦 Package prepared for publishing"

# Publish to npm (requires authentication)
publish: prepare
	@echo "🚀 Publishing to npm..."
	npm publish

# Publish as beta version
publish-beta: prepare
	@echo "🚀 Publishing beta version to npm..."
	npm publish --tag beta

# Install package globally for testing
install-global: build
	@echo "🌍 Installing package globally..."
	npm pack
	npm install -g *.tgz
	rm -f *.tgz

# Uninstall global package
uninstall-global:
	@echo "🗑️  Uninstalling global package..."
	npm uninstall -g @gizzahub/gzh-manager

# Development server (if applicable)
serve:
	@echo "🚀 Starting development server..."
	node -r ts-node/register src/examples/basic-usage.ts

# Run examples
examples: build
	@echo "📚 Running examples..."
	@echo "Running basic usage example..."
	node dist/examples/basic-usage.js
	@echo "Running advanced usage example..."
	node dist/examples/advanced-usage.js

# Check package health
check: lint test
	@echo "🏥 Checking package health..."
	npm audit
	npm outdated || true

# Update dependencies
update:
	@echo "⬆️  Updating dependencies..."
	npm update
	npm audit fix

# Generate documentation
docs:
	@echo "📖 Generating documentation..."
	npx typedoc src/index.ts --out docs/

# Setup development environment
setup: install
	@echo "🛠️  Setting up development environment..."
	npm run prepare
	@echo "✅ Development environment ready"

# Docker build
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t gzh-manager-nodejs .

# Docker test
docker-test: docker-build
	@echo "🐳 Running tests in Docker..."
	docker run --rm gzh-manager-nodejs npm test

# Cross-platform build
cross-build:
	@echo "🌐 Building for multiple platforms..."
	npm run build:native -- --target_arch=x64
	npm run build:native -- --target_arch=arm64

# Release preparation checklist
release-check:
	@echo "📋 Release checklist:"
	@echo "  ✓ Version bumped in package.json"
	@echo "  ✓ CHANGELOG.md updated"
	@echo "  ✓ Tests passing"
	@echo "  ✓ No linting errors"
	@echo "  ✓ Documentation updated"
	@echo "  ✓ Examples working"
	@make check

# Benchmark tests
benchmark: build
	@echo "⚡ Running benchmark tests..."
	node -r ts-node/register benchmarks/performance.ts

# Memory leak detection
memory-test: build
	@echo "🧠 Running memory leak detection..."
	node --inspect --expose-gc dist/examples/memory-test.js

# Security audit
security:
	@echo "🔒 Running security audit..."
	npm audit
	npx audit-ci

# Performance profiling
profile: build
	@echo "📊 Running performance profiling..."
	node --prof dist/examples/basic-usage.js
	node --prof-process isolate-*.log > profile.txt
	@echo "Profile saved to profile.txt"

# Help target
help:
	@echo "📚 Available targets:"
	@echo "  all          - Build everything (default)"
	@echo "  install      - Install dependencies"
	@echo "  build        - Build Go library, native addon, and TypeScript"
	@echo "  build-go     - Build Go shared library only"
	@echo "  build-native - Build native Node.js addon only"
	@echo "  build-ts     - Build TypeScript only"
	@echo "  dev          - Development build (faster)"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-watch   - Run tests in watch mode"
	@echo "  lint         - Lint code"
	@echo "  format       - Format code"
	@echo "  prepare      - Prepare for publishing"
	@echo "  publish      - Publish to npm"
	@echo "  publish-beta - Publish beta version"
	@echo "  examples     - Run example scripts"
	@echo "  check        - Check package health"
	@echo "  update       - Update dependencies"
	@echo "  docs         - Generate documentation"
	@echo "  setup        - Setup development environment"
	@echo "  help         - Show this help"

# Environment info
env-info:
	@echo "🌍 Environment Information:"
	@echo "Node.js version: $(shell node --version)"
	@echo "npm version: $(shell npm --version)"
	@echo "Platform: $(shell uname -s)"
	@echo "Architecture: $(shell uname -m)"
	@echo "Go version: $(shell go version 2>/dev/null || echo 'Go not found')"
	@echo "Python version: $(shell python3 --version 2>/dev/null || echo 'Python not found')"