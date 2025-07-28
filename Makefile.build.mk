# Makefile.build - Build and Installation targets for gzh-manager-go
# Build, compilation, and installation management

# ==============================================================================
# Build Configuration
# ==============================================================================

# ==============================================================================
# Build Targets
# ==============================================================================

.PHONY: build build-git-extensions install run bootstrap clean release-dry-run release-snapshot release-check deploy

build: ## build golang binary
	@echo -e "$(CYAN)Building $(executablename)...$(RESET)"
	@go build -ldflags "-X main.version=$(VERSION)" -o $(executablename)
	@echo -e "$(GREEN)✅ Built $(executablename) successfully$(RESET)"

build-git-extensions: ## build git extension binaries
	@echo -e "$(CYAN)Building git-synclone...$(RESET)"
	@go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$$(date -u +%Y-%m-%d_%H:%M:%S)" -o git-synclone ./cmd/git-synclone
	@echo -e "$(GREEN)✅ Built git-synclone successfully$(RESET)"

install: build ## install golang binary
	@echo -e "$(CYAN)Installing $(executablename)...$(RESET)"
	@mv $(executablename) $(shell go env GOPATH)/bin/
	@echo -e "$(GREEN)✅ Installed $(executablename) to $(shell go env GOPATH)/bin/$(RESET)"

install-git-extensions: build-git-extensions ## install git extension binaries
	@echo -e "$(CYAN)Installing git-synclone...$(RESET)"
	@mv git-synclone $(shell go env GOPATH)/bin/
	@echo -e "$(GREEN)✅ Installed git-synclone to $(shell go env GOPATH)/bin/$(RESET)"
	@echo -e "$(YELLOW)You can now use: git synclone$(RESET)"

run: ## run the application (usage: make run [args...] or ARGS="args" make run)
	@echo -e "$(CYAN)Running application with version $(VERSION)...$(RESET)"
	@if [ "$(words $(MAKECMDGOALS))" -gt 1 ]; then \
		ARGS="$(filter-out run,$(MAKECMDGOALS))"; \
		echo -e "$(YELLOW)Arguments: $$ARGS$(RESET)"; \
		go run -ldflags "-X main.version=$(VERSION)" main.go $$ARGS; \
	elif [ -n "$(ARGS)" ]; then \
		echo -e "$(YELLOW)Arguments: $(ARGS)$(RESET)"; \
		go run -ldflags "-X main.version=$(VERSION)" main.go $(ARGS); \
	else \
		go run -ldflags "-X main.version=$(VERSION)" main.go; \
	fi

# Prevent make from interpreting arguments as targets
%:
	@:

bootstrap: ## install build dependencies
	@echo -e "$(CYAN)Installing build dependencies...$(RESET)"
	go generate -tags tools tools/tools.go
	@echo -e "$(GREEN)✅ Build dependencies installed$(RESET)"

clean: ## clean up environment
	@echo -e "$(CYAN)Cleaning up build artifacts...$(RESET)"
	@rm -rf coverage.out coverage.html dist/ $(executablename) git-synclone
	@rm -f $(shell go env GOPATH)/bin/$(executablename)
	@rm -f $(shell go env GOPATH)/bin/$(projectname)
	@rm -f $(shell go env GOPATH)/bin/git-synclone
	@rm -f lint-report.json gosec-report.json
	@echo -e "$(GREEN)✅ Cleanup completed$(RESET)"

# ==============================================================================
# Release Targets
# ==============================================================================

release-dry-run: ## run goreleaser in dry-run mode
	@echo -e "$(CYAN)Running goreleaser in dry-run mode...$(RESET)"
	@command -v goreleaser >/dev/null 2>&1 || { echo -e "$(RED)goreleaser not found. Install with: make install-goreleaser$(RESET)"; exit 1; }
	@goreleaser release --snapshot --clean --skip=publish

release-snapshot: ## create a snapshot release
	@echo -e "$(CYAN)Creating snapshot release...$(RESET)"
	@command -v goreleaser >/dev/null 2>&1 || { echo -e "$(RED)goreleaser not found. Install with: make install-goreleaser$(RESET)"; exit 1; }
	@goreleaser release --snapshot --clean

release-check: ## check goreleaser configuration
	@echo -e "$(CYAN)Checking goreleaser configuration...$(RESET)"
	@command -v goreleaser >/dev/null 2>&1 || { echo -e "$(RED)goreleaser not found. Install with: make install-goreleaser$(RESET)"; exit 1; }
	@goreleaser check

deploy: release-dry-run ## alias for release-dry-run

# ==============================================================================
# Build Information
# ==============================================================================

.PHONY: build-info

build-info: ## show build information and current configuration
	@echo -e "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo -e "║                         $(YELLOW)Build Information$(CYAN)                              ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo -e "$(RESET)"
	@echo -e "$(GREEN)📋 Project Details:$(RESET)"
	@echo -e "  Name:           $(YELLOW)$(projectname)$(RESET)"
	@echo -e "  Executable:     $(YELLOW)$(executablename)$(RESET)"
	@echo -e "  Version:        $(YELLOW)$(VERSION)$(RESET)"
	@echo ""
	@echo -e "$(GREEN)🏗️  Build Environment:$(RESET)"
	@echo "  Go Version:     $$(go version | cut -d' ' -f3)"
	@echo -e "  GOPROXY:        $(GOPROXY)"
	@echo -e "  GOSUMDB:        $(GOSUMDB)"
	@echo "  GOPATH:         $$(go env GOPATH)"
	@echo "  GOROOT:         $$(go env GOROOT)"
	@echo ""
	@echo -e "$(GREEN)🎯 Build Targets:$(RESET)"
	@echo -e "  • $(CYAN)build$(RESET)               Build golang binary"
	@echo -e "  • $(CYAN)install$(RESET)             Install golang binary to GOPATH/bin"
	@echo -e "  • $(CYAN)run$(RESET)                 Run the application"
	@echo -e "  • $(CYAN)bootstrap$(RESET)           Install build dependencies"
	@echo -e "  • $(CYAN)clean$(RESET)               Clean up build artifacts"
	@echo -e "  • $(CYAN)release-dry-run$(RESET)     Test goreleaser configuration"
	@echo -e "  • $(CYAN)release-snapshot$(RESET)    Create snapshot release"
	@echo -e "  • $(CYAN)release-check$(RESET)       Check goreleaser configuration"
