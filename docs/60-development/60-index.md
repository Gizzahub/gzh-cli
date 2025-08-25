# 🛠️ Development Guide

Comprehensive development documentation for contributors and maintainers of gzh-cli.

## 📋 Table of Contents

- [Code Quality](#code-quality)
- [Development Process](#development-process)
- [Testing Strategy](#testing-strategy)
- [Debugging & Troubleshooting](#debugging--troubleshooting)
- [Build & Deployment](#build--deployment)

## 🎯 Quick Start for Contributors

### Prerequisites

- **Go 1.24.0+** (with toolchain go1.24.5)
- **Git** (any recent version)
- **Make** (for build automation)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/gizzahub/gzh-cli.git
cd gzh-cli

# Install development dependencies
make bootstrap

# Build and test
make build
make test

# Install pre-commit hooks
make pre-commit-install
```

## 📚 Development Documentation

### Code Quality & Standards

- **[Code Quality Guide](60-code-quality.md)** - Coding standards and quality practices
- **[Lint Standards](66-lint-standards.md)** - Linting rules and configuration

### Development Process

- **[Debugging Guide](61-debugging-guide.md)** - Debugging techniques and tools
- **[Mocking Strategy](62-mocking-strategy.md)** - Testing with mocks and interfaces
- **[Pre-commit Guide](63-pre-commit-guide.md)** - Pre-commit hook configuration
- **[Pre-commit Hooks](64-pre-commit-hooks.md)** - Hook implementation details

### Dependency & Build Management

- **[Dependency Management](65-dependency-management.md)** - Go modules and dependency handling
- **[Testing Strategy](67-testing-strategy.md)** - Comprehensive testing approach

## 🚀 Development Workflow

### Daily Development

```bash
# 1. Update dependencies
go mod tidy
go mod download

# 2. Run quality checks
make fmt        # Format code
make lint       # Run linters
make test       # Run tests

# 3. Build and verify
make build
make install
```

### Pre-commit Workflow

```bash
# Automatic pre-commit checks (after make pre-commit-install)
git commit -m "feat: add new feature"

# Manual pre-commit checks
make pre-commit

# Fix issues and retry
make fmt
make lint
git add .
git commit -m "feat: add new feature"
```

### Testing Workflow

```bash
# Unit tests
go test ./...

# Specific package tests
go test ./cmd/synclone -v
go test ./pkg/github -v

# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# Test coverage
make cover
```

## 🏗️ Architecture for Developers

### Code Organization

```
gzh-cli/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command and CLI setup
│   ├── synclone/          # Synclone command
│   ├── git/               # Git command suite
│   ├── quality/           # Quality management
│   └── ...
├── internal/              # Private packages
│   ├── config/           # Configuration management
│   ├── git/              # Git operations
│   ├── cli/              # CLI utilities
│   └── logger/           # Logging infrastructure
├── pkg/                   # Public packages
│   ├── config/           # Configuration types
│   ├── github/           # GitHub provider
│   ├── gitlab/           # GitLab provider
│   └── synclone/         # Synclone engine
└── test/                  # Test utilities and fixtures
```

### Key Design Patterns

#### Interface-Driven Development

```go
// Define interfaces first
type Provider interface {
    GetRepositories(ctx context.Context, opts GetRepositoriesOptions) ([]Repository, error)
    CloneRepository(ctx context.Context, repo Repository, opts CloneOptions) error
}

// Implement with concrete types
type GitHubProvider struct {
    BaseProvider
    // GitHub-specific fields
}
```

#### Command Pattern

```go
// Standard command structure
func NewCommandCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "command",
        Short: "Brief description",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Load configuration
            // 2. Create service
            // 3. Execute business logic
            // 4. Format output
        },
    }
    return cmd
}
```

#### Provider Registry Pattern

```go
// Register providers at startup
providerRegistry := provider.NewRegistry()
providerRegistry.Register("github", github.NewProvider())
providerRegistry.Register("gitlab", gitlab.NewProvider())
```

## 🧪 Testing Guidelines

### Test Categories

1. **Unit Tests** - Test individual functions and methods
1. **Integration Tests** - Test component interactions
1. **End-to-End Tests** - Test complete user workflows
1. **Performance Tests** - Test performance characteristics

### Testing Patterns

#### Table-Driven Tests

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### Mock Usage

```go
func TestServiceWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockProvider := mocks.NewMockProvider(ctrl)
    mockProvider.EXPECT().
        GetRepositories(gomock.Any(), gomock.Any()).
        Return(expectedRepos, nil)

    service := NewService(mockProvider)
    result, err := service.DoSomething()

    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
}
```

## 🔧 Build System

### Makefile Targets

```bash
# Development
make bootstrap      # Install build dependencies
make build         # Build binary
make install       # Install to GOPATH/bin
make clean         # Clean build artifacts

# Code Quality
make fmt           # Format code (gofumpt, gci)
make lint          # Run golangci-lint
make lint-all      # Format + lint + pre-commit
make test          # Run tests with coverage

# Testing
make test-unit     # Unit tests only
make test-integration  # Integration tests
make test-e2e      # End-to-end tests
make cover         # Test coverage report

# Pre-commit
make pre-commit-install  # Install pre-commit hooks
make pre-commit         # Run pre-commit checks
make pre-push          # Run pre-push checks

# Release
make release       # Create release build
make docker        # Build Docker image
```

### Build Configuration

The project uses a modular Makefile system:

- `Makefile` - Main build targets
- `scripts/` - Build and utility scripts
- `.golangci.yml` - Linter configuration
- `.pre-commit-config.yaml` - Pre-commit hook configuration

## 📦 Release Process

### Version Management

```bash
# Update version
make version-bump BUMP=minor  # patch, minor, major

# Tag release
git tag v1.2.0
git push origin v1.2.0

# Build release artifacts
make release
```

### Release Checklist

1. **Code Quality**

   - [ ] All tests pass
   - [ ] Lint checks pass
   - [ ] Code coverage ≥ 80%
   - [ ] Documentation updated

1. **Version Management**

   - [ ] Version bumped appropriately
   - [ ] CHANGELOG.md updated
   - [ ] Git tag created

1. **Testing**

   - [ ] Integration tests pass
   - [ ] E2E tests pass
   - [ ] Manual testing completed

1. **Documentation**

   - [ ] README.md updated
   - [ ] API documentation current
   - [ ] Migration guides (if needed)

## 🐛 Debugging & Troubleshooting

### Debug Build

```bash
# Build with debug symbols
make build-debug

# Run with debug logging
gz --debug --verbose synclone github --org myorg

# Enable Go runtime debugging
GODEBUG=gctrace=1 gz synclone github --org myorg
```

### Common Development Issues

#### Build Issues

```bash
# Clean and rebuild
make clean
make bootstrap
make build

# Check Go version
go version

# Update dependencies
go mod tidy
go mod download
```

#### Test Issues

```bash
# Run specific tests
go test -v ./cmd/synclone

# Run with race detection
go test -race ./...

# Generate test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Lint Issues

```bash
# Fix formatting
make fmt

# Run specific linter
golangci-lint run --enable-only=gofmt

# Fix common issues
golangci-lint run --fix
```

## 📋 Contributing Guidelines

### Code Review Process

1. **Pre-submission**

   - Run all quality checks locally
   - Ensure tests pass
   - Update documentation

1. **Pull Request**

   - Descriptive title and description
   - Link to related issues
   - Small, focused changes

1. **Review Criteria**

   - Code quality and style
   - Test coverage
   - Performance impact
   - Documentation completeness

### Commit Message Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:

```
feat(synclone): add GitLab support
fix(config): resolve YAML parsing issue
docs(api): update command reference
test(github): add integration tests
```

## 🔗 Development Resources

### External Tools

- **golangci-lint** - Comprehensive linter
- **gomock** - Mock generation
- **testify** - Testing toolkit
- **cobra** - CLI framework
- **viper** - Configuration management

### IDE Setup

- **VS Code** - `.vscode/settings.json` for project settings
- **GoLand** - IntelliJ IDEA configuration
- **Vim/Neovim** - Go plugin configuration

### Documentation

- **[Architecture Overview](../20-architecture/20-system-overview.md)** - System design
- **[Configuration Guide](../40-configuration/40-configuration-guide.md)** - Configuration system
- **[API Reference](../50-api-reference/50-index.md)** - Complete API documentation

______________________________________________________________________

**Go Version**: 1.24.0+ (toolchain: go1.24.5)
**Testing Framework**: Go standard library + testify
**Mocking**: gomock
**CLI Framework**: cobra + viper
**Build System**: Make + modular scripts
