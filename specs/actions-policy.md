<!-- 🚫 AI_MODIFY_PROHIBITED -->

<!-- This file should not be modified by AI agents -->

# GitHub Actions Policy Management Specification

## Overview

The `actions-policy` command provides comprehensive GitHub Actions policy management and enforcement capabilities. It enables organizations to create, validate, and enforce security policies across repositories, ensuring consistent governance and compliance for GitHub Actions workflows.

## Commands

### Core Commands

- `gz actions-policy create` - Create a new Actions policy
- `gz actions-policy enforce` - Enforce Actions policy on a repository
- `gz actions-policy validate` - Validate repository against Actions policy
- `gz actions-policy list` - List all Actions policies
- `gz actions-policy show` - Show Actions policy details
- `gz actions-policy delete` - Delete an Actions policy
- `gz actions-policy monitor` - Monitor policy compliance across organization

### Create Policy (`gz actions-policy create`)

**Purpose**: Create a new GitHub Actions policy with specified configuration

**Features**:

- Multiple policy templates (default, strict, permissive)
- Organization and repository-level policies
- Configurable security settings
- Policy versioning and metadata
- Automated policy ID generation

**Usage**:

```bash
gz actions-policy create my-security-policy --org myorg --template strict      # Create strict security policy
gz actions-policy create dev-policy --org myorg --repo myrepo --template permissive  # Repository-specific policy
gz actions-policy create custom-policy --org myorg --description "Custom security rules"  # Custom policy
gz actions-policy create audit-policy --org myorg --template default --tags security,audit  # Tagged policy
```

**Parameters**:

- `policy-name` (required): Name of the policy to create
- `--org` (required): Target organization
- `--repo`: Target repository (optional, for repo-level policies)
- `--template` (default: default): Policy template (default, strict, permissive)
- `--description`: Policy description
- `--tags`: Policy tags for categorization
- `--enabled` (default: true): Enable policy immediately

**Templates**:

- `default`: Balanced security policy with moderate restrictions
- `strict`: High security policy with restrictive settings
- `permissive`: Permissive policy for development environments

### Enforce Policy (`gz actions-policy enforce`)

**Purpose**: Apply and enforce a specific Actions policy on a repository

**Features**:

- Dry-run mode for validation preview
- Force enforcement option
- Timeout configuration
- Detailed change reporting
- Rollback capability on failure

**Usage**:

```bash
gz actions-policy enforce policy-123 myorg myrepo                    # Enforce policy on repository
gz actions-policy enforce policy-123 myorg myrepo --dry-run          # Preview changes without applying
gz actions-policy enforce policy-123 myorg myrepo --force --timeout 600  # Force enforcement with custom timeout
```

**Parameters**:

- `policy-id` (required): Policy ID to enforce
- `org` (required): Target organization
- `repo` (required): Target repository
- `--dry-run` (default: false): Perform validation only, don't apply changes
- `--force` (default: false): Force enforcement even if validation fails
- `--timeout` (default: 300): Enforcement timeout in seconds

### Validate Policy (`gz actions-policy validate`)

**Purpose**: Validate a repository's current configuration against an Actions policy

**Features**:

- Comprehensive compliance checking
- Severity-based filtering
- Detailed validation reports
- Suggestion generation for remediation
- Configuration gap analysis

**Usage**:

```bash
gz actions-policy validate policy-123 myorg myrepo                   # Basic validation
gz actions-policy validate policy-123 myorg myrepo --detailed        # Detailed validation results
gz actions-policy validate policy-123 myorg myrepo --severity high   # Filter by severity level
```

**Parameters**:

- `policy-id` (required): Policy ID to validate against
- `org` (required): Target organization
- `repo` (required): Target repository
- `--detailed` (default: false): Show detailed validation results
- `--severity` (default: all): Filter by severity (all, low, medium, high, critical)

### List Policies (`gz actions-policy list`)

**Purpose**: Display all available Actions policies with filtering options

**Features**:

- Organization-based filtering
- Tag-based filtering
- Status filtering (enabled/disabled)
- Multiple output formats
- Policy metadata display

**Usage**:

```bash
gz actions-policy list                                              # List all policies
gz actions-policy list --org myorg                                  # Filter by organization
gz actions-policy list --tags security,audit --enabled-only        # Filter by tags and status
gz actions-policy list --format json                               # Output as JSON
```

**Parameters**:

- `--org`: Filter by organization
- `--tags`: Filter by tags
- `--enabled-only` (default: false): Show only enabled policies
- `--format` (default: table): Output format (table, json, yaml)

### Show Policy Details (`gz actions-policy show`)

**Purpose**: Display detailed information about a specific Actions policy

**Features**:

- Complete policy configuration display
- Security settings breakdown
- Permission analysis
- Metadata and versioning information
- Configuration summary

**Usage**:

```bash
gz actions-policy show policy-123                                  # Show policy details
gz actions-policy show policy-123 --format json                    # Output as JSON
```

**Parameters**:

- `policy-id` (required): Policy ID to display
- `--format` (default: table): Output format (table, json, yaml)

### Delete Policy (`gz actions-policy delete`)

**Purpose**: Remove an Actions policy from the system

**Features**:

- Safe deletion with confirmation
- Policy dependency checking
- Audit trail maintenance

**Usage**:

```bash
gz actions-policy delete policy-123                                # Delete specific policy
```

**Parameters**:

- `policy-id` (required): Policy ID to delete

### Monitor Compliance (`gz actions-policy monitor`)

**Purpose**: Continuously monitor policy compliance across all repositories in an organization

**Features**:

- Real-time compliance monitoring
- Configurable monitoring intervals
- Webhook integration for alerts
- Continuous or one-time execution
- Compliance reporting

**Usage**:

```bash
gz actions-policy monitor myorg                                    # One-time compliance check
gz actions-policy monitor myorg --continuous --interval 10m        # Continuous monitoring every 10 minutes
gz actions-policy monitor myorg --webhook-url https://alerts.example.com/webhook  # With webhook alerts
```

**Parameters**:

- `org` (required): Organization to monitor
- `--interval` (default: 5m): Monitoring interval
- `--continuous` (default: false): Run continuously until interrupted
- `--webhook-url`: Webhook URL for compliance alerts

## Policy Configuration

### Policy Structure

Actions policies contain the following configuration categories:

#### Permission Settings

- Default workflow permissions (read/write/restricted)
- Token permissions scope
- Repository access levels

#### Security Settings

- Fork pull request handling
- GitHub-owned actions allowance
- Marketplace actions policy
- Self-hosted runner policies

#### Secrets Management

- Maximum secret count limits
- Secret naming conventions
- Environment-specific restrictions

#### Runner Configuration

- Allowed runner types
- Self-hosted runner policies
- Resource limitations

### Policy Templates

#### Default Template

```yaml
permission_level: selected_actions
workflow_permissions:
  default_permissions: restricted
security_settings:
  allow_fork_prs: false
  allow_github_owned_actions: true
  allow_marketplace_actions: verified_creator
secrets_policy:
  max_secret_count: 50
runners:
  allowed_runner_types: ["ubuntu-latest", "windows-latest", "macos-latest"]
```

#### Strict Template

```yaml
permission_level: selected_actions
workflow_permissions:
  default_permissions: restricted
security_settings:
  allow_fork_prs: false
  allow_github_owned_actions: true
  allow_marketplace_actions: disabled
secrets_policy:
  max_secret_count: 10
runners:
  allowed_runner_types: ["ubuntu-latest"]
```

#### Permissive Template

```yaml
permission_level: all
workflow_permissions:
  default_permissions: write
security_settings:
  allow_fork_prs: true
  allow_github_owned_actions: true
  allow_marketplace_actions: all
secrets_policy:
  max_secret_count: 100
runners:
  allowed_runner_types: ["ubuntu-latest", "windows-latest", "macos-latest", "self-hosted"]
```

## Validation Rules

### Security Validation

- Permission level compliance
- Secret count limits
- Runner type restrictions
- Fork pull request policies

### Compliance Checks

- Workflow permission auditing
- Action marketplace compliance
- Self-hosted runner governance
- Environment protection rules

### Severity Levels

- **Critical**: Security vulnerabilities, unauthorized permissions
- **High**: Policy violations, non-compliant configurations
- **Medium**: Best practice deviations, optimization opportunities
- **Low**: Minor configuration issues, style preferences

## Integration

### GitHub API Integration

The actions-policy command integrates with GitHub APIs for:

- Repository configuration management
- Organization settings access
- Actions permissions control
- Security policy enforcement

### Authentication

Supports multiple authentication methods:

- GitHub personal access tokens
- GitHub Apps authentication
- Environment variable configuration

```bash
export GITHUB_TOKEN="ghp_your_token_here"
gz actions-policy list --org myorg
```

### Webhook Integration

Monitor command supports webhook notifications for:

- Policy violations
- Compliance changes
- Enforcement results
- Audit events

## Examples

### Complete Policy Management Workflow

```bash
# Create a new security policy
gz actions-policy create security-v1 --org myorg --template strict --description "Security policy v1.0"

# Validate existing repositories
gz actions-policy validate security-v1 myorg frontend-app --detailed
gz actions-policy validate security-v1 myorg backend-api --severity high

# Enforce policy with dry-run first
gz actions-policy enforce security-v1 myorg frontend-app --dry-run
gz actions-policy enforce security-v1 myorg frontend-app

# Monitor compliance across organization
gz actions-policy monitor myorg --continuous --interval 15m

# List and manage policies
gz actions-policy list --org myorg --format json
gz actions-policy show security-v1 --format yaml
```

### Automated Compliance Pipeline

```bash
#!/bin/bash
# Automated compliance check script

ORG="myorg"
POLICY="security-v1"

# Get list of repositories
REPOS=$(gz actions-policy list --org $ORG --format json | jq -r '.[] | .name')

# Validate each repository
for repo in $REPOS; do
  echo "Validating $repo..."
  gz actions-policy validate $POLICY $ORG $repo --severity high

  if [ $? -eq 0 ]; then
    echo "✅ $repo is compliant"
  else
    echo "❌ $repo has compliance issues"
    # Optional: Auto-enforce policy
    # gz actions-policy enforce $POLICY $ORG $repo --force
  fi
done
```

### Organization-wide Policy Enforcement

```bash
# Create organization-wide security policy
gz actions-policy create org-security --org myorg --template strict \
  --description "Organization-wide security policy" \
  --tags security,compliance

# Monitor and enforce across all repositories
gz actions-policy monitor myorg --webhook-url https://compliance.example.com/alerts

# Generate compliance report
gz actions-policy list --org myorg --format json > compliance-report.json
```

## Error Handling

### Common Errors

- **Policy not found**: Invalid policy ID provided
- **Insufficient permissions**: Missing GitHub API permissions
- **Rate limit exceeded**: GitHub API rate limiting
- **Validation failures**: Repository configuration conflicts
- **Network errors**: GitHub API connectivity issues

### Recovery Strategies

- **Authentication errors**: Verify GitHub token permissions
- **Rate limiting**: Implement exponential backoff
- **Policy conflicts**: Use force flag or resolve conflicts manually
- **Network issues**: Retry with timeout adjustments

## Security Considerations

### Permissions Required

GitHub token requires the following scopes:

- `repo`: Repository access
- `admin:org`: Organization administration
- `workflow`: Actions workflow management

### Audit Trail

All policy operations generate audit logs including:

- Policy creation and modifications
- Enforcement actions
- Validation results
- Compliance monitoring events

### Best Practices

- Use principle of least privilege for policies
- Regularly review and update policy templates
- Monitor compliance continuously
- Implement gradual rollout for new policies
- Maintain policy versioning and documentation
