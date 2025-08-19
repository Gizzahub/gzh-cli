# Complete Command Reference

Comprehensive reference documentation for all `gz` commands and their options.

## Table of Contents

1. [Overview](#overview)
2. [Global Options](#global-options)
3. [Core Commands](#core-commands)
4. [Environment Management](#environment-management)
5. [Repository Management](#repository-management)
6. [Platform Integration](#platform-integration)
7. [Configuration](#configuration)
8. [Examples](#examples)
9. [Troubleshooting](#troubleshooting)

## Overview

The `gz` CLI provides comprehensive tools for development environment management, repository operations, and code quality control:

```bash
gz <command> [subcommand] [flags]
```

### Command Categories

#### Core Features
- **[synclone](#synclone)** - Multi-platform repository synchronization
- **[git](#git)** - Unified Git operations and platform management
- **[quality](#quality)** - Multi-language code quality management
- **[ide](#ide)** - JetBrains IDE monitoring and management
- **[profile](#profile)** - Performance profiling and analysis

#### Environment Management
- **[dev-env](#dev-env)** - Development environment configuration
- **[net-env](#net-env)** - Network environment transitions
- **[pm](#pm)** - Package manager updates and management

#### Repository Management
- **[repo-config](#repo-config)** - GitHub repository configuration management

## Global Options

All commands support these global flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Configuration file path | Auto-detected |
| `--debug` | Enable debug logging | `false` |
| `--help` | Show help information | - |
| `--log-level` | Set log level (debug, info, warn, error) | `info` |
| `--quiet` | Suppress non-error output | `false` |
| `--verbose` | Enable verbose output | `false` |
| `--version` | Show version information | - |

### Environment Variables

```bash
# Authentication
export GITHUB_TOKEN="ghp_..."
export GITLAB_TOKEN="glpat-..."
export GITEA_TOKEN="..."

# Global settings
export GZ_DEBUG="true"
export GZ_LOG_LEVEL="debug"
export GZ_CONFIG_DIR="~/.config/gzh-manager"
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Misuse of command (invalid arguments) |
| `3` | Authentication error |
| `4` | Network error |
| `5` | Configuration error |
| `6` | File system error |

## Core Commands

### synclone

Multi-platform repository synchronization with support for GitHub, GitLab, Gitea, and Gogs.

#### Basic Usage

```bash
gz synclone <platform> [flags]
gz synclone --config <config-file>
```

#### Platform Subcommands

##### `gz synclone github`

Clone and synchronize GitHub organizations.

```bash
gz synclone github --org <organization> [flags]
```

**Key Flags:**
- `--org`, `-o` - Organization name (required)
- `--target`, `-t` - Target directory (default: current directory)
- `--strategy` - Clone/update strategy: reset, pull, fetch, rebase, clone, skip
- `--include-archived` - Include archived repositories
- `--include-forks` - Include forked repositories
- `--language` - Filter by programming language
- `--topic` - Filter by repository topic
- `--min-stars` - Minimum star count
- `--parallel` - Number of concurrent operations (default: 5)
- `--shallow` - Use shallow clones
- `--depth` - Depth for shallow clones (default: 1)

**Examples:**
```bash
# Basic organization clone
gz synclone github --org kubernetes

# Clone with filters
gz synclone github --org prometheus --language Go --min-stars 100

# Shallow clone for space efficiency
gz synclone github --org large-org --shallow --depth 1
```

##### `gz synclone gitlab`

Clone and synchronize GitLab groups.

```bash
gz synclone gitlab --group <group> [flags]
```

**Key Flags:**
- `--group`, `-g` - Group ID or path (required)
- `--include-subgroups` - Include subgroups recursively
- `--visibility` - Filter by visibility: public, internal, private
- `--base-url` - GitLab instance URL (default: https://gitlab.com)

**Examples:**
```bash
# Clone GitLab group with subgroups
gz synclone gitlab --group mygroup --include-subgroups

# Self-hosted GitLab
gz synclone gitlab --group mygroup --base-url https://gitlab.company.com
```

##### `gz synclone gitea`

Clone and synchronize Gitea organizations.

```bash
gz synclone gitea --org <organization> --base-url <url> [flags]
```

**Key Flags:**
- `--org`, `-o` - Organization name (required)
- `--base-url` - Gitea instance URL (required)
- `--token` - Authentication token

#### Clone Strategies

| Strategy | Behavior | Use Case |
|----------|----------|----------|
| `reset` | Hard reset to match remote (default) | CI/CD, mirrors |
| `pull` | Merge remote changes | Active development |
| `fetch` | Update refs only | Inspection |
| `rebase` | Rebase local changes on remote | Clean history |
| `clone` | Fresh clone (removes existing) | Clean start |
| `skip` | Skip existing repositories | Initial clone only |

### git

Unified Git operations and platform management.

#### Basic Usage

```bash
gz git <subcommand> [flags]
```

#### Repository Operations

##### `gz git repo clone-or-update`

Intelligently clone new repositories or update existing ones.

```bash
gz git repo clone-or-update <repository-url> [target-path] [flags]
```

**Arguments:**
- `repository-url` - Git repository URL (HTTPS, SSH, or ssh:// format)
- `target-path` - Optional target directory (auto-extracts repo name if omitted)

**Key Flags:**
- `--strategy`, `-s` - Update strategy: rebase, reset, clone, skip, pull, fetch
- `--branch`, `-b` - Branch to check out
- `--shallow` - Use shallow clone
- `--depth` - Depth for shallow clone (default: 1)

**Examples:**
```bash
# Clone new repository (auto-extracts name)
gz git repo clone-or-update https://github.com/user/awesome-project.git

# Clone to specific directory
gz git repo clone-or-update https://github.com/user/repo.git my-local-name

# Update with specific strategy
gz git repo clone-or-update https://github.com/user/repo.git --strategy reset

# Clone specific branch
gz git repo clone-or-update https://github.com/user/repo.git -b develop
```

#### Configuration Management

##### `gz git config`

Manage repository configurations at scale across organizations.

```bash
gz git config <action> [flags]
```

**Actions:**

###### `gz git config audit`
```bash
gz git config audit --org <organization> [flags]
```

**Key Flags:**
- `--org` - Organization name (required)
- `--framework` - Compliance framework: SOC2, GDPR, HIPAA, PCI-DSS
- `--output` - Output format: table, json, yaml, csv, html
- `--severity` - Minimum severity: low, medium, high, critical

###### `gz git config apply`
```bash
gz git config apply --config <config-file> [flags]
```

**Key Flags:**
- `--config` - Configuration file path (required)
- `--dry-run` - Preview changes without applying
- `--force` - Apply changes without confirmation

###### `gz git config diff`
```bash
gz git config diff --org <organization> [flags]
```

**Key Flags:**
- `--org` - Organization name (required)
- `--baseline` - Baseline configuration file
- `--output` - Output format: unified, side-by-side, json

#### Webhook Management

##### `gz git webhook`

Create, manage, and monitor webhooks.

```bash
gz git webhook <action> [flags]
```

**Actions:**

###### `gz git webhook create`
```bash
gz git webhook create --org <org> --repo <repo> --url <webhook-url> [flags]
```

**Key Flags:**
- `--org` - Organization name (required)
- `--repo` - Repository name (required)
- `--url` - Webhook URL (required)
- `--events` - Webhook events (comma-separated, default: push)
- `--secret` - Webhook secret

###### `gz git webhook list`
```bash
gz git webhook list --org <organization> [flags]
```

### quality

Multi-language code quality management with integrated formatters and linters.

#### Basic Usage

```bash
gz quality <action> [flags]
```

#### Supported Languages

- **Go** - gofumpt, gci, golangci-lint, staticcheck
- **Python** - black, isort, ruff, flake8, pylint, mypy
- **JavaScript/TypeScript** - prettier, eslint, tsc
- **Rust** - rustfmt, clippy
- **Java** - google-java-format, checkstyle, spotbugs, pmd
- **C/C++** - clang-format, clang-tidy, cppcheck

#### Actions

##### `gz quality run`

Run formatters and linters for detected languages.

```bash
gz quality run [path] [flags]
```

**Key Flags:**
- `--languages` - Comma-separated list of languages to process
- `--tools` - Specific tools to run (comma-separated)
- `--exclude-tools` - Tools to exclude (comma-separated)
- `--auto-fix` - Automatically fix issues when possible (default: true)
- `--parallel` - Run tools in parallel (default: true)
- `--output` - Output format: text, json, sarif, checkstyle

**Examples:**
```bash
# Run all quality checks
gz quality run

# Run only Go tools
gz quality run --languages go

# Generate SARIF report
gz quality run --output sarif --output-file quality.sarif
```

##### `gz quality install`

Install or update quality tools.

```bash
gz quality install [tools] [flags]
```

**Key Flags:**
- `--languages` - Install tools for specific languages
- `--force` - Force reinstall even if already present
- `--version` - Install specific versions (format: tool@version)

**Examples:**
```bash
# Install all tools
gz quality install

# Install Go tools only
gz quality install --languages go

# Install specific versions
gz quality install --version golangci-lint@1.54.2,black@23.7.0
```

##### `gz quality check`

Check code quality without making changes.

```bash
gz quality check [path] [flags]
```

##### `gz quality list`

List available and installed tools.

```bash
gz quality list [flags]
```

### ide

JetBrains IDE monitoring and configuration management.

#### Basic Usage

```bash
gz ide <action> [flags]
```

#### Supported IDEs

- IntelliJ IDEA (Community, Ultimate)
- PyCharm (Community, Professional)
- WebStorm, PhpStorm, RubyMine
- CLion, GoLand, DataGrip
- Android Studio, Rider

#### Actions

##### `gz ide monitor`

Monitor IDE configuration changes in real-time.

```bash
gz ide monitor [flags]
```

**Key Flags:**
- `--product` - Specific IDE product to monitor
- `--interval` - Monitoring interval (default: 1s)
- `--auto-fix` - Automatically fix sync issues

**Examples:**
```bash
# Monitor all IDEs
gz ide monitor

# Monitor specific product
gz ide monitor --product IntelliJIdea2023.2
```

##### `gz ide fix-sync`

Fix IDE synchronization issues.

```bash
gz ide fix-sync [flags]
```

**Key Flags:**
- `--dry-run` - Preview fixes without applying
- `--backup` - Create backup before fixing

##### `gz ide list`

List installed JetBrains IDEs.

```bash
gz ide list [flags]
```

### profile

Performance profiling and analysis using Go pprof.

#### Basic Usage

```bash
gz profile <action> [flags]
```

#### Actions

##### `gz profile stats`

Show runtime statistics.

```bash
gz profile stats [flags]
```

##### `gz profile server`

Start pprof HTTP server.

```bash
gz profile server [flags]
```

**Key Flags:**
- `--port` - Server port (default: 6060)
- `--host` - Server host (default: localhost)

##### `gz profile cpu`

Perform CPU profiling.

```bash
gz profile cpu [flags]
```

**Key Flags:**
- `--duration` - Profiling duration (default: 30s)
- `--output` - Output file name

##### `gz profile memory`

Perform memory profiling.

```bash
gz profile memory [flags]
```

## Environment Management

### dev-env

Development environment configuration management.

#### Basic Usage

```bash
gz dev-env <provider> <action> [flags]
```

#### Supported Providers

- **AWS** - AWS CLI and configuration management
- **Docker** - Docker environment setup and management
- **Kubernetes** - Kubernetes cluster configuration
- **SSH** - SSH key and configuration management

#### Actions

##### `gz dev-env aws`

AWS environment management.

```bash
gz dev-env aws <action> [flags]
```

**Actions:**
- `configure` - Set up AWS configuration
- `backup` - Backup AWS configuration
- `restore` - Restore AWS configuration
- `status` - Show AWS status

##### `gz dev-env docker`

Docker environment management.

```bash
gz dev-env docker <action> [flags]
```

##### `gz dev-env k8s`

Kubernetes environment management.

```bash
gz dev-env k8s <action> [flags]
```

### net-env

Network environment transitions and management.

#### Basic Usage

```bash
gz net-env <action> [flags]
```

#### Actions

##### `gz net-env monitor`

Monitor network changes and automatically switch configurations.

```bash
gz net-env monitor [flags]
```

##### `gz net-env status`

Show current network status.

```bash
gz net-env status [flags]
```

##### `gz net-env switch`

Manually switch network profile.

```bash
gz net-env switch <profile> [flags]
```

### pm

Package manager updates and management.

#### Basic Usage

```bash
gz pm <action> [flags]
```

#### Supported Package Managers

- **Language Managers**: asdf, nvm, pyenv, rbenv
- **System Managers**: Homebrew (macOS), apt (Ubuntu), yum (CentOS)
- **Development Tools**: npm, pip, cargo, go modules
- **Cloud Tools**: SDKMAN, kubectl, helm

#### Actions

##### `gz pm update`

Update package managers and their packages.

```bash
gz pm update [flags]
```

**Key Flags:**
- `--all` - Update all package managers
- `--managers` - Specific managers to update (comma-separated)
- `--dry-run` - Show what would be updated

**Examples:**
```bash
# Update all package managers
gz pm update --all

# Update specific managers
gz pm update --managers homebrew,asdf

# Show what would be updated
gz pm update --all --dry-run
```

##### `gz pm list`

List installed package managers.

```bash
gz pm list [flags]
```

##### `gz pm status`

Show package manager status.

```bash
gz pm status [flags]
```

## Repository Management

### repo-config

GitHub repository configuration management for organization-wide policy enforcement.

#### Basic Usage

```bash
gz repo-config <action> [flags]
```

**Note:** This command is also available as `gz git config`.

#### Actions

##### `gz repo-config audit`

Audit repository settings against compliance frameworks.

```bash
gz repo-config audit --org <organization> [flags]
```

**Key Flags:**
- `--org` - Organization name (required)
- `--framework` - Compliance framework: SOC2, GDPR, HIPAA, PCI-DSS
- `--output` - Output format: table, json, yaml, csv, html
- `--severity` - Minimum severity: low, medium, high, critical

##### `gz repo-config apply`

Apply configuration policies to repositories.

```bash
gz repo-config apply --config <config-file> [flags]
```

**Key Flags:**
- `--config` - Configuration file path (required)
- `--dry-run` - Preview changes without applying
- `--force` - Apply changes without confirmation

##### `gz repo-config generate`

Generate configuration templates.

```bash
gz repo-config generate --org <organization> [flags]
```

**Key Flags:**
- `--template` - Template type: minimal, standard, enterprise
- `--output` - Output file name

## Configuration

### Configuration Priority

1. **Command-Line Flags** (Highest Priority)
2. **Environment Variables** (Second Priority)
3. **Configuration Files** (Third Priority)
4. **Default Values** (Lowest Priority)

### Configuration File Locations

Commands look for configuration in this order:

1. `--config` flag value
2. Current directory: `./gzh.yaml`
3. User config: `~/.config/gzh-manager/gzh.yaml`
4. System config: `/etc/gzh-manager/gzh.yaml`

### Common Configuration

```yaml
# gzh.yaml
version: "1.0.0"
default_provider: github

global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset
  concurrency:
    clone_workers: 10

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/github/myorg"

ide:
  enabled: true
  auto_fix_sync: true

quality:
  enabled: true
  auto_fix: true
  parallel: true
```

## Examples

### Repository Workflow

```bash
# Clone/update repository
gz git repo clone-or-update https://github.com/myorg/service.git

# Run quality checks
gz quality run

# Audit repository configuration
gz git config audit --org myorg

# Monitor IDE changes
gz ide monitor
```

### Organization Setup

```bash
# Sync all repositories
gz synclone github --org myorg

# Apply security policies
gz git config apply --config security-policy.yaml --org myorg

# Set up webhooks
gz git webhook create --org myorg --repo service --url https://ci.example.com/webhook

# Generate compliance report
gz git config audit --org myorg --framework SOC2 --output html
```

### Development Environment

```bash
# Update all package managers
gz pm update --all

# Set up development environment
gz dev-env aws configure
gz dev-env docker setup

# Start network monitoring
gz net-env monitor

# Profile application performance
gz profile cpu --duration 60s
```

### CI/CD Integration

```bash
# Quality checks in CI
gz quality run --output sarif --fail-on-error

# Repository compliance check
gz git config audit --org myorg --exit-on-fail

# Performance profiling
gz profile stats
```

## Troubleshooting

### Common Issues

#### 1. Command Not Found

```bash
# Check installation
which gz

# Check PATH
echo $PATH

# Reinstall if needed
make install
```

#### 2. Authentication Issues

```bash
# Check tokens
echo $GITHUB_TOKEN

# Test connectivity
gz git config audit --org myorg --dry-run
```

#### 3. Configuration Errors

```bash
# Validate configuration
gz config validate

# Show configuration sources
gz config show

# Debug configuration loading
gz --debug config show
```

#### 4. Performance Issues

```bash
# Reduce parallelism
gz synclone github --org myorg --parallel 2

# Use shallow clones
gz synclone github --org myorg --shallow

# Enable caching
gz quality run --cache
```

### Debug Mode

```bash
# Enable debug logging for any command
gz --debug <command>

# Set specific log level
gz --log-level debug <command>

# Verbose output
gz --verbose <command>
```

### Help System

```bash
# Show all commands
gz help

# Command-specific help
gz help <command>
gz <command> --help

# Subcommand help
gz <command> <subcommand> --help
```

## Support

For additional help:

1. Run `gz <command> --help` for command-specific options
2. Check the [Configuration Guide](../30-configuration/30-configuration-guide.md)
3. Review the [Examples](../../examples/) directory
4. Open an issue on [GitHub](https://github.com/gizzahub/gzh-cli/issues)
