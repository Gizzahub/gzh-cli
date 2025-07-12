# Makefile for GZH Manager Python Client

# Variables
PYTHON := python3
PIP := pip3
GO := go
LIBRARY_NAME := libgzh
PACKAGE_NAME := gzhclient

# Platform-specific library extension
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	LIB_EXT := .so
endif
ifeq ($(UNAME_S),Darwin)
	LIB_EXT := .dylib
endif
ifeq ($(UNAME_S),Windows)
	LIB_EXT := .dll
endif

LIBRARY := $(LIBRARY_NAME)$(LIB_EXT)

.PHONY: help build install install-dev clean test lint format type-check docs build-dist upload-test upload-prod

help: ## Show this help message
	@echo "GZH Manager Python Client - Available targets:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go shared library
	@echo "🔨 Building Go shared library..."
	$(GO) build -buildmode=c-shared -o $(LIBRARY) libgzh.go
	@echo "✅ Built $(LIBRARY)"

install: build ## Install the package
	@echo "📦 Installing package..."
	$(PIP) install .
	@echo "✅ Package installed"

install-dev: build ## Install the package in development mode
	@echo "🔧 Installing package in development mode..."
	$(PIP) install -e ".[dev]"
	@echo "✅ Development installation complete"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf build/
	rm -rf dist/
	rm -rf *.egg-info/
	rm -rf .pytest_cache/
	rm -rf .coverage
	rm -rf htmlcov/
	rm -rf .mypy_cache/
	rm -rf .tox/
	rm -f $(LIBRARY)
	rm -f libgzh.h
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	@echo "✅ Clean complete"

test: build ## Run tests
	@echo "🧪 Running tests..."
	pytest tests/ -v
	@echo "✅ Tests complete"

test-cov: build ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	pytest tests/ -v --cov=$(PACKAGE_NAME) --cov-report=html --cov-report=term
	@echo "✅ Coverage report generated in htmlcov/"

lint: ## Run linting
	@echo "🔍 Running linter..."
	flake8 $(PACKAGE_NAME)/ examples/ tests/
	@echo "✅ Linting complete"

format: ## Format code
	@echo "🎨 Formatting code..."
	black $(PACKAGE_NAME)/ examples/ tests/
	@echo "✅ Code formatted"

type-check: ## Run type checking
	@echo "🔍 Running type checker..."
	mypy $(PACKAGE_NAME)/
	@echo "✅ Type checking complete"

check: format lint type-check test ## Run all checks (format, lint, type-check, test)

docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@echo "Documentation is in README.md and docstrings"
	@echo "✅ Documentation ready"

build-dist: clean build ## Build distribution packages
	@echo "📦 Building distribution packages..."
	$(PYTHON) -m build
	@echo "✅ Distribution packages built in dist/"

upload-test: build-dist ## Upload to Test PyPI
	@echo "🚀 Uploading to Test PyPI..."
	twine upload --repository testpypi dist/*
	@echo "✅ Uploaded to Test PyPI"

upload-prod: build-dist ## Upload to Production PyPI
	@echo "🚀 Uploading to Production PyPI..."
	twine upload dist/*
	@echo "✅ Uploaded to Production PyPI"

example: build ## Run basic example
	@echo "🏃 Running basic example..."
	cd examples && $(PYTHON) basic_usage.py

example-bulk: build ## Run bulk clone example (dry run)
	@echo "🏃 Running bulk clone example (dry run)..."
	cd examples && DRY_RUN=true $(PYTHON) bulk_clone_example.py

setup-dev: ## Setup development environment
	@echo "🔧 Setting up development environment..."
	$(PYTHON) -m venv venv
	@echo "Activate with: source venv/bin/activate (Linux/macOS) or venv\\Scripts\\activate (Windows)"
	@echo "Then run: make install-dev"

verify-go: ## Verify Go installation
	@echo "🔍 Verifying Go installation..."
	$(GO) version
	@echo "✅ Go is installed and available"

verify-python: ## Verify Python installation
	@echo "🔍 Verifying Python installation..."
	$(PYTHON) --version
	$(PIP) --version
	@echo "✅ Python is installed and available"

verify: verify-go verify-python ## Verify all dependencies

install-deps: ## Install Python dependencies
	@echo "📦 Installing Python dependencies..."
	$(PIP) install --upgrade pip setuptools wheel
	$(PIP) install build twine pytest pytest-cov black flake8 mypy
	@echo "✅ Dependencies installed"

# Development workflow targets
dev-setup: verify setup-dev install-deps ## Complete development setup
	@echo "✅ Development environment ready!"
	@echo "Next steps:"
	@echo "  1. Activate virtual environment: source venv/bin/activate"
	@echo "  2. Install package: make install-dev"
	@echo "  3. Run tests: make test"
	@echo "  4. Run example: make example"

dev-workflow: format lint type-check test ## Complete development workflow

release-prep: clean check build-dist ## Prepare for release
	@echo "🚀 Release preparation complete!"
	@echo "Distribution files ready in dist/"
	@echo "Next: Run 'make upload-test' to test upload, then 'make upload-prod' for production"

# Help is default target
.DEFAULT_GOAL := help