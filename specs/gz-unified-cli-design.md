<!-- 🚫 AI_MODIFY_PROHIBITED -->
# GZ Unified CLI Design

A comprehensive design for unified Git platform management through the `gz` CLI tool, combining Git extensions and platform management capabilities.

## Overview

This document consolidates the design for a unified Git platform management system that provides:

1. **Git Extensions** - Native Git integration for core repository operations
2. **Platform Management** - Unified interface for cross-provider operations
3. **Consistent Experience** - Coherent CLI design across all functionalities

### Design Philosophy

- **Clear Separation**: Git extensions enhance Git core functionality; platform management handles DevOps operations
- **Unified Interface**: Consistent command patterns and user experience
- **Provider Agnostic**: Abstract provider differences through common interfaces
- **Incremental Development**: Leverage existing codebase with gradual enhancement
- **Practical Implementation**: Focus on real-world use cases and implementation feasibility

## Architecture Overview

### Two-Tier Command Structure

```
┌─────────────────────────────────────────────────────────────┐
│                     GZ CLI ECOSYSTEM                       │
├─────────────────────────────────────────────────────────────┤
│  Git Extensions          │  Platform Management             │
│  ├─ git synclone         │  ├─ gz git repo [action]          │
│  ├─ git remote-sync      │  ├─ gz git config [action]        │
│  └─ git provider-auth    │  ├─ gz git webhook [action]       │
│                          │  ├─ gz git event [action]         │
│                          │  ├─ gz git auth [action]          │
│                          │  └─ gz git sync [action]          │
└─────────────────────────────────────────────────────────────┘
```

### Provider Abstraction Layer

```go
// GitProvider defines the unified interface for all Git platforms
type GitProvider interface {
    // Core Operations
    GetName() string
    GetCapabilities() []Capability
    Authenticate(ctx context.Context, creds Credentials) error

    // Repository Operations
    CloneRepositories(ctx context.Context, target CloneTarget) error
    ListRepositories(ctx context.Context, filter RepoFilter) ([]Repository, error)
    GetRepository(ctx context.Context, id string) (*Repository, error)
    SyncRepository(ctx context.Context, source, target RepositoryRef) error

    // Configuration Management
    ApplyConfiguration(ctx context.Context, config RepoConfig) error
    GetConfiguration(ctx context.Context, repoID string) (*RepoConfig, error)
    ValidateConfiguration(config RepoConfig) error

    // Webhook Management
    CreateWebhook(ctx context.Context, webhook Webhook) (*Webhook, error)
    ListWebhooks(ctx context.Context, filter WebhookFilter) ([]Webhook, error)
    UpdateWebhook(ctx context.Context, webhook Webhook) error
    DeleteWebhook(ctx context.Context, id string) error

    // Event Handling
    GetEvents(ctx context.Context, filter EventFilter) ([]Event, error)
    HandleWebhookEvent(ctx context.Context, event WebhookEvent) error
    StreamEvents(ctx context.Context, filter EventFilter) (<-chan Event, error)
}

// Common types across all providers
type Repository struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Owner       string                 `json:"owner"`
    CloneURL    string                 `json:"clone_url"`
    Private     bool                   `json:"private"`
    Archived    bool                   `json:"archived"`
    Provider    string                 `json:"provider"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type Capability string

const (
    CapabilityWebhooks      Capability = "webhooks"
    CapabilityEvents        Capability = "events"
    CapabilityConfiguration Capability = "configuration"
    CapabilityBulkOps       Capability = "bulk_operations"
    CapabilityAuthentication Capability = "authentication"
)
```

## Git Extensions

### git synclone - Intelligent Repository Cloning

**Purpose**: Enhanced Git cloning with provider awareness and advanced features.

**Integration**: Native Git command that appears as `git synclone` in user's PATH.

**Core Features**:
- Bulk cloning from multiple providers
- Parallel execution with configurable concurrency
- Resume capability for interrupted operations
- Automatic repository organization
- Orphan directory cleanup and validation

**Usage Examples**:
```bash
# Clone all repositories from an organization
git synclone github --org myorg --target ~/repos

# Clone with advanced filtering
git synclone gitlab --group mygroup --filter "name:api-*" --archived=false

# Resume interrupted cloning session
git synclone --resume --session-id abc123

# Clone from multiple providers with unified config
git synclone --all-providers --config synclone.yaml
```

**Implementation Structure**:
```
cmd/git-synclone/
├── main.go              # Entry point and CLI setup
├── providers/           # Provider-specific implementations
├── clone/               # Core cloning logic and parallelization
├── session/             # Resume capability and state management
├── config/              # Configuration loading and validation
└── filters/             # Repository filtering and selection
```

### Future Git Extensions

**git remote-sync**: Synchronize remote configurations across providers
**git provider-auth**: Unified Git credential helper for multiple providers

## Platform Management Commands

### Unified Command Interface

**Structure**: `gz git [resource] [action] --provider [provider] [options]`

**Core Resources**:
- **repo** - Repository lifecycle management
- **config** - Configuration and compliance management
- **webhook** - Webhook lifecycle and monitoring
- **event** - Event processing and analytics
- **auth** - Authentication and credential management
- **sync** - Cross-provider synchronization

**New Unified Commands**:
- All functionality consolidated under `gz git [resource]` structure
- Clean, consistent interface across all Git platform operations

### Repository Management (`gz git repo`)

**Purpose**: Comprehensive repository operations across all providers.

**Key Operations**:
```bash
# Repository Discovery and Cloning
gz git repo clone --provider github --org myorg --target /repos
gz git repo list --provider gitlab --group mygroup --format table
gz git repo search --query "language:go stars:>100" --all-providers

# Repository Lifecycle
gz git repo create --provider github --org myorg --name newproject --template api-template
gz git repo archive --provider gitlab --group mygroup --pattern "legacy-*"
gz git repo delete --provider github --repo myorg/deprecated --confirm

# Bulk Operations
gz git repo sync --all --dry-run
gz git repo migrate --from github:oldorg --to gitlab:newgroup --include-issues
```

### Configuration Management (`gz git config`)

**Purpose**: Cross-provider repository configuration with enterprise compliance.

**Key Features**:
- Template-based configuration application
- Policy validation and compliance checking
- Configuration diffing and auditing
- Bulk configuration operations

**Operations**:
```bash
# Configuration Application
gz git config apply --template security.yaml --provider github --org myorg
gz git config apply --template compliance.yaml --all-repos --dry-run

# Configuration Analysis
gz git config diff --repo1 github:org/repo --repo2 gitlab:group/repo
gz git config export --provider github --org myorg --output current.yaml
gz git config validate --policy company-policy.yaml --all-repos

# Compliance and Auditing
gz git config audit --provider github --org myorg --policy security-baseline
gz git config report --format html --output compliance-report.html
```

### Webhook Management (`gz git webhook`)

**Purpose**: Unified webhook management with advanced routing and monitoring.

**Key Features**:
- Provider-agnostic webhook creation
- Webhook health monitoring and alerting
- Event routing and transformation
- Bulk webhook operations

**Operations**:
```bash
# Webhook Lifecycle
gz git webhook create --url https://ci.company.com/hook --events push,pr --all-repos
gz git webhook list --provider gitlab --group mygroup --format table
gz git webhook update --id 123 --events push,issue --provider github

# Advanced Features
gz git webhook forward --from gitlab:project --to github:org/repo
gz git webhook test --provider github --repo org/repo --event push
gz git webhook monitor --dashboard --port 9090 --metrics-endpoint /metrics
```

### Event Management (`gz git event`)

**Purpose**: Unified event monitoring, processing, and analytics across providers.

**Key Features**:
- Real-time event streaming with filtering
- Historical event querying and analysis
- Cross-provider event correlation
- Event replay for testing and debugging

**Operations**:
```bash
# Event Monitoring
gz git event server --port 8080 --providers github,gitlab --dashboard
gz git event stream --filter "repo:critical-*" --output json --follow
gz git event tail --provider github --org myorg --events push,pr

# Event Analysis
gz git event query --since 2024-01-01 --type pull_request --action closed
gz git event analytics --provider github --org myorg --period month --format chart
gz git event correlation --events "push,deployment" --timeframe 1h

# Event Replay and Testing
gz git event replay --from 2024-01-01T10:00:00 --to 2024-01-01T11:00:00
gz git event simulate --type push --repo github:org/repo --count 10
```

### Authentication Management (`gz git auth`)

**Purpose**: Unified authentication and credential management across providers.

**Key Features**:
- Secure credential storage and retrieval
- Automatic token rotation with validation
- SSO integration and enterprise authentication
- Multi-factor authentication support

**Operations**:
```bash
# Credential Management
gz git auth add github --token ${GITHUB_TOKEN} --scope repo,admin:org
gz git auth add gitlab --oauth --client-id ${ID} --client-secret ${SECRET}
gz git auth add gitea --token ${GITEA_TOKEN} --endpoint https://gitea.company.com

# Authentication Operations
gz git auth validate --all-providers --verbose
gz git auth rotate --provider github --auto-update-webhooks
gz git auth test --provider gitlab --endpoint https://gitlab.company.com
gz git auth list --show-scopes --mask-tokens
```

### Cross-Provider Synchronization (`gz git sync`)

**Purpose**: Synchronize repositories, configurations, and metadata across providers.

**Key Features**:
- Bidirectional synchronization with conflict resolution
- Selective sync (branches, tags, issues, metadata)
- Scheduled synchronization with monitoring
- Migration capabilities with full history preservation

**Operations**:
```bash
# Repository Synchronization
gz git sync --from github:org/repo --to gitlab:group/repo --include-issues
gz git sync watch --config sync-rules.yaml --dashboard --port 8080
gz git sync status --all --format table

# Migration Operations
gz git sync migrate --source github:old/repo --target gitea:new/repo --include-all
gz git sync validate --config migration.yaml --dry-run --report migration-plan.html

# Scheduled Operations
gz git sync schedule --cron "0 2 * * *" --config nightly-sync.yaml
gz git sync jobs list --status active --format table
```

## Implementation Strategy

### Phase 1: Foundation (Weeks 1-2)
**Objective**: Establish core architecture and provider abstraction

**Deliverables**:
1. Provider interface definition and common types
2. Provider factory and registry implementation
3. Configuration system for multi-provider setup
4. Basic CLI framework with cobra integration

**File Structure**:
```
cmd/
└── git/
    ├── main.go              # Main CLI entry point
    ├── common/
    │   ├── flags.go         # Shared CLI flags and validation
    │   ├── types.go         # Common data structures
    │   ├── config.go        # Configuration loading and validation
    │   └── output.go        # Output formatting (table, json, yaml)
    ├── providers/
    │   ├── interface.go     # GitProvider interface definition
    │   ├── factory.go       # Provider factory and registry
    │   ├── github/          # GitHub provider implementation
    │   ├── gitlab/          # GitLab provider implementation
    │   └── gitea/           # Gitea provider implementation
    └── commands/
        └── root.go          # Root command setup
```

### Phase 2: Core Repository Operations (Weeks 3-4)
**Objective**: Implement repository management with existing synclone integration

**Deliverables**:
1. `gz git repo` command family implementation
2. Integration with existing synclone logic
3. Basic provider support (GitHub priority)
4. Configuration-driven operations

**Implementation Focus**:
- Leverage existing `pkg/synclone` and `pkg/github` packages
- Adapter pattern for existing implementations
- Unified configuration schema
- Error handling and user feedback

### Phase 3: Configuration and Webhook Management (Weeks 5-6)
**Objective**: Implement configuration and webhook management features

**Deliverables**:
1. `gz git config` command family
2. `gz git webhook` command family
3. Template-based configuration system
4. Policy validation framework

**Integration Points**:
- Extend existing `cmd/repo-config` functionality
- Integrate with existing webhook management
- Cross-provider configuration templates
- Compliance and auditing features

### Phase 4: Events and Monitoring (Weeks 7-8)
**Objective**: Implement event processing and monitoring capabilities

**Deliverables**:
1. `gz git event` command family
2. Real-time event streaming
3. Event analytics and reporting
4. Monitoring dashboard integration

**Technical Components**:
- Event server with WebSocket support
- Event storage and querying
- Analytics engine for event correlation
- Dashboard for real-time monitoring

### Phase 5: Authentication and Sync (Weeks 9-10)
**Objective**: Complete the platform management suite

**Deliverables**:
1. `gz git auth` command family
2. `gz git sync` command family
3. Cross-provider operations
4. Advanced synchronization features

**Advanced Features**:
- Secure credential management
- Token rotation automation
- Bidirectional sync with conflict resolution
- Migration utilities

### Phase 6: Migration and Documentation (Weeks 11-12)
**Objective**: Smooth transition from existing commands

**Deliverables**:
1. Migration utilities and documentation
2. Deprecation notices for old commands
3. Comprehensive user documentation
4. Performance optimization and testing

## Migration and Compatibility Strategy

### Backward Compatibility Approach

**Existing Command Preservation**:
```bash
# Existing commands continue to work with deprecation warnings
gz synclone github -o myorg -t /repos
# Warning: This command is deprecated. Use: gz git repo clone --provider github --org myorg --target /repos

gz repo-config apply --org myorg --template security.yaml
# Warning: This command is deprecated. Use: gz git config apply --provider github --org myorg --template security.yaml
```

**Command Consolidation**:
```bash
# Old commands → New unified structure
gz webhook → REMOVED (replaced by gz git webhook)
gz event → REMOVED (replaced by gz git event)
gz synclone → gz git repo clone (unified repository management)
gz repo-config → gz git config (extended configuration management)
```

**Migration Tools**:
```bash
# Configuration migration utility
gz migrate config --from-legacy --output unified-config.yaml

# Command migration helper
gz migrate commands --show-mapping --format table
```

### Phased Deprecation Timeline

**Phase 1 (Months 1-3)**: Soft deprecation
- Add deprecation warnings to existing commands
- Provide migration suggestions
- Both old and new commands work

**Phase 2 (Months 4-6)**: Documentation migration
- Update all documentation to use new commands
- Provide migration guides and examples
- Training materials for new interface

**Phase 3 (Months 7-12)**: Hard deprecation
- Remove old commands from main help
- Require explicit flag to use deprecated commands
- Clear timeline for complete removal

## Security and Configuration

### Security Framework

**Credential Management**:
- OS keyring integration for secure storage
- Token scoping and permission validation
- Automatic token rotation with webhook updates
- Audit logging for all authentication events

**Network Security**:
- TLS certificate validation and pinning
- Proxy support for enterprise environments
- Rate limiting and backoff strategies
- Network timeout and retry policies

**Access Control**:
- Role-based access control (RBAC) support
- Provider-specific permission validation
- Operation approval workflows for sensitive actions
- Audit trails for compliance requirements

### Configuration System

**Unified Configuration Schema**:
```yaml
# ~/.config/gz/config.yaml
providers:
  github:
    token: ${GITHUB_TOKEN}
    endpoint: https://api.github.com
    org: myorg
    default_clone_target: ~/repos/github
  gitlab:
    token: ${GITLAB_TOKEN}
    endpoint: https://gitlab.com/api/v4
    group: mygroup
    default_clone_target: ~/repos/gitlab
  gitea:
    token: ${GITEA_TOKEN}
    endpoint: https://gitea.example.com/api/v1
    org: myorg
    default_clone_target: ~/repos/gitea

defaults:
  parallel: 10
  timeout: 30s
  retry: 3
  output_format: table

operations:
  clone:
    parallel: 5
    resume: true
    cleanup_orphans: true
  config:
    dry_run_default: true
    backup_before_apply: true
  webhook:
    health_check_interval: 5m
    retry_failed_deliveries: true

monitoring:
  enabled: true
  metrics_port: 9090
  log_level: info
  dashboard_port: 8080
```

**Environment Integration**:
- Environment variable support for all configurations
- Multiple configuration file locations
- Configuration validation and schema checking
- Dynamic configuration reloading

## Installation and Distribution

### 1. Core CLI Installation
```bash
# Install gz CLI tool
go install github.com/gizzahub/gzh-manager-go/cmd/gz@latest

# Or via package managers
brew install gizzahub/tap/gz
apt-get install gz-cli
yum install gz-cli
```

### 2. Git Extensions (Separate Installation)
```bash
# Install specific Git extension
go install github.com/gizzahub/gzh-manager-go/cmd/git-synclone@latest

# Install all Git extensions
curl -sSL https://install.gizzahub.com/git-extensions.sh | bash
```

### 3. Shell Completion (One-time Script Installation)

**Rationale**: Shell completion is installed via lightweight shell scripts:
- No additional binaries required
- One-time installation process
- Supports all major shells independently
- Easy uninstallation and updates

**Installation Options**:
```bash
# Built-in completion generation (from main gz binary)
gz completion bash > /etc/bash_completion.d/gz
gz completion zsh > "${fpath[1]}/_gz"
gz completion fish > ~/.config/fish/completions/gz.fish
gz completion powershell > ~/.config/powershell/Microsoft.PowerShell_profile.ps1

# Automatic detection and installation
gz completion install --auto-detect

# Verify completion installation
gz completion test
```

**One-time Script Installation**:
```bash
# Standalone completion installer (shell script only)
curl -sSL https://install.gizzahub.com/completion.sh | bash

# Manual installation via script
wget -qO- https://install.gizzahub.com/completion.sh | bash -s -- --shell bash
wget -qO- https://install.gizzahub.com/completion.sh | bash -s -- --shell zsh

# Package managers include completion in main package
brew install gizzahub/tap/gz  # completion included
apt-get install gz-cli        # completion included
yum install gz-cli           # completion included

# Uninstallation
curl -sSL https://install.gizzahub.com/completion.sh | bash -s -- --uninstall
```

**Completion Features**:
- Context-aware command completion
- Provider name completion (github, gitlab, gitea)
- Repository name completion from configured providers
- Configuration file path completion
- Dynamic option completion based on provider capabilities

### 4. Development Environment Setup
```bash
# For developers contributing to gz
git clone https://github.com/gizzahub/gzh-manager-go
cd gzh-manager-go
make bootstrap  # Install development dependencies
make build      # Build local binary
make install    # Install to GOPATH/bin

# Development with completion
make completion-dev  # Generate and install development completion
```

### 5. Docker and Container Support
```bash
# Official Docker images
docker pull gizzahub/gz:latest
docker pull gizzahub/gz:v1.2.3

# Run in container with completion support
docker run -it \
  -v ~/.config/gz:/root/.config/gz \
  -v ~/.gitconfig:/root/.gitconfig \
  gizzahub/gz:latest

# Kubernetes deployment
kubectl apply -f https://install.gizzahub.com/k8s/gz-deployment.yaml
```

## Future Enhancements

### Advanced Features

**Multi-Provider Operations**:
```bash
# Operate across all configured providers
gz git repo sync --all-providers
gz git config audit --compare-providers
gz git event correlate --cross-provider
```

**Plugin System**:
```bash
# Third-party provider plugins
gz git provider install custom-gitea-plugin
gz git provider list --available
```

**Interactive Mode**:
```bash
# Terminal UI for complex operations
gz git interactive
gz git wizard setup-new-organization
```

**Automation and Scripting**:
```bash
# Workflow automation
gz git script run migration-workflow.yaml
gz git template apply --workflow ci-cd-setup
```

**AI-Powered Features**:
```bash
# Intelligent suggestions and automation
gz git suggest --analyze-repos --recommend-configs
gz git analyze --security-issues --compliance-gaps
```

## Benefits and Value Proposition

### For Development Teams

1. **Unified Experience**: Single interface for all Git platform operations
2. **Reduced Context Switching**: Consistent commands across platforms
3. **Improved Productivity**: Bulk operations and automation capabilities
4. **Better Compliance**: Automated policy validation and auditing

### For DevOps Teams

1. **Infrastructure as Code**: Configuration templates and automation
2. **Monitoring and Analytics**: Real-time insights across platforms
3. **Migration Support**: Seamless platform transitions
4. **Security Integration**: Unified authentication and compliance

### For Organizations

1. **Vendor Independence**: Avoid lock-in with provider abstraction
2. **Cost Optimization**: Efficient multi-provider management
3. **Risk Reduction**: Automated compliance and security policies
4. **Scalability**: Handle large-scale operations efficiently

## Conclusion

The GZ Unified CLI Design provides a comprehensive solution for Git platform management that:

- **Consolidates** existing functionality into a coherent system
- **Extends** capabilities through unified provider interfaces
- **Simplifies** operations with consistent command patterns
- **Enables** advanced features like cross-provider operations
- **Maintains** backward compatibility during migration

This design leverages existing codebase investments while providing a clear path forward for enhanced Git platform management capabilities. The phased implementation approach ensures practical development while delivering incremental value to users.

## Related Documentation

- [Git Extension Commands](git-extension-commands.md) - Git-specific extension details
- [Synclone Configuration](synclone.md) - Repository cloning configuration
- [Package Manager Integration](package-manager.md) - Development environment integration
- [Network Environment Management](net-env.md) - Network configuration management
