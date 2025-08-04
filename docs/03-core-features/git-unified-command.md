# Git Unified Command Guide

The `gz git` command provides a unified interface for managing Git operations and platforms, consolidating repository cloning, configuration management, webhook handling, and event processing into a single, cohesive command structure.

## Overview

The git unified command groups all Git-related operations under a single namespace, providing a consistent interface for both local Git operations and platform management:

```bash
gz git <subcommand> [options]
```

## Available Subcommands

### 1. Repository Clone or Update (`gz git repo clone-or-update`)

Intelligently clone new repositories or update existing ones with multiple strategies.

```bash
# Clone a new repository (auto-extracts repo name from URL)
gz git repo clone-or-update https://github.com/user/repo.git

# Clone to specific directory
gz git repo clone-or-update https://github.com/user/repo.git my-local-repo

# Update existing repository with rebase (default)
gz git repo clone-or-update https://github.com/user/repo.git --strategy rebase

# Force fresh clone
gz git repo clone-or-update https://github.com/user/repo.git --strategy clone

# Specify branch
gz git repo clone-or-update https://github.com/user/repo.git -b develop
```

**Supported Strategies:**
- `rebase` (default): Rebase local changes on top of remote
- `reset`: Hard reset to match remote state
- `clone`: Remove existing and perform fresh clone
- `skip`: Leave existing repository unchanged
- `pull`: Standard git pull (merge)
- `fetch`: Only fetch remote changes

**URL Support:**
- HTTPS: `https://github.com/user/repo.git`
- SSH: `git@github.com:user/repo.git`
- SSH Protocol: `ssh://git@github.com/user/repo.git`

### 2. Repository Configuration (`gz git config`)

Manage repository configurations at scale across your organization.

```bash
# Audit repository settings
gz git config audit --org myorg --framework SOC2

# Apply configuration from file
gz git config apply --config repo-config.yaml --dry-run

# Show configuration differences
gz git config diff --org myorg --repo myrepo
```

**Note**: This command delegates to the `repo-config` functionality, providing the same features through a more intuitive interface.

### 3. Webhook Management (`gz git webhook`)

Create, manage, and monitor webhooks across repositories and organizations.

```bash
# Create a webhook
gz git webhook create --org myorg --repo myrepo --url https://example.com/webhook

# List all webhooks
gz git webhook list --org myorg

# Delete a webhook
gz git webhook delete --id 12345

# Bulk webhook operations
gz git webhook bulk create --org myorg --config webhooks.yaml
```

### 4. Event Processing (`gz git event`)

Monitor and process GitHub events in real-time.

```bash
# Start webhook server
gz git event server --port 8080 --secret mysecret

# List recent events
gz git event list --org myorg --type push --limit 50

# Get specific event details
gz git event get --id event123

# View event metrics
gz git event metrics --output json
```

## Why Use the Unified Command?

### 1. **Logical Grouping**
All Git operations are organized under a single command, from basic repository operations to advanced platform management.

### 2. **Consistent Interface**
The unified command provides a consistent experience across different Git operations, with similar flags and patterns.

### 3. **Smart Operations**
Commands like `clone-or-update` intelligently handle both new and existing repositories, reducing manual decision-making.

### 4. **Future-Proof**
New Git-related features can be added as subcommands without cluttering the top-level command namespace.

### 5. **Simplified Learning Curve**
Users only need to remember `gz git` for all Git operations, rather than multiple standalone commands.

## Migration from Standalone Commands

If you're currently using standalone commands, here's how to migrate:

| Old Command | New Command | Notes |
|-------------|-------------|-------|
| `gz repo-config` | `gz git config` | Identical functionality |
| `gz repo-config webhook` | `gz git webhook` | Moved to top level under git |
| `git clone` + manual update | `gz git repo clone-or-update` | Smart clone/update logic |
| N/A | `gz git event` | New functionality |

### Backward Compatibility

The `gz repo-config` command remains available for backward compatibility, but we recommend using `gz git config` for new scripts and workflows.

## Examples

### Complete Workflow Example

```bash
# 1. Clone or update repository
gz git repo clone-or-update https://github.com/myorg/myrepo.git

# 2. Configure repository settings
gz git config apply --config repo-standards.yaml

# 3. Set up webhooks for monitoring
gz git webhook create --org myorg --url https://monitor.example.com/events

# 4. Start event processing server
gz git event server --port 8080 --secret webhook-secret

# 5. Monitor events in real-time
gz git event list --follow --type push
```

### Repository Management Example

```bash
# Clone multiple repositories with smart update
for repo in api-service web-frontend mobile-app; do
  gz git repo clone-or-update https://github.com/myorg/$repo.git
done

# Update all repositories with rebase strategy
find . -name ".git" -type d | while read gitdir; do
  repo_dir=$(dirname "$gitdir")
  cd "$repo_dir"
  gz git repo clone-or-update . --strategy rebase
  cd -
done
```

### Organization-Wide Setup

```bash
# Audit current state
gz git config audit --org mycompany --output audit-report.json

# Apply security policies
gz git config apply --config security-policies.yaml --org mycompany

# Configure webhooks for all repositories
gz git webhook bulk create --org mycompany --config webhook-config.yaml

# Monitor compliance
gz git event metrics --org mycompany --output dashboard
```

## Configuration

The git unified command uses the same configuration files as the underlying commands:

- Repository configuration: `~/.gzh/repo-config.yaml`
- Webhook definitions: `~/.gzh/webhooks.yaml`
- Event processing: `~/.gzh/events.yaml`

## Best Practices

1. **Smart Cloning**: Use `clone-or-update` instead of manual git operations
2. **Choose Strategy Wisely**:
   - Use `rebase` for clean history in development
   - Use `reset` for CI/CD and read-only mirrors
   - Use `pull` when preserving merge history
3. **Use Templates**: Define repository templates in your configuration for consistency
4. **Dry Run First**: Always use `--dry-run` before applying changes
5. **Audit Regularly**: Run audit commands to ensure compliance
6. **Monitor Events**: Set up event monitoring for security and compliance

## Related Documentation

- [Repository Synchronization Guide](synclone-guide.md)
- [Repository Configuration Guide](repository-management/repo-config-user-guide.md)
- [Webhook Management Guide](../08-integrations/webhook-management-guide.md)
- [Getting Started](../01-getting-started/)
