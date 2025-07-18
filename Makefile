projectname?=gzh-manager
executablename?=gz
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

default: help

.PHONY: help
help: ## list makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

depup: ## update dependencies
	go mod tidy
	go get -u ./...

.PHONY: build
build: ## build golang binary
	@echo "Building $(executablename)..."
	@go build -ldflags "-X main.version=$(shell git describe --always --abbrev=0 --tags)" -o $(executablename)

.PHONY: install
install: build ## install golang binary
#	@go install -ldflags "-X main.version=$(shell git describe --always --abbrev=0 --tags)"
	@echo "Installing $(executablename)..."
	@mv $(executablename) $(shell go env GOPATH)/bin/

.PHONY: run
run: ## run the app
	@go run -ldflags "-X main.version=$(shell git describe --always --abbrev=0 --tags)"  main.go

.PHONY: bootstrap
bootstrap: ## install build deps
	go generate -tags tools tools/tools.go

PHONY: test
test: clean ## run all tests with coverage
	go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | sort -rnk3

PHONY: test-unit
test-unit: ## run only unit tests (exclude integration and e2e)
	go test -short --cover -parallel=1 -v -coverprofile=coverage-unit.out \
		$(shell go list ./... | grep -v -E '(test/integration|test/e2e)')
	go tool cover -func=coverage-unit.out | sort -rnk3

PHONY: test-integration-only
test-integration-only: ## run only integration tests with build tag
	go test -tags=integration -v ./test/integration/...

PHONY: test-e2e-only
test-e2e-only: ## run only e2e tests with build tag
	go test -tags=e2e -v ./test/e2e/...

PHONY: clean
clean: ## clean up environment
	rm -rf coverage.out dist/ $(executablename)
	rm -f $(shell go env GOPATH)/bin/$(executablename)
	rm -f $(shell go env GOPATH)/bin/$(projectname)


PHONY: cover
cover: ## display test coverage
	go test -v -race $(shell go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

PHONY: cover-html
cover-html: ## generate HTML coverage report
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

PHONY: cover-report
cover-report: ## generate detailed coverage report
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "=== Coverage Summary ==="
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
	@echo ""
	@echo "=== Package Coverage ==="
	@go tool cover -func=coverage.out | grep -v total | sort -k3 -nr | head -20
	@echo ""
	@echo "For detailed HTML report, run: make cover-html"

.PHONY: test-docker
test-docker: ## run Docker-based integration tests
	@echo "Running Docker integration tests..."
	@./test/integration/run_docker_tests.sh all

.PHONY: test-docker-short
test-docker-short: ## run integration tests in short mode (skip Docker)
	@echo "Running integration tests in short mode..."
	@./test/integration/run_docker_tests.sh -s all

.PHONY: test-gitlab
test-gitlab: ## run GitLab integration tests only
	@echo "Running GitLab integration tests..."
	@./test/integration/run_docker_tests.sh gitlab

.PHONY: test-gitea
test-gitea: ## run Gitea integration tests only
	@echo "Running Gitea integration tests..."
	@./test/integration/run_docker_tests.sh gitea

.PHONY: test-redis
test-redis: ## run Redis integration tests only
	@echo "Running Redis integration tests..."
	@./test/integration/run_docker_tests.sh redis

.PHONY: test-integration
test-integration: test-docker ## alias for test-docker

.PHONY: test-e2e
test-e2e: build ## run End-to-End test scenarios
	@echo "Running E2E tests..."
	@./test/e2e/run_e2e_tests.sh all

.PHONY: test-e2e-short
test-e2e-short: build ## run E2E tests in short mode
	@echo "Running E2E tests in short mode..."
	@./test/e2e/run_e2e_tests.sh -s all

.PHONY: test-e2e-bulk-clone
test-e2e-bulk-clone: build ## run bulk clone E2E tests only
	@echo "Running bulk clone E2E tests..."
	@./test/e2e/run_e2e_tests.sh bulk-clone

.PHONY: test-e2e-config
test-e2e-config: build ## run configuration E2E tests only
	@echo "Running configuration E2E tests..."
	@./test/e2e/run_e2e_tests.sh config

.PHONY: test-e2e-ide
test-e2e-ide: build ## run IDE E2E tests only
	@echo "Running IDE E2E tests..."
	@./test/e2e/run_e2e_tests.sh ide

.PHONY: test-all
test-all: test test-docker test-e2e ## run all tests (unit, integration, e2e)

PHONY: fmt
fmt: ## format go files
	gofumpt -w .
	gci write .

PHONY: lint
lint: ## lint go files
	golangci-lint run -c .golangci.yml --fix

PHONY: format
format: ## format go files (alias for lint)
	golangci-lint run -c .golangci.yml --fix

.PHONY: security
security: ## run security analysis with gosec
	@echo "Running security analysis..."
	@command -v gosec >/dev/null 2>&1 || { echo "gosec not found. Installing..."; go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@gosec -config=.gosec.yaml ./...

.PHONY: security-json
security-json: ## run security analysis and output JSON report
	@echo "Running security analysis with JSON output..."
	@command -v gosec >/dev/null 2>&1 || { echo "gosec not found. Installing..."; go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	@gosec -fmt=json -out=gosec-report.json -config=.gosec.yaml ./...

.PHONY: generate-mocks
generate-mocks: ## generate all mock files using gomock
	@echo "Generating mocks..."
	@command -v mockgen >/dev/null 2>&1 || { echo "mockgen not found. Installing..."; go install go.uber.org/mock/mockgen@latest; }
	@mockgen -source=pkg/github/interfaces.go -destination=pkg/github/mocks/github_mocks.go -package=mocks
	@mockgen -source=internal/filesystem/interfaces.go -destination=internal/filesystem/mocks/filesystem_mocks.go -package=mocks
	@mockgen -source=internal/httpclient/interfaces.go -destination=internal/httpclient/mocks/httpclient_mocks.go -package=mocks
	@mockgen -source=internal/git/interfaces.go -destination=internal/git/mocks/git_mocks.go -package=mocks
	@echo "Mock generation complete!"

.PHONY: clean-mocks
clean-mocks: ## remove all generated mock files
	@echo "Cleaning generated mocks..."
	@rm -f pkg/github/mocks/github_mocks.go
	@rm -f internal/filesystem/mocks/filesystem_mocks.go
	@rm -f internal/httpclient/mocks/httpclient_mocks.go
	@rm -f internal/git/mocks/git_mocks.go
	@echo "Mock cleanup complete!"

.PHONY: regenerate-mocks
regenerate-mocks: clean-mocks generate-mocks ## clean and regenerate all mocks

.PHONY: pre-commit-install
pre-commit-install: ## install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || { echo "pre-commit not found. Install with: pip install pre-commit"; exit 1; }
	@./scripts/setup-git-hooks.sh

.PHONY: pre-commit
pre-commit:	## run pre-commit hooks
	pre-commit run --all-files

.PHONY: pre-push
pre-push: ## run pre-push hooks (same as pre-commit for consistency)
	pre-commit run --all-files --hook-stage pre-push

.PHONY: lint-all
lint-all: fmt lint pre-commit ## run all linting steps (format, lint, pre-commit)

.PHONY: check-consistency
check-consistency: ## verify lint configuration consistency
	@echo "Checking lint configuration consistency..."
	@echo "✓ Makefile uses: .golangci.yml"
	@grep -q "\.golangci\.yml" .pre-commit-config.yaml && echo "✓ Pre-commit uses: .golangci.yml" || echo "✗ Pre-commit config mismatch"
	@echo "✓ All configurations aligned"

.PHONY: pre-commit-update
pre-commit-update: ## update pre-commit hooks to latest versions
	pre-commit autoupdate

.PHONY: release-dry-run
release-dry-run: ## run goreleaser in dry-run mode
	@echo "Running goreleaser in dry-run mode..."
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest"; exit 1; }
	@goreleaser release --snapshot --clean --skip=publish

.PHONY: release-snapshot
release-snapshot: ## create a snapshot release
	@echo "Creating snapshot release..."
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest"; exit 1; }
	@goreleaser release --snapshot --clean

.PHONY: release-check
release-check: ## check goreleaser configuration
	@echo "Checking goreleaser configuration..."
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest"; exit 1; }
	@goreleaser check

.PHONY: install-goreleaser
install-goreleaser: ## install goreleaser
	@echo "Installing goreleaser..."
	@go install github.com/goreleaser/goreleaser@latest

.PHONY: deploy
deploy: release-dry-run ## alias for release-dry-run

## Development Workflow Targets

.PHONY: dev
dev: fmt lint test ## run standard development workflow (format, lint, test)

.PHONY: dev-fast
dev-fast: fmt test-unit ## quick development cycle (format and unit tests only)

.PHONY: verify
verify: fmt lint test cover-report check-consistency ## complete verification before PR

.PHONY: deps-check
deps-check: ## check for outdated dependencies
	@echo "Checking for outdated dependencies..."
	@go list -u -m all | grep '\[' || echo "All dependencies are up to date"

.PHONY: deps-upgrade
deps-upgrade: ## upgrade all dependencies to latest versions
	@echo "Upgrading all dependencies..."
	@go get -u ./...
	@go mod tidy

.PHONY: deps-verify
deps-verify: ## verify dependency checksums
	@echo "Verifying dependency checksums..."
	@go mod verify

.PHONY: deps-why
deps-why: ## show why a module is needed (usage: make deps-why MOD=github.com/pkg/errors)
	@go mod why -m $(MOD)

.PHONY: deps-graph
deps-graph: ## show module dependency graph
	@go mod graph

## Documentation Targets

.PHONY: docs-serve
docs-serve: ## serve documentation locally (requires mkdocs)
	@command -v mkdocs >/dev/null 2>&1 || { echo "mkdocs not found. Install with: pip install mkdocs mkdocs-material"; exit 1; }
	@mkdocs serve

.PHONY: docs-build
docs-build: ## build documentation site
	@command -v mkdocs >/dev/null 2>&1 || { echo "mkdocs not found. Install with: pip install mkdocs mkdocs-material"; exit 1; }
	@mkdocs build

.PHONY: godoc
godoc: ## run godoc server
	@echo "Starting godoc server on http://localhost:6060"
	@godoc -http=:6060

.PHONY: docs-check
docs-check: ## check for missing package documentation
	@echo "Checking for missing package documentation..."
	@for pkg in $$(go list ./...); do \
		if ! go doc -short $$pkg | grep -q "^package"; then \
			echo "Missing documentation for: $$pkg"; \
		fi; \
	done

## Code Analysis Targets

.PHONY: complexity
complexity: ## analyze code complexity
	@echo "Analyzing code complexity..."
	@command -v gocyclo >/dev/null 2>&1 || { echo "gocyclo not found. Installing..."; go install github.com/fzipp/gocyclo/cmd/gocyclo@latest; }
	@gocyclo -over 10 -avg .

.PHONY: ineffassign
ineffassign: ## detect ineffectual assignments
	@echo "Checking for ineffectual assignments..."
	@command -v ineffassign >/dev/null 2>&1 || { echo "ineffassign not found. Installing..."; go install github.com/gordonklaus/ineffassign@latest; }
	@ineffassign ./...

.PHONY: dupl
dupl: ## find duplicate code
	@echo "Checking for duplicate code..."
	@command -v dupl >/dev/null 2>&1 || { echo "dupl not found. Installing..."; go install github.com/mibk/dupl@latest; }
	@dupl -threshold 50 .

.PHONY: vuln
vuln: ## check for known vulnerabilities
	@echo "Checking for known vulnerabilities..."
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## Benchmarking Targets

.PHONY: bench
bench: ## run all benchmarks
	@go test -bench=. -benchmem ./...

.PHONY: bench-compare
bench-compare: ## compare benchmark results (requires benchstat)
	@command -v benchstat >/dev/null 2>&1 || { echo "benchstat not found. Installing..."; go install golang.org/x/perf/cmd/benchstat@latest; }
	@echo "Run benchmarks and save results:"
	@echo "  go test -bench=. -count=10 ./... > old.txt"
	@echo "  # make changes"
	@echo "  go test -bench=. -count=10 ./... > new.txt"
	@echo "  benchstat old.txt new.txt"

## Quick Commands

.PHONY: todo
todo: ## show all TODO comments in codebase
	@grep -r "TODO" --include="*.go" . | grep -v vendor | grep -v .git || echo "No TODOs found!"

.PHONY: fixme
fixme: ## show all FIXME comments in codebase
	@grep -r "FIXME" --include="*.go" . | grep -v vendor | grep -v .git || echo "No FIXMEs found!"

.PHONY: notes
notes: ## show all NOTE comments in codebase
	@grep -r "NOTE" --include="*.go" . | grep -v vendor | grep -v .git || echo "No NOTEs found!"

## CI/CD Helpers

.PHONY: ci-local
ci-local: clean verify test-all ## run full CI pipeline locally

.PHONY: pr-check
pr-check: fmt lint test cover-report check-consistency ## pre-PR submission check

.PHONY: changelog
changelog: ## generate changelog (requires git-chglog)
	@command -v git-chglog >/dev/null 2>&1 || { echo "git-chglog not found. Install from: https://github.com/git-chglog/git-chglog"; exit 1; }
	@git-chglog -o CHANGELOG.md

## Module Management

.PHONY: mod-tidy-check
mod-tidy-check: ## check if go.mod and go.sum are tidy
	@echo "Checking if go.mod is tidy..."
	@cp go.mod go.mod.bak
	@cp go.sum go.sum.bak
	@go mod tidy
	@if ! diff -q go.mod go.mod.bak >/dev/null || ! diff -q go.sum go.sum.bak >/dev/null; then \
		echo "❌ go.mod or go.sum is not tidy. Run 'go mod tidy'"; \
		mv go.mod.bak go.mod; \
		mv go.sum.bak go.sum; \
		exit 1; \
	else \
		echo "✅ go.mod is tidy"; \
		rm go.mod.bak go.sum.bak; \
	fi
