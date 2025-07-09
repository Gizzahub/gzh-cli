# Repository Configuration Diff Guide

## Overview

The `gz repo-config diff` command compares the current state of your GitHub repositories with the desired configuration defined in your configuration files. This helps you identify what changes would be made before applying them.

## Usage

```bash
# Show all differences for an organization
gz repo-config diff --org myorg

# Filter by repository pattern
gz repo-config diff --org myorg --filter "^api-.*"

# Show in different formats
gz repo-config diff --org myorg --format json
gz repo-config diff --org myorg --format unified

# Include current values in the output
gz repo-config diff --org myorg --show-values
```

## Configuration

The diff command requires:

1. **GitHub Token**: Set via `GITHUB_TOKEN` environment variable or `--token` flag
2. **Configuration File**: A `repo-config.yaml` file defining your desired state

### Example Configuration

```yaml
version: "1.0.0"
organization: "myorg"

templates:
  microservice:
    description: "Microservice template"
    settings:
      has_issues: true
      has_wiki: false
      delete_branch_on_merge: true
    security:
      branch_protection:
        main:
          required_reviews: 2
          enforce_admins: true

repositories:
  patterns:
    - match: "*-service"
      template: "microservice"
```

## Output Formats

### Table Format (Default)

Shows differences in a human-readable table:

```
REPOSITORY           SETTING                        IMPACT     ACTION     TEMPLATE
────────────────────────────────────────────────────────────────────────────────
api-service          branch_protection.main.requir  🟡 Med     🔄         microservice
web-frontend         features.wiki                  🟢 Low     🔄         frontend
legacy-service       security.delete_head_branches  🔴 High    ➕         none
```

### JSON Format

Provides structured output for programmatic use:

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

### Unified Diff Format

Shows changes in a familiar diff format:

```diff
--- api-service (current)
+++ api-service (target)
@@ branch_protection.main.required_reviews @@
-branch_protection.main.required_reviews: 1
+branch_protection.main.required_reviews: 2
```

## Understanding Impact Levels

- **🔴 High**: Security-critical changes (visibility, admin enforcement, etc.)
- **🟡 Medium**: Important changes (branch protection, permissions, merge settings)
- **🟢 Low**: Minor changes (description, features like wiki/issues)

## Common Use Cases

### 1. Pre-Apply Validation

Before applying configuration changes, always run diff first:

```bash
# Check what would change
gz repo-config diff --org myorg

# If satisfied, apply the changes
gz repo-config apply --org myorg
```

### 2. Compliance Auditing

Regular compliance checks:

```bash
# Check specific repositories
gz repo-config diff --org myorg --filter "^production-.*"

# Export to JSON for reporting
gz repo-config diff --org myorg --format json > compliance-report.json
```

### 3. Template Validation

Verify template effects before deployment:

```bash
# See what a new template would change
gz repo-config diff --org myorg --filter "^api-.*" --show-values
```

## Troubleshooting

### No Configuration File Found

```
Error: configuration file not found
```

**Solution**: Create a `repo-config.yaml` file or specify path with `--config`:

```bash
gz repo-config diff --org myorg --config /path/to/config.yaml
```

### GitHub Token Issues

```
Error: GitHub token not found
```

**Solution**: Set the GitHub token:

```bash
export GITHUB_TOKEN=ghp_your_token_here
gz repo-config diff --org myorg
```

### Organization Mismatch

```
Error: organization mismatch: config file is for 'org1', but diff requested for 'org2'
```

**Solution**: Ensure the organization in your config file matches the `--org` flag.

## Best Practices

1. **Regular Diffs**: Run diffs regularly to ensure repositories stay compliant
2. **Filter Strategically**: Use filters to focus on specific repository groups
3. **Review High Impact**: Always carefully review high-impact changes
4. **Document Templates**: Keep template documentation up-to-date
5. **Version Control**: Keep your configuration files in version control

## Integration with CI/CD

The diff command can be integrated into CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Check Repository Compliance
  run: |
    gz repo-config diff --org ${{ github.repository_owner }} --format json > diff.json
    if [ $(jq '.summary.total_changes' diff.json) -gt 0 ]; then
      echo "Non-compliant repositories found!"
      exit 1
    fi
```

## See Also

- [Repository Configuration User Guide](repo-config-user-guide.md)
- [Configuration Schema Reference](repo-config-schema.yaml)
- [Template Examples](repo-config-policy-examples.md)