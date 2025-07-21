# Environment Variables

This document describes all environment variables used by gzh-manager.

## Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GZH_CONFIG_PATH` | Path to main configuration file | See [Configuration Priority](./priority-system.md) | `/etc/gzh/config.yaml` |
| `GZH_CONFIG_DIR` | Directory containing configuration files | `~/.config/gzh-manager` | `/opt/gzh/config` |
| `GZH_CLOUD_CONFIG` | Path to cloud configuration file | None | `/etc/gzh/cloud.yaml` |
| `GZH_BULK_CLONE_CONFIG` | Path to bulk-clone specific configuration | None | `./bulk-clone.yaml` |

## Authentication

### Standard Token Variables

These follow industry-standard naming conventions and are checked first:

| Variable | Description | Provider |
|----------|-------------|----------|
| `GITHUB_TOKEN` | GitHub personal access token | GitHub |
| `GITLAB_TOKEN` | GitLab personal access token | GitLab |
| `GITEA_TOKEN` | Gitea personal access token | Gitea |

### GZH-Prefixed Token Variables

Use these to avoid conflicts with other tools:

| Variable | Description | Provider |
|----------|-------------|----------|
| `GZH_GITHUB_TOKEN` | Alternative GitHub token | GitHub |
| `GZH_GITLAB_TOKEN` | Alternative GitLab token | GitLab |
| `GZH_GITEA_TOKEN` | Alternative Gitea token | Gitea |

**Note**: GZH-prefixed tokens take precedence over standard tokens.

## Logging and Debug

| Variable | Description | Values | Default |
|----------|-------------|--------|---------|
| `GZH_DEBUG` | Enable debug mode | `true`, `false` | `false` |
| `GZH_LOG_LEVEL` | Set logging level | `debug`, `info`, `warn`, `error` | `info` |
| `GZH_NO_COLOR` | Disable colored output | `true`, `false` | `false` |
| `GZH_PROGRESS_BAR` | Control progress bar display | `auto`, `always`, `never` | `auto` |

## Performance Tuning

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GZH_MAX_WORKERS` | Maximum concurrent operations | CPU cores | `8` |
| `GZH_TIMEOUT` | Default operation timeout | `30s` | `5m` |
| `GZH_RETRY_ATTEMPTS` | Number of retry attempts | `3` | `5` |
| `GZH_RATE_LIMIT` | API rate limit per hour | Provider default | `5000` |

## Network Settings

| Variable | Description | Example |
|----------|-------------|---------|
| `GZH_HTTP_PROXY` | HTTP proxy URL | `http://proxy.company.com:8080` |
| `GZH_HTTPS_PROXY` | HTTPS proxy URL | `https://proxy.company.com:8443` |
| `GZH_NO_PROXY` | Hosts to bypass proxy | `localhost,127.0.0.1,.company.com` |

## Provider API Endpoints

For self-hosted or enterprise instances:

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GZH_GITHUB_API` | GitHub API base URL | `https://api.github.com` | `https://github.company.com/api/v3` |
| `GZH_GITLAB_API` | GitLab API base URL | `https://gitlab.com` | `https://gitlab.company.com` |
| `GZH_GITEA_API` | Gitea API base URL | None | `https://gitea.company.com` |

## Command-Specific Variables

Some commands support additional environment variables:

### Bulk Clone

- `GZH_BULK_CLONE_CONFIG` - Path to bulk-clone configuration file

### Net Env

- `GZH_NET_ENV_CONFIG` - Path to network environment configuration

### Dev Env

- `GZH_DEV_ENV_CONFIG` - Path to development environment configuration

## Precedence Rules

1. Command-line flags always take precedence
2. GZH-prefixed environment variables
3. Standard environment variables
4. Configuration file values
5. Default values

## Examples

### Basic Usage

```bash
# Set GitHub token
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# Run bulk clone
gz bulk-clone github myorg
```

### Using GZH-Prefixed Variables

```bash
# Avoid conflicts with other tools
export GZH_GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GZH_CONFIG_PATH="/opt/gzh/config.yaml"
export GZH_MAX_WORKERS="16"

gz bulk-clone github myorg
```

### Debug Mode

```bash
# Enable debug logging
export GZH_DEBUG=true
export GZH_LOG_LEVEL=debug

gz doctor
```

### Proxy Configuration

```bash
# Configure proxy
export GZH_HTTP_PROXY="http://proxy:8080"
export GZH_HTTPS_PROXY="http://proxy:8080"
export GZH_NO_PROXY="localhost,127.0.0.1"

gz bulk-clone github myorg
```

## Best Practices

1. **Use GZH-prefixed variables** for gzh-specific settings to avoid conflicts
2. **Store tokens securely** - never commit them to version control
3. **Use configuration files** for complex setups instead of many env vars
4. **Document your setup** if using custom environment variables
5. **Test with GZH_DEBUG** when troubleshooting issues
