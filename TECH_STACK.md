# Tech Stack

## Core Technologies

- **Language**: Go 1.24.0+ (toolchain: go1.24.5)
- **Framework**: Cobra CLI framework
- **Database**: File-based configuration (YAML/JSON)
- **Cloud Platform**: Multi-cloud support (AWS, GCP, Azure)

## Development Tools

- **Build System**: Make + Go modules
- **Testing**: testify framework with gomock
- **Linting**: golangci-lint v2 with comprehensive rules
- **Package Manager**: Go modules
- **Performance Monitoring**: Standard Go pprof integration

## External Dependencies

### Core Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `gopkg.in/yaml.v3` - YAML processing
- `github.com/go-git/go-git/v5` - Git operations
- `github.com/google/go-github/v45` - GitHub API
- `github.com/xanzy/go-gitlab` - GitLab API
- `github.com/fatih/color` - Terminal colors
- `github.com/schollz/progressbar/v3` - Progress bars
- `github.com/fsnotify/fsnotify` - File system monitoring (IDE features)

### Development Dependencies

- `github.com/stretchr/testify` - Testing framework
- `github.com/golang/mock/gomock` - Mock generation
- `github.com/golangci/golangci-lint` - Code linting
- `golang.org/x/tools/go/packages` - Code analysis
- `github.com/xeipuuv/gojsonschema` - JSON schema validation

### UI/TUI Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - TUI styling

### Cloud SDKs

- `github.com/aws/aws-sdk-go-v2` - AWS SDK
- Cloud provider abstractions for multi-cloud support

## Architecture Overview

gzh-cli follows a **simplified CLI architecture** optimized for developer productivity and maintainability. The architecture was recently simplified (2025-01) to remove over-engineered components and focus on CLI-appropriate patterns.

### Command Structure

```
cmd/
├── root.go              # Main CLI entry point with direct constructors
├── git/                 # Git platform integration (unified interface)
├── synclone/           # Multi-platform repository cloning
├── ide/                # JetBrains IDE monitoring and management
├── quality/            # Multi-language code quality tools
├── profile/            # Performance profiling (simplified pprof)
├── dev-env/            # Development environment management
├── net-env/            # Network environment transitions
├── pm/                 # Package manager integration
├── repo-config/        # GitHub repository configuration
├── doctor/             # System diagnostics (hidden)
├── shell/              # Interactive debugging shell (debug mode)
├── actions-policy/     # GitHub Actions policy (planned)
└── man.go              # Manual page generation (planned)
```

### Package Architecture

```
pkg/
├── synclone/           # Configuration loading and validation
├── github/             # GitHub API integration
├── gitlab/             # GitLab API integration
├── gitea/              # Gitea API integration
├── gogs/               # Gogs API integration (planned)
└── example/            # Example package structure
```

### Internal Architecture (Simplified)

```
internal/
├── git/                # Core Git operations with dependency injection
│   ├── interfaces.go   # Client, StrategyExecutor, BulkOperator interfaces
│   └── constructors.go # Direct constructor implementations
├── simpleprof/         # Simplified profiling using standard Go pprof
├── logger/             # Logging abstraction with SimpleLogger
└── testlib/            # Testing utilities and environment checkers
```

### Helper Utilities

```
helpers/
└── git_helper.go       # Git repository operations and testing utilities
```

## Architecture Evolution (2025-01 Simplification)

### Removed Components

The architecture was recently simplified to remove over-engineered components inappropriate for CLI tools:

1. **Dependency Injection Container** (`internal/container/`):

   - **Removed**: ~1,188 lines of complex DI container code
   - **Replaced with**: Direct constructor calls in command initialization
   - **Rationale**: CLI tools don't need runtime service discovery

1. **Complex Profiling System** (`internal/profiling/`):

   - **Removed**: Custom HTTP server with multiple abstractions
   - **Replaced with**: Standard Go pprof integration via `internal/simpleprof/`
   - **Rationale**: Standard pprof tooling is more appropriate and familiar

### Current Design Patterns

1. **Simplified Architecture**: Clean, direct implementation focused on CLI tool needs

   - Direct constructor pattern instead of dependency injection containers
   - Standard Go pprof integration instead of custom profiling systems
   - Minimal abstractions with clear benefits

1. **Service-specific implementations**: Each Git platform (GitHub, GitLab, Gitea, Gogs) has dedicated packages following common interfaces

1. **Configuration-driven design**: Extensive YAML configuration support with schema validation (see `examples/` directory)

1. **Cobra CLI framework**: All commands use cobra with consistent flag patterns and help documentation

1. **Cross-platform support**: Native OS detection and platform-specific implementations (Linux, macOS, Windows)

1. **Environment variable integration**: Support for token authentication and configuration overrides

1. **Atomic operations**: Commands designed for safe execution with backup and rollback capabilities

1. **Comprehensive testing**: testify framework with mock services and environment-specific tests

1. **Standard Go tooling integration**: Uses standard `runtime/pprof` for profiling instead of custom implementations

1. **URL Parsing Strategy**: Robust URL parsing for multiple Git hosting formats (HTTPS, SSH, ssh://) in `extractRepoNameFromURL`

1. **Performance monitoring**: Automated benchmarking system with regression detection (see `scripts/` directory)

## Performance Monitoring System

### Automated Benchmarking

The project includes automated performance monitoring to prevent regressions:

```bash
scripts/
├── simple-benchmark.sh          # Quick performance checks (startup, size, memory)
└── benchmark-performance.sh     # Comprehensive benchmarking with baselines
```

### Performance Metrics

- **Startup Time**: Target < 50ms (currently ~10ms)
- **Binary Size**: ~33MB maintained
- **Memory Usage**: Minimal heap allocation
- **Command Response**: Most commands < 100ms

### Profiling Integration

```bash
# Simplified profiling using standard Go pprof
gz profile stats                 # Runtime statistics
gz profile server --port 6060   # HTTP pprof server
gz profile cpu --duration 30s   # CPU profiling
gz profile memory                # Memory profiling
```

## Deployment

### Binary Distribution

- **Binary Name**: `gz`
- **Platforms**: Linux, macOS, Windows (64-bit)
- **Installation**:
  - Go install: `go install github.com/gizzahub/gzh-cli@latest`
  - Manual build: `make build && make install`
  - Pre-compiled binaries: Available in GitHub Releases

### Build Process

```bash
# Development setup
make bootstrap          # Install build dependencies (one-time)
make build             # Create 'gz' executable
make install           # Install to GOPATH/bin

# Quality assurance (always run before commit)
make fmt               # Format code with gofumpt and gci
make lint              # Run golangci-lint with auto-fix
make test              # Run tests with coverage
make lint-all          # Complete quality check pipeline
```

### Configuration

- **Config Hierarchy**:

  1. Environment variable: `GZH_CONFIG_PATH`
  1. Current directory: `./synclone.yaml` or `./synclone.yml`
  1. User config: `~/.config/gzh-manager/synclone.yaml`
  1. System config: `/etc/gzh-manager/synclone.yaml`

- **Schema Validation**:

  - JSON Schema: `docs/synclone-schema.json`
  - YAML Schema: `docs/synclone-schema.yaml`
  - Built-in validator: `gz synclone validate`

## Quality Tools Integration

### Code Quality Management

The project includes a comprehensive code quality system (`gz quality`):

**Supported Languages and Tools**:

- **Go**: gofumpt, golangci-lint, goimports, gci
- **Python**: ruff (format + lint), black, isort, flake8, mypy
- **JavaScript/TypeScript**: prettier, eslint, dprint
- **Rust**: rustfmt, clippy
- **Java**: google-java-format, checkstyle, spotbugs
- **C/C++**: clang-format, clang-tidy
- **Other**: YAML, JSON, Markdown, Shell script support

**Quality Commands**:

```bash
gz quality run          # Run all formatting and linting tools
gz quality check        # Lint-only mode (no changes)
gz quality init         # Generate project configurations
gz quality install     # Install quality tools
gz quality analyze     # Project analysis and recommendations
```

## IDE Integration

### JetBrains IDE Support

Comprehensive JetBrains IDE monitoring and management (`gz ide`):

**Supported Products**:

- IntelliJ IDEA (Community, Ultimate)
- PyCharm (Community, Professional)
- WebStorm, PhpStorm, RubyMine
- CLion, GoLand, DataGrip
- Android Studio, Rider

**Features**:

- Real-time settings monitoring
- Cross-platform support (Linux, macOS, Windows)
- Automatic sync issue detection and fixes
- Configuration backup and restore

```bash
gz ide monitor          # Real-time settings monitoring
gz ide fix-sync        # Fix synchronization issues
gz ide list            # List detected IDE installations
```

## Git Platform Integration

### Multi-Platform Support

Unified Git platform management through `gz git` command:

**Supported Platforms**:

- GitHub (Organizations, Personal repos)
- GitLab (Groups, Projects)
- Gitea (Organizations, Personal repos)
- Gogs (planned)

**Key Features**:

- Smart clone-or-update strategies (rebase, reset, pull, fetch, clone, skip)
- Bulk repository operations (up to 50 parallel)
- Webhook management across platforms
- Event processing and automation

## Development Environment

### Required Tools

- **Go**: 1.24.0+ (as specified in go.mod)
- **Make**: For build automation
- **Git**: 2.0+ for repository operations
- **golangci-lint**: v2+ for code linting (auto-installed via `make bootstrap`)

### Optional Tools

- **Docker**: For containerized testing
- **jq**: For JSON processing in benchmarks
- **bc**: For benchmark calculations

### Development Workflow

```bash
# One-time setup
make bootstrap          # Install all build dependencies

# Development cycle
make build             # Build gz executable
make test              # Run tests
make fmt               # Format code
make lint              # Run linters
make lint-all          # Complete quality pipeline

# Pre-commit setup (optional but recommended)
make pre-commit-install # Install pre-commit hooks
make pre-commit        # Run pre-commit checks manually

# Performance monitoring
./scripts/simple-benchmark.sh  # Quick performance check
```

## Testing Strategy

### Test Organization

- **Unit Tests**: `*_test.go` files alongside source code
- **Integration Tests**: Environment-specific tests with external service mocking
- **Mock Generation**: `gomock` for interface mocking
- **Test Coverage**: Maintained through `make test` and `make cover`

### Test Execution

```bash
# Run all tests
make test

# Run specific package tests
go test ./cmd/ide -v
go test ./cmd/quality -v
go test ./pkg/github -v

# Run specific test functions
go test ./cmd/git -run "TestExtractRepoNameFromURL" -v

# Coverage with race detection
make cover
```

### Mock Management

```bash
make generate-mocks     # Generate all interface mocks
make clean-mocks       # Remove generated mocks
make regenerate-mocks  # Clean and regenerate all mocks
```

## Security Considerations

### Code Security

- **Static Analysis**: gosec integration through golangci-lint
- **Dependency Scanning**: Regular dependency vulnerability checks
- **Secret Management**: Environment variable-based authentication
- **File Permissions**: Secure handling of configuration and backup files

### Authentication

- **Token-based**: GitHub, GitLab, Gitea API tokens
- **Environment Variables**: Secure token storage
- **Multi-platform**: Consistent auth across Git platforms

## Observability

### Logging

- **Structured Logging**: Level-based logging (debug, info, warn, error)
- **Global Flags**: `--verbose`, `--debug`, `--quiet` for log control
- **Context Propagation**: Proper context handling across operations

### Metrics and Monitoring

- **Performance Benchmarking**: Automated startup time and resource monitoring
- **Profiling**: Standard Go pprof integration for performance analysis
- **System Health**: `gz doctor` command for system diagnostics (hidden utility)

## Future Architecture Considerations

### Planned Enhancements

- **Plugin System**: Extensible architecture for custom commands
- **Configuration Management**: Enhanced configuration validation and templating
- **Cloud Integration**: Expanded cloud provider support
- **Telemetry**: Optional usage analytics (privacy-preserving)

### Scalability

The current architecture is designed for:

- **Single-user CLI usage**: Optimized for individual developer workflows
- **Large-scale repository operations**: Up to 50 parallel clone operations
- **Cross-platform consistency**: Uniform behavior across operating systems
- **Resource efficiency**: Minimal memory footprint and fast startup times

## Specification-Driven Development

The project follows a specification-first approach:

- **Specifications**: All features documented in `specs/` directory before implementation
- **Authority Hierarchy**: `specs/` → source code → `docs/` (in order of authority)
- **Implementation Status**: Clear tracking of ✅ implemented, 🚧 in-progress, 📋 planned features

This ensures consistency between documentation and implementation, with specifications serving as the "source of truth" for all functionality.
