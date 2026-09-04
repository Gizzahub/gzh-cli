# Makefile.tools - Tool Installation and Management for gzh-cli
# Development tools, linters, formatters, and utilities

# ==============================================================================
# Tool Configuration
# ==============================================================================

GOLANGCI_LINT_VERSION := v2.13.1
GOLANGCI_LINT_MODULE := github.com/golangci/golangci-lint/v2
GOLANGCI_LINT_INSTALL := $(GOLANGCI_LINT_MODULE)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT_DIR := $(CURDIR)/bin/tools
GOLANGCI_LINT := $(GOLANGCI_LINT_DIR)/golangci-lint$(shell go env GOEXE)
GOLANGCI_LINT_RUN_FLAGS := --allow-serial-runners
GOSEC_VERSION := v2.28.0
GOSEC_MODULE := github.com/securego/gosec/v2
GOSEC_INSTALL := $(GOSEC_MODULE)/cmd/gosec@$(GOSEC_VERSION)
GOSEC_DIR := $(CURDIR)/bin/tools
GOSEC := $(GOSEC_DIR)/gosec$(shell go env GOEXE)
GOSEC_SCAN_FLAGS := -conf=.gosec.json -exclude-generated -exclude-dir=vendor -exclude-dir=node_modules -exclude-dir=.git -exclude-dir=tmp -tests -confidence=medium -severity=medium

# ------------------------------------------------------------------------------
# Format toolchain pins
# ------------------------------------------------------------------------------
# These MUST equal the versions the pinned golangci-lint embeds. `make lint` runs
# golangci-lint's own copies of gofumpt/gci and the x/tools import machinery;
# `make fmt` and `make format-strict` run these standalone binaries. When the two
# drift, the formatter that writes the file and the formatter that judges it are
# different programs, and the repository oscillates. verify-format-pins turns that
# requirement into a check instead of a comment — run it after bumping
# GOLANGCI_LINT_VERSION.
#
# Before this pin the installers guarded with `which <tool>`, which only asks
# whether a binary exists. On the machine this was written on that had already
# drifted: gofumpt v0.10.0 / gci v0.14.0 / x-tools v0.38.0 on PATH against the
# v0.11.0 / v0.13.7 / v0.49.0 that golangci-lint v2.13.1 actually uses.
FORMAT_TOOLS_DIR := $(CURDIR)/bin/tools

GOFUMPT_VERSION := v0.11.0
GOFUMPT_MODULE := mvdan.cc/gofumpt
GOFUMPT_INSTALL := $(GOFUMPT_MODULE)@$(GOFUMPT_VERSION)
GOFUMPT := $(FORMAT_TOOLS_DIR)/gofumpt$(shell go env GOEXE)

GCI_VERSION := v0.13.7
GCI_MODULE := github.com/daixiang0/gci
GCI_INSTALL := $(GCI_MODULE)@$(GCI_VERSION)
GCI := $(FORMAT_TOOLS_DIR)/gci$(shell go env GOEXE)

# goimports ships from the x/tools module, so `go version -m` reports the module
# path rather than the command path. The version below is x/tools', not a
# separate goimports version.
GOIMPORTS_VERSION := v0.49.0
GOIMPORTS_MODULE := golang.org/x/tools
GOIMPORTS_INSTALL := $(GOIMPORTS_MODULE)/cmd/goimports@$(GOIMPORTS_VERSION)
GOIMPORTS := $(FORMAT_TOOLS_DIR)/goimports$(shell go env GOEXE)

# mdformat is a Python package, so it cannot live in bin/tools next to the Go
# binaries. `pip install --user` puts the launcher in ~/.local/bin and the
# package in ~/.local/lib/python3.X/site-packages. Those are two directories, and
# CI caches only the first — a restored launcher whose site-packages are gone
# passes `which mdformat` and then dies with ModuleNotFoundError. Verifying with
# `mdformat --version` instead of `which` catches both that and a stale pin,
# because a broken install cannot print its version.
MDFORMAT_VERSION := 0.7.22
MDFORMAT_GFM_VERSION := 1.0.0
MDFORMAT_TABLES_VERSION := 1.0.0

# Exit 0 only when the installed binary really is the pinned module version and
# was built with the current Go toolchain — same test install-golangci-lint uses.
GOFUMPT_VERSION_OK = go version -m "$(GOFUMPT)" 2>/dev/null | \
	awk -v want="$$(go env GOVERSION)" 'NR == 1 { built = $$NF } $$1 == "mod" && $$2 == "$(GOFUMPT_MODULE)" && $$3 == "$(GOFUMPT_VERSION)" { found = 1 } END { exit !(found && built == want) }'
GCI_VERSION_OK = go version -m "$(GCI)" 2>/dev/null | \
	awk -v want="$$(go env GOVERSION)" 'NR == 1 { built = $$NF } $$1 == "mod" && $$2 == "$(GCI_MODULE)" && $$3 == "$(GCI_VERSION)" { found = 1 } END { exit !(found && built == want) }'
GOIMPORTS_VERSION_OK = go version -m "$(GOIMPORTS)" 2>/dev/null | \
	awk -v want="$$(go env GOVERSION)" 'NR == 1 { built = $$NF } $$1 == "mod" && $$2 == "$(GOIMPORTS_MODULE)" && $$3 == "$(GOIMPORTS_VERSION)" { found = 1 } END { exit !(found && built == want) }'
MDFORMAT_VERSION_OK = mdformat --version 2>/dev/null | \
	grep -Fq "mdformat $(MDFORMAT_VERSION) (mdformat_tables $(MDFORMAT_TABLES_VERSION), mdformat-gfm $(MDFORMAT_GFM_VERSION))"

# ------------------------------------------------------------------------------
# Release toolchain pins
# ------------------------------------------------------------------------------
# These are the local half of the release toolchain contract. The CI half lives
# in .github/workflows/release.yml and MUST pin the same three versions.
# Bump both together and re-run `make release-check release-snapshot`.
# Rationale and the upstream evidence for each pin: docs/70-deployment/70-releases.md
#
# Locally the tools are `go install`ed into bin/tools and verified through
# `go version -m`, which reports the true module version even when the binary
# was not built with the project's release ldflags. CI installs the official
# release binaries instead and verifies them with each tool's own `version`
# output. Both paths are exact; neither falls back to "latest".
RELEASE_TOOLS_DIR := $(CURDIR)/bin/tools

# The CI half of the pin contract. verify-release-pins enforces that the two
# files agree, so the "MUST stay in sync" comments below are checkable rather
# than aspirational.
RELEASE_WORKFLOW := $(CURDIR)/.github/workflows/release.yml

GORELEASER_VERSION := v2.18.0
GORELEASER_MODULE := github.com/goreleaser/goreleaser/v2
GORELEASER_INSTALL := $(GORELEASER_MODULE)@$(GORELEASER_VERSION)
GORELEASER := $(RELEASE_TOOLS_DIR)/goreleaser$(shell go env GOEXE)

SYFT_VERSION := v1.51.1
SYFT_MODULE := github.com/anchore/syft
SYFT_INSTALL := $(SYFT_MODULE)/cmd/syft@$(SYFT_VERSION)
SYFT := $(RELEASE_TOOLS_DIR)/syft$(shell go env GOEXE)

COSIGN_VERSION := v3.1.3
COSIGN_MODULE := github.com/sigstore/cosign/v3
COSIGN_INSTALL := $(COSIGN_MODULE)/cmd/cosign@$(COSIGN_VERSION)
COSIGN := $(RELEASE_TOOLS_DIR)/cosign$(shell go env GOEXE)

# GoReleaser shells out to `syft` and `cosign` by name, so the pinned copies
# must win over anything else on the developer's PATH.
RELEASE_PATH := $(RELEASE_TOOLS_DIR):$(PATH)

# Each *_VERSION_OK is a shell test that exits 0 only when the binary really was
# built from the pinned module version. `go version -m` is authoritative here:
# unlike `<tool> version`, it cannot report "unknown" just because the binary was
# built without the upstream release ldflags.
GORELEASER_VERSION_OK = go version -m "$(GORELEASER)" 2>/dev/null | \
	awk '$$1 == "mod" && $$2 == "$(GORELEASER_MODULE)" && $$3 == "$(GORELEASER_VERSION)" { found = 1 } END { exit !found }'
SYFT_VERSION_OK = go version -m "$(SYFT)" 2>/dev/null | \
	awk '$$1 == "mod" && $$2 == "$(SYFT_MODULE)" && $$3 == "$(SYFT_VERSION)" { found = 1 } END { exit !found }'
COSIGN_VERSION_OK = go version -m "$(COSIGN)" 2>/dev/null | \
	awk '$$1 == "mod" && $$2 == "$(COSIGN_MODULE)" && $$3 == "$(COSIGN_VERSION)" { found = 1 } END { exit !found }'

# Reported version of a pinned tool, or empty when it is missing/unreadable.
GORELEASER_FOUND = $$(go version -m "$(GORELEASER)" 2>/dev/null | awk '$$1 == "mod" && $$2 == "$(GORELEASER_MODULE)" { print $$3 }')
SYFT_FOUND = $$(go version -m "$(SYFT)" 2>/dev/null | awk '$$1 == "mod" && $$2 == "$(SYFT_MODULE)" { print $$3 }')
COSIGN_FOUND = $$(go version -m "$(COSIGN)" 2>/dev/null | awk '$$1 == "mod" && $$2 == "$(COSIGN_MODULE)" { print $$3 }')

# ==============================================================================
# Core Tool Installation
# ==============================================================================

.PHONY: install-tools install-format-tools install-analysis-tools install-goreleaser
.PHONY: install-gofumpt install-gci install-goimports install-mdformat verify-format-pins
.PHONY: install-golangci-lint install-gosec install-pre-commit-tools install-docs-tools
.PHONY: install-syft install-cosign install-release-tools verify-release-tools
.PHONY: verify-release-pins

install-tools: install-format-tools install-analysis-tools install-golangci-lint install-release-tools ## install all development tools
	@echo -e "$(GREEN)✅ All development tools installed!$(RESET)"

install-format-tools: install-gofumpt install-gci install-goimports install-mdformat ## install the exact pinned formatting toolchain
	@echo -e "$(GREEN)✅ All formatting tools installed!$(RESET)"

install-gofumpt: ## install the pinned gofumpt release
	@mkdir -p "$(FORMAT_TOOLS_DIR)"
	@if ! $(GOFUMPT_VERSION_OK); then \
		echo "Installing gofumpt $(GOFUMPT_VERSION) to $(GOFUMPT)..."; \
		GOBIN="$(FORMAT_TOOLS_DIR)" go install $(GOFUMPT_INSTALL); \
	fi
	@$(GOFUMPT_VERSION_OK) || { \
		echo "gofumpt installation did not produce $(GOFUMPT_MODULE) $(GOFUMPT_VERSION) built with $$(go env GOVERSION): $(GOFUMPT)" >&2; \
		exit 1; \
	}

install-gci: ## install the pinned gci release
	@mkdir -p "$(FORMAT_TOOLS_DIR)"
	@if ! $(GCI_VERSION_OK); then \
		echo "Installing gci $(GCI_VERSION) to $(GCI)..."; \
		GOBIN="$(FORMAT_TOOLS_DIR)" go install $(GCI_INSTALL); \
	fi
	@$(GCI_VERSION_OK) || { \
		echo "gci installation did not produce $(GCI_MODULE) $(GCI_VERSION) built with $$(go env GOVERSION): $(GCI)" >&2; \
		exit 1; \
	}

install-goimports: ## install the pinned goimports (x/tools) release
	@mkdir -p "$(FORMAT_TOOLS_DIR)"
	@if ! $(GOIMPORTS_VERSION_OK); then \
		echo "Installing goimports from $(GOIMPORTS_MODULE) $(GOIMPORTS_VERSION) to $(GOIMPORTS)..."; \
		GOBIN="$(FORMAT_TOOLS_DIR)" go install $(GOIMPORTS_INSTALL); \
	fi
	@$(GOIMPORTS_VERSION_OK) || { \
		echo "goimports installation did not produce $(GOIMPORTS_MODULE) $(GOIMPORTS_VERSION) built with $$(go env GOVERSION): $(GOIMPORTS)" >&2; \
		exit 1; \
	}

install-mdformat: ## install the pinned mdformat and its plugins
	@if ! $(MDFORMAT_VERSION_OK); then \
		echo "Installing mdformat $(MDFORMAT_VERSION)..."; \
		pip install --user --upgrade \
			"mdformat==$(MDFORMAT_VERSION)" \
			"mdformat-gfm==$(MDFORMAT_GFM_VERSION)" \
			"mdformat-tables==$(MDFORMAT_TABLES_VERSION)"; \
	fi
	@$(MDFORMAT_VERSION_OK) || { \
		echo "mdformat is not the pinned $(MDFORMAT_VERSION) (mdformat_tables $(MDFORMAT_TABLES_VERSION), mdformat-gfm $(MDFORMAT_GFM_VERSION)); got: $$(mdformat --version 2>&1 || echo 'not runnable')" >&2; \
		exit 1; \
	}

verify-format-pins: ## fail unless the format pins match what the pinned golangci-lint embeds
	@if [ ! -x "$(GOLANGCI_LINT)" ]; then \
		echo "verify-format-pins needs the pinned golangci-lint; run 'make install-golangci-lint' first" >&2; \
		exit 1; \
	fi
	@status=0; \
	for spec in "$(GOFUMPT_MODULE) $(GOFUMPT_VERSION)" "$(GCI_MODULE) $(GCI_VERSION)" "$(GOIMPORTS_MODULE) $(GOIMPORTS_VERSION)"; do \
		module=$${spec%% *}; want=$${spec##* }; \
		got=$$(go version -m "$(GOLANGCI_LINT)" 2>/dev/null | awk -v m="$$module" '$$1 == "dep" && $$2 == m { print $$3 }'); \
		if [ -z "$$got" ]; then \
			echo "verify-format-pins: golangci-lint $(GOLANGCI_LINT_VERSION) does not depend on $$module" >&2; \
			status=1; \
		elif [ "$$got" != "$$want" ]; then \
			echo "verify-format-pins: $$module pinned to $$want here but golangci-lint $(GOLANGCI_LINT_VERSION) embeds $$got" >&2; \
			status=1; \
		fi; \
	done; \
	exit $$status
	@echo -e "$(GREEN)✅ Format pins match golangci-lint $(GOLANGCI_LINT_VERSION)$(RESET)"

# ------------------------------------------------------------------------------
# Analysis, mock and docs toolchain pins
# ------------------------------------------------------------------------------
# Before this, all seven of these installed with `@latest`: five behind a
# `command -v` guard and two (benchstat, godoc) with no guard at all, reinstalled
# on every invocation -- see install-docs-tools below. Either way the version any
# given machine ran was decided by the date it first happened to be missing the
# binary. Two machines, or one machine before and
# after a cache eviction, could disagree about what `make analyze` even means.
# The guard made it worse rather than better: `command -v` only asks whether a
# name resolves on PATH, so an ancient binary from an unrelated project silently
# satisfied it forever.
#
# The drift this fixes is a local-workstation one, not a CI one. An earlier
# draft of this comment blamed CI's tool cache; that was wrong and an
# independent review caught it. CI's four make invocations are format-strict,
# lint-check, security-json (main.yml) and release-snapshot (release.yml), and
# `make -n` on each shows none of them reaches any of these seven -- they pull
# gci/gofumpt/goimports, golangci-lint and gosec only. So `@latest` was never
# resolved on a runner at all. It was resolved once per developer machine, into
# ~/go/bin, and then never again, because `command -v` kept finding it.
#
# Installing from here into bin/tools with GOBIN is still the right place for
# the pins: main.yml's tool cache key already hashes Makefile and .make/tools.mk
# and its paths already include bin/tools, so if CI ever does run `make analyze`
# the version is decided by a file the key can see -- no workflow edit needed.
#
# Versions resolved with `go list -m -versions` on 2026-09-04 and recorded here
# rather than left to resolution time. One is not a guess at all: mockgen ships
# from go.uber.org/mock, already a direct dependency in go.mod, so its pin is
# read off go.mod rather than chosen. godoc looked like it would be the same
# story against GOIMPORTS_VERSION and turned out not to be -- see its own note.
ANALYSIS_TOOLS_DIR := $(CURDIR)/bin/tools

GOCYCLO_VERSION := v0.6.0
GOCYCLO_MODULE := github.com/fzipp/gocyclo
GOCYCLO_INSTALL := $(GOCYCLO_MODULE)/cmd/gocyclo@$(GOCYCLO_VERSION)
GOCYCLO := $(ANALYSIS_TOOLS_DIR)/gocyclo$(shell go env GOEXE)

STATICCHECK_VERSION := v0.8.1
STATICCHECK_MODULE := honnef.co/go/tools
STATICCHECK_INSTALL := $(STATICCHECK_MODULE)/cmd/staticcheck@$(STATICCHECK_VERSION)
STATICCHECK := $(ANALYSIS_TOOLS_DIR)/staticcheck$(shell go env GOEXE)

DUPL_VERSION := v1.1.0
DUPL_MODULE := github.com/mibk/dupl
DUPL_INSTALL := $(DUPL_MODULE)@$(DUPL_VERSION)
DUPL := $(ANALYSIS_TOOLS_DIR)/dupl$(shell go env GOEXE)

INEFFASSIGN_VERSION := v0.2.0
INEFFASSIGN_MODULE := github.com/gordonklaus/ineffassign
INEFFASSIGN_INSTALL := $(INEFFASSIGN_MODULE)@$(INEFFASSIGN_VERSION)
INEFFASSIGN := $(ANALYSIS_TOOLS_DIR)/ineffassign$(shell go env GOEXE)

# Must equal go.mod's go.uber.org/mock. The generated mocks and the library that
# runs them come from one module; a mockgen from a different release can emit
# code the vendored gomock does not accept.
MOCKGEN_VERSION := v0.6.0
MOCKGEN_MODULE := go.uber.org/mock
MOCKGEN_INSTALL := $(MOCKGEN_MODULE)/mockgen@$(MOCKGEN_VERSION)
MOCKGEN := $(ANALYSIS_TOOLS_DIR)/mockgen$(shell go env GOEXE)

# golang.org/x/perf has never been tagged -- `go list -m -versions` returns an
# empty list -- so a pseudo-version is the only exact pin available. It is still
# exact; it names one commit.
BENCHSTAT_VERSION := v0.0.0-20260825160852-19be9d8e6c70
BENCHSTAT_MODULE := golang.org/x/perf
BENCHSTAT_INSTALL := $(BENCHSTAT_MODULE)/cmd/benchstat@$(BENCHSTAT_VERSION)
BENCHSTAT := $(ANALYSIS_TOOLS_DIR)/benchstat$(shell go env GOEXE)

# Not $(GOIMPORTS_VERSION), even though the import path still starts with
# golang.org/x/tools. cmd/godoc was split into its own nested module and tagged
# `v0.1.0-deprecated`; the main x/tools module no longer contains the package, so
# `go install golang.org/x/tools/cmd/godoc@v0.49.0` fails outright.
#
# This is the sharpest evidence for the whole card. `@latest` here never resolved
# to the x/tools version this repo pins -- it silently walked off to a separate,
# upstream-deprecated module, and the `command -v` guard meant nobody would ever
# see it happen. The pin is exact; whether a deprecated godoc should still be a
# make target at all is a separate call and not made here.
GODOC_VERSION := v0.1.0-deprecated
GODOC_MODULE := golang.org/x/tools/cmd/godoc
GODOC_INSTALL := $(GODOC_MODULE)@$(GODOC_VERSION)
GODOC := $(ANALYSIS_TOOLS_DIR)/godoc$(shell go env GOEXE)

# The same freshness test install-golangci-lint and the format pins already use,
# parameterised so seven more tools do not need seven more copies of the awk.
# $(1)=binary  $(2)=module  $(3)=version
tool_version_ok = go version -m "$(1)" 2>/dev/null | \
	awk -v want="$$(go env GOVERSION)" -v mod="$(2)" -v ver="$(3)" 'NR == 1 { built = $$NF } $$1 == "mod" && $$2 == mod && $$3 == ver { found = 1 } END { exit !(found && built == want) }'

# Canned recipe: install the pin if the binary is not already exactly it, then
# verify. The second check is the one that matters -- it is what makes a missing
# or wrong tool a non-zero exit instead of a target that quietly did nothing.
#
# The `rm -f` is load-bearing, and measured rather than assumed: `go install`
# overwrites an existing Go binary happily, but refuses a target path that is not
# an object file at all ("build output ... already exists and is not an object
# file"), so a truncated or half-written file in bin/tools would otherwise wedge
# the target forever behind an error that does not name its own fix.
#
# It is gated on exactly that condition, and no wider one. An unconditional `rm`
# turns every install failure -- offline, proxy 5xx, a version that does not
# exist, a bad GOFLAGS -- into the loss of a working pinned binary, because the
# file is already gone when `go install` starts and nothing replaces it. Gated,
# the failure mode is "the old binary is still there and the pin check fails
# loudly", which is recoverable; ungated it is "the tool is gone and `make lint`
# stops working", which is not.
#
# The gate reads `go version`'s stdout, not its exit status, because the exit
# status does not answer the question. Measured with stderr discarded, go1.26.7:
#
#   input               rc  stdout                         `: go`?
#   text file            0  (empty)                        no
#   /bin/echo            0  (empty)                        no
#   real Go binary       0  "bin/tools/gocyclo: go1.26.7"  YES
#   missing file         1  (empty)                        no
#
# rc is 0 for three of the four, including the text file -- the one case the
# `rm` exists for -- so gating on rc would break self-healing outright. The
# diagnostics ("not a Go executable", "unrecognized file format") go to stderr,
# which the pipe already discards, so stdout is either a version line or nothing
# and `: go` separates all four.
# $(1)=display name  $(2)=binary  $(3)=module  $(4)=version  $(5)=install spec
define install_pinned_tool
@mkdir -p "$(ANALYSIS_TOOLS_DIR)"
@if ! $(call tool_version_ok,$(2),$(3),$(4)); then \
	echo "Installing $(1) $(4) to $(2)..."; \
	if [ -e "$(2)" ] && ! go version "$(2)" 2>/dev/null | grep -q ": go"; then \
		rm -f "$(2)"; \
	fi; \
	GOBIN="$(ANALYSIS_TOOLS_DIR)" go install $(5); \
fi
@$(call tool_version_ok,$(2),$(3),$(4)) || { \
	echo "$(1) installation did not produce $(3) $(4) built with $$(go env GOVERSION): $(2)" >&2; \
	exit 1; \
}
endef

.PHONY: install-gocyclo install-staticcheck install-dupl install-ineffassign
.PHONY: install-mockgen install-benchstat install-godoc

install-gocyclo: ## install the pinned gocyclo
	$(call install_pinned_tool,gocyclo,$(GOCYCLO),$(GOCYCLO_MODULE),$(GOCYCLO_VERSION),$(GOCYCLO_INSTALL))

install-staticcheck: ## install the pinned staticcheck
	$(call install_pinned_tool,staticcheck,$(STATICCHECK),$(STATICCHECK_MODULE),$(STATICCHECK_VERSION),$(STATICCHECK_INSTALL))

install-dupl: ## install the pinned dupl
	$(call install_pinned_tool,dupl,$(DUPL),$(DUPL_MODULE),$(DUPL_VERSION),$(DUPL_INSTALL))

install-ineffassign: ## install the pinned ineffassign
	$(call install_pinned_tool,ineffassign,$(INEFFASSIGN),$(INEFFASSIGN_MODULE),$(INEFFASSIGN_VERSION),$(INEFFASSIGN_INSTALL))

install-mockgen: ## install the pinned mockgen
	$(call install_pinned_tool,mockgen,$(MOCKGEN),$(MOCKGEN_MODULE),$(MOCKGEN_VERSION),$(MOCKGEN_INSTALL))

install-benchstat: ## install the pinned benchstat
	$(call install_pinned_tool,benchstat,$(BENCHSTAT),$(BENCHSTAT_MODULE),$(BENCHSTAT_VERSION),$(BENCHSTAT_INSTALL))

install-godoc: ## install the pinned godoc
	$(call install_pinned_tool,godoc,$(GODOC),$(GODOC_MODULE),$(GODOC_VERSION),$(GODOC_INSTALL))

install-analysis-tools: install-gosec install-gocyclo install-ineffassign install-dupl install-staticcheck ## install code analysis tools
	@echo -e "$(GREEN)✅ All analysis tools installed!$(RESET)"

# The freshness check reads build info, not the release string alone. `go
# install` builds with the *active* toolchain and ignores go.mod's `toolchain`
# directive, so the Go that built this binary is decided by whatever mise
# resolves — which differs between the devbox tree and a ~/worktrees task
# worktree. When those disagree, golangci-lint's go/types reads a stdlib it
# cannot parse and the run dies mid-analysis instead of reporting findings:
#
#   panic: file requires newer Go version go1.27 (application built with go1.26)
#
# Comparing only the release string cannot see that: the version never moved,
# so the old check printed "Ensuring golangci-lint v2.13.1..." and reinstalled
# nothing, leaving the mismatched binary in place (measured 2026-09-02 —
# cold install under go1.26.7, then `mise exec go@1.27.0 -- make
# install-golangci-lint` skipped the reinstall and left a go1.26.7 binary).
# install-gosec below already reads `go version -m`; this makes golangci-lint
# match its sibling rather than introduce a second idiom.
install-golangci-lint: ## install the pinned golangci-lint v2 release
	@echo -e "$(CYAN)Ensuring golangci-lint $(GOLANGCI_LINT_VERSION)...$(RESET)"
	@mkdir -p "$(GOLANGCI_LINT_DIR)"
	@if [ ! -x "$(GOLANGCI_LINT)" ] || ! go version -m "$(GOLANGCI_LINT)" 2>/dev/null | \
		awk -v want="$$(go env GOVERSION)" 'NR == 1 { built = $$NF } $$1 == "mod" && $$2 == "$(GOLANGCI_LINT_MODULE)" && $$3 == "$(GOLANGCI_LINT_VERSION)" { found = 1 } END { exit !(found && built == want) }'; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) to $(GOLANGCI_LINT)..."; \
		GOBIN="$(GOLANGCI_LINT_DIR)" go install $(GOLANGCI_LINT_INSTALL); \
	fi
	@go version -m "$(GOLANGCI_LINT)" 2>/dev/null | \
		awk -v want="$$(go env GOVERSION)" 'NR == 1 { built = $$NF } $$1 == "mod" && $$2 == "$(GOLANGCI_LINT_MODULE)" && $$3 == "$(GOLANGCI_LINT_VERSION)" { found = 1 } END { exit !(found && built == want) }' || { \
		echo "golangci-lint installation did not produce $(GOLANGCI_LINT_MODULE) $(GOLANGCI_LINT_VERSION) built with $$(go env GOVERSION): $(GOLANGCI_LINT)" >&2; \
		exit 1; \
	}

install-gosec: ## install the pinned gosec release
	@echo -e "$(CYAN)Ensuring gosec $(GOSEC_VERSION)...$(RESET)"
	@mkdir -p "$(GOSEC_DIR)"
	@if [ ! -x "$(GOSEC)" ] || ! go version -m "$(GOSEC)" 2>/dev/null | \
		awk '$$1 == "mod" && $$2 == "$(GOSEC_MODULE)" && $$3 == "$(GOSEC_VERSION)" { found = 1 } END { exit !found }'; then \
		echo "Installing gosec $(GOSEC_VERSION) to $(GOSEC)..."; \
		GOBIN="$(GOSEC_DIR)" go install $(GOSEC_INSTALL); \
	fi
	@go version -m "$(GOSEC)" 2>/dev/null | \
		awk '$$1 == "mod" && $$2 == "$(GOSEC_MODULE)" && $$3 == "$(GOSEC_VERSION)" { found = 1 } END { exit !found }' || { \
		echo "gosec installation did not produce $(GOSEC_MODULE) $(GOSEC_VERSION): $(GOSEC)" >&2; \
		exit 1; \
	}

install-goreleaser: ## install the pinned GoReleaser v2 release
	@echo -e "$(CYAN)Ensuring goreleaser $(GORELEASER_VERSION)...$(RESET)"
	@mkdir -p "$(RELEASE_TOOLS_DIR)"
	@if [ ! -x "$(GORELEASER)" ] || ! $(GORELEASER_VERSION_OK); then \
		echo "Installing goreleaser $(GORELEASER_VERSION) to $(GORELEASER)..."; \
		GOBIN="$(RELEASE_TOOLS_DIR)" go install $(GORELEASER_INSTALL); \
	fi
	@$(GORELEASER_VERSION_OK) || { \
		echo "goreleaser installation did not produce $(GORELEASER_MODULE) $(GORELEASER_VERSION): $(GORELEASER)" >&2; \
		exit 1; \
	}

install-syft: ## install the pinned syft release (SBOM generation)
	@echo -e "$(CYAN)Ensuring syft $(SYFT_VERSION)...$(RESET)"
	@mkdir -p "$(RELEASE_TOOLS_DIR)"
	@if [ ! -x "$(SYFT)" ] || ! $(SYFT_VERSION_OK); then \
		echo "Installing syft $(SYFT_VERSION) to $(SYFT)..."; \
		GOBIN="$(RELEASE_TOOLS_DIR)" go install $(SYFT_INSTALL); \
	fi
	@$(SYFT_VERSION_OK) || { \
		echo "syft installation did not produce $(SYFT_MODULE) $(SYFT_VERSION): $(SYFT)" >&2; \
		exit 1; \
	}

install-cosign: ## install the pinned cosign release (artifact signing)
	@echo -e "$(CYAN)Ensuring cosign $(COSIGN_VERSION)...$(RESET)"
	@mkdir -p "$(RELEASE_TOOLS_DIR)"
	@if [ ! -x "$(COSIGN)" ] || ! $(COSIGN_VERSION_OK); then \
		echo "Installing cosign $(COSIGN_VERSION) to $(COSIGN)..."; \
		GOBIN="$(RELEASE_TOOLS_DIR)" go install $(COSIGN_INSTALL); \
	fi
	@$(COSIGN_VERSION_OK) || { \
		echo "cosign installation did not produce $(COSIGN_MODULE) $(COSIGN_VERSION): $(COSIGN)" >&2; \
		exit 1; \
	}

install-release-tools: install-goreleaser install-syft install-cosign ## install the exact pinned release toolchain
	@echo -e "$(GREEN)✅ Release toolchain installed$(RESET)"

verify-release-tools: ## fail fast unless the exact pinned release toolchain is present
	@echo -e "$(CYAN)Verifying pinned release toolchain...$(RESET)"
	@failed=0; \
	if $(GORELEASER_VERSION_OK); then echo "  ✓ goreleaser $(GORELEASER_VERSION)"; \
	elif [ ! -x "$(GORELEASER)" ]; then \
		echo "release toolchain: goreleaser is missing at $(GORELEASER) (want $(GORELEASER_VERSION)); run: make install-goreleaser" >&2; \
		failed=1; \
	else \
		echo "release toolchain: goreleaser at $(GORELEASER) is '$(GORELEASER_FOUND)' but $(GORELEASER_VERSION) is pinned; run: make install-goreleaser" >&2; \
		failed=1; \
	fi; \
	if $(SYFT_VERSION_OK); then echo "  ✓ syft $(SYFT_VERSION)"; \
	elif [ ! -x "$(SYFT)" ]; then \
		echo "release toolchain: syft is missing at $(SYFT) (want $(SYFT_VERSION)); run: make install-syft" >&2; \
		failed=1; \
	else \
		echo "release toolchain: syft at $(SYFT) is '$(SYFT_FOUND)' but $(SYFT_VERSION) is pinned; run: make install-syft" >&2; \
		failed=1; \
	fi; \
	if $(COSIGN_VERSION_OK); then echo "  ✓ cosign $(COSIGN_VERSION)"; \
	elif [ ! -x "$(COSIGN)" ]; then \
		echo "release toolchain: cosign is missing at $(COSIGN) (want $(COSIGN_VERSION)); run: make install-cosign" >&2; \
		failed=1; \
	else \
		echo "release toolchain: cosign at $(COSIGN) is '$(COSIGN_FOUND)' but $(COSIGN_VERSION) is pinned; run: make install-cosign" >&2; \
		failed=1; \
	fi; \
	if [ "$$failed" -ne 0 ]; then \
		echo -e "$(RED)Release toolchain verification failed: refusing to run a release step that would silently skip SBOM or signing.$(RESET)" >&2; \
		exit 1; \
	fi
	@echo -e "$(GREEN)✅ Release toolchain matches the pinned versions$(RESET)"

verify-release-pins: ## fail unless the release workflow pins the same versions as this file
	@echo -e "$(CYAN)Verifying release pin parity with the workflow...$(RESET)"
	@[ -f "$(RELEASE_WORKFLOW)" ] || { \
		echo "release pins: workflow not found at $(RELEASE_WORKFLOW)" >&2; \
		exit 1; \
	}
	@failed=0; \
	check_pin() { \
		got=$$(awk -v k="$$1:" '$$1 == k { print $$2; exit }' "$(RELEASE_WORKFLOW)"); \
		if [ -z "$$got" ]; then \
			echo "release pins: $$1 is not pinned in $(RELEASE_WORKFLOW) (.make/tools.mk pins $$2)" >&2; \
			return 1; \
		fi; \
		if [ "$$got" != "$$2" ]; then \
			echo "release pins: $$1 is $$got in the workflow but $$2 in .make/tools.mk; CI would not release what make validated" >&2; \
			return 1; \
		fi; \
		echo "  ✓ $$1 $$2"; \
	}; \
	check_pin GORELEASER_VERSION "$(GORELEASER_VERSION)" || failed=1; \
	check_pin SYFT_VERSION "$(SYFT_VERSION)" || failed=1; \
	check_pin COSIGN_VERSION "$(COSIGN_VERSION)" || failed=1; \
	if [ "$$failed" -ne 0 ]; then \
		echo -e "$(RED)Release pin verification failed: .make/tools.mk and .github/workflows/release.yml must pin the same versions.$(RESET)" >&2; \
		exit 1; \
	fi
	@echo -e "$(GREEN)✅ Release pins match the workflow$(RESET)"

# ==============================================================================
# Mock and Generation Tools
# ==============================================================================

.PHONY: install-mock-tools generate-mocks clean-mocks regenerate-mocks

install-mock-tools: install-mockgen ## install mock generation tools
	@echo -e "$(GREEN)✅ Mock generation tools installed!$(RESET)"

# This target is now the only way to regenerate mocks: the three sources that
# used to carry their own `//go:generate mockgen ...` directive
# (internal/analysis/quality_analyzer.go, pkg/config/interfaces.go,
# pkg/synclone/facade.go) have those directives folded in below, so `go
# generate` no longer resolves a second, unpinned mockgen from PATH.
generate-mocks: install-mock-tools ## generate all mock files using gomock (the pinned path; see note above)
	@echo -e "$(CYAN)Generating mocks...$(RESET)"
	@echo "Generating GitHub interface mocks..."
	@if [ -f "pkg/github/interfaces.go" ]; then \
		"$(MOCKGEN)" -source=pkg/github/interfaces.go -destination=pkg/github/mocks/github_mocks.go -package=mocks; \
		echo "  ✅ GitHub mocks generated"; \
	else \
		echo "  ⚠️  pkg/github/interfaces.go not found"; \
	fi
	@echo "Generating filesystem interface mocks..."
	@if [ -f "internal/filesystem/interfaces.go" ]; then \
		"$(MOCKGEN)" -source=internal/filesystem/interfaces.go -destination=internal/filesystem/mocks/filesystem_mocks.go -package=mocks; \
		echo "  ✅ Filesystem mocks generated"; \
	else \
		echo "  ⚠️  internal/filesystem/interfaces.go not found"; \
	fi
	@echo "Generating HTTP client interface mocks..."
	@if [ -f "internal/httpclient/interfaces.go" ]; then \
		"$(MOCKGEN)" -source=internal/httpclient/interfaces.go -destination=internal/httpclient/mocks/httpclient_mocks.go -package=mocks; \
		echo "  ✅ HTTP client mocks generated"; \
	else \
		echo "  ⚠️  internal/httpclient/interfaces.go not found"; \
	fi
	@echo "Generating Git interface mocks..."
	@if [ -f "internal/git/interfaces.go" ]; then \
		"$(MOCKGEN)" -source=internal/git/interfaces.go -destination=internal/git/mocks/git_mocks.go -package=mocks; \
		echo "  ✅ Git mocks generated"; \
	else \
		echo "  ⚠️  internal/git/interfaces.go not found"; \
	fi
	@echo "Generating quality analyzer interface mocks..."
	@if [ -f "internal/analysis/quality_analyzer.go" ]; then \
		"$(MOCKGEN)" -source=internal/analysis/quality_analyzer.go -destination=internal/analysis/mocks/quality_analyzer_mock.go -package=mocks; \
		echo "  ✅ Quality analyzer mocks generated"; \
	else \
		echo "  ⚠️  internal/analysis/quality_analyzer.go not found"; \
	fi
	@echo "Generating config interface mocks..."
	@if [ -f "pkg/config/interfaces.go" ]; then \
		"$(MOCKGEN)" -source=pkg/config/interfaces.go -destination=pkg/config/mocks/config_mocks.go -package=mocks; \
		echo "  ✅ Config mocks generated"; \
	else \
		echo "  ⚠️  pkg/config/interfaces.go not found"; \
	fi
	@echo "Generating synclone facade interface mocks..."
	@if [ -f "pkg/synclone/facade.go" ]; then \
		"$(MOCKGEN)" -source=pkg/synclone/facade.go -destination=pkg/synclone/mocks/bulk_clone_manager_mock.go -package=mocks; \
		echo "  ✅ Synclone facade mocks generated"; \
	else \
		echo "  ⚠️  pkg/synclone/facade.go not found"; \
	fi
	@echo -e "$(GREEN)✅ Mock generation complete!$(RESET)"

clean-mocks: ## remove all generated mock files
	@echo -e "$(CYAN)Cleaning generated mocks...$(RESET)"
	@rm -f pkg/github/mocks/github_mocks.go
	@rm -f internal/filesystem/mocks/filesystem_mocks.go
	@rm -f internal/httpclient/mocks/httpclient_mocks.go
	@rm -f internal/git/mocks/git_mocks.go
	@echo -e "$(GREEN)✅ Mock cleanup complete!$(RESET)"

regenerate-mocks: clean-mocks generate-mocks ## clean and regenerate all mocks

# ==============================================================================
# Pre-commit and Git Hooks
# ==============================================================================

.PHONY: install-pre-commit-tools pre-commit-update

install-pre-commit-tools: ## install pre-commit and related tools
	@echo -e "$(CYAN)Installing pre-commit tools...$(RESET)"
	@command -v pre-commit >/dev/null 2>&1 || { echo -e "$(RED)pre-commit not found. Install with: pip install pre-commit$(RESET)"; }
	@echo -e "$(GREEN)✅ Pre-commit tools ready!$(RESET)"


# ==============================================================================
# Documentation Tools
# ==============================================================================

.PHONY: install-docs-tools

# benchstat and godoc used to be installed here unconditionally on every
# invocation -- no guard at all, so this target reinstalled two tools from
# `@latest` every time anything depended on it. They are now ordinary pinned
# prerequisites, which also means they are no-ops once present.
install-docs-tools: install-benchstat install-godoc ## install documentation tools
	@echo -e "$(CYAN)Installing documentation tools...$(RESET)"
	@which git-chglog >/dev/null 2>&1 || echo -e "$(YELLOW)Consider installing git-chglog for changelog generation$(RESET)"
	@which mkdocs >/dev/null 2>&1 || echo -e "$(YELLOW)Consider installing mkdocs for documentation: pip install mkdocs mkdocs-material$(RESET)"
	@echo -e "$(GREEN)✅ Documentation tools installed$(RESET)"

# ==============================================================================
# Security Tools
# ==============================================================================

.PHONY: install-security-tools

install-security-tools: install-gosec ## install security analysis tools
	@echo -e "$(CYAN)Installing security tools...$(RESET)"
	@echo -e "$(GREEN)✅ Security tools installed!$(RESET)"

# ==============================================================================
# Vulnerability Scanning
# ==============================================================================

.PHONY: install-vuln-tools

install-vuln-tools: ## install vulnerability scanning tools
	@echo -e "$(CYAN)Installing vulnerability scanning tools...$(RESET)"
	@echo "govulncheck is available as: go run golang.org/x/vuln/cmd/govulncheck@latest"
	@echo -e "$(GREEN)✅ Vulnerability tools ready!$(RESET)"

# ==============================================================================
# Tool Status and Information
# ==============================================================================

.PHONY: tools-status tools-info

tools-status: ## show installed tool status
	@echo -e "$(CYAN)Checking development tool status...$(RESET)"
	@echo ""
	@echo -e "$(GREEN)📦 Core Tools:$(RESET)"
	@printf "  %-20s " "go:"; go version 2>/dev/null | cut -d' ' -f3 || echo -e "$(RED)Not found$(RESET)"
	@printf "  %-20s " "git:"; git --version 2>/dev/null | cut -d' ' -f3 || echo -e "$(RED)Not found$(RESET)"
	@echo ""
	@echo -e "$(GREEN)🔧 Release Toolchain (pinned):$(RESET)"
	@printf "  %-20s " "goreleaser:"; VERSION=$$(go version -m "$(GORELEASER)" 2>/dev/null | awk '$$1 == "mod" && $$2 == "$(GORELEASER_MODULE)" { print $$3 }'); \
		if [ -n "$$VERSION" ]; then echo "$$VERSION (want $(GORELEASER_VERSION))"; else echo -e "$(RED)Not installed$(RESET) (want $(GORELEASER_VERSION))"; fi
	@printf "  %-20s " "syft:"; VERSION=$$(go version -m "$(SYFT)" 2>/dev/null | awk '$$1 == "mod" && $$2 == "$(SYFT_MODULE)" { print $$3 }'); \
		if [ -n "$$VERSION" ]; then echo "$$VERSION (want $(SYFT_VERSION))"; else echo -e "$(RED)Not installed$(RESET) (want $(SYFT_VERSION))"; fi
	@printf "  %-20s " "cosign:"; VERSION=$$(go version -m "$(COSIGN)" 2>/dev/null | awk '$$1 == "mod" && $$2 == "$(COSIGN_MODULE)" { print $$3 }'); \
		if [ -n "$$VERSION" ]; then echo "$$VERSION (want $(COSIGN_VERSION))"; else echo -e "$(RED)Not installed$(RESET) (want $(COSIGN_VERSION))"; fi
	@echo ""
	@echo -e "$(GREEN)✨ Format Tools:$(RESET)"
	@printf "  %-20s " "gofumpt:"; "$(GOFUMPT)" --version 2>/dev/null || echo -e "$(RED)Not installed$(RESET)"
	@printf "  %-20s " "gci:"; "$(GCI)" --version 2>/dev/null || echo -e "$(RED)Not installed$(RESET)"
	@echo ""
	@echo -e "$(GREEN)🔍 Lint Tools:$(RESET)"
	@printf "  %-20s " "golangci-lint:"; "$(GOLANGCI_LINT)" version --short 2>/dev/null || echo -e "$(RED)Not installed$(RESET)"
	@printf "  %-20s " "staticcheck:"; "$(STATICCHECK)" -version 2>/dev/null || echo -e "$(RED)Not installed$(RESET) (want $(STATICCHECK_VERSION))"
	@echo ""
	@echo -e "$(GREEN)🛡️  Security Tools:$(RESET)"
	@printf "  %-20s " "gosec:"; VERSION=$$(go version -m "$(GOSEC)" 2>/dev/null | awk '$$1 == "mod" && $$2 == "$(GOSEC_MODULE)" { print $$3 }'); \
		if [ -n "$$VERSION" ]; then echo "$$VERSION"; else echo -e "$(RED)Not installed$(RESET)"; fi
	@echo ""
	@echo -e "$(GREEN)🎭 Mock Tools:$(RESET)"
	@printf "  %-20s " "mockgen:"; "$(MOCKGEN)" --version 2>/dev/null || echo -e "$(RED)Not installed$(RESET) (want $(MOCKGEN_VERSION))"
	@echo ""
	@echo -e "$(GREEN)🎣 Git Hooks:$(RESET)"
	@printf "  %-20s " "pre-commit:"; pre-commit --version 2>/dev/null || echo -e "$(RED)Not installed$(RESET)"

tools-info: ## show comprehensive tool information
	@echo -e "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo -e "║                         $(YELLOW)Development Tools Information$(CYAN)                   ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo -e "$(RESET)"
	@echo -e "$(GREEN)🔧 Available Tool Categories:$(RESET)"
	@echo -e "  • $(CYAN)Format Tools$(RESET)        Code formatting (gofumpt, gci)"
	@echo -e "  • $(CYAN)Analysis Tools$(RESET)      Static analysis (staticcheck, gosec)"
	@echo -e "  • $(CYAN)Lint Tools$(RESET)          Code linting (golangci-lint)"
	@echo -e "  • $(CYAN)Release Tools$(RESET)       Pinned release toolchain (goreleaser, syft, cosign)"
	@echo -e "  • $(CYAN)Mock Tools$(RESET)          Mock generation (mockgen)"
	@echo -e "  • $(CYAN)Security Tools$(RESET)      Security scanning (gosec, govulncheck)"
	@echo -e "  • $(CYAN)Git Hooks$(RESET)           Pre-commit hooks and validation"
	@echo -e "  • $(CYAN)Documentation$(RESET)       Documentation tools (godoc, benchstat)"
	@echo ""
	@echo -e "$(GREEN)🚀 Quick Installation:$(RESET)"
	@echo -e "  $(CYAN)make install-tools$(RESET)        Install all development tools"
	@echo -e "  $(CYAN)make tools-status$(RESET)         Check current tool installation status"
	@echo ""
	@echo -e "$(GREEN)💡 Individual Categories:$(RESET)"
	@echo -e "  $(CYAN)make install-format-tools$(RESET)     Format tools only"
	@echo -e "  $(CYAN)make install-analysis-tools$(RESET)   Analysis tools only"
	@echo -e "  $(CYAN)make install-security-tools$(RESET)   Security tools only"
	@echo -e "  $(CYAN)make install-release-tools$(RESET)    Pinned release toolchain only"
	@echo -e "  $(CYAN)make verify-release-tools$(RESET)     Fail fast on a drifted release toolchain"
	@echo -e "  $(CYAN)make verify-release-pins$(RESET)      Fail fast when local and CI pins disagree"
	@echo -e "  $(CYAN)make install-mock-tools$(RESET)       Mock generation tools only"
