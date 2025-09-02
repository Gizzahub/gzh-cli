# Command: gz repo-config validate

## Scenario: Validate repository configuration against templates

### Input

**Command**:
```bash
gz repo-config validate --template enterprise --repo https://github.com/myorg/myrepo.git
```

**Prerequisites**:

- [ ] GitHub/GitLab authentication configured
- [ ] Repository read access
- [ ] Configuration templates available

### Expected Output

**Fully Compliant Repository**:
```text
🔍 Validating repository configuration: enterprise

📂 Target Repository: myorg/myrepo
   🌐 Platform: GitHub
   🔗 URL: https://github.com/myorg/myrepo.git
   👤 Owner: myorg

📋 Validation Results:

✅ Repository Settings (6/6)
   ✅ Visibility: private (required: private)
   ✅ Default branch: main (required: main)
   ✅ Delete head branches: enabled (required: enabled)
   ✅ Allow merge commits: disabled (required: disabled)
   ✅ Allow squash merging: enabled (required: enabled)
   ✅ Allow rebase merging: enabled (required: enabled)

✅ Branch Protection - main (7/7)
   ✅ Require PR reviews: 2 reviewers (required: ≥2)
   ✅ Dismiss stale reviews: enabled (required: enabled)
   ✅ Require status checks: enabled (required: enabled)
   ✅ Require up-to-date branches: enabled (required: enabled)
   ✅ Include administrators: disabled (required: disabled)
   ✅ Allow force pushes: disabled (required: disabled)
   ✅ Allow deletions: disabled (required: disabled)

✅ Required Status Checks (3/3)
   ✅ continuous-integration/github-actions (present)
   ✅ security/codeql (present)
   ✅ quality/sonarcloud (present)

✅ Security Settings (4/4)
   ✅ Vulnerability alerts: enabled (required: enabled)
   ✅ Security updates: enabled (required: enabled)
   ✅ Secret scanning: enabled (required: enabled)
   ✅ Push protection: enabled (required: enabled)

✅ Access Control (3/3)
   ✅ @myorg/developers: write access (required: write)
   ✅ @myorg/maintainers: admin access (required: admin)
   ✅ Required secrets: all present and valid

🎉 Repository fully compliant with 'enterprise' template!

📊 Compliance Score: 100% (23/23 checks passed)

stderr: (empty)
Exit Code: 0
```

**Non-Compliant Repository**:
```text
🔍 Validating repository configuration: enterprise

📂 Target Repository: myorg/legacy-repo

📋 Validation Results:

⚠️  Repository Settings (4/6)
   ✅ Visibility: private (required: private)
   ❌ Default branch: master (required: main)
   ✅ Delete head branches: enabled (required: enabled)
   ❌ Allow merge commits: enabled (required: disabled)
   ✅ Allow squash merging: enabled (required: enabled)
   ✅ Allow rebase merging: enabled (required: enabled)

❌ Branch Protection - master (2/7)
   ❌ Require PR reviews: 1 reviewer (required: ≥2)
   ✅ Dismiss stale reviews: enabled (required: enabled)
   ❌ Require status checks: disabled (required: enabled)
   ❌ Require up-to-date branches: disabled (required: enabled)
   ❌ Include administrators: enabled (required: disabled)
   ❌ Allow force pushes: enabled (required: disabled)
   ✅ Allow deletions: disabled (required: disabled)

❌ Required Status Checks (0/3)
   ❌ continuous-integration/github-actions (missing)
   ❌ security/codeql (missing)
   ❌ quality/sonarcloud (missing)

⚠️  Security Settings (2/4)
   ✅ Vulnerability alerts: enabled (required: enabled)
   ❌ Security updates: disabled (required: enabled)
   ✅ Secret scanning: enabled (required: enabled)
   ❌ Push protection: disabled (required: enabled)

❌ Access Control (1/3)
   ❌ @myorg/developers: missing (required: write)
   ✅ @myorg/maintainers: admin access (required: admin)
   ❌ Required secrets: 2/5 missing

❌ Repository NOT compliant with 'enterprise' template!

📊 Compliance Score: 43% (10/23 checks passed)

⚠️  Critical Issues (5):
   • Branch protection insufficient on default branch
   • Required status checks missing
   • Security updates disabled
   • Push protection disabled
   • Team access not configured

💡 Apply template to fix issues:
   gz repo-config apply --template enterprise --repo myorg/legacy-repo

🔧 Manual fixes needed:
   • Rename default branch: master → main
   • Configure missing CI/CD integrations
   • Add missing repository secrets

stderr: repository not compliant
Exit Code: 1
```

**Template Not Found**:
```text
🔍 Validating repository configuration: nonexistent

❌ Configuration template 'nonexistent' not found!

📋 Available templates:
   • basic - Basic repository settings (15 checks)
   • opensource - Open source project configuration (12 checks)
   • enterprise - Enterprise security and compliance (23 checks)
   • minimal - Minimal required settings (8 checks)

💡 List template details: gz repo-config templates show <template-name>
💡 Create custom template: gz repo-config templates create nonexistent

🚫 Configuration validation failed.

stderr: template not found
Exit Code: 1
```

**Repository Access Error**:
```text
🔍 Validating repository configuration: enterprise

📂 Target Repository: private-org/secret-repo

❌ Repository access denied:
   • Repository: private-org/secret-repo not found or inaccessible
   • Authentication: token valid but insufficient permissions
   • Required access: read permissions for repository metadata

💡 Check repository URL and permissions:
   - Verify repository exists and is accessible
   - Ensure authentication token has repo:read scope
   - For private repositories, confirm team membership

🚫 Cannot validate inaccessible repository.

stderr: repository access denied
Exit Code: 2
```

### Side Effects

**Files Created**:
- `~/.gzh/repo-config/validation-<repo>-<timestamp>.json` - Validation report
- `~/.gzh/repo-config/compliance-summary.json` - Compliance summary cache

**Files Modified**: None (read-only validation)
**State Changes**: Compliance cache updated with latest validation results

### Validation

**Automated Tests**:
```bash
# Test validation with compliant repository
result=$(gz repo-config validate --template test-config --repo "test-org/compliant-repo" 2>&1)
exit_code=$?

assert_contains "$result" "Validating repository configuration"
assert_contains "$result" "Compliance Score:"
# Exit code: 0 (compliant), 1 (non-compliant), 2 (access error)

# Check validation report creation
assert_file_exists "$HOME/.gzh/repo-config/validation-compliant-repo-*.json"
report_content=$(cat "$HOME/.gzh/repo-config/validation-"*".json" | head -1)
assert_contains "$report_content" '"compliance_score":'
assert_contains "$report_content" '"checks":'
```

**Manual Verification**:
1. Validate against known compliant repository
2. Test with non-compliant repository
3. Verify compliance score calculations
4. Check detailed violation reports
5. Test with different template types
6. Confirm access error handling

### Edge Cases

**Repository States**:
- Archived repositories (limited configuration access)
- Empty repositories (missing default branch)
- Forked repositories (restricted permissions)
- Template repositories (special configurations)

**Template Variations**:
- Custom organization templates
- Template inheritance and overrides
- Conditional rules based on repository type
- Language-specific requirements

**API Limitations**:
- Rate limiting during bulk validations
- Incomplete API responses
- Deprecated API endpoints
- Platform-specific feature availability

**Complex Configurations**:
- Multiple branch protection rules
- Nested team permissions
- Environment-specific secrets
- Custom webhook configurations

### Performance Expectations

**Response Time**:
- Single repository: < 15 seconds
- Batch validation: < 2 minutes per 10 repositories
- Complex templates: < 30 seconds
- Cached results: < 2 seconds

**Resource Usage**:
- Memory: < 50MB
- Network: Read-only API calls
- CPU: Low impact except JSON processing

**Validation Coverage**:
- Repository settings: 15-25 checks
- Branch protection: 5-10 rules
- Security features: 3-8 settings
- Access control: 2-5 permissions
- Integration points: 1-10 services

## Notes

- Comprehensive compliance reporting
- Customizable validation rules per template
- Batch validation for multiple repositories
- Export capabilities (JSON, CSV, PDF reports)
- Integration with compliance dashboards
- Historical compliance tracking
- Automated compliance monitoring (scheduled)
- Template versioning and migration support
- Organization-wide compliance overview
