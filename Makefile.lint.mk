# Makefile.lint.mk - Code Quality and Linting for Go Project
# Separated from main Makefile for better organization

# ==============================================================================
# Code Quality Configuration
# ==============================================================================

.PHONY: fmt lint format security security-json generate-mocks clean-mocks regenerate-mocks
.PHONY: pre-commit-install pre-commit pre-push lint-all check-consistency pre-commit-update
.PHONY: complexity ineffassign dupl vuln dev dev-fast verify ci-local pr-check

# ==============================================================================
# Code Formatting and Linting
# ==============================================================================

fmt: ## format go files
	gofumpt -w .
	gci write .

lint: ## lint go files
	golangci-lint run -c .golangci.yml --fix

format: ## format go files (alias for lint)
	golangci-lint run -c .golangci.yml --fix

# ==============================================================================
# Security Analysis
# ==============================================================================

security: ## run security analysis with gosec
	@echo "Running security analysis..."
	@command -v gosec >/dev/null 2>&1 || { echo "gosec not found. Installing..."; go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@gosec -config=.gosec.yaml ./...

security-json: ## run security analysis and output JSON report
	@echo "Running security analysis with JSON output..."
	@command -v gosec >/dev/null 2>&1 || { echo "gosec not found. Installing..."; go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@gosec -fmt=json -out=gosec-report.json -config=.gosec.yaml ./...

# ==============================================================================
# Mock Generation
# ==============================================================================

generate-mocks: ## generate all mock files using gomock
	@echo "Generating mocks..."
	@command -v mockgen >/dev/null 2>&1 || { echo "mockgen not found. Installing..."; go install go.uber.org/mock/mockgen@latest; }
	@mockgen -source=pkg/github/interfaces.go -destination=pkg/github/mocks/github_mocks.go -package=mocks
	@mockgen -source=internal/filesystem/interfaces.go -destination=internal/filesystem/mocks/filesystem_mocks.go -package=mocks
	@mockgen -source=internal/httpclient/interfaces.go -destination=internal/httpclient/mocks/httpclient_mocks.go -package=mocks
	@mockgen -source=internal/git/interfaces.go -destination=internal/git/mocks/git_mocks.go -package=mocks
	@echo "Mock generation complete!"

clean-mocks: ## remove all generated mock files
	@echo "Cleaning generated mocks..."
	@rm -f pkg/github/mocks/github_mocks.go
	@rm -f internal/filesystem/mocks/filesystem_mocks.go
	@rm -f internal/httpclient/mocks/httpclient_mocks.go
	@rm -f internal/git/mocks/git_mocks.go
	@echo "Mock cleanup complete!"

regenerate-mocks: clean-mocks generate-mocks ## clean and regenerate all mocks

# ==============================================================================
# Pre-commit Integration
# ==============================================================================

pre-commit-install: ## install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || { echo "pre-commit not found. Install with: pip install pre-commit"; exit 1; }
	@./scripts/setup-git-hooks.sh

pre-commit:	## run pre-commit hooks (format + light checks)
	pre-commit run --all-files

pre-push: ## run pre-push hooks (comprehensive checks)
	pre-commit run --all-files --hook-stage pre-push

lint-all: fmt lint pre-commit ## run all linting steps (format, lint, pre-commit)

check-consistency: ## verify lint configuration consistency
	@echo "Checking lint configuration consistency..."
	@echo "✓ Makefile uses: .golangci.yml"
	@grep -q "\.golangci\.yml" .pre-commit-config.yaml && echo "✓ Pre-commit uses: .golangci.yml" || echo "✗ Pre-commit config mismatch"
	@echo "✓ All configurations aligned"

pre-commit-update: ## update pre-commit hooks to latest versions
	pre-commit autoupdate

# ==============================================================================
# Code Analysis Tools
# ==============================================================================

complexity: ## analyze code complexity
	@echo "Analyzing code complexity..."
	@command -v gocyclo >/dev/null 2>&1 || { echo "gocyclo not found. Installing..."; go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; }
	@gocyclo -over 10 -avg .

ineffassign: ## detect ineffectual assignments
	@echo "Checking for ineffectual assignments..."
	@command -v ineffassign >/dev/null 2>&1 || { echo "ineffassign not found. Installing..."; go install github.com/gordonklaus/ineffassign@latest; }
	@ineffassign ./...

dupl: ## find duplicate code
	@echo "Checking for duplicate code..."
	@command -v dupl >/dev/null 2>&1 || { echo "dupl not found. Installing..."; go install github.com/mibk/dupl@latest; }
	@dupl -threshold 50 .

vuln: ## check for known vulnerabilities
	@echo "Checking for known vulnerabilities..."
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# ==============================================================================
# Development Workflow Targets
# ==============================================================================

dev: fmt lint test ## run standard development workflow (format, lint, test)

dev-fast: fmt test-unit ## quick development cycle (format and unit tests only)

verify: fmt lint test cover-report check-consistency ## complete verification before PR

ci-local: clean verify test-all ## run full CI pipeline locally

pr-check: fmt lint test cover-report check-consistency ## pre-PR submission check

# ==============================================================================
# Help
# ==============================================================================

lint-help: ## show help for linting targets
	@echo "Code Quality and Linting Commands:"
	@echo ""
	@echo "Formatting and Linting:"
	@echo "  make fmt             Format Go files with gofumpt and gci"
	@echo "  make lint            Run golangci-lint with auto-fix"
	@echo "  make format          Alias for lint"
	@echo "  make lint-all        Run all linting steps (format, lint, pre-commit)"
	@echo ""
	@echo "Security Analysis:"
	@echo "  make security        Run gosec security analysis"
	@echo "  make security-json   Run gosec with JSON output"
	@echo ""
	@echo "Mock Generation:"
	@echo "  make generate-mocks  Generate all mock files using gomock"
	@echo "  make clean-mocks     Remove all generated mock files"
	@echo "  make regenerate-mocks Clean and regenerate all mocks"
	@echo ""
	@echo "Pre-commit Integration:"
	@echo "  make pre-commit-install    Install pre-commit hooks"
	@echo "  make pre-commit           Run pre-commit hooks"
	@echo "  make pre-push             Run pre-push hooks"
	@echo "  make pre-commit-update    Update pre-commit hooks"
	@echo "  make check-consistency    Verify lint configuration consistency"
	@echo ""
	@echo "Code Analysis:"
	@echo "  make complexity      Analyze code complexity with gocyclo"
	@echo "  make ineffassign     Detect ineffectual assignments"
	@echo "  make dupl            Find duplicate code"
	@echo "  make vuln            Check for known vulnerabilities"
	@echo ""
	@echo "Development Workflows:"
	@echo "  make dev             Standard development workflow (format, lint, test)"
	@echo "  make dev-fast        Quick development cycle (format and unit tests only)"
	@echo "  make verify          Complete verification before PR"
	@echo "  make ci-local        Run full CI pipeline locally"
	@echo "  make pr-check        Pre-PR submission check"
	@echo ""
	@echo "Configuration Files:"
	@echo "  .golangci.yml        golangci-lint configuration"
	@echo "  .pre-commit-config.yaml  Pre-commit hooks configuration"
	@echo "  .gosec.yaml         gosec security scanner configuration"