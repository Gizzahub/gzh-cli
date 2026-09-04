# Makefile.build - Build and Installation targets for gzh-cli
# Build, compilation, and installation management

# ==============================================================================
# Build Configuration
# ==============================================================================

# Go environment configuration
# Use system Go if asdf golang is not available
ASDF_GOLANG_PATH := $(shell asdf where golang 2>/dev/null || echo "")
ifneq ($(ASDF_GOLANG_PATH),)
export GOROOT := $(ASDF_GOLANG_PATH)/go
else
# Use system Go installation
export GOROOT := $(shell go env GOROOT)
endif

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
.PHONY: release-healthcheck
.PHONY: completions completions-check

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
	@rm -f lint-report.json gosec-report.json
	@echo -e "$(GREEN)✅ Cleanup completed$(RESET)"

# ==============================================================================
# Shell Completions
# ==============================================================================

# `.goreleaser.yml`'s before.hooks runs ./scripts/completions.sh on every release
# and snapshot, and that script starts with `rm -rf completions`. So whatever is
# tracked here is NOT what ships unless it is byte-identical to the generator's
# output -- a reader auditing completions/gzh-manager.zsh in git is otherwise
# reading a file the release overwrites. completions-check is what makes the two
# the same artifact instead of two files that happen to look alike.
#
# Note for anyone adding a gate here: .make/build.mk defines a `%:` catchall that
# answers every undefined target with success, so `make some-typo` exits 0. A
# check is only real if the target name actually resolves; grep for the rule when
# binding this to an acceptance criterion.

completions: ## regenerate the tracked shell completions from the generator
	@echo -e "$(CYAN)Regenerating shell completions...$(RESET)"
	@for sh in bash zsh fish; do \
		go run ./scripts/completiongen $$sh > "completions/gzh-manager.$$sh"; \
	done
	@echo -e "$(GREEN)✅ completions/ regenerated$(RESET)"

completions-check: ## fail if tracked completions/ drifted from the generator
	@echo -e "$(CYAN)Checking completions for drift...$(RESET)"
	@for sh in bash zsh fish; do \
		go run ./scripts/completiongen $$sh | diff -u "completions/gzh-manager.$$sh" - || { \
			echo "❌ completions/gzh-manager.$$sh is not what ./scripts/completiongen $$sh produces."; \
			echo "   The release regenerates it, so the tracked file would be discarded."; \
			echo "   Run: make completions   (then commit the result)"; \
			exit 1; \
		}; \
	done
	@echo -e "$(GREEN)✅ completions/ matches the generator$(RESET)"

# ==============================================================================
# Release Targets
# ==============================================================================

# Every release target runs against the pinned toolchain only. `bin/tools` is
# prepended to PATH because GoReleaser resolves `syft` and `cosign` by name, and
# an unpinned copy earlier on the developer's PATH would silently win.
# verify-release-tools and release-healthcheck live in .make/tools.mk / here so
# that a missing or drifted tool fails before any artifact is produced, instead
# of GoReleaser quietly dropping the SBOM or signature stage.

release-healthcheck: verify-release-tools verify-release-pins ## fail fast on tools GoReleaser itself needs
	@echo -e "$(CYAN)Running goreleaser healthcheck...$(RESET)"
	@PATH="$(RELEASE_PATH)" "$(GORELEASER)" healthcheck

# `sign` is excluded from both local targets on purpose. Signing is keyless: it
# mints a Fulcio certificate from an ambient OIDC token and writes an entry to
# the public Rekor transparency log, which is irreversible. The release job has
# that ambient token (id-token: write); a laptop does not, so locally cosign
# would fall back to an interactive browser flow and, if completed, would
# publish a throwaway snapshot to a public log. The signing contract is instead
# gated by `goreleaser check` and by release-healthcheck, which fail if cosign
# is missing or the config is invalid.

release-dry-run: release-healthcheck ## dry-run release: builds and Docker images, no publish, no signing
	@echo -e "$(CYAN)Running goreleaser in dry-run mode (publish and sign excluded)...$(RESET)"
	@PATH="$(RELEASE_PATH)" "$(GORELEASER)" release --snapshot --clean --skip=publish,sign

release-snapshot: release-healthcheck ## snapshot release with publish, Docker and signing excluded
	@echo -e "$(CYAN)Creating snapshot release (publish, docker and sign excluded)...$(RESET)"
	@PATH="$(RELEASE_PATH)" "$(GORELEASER)" release --snapshot --clean --skip=publish,docker,sign

release-check: install-goreleaser ## check goreleaser configuration with the pinned version
	@echo -e "$(CYAN)Checking goreleaser configuration with $(GORELEASER_VERSION)...$(RESET)"
	@"$(GORELEASER)" check

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
	@echo -e "  • $(CYAN)release-healthcheck$(RESET) Verify the pinned release toolchain"
	@echo -e "  • $(CYAN)release-dry-run$(RESET)     Dry-run release, publish and sign excluded"
	@echo -e "  • $(CYAN)release-snapshot$(RESET)    Snapshot release, publish/docker/sign excluded"
	@echo -e "  • $(CYAN)release-check$(RESET)       Check goreleaser configuration"
