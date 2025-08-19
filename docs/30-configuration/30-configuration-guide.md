# Configuration System Guide

Comprehensive guide for the gzh-cli unified configuration system supporting all commands through a single `gzh.yaml` configuration file.

## Table of Contents

1. [Overview](#overview)
2. [Configuration Priority](#configuration-priority)
3. [Configuration Structure](#configuration-structure)
4. [Platform Configuration](#platform-configuration)
5. [Environment Variables](#environment-variables)
6. [Configuration Management](#configuration-management)
7. [Migration Guide](#migration-guide)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Overview

The gzh-cli project uses a unified configuration system that supports all commands through a single `gzh.yaml` configuration file. This system provides:

- **Unified Configuration**: Single file for all commands and features
- **Priority System**: Predictable configuration resolution
- **Environment Integration**: Support for environment variables
- **Schema Validation**: Built-in validation with clear error messages
- **Migration Support**: Automatic migration from legacy formats

### Key Features

- **Command-Line Override**: CLI flags always take precedence
- **Environment Variable Support**: Secure token management
- **Hierarchical Configuration**: User, project, and system-level configs
- **Schema Validation**: JSON Schema validation for structure
- **Auto-migration**: Legacy configuration file migration

## Configuration Priority

The configuration system follows a strict priority order where higher priority sources override lower priority ones.

### Priority Order (Highest to Lowest)

1. **Command-Line Flags** (Highest Priority)
2. **Environment Variables** (Second Priority)
3. **Configuration Files** (Third Priority)
4. **Default Values** (Lowest Priority)

### Detailed Priority Explanation

#### 1. Command-Line Flags (Highest Priority)

Command-line flags always take precedence over all other configuration sources. When a flag is specified, it overrides any corresponding setting from environment variables, configuration files, or default values.

**Examples:**

```bash
# Override configuration file strategy
gz synclone github --org myorg --strategy=pull

# Override environment variable token
gz synclone github --org myorg --token=ghp_custom_token

# Override configuration file parallel workers
gz synclone github --org myorg --parallel=20
```

#### 2. Environment Variables (Second Priority)

Environment variables override configuration file values but are overridden by command-line flags.

**Key Environment Variables:**

- `GZH_CONFIG_PATH`: Override config file location
- `GITHUB_TOKEN`: GitHub authentication token
- `GITLAB_TOKEN`: GitLab authentication token
- `GITEA_TOKEN`: Gitea authentication token

**Examples:**

```bash
# Environment variable overrides config file
export GITHUB_TOKEN=ghp_env_token
gz synclone github --org myorg  # Uses ghp_env_token

# But command-line flag overrides environment variable
gz synclone github --org myorg --token=ghp_flag_token  # Uses ghp_flag_token
```

#### 3. Configuration Files (Third Priority)

Configuration files provide the base configuration but are overridden by environment variables and command-line flags.

**In configuration files, you can reference environment variables:**

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}" # Expands to environment variable value
```

#### 4. Default Values (Lowest Priority)

Default values are used when no higher priority source provides a value.

**Common defaults:**

- `strategy: reset`
- `parallel: 10`
- `visibility: all`
- `timeout: 30s`

### Configuration File Search Order

When no explicit config path is provided, the system searches in this order:

1. **Environment Variable**: `GZH_CONFIG_PATH`
2. **Current Directory**: `./gzh.yaml`, `./gzh.yml`
3. **User Config**: `~/.config/gzh-manager/gzh.yaml`
4. **System Config**: `/etc/gzh-manager/gzh.yaml`
5. **Legacy Files**: `./synclone.yaml`, `./synclone.yml` (auto-migrated)

### Priority Resolution Examples

#### Example 1: Token Resolution

```yaml
# config.yaml
providers:
  github:
    token: "ghp_config_token"
```

```bash
# Environment variable
export GITHUB_TOKEN=ghp_env_token

# Command execution
gz synclone github --org myorg --token=ghp_flag_token
```

**Resolution:** `ghp_flag_token` (CLI flag wins)

#### Example 2: Strategy Resolution

```yaml
# config.yaml
global:
  default_strategy: "reset"
```

```bash
# Command execution
gz synclone github --org myorg --strategy=pull
```

**Resolution:** `pull` (CLI flag overrides config file)

### Environment Variable Expansion

Configuration files support environment variable expansion using `${VAR_NAME}` syntax:

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    api_url: "${GITHUB_API_URL:-https://api.github.com}" # With default
```

**Priority for expanded variables:**

1. Command-line flags (if applicable)
2. Environment variables (used in expansion)
3. Default values in expansion syntax
4. Configuration file literal values

## Configuration Structure

### Basic Structure

```yaml
# gzh-manager unified configuration
version: "1.0.0"
default_provider: github

# Global settings that apply to all commands
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset
  default_visibility: all
  timeouts:
    http_timeout: 30s
    git_timeout: 5m
    rate_limit_timeout: 1h
  concurrency:
    clone_workers: 10
    update_workers: 15
    api_workers: 5

# Provider configurations (GitHub, GitLab, Gitea, Gogs)
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/github/myorg"
        visibility: all
        strategy: reset
  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.example.com/api/v4"
    organizations:
      - name: "mygroup"
        clone_dir: "$HOME/repos/gitlab/mygroup"
        recursive: true

# IDE configuration
ide:
  enabled: true
  watch_directories:
    - "$HOME/.config"
    - "$HOME/.local/share/JetBrains"
  exclude_patterns:
    - "\.git/.*"
    - "node_modules/.*"
    - "\.DS_Store"
  jetbrains_products:
    - "IntelliJ"
    - "PyCharm"
    - "GoLand"
    - "WebStorm"
  auto_fix_sync: true
  sync_settings:
    enabled: true
    interval: 5m
    sync_types:
      - "keymap"
      - "editor"
      - "ui"
      - "plugins"
    backup_before_sync: true
  logging:
    level: info
    file_path: "$HOME/.local/share/gzh-manager/logs/ide.log"
    console: true

# Development environment configuration
dev_env:
  enabled: true
  backup_location: "$HOME/.gz/backups"
  auto_backup: true
  providers:
    aws:
      default_profile: default
      preferred_regions:
        - us-west-2
        - us-east-1
      credentials_file: "$HOME/.aws/credentials"
      config_file: "$HOME/.aws/config"
    gcp:
      default_project: "my-project"
      preferred_regions:
        - us-central1
        - us-west1
      use_adc: true
    azure:
      default_subscription: "my-subscription"
      preferred_regions:
        - westus2
        - eastus
  containers:
    default_runtime: docker
    docker:
      socket_path: /var/run/docker.sock
      default_registry: docker.io
  kubernetes:
    kubeconfig_path: "$HOME/.kube/config"
    default_namespace: default
    auto_discovery: true

# Network environment configuration
net_env:
  enabled: true
  wifi_detection:
    enabled: true
    interval: 5s
    known_networks:
      "Home-WiFi":
        ssid: "Home-WiFi"
        type: home
        dns_servers:
          - "192.168.1.1"
      "Office-WiFi":
        ssid: "Office-WiFi"
        type: work
        vpn_config: "work-vpn"
        dns_servers:
          - "10.0.0.1"
  vpn:
    profiles:
      work-vpn:
        type: openvpn
        config_file: "$HOME/.config/vpn/work.ovpn"
        auto_connect_networks:
          - "Office-WiFi"
    default_profile: work-vpn
  dns:
    default_servers:
      - "1.1.1.1"
      - "1.0.0.1"
  proxy:
    profiles:
      corporate:
        type: http
        host: "proxy.company.com"
        port: 8080
        username: "${PROXY_USERNAME}"
        password: "${PROXY_PASSWORD}"
```

## Platform Configuration

### GitHub Configuration

```yaml
providers:
  github:
    # Authentication
    token: "${GITHUB_TOKEN}"
    api_url: "https://api.github.com"  # For GitHub Enterprise

    # Organizations
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/github/myorg"
        visibility: all  # all, public, private
        strategy: reset  # reset, pull, fetch, rebase
        
        # Filtering
        include_archived: false
        include_forks: false
        filters:
          languages: ["Go", "Python"]
          topics: ["microservice", "api"]
          min_stars: 10
          updated_after: "2024-01-01"
        
        # Repository-specific settings
        repositories:
          - name: "critical-service"
            strategy: "pull"
            branch: "production"
```

### GitLab Configuration

```yaml
providers:
  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.com/api/v4"  # For self-hosted GitLab
    
    groups:
      - name: "mygroup"
        clone_dir: "$HOME/repos/gitlab/mygroup"
        include_subgroups: true
        visibility: public  # public, internal, private
        strategy: fetch
```

### Gitea Configuration

```yaml
providers:
  gitea:
    token: "${GITEA_TOKEN}"
    api_url: "https://gitea.company.com/api/v1"
    
    organizations:
      - name: "infrastructure"
        clone_dir: "$HOME/repos/gitea/infrastructure"
        strategy: reset
```

### Gogs Configuration

```yaml
providers:
  gogs:
    token: "${GOGS_TOKEN}"
    api_url: "https://gogs.company.com/api/v1"
    
    organizations:
      - name: "legacy"
        clone_dir: "$HOME/repos/gogs/legacy"
        strategy: fetch
```

## Environment Variables

### Authentication Tokens

```bash
# GitHub
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# GitLab
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"

# Gitea
export GITEA_TOKEN="xxxxxxxxxxxx"

# Gogs
export GOGS_TOKEN="xxxxxxxxxxxx"
```

### Configuration Paths

```bash
# Override default configuration file location
export GZH_CONFIG_PATH="/custom/path/config.yaml"

# Debug mode
export GZH_DEBUG_SHELL=1

# IDE-specific
export JETBRAINS_CONFIG_PATH="/custom/jetbrains/config"
export IDE_MONITOR_INTERVAL="1s"

# Quality tools
export QUALITY_PARALLEL=true
export QUALITY_TIMEOUT=300
```

### Network Configuration

```bash
# Proxy settings
export PROXY_USERNAME="user"
export PROXY_PASSWORD="pass"
export HTTP_PROXY="http://proxy:8080"
export HTTPS_PROXY="http://proxy:8080"

# Cloud provider settings
export AWS_PROFILE="default"
export AWS_REGION="us-west-2"
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
```

## Configuration Management

### Validation

The configuration system includes comprehensive validation:

- **Schema Validation**: JSON Schema validation for structure
- **Field Validation**: Type checking and value constraints
- **Environment Variable Validation**: Checks for required environment variables
- **Path Validation**: Validates file and directory paths
- **Network Validation**: Validates URLs and network configurations

### Management Commands

```bash
# Validate configuration
gz config validate

# Show effective configuration after all priorities applied
gz config show

# Show configuration sources and their priority
gz config sources

# Create default configuration
gz config init

# Test configuration
gz config test

# Show configuration file search paths
gz config paths
```

### Debug Commands

```bash
# Show configuration loading process
gz config show --debug

# Validate with verbose output
gz config validate --verbose

# Show environment variable expansion
gz config show --expand-env
```

## Migration Guide

### From Legacy Formats

The system automatically migrates from legacy configuration formats:

- **synclone.yaml**: Automatically migrated to unified format
- **bulk-clone.yaml**: Integrated into unified configuration
- **Command-specific configs**: Integrated into unified configuration
- **Backup Creation**: Original files are backed up before migration
- **Migration Tracking**: Migration information is recorded in the configuration

### Format Comparison

#### Legacy synclone.yaml

```yaml
version: "0.1"
target: "./repos"
strategy: "reset"

github:
  organizations:
    - name: "myorg"
      target: "./github/myorg"
```

#### New gzh.yaml

```yaml
version: "1.0.0"
default_provider: github

global:
  clone_base_dir: "./repos"
  default_strategy: reset

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "./github/myorg"
```

### Migration Benefits

#### Immediate Benefits

- ✅ **Better Security**: Token-based authentication
- ✅ **Multi-Provider**: Support for GitLab, Gitea, Gogs
- ✅ **Better Validation**: Schema validation with clear error messages
- ✅ **Granular Control**: Per-organization settings

#### Advanced Benefits

- ✅ **Flexible Filtering**: Regex matching and visibility filters
- ✅ **Directory Control**: Flatten option for better organization
- ✅ **Update Strategies**: Choose how repositories are updated
- ✅ **Environment Integration**: Better environment variable support

### Migration Steps

1. **Backup existing configuration**:
   ```bash
   cp synclone.yaml synclone.yaml.backup
   ```

2. **Run migration tool** (if available):
   ```bash
   gz config migrate --from synclone.yaml --to gzh.yaml
   ```

3. **Manual migration**:
   - Convert structure to new format
   - Update authentication to use tokens
   - Migrate filtering options
   - Test new configuration

4. **Validate new configuration**:
   ```bash
   gz config validate --config gzh.yaml
   ```

## Best Practices

### 1. Security

- **Use environment variables for secrets**: Store tokens in environment variables
- **Restrict file permissions**: Ensure config files are not world-readable
- **Regular token rotation**: Rotate authentication tokens periodically

```bash
# Set secure permissions
chmod 600 ~/.config/gzh-manager/gzh.yaml

# Use environment variables for tokens
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
```

### 2. Organization

- **Separate environments**: Use different configs for dev/staging/prod
- **Version control**: Include config templates in version control (without secrets)
- **Documentation**: Document custom configurations and overrides

```yaml
# dev-config.yaml
providers:
  github:
    token: "${DEV_GITHUB_TOKEN}"
    organizations:
      - name: "dev-org"
        strategy: pull  # Allow local changes in dev

# prod-config.yaml
providers:
  github:
    token: "${PROD_GITHUB_TOKEN}"
    organizations:
      - name: "prod-org"
        strategy: reset  # Always match remote in prod
```

### 3. Performance

- **Tune concurrency**: Adjust parallel workers based on your environment
- **Use appropriate strategies**: Choose the right strategy for each use case
- **Monitor timeouts**: Adjust timeouts based on network conditions

```yaml
global:
  concurrency:
    clone_workers: 5     # Conservative for limited bandwidth
    update_workers: 10   # More aggressive for updates
    api_workers: 3       # Respect API rate limits
  timeouts:
    http_timeout: 60s    # Longer for slow connections
    git_timeout: 10m     # Longer for large repositories
```

### 4. Maintenance

- **Regular validation**: Validate configuration regularly
- **Monitor deprecations**: Watch for deprecated options
- **Update regularly**: Keep configuration format up to date

```bash
# Regular maintenance script
#!/bin/bash
gz config validate
gz config show --warnings
gz config update-schema  # Future feature
```

## Troubleshooting

### Common Issues

#### 1. Configuration Not Found

**Problem**: "configuration file not found"

**Solutions**:
- Check file paths and permissions
- Use `gz config paths` to see search locations
- Set `GZH_CONFIG_PATH` environment variable

#### 2. Environment Variables Not Expanded

**Problem**: `${GITHUB_TOKEN}` appears literally in logs

**Solutions**:
- Ensure environment variable is set: `echo $GITHUB_TOKEN`
- Check variable name spelling
- Use `gz config show --expand-env` to debug

#### 3. Validation Errors

**Problem**: "configuration validation failed"

**Solutions**:
- Use `gz config validate --verbose` for details
- Check schema documentation
- Verify required fields are present

#### 4. Permission Errors

**Problem**: "permission denied" when reading config

**Solutions**:
- Check file permissions: `ls -la ~/.config/gzh-manager/`
- Fix permissions: `chmod 644 ~/.config/gzh-manager/gzh.yaml`
- Check directory permissions

#### 5. Migration Errors

**Problem**: Legacy configuration not migrated

**Solutions**:
- Check legacy configuration format
- Run manual migration
- Review migration logs

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# Debug configuration loading
gz config show --debug

# Debug specific command
gz synclone github --org myorg --debug

# Debug with verbose output
gz --verbose config validate
```

### Common Pitfalls

1. **Environment variable expansion**: Remember that `${VAR}` in config files is expanded at runtime
2. **Configuration file precedence**: Files in current directory take precedence over user config
3. **Default value confusion**: Not all settings have the same default values across commands
4. **Token inheritance**: Different commands may use different token environment variables

## Command-Specific Configuration

### synclone Command

Uses the `providers` section and `global` settings:

```bash
# Use GitHub provider with specific organization
gz synclone github --org myorg

# Use configuration file
gz synclone --config gzh.yaml

# Override strategy
gz synclone github --org myorg --strategy pull
```

### ide Command

Uses the `ide` section:

```bash
# Enable IDE monitoring
gz ide monitor

# Monitor specific product
gz ide monitor --product IntelliJ
```

### dev-env Command

Uses the `dev_env` section:

```bash
# Backup AWS configuration
gz dev-env aws backup

# Restore Docker configuration
gz dev-env docker restore
```

### net-env Command

Uses the `net_env` section:

```bash
# Start network environment monitoring
gz net-env monitor

# Show current network status
gz net-env status
```

## Schema Reference

The configuration schema is defined in JSON Schema format and includes:

- **Type Definitions**: All configuration types and structures
- **Validation Rules**: Field constraints and validation rules
- **Examples**: Sample configurations for each section
- **Default Values**: Default values for all configuration options

For detailed schema documentation, see the [JSON Schema files](schemas/).

## Support

For additional help:

1. Run `gz config --help` for command options
2. Use `gz config validate` to check your configuration
3. Check the [examples directory](../../examples/) for sample configurations
4. Open an issue on [GitHub](https://github.com/gizzahub/gzh-cli/issues)