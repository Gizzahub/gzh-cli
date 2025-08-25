# 📝 YAML Configuration Guide

Comprehensive guide to gzh-cli YAML configuration with examples, best practices, and troubleshooting.

## 📋 Table of Contents

1. [Quick Start](#quick-start)
1. [Configuration Structure](#configuration-structure)
1. [Provider Configuration](#provider-configuration)
1. [Advanced Settings](#advanced-settings)
1. [Example Configurations](#example-configurations)
1. [Schema Validation](#schema-validation)
1. [Troubleshooting](#troubleshooting)

## 🚀 Quick Start

### Minimal Configuration

```yaml
# ~/.config/gzh-manager/gzh.yaml
global:
  clone_base_dir: "$HOME/repos"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "my-org"
        clone_dir: "$HOME/repos/my-org"
```

### Basic Multi-Platform Setup

```yaml
# ~/.config/gzh-manager/gzh.yaml
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset
  log_level: info

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "my-company"
        clone_dir: "$HOME/repos/github/my-company"

  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.com/api/v4"
    groups:
      - name: "my-group"
        clone_dir: "$HOME/repos/gitlab/my-group"

  gitea:
    token: "${GITEA_TOKEN}"
    api_url: "https://gitea.company.com/api/v1"
    organizations:
      - name: "infrastructure"
        clone_dir: "$HOME/repos/gitea/infrastructure"
```

## 🏗️ Configuration Structure

### Global Settings

```yaml
global:
  # Base directory for all cloned repositories
  clone_base_dir: "$HOME/repos"

  # Default strategy for repository updates
  default_strategy: reset  # reset, pull, fetch, rebase, clone

  # Logging configuration
  log_level: info          # debug, info, warn, error
  log_file: "$HOME/.config/gzh-manager/gzh.log"

  # Output format
  output_format: table     # table, json, yaml, csv

  # Concurrency settings
  concurrent_jobs: 5
  timeout: "30m"
  retry_attempts: 3
  retry_delay: "5s"
```

### Command-Specific Settings

```yaml
commands:
  synclone:
    # Concurrent repository operations
    concurrent_jobs: 10
    retry_attempts: 3

    # Progress reporting
    show_progress: true
    progress_interval: "1s"

  quality:
    # Automatic tool installation
    auto_install_tools: true

    # Default languages to check
    default_languages: ["go", "python", "javascript"]

    # Output settings
    output_format: table
    fail_on_errors: true

  git:
    # Default branch name
    default_branch: "main"

    # Automatic upstream setup
    auto_setup_upstream: true

    # Git operation timeout
    timeout: "10m"

  ide:
    # Automatic monitoring
    auto_monitor: true

    # Monitoring interval
    monitor_interval: "1s"

    # Backup settings
    backup:
      enabled: true
      location: "$HOME/.config/gzh-manager/ide-backups"
      retention_days: 30
```

## 🔧 Provider Configuration

### GitHub Configuration

```yaml
providers:
  github:
    # Authentication
    token: "${GITHUB_TOKEN}"

    # API settings
    api_url: "https://api.github.com"  # For GitHub Enterprise

    # Organizations
    organizations:
      - name: "my-company"
        clone_dir: "$HOME/repos/github/my-company"
        visibility: all              # all, public, private
        strategy: reset              # Override global strategy

        # Filtering options
        include_archived: false
        include_forks: false
        include_private: true

        # Advanced filtering
        filters:
          topics: ["microservice", "api"]
          languages: ["go", "python"]
          min_stars: 10
          max_age_days: 365
          updated_after: "2024-01-01"

        # Repository-specific overrides
        repositories:
          - name: "critical-service"
            strategy: rebase
            branch: "develop"
            clone_dir: "$HOME/projects/critical-service"

          - name: "archived-project"
            exclude: true            # Skip this repository
```

### GitLab Configuration

```yaml
providers:
  gitlab:
    # Authentication
    token: "${GITLAB_TOKEN}"

    # API settings (for self-hosted GitLab)
    api_url: "https://gitlab.company.com/api/v4"

    # Groups
    groups:
      - name: "development"
        clone_dir: "$HOME/repos/gitlab/development"
        visibility: internal        # public, internal, private

        # Subgroup support
        include_subgroups: true

        # Project filtering
        filters:
          archived: false
          starred: true
          last_activity_after: "2024-01-01"

      - name: "infrastructure"
        clone_dir: "$HOME/repos/gitlab/infrastructure"
        visibility: private

        # Project-specific settings
        projects:
          - name: "kubernetes-configs"
            branch: "production"
            strategy: fetch
```

### Gitea Configuration

```yaml
providers:
  gitea:
    # Authentication
    token: "${GITEA_TOKEN}"

    # API settings
    api_url: "https://gitea.company.com/api/v1"

    # Organizations
    organizations:
      - name: "infrastructure"
        clone_dir: "$HOME/repos/gitea/infrastructure"

        # Gitea-specific filters
        filters:
          private: false
          mirror: false
          template: false

      - name: "development"
        clone_dir: "$HOME/repos/gitea/development"

        # Repository overrides
        repositories:
          - name: "legacy-system"
            exclude: true
```

### Gogs Configuration

```yaml
providers:
  gogs:
    # Authentication
    token: "${GOGS_TOKEN}"

    # API settings
    api_url: "https://gogs.company.com/api/v1"

    # Organizations
    organizations:
      - name: "legacy"
        clone_dir: "$HOME/repos/gogs/legacy"

        # Basic filtering
        filters:
          private: false
```

## ⚙️ Advanced Settings

### Environment Variables

```yaml
# Environment variable expansion
global:
  clone_base_dir: "${REPOS_BASE:-$HOME/repos}"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "${COMPANY_ORG:-my-company}"
        clone_dir: "${REPOS_BASE:-$HOME/repos}/github/${COMPANY_ORG:-my-company}"
```

### Conditional Configuration

```yaml
# Platform-specific settings
global:
  clone_base_dir: !if
    condition: "${OS}" == "windows"
    then: "C:/repos"
    else: "$HOME/repos"

# Environment-specific providers
providers: !if
  condition: "${ENV}" == "production"
  then:
    github:
      token: "${PROD_GITHUB_TOKEN}"
      organizations:
        - name: "production-org"
  else:
    github:
      token: "${DEV_GITHUB_TOKEN}"
      organizations:
        - name: "development-org"
```

### Performance Tuning

```yaml
global:
  # Optimize for large organizations
  concurrent_jobs: 20
  timeout: "60m"

  # Memory management
  max_memory_usage: "2GB"
  gc_interval: "5m"

  # Network optimization
  http_timeout: "30s"
  http_retry_attempts: 5
  http_retry_delay: "10s"

# Provider-specific rate limiting
providers:
  github:
    rate_limiting:
      requests_per_hour: 4500
      burst_size: 100

  gitlab:
    rate_limiting:
      requests_per_minute: 300
      burst_size: 50
```

## 📋 Example Configurations

### Personal Developer Setup

```yaml
# Personal development environment
global:
  clone_base_dir: "$HOME/code"
  default_strategy: pull
  log_level: info

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "my-username"
        clone_dir: "$HOME/code/personal"
        visibility: all
        include_forks: true

commands:
  quality:
    auto_install_tools: true
    default_languages: ["go", "python", "javascript", "rust"]

  ide:
    auto_monitor: true
    backup_enabled: true

  dev_env:
    default_provider: aws
    backup_location: "$HOME/.config/gzh-manager/backups"
```

### Enterprise Team Setup

```yaml
# Enterprise team configuration
global:
  clone_base_dir: "$HOME/work/repos"
  default_strategy: reset
  log_level: warn
  concurrent_jobs: 15

providers:
  github:
    token: "${GITHUB_ENTERPRISE_TOKEN}"
    api_url: "https://github.company.com/api/v3"
    organizations:
      - name: "platform-team"
        clone_dir: "$HOME/work/repos/platform"
        visibility: private
        filters:
          topics: ["platform", "infrastructure"]

      - name: "product-team"
        clone_dir: "$HOME/work/repos/products"
        visibility: private
        filters:
          topics: ["product", "api"]

  gitlab:
    token: "${GITLAB_ENTERPRISE_TOKEN}"
    api_url: "https://gitlab.company.com/api/v4"
    groups:
      - name: "infrastructure"
        clone_dir: "$HOME/work/repos/infrastructure"
        visibility: internal

commands:
  synclone:
    concurrent_jobs: 20
    show_progress: false  # Reduce noise in CI

  quality:
    auto_install_tools: false  # Managed by team
    default_checks: ["format", "lint", "security"]
    fail_on_errors: true

  git:
    default_branch: "main"
    auto_setup_upstream: true
```

### CI/CD Pipeline Configuration

```yaml
# Optimized for CI/CD environments
global:
  clone_base_dir: "/tmp/repos"
  default_strategy: clone
  log_level: error
  concurrent_jobs: 5
  timeout: "15m"

providers:
  github:
    token: "${CI_GITHUB_TOKEN}"
    organizations:
      - name: "${CI_TARGET_ORG}"
        clone_dir: "/tmp/repos/${CI_TARGET_ORG}"
        visibility: private

        # CI-specific filtering
        filters:
          updated_after: "${CI_START_DATE}"
          topics: ["${CI_PROJECT_TYPE}"]

commands:
  synclone:
    show_progress: false
    retry_attempts: 1

  quality:
    auto_install_tools: true
    output_format: json
    fail_on_errors: true

    # CI-specific quality settings
    checks: ["lint", "security"]
    output_file: "quality-report.json"
```

## 🔍 Schema Validation

### Validate Configuration

```bash
# Validate current configuration
gz config validate

# Validate specific file
gz config validate --file custom-config.yaml

# Show effective configuration
gz config show

# Export configuration schema
gz config schema --output schema.json
```

### Configuration Testing

```bash
# Test configuration without executing
gz synclone --dry-run --config test-config.yaml

# Test provider authentication
gz config test-auth --provider github

# Validate all providers
gz config test-auth --all
```

## 🔧 Configuration Management

### Multiple Configurations

```bash
# Use specific configuration file
gz synclone --config ~/.config/gzh-manager/work.yaml

# Environment-specific configurations
export GZH_CONFIG_PATH=~/.config/gzh-manager/production.yaml
gz synclone

# Project-specific configuration
cd my-project
echo "providers: ..." > .gzh.yaml
gz synclone  # Uses .gzh.yaml in current directory
```

### Configuration Inheritance

```yaml
# base-config.yaml
global:
  clone_base_dir: "$HOME/repos"
  log_level: info

providers:
  github:
    token: "${GITHUB_TOKEN}"

---
# work-config.yaml (inherits from base-config.yaml)
inherit: "base-config.yaml"

global:
  clone_base_dir: "$HOME/work/repos"  # Override

providers:
  github:
    organizations:                     # Extend
      - name: "work-org"
```

## 🆘 Troubleshooting

### Common Configuration Issues

#### YAML Syntax Errors

```bash
# Validate YAML syntax
gz config validate

# Common issues:
# 1. Incorrect indentation
# 2. Missing quotes around special characters
# 3. Duplicate keys
# 4. Invalid YAML structure
```

#### Environment Variable Issues

```bash
# Check environment variable expansion
gz config show --expand-vars

# Debug environment variables
echo $GITHUB_TOKEN
env | grep -E "(GITHUB|GITLAB|GITEA)_TOKEN"
```

#### Provider Authentication Problems

```bash
# Test authentication
gz config test-auth --provider github --verbose

# Debug API connectivity
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
```

#### File Permission Issues

```bash
# Check configuration file permissions
ls -la ~/.config/gzh-manager/gzh.yaml

# Fix permissions
chmod 600 ~/.config/gzh-manager/gzh.yaml
chmod 700 ~/.config/gzh-manager/
```

### Configuration Debugging

```bash
# Enable debug logging
gz synclone --debug --verbose

# Show effective configuration
gz config show --format yaml

# Trace configuration loading
gz config trace --file ~/.config/gzh-manager/gzh.yaml

# Export merged configuration
gz config export --output effective-config.yaml
```

### Performance Issues

```bash
# Reduce concurrent jobs
# In configuration:
global:
  concurrent_jobs: 3

# Increase timeouts
global:
  timeout: "60m"
  http_timeout: "30s"

# Enable compression
providers:
  github:
    compression: true
```

## 📖 Best Practices

### Security

- Use environment variables for tokens
- Set restrictive file permissions (600)
- Never commit tokens to version control
- Rotate tokens regularly

### Performance

- Adjust concurrent_jobs based on system resources
- Use appropriate timeouts for your network
- Filter repositories to avoid unnecessary clones
- Use reset strategy for automated environments

### Maintainability

- Use descriptive names for configurations
- Comment complex settings
- Separate personal and work configurations
- Use configuration inheritance for common settings

______________________________________________________________________

**Configuration File**: `~/.config/gzh-manager/gzh.yaml`
**Schema Validation**: `gz config validate`
**Environment Variables**: Expanded with `${VAR}` syntax
**File Permissions**: 600 (user read/write only)
