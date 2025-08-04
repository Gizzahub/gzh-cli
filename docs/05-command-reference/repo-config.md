# repo-config Command Reference

GitHub repository configuration management for organization-wide policy enforcement and compliance.

## Synopsis

```bash
gz repo-config <action> [flags]
gz repo-config <action> --config <config-file>
```

## Description

The `repo-config` command manages GitHub repository configurations at scale, enabling policy enforcement, compliance auditing, and standardization across organizations.

**Note:** This command is also available as `gz git config`. The `repo-config` command remains for backward compatibility.

## Actions

### `gz repo-config audit`

Audit repository settings against compliance frameworks.

```bash
gz repo-config audit --org <organization> [flags]
```

**Flags:**
- `--org` - Organization name (required)
- `--repo` - Specific repository (format: org/repo)
- `--framework` - Compliance framework: SOC2, GDPR, HIPAA, PCI-DSS
- `--output` - Output format: table, json, yaml, csv
- `--output-file` - Save results to file
- `--severity` - Minimum severity: low, medium, high, critical

**Examples:**
```bash
# Basic audit
gz repo-config audit --org myorg

# SOC2 compliance audit
gz repo-config audit --org myorg --framework SOC2

# Audit specific repository
gz repo-config audit --org myorg --repo important-service

# Export audit results
gz repo-config audit --org myorg --output json --output-file audit.json
```

### `gz repo-config apply`

Apply configuration policies to repositories.

```bash
gz repo-config apply --config <config-file> [flags]
```

**Flags:**
- `--config` - Configuration file path (required)
- `--org` - Target organization
- `--repo` - Target repository
- `--dry-run` - Preview changes without applying
- `--force` - Apply changes without confirmation
- `--parallel` - Number of concurrent operations (default: 5)

**Examples:**
```bash
# Apply configuration
gz repo-config apply --config repo-standards.yaml

# Dry run first
gz repo-config apply --config security-policy.yaml --dry-run

# Apply to specific organization
gz repo-config apply --config policy.yaml --org myorg

# Force application without prompts
gz repo-config apply --config policy.yaml --force
```

### `gz repo-config diff`

Show configuration differences between current state and desired policy.

```bash
gz repo-config diff --org <organization> [flags]
```

**Flags:**
- `--org` - Organization name (required)
- `--repo` - Specific repository
- `--baseline` - Baseline configuration file
- `--output` - Output format: unified, side-by-side, json

**Examples:**
```bash
# Show differences for organization
gz repo-config diff --org myorg

# Compare specific repository
gz repo-config diff --org myorg --repo critical-service

# Compare against baseline
gz repo-config diff --org myorg --baseline security-baseline.yaml
```

### `gz repo-config generate`

Generate configuration templates based on existing repositories.

```bash
gz repo-config generate --org <organization> [flags]
```

**Flags:**
- `--org` - Organization name (required)
- `--template` - Template type: minimal, standard, enterprise
- `--output` - Output file name
- `--sample-size` - Number of repositories to sample

**Examples:**
```bash
# Generate standard template
gz repo-config generate --org myorg --template standard

# Generate from sample repositories
gz repo-config generate --org myorg --sample-size 10

# Save to specific file
gz repo-config generate --org myorg --output custom-policy.yaml
```

### `gz repo-config validate`

Validate configuration file syntax and policies.

```bash
gz repo-config validate --config <config-file> [flags]
```

**Flags:**
- `--config` - Configuration file to validate (required)
- `--schema` - Schema file for validation
- `--strict` - Enable strict validation mode

**Examples:**
```bash
# Validate configuration
gz repo-config validate --config repo-policy.yaml

# Strict validation
gz repo-config validate --config policy.yaml --strict
```

## Configuration File Format

### Repository Policy Configuration

```yaml
version: "1.0"

# Global settings
organization: "myorg"
apply_to: "all"  # or specific repo list

# Repository settings
repository:
  # Basic settings
  description_required: true
  topics_required: true
  homepage_url_pattern: "^https://(www\\.)?mycompany\\.com"

  # Feature settings
  features:
    issues: true
    projects: false
    wiki: false
    downloads: false
    pages: false

  # Merge settings
  merge_options:
    allow_merge_commit: false
    allow_squash_merge: true
    allow_rebase_merge: true
    delete_branch_on_merge: true

# Security settings
security:
  vulnerability_alerts: true
  security_advisories: true

  # Dependency scanning
  dependency_scanning:
    enabled: true
    auto_fix: true

  # Secret scanning
  secret_scanning:
    enabled: true
    push_protection: true

  # Code scanning
  code_scanning:
    enabled: true
    default_setup: true

# Branch protection rules
branch_protection:
  patterns:
    - pattern: "main"
      protection:
        required_status_checks:
          strict: true
          contexts:
            - "ci/tests"
            - "ci/lint"
        enforce_admins: true
        required_pull_request_reviews:
          required_approving_review_count: 2
          dismiss_stale_reviews: true
          require_code_owner_reviews: true
          require_last_push_approval: true
        restrictions:
          users: []
          teams: ["admin"]

    - pattern: "release/*"
      protection:
        required_status_checks:
          strict: true
          contexts:
            - "ci/tests"
            - "ci/security-scan"
        required_pull_request_reviews:
          required_approving_review_count: 3

# Team permissions
permissions:
  teams:
    - name: "developers"
      permission: "write"
    - name: "contractors"
      permission: "read"
    - name: "admin"
      permission: "admin"

# Compliance rules
compliance:
  frameworks:
    - "SOC2"
    - "GDPR"

  required_labels:
    - "type/enhancement"
    - "type/bug"
    - "priority/high"
    - "priority/medium"
    - "priority/low"

  required_files:
    - "README.md"
    - "LICENSE"
    - "CONTRIBUTING.md"
    - ".github/CODEOWNERS"

# Webhooks
webhooks:
  - name: "ci-webhook"
    url: "https://ci.company.com/webhook"
    events: ["push", "pull_request"]
    secret: "${CI_WEBHOOK_SECRET}"

  - name: "security-webhook"
    url: "https://security.company.com/webhook"
    events: ["security_advisory"]
```

### Audit Configuration

```yaml
version: "1.0"

audit:
  frameworks: ["SOC2", "GDPR"]

  checks:
    - name: "branch-protection"
      severity: "high"
      description: "Main branch must be protected"

    - name: "required-reviews"
      severity: "medium"
      description: "Pull requests require at least 2 reviews"

    - name: "secret-scanning"
      severity: "critical"
      description: "Secret scanning must be enabled"

  reporting:
    format: "json"
    include_compliant: false
    group_by_severity: true
```

## Examples

### Organization-Wide Policy Enforcement

```bash
# 1. Audit current state
gz repo-config audit --org mycompany --output json > current-state.json

# 2. Generate baseline policy
gz repo-config generate --org mycompany --template enterprise > policy.yaml

# 3. Preview changes
gz repo-config apply --config policy.yaml --dry-run

# 4. Apply policy
gz repo-config apply --config policy.yaml --org mycompany

# 5. Verify compliance
gz repo-config audit --org mycompany --framework SOC2
```

### Security Hardening

```bash
# Create security-focused policy
cat > security-policy.yaml << EOF
version: "1.0"
security:
  vulnerability_alerts: true
  secret_scanning:
    enabled: true
    push_protection: true
  dependency_scanning:
    enabled: true
    auto_fix: true
branch_protection:
  patterns:
    - pattern: "main"
      protection:
        required_status_checks:
          strict: true
          contexts: ["security-scan"]
        required_pull_request_reviews:
          required_approving_review_count: 2
EOF

# Apply security policy
gz repo-config apply --config security-policy.yaml --org myorg
```

### Compliance Reporting

```bash
# Generate compliance report
gz repo-config audit --org myorg --framework SOC2 --output json > soc2-report.json

# Create remediation plan
gz repo-config diff --org myorg --baseline compliance-baseline.yaml > remediation-plan.txt

# Track progress
gz repo-config audit --org myorg --framework SOC2 --severity high
```

## Integration Examples

### CI/CD Pipeline

```yaml
# .github/workflows/repo-governance.yml
name: Repository Governance
on:
  schedule:
    - cron: '0 9 * * MON'  # Weekly on Monday

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install gz
        run: |
          # Install gz binary

      - name: Audit Repositories
        run: |
          gz repo-config audit --org ${{ github.repository_owner }} \
            --framework SOC2 --output json --output-file audit.json
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Upload Audit Results
        uses: actions/upload-artifact@v3
        with:
          name: audit-results
          path: audit.json
```

### Bulk Repository Setup

```bash
#!/bin/bash
# Setup script for new organization

ORG="newcompany"
POLICY="enterprise-policy.yaml"

echo "Setting up repository governance for $ORG..."

# Generate initial policy
gz repo-config generate --org "$ORG" --template enterprise > "$POLICY"

# Review and edit policy
echo "Please review $POLICY and make necessary changes"
read -p "Press enter to continue..."

# Apply policy with dry run first
echo "Previewing changes..."
gz repo-config apply --config "$POLICY" --org "$ORG" --dry-run

read -p "Apply changes? (y/N) " -n 1 -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    gz repo-config apply --config "$POLICY" --org "$ORG"
    echo "Policy applied successfully"
else
    echo "Policy application cancelled"
fi
```

## Error Handling

### Common Issues

1. **Permission Denied**
   ```bash
   # Check GitHub token permissions
   curl -H "Authorization: token $GITHUB_TOKEN" \
        https://api.github.com/user
   ```

2. **Rate Limiting**
   ```bash
   # Reduce parallelism
   gz repo-config apply --config policy.yaml --parallel 2

   # Check rate limit status
   gz repo-config status --rate-limit
   ```

3. **Configuration Conflicts**
   ```bash
   # Validate configuration
   gz repo-config validate --config policy.yaml --strict

   # Check for conflicts
   gz repo-config diff --org myorg --baseline policy.yaml
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

- [`gz git config`](git.md#config) - New unified Git command interface
- [`gz git webhook`](git.md#webhook) - Webhook management

## See Also

- [Repository Configuration Guide](../03-core-features/repository-management/repo-config-user-guide.md)
- [GitHub Repository Examples](../../examples/github/)
- [Policy Examples](../03-core-features/repository-management/repo-config-policy-examples.md)
