# Configuration Priority System

The gzh-manager tool implements a robust configuration priority system that ensures predictable behavior when multiple configuration sources are present.

## Priority Order (Highest to Lowest)

1. **Command-Line Flags** (Highest Priority)
2. **Environment Variables** (Second Priority)
3. **Configuration Files** (Third Priority)
4. **Default Values** (Lowest Priority)

## Detailed Priority Rules

### 1. Command-Line Flags (Highest Priority)

Command-line flags always take precedence over all other configuration sources. This ensures that explicit user intent via CLI arguments is always respected.

**Common CLI Flags:**

- `--config`, `-c`: Configuration file path
- `--strategy`, `-s`: Clone/sync strategy
- `--parallel`, `-p`: Number of parallel workers
- `--token`, `-t`: Authentication token
- `--provider`: Git provider filter
- `--dry-run`: Preview mode

**Example:**

```bash
# All flags override any configuration file or environment variable
gz bulk-clone --strategy=pull --parallel=20 --token=ghp_override_token
```

### 2. Environment Variables (Second Priority)

Environment variables provide a way to configure the tool without modifying files, but they are overridden by command-line flags.

**Key Environment Variables:**

- `GZH_CONFIG_PATH`: Override default configuration file location
- `GITHUB_TOKEN`: GitHub API authentication token
- `GITLAB_TOKEN`: GitLab API authentication token
- `GITEA_TOKEN`: Gitea API authentication token
- `GOGS_TOKEN`: Gogs API authentication token

**Example:**

```bash
# Environment variable sets the default
export GITHUB_TOKEN=ghp_env_token
gz bulk-clone  # Uses ghp_env_token

# CLI flag overrides environment variable
gz bulk-clone --token=ghp_flag_token  # Uses ghp_flag_token
```

### 3. Configuration Files (Third Priority)

Configuration files provide the base configuration but can be overridden by environment variables and command-line flags.

**Configuration File Search Order:**

1. Path specified by `GZH_CONFIG_PATH` environment variable
2. Path specified by `--config` CLI flag
3. Current directory: `./gzh.yaml`, `./gzh.yml`
4. User config: `~/.config/gzh-manager/gzh.yaml`
5. System config: `/etc/gzh-manager/gzh.yaml`
6. Legacy files: `./bulk-clone.yaml`, `./bulk-clone.yml` (auto-migrated)

**Example:**

```yaml
# gzh.yaml
version: "1.0.0"
global:
  default_strategy: reset
  concurrency:
    clone_workers: 10
providers:
  github:
    token: "${GITHUB_TOKEN}"
```

### 4. Default Values (Lowest Priority)

Default values are hardcoded in the application and used when no higher priority source provides a value.

**Common Defaults:**

- `strategy`: `reset`
- `parallel`: `10`
- `visibility`: `all`
- `timeout`: `30s`
- `clone_base_dir`: `~/repos`

## Priority Resolution Examples

### Example 1: Token Configuration

**Sources:**

```yaml
# ~/.config/gzh-manager/gzh.yaml
providers:
  github:
    token: "ghp_config_token"
```

```bash
# Environment
export GITHUB_TOKEN=ghp_env_token

# Command
gz bulk-clone --token=ghp_flag_token
```

**Resolution:** `ghp_flag_token` (CLI flag has highest priority)

### Example 2: Strategy Configuration

**Sources:**

```yaml
# gzh.yaml
global:
  default_strategy: "reset"
```

```bash
# No environment variable for strategy
gz bulk-clone --strategy=pull
```

**Resolution:** `pull` (CLI flag overrides config file)

### Example 3: Parallel Workers

**Sources:**

```yaml
# gzh.yaml
global:
  concurrency:
    clone_workers: 15
```

```bash
# No CLI flag or environment variable
gz bulk-clone
```

**Resolution:** `15` (from configuration file)

### Example 4: Configuration File Location

**Sources:**

```bash
# Environment variable
export GZH_CONFIG_PATH=/custom/path/config.yaml

# CLI flag
gz bulk-clone --config=/override/path/config.yaml
```

**Resolution:** `/override/path/config.yaml` (CLI flag overrides environment variable)

## Environment Variable Expansion

Configuration files support environment variable expansion using `${VAR_NAME}` syntax:

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    api_url: "${GITHUB_API_URL:-https://api.github.com}" # With default value
```

**Priority for expanded variables:**

1. Command-line flags (if applicable)
2. Environment variables (used in expansion)
3. Default values in expansion syntax (`${VAR:-default}`)
4. Configuration file literal values

## Command-Specific Priority

Different commands may have different configuration sections but follow the same priority rules:

### bulk-clone Command

- CLI flags: `--strategy`, `--parallel`, `--token`, `--provider`
- Environment: `GITHUB_TOKEN`, `GITLAB_TOKEN`, `GITEA_TOKEN`
- Config: `global.default_strategy`, `global.concurrency.clone_workers`

### ide Command

- CLI flags: `--ide-type`, `--sync-interval`, `--backup-enabled`
- Environment: `GZH_IDE_TYPE`, `GZH_SYNC_INTERVAL`
- Config: `ide.enabled`, `ide.sync_interval`, `ide.backup_enabled`

### dev-env Command

- CLI flags: `--profile`, `--auto-switch`
- Environment: `GZH_DEV_PROFILE`, `AWS_PROFILE`
- Config: `dev_env.enabled`, `dev_env.profiles`

### net-env Command

- CLI flags: `--wifi-monitor`, `--auto-switch`
- Environment: `GZH_WIFI_MONITOR`
- Config: `net_env.enabled`, `net_env.wifi_detection`

## Debugging Configuration Priority

Use these commands to understand which configuration sources are being used:

```bash
# Show effective configuration after all priorities applied
gz config show

# Show configuration sources and their priority
gz config sources

# Validate current configuration
gz config validate

# Show configuration file search paths
gz config paths
```

## Best Practices

1. **Use configuration files for defaults**: Set your most common settings in configuration files
2. **Use environment variables for secrets**: Keep tokens and sensitive data in environment variables
3. **Use CLI flags for overrides**: Override specific settings for individual commands
4. **Test priority resolution**: Use `gz config show` to verify your configuration is resolved correctly
5. **Keep it simple**: Don't over-complicate your configuration hierarchy

## Common Pitfalls

1. **Environment variable expansion**: Remember that `${VAR}` in config files is expanded at runtime
2. **Configuration file precedence**: Files in current directory take precedence over user config
3. **Default value confusion**: Not all settings have the same default values across commands
4. **Token inheritance**: Different commands may use different token environment variables

## Migration from Legacy Configuration

Legacy `bulk-clone.yaml` files are automatically migrated to the new unified format, but priority rules still apply:

1. CLI flags override legacy config values
2. Environment variables override legacy config values
3. New unified config files take precedence over legacy files
4. Migration preserves existing priority behavior

For detailed migration guidance, see [Configuration Migration Guide](configuration-migration.md).
