# Command: gz git webhook list

## Scenario: List webhooks for repositories

### Input

**Command**:

```bash
gz git webhook list --repo myorg/myrepo
```

**Prerequisites**:

- [ ] GitHub/GitLab authentication configured
- [ ] Repository read access
- [ ] Webhook read permissions

### Expected Output

**Multiple Webhooks Found**:

```text
🔗 Repository Webhooks: myorg/myrepo

📂 Repository: myorg/myrepo
   🌐 Platform: GitHub
   👤 Owner: myorg
   🔗 URL: https://github.com/myorg/myrepo.git

📋 Webhooks (3 total):

🟢 Webhook #1
   🆔 ID: 12345678
   🌐 URL: https://api.example.com/webhook
   📡 Events (4): push, pull_request, issues, release
   📅 Created: 2025-08-15T10:20:30Z
   📅 Updated: 2025-09-01T14:15:22Z
   ⚡ Status: active
   📄 Content Type: application/json
   🔒 Secret: configured
   🔐 SSL Verification: enabled
   📊 Recent Deliveries: 145 successful, 2 failed (last 30 days)

🟡 Webhook #2
   🆔 ID: 23456789
   🌐 URL: https://ci.company.com/github-hook
   📡 Events (2): push, pull_request
   📅 Created: 2025-07-20T09:45:15Z
   📅 Updated: 2025-07-20T09:45:15Z
   ⚡ Status: active
   📄 Content Type: application/x-www-form-urlencoded
   🔒 Secret: not configured
   🔐 SSL Verification: enabled
   📊 Recent Deliveries: 89 successful, 0 failed (last 30 days)
   ⚠️  Warning: No secret configured (security risk)

🔴 Webhook #3
   🆔 ID: 34567890
   🌐 URL: https://old-system.defunct.com/webhook
   📡 Events (1): push
   📅 Created: 2024-12-01T15:30:45Z
   📅 Updated: 2024-12-01T15:30:45Z
   ⚡ Status: inactive
   📄 Content Type: application/json
   🔒 Secret: configured
   🔐 SSL Verification: disabled
   📊 Recent Deliveries: 0 successful, 23 failed (last 30 days)
   ❌ Error: Endpoint unreachable for 30+ days

📊 Summary:
   Total webhooks: 3
   Active: 2, Inactive: 1
   With secrets: 2, Without secrets: 1
   Recent success rate: 94.5% (234/248 deliveries)

⚠️  Issues detected:
   • Webhook #2: Missing secret (security risk)
   • Webhook #3: Endpoint unreachable (consider removal)

💡 Manage webhooks:
   gz git webhook delete --id 34567890      # Remove broken webhook
   gz git webhook update --id 23456789      # Add security
   gz git webhook test --id 12345678        # Test delivery

stderr: (empty)
Exit Code: 0
```

**No Webhooks Found**:

```text
🔗 Repository Webhooks: myorg/empty-repo

📂 Repository: myorg/empty-repo
   🌐 Platform: GitHub
   👤 Owner: myorg

❌ No webhooks configured for this repository.

💡 Create your first webhook:
   gz git webhook create --repo myorg/empty-repo --url https://your-endpoint.com

💡 Common webhook use cases:
   • CI/CD integration: --events push,pull_request
   • Issue tracking: --events issues,issue_comment
   • Release automation: --events release,create
   • Security notifications: --events push --branch main

🚫 No webhooks to display.

stderr: (empty)
Exit Code: 1
```

**Organization-wide Listing**:

```text
# Command: gz git webhook list --org myorg

🔗 Organization Webhooks: myorg

🏢 Organization: myorg
   🌐 Platform: GitHub
   👥 Repositories: 25 total

📋 Organization-Level Webhooks (2):

🟢 Organization Webhook #1
   🆔 ID: 45678901
   🌐 URL: https://security.myorg.com/github-webhook
   📡 Events (3): push, repository, member
   🎯 Scope: organization
   📅 Created: 2025-01-15T08:00:00Z
   ⚡ Status: active
   🔒 Secret: configured
   📊 Recent Deliveries: 1,234 successful, 5 failed

🟢 Organization Webhook #2
   🆔 ID: 56789012
   🌐 URL: https://compliance.myorg.com/audit-hook
   📡 Events (5): repository, team, member, organization, meta
   🎯 Scope: organization
   📅 Created: 2025-02-01T12:30:00Z
   ⚡ Status: active
   🔒 Secret: configured
   📊 Recent Deliveries: 456 successful, 1 failed

📊 Repository Webhooks Summary:
   🔗 Total across 25 repositories: 47 webhooks
   📊 Average per repository: 1.9 webhooks
   ⚡ Active: 43, Inactive: 4
   🔒 With secrets: 41, Without secrets: 6

⚠️  Repositories with issues:
   • myorg/legacy-app: 2 inactive webhooks
   • myorg/test-repo: 1 webhook without secret
   • myorg/archive-project: 3 failed webhooks

💡 Organization management:
   gz git webhook list --org myorg --show-issues    # Show problematic webhooks
   gz git webhook cleanup --org myorg                # Remove inactive webhooks
   gz git webhook audit --org myorg                  # Security audit

stderr: (empty)
Exit Code: 0
```

**Permission Denied**:

```text
🔗 Repository Webhooks: private-org/secret-repo

📂 Repository: private-org/secret-repo

❌ Insufficient permissions to list webhooks:
   • Repository access: none (repository may not exist)
   • Required access: read permissions for webhook management
   • Authentication: token valid but insufficient scope

💡 Check permissions:
   - Verify repository exists and is accessible
   - Ensure authentication token has repo:read or admin:repo_hook scope
   - For organization repositories, confirm team membership

🚫 Cannot list webhooks for inaccessible repository.

stderr: repository access denied
Exit Code: 2
```

### Side Effects

**Files Created**:

- `~/.gzh/git/webhooks/webhook-list-cache.json` - Webhook listing cache
- `~/.gzh/git/webhook-summary-<timestamp>.json` - Summary report

**Files Modified**: None (read-only operation)
**State Changes**: Webhook cache updated with latest information

### Validation

**Automated Tests**:

```bash
# Test webhook listing (requires repository with webhooks)
result=$(gz git webhook list --repo "test-org/test-repo" 2>&1)
exit_code=$?

assert_contains "$result" "Repository Webhooks:"
# Exit code: 0 (webhooks found), 1 (no webhooks), 2 (access denied)

# Check cache file creation
assert_file_exists "$HOME/.gzh/git/webhooks/webhook-list-cache.json"
cache_content=$(cat "$HOME/.gzh/git/webhooks/webhook-list-cache.json")
assert_contains "$cache_content" '"webhooks":'
assert_contains "$cache_content" '"repository":'
```

**Manual Verification**:

1. List webhooks for repository with multiple webhooks
1. Test with repository having no webhooks
1. Verify organization-wide listing works
1. Check permission error handling
1. Validate webhook status indicators
1. Confirm delivery statistics accuracy

### Edge Cases

**Webhook States and Issues**:

- Webhooks with delivery failures
- SSL certificate issues
- Endpoint URL changes/redirects
- Webhooks with expired secrets

**Large-scale Operations**:

- Organizations with hundreds of repositories
- Repositories with many webhooks
- Pagination handling for large result sets
- Performance with slow API responses

**Platform Differences**:

- GitHub vs GitLab webhook structure
- Organization vs group level webhooks
- Platform-specific event types
- Different authentication mechanisms

**Data Consistency**:

- Recently created/deleted webhooks
- Webhook configuration caching
- API response variations
- Network timeout handling

### Performance Expectations

**Response Time**:

- Single repository: < 3 seconds
- Organization overview: < 15 seconds
- Large organizations: < 45 seconds with progress
- Cached results: < 1 second

**Resource Usage**:

- Memory: < 50MB for large result sets
- Network: Read-only API calls
- CPU: Low impact except JSON processing

**Display Limits**:

- Repository webhooks: unlimited
- Organization summary: paginated for >100 repos
- Delivery statistics: last 30 days by default
- History retention: 90 days in cache

## Notes

- Comprehensive webhook status and health monitoring
- Organization-wide webhook management and auditing
- Delivery statistics and failure analysis
- Security assessment (missing secrets, SSL issues)
- Inactive webhook detection and cleanup suggestions
- Export capabilities for webhook inventories
- Integration with webhook testing and management tools
- Historical webhook configuration tracking
