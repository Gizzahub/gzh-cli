# Git Extension Commands Design

This document outlines the design and implementation of Git extension commands that enhance Git's native functionality with cross-platform capabilities.

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
- Bulk cloning from multiple providers
- Parallel execution for performance
- Resume capability for interrupted clones
- Automatic organization of cloned repositories
- Orphan directory cleanup

**Usage**:
```bash
# Clone all repositories from an organization
git synclone github --org myorg --target ~/repos

# Clone from multiple providers
git synclone --all-providers --config synclone.yaml

# Resume interrupted cloning
git synclone --resume --session-id abc123

# Clone with filters
git synclone gitlab --group mygroup --filter "name:api-*" --archived=false
```

**Implementation Structure**:
```
git-synclone
├── main.go              # Entry point
├── providers/           # Provider-specific implementations
├── clone/               # Core cloning logic
├── session/             # Resume capability
└── config/              # Configuration handling
```

## Related Platform Management Features

The following features are more suitable as standalone platform management tools rather than Git extensions. See [`gz-unified-cli-design.md`](gz-unified-cli-design.md) for comprehensive platform management design:

- **Repository Configuration Management** (`gz config`) - Cross-provider repository settings
- **Webhook Management** (`gz webhook`) - Unified webhook operations
- **Event Monitoring** (`gz event`) - Real-time event streaming and analytics
- **Repository Synchronization** (`gz sync`) - Cross-provider repository sync
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

Use consistent configuration format:
```yaml
providers:
  github:
    token: ${GITHUB_TOKEN}
    endpoint: https://api.github.com
  gitlab:
    token: ${GITLAB_TOKEN}
    endpoint: https://gitlab.com/api/v4
  gitea:
    token: ${GITEA_TOKEN}
    endpoint: https://gitea.example.com/api/v1

defaults:
  parallel: 10
  timeout: 30s
  retry: 3
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

## Integration with Core gzh-git

While these commands work standalone, they integrate with the main gzh-git command:

```bash
# Standalone
git synclone github --org myorg

# Via gzh-git
gz-git extensions exec synclone github --org myorg

# Extension management
gz-git extensions list
gz-git extensions update
gz-git extensions config
```

## Security Considerations

1. **Credential Security**: Never store credentials in plain text
2. **Token Scopes**: Request minimum required permissions
3. **Audit Logging**: Log all operations for compliance
4. **Network Security**: Support proxy and TLS configurations
5. **Input Validation**: Sanitize all user inputs

## Future Git Extensions

Potential future Git extensions that enhance core Git functionality:
- `git remote-sync`: Synchronize remote configurations across multiple providers
- `git multi-clone`: Enhanced cloning with Git-specific features (submodules, LFS, etc.)
- `git provider-auth`: Git credential helper for multiple providers
