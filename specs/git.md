# Git Unified Command Specification (Updated)

## Overview

The `gz git` command provides a unified interface for managing Git platforms (GitHub, GitLab, Gitea, Gogs), consolidating repository management, webhook operations, event processing, and cross-platform synchronization into a single, cohesive command structure.

## Purpose

The git unified command addresses the need for:

- **Consistent Interface**: Single entry point for all Git platform operations
- **Cross-Platform Support**: Abstract differences between GitHub, GitLab, Gitea, and Gogs
- **Repository Lifecycle**: Complete repository management from creation to archival
- **Cross-Platform Sync**: Synchronize repositories between different Git platforms
- **Simplified Learning**: Users only need to remember `gz git` for all Git operations
- **Future Expansion**: Extensible architecture for additional Git platform features

## Command Structure

```
gz git <subcommand> [options]
```

## Current Implementation (2025-08)

### Available Subcommands

| Subcommand | Purpose                                  | Implementation Status     |
| ---------- | ---------------------------------------- | ------------------------- |
| `repo`     | Repository lifecycle and sync management | ✅ Implemented (Enhanced) |
| `webhook`  | Webhook management and automation        | ✅ Implemented            |
| `event`    | Event processing and monitoring          | ✅ Implemented            |

### Command Hierarchy

```
gz git
   repo             # Repository management (NEW STRUCTURE)
      clone-or-update  # Smart clone/update with strategies
      create         # Create new repository
      delete         # Delete repository
      archive        # Archive repository
      list           # List repositories
      sync           # Cross-platform synchronization (NEW)
   webhook          # Webhook management
      create        # Create webhooks
      list          # List webhooks
      update        # Update webhook configuration
      delete        # Delete webhooks
      bulk          # Bulk webhook operations
   event            # Event processing
      server        # Run webhook event server
      list          # List received events
      get           # Get specific event details
      metrics       # Event processing metrics
      test          # Test event handling
```

## Subcommand Specifications

### 1. Repository Management (`gz git repo`)

**Purpose**: Comprehensive repository lifecycle management and cross-platform synchronization.

**Description**: This command manages the complete repository lifecycle including creation, deletion, archival, smart cloning/updating, and cross-platform synchronization between GitHub, GitLab, and Gitea.

#### Subcommands

##### `gz git repo clone-or-update`

Smart repository clone with update strategies.

```bash
gz git repo clone-or-update <url> [target-path] [--strategy <strategy>] [--branch <branch>]
```

**Strategies**:

- `rebase` (default): Rebase local changes on remote
- `reset`: Hard reset to match remote state
- `clone`: Remove existing and fresh clone
- `skip`: Leave existing repository unchanged
- `pull`: Standard git pull (merge)
- `fetch`: Only fetch remote changes

**Examples**:

```bash
# Basic clone/update
gz git repo clone-or-update https://github.com/user/repo.git

# With specific strategy
gz git repo clone-or-update https://github.com/user/repo.git --strategy rebase

# Clone to specific path
gz git repo clone-or-update https://github.com/user/repo.git ~/projects/myrepo
```

##### `gz git repo create`

Create a new repository on a Git platform.

```bash
gz git repo create --name <name> --org <org> [--private] [--description <desc>]
```

**Options**:

- `--name` - Repository name (required)
- `--org` - Organization/owner (required)
- `--private` - Create as private repository
- `--description` - Repository description
- `--template` - Template repository to use
- `--init` - Initialize with README
- `--gitignore` - Language-specific .gitignore template
- `--license` - License type (MIT, Apache-2.0, GPL-3.0, etc.)

**Examples**:

```bash
# Create public repository
gz git repo create --name myapp --org myorg --description "My application"

# Create private repository with initialization
gz git repo create --name internal-tool --org myorg --private --init --license MIT
```

##### `gz git repo delete`

Delete a repository from a Git platform.

```bash
gz git repo delete --name <name> --org <org> --confirm
```

**Options**:

- `--name` - Repository name (required)
- `--org` - Organization/owner (required)
- `--confirm` - Confirm deletion (required for safety)
- `--backup` - Create backup before deletion

**Examples**:

```bash
# Delete repository with confirmation
gz git repo delete --name old-project --org myorg --confirm

# Delete with backup
gz git repo delete --name deprecated-app --org myorg --backup --confirm
```

##### `gz git repo archive`

Archive a repository (make read-only).

```bash
gz git repo archive --name <name> --org <org>
```

**Options**:

- `--name` - Repository name (required)
- `--org` - Organization/owner (required)
- `--message` - Archive reason/message

##### `gz git repo list`

List repositories with filtering options.

```bash
gz git repo list --org <org> [--filter <filter>] [--format <format>]
```

**Options**:

- `--org` - Organization to list (required)
- `--filter` - Filter repositories (archived, private, public, fork)
- `--format` - Output format (table, json, yaml, csv)
- `--sort` - Sort by (name, created, updated, size)

##### `gz git repo sync` (NEW)

Synchronize repositories between different Git platforms.

```bash
gz git repo sync --from <source> --to <destination> [options]
```

**Options**:

- `--from` - Source (e.g., `github:org/repo` or `github:org`)
- `--to` - Destination (e.g., `gitlab:group/repo` or `gitlab:group`)
- `--create-missing` - Create repos that don't exist in destination
- `--include-issues` - Sync issues and pull requests
- `--include-wiki` - Sync wiki content
- `--include-releases` - Sync releases and tags
- `--dry-run` - Preview changes without applying

**Examples**:

```bash
# Sync single repository
gz git repo sync --from github:myorg/repo --to gitlab:mygroup/repo

# Sync entire organization
gz git repo sync --from github:myorg --to gitea:myorg --create-missing

# Sync with all features
gz git repo sync --from github:org/repo --to gitlab:group/repo \
  --include-issues --include-wiki --include-releases

# Dry run to preview
gz git repo sync --from github:org/repo --to gitlab:group/repo --dry-run
```

### 2. Webhook Management (`gz git webhook`)

**Purpose**: Manage webhooks across Git platforms.

**Description**: Create, update, delete, and manage webhooks for repositories and organizations.

#### Subcommands

##### `gz git webhook create`

Create a new webhook.

```bash
gz git webhook create --org <org> --repo <repo> --url <webhook-url> [options]
```

**Options**:

- `--org` - Organization (required)
- `--repo` - Repository name (optional, for repo-specific webhook)
- `--url` - Webhook URL (required)
- `--events` - Comma-separated list of events
- `--secret` - Webhook secret for validation
- `--active` - Whether webhook is active (default: true)

##### `gz git webhook list`

List webhooks for an organization or repository.

```bash
gz git webhook list --org <org> [--repo <repo>] [--format <format>]
```

##### `gz git webhook delete`

Delete a webhook.

```bash
gz git webhook delete --org <org> --repo <repo> --id <webhook-id>
```

### 3. Event Processing (`gz git event`)

**Purpose**: Process and handle Git platform events.

**Description**: Run event servers, process webhooks, and manage event-driven automation.

#### Subcommands

##### `gz git event server`

Run a webhook event server.

```bash
gz git event server --port <port> [--handler <handler>]
```

**Options**:

- `--port` - Server port (default: 8080)
- `--handler` - Event handler script or command
- `--secret` - Webhook secret for validation
- `--log-level` - Logging level

## Architecture Changes (2025-08)

### Repository Module Restructuring

The Git command has been restructured into modular components:

```
cmd/git/
├── git.go              # Main command entry
├── repo/               # Repository subcommands
│   ├── repo_root.go    # Repo command root
│   ├── repo_clone_or_update.go
│   ├── repo_create.go
│   ├── repo_delete.go
│   ├── repo_archive.go
│   ├── repo_list.go
│   ├── repo_sync.go    # NEW: Cross-platform sync
│   └── adapter.go      # Platform adapters
├── webhook/            # Webhook subcommands
│   ├── webhook.go
│   └── adapter.go
└── event/              # Event subcommands
    ├── event.go
    └── adapter.go
```

### Test Coverage Improvements

- Git package: 91.7% coverage achieved
- Comprehensive unit tests for all repo operations
- Integration tests for cross-platform sync
- Mock-based testing for platform APIs

## Configuration

### Repository Sync Configuration

```yaml
git:
  sync:
    default_options:
      create_missing: true
      include_issues: false
      include_wiki: false
      include_releases: true

    mappings:
      - from: github:myorg
        to: gitlab:mygroup
        options:
          include_issues: true

      - from: github:important-org/critical-repo
        to: gitea:backup-org/critical-repo
        options:
          include_wiki: true
          include_releases: true
```

## Migration Path

### From Separate Commands to Unified Git

Users migrating from older versions should update their scripts:

**Old**:

```bash
gz repo-config apply --config config.yaml
gz synclone github --org myorg
```

**New**:

```bash
gz git config apply --config config.yaml  # Delegated to repo-config
gz git repo sync --from github:myorg --to local:~/repos
```

## Future Enhancements

1. **Git Flow Integration**: Support for git flow workflows
1. **PR/MR Management**: Create and manage pull/merge requests
1. **Code Review**: Integration with code review workflows
1. **CI/CD Triggers**: Direct CI/CD pipeline management
1. **Repository Templates**: Advanced template management
1. **Access Control**: Team and permission management
1. **Repository Insights**: Statistics and analytics

## Testing

### Unit Tests

- Repository operations: `cmd/git/repo/*_test.go`
- Webhook management: `cmd/git/webhook/*_test.go`
- Event processing: `cmd/git/event/*_test.go`

### Integration Tests

- Cross-platform sync: `test/integration/git/sync_test.go`
- Webhook delivery: `test/integration/git/webhook_test.go`

### Coverage Goals

- Unit test coverage: >90%
- Integration test coverage: >70%
- Current achievement: 91.7%

## Documentation

- User Guide: `docs/30-features/git-management.md`
- API Reference: `docs/50-api-reference/git-commands.md`
- Configuration: `docs/40-configuration/git-config.md`

## Compatibility

- **Git Platforms**: GitHub, GitLab, Gitea, Gogs
- **Git Versions**: 2.0+
- **Operating Systems**: Linux, macOS, Windows (WSL recommended)
- **Authentication**: Token-based (OAuth2, Personal Access Tokens)

## Security Considerations

1. **Token Management**: Never log or expose authentication tokens
1. **Webhook Secrets**: Use strong secrets for webhook validation
1. **Repository Access**: Respect platform permissions and ACLs
1. **Data Sync**: Validate data integrity during cross-platform sync
1. **Rate Limiting**: Implement adaptive rate limiting for API calls
