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

**Usage**:
```bash
--config custom.yaml # Clone using specific config
--strategy pull      # Use pull strategy instead of reset
--parallel 5         # Set parallel clone limit
--resume             # Resume interrupted clone operation
--cleanup-orphans    # Remove directories not in organization
--use-gzh-config     # Use gzh.yaml configuration format
```

### GitHub Cloning (`gz synclone github`)

**Purpose**: Clone repositories from GitHub organizations

**Features**:
- Clone entire GitHub organizations
- Filter repositories by visibility
- Match repositories with regex patterns
- Support for both HTTPS and SSH protocols
- Private repository support with token authentication
- Custom directory structures
- Branch selection support
- Shallow cloning capability
- Bare repository support

**Path Resolution Rules**:
1. Without `-t` flag: Clones to organization name directory in current path
   - `gz synclone github -o ScriptonBasestar` → `./ScriptonBasestar/`
2. With relative `-t` flag: Clones to specified directory relative to current path
   - `gz synclone github -o ScriptonBasestar -t scripton` → `./scripton/`
3. With absolute `-t` flag: Clones to specified absolute path
   - `gz synclone github -o ScriptonBasestar -t ~/work/scripton` → `~/work/scripton/`

**Usage**:
```bash
gz synclone github -o Gizzahub              # Clone Gizzahub organization
gz synclone github -o Gizzahub -t ~/workspace # Clone to specific directory
gz synclone github -o Gizzahub --match ".*-api" # Clone only matching repos
gz synclone github -o Gizzahub --visibility private # Clone only private repos
gz synclone github -o Gizzahub --protocol ssh # Use SSH protocol
gz synclone github -o Gizzahub --branch develop # Clone develop branch
gz synclone github -o Gizzahub --bare       # Clone as bare repositories
gz synclone github -o Gizzahub --depth 1    # Shallow clone with depth 1
```

### GitLab Cloning (`gz synclone gitlab`)

**Purpose**: Clone repositories from GitLab groups

**Features**:
- Clone entire GitLab groups
- Support for subgroups
- Filter repositories by visibility
- Match repositories with regex patterns
- Support for both HTTPS and SSH protocols
- Private repository support with token authentication
- Custom directory structures
- Branch selection support
- Shallow cloning capability
- Bare repository support

**Path Resolution Rules**:
1. Without `-t` flag: Clones to group name directory in current path
   - `gz synclone gitlab -g mygroup` → `./mygroup/`
2. With relative `-t` flag: Clones to specified directory relative to current path
   - `gz synclone gitlab -g mygroup -t projects` → `./projects/`
3. With absolute `-t` flag: Clones to specified absolute path
   - `gz synclone gitlab -g mygroup -t ~/work/projects` → `~/work/projects/`

**Usage**:
```bash
gz synclone gitlab -g mygroup               # Clone mygroup group
gz synclone gitlab -g mygroup -t ~/workspace # Clone to specific directory
gz synclone gitlab -g mygroup --match ".*-service" # Clone only matching repos
gz synclone gitlab -g mygroup --visibility public # Clone only public repos
gz synclone gitlab -g mygroup --protocol ssh # Use SSH protocol
gz synclone gitlab -g mygroup --branch develop # Clone develop branch
gz synclone gitlab -g mygroup --bare       # Clone as bare repositories
gz synclone gitlab -g mygroup --depth 1    # Shallow clone with depth 1
```

### Gitea Cloning (`gz synclone gitea`)

**Purpose**: Clone repositories from Gitea organizations

**Features**:
- Clone entire Gitea organizations
- Filter repositories by visibility
- Match repositories with regex patterns
- Support for both HTTPS and SSH protocols
- Private repository support with token authentication
- Custom directory structures
- Branch selection support
- Shallow cloning capability
- Bare repository support

**Path Resolution Rules**:
1. Without `-t` flag: Clones to organization name directory in current path
   - `gz synclone gitea -o myorg` → `./myorg/`
2. With relative `-t` flag: Clones to specified directory relative to current path
   - `gz synclone gitea -o myorg -t repositories` → `./repositories/`
3. With absolute `-t` flag: Clones to specified absolute path
   - `gz synclone gitea -o myorg -t ~/work/repositories` → `~/work/repositories/`

**Usage**:
```bash
gz synclone gitea -o myorg                  # Clone myorg organization
gz synclone gitea -o myorg -t ~/workspace   # Clone to specific directory
gz synclone gitea -o myorg --match ".*-app" # Clone only matching repos
gz synclone gitea -o myorg --visibility all # Clone all repos
gz synclone gitea -o myorg --protocol ssh   # Use SSH protocol
gz synclone gitea -o myorg --branch develop # Clone develop branch
gz synclone gitea -o myorg --bare          # Clone as bare repositories
gz synclone gitea -o myorg --depth 1       # Shallow clone with depth 1
```

### Configuration Validation (`gz synclone validate`)

**Purpose**: Validate synclone configuration files against schema

**Features**:
- JSON Schema validation
- Configuration syntax checking
- Token validation
- Platform connectivity testing
- Repository access verification

**Usage**:
```bash
gz synclone validate                   # Validate default config
gz synclone validate --config custom.yaml # Validate specific config
gz synclone validate --strict         # Strict validation mode
gz synclone validate --check-tokens   # Validate authentication tokens
```


## Directory Restrictions and Validation

### gzh.yml Validation

When executing synclone in an existing directory, the command will check for the presence of a `gzh.yml` configuration file:

1. **Existing Directory with gzh.yml**: Synclone will use the configuration from the existing `gzh.yml` file
2. **Existing Directory without gzh.yml**: Command will fail with an error to prevent accidental overwrites
3. **New/Empty Directory**: Synclone will proceed and create a new `gzh.yml` configuration file

### Directory Requirements

- **First-time clone**: Target directory must be either:
  - Non-existent (will be created)
  - Empty directory
  - Directory with valid `gzh.yml` configuration
  
- **Subsequent synclone**: Target directory must contain valid `gzh.yml` file

## Configuration

### Global Configuration

Repository synchronization configurations are stored in:
- `~/.config/gzh-manager/synclone.yaml` - User-specific settings
- `/etc/gzh-manager/synclone.yaml` - System-wide settings
- Environment variable: `GZH_SYNCLONE_CONFIG`
- Current directory: `./synclone.yaml` or `./synclone.yml`

### Configuration Hierarchy

1. Environment variable: `GZH_SYNCLONE_CONFIG`
2. Current directory: `./synclone.yaml` or `./synclone.yml`
3. User config: `~/.config/gzh-manager/synclone.yaml`
4. System config: `/etc/gzh-manager/synclone.yaml`

### Configuration Structure

```yaml
# gzh-manager synclone configuration
# Version: 1.0.0
# Documentation: https://github.com/gizzahub/gzh-manager-go/docs/synclone.md

version: "1.0.0"
default_provider: github

# Sync mode configuration for subsequent synclone operations
sync_mode:
  # Remove directories not defined in gzh.yml
  cleanup_orphans: true
  
  # Conflict resolution strategy
  # Options:
  # - "remote-overwrite": Overwrite local changes with remote version (hard reset)
  # - "local-preserve": Preserve local changes and ignore remote updates
  # - "rebase-attempt": Attempt rebase and leave conflicts unresolved
  # - "conflict-skip": Skip repositories with conflicts, leave unchanged
  conflict_resolution: "remote-overwrite"

# Global settings that apply to all providers
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset      # Options: reset, pull, fetch
  default_visibility: all      # Options: all, public, private
  default_protocol: https      # Options: http, https, ssh
  global_ignores:
    - "^test-.*"        # Repos starting with 'test-'
    - ".*-archive$"     # Repos ending with '-archive'
    - "^temp.*"         # Repos starting with 'temp'
    - ".*-deprecated$"  # Deprecated repositories
    - "^\\."            # Hidden repositories (starting with .)
  timeouts:
    http_timeout: 30s
    git_timeout: 5m
    rate_limit_timeout: 1h
  concurrency:
    clone_workers: 10
    update_workers: 15
    api_workers: 5

# Provider configurations for Git hosting services
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "mycompany"
        clone_dir: "$HOME/work/mycompany"
        visibility: all
        strategy: reset
        protocol: ssh
        branch: main              # Default branch to clone
        bare: false              # Clone as bare repository
        depth: 0                 # Clone depth (0 = full history)
        include: ""              # Regex pattern to include repos
        exclude:
          - ".*-archive$"
          - ".*-deprecated$"
        flatten: false           # Flatten directory structure
        auth:
          token: "${GITHUB_COMPANY_TOKEN}"  # Org-specific token
          ssh_key: "~/.ssh/github_rsa"      # Org-specific SSH key
      
      - name: "kubernetes"
        clone_dir: "$HOME/opensource/kubernetes"
        visibility: public
        strategy: pull
        protocol: https
        branch: main
    
    settings:
      rate_limit:
        requests_per_hour: 5000
        burst_limit: 50
        auto_detect: true
      retry:
        max_attempts: 3
        base_delay: 1s
        max_delay: 30s
        exponential_backoff: true
      auth:
        token_env_var: "GITHUB_TOKEN"
        use_ssh: true
        ssh_key_path: "$HOME/.ssh/id_rsa"

  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.com"
    groups:
      - name: "backend-team"
        clone_dir: "$HOME/work/gitlab-internal"
        visibility: private
        strategy: reset
        protocol: https
        branch: develop
        bare: false
        depth: 0
        recursive: true          # Include subgroups
        api_url: "https://gitlab.company.com"  # Custom GitLab instance
        auth:
          token: "${COMPANY_GITLAB_TOKEN}"
          ssh_key: "~/.ssh/company_rsa"
    
    settings:
      rate_limit:
        requests_per_hour: 2000
        burst_limit: 30
        auto_detect: true
      retry:
        max_attempts: 3
        base_delay: 2s
        max_delay: 60s
        exponential_backoff: true

  gitea:
    token: "${GITEA_TOKEN}"
    api_url: "https://gitea.com"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/gitea/myorg"
        visibility: all
        strategy: reset
        protocol: ssh
        branch: main
        bare: false
        depth: 0
        api_url: "https://gitea.company.com"  # Custom Gitea instance
        auth:
          token: "${COMPANY_GITEA_TOKEN}"
          ssh_key: "~/.ssh/gitea_rsa"
    
    settings:
      rate_limit:
        requests_per_hour: 1000
        burst_limit: 20
        auto_detect: true
      retry:
        max_attempts: 3
        base_delay: 1s
        max_delay: 30s
        exponential_backoff: true

# Auto-sync configuration
auto_sync:
  enabled: false
  interval: "6h"
  on_network_change: true

# Progress display configuration
progress:
  mode: "bar"         # Options: bar, dots, spinner, quiet
  update_interval: "1s"

# Logging configuration
logging:
  level: "info"       # Options: debug, info, warn, error
  file: "~/.config/gzh-manager/synclone.log"
  format: "json"      # Options: json, text
  max_size: "100MB"
  max_age: "30d"
  compress: true
```

### Environment Variables

- `GZH_SYNCLONE_CONFIG` - Path to configuration file
- `GITHUB_TOKEN` - GitHub authentication token
- `GITLAB_TOKEN` - GitLab authentication token
- `GITEA_TOKEN` - Gitea authentication token
- `GZH_SYNCLONE_STRATEGY` - Override default clone strategy
- `GZH_SYNCLONE_PARALLEL` - Override parallel clone limit
- `GZH_SYNCLONE_PROGRESS_MODE` - Override progress display mode

## Examples

### Basic Organization Cloning

```bash
# Clone GitHub organization
gz synclone github -o myorg

# Clone GitLab group
gz synclone gitlab -g mygroup

# Clone Gitea organization
gz synclone gitea -o myorg

# Clone with custom target directory
gz synclone github -o myorg -t ~/workspace/myorg
```

### Advanced Filtering

```bash
# Clone only repositories matching pattern
gz synclone github -o myorg --match ".*-service"

# Clone only private repositories
gz synclone github -o myorg --visibility private

# Clone with SSH protocol
gz synclone github -o myorg --protocol ssh

# Clone with specific strategy
gz synclone github -o myorg --strategy pull
```

### Configuration-Based Cloning

```bash
# Clone using configuration file
gz synclone --config synclone.yaml

# Clone using gzh.yaml format
gz synclone --use-gzh-config

# Validate configuration before cloning
gz synclone validate --config synclone.yaml

# Clone with parallel processing
gz synclone --parallel 5 --config synclone.yaml
```

### Resumable Operations

```bash
# Start cloning with resume capability
gz synclone --resume --config synclone.yaml

# Resume interrupted operation
gz synclone --resume
```

### Repository Management

```bash
# Clone and cleanup orphan directories
gz synclone --cleanup-orphans --config synclone.yaml

# Validate configuration
gz synclone validate --config synclone.yaml

# Validate with token checks
gz synclone validate --check-tokens --config synclone.yaml
```

## Sync Modes and Conflict Resolution

### Sync Mode Configuration

When performing subsequent synclone operations on an existing directory with `gzh.yml`, the following sync modes control the behavior:

#### cleanup_orphans
- **true**: Removes any directories that exist locally but are not defined in the `gzh.yml` configuration
- **false**: Preserves all existing directories, only updates defined repositories

#### conflict_resolution
Controls how to handle conflicts when local changes exist:

1. **remote-overwrite** (default)
   - Description: Overwrite local changes with remote version (hard reset to remote state)
   - Use case: When you want to ensure local matches remote exactly
   - Git operation: `git reset --hard origin/branch`

2. **local-preserve**
   - Description: Preserve local changes and ignore remote updates
   - Use case: When local work takes priority over remote changes
   - Git operation: No pull/fetch performed

3. **rebase-attempt**
   - Description: Attempt rebase and leave conflicts unresolved for manual handling
   - Use case: When you want to integrate changes but handle conflicts manually
   - Git operation: `git pull --rebase`, leaves conflicts for manual resolution

4. **conflict-skip**
   - Description: Skip repositories with conflicts and leave them unchanged
   - Use case: When you want to update only clean repositories
   - Git operation: Checks for conflicts first, skips if any exist

### Example Configuration

```yaml
sync:
  sync_mode:
    cleanup_orphans: true
    conflict_resolution: "remote-overwrite"
```

## Integration Points

- **Development Environment**: Coordinates with `dev-env` for environment-specific repository access
- **Network Management**: Integrates with `net-env` for proxy and VPN-aware cloning
- **SSH Configuration**: Works with `ssh-config` for Git authentication
- **Configuration Generation**: Generates repository-specific configurations through `gen-config`

## Security Considerations

- **Token Security**: Secure storage and handling of authentication tokens
- **SSH Key Management**: Integration with SSH agent and key management
- **Repository Verification**: Verification of repository authenticity and integrity
- **Access Control**: Respect for repository permissions and access controls
- **Audit Logging**: Complete logging of all repository operations
- **Network Security**: Support for VPN and proxy configurations
- **Backup Encryption**: Encrypted backups of repository metadata

## Platform Support

- **GitHub**: Full support for organizations, repositories, and GitHub Enterprise
- **GitLab**: Support for groups, subgroups, and GitLab CE/EE instances
- **Gitea**: Support for organizations and custom Gitea instances
- **Gogs**: Support for organizations and custom Gogs instances
- **Git**: Support for any Git repository with HTTPS/SSH access

## Performance Optimization

- **Parallel Processing**: Configurable parallel cloning and updates
- **Rate Limiting**: Respect for API rate limits and burst controls
- **Incremental Updates**: Smart detection of changes for efficient updates
- **Caching**: API response caching to reduce unnecessary requests
- **Bandwidth Management**: Configurable bandwidth limits and scheduling
- **Storage Optimization**: Efficient storage usage and cleanup
