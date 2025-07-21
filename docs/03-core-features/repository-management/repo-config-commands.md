# Repository Configuration Management Commands

The `gz repo-config` command provides powerful tools for managing GitHub repository configurations at scale, including compliance auditing and configuration diffing.

## Table of Contents

- [Overview](#overview)
- [diff Command](#diff-command)
- [audit Command](#audit-command)
- [Configuration](#configuration)
- [Policy Templates](#policy-templates)
- [Examples](#examples)

## Overview

Repository configuration management helps organizations:
- Ensure consistent repository settings across all projects
- Identify configuration drift from desired state
- Enforce security and compliance policies
- Generate audit reports for compliance frameworks
- Track configuration changes over time

## diff Command

The `diff` command compares current repository configurations against desired state defined in configuration files.

### Usage

```bash
gz repo-config diff --org <organization> [flags]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--org` | GitHub organization name | Required |
| `--config` | Configuration file path | Auto-detected |
| `--filter` | Filter repositories by name pattern (regex) | All repos |
| `--format` | Output format (table, json, unified) | table |
| `--show-values` | Include current values in output | false |
| `--dry-run` | Preview changes without applying | false |
| `--parallel` | Number of parallel operations | 5 |
| `--timeout` | API timeout duration | 30s |
| `--token` | GitHub personal access token | $GITHUB_TOKEN |

### Output Formats

#### Table Format (Default)
```bash
gz repo-config diff --org myorg

📊 Repository Configuration Differences
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Organization: myorg
Total Repositories: 25
Different: 12 (48.0%)

REPOSITORY       TEMPLATE    DIFFERENCES   IMPACT    CHANGES
────────────────────────────────────────────────────────────────
api-server       backend     3             HIGH      🔒 Security
web-app          frontend    2             MEDIUM    ⚙️ Settings
docs-site        -           5             LOW       📋 Features
```

#### JSON Format
```bash
gz repo-config diff --org myorg --format json
```

```json
{
  "organization": "myorg",
  "timestamp": "2024-01-15T10:30:00Z",
  "summary": {
    "total_repositories": 25,
    "different_count": 12,
    "difference_percentage": 48.0
  },
  "differences": [
    {
      "repository": "api-server",
      "template": "backend",
      "changes": [
        {
          "setting": "branch_protection.main.required_reviews",
          "current": 1,
          "desired": 2,
          "impact": "high"
        }
      ]
    }
  ]
}
```

#### Unified Diff Format
```bash
gz repo-config diff --org myorg --format unified
```

Shows changes in standard unified diff format, useful for reviewing changes before applying.

### Examples

```bash
# Basic diff for organization
gz repo-config diff --org myorg

# Filter specific repositories
gz repo-config diff --org myorg --filter "^api-.*"

# Show current values alongside desired values
gz repo-config diff --org myorg --show-values

# Output as JSON for automation
gz repo-config diff --org myorg --format json > diff.json

# Use specific configuration file
gz repo-config diff --org myorg --config repo-config.yaml
```

## audit Command

The `audit` command generates comprehensive compliance audit reports for repository configurations.

### Usage

```bash
gz repo-config audit --org <organization> [flags]
```

### Flags

#### Basic Options
| Flag | Description | Default |
|------|-------------|---------|
| `--org` | GitHub organization name | Required |
| `--format` | Output format (table, json, html, csv, sarif, junit) | table |
| `--output` | Output file path | stdout |
| `--detailed` | Include detailed violation information | false |
| `--policy` | Audit specific policy only | All policies |

#### Repository Filters
| Flag | Description | Example |
|------|-------------|---------|
| `--filter-visibility` | Filter by visibility (public, private, all) | `--filter-visibility private` |
| `--filter-template` | Filter by template name | `--filter-template backend` |
| `--filter-topics` | Filter by repository topics | `--filter-topics security,compliance` |
| `--filter-team` | Filter by team ownership | `--filter-team platform` |
| `--filter-modified` | Filter by last modified time | `--filter-modified 30d` |
| `--filter-pattern` | Filter by repository name pattern (regex) | `--filter-pattern ^api-.*` |

#### Policy Options
| Flag | Description | Available Options |
|------|-------------|-------------------|
| `--policy-group` | Audit specific policy group | security, compliance, best-practice |
| `--policy-preset` | Use predefined policy preset | soc2, iso27001, nist, pci-dss, hipaa, gdpr, minimal, enterprise |

#### CI/CD Integration
| Flag | Description | Default |
|------|-------------|---------|
| `--exit-on-fail` | Exit with non-zero code if compliance fails | false |
| `--fail-threshold` | Compliance percentage threshold for failure | 80.0 |
| `--baseline` | Compare against baseline file | none |

#### Trend Analysis
| Flag | Description | Default |
|------|-------------|---------|
| `--save-trend` | Save audit results for trend analysis | false |
| `--show-trend` | Show trend analysis report | false |
| `--trend-period` | Trend analysis period | 30d |

#### Notifications
| Flag | Description | Example |
|------|-------------|---------|
| `--notify-webhook` | Send audit results to webhook URL | `--notify-webhook https://slack.com/webhook` |
| `--notify-email` | Send audit results to email address | `--notify-email admin@company.com` |

### Output Formats

#### Table Format (Default)
Shows human-readable compliance summary with risk analysis.

#### HTML Format
Generates interactive HTML report with charts and visualizations.

```bash
gz repo-config audit --org myorg --format html --output audit-report.html
```

#### SARIF Format
Static Analysis Results Interchange Format for integration with GitHub Advanced Security.

```bash
gz repo-config audit --org myorg --format sarif --output results.sarif
```

#### JUnit XML Format
For CI/CD integration and test reporting.

```bash
gz repo-config audit --org myorg --format junit --output junit.xml
```

### Policy Presets

#### SOC 2 Type II
```bash
gz repo-config audit --org myorg --policy-preset soc2
```
Enforces Service Organization Control 2 compliance requirements.

#### ISO 27001:2022
```bash
gz repo-config audit --org myorg --policy-preset iso27001
```
Information Security Management System requirements.

#### NIST Cybersecurity Framework
```bash
gz repo-config audit --org myorg --policy-preset nist
```
NIST CSF security controls.

#### PCI DSS v4.0
```bash
gz repo-config audit --org myorg --policy-preset pci-dss
```
Payment Card Industry Data Security Standard.

### Examples

```bash
# Basic audit
gz repo-config audit --org myorg

# Security policies only
gz repo-config audit --org myorg --policy-group security

# SOC2 compliance check
gz repo-config audit --org myorg --policy-preset soc2

# Filter private repos modified in last 30 days
gz repo-config audit --org myorg --filter-visibility private --filter-modified 30d

# CI pipeline with failure on low compliance
gz repo-config audit --org myorg --format junit --exit-on-fail --fail-threshold 90

# Generate SARIF report for GitHub Advanced Security
gz repo-config audit --org myorg --format sarif --output results.sarif

# Save trend data
gz repo-config audit --org myorg --save-trend

# Show 7-day trend analysis
gz repo-config audit --org myorg --show-trend --trend-period 7d

# Send notifications
gz repo-config audit --org myorg --notify-webhook https://slack.com/webhook
```

## Configuration

### Repository Configuration File

Define desired repository configurations in YAML:

```yaml
version: "1.0"
organization: "myorg"

defaults:
  template: "default"
  settings:
    private: true
    has_issues: true
    has_wiki: false
    has_projects: false
    allow_squash_merge: true
    allow_merge_commit: false
    allow_rebase_merge: true
    delete_branch_on_merge: true

templates:
  backend:
    description: "Backend service template"
    settings:
      has_wiki: false
    security:
      vulnerability_alerts: true
      security_advisories: true
      branch_protection:
        main:
          required_reviews: 2
          dismiss_stale_reviews: true
          require_code_owner_reviews: true
          enforce_admins: true

repositories:
  patterns:
    - match: "^api-.*"
      template: "backend"
    - match: "^web-.*"
      template: "frontend"
  specific:
    - name: "legacy-service"
      exceptions:
        - policy: "branch_protection"
          rule: "require_reviews"
          reason: "Legacy service with limited maintainers"
          approved_by: "security-team"
```

### Policy Configuration

Define custom policies:

```yaml
policies:
  custom_security:
    description: "Custom security requirements"
    group: "security"
    severity: "critical"
    rules:
      branch_protection_enabled:
        type: "branch_protection"
        value: true
        enforcement: "required"
        message: "Main branch must be protected"
```

## Policy Templates

### Built-in Policy Groups

1. **Security Policies** (40% weight)
   - Branch protection
   - Vulnerability management
   - Access control

2. **Compliance Policies** (35% weight)
   - Required documentation
   - Audit logging

3. **Best Practice Policies** (25% weight)
   - CI/CD pipeline
   - Code quality
   - Repository hygiene

### Risk Scoring

Repositories are assessed across multiple risk factors:

- **Security Risk**: Based on security policy violations
- **Compliance Risk**: Based on regulatory requirements
- **Operational Risk**: Based on best practices
- **Exposure Risk**: Public repositories with violations

Risk levels:
- **Critical**: Score ≥ 75
- **High**: Score ≥ 50
- **Medium**: Score ≥ 25
- **Low**: Score < 25

## Examples

### Complete Workflow

```bash
# 1. Check current state vs desired state
gz repo-config diff --org myorg --show-values

# 2. Run compliance audit
gz repo-config audit --org myorg --policy-preset enterprise --detailed

# 3. Save baseline for future comparison
gz repo-config audit --org myorg --format json --output baseline.json

# 4. Apply changes (if using apply command - future feature)
# gz repo-config apply --org myorg --dry-run

# 5. Re-audit and compare with baseline
gz repo-config audit --org myorg --baseline baseline.json

# 6. Generate report for management
gz repo-config audit --org myorg --format html --output compliance-report.html
```

### CI/CD Integration

```yaml
# .github/workflows/compliance.yml
name: Repository Compliance Check

on:
  schedule:
    - cron: '0 0 * * 1' # Weekly on Monday
  workflow_dispatch:

jobs:
  compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install gzh-manager
        run: |
          curl -L https://github.com/gizzahub/gzh-manager-go/releases/latest/download/gz-linux-amd64 -o gz
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

### Trend Analysis

```bash
# Set up daily audit job to collect trend data
0 0 * * * gz repo-config audit --org myorg --save-trend

# Weekly trend report
gz repo-config audit --org myorg --show-trend --trend-period 7d

# Monthly compliance review
gz repo-config audit --org myorg --show-trend --trend-period 30d --format html --output monthly-report.html
```
