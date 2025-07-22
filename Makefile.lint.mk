# Makefile.lint.mk - Enhanced Code Quality and Linting for gzh-manager-go
# Upgraded with advanced features, better organization, and comprehensive tooling

# ==============================================================================
# Code Quality Configuration
# ==============================================================================

.PHONY: fmt lint format security security-json analyze analyze-complexity analyze-unused
.PHONY: quality quality-fix lint-fix lint-new lint-ci format-all format-check format-diff
.PHONY: format-imports format-simplify format-ci install-golangci-lint install-format-tools
.PHONY: generate-mocks clean-mocks regenerate-mocks pre-commit-install pre-commit pre-push
.PHONY: lint-all check-consistency pre-commit-update complexity ineffassign dupl vuln
.PHONY: dev dev-fast verify ci-local pr-check lint-count lint-summary lint-stats lint-status
.PHONY: lint-json lint-help install-analysis-tools

# Colors for output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
BLUE := \033[34m
MAGENTA := \033[35m
RESET := \033[0m

# ==============================================================================
# Enhanced Code Formatting
# ==============================================================================

fmt: ## format go files with gofumpt and gci
	@echo "$(CYAN)Formatting Go code...$(RESET)"
	@echo "1. Running gofumpt..."
	@gofumpt -w .
	@echo "2. Running gci (import organization)..."
	@gci write --skip-generated .
	@echo "$(GREEN)✅ Code formatting complete!$(RESET)"

format-all: install-format-tools ## run all formatters including advanced ones
	@echo "$(CYAN)Running comprehensive code formatting...$(RESET)"
	@echo "1. Standard formatting..."
	@gofmt -w .
	@echo "2. Simplifying code..."
	@gofmt -s -w .
	@echo "3. Running gofumpt (strict formatting)..."
	@gofumpt -w -extra .
	@echo "4. Running gci (import grouping)..."
	@gci write --skip-generated -s standard -s default -s "prefix(github.com/gizzahub/gzh-manager-go)" .
	@echo "$(GREEN)✅ All formatting complete!$(RESET)"

format-check: ## check code formatting without fixing
	@echo "$(CYAN)Checking code formatting...$(RESET)"
	@if [ -n "$$(gofumpt -l .)" ]; then \
		echo "$(RED)❌ The following files need formatting:$(RESET)"; \
		gofumpt -l .; \
		echo "$(YELLOW)Run 'make fmt' to fix.$(RESET)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ All files are properly formatted$(RESET)"; \
	fi

format-diff: ## show formatting differences
	@echo "$(CYAN)Showing formatting differences...$(RESET)"
	@gofumpt -d .

format-imports: ## organize imports only
	@echo "$(CYAN)Organizing imports...$(RESET)"
	@gci write --skip-generated .
	@echo "$(GREEN)✅ Imports organized!$(RESET)"

format-simplify: ## simplify code with gofmt -s
	@echo "$(CYAN)Simplifying code...$(RESET)"
	@gofmt -s -w .
	@echo "$(GREEN)✅ Code simplified!$(RESET)"

install-format-tools: ## install advanced formatting tools
	@echo "$(CYAN)Installing formatting tools...$(RESET)"
	@which gofumpt > /dev/null || (echo "Installing gofumpt..." && go install mvdan.cc/gofumpt@latest)
	@which gci > /dev/null || (echo "Installing gci..." && go install github.com/daixiang0/gci@latest)
	@echo "$(GREEN)✅ All formatting tools installed!$(RESET)"

format-ci: format-check ## CI-friendly format check
	@echo "$(GREEN)✅ CI format check passed!$(RESET)"

# ==============================================================================
# Enhanced Linting with Advanced Reporting
# ==============================================================================

install-golangci-lint: ## install golangci-lint
	@echo "$(CYAN)Installing golangci-lint...$(RESET)"
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	@echo "$(GREEN)✅ golangci-lint installed!$(RESET)"

lint-check: install-golangci-lint ## check lint issues without fixing (exit code reflects status)
	@echo "$(CYAN)Running golangci-lint...$(RESET)"
	golangci-lint run -c .golangci.yml

lint: lint-check ## alias for lint-check

lint-fix: install-golangci-lint ## run golangci-lint with auto-fix
	@echo "$(CYAN)Running golangci-lint with auto-fix...$(RESET)"
	golangci-lint run -c .golangci.yml --fix

format: lint-fix ## format go files (alias for lint-fix)

lint-new: install-golangci-lint ## run golangci-lint on new code only
	@echo "$(CYAN)Running golangci-lint on new code only...$(RESET)"
	golangci-lint run -c .golangci.yml --new-from-rev=HEAD~

lint-ci: install-golangci-lint ## run golangci-lint for CI
	@echo "$(CYAN)Running golangci-lint for CI...$(RESET)"
	golangci-lint run -c .golangci.yml --out-format=github-actions

# ==============================================================================
# Advanced Lint Reporting and Statistics
# ==============================================================================

lint-count: install-golangci-lint ## count total lint issues without fixing
	@echo "$(CYAN)Counting lint issues...$(RESET)"
	@ISSUES=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | grep -E "^[^[:space:]].*\([^)]+\)$$" | wc -l); \
	echo "$(YELLOW)Total lint issues: $$ISSUES$(RESET)"

lint-summary: install-golangci-lint ## show lint issues summary by linter
	@echo "$(CYAN)Lint issues summary:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\([^)]+\)$$" | sed 's/.*(\([^)]*\))$$/\1/' | sort | uniq -c | sort -nr | \
	awk '{printf "  $(YELLOW)%-15s$(RESET) %d issues\n", $$2, $$1}'

lint-stats: install-golangci-lint ## show detailed lint statistics with golangci-lint built-in stats
	@echo "$(CYAN)=== Lint Statistics ===$(RESET)"
	@golangci-lint run -c .golangci.yml --show-stats --max-issues-per-linter=0 --max-same-issues=0

lint-status: install-golangci-lint ## comprehensive lint status report
	@echo "$(BLUE)🔍 Comprehensive Lint Status Report$(RESET)"
	@echo "$(BLUE)==================================$(RESET)"
	@echo ""
	@echo "$(GREEN)📊 Quick Stats:$(RESET)"
	@TOTAL=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | grep -E "^[^[:space:]].*\([^)]+\)$$" | wc -l); \
	ERRORS=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=json 2>/dev/null | jq -r '.Issues[]? | select(.Severity=="error") | .Severity' 2>/dev/null | wc -l || echo "0"); \
	WARNINGS=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=json 2>/dev/null | jq -r '.Issues[]? | select(.Severity=="warning") | .Severity' 2>/dev/null | wc -l || echo "0"); \
	echo "  $(YELLOW)Total Issues: $$TOTAL$(RESET)"; \
	echo "  $(RED)Errors: $$ERRORS$(RESET)"; \
	echo "  $(YELLOW)Warnings: $$WARNINGS$(RESET)"
	@echo ""
	@echo "$(GREEN)🏷️  Top 10 Linters:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\([^)]+\)$$" | sed 's/.*(\([^)]*\))$$/\1/' | sort | uniq -c | sort -nr | head -10 | \
	awk '{printf "  $(CYAN)%-15s$(RESET) %d issues\n", $$2, $$1}'
	@echo ""
	@echo "$(GREEN)📁 Most Problematic Files:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\([^)]+\)$$" | sed 's/^\([^:]*\):.*/\1/' | sort | uniq -c | sort -nr | head -5 | \
	awk '{printf "  $(MAGENTA)%-40s$(RESET) %d issues\n", $$2, $$1}'

lint-json: install-golangci-lint ## export lint results to JSON for further analysis
	@echo "$(CYAN)Exporting lint results to lint-report.json...$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=json > lint-report.json 2>/dev/null || true
	@echo "$(GREEN)✅ Report saved to lint-report.json$(RESET)"
	@if command -v jq >/dev/null 2>&1; then \
		echo ""; \
		echo "$(YELLOW)📈 JSON Report Summary:$(RESET)"; \
		echo "  Total Issues: $$(jq '.Issues | length' lint-report.json 2>/dev/null || echo '0')"; \
		echo "  Unique Files: $$(jq -r '.Issues[]? | .Pos.Filename' lint-report.json 2>/dev/null | sort | uniq | wc -l || echo '0')"; \
	fi

# ==============================================================================
# Enhanced Security Analysis
# ==============================================================================

security: security-deps security-code ## run all security checks
	@echo "$(GREEN)✅ Security checks completed!$(RESET)"

security-deps: ## check dependencies for vulnerabilities
	@echo "$(CYAN)Checking dependencies for vulnerabilities...$(RESET)"
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./... || echo "$(RED)❌ Vulnerabilities found$(RESET)"

security-code: ## run security code analysis
	@echo "$(CYAN)Running security code analysis with gosec...$(RESET)"
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@gosec -config=.gosec.yaml ./... 2>/dev/null || echo "$(YELLOW)No gosec config found, using defaults$(RESET)"

security-json: ## run security analysis and output JSON report
	@echo "$(CYAN)Running security analysis with JSON output...$(RESET)"
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@gosec -fmt=json -out=gosec-report.json -config=.gosec.yaml ./... 2>/dev/null || \
		gosec -fmt=json -out=gosec-report.json ./... 2>/dev/null || true
	@echo "$(GREEN)✅ Security report generated: gosec-report.json$(RESET)"

# ==============================================================================
# Enhanced Code Analysis
# ==============================================================================

install-analysis-tools: ## install code analysis tools
	@echo "$(CYAN)Installing code analysis tools...$(RESET)"
	@command -v gocyclo >/dev/null 2>&1 || { echo "Installing gocyclo..." && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; }
	@command -v ineffassign >/dev/null 2>&1 || { echo "Installing ineffassign..." && go install github.com/gordonklaus/ineffassign@latest; }
	@command -v dupl >/dev/null 2>&1 || { echo "Installing dupl..." && go install github.com/mibk/dupl@latest; }
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest; }
	@echo "$(GREEN)✅ All analysis tools installed!$(RESET)"

analyze: analyze-complexity analyze-unused analyze-dupl ## run comprehensive code analysis
	@echo "$(GREEN)✅ Code analysis complete!$(RESET)"

analyze-complexity: ## analyze code complexity
	@echo "$(CYAN)Analyzing code complexity...$(RESET)"
	@command -v gocyclo >/dev/null 2>&1 || { echo "Installing gocyclo..." && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; }
	@gocyclo -over 10 -avg .

analyze-unused: ## find unused code
	@echo "$(CYAN)Finding unused code...$(RESET)"
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest; }
	@staticcheck -checks U1000 ./...

analyze-dupl: ## find duplicate code
	@echo "$(CYAN)Checking for duplicate code...$(RESET)"
	@command -v dupl >/dev/null 2>&1 || { echo "Installing dupl..." && go install github.com/mibk/dupl@latest; }
	@dupl -threshold 50 .

# Legacy aliases for backward compatibility
complexity: analyze-complexity ## analyze code complexity (legacy alias)
ineffassign: ## detect ineffectual assignments (legacy)
	@echo "$(CYAN)Checking for ineffectual assignments...$(RESET)"
	@command -v ineffassign >/dev/null 2>&1 || { echo "Installing ineffassign..." && go install github.com/gordonklaus/ineffassign@latest; }
	@ineffassign ./...

dupl: analyze-dupl ## find duplicate code (legacy alias)

vuln: security-deps ## check for known vulnerabilities (legacy alias)

# ==============================================================================
# Enhanced Mock Generation
# ==============================================================================

generate-mocks: ## generate all mock files using gomock
	@echo "$(CYAN)Generating mocks...$(RESET)"
	@command -v mockgen >/dev/null 2>&1 || { echo "Installing mockgen..." && go install go.uber.org/mock/mockgen@latest; }
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
# Enhanced Pre-commit Integration
# ==============================================================================

pre-commit-install: ## install pre-commit hooks
	@echo "$(CYAN)Installing pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	@if [ -f "./scripts/setup-git-hooks.sh" ]; then \
		./scripts/setup-git-hooks.sh; \
	else \
		pre-commit install --hook-type pre-commit --hook-type commit-msg --hook-type pre-push; \
	fi
	@echo "$(GREEN)✅ Pre-commit hooks installed!$(RESET)"

pre-commit: ## run pre-commit hooks (format + light checks)
	@echo "$(CYAN)Running pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit run --all-files

pre-push: ## run pre-push hooks (comprehensive checks)
	@echo "$(CYAN)Running pre-push hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit run --all-files --hook-stage pre-push

check-consistency: ## verify lint configuration consistency
	@echo "$(CYAN)Checking lint configuration consistency...$(RESET)"
	@echo "$(GREEN)✓$(RESET) Makefile uses: .golangci.yml"
	@if [ -f ".pre-commit-config.yaml" ]; then \
		grep -q "\.golangci\.yml" .pre-commit-config.yaml && echo "$(GREEN)✓$(RESET) Pre-commit uses: .golangci.yml" || echo "$(RED)✗$(RESET) Pre-commit config mismatch"; \
	else \
		echo "$(YELLOW)⚠$(RESET) No pre-commit config found"; \
	fi
	@echo "$(GREEN)✅ Configuration consistency checked$(RESET)"

pre-commit-update: ## update pre-commit hooks to latest versions
	@echo "$(CYAN)Updating pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit autoupdate
	@echo "$(GREEN)✅ Pre-commit hooks updated!$(RESET)"

# ==============================================================================
# Enhanced Quality Assurance Workflow Targets
# ==============================================================================

quality: fmt lint-check test-coverage security ## run comprehensive quality checks
	@echo "$(GREEN)✅ All quality checks passed!$(RESET)"

quality-fix: fmt lint-fix ## apply automatic quality fixes
	@echo "$(GREEN)✅ Code quality fixes applied!$(RESET)"

lint-all: fmt lint-check pre-commit ## run all linting steps (format, lint, pre-commit)
	@echo "$(GREEN)✅ All linting steps completed!$(RESET)"

# ==============================================================================
# Enhanced Development Workflow Targets
# ==============================================================================

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
# Enhanced Help System
# ==============================================================================

lint-help: ## show comprehensive help for linting targets
	@echo "$(BLUE)Code Quality and Linting Commands:$(RESET)"
	@echo ""
	@echo "$(YELLOW)🎨 Formatting:$(RESET)"
	@echo "  $(CYAN)fmt$(RESET)                   Format Go files with gofumpt and gci"
	@echo "  $(CYAN)format-all$(RESET)            Run all formatters including advanced ones"
	@echo "  $(CYAN)format-check$(RESET)          Check code formatting without fixing"
	@echo "  $(CYAN)format-diff$(RESET)           Show formatting differences"
	@echo "  $(CYAN)format-imports$(RESET)        Organize imports only"
	@echo "  $(CYAN)format-simplify$(RESET)       Simplify code with gofmt -s"
	@echo ""
	@echo "$(YELLOW)🔍 Linting:$(RESET)"
	@echo "  $(CYAN)lint$(RESET)                  Check lint issues without fixing"
	@echo "  $(CYAN)lint-fix$(RESET)              Run golangci-lint with auto-fix"
	@echo "  $(CYAN)lint-new$(RESET)              Run golangci-lint on new code only"
	@echo "  $(CYAN)lint-ci$(RESET)               Run golangci-lint for CI"
	@echo "  $(CYAN)lint-count$(RESET)            Count total lint issues"
	@echo "  $(CYAN)lint-summary$(RESET)          Show lint issues summary by linter"
	@echo "  $(CYAN)lint-stats$(RESET)            Show detailed lint statistics"
	@echo "  $(CYAN)lint-status$(RESET)           Comprehensive lint status report"
	@echo "  $(CYAN)lint-json$(RESET)             Export lint results to JSON"
	@echo ""
	@echo "$(YELLOW)🔒 Security Analysis:$(RESET)"
	@echo "  $(CYAN)security$(RESET)              Run all security checks"
	@echo "  $(CYAN)security-deps$(RESET)         Check dependencies for vulnerabilities"
	@echo "  $(CYAN)security-code$(RESET)         Run security code analysis with gosec"
	@echo "  $(CYAN)security-json$(RESET)         Security analysis with JSON output"
	@echo ""
	@echo "$(YELLOW)📊 Code Analysis:$(RESET)"
	@echo "  $(CYAN)analyze$(RESET)               Run comprehensive code analysis"
	@echo "  $(CYAN)analyze-complexity$(RESET)    Analyze code complexity"
	@echo "  $(CYAN)analyze-unused$(RESET)        Find unused code"
	@echo "  $(CYAN)analyze-dupl$(RESET)          Find duplicate code"
	@echo ""
	@echo "$(YELLOW)🔧 Mock Generation:$(RESET)"
	@echo "  $(CYAN)generate-mocks$(RESET)        Generate all mock files using gomock"
	@echo "  $(CYAN)clean-mocks$(RESET)           Remove all generated mock files"
	@echo "  $(CYAN)regenerate-mocks$(RESET)      Clean and regenerate all mocks"
	@echo ""
	@echo "$(YELLOW)🎣 Pre-commit Integration:$(RESET)"
	@echo "  $(CYAN)pre-commit-install$(RESET)    Install pre-commit hooks"
	@echo "  $(CYAN)pre-commit$(RESET)            Run pre-commit hooks"
	@echo "  $(CYAN)pre-push$(RESET)              Run pre-push hooks"
	@echo "  $(CYAN)pre-commit-update$(RESET)     Update pre-commit hooks"
	@echo "  $(CYAN)check-consistency$(RESET)     Verify lint configuration consistency"
	@echo ""
	@echo "$(YELLOW)🔄 Development Workflows:$(RESET)"
	@echo "  $(CYAN)dev$(RESET)                   Standard development workflow"
	@echo "  $(CYAN)dev-fast$(RESET)              Quick development cycle"
	@echo "  $(CYAN)verify$(RESET)                Complete verification before PR"
	@echo "  $(CYAN)ci-local$(RESET)              Run full CI pipeline locally"
	@echo "  $(CYAN)pr-check$(RESET)              Pre-PR submission check"
	@echo "  $(CYAN)quality$(RESET)               Run comprehensive quality checks"
	@echo "  $(CYAN)quality-fix$(RESET)           Apply automatic quality fixes"
	@echo "  $(CYAN)lint-all$(RESET)              Run all linting steps"
	@echo ""
	@echo "$(YELLOW)📁 Configuration Files:$(RESET)"
	@echo "  .golangci.yml             golangci-lint configuration"
	@echo "  .pre-commit-config.yaml   Pre-commit hooks configuration"
	@echo "  .gosec.yaml              gosec security scanner configuration"
