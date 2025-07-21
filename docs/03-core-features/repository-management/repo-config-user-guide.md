# Repository Configuration Management User Guide

This guide explains how to use the `gz repo-config` command to manage GitHub repository configurations at scale.

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Configuration File Structure](#configuration-file-structure)
4. [Templates](#templates)
5. [Policies](#policies)
6. [Commands](#commands)
7. [Examples](#examples)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Overview

The `gz repo-config` feature allows you to:

- Define repository configuration templates
- Apply configurations to multiple repositories at once
- Enforce security and compliance policies
- Audit repository settings across your organization
- Manage exceptions for specific repositories

## Getting Started

### Prerequisites

1. **GitHub Personal Access Token**
   ```bash
   export GITHUB_TOKEN="your-github-token"
   ```
   Required scopes:
   - `repo` - Full control of private repositories
   - `admin:org` - Read and write org and team membership
   - `admin:repo_hook` - Full control of repository hooks

2. **Install gzh-manager**
   ```bash
   go install github.com/gizzahub/gzh-manager-go@latest
   ```

### Quick Start

1. Create a basic configuration file:
   ```bash
   gz repo-config init --org your-org-name
   ```

2. Apply configuration to all repositories:
   ```bash
   gz repo-config apply --config repo-config.yaml
   ```

3. Run a compliance audit:
   ```bash
   gz repo-config audit --config repo-config.yaml
   ```

## Configuration File Structure

### Basic Structure

```yaml
version: "1.0.0"
organization: "your-org-name"

# Define reusable templates
templates:
  standard:
    description: "Standard repository settings"
    settings:
      private: false
      has_issues: true
      has_wiki: true

# Define compliance policies
policies:
  security:
    description: "Security requirements"
    rules:
      vulnerability_alerts:
        type: "security_feature"
        value: true
        enforcement: "required"

# Apply configurations to repositories
repositories:
  - name: "*"
    template: "standard"
    policies: ["security"]
```

### Configuration Sections

#### Version
- **Required**: Yes
- **Type**: String
- **Description**: Configuration schema version
- **Current**: "1.0.0"

#### Organization
- **Required**: Yes
- **Type**: String
- **Description**: GitHub organization name

#### Templates
- **Required**: No
- **Type**: Map of template definitions
- **Description**: Reusable configuration templates

#### Policies
- **Required**: No
- **Type**: Map of policy definitions
- **Description**: Compliance and security policies

#### Repositories
- **Required**: No
- **Type**: List of repository configurations
- **Description**: Repository-specific settings

#### Patterns
- **Required**: No
- **Type**: List of pattern-based configurations
- **Description**: Apply settings based on repository name patterns

## Templates

Templates define reusable repository configurations that can be applied to multiple repositories.

### Template Structure

```yaml
templates:
  template-name:
    description: "Template description"
    base: "parent-template"  # Optional: inherit from another template
    settings:
      # Repository settings
    security:
      # Security settings
    permissions:
      # Access permissions
    webhooks:
      # Webhook configurations
    required_files:
      # Required file definitions
```

### Template Inheritance

Templates can inherit from other templates using the `base` field:

```yaml
templates:
  base-template:
    settings:
      has_issues: true
      has_wiki: true

  extended-template:
    base: "base-template"
    settings:
      has_wiki: false  # Override parent setting
      has_projects: true  # Add new setting
```

### Settings Reference

#### Repository Settings
```yaml
settings:
  private: true/false
  has_issues: true/false
  has_wiki: true/false
  has_projects: true/false
  has_downloads: true/false
  has_pages: true/false
  default_branch: "main"
  allow_squash_merge: true/false
  allow_merge_commit: true/false
  allow_rebase_merge: true/false
  delete_branch_on_merge: true/false
  allow_auto_merge: true/false
  allow_update_branch: true/false
  web_commit_signoff_required: true/false
```

#### Security Settings
```yaml
security:
  vulnerability_alerts: true/false
  automated_security_fixes: true/false
  secret_scanning: true/false
  secret_scanning_push_protection: true/false
  dependency_graph: true/false
  security_advisories: true/false

  branch_protection:
    branch-name:
      required_reviews: 2
      dismiss_stale_reviews: true
      require_code_owner_reviews: true
      required_status_checks:
        - "ci/build"
        - "ci/test"
      strict_status_checks: true
      enforce_admins: true
      restrict_push_access:
        - "team-name"
      allow_force_pushes: false
      allow_deletions: false
      require_conversation_resolution: true
      require_linear_history: false
      required_signatures: true
```

#### Permissions
```yaml
permissions:
  teams:
    team-name: "admin|maintain|push|triage|pull"
  users:
    username: "admin|maintain|push|triage|pull"
```

#### Webhooks
```yaml
webhooks:
  - name: "webhook-name"
    url: "https://example.com/webhook"
    events:
      - push
      - pull_request
      - issues
    active: true
    secret: "${WEBHOOK_SECRET}"  # Use environment variable
```

#### Required Files
```yaml
required_files:
  - path: "README.md"
    content: |
      # Project Name
      Description here
  - path: "LICENSE"
    source: "templates/MIT-LICENSE.txt"
  - path: ".github/CODEOWNERS"
    content: |
      * @your-org/team-name
```

## Policies

Policies define rules that repositories must follow for compliance.

### Policy Structure

```yaml
policies:
  policy-name:
    description: "Policy description"
    rules:
      rule-name:
        type: "rule-type"
        value: expected-value
        enforcement: "required|recommended|optional"
        message: "Violation message"
```

### Rule Types

#### Visibility Rules
```yaml
rules:
  must_be_private:
    type: "visibility"
    value: "private"
    enforcement: "required"
    message: "Repository must be private"
```

#### Security Feature Rules
```yaml
rules:
  vulnerability_alerts:
    type: "security_feature"
    value: true
    enforcement: "required"
    message: "Vulnerability alerts must be enabled"
```

#### Branch Protection Rules
```yaml
rules:
  main_branch_protected:
    type: "branch_protection"
    value: true
    enforcement: "required"
    message: "Main branch must be protected"
```

#### File Existence Rules
```yaml
rules:
  has_readme:
    type: "file_exists"
    value: "README.md"
    enforcement: "required"
    message: "Repository must have a README.md"
```

#### Webhook Rules
```yaml
rules:
  has_ci_webhook:
    type: "webhook_exists"
    value: "ci-webhook"
    enforcement: "recommended"
    message: "CI webhook is recommended"
```

### Policy Exceptions

You can define exceptions for specific repositories:

```yaml
repositories:
  - name: "special-repo"
    template: "standard"
    exceptions:
      - policy: "security-policy"
        rule: "must_be_private"
        reason: "Public documentation repository"
        approved_by: "security-team"
        expires_at: "2024-12-31"
```

## Commands

### Initialize Configuration

Create a new configuration file:
```bash
gz repo-config init --org your-org-name [--output config.yaml]
```

### Validate Configuration

Check if your configuration file is valid:
```bash
gz repo-config validate --config repo-config.yaml
```

### Apply Configuration

Apply configuration to repositories:
```bash
# Dry run (preview changes)
gz repo-config apply --config repo-config.yaml --dry-run

# Apply changes
gz repo-config apply --config repo-config.yaml

# Apply to specific repositories
gz repo-config apply --config repo-config.yaml --repos repo1,repo2

# Force apply without confirmation
gz repo-config apply --config repo-config.yaml --force
```

### Audit Compliance

Run compliance audit:
```bash
# Basic audit
gz repo-config audit --config repo-config.yaml

# Output to file
gz repo-config audit --config repo-config.yaml --output audit-report.json

# Different output formats
gz repo-config audit --config repo-config.yaml --format html
gz repo-config audit --config repo-config.yaml --format markdown
```

### Show Differences

Compare current state with desired state:
```bash
gz repo-config diff --config repo-config.yaml --repo repo-name
```

### List Templates

Show available templates:
```bash
gz repo-config templates --config repo-config.yaml
```

### Generate Report

Generate detailed configuration report:
```bash
gz repo-config report --config repo-config.yaml --output report.html
```

## Examples

### Example 1: Basic Configuration

```yaml
version: "1.0.0"
organization: "my-company"

templates:
  default:
    description: "Default settings for all repositories"
    settings:
      has_issues: true
      has_wiki: false
      delete_branch_on_merge: true

repositories:
  - name: "*"
    template: "default"
```

### Example 2: Security-Focused Configuration

```yaml
version: "1.0.0"
organization: "my-company"

templates:
  secure:
    description: "Security-enhanced configuration"
    settings:
      private: true
      web_commit_signoff_required: true
    security:
      vulnerability_alerts: true
      secret_scanning: true
      branch_protection:
        main:
          required_reviews: 2
          enforce_admins: true
          require_code_owner_reviews: true

policies:
  security-baseline:
    description: "Minimum security requirements"
    rules:
      private_repos:
        type: "visibility"
        value: "private"
        enforcement: "required"
      vulnerability_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"

repositories:
  - name: "*-production"
    template: "secure"
    policies: ["security-baseline"]
```

### Example 3: Open Source Project

```yaml
version: "1.0.0"
organization: "my-oss-org"

templates:
  opensource:
    description: "Open source project template"
    settings:
      private: false
      has_issues: true
      has_wiki: true
      has_projects: true
    security:
      vulnerability_alerts: true
    required_files:
      - path: "LICENSE"
        content: |
          MIT License
          Copyright (c) 2024 My Organization
      - path: "CONTRIBUTING.md"
        source: "templates/CONTRIBUTING.md"
      - path: "CODE_OF_CONDUCT.md"
        source: "templates/CODE_OF_CONDUCT.md"

repositories:
  - name: "*"
    template: "opensource"
```

### Example 4: Enterprise with Multiple Templates

```yaml
version: "1.0.0"
organization: "enterprise-org"

templates:
  base:
    description: "Base template for all repos"
    settings:
      has_issues: true
      delete_branch_on_merge: true

  backend:
    base: "base"
    description: "Backend service template"
    settings:
      private: true
    security:
      secret_scanning: true

  frontend:
    base: "base"
    description: "Frontend application template"
    settings:
      has_pages: true
    webhooks:
      - name: "deploy-preview"
        url: "https://deploy.example.com/preview"
        events: ["pull_request"]

patterns:
  - pattern: "*-api"
    template: "backend"
  - pattern: "*-ui"
    template: "frontend"
```

## Best Practices

### 1. Use Templates for Consistency

Create templates for different types of repositories:
- `backend-service`
- `frontend-app`
- `documentation`
- `infrastructure`

### 2. Implement Progressive Security

Start with baseline security and add stricter rules:
```yaml
templates:
  security-baseline:
    security:
      vulnerability_alerts: true

  security-enhanced:
    base: "security-baseline"
    security:
      secret_scanning: true
      branch_protection:
        main:
          required_reviews: 2
```

### 3. Use Pattern Matching

Apply configurations based on naming conventions:
```yaml
patterns:
  - pattern: "*-prod"
    template: "production"
  - pattern: "*-dev"
    template: "development"
  - pattern: "test-*"
    template: "testing"
```

### 4. Document Exceptions

Always document why exceptions are needed:
```yaml
exceptions:
  - policy: "must-be-private"
    rule: "visibility"
    reason: "Public documentation site"
    approved_by: "john.doe@company.com"
    expires_at: "2024-12-31"
```

### 5. Regular Audits

Schedule regular compliance audits:
```bash
# Add to CI/CD pipeline
gz repo-config audit --config repo-config.yaml --fail-on-violations
```

### 6. Version Control

Keep your configuration files in version control:
```
repo-config/
├── repo-config.yaml
├── templates/
│   ├── CONTRIBUTING.md
│   ├── CODE_OF_CONDUCT.md
│   └── LICENSE
└── policies/
    ├── security.yaml
    └── compliance.yaml
```

## Troubleshooting

### Common Issues

#### 1. Authentication Errors

**Problem**: "Bad credentials" or "401 Unauthorized"

**Solution**:
- Check your GitHub token has required scopes
- Ensure token hasn't expired
- Verify `GITHUB_TOKEN` environment variable is set

#### 2. Permission Errors

**Problem**: "403 Forbidden" when updating repositories

**Solution**:
- Ensure you have admin access to the repository
- Check organization settings allow token access
- Verify token has `admin:org` scope

#### 3. Rate Limiting

**Problem**: "API rate limit exceeded"

**Solution**:
- Use `--concurrency` flag to limit parallel requests
- Add delays between operations with `--delay`
- Consider using GitHub App for higher limits

#### 4. Template Not Found

**Problem**: "Template 'x' not found"

**Solution**:
- Check template name spelling
- Ensure template is defined before use
- Verify no circular dependencies

#### 5. Policy Violations

**Problem**: Repositories failing compliance checks

**Solution**:
- Review audit report for specific violations
- Add exceptions for legitimate cases
- Update repository settings to comply

### Debug Mode

Enable debug logging for troubleshooting:
```bash
gz repo-config apply --config repo-config.yaml --debug
```

### Validation Errors

Common validation errors and fixes:

1. **Missing required field**
   ```
   Error: organization is required
   Fix: Add 'organization: "your-org"' to config
   ```

2. **Invalid template reference**
   ```
   Error: Unknown template 'invalid'
   Fix: Check template name exists in templates section
   ```

3. **Circular dependency**
   ```
   Error: Circular dependency detected: a -> b -> a
   Fix: Remove circular references in template inheritance
   ```

## Support

For additional help:

1. Run `gz repo-config --help` for command options
2. Check the [examples directory](../samples/) for more configurations
3. Open an issue on [GitHub](https://github.com/gizzahub/gzh-manager-go/issues)
4. Review the [API documentation](./repository-configuration-api.md)
