# Repository Synchronization Guide (synclone)

The `gz synclone` command provides powerful multi-platform repository synchronization capabilities, allowing you to clone and manage entire organizations or groups across GitHub, GitLab, Gitea, and Gogs.

## Overview

Synclone (synchronize + clone) is designed for developers and teams who need to:
- Clone all repositories from an organization or group
- Keep local repositories synchronized with remote changes
- Manage repositories across multiple Git hosting platforms
- Automate repository backup and mirroring workflows

## Supported Platforms

- **GitHub**: Public and Enterprise Server
- **GitLab**: SaaS and Self-hosted
- **Gitea**: Self-hosted Git service
- **Gogs**: Lightweight Git service

## Quick Start

### Basic Usage

```bash
# Clone all repositories from a GitHub organization
gz synclone github --org myorganization

# Clone from GitLab group
gz synclone gitlab --group mygroup

# Clone from Gitea organization
gz synclone gitea --org myorg --base-url https://gitea.company.com

# Use configuration file
gz synclone --config synclone.yaml
```

### Authentication

Set up authentication tokens:

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

## Configuration

### Configuration File Structure

Create a `synclone.yaml` file:

```yaml
# Global settings
target: "./repos"  # Base directory for all clones
strategy: "reset"  # Default strategy: reset, pull, fetch, rebase
shallow: false     # Use shallow clones
depth: 1          # Depth for shallow clones (when shallow: true)
parallel: 5       # Number of concurrent operations

# Platform configurations
github:
  enabled: true
  organizations:
    - name: "kubernetes"
      target: "./k8s-repos"
      include_archived: false
      include_forks: false
      filters:
        languages: ["Go", "Shell"]
        topics: ["kubernetes", "cloud-native"]
    - name: "prometheus"
      strategy: "pull"  # Override global strategy

gitlab:
  enabled: true
  base_url: "https://gitlab.com"
  groups:
    - id: "gitlab-org"
      target: "./gitlab-repos"
      include_subgroups: true
      visibility: "public"  # public, internal, private

gitea:
  enabled: true
  base_url: "https://gitea.company.com"
  organizations:
    - name: "infrastructure"
      target: "./infra-repos"

gogs:
  enabled: false
  base_url: "https://gogs.company.com"
  organizations:
    - name: "legacy"
      target: "./legacy-repos"
```

### Configuration Hierarchy

1. Environment variables (highest priority)
2. Command-line flags
3. Configuration file
4. Default values

### Advanced Configuration

```yaml
# Advanced example with all options
github:
  enabled: true
  base_url: "https://github.enterprise.com"  # For GitHub Enterprise
  api_version: "v3"
  organizations:
    - name: "myorg"
      target: "./github/myorg"

      # Authentication
      token: "${GITHUB_ENTERPRISE_TOKEN}"  # Environment variable reference

      # Clone settings
      strategy: "reset"
      branch: "main"  # Default branch to check out
      shallow: true
      depth: 10

      # Filtering
      include_archived: false
      include_forks: false
      include_private: true
      include_templates: false

      # Advanced filters
      filters:
        languages: ["Go", "Python", "JavaScript"]
        topics: ["microservice", "api"]
        name_pattern: "^service-.*"  # Regex pattern
        exclude_pattern: ".*-deprecated$"
        min_stars: 10
        min_size: 100  # KB
        max_size: 1000000  # KB
        updated_after: "2024-01-01"

      # Repository-specific overrides
      repositories:
        - name: "critical-service"
          strategy: "pull"  # Never reset this repo
          branch: "production"
        - name: "experimental-feature"
          skip: true  # Don't clone this repo

      # Hooks
      hooks:
        pre_clone: "scripts/pre-clone.sh"
        post_clone: "scripts/post-clone.sh"
        on_error: "scripts/on-error.sh"
```

## Clone Strategies

### reset (default)
Hard reset to match remote state, discarding all local changes:
```bash
gz synclone github --org myorg --strategy reset
```
Best for: Read-only mirrors, CI/CD environments

### pull
Merge remote changes with local changes:
```bash
gz synclone github --org myorg --strategy pull
```
Best for: Active development with local modifications

### fetch
Only update remote references without changing working tree:
```bash
gz synclone github --org myorg --strategy fetch
```
Best for: Inspecting changes before merging

### rebase
Rebase local changes on top of remote changes:
```bash
gz synclone github --org myorg --strategy rebase
```
Best for: Maintaining linear history with local commits

### clone
Always perform fresh clone (removes existing directory):
```bash
gz synclone github --org myorg --strategy clone
```
Best for: Clean environments, resolving corruption

### skip
Skip existing repositories:
```bash
gz synclone github --org myorg --strategy skip
```
Best for: Initial bulk clone without updates

## Command Reference

### GitHub Operations

```bash
# Basic organization clone
gz synclone github --org myorg

# With custom target directory
gz synclone github --org myorg --target ./my-repos

# Include archived repositories
gz synclone github --org myorg --include-archived

# Filter by language
gz synclone github --org myorg --language Go

# Filter by topic
gz synclone github --org myorg --topic kubernetes

# Multiple organizations
gz synclone github --org org1,org2,org3

# GitHub Enterprise
gz synclone github --org myorg --base-url https://github.company.com
```

### GitLab Operations

```bash
# Clone GitLab group
gz synclone gitlab --group mygroup

# Include subgroups
gz synclone gitlab --group mygroup --include-subgroups

# Filter by visibility
gz synclone gitlab --group mygroup --visibility private

# Self-hosted GitLab
gz synclone gitlab --group mygroup --base-url https://gitlab.company.com
```

### Gitea Operations

```bash
# Clone Gitea organization
gz synclone gitea --org myorg --base-url https://gitea.company.com

# With authentication
gz synclone gitea --org myorg --token $GITEA_TOKEN
```

### Multi-Platform Operations

```bash
# Use configuration file
gz synclone --config synclone.yaml

# Validate configuration
gz synclone validate --config synclone.yaml

# Dry run (show what would be cloned)
gz synclone --config synclone.yaml --dry-run

# Force refresh (ignore cache)
gz synclone --config synclone.yaml --force-refresh
```

## Filtering and Selection

### Language Filters

```bash
# Single language
gz synclone github --org myorg --language Go

# Multiple languages
gz synclone github --org myorg --language Go,Python,JavaScript
```

### Topic Filters

```bash
# Single topic
gz synclone github --org myorg --topic microservices

# Multiple topics (AND logic)
gz synclone github --org myorg --topic kubernetes,production
```

### Pattern Matching

```bash
# Name pattern (regex)
gz synclone github --org myorg --name-pattern "^api-.*"

# Exclude pattern
gz synclone github --org myorg --exclude-pattern ".*-deprecated$"
```

### Size and Activity Filters

```bash
# Minimum stars
gz synclone github --org myorg --min-stars 50

# Size range (in KB)
gz synclone github --org myorg --min-size 100 --max-size 50000

# Recently updated
gz synclone github --org myorg --updated-after 2024-01-01
```

## Progress and Monitoring

### Progress Display

```bash
$ gz synclone github --org kubernetes
🔍 Fetching repositories from kubernetes...
📊 Found 147 repositories matching criteria

Cloning repositories:
[##########··········] 50% | 73/147 | kubernetes/dashboard ⠋
✓ kubernetes/kubernetes (25.3 MB)
✓ kubernetes/minikube (18.7 MB)
⠼ kubernetes/dashboard (5.2 MB)
⠿ kubernetes/client-go (queued)

Summary:
✅ Successful: 145
⚠️  Warnings: 1
❌ Failed: 1
⏱️  Duration: 4m 23s
```

### Verbose Output

```bash
# Show detailed progress
gz synclone github --org myorg --verbose

# Show debug information
gz synclone github --org myorg --debug
```

### Logging

```bash
# Log to file
gz synclone github --org myorg --log-file synclone.log

# JSON format logging
gz synclone github --org myorg --log-format json
```

## Error Handling

### Retry Logic

```bash
# Set retry attempts
gz synclone github --org myorg --retry 5

# Set retry delay
gz synclone github --org myorg --retry-delay 10s
```

### Error Recovery

```bash
# Continue on errors
gz synclone github --org myorg --continue-on-error

# Skip problematic repos
gz synclone github --org myorg --skip-on-error
```

### Common Issues

1. **Authentication Failures**
   ```bash
   # Verify token
   gz synclone github --org myorg --verify-auth

   # Use different token
   gz synclone github --org myorg --token $BACKUP_TOKEN
   ```

2. **Rate Limiting**
   ```bash
   # Reduce parallelism
   gz synclone github --org myorg --parallel 2

   # Add delay between requests
   gz synclone github --org myorg --request-delay 1s
   ```

3. **Network Issues**
   ```bash
   # Increase timeout
   gz synclone github --org myorg --timeout 300s

   # Use proxy
   gz synclone github --org myorg --proxy http://proxy:8080
   ```

## Best Practices

### 1. Organization Structure

```
repos/
├── github/
│   ├── kubernetes/
│   ├── prometheus/
│   └── grafana/
├── gitlab/
│   ├── gitlab-org/
│   └── gnome/
└── internal/
    ├── gitea/
    └── gogs/
```

### 2. Backup Strategy

```bash
# Daily backup script
#!/bin/bash
gz synclone --config backup.yaml \
  --strategy fetch \
  --log-file logs/backup-$(date +%Y%m%d).log
```

### 3. CI/CD Integration

```yaml
# GitHub Actions example
- name: Sync repositories
  run: |
    gz synclone --config synclone.yaml \
      --strategy reset \
      --continue-on-error
```

### 4. Performance Optimization

```bash
# Use shallow clones for large repos
gz synclone github --org myorg --shallow --depth 1

# Limit parallelism for stability
gz synclone github --org myorg --parallel 3

# Cache repository metadata
gz synclone github --org myorg --cache-ttl 1h
```

## Migration from bulk-clone

If you're migrating from the old `bulk-clone` command:

```bash
# Old command
gz bulk-clone --org myorg

# New command
gz synclone github --org myorg

# Convert old config
gz synclone migrate --old-config bulk-clone.yaml --output synclone.yaml
```

Key differences:
- Platform-specific subcommands (github, gitlab, etc.)
- More granular filtering options
- Better progress reporting
- Improved error handling
- Support for multiple strategies

## Related Documentation

- [Configuration Guide](../04-configuration/configuration-guide.md)
- [Git Unified Command](git-unified-command.md)
- [Repository Management](repository-management/)
- [Migration Guide](../01-getting-started/migration-guides/bulk-clone-to-gzh.md)
