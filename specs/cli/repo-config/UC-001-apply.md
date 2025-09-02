# Command: gz repo-config apply

## Scenario: Apply repository configuration templates

### Input

**Command**:
```bash
gz repo-config apply --template enterprise --repo https://github.com/myorg/myrepo.git
```

**Prerequisites**:

- [ ] GitHub/GitLab authentication configured
- [ ] Repository write access
- [ ] Configuration templates available

### Expected Output

**Success Case**:
```text
🔧 Applying repository configuration: enterprise

📂 Target Repository: myorg/myrepo
   🌐 Platform: GitHub
   🔗 URL: https://github.com/myorg/myrepo.git
   👤 Owner: myorg

📋 Configuration Changes:

✅ Repository Settings:
   • Visibility: public → private
   • Default branch: main (unchanged)
   • Delete head branches: disabled → enabled
   • Allow merge commits: enabled → disabled
   • Allow squash merging: disabled → enabled
   • Allow rebase merging: disabled → enabled

✅ Branch Protection (main):
   • Require PR reviews: enabled (2 reviewers)
   • Dismiss stale reviews: enabled
   • Require status checks: enabled
   • Require up-to-date branches: enabled
   • Include administrators: disabled
   • Allow force pushes: disabled
   • Allow deletions: disabled

✅ Required Status Checks:
   • continuous-integration/github-actions
   • security/codeql
   • quality/sonarcloud

✅ Security Settings:
   • Vulnerability alerts: enabled
   • Security updates: enabled
   • Secret scanning: enabled
   • Push protection: enabled

✅ Collaborators & Teams:
   • Added: @myorg/developers (write)
   • Added: @myorg/maintainers (admin)
   • Updated: individual-user (read → write)

✅ Repository Secrets:
   • DOCKER_REGISTRY_TOKEN: ●●●●●●●● (updated)
   • SONAR_TOKEN: ●●●●●●●● (created)
   • DEPLOY_KEY: ●●●●●●●● (unchanged)

✅ GitHub Actions:
   • Workflow permissions: read → restricted
   • Allow actions: selected → all
   • Default workflow token: read → write

🎉 Repository configuration applied successfully!

📊 Summary:
   • Repository settings: 6 changes
   • Branch protection: 7 rules applied
   • Status checks: 3 checks required
   • Security features: 4 enabled
   • Access permissions: 3 updates
   • Secrets: 2 updated, 1 created

⏰ Changes effective immediately.

stderr: (empty)
Exit Code: 0
```

**Template Not Found**:
```text
🔧 Applying repository configuration: enterprise

❌ Configuration template 'enterprise' not found!

📋 Available templates:
   • basic - Basic repository settings
   • opensource - Open source project configuration
   • enterprise - Enterprise security and compliance
   • minimal - Minimal required settings

💡 List all templates: gz repo-config templates list
💡 Create custom template: gz repo-config templates create enterprise

🚫 Configuration apply failed.

stderr: template not found
Exit Code: 1
```

**Permission Denied**:
```text
🔧 Applying repository configuration: enterprise

📂 Target Repository: myorg/myrepo

❌ Insufficient permissions:
   • Repository access: admin required (current: write)
   • Organization settings: owner required for secrets
   • Branch protection: write required (not available)

💡 Required permissions:
   - Repository: admin access to modify settings
   - Organization: owner/admin for organization secrets
   - Teams: maintain permission for team assignments

⚠️  Some settings may require additional permissions.
Contact repository/organization administrator.

stderr: insufficient permissions
Exit Code: 2
```

**Partial Application**:
```text
🔧 Applying repository configuration: enterprise

📂 Target Repository: myorg/myrepo

📋 Configuration Results:

✅ Repository Settings: applied (6/6)
✅ Branch Protection: applied (7/7)
✅ Security Settings: applied (4/4)
⚠️  Status Checks: partial (2/3)
   ✅ continuous-integration/github-actions
   ✅ security/codeql
   ❌ quality/sonarcloud - service unavailable

❌ Team Permissions: failed (0/2)
   ❌ @myorg/developers - team not found
   ❌ @myorg/maintainers - team not found

✅ Repository Secrets: applied (2/3)
   ✅ DOCKER_REGISTRY_TOKEN: updated
   ❌ SONAR_TOKEN: quota exceeded
   ✅ DEPLOY_KEY: unchanged

⚠️  Configuration partially applied!

💡 Retry failed items:
   gz repo-config apply --template enterprise --retry-failed

🔧 Manual fixes needed for team permissions and SonarCloud integration.

stderr: partial application completed
Exit Code: 1
```

### Side Effects

**Files Created**:
- `~/.gzh/repo-config/applied-<repo>-<timestamp>.json` - Application log
- `~/.gzh/repo-config/rollback-<repo>-<timestamp>.json` - Rollback data

**Files Modified**:
- Repository configuration via API calls
- Branch protection rules
- Team and collaborator permissions
- Repository secrets and environment variables

**State Changes**:
- Repository settings updated
- Security features enabled/configured
- Access control rules applied
- CI/CD configurations modified

### Validation

**Automated Tests**:
```bash
# Test configuration application (requires test repository)
result=$(gz repo-config apply --template test-config --repo "test-org/test-repo" 2>&1)
exit_code=$?

assert_contains "$result" "Applying repository configuration"
# Exit code: 0 (success), 1 (partial/failed), 2 (permission)

# Check application log creation
assert_file_exists "$HOME/.gzh/repo-config/applied-test-repo-*.json"
log_content=$(cat "$HOME/.gzh/repo-config/applied-"*".json" | head -1)
assert_contains "$log_content" '"template":'
assert_contains "$log_content" '"repository":'
```

**Manual Verification**:
1. Apply configuration to test repository
2. Verify repository settings match template
3. Check branch protection rules are active
4. Confirm security features are enabled
5. Test access permissions work correctly
6. Validate CI/CD integrations function properly

### Edge Cases

**Template Conflicts**:
- Existing repository settings conflict with template
- Custom branch protection rules vs template rules
- Organization-level policies override template settings
- Legacy webhook configurations interfering

**API Rate Limiting**:
- GitHub/GitLab API rate limits during bulk operations
- Retry mechanisms with exponential backoff
- Progress indication for long-running operations
- Graceful degradation when rate limited

**Repository States**:
- Archived repositories (read-only)
- Private repositories with restricted access
- Forked repositories with limited permissions
- Template repositories with special settings

**Network and Service Issues**:
- GitHub/GitLab service outages
- Third-party service integrations (SonarCloud, etc.)
- Webhook endpoint validation failures
- Secret/token validation errors

### Performance Expectations

**Response Time**:
- Simple templates: < 30 seconds
- Complex templates: < 2 minutes
- Organization-wide: < 5 minutes with progress
- Retry operations: < 1 minute

**Resource Usage**:
- Memory: < 100MB
- Network: API calls (varies by template complexity)
- CPU: Low impact except during JSON processing

**API Efficiency**:
- Batch operations where possible
- Minimize redundant API calls
- Cache repository state for comparisons
- Intelligent diff-based updates

## Notes

- Template-driven configuration management
- Rollback capability for failed applications
- Dry-run mode for preview before applying
- Support for custom organization templates
- Integration with compliance and security policies
- Audit logging for all configuration changes
- Multi-repository batch operations support
- Custom validation rules per template type
