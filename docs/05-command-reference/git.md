# git Command Reference

Unified Git operations and platform management, combining repository operations, configuration management, webhook handling, and event processing.

## Synopsis

```bash
gz git <subcommand> [flags]
```

## Description

The `git` command provides a unified interface for Git-related operations, from basic repository management to advanced platform administration.

## Subcommands

### Repository Operations

#### `gz git repo clone-or-update`

Intelligently clone new repositories or update existing ones with multiple strategies.

```bash
gz git repo clone-or-update <repository-url> [target-path] [flags]
```

**Arguments:**
- `repository-url` - Git repository URL (HTTPS, SSH, or ssh:// format)
- `target-path` - Optional target directory (auto-extracts repo name if omitted)

**Flags:**
- `--strategy`, `-s` - Update strategy: rebase, reset, clone, skip, pull, fetch (default: rebase)
- `--branch`, `-b` - Branch to check out
- `--shallow` - Use shallow clone (default: false)
- `--depth` - Depth for shallow clone (default: 1)

**Supported URL Formats:**
- HTTPS: `https://github.com/user/repo.git`
- SSH: `git@github.com:user/repo.git`
- SSH Protocol: `ssh://git@github.com/user/repo.git`

**Strategies:**
- `rebase` (default): Rebase local changes on top of remote
- `reset`: Hard reset to match remote state (discards local changes)
- `clone`: Remove existing directory and perform fresh clone
- `skip`: Leave existing repository unchanged
- `pull`: Standard git pull (merge remote changes)
- `fetch`: Only fetch remote changes without updating working tree

**Examples:**
```bash
# Clone new repository (auto-extracts name)
gz git repo clone-or-update https://github.com/user/awesome-project.git

# Clone to specific directory
gz git repo clone-or-update https://github.com/user/repo.git my-local-name

# Update with specific strategy
gz git repo clone-or-update https://github.com/user/repo.git --strategy reset

# Clone specific branch
gz git repo clone-or-update https://github.com/user/repo.git -b develop

# Force fresh clone
gz git repo clone-or-update https://github.com/user/repo.git --strategy clone
```

### Configuration Management

#### `gz git config`

Manage repository configurations at scale across organizations.

```bash
gz git config <action> [flags]
```

**Actions:**

##### `gz git config audit`
Audit repository settings against compliance frameworks.

```bash
gz git config audit --org <organization> [flags]
```

**Flags:**
- `--org` - Organization name (required)
- `--framework` - Compliance framework: SOC2, GDPR, HIPAA, PCI-DSS
- `--output` - Output format: table, json, yaml, csv (default: table)
- `--output-file` - Save results to file
- `--severity` - Minimum severity: low, medium, high, critical

**Examples:**
```bash
# Basic audit
gz git config audit --org myorg

# SOC2 compliance check
gz git config audit --org myorg --framework SOC2

# Export results
gz git config audit --org myorg --output json --output-file audit.json
```

##### `gz git config apply`
Apply configuration from file to repositories.

```bash
gz git config apply --config <config-file> [flags]
```

**Flags:**
- `--config` - Configuration file path (required)
- `--org` - Target organization
- `--repo` - Target repository (format: org/repo)
- `--dry-run` - Preview changes without applying
- `--force` - Apply changes without confirmation
- `--parallel` - Number of concurrent operations (default: 5)

**Examples:**
```bash
# Apply configuration
gz git config apply --config repo-standards.yaml

# Dry run first
gz git config apply --config repo-standards.yaml --dry-run

# Apply to specific organization
gz git config apply --config repo-standards.yaml --org myorg
```

##### `gz git config diff`
Show configuration differences.

```bash
gz git config diff --org <organization> [flags]
```

**Flags:**
- `--org` - Organization name (required)
- `--repo` - Specific repository
- `--baseline` - Baseline configuration file
- `--output` - Output format: unified, side-by-side, json

**Examples:**
```bash
# Show differences for organization
gz git config diff --org myorg

# Compare specific repository
gz git config diff --org myorg --repo important-service

# Compare against baseline
gz git config diff --org myorg --baseline security-baseline.yaml
```

### Webhook Management

#### `gz git webhook`

Create, manage, and monitor webhooks across repositories and organizations.

```bash
gz git webhook <action> [flags]
```

**Actions:**

##### `gz git webhook create`
Create a new webhook.

```bash
gz git webhook create --org <org> --repo <repo> --url <webhook-url> [flags]
```

**Flags:**
- `--org` - Organization name (required)
- `--repo` - Repository name (required)
- `--url` - Webhook URL (required)
- `--events` - Webhook events (comma-separated, default: push)
- `--secret` - Webhook secret
- `--content-type` - Content type: json, form (default: json)
- `--active` - Webhook active state (default: true)

**Examples:**
```bash
# Basic webhook
gz git webhook create --org myorg --repo myrepo --url https://example.com/webhook

# Webhook with specific events
gz git webhook create --org myorg --repo myrepo --url https://example.com/webhook --events push,pull_request

# Webhook with secret
gz git webhook create --org myorg --repo myrepo --url https://example.com/webhook --secret mysecret
```

##### `gz git webhook list`
List webhooks for organization or repository.

```bash
gz git webhook list --org <organization> [flags]
```

**Flags:**
- `--org` - Organization name (required)
- `--repo` - Repository name (optional, lists all org webhooks if omitted)
- `--output` - Output format: table, json, yaml
- `--active-only` - Show only active webhooks

**Examples:**
```bash
# List organization webhooks
gz git webhook list --org myorg

# List repository webhooks
gz git webhook list --org myorg --repo myrepo

# JSON output
gz git webhook list --org myorg --output json
```

##### `gz git webhook delete`
Delete a webhook.

```bash
gz git webhook delete --id <webhook-id> [flags]
```

**Flags:**
- `--id` - Webhook ID (required)
- `--org` - Organization name
- `--repo` - Repository name
- `--confirm` - Skip confirmation prompt

##### `gz git webhook bulk`
Bulk webhook operations.

```bash
gz git webhook bulk <action> --org <organization> [flags]
```

**Actions:**
- `create` - Create webhooks from configuration
- `update` - Update existing webhooks
- `delete` - Delete webhooks matching criteria

**Flags:**
- `--config` - Configuration file for bulk operations
- `--filter` - Filter criteria for bulk operations

**Examples:**
```bash
# Bulk create from config
gz git webhook bulk create --org myorg --config webhooks.yaml

# Bulk delete inactive webhooks
gz git webhook bulk delete --org myorg --filter "active=false"
```

### Event Processing

#### `gz git event`

Monitor and process GitHub events in real-time.

```bash
gz git event <action> [flags]
```

**Actions:**

##### `gz git event server`
Start webhook event server.

```bash
gz git event server --port <port> [flags]
```

**Flags:**
- `--port` - Server port (default: 8080)
- `--host` - Server host (default: localhost)
- `--secret` - Webhook secret for validation
- `--ssl-cert` - SSL certificate file
- `--ssl-key` - SSL private key file

**Examples:**
```bash
# Basic server
gz git event server --port 8080

# Secure server with secret
gz git event server --port 8080 --secret webhook-secret

# HTTPS server
gz git event server --port 8443 --ssl-cert cert.pem --ssl-key key.pem
```

##### `gz git event list`
List recent events.

```bash
gz git event list --org <organization> [flags]
```

**Flags:**
- `--org` - Organization name (required)
- `--repo` - Repository name (optional)
- `--type` - Event type filter: push, pull_request, issues, etc.
- `--limit` - Maximum number of events (default: 50)
- `--since` - Show events since timestamp
- `--follow` - Follow events in real-time

**Examples:**
```bash
# List recent push events
gz git event list --org myorg --type push

# Follow events in real-time
gz git event list --org myorg --follow

# Last 100 events
gz git event list --org myorg --limit 100
```

##### `gz git event get`
Get specific event details.

```bash
gz git event get --id <event-id> [flags]
```

**Flags:**
- `--id` - Event ID (required)
- `--output` - Output format: json, yaml, table

##### `gz git event metrics`
View event processing metrics.

```bash
gz git event metrics [flags]
```

**Flags:**
- `--org` - Organization filter
- `--period` - Time period: 1h, 24h, 7d, 30d (default: 24h)
- `--output` - Output format: table, json, dashboard

**Examples:**
```bash
# Basic metrics
gz git event metrics

# Organization-specific metrics
gz git event metrics --org myorg --period 7d

# Dashboard output
gz git event metrics --output dashboard
```

## Configuration

### Repository Configuration Format

```yaml
# repo-config.yaml
version: "1.0"

settings:
  # Repository settings
  has_issues: true
  has_projects: false
  has_wiki: false
  allow_merge_commit: false
  allow_squash_merge: true
  allow_rebase_merge: true
  delete_branch_on_merge: true

  # Security settings
  security:
    vulnerability_alerts: true
    security_and_analysis:
      secret_scanning: true
      secret_scanning_push_protection: true

  # Branch protection
  branch_protection:
    main:
      required_status_checks:
        strict: true
        contexts: ["ci/tests"]
      enforce_admins: true
      required_pull_request_reviews:
        required_approving_review_count: 2
        dismiss_stale_reviews: true
        require_code_owner_reviews: true
```

### Webhook Configuration Format

```yaml
# webhooks.yaml
version: "1.0"

webhooks:
  - name: "ci-webhook"
    url: "https://ci.example.com/webhook"
    events: ["push", "pull_request"]
    secret: "${WEBHOOK_SECRET}"
    content_type: "json"
    active: true

  - name: "notification-webhook"
    url: "https://notify.example.com/webhook"
    events: ["issues", "pull_request_review"]
    active: true
```

## Global Flags

- `--config` - Configuration file path
- `--debug` - Enable debug logging
- `--dry-run` - Preview operations without executing
- `--output` - Output format: table, json, yaml, csv
- `--quiet` - Suppress non-error output
- `--verbose` - Enable verbose output

## Authentication

Set up authentication using environment variables:

```bash
export GITHUB_TOKEN="ghp_..."
export GITHUB_ENTERPRISE_TOKEN="ghp_..."  # For GitHub Enterprise
```

Or in configuration files:
```yaml
github:
  token: "${GITHUB_TOKEN}"
  enterprise:
    token: "${GITHUB_ENTERPRISE_TOKEN}"
    base_url: "https://github.company.com"
```

## Examples

### Complete Workflow

```bash
# 1. Clone or update a repository
gz git repo clone-or-update https://github.com/myorg/important-service.git

# 2. Audit repository configuration
gz git config audit --org myorg --framework SOC2

# 3. Apply security policies
gz git config apply --config security-policies.yaml --org myorg

# 4. Set up monitoring webhooks
gz git webhook create --org myorg --repo important-service --url https://monitor.example.com/webhook

# 5. Start event processing
gz git event server --port 8080 --secret webhook-secret
```

### Repository Management Script

```bash
#!/bin/bash
set -e

ORG="mycompany"
REPOS=("api-service" "web-frontend" "mobile-app")

# Clone/update all repositories
for repo in "${REPOS[@]}"; do
    echo "Processing $repo..."
    gz git repo clone-or-update "https://github.com/$ORG/$repo.git" --strategy rebase
done

# Apply configuration standards
gz git config apply --config standards.yaml --org "$ORG" --dry-run
read -p "Apply changes? (y/N) " -n 1 -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    gz git config apply --config standards.yaml --org "$ORG"
fi
```

### Organization Setup

```bash
# Comprehensive organization setup
gz git config audit --org myorg --output json > audit.json
gz git config apply --config enterprise-policies.yaml --org myorg
gz git webhook bulk create --org myorg --config webhooks.yaml
gz git event server --port 8080 &
```

## Error Handling

### Common Issues

1. **Authentication Error**
   ```bash
   # Test authentication
   gz git config audit --org myorg --dry-run
   ```

2. **Rate Limiting**
   ```bash
   # Reduce parallelism
   gz git config apply --config policy.yaml --parallel 2
   ```

3. **Repository Not Found**
   ```bash
   # Check repository URL and permissions
   gz git repo clone-or-update https://github.com/org/repo.git --debug
   ```

## Related Commands

- [`gz synclone`](synclone.md) - Multi-repository synchronization
- [`gz quality`](quality.md) - Code quality management

## See Also

- [Git Unified Command Guide](../03-core-features/git-unified-command.md)
- [Repository Configuration Examples](../../examples/github/)
- [Webhook Management Guide](../08-integrations/webhook-management-guide.md)
