<!-- 🚫 AI_MODIFY_PROHIBITED -->

<!-- This file should not be modified by AI agents -->

# Repository Configuration Management Specification

## Overview

The `repo-config` command provides comprehensive repository configuration management capabilities for standardizing and enforcing repository settings across organizations. It enables batch application of configuration templates, security policies, and compliance rules to multiple repositories simultaneously.

## Commands

### Core Commands

- `gz repo-config apply` - Apply repository configuration
- `gz repo-config audit` - Audit repository configurations
- `gz repo-config template` - Manage configuration templates
- `gz repo-config webhook` - Manage repository webhooks
- `gz repo-config validate` - Validate repository configurations
- `gz repo-config export` - Export repository configurations
- `gz repo-config import` - Import repository configurations
- `gz repo-config diff` - Compare repository configurations

### Apply Configuration (`gz repo-config apply`)

**Purpose**: Apply repository configuration to organizations and repositories

**Features**:

- Template-based configuration deployment
- Organization-wide and selective application
- Dry-run mode for preview
- Interactive confirmation
- Rollback capability
- Progress tracking and reporting

**Usage**:

```bash
gz repo-config apply --org myorg                               # Apply to all repositories
gz repo-config apply --filter "^api-.*"                        # Apply to matching repositories
gz repo-config apply --template security --org myorg          # Apply specific template
gz repo-config apply --dry-run --org myorg                     # Preview changes only
gz repo-config apply --interactive --template compliance       # Interactive confirmation
gz repo-config apply --config custom-config.yaml --org myorg  # Apply custom configuration
```

**Parameters**:

- `--org`: Target organization
- `--repo`: Specific repository (alternative to org)
- `--filter`: Repository name filter (regex pattern)
- `--template`: Configuration template name
- `--config`: Custom configuration file path
- `--dry-run` (default: false): Preview changes without applying
- `--interactive` (default: false): Prompt for confirmation
- `--force` (default: false): Apply changes without validation
- `--concurrency` (default: 5): Number of concurrent operations

### Audit Repositories (`gz repo-config audit`)

**Purpose**: Audit repository configurations for compliance and consistency

**Features**:

- Comprehensive configuration analysis
- Compliance rule validation
- Security setting verification
- Configuration drift detection
- Detailed audit reports

**Usage**:

```bash
gz repo-config audit --org myorg                               # Audit all repositories
gz repo-config audit --org myorg --template security          # Audit against template
gz repo-config audit --org myorg --format json --output audit.json  # Export audit results
gz repo-config audit --severity high --org myorg              # Filter by severity
```

**Parameters**:

- `--org` (required): Target organization
- `--template`: Audit against specific template
- `--severity`: Filter by issue severity (low, medium, high, critical)
- `--format` (default: table): Output format (table, json, yaml, csv)
- `--output`: Output file path
- `--include-private` (default: true): Include private repositories
- `--since`: Audit repositories modified since date

### Template Management (`gz repo-config template`)

**Purpose**: Manage configuration templates for consistent repository setup

**Features**:

- Template creation and modification
- Template validation and testing
- Template versioning
- Built-in template library
- Custom template support

**Usage**:

```bash
gz repo-config template list                                   # List available templates
gz repo-config template show security                          # Show template details
gz repo-config template create my-template --from security     # Create from existing template
gz repo-config template validate security                      # Validate template
gz repo-config template export security --output security.yaml # Export template
gz repo-config template import --file custom.yaml              # Import custom template
```

**Subcommands**:

- `list`: List available templates
- `show`: Display template details
- `create`: Create new template
- `validate`: Validate template syntax
- `export`: Export template to file
- `import`: Import template from file
- `delete`: Delete template

### Webhook Management (`gz repo-config webhook`)

**Purpose**: Manage repository webhooks across multiple repositories

**Features**:

- Bulk webhook creation and modification
- Webhook template support
- Event configuration management
- Webhook validation and testing
- Centralized webhook policies

**Usage**:

```bash
gz repo-config webhook list --org myorg                        # List all webhooks
gz repo-config webhook create --org myorg --url https://api.example.com/webhook  # Create webhook
gz repo-config webhook update --org myorg --id 123 --events push,pull_request  # Update webhook
gz repo-config webhook delete --org myorg --id 123            # Delete webhook
gz repo-config webhook test --org myorg --id 123              # Test webhook
```

**Subcommands**:

- `list`: List repository webhooks
- `create`: Create new webhook
- `update`: Update existing webhook
- `delete`: Delete webhook
- `test`: Test webhook functionality

### Validate Configuration (`gz repo-config validate`)

**Purpose**: Validate repository configurations against policies and best practices

**Features**:

- Policy compliance validation
- Security configuration checks
- Best practice verification
- Configuration syntax validation
- Custom rule support

**Usage**:

```bash
gz repo-config validate --org myorg                            # Validate all repositories
gz repo-config validate --org myorg --policy security         # Validate against policy
gz repo-config validate --config custom.yaml --rules strict   # Validate custom config
```

**Parameters**:

- `--org`: Target organization
- `--policy`: Validation policy name
- `--rules`: Validation rule set (basic, standard, strict)
- `--config`: Configuration file to validate
- `--fix` (default: false): Automatically fix issues where possible

### Export Configuration (`gz repo-config export`)

**Purpose**: Export repository configurations for backup or migration

**Features**:

- Complete configuration backup
- Selective configuration export
- Multiple output formats
- Batch export capabilities
- Configuration versioning

**Usage**:

```bash
gz repo-config export --org myorg --output backup.yaml        # Export all repositories
gz repo-config export --repo myorg/myrepo --format json       # Export single repository
gz repo-config export --org myorg --filter "^prod-.*" --output prod-configs/  # Export filtered repositories
```

**Parameters**:

- `--org`: Source organization
- `--repo`: Specific repository
- `--filter`: Repository name filter
- `--output` (required): Output file or directory
- `--format` (default: yaml): Output format (yaml, json, toml)
- `--include-secrets` (default: false): Include secret configurations

### Import Configuration (`gz repo-config import`)

**Purpose**: Import repository configurations from external sources

**Features**:

- Configuration file import
- Migration from other platforms
- Bulk configuration application
- Validation during import
- Conflict resolution

**Usage**:

```bash
gz repo-config import --file backup.yaml --org myorg          # Import from file
gz repo-config import --directory configs/ --org myorg        # Import from directory
gz repo-config import --file config.json --dry-run            # Preview import
```

**Parameters**:

- `--file`: Configuration file to import
- `--directory`: Directory containing configuration files
- `--org` (required): Target organization
- `--dry-run` (default: false): Preview import without applying
- `--overwrite` (default: false): Overwrite existing configurations
- `--merge` (default: true): Merge with existing configurations

### Compare Configurations (`gz repo-config diff`)

**Purpose**: Compare repository configurations between repositories or templates

**Features**:

- Repository-to-repository comparison
- Template-to-repository comparison
- Configuration drift analysis
- Detailed difference reporting
- Side-by-side comparison views

**Usage**:

```bash
gz repo-config diff --repo1 myorg/repo1 --repo2 myorg/repo2   # Compare repositories
gz repo-config diff --repo myorg/repo1 --template security    # Compare with template
gz repo-config diff --org myorg --baseline-template standard  # Compare org against template
```

**Parameters**:

- `--repo1`: First repository for comparison
- `--repo2`: Second repository for comparison
- `--repo`: Repository to compare
- `--template`: Template for comparison
- `--org`: Organization for bulk comparison
- `--baseline-template`: Baseline template for organization comparison
- `--format` (default: unified): Diff format (unified, side-by-side, json)

## Configuration Templates

### Built-in Templates

#### Security Template

- Branch protection rules
- Required status checks
- Merge restrictions
- Secret scanning enablement
- Dependency vulnerability alerts

#### Compliance Template

- Audit logging configuration
- Access control policies
- Documentation requirements
- License compliance settings

#### Development Template

- CI/CD pipeline configuration
- Code quality checks
- Automated testing setup
- Development workflow rules

### Template Structure

```yaml
name: "Security Template"
version: "1.0.0"
description: "Standard security configuration for repositories"

repository:
  # Basic repository settings
  has_issues: true
  has_projects: true
  has_wiki: false
  allow_merge_commit: true
  allow_squash_merge: true
  allow_rebase_merge: false
  delete_branch_on_merge: true

  # Security settings
  security:
    secret_scanning: true
    dependency_vulnerability_alerts: true
    automated_security_fixes: true

branch_protection:
  main:
    required_status_checks:
      strict: true
      contexts:
        - "ci/tests"
        - "ci/security-scan"
    enforce_admins: true
    required_pull_request_reviews:
      required_approving_review_count: 2
      dismiss_stale_reviews: true
      require_code_owner_reviews: true
    restrictions:
      users: []
      teams: ["security-team"]

webhooks:
  - url: "https://api.example.com/webhook"
    content_type: "json"
    events:
      - "push"
      - "pull_request"
    active: true
```

### Custom Templates

Users can create custom templates by:

1. Defining configuration in YAML format
1. Validating template syntax
1. Testing template application
1. Importing into template library

## Validation Rules

### Security Validation

- Branch protection enforcement
- Secret scanning enablement
- Dependency vulnerability alerts
- Access control verification

### Compliance Validation

- Required file presence (LICENSE, CONTRIBUTING.md)
- Documentation standards
- Audit trail requirements
- Policy adherence

### Best Practice Validation

- CI/CD pipeline presence
- Code quality checks
- Testing requirements
- Documentation coverage

## Integration

### Git Platform Integration

Supports multiple Git platforms:

- GitHub (primary)
- GitLab (via adapters)
- Gitea (via adapters)
- Bitbucket (planned)

### Authentication

```bash
# GitHub
export GITHUB_TOKEN="ghp_your_token_here"

# GitLab
export GITLAB_TOKEN="glpat_your_token_here"

# Gitea
export GITEA_TOKEN="your_token_here"
export GITEA_URL="https://gitea.example.com"
```

### CI/CD Integration

Repository configuration can be automated in CI/CD pipelines:

```yaml
# GitHub Actions example
name: Repository Configuration
on:
  schedule:
    - cron: '0 2 * * 1'  # Weekly on Monday

jobs:
  config-sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Apply repository configuration
        run: |
          gz repo-config apply --org ${{ github.repository_owner }} --template security
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Examples

### Organization-wide Configuration

```bash
# Apply security template to all repositories
gz repo-config apply --org myorg --template security --dry-run
gz repo-config apply --org myorg --template security

# Audit compliance across organization
gz repo-config audit --org myorg --template compliance --format json --output compliance-report.json

# Export all configurations for backup
gz repo-config export --org myorg --output backup/$(date +%Y-%m-%d)-configs.yaml
```

### Selective Repository Management

```bash
# Apply configuration to API repositories only
gz repo-config apply --org myorg --filter "^api-.*" --template microservice

# Compare production repositories with standard template
gz repo-config diff --org myorg --filter "^prod-.*" --baseline-template production

# Validate specific repositories
gz repo-config validate --org myorg --filter "^critical-.*" --rules strict
```

### Template Management Workflow

```bash
# Create custom template from existing
gz repo-config template create my-custom --from security
gz repo-config template show my-custom > my-custom.yaml

# Edit template file...
# vim my-custom.yaml

# Import modified template
gz repo-config template import --file my-custom.yaml

# Validate and test template
gz repo-config template validate my-custom
gz repo-config apply --template my-custom --repo myorg/test-repo --dry-run
```

### Webhook Management

```bash
# Set up organization-wide webhooks
gz repo-config webhook create --org myorg \
  --url "https://api.example.com/webhook" \
  --events push,pull_request,issues \
  --content-type json

# Update all webhooks with new events
gz repo-config webhook update --org myorg \
  --url "https://api.example.com/webhook" \
  --events push,pull_request,issues,release

# Test all webhooks
gz repo-config webhook test --org myorg
```

## Error Handling

### Common Errors

- **Authentication failures**: Invalid or expired tokens
- **Permission errors**: Insufficient repository access
- **Rate limiting**: API rate limit exceeded
- **Configuration conflicts**: Template validation failures
- **Network errors**: API connectivity issues

### Recovery Strategies

- **Rate limiting**: Automatic retry with exponential backoff
- **Permission errors**: Detailed permission requirement reporting
- **Configuration errors**: Validation with suggested fixes
- **Rollback support**: Automatic configuration restore on failure

## Security Considerations

### Access Control

Repository configuration operations require appropriate permissions:

- Repository administration rights
- Organization membership (for org-wide operations)
- Secret management permissions (for webhook secrets)

### Audit Trail

All configuration changes are logged including:

- User identification
- Timestamp of changes
- Configuration differences
- Rollback information

### Secret Management

- Webhook secrets are encrypted in transit and at rest
- Token-based authentication with scope validation
- Support for secret rotation and management

## Best Practices

### Template Design

- Use version control for templates
- Test templates on non-production repositories
- Implement gradual rollout strategies
- Document template purpose and usage

### Configuration Management

- Regular audit and compliance checks
- Backup configurations before major changes
- Use dry-run mode for validation
- Monitor configuration drift

### Security

- Regular review of access permissions
- Automated security configuration enforcement
- Incident response procedures for configuration breaches
- Regular token rotation and management
