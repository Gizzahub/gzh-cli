# Test Coverage Configuration for Git Repo Commands
# Include this in the main Makefile for coverage targets

.PHONY: test-coverage test-coverage-html test-integration-coverage coverage-report

# Generate test coverage for unit tests
test-coverage:
	@echo "🧪 Running unit tests with coverage..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./cmd/git/...
	go tool cover -func=coverage.out

# Generate HTML coverage report
test-coverage-html: test-coverage
	@echo "📊 Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests with coverage (requires tokens)
test-integration-coverage:
	@echo "🔗 Running integration tests with coverage..."
	@if [ -z "$$GITHUB_TOKEN" ] && [ -z "$$GITLAB_TOKEN" ] && [ -z "$$GITEA_TOKEN" ]; then \
		echo "⚠️  No authentication tokens found, skipping integration tests"; \
		echo "Set GITHUB_TOKEN, GITLAB_TOKEN, or GITEA_TOKEN to run integration tests"; \
	else \
		go test -tags=integration -race -coverprofile=integration_coverage.out -covermode=atomic ./cmd/git/...; \
	fi

# Generate combined coverage report
coverage-report: test-coverage test-integration-coverage
	@echo "📋 Generating combined coverage report..."
	@if [ -f integration_coverage.out ]; then \
		echo "mode: atomic" > combined_coverage.out; \
		grep -h -v "mode: atomic" coverage.out integration_coverage.out >> combined_coverage.out; \
		go tool cover -func=combined_coverage.out; \
		go tool cover -html=combined_coverage.out -o combined_coverage.html; \
		echo "Combined coverage report generated: combined_coverage.html"; \
	else \
		echo "Only unit test coverage available"; \
		cp coverage.out combined_coverage.out; \
		cp coverage.html combined_coverage.html; \
	fi

# Clean coverage files
clean-coverage:
	@echo "🧹 Cleaning coverage files..."
	rm -f coverage.out coverage.html integration_coverage.out combined_coverage.out combined_coverage.html

# Coverage threshold check (requires go-cover-treemap or similar tool)
coverage-check:
	@echo "🎯 Checking coverage threshold..."
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
	THRESHOLD=80; \
	if [ "$$(echo "$$COVERAGE >= $$THRESHOLD" | bc -l)" -eq 1 ]; then \
		echo "✅ Coverage $$COVERAGE% meets threshold of $$THRESHOLD%"; \
	else \
		echo "❌ Coverage $$COVERAGE% below threshold of $$THRESHOLD%"; \
		exit 1; \
	fi