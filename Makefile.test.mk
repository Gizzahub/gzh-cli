# Makefile.test - Testing targets for gzh-manager-go
# Unit tests, integration tests, benchmarks, and coverage

# ==============================================================================
# Testing Configuration
# ==============================================================================

# Colors for output
CYAN := \\033[36m
GREEN := \\033[32m
YELLOW := \\033[33m
RED := \\033[31m
BLUE := \\033[34m
RESET := \\033[0m

# ==============================================================================
# Testing Targets
# ==============================================================================

.PHONY: test test-unit test-integration test-integration-only test-e2e test-e2e-only test-all
.PHONY: cover cover-html cover-report bench test-coverage test-docker

test: clean ## run all tests with coverage
	@echo "$(CYAN)Running all tests with coverage...$(RESET)"
	go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | sort -rnk3
	@echo "$(GREEN)✅ Tests completed$(RESET)"

test-unit: ## run only unit tests (exclude integration and e2e)
	@echo "$(CYAN)Running unit tests...$(RESET)"
	go test -short --cover -parallel=1 -v -coverprofile=coverage-unit.out \
		$$(go list ./... | grep -v -E '(test/integration|test/e2e)')
	go tool cover -func=coverage-unit.out | sort -rnk3
	@echo "$(GREEN)✅ Unit tests completed$(RESET)"

test-integration-only: ## run only integration tests with build tag
	@echo "$(CYAN)Running integration tests...$(RESET)"
	go test -tags=integration -v ./test/integration/...
	@echo "$(GREEN)✅ Integration tests completed$(RESET)"

test-e2e-only: ## run only e2e tests with build tag
	@echo "$(CYAN)Running E2E tests...$(RESET)"
	go test -tags=e2e -v ./test/e2e/...
	@echo "$(GREEN)✅ E2E tests completed$(RESET)"

test-integration: ## run Docker-based integration tests (alias for test-docker)
	@echo "$(CYAN)Running Docker integration tests...$(RESET)"
	@if [ -f "./test/integration/run_docker_tests.sh" ]; then \
		./test/integration/run_docker_tests.sh all; \
	else \
		echo "$(YELLOW)No Docker integration test script found$(RESET)"; \
		make test-integration-only; \
	fi
	@echo "$(GREEN)✅ Integration tests completed$(RESET)"

test-e2e: build ## run End-to-End test scenarios
	@echo "$(CYAN)Running E2E tests...$(RESET)"
	@if [ -f "./test/e2e/run_e2e_tests.sh" ]; then \
		./test/e2e/run_e2e_tests.sh all; \
	else \
		echo "$(YELLOW)No E2E test script found$(RESET)"; \
		make test-e2e-only; \
	fi
	@echo "$(GREEN)✅ E2E tests completed$(RESET)"

test-all: test test-integration test-e2e ## run all tests (unit, integration, e2e)
	@echo "$(GREEN)✅ All tests completed successfully!$(RESET)"

test-docker: test-integration ## alias for test-integration

# ==============================================================================
# Coverage Targets
# ==============================================================================

cover: ## display test coverage
	@echo "$(CYAN)Generating test coverage report...$(RESET)"
	go test -v -race $$(go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@echo "$(GREEN)✅ Coverage report generated$(RESET)"

cover-html: ## generate HTML coverage report
	@echo "$(CYAN)Generating HTML coverage report...$(RESET)"
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(RESET)"

cover-report: ## generate detailed coverage report
	@echo "$(CYAN)Generating detailed coverage report...$(RESET)"
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "$(YELLOW)=== Coverage Summary ===$(RESET)"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
	@echo ""
	@echo "$(YELLOW)=== Package Coverage ===$(RESET)"
	@go tool cover -func=coverage.out | grep -v total | sort -k3 -nr | head -20
	@echo ""
	@echo "$(BLUE)For detailed HTML report, run: make cover-html$(RESET)"

test-coverage: cover-report ## alias for cover-report

# ==============================================================================
# Benchmark Targets
# ==============================================================================

bench: ## run all benchmarks
	@echo "$(CYAN)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem ./...
	@echo "$(GREEN)✅ Benchmarks completed$(RESET)"

bench-cpu: ## run CPU benchmarks with profiling
	@echo "$(CYAN)Running CPU benchmarks with profiling...$(RESET)"
	@go test -bench=. -benchmem -cpuprofile=cpu.prof ./...
	@echo "$(GREEN)✅ CPU benchmarks completed$(RESET)"
	@echo "$(YELLOW)Use 'go tool pprof cpu.prof' to analyze$(RESET)"

bench-mem: ## run memory benchmarks with profiling
	@echo "$(CYAN)Running memory benchmarks with profiling...$(RESET)"
	@go test -bench=. -benchmem -memprofile=mem.prof ./...
	@echo "$(GREEN)✅ Memory benchmarks completed$(RESET)"
	@echo "$(YELLOW)Use 'go tool pprof mem.prof' to analyze$(RESET)"

bench-compare: ## compare benchmarks (requires benchstat)
	@echo "$(CYAN)Comparing benchmarks...$(RESET)"
	@command -v benchstat >/dev/null 2>&1 || { echo "Installing benchstat..." && go install golang.org/x/perf/cmd/benchstat@latest; }
	@go test -bench=. -count=5 ./... > new.bench
	@echo "$(GREEN)✅ Benchmark comparison data generated: new.bench$(RESET)"
	@echo "$(YELLOW)Run 'benchstat old.bench new.bench' to compare$(RESET)"

# ==============================================================================
# Test Utilities
# ==============================================================================

test-race: ## run tests with race detection
	@echo "$(CYAN)Running tests with race detection...$(RESET)"
	@go test -race -short ./...
	@echo "$(GREEN)✅ Race detection tests completed$(RESET)"

test-verbose: ## run tests with verbose output
	@echo "$(CYAN)Running tests with verbose output...$(RESET)"
	@go test -v ./...
	@echo "$(GREEN)✅ Verbose tests completed$(RESET)"

test-timeout: ## run tests with custom timeout
	@echo "$(CYAN)Running tests with 30s timeout...$(RESET)"
	@go test -timeout=30s ./...
	@echo "$(GREEN)✅ Timeout tests completed$(RESET)"

test-list: ## list all available tests
	@echo "$(CYAN)Listing all available tests...$(RESET)"
	@go test -list . ./... | grep -E '^Test|^Benchmark'
	@echo "$(GREEN)✅ Test listing completed$(RESET)"

# ==============================================================================
# Test Information
# ==============================================================================

.PHONY: test-info

test-info: ## show testing information and available targets
	@echo "$(CYAN)"
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                         $(YELLOW)Testing Information$(CYAN)                             ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo "$(RESET)"
	@echo "$(GREEN)🧪 Test Categories:$(RESET)"
	@echo "  • $(CYAN)Unit Tests$(RESET)          Fast, isolated component tests"
	@echo "  • $(CYAN)Integration Tests$(RESET)   Docker-based service integration"
	@echo "  • $(CYAN)E2E Tests$(RESET)           End-to-end scenario testing"
	@echo ""
	@echo "$(GREEN)📊 Coverage Targets:$(RESET)"
	@echo "  • $(CYAN)cover$(RESET)               Display test coverage"
	@echo "  • $(CYAN)cover-html$(RESET)          Generate HTML coverage report"
	@echo "  • $(CYAN)cover-report$(RESET)        Detailed coverage analysis"
	@echo ""
	@echo "$(GREEN)⚡ Benchmark Targets:$(RESET)"
	@echo "  • $(CYAN)bench$(RESET)               Run all benchmarks"
	@echo "  • $(CYAN)bench-cpu$(RESET)           CPU benchmarks with profiling"
	@echo "  • $(CYAN)bench-mem$(RESET)           Memory benchmarks with profiling"
	@echo "  • $(CYAN)bench-compare$(RESET)       Compare benchmark results"
	@echo ""
	@echo "$(GREEN)🔧 Test Utilities:$(RESET)"
	@echo "  • $(CYAN)test-race$(RESET)           Run with race detection"
	@echo "  • $(CYAN)test-verbose$(RESET)        Run with verbose output"
	@echo "  • $(CYAN)test-timeout$(RESET)        Run with custom timeout"
	@echo "  • $(CYAN)test-list$(RESET)           List all available tests"
