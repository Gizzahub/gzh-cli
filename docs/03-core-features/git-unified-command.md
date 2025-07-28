# Git Unified Command Guide

The `gz git` command provides a unified interface for managing Git platforms, consolidating repository configuration, webhook management, and event processing into a single, cohesive command structure.

## Overview

Instead of using multiple standalone commands, the git unified command groups all Git-related operations under a single namespace:

```bash
gz git <subcommand> [options]
```

## Available Subcommands

### 1. Repository Configuration (`gz git config`)

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

### 2. Webhook Management (`gz git webhook`)

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

### 3. Event Processing (`gz git event`)

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
All Git platform operations are organized under a single command, making it easier to discover and use related features.

### 2. **Consistent Interface**
The unified command provides a consistent experience across different Git operations, with similar flags and patterns.

### 3. **Future-Proof**
New Git-related features can be added as subcommands without cluttering the top-level command namespace.

### 4. **Simplified Learning Curve**
Users only need to remember `gz git` for all Git platform operations, rather than multiple standalone commands.

## Migration from Standalone Commands

If you're currently using standalone commands, here's how to migrate:

| Old Command | New Command | Notes |
|-------------|-------------|-------|
| `gz repo-config` | `gz git config` | Identical functionality |
| `gz repo-config webhook` | `gz git webhook` | Moved to top level under git |
| N/A | `gz git event` | New functionality |

### Backward Compatibility

The `gz repo-config` command remains available for backward compatibility, but we recommend using `gz git config` for new scripts and workflows.

## Examples

### Complete Workflow Example

```bash
# 1. Configure repositories
gz git config apply --config repo-standards.yaml

# 2. Set up webhooks for monitoring
gz git webhook create --org myorg --url https://monitor.example.com/events

# 3. Start event processing server
gz git event server --port 8080 --secret webhook-secret

# 4. Monitor events in real-time
gz git event list --follow --type push
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

1. **Use Templates**: Define repository templates in your configuration for consistency
2. **Dry Run First**: Always use `--dry-run` before applying changes
3. **Audit Regularly**: Run audit commands to ensure compliance
4. **Monitor Events**: Set up event monitoring for security and compliance

## Related Documentation

- [Repository Configuration Guide](repository-management/repo-config-user-guide.md)
- [Webhook Management Guide](../08-integrations/webhook-management-guide.md)
- [Command Structure Overview](../analysis/current-command-structure.md)