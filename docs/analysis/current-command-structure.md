# Current Command Structure

**Last Updated**: 2025-01-27
**Status**: Reflects actual implementation based on code analysis

## Overview

The gzh-cli CLI tool (`gz`) provides a comprehensive set of commands for managing development environments, repositories, and system configurations.

## Command List

### Core Commands

1. **synclone** - Multi-platform repository cloning and synchronization
   - Supports GitHub, GitLab, Gitea, and Gogs
   - Bulk operations for organizations and groups
   - Configuration management with subcommands:
     - `config generate` - Generate configuration files
     - `config validate` - Validate configuration syntax
     - `config convert` - Convert between formats
   - State management with subcommands:
     - `state list` - List tracked operations
     - `state show` - Show operation details
     - `state clean` - Clean up state files

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

6. **pm** - Unified package manager interface
   - Core unified commands: `status`, `install`, `update`, `sync`, `export`, `validate`, `clean`, `bootstrap`
   - Legacy direct access commands: `brew`, `asdf`, `sdkman`, `apt`, `port`, `rbenv`, `pip`, `npm`
   - Configuration-based management for all package managers
   - Coordinated version management across tools

7. **doctor** - System health and configuration validation
   - Check system dependencies
   - Validate configurations
   - Diagnose common issues

### Git Platform Management

8. **git** - Unified Git platform management interface
   - `git config` - Repository configuration management (delegates to repo-config)
   - `git webhook` - Webhook management (via repo-config webhook)
   - `git event` - Event processing and monitoring
   - Provides consistent interface for all Git-related operations

### Utility Commands

9. **version** - Display version information
10. **help** - Show help for commands

### Special Commands

11. **shell** - Interactive shell mode (debug feature)
    - Available via `--debug-shell` flag
    - For development and debugging
    - Hidden from normal help output

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
