# Makefile.quality - Code Quality and Analysis for gzh-manager-go
# Formatting, linting, security analysis, and code quality checks

# ==============================================================================
# Quality Configuration
# ==============================================================================

.PHONY: fmt format format-all format-check format-diff format-imports format-simplify format-ci format-strict format-list format-file format-install-tools
.PHONY: pre-commit-install dev dev-fast verify ci-local pr-check lint-help

# ==============================================================================
# Code Formatting Targets
# ==============================================================================

format: format-simplify ## quick and simple formatting (default)
fmt: format-simplify

format-simplify: ## quick basic formatting with gofumpt and goimports
	@echo -e "$(CYAN)🚀 Quick formatting...$(RESET)"
	@echo "1. Running gofumpt (includes go fmt + simplification)..."
	@gofumpt -w .
	@echo "2. Organizing imports..."
	@goimports -w -local github.com/gizzahub/gzh-manager-go .
	@echo -e "$(GREEN)✅ Quick formatting complete!$(RESET)"

format-strict: format-install-tools ## comprehensive formatting with all tools
	@echo -e "$(CYAN)🔧 Strict formatting (all tools)...$(RESET)"
	@echo "1. Running gofumpt (strict formatting + simplification)..."
	@gofumpt -w -extra .
	@echo "2. Running gci (import organization)..."
	@gci write --skip-generated .
	@echo "3. Organizing imports with goimports..."
	@goimports -w -local github.com/gizzahub/gzh-manager-go .
	@echo "4. Final gci (import grouping)..."
	@gci write --skip-generated -s standard -s default -s "prefix(github.com/gizzahub/gzh-manager-go)" .
	@echo -e "$(GREEN)✅ Strict formatting complete!$(RESET)"

format-list: ## show files that need formatting
	@echo -e "$(CYAN)📋 Files that need formatting:$(RESET)"
	@FILES=$$(gofmt -l .); \
	if [ -n "$$FILES" ]; then \
		echo "$$FILES" | while read file; do echo "  $(YELLOW)$$file$(RESET)"; done; \
		echo ""; \
		echo -e "$(YELLOW)Total: $$(echo "$$FILES" | wc -l) files need formatting$(RESET)"; \
		echo -e "$(CYAN)Run 'make format-simplify' or 'make format-strict' to fix$(RESET)"; \
	else \
		echo -e "$(GREEN)✅ All files are properly formatted!$(RESET)"; \
	fi

format-diff: ## show formatting differences
	@echo -e "$(CYAN)📝 Formatting differences:$(RESET)"
	@DIFF_OUTPUT=$$(gofmt -d .); \
	if [ -n "$$DIFF_OUTPUT" ]; then \
		echo "$$DIFF_OUTPUT"; \
	else \
		echo -e "$(GREEN)✅ No formatting differences found!$(RESET)"; \
	fi

format-install-tools: ## install advanced formatting tools
	@echo -e "$(CYAN)Installing formatting tools...$(RESET)"
	@which goimports > /dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	@which gofumpt > /dev/null || (echo "Installing gofumpt..." && go install mvdan.cc/gofumpt@latest)
	@which gci > /dev/null || (echo "Installing gci..." && go install github.com/daixiang0/gci@latest)
	@echo -e "$(GREEN)✅ All formatting tools installed!$(RESET)"

format-file: ## format specific files with gofumpt and goimports (usage: make format-file file1.go file2.go ...)
	@if [ -z "$(MAKECMDGOALS)" ] || [ "$(words $(MAKECMDGOALS))" -eq 1 ]; then \
		echo "$(RED)❌ Error: At least one file must be specified$(RESET)"; \
		echo "$(YELLOW)Usage: make format-file file1.go file2.go ...$(RESET)"; \
		exit 1; \
	fi
	@echo "$(CYAN)🔄 Processing files...$(RESET)"
	@for file in $(filter-out format-file,$(MAKECMDGOALS)); do \
		if [ -n "$$file" ]; then \
			if [ ! -f "$$file" ]; then \
				echo "$(RED)❌ Error: File '$$file' does not exist$(RESET)"; \
				continue; \
			fi; \
			if ! echo "$$file" | grep -q "\.go$$"; then \
				echo "$(YELLOW)⚠️  Warning: File '$$file' is not a Go file (.go extension), skipping$(RESET)"; \
				continue; \
			fi; \
			echo "$(CYAN)📝 Formatting file: $$file$(RESET)"; \
			echo "  1. Running gofumpt..."; \
			gofumpt -w "$$file" || echo "$(RED)❌ gofumpt failed for $$file$(RESET)"; \
			echo "  2. Running goimports..."; \
			goimports -w -local github.com/gizzahub/gzh-manager-go "$$file" || echo "$(RED)❌ goimports failed for $$file$(RESET)"; \
			echo "$(GREEN)✅ File '$$file' formatted successfully!$(RESET)"; \
		fi; \
	done
	@echo "$(GREEN)🎉 All files processed!$(RESET)"

# ==============================================================================
# Linting and Static Analysis
# ==============================================================================

.PHONY: lint format lint-check lint-fix lint-new lint-ci lint-count lint-summary lint-stats lint-status lint-json

lint-check: install-golangci-lint ## check lint issues without fixing (exit code reflects status)
	@echo -e "$(CYAN)Running golangci-lint...$(RESET)"
	golangci-lint run -c .golangci.yml

lint: lint-check ## alias for lint-check

lint-fix: install-golangci-lint ## run golangci-lint with auto-fix
	@echo -e "$(CYAN)Running golangci-lint with auto-fix...$(RESET)"
	golangci-lint run -c .golangci.yml --fix

lint-new: install-golangci-lint ## run golangci-lint on new code only
	@echo -e "$(CYAN)Running golangci-lint on new code only...$(RESET)"
	golangci-lint run -c .golangci.yml --new-from-rev=HEAD~

lint-ci: install-golangci-lint ## run golangci-lint for CI
	@echo -e "$(CYAN)Running golangci-lint for CI...$(RESET)"
	golangci-lint run -c .golangci.yml --out-format=github-actions

lint-count: install-golangci-lint ## count total lint issues without fixing
	@echo -e "$(CYAN)Counting lint issues...$(RESET)"
	@ISSUES=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | grep -E "^[^[:space:]].*\\([^)]+\\)$$" | wc -l); \
	echo -e "$(YELLOW)Total lint issues: $$ISSUES$(RESET)"

lint-summary: install-golangci-lint ## show lint issues summary by linter
	@echo -e "$(CYAN)Lint issues summary:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" | sed 's/.*(\\([^)]*\\))$$/\\1/' | sort | uniq -c | sort -nr | \
	awk '{printf "  $(YELLOW)%-15s$(RESET) %d issues\\n", $$2, $$1}'

lint-stats: install-golangci-lint ## show detailed lint statistics with golangci-lint built-in stats
	@echo -e "$(CYAN)=== Lint Statistics ===$(RESET)"
	@golangci-lint run -c .golangci.yml --show-stats --max-issues-per-linter=0 --max-same-issues=0

lint-status: install-golangci-lint ## comprehensive lint status report
	@echo -e "$(BLUE)🔍 Comprehensive Lint Status Report$(RESET)"
	@echo -e "$(BLUE)==================================$(RESET)"
	@echo ""
	@echo -e "$(GREEN)📊 Quick Stats:$(RESET)"
	@TOTAL=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | grep -E "^[^[:space:]].*\\([^)]+\\)$$" | wc -l); \
	ERRORS=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=json 2>/dev/null | jq -r '.Issues[]? | select(.Severity=="error") | .Severity' 2>/dev/null | wc -l || echo "0"); \
	WARNINGS=$$(golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=json 2>/dev/null | jq -r '.Issues[]? | select(.Severity=="warning") | .Severity' 2>/dev/null | wc -l || echo "0"); \
	echo "  $(YELLOW)Total Issues: $$TOTAL$(RESET)"; \
	echo "  $(RED)Errors: $$ERRORS$(RESET)"; \
	echo "  $(YELLOW)Warnings: $$WARNINGS$(RESET)"
	@echo ""
	@echo -e "$(GREEN)🏷️  Top 10 Linters:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" | sed 's/.*(\\([^)]*\\))$$/\\1/' | sort | uniq -c | sort -nr | head -10 | \
	awk '{printf "  $(CYAN)%-15s$(RESET) %d issues\\n", $$2, $$1}'
	@echo ""
	@echo -e "$(GREEN)📁 Most Problematic Files:$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=line-number 2>/dev/null | \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" | sed 's/^\\([^:]*\\):.*/\\1/' | sort | uniq -c | sort -nr | head -5 | \
	awk '{printf "  $(MAGENTA)%-40s$(RESET) %d issues\\n", $$2, $$1}'

lint-json: install-golangci-lint ## export lint results to JSON for further analysis
	@echo -e "$(CYAN)Exporting lint results to lint-report.json...$(RESET)"
	@golangci-lint run -c .golangci.yml --max-issues-per-linter=0 --max-same-issues=0 --out-format=json > lint-report.json 2>/dev/null || true
	@echo -e "$(GREEN)✅ Report saved to lint-report.json$(RESET)"
	@if command -v jq >/dev/null 2>&1; then \
		echo ""; \
		echo "$(YELLOW)📈 JSON Report Summary:$(RESET)"; \
		echo "  Total Issues: $$(jq '.Issues | length' lint-report.json 2>/dev/null || echo '0')"; \
		echo "  Unique Files: $$(jq -r '.Issues[]? | .Pos.Filename' lint-report.json 2>/dev/null | sort | uniq | wc -l || echo '0')"; \
	fi

# ==============================================================================
# Enhanced Code Analysis
# ==============================================================================

# ==============================================================================
# Security Analysis
# ==============================================================================

.PHONY: security security-deps security-code security-json vuln

security: security-deps security-code ## run all security checks
	@echo -e "$(GREEN)✅ Security checks completed!$(RESET)"

security-deps: ## check dependencies for vulnerabilities
	@echo -e "$(CYAN)Checking dependencies for vulnerabilities...$(RESET)"
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./... || echo "$(RED)❌ Vulnerabilities found$(RESET)"

security-code: ## run security code analysis
	@echo -e "$(CYAN)Running security code analysis with gosec...$(RESET)"
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@gosec -config=.gosec.yaml ./... 2>/dev/null || echo "$(YELLOW)No gosec config found, using defaults$(RESET)"

security-json: ## run security analysis and output JSON report
	@echo -e "$(CYAN)Running security analysis with JSON output...$(RESET)"
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@gosec -fmt=json -out=gosec-report.json -config=.gosec.yaml ./... 2>/dev/null || \
		gosec -fmt=json -out=gosec-report.json ./... 2>/dev/null || true
	@echo -e "$(GREEN)✅ Security report generated: gosec-report.json$(RESET)"

# ==============================================================================
# Code Analysis
# ==============================================================================

.PHONY: analyze analyze-complexity analyze-unused analyze-dupl complexity ineffassign dupl

analyze: analyze-complexity analyze-unused analyze-dupl ## run comprehensive code analysis
	@echo -e "$(GREEN)✅ Code analysis complete!$(RESET)"

analyze-complexity: ## analyze code complexity
	@echo -e "$(CYAN)Analyzing code complexity...$(RESET)"
	@command -v gocyclo >/dev/null 2>&1 || { echo "Installing gocyclo..." && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; }
	@gocyclo -over 10 -avg .

analyze-unused: ## find unused code
	@echo -e "$(CYAN)Finding unused code...$(RESET)"
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest; }
	@staticcheck -checks U1000 ./...

analyze-dupl: ## find duplicate code
	@echo -e "$(CYAN)Checking for duplicate code...$(RESET)"
	@command -v dupl >/dev/null 2>&1 || { echo "Installing dupl..." && go install github.com/mibk/dupl@latest; }
	@dupl -threshold 50 .

# ==============================================================================
# Pre-commit Integration
# ==============================================================================

.PHONY: pre-commit-install pre-commit pre-push check-consistency pre-commit-update

pre-commit-install: ## install pre-commit hooks
	@echo -e "$(CYAN)Installing pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	@if [ -f "./scripts/setup-git-hooks.sh" ]; then \
		./scripts/setup-git-hooks.sh; \
	else \
		pre-commit install --hook-type pre-commit --hook-type commit-msg --hook-type pre-push; \
	fi
	@echo -e "$(GREEN)✅ Pre-commit hooks installed!$(RESET)"

pre-commit: ## run pre-commit hooks (format + light checks)
	@echo -e "$(CYAN)Running pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit run --all-files

pre-push: ## run pre-push hooks (comprehensive checks)
	@echo -e "$(CYAN)Running pre-push hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit run --all-files --hook-stage pre-push

check-consistency: ## verify lint configuration consistency
	@echo -e "$(CYAN)Checking lint configuration consistency...$(RESET)"
	@echo -e "$(GREEN)✓$(RESET) Makefile uses: .golangci.yml"
	@if [ -f ".pre-commit-config.yaml" ]; then \
		grep -q "\\.golangci\\.yml" .pre-commit-config.yaml && echo "$(GREEN)✓$(RESET) Pre-commit uses: .golangci.yml" || echo "$(RED)✗$(RESET) Pre-commit config mismatch"; \
	else \
		echo "$(YELLOW)⚠$(RESET) No pre-commit config found"; \
	fi
	@echo -e "$(GREEN)✅ Configuration consistency checked$(RESET)"

pre-commit-update: ## update pre-commit hooks to latest versions
	@echo -e "$(CYAN)Updating pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit autoupdate
	@echo -e "$(GREEN)✅ Pre-commit hooks updated!$(RESET)"

# ==============================================================================
# Quality Assurance Workflows
# ==============================================================================

.PHONY: quality quality-fix lint-all

quality: fmt lint-check test-coverage security ## run comprehensive quality checks
	@echo -e "$(GREEN)✅ All quality checks passed!$(RESET)"

quality-fix: fmt lint-fix ## apply automatic quality fixes
	@echo -e "$(GREEN)✅ Code quality fixes applied!$(RESET)"

lint-all: fmt lint-check pre-commit ## run all linting steps (format, lint, pre-commit)
	@echo -e "$(GREEN)✅ All linting steps completed!$(RESET)"

# ==============================================================================
# Quality Information and Help
# ==============================================================================

.PHONY: quality-info quality-help

quality-info: ## show code quality information and targets
	@echo -e "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo -e "║                         $(YELLOW)Code Quality Information$(CYAN)                        ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo -e "$(RESET)"
	@echo -e "$(GREEN)🎨 Formatting Tools:$(RESET)"
	@echo -e "  • $(CYAN)format$(RESET)                기본 포맷팅 (format-simplify 실행)"
	@echo -e "  • $(CYAN)format-simplify$(RESET)       신속한 기본 포맷팅"
	@echo -e "  • $(CYAN)format-strict$(RESET)         엄격한 포맷팅 (모든 도구 사용)"
	@echo -e "  • $(CYAN)format-list$(RESET)           포맷팅 필요한 파일 목록"
	@echo -e "  • $(CYAN)format-diff$(RESET)           포맷팅 차이점 표시"
	@echo -e "  • $(CYAN)format-file$(RESET)           특정 파일 포맷팅"
	@echo ""
	@echo -e "$(GREEN)🔍 Linting & Analysis:$(RESET)"
	@echo -e "  • $(CYAN)lint-check$(RESET)            Run golangci-lint checks"
	@echo -e "  • $(CYAN)lint-fix$(RESET)              Auto-fix lint issues where possible"
	@echo -e "  • $(CYAN)lint-status$(RESET)           Comprehensive lint status report"
	@echo -e "  • $(CYAN)analyze$(RESET)               Code complexity and quality analysis"
	@echo ""
	@echo -e "$(GREEN)🛡️  Security Analysis:$(RESET)"
	@echo -e "  • $(CYAN)security$(RESET)              All security checks (deps + code)"
	@echo -e "  • $(CYAN)security-deps$(RESET)         Check dependencies for vulnerabilities"
	@echo -e "  • $(CYAN)security-code$(RESET)         Static security analysis with gosec"
	@echo ""
	@echo -e "$(GREEN)🔄 Quality Workflows:$(RESET)"
	@echo -e "  • $(CYAN)quality$(RESET)               Comprehensive quality pipeline"
	@echo -e "  • $(CYAN)quality-fix$(RESET)           Apply all automatic fixes"
	@echo -e "  • $(CYAN)lint-all$(RESET)              Complete linting workflow"

quality-help: quality-info ## alias for quality-info

# ==============================================================================
# Enhanced Help System
# ==============================================================================

lint-help: ## show comprehensive help for linting targets
	@echo -e "$(BLUE)Code Quality and Linting Commands:$(RESET)"
	@echo ""
	@echo -e "$(YELLOW)🎨 Formatting:$(RESET)"
	@echo -e "  $(CYAN)format$(RESET)                기본 포맷팅 (format-simplify 실행)"
	@echo -e "  $(CYAN)format-simplify$(RESET)       신속한 기본 포맷팅"
	@echo -e "  $(CYAN)format-strict$(RESET)         엄격한 포맷팅 (모든 도구 사용)"
	@echo -e "  $(CYAN)format-list$(RESET)           포맷팅 필요한 파일 목록"
	@echo -e "  $(CYAN)format-diff$(RESET)           포맷팅 차이점 표시"
	@echo -e "  $(CYAN)format-file$(RESET)           특정 파일 포맷팅 (FILE= 옵션 사용)"
	@echo -e "  $(CYAN)format-check$(RESET)          Check code formatting without fixing"
	@echo -e "  $(CYAN)format-imports$(RESET)        Organize imports only"
	@echo -e "  $(CYAN)fmt$(RESET)                   Alias for format-simplify (backward compatibility)"
	@echo -e "  $(CYAN)format-all$(RESET)            Alias for format-strict (backward compatibility)"
	@echo ""
	@echo -e "$(YELLOW)🔍 Linting:$(RESET)"
	@echo -e "  $(CYAN)lint$(RESET)                  Check lint issues without fixing"
	@echo -e "  $(CYAN)lint-fix$(RESET)              Run golangci-lint with auto-fix"
	@echo -e "  $(CYAN)lint-new$(RESET)              Run golangci-lint on new code only"
	@echo -e "  $(CYAN)lint-ci$(RESET)               Run golangci-lint for CI"
	@echo -e "  $(CYAN)lint-count$(RESET)            Count total lint issues"
	@echo -e "  $(CYAN)lint-summary$(RESET)          Show lint issues summary by linter"
	@echo -e "  $(CYAN)lint-stats$(RESET)            Show detailed lint statistics"
	@echo -e "  $(CYAN)lint-status$(RESET)           Comprehensive lint status report"
	@echo -e "  $(CYAN)lint-json$(RESET)             Export lint results to JSON"
	@echo ""
	@echo -e "$(YELLOW)🔒 Security Analysis:$(RESET)"
	@echo -e "  $(CYAN)security$(RESET)              Run all security checks"
	@echo -e "  $(CYAN)security-deps$(RESET)         Check dependencies for vulnerabilities"
	@echo -e "  $(CYAN)security-code$(RESET)         Run security code analysis with gosec"
	@echo -e "  $(CYAN)security-json$(RESET)         Security analysis with JSON output"
	@echo ""
	@echo -e "$(YELLOW)📊 Code Analysis:$(RESET)"
	@echo -e "  $(CYAN)analyze$(RESET)               Run comprehensive code analysis"
	@echo -e "  $(CYAN)analyze-complexity$(RESET)    Analyze code complexity"
	@echo -e "  $(CYAN)analyze-unused$(RESET)        Find unused code"
	@echo -e "  $(CYAN)analyze-dupl$(RESET)          Find duplicate code"
	@echo ""
	@echo -e "$(YELLOW)🔧 Mock Generation:$(RESET)"
	@echo -e "  $(CYAN)generate-mocks$(RESET)        Generate all mock files using gomock"
	@echo -e "  $(CYAN)clean-mocks$(RESET)           Remove all generated mock files"
	@echo -e "  $(CYAN)regenerate-mocks$(RESET)      Clean and regenerate all mocks"
	@echo ""
	@echo -e "$(YELLOW)🎣 Pre-commit Integration:$(RESET)"
	@echo -e "  $(CYAN)pre-commit-install$(RESET)    Install pre-commit hooks"
	@echo -e "  $(CYAN)pre-commit$(RESET)            Run pre-commit hooks"
	@echo -e "  $(CYAN)pre-push$(RESET)              Run pre-push hooks"
	@echo -e "  $(CYAN)pre-commit-update$(RESET)     Update pre-commit hooks"
	@echo -e "  $(CYAN)check-consistency$(RESET)     Verify lint configuration consistency"
	@echo ""
	@echo -e "$(YELLOW)🔄 Development Workflows:$(RESET)"
	@echo -e "  $(CYAN)dev$(RESET)                   Standard development workflow"
	@echo -e "  $(CYAN)dev-fast$(RESET)              Quick development cycle"
	@echo -e "  $(CYAN)verify$(RESET)                Complete verification before PR"
	@echo -e "  $(CYAN)ci-local$(RESET)              Run full CI pipeline locally"
	@echo -e "  $(CYAN)pr-check$(RESET)              Pre-PR submission check"
	@echo -e "  $(CYAN)quality$(RESET)               Run comprehensive quality checks"
	@echo -e "  $(CYAN)quality-fix$(RESET)           Apply automatic quality fixes"
	@echo -e "  $(CYAN)lint-all$(RESET)              Run all linting steps"
	@echo ""
	@echo -e "$(YELLOW)📁 Configuration Files:$(RESET)"
	@echo "  .golangci.yml             golangci-lint configuration"
	@echo "  .pre-commit-config.yaml   Pre-commit hooks configuration"
	@echo "  .gosec.yaml              gosec security scanner configuration"
