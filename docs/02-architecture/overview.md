# GZH Manager Architecture Documentation

## Overview

GZH Manager (`gz`) is a comprehensive CLI tool designed for managing development environments and Git repositories across multiple platforms. This document provides a high-level architectural overview of the system design, components, and their interactions.

## Table of Contents

- [System Architecture](#system-architecture)
- [Component Overview](#component-overview)
- [Package Structure](#package-structure)
- [Design Patterns](#design-patterns)
- [Data Flow](#data-flow)
- [Extension Points](#extension-points)
- [Security Architecture](#security-architecture)
- [Performance Considerations](#performance-considerations)

## System Architecture

```mermaid
graph TB
    subgraph "CLI Layer"
        CLI[CLI Interface]
        ROOT[Root Command]
        BULK[Bulk Clone]
        IDE[IDE Management]
        NETENV[Network Environment]
        DEVENV[Development Environment]
    end

    subgraph "Service Layer"
        GITHUB[GitHub Service]
        GITLAB[GitLab Service]
        GITEA[Gitea Service]
        GOGS[Gogs Service]
        CONFIG[Configuration Service]
        ERROR[Error Handling]
    end

    subgraph "Infrastructure Layer"
        FS[File System]
        NET[Network Client]
        GIT[Git Operations]
        CACHE[Caching]
        ASYNC[Async Processing]
    end

    subgraph "External Systems"
        GHAPI[GitHub API]
        GLAPI[GitLab API]
        GTAPI[Gitea API]
        GSAPI[Gogs API]
        LOCAL[Local Git Repos]
    end

    CLI --> ROOT
    ROOT --> BULK
    ROOT --> IDE
    ROOT --> NETENV
    ROOT --> DEVENV

    BULK --> GITHUB
    BULK --> GITLAB
    BULK --> GITEA
    BULK --> GOGS

    GITHUB --> GHAPI
    GITLAB --> GLAPI
    GITEA --> GTAPI
    GOGS --> GSAPI

    GITHUB --> FS
    GITLAB --> FS
    GITEA --> FS
    GOGS --> FS

    GITHUB --> NET
    GITLAB --> NET
    GITEA --> NET
    GOGS --> NET

    FS --> LOCAL
    NET --> GHAPI
    NET --> GLAPI
    NET --> GTAPI
    NET --> GSAPI
```

## Component Overview

### 1. CLI Layer (`cmd/`)

The CLI layer provides the user interface and command-line interface functionality:

- **Root Command**: Main entry point that coordinates all sub-commands
- **Bulk Clone**: Multi-platform repository cloning operations
- **IDE Management**: JetBrains IDE settings synchronization
- **Network Environment**: Network configuration and VPN management
- **Development Environment**: Development tool and environment management

### 2. Service Layer (`pkg/`)

The service layer contains business logic and provider-specific implementations:

- **Git Platform Services**: GitHub, GitLab, Gitea, Gogs integrations
- **Configuration Management**: YAML configuration loading and validation
- **Error Handling**: User-friendly error processing and recovery guidance
- **Async Processing**: Background tasks and concurrent operations
- **Caching**: Performance optimization through intelligent caching

### 3. Infrastructure Layer (`internal/`)

The infrastructure layer provides core utilities and abstractions:

- **File System Operations**: File and directory management
- **HTTP Client**: Network communication abstraction
- **Git Operations**: Git command execution and repository management
- **Testing Utilities**: Test helpers and mock services
- **Utilities**: Common functionality and helper functions

## Package Structure

### Core Packages

```
gzh-cli/
├── cmd/                    # CLI commands and user interface
│   ├── root.go            # Main CLI entry point
│   ├── bulk-clone/        # Repository bulk cloning
│   ├── ide/              # IDE management
│   ├── net-env/          # Network environment
│   └── dev-env/          # Development environment
├── pkg/                   # Public packages (importable)
│   ├── github/           # GitHub API integration
│   ├── gitlab/           # GitLab API integration
│   ├── gitea/            # Gitea API integration
│   ├── bulk-clone/       # Configuration and orchestration
│   ├── errors/           # Error handling and user guidance
│   ├── async/            # Asynchronous processing
│   ├── cache/            # Caching implementations
│   ├── memory/           # Memory management
│   └── recovery/         # Error recovery mechanisms
└── internal/             # Private packages
    ├── filesystem/       # File system abstraction
    ├── httpclient/       # HTTP client abstraction
    ├── git/              # Git operations
    ├── testlib/          # Testing infrastructure
    └── utils/            # Common utilities
```

### Configuration Architecture

```mermaid
graph LR
    subgraph "Configuration Sources"
        ENV[Environment Variables]
        CLI_FLAGS[CLI Flags]
        YAML[YAML Files]
        DEFAULTS[Defaults]
    end

    subgraph "Configuration Loading"
        LOADER[Configuration Loader]
        VALIDATOR[Schema Validator]
        MERGER[Config Merger]
    end

    subgraph "Configuration Usage"
        SERVICES[Service Layer]
        COMMANDS[CLI Commands]
    end

    ENV --> LOADER
    CLI_FLAGS --> LOADER
    YAML --> LOADER
    DEFAULTS --> LOADER

    LOADER --> VALIDATOR
    VALIDATOR --> MERGER
    MERGER --> SERVICES
    MERGER --> COMMANDS
```

## Design Patterns

### 1. Factory Pattern

Used for creating provider-specific clients and services:

```go
// Factory creates provider-specific implementations
type GitProviderFactory interface {
    CreateGitHubClient(config GitHubConfig) GitHubService
    CreateGitLabClient(config GitLabConfig) GitLabService
    CreateGiteaClient(config GiteaConfig) GiteaService
}
```

### 2. Strategy Pattern

Implemented for different cloning strategies:

```go
type CloneStrategy interface {
    Execute(ctx context.Context, repo Repository, target string) error
}

// Strategies: reset, pull, fetch
type ResetStrategy struct{}
type PullStrategy struct{}
type FetchStrategy struct{}
```

### 3. Observer Pattern

Used in the event system:

```go
type EventBus interface {
    Subscribe(eventType string, handler EventHandler)
    Publish(event Event)
}
```

### 4. Builder Pattern

Applied in configuration building and error construction:

```go
type ErrorBuilder interface {
    Message(string) ErrorBuilder
    Description(string) ErrorBuilder
    Suggest(string) ErrorBuilder
    Build() *UserError
}
```

### 5. Facade Pattern

Provides simplified interfaces for complex subsystems:

```go
type BulkCloneFacade interface {
    CloneOrganizations(ctx context.Context, config BulkCloneConfig) error
    ValidateConfiguration(config BulkCloneConfig) error
    GetProgress() ProgressInfo
}
```

## Data Flow

### Repository Cloning Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Config
    participant Provider
    participant Git
    participant FS

    User->>CLI: gz bulk-clone
    CLI->>Config: Load configuration
    Config->>CLI: Return config
    CLI->>Provider: List repositories
    Provider->>Provider: Authenticate
    Provider->>CLI: Return repo list
    CLI->>Git: Clone repositories
    Git->>FS: Write to disk
    FS->>CLI: Confirm success
    CLI->>User: Display results
```

### Error Handling Flow

```mermaid
sequenceDiagram
    participant Operation
    participant ErrorSystem
    participant KnowledgeBase
    participant User

    Operation->>ErrorSystem: System error occurs
    ErrorSystem->>KnowledgeBase: Look up solutions
    KnowledgeBase->>ErrorSystem: Return guidance
    ErrorSystem->>User: Present friendly error
    User->>ErrorSystem: Request more help
    ErrorSystem->>User: Provide detailed steps
```

## Extension Points

### 1. Git Provider Extensions

Add support for new Git hosting services:

```go
type GitProvider interface {
    GetDefaultBranch(ctx context.Context, org, repo string) (string, error)
    List(ctx context.Context, org string) ([]string, error)
    Clone(ctx context.Context, target, org, repo, branch string) error
}
```

### 2. Command Extensions

Extend CLI functionality with new commands:

```go
func NewCustomCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "custom",
        Short: "Custom functionality",
        RunE:  runCustomCommand,
    }
}
```

### 3. Configuration Extensions

Add new configuration sections:

```yaml
custom_provider:
  organizations:
    - name: "custom-org"
      target: "./custom-repos"
  base_url: "https://custom-git.com"
```

## Security Architecture

### Authentication & Authorization

```mermaid
graph TB
    subgraph "Authentication Sources"
        ENV_TOKENS[Environment Variables]
        CONFIG_TOKENS[Configuration Files]
        CLI_TOKENS[CLI Parameters]
    end

    subgraph "Token Management"
        TOKEN_STORE[Token Store]
        TOKEN_VALIDATOR[Token Validator]
        TOKEN_ROTATION[Token Rotation]
    end

    subgraph "API Clients"
        GITHUB_CLIENT[GitHub Client]
        GITLAB_CLIENT[GitLab Client]
        GITEA_CLIENT[Gitea Client]
    end

    ENV_TOKENS --> TOKEN_STORE
    CONFIG_TOKENS --> TOKEN_STORE
    CLI_TOKENS --> TOKEN_STORE

    TOKEN_STORE --> TOKEN_VALIDATOR
    TOKEN_VALIDATOR --> GITHUB_CLIENT
    TOKEN_VALIDATOR --> GITLAB_CLIENT
    TOKEN_VALIDATOR --> GITEA_CLIENT
```

### Security Features

1. **Token Security**
   - Environment variable-based token storage
   - No token persistence in configuration files
   - Token validation before API calls

2. **Input Validation**
   - YAML schema validation
   - URL sanitization
   - Path traversal prevention

3. **Network Security**
   - TLS/HTTPS enforcement
   - Certificate validation
   - Timeout and rate limiting

## Performance Considerations

### Concurrency Model

```mermaid
graph TB
    subgraph "Request Layer"
        USER_REQUEST[User Request]
        RATE_LIMITER[Rate Limiter]
    end

    subgraph "Processing Layer"
        WORKER_POOL[Worker Pool]
        TASK_QUEUE[Task Queue]
        SEMAPHORE[Semaphore]
    end

    subgraph "Resource Layer"
        API_CLIENTS[API Clients]
        CACHE[Cache Layer]
        DISK_IO[Disk I/O]
    end

    USER_REQUEST --> RATE_LIMITER
    RATE_LIMITER --> WORKER_POOL
    WORKER_POOL --> TASK_QUEUE
    TASK_QUEUE --> SEMAPHORE
    SEMAPHORE --> API_CLIENTS
    SEMAPHORE --> CACHE
    SEMAPHORE --> DISK_IO
```

### Optimization Strategies

1. **Concurrency**
   - Worker pools for parallel repository operations
   - Semaphores for resource limiting
   - Context-based cancellation

2. **Caching**
   - LRU cache for API responses
   - Redis support for distributed caching
   - TTL-based cache invalidation

3. **Memory Management**
   - Object pooling for frequent allocations
   - Garbage collection tuning

4. **Network Optimization**
   - Request batching and deduplication
   - HTTP/2 connection reuse
   - Retry mechanisms with exponential backoff

### Metrics Collection

```mermaid
graph LR
    subgraph "Metrics Sources"
        APP_METRICS[Application Metrics]
        SYS_METRICS[System Metrics]
        CUSTOM_METRICS[Custom Metrics]
    end

    subgraph "Collection Layer"
        PROMETHEUS[Prometheus]
        GRAFANA[Grafana]
        ALERTS[Alert Manager]
    end

    subgraph "Storage & Analysis"
        TSDB[Time Series DB]
        NOTIFICATIONS[Notifications]
    end

    APP_METRICS --> PROMETHEUS
    SYS_METRICS --> PROMETHEUS
    CUSTOM_METRICS --> PROMETHEUS

    PROMETHEUS --> GRAFANA
    PROMETHEUS --> ALERTS

    GRAFANA --> TSDB
    ALERTS --> NOTIFICATIONS
```

### Logging Architecture

The logging system has been significantly enhanced with RFC 5424 compliant structured logging and centralized log management:

- **RFC 5424 Compliance**: Standardized log format with severity levels (0-7)
- **Structured Logging**: JSON, logfmt, and console output formats for machine and human processing
- **Centralized Integration**: Seamless bridge between structured and centralized logging systems
- **Dynamic Log Control**: Real-time log level management with rule-based conditional logging
- **Remote Log Shipping**: Support for Elasticsearch, Loki, Fluentd, and HTTP endpoints
- **Performance Optimization**: Async logging, sampling, and buffering for high-throughput scenarios
- **Distributed Tracing**: OpenTelemetry integration with trace and span ID propagation
- **Adaptive Sampling**: Performance-aware log sampling based on system metrics
- **Multi-destination Routing**: Configurable log routing with fallback mechanisms

#### Enhanced Logging Flow

```mermaid
graph TB
    subgraph "Application Layer"
        APP[Application Code]
        MODULE[Module Loggers]
    end

    subgraph "Structured Logging Layer"
        SL[StructuredLogger]
        ESL[EnhancedStructuredLogger]
        BRIDGE[CentralizedBridge]
    end

    subgraph "Log Level Management"
        LLM[LogLevelManager]
        RULES[Dynamic Rules]
        PROFILES[Log Profiles]
        HTTP_API[HTTP Control API]
    end

    subgraph "Centralized Logging Layer"
        CL[CentralizedLogger]
        PROCESSORS[Log Processors]
        OUTPUTS[Output Destinations]
    end

    subgraph "Remote Destinations"
        ES[Elasticsearch]
        LOKI[Grafana Loki]
        FLUENTD[Fluentd]
        HTTP_DEST[HTTP Endpoints]
        WEBSOCKET[WebSocket Streams]
    end

    APP --> MODULE
    MODULE --> SL
    SL --> ESL
    ESL --> BRIDGE
    BRIDGE --> CL

    LLM --> SL
    RULES --> LLM
    PROFILES --> LLM
    HTTP_API --> LLM

    CL --> PROCESSORS
    PROCESSORS --> OUTPUTS
    OUTPUTS --> ES
    OUTPUTS --> LOKI
    OUTPUTS --> FLUENTD
    OUTPUTS --> HTTP_DEST
    OUTPUTS --> WEBSOCKET
```

#### Key Components

1. **StructuredLogger**: RFC 5424 compliant logging with OpenTelemetry integration
2. **EnhancedStructuredLogger**: Structured logger with centralized forwarding capabilities
3. **CentralizedLoggerBridge**: Asynchronous bridge for log forwarding with buffering
4. **LogLevelManager**: Dynamic log level control with rule-based conditions
5. **IntegratedLoggingSetup**: Unified configuration and management for both logging systems

## Architecture Evolution (2025-01 Simplification)

### Removed Components

The architecture was recently simplified to remove over-engineered components inappropriate for CLI tools:

1. **Dependency Injection Container** (`internal/container/`):
   - **Removed**: ~1,188 lines of complex DI container code
   - **Replaced with**: Direct constructor calls in command initialization
   - **Rationale**: CLI tools don't need runtime service discovery

2. **Complex Profiling System** (`internal/profiling/`):
   - **Removed**: Custom HTTP server with multiple abstractions
   - **Replaced with**: Standard Go pprof integration via `internal/simpleprof/`
   - **Rationale**: Standard pprof tooling is more appropriate and familiar

### Current Design Philosophy

- **Simplicity First**: Direct, clear implementations without unnecessary abstractions
- **Standard Tools**: Leverage Go's built-in tooling (pprof, testing, etc.)
- **CLI-Appropriate Patterns**: Design patterns that make sense for command-line tools
- **Performance**: Maintain fast startup times and minimal memory usage

## Development Guidelines

### Code Organization

1. **Package Boundaries**: Clear separation between public (`pkg/`) and private (`internal/`) packages
2. **Interface Design**: Use interfaces for testability and flexibility
3. **Error Handling**: Comprehensive error handling with user-friendly messages
4. **Testing Strategy**: Unit tests, integration tests, and end-to-end tests

### Configuration Management

1. **Schema-Driven**: JSON Schema validation for all configuration
2. **Environment-Aware**: Support for different environments (dev, staging, prod)
3. **Backward Compatibility**: Migration support for configuration changes
4. **Documentation**: Comprehensive configuration documentation and examples

### Release Process

1. **Semantic Versioning**: Following SemVer for version management
2. **Automated Testing**: CI/CD pipeline with comprehensive test coverage
3. **Documentation Updates**: Keep architecture and API documentation current
4. **Migration Guides**: Provide clear upgrade paths for breaking changes

## Future Considerations

### Scalability

- **Horizontal Scaling**: Support for distributed processing
- **Cloud Native**: Kubernetes operator for large-scale deployments
- **API Gateway**: REST API for programmatic access

### Extensibility

- **Plugin System**: Dynamic plugin loading for custom functionality
- **Webhook Support**: Event-driven integrations
- **Custom Providers**: Framework for adding new Git hosting services

### User Experience

- **Web Interface**: Browser-based management console
- **Real-time Updates**: WebSocket-based progress updates

## Technology Stack

### Core Technologies

- **Language**: Go 1.24.0+ (toolchain: go1.24.5)
- **CLI Framework**: Cobra with direct command initialization
- **Configuration**: Viper with YAML/JSON schema validation
- **Git Operations**: go-git v5 with strategy pattern
- **API Clients**:
  - GitHub: google/go-github/v66
  - GitLab: xanzy/go-gitlab
  - Gitea/Gogs: Custom implementations
- **Testing**: testify with gomock for mocking
- **Profiling**: Standard Go pprof (simplified from custom solution)
- **Logging**: Structured logging with RFC 5424 compliance

### Key Dependencies

- **UI/TUI**: Charm libraries (bubbletea, bubbles, lipgloss)
- **File Watching**: fsnotify for IDE monitoring
- **Progress**: schollz/progressbar for visual feedback
- **Validation**: go-playground/validator
- **Schema**: xeipuuv/gojsonschema for configuration validation

---

This architecture documentation provides a comprehensive overview of the GZH Manager system design, reflecting the simplified architecture adopted in January 2025. For detailed implementation information, refer to the package-specific documentation and code comments.
