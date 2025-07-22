projectname?=gzh-manager
executablename?=gz
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

# Include lint-related targets
include Makefile.lint.mk

# Include dependency management targets
include Makefile.deps.mk

default: help

.PHONY: help
help: ## list makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

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

.PHONY: comments
comments: ## show all TODO/FIXME/NOTE comments in codebase
	@echo "=== TODO comments ==="
	@grep -r "TODO" --include="*.go" . | grep -v vendor | grep -v .git || echo "No TODOs found!"
	@echo ""
	@echo "=== FIXME comments ==="
	@grep -r "FIXME" --include="*.go" . | grep -v vendor | grep -v .git || echo "No FIXMEs found!"
	@echo ""
	@echo "=== NOTE comments ==="
	@grep -r "NOTE" --include="*.go" . | grep -v vendor | grep -v .git || echo "No NOTEs found!"

# Aliases for backward compatibility
.PHONY: todo fixme notes
todo: comments ## show all TODO comments (alias for comments)
fixme: comments ## show all FIXME comments (alias for comments)
notes: comments ## show all NOTE comments (alias for comments)

## CI/CD Helpers

.PHONY: changelog
changelog: ## generate changelog (requires git-chglog)
	@command -v git-chglog >/dev/null 2>&1 || { echo "git-chglog not found. Install from: https://github.com/git-chglog/git-chglog"; exit 1; }
	@git-chglog -o CHANGELOG.md
