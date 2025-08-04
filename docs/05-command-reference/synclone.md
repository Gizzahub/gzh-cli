# synclone Command Reference

Multi-platform repository synchronization with support for GitHub, GitLab, Gitea, and Gogs.

## Synopsis

```bash
gz synclone [platform] [flags]
gz synclone --config <config-file>
```

## Description

The `synclone` command provides powerful repository synchronization capabilities across multiple Git hosting platforms. It can clone entire organizations/groups and keep them synchronized with configurable strategies.

## Subcommands

### `gz synclone github`

Clone and synchronize GitHub organizations.

```bash
gz synclone github --org <organization> [flags]
```

**Flags:**
- `--org`, `-o` - Organization name (required)
- `--target`, `-t` - Target directory (default: current directory)
- `--strategy` - Clone/update strategy: reset, pull, fetch, rebase, clone, skip (default: reset)
- `--include-archived` - Include archived repositories (default: false)
- `--include-forks` - Include forked repositories (default: false)
- `--include-private` - Include private repositories (default: true)
- `--language` - Filter by programming language
- `--topic` - Filter by repository topic
- `--min-stars` - Minimum star count
- `--parallel` - Number of concurrent operations (default: 5)
- `--shallow` - Use shallow clones (default: false)
- `--depth` - Depth for shallow clones (default: 1)

**Examples:**
```bash
# Basic organization clone
gz synclone github --org kubernetes

# Clone with filters
gz synclone github --org prometheus --language Go --min-stars 100

# Clone to specific directory with custom strategy
gz synclone github --org myorg --target ./repos --strategy pull

# Shallow clone for space efficiency
gz synclone github --org large-org --shallow --depth 1
```

### `gz synclone gitlab`

Clone and synchronize GitLab groups.

```bash
gz synclone gitlab --group <group> [flags]
```

**Flags:**
- `--group`, `-g` - Group ID or path (required)
- `--target`, `-t` - Target directory
- `--strategy` - Clone/update strategy
- `--include-subgroups` - Include subgroups recursively (default: false)
- `--visibility` - Filter by visibility: public, internal, private
- `--base-url` - GitLab instance URL (default: https://gitlab.com)
- `--parallel` - Number of concurrent operations

**Examples:**
```bash
# Clone GitLab group
gz synclone gitlab --group gitlab-org

# Include subgroups
gz synclone gitlab --group mygroup --include-subgroups

# Self-hosted GitLab
gz synclone gitlab --group mygroup --base-url https://gitlab.company.com
```

### `gz synclone gitea`

Clone and synchronize Gitea organizations.

```bash
gz synclone gitea --org <organization> --base-url <url> [flags]
```

**Flags:**
- `--org`, `-o` - Organization name (required)
- `--base-url` - Gitea instance URL (required)
- `--target`, `-t` - Target directory
- `--strategy` - Clone/update strategy
- `--token` - Authentication token

**Examples:**
```bash
# Clone Gitea organization
gz synclone gitea --org myorg --base-url https://gitea.company.com

# With authentication
gz synclone gitea --org myorg --base-url https://gitea.company.com --token $GITEA_TOKEN
```

### `gz synclone gogs`

Clone and synchronize Gogs organizations.

```bash
gz synclone gogs --org <organization> --base-url <url> [flags]
```

**Flags:**
- `--org`, `-o` - Organization name (required)
- `--base-url` - Gogs instance URL (required)
- `--target`, `-t` - Target directory
- `--strategy` - Clone/update strategy

## Configuration File Usage

Use a configuration file for complex setups:

```bash
gz synclone --config synclone.yaml
```

**Configuration file structure:**
```yaml
version: "1.0"
target: "./repos"
strategy: "reset"
parallel: 5

github:
  enabled: true
  organizations:
    - name: "kubernetes"
      target: "./k8s-repos"
      filters:
        languages: ["Go"]
        min_stars: 100

gitlab:
  enabled: true
  base_url: "https://gitlab.com"
  groups:
    - id: "gitlab-org"
      include_subgroups: true
```

## Global Flags

- `--config` - Configuration file path
- `--dry-run` - Show what would be done without executing
- `--force-refresh` - Ignore cache and fetch fresh data
- `--quiet`, `-q` - Suppress non-error output
- `--verbose`, `-v` - Enable verbose output
- `--debug` - Enable debug logging

## Clone Strategies

| Strategy | Behavior | Use Case |
|----------|----------|----------|
| `reset` | Hard reset to match remote (default) | CI/CD, mirrors |
| `pull` | Merge remote changes | Active development |
| `fetch` | Update refs only | Inspection |
| `rebase` | Rebase local changes on remote | Clean history |
| `clone` | Fresh clone (removes existing) | Clean start |
| `skip` | Skip existing repositories | Initial clone only |

## Authentication

### Environment Variables

```bash
export GITHUB_TOKEN="ghp_..."
export GITLAB_TOKEN="glpat-..."
export GITEA_TOKEN="..."
export GOGS_TOKEN="..."
```

### Configuration File

```yaml
github:
  organizations:
    - name: "myorg"
      token: "${GITHUB_TOKEN}"  # Environment variable reference
      # or
      token: "ghp_direct_token"  # Direct token (not recommended)
```

## Filtering Options

### Language Filtering

```bash
# Single language
gz synclone github --org myorg --language Go

# Multiple languages
gz synclone github --org myorg --language Go,Python,JavaScript
```

### Topic Filtering

```bash
# Single topic
gz synclone github --org myorg --topic kubernetes

# Multiple topics (AND logic)
gz synclone github --org myorg --topic kubernetes,cloud-native
```

### Size and Activity

```bash
# Star count filtering
gz synclone github --org myorg --min-stars 50

# Recently updated repositories
gz synclone github --org myorg --updated-after 2024-01-01
```

## Examples

### Personal Repository Management

```bash
# Clone your personal repositories
gz synclone github --org yourusername --target ~/repos/personal

# Clone work repositories with SSH
gz synclone github --org company --target ~/repos/work --strategy pull
```

### Multi-Platform Setup

```yaml
# synclone.yaml
version: "1.0"
target: "./repos"

github:
  enabled: true
  organizations:
    - name: "kubernetes"
      target: "./github/kubernetes"

gitlab:
  enabled: true
  groups:
    - id: "gitlab-org"
      target: "./gitlab/gitlab-org"

gitea:
  enabled: true
  base_url: "https://gitea.company.com"
  organizations:
    - name: "internal"
      target: "./gitea/internal"
```

### Enterprise Deployment

```bash
# Large-scale deployment with specific filters
gz synclone github \
  --org enterprise-org \
  --target /data/repositories \
  --language Go,Python,JavaScript \
  --min-stars 10 \
  --parallel 10 \
  --strategy reset
```

## Error Handling

### Common Errors

1. **Authentication Failed**
   ```
   Error: authentication failed for GitHub
   ```
   **Solution:** Check token validity and permissions

2. **Rate Limited**
   ```
   Error: rate limit exceeded
   ```
   **Solution:** Wait or reduce parallelism with `--parallel 1`

3. **Network Timeout**
   ```
   Error: context deadline exceeded
   ```
   **Solution:** Check network connectivity or increase timeout

### Retry and Recovery

```bash
# Retry failed operations
gz synclone github --org myorg --retry 3

# Continue on errors
gz synclone github --org myorg --continue-on-error

# Skip problematic repositories
gz synclone github --org myorg --skip-on-error
```

## Performance Tuning

### Concurrency

```bash
# High performance (good network)
gz synclone github --org myorg --parallel 10

# Conservative (limited resources)
gz synclone github --org myorg --parallel 2
```

### Shallow Clones

```bash
# Save disk space and time
gz synclone github --org myorg --shallow --depth 1

# Recent history only
gz synclone github --org myorg --shallow --depth 10
```

### Caching

```bash
# Use cached repository metadata
gz synclone github --org myorg --cache-ttl 1h

# Force refresh
gz synclone github --org myorg --force-refresh
```

## Validation

```bash
# Validate configuration file
gz synclone validate --config synclone.yaml

# Test connection and authentication
gz synclone test --config synclone.yaml

# Dry run to preview operations
gz synclone --config synclone.yaml --dry-run
```

## Integration

### CI/CD

```bash
# CI-friendly execution
export GITHUB_TOKEN="${CI_GITHUB_TOKEN}"
gz synclone github --org myorg --quiet --strategy reset
```

### Scripting

```bash
#!/bin/bash
set -e

# Backup script
gz synclone --config backup.yaml --log-file "backup-$(date +%Y%m%d).log"

# Check exit code
if [ $? -eq 0 ]; then
    echo "Backup completed successfully"
else
    echo "Backup failed" >&2
    exit 1
fi
```

## Related Commands

- [`gz git repo clone-or-update`](git.md#repo-clone-or-update) - Single repository operations
- [`gz git config`](git.md#config) - Repository configuration management
- [`gz profile`](profile.md) - Performance profiling

## See Also

- [Repository Synchronization Guide](../03-core-features/synclone-guide.md)
- [Configuration Schema](../04-configuration/schemas/synclone-schema.yaml)
- [Examples](../../examples/synclone/)
