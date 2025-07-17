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
test: clean ## display test coverage
	go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | sort -rnk3

PHONY: clean
clean: ## clean up environment
	rm -rf coverage.out dist/ $(executablename)
	rm -f $(shell go env GOPATH)/bin/$(executablename)
	rm -f $(shell go env GOPATH)/bin/$(projectname)


PHONY: cover
cover: ## display test coverage
	go test -v -race $(shell go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

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
	golangci-lint run -c .golangci.yml

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
	@pre-commit install --install-hooks
	@pre-commit install --hook-type commit-msg
	@pre-commit install --hook-type pre-push
	@echo "Pre-commit hooks installed successfully!"

.PHONY: pre-commit
pre-commit:	## run pre-commit hooks
	pre-commit run --all-files

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
