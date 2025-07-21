# Testing Strategy

## Test Organization

gzh-manager uses a three-tier testing strategy:

### 1. Unit Tests
- Located alongside the code in each package
- Fast, isolated tests with no external dependencies
- Run with `make test-unit`

### 2. Integration Tests
- Located in `test/integration/`
- Test interactions with external systems (Docker, APIs)
- Use build tag `//go:build integration`
- Run with `make test-integration`

### 3. End-to-End (E2E) Tests
- Located in `test/e2e/`
- Test complete user workflows
- Use build tag `//go:build e2e`
- Run with `make test-e2e`

## Running Tests

### Run all tests
```bash
make test-all
```

### Run only unit tests (fast)
```bash
make test-unit
```

### Run integration tests
```bash
make test-integration
# or with Docker
make test-docker
```

### Run E2E tests
```bash
make test-e2e
```

## Writing Tests

### Unit Test Example
```go
// internal/git/operations_test.go
package git

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestOperations_Clone(t *testing.T) {
    // Unit test - no external dependencies
    ops := NewOperations(false)
    // ... test logic
}
```

### Integration Test Example
```go
//go:build integration
// +build integration

// test/integration/github/repo_config_integration_test.go
package github_test

import (
    "testing"
    "os"
)

func TestGitHubIntegration(t *testing.T) {
    if os.Getenv("GITHUB_TOKEN") == "" {
        t.Skip("GITHUB_TOKEN not set")
    }
    // ... integration test logic
}
```

### E2E Test Example
```go
//go:build e2e
// +build e2e

// test/e2e/scenarios/bulk_clone_e2e_test.go
package scenarios

import (
    "testing"
    "os/exec"
)

func TestBulkCloneE2E(t *testing.T) {
    // Test complete workflow
    cmd := exec.Command("gz", "bulk-clone", "github", "myorg")
    // ... e2e test logic
}
```

## Test Coverage

### Generate coverage report
```bash
# Unit test coverage
make test-unit
# Creates coverage-unit.out

# All tests coverage
make test
# Creates coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Coverage Goals
- Unit tests: 80% coverage minimum
- Critical packages: 90% coverage
- Integration tests: Focus on API contracts
- E2E tests: Cover main user workflows

## Best Practices

1. **Use build tags** for integration/e2e tests
2. **Mock external dependencies** in unit tests
3. **Use testify** for assertions
4. **Table-driven tests** for multiple scenarios
5. **Parallel tests** where possible (`t.Parallel()`)
6. **Skip tests** when prerequisites missing
7. **Clean up** resources in tests

## CI/CD Integration

Tests run in different stages:

1. **PR Checks**: Unit tests only (fast feedback)
2. **Merge Queue**: Unit + Integration tests
3. **Nightly**: Full test suite including E2E

## Debugging Tests

### Run specific test
```bash
go test -v -run TestOperations_Clone ./internal/git/
```

### Run with race detector
```bash
go test -race ./...
```

### Verbose output
```bash
go test -v ./...
```

### Test with specific tags
```bash
go test -tags=integration -v ./test/integration/github/...
```
