# Current Command Structure

## Overview

The gzh-manager-go CLI tool (`gz`) provides a comprehensive set of commands for managing development environments, repositories, and system configurations.

## Command List

### Core Commands

1. **synclone** - Multi-platform repository cloning and synchronization
   - Supports GitHub, GitLab, Gitea, and Gogs
   - Bulk operations for organizations and groups
   - Configuration management with subcommands

2. **dev-env** - Development environment configuration management
   - AWS, GCP, Azure cloud configurations
   - Docker and Kubernetes settings
   - SSH configuration management

3. **net-env** - Network environment transitions and monitoring
   - WiFi network detection and actions
   - VPN connection management
   - DNS and proxy configuration
   - Container network monitoring

4. **repo-config** - GitHub repository configuration management
   - Apply organization-wide policies
   - Audit repository settings
   - Manage webhooks and security settings

### Tool Commands

5. **ide** - JetBrains IDE settings monitoring and sync
   - Monitor configuration changes
   - Fix sync issues
   - Backup and restore settings

6. **pm** (formerly always-latest) - Package manager updates
   - Homebrew, asdf, SDKMAN management
   - npm, pip, gem package updates
   - Coordinated version management

7. **doctor** - System health and configuration validation
   - Check system dependencies
   - Validate configurations
   - Diagnose common issues

### Repository Management

8. **event** - GitHub event management
   - Monitor repository events
   - Event filtering and processing
   - Webhook event handling

9. **webhook** - GitHub webhook management
   - Create and manage webhooks
   - Update webhook configurations
   - List and delete webhooks

### Utility Commands

10. **version** - Display version information
11. **help** - Show help for commands
12. **man** - Display manual pages

### Special Commands

13. **shell** - Interactive shell mode (debug feature)
    - Available via `--debug-shell` flag
    - For development and debugging

## Command Structure Pattern

Most commands follow a consistent pattern:

```
gz <command> [subcommand] [flags]
```

### Common Subcommands

- `config` - Configuration management for the specific command
- `validate` - Validate settings and configurations
- `status` - Show current status
- `list` - List available items
- `apply` - Apply configurations or changes

### Global Flags

- `--config` - Specify configuration file
- `--verbose` - Enable verbose output
- `--debug` - Enable debug mode
- `--help` - Show help for any command

## Configuration

All commands can be configured through:

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration files (`gzh.yaml`)
4. Default values (lowest priority)

## Command Categories

### Repository Operations
- synclone - Clone and manage repositories
- repo-config - Configure repository settings
- event - Monitor repository events
- webhook - Manage webhooks

### Development Environment
- dev-env - Manage development configurations
- pm - Package manager operations
- ide - IDE settings management

### Network Management
- net-env - Network environment transitions

### System Utilities
- doctor - System diagnostics
- version - Version information
- help - Command help
- man - Manual pages

## Integration

Commands are designed to work together:

- `synclone` uses network settings from `net-env`
- `dev-env` coordinates with `pm` for package management
- `repo-config` works with `webhook` and `event`
- `doctor` validates all command configurations