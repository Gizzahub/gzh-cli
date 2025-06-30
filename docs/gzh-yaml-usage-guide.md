# gzh.yaml Usage Guide

This guide provides comprehensive documentation for using the `gzh.yaml` configuration system in gzh-manager-go.

## Table of Contents

- [Overview](#overview)
- [Configuration File Structure](#configuration-file-structure)
- [Basic Usage](#basic-usage)
- [Advanced Configuration](#advanced-configuration)
- [Environment Variables](#environment-variables)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

The `gzh.yaml` configuration system provides a unified way to manage bulk clone operations across multiple Git providers (GitHub, GitLab, Gitea, Gogs). It supports advanced features like repository filtering, directory structure customization, and environment variable substitution.

### Key Features

- **Multi-provider support**: GitHub, GitLab, Gitea, Gogs
- **Environment variable substitution**: Secure token management
- **Advanced filtering**: Regular expressions and exclude patterns
- **Directory structure control**: Flatten or nested organization
- **Validation**: Built-in schema validation with helpful error messages

## Configuration File Structure

```yaml
# gzh.yaml
version: "1.0.0"                    # Required: Schema version
default_provider: github            # Optional: Default provider

providers:                          # Provider configurations
  github:                          # GitHub configuration
    token: "${GITHUB_TOKEN}"       # Authentication token
    orgs:                          # Organizations to clone
      - name: "my-org"             # Organization name
        visibility: "all"          # public, private, all
        clone_dir: "./github"      # Target directory
        match: "^my-.*"            # Regex filter (optional)
        exclude: ["test-*"]        # Exclude patterns (optional)
        strategy: "reset"          # reset, pull, fetch
        flatten: false             # Directory structure
  
  gitlab:                          # GitLab configuration
    token: "${GITLAB_TOKEN}"
    groups:                        # Groups to clone
      - name: "my-group"
        recursive: true            # Include subgroups
        
  gitea:                          # Gitea configuration
    token: "${GITEA_TOKEN}"
    orgs:
      - name: "my-gitea-org"
      
  gogs:                           # Gogs configuration
    token: "${GOGS_TOKEN}"
    orgs:
      - name: "my-gogs-org"
```

## Basic Usage

### 1. Creating Your First Configuration

Create a minimal `gzh.yaml` file:

```yaml
version: "1.0.0"
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "your-org-name"
```

### 2. Using the Configuration

```bash
# Clone using gzh.yaml configuration
gzh bulk-clone --use-gzh-config

# Use specific config file
gzh bulk-clone --config-file /path/to/gzh.yaml
```

### 3. File Location Priority

The tool searches for configuration files in this order:

1. Command line flag: `--config-file path/to/gzh.yaml`
2. Current directory: `./gzh.yaml`
3. User config directory: `~/.config/gzh.yaml`
4. Environment variable: `$GZH_CONFIG_PATH`

## Advanced Configuration

### Repository Filtering

#### Visibility Filtering

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "my-org"
        visibility: "public"    # Only public repositories
      - name: "private-org"
        visibility: "private"   # Only private repositories
      - name: "all-repos"
        visibility: "all"       # All repositories (default)
```

#### Pattern Matching

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "my-org"
        match: "^gzh-.*"        # Only repositories starting with "gzh-"
        exclude:                # Exclude specific patterns
          - "gzh-temp-*"        # Exclude temporary repos
          - "gzh-archive"       # Exclude specific repo
          - "*-backup"          # Exclude backup repos
```

#### Regular Expression Examples

```yaml
# Common patterns
match: "^api-.*"           # APIs only
match: ".*-service$"       # Services only
match: "^(web|app)-.*"     # Web and app projects
match: "\\.(js|ts)$"       # JavaScript/TypeScript projects
```

### Directory Structure

#### Nested Structure (Default)

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "my-org"
        clone_dir: "./repos"
        flatten: false      # Creates: ./repos/my-org/repo-name
```

#### Flattened Structure

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "my-org"
        clone_dir: "./all-repos"
        flatten: true       # Creates: ./all-repos/repo-name
```

### Clone Strategies

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "production-org"
        strategy: "reset"       # Hard reset to origin (default)
      - name: "development-org"
        strategy: "pull"        # Git pull if repo exists
      - name: "mirror-org"
        strategy: "fetch"       # Git fetch only
```

### GitLab-Specific Features

```yaml
providers:
  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "parent-group"
        recursive: true         # Include all subgroups
        visibility: "private"
      - name: "public-group"
        recursive: false        # Top-level group only
        visibility: "public"
```

## Environment Variables

### Token Management

```bash
# Set environment variables
export GITHUB_TOKEN="ghp_your_token_here"
export GITLAB_TOKEN="glpat_your_token_here"
export GITEA_TOKEN="your_gitea_token_here"
export GOGS_TOKEN="your_gogs_token_here"
```

### Path Variables

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "my-org"
        clone_dir: "${HOME}/projects/github"    # Uses $HOME variable
      - name: "work-org"
        clone_dir: "${WORKSPACE}/repos"         # Custom environment variable
```

### Default Values

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN:default-token}"     # Uses default if not set
    orgs:
      - name: "my-org"
        clone_dir: "${CLONE_DIR:./repos}"      # Default directory
```

## Examples

### Example 1: Simple GitHub Organization

```yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "gizzahub"
        visibility: "public"
        clone_dir: "./github"
```

### Example 2: Multi-Provider Setup

```yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "work-org"
        visibility: "private"
        clone_dir: "${HOME}/work/github"
        match: "^project-.*"
        strategy: "pull"
      - name: "oss-org"
        visibility: "public"
        clone_dir: "${HOME}/oss/github"
        flatten: true

  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "infrastructure"
        recursive: true
        visibility: "private"
        clone_dir: "${HOME}/work/gitlab"
        exclude: ["archive-*", "*-backup"]

  gitea:
    token: "${GITEA_TOKEN}"
    orgs:
      - name: "personal"
        visibility: "all"
        clone_dir: "${HOME}/personal/gitea"
```

### Example 3: Development Environment

```yaml
version: "1.0.0"
default_provider: github

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "frontend-team"
        visibility: "all"
        clone_dir: "./frontend"
        match: "^(web|app|mobile)-.*"
        strategy: "pull"
        flatten: true
      
      - name: "backend-team"
        visibility: "all"
        clone_dir: "./backend"
        match: "^(api|service|worker)-.*"
        strategy: "pull"
        flatten: true
      
      - name: "devops-team"
        visibility: "private"
        clone_dir: "./infrastructure"
        match: "^(infra|deploy|ci)-.*"
        strategy: "reset"
```

### Example 4: Research and Archival

```yaml
version: "1.0.0"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "research-org"
        visibility: "public"
        clone_dir: "./research/github"
        exclude: ["workshop-*", "temp-*"]
        strategy: "fetch"  # Read-only mirroring
      
      - name: "archive-org"
        visibility: "all"
        clone_dir: "./archive"
        strategy: "fetch"
        flatten: true
```

## Best Practices

### 1. Token Security

```yaml
# ✅ Good: Use environment variables
token: "${GITHUB_TOKEN}"

# ❌ Bad: Hardcoded tokens
token: "ghp_hardcoded_token_here"
```

### 2. Directory Organization

```yaml
# ✅ Good: Organized structure
providers:
  github:
    orgs:
      - name: "work-org"
        clone_dir: "${HOME}/work/github"
      - name: "personal-org"
        clone_dir: "${HOME}/personal/github"

# ❌ Bad: Everything in one place
clone_dir: "./all-repos"
```

### 3. Filtering Strategy

```yaml
# ✅ Good: Specific filtering
match: "^api-.*"           # Clear intent
exclude: ["*-archive", "*-backup"]

# ❌ Bad: Overly broad or complex
match: ".*(api|service|worker|job|task|cron).*"
```

### 4. Version Control

```yaml
# Always specify version
version: "1.0.0"

# Consider environment-specific configs
# gzh.dev.yaml, gzh.prod.yaml, gzh.personal.yaml
```

### 5. Documentation

```yaml
# Add comments for complex configurations
providers:
  github:
    orgs:
      - name: "complex-org"
        # Only clone active microservices, exclude legacy systems
        match: "^service-.*"
        exclude: ["service-legacy-*", "service-deprecated-*"]
        # Use pull strategy for active development
        strategy: "pull"
```

## Troubleshooting

### Common Issues

#### 1. Authentication Errors

```
Error: configuration validation failed: missing required field: token
```

**Solution**: Ensure environment variables are set:
```bash
echo $GITHUB_TOKEN  # Should output your token
export GITHUB_TOKEN="your_token_here"
```

#### 2. Invalid Regular Expression

```
Error: configuration validation failed: invalid regex pattern
```

**Solution**: Test your regex pattern:
```bash
# Test regex online or with tools
echo "test-repo" | grep -E "^test-.*"
```

#### 3. Directory Permission Issues

```
Error: failed to create directory: permission denied
```

**Solution**: Check directory permissions or use accessible paths:
```yaml
clone_dir: "${HOME}/repos"  # User directory
# Instead of:
clone_dir: "/usr/local/repos"  # System directory
```

#### 4. Configuration File Not Found

```
Error: configuration file not found
```

**Solution**: Check file location and permissions:
```bash
# Check if file exists
ls -la gzh.yaml

# Check configuration search paths
gzh bulk-clone --help | grep -A5 "config"
```

### Validation Commands

```bash
# Validate configuration file
gzh config validate

# Dry run to see what would be cloned
gzh bulk-clone --dry-run --use-gzh-config

# Verbose output for debugging
gzh bulk-clone --verbose --use-gzh-config
```

### Common Patterns

#### Multiple Environments

```yaml
# gzh.dev.yaml
version: "1.0.0"
providers:
  github:
    token: "${GITHUB_DEV_TOKEN}"
    orgs:
      - name: "dev-org"
        strategy: "pull"  # Active development

# gzh.prod.yaml  
version: "1.0.0"
providers:
  github:
    token: "${GITHUB_PROD_TOKEN}"
    orgs:
      - name: "prod-org"
        strategy: "fetch"  # Read-only
```

#### Incremental Migration

```yaml
# Start with one provider
version: "1.0.0"
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "main-org"

# Add more providers over time
# gitlab: ...
# gitea: ...
```

For more advanced usage and examples, see the [examples directory](../samples/) and [schema documentation](./gzh-schema.yaml).