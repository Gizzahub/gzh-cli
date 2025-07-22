# Makefile.dev - Development Workflow for gzh-manager-go
# Development environment, workflow automation, and quick iteration

# ==============================================================================
# Development Configuration
# ==============================================================================

# Colors for output
CYAN := \\033[36m
GREEN := \\033[32m
YELLOW := \\033[33m
RED := \\033[31m
BLUE := \\033[34m
MAGENTA := \\033[35m
RESET := \\033[0m

# ==============================================================================
# Quick Access Aliases for Development
# ==============================================================================

.PHONY: start stop restart status logs quick full setup-all

# Quick development aliases
start: run                     ## quick start: run development server
stop:                         ## stop running development server
	@echo "$(YELLOW)Stopping development server...$(RESET)"
	@pkill -f "$(executablename)" || echo "$(GREEN)No running $(executablename) processes found$(RESET)"

restart: stop start           ## restart development server

status:                       ## check development server status
	@echo "$(CYAN)Checking for running $(executablename) processes...$(RESET)"
	@pgrep -f "$(executablename)" > /dev/null && echo "$(GREEN)✅ $(executablename) is running$(RESET)" || echo "$(RED)❌ $(executablename) is not running$(RESET)"

logs:                         ## show recent log files
	@echo "$(CYAN)Recent log files:$(RESET)"
	@find . -name "*.log" -type f -exec ls -la {} \; 2>/dev/null || echo "$(YELLOW)No log files found$(RESET)"

# ==============================================================================
# Development Workflow Targets
# ==============================================================================

.PHONY: dev dev-fast verify ci-local pr-check

dev: fmt lint-check test ## run standard development workflow (format, lint, test)
	@echo "$(GREEN)✅ Standard development workflow completed!$(RESET)"

dev-fast: fmt test-unit ## quick development cycle (format and unit tests only)
	@echo "$(GREEN)✅ Fast development cycle completed!$(RESET)"

verify: fmt lint-check test cover-report check-consistency ## complete verification before PR
	@echo "$(GREEN)✅ Complete verification completed!$(RESET)"

ci-local: clean verify test-all security ## run full CI pipeline locally
	@echo "$(GREEN)✅ Local CI pipeline completed!$(RESET)"

pr-check: fmt lint-check test cover-report check-consistency ## pre-PR submission check
	@echo "$(GREEN)✅ Pre-PR check completed - ready for submission!$(RESET)"

# ==============================================================================
# Main Workflow Aliases
# ==============================================================================

quick: fmt lint-check test-unit ## quick development check (format + lint + unit tests)
	@echo "$(GREEN)✅ Quick development check completed!$(RESET)"

full: fmt lint test cover-report ## full quality check (comprehensive)
	@echo "$(GREEN)✅ Full quality check completed!$(RESET)"

setup-all: bootstrap install-tools ## complete project setup (dependencies + all tools)
	@echo "$(GREEN)🎉 Complete project setup finished!$(RESET)"

# ==============================================================================
# Code Analysis and Comments
# ==============================================================================

.PHONY: comments todo fixme notes deps-graph

comments: ## show all TODO/FIXME/NOTE comments in codebase
	@echo "$(CYAN)=== TODO comments ===$(RESET)"
	@grep -r "TODO" --include="*.go" . | grep -v vendor | grep -v .git || echo "$(GREEN)No TODOs found!$(RESET)"
	@echo ""
	@echo "$(CYAN)=== FIXME comments ===$(RESET)"
	@grep -r "FIXME" --include="*.go" . | grep -v vendor | grep -v .git || echo "$(GREEN)No FIXMEs found!$(RESET)"
	@echo ""
	@echo "$(CYAN)=== NOTE comments ===$(RESET)"
	@grep -r "NOTE" --include="*.go" . | grep -v vendor | grep -v .git || echo "$(GREEN)No NOTEs found!$(RESET)"

# Aliases for backward compatibility
todo: comments ## show all TODO comments (alias for comments)
fixme: comments ## show all FIXME comments (alias for comments)
notes: comments ## show all NOTE comments (alias for comments)

deps-graph: ## show module dependency graph
	@echo "$(CYAN)Module dependency graph:$(RESET)"
	@go mod graph

# ==============================================================================
# Documentation Generation
# ==============================================================================

.PHONY: changelog docs-serve docs-build godoc docs-check

changelog: ## generate changelog (requires git-chglog)
	@command -v git-chglog >/dev/null 2>&1 || { echo "$(RED)git-chglog not found. Install from: https://github.com/git-chglog/git-chglog$(RESET)"; exit 1; }
	@echo "$(CYAN)Generating changelog...$(RESET)"
	@git-chglog -o CHANGELOG.md
	@echo "$(GREEN)✅ Changelog generated: CHANGELOG.md$(RESET)"

docs-serve: ## serve documentation locally (requires mkdocs)
	@command -v mkdocs >/dev/null 2>&1 || { echo "$(RED)mkdocs not found. Install with: pip install mkdocs mkdocs-material$(RESET)"; exit 1; }
	@echo "$(CYAN)Starting documentation server...$(RESET)"
	@mkdocs serve

docs-build: ## build documentation site
	@command -v mkdocs >/dev/null 2>&1 || { echo "$(RED)mkdocs not found. Install with: pip install mkdocs mkdocs-material$(RESET)"; exit 1; }
	@echo "$(CYAN)Building documentation site...$(RESET)"
	@mkdocs build

godoc: ## run godoc server
	@echo "$(CYAN)Starting godoc server on http://localhost:6060$(RESET)"
	@godoc -http=:6060

docs-check: ## check for missing package documentation
	@echo "$(CYAN)Checking for missing package documentation...$(RESET)"
	@for pkg in $$(go list ./...); do \
		if ! go doc -short $$pkg | grep -q "^package"; then \
			echo "$(RED)Missing documentation for: $$pkg$(RESET)"; \
		fi; \
	done || echo "$(GREEN)✅ All packages have documentation$(RESET)"

# ==============================================================================
# Development Environment Information
# ==============================================================================

.PHONY: dev-info dev-status

dev-info: ## show development environment information
	@echo "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                         $(MAGENTA)Development Environment$(CYAN)                         ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo "$(RESET)"
	@echo "$(GREEN)🏗️  Environment Details:$(RESET)"
	@echo "  Go Version:     $$(go version | cut -d' ' -f3)"
	@echo "  GOPROXY:        $(GOPROXY)"
	@echo "  GOSUMDB:        $(GOSUMDB)"
	@echo "  GOPATH:         $$(go env GOPATH)"
	@echo "  GOROOT:         $$(go env GOROOT)"
	@echo ""
	@echo "$(GREEN)🔄 Development Workflows:$(RESET)"
	@echo "  • $(CYAN)dev$(RESET)                 Standard development workflow"
	@echo "  • $(CYAN)dev-fast$(RESET)            Quick development cycle"
	@echo "  • $(CYAN)quick$(RESET)               Quick check (format + lint + unit tests)"
	@echo "  • $(CYAN)full$(RESET)                Full quality check"
	@echo "  • $(CYAN)verify$(RESET)              Complete verification before PR"
	@echo "  • $(CYAN)ci-local$(RESET)            Run full CI pipeline locally"
	@echo "  • $(CYAN)pr-check$(RESET)            Pre-PR submission check"
	@echo ""
	@echo "$(GREEN)🚀 Quick Commands:$(RESET)"
	@echo "  • $(CYAN)start$(RESET)               Start development server"
	@echo "  • $(CYAN)stop$(RESET)                Stop development server"
	@echo "  • $(CYAN)restart$(RESET)             Restart development server"
	@echo "  • $(CYAN)status$(RESET)              Check server status"
	@echo "  • $(CYAN)logs$(RESET)                Show recent log files"

dev-status: ## show current development status
	@echo "$(CYAN)Development Status Check$(RESET)"
	@echo "$(BLUE)========================$(RESET)"
	@echo ""
	@echo "$(GREEN)📊 Project Status:$(RESET)"
	@printf "  %-20s " "Git Status:"; if git status --porcelain | grep -q .; then echo "$(YELLOW)Modified files$(RESET)"; else echo "$(GREEN)Clean$(RESET)"; fi
	@printf "  %-20s " "Current Branch:"; git branch --show-current 2>/dev/null || echo "$(RED)Unknown$(RESET)"
	@printf "  %-20s " "Last Commit:"; git log -1 --format="%h %s" 2>/dev/null | cut -c1-50 || echo "$(RED)No commits$(RESET)"
	@echo ""
	@echo "$(GREEN)🔧 Build Status:$(RESET)"
	@printf "  %-20s " "Binary Exists:"; if [ -f "$(executablename)" ]; then echo "$(GREEN)Yes$(RESET)"; else echo "$(YELLOW)No$(RESET)"; fi
	@printf "  %-20s " "Coverage File:"; if [ -f "coverage.out" ]; then echo "$(GREEN)Yes$(RESET)"; else echo "$(YELLOW)No$(RESET)"; fi
	@echo ""
	@echo "$(GREEN)🎯 Quick Actions:$(RESET)"
	@echo "  • $(CYAN)make quick$(RESET)          Quick development check"
	@echo "  • $(CYAN)make dev$(RESET)            Full development workflow"
	@echo "  • $(CYAN)make setup-all$(RESET)      Set up everything from scratch"
