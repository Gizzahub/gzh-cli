# Makefile.build - Build and Installation targets for gzh-manager-go
# Build, compilation, and installation management

# ==============================================================================
# Build Configuration
# ==============================================================================

# Colors for output
CYAN := \\033[36m
GREEN := \\033[32m
YELLOW := \\033[33m
RED := \\033[31m
BLUE := \\033[34m
RESET := \\033[0m

# ==============================================================================
# Build Targets
# ==============================================================================

.PHONY: build install run bootstrap clean release-dry-run release-snapshot release-check deploy

build: ## build golang binary
	@echo "$(CYAN)Building $(executablename)...$(RESET)"
	@go build -ldflags "-X main.version=$(VERSION)" -o $(executablename)
	@echo "$(GREEN)✅ Built $(executablename) successfully$(RESET)"

install: build ## install golang binary
	@echo "$(CYAN)Installing $(executablename)...$(RESET)"
	@mv $(executablename) $(shell go env GOPATH)/bin/
	@echo "$(GREEN)✅ Installed $(executablename) to $(shell go env GOPATH)/bin/$(RESET)"

run: ## run the application
	@echo "$(CYAN)Running application with version $(VERSION)...$(RESET)"
	@go run -ldflags "-X main.version=$(VERSION)" main.go

bootstrap: ## install build dependencies
	@echo "$(CYAN)Installing build dependencies...$(RESET)"
	go generate -tags tools tools/tools.go
	@echo "$(GREEN)✅ Build dependencies installed$(RESET)"

clean: ## clean up environment
	@echo "$(CYAN)Cleaning up build artifacts...$(RESET)"
	@rm -rf coverage.out coverage.html dist/ $(executablename)
	@rm -f $(shell go env GOPATH)/bin/$(executablename)
	@rm -f $(shell go env GOPATH)/bin/$(projectname)
	@rm -f lint-report.json gosec-report.json
	@echo "$(GREEN)✅ Cleanup completed$(RESET)"

# ==============================================================================
# Release Targets
# ==============================================================================

release-dry-run: ## run goreleaser in dry-run mode
	@echo "$(CYAN)Running goreleaser in dry-run mode...$(RESET)"
	@command -v goreleaser >/dev/null 2>&1 || { echo "$(RED)goreleaser not found. Install with: make install-goreleaser$(RESET)"; exit 1; }
	@goreleaser release --snapshot --clean --skip=publish

release-snapshot: ## create a snapshot release
	@echo "$(CYAN)Creating snapshot release...$(RESET)"
	@command -v goreleaser >/dev/null 2>&1 || { echo "$(RED)goreleaser not found. Install with: make install-goreleaser$(RESET)"; exit 1; }
	@goreleaser release --snapshot --clean

release-check: ## check goreleaser configuration
	@echo "$(CYAN)Checking goreleaser configuration...$(RESET)"
	@command -v goreleaser >/dev/null 2>&1 || { echo "$(RED)goreleaser not found. Install with: make install-goreleaser$(RESET)"; exit 1; }
	@goreleaser check

deploy: release-dry-run ## alias for release-dry-run

# ==============================================================================
# Build Information
# ==============================================================================

.PHONY: build-info

build-info: ## show build information and current configuration
	@echo "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                         $(YELLOW)Build Information$(CYAN)                              ║"
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
	@echo "$(GREEN)🎯 Build Targets:$(RESET)"
	@echo "  • $(CYAN)build$(RESET)               Build golang binary"
	@echo "  • $(CYAN)install$(RESET)             Install golang binary to GOPATH/bin"
	@echo "  • $(CYAN)run$(RESET)                 Run the application"
	@echo "  • $(CYAN)bootstrap$(RESET)           Install build dependencies"
	@echo "  • $(CYAN)clean$(RESET)               Clean up build artifacts"
	@echo "  • $(CYAN)release-dry-run$(RESET)     Test goreleaser configuration"
	@echo "  • $(CYAN)release-snapshot$(RESET)    Create snapshot release"
	@echo "  • $(CYAN)release-check$(RESET)       Check goreleaser configuration"
