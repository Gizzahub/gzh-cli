# Makefile.build - Build and Installation targets for gzh-cli
# Build, compilation, and installation management

# ==============================================================================
# Build Configuration
# ==============================================================================

# Go environment configuration for crypto/mlkem and iter package support
# Use system Go if asdf golang is not available
ASDF_GOLANG_PATH := $(shell asdf where golang 2>/dev/null || echo "")
ifneq ($(ASDF_GOLANG_PATH),)
export GOROOT := $(ASDF_GOLANG_PATH)/go
else
# Use system Go installation
export GOROOT := $(shell go env GOROOT)
endif
export GOEXPERIMENT := rangefunc

# Detect OS-specific executable extension (e.g., .exe on Windows)
BINEXT := $(shell go env GOEXE)
BINARY := $(executablename)$(BINEXT)
GOBIN := $(shell go env GOBIN)
GOPATH := $(shell go env GOPATH)

# OS-specific path separator and binary install dir
ifeq ($(OS),Windows_NT)
SEP := \\\\
else
SEP := /
endif

ifeq ($(strip $(GOBIN)),)
  ifeq ($(OS),Windows_NT)
    BINDIR := $(GOPATH)$(SEP)bin
  else
    BINDIR := $(GOPATH)$(SEP)bin
  endif
else
  BINDIR := $(GOBIN)
endif

# ==============================================================================
# Build Targets
# ==============================================================================

# Version file (single source of truth)
VERSION_FILE := internal/version/version.go

.PHONY: build install run bootstrap clean release-dry-run release-snapshot release-check deploy bump-version

## bump-version: Bump patch version if there are changes or new commit
bump-version:
	@LAST_COMMIT=$$(cat .last_built_commit 2>/dev/null || echo ""); \
	CURRENT_COMMIT=$$(git rev-parse HEAD 2>/dev/null || echo "none"); \
	DIRTY=$$(git status --porcelain | grep -v '$(VERSION_FILE)'); \
	if [ "$$CURRENT_COMMIT" != "$$LAST_COMMIT" ] || [ -n "$$DIRTY" ]; then \
		CURRENT_VERSION=$$(grep 'Version =' $(VERSION_FILE) | cut -d'"' -f2); \
		MAJOR=$$(echo $$CURRENT_VERSION | cut -d. -f1); \
		MINOR=$$(echo $$CURRENT_VERSION | cut -d. -f2); \
		PATCH=$$(echo $$CURRENT_VERSION | cut -d. -f3); \
		NEW_PATCH=$$(($$PATCH + 1)); \
		NEW_VERSION="$$MAJOR.$$MINOR.$$NEW_PATCH"; \
		sed -i.bak "s/Version = \"$$CURRENT_VERSION\"/Version = \"$$NEW_VERSION\"/" $(VERSION_FILE) && rm -f $(VERSION_FILE).bak; \
		echo "$$CURRENT_COMMIT" > .last_built_commit; \
		printf "$(YELLOW)Version bumped: %s → %s$(RESET)\n" "$$CURRENT_VERSION" "$$NEW_VERSION"; \
	else \
		printf "$(GREEN)Version unchanged: %s$(RESET)\n" "$$(grep 'Version =' $(VERSION_FILE) | cut -d'\"' -f2)"; \
	fi

build: ## build golang binary
	$(eval VERSION := $(shell grep 'Version =' $(VERSION_FILE) | cut -d'"' -f2))
	@printf "$(CYAN)Building %s v%s...$(RESET)\n" "$(BINARY)" "$(VERSION)"
	@go build -trimpath -o $(BINARY) ./cmd/gz
	@printf "$(GREEN)Built %s v%s successfully$(RESET)\n" "$(BINARY)" "$(VERSION)"

install: bump-version build ## install golang binary (auto bumps patch version)
	@printf "$(CYAN)Installing to %s$(RESET)\n" "$(BINDIR)$(SEP)$(BINARY)"
	@mv $(BINARY) "$(BINDIR)"/
	@printf "$(GREEN)Installed %s to %s$(RESET)\n" "$(BINARY)" "$(BINDIR)$(SEP)$(BINARY)"

run: ## run the application (usage: make run [args...] or ARGS="args" make run)
	$(eval VERSION := $(shell grep 'Version =' $(VERSION_FILE) | cut -d'"' -f2))
	@echo -e "$(CYAN)Running application with version $(VERSION)...$(RESET)"
	@if [ "$(words $(MAKECMDGOALS))" -gt 1 ]; then \
		ARGS="$(filter-out run,$(MAKECMDGOALS))"; \
		echo -e "$(YELLOW)Arguments: $$ARGS$(RESET)"; \
		go run ./cmd/gz $$ARGS; \
	elif [ -n "$(ARGS)" ]; then \
		echo -e "$(YELLOW)Arguments: $(ARGS)$(RESET)"; \
		go run ./cmd/gz $(ARGS); \
	else \
		go run ./cmd/gz; \
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
	@rm -rf coverage.out coverage.html dist/ $(executablename) $(BINARY)
	@rm -f $(shell go env GOPATH)/bin/$(executablename)
	@rm -f $(shell go env GOPATH)/bin/$(BINARY)
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
	@echo -e "  Executable:     $(YELLOW)$(BINARY)$(RESET)"
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
