# Makefile.quality - Code Quality and Analysis for gzh-cli
# Formatting, linting, security analysis, and code quality checks

# ==============================================================================
# Quality Configuration
# ==============================================================================

.PHONY: fmt fmt-check format format-all format-check format-diff format-imports format-simplify format-ci format-strict format-list format-file format-install-tools format-md format-md-check format-md-diff
.PHONY: pre-commit-install dev dev-fast verify ci-local pr-check lint-help fmt-diff lint-diff quality-fast quality-push

# ==============================================================================
# Code Formatting Targets
# ==============================================================================

format: format-simplify ## quick and simple formatting (default)
fmt: format-simplify

format-simplify: format-install-tools ## quick basic formatting with gofumpt, goimports, and mdformat
	@echo -e "$(CYAN)🚀 Quick formatting...$(RESET)"
	@echo "1. Running gofumpt (includes go fmt + simplification)..."
	@"$(GOFUMPT)" -w .
	@echo "2. Organizing imports..."
	@"$(GOIMPORTS)" -w -local github.com/gizzahub/gzh-cli .
	@echo "3. Formatting markdown files..."
	@find . -name "*.md" -type f -not -path "./vendor/*" -not -path "./.git/*" | xargs -r mdformat || true
	@echo -e "$(GREEN)✅ Quick formatting complete!$(RESET)"

format-md: install-mdformat ## format all markdown files with mdformat
	@echo -e "$(CYAN)📝 Formatting markdown files...$(RESET)"
	@find . -name "*.md" -type f -not -path "./vendor/*" -not -path "./.git/*" | xargs -r mdformat
	@echo -e "$(GREEN)✅ Markdown formatting complete!$(RESET)"

# The `|| echo` this replaced turned a check into a report: mdformat --check named
# the unformatted file, the recipe printed a warning, and make still exited 0, so
# nothing downstream could act on it -- "checked and clean" and "checked and dirty"
# were the same exit code. The pipeline's status is xargs's, and xargs reports a
# non-zero status when mdformat does (measured here: rc=1 on an unformatted file,
# rc=0 once formatted), so dropping the `|| echo` is the entire fix.
# pipefail is deliberately not added: it is not portable across the shells make may
# pick, and it would only cover `find` itself failing, not this target's job.
format-md-check: install-mdformat ## check markdown files that need formatting
	@echo -e "$(CYAN)📋 Checking markdown formatting...$(RESET)"
	@find . -name "*.md" -type f -not -path "./vendor/*" -not -path "./.git/*" | xargs -r mdformat --check

# `.PHONY` above has listed format-check since this file was written and no rule
# ever stood behind it, and scripts/pre-commit-lint.sh has called `make fmt-check`
# just as long. Both were answered by the match-anything rule in .make/build.mk,
# which returned success for any undefined target -- so step 1 of that hook, the
# format check, has never checked anything. This is the target both of them meant.
#
# Non-mutating on purpose. `fmt` and `format-strict` rewrite the tree, which is
# the wrong thing for a gate to do to a working copy it is only supposed to
# judge. It checks with the same two pinned tools format-strict applies, in the
# same modes, so what it verifies is exactly what `make format-strict` produces.
format-check: format-install-tools ## fail if Go files are not formatted (non-mutating)
	@echo -e "$(CYAN)📋 Checking Go formatting...$(RESET)"
	@unformatted="$$("$(GOFUMPT)" -l -extra .)"; \
	if [ -n "$$unformatted" ]; then \
		echo -e "$(RED)❌ gofumpt: not formatted:$(RESET)"; \
		printf '%s\n' "$$unformatted"; \
		echo "   Run: make format-strict"; \
		exit 1; \
	fi
	@imports="$$("$(GCI)" diff --skip-generated -s standard -s default -s "prefix(github.com/gizzahub/gzh-cli)" .)"; \
	if [ -n "$$imports" ]; then \
		echo -e "$(RED)❌ gci: import grouping not normalized:$(RESET)"; \
		printf '%s\n' "$$imports"; \
		echo "   Run: make format-strict"; \
		exit 1; \
	fi
	@echo -e "$(GREEN)✅ Go files are properly formatted$(RESET)"

fmt-check: format-check ## alias for format-check (the name pre-commit-lint.sh uses)

format-md-diff: install-mdformat ## format only changed markdown files
	@echo -e "$(CYAN)🚀 Formatting changed markdown files...$(RESET)"
	@CHANGED_FILES=$$(git diff --name-only --diff-filter=d HEAD | grep '\.md$$' || true); \
	if [ -n "$$CHANGED_FILES" ]; then \
		echo "$$CHANGED_FILES" | xargs -r mdformat; \
		echo -e "$(GREEN)✅ Changed markdown files formatted!$(RESET)"; \
	else \
		echo -e "$(YELLOW)No markdown files changed$(RESET)"; \
	fi

format-strict: format-install-tools ## comprehensive formatting with all tools
	@echo -e "$(CYAN)🔧 Strict formatting (all tools)...$(RESET)"
	@echo "1. Running gofumpt (strict formatting + simplification)..."
	@"$(GOFUMPT)" -w -extra .
	@echo "2. Running gci (import organization)..."
	@"$(GCI)" write --skip-generated .
	@echo "3. Organizing imports with goimports..."
	@"$(GOIMPORTS)" -w -local github.com/gizzahub/gzh-cli .
	@echo "4. Final gci (import grouping)..."
	@"$(GCI)" write --skip-generated -s standard -s default -s "prefix(github.com/gizzahub/gzh-cli)" .
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

# Kept as the name every formatting target already depends on; the pins and the
# version checks live in .make/tools.mk next to the other pinned toolchains.
format-install-tools: install-format-tools ## install the exact pinned formatting toolchain

format-file: format-install-tools ## format specific files with gofumpt and goimports (usage: make format-file file1.go file2.go ...)
	@if [ -z "$(MAKECMDGOALS)" ] || [ "$(words $(MAKECMDGOALS))" -eq 1 ]; then \
		echo -e "$(RED)❌ Error: At least one file must be specified$(RESET)"; \
		echo -e "$(YELLOW)Usage: make format-file file1.go file2.go ...$(RESET)"; \
		exit 1; \
	fi
	@echo -e "$(CYAN)🔄 Processing files...$(RESET)"
	@for file in $(filter-out format-file,$(MAKECMDGOALS)); do \
		if [ -n "$$file" ]; then \
			if [ ! -f "$$file" ]; then \
				echo -e "$(RED)❌ Error: File '$$file' does not exist$(RESET)"; \
				continue; \
			fi; \
			if ! echo "$$file" | grep -q "\.go$$"; then \
				echo -e "$(YELLOW)⚠️  Warning: File '$$file' is not a Go file (.go extension), skipping$(RESET)"; \
				continue; \
			fi; \
			echo -e "$(CYAN)📝 Formatting file: $$file$(RESET)"; \
			echo "  1. Running gofumpt..."; \
			"$(GOFUMPT)" -w "$$file" || echo -e "$(RED)❌ gofumpt failed for $$file$(RESET)"; \
			echo "  2. Running goimports..."; \
			"$(GOIMPORTS)" -w -local github.com/gizzahub/gzh-cli "$$file" || echo -e "$(RED)❌ goimports failed for $$file$(RESET)"; \
			echo -e "$(GREEN)✅ File '$$file' formatted successfully!$(RESET)"; \
		fi; \
	done
	@echo -e "$(GREEN)🎉 All files processed!$(RESET)"

fmt-diff: format-install-tools ## format only changed files (fast, for pre-commit)
	@echo -e "$(CYAN)🚀 Formatting changed files only...$(RESET)"
	@CHANGED_FILES=$$(git diff --name-only --diff-filter=d HEAD | grep '\.go$$' || true); \
	if [ -n "$$CHANGED_FILES" ]; then \
		echo "$$CHANGED_FILES" | while read file; do \
			if [ -f "$$file" ]; then \
				echo -e "$(CYAN)📝 Formatting: $$file$(RESET)"; \
				"$(GOFUMPT)" -w "$$file" || echo -e "$(RED)❌ gofumpt failed for $$file$(RESET)"; \
				"$(GOIMPORTS)" -w -local github.com/gizzahub/gzh-cli "$$file" || echo -e "$(RED)❌ goimports failed for $$file$(RESET)"; \
			fi; \
		done; \
		echo -e "$(GREEN)✅ Changed files formatted!$(RESET)"; \
	else \
		echo -e "$(YELLOW)No Go files changed$(RESET)"; \
	fi

# ==============================================================================
# Linting and Static Analysis
# ==============================================================================

.PHONY: lint format lint-check lint-fix lint-new lint-ci lint-count lint-summary lint-stats lint-status lint-json

lint-check: install-golangci-lint ## check lint issues without fixing (exit code reflects status)
	@echo -e "$(CYAN)Running golangci-lint...$(RESET)"
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml

lint: lint-check ## alias for lint-check

lint-fix: install-golangci-lint ## run golangci-lint with auto-fix
	@echo -e "$(CYAN)Running golangci-lint with auto-fix...$(RESET)"
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --fix

lint-new: install-golangci-lint ## run golangci-lint on new code only
	@echo -e "$(CYAN)Running golangci-lint on new code only...$(RESET)"
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --new-from-rev=HEAD~

lint-ci: install-golangci-lint ## run golangci-lint for CI
	@echo -e "$(CYAN)Running golangci-lint for CI...$(RESET)"
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --output.text.path=stdout --output.text.colors=false

lint-count: install-golangci-lint ## count total lint issues without fixing
	@echo -e "$(CYAN)Counting lint issues...$(RESET)"
	@set -e; \
	OUTPUT=$$(mktemp); \
	trap 'rm -f "$$OUTPUT"' EXIT; \
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --issues-exit-code=0 --max-issues-per-linter=0 --max-same-issues=0 --output.text.path=stdout --output.text.colors=false >"$$OUTPUT"; \
	ISSUES=$$(grep -E "^[^[:space:]].*\\([^)]+\\)$$" "$$OUTPUT" | wc -l); \
	echo -e "$(YELLOW)Total lint issues: $$ISSUES$(RESET)"

lint-summary: install-golangci-lint ## show lint issues summary by linter
	@echo -e "$(CYAN)Lint issues summary:$(RESET)"
	@set -e; \
	OUTPUT=$$(mktemp); \
	trap 'rm -f "$$OUTPUT"' EXIT; \
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --issues-exit-code=0 --max-issues-per-linter=0 --max-same-issues=0 --output.text.path=stdout --output.text.colors=false >"$$OUTPUT"; \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" "$$OUTPUT" | sed 's/.*(\\([^)]*\\))$$/\\1/' | sort | uniq -c | sort -nr | \
	awk '{printf "  $(YELLOW)%-15s$(RESET) %d issues\\n", $$2, $$1}'

lint-stats: install-golangci-lint ## show detailed lint statistics with golangci-lint built-in stats
	@echo -e "$(CYAN)=== Lint Statistics ===$(RESET)"
	@"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --show-stats --max-issues-per-linter=0 --max-same-issues=0

lint-status: install-golangci-lint ## comprehensive lint status report
	@echo -e "$(BLUE)🔍 Comprehensive Lint Status Report$(RESET)"
	@echo -e "$(BLUE)==================================$(RESET)"
	@echo ""
	@echo -e "$(GREEN)📊 Quick Stats:$(RESET)"
	@set -e; \
	TEXT_OUTPUT=$$(mktemp); JSON_OUTPUT=$$(mktemp); \
	trap 'rm -f "$$TEXT_OUTPUT" "$$JSON_OUTPUT"' EXIT; \
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --issues-exit-code=0 --max-issues-per-linter=0 --max-same-issues=0 --output.text.path=stdout --output.text.colors=false >"$$TEXT_OUTPUT"; \
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --issues-exit-code=0 --max-issues-per-linter=0 --max-same-issues=0 --show-stats=false --output.text.path=stderr --output.json.path="$$JSON_OUTPUT" >/dev/null; \
	jq -e '.Issues | type == "array"' "$$JSON_OUTPUT" >/dev/null; \
	TOTAL=$$(grep -E "^[^[:space:]].*\\([^)]+\\)$$" "$$TEXT_OUTPUT" | wc -l); \
	ERRORS=$$(jq '[.Issues[]? | select(.Severity=="error")] | length' "$$JSON_OUTPUT"); \
	WARNINGS=$$(jq '[.Issues[]? | select(.Severity=="warning")] | length' "$$JSON_OUTPUT"); \
	echo "  $(YELLOW)Total Issues: $$TOTAL$(RESET)"; \
	echo "  $(RED)Errors: $$ERRORS$(RESET)"; \
	echo "  $(YELLOW)Warnings: $$WARNINGS$(RESET)"; \
	echo ""; \
	echo -e "$(GREEN)🏷️  Top 10 Linters:$(RESET)"; \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" "$$TEXT_OUTPUT" | sed 's/.*(\\([^)]*\\))$$/\\1/' | sort | uniq -c | sort -nr | head -10 | \
	awk '{printf "  $(CYAN)%-15s$(RESET) %d issues\\n", $$2, $$1}'; \
	echo ""; \
	echo -e "$(GREEN)📁 Most Problematic Files:$(RESET)"; \
	grep -E "^[^[:space:]].*\\([^)]+\\)$$" "$$TEXT_OUTPUT" | sed 's/^\\([^:]*\\):.*/\\1/' | sort | uniq -c | sort -nr | head -5 | \
	awk '{printf "  $(MAGENTA)%-40s$(RESET) %d issues\\n", $$2, $$1}'

lint-diff: install-golangci-lint ## lint only changed files (fast, for pre-commit)
	@echo -e "$(CYAN)🔍 Linting changed files only...$(RESET)"
	@set -eu; \
	CHANGED_FILES=$$(mktemp "$${TMPDIR:-/tmp}/gzh-cli-lint-diff.XXXXXX"); \
	trap 'rm -f "$$CHANGED_FILES"' EXIT HUP INT TERM; \
	git diff --name-only -z --diff-filter=d HEAD -- '*.go' >"$$CHANGED_FILES"; \
	if [ -s "$$CHANGED_FILES" ]; then \
		xargs -0 "$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --new-from-rev=HEAD~1 -- <"$$CHANGED_FILES"; \
	else \
		echo -e "$(YELLOW)No Go files changed$(RESET)"; \
	fi

lint-json: install-golangci-lint ## export lint results to JSON for further analysis
	@echo -e "$(CYAN)Exporting lint results to lint-report.json...$(RESET)"
	@set -e; \
	REPORT=$$(mktemp .lint-report.json.XXXXXX); \
	trap 'rm -f "$$REPORT"' EXIT; \
	"$(GOLANGCI_LINT)" run $(GOLANGCI_LINT_RUN_FLAGS) -c .golangci.yml --issues-exit-code=0 --max-issues-per-linter=0 --max-same-issues=0 --output.text.path=stderr --output.json.path="$$REPORT" >/dev/null; \
	test -s "$$REPORT"; \
	if command -v jq >/dev/null 2>&1; then jq -e '.Issues | type == "array"' "$$REPORT" >/dev/null; fi; \
	mv "$$REPORT" lint-report.json; \
	trap - EXIT
	@echo -e "$(GREEN)✅ Report saved to lint-report.json$(RESET)"
	@if command -v jq >/dev/null 2>&1; then \
		echo ""; \
		echo -e "$(YELLOW)📈 JSON Report Summary:$(RESET)"; \
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
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./... || echo -e "$(RED)❌ Vulnerabilities found$(RESET)"

security-code: install-gosec ## run security code analysis
	@echo -e "$(CYAN)Running security code analysis with gosec...$(RESET)"
	@"$(GOSEC)" $(GOSEC_SCAN_FLAGS) -no-fail ./...

security-json: install-gosec ## run security analysis and output JSON/SARIF report
	@echo -e "$(CYAN)Running security analysis with JSON/SARIF output...$(RESET)"
	@set -e; \
	REPORT=$$(mktemp .gosec-report.sarif.XXXXXX); \
	DETAILS=$$(mktemp .gosec-report.json.XXXXXX); \
	trap 'rm -f "$$REPORT" "$$DETAILS"' EXIT; \
	command -v jq >/dev/null 2>&1 || { echo "jq is required to validate gosec reports" >&2; exit 1; }; \
	"$(GOSEC)" $(GOSEC_SCAN_FLAGS) -no-fail -fmt=sarif -out="$$REPORT" -stdout -verbose=json ./... >"$$DETAILS"; \
	test -s "$$REPORT"; test -s "$$DETAILS"; \
	jq -e '.version == "2.1.0" and (.runs | type == "array" and length > 0) and all(.runs[]; (.tool.driver.name | type == "string" and length > 0) and (.results | type == "array"))' "$$REPORT" >/dev/null; \
	jq -e '(.Stats | type == "object") and (.Stats.files | type == "number" and . > 0) and (.["Golang errors"] | type == "object" and length == 0) and (.Issues | type == "array")' "$$DETAILS" >/dev/null; \
	mv "$$REPORT" gosec-report.json; \
	rm -f "$$DETAILS"; \
	trap - EXIT
	@echo -e "$(GREEN)✅ Security report generated: gosec-report.json$(RESET)"

# ==============================================================================
# Code Analysis
# ==============================================================================

.PHONY: analyze analyze-complexity analyze-unused analyze-dupl complexity ineffassign dupl

analyze: analyze-complexity analyze-unused analyze-dupl ## run comprehensive code analysis
	@echo -e "$(GREEN)✅ Code analysis complete!$(RESET)"

analyze-complexity: install-gocyclo ## analyze code complexity
	@echo -e "$(CYAN)Analyzing code complexity...$(RESET)"
	@"$(GOCYCLO)" -over 10 -avg .

analyze-unused: install-staticcheck ## find unused code
	@echo -e "$(CYAN)Finding unused code...$(RESET)"
	@"$(STATICCHECK)" -checks U1000 ./...

analyze-dupl: install-dupl ## find duplicate code
	@echo -e "$(CYAN)Checking for duplicate code...$(RESET)"
	@"$(DUPL)" -threshold 50 .

# ==============================================================================
# Pre-commit Integration
# ==============================================================================

.PHONY: pre-commit-install pre-commit pre-push check-consistency pre-commit-update

pre-commit-install: ## install pre-commit hooks
	@echo -e "$(CYAN)Installing pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo -e "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	@if [ -f "./scripts/setup-git-hooks.sh" ]; then \
		./scripts/setup-git-hooks.sh; \
	else \
		pre-commit install --hook-type pre-commit --hook-type commit-msg --hook-type pre-push; \
	fi
	@echo -e "$(GREEN)✅ Pre-commit hooks installed!$(RESET)"

pre-commit: ## run pre-commit hooks (format + light checks)
	@echo -e "$(CYAN)Running pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo -e "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit run --all-files

pre-push: ## run pre-push hooks (comprehensive checks)
	@echo -e "$(CYAN)Running pre-push hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo -e "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit run --all-files --hook-stage pre-push

check-consistency: ## verify lint configuration consistency
	@echo -e "$(CYAN)Checking lint configuration consistency...$(RESET)"
	@echo -e "$(GREEN)✓$(RESET) Makefile uses: .golangci.yml"
	@if [ -f ".pre-commit-config.yaml" ]; then \
		grep -q "\\.golangci\\.yml" .pre-commit-config.yaml && echo -e "$(GREEN)✓$(RESET) Pre-commit uses: .golangci.yml" || echo -e "$(RED)✗$(RESET) Pre-commit config mismatch"; \
	else \
		echo -e "$(YELLOW)⚠$(RESET) No pre-commit config found"; \
	fi
	@echo -e "$(GREEN)✅ Configuration consistency checked$(RESET)"

pre-commit-update: ## update pre-commit hooks to latest versions
	@echo -e "$(CYAN)Updating pre-commit hooks...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo -e "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; exit 1; }
	pre-commit autoupdate
	@echo -e "$(GREEN)✅ Pre-commit hooks updated!$(RESET)"

# ==============================================================================
# Quality Assurance Workflows
# ==============================================================================

.PHONY: quality quality-fix lint-all

quality: fmt security ## run comprehensive quality checks (without lint-check for now)
	@echo -e "$(GREEN)✅ All quality checks passed!$(RESET)"

quality-strict: fmt lint-check security ## run strict quality checks with linting
	@echo -e "$(GREEN)✅ All strict quality checks passed!$(RESET)"

quality-fix: fmt lint-fix ## apply automatic quality fixes
	@echo -e "$(GREEN)✅ Code quality fixes applied!$(RESET)"

lint-all: fmt lint-check pre-commit ## run all linting steps (format, lint, pre-commit)
	@echo -e "$(GREEN)✅ All linting steps completed!$(RESET)"

quality-fast: fmt-diff lint-diff format-md-diff ## fast quality check for pre-commit (changed files only, <3s)
	@echo -e "$(GREEN)⚡ Fast quality check completed!$(RESET)"

quality-push: format-strict lint-fix ## comprehensive quality check for pre-push
	@echo -e "$(GREEN)✅ Pre-push quality check completed!$(RESET)"

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
	@echo -e "  • $(CYAN)format$(RESET)                기본 포맷팅 (Go + Markdown)"
	@echo -e "  • $(CYAN)format-simplify$(RESET)       신속한 기본 포맷팅 (Go + Markdown)"
	@echo -e "  • $(CYAN)format-strict$(RESET)         엄격한 포맷팅 (모든 Go 도구 사용)"
	@echo -e "  • $(CYAN)format-md$(RESET)             Markdown 파일 포맷팅"
	@echo -e "  • $(CYAN)format-md-check$(RESET)       Markdown 포맷팅 필요 파일 확인"
	@echo -e "  • $(CYAN)format-md-diff$(RESET)        변경된 Markdown 파일만 포맷팅"
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
	@echo -e "  $(CYAN)format$(RESET)                기본 포맷팅 (Go + Markdown)"
	@echo -e "  $(CYAN)format-simplify$(RESET)       신속한 기본 포맷팅 (Go + Markdown)"
	@echo -e "  $(CYAN)format-strict$(RESET)         엄격한 포맷팅 (모든 Go 도구 사용)"
	@echo -e "  $(CYAN)format-md$(RESET)             Markdown 파일 포맷팅"
	@echo -e "  $(CYAN)format-md-check$(RESET)       Markdown 포맷팅 필요 파일 확인"
	@echo -e "  $(CYAN)format-md-diff$(RESET)        변경된 Markdown 파일만 포맷팅"
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
	@echo "  .gosec.json              gosec security scanner configuration"
