# Makefile.tools - Tool Installation and Management for gzh-manager-go
# Development tools, linters, formatters, and utilities

# ==============================================================================
# Tool Configuration
# ==============================================================================

# Colors for output
CYAN := \\033[36m
GREEN := \\033[32m
YELLOW := \\033[33m
RED := \\033[31m
BLUE := \\033[34m
RESET := \\033[0m

# ==============================================================================
# Core Tool Installation
# ==============================================================================

.PHONY: install-tools install-format-tools install-analysis-tools install-goreleaser
.PHONY: install-golangci-lint install-pre-commit-tools install-docs-tools

install-tools: install-format-tools install-analysis-tools install-golangci-lint install-goreleaser ## install all development tools
	@echo "$(GREEN)✅ All development tools installed!$(RESET)"

install-format-tools: ## install advanced formatting tools
	@echo "$(CYAN)Installing formatting tools...$(RESET)"
	@which gofumpt > /dev/null || (echo "Installing gofumpt..." && go install mvdan.cc/gofumpt@latest)
	@which gci > /dev/null || (echo "Installing gci..." && go install github.com/daixiang0/gci@latest)
	@echo "$(GREEN)✅ All formatting tools installed!$(RESET)"

install-analysis-tools: ## install code analysis tools
	@echo "$(CYAN)Installing code analysis tools...$(RESET)"
	@command -v gocyclo >/dev/null 2>&1 || { echo "Installing gocyclo..." && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; }
	@command -v ineffassign >/dev/null 2>&1 || { echo "Installing ineffassign..." && go install github.com/gordonklaus/ineffassign@latest; }
	@command -v dupl >/dev/null 2>&1 || { echo "Installing dupl..." && go install github.com/mibk/dupl@latest; }
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest; }
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@echo "$(GREEN)✅ All analysis tools installed!$(RESET)"

install-golangci-lint: ## install golangci-lint
	@echo "$(CYAN)Installing golangci-lint...$(RESET)"
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	@echo "$(GREEN)✅ golangci-lint installed!$(RESET)"

install-goreleaser: ## install goreleaser
	@echo "$(CYAN)Installing goreleaser...$(RESET)"
	@go install github.com/goreleaser/goreleaser@latest
	@echo "$(GREEN)✅ goreleaser installed$(RESET)"

# ==============================================================================
# Mock and Generation Tools
# ==============================================================================

.PHONY: install-mock-tools generate-mocks clean-mocks regenerate-mocks

install-mock-tools: ## install mock generation tools
	@echo "$(CYAN)Installing mock generation tools...$(RESET)"
	@command -v mockgen >/dev/null 2>&1 || { echo "Installing mockgen..." && go install go.uber.org/mock/mockgen@latest; }
	@echo "$(GREEN)✅ Mock generation tools installed!$(RESET)"

generate-mocks: install-mock-tools ## generate all mock files using gomock
	@echo "$(CYAN)Generating mocks...$(RESET)"
	@echo "Generating GitHub interface mocks..."
	@if [ -f "pkg/github/interfaces.go" ]; then \
		mockgen -source=pkg/github/interfaces.go -destination=pkg/github/mocks/github_mocks.go -package=mocks; \
		echo "  ✅ GitHub mocks generated"; \
	else \
		echo "  ⚠️  pkg/github/interfaces.go not found"; \
	fi
	@echo "Generating filesystem interface mocks..."
	@if [ -f "internal/filesystem/interfaces.go" ]; then \
		mockgen -source=internal/filesystem/interfaces.go -destination=internal/filesystem/mocks/filesystem_mocks.go -package=mocks; \
		echo "  ✅ Filesystem mocks generated"; \
	else \
		echo "  ⚠️  internal/filesystem/interfaces.go not found"; \
	fi
	@echo "Generating HTTP client interface mocks..."
	@if [ -f "internal/httpclient/interfaces.go" ]; then \
		mockgen -source=internal/httpclient/interfaces.go -destination=internal/httpclient/mocks/httpclient_mocks.go -package=mocks; \
		echo "  ✅ HTTP client mocks generated"; \
	else \
		echo "  ⚠️  internal/httpclient/interfaces.go not found"; \
	fi
	@echo "Generating Git interface mocks..."
	@if [ -f "internal/git/interfaces.go" ]; then \
		mockgen -source=internal/git/interfaces.go -destination=internal/git/mocks/git_mocks.go -package=mocks; \
		echo "  ✅ Git mocks generated"; \
	else \
		echo "  ⚠️  internal/git/interfaces.go not found"; \
	fi
	@echo "$(GREEN)✅ Mock generation complete!$(RESET)"

clean-mocks: ## remove all generated mock files
	@echo "$(CYAN)Cleaning generated mocks...$(RESET)"
	@rm -f pkg/github/mocks/github_mocks.go
	@rm -f internal/filesystem/mocks/filesystem_mocks.go
	@rm -f internal/httpclient/mocks/httpclient_mocks.go
	@rm -f internal/git/mocks/git_mocks.go
	@echo "$(GREEN)✅ Mock cleanup complete!$(RESET)"

regenerate-mocks: clean-mocks generate-mocks ## clean and regenerate all mocks

# ==============================================================================
# Pre-commit and Git Hooks
# ==============================================================================

.PHONY: install-pre-commit-tools pre-commit-install pre-commit-update

install-pre-commit-tools: ## install pre-commit and related tools
	@echo "$(CYAN)Installing pre-commit tools...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; }
	@echo "$(GREEN)✅ Pre-commit tools ready!$(RESET)"

pre-commit-install: install-pre-commit-tools ## install pre-commit hooks
	@echo "$(CYAN)Installing pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	@if [ -f "./scripts/setup-git-hooks.sh" ]; then \
		./scripts/setup-git-hooks.sh; \
	else \
		pre-commit install --hook-type pre-commit --hook-type commit-msg --hook-type pre-push; \
	fi
	@echo "$(GREEN)✅ Pre-commit hooks installed!$(RESET)"


# ==============================================================================
# Documentation Tools
# ==============================================================================

.PHONY: install-docs-tools

install-docs-tools: ## install documentation tools
	@echo "$(CYAN)Installing documentation tools...$(RESET)"
	@go install golang.org/x/perf/cmd/benchstat@latest
	@go install golang.org/x/tools/cmd/godoc@latest
	@which git-chglog >/dev/null 2>&1 || echo "$(YELLOW)Consider installing git-chglog for changelog generation$(RESET)"
	@which mkdocs >/dev/null 2>&1 || echo "$(YELLOW)Consider installing mkdocs for documentation: pip install mkdocs mkdocs-material$(RESET)"
	@echo "$(GREEN)✅ Documentation tools installed$(RESET)"

# ==============================================================================
# Security Tools
# ==============================================================================

.PHONY: install-security-tools

install-security-tools: ## install security analysis tools
	@echo "$(CYAN)Installing security tools...$(RESET)"
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@echo "$(GREEN)✅ Security tools installed!$(RESET)"

# ==============================================================================
# Vulnerability Scanning
# ==============================================================================

.PHONY: install-vuln-tools

install-vuln-tools: ## install vulnerability scanning tools
	@echo "$(CYAN)Installing vulnerability scanning tools...$(RESET)"
	@echo "govulncheck is available as: go run golang.org/x/vuln/cmd/govulncheck@latest"
	@echo "$(GREEN)✅ Vulnerability tools ready!$(RESET)"

# ==============================================================================
# Tool Status and Information
# ==============================================================================

.PHONY: tools-status tools-info

tools-status: ## show installed tool status
	@echo "$(CYAN)Checking development tool status...$(RESET)"
	@echo ""
	@echo "$(GREEN)📦 Core Tools:$(RESET)"
	@printf "  %-20s " "go:"; go version 2>/dev/null | cut -d' ' -f3 || echo "$(RED)Not found$(RESET)"
	@printf "  %-20s " "git:"; git --version 2>/dev/null | cut -d' ' -f3 || echo "$(RED)Not found$(RESET)"
	@echo ""
	@echo "$(GREEN)🔧 Build Tools:$(RESET)"
	@printf "  %-20s " "goreleaser:"; goreleaser --version 2>/dev/null | head -1 | awk '{print $3}' || echo "$(RED)Not installed$(RESET)"
	@echo ""
	@echo "$(GREEN)✨ Format Tools:$(RESET)"
	@printf "  %-20s " "gofumpt:"; gofumpt --version 2>/dev/null || echo "$(RED)Not installed$(RESET)"
	@printf "  %-20s " "gci:"; gci --version 2>/dev/null || echo "$(RED)Not installed$(RESET)"
	@echo ""
	@echo "$(GREEN)🔍 Lint Tools:$(RESET)"
	@printf "  %-20s " "golangci-lint:"; golangci-lint --version 2>/dev/null | cut -d' ' -f4 || echo "$(RED)Not installed$(RESET)"
	@printf "  %-20s " "staticcheck:"; staticcheck -version 2>/dev/null || echo "$(RED)Not installed$(RESET)"
	@echo ""
	@echo "$(GREEN)🛡️  Security Tools:$(RESET)"
	@printf "  %-20s " "gosec:"; gosec -version 2>/dev/null || echo "$(RED)Not installed$(RESET)"
	@echo ""
	@echo "$(GREEN)🎭 Mock Tools:$(RESET)"
	@printf "  %-20s " "mockgen:"; mockgen --version 2>/dev/null || echo "$(RED)Not installed$(RESET)"
	@echo ""
	@echo "$(GREEN)🎣 Git Hooks:$(RESET)"
	@printf "  %-20s " "pre-commit:"; pre-commit --version 2>/dev/null || echo "$(RED)Not installed$(RESET)"

tools-info: ## show comprehensive tool information
	@echo "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                         $(YELLOW)Development Tools Information$(CYAN)                   ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo "$(RESET)"
	@echo "$(GREEN)🔧 Available Tool Categories:$(RESET)"
	@echo "  • $(CYAN)Format Tools$(RESET)        Code formatting (gofumpt, gci)"
	@echo "  • $(CYAN)Analysis Tools$(RESET)      Static analysis (staticcheck, gosec)"
	@echo "  • $(CYAN)Lint Tools$(RESET)          Code linting (golangci-lint)"
	@echo "  • $(CYAN)Build Tools$(RESET)         Build and release (goreleaser)"
	@echo "  • $(CYAN)Mock Tools$(RESET)          Mock generation (mockgen)"
	@echo "  • $(CYAN)Security Tools$(RESET)      Security scanning (gosec, govulncheck)"
	@echo "  • $(CYAN)Git Hooks$(RESET)           Pre-commit hooks and validation"
	@echo "  • $(CYAN)Documentation$(RESET)       Documentation tools (godoc, benchstat)"
	@echo ""
	@echo "$(GREEN)🚀 Quick Installation:$(RESET)"
	@echo "  $(CYAN)make install-tools$(RESET)        Install all development tools"
	@echo "  $(CYAN)make tools-status$(RESET)         Check current tool installation status"
	@echo ""
	@echo "$(GREEN)💡 Individual Categories:$(RESET)"
	@echo "  $(CYAN)make install-format-tools$(RESET)     Format tools only"
	@echo "  $(CYAN)make install-analysis-tools$(RESET)   Analysis tools only"
	@echo "  $(CYAN)make install-security-tools$(RESET)   Security tools only"
	@echo "  $(CYAN)make install-mock-tools$(RESET)       Mock generation tools only"
