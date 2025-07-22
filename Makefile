# Makefile - gzh-manager-go CLI Tool
# Modular Makefile structure with comprehensive functionality
# Git Repository Management CLI Tool

# ==============================================================================
# Project Configuration
# ==============================================================================

# Project metadata
projectname := gzh-manager
executablename := gz
VERSION ?= $(shell git describe --always --abbrev=0 --tags 2>/dev/null || echo "dev")

# Go configuration
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

# Colors for output (shared across all include files)
export CYAN := \\033[36m
export GREEN := \\033[32m
export YELLOW := \\033[33m
export RED := \\033[31m
export BLUE := \\033[34m
export MAGENTA := \\033[35m
export RESET := \\033[0m

# ==============================================================================
# Include Modular Makefiles
# ==============================================================================

include Makefile.deps.mk    # Dependency management
include Makefile.build.mk   # Build and installation
include Makefile.test.mk    # Testing and coverage
include Makefile.quality.mk # Code quality and linting
include Makefile.tools.mk   # Tool installation and management
include Makefile.dev.mk     # Development workflow
include Makefile.docker.mk  # Docker operations

# ==============================================================================
# Enhanced Help System
# ==============================================================================

.DEFAULT_GOAL := help

.PHONY: help help-build help-test help-quality help-deps help-dev help-docker help-tools

help: ## show main help menu with categories
	@echo "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                           $(MAGENTA)gzh-manager-go Makefile Help$(CYAN)                       ║"
	@echo "║                    $(YELLOW)Git Repository Management CLI Tool$(CYAN)                      ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo "$(RESET)"
	@echo "$(GREEN)📋 Main Categories:$(RESET)"
	@echo "  $(YELLOW)make help-build$(RESET)    🔨 Build, installation, and deployment"
	@echo "  $(YELLOW)make help-test$(RESET)     🧪 Testing, benchmarks, and validation"
	@echo "  $(YELLOW)make help-quality$(RESET)  ✨ Code quality, formatting, and linting"
	@echo "  $(YELLOW)make help-deps$(RESET)     📦 Dependency management and updates"
	@echo "  $(YELLOW)make help-dev$(RESET)      🛠️  Development tools and workflow"
	@echo "  $(YELLOW)make help-docker$(RESET)   🐳 Docker operations and containers"
	@echo "  $(YELLOW)make help-tools$(RESET)    🔧 Tool installation and management"
	@echo ""
	@echo "$(GREEN)🚀 Quick Commands:$(RESET)"
	@echo "  $(CYAN)make start$(RESET)         Start development (run)"
	@echo "  $(CYAN)make stop$(RESET)          Stop development server"
	@echo "  $(CYAN)make restart$(RESET)       Restart development server"
	@echo "  $(CYAN)make status$(RESET)        Check development server status"
	@echo "  $(CYAN)make quick$(RESET)         Quick check (format + lint + unit tests)"
	@echo "  $(CYAN)make full$(RESET)          Full quality check (comprehensive)"
	@echo "  $(CYAN)make setup-all$(RESET)     Complete project setup"
	@echo ""
	@echo "$(GREEN)💡 Pro Tips:$(RESET)"
	@echo "  • Use $(YELLOW)'make quick'$(RESET) for fast development iteration"
	@echo "  • Use $(YELLOW)'make full'$(RESET) before pushing to ensure quality"
	@echo "  • Use $(YELLOW)'make setup-all'$(RESET) for first-time project setup"
	@echo "  • All commands support tab completion if bash-completion is installed"
	@echo ""
	@echo "$(BLUE)📖 Documentation: $(RESET)https://github.com/gizzahub/gzh-manager-go"

help-build: ## show build and deployment help
	@echo "$(GREEN)🔨 Build and Installation Commands:$(RESET)"
	@echo "  $(CYAN)build$(RESET)              Build golang binary ($(executablename))"
	@echo "  $(CYAN)install$(RESET)            Install golang binary to GOPATH/bin"
	@echo "  $(CYAN)run$(RESET)                Run the application"
	@echo "  $(CYAN)bootstrap$(RESET)          Install build dependencies"
	@echo "  $(CYAN)clean$(RESET)              Clean up build artifacts and binaries"
	@echo "  $(CYAN)release-dry-run$(RESET)    Run goreleaser in dry-run mode"
	@echo "  $(CYAN)release-snapshot$(RESET)   Create a snapshot release"
	@echo "  $(CYAN)release-check$(RESET)      Check goreleaser configuration"
	@echo "  $(CYAN)build-info$(RESET)         Show build environment information"

help-test: ## show testing help
	@echo "$(GREEN)🧪 Testing and Validation Commands:$(RESET)"
	@echo "  $(CYAN)test$(RESET)               Run all tests with coverage"
	@echo "  $(CYAN)test-unit$(RESET)          Run only unit tests (exclude integration/e2e)"
	@echo "  $(CYAN)test-integration$(RESET)   Run Docker-based integration tests"
	@echo "  $(CYAN)test-e2e$(RESET)           Run End-to-End test scenarios"
	@echo "  $(CYAN)test-all$(RESET)           Run all tests (unit, integration, e2e)"
	@echo "  $(CYAN)cover$(RESET)              Display test coverage"
	@echo "  $(CYAN)cover-html$(RESET)         Generate HTML coverage report"
	@echo "  $(CYAN)cover-report$(RESET)       Generate detailed coverage report"
	@echo "  $(CYAN)bench$(RESET)              Run all benchmarks"
	@echo "  $(CYAN)test-info$(RESET)          Show testing information and targets"

help-quality: ## show quality help
	@echo "$(GREEN)✨ Code Quality Commands:$(RESET)"
	@echo "  $(CYAN)fmt$(RESET)                Format Go files with gofumpt and gci"
	@echo "  $(CYAN)lint-check$(RESET)         Check lint issues without fixing"
	@echo "  $(CYAN)lint-fix$(RESET)           Run golangci-lint with auto-fix"
	@echo "  $(CYAN)security$(RESET)           Run all security checks"
	@echo "  $(CYAN)analyze$(RESET)            Run comprehensive code analysis"
	@echo "  $(CYAN)quality$(RESET)            Run comprehensive quality checks"
	@echo "  $(CYAN)quality-fix$(RESET)        Apply automatic quality fixes"
	@echo "  $(CYAN)pre-commit$(RESET)         Run pre-commit hooks"
	@echo "  $(CYAN)quality-info$(RESET)       Show quality tools and targets"

help-deps: ## show dependency help
	@echo "$(GREEN)📦 Dependency Management Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile.deps.mk 2>/dev/null | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\\n", $$1, $$2}' | head -10 || echo "  $(YELLOW)Run 'make deps-help' for dependency commands$(RESET)"

help-dev: ## show development workflow help
	@echo "$(GREEN)🛠️  Development Workflow Commands:$(RESET)"
	@echo "  $(CYAN)dev$(RESET)                Standard development workflow (format, lint, test)"
	@echo "  $(CYAN)dev-fast$(RESET)           Quick development cycle (format and unit tests only)"
	@echo "  $(CYAN)verify$(RESET)             Complete verification before PR"
	@echo "  $(CYAN)ci-local$(RESET)           Run full CI pipeline locally"
	@echo "  $(CYAN)pr-check$(RESET)           Pre-PR submission check"
	@echo "  $(CYAN)comments$(RESET)           Show all TODO/FIXME/NOTE comments in codebase"
	@echo "  $(CYAN)changelog$(RESET)          Generate changelog"
	@echo "  $(CYAN)docs-serve$(RESET)         Serve documentation locally"
	@echo "  $(CYAN)dev-info$(RESET)           Show development environment information"

help-docker: ## show Docker help
	@echo "$(GREEN)🐳 Docker Commands:$(RESET)"
	@echo "  $(CYAN)docker-build$(RESET)       Build Docker image"
	@echo "  $(CYAN)docker-run$(RESET)         Run Docker container"
	@echo "  $(CYAN)docker-stop$(RESET)        Stop and remove Docker containers"
	@echo "  $(CYAN)docker-logs$(RESET)        Show Docker container logs"
	@echo "  $(CYAN)docker-optimize$(RESET)    Analyze Docker image for optimization"
	@echo "  $(CYAN)docker-scan$(RESET)        Scan Docker image for vulnerabilities"
	@echo "  $(CYAN)docker-clean$(RESET)       Clean up Docker containers and images"
	@echo "  $(CYAN)docker-info$(RESET)        Show Docker information and targets"

help-tools: ## show tools help
	@echo "$(GREEN)🔧 Tool Management Commands:$(RESET)"
	@echo "  $(CYAN)install-tools$(RESET)      Install all development tools"
	@echo "  $(CYAN)tools-status$(RESET)       Check installed tool status"
	@echo "  $(CYAN)generate-mocks$(RESET)     Generate all mock files using gomock"
	@echo "  $(CYAN)pre-commit-install$(RESET) Install pre-commit hooks"
	@echo "  $(CYAN)tools-info$(RESET)         Show comprehensive tool information"

# ==============================================================================
# Project Information
# ==============================================================================

.PHONY: info about

info: ## show project information and current configuration
	@echo "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                         $(MAGENTA)gzh-manager-go Project Information$(CYAN)                   ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo "$(RESET)"
	@echo "$(GREEN)📋 Project Details:$(RESET)"
	@echo "  Name:           $(YELLOW)$(projectname)$(RESET)"
	@echo "  Executable:     $(YELLOW)$(executablename)$(RESET)"
	@echo "  Version:        $(YELLOW)$(VERSION)$(RESET)"
	@echo ""
	@echo "$(GREEN)🏗️  Build Environment:$(RESET)"
	@echo "  Go Version:     $$(go version | cut -d' ' -f3)"
	@echo "  GOPROXY:        $(GOPROXY)"
	@echo "  GOSUMDB:        $(GOSUMDB)"
	@echo "  GOPATH:         $$(go env GOPATH)"
	@echo "  GOROOT:         $$(go env GOROOT)"
	@echo ""
	@echo "$(GREEN)📁 Key Features:$(RESET)"
	@echo "  • Multi-platform Git repository cloning (GitHub, GitLab, Gitea, Gogs)"
	@echo "  • Package manager updates (asdf, Homebrew, SDKMAN)"
	@echo "  • Development environment management (AWS, Docker, Kubernetes)"
	@echo "  • Network environment transitions (WiFi, VPN, DNS, proxy)"
	@echo "  • JetBrains IDE settings monitoring and sync fixes"
	@echo ""
	@echo "$(GREEN)🔧 Available Modules:$(RESET)"
	@echo "  • $(CYAN)Build & Deploy$(RESET)      (Makefile.build.mk)  - Build, installation, and release"
	@echo "  • $(CYAN)Testing$(RESET)             (Makefile.test.mk)   - Unit, integration, and e2e tests"
	@echo "  • $(CYAN)Code Quality$(RESET)        (Makefile.quality.mk) - Formatting, linting, and security"
	@echo "  • $(CYAN)Dependencies$(RESET)        (Makefile.deps.mk)   - Dependency management and updates"
	@echo "  • $(CYAN)Development$(RESET)         (Makefile.dev.mk)    - Development workflow and tools"
	@echo "  • $(CYAN)Docker$(RESET)              (Makefile.docker.mk) - Container operations and optimization"
	@echo "  • $(CYAN)Tools$(RESET)               (Makefile.tools.mk)  - Tool installation and management"
