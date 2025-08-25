<!-- 🚫 AI_MODIFY_PROHIBITED -->

<!-- This file should not be modified by AI agents -->

# Git Extension Commands Design

This document outlines the design and implementation of Git extension commands that enhance Git's native functionality with cross-platform capabilities.

## Current Implementation Status

The `synclone` command has evolved significantly from a simple cloning tool to a comprehensive repository management system. This document describes the Git extension pattern (`git synclone`) that provides seamless integration with Git's CLI.

## Overview

Git extension commands are custom commands that integrate seamlessly with Git's command-line interface. They appear as subcommands of git (e.g., `git synclone`) and provide enhanced functionality for managing repositories across multiple Git providers.

## How Git Extensions Work

Git automatically recognizes any executable in your PATH that follows the naming pattern `git-{command}`. When you type `git {command}`, Git will execute the corresponding `git-{command}` executable.

```bash
# When you type:
git synclone

# Git looks for and executes:
git-synclone
```

## Git Extension Command

### git synclone

**Purpose**: Intelligent repository cloning with provider awareness and advanced features.

**Features**:

- Bulk cloning from multiple providers (GitHub, GitLab, Gitea, Gogs)
- Parallel execution with configurable concurrency
- Resume capability for interrupted clones
- Automatic organization of cloned repositories
- Orphan directory cleanup
- Configuration-driven operations
- State management and operation tracking
- Multi-strategy cloning (reset, pull, fetch)
- Advanced filtering and pattern matching

**Usage**:

```bash
# Basic cloning operations
git synclone github --org myorg --target ~/repos
git synclone gitlab --group mygroup --recursive
git synclone gitea --org myorg --api-url https://gitea.example.com

# Configuration-based operations
git synclone --config synclone.yaml
git synclone --config synclone.yaml --strategy pull
git synclone --config synclone.yaml --cleanup-orphans

# Resume and state management
git synclone --resume
git synclone state list
git synclone state show --last

# Configuration management
git synclone config generate init
git synclone config validate --strict
git synclone config convert --from yaml --to json

# Advanced filtering
git synclone github --org myorg --match ".*-api" --visibility private
git synclone gitlab --group mygroup --exclude ".*-deprecated$"
```

**Implementation Structure**:

```
git-synclone
├── main.go              # Entry point
├── cmd/
│   ├── root.go          # Main command setup
│   ├── github.go        # GitHub-specific commands
│   ├── gitlab.go        # GitLab-specific commands
│   ├── gitea.go         # Gitea-specific commands
│   ├── config.go        # Configuration management commands
│   └── state.go         # State management commands
├── pkg/
│   ├── synclone/        # Core synclone logic
│   ├── github/          # GitHub API integration
│   ├── gitlab/          # GitLab API integration
│   └── gitea/           # Gitea API integration
└── internal/
    ├── config/          # Configuration handling
    ├── state/           # State management
    └── git/             # Git operations
```

## Subcommands and Features

### Configuration Management (`git synclone config`)

Manage synclone configuration files:

```bash
# Generate new configuration
git synclone config generate init
git synclone config generate template --template enterprise
git synclone config generate discover --path ~/repos

# Validate configuration
git synclone config validate --config synclone.yaml
git synclone config validate --strict --check-tokens

# Convert between formats
git synclone config convert --from synclone.yaml --to synclone.json
git synclone config convert --from synclone.yaml --format gzh
```

### State Management (`git synclone state`)

Track and manage clone operations:

```bash
# List operations
git synclone state list
git synclone state list --active
git synclone state list --failed

# Show operation details
git synclone state show <state-id>
git synclone state show --last

# Clean up state files
git synclone state clean --age 7d
git synclone state clean --failed
git synclone state clean --id <state-id>
```

## Related Platform Management Features

The following features are more suitable as standalone platform management tools rather than Git extensions. See [`gz-unified-cli-design.md`](gz-unified-cli-design.md) for comprehensive platform management design:

- **Repository Configuration Management** (`gz repo-config`) - Cross-provider repository settings
- **Webhook Management** (`gz webhook`) - Unified webhook operations
- **Event Monitoring** (`gz event`) - Real-time event streaming and analytics
- **Authentication Management** (`gz auth`) - Unified credential management

## Implementation Guidelines

### 1. Command Structure

Each extension should follow this structure:

```go
package main

import (
    "github.com/spf13/cobra"
    "github.com/gizzahub/gzh-git/pkg/providers"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "git-{command}",
        Short: "Brief description",
        Long:  "Detailed description",
    }

    // Add subcommands
    rootCmd.AddCommand(
        newGitHubCmd(),
        newGitLabCmd(),
        newGiteaCmd(),
        newAllCmd(),
    )

    rootCmd.Execute()
}
```

### 2. Provider Integration

Extensions should use the common provider interface:

```go
type ProviderRegistry struct {
    providers map[string]providers.GitProvider
}

func (r *ProviderRegistry) Execute(providerName string, fn func(p GitProvider) error) error {
    provider, exists := r.providers[providerName]
    if !exists {
        return fmt.Errorf("provider %s not configured", providerName)
    }
    return fn(provider)
}
```

### 3. Configuration

The git extension uses the same configuration format as the standalone `gz synclone` command:

```yaml
version: "1.0.0"
default_provider: github

# Sync mode for subsequent operations
sync_mode:
  cleanup_orphans: true
  conflict_resolution: "remote-overwrite"

# Global settings
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset
  default_visibility: all
  default_protocol: https
  global_ignores:
    - "^test-.*"
    - ".*-archive$"
  timeouts:
    http_timeout: 30s
    git_timeout: 5m
  concurrency:
    clone_workers: 10
    update_workers: 15

# Provider configurations
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "mycompany"
        clone_dir: "$HOME/work/mycompany"
        visibility: all
        strategy: reset
        protocol: ssh
        branch: main
        exclude:
          - ".*-archive$"
    settings:
      rate_limit:
        requests_per_hour: 5000
        burst_limit: 50

  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.com"
    groups:
      - name: "backend-team"
        clone_dir: "$HOME/work/gitlab-internal"
        recursive: true

  gitea:
    token: "${GITEA_TOKEN}"
    api_url: "https://gitea.com"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/gitea/myorg"
```

### 4. Error Handling

Consistent error messages and exit codes:

```go
const (
    ExitSuccess = 0
    ExitError = 1
    ExitAuthError = 2
    ExitNetworkError = 3
    ExitConfigError = 4
)
```

## Installation and Distribution

### 1. Individual Commands

```bash
# Install specific extension
go install github.com/gizzahub/gzh-git/cmd/git-synclone@latest
```

### 2. All Extensions

```bash
# Install all extensions
curl -sSL https://gizzahub.com/install.sh | bash
```

### 3. Package Managers

```bash
# Homebrew
brew install gizzahub/tap/git-extensions

# APT
apt-get install git-gzh-extensions

# YUM
yum install git-gzh-extensions
```

## Integration with gz Command

The git extension provides identical functionality to the main `gz synclone` command:

```bash
# Git extension usage
git synclone github --org myorg
git synclone config validate
git synclone state list

# Equivalent gz command usage
gz synclone github --org myorg
gz synclone config validate
gz synclone state list
```

Both commands share:

- Same configuration files and format
- Same state management system
- Same provider integrations
- Same command-line options

## Advanced Features

### Clone Strategies

- **reset** (default): Hard reset + pull (discards local changes)
- **pull**: Merge remote changes with local changes
- **fetch**: Update remote tracking without changing working directory

### Conflict Resolution Modes

- **remote-overwrite**: Hard reset to remote state (default)
- **local-preserve**: Keep local changes
- **rebase-attempt**: Try rebase, leave conflicts for manual resolution
- **conflict-skip**: Skip repositories with conflicts

### Performance Optimization

- Parallel cloning with configurable worker limits
- Rate limiting with automatic detection
- Incremental updates for large organizations
- Resume capability for interrupted operations

## Security Considerations

1. **Credential Security**: Never store credentials in plain text
1. **Token Scopes**: Request minimum required permissions
1. **Audit Logging**: Log all operations for compliance
1. **Network Security**: Support proxy and TLS configurations
1. **Input Validation**: Sanitize all user inputs

## Platform-Specific Features

### GitHub

- Organization and user repository cloning
- Visibility filtering (public/private/all)
- Branch selection and shallow cloning
- GitHub Enterprise support

### GitLab

- Group and subgroup cloning
- Self-hosted GitLab instance support
- Recursive group traversal

### Gitea

- Organization cloning
- Custom Gitea instance support
- Full API compatibility

## Future Git Extensions

Potential future Git extensions that enhance core Git functionality:

- `git remote-sync`: Synchronize remote configurations across multiple providers
- `git multi-clone`: Enhanced cloning with Git-specific features (submodules, LFS, etc.)
- `git provider-auth`: Git credential helper for multiple providers
- `git bulk-update`: Mass update operations across repositories
- `git cross-provider`: Cross-provider repository migration
