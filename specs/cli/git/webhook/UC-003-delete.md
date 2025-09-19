# Command: gz git webhook delete

## Scenario: Delete repository webhooks

### Input

**Command**:

```bash
gz git webhook delete --id 12345678 --repo myorg/myrepo
```

**Prerequisites**:

- [ ] GitHub/GitLab authentication configured
- [ ] Repository admin access
- [ ] Webhook exists and is accessible

### Expected Output

**Success Case**:

```text
🗑️  Deleting webhook from repository: myorg/myrepo

📂 Repository: myorg/myrepo
   🌐 Platform: GitHub
   👤 Owner: myorg

🔍 Locating webhook...

📋 Webhook Details:
   🆔 ID: 12345678
   🌐 URL: https://api.example.com/webhook
   📡 Events (4): push, pull_request, issues, release
   📅 Created: 2025-08-15T10:20:30Z
   ⚡ Status: active
   📊 Recent Activity: 145 deliveries (last 30 days)

⚠️  WARNING: This action cannot be undone!

📋 Impact Assessment:
   • CI/CD pipelines may stop receiving notifications
   • Issue tracking integrations will be disabled
   • Release automation may be affected
   • 145 successful deliveries in last 30 days (active webhook)

Confirm webhook deletion [y/N]: y

🗑️  Deleting webhook...

✅ Webhook deleted successfully!

📊 Deletion Summary:
   🆔 Webhook ID: 12345678 (deleted)
   📅 Deleted at: 2025-09-02T15:45:30Z
   👤 Deleted by: github-user
   🔗 Endpoint: https://api.example.com/webhook (no longer receiving events)

💡 Recovery options:
   • Recreate webhook: gz git webhook create --repo myorg/myrepo --url https://api.example.com/webhook
   • View deletion audit: gz git webhook audit --repo myorg/myrepo

🎉 Webhook removal completed.

stderr: (empty)
Exit Code: 0
```

**Webhook Not Found**:

```text
🗑️  Deleting webhook from repository: myorg/myrepo

📂 Repository: myorg/myrepo

🔍 Locating webhook...

❌ Webhook not found!
   🆔 Requested ID: 99999999
   📂 Repository: myorg/myrepo

📋 Available webhooks in repository:
   🆔 12345678: https://api.example.com/webhook (active)
   🆔 23456789: https://ci.company.com/github-hook (active)
   🆔 34567890: https://old-system.defunct.com/webhook (inactive)

💡 List all webhooks: gz git webhook list --repo myorg/myrepo
💡 Delete by URL pattern: gz git webhook delete --url "*.defunct.com" --repo myorg/myrepo

🚫 Webhook deletion failed - webhook not found.

stderr: webhook not found
Exit Code: 1
```

**Permission Denied**:

```text
🗑️  Deleting webhook from repository: myorg/myrepo

📂 Repository: myorg/myrepo

❌ Insufficient permissions to delete webhooks:
   • Current access: write (required: admin)
   • Webhook management requires repository admin permissions
   • Repository: myorg/myrepo

💡 Required access:
   - Repository admin permissions
   - Organization webhook management (if applicable)

⚠️  Contact repository administrator to:
   1. Grant admin access to your account
   2. Delete webhook manually through GitHub/GitLab UI
   3. Use organization-level webhook management

🚫 Webhook deletion failed due to insufficient permissions.

stderr: insufficient permissions
Exit Code: 2
```

**Interactive Deletion Cancelled**:

```text
🗑️  Deleting webhook from repository: myorg/myrepo

📋 Webhook Details:
   🆔 ID: 12345678
   🌐 URL: https://api.example.com/webhook
   📡 Events (4): push, pull_request, issues, release
   ⚡ Status: active
   📊 Recent Activity: 145 deliveries (last 30 days)

⚠️  WARNING: This action cannot be undone!

📋 Impact Assessment:
   • Active webhook with recent deliveries
   • May break CI/CD integrations
   • Consider testing endpoint before deletion

Confirm webhook deletion [y/N]: n

🚫 Webhook deletion cancelled by user.

💡 Alternative actions:
   • Disable temporarily: gz git webhook update --id 12345678 --disable
   • Test endpoint health: gz git webhook test --id 12345678
   • View delivery history: gz git webhook deliveries --id 12345678

stderr: (empty)
Exit Code: 1
```

**Bulk Deletion**:

```text
# Command: gz git webhook delete --inactive --repo myorg/myrepo --confirm

🗑️  Bulk deleting inactive webhooks: myorg/myrepo

📂 Repository: myorg/myrepo

🔍 Scanning for inactive webhooks...

📋 Inactive Webhooks Found (2):

🔴 Webhook #1
   🆔 ID: 34567890
   🌐 URL: https://old-system.defunct.com/webhook
   📅 Last successful delivery: 2025-07-15 (48 days ago)
   📊 Recent failures: 23 consecutive failures

🔴 Webhook #2
   🆔 ID: 45678901
   🌐 URL: https://temp-service.example.com/hook
   📅 Last successful delivery: never
   📊 Recent failures: 156 consecutive failures

⚠️  Bulk deletion confirmation: --confirm flag detected

🗑️  Deleting inactive webhooks...

✅ Webhook 34567890: deleted
✅ Webhook 45678901: deleted

📊 Bulk Deletion Summary:
   🗑️  Deleted: 2 webhooks
   ⏱️  Time saved: ~15 minutes of manual cleanup
   🔗 Endpoints cleaned: 2 unreachable URLs
   📉 Failure rate improvement: 23 + 156 = 179 fewer failures

🎉 Inactive webhook cleanup completed!

💡 Remaining active webhooks: 1
   gz git webhook list --repo myorg/myrepo

stderr: (empty)
Exit Code: 0
```

### Side Effects

**Files Created**:

- `~/.gzh/git/webhooks/deletion-audit-<timestamp>.json` - Deletion audit log
- `~/.gzh/git/webhook-backups/<webhook-id>.json` - Webhook configuration backup

**Files Modified**:

- Repository webhook configuration via API
- Webhook registry updated (webhook removed)

**State Changes**:

- Webhook removed from repository
- Event delivery stopped to deleted webhook endpoint
- Webhook delivery history archived

### Validation

**Automated Tests**:

```bash
# Test webhook deletion (requires test repository with admin access)
# First create a test webhook
webhook_result=$(gz git webhook create --repo "test-org/test-repo" --url "https://webhook.site/test" --events "push" 2>&1)
webhook_id=$(echo "$webhook_result" | grep "Webhook ID:" | cut -d: -f2 | tr -d ' ')

# Then delete it
result=$(gz git webhook delete --id "$webhook_id" --repo "test-org/test-repo" --confirm 2>&1)
exit_code=$?

assert_contains "$result" "Deleting webhook from repository"
# Exit code: 0 (success), 1 (not found/cancelled), 2 (permission denied)

# Check deletion audit log creation
assert_file_exists "$HOME/.gzh/git/webhooks/deletion-audit-*.json"
audit_content=$(cat "$HOME/.gzh/git/webhooks/deletion-audit-"*".json" | head -1)
assert_contains "$audit_content" '"webhook_id":'
assert_contains "$audit_content" '"deleted_at":'
```

**Manual Verification**:

1. Delete webhook and confirm removal from repository
1. Test webhook not found error handling
1. Verify permission error with read-only access
1. Check interactive confirmation workflow
1. Test bulk deletion of inactive webhooks
1. Validate audit log creation and backup

### Edge Cases

**Webhook Dependencies**:

- Webhooks with active CI/CD integrations
- Webhooks shared across multiple systems
- Critical production webhook endpoints
- Webhooks with complex event filtering

**Timing and Concurrency**:

- Webhook deletion during active delivery
- Multiple users managing webhooks simultaneously
- API rate limiting during bulk operations
- Webhook recreation immediately after deletion

**Data Integrity**:

- Delivery history preservation
- Configuration backup before deletion
- Audit trail maintenance
- Recovery information storage

**Organization Policies**:

- Organization-level webhook policies
- Required webhooks that cannot be deleted
- Approval workflows for webhook changes
- Compliance and audit requirements

### Performance Expectations

**Response Time**:

- Single webhook deletion: < 5 seconds
- Bulk deletion: < 30 seconds for 10 webhooks
- Interactive confirmation: immediate UI response
- Audit log creation: < 2 seconds

**Resource Usage**:

- Memory: < 30MB
- Network: Single API call per webhook
- CPU: Low impact deletion operations

**Safety Measures**:

- Configuration backup before deletion
- Interactive confirmation for active webhooks
- Audit logging for all deletions
- Recovery guidance provided

## Notes

- Comprehensive impact assessment before deletion
- Interactive confirmation with cancellation option
- Bulk deletion capabilities for cleanup operations
- Audit logging and configuration backup
- Recovery guidance and webhook recreation support
- Integration with webhook health monitoring
- Organization-level webhook policy enforcement
- Safe deletion with rollback information
