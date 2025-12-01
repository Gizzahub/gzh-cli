# Repository Management Guide

> **🔗 Powered by**: [gzh-cli-git](https://github.com/gizzahub/gzh-cli-git) (for local Git operations)
>
> - **독립 설치**: `go install github.com/gizzahub/gzh-cli-git/cmd/gzh-git@latest`
> - **상세 문서**: [gzh-cli-git README](https://github.com/gizzahub/gzh-cli-git#readme)
> - **통합 가이드**: [Subprojects Integration Guide](../integration/00-SUBPROJECTS_GUIDE.md#1-gzh-cli-git)

Comprehensive guide for managing GitHub repository configurations at scale using `gz repo-config` and `gz git config` commands.

## Table of Contents

1. [Overview](#overview)
1. [Quick Start](#quick-start)
1. [Configuration System](#configuration-system)
1. [Command Reference](#command-reference)
1. [Policy Templates](#policy-templates)
1. [Audit Reports](#audit-reports)
1. [Diff Analysis](#diff-analysis)
1. [Best Practices](#best-practices)
1. [Examples](#examples)
1. [Troubleshooting](#troubleshooting)

## Overview

The `gz repo-config` command (also available as `gz git config`) provides powerful tools for managing GitHub repository configurations at scale, including:

- **Policy Enforcement**: Apply consistent settings across all repositories
- **Compliance Auditing**: Generate audit reports for security and compliance frameworks
- **Configuration Drift Detection**: Compare current state vs desired configuration
- **Template Management**: Reusable configuration templates for different repository types
- **Exception Handling**: Document and manage policy exceptions

### Key Features

- **Multi-Organization Support**: Manage repositories across multiple GitHub organizations
- **Template Inheritance**: Build complex configurations from simple base templates
- **Policy Validation**: Enforce security, compliance, and quality policies
- **Interactive Reports**: Generate HTML reports with visualizations and filtering
- **CI/CD Integration**: Integrate with automated workflows and compliance checks

## Quick Start

### Prerequisites

1. **GitHub Personal Access Token**

   ```bash
   export GITHUB_TOKEN="your-github-token"
   ```

   Required scopes: `repo`, `admin:org`, `admin:repo_hook`

1. **Install gzh-cli**

   ```bash
   go install github.com/gizzahub/gzh-cli@latest
   ```

### 5-Minute Setup

#### 1. Basic Configuration (1 minute)

Create a simple configuration:

```yaml
# repo-config.yaml
version: "1.0.0"
organization: "your-org"

templates:
  standard:
    description: "Standard repository settings"
    settings:
      has_issues: true
      has_wiki: false
      delete_branch_on_merge: true
    security:
      vulnerability_alerts: true

repositories:
  - name: "*"
    template: "standard"
```

Apply it:

```bash
gz repo-config apply --config repo-config.yaml --dry-run
```

#### 2. Add Security (2 minutes)

Enhance with security policies:

```yaml
# repo-config.yaml (enhanced)
version: "1.0.0"
organization: "your-org"

templates:
  standard:
    description: "Standard repository settings"
    settings:
      has_issues: true
      has_wiki: false
      delete_branch_on_merge: true
    security:
      vulnerability_alerts: true
      branch_protection:
        main:
          required_reviews: 2
          enforce_admins: true

policies:
  security:
    description: "Basic security requirements"
    rules:
      branch_protection:
        type: "branch_protection"
        value: true
        enforcement: "required"
        message: "Main branch must be protected"

repositories:
  - name: "*"
    template: "standard"
    policies: ["security"]
```

#### 3. Repository Types (3 minutes)

Configure different templates for different repository types:

```yaml
templates:
  backend:
    description: "Backend service configuration"
    settings:
      private: true
      has_issues: true
    security:
      secret_scanning: true
      vulnerability_alerts: true
      branch_protection:
        main:
          required_reviews: 2
          required_status_checks:
            - "ci/build"
            - "ci/test"

  frontend:
    description: "Frontend application configuration"
    settings:
      has_issues: true
      has_pages: true
    security:
      vulnerability_alerts: true
      branch_protection:
        main:
          required_reviews: 1

patterns:
  - pattern: "*-api"
    template: "backend"
  - pattern: "*-service"
    template: "backend"
  - pattern: "*-ui"
    template: "frontend"
  - pattern: "*-web"
    template: "frontend"
```

## Configuration System

### Configuration File Structure

```yaml
version: "1.0.0"                    # Configuration schema version
organization: "your-org-name"       # GitHub organization

# Define reusable templates
templates:
  template-name:
    description: "Template description"
    base: "parent-template"         # Optional: inherit from another template
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

# Define compliance policies
policies:
  policy-name:
    description: "Policy description"
    rules:
      rule-name:
        type: "rule-type"
        value: expected-value
        enforcement: "required|recommended|optional"
        message: "Violation message"

# Apply configurations to repositories
repositories:
  - name: "repository-name"        # Specific repository
    template: "template-name"
    policies: ["policy-name"]
    exceptions:                    # Policy exceptions
      - policy: "policy-name"
        rule: "rule-name"
        reason: "Exception reason"

# Pattern-based configuration
patterns:
  - pattern: "*-service"           # Regex pattern
    template: "backend"
    policies: ["security"]
```

### Template Inheritance

Templates can inherit from other templates:

```yaml
templates:
  base-template:
    settings:
      has_issues: true
      has_wiki: true

  extended-template:
    base: "base-template"
    settings:
      has_wiki: false         # Override parent setting
      has_projects: true      # Add new setting
```

### Repository Settings Reference

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

### Security Settings Reference

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

## Command Reference

### Core Commands

#### `gz repo-config audit`

Audit repository settings against compliance frameworks:

```bash
gz repo-config audit --org <organization> [flags]
```

**Key Flags:**

- `--org` - Organization name (required)
- `--framework` - Compliance framework: SOC2, GDPR, HIPAA, PCI-DSS
- `--output` - Output format: table, json, yaml, csv, html
- `--output-file` - Save results to file
- `--severity` - Minimum severity: low, medium, high, critical

**Examples:**

```bash
# Basic audit
gz repo-config audit --org myorg

# SOC2 compliance audit
gz repo-config audit --org myorg --framework SOC2

# Generate HTML report
gz repo-config audit --org myorg --output html --output-file audit.html
```

#### `gz repo-config apply`

Apply configuration policies to repositories:

```bash
gz repo-config apply --config <config-file> [flags]
```

**Key Flags:**

- `--config` - Configuration file path (required)
- `--dry-run` - Preview changes without applying
- `--force` - Apply changes without confirmation
- `--parallel` - Number of concurrent operations (default: 5)

**Examples:**

```bash
# Preview changes
gz repo-config apply --config repo-standards.yaml --dry-run

# Apply configuration
gz repo-config apply --config security-policy.yaml

# Force application without prompts
gz repo-config apply --config policy.yaml --force
```

#### `gz repo-config diff`

Show configuration differences between current state and desired policy:

```bash
gz repo-config diff --org <organization> [flags]
```

**Key Flags:**

- `--org` - Organization name (required)
- `--repo` - Specific repository
- `--baseline` - Baseline configuration file
- `--output` - Output format: unified, side-by-side, json
- `--show-values` - Include current values in output

**Examples:**

```bash
# Show differences for organization
gz repo-config diff --org myorg

# Compare specific repository
gz repo-config diff --org myorg --repo critical-service

# Compare against baseline
gz repo-config diff --org myorg --baseline security-baseline.yaml
```

#### `gz repo-config generate`

Generate configuration templates based on existing repositories:

```bash
gz repo-config generate --org <organization> [flags]
```

**Key Flags:**

- `--org` - Organization name (required)
- `--template` - Template type: minimal, standard, enterprise
- `--output` - Output file name
- `--sample-size` - Number of repositories to sample

#### `gz repo-config validate`

Validate configuration file syntax and policies:

```bash
gz repo-config validate --config <config-file> [flags]
```

**Key Flags:**

- `--config` - Configuration file to validate (required)
- `--schema` - Schema file for validation
- `--strict` - Enable strict validation mode

### Output Formats

#### Table Format (Default)

Human-readable compliance summary with risk analysis.

#### HTML Format

Interactive HTML report with charts and visualizations:

```bash
gz repo-config audit --org myorg --format html --output audit-report.html
```

#### JSON Format

Structured output for programmatic use:

```bash
gz repo-config audit --org myorg --format json --output results.json
```

#### SARIF Format

Static Analysis Results Interchange Format for GitHub Advanced Security:

```bash
gz repo-config audit --org myorg --format sarif --output results.sarif
```

## Policy Templates

### Security Policies

#### Basic Security Policy

```yaml
policies:
  basic-security:
    description: "Basic security requirements"
    rules:
      vulnerability_alerts:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Vulnerability alerts must be enabled"

      default_branch_protection:
        type: "branch_protection"
        value: true
        enforcement: "required"
        message: "Default branch must be protected"
```

#### Enhanced Security Policy

```yaml
policies:
  enhanced-security:
    description: "Enhanced security for sensitive repositories"
    rules:
      must_be_private:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "Sensitive repositories must be private"

      secret_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Secret scanning must be enabled"

      branch_protection_reviews:
        type: "branch_protection_setting"
        value:
          required_reviews: 2
          dismiss_stale_reviews: true
        enforcement: "required"
        message: "Main branch must require 2 reviews"
```

### Compliance Policies

#### SOC2 Compliance

```yaml
policies:
  soc2-compliance:
    description: "SOC2 compliance requirements"
    rules:
      access_logging:
        type: "audit_log"
        value: true
        enforcement: "required"
        message: "Audit logging must be enabled for SOC2"

      code_review_required:
        type: "branch_protection_setting"
        value:
          required_reviews: 2
          require_code_owner_reviews: true
        enforcement: "required"
        message: "Code review is mandatory for SOC2"

      vulnerability_management:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Vulnerability scanning required for SOC2"
```

#### GDPR Compliance

```yaml
policies:
  gdpr-compliance:
    description: "GDPR data protection compliance"
    rules:
      private_repos:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "GDPR data must be in private repositories"

      access_control:
        type: "permissions"
        value:
          max_permission: "push"
          require_2fa: true
        enforcement: "required"
        message: "Strict access control for GDPR compliance"
```

### Open Source Policies

#### Basic Open Source

```yaml
policies:
  open-source-basic:
    description: "Basic open source project requirements"
    rules:
      must_be_public:
        type: "visibility"
        value: "public"
        enforcement: "required"
        message: "Open source projects must be public"

      has_license:
        type: "file_exists"
        value: "LICENSE"
        enforcement: "required"
        message: "Open source projects must have a license"

      community_features:
        type: "settings"
        value:
          has_issues: true
          has_discussions: true
        enforcement: "required"
        message: "Community features must be enabled"
```

## Audit Reports

### HTML Report Features

The enhanced HTML audit reports provide:

1. **Interactive Dashboard**

   - Real-time filtering by repository name, status, and policies
   - Searchable repository list
   - Print-friendly layout

1. **Visual Compliance Score**

   - Circular progress indicator showing overall compliance percentage
   - Color-coded based on compliance level:
     - Green (≥80%): Excellent compliance
     - Yellow (60-79%): Good compliance
     - Red (\<60%): Needs improvement

1. **Key Metrics Cards**

   - Total repositories
   - Compliant repositories
   - Non-compliant repositories
   - Total violations

1. **Policy Overview**

   - List of active policies with enforcement levels
   - Violation counts per policy
   - Visual badges for policy types

1. **Detailed Repository Table**

   - Status indicators
   - Violation details with policy and rule information
   - Applied policies per repository
   - Last checked timestamp

1. **Compliance Trend Chart**

   - 30-day trend visualization
   - Track compliance improvements over time
   - Interactive chart with Chart.js

### Report Customization

The HTML template uses:

- Bootstrap 5.3 for responsive layout
- Font Awesome 6.4 for icons
- Chart.js 4.3 for visualizations
- Custom CSS variables for theming

## Diff Analysis

### Understanding Diff Output

The `diff` command shows configuration differences using impact levels:

- **🔴 High**: Security-critical changes (visibility, admin enforcement, etc.)
- **🟡 Medium**: Important changes (branch protection, permissions, merge settings)
- **🟢 Low**: Minor changes (description, features like wiki/issues)

### Output Formats

#### Table Format (Default)

```
REPOSITORY           SETTING                        IMPACT     ACTION     TEMPLATE
────────────────────────────────────────────────────────────────────────────────
api-service          branch_protection.main.requir  🟡 Med     🔄         microservice
web-frontend         features.wiki                  🟢 Low     🔄         frontend
legacy-service       security.delete_head_branches  🔴 High    ➕         none
```

#### JSON Format

```json
{
  "differences": [
    {
      "repository": "api-service",
      "setting": "branch_protection.main.required_reviews",
      "current_value": "1",
      "target_value": "2",
      "change_type": "update",
      "impact": "medium",
      "template": "microservice",
      "compliant": false
    }
  ],
  "summary": {
    "total_changes": 4,
    "affected_repos": 3
  }
}
```

#### Unified Diff Format

```diff
--- api-service (current)
+++ api-service (target)
@@ branch_protection.main.required_reviews @@
-branch_protection.main.required_reviews: 1
+branch_protection.main.required_reviews: 2
```

## Best Practices

### 1. Start Simple, Build Complexity

Begin with basic policies and gradually add complexity:

```yaml
# Phase 1: Basic requirements
policies:
  basic:
    rules:
      has_readme:
        type: "file_exists"
        value: "README.md"
        enforcement: "required"

# Phase 2: Add security
policies:
  security:
    rules:
      vulnerability_alerts:
        type: "security_feature"
        value: true
        enforcement: "required"
```

### 2. Use Template Inheritance

Build complex templates from simple ones:

```yaml
templates:
  base:
    settings:
      has_issues: true

  secure:
    base: "base"
    settings:
      private: true

  production:
    base: "secure"
    security:
      secret_scanning: true
```

### 3. Document Exceptions

Always document why exceptions exist:

```yaml
exceptions:
  - policy: "must-be-private"
    rule: "visibility"
    reason: "Public API documentation"
    approved_by: "security-team"
    expires_at: "2024-12-31"
```

### 4. Regular Audits

Schedule regular compliance audits:

```bash
# Add to CI/CD pipeline
gz repo-config audit --config repo-config.yaml --fail-on-violations
```

### 5. Version Control Configuration

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

### 6. Use Pattern Matching

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

## Examples

### Complete Workflow

```bash
# 1. Check current state vs desired state
gz repo-config diff --org myorg --show-values

# 2. Run compliance audit
gz repo-config audit --org myorg --policy-preset enterprise --detailed

# 3. Save baseline for future comparison
gz repo-config audit --org myorg --format json --output baseline.json

# 4. Apply changes (dry run first)
gz repo-config apply --org myorg --dry-run

# 5. Apply actual changes
gz repo-config apply --org myorg

# 6. Re-audit and compare with baseline
gz repo-config audit --org myorg --baseline baseline.json

# 7. Generate report for management
gz repo-config audit --org myorg --format html --output compliance-report.html
```

### CI/CD Integration

```yaml
# .github/workflows/compliance.yml
name: Repository Compliance Check

on:
  schedule:
    - cron: "0 0 * * 1" # Weekly on Monday
  workflow_dispatch:

jobs:
  compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install gzh-cli
        run: |
          curl -L https://github.com/gizzahub/gzh-cli/releases/latest/download/gz-linux-amd64 -o gz
          chmod +x gz

      - name: Run Compliance Audit
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          ./gz repo-config audit \
            --org ${{ github.repository_owner }} \
            --policy-preset enterprise \
            --format sarif \
            --output results.sarif \
            --exit-on-fail \
            --fail-threshold 85

      - name: Upload SARIF results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: results.sarif

      - name: Generate HTML Report
        if: always()
        run: |
          ./gz repo-config audit \
            --org ${{ github.repository_owner }} \
            --format html \
            --output compliance-report.html

      - name: Upload HTML Report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: compliance-report
          path: compliance-report.html
```

### Enterprise Configuration Example

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

policies:
  security-baseline:
    description: "Security baseline for all repos"
    rules:
      vulnerability_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Vulnerability scanning is mandatory"

      code_review:
        type: "branch_protection_setting"
        value:
          required_reviews: 2
        enforcement: "required"
        message: "Code review is required for compliance"
```

## Troubleshooting

### Common Issues

#### 1. Authentication Errors

**Problem**: "Bad credentials" or "401 Unauthorized"

**Solution**:

- Check your GitHub token has required scopes: `repo`, `admin:org`, `admin:repo_hook`
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

- Use `--parallel` flag to limit concurrent requests
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

1. **Invalid template reference**

   ```
   Error: Unknown template 'invalid'
   Fix: Check template name exists in templates section
   ```

1. **Circular dependency**

   ```
   Error: Circular dependency detected: a -> b -> a
   Fix: Remove circular references in template inheritance
   ```

## Migration from Standalone Command

The `repo-config` command is now available as `gz git config`. Update your scripts:

```bash
# Old command
gz repo-config audit --org myorg

# New command (preferred)
gz git config audit --org myorg
```

## Related Commands

- [`gz git config`](../40-api-reference/git.md#config) - New unified Git command interface
- [`gz git webhook`](../40-api-reference/git.md#webhook) - Webhook management

## Support

For additional help:

1. Run `gz repo-config --help` for command options
1. Check the [examples directory](../../examples/) for more configurations
1. Open an issue on [GitHub](https://github.com/gizzahub/gzh-cli/issues)
1. Review the [API documentation](../40-api-reference/repository-configuration-api.md)
