# End-to-End (E2E) Test Scenarios

This directory contains end-to-end test scenarios that validate complete user workflows across the entire `gz` CLI application.

## Overview

E2E tests simulate real user interactions with the CLI tool, testing complete workflows from start to finish. These tests validate that all components work together correctly and provide confidence that the application functions as expected from a user's perspective.

## Test Scenarios

### 1. **Bulk Clone Workflow** (`bulk_clone_e2e_test.go`)

Complete repository cloning workflow across multiple Git providers.

**Scenarios:**

- Configure multiple Git providers (GitHub, GitLab, Gitea)
- Generate configuration from existing directories
- Execute bulk clone operations with different strategies
- Validate cloned repositories and directory structure
- Handle authentication and error scenarios

**Commands Tested:**

- `gz bulk-clone --config config.yaml`
- `gz synclone config generate --provider github --org myorg`
- `gz config validate --config config.yaml`

### 2. **Development Environment Setup** (`dev_env_e2e_test.go`)

End-to-end development environment configuration and management.

**Scenarios:**

- Configure AWS, Docker, and Kubernetes environments
- Manage credentials and configuration files
- Switch between different environment profiles
- Validate environment health and connectivity

**Commands Tested:**

- `gz dev-env aws configure`
- `gz dev-env docker setup`
- `gz dev-env kubeconfig switch`

### 3. **IDE Integration Workflow** (`ide_e2e_test.go`)

JetBrains IDE settings monitoring and synchronization.

**Scenarios:**

- Monitor IDE settings changes in real-time
- Fix synchronization issues automatically
- Handle multiple IDE installations
- Backup and restore settings

**Commands Tested:**

- `gz ide monitor --daemon`
- `gz ide list`
- `gz ide fix-sync`

### 4. **Repository Configuration Management** (`repo_config_e2e_test.go`)

Complete repository configuration lifecycle management.

**Scenarios:**

- Define repository configuration templates
- Apply configurations to multiple repositories
- Audit compliance across organizations
- Handle configuration inheritance and exceptions

**Commands Tested:**

- `gz repo-config apply --template security`
- `gz repo-config audit --org myorg`
- `gz repo-config validate --config config.yaml`

### 5. **Network Environment Transitions** (`net_env_e2e_test.go`)

Network environment detection and automated transitions.

**Scenarios:**

- Detect WiFi network changes
- Automatically configure VPN connections
- Update DNS and proxy settings
- Handle network failure scenarios

**Commands Tested:**

- `gz net-env monitor`
- `gz net-env configure --network office`
- `gz net-env status`

### 6. **Configuration Management** (`config_e2e_test.go`)

Comprehensive configuration management across all modules.

**Scenarios:**

- Initialize configuration from scratch
- Migrate between configuration versions
- Validate configuration schemas
- Handle configuration inheritance and profiles

**Commands Tested:**

- `gz config init`
- `gz config migrate --from v1 --to v2`
- `gz config validate --all`

## Test Structure

```
test/e2e/
├── README.md                   # This file
├── fixtures/                   # Test data and configuration files
│   ├── configs/               # Sample configuration files
│   ├── repositories/          # Test repository data
│   └── templates/             # Configuration templates
├── helpers/                    # Common test utilities
│   ├── cli.go                 # CLI execution helpers
│   ├── filesystem.go          # File system utilities
│   ├── assertions.go          # Custom assertions
│   └── cleanup.go             # Test cleanup utilities
├── scenarios/                  # Individual test scenarios
│   ├── bulk_clone_e2e_test.go
│   ├── dev_env_e2e_test.go
│   ├── ide_e2e_test.go
│   ├── repo_config_e2e_test.go
│   ├── net_env_e2e_test.go
│   └── config_e2e_test.go
└── run_e2e_tests.sh           # Test runner script
```

## Running E2E Tests

### Prerequisites

1. **Build the Application**

   ```bash
   make build
   ```

2. **Install Dependencies**

   ```bash
   make bootstrap
   ```

3. **Set Environment Variables** (optional for external services)
   ```bash
   export GITHUB_TOKEN="your-token"
   export GITLAB_TOKEN="your-token"
   export GITEA_TOKEN="your-token"
   ```

### Running Tests

#### Run All E2E Tests

```bash
# Using make target
make test-e2e

# Using test runner script
./test/e2e/run_e2e_tests.sh

# Using go test directly
go test ./test/e2e/scenarios/... -v -timeout 30m
```

#### Run Specific Scenarios

```bash
# Bulk clone scenarios only
go test ./test/e2e/scenarios -v -run TestBulkClone

# Development environment scenarios
go test ./test/e2e/scenarios -v -run TestDevEnv

# Configuration management scenarios
go test ./test/e2e/scenarios -v -run TestConfig
```

#### Run with Different Modes

```bash
# Quick mode (skip slow tests)
go test ./test/e2e/scenarios -v -short

# With specific timeout
go test ./test/e2e/scenarios -v -timeout 45m

# With verbose output
go test ./test/e2e/scenarios -v -x
```

## Test Environment

### Isolation

- Each test runs in isolated temporary directories
- Tests clean up after themselves (files, processes, etc.)
- No interference between test scenarios

### External Dependencies

- **Optional**: Real Git service tokens for complete integration
- **Mock Mode**: Use mock servers when tokens unavailable
- **Docker**: Some tests may use testcontainers for isolated services

### File System Layout

```
/tmp/gz-e2e-tests-{random}/
├── home/                    # Simulated user home directory
├── work/                    # Working directory for operations
├── configs/                 # Generated configuration files
├── repositories/            # Cloned/test repositories
└── logs/                    # Application logs
```

## Writing New E2E Tests

### Test Structure Template

```go
func TestNewFeature_E2E(t *testing.T) {
    // Setup test environment
    env := helpers.NewTestEnvironment(t)
    defer env.Cleanup()

    // Given: Initial state setup
    env.CreateConfig("feature-config.yaml", configData)

    // When: Execute CLI commands
    result := env.RunCommand("gz", "feature", "command", "--flag", "value")

    // Then: Validate results
    require.NoError(t, result.Error)
    assert.Contains(t, result.Output, "expected-output")
    env.AssertFileExists("expected/file/path")
}
```

### Best Practices

1. **Descriptive Test Names**: Use format `TestFeature_Scenario_E2E`
2. **Isolation**: Each test should be independent and idempotent
3. **Cleanup**: Always clean up resources (use defer statements)
4. **Assertions**: Use clear, specific assertions with helpful error messages
5. **Documentation**: Document complex test scenarios and their purpose
6. **Performance**: Consider test execution time and optimize where possible

### Common Patterns

#### Configuration Setup

```go
config := &Config{
    Version: "1.0.0",
    Providers: map[string]Provider{
        "github": {Token: "test-token", Orgs: []string{"test-org"}},
    },
}
env.WriteConfig("bulk-clone.yaml", config)
```

#### Command Execution

```go
result := env.RunCommand("gz", "bulk-clone", "--config", "bulk-clone.yaml", "--dry-run")
require.NoError(t, result.Error)
assert.Contains(t, result.Output, "Would clone")
```

#### File System Validation

```go
env.AssertFileExists("cloned-repos/org/repo/.git")
env.AssertFileContains("config.yaml", "expected-content")
env.AssertDirectoryNotEmpty("logs/")
```

## Troubleshooting

### Common Issues

#### Test Timeouts

```
Error: Test exceeded timeout
```

**Solution**: Increase timeout or optimize test performance

```bash
go test ./test/e2e/scenarios -v -timeout 60m
```

#### Permission Errors

```
Error: Permission denied accessing test directory
```

**Solution**: Check directory permissions and cleanup

```bash
sudo rm -rf /tmp/gz-e2e-tests-*
```

#### CLI Not Found

```
Error: gz command not found
```

**Solution**: Build the application first

```bash
make build
export PATH="$PWD:$PATH"
```

#### External Service Errors

```
Error: GitHub API authentication failed
```

**Solution**: Set up proper tokens or run in mock mode

```bash
export GITHUB_TOKEN="your-token"
# or
go test ./test/e2e/scenarios -v -tags mock
```

### Debug Mode

Enable debug output for troubleshooting:

```bash
# Enable debug logging
export GZ_DEBUG=true

# Enable trace logging
export GZ_TRACE=true

# Run specific test with verbose output
go test ./test/e2e/scenarios -v -run TestBulkClone -x
```

### Log Analysis

Test logs are available in:

- Test output: `go test -v` output
- Application logs: `/tmp/gz-e2e-tests-*/logs/`
- System logs: Check system logs for any issues

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.24

      - name: Build Application
        run: make build

      - name: Run E2E Tests
        env:
          GITHUB_TOKEN: ${{ secrets.E2E_GITHUB_TOKEN }}
        run: make test-e2e
```

### Performance Considerations

- **Parallel Execution**: Tests run in parallel where possible
- **Resource Cleanup**: Aggressive cleanup to prevent resource leaks
- **Timeout Management**: Appropriate timeouts for different scenario types
- **Mock Integration**: Fast fallback when external services unavailable

## Maintenance

### Regular Tasks

1. **Update Test Data**: Keep test fixtures current with schema changes
2. **Review Scenarios**: Ensure scenarios cover new features
3. **Performance Monitoring**: Track test execution times
4. **Dependency Updates**: Update test dependencies regularly

### Version Compatibility

- E2E tests validate current version behavior
- Migration tests ensure backward compatibility
- Deprecation warnings for removed features
