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

## CLI Specification Testing

gzh-cli follows **SDD (Specification-Driven Development)** for CLI commands, where specifications define expected behavior before implementation.

### 4. CLI Specification Tests

- Located in `docs/testing/{command}/`
- Define complete input-output contracts
- Use structured scenario format
- Include success, error, and edge cases
- Run with `make test-specs`

### CLI Specification Format

Each command specification follows a standard format:

```markdown
# Command: gz [subcommand] [options]

## Scenario: [Brief description]

### Input
**Command**: `gz command --flag value`
**Prerequisites**: [Required conditions]

### Expected Output
**Success Case**: [Expected stdout, stderr, exit code]
**Error Cases**: [Error conditions and outputs]

### Side Effects
**Files Created**: [List of files/directories]
**State Changes**: [Configuration or cache changes]

### Validation
**Automated Tests**: [Assertion commands]
**Manual Verification**: [Step-by-step checks]
```

### Command Contract Testing

Validate CLI contracts using structured assertions:

```bash
# Example contract test
test_synclone_github_success() {
    result=$(gz synclone github -o test-org 2>&1)
    exit_code=$?
    
    # Contract assertions
    assert_contains "$result" "📋 Found"
    assert_contains "$result" "repositories in organization test-org"
    assert_exit_code 0
    assert_directory_exists "./test-org"
    assert_file_exists "./test-org/gzh.yaml"
}

# Error contract test
test_synclone_rate_limit() {
    unset GITHUB_TOKEN
    result=$(gz synclone github -o microsoft 2>&1)
    exit_code=$?
    
    # Critical: NO Usage block for rate limit errors
    assert_not_contains "$result" "Usage:"
    assert_contains "$result" "🚫 GitHub API Rate Limit Exceeded!"
    assert_exit_code 1
}
```

### Specification-Based Test Categories

**Success Scenarios**:
- Happy path with valid inputs
- Expected output patterns
- Correct side effects

**Error Scenarios**:
- Rate limiting (no Usage block)
- Authentication failures (no Usage block)
- Network errors
- Invalid inputs

**Edge Cases**:
- Large datasets (pagination)
- Empty results
- Special characters
- Resource constraints

### Example Specifications

See detailed command specifications:
- [`specs/cli/synclone/UC-002-github-clone.md`](../../specs/cli/synclone/UC-002-github-clone.md) - Successful clone
- [`specs/cli/synclone/UC-003-rate-limit.md`](../../specs/cli/synclone/UC-003-rate-limit.md) - Rate limit handling
- [`specs/cli/synclone/UC-004-auth-error.md`](../../specs/cli/synclone/UC-004-auth-error.md) - Authentication errors
- [`specs/cli/synclone/UC-005-pagination.md`](../../specs/cli/synclone/UC-005-pagination.md) - Large organization pagination

## Best Practices

1. **Use build tags** for integration/e2e tests
1. **Mock external dependencies** in unit tests
1. **Use testify** for assertions
1. **Table-driven tests** for multiple scenarios
1. **Parallel tests** where possible (`t.Parallel()`)
1. **Skip tests** when prerequisites missing
1. **Clean up** resources in tests
1. **Write specifications before implementation** (SDD approach)
1. **Validate CLI contracts** with structured assertions
1. **Test error cases** without Usage blocks

## CI/CD Integration

Tests run in different stages:

1. **PR Checks**: Unit tests + CLI specification validation (fast feedback)
1. **Merge Queue**: Unit + Integration + CLI contract tests
1. **Nightly**: Full test suite including E2E and specification compliance

### CLI Specification Integration

```yaml
# .github/workflows/cli-specs.yml
name: CLI Specification Validation
on: [push, pull_request]

jobs:
  validate-specs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build CLI
        run: make build
      - name: Validate CLI Specifications
        run: make test-specs
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

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

### Run CLI specification tests

```bash
# Run all specification-based tests
make test-specs

# Test specific command specifications
make test-specs-synclone

# Validate specification format
make validate-specs

# Generate specification documentation
make generate-spec-docs
```

## Related Documentation

- [CLI Specification Strategy](68-cli-specification-strategy.md) - Complete SDD methodology
- [CLI Specifications](../../specs/cli/) - All CLI command specifications
- [CLI Template](../../specs/cli/template.md) - Standard specification format
