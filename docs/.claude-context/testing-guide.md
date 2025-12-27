# Testing Guide - gzh-cli

## Test Organization

```
cmd/{module}/
├── {feature}.go
├── {feature}_test.go      # Unit tests
└── AGENTS.md              # Module-specific guide

internal/{package}/
├── {file}.go
└── {file}_test.go         # Unit tests

test/integration/
└── {feature}_test.go      # Integration tests (Docker)
```

## Running Tests

```bash
# All tests
make test

# Specific package
go test ./cmd/{module} -v

# Specific test function
go test ./cmd/git -run "TestCloneOrUpdate" -v

# Coverage
make cover

# With race detection
go test -race ./...

# Integration tests
cd test/integration && go test -v
```

## Test Coverage Target

- **Core logic**: 80%+
- **CLI commands**: Test critical paths
- **Integration tests**: Key workflows

## Mocking

### Generate mocks

```bash
# Generate all mocks
make generate-mocks

# Regenerate (clean + generate)
make regenerate-mocks
```

### Using gomock

```go
import (
    "testing"
    "github.com/golang/mock/gomock"
)

func TestWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockClient := mocks.NewMockClient(ctrl)
    mockClient.EXPECT().
        Clone(gomock.Any(), gomock.Any()).
        Return(nil)

    // Test with mock
}
```

## Environment-Specific Tests

### Skipping tests without credentials

```go
func TestGitHubAPI(t *testing.T) {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        t.Skip("GITHUB_TOKEN not set - skipping integration test")
    }
    // Test with real API
}
```

### Required environment variables

- `GITHUB_TOKEN` - GitHub API tests
- `GITLAB_TOKEN` - GitLab API tests
- `GITEA_URL` - Gitea API tests

## Integration Tests

### Location

`test/integration/`

### Running

```bash
cd test/integration
go test -v
```

### Docker containers

Integration tests may use Docker for:

- Git servers (GitHub Enterprise, GitLab)
- Package managers
- Database systems

## Test Helpers

### From gzh-cli-core

```go
import "github.com/gizzahub/gzh-cli-core/testutil"

// Temp directory (auto-cleanup)
tempDir := testutil.TempDir(t)

// Assertions
testutil.AssertNoError(t, err)
testutil.AssertEqual(t, expected, actual)

// Capture output
output := testutil.CaptureOutput(func() {
    fmt.Println("test")
})
```

### Git-specific helpers

Check `cmd/git/testutil/` for Git-specific test utilities.

## Test Naming Conventions

```go
// Unit tests
func TestCloneRepository(t *testing.T) {}

// Table-driven tests
func TestCloneRepository_MultipleScenarios(t *testing.T) {
    tests := []struct{
        name string
        // ...
    }{
        {name: "successful clone"},
        {name: "repository exists"},
        {name: "network error"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}

// Integration tests
func TestIntegration_FullWorkflow(t *testing.T) {}
```
