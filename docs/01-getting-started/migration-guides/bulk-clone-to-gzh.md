# Migration Guide: bulk-clone.yaml → gzh.yaml

This guide helps you migrate from the legacy `bulk-clone.yaml` configuration format to the new unified `gzh.yaml` configuration system.

## Table of Contents

- [Overview](#overview)
- [Key Differences](#key-differences)
- [Migration Steps](#migration-steps)
- [Configuration Mapping](#configuration-mapping)
- [Examples](#examples)
- [Automated Migration Tool](#automated-migration-tool)
- [Troubleshooting](#troubleshooting)

## Overview

The new `gzh.yaml` configuration format provides:

- **Unified multi-provider support**: GitHub, GitLab, Gitea, Gogs in one file
- **Improved organization**: Cleaner structure with provider-specific sections
- **Enhanced filtering**: Better regex support and exclude patterns
- **Environment variables**: Built-in support for secure token management
- **Advanced features**: Directory flattening, clone strategies, validation

## Key Differences

### Schema Version

```yaml
# OLD: bulk-clone.yaml
version: "0.1"

# NEW: gzh.yaml
version: "1.0.0"
```

### Structure Changes

| bulk-clone.yaml            | gzh.yaml                      | Notes                               |
| -------------------------- | ----------------------------- | ----------------------------------- |
| `repo_roots[]`             | `providers.{provider}.orgs[]` | More organized provider structure   |
| `ignore_names[]`           | `exclude[]` per organization  | Per-organization exclusion patterns |
| `default.github.root_path` | `clone_dir` per organization  | More flexible path management       |
| `protocol`                 | Removed                       | Authentication handled via tokens   |
| `org_name`                 | `name`                        | Simpler naming                      |

### Authentication

```yaml
# OLD: Protocol-based authentication
protocol: "ssh" # or "https"

# NEW: Token-based authentication
token: "${GITHUB_TOKEN}"
```

## Migration Steps

### Step 1: Backup Your Current Configuration

```bash
# Backup your existing configuration
cp bulk-clone.yaml bulk-clone.yaml.backup
```

### Step 2: Create New gzh.yaml Structure

Start with the basic structure:

```yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs: []
  # Add other providers as needed
```

### Step 3: Migrate Organizations

Convert each `repo_roots` entry to the new format.

### Step 4: Update Environment Variables

Set up authentication tokens:

```bash
# Replace protocol-based auth with tokens
export GITHUB_TOKEN="your_github_token"
export GITLAB_TOKEN="your_gitlab_token"  # if using GitLab
```

### Step 5: Test and Validate

```bash
# Validate new configuration
gzh config validate

# Test with dry run
gzh bulk-clone --dry-run --use-gzh-config
```

## Configuration Mapping

### Basic Organization Migration

#### Before (bulk-clone.yaml)

```yaml
version: "0.1"

default:
  protocol: https
  github:
    root_path: "$HOME/github-repos"

repo_roots:
  - root_path: "$HOME/work/mycompany"
    provider: "github"
    protocol: "ssh"
    org_name: "mycompany"

  - root_path: "$HOME/opensource"
    provider: "github"
    protocol: "https"
    org_name: "kubernetes"

ignore_names:
  - "test-.*"
  - ".*-archive"
```

#### After (gzh.yaml)

```yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "mycompany"
        clone_dir: "$HOME/work/mycompany"
        exclude: ["test-.*", ".*-archive"]
        strategy: "reset"

      - name: "kubernetes"
        clone_dir: "$HOME/opensource"
        exclude: ["test-.*", ".*-archive"]
        strategy: "reset"
```

### Advanced Migration with Multiple Providers

#### Before (bulk-clone.yaml)

```yaml
version: "0.1"

default:
  protocol: https
  github:
    root_path: "$HOME/github"
  gitlab:
    root_path: "$HOME/gitlab"
    url: "https://gitlab.com"
    recursive: false

repo_roots:
  - root_path: "$HOME/work/company"
    provider: "github"
    protocol: "ssh"
    org_name: "mycompany"

  - root_path: "$HOME/personal"
    provider: "github"
    protocol: "https"
    org_name: "myusername"

ignore_names:
  - "^test-.*"
  - ".*-deprecated$"
```

#### After (gzh.yaml)

```yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "mycompany"
        visibility: "private"
        clone_dir: "$HOME/work/company"
        exclude: ["^test-.*", ".*-deprecated$"]
        strategy: "reset"

      - name: "myusername"
        visibility: "public"
        clone_dir: "$HOME/personal"
        exclude: ["^test-.*", ".*-deprecated$"]
        strategy: "pull"

  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      # Add GitLab groups as needed
      - name: "my-gitlab-group"
        visibility: "all"
        recursive: false
        clone_dir: "$HOME/gitlab"
        exclude: ["^test-.*", ".*-deprecated$"]
```

## Examples

### Example 1: Simple Personal Setup

#### Old Configuration

```yaml
# bulk-clone.yaml
version: "0.1"

repo_roots:
  - root_path: "$HOME/github"
    provider: "github"
    protocol: "https"
    org_name: "myusername"

ignore_names:
  - "test-*"
```

#### New Configuration

```yaml
# gzh.yaml
version: "1.0.0"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "myusername"
        clone_dir: "$HOME/github"
        exclude: ["test-*"]
```

### Example 2: Multi-Organization Development

#### Old Configuration

```yaml
# bulk-clone.yaml
version: "0.1"

default:
  protocol: ssh

repo_roots:
  - root_path: "$HOME/work/frontend"
    provider: "github"
    org_name: "frontend-team"

  - root_path: "$HOME/work/backend"
    provider: "github"
    org_name: "backend-team"

  - root_path: "$HOME/opensource"
    provider: "github"
    protocol: "https"
    org_name: "kubernetes"

ignore_names:
  - ".*-archive"
  - "temp-.*"
  - "test-.*"
```

#### New Configuration

```yaml
# gzh.yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "frontend-team"
        visibility: "private"
        clone_dir: "$HOME/work/frontend"
        match: "^(web|app|ui)-.*"
        exclude: [".*-archive", "temp-.*", "test-.*"]
        strategy: "pull"
        flatten: true

      - name: "backend-team"
        visibility: "private"
        clone_dir: "$HOME/work/backend"
        match: "^(api|service|worker)-.*"
        exclude: [".*-archive", "temp-.*", "test-.*"]
        strategy: "pull"
        flatten: true

      - name: "kubernetes"
        visibility: "public"
        clone_dir: "$HOME/opensource"
        exclude: [".*-archive", "temp-.*", "test-.*"]
        strategy: "fetch"
```

### Example 3: Enterprise with Multiple Providers

#### Old Configuration

```yaml
# bulk-clone.yaml
version: "0.1"

default:
  protocol: ssh
  github:
    root_path: "$HOME/work/github"
  gitlab:
    root_path: "$HOME/work/gitlab"
    url: "https://gitlab.company.com"

repo_roots:
  - root_path: "$HOME/work/platform"
    provider: "github"
    org_name: "company-platform"

  - root_path: "$HOME/work/tools"
    provider: "github"
    org_name: "company-tools"

ignore_names:
  - ".*-deprecated"
  - ".*-archive"
  - "test-.*"
  - "temp-.*"
```

#### New Configuration

```yaml
# gzh.yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_ENTERPRISE_TOKEN}"
    orgs:
      - name: "company-platform"
        visibility: "private"
        clone_dir: "$HOME/work/platform"
        match: "^(core|api|service)-.*"
        exclude: [".*-deprecated", ".*-archive", "test-.*", "temp-.*"]
        strategy: "reset"
        flatten: false

      - name: "company-tools"
        visibility: "private"
        clone_dir: "$HOME/work/tools"
        match: "^(tool|util|cli)-.*"
        exclude: [".*-deprecated", ".*-archive", "test-.*", "temp-.*"]
        strategy: "pull"
        flatten: true

  gitlab:
    token: "${GITLAB_ENTERPRISE_TOKEN}"
    groups:
      - name: "infrastructure"
        visibility: "private"
        recursive: true
        clone_dir: "$HOME/work/gitlab/infra"
        exclude: [".*-deprecated", ".*-archive", "test-.*", "temp-.*"]
        strategy: "reset"
```

## Automated Migration Tool

We provide a command-line tool to help automate the migration process:

```bash
# Convert existing bulk-clone.yaml to gzh.yaml
gzh config migrate --input bulk-clone.yaml --output gzh.yaml

# Migrate with validation
gzh config migrate --input bulk-clone.yaml --output gzh.yaml --validate

# Preview migration without writing file
gzh config migrate --input bulk-clone.yaml --dry-run
```

### Migration Tool Features

- **Automatic conversion**: Converts structure and field mappings
- **Token guidance**: Provides instructions for setting up authentication
- **Validation**: Checks the converted configuration for correctness
- **Backup creation**: Automatically creates backups of original files
- **Interactive mode**: Asks for user input when automatic conversion isn't possible

## Post-Migration Checklist

### 1. Environment Setup

```bash
# Set required environment variables
export GITHUB_TOKEN="your_github_personal_access_token"
export GITLAB_TOKEN="your_gitlab_access_token"  # if using GitLab
export GITEA_TOKEN="your_gitea_token"           # if using Gitea
```

### 2. Validation

```bash
# Validate the new configuration
gzh config validate

# Check for syntax errors
gzh config validate --strict
```

### 3. Test Run

```bash
# Perform a dry run to see what would be cloned
gzh bulk-clone --dry-run --use-gzh-config

# Test with verbose output
gzh bulk-clone --dry-run --verbose --use-gzh-config
```

### 4. Update Command Usage

```bash
# OLD: Using bulk-clone.yaml
gzh bulk-clone github --use-config -o myorg

# NEW: Using gzh.yaml
gzh bulk-clone --use-gzh-config
```

## Troubleshooting

### Common Migration Issues

#### 1. Authentication Errors

```
Error: missing required field: token
```

**Solution**: Set up environment variables for authentication:

```bash
export GITHUB_TOKEN="ghp_your_token_here"
```

#### 2. Path Resolution Issues

```
Error: failed to create directory
```

**Solution**: Check if environment variables are properly resolved:

```yaml
# Ensure environment variables are accessible
clone_dir: "${HOME}/repos" # Use ${HOME} instead of ~
```

#### 3. Regex Pattern Issues

```
Error: invalid regex pattern
```

**Solution**: Update regex patterns to use Go regex syntax:

```yaml
# OLD: Shell glob patterns
exclude: ["test-*"]

# NEW: Go regex patterns
exclude: ["^test-.*", ".*-test$"]
```

#### 4. Provider Configuration

```
Error: unsupported provider
```

**Solution**: Update provider structure:

```yaml
# OLD: Mixed in repo_roots
repo_roots:
  - provider: "github"

# NEW: Organized by provider
providers:
  github:
    orgs: [...]
```

### Migration Validation

```bash
# Compare old vs new behavior
gzh bulk-clone github --config bulk-clone.yaml --dry-run -o myorg
gzh bulk-clone --config gzh.yaml --dry-run --use-gzh-config

# Validate that both produce similar results
```

### Performance Considerations

The new gzh.yaml format provides better performance through:

- **Parallel processing**: Better support for concurrent operations
- **Efficient filtering**: Regex compilation happens once
- **Reduced API calls**: Smarter repository discovery

## Additional Resources

- [gzh.yaml Usage Guide](./gzh-yaml-usage-guide.md) - Comprehensive usage documentation
- [gzh.yaml Quick Reference](./gzh-yaml-quick-reference.md) - Quick reference for common patterns
- [Sample Configurations](../samples/) - Example configurations for different use cases
- [Schema Documentation](./gzh-schema.yaml) - Complete schema reference

## Getting Help

If you encounter issues during migration:

1. **Check the logs**: Run with `--verbose` for detailed output
2. **Validate configuration**: Use `gzh config validate`
3. **Use dry run**: Test with `--dry-run` before actual execution
4. **Compare outputs**: Ensure both old and new configs produce similar results
5. **Ask for help**: Open an issue with your configuration and error details

The migration tool and this guide should handle most common scenarios, but complex configurations may require manual adjustment.
