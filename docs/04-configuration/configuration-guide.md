# Configuration System

The gzh-manager-go project uses a unified configuration system that supports all commands through a single `gzh.yaml` configuration file.

## Configuration Priority

The configuration system follows a strict priority order where higher priority sources override lower priority ones:

### Priority Order (Highest to Lowest)

1. **Command-Line Flags** (Highest Priority)
2. **Environment Variables** (Second Priority)
3. **Configuration Files** (Third Priority)
4. **Default Values** (Lowest Priority)

> **📖 For comprehensive priority documentation, see [Configuration Priority Guide](configuration-priority.md)**

### Detailed Priority Explanation

#### 1. Command-Line Flags (Highest Priority)

Command-line flags always take precedence over all other configuration sources. When a flag is specified, it overrides any corresponding setting from environment variables, configuration files, or default values.

**Examples:**

```bash
# Override configuration file strategy
gz bulk-clone --strategy=pull

# Override environment variable token
gz bulk-clone --token=ghp_custom_token

# Override configuration file parallel workers
gz bulk-clone --parallel=20
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
gz bulk-clone  # Uses ghp_env_token

# But command-line flag overrides environment variable
gz bulk-clone --token=ghp_flag_token  # Uses ghp_flag_token
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
gz bulk-clone --token=ghp_flag_token
```

**Resolution:** `ghp_flag_token` (CLI flag wins)

#### Example 2: Strategy Resolution

```yaml
# config.yaml
global:
  default_strategy: "reset"
```

```bash
# Environment variable (not applicable for strategy)
# Command execution
gz bulk-clone --strategy=pull
```

**Resolution:** `pull` (CLI flag overrides config file)

#### Example 3: Parallel Workers Resolution

```yaml
# config.yaml
global:
  concurrency:
    clone_workers: 15
```

```bash
# No environment variable for parallel workers
# Command execution
gz bulk-clone  # No --parallel flag
```

**Resolution:** `15` (from configuration file)

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

### Configuration File Search Order

When no explicit config path is provided, the system searches in this order:

1. **Environment Variable**: `GZH_CONFIG_PATH`
2. **Current Directory**: `./gzh.yaml`, `./gzh.yml`
3. **User Config**: `~/.config/gzh-manager/gzh.yaml`
4. **System Config**: `/etc/gzh-manager/gzh.yaml`
5. **Legacy Files**: `./bulk-clone.yaml`, `./bulk-clone.yml` (auto-migrated)

### Debugging Configuration Priority

Use these commands to understand configuration resolution:

```bash
# Show effective configuration after all priorities applied
gz config show

# Show configuration sources being used
gz config sources

# Validate current configuration
gz config validate
```

## Configuration File Hierarchy

Configuration files are searched in the following order:

1. **Environment Variable**: `GZH_CONFIG_PATH`
2. **Current Directory**: `./gzh.yaml`, `./gzh.yml`, `./config.yaml`, `./config.yml`
3. **User Config**: `~/.config/gzh-manager/gzh.yaml`
4. **System Config**: `/etc/gzh-manager/gzh.yaml`
5. **Legacy Files**: `./bulk-clone.yaml`, `./bulk-clone.yml` (automatically migrated)

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
    rotation:
      max_size_mb: 10
      max_backups: 5
      max_age_days: 30
      compress: true

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
      enable_mfa: false
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
      use_managed_identity: false
  containers:
    default_runtime: docker
    docker:
      socket_path: /var/run/docker.sock
      default_registry: docker.io
      registry_auth:
        docker.io:
          username: "${DOCKER_USERNAME}"
          password: "${DOCKER_PASSWORD}"
      build_options:
        default_context: "."
        enable_buildkit: true
  kubernetes:
    kubeconfig_path: "$HOME/.kube/config"
    default_namespace: default
    auto_discovery: true
  backup:
    enabled: false
    interval: 24h
    retention_period: 720h  # 30 days
    compression: gzip
    destinations:
      - "/backup/location"
    encryption:
      enabled: false
      method: aes256

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
        on_connect:
          - "sync-time"
      "Office-WiFi":
        ssid: "Office-WiFi"
        type: work
        vpn_config: "work-vpn"
        dns_servers:
          - "10.0.0.1"
        on_connect:
          - "connect-vpn"
          - "sync-time"
    default_actions:
      - "update-dns"
      - "check-vpn"
  vpn:
    profiles:
      work-vpn:
        type: openvpn
        config_file: "$HOME/.config/vpn/work.ovpn"
        connect_command: "openvpn --config $HOME/.config/vpn/work.ovpn"
        disconnect_command: "pkill openvpn"
        auto_connect_networks:
          - "Office-WiFi"
    default_profile: work-vpn
    auto_connect:
      enabled: true
      on_untrusted_networks: true
      trusted_networks:
        - "Home-WiFi"
        - "Office-WiFi"
      retry_attempts: 3
      retry_delay: 5s
  dns:
    default_servers:
      - "1.1.1.1"
      - "1.0.0.1"
    enable_doh: false
    doh_provider: cloudflare
    profiles:
      home:
        servers:
          - "192.168.1.1"
        search_domains:
          - "local"
      work:
        servers:
          - "10.0.0.1"
          - "10.0.0.2"
        search_domains:
          - "company.com"
  proxy:
    profiles:
      corporate:
        type: http
        host: "proxy.company.com"
        port: 8080
        username: "${PROXY_USERNAME}"
        password: "${PROXY_PASSWORD}"
        no_proxy:
          - "localhost"
          - "127.0.0.1"
          - "*.company.com"
    auto_configure: false
  actions:
    on_network_change:
      - "update-dns"
      - "check-vpn"
    on_wifi_connect:
      - "sync-time"
    on_wifi_disconnect:
      - "pause-sync"
    on_vpn_connect:
      - "update-routes"
    on_vpn_disconnect:
      - "restore-routes"
    custom_actions:
      sync-time:
        name: "Sync system time"
        command: "ntpdate -s time.nist.gov"
        timeout: 30s
        retry:
          max_attempts: 3
          delay: 5s
      update-dns:
        name: "Update DNS configuration"
        command: "systemctl restart systemd-resolved"
        run_as_user: root
        timeout: 10s
  daemon:
    enabled: false
    pid_file: "/var/run/gzh-manager-netenv.pid"
    log_file: "/var/log/gzh-manager-netenv.log"
    log_level: info
    systemd_integration: true

```

## Environment Variable Support

All configuration values support environment variable expansion using the `${VAR_NAME}` syntax:

- `${GITHUB_TOKEN}` - GitHub personal access token
- `${GITLAB_TOKEN}` - GitLab personal access token
- `${HOME}` - User home directory
- `${USER}` - Current username
- `${PWD}` - Current working directory

## Configuration Validation

The configuration system includes comprehensive validation:

- **Schema Validation**: JSON Schema validation for structure
- **Field Validation**: Type checking and value constraints
- **Environment Variable Validation**: Checks for required environment variables
- **Path Validation**: Validates file and directory paths
- **Network Validation**: Validates URLs and network configurations

## Migration from Legacy Formats

The system automatically migrates from legacy configuration formats:

- **bulk-clone.yaml**: Automatically migrated to unified format
- **Command-specific configs**: Integrated into unified configuration
- **Backup Creation**: Original files are backed up before migration
- **Migration Tracking**: Migration information is recorded in the configuration

## Command-Specific Configuration

### bulk-clone Command

Uses the `providers` section and `global` settings:

```bash
# Use GitHub provider with specific organization
gz bulk-clone github --org myorg

# Use configuration file
gz bulk-clone --use-config

# Override strategy
gz bulk-clone --strategy pull
```

### ide Command

Uses the `ide` section:

```bash
# Enable IDE sync
gz ide sync

# Sync IDE settings
gz ide sync --product IntelliJ
```

### dev-env Command

Uses the `dev_env` section:

```bash
# Backup AWS configuration
gz dev-env backup aws

# Restore Docker configuration
gz dev-env restore docker
```

### net-env Command

Uses the `net_env` section:

```bash
# Start network environment daemon
gz net-env daemon start

# Show current network status
gz net-env status
```


## Best Practices

1. **Use Environment Variables**: Store sensitive data in environment variables
2. **Version Control**: Include configuration in version control (without secrets)
3. **Backup Configuration**: Regular backups of configuration files
4. **Validate Configuration**: Use `gz config validate` to check configuration
5. **Watch Changes**: Enable configuration file watching for automatic reloads
6. **Documentation**: Document custom configurations and overrides

## Configuration Management Commands

```bash
# Validate configuration
gz config validate

# Show configuration summary
gz config show

# Create default configuration
gz config init

# Test configuration
gz config test
```

## Troubleshooting

### Common Issues

1. **Configuration Not Found**: Check file paths and permissions
2. **Environment Variables**: Ensure required variables are set
3. **Migration Errors**: Check legacy configuration format
4. **Validation Errors**: Use `gz config validate` for details
5. **Permission Errors**: Check file and directory permissions

### Debug Commands

```bash
# Show configuration loading process
gz config show --debug

# Validate with verbose output
gz config validate --verbose

# Show environment variable expansion
gz config show --expand-env
```

## Schema Reference

The configuration schema is defined in JSON Schema format and includes:

- **Type Definitions**: All configuration types and structures
- **Validation Rules**: Field constraints and validation rules
- **Examples**: Sample configurations for each section
- **Default Values**: Default values for all configuration options

For detailed schema documentation, see the [JSON Schema files](../docs/schema/).
