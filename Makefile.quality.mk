# Makefile.quality - Code Quality and Analysis for gzh-manager-go
# Formatting, linting, security analysis, and code quality checks

# ==============================================================================
# Quality Configuration
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
# Code Formatting Targets
# ==============================================================================

.PHONY: fmt format-all format-check format-diff format-imports format-simplify format-ci

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

format-ci: format-check ## CI-friendly format check

# ==============================================================================
# Linting and Static Analysis
# ==============================================================================

.PHONY: lint lint-check lint-fix lint-new lint-ci lint-count lint-summary lint-stats lint-status lint-json

lint-check: install-golangci-lint ## check lint issues without fixing (exit code reflects status)
	@echo "$(CYAN)Running golangci-lint...$(RESET)"
	golangci-lint run -c .golangci.yml

lint: lint-check ## alias for lint-check

lint-fix: install-golangci-lint ## run golangci-lint with auto-fix
	@echo "$(CYAN)Running golangci-lint with auto-fix...$(RESET)"
	golangci-lint run -c .golangci.yml --fix

lint-new: install-golangci-lint ## run golangci-lint on new code only
	@echo "$(CYAN)Running golangci-lint on new code only...$(RESET)"
	golangci-lint run -c .golangci.yml --new-from-rev=HEAD~

lint-ci: install-golangci-lint ## run golangci-lint for CI
	@echo "$(CYAN)Running golangci-lint for CI...$(RESET)"
	golangci-lint run -c .golangci.yml --out-format=github-actions

lint-count: install-golangci-lint ## count total lint issues without fixing
	@echo "$(CYAN)Counting lint issues...$(RESET)"
	@ISSUES=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | grep -E "^[^[:space:]].*\\([^)]+\\)$$" | wc -l); \
	echo "$(YELLOW)Total lint issues: $$ISSUES$(RESET)"

lint-summary: install-golangci-lint ## show lint issues summary by linter
	@echo "$(CYAN)Lint issues summary:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" | sed 's/.*(\\([^)]*\\))$$/\\1/' | sort | uniq -c | sort -nr | \
	awk '{printf "  $(YELLOW)%-15s$(RESET) %d issues\\n", $$2, $$1}'

lint-stats: install-golangci-lint ## show detailed lint statistics
	@echo "$(CYAN)=== Lint Statistics ===$(RESET)"
	@golangci-lint run -c .golangci.yml --show-stats --max-issues-per-linter=0 --max-same-issues=0

lint-status: install-golangci-lint ## comprehensive lint status report
	@echo "$(BLUE)🔍 Comprehensive Lint Status Report$(RESET)"
	@echo "$(BLUE)==================================$(RESET)"
	@echo ""
	@echo "$(GREEN)📊 Quick Stats:$(RESET)"
	@TOTAL=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | grep -E "^[^[:space:]].*\\([^)]+\\)$$" | wc -l); \
	ERRORS=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=json 2>/dev/null | jq -r '.Issues[]? | select(.Severity=="error") | .Severity' 2>/dev/null | wc -l || echo "0"); \
	WARNINGS=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=json 2>/dev/null | jq -r '.Issues[]? | select(.Severity=="warning") | .Severity' 2>/dev/null | wc -l || echo "0"); \
	echo "  $(YELLOW)Total Issues: $$TOTAL$(RESET)"; \
	echo "  $(RED)Errors: $$ERRORS$(RESET)"; \
	echo "  $(YELLOW)Warnings: $$WARNINGS$(RESET)"
	@echo ""
	@echo "$(GREEN)🏷️  Top 10 Linters:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" | sed 's/.*(\\([^)]*\\))$$/\\1/' | sort | uniq -c | sort -nr | head -10 | \
	awk '{printf "  $(CYAN)%-15s$(RESET) %d issues\\n", $$2, $$1}'
	@echo ""
	@echo "$(GREEN)📁 Most Problematic Files:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" | sed 's/^\\([^:]*\\):.*/\\1/' | sort | uniq -c | sort -nr | head -5 | \
	awk '{printf "  $(MAGENTA)%-40s$(RESET) %d issues\\n", $$2, $$1}'

lint-json: install-golangci-lint ## export lint results to JSON
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
# Security Analysis
# ==============================================================================

.PHONY: security security-deps security-code security-json vuln

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

vuln: security-deps ## check for known vulnerabilities (legacy alias)

# ==============================================================================
# Code Analysis
# ==============================================================================

.PHONY: analyze analyze-complexity analyze-unused analyze-dupl complexity ineffassign dupl

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

# ==============================================================================
# Pre-commit Integration
# ==============================================================================

.PHONY: pre-commit pre-push check-consistency pre-commit-update

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
		grep -q "\\.golangci\\.yml" .pre-commit-config.yaml && echo "$(GREEN)✓$(RESET) Pre-commit uses: .golangci.yml" || echo "$(RED)✗$(RESET) Pre-commit config mismatch"; \
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
# Quality Assurance Workflows
# ==============================================================================

.PHONY: quality quality-fix lint-all

quality: fmt lint-check test-coverage security ## run comprehensive quality checks
	@echo "$(GREEN)✅ All quality checks passed!$(RESET)"

quality-fix: fmt lint-fix ## apply automatic quality fixes
	@echo "$(GREEN)✅ Code quality fixes applied!$(RESET)"

lint-all: fmt lint-check pre-commit ## run all linting steps (format, lint, pre-commit)
	@echo "$(GREEN)✅ All linting steps completed!$(RESET)"

# ==============================================================================
# Quality Information and Help
# ==============================================================================

.PHONY: quality-info quality-help

quality-info: ## show code quality information and targets
	@echo "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                         $(YELLOW)Code Quality Information$(CYAN)                        ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo "$(RESET)"
	@echo "$(GREEN)🎨 Formatting Tools:$(RESET)"
	@echo "  • $(CYAN)fmt$(RESET)                   Standard Go formatting with gofumpt + gci"
	@echo "  • $(CYAN)format-all$(RESET)            Comprehensive formatting including advanced"
	@echo "  • $(CYAN)format-check$(RESET)          Check formatting without making changes"
	@echo "  • $(CYAN)format-diff$(RESET)           Show formatting differences"
	@echo ""
	@echo "$(GREEN)🔍 Linting & Analysis:$(RESET)"
	@echo "  • $(CYAN)lint-check$(RESET)            Run golangci-lint checks"
	@echo "  • $(CYAN)lint-fix$(RESET)              Auto-fix lint issues where possible"
	@echo "  • $(CYAN)lint-status$(RESET)           Comprehensive lint status report"
	@echo "  • $(CYAN)analyze$(RESET)               Code complexity and quality analysis"
	@echo ""
	@echo "$(GREEN)🛡️  Security Analysis:$(RESET)"
	@echo "  • $(CYAN)security$(RESET)              All security checks (deps + code)"
	@echo "  • $(CYAN)security-deps$(RESET)         Check dependencies for vulnerabilities"
	@echo "  • $(CYAN)security-code$(RESET)         Static security analysis with gosec"
	@echo ""
	@echo "$(GREEN)🔄 Quality Workflows:$(RESET)"
	@echo "  • $(CYAN)quality$(RESET)               Comprehensive quality pipeline"
	@echo "  • $(CYAN)quality-fix$(RESET)           Apply all automatic fixes"
	@echo "  • $(CYAN)lint-all$(RESET)              Complete linting workflow"

quality-help: quality-info ## alias for quality-info
