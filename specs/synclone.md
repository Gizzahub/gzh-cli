<!-- 🚫 AI_MODIFY_PROHIBITED -->

<!-- This file should not be modified by AI agents -->

# Repository Synchronization and Cloning Specification

## Overview

The `synclone` command (previously `bulk-clone`) provides comprehensive repository synchronization and cloning capabilities across multiple Git platforms. It enables bulk operations for cloning entire organizations, managing repository configurations, and maintaining synchronized development environments across GitHub, GitLab, Gitea, and Gogs platforms.

## Commands

### Core Commands

- `gz synclone` - Main command for synchronized repository cloning
- `gz synclone github` - Clone repositories from GitHub organizations
- `gz synclone gitlab` - Clone repositories from GitLab groups
- `gz synclone gitea` - Clone repositories from Gitea organizations
- `gz synclone validate` - Validate configuration files

### Configuration Commands

- `gz synclone config` - Configuration file management
- `gz synclone config generate` - Generate configuration files
- `gz synclone config validate` - Validate configuration syntax
- `gz synclone config convert` - Convert between formats

### State Management Commands

- `gz synclone state` - Manage operation state
- `gz synclone state list` - List tracked operations
- `gz synclone state show` - Show operation details
- `gz synclone state clean` - Clean up state files

## Command Details

### Repository Cloning (`gz synclone`)

**Purpose**: Clone repositories from multiple Git platforms with configurable strategies

**Features**:

- Multi-platform support (GitHub, GitLab, Gitea, Gogs)
- Organization and group-based cloning
- Configurable clone strategies (reset, pull, fetch)
- Authentication token management
- Parallel cloning with rate limiting
- Repository filtering and exclusions
- Custom clone destinations
- Orphan directory cleanup

**Options**:

```bash
--config custom.yaml    # Clone using specific config
--strategy pull         # Use pull strategy instead of reset
--parallel 5            # Set parallel clone limit
--resume                # Resume interrupted clone operation
--cleanup-orphans       # Remove directories not in organization
--use-gzh-config        # Use gzh.yaml configuration format
```

### Platform-Specific Cloning

#### GitHub (`gz synclone github`)

**Features**:

- Clone entire GitHub organizations
- Filter by visibility (public/private/all)
- Match repositories with regex patterns
- Support HTTPS and SSH protocols
- Token authentication for private repos
- Branch selection and shallow cloning
- Bare repository support

**Usage**:

```bash
gz synclone github -o Gizzahub                      # Clone organization
gz synclone github -o Gizzahub -t ~/workspace      # Custom directory
gz synclone github -o Gizzahub --match ".*-api"    # Pattern matching
gz synclone github -o Gizzahub --visibility private # Private repos only
gz synclone github -o Gizzahub --protocol ssh      # Use SSH
gz synclone github -o Gizzahub --branch develop    # Specific branch
gz synclone github -o Gizzahub --bare              # Bare repos
gz synclone github -o Gizzahub --depth 1           # Shallow clone
```

#### GitLab (`gz synclone gitlab`)

**Features**:

- Clone entire GitLab groups
- Support for subgroups
- All features from GitHub cloning
- Custom GitLab instance support

**Usage**:

```bash
gz synclone gitlab -g mygroup                       # Clone group
gz synclone gitlab -g mygroup -t ~/workspace       # Custom directory
gz synclone gitlab -g mygroup --match ".*-service" # Pattern matching
gz synclone gitlab -g mygroup --recursive          # Include subgroups
```

#### Gitea (`gz synclone gitea`)

**Features**:

- Clone entire Gitea organizations
- Custom Gitea instance support
- All standard cloning features

**Usage**:

```bash
gz synclone gitea -o myorg                         # Clone organization
gz synclone gitea -o myorg -t ~/workspace         # Custom directory
gz synclone gitea -o myorg --api-url https://...  # Custom instance
```

### Path Resolution Rules

For all platforms:

1. Without `-t` flag: Clones to organization/group name in current directory
1. With relative `-t` flag: Clones relative to current directory
1. With absolute `-t` flag: Clones to absolute path

Examples:

- `gz synclone github -o Gizzahub` → `./Gizzahub/`
- `gz synclone github -o Gizzahub -t workspace` → `./workspace/`
- `gz synclone github -o Gizzahub -t ~/work` → `~/work/`

### Configuration Management (`gz synclone config`)

**Purpose**: Advanced configuration file management tools

#### Generate (`gz synclone config generate`)

Create configuration files with various strategies:

```bash
# Initialize new configuration
gz synclone config generate init

# Generate from template
gz synclone config generate template --template enterprise

# Discover existing repositories
gz synclone config generate discover --path ~/repos

# Generate GitHub-specific config
gz synclone config generate github --org mycompany
```

#### Validate (`gz synclone config validate`)

Validate configuration syntax and structure:

```bash
# Basic validation
gz synclone config validate --config synclone.yaml

# Strict validation with schema checking
gz synclone config validate --strict --config synclone.yaml
```

#### Convert (`gz synclone config convert`)

Convert between configuration formats:

```bash
# YAML to JSON
gz synclone config convert --from synclone.yaml --to synclone.json

# Convert to gzh.yml format
gz synclone config convert --from synclone.yaml --format gzh
```

### State Management (`gz synclone state`)

**Purpose**: Track and manage clone operations

#### List (`gz synclone state list`)

View tracked operations:

```bash
# All operations
gz synclone state list

# Active operations only
gz synclone state list --active

# Failed operations
gz synclone state list --failed
```

#### Show (`gz synclone state show`)

Display operation details:

```bash
# Show by ID
gz synclone state show <state-id>

# Show last operation
gz synclone state show --last
```

#### Clean (`gz synclone state clean`)

Clean up state files:

```bash
# Clean old operations
gz synclone state clean --age 7d

# Clean failed operations
gz synclone state clean --failed

# Clean specific operation
gz synclone state clean --id <state-id>
```

## Configuration

### Configuration Hierarchy

1. Environment variable: `GZH_SYNCLONE_CONFIG`
1. Current directory: `./synclone.yaml` or `./synclone.yml`
1. User config: `~/.config/gzh-manager/synclone.yaml`
1. System config: `/etc/gzh-manager/synclone.yaml`

### Configuration Structure

```yaml
version: "1.0.0"
default_provider: github

# Sync mode for subsequent operations
sync_mode:
  cleanup_orphans: true
  conflict_resolution: "remote-overwrite"  # remote-overwrite, local-preserve, rebase-attempt, conflict-skip

# Global settings
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset      # reset, pull, fetch
  default_visibility: all      # all, public, private
  default_protocol: https      # http, https, ssh
  global_ignores:
    - "^test-.*"
    - ".*-archive$"
    - ".*-deprecated$"
  timeouts:
    http_timeout: 30s
    git_timeout: 5m
  concurrency:
    clone_workers: 10
    update_workers: 15

# Provider configurations
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "mycompany"
        clone_dir: "$HOME/work/mycompany"
        visibility: all
        strategy: reset
        protocol: ssh
        branch: main
        exclude:
          - ".*-archive$"
          - ".*-deprecated$"
        auth:
          token: "${GITHUB_COMPANY_TOKEN}"
          ssh_key: "~/.ssh/github_rsa"

    settings:
      rate_limit:
        requests_per_hour: 5000
        burst_limit: 50
        auto_detect: true
      retry:
        max_attempts: 3
        exponential_backoff: true

  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.com"
    groups:
      - name: "backend-team"
        clone_dir: "$HOME/work/gitlab-internal"
        recursive: true
        api_url: "https://gitlab.company.com"

  gitea:
    token: "${GITEA_TOKEN}"
    api_url: "https://gitea.com"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/gitea/myorg"
```

## Directory Management

### gzh.yml Validation

When executing in an existing directory:

1. **With gzh.yml**: Uses existing configuration
1. **Without gzh.yml**: Fails to prevent accidental overwrites
1. **New/Empty directory**: Creates new gzh.yml

### Sync Modes

#### cleanup_orphans

- `true`: Remove directories not in configuration
- `false`: Preserve existing directories

#### conflict_resolution

- **remote-overwrite**: Hard reset to remote state (default)
- **local-preserve**: Keep local changes
- **rebase-attempt**: Try rebase, leave conflicts for manual resolution
- **conflict-skip**: Skip repositories with conflicts

## State Management and Recovery

### Operation Tracking

Synclone tracks all operations for:

- Resume capability for interrupted operations
- Operation history and audit trails
- Cleanup of failed or partial clones
- Performance metrics and statistics

### State Storage

- Location: `~/.config/gzh-manager/synclone/state/`
- Individual state files per operation
- Automatic cleanup of old state files

### Resume Capability

When using `--resume`:

1. Checks for incomplete operations
1. Identifies repositories needing retry
1. Continues from last successful repository
1. Maintains same configuration and options

### State File Contents

- Operation ID and timestamp
- Configuration snapshot
- Repository list and status
- Success/failure metrics
- Error logs for failed repositories

## Examples

### Basic Operations

```bash
# Clone GitHub organization
gz synclone github -o myorg

# Clone with configuration
gz synclone --config synclone.yaml

# Resume interrupted operation
gz synclone --resume

# Validate before cloning
gz synclone validate --config synclone.yaml
```

### Advanced Configuration Management

```bash
# Generate config from existing repos
gz synclone config generate discover --path ~/projects

# Convert formats
gz synclone config convert --from synclone.yaml --to synclone.json

# Validate with token checks
gz synclone validate --check-tokens --config synclone.yaml
```

### State Management

```bash
# View operation history
gz synclone state list

# Check last operation
gz synclone state show --last

# Clean old operations
gz synclone state clean --age 30d
```

### Multi-Organization Setup

```bash
# First time setup
gz synclone config generate init
# Edit configuration to add organizations

# Clone all configured organizations
gz synclone --config synclone.yaml

# Update existing clones
gz synclone --config synclone.yaml --strategy pull

# Clean orphan directories
gz synclone --config synclone.yaml --cleanup-orphans
```

## Integration Points

- **Development Environment**: Coordinates with `dev-env` for environment-specific access
- **Network Management**: Integrates with `net-env` for proxy and VPN-aware cloning
- **SSH Configuration**: Works with SSH keys for authentication
- **Git Configuration**: Respects global and local git settings

## Security Considerations

- Secure token storage and handling
- SSH key management integration
- Repository authenticity verification
- Access control enforcement
- Complete audit logging
- VPN and proxy support
- Encrypted backup of metadata

## Performance Optimization

- Parallel processing with configurable limits
- Rate limiting with burst control
- Incremental updates for efficiency
- API response caching
- Bandwidth management
- Smart storage optimization

## Platform Support

- **GitHub**: Full support including GitHub Enterprise
- **GitLab**: Groups, subgroups, and self-hosted instances
- **Gitea**: Organizations and custom instances
- **Gogs**: Organizations and custom instances
- **Git**: Any Git repository with HTTPS/SSH access
