# Makefile.deps.mk - Enhanced Dependency Management for gzh-manager-go
# Alternative to Dependabot for controlled local updates

# ==============================================================================
# Dependency Management Configuration
# ==============================================================================

.PHONY: deps-check deps-update deps-upgrade deps-update-go deps-update-actions deps-update-docker
.PHONY: deps-outdated deps-security deps-audit deps-report deps-clean deps-help
.PHONY: deps-update-minor deps-update-patch deps-update-major deps-interactive
.PHONY: deps-verify deps-why deps-weekly deps-monthly deps-tidy

# ==============================================================================
# Go Dependencies
# ==============================================================================

deps-check: ## check for outdated Go dependencies
	@echo -e "$(CYAN)Checking for outdated Go dependencies...$(RESET)"
	@go list -u -m all | grep '\[' || echo "$(GREEN)✅ All Go dependencies are up to date$(RESET)"

deps-outdated: ## detailed report of outdated dependencies
	@echo -e "$(CYAN)Generating detailed outdated dependencies report...$(RESET)"
	@echo -e "$(YELLOW)Go Modules:$(RESET)"
	@go list -u -m all | grep '\[' | while read line; do \
		echo "  $(RED)→$(RESET) $$line"; \
	done || echo "  $(GREEN)✅ All Go modules are up to date$(RESET)"
	@echo ""
	@echo -e "$(YELLOW)Direct Dependencies:$(RESET)"
	@go list -u -m all | grep '\[' | grep -v 'indirect' | while read line; do \
		echo "  $(RED)→$(RESET) $$line"; \
	done || echo "  $(GREEN)✅ All direct dependencies are up to date$(RESET)"

deps-tidy: ## run go mod tidy to clean up dependencies
	@echo -e "$(CYAN)Tidying Go modules...$(RESET)"
	@go mod tidy
	@echo -e "$(GREEN)✅ Go modules tidied$(RESET)"

deps-update: ## update all dependencies (safe: patch + minor only)
	@echo -e "$(CYAN)Updating dependencies safely (patch + minor versions)...$(RESET)"
	@echo -e "$(YELLOW)Before update:$(RESET)"
	@go mod tidy
	@cp go.mod go.mod.backup
	@cp go.sum go.sum.backup
	@echo -e "$(CYAN)Updating Go dependencies...$(RESET)"
	@go get -u=patch ./...
	@go mod tidy
	@echo -e "$(GREEN)✅ Dependencies updated safely$(RESET)"
	@echo -e "$(YELLOW)Changes:$(RESET)"
	@diff go.mod.backup go.mod || echo "  No changes in go.mod"
	@rm go.mod.backup go.sum.backup

deps-update-minor: ## update to latest minor versions (more aggressive)
	@echo -e "$(CYAN)Updating to latest minor versions...$(RESET)"
	@cp go.mod go.mod.backup
	@cp go.sum go.sum.backup
	@go get -u ./...
	@go mod tidy
	@echo -e "$(GREEN)✅ Dependencies updated to latest minor versions$(RESET)"
	@echo -e "$(YELLOW)Changes:$(RESET)"
	@diff go.mod.backup go.mod || echo "  No changes in go.mod"
	@rm go.mod.backup go.sum.backup

deps-update-patch: ## update to latest patch versions only (safest)
	@echo -e "$(CYAN)Updating to latest patch versions only...$(RESET)"
	@cp go.mod go.mod.backup
	@cp go.sum go.sum.backup
	@go get -u=patch ./...
	@go mod tidy
	@echo -e "$(GREEN)✅ Dependencies updated to latest patch versions$(RESET)"
	@echo -e "$(YELLOW)Changes:$(RESET)"
	@diff go.mod.backup go.mod || echo "  No changes in go.mod"
	@rm go.mod.backup go.sum.backup

deps-update-major: ## update to latest major versions (use with caution!)
	@echo -e "$(RED)⚠️  WARNING: This will update to latest major versions!$(RESET)"
	@echo -e "$(YELLOW)This may introduce breaking changes. Continue? [y/N]$(RESET)"
	@read -r confirm && [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ] || exit 1
	@cp go.mod go.mod.backup
	@cp go.sum go.sum.backup
	@go list -u -m all | grep '\[' | cut -d' ' -f1 | xargs -I {} go get {}@latest
	@go mod tidy
	@echo -e "$(GREEN)✅ Dependencies updated to latest major versions$(RESET)"
	@echo -e "$(YELLOW)Changes:$(RESET)"
	@diff go.mod.backup go.mod || echo "  No changes in go.mod"
	@rm go.mod.backup go.sum.backup

deps-interactive: ## interactive dependency update (choose which ones to update)
	@echo -e "$(CYAN)Interactive dependency update...$(RESET)"
	@outdated=$$(go list -u -m all | grep '\['); \
	if [ -z "$$outdated" ]; then \
		echo "$(GREEN)✅ All dependencies are up to date$(RESET)"; \
		exit 0; \
	fi; \
	echo "$$outdated" | while read line; do \
		pkg=$$(echo $$line | cut -d' ' -f1); \
		current=$$(echo $$line | cut -d' ' -f2); \
		latest=$$(echo $$line | sed 's/.*\[\(.*\)\].*/\1/'); \
		echo "$(YELLOW)Update $$pkg from $$current to $$latest? [y/N]$(RESET)"; \
		read -r confirm; \
		if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
			echo "$(CYAN)Updating $$pkg...$(RESET)"; \
			go get $$pkg@$$latest; \
		fi; \
	done; \
	go mod tidy

# ==============================================================================
# GitHub Actions Dependencies
# ==============================================================================

deps-update-actions: ## check and show GitHub Actions that need updates
	@echo -e "$(CYAN)Checking GitHub Actions dependencies...$(RESET)"
	@if [ -d ".github/workflows" ]; then \
		echo "$(YELLOW)GitHub Actions in use:$(RESET)"; \
		grep -r "uses:" .github/workflows/ | sed 's/.*uses: */  → /' | sort | uniq; \
		echo ""; \
		echo "$(YELLOW)To update GitHub Actions, manually edit .github/workflows/*.yml files$(RESET)"; \
		echo "$(YELLOW)Common updates:$(RESET)"; \
		echo "  → actions/checkout@v4"; \
		echo "  → actions/setup-go@v5"; \
		echo "  → actions/cache@v4"; \
		echo "  → codecov/codecov-action@v4"; \
	else \
		echo "$(GREEN)✅ No GitHub Actions found$(RESET)"; \
	fi

# ==============================================================================
# Docker Dependencies
# ==============================================================================

deps-update-docker: ## check and show Docker base images that need updates
	@echo -e "$(CYAN)Checking Docker dependencies...$(RESET)"
	@if [ -f "Dockerfile" ]; then \
		echo "$(YELLOW)Docker base images in use:$(RESET)"; \
		grep -E "^FROM" Dockerfile | sed 's/FROM */  → /'; \
		echo ""; \
		echo "$(YELLOW)To update Docker images, manually edit Dockerfile$(RESET)"; \
		echo "$(YELLOW)Consider using specific version tags instead of 'latest'$(RESET)"; \
	else \
		echo "$(GREEN)✅ No Dockerfile found$(RESET)"; \
	fi
	@if [ -f "docker-compose.yml" ]; then \
		echo ""; \
		echo "$(YELLOW)Docker Compose images in use:$(RESET)"; \
		grep -E "image:" docker-compose.yml | sed 's/.*image: */  → /' | sort | uniq; \
	fi

# ==============================================================================
# Security and Audit
# ==============================================================================

deps-security: ## run security audit on dependencies
	@echo -e "$(CYAN)Running security audit...$(RESET)"
	@echo -e "$(YELLOW)Checking for known vulnerabilities...$(RESET)"
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./... || echo "$(RED)❌ Vulnerabilities found$(RESET)"

deps-audit: ## comprehensive dependency audit and report
	@echo -e "$(CYAN)Comprehensive dependency audit...$(RESET)"
	@echo -e "$(BLUE)=== Go Module Information ===$(RESET)"
	@go version
	@echo "Module: $$(go list -m)"
	@echo "Go version: $$(go list -m -f '{{.GoVersion}}')"
	@echo ""
	@echo -e "$(BLUE)=== Direct Dependencies ===$(RESET)"
	@go list -m -f '{{if not .Indirect}}{{.Path}} {{.Version}}{{end}}' all | grep -v "^$$" | head -20
	@echo ""
	@echo -e "$(BLUE)=== Outdated Dependencies ===$(RESET)"
	@make --no-print-directory deps-outdated
	@echo ""
	@echo -e "$(BLUE)=== Security Check ===$(RESET)"
	@make --no-print-directory deps-security

deps-verify: ## verify dependency checksums
	@echo -e "$(CYAN)Verifying dependency checksums...$(RESET)"
	@go mod verify
	@echo -e "$(GREEN)✅ All dependency checksums verified$(RESET)"

deps-why: ## show why a specific module is needed (usage: make deps-why MOD=github.com/spf13/cobra)
	@if [ -z "$(MOD)" ]; then \
		echo "$(RED)Usage: make deps-why MOD=github.com/spf13/cobra$(RESET)"; \
		exit 1; \
	fi
	@echo -e "$(CYAN)Showing why $(MOD) is needed...$(RESET)"
	@go mod why -m $(MOD)

# ==============================================================================
# Dependency Reports
# ==============================================================================

deps-report: ## generate comprehensive dependency report
	@echo -e "$(CYAN)Generating dependency report...$(RESET)"
	@report_file="dependency-report-$$(date +%Y%m%d-%H%M%S).md"; \
	echo "# Dependency Report - gzh-manager-go" > $$report_file; \
	echo "Generated: $$(date)" >> $$report_file; \
	echo "" >> $$report_file; \
	echo "## Go Module Information" >> $$report_file; \
	echo "\`\`\`" >> $$report_file; \
	go version >> $$report_file; \
	echo "Module: $$(go list -m)" >> $$report_file; \
	echo "Go version: $$(go list -m -f '{{.GoVersion}}')" >> $$report_file; \
	echo "\`\`\`" >> $$report_file; \
	echo "" >> $$report_file; \
	echo "## Direct Dependencies" >> $$report_file; \
	echo "\`\`\`" >> $$report_file; \
	go list -m -f '{{if not .Indirect}}{{.Path}} {{.Version}}{{end}}' all | grep -v "^$$" >> $$report_file; \
	echo "\`\`\`" >> $$report_file; \
	echo "" >> $$report_file; \
	echo "## Outdated Dependencies" >> $$report_file; \
	echo "\`\`\`" >> $$report_file; \
	go list -u -m all | grep '\[' >> $$report_file || echo "All dependencies are up to date" >> $$report_file; \
	echo "\`\`\`" >> $$report_file; \
	echo "$(GREEN)✅ Report generated: $$report_file$(RESET)"

# ==============================================================================
# Cleanup and Maintenance
# ==============================================================================

deps-clean: ## clean up dependency cache and temporary files
	@echo -e "$(CYAN)Cleaning dependency cache...$(RESET)"
	@go clean -modcache
	@go clean -cache
	@rm -f go.mod.backup go.sum.backup
	@rm -f go.mod.monthly-backup go.sum.monthly-backup
	@echo -e "$(GREEN)✅ Dependency cache cleaned$(RESET)"

# ==============================================================================
# Dependabot Alternative Workflow
# ==============================================================================

deps-weekly: ## run weekly dependency maintenance (safe updates)
	@echo -e "$(BLUE)🗓️  Running weekly dependency maintenance...$(RESET)"
	@echo -e "$(YELLOW)1. Checking current status...$(RESET)"
	@make --no-print-directory deps-check
	@echo ""
	@echo -e "$(YELLOW)2. Running security audit...$(RESET)"
	@make --no-print-directory deps-security
	@echo ""
	@echo -e "$(YELLOW)3. Updating patch versions (safest)...$(RESET)"
	@make --no-print-directory deps-update-patch
	@echo ""
	@echo -e "$(YELLOW)4. Running tests after update...$(RESET)"
	@go test ./... -short
	@echo ""
	@echo -e "$(GREEN)✅ Weekly maintenance completed$(RESET)"

deps-monthly: ## run monthly dependency maintenance (minor updates)
	@echo -e "$(BLUE)📅 Running monthly dependency maintenance...$(RESET)"
	@echo -e "$(YELLOW)1. Creating backup...$(RESET)"
	@cp go.mod go.mod.monthly-backup
	@cp go.sum go.sum.monthly-backup
	@echo ""
	@echo -e "$(YELLOW)2. Updating minor versions...$(RESET)"
	@make --no-print-directory deps-update-minor
	@echo ""
	@echo -e "$(YELLOW)3. Running full test suite...$(RESET)"
	@go test ./...
	@echo ""
	@echo -e "$(YELLOW)4. Running security audit...$(RESET)"
	@make --no-print-directory deps-security
	@echo ""
	@if [ -f "go.mod.monthly-backup" ]; then \
		echo "$(YELLOW)Backup files created:$(RESET)"; \
		echo "  → go.mod.monthly-backup"; \
		echo "  → go.sum.monthly-backup"; \
	fi
	@echo -e "$(GREEN)✅ Monthly maintenance completed$(RESET)"

# ==============================================================================
# Help System
# ==============================================================================

deps-help: ## show comprehensive help for dependency management commands
	@echo -e "$(BLUE)📦 Dependency Management Commands:$(RESET)"
	@echo ""
	@echo -e "$(YELLOW)📋 Daily Operations:$(RESET)"
	@echo -e "  $(CYAN)deps-check$(RESET)            Check for outdated dependencies"
	@echo -e "  $(CYAN)deps-outdated$(RESET)         Detailed outdated dependencies report"
	@echo -e "  $(CYAN)deps-tidy$(RESET)             Run go mod tidy to clean up dependencies"
	@echo -e "  $(CYAN)deps-update$(RESET)           Safe update (patch + minor only)"
	@echo -e "  $(CYAN)deps-interactive$(RESET)      Interactive dependency updates"
	@echo ""
	@echo -e "$(YELLOW)🔄 Update Levels:$(RESET)"
	@echo -e "  $(CYAN)deps-update-patch$(RESET)     Update patch versions only (safest)"
	@echo -e "  $(CYAN)deps-update-minor$(RESET)     Update minor versions (moderate risk)"
	@echo -e "  $(CYAN)deps-update-major$(RESET)     Update major versions (⚠️  breaking changes!)"
	@echo ""
	@echo -e "$(YELLOW)🔒 Security & Audit:$(RESET)"
	@echo -e "  $(CYAN)deps-security$(RESET)         Run security vulnerability scan"
	@echo -e "  $(CYAN)deps-audit$(RESET)            Comprehensive dependency audit"
	@echo -e "  $(CYAN)deps-verify$(RESET)           Verify dependency checksums"
	@echo ""
	@echo -e "$(YELLOW)📊 Analysis & Reporting:$(RESET)"
	@echo -e "  $(CYAN)deps-report$(RESET)           Generate comprehensive dependency report"
	@echo -e "  $(CYAN)deps-why MOD=...$(RESET)      Show why a module is needed"
	@echo ""
	@echo -e "$(YELLOW)🔄 Other Dependencies:$(RESET)"
	@echo -e "  $(CYAN)deps-update-actions$(RESET)   Check GitHub Actions updates"
	@echo -e "  $(CYAN)deps-update-docker$(RESET)    Check Docker base image updates"
	@echo ""
	@echo -e "$(YELLOW)📅 Maintenance Workflows:$(RESET)"
	@echo -e "  $(CYAN)deps-weekly$(RESET)           Weekly maintenance (patch updates + security)"
	@echo -e "  $(CYAN)deps-monthly$(RESET)          Monthly maintenance (minor updates + full tests)"
	@echo ""
	@echo -e "$(YELLOW)🧹 Cleanup:$(RESET)"
	@echo -e "  $(CYAN)deps-clean$(RESET)            Clean dependency cache and temporary files"
	@echo ""
	@echo -e "$(YELLOW)💡 Usage Examples:$(RESET)"
	@echo -e "  $(GREEN)make deps-check$(RESET)                    # Check what's outdated"
	@echo -e "  $(GREEN)make deps-weekly$(RESET)                   # Safe weekly maintenance"
	@echo -e "  $(GREEN)make deps-interactive$(RESET)              # Choose what to update"
	@echo -e "  $(GREEN)make deps-why MOD=github.com/spf13/cobra$(RESET)  # Why is cobra needed?"
	@echo ""
	@echo -e "$(BLUE)📝 Configuration:$(RESET)"
	@echo "  This replaces Dependabot for more controlled dependency management"
	@echo -e "  Recommended: Run $(YELLOW)deps-weekly$(RESET) every week for maintenance"
