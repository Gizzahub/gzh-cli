# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gzh-cli is a comprehensive CLI tool (binary name: `gz`) for managing development environments and Git repositories across multiple platforms. It provides unified commands for repository operations, development environment management, code quality control, and network environment transitions. The project follows a simplified CLI architecture optimized for developer productivity.

Each command directory under `cmd/` contains its own `AGENTS.md` with coding
and testing conventions. Review the relevant module's file before making
changes.

## Essential Commands

### Development Setup

```bash
make bootstrap  # Install all build dependencies - run this first
```

### Building and Running

```bash
make build      # Creates 'gz' executable
make install    # Installs to GOPATH/bin
make run        # Run with version tag
```

### Code Quality - ALWAYS RUN BEFORE COMMITTING

```bash
make fmt        # Format code with gofumpt and gci
make lint       # Run golangci-lint checks with auto-fix
make test       # Run all tests with coverage
make lint-all   # Run all linting steps (format + lint + pre-commit)
```

### Single Package Testing

```bash
# Test specific packages
go test ./cmd/git -v                    # Test git command package
go test ./cmd/synclone -v               # Test synclone package
go test ./pkg/github -v                 # Test GitHub integration
go test ./internal/git -v               # Test internal git operations

# Run specific test functions
go test ./cmd/git -run "TestExtractRepoNameFromURL" -v
go test ./cmd/git -run "TestCloneOrUpdate" -v
```

### Pre-commit Hooks Setup

```bash
make pre-commit-install    # Install pre-commit hooks (one-time setup)
make pre-commit           # Run pre-commit hooks manually
make pre-push             # Run pre-push hooks manually
make check-consistency    # Verify lint configuration consistency
```

### Testing

```bash
make test       # Run unit tests with coverage
make cover      # Show coverage with race detection
go test ./cmd/synclone -v          # Run specific package tests
go test ./cmd/ide -v               # Run IDE package tests
go test ./pkg/github -v            # Run GitHub integration tests
```

### Performance Monitoring

```bash
./scripts/simple-benchmark.sh                    # Quick performance check
./scripts/benchmark-performance.sh --baseline    # Create performance baseline
./scripts/benchmark-performance.sh --compare baseline.json  # Compare against baseline
./scripts/benchmark-performance.sh --format human          # Human-readable output
```

### Mocking Strategy

```bash
make generate-mocks    # Generate all interface mocks using gomock
make clean-mocks      # Remove all generated mock files
make regenerate-mocks # Clean and regenerate all mocks
```

## Architecture

gzh-cli follows a **simplified CLI architecture** (refactored 2025-01) that prioritizes developer productivity over abstract patterns. The architecture centers around direct constructors, interface-based abstractions, and modular command organization.

### High-Level Architecture Principles

1. **Interface-Driven Design**: Core abstractions through interfaces with concrete implementations
1. **Direct Constructor Pattern**: Avoid over-engineering with DI containers, use simple constructors
1. **Command-Centric Organization**: Each major feature is a top-level command with subcommands
1. **Configuration-First**: Unified YAML configuration system with schema validation
1. **Multi-Platform Support**: Abstracted platform providers for GitHub, GitLab, Gitea, Gogs

### Command Structure (cmd/)

```
cmd/
├── root.go              # Main CLI entry with all command registrations
├── git/                 # Unified Git platform management
│   ├── repo_clone_or_update.go  # Smart cloning with strategies
│   ├── repo_list.go     # Repository listing with output formats
│   ├── webhook.go       # Webhook management
│   └── event.go         # GitHub event processing
├── synclone/            # Multi-platform repository synchronization
├── quality/             # Multi-language code quality management
├── repo-config/         # Repository configuration management
├── dev-env/             # Development environment management
├── net-env/             # Network environment transitions
├── ide/                 # JetBrains IDE monitoring
├── pm/                  # Package manager updates
├── profile/             # Performance profiling (Go pprof)
└── doctor/              # System health diagnostics
```

### Core Architecture Layers

#### 1. Internal Layer (internal/)

**Purpose**: Private abstractions and implementations

- **`git/`** - Core Git operations with strategy pattern
  - `interfaces.go` - Client, StrategyExecutor, BulkOperator interfaces
  - `constructors.go` - Concrete implementations with dependency injection
  - `operations.go` - Git operations (clone, pull, push, reset)
- **`config/`** - Configuration management with validation
- **`logger/`** - Structured logging abstractions
- **`cli/`** - Command builder and output formatting

#### 2. Package Layer (pkg/)

**Purpose**: Public APIs and platform implementations

- **`config/`** - Unified configuration system with schema validation
- **`github/`, `gitlab/`, `gitea/`** - Platform-specific API implementations
- **`synclone/`** - Multi-platform synchronization logic
- **`git/provider/`** - Git provider abstraction layer

#### 3. Command Layer (cmd/)

**Purpose**: CLI command implementations using Cobra framework

### Key Architectural Patterns

#### Interface-Based Abstractions

```go
// Git operations abstraction
type Client interface {
    Clone(ctx context.Context, options CloneOptions) error
    Pull(ctx context.Context, options PullOptions) error
    // ...
}

// Configuration loading abstraction
type Loader interface {
    LoadConfig(ctx context.Context) (*Config, error)
    LoadConfigFromFile(ctx context.Context, filename string) (*Config, error)
}
```

#### Provider Registry Pattern

```go
// Platform providers registered at startup
providerRegistry := provider.NewRegistry()
providerRegistry.Register("github", github.NewProvider())
providerRegistry.Register("gitlab", gitlab.NewProvider())
```

#### Strategy Pattern for Git Operations

- **rebase**: Rebase local changes on remote
- **reset**: Hard reset to match remote state
- **clone**: Fresh clone (remove existing)
- **pull**: Standard git pull (merge)
- **fetch**: Update refs only

### Configuration Architecture

#### Unified Configuration System

- **Single file**: `gzh.yaml` for all commands
- **Priority system**: CLI flags > env vars > config files > defaults
- **Schema validation**: JSON Schema with detailed error messages
- **Environment variable expansion**: `${GITHUB_TOKEN}` support

#### Configuration Structure

```yaml
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/github/myorg"
```

### Testing Architecture

#### Test Organization

- **Unit tests**: `*_test.go` files alongside source
- **Integration tests**: `test/integration/` with Docker containers
- **E2E tests**: `test/e2e/` with real CLI execution
- **Mocking**: Generated mocks with `gomock` for interfaces

#### Test Categories

```bash
make test-unit          # Fast unit tests
make test-integration   # Docker-based integration tests
make test-e2e           # End-to-end CLI tests
make test-all           # Complete test suite
```

### Data Flow Architecture

#### Typical Command Execution Flow

1. **Command parsing** (Cobra) → **Flag validation**
1. **Configuration loading** (unified config system)
1. **Provider factory** → **Interface implementation**
1. **Business logic execution** → **Result formatting**
1. **Output generation** (table/JSON/YAML/CSV)

## Configuration and Schema

### Configuration File Hierarchy

1. Environment variable: `GZH_CONFIG_PATH`
1. Current directory: `./synclone.yaml` or `./synclone.yml`
1. User config: `~/.config/gzh-manager/synclone.yaml`
1. System config: `/etc/gzh-manager/synclone.yaml`

### Schema Validation

- JSON Schema: `docs/synclone-schema.json`
- YAML Schema: `docs/synclone-schema.yaml`
- Built-in validator: `gz synclone validate`

### Sample Configurations

- `examples/synclone-simple.yaml` - Minimal working example
- `examples/synclone-example.yaml` - Comprehensive with comments
- `examples/synclone.yml` - Advanced features

## Testing Guidelines

- Test files use `*_test.go` convention
- Uses testify for assertions
- Environment-specific tests check for tokens (GITHUB_TOKEN, GITLAB_TOKEN)
- Integration tests mock external services when possible
- Cross-platform testing for path handling and OS-specific features

## Important Notes

- **Binary name**: `gz` (not `gzh-cli`)
- **Always run**: `make fmt` before committing code
- **Configuration**: Unified `gzh.yaml` supports all commands
- **Authentication**: Token-based auth for all Git platforms
- **Cross-platform**: Supports Linux, macOS, Windows
- **Modular Makefiles**: Use appropriate `make help-*` commands
- **Mock generation**: Interfaces marked with `//go:generate mockgen`
- **Schema validation**: All configs validated against JSON Schema
- **Performance monitoring**: Built-in benchmarking with `scripts/`

## Command Categories

### Repository Operations

- `gz git repo clone-or-update` - **NEW**: Intelligent single repository management
  - Optional target-path (auto-extracts from URL like `git clone`)
  - Multiple strategies: rebase (default), reset, clone, skip, pull, fetch
  - Branch specification with `-b/--branch` flag
  - Examples: `gz git repo clone-or-update https://github.com/user/repo.git`
- `gz synclone` - Clone entire organizations from GitHub, GitLab, Gitea, Gogs
- `gz repo-config` - GitHub repository configuration management
- `gz actions-policy` - GitHub Actions policy management
- `gz quality` - Code quality checks and improvements
- `gz shell` - Shell integration and automation

### Development Environment

- `gz dev-env` - Manage AWS, Docker, Kubernetes, SSH configurations
- `gz pm` - Update package managers (asdf, Homebrew, SDKMAN, npm, pip, etc.)
- `gz ide` - Monitor JetBrains IDE settings and fix sync issues
- `gz doctor` - Diagnose system health and configuration issues
- `gz profile` - Performance profiling using standard Go pprof (server, cpu, memory, stats)

### Network Management

- `gz net-env` - WiFi change detection, VPN/DNS/proxy management

## Repository Clone Strategies

### Single Repository (gz git repo clone-or-update)

- `rebase` (default): Rebase local changes on top of remote changes
- `reset`: Hard reset to match remote state (discards local changes)
- `clone`: Remove existing directory and perform fresh clone
- `skip`: Leave existing repository unchanged
- `pull`: Standard git pull (merge remote changes)
- `fetch`: Only fetch remote changes without updating working directory

### Bulk Operations (gz synclone)

- `reset` (default): Hard reset + pull (discards local changes)
- `pull`: Merge remote changes with local changes
- `fetch`: Update remote tracking without changing working directory

## Authentication

- Token-based authentication for private repositories
- Environment variable support (GITHUB_TOKEN, GITLAB_TOKEN, etc.)
- SSH key management and configuration
