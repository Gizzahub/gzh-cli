# Repository Synchronization Guide (synclone)

Comprehensive guide for multi-platform repository synchronization using `gz synclone` command.

## Table of Contents

1. [Overview](#overview)
1. [Quick Start](#quick-start)
1. [Command Reference](#command-reference)
1. [Configuration System](#configuration-system)
1. [Clone Strategies](#clone-strategies)
1. [Platform Support](#platform-support)
1. [Filtering and Selection](#filtering-and-selection)
1. [Performance and Monitoring](#performance-and-monitoring)
1. [Best Practices](#best-practices)
1. [Migration Guide](#migration-guide)
1. [Troubleshooting](#troubleshooting)

## Overview

The `gz synclone` command provides powerful multi-platform repository synchronization capabilities, allowing you to clone and manage entire organizations or groups across GitHub, GitLab, Gitea, and Gogs. Synclone (synchronize + clone) is designed for developers and teams who need to:

- Clone all repositories from an organization or group
- Keep local repositories synchronized with remote changes
- Manage repositories across multiple Git hosting platforms
- Automate repository backup and mirroring workflows

### Key Features

- **Multi-Platform Support**: GitHub, GitLab, Gitea, and Gogs
- **Flexible Strategies**: Multiple clone/update strategies for different use cases
- **Advanced Filtering**: Language, topic, size, activity, and pattern-based filters
- **Parallel Operations**: Concurrent cloning for improved performance
- **Progress Monitoring**: Real-time progress tracking and detailed reporting
- **Error Recovery**: Robust retry logic and error handling
- **Configuration Management**: YAML-based configuration with environment variable support

### Supported Platforms

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

### Authentication Setup

Set up authentication tokens for the platforms you use:

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

### First Sync

```bash
# Clone GitHub organization with filters
gz synclone github --org kubernetes --language Go --min-stars 100

# Clone to specific directory
gz synclone github --org myorg --target ./my-repos

# Use pull strategy for active development
gz synclone github --org myorg --strategy pull
```

## Command Reference

### Platform-Specific Commands

#### `gz synclone github`

Clone and synchronize GitHub organizations:

```bash
gz synclone github --org <organization> [flags]
```

**Key Flags:**

- `--org`, `-o` - Organization name (required)
- `--target`, `-t` - Target directory (default: current directory)
- `--strategy` - Clone/update strategy: reset, pull, fetch, rebase, clone, skip
- `--include-archived` - Include archived repositories
- `--include-forks` - Include forked repositories
- `--include-private` - Include private repositories (default: true)
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

# Clone to specific directory with custom strategy
gz synclone github --org myorg --target ./repos --strategy pull

# Shallow clone for space efficiency
gz synclone github --org large-org --shallow --depth 1

# GitHub Enterprise
gz synclone github --org myorg --base-url https://github.company.com
```

#### `gz synclone gitlab`

Clone and synchronize GitLab groups:

```bash
gz synclone gitlab --group <group> [flags]
```

**Key Flags:**

- `--group`, `-g` - Group ID or path (required)
- `--target`, `-t` - Target directory
- `--strategy` - Clone/update strategy
- `--include-subgroups` - Include subgroups recursively
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

# Filter by visibility
gz synclone gitlab --group mygroup --visibility private
```

#### `gz synclone gitea`

Clone and synchronize Gitea organizations:

```bash
gz synclone gitea --org <organization> --base-url <url> [flags]
```

**Key Flags:**

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

#### `gz synclone gogs`

Clone and synchronize Gogs organizations:

```bash
gz synclone gogs --org <organization> --base-url <url> [flags]
```

**Key Flags:**

- `--org`, `-o` - Organization name (required)
- `--base-url` - Gogs instance URL (required)
- `--target`, `-t` - Target directory
- `--strategy` - Clone/update strategy

### Global Commands

#### Configuration File Usage

```bash
# Use configuration file for complex setups
gz synclone --config synclone.yaml

# Validate configuration
gz synclone validate --config synclone.yaml

# Test connection and authentication
gz synclone test --config synclone.yaml

# Dry run to preview operations
gz synclone --config synclone.yaml --dry-run
```

#### Global Flags

- `--config` - Configuration file path
- `--dry-run` - Show what would be done without executing
- `--force-refresh` - Ignore cache and fetch fresh data
- `--quiet`, `-q` - Suppress non-error output
- `--verbose`, `-v` - Enable verbose output
- `--debug` - Enable debug logging

## Configuration System

### Configuration File Structure

Create a `synclone.yaml` file for complex setups:

```yaml
# Global settings
version: "1.0"
target: "./repos"          # Base directory for all clones
strategy: "reset"          # Default strategy: reset, pull, fetch, rebase
shallow: false             # Use shallow clones
depth: 1                  # Depth for shallow clones (when shallow: true)
parallel: 5               # Number of concurrent operations

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
        min_stars: 10
        updated_after: "2024-01-01"

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
      branch: "main"        # Default branch to check out
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
        name_pattern: "^service-.*"      # Regex pattern
        exclude_pattern: ".*-deprecated$"
        min_stars: 10
        min_size: 100        # KB
        max_size: 1000000    # KB
        updated_after: "2024-01-01"

      # Repository-specific overrides
      repositories:
        - name: "critical-service"
          strategy: "pull"    # Never reset this repo
          branch: "production"
        - name: "experimental-feature"
          skip: true         # Don't clone this repo

      # Hooks (future feature)
      hooks:
        pre_clone: "scripts/pre-clone.sh"
        post_clone: "scripts/post-clone.sh"
        on_error: "scripts/on-error.sh"
```

### Configuration Hierarchy

Configuration values are resolved in this priority order:

1. **Environment variables** (highest priority)
1. **Command-line flags**
1. **Configuration file**
1. **Default values** (lowest priority)

### Authentication Configuration

#### Environment Variables

```bash
export GITHUB_TOKEN="ghp_..."
export GITLAB_TOKEN="glpat-..."
export GITEA_TOKEN="..."
export GOGS_TOKEN="..."
```

#### Configuration File

```yaml
github:
  organizations:
    - name: "myorg"
      token: "${GITHUB_TOKEN}"      # Environment variable reference
      # or
      token: "ghp_direct_token"     # Direct token (not recommended)
```

## Clone Strategies

Synclone supports multiple strategies for handling existing repositories:

### Strategy Comparison

| Strategy | Behavior                             | Use Case           | Risk Level                            |
| -------- | ------------------------------------ | ------------------ | ------------------------------------- |
| `reset`  | Hard reset to match remote (default) | CI/CD, mirrors     | ⚠️ **High** - Discards local changes  |
| `pull`   | Merge remote changes                 | Active development | 🟢 **Low** - Preserves local work     |
| `fetch`  | Update refs only                     | Inspection         | 🟢 **Low** - No working tree changes  |
| `rebase` | Rebase local changes on remote       | Clean history      | 🟡 **Medium** - May create conflicts  |
| `clone`  | Fresh clone (removes existing)       | Clean start        | 🔴 **Very High** - Deletes everything |
| `skip`   | Skip existing repositories           | Initial clone only | 🟢 **Low** - No changes to existing   |

### Strategy Details

#### reset (default)

Hard reset to match remote state, discarding all local changes:

```bash
gz synclone github --org myorg --strategy reset
```

- **Best for**: Read-only mirrors, CI/CD environments, backup systems
- **Warning**: All local changes will be lost

#### pull

Merge remote changes with local changes:

```bash
gz synclone github --org myorg --strategy pull
```

- **Best for**: Active development with local modifications
- **Note**: May create merge commits

#### fetch

Only update remote references without changing working tree:

```bash
gz synclone github --org myorg --strategy fetch
```

- **Best for**: Inspecting changes before merging, backup systems
- **Note**: Safest option, never modifies working tree

#### rebase

Rebase local changes on top of remote changes:

```bash
gz synclone github --org myorg --strategy rebase
```

- **Best for**: Maintaining linear history with local commits
- **Warning**: May fail if conflicts arise

#### clone

Always perform fresh clone (removes existing directory):

```bash
gz synclone github --org myorg --strategy clone
```

- **Best for**: Clean environments, resolving corruption
- **Warning**: Completely removes existing repositories

#### skip

Skip existing repositories:

```bash
gz synclone github --org myorg --strategy skip
```

- **Best for**: Initial bulk clone without updates
- **Note**: Only clones missing repositories

## Platform Support

### GitHub

**Features:**

- Public and GitHub Enterprise Server support
- Organization and user repository cloning
- Advanced filtering by language, topics, stars, size
- Fork and archived repository inclusion options
- Private repository support with authentication

**Authentication:**

- Personal Access Token (PAT)
- GitHub App token (future)

**Enterprise Support:**

```bash
gz synclone github --org myorg --base-url https://github.enterprise.com
```

### GitLab

**Features:**

- GitLab.com and self-hosted GitLab support
- Group and subgroup cloning
- Visibility filtering (public, internal, private)
- Project filtering by various criteria

**Authentication:**

- Personal Access Token
- Project Access Token
- OAuth token

**Self-hosted GitLab:**

```bash
gz synclone gitlab --group mygroup --base-url https://gitlab.company.com
```

### Gitea

**Features:**

- Self-hosted Gitea instances
- Organization repository cloning
- Basic filtering capabilities
- Authentication support

**Example:**

```bash
gz synclone gitea --org myorg --base-url https://gitea.company.com
```

### Gogs

**Features:**

- Lightweight Git service support
- Organization repository cloning
- Basic functionality similar to Gitea

**Example:**

```bash
gz synclone gogs --org myorg --base-url https://gogs.company.com
```

## Filtering and Selection

### Language Filters

Filter repositories by programming language:

```bash
# Single language
gz synclone github --org myorg --language Go

# Multiple languages
gz synclone github --org myorg --language Go,Python,JavaScript
```

**Configuration:**

```yaml
filters:
  languages: ["Go", "Python", "JavaScript"]
```

### Topic Filters

Filter repositories by GitHub topics:

```bash
# Single topic
gz synclone github --org myorg --topic microservices

# Multiple topics (AND logic)
gz synclone github --org myorg --topic kubernetes,production
```

**Configuration:**

```yaml
filters:
  topics: ["microservice", "api", "cloud-native"]
```

### Pattern Matching

Use regex patterns for repository names:

```bash
# Name pattern (regex)
gz synclone github --org myorg --name-pattern "^api-.*"

# Exclude pattern
gz synclone github --org myorg --exclude-pattern ".*-deprecated$"
```

**Configuration:**

```yaml
filters:
  name_pattern: "^service-.*"
  exclude_pattern: ".*-deprecated$"
```

### Size and Activity Filters

Filter by repository characteristics:

```bash
# Minimum stars
gz synclone github --org myorg --min-stars 50

# Size range (in KB)
gz synclone github --org myorg --min-size 100 --max-size 50000

# Recently updated
gz synclone github --org myorg --updated-after 2024-01-01
```

**Configuration:**

```yaml
filters:
  min_stars: 10
  min_size: 100      # KB
  max_size: 1000000  # KB
  updated_after: "2024-01-01"
```

### Repository Type Filters

Control which types of repositories to include:

```bash
# Include archived repositories
gz synclone github --org myorg --include-archived

# Include forked repositories
gz synclone github --org myorg --include-forks

# Exclude private repositories
gz synclone github --org myorg --include-private=false
```

**Configuration:**

```yaml
include_archived: false
include_forks: false
include_private: true
include_templates: false
```

## Performance and Monitoring

### Progress Display

Synclone provides real-time progress monitoring with enhanced UX:

**Normal Mode (Clean Output)**:

```bash
$ gz synclone github --org Gizzahub
🔍 Fetching repository list from GitHub organization: Gizzahub
📋 Found 5 repositories in organization Gizzahub
📝 Generated gzh.yaml with 5 repositories
📦 Processing 5 repositories (5 remaining)
[░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 0.0% (0/5) • ✓ 0 • ✗ 0 • ⏳ 5 • 0s
[████████████░░░░░░░░░░░░░░░░░░] 40.0% (2/5) • ✓ 2 • ✗ 0 • ⏳ 0 • 2s
[██████████████████████████████] 100.0% (5/5) • ✓ 5 • ✗ 0 • ⏳ 0 • 3s
✅ Clone operation completed successfully
```

**Debug Mode (With Detailed Logs)**:

```bash
$ gz synclone github --org Gizzahub --debug
22:13:47 INFO  [component=gzh-cli org=Gizzahub] Starting GitHub synclone operation
22:13:47 INFO  [component=gzh-cli org=Gizzahub] Starting synclone workflow: fetching repository list from GitHub
🔍 Fetching repository list from GitHub organization: Gizzahub
📋 Found 5 repositories in organization Gizzahub
📝 Generated gzh.yaml with 5 repositories
22:13:47 INFO  [component=gzh-cli org=Gizzahub] Using resumable parallel cloning
📦 Processing 5 repositories (5 remaining)
[░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 0.0% (0/5) • ✓ 0 • ✗ 0 • ⏳ 5 • 0s
[████████████░░░░░░░░░░░░░░░░░░] 40.0% (2/5) • ✓ 2 • ✗ 0 • ⏳ 0 • 2s
[██████████████████████████████] 100.0% (5/5) • ✓ 5 • ✗ 0 • ⏳ 0 • 3s
✅ Clone operation completed successfully
22:13:50 INFO  [component=gzh-cli org=Gizzahub] Operation 'github-synclone-completed' completed in 2.920s (Memory: 2.68 MB) [org_name=Gizzahub strategy=reset parallel=2]
22:13:50 INFO  [component=gzh-cli org=Gizzahub] GitHub synclone operation completed successfully
```

**Key UX Improvements (2025-09)**:

- **Clean Output**: Logs only appear with `--debug` flag for better user experience
- **Accurate Progress**: Progress bar starts from 0/total instead of jumping to middle values
- **Real-time Updates**: Progress updates every 500ms with precise status indicators
- **Human-readable Performance**: Performance metrics in text format, not JSON

### Concurrency Control

Optimize performance with parallel operations:

```bash
# High performance (good network)
gz synclone github --org myorg --parallel 10

# Conservative (limited resources)
gz synclone github --org myorg --parallel 2

# Single-threaded (debugging)
gz synclone github --org myorg --parallel 1
```

### Shallow Clones

Save disk space and time with shallow clones:

```bash
# Save disk space and time
gz synclone github --org myorg --shallow --depth 1

# Recent history only
gz synclone github --org myorg --shallow --depth 10
```

**Configuration:**

```yaml
shallow: true
depth: 1
```

### Caching

Improve performance with metadata caching:

```bash
# Use cached repository metadata
gz synclone github --org myorg --cache-ttl 1h

# Force refresh (ignore cache)
gz synclone github --org myorg --force-refresh
```

### Logging and Monitoring

**Enhanced Logging System (2025-09)**:

```bash
# Clean output (default) - only console messages
gz synclone github --org myorg

# Show detailed progress with logs
gz synclone github --org myorg --verbose

# Show debug information with detailed logs
gz synclone github --org myorg --debug

# Log to file
gz synclone github --org myorg --log-file synclone.log

# Quiet mode - suppress all logs except errors
gz synclone github --org myorg --quiet
```

**Logging Modes**:

- **Default Mode**: Clean console output with progress indicators only

  - Shows: 🔍, 📋, ✅ progress messages
  - Hides: Timestamp logs, debug information, performance metrics

- **Verbose Mode**: Adds informational logs

  - Shows: All default output + INFO level logs
  - Use: When you need more context about operations

- **Debug Mode**: Complete logging with technical details

  - Shows: All output + DEBUG logs + performance metrics
  - Use: For troubleshooting and development

- **Quiet Mode**: Error-only output

  - Shows: Only errors and critical failures
  - Use: In automated scripts or CI/CD environments

**Performance Metrics**: Now displayed in human-readable format instead of JSON:

```
Operation 'github-synclone-completed' completed in 2.920s (Memory: 2.68 MB)
```

## Best Practices

### 1. Organization Structure

Organize repositories by platform and organization:

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

### 2. Configuration Management

Use configuration files for complex setups:

```yaml
# production-sync.yaml
version: "1.0"
target: "/data/repositories"
strategy: "fetch"  # Safe for production
parallel: 3        # Conservative

github:
  enabled: true
  organizations:
    - name: "production-org"
      filters:
        exclude_pattern: ".*-experimental$"
```

### 3. Backup Strategy

Set up automated backups:

```bash
#!/bin/bash
# daily-backup.sh

LOG_FILE="logs/backup-$(date +%Y%m%d).log"

gz synclone --config backup.yaml \
  --strategy fetch \
  --log-file "$LOG_FILE" \
  --quiet

if [ $? -eq 0 ]; then
    echo "Backup completed successfully" | tee -a "$LOG_FILE"
else
    echo "Backup failed" | tee -a "$LOG_FILE" >&2
    exit 1
fi
```

### 4. CI/CD Integration

Example GitHub Actions workflow:

```yaml
# .github/workflows/sync-repos.yml
name: Sync Repositories

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install gzh-cli
        run: |
          curl -L https://github.com/gizzahub/gzh-cli/releases/latest/download/gz-linux-amd64 -o gz
          chmod +x gz

      - name: Sync repositories
        env:
          GITHUB_TOKEN: ${{ secrets.SYNC_TOKEN }}
        run: |
          ./gz synclone --config synclone.yaml \
            --strategy reset \
            --continue-on-error \
            --log-file sync.log

      - name: Upload logs
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: sync-logs
          path: sync.log
```

### 5. Performance Optimization

For large organizations:

```bash
# Use shallow clones to save space and time
gz synclone github --org large-org --shallow --depth 1

# Limit parallelism for stability
gz synclone github --org large-org --parallel 3

# Use filters to reduce scope
gz synclone github --org large-org --language Go --min-stars 10

# Cache repository metadata
gz synclone github --org large-org --cache-ttl 1h
```

### 6. Error Handling Strategy

Implement robust error handling:

```bash
# Retry failed operations
gz synclone github --org myorg --retry 3 --retry-delay 10s

# Continue on errors (for bulk operations)
gz synclone github --org myorg --continue-on-error

# Skip problematic repositories
gz synclone github --org myorg --skip-on-error
```

## Migration Guide

### From bulk-clone Command

If you're migrating from the old `bulk-clone` command:

```bash
# Old command
gz bulk-clone --org myorg

# New command
gz synclone github --org myorg

# Convert old config (if available)
gz synclone migrate --old-config bulk-clone.yaml --output synclone.yaml
```

### Key Differences from bulk-clone

1. **Platform-specific subcommands**: Use `github`, `gitlab`, etc.
1. **More granular filtering**: Language, topic, size, activity filters
1. **Better progress reporting**: Real-time progress with detailed metrics
1. **Improved error handling**: Retry logic and continue-on-error options
1. **Multiple strategies**: Different approaches for existing repositories
1. **Configuration file format**: New YAML structure with more options

### Migration Steps

1. **Update commands**:

   ```bash
   # Old
   gz bulk-clone --org myorg --target ./repos

   # New
   gz synclone github --org myorg --target ./repos
   ```

1. **Convert configuration**:

   ```yaml
   # Old bulk-clone.yaml format
   organizations:
     - name: myorg
       target: ./repos

   # New synclone.yaml format
   github:
     enabled: true
     organizations:
       - name: myorg
         target: ./repos
   ```

1. **Update scripts and automation**:

   - Replace `bulk-clone` with `synclone github/gitlab/gitea`
   - Update configuration file references
   - Test new filtering and strategy options

## Troubleshooting

### Common Issues

#### 1. Authentication Errors

**Problem**: "Authentication failed" or "401 Unauthorized"

**Solutions:**

```bash
# Verify token validity
gz synclone github --org myorg --verify-auth

# Check token permissions
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Use different token
gz synclone github --org myorg --token $BACKUP_TOKEN
```

#### 2. Rate Limiting

**Problem**: "Rate limit exceeded" or "429 Too Many Requests"

**Solutions:**

```bash
# Reduce parallelism
gz synclone github --org myorg --parallel 1

# Add delay between requests
gz synclone github --org myorg --request-delay 1s

# Check rate limit status
curl -H "Authorization: token $GITHUB_TOKEN" \
     https://api.github.com/rate_limit
```

#### 3. Network Issues

**Problem**: "Connection timeout" or "Network unreachable"

**Solutions:**

```bash
# Increase timeout
gz synclone github --org myorg --timeout 300s

# Use proxy
gz synclone github --org myorg --proxy http://proxy:8080

# Retry with backoff
gz synclone github --org myorg --retry 5 --retry-delay 30s
```

#### 4. Disk Space Issues

**Problem**: "No space left on device"

**Solutions:**

```bash
# Use shallow clones
gz synclone github --org myorg --shallow --depth 1

# Filter by size
gz synclone github --org myorg --max-size 50000

# Clean up existing clones
find ./repos -type d -name .git -exec du -sh {} \; | sort -h
```

#### 5. Configuration Errors

**Problem**: "Configuration validation failed"

**Solutions:**

```bash
# Validate configuration
gz synclone validate --config synclone.yaml

# Check for common issues
# - Missing required fields
# - Invalid URL formats
# - Circular dependencies

# Use minimal configuration to test
gz synclone github --org myorg --target ./test
```

### Debug Mode

Enable detailed debugging:

```bash
# Debug output
gz synclone github --org myorg --debug

# Verbose logging
gz synclone github --org myorg --verbose --log-file debug.log

# Dry run for testing
gz synclone github --org myorg --dry-run
```

### Error Recovery

Handle partial failures:

```bash
# Continue on errors
gz synclone github --org myorg --continue-on-error

# Skip failed repositories
gz synclone github --org myorg --skip-on-error

# Retry specific repositories
gz synclone github --org myorg --only failed-repo1,failed-repo2
```

### Performance Issues

Optimize for your environment:

```bash
# Reduce memory usage
gz synclone github --org myorg --parallel 2 --shallow

# Network optimization
gz synclone github --org myorg --timeout 60s --retry 3

# Disk I/O optimization
gz synclone github --org myorg --target /fast-storage/repos
```

## Examples

### Personal Repository Management

```bash
# Clone your personal repositories
gz synclone github --org yourusername --target ~/repos/personal

# Clone work repositories with SSH
gz synclone github --org company --target ~/repos/work --strategy pull
```

### Multi-Platform Enterprise Setup

```yaml
# enterprise-sync.yaml
version: "1.0"
target: "/data/repositories"
strategy: "fetch"
parallel: 5

github:
  enabled: true
  base_url: "https://github.enterprise.com"
  organizations:
    - name: "platform-team"
      target: "./github/platform"
      filters:
        languages: ["Go", "Python"]
        min_stars: 5

gitlab:
  enabled: true
  base_url: "https://gitlab.company.com"
  groups:
    - id: "infrastructure"
      target: "./gitlab/infra"
      include_subgroups: true

gitea:
  enabled: true
  base_url: "https://gitea.company.com"
  organizations:
    - name: "legacy-systems"
      target: "./gitea/legacy"
```

### Backup and Mirror Setup

```bash
#!/bin/bash
# mirror-setup.sh

# Create mirror directory structure
mkdir -p /backup/git-mirrors/{github,gitlab,gitea}

# Mirror critical organizations
gz synclone github --org kubernetes \
  --target /backup/git-mirrors/github/kubernetes \
  --strategy reset \
  --shallow --depth 1 \
  --parallel 5

gz synclone gitlab --group gitlab-org \
  --target /backup/git-mirrors/gitlab/gitlab-org \
  --strategy fetch \
  --include-subgroups

# Log results
echo "Mirror update completed at $(date)" >> /backup/logs/mirror.log
```

## Related Documentation

- [Configuration Guide](../30-configuration/configuration-guide.md)
- [Git Unified Command](../40-api-reference/git.md)
- [Repository Management](21-repository-management.md)
- [Performance Profiling](../40-api-reference/profile.md)

## Support

For additional help:

1. Run `gz synclone --help` for command options
1. Check the [examples directory](../../examples/synclone/) for more configurations
1. Open an issue on [GitHub](https://github.com/gizzahub/gzh-cli/issues)
1. Review the [configuration schema](../30-configuration/schemas/synclone-schema.yaml)
