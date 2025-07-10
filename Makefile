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
build: build-frontend ## build golang binary
	@echo "Building $(executablename)..."
	@go build -ldflags "-X main.version=$(shell git describe --always --abbrev=0 --tags)" -o $(executablename)

.PHONY: build-frontend
build-frontend: ## build React frontend
	@echo "Building React frontend..."
	@cd web && npm ci && npm run build

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
	rm -rf web/build web/node_modules

.PHONY: dev-frontend
dev-frontend: ## start React development server
	@echo "Starting React development server..."
	@cd web && npm start

PHONY: cover
cover: ## display test coverage
	go test -v -race $(shell go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

PHONY: fmt
fmt: ## format go files
	gofumpt -w .
	gci write .

PHONY: lint
lint: ## lint go files
	golangci-lint run -c .golang-ci.yml

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

.PHONY: pre-commit
pre-commit:	## run pre-commit hooks
	pre-commit run --all-files

.PHONY: deploy
deploy:
	# TODO ...
	# $build and deploy
	cp * .$(executablename)
