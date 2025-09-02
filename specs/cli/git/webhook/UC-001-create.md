# Command: gz git webhook create

## Scenario: Create webhook for repository events

### Input

**Command**:
```bash
gz git webhook create --repo myorg/myrepo --url https://api.example.com/webhook --events push,pull_request --secret mysecret
```

**Prerequisites**:

- [ ] GitHub/GitLab authentication configured
- [ ] Repository admin access
- [ ] Webhook endpoint URL accessible

### Expected Output

**Success Case**:
```text
🔗 Creating webhook for repository: myorg/myrepo

📂 Target Repository: myorg/myrepo
   🌐 Platform: GitHub
   🔗 URL: https://github.com/myorg/myrepo.git
   👤 Owner: myorg

📋 Webhook Configuration:
   🌐 Endpoint URL: https://api.example.com/webhook
   📡 Events: push, pull_request
   🔒 Secret: ●●●●●●●● (configured)
   📄 Content Type: application/json
   🔐 SSL Verification: enabled

🔍 Validating webhook endpoint...
   ✅ URL accessible: https://api.example.com/webhook
   ✅ SSL certificate valid
   ✅ Response time: 234ms (acceptable)

🚀 Creating webhook...

✅ Webhook created successfully!

📊 Webhook Details:
   🆔 Webhook ID: 12345678
   📅 Created: 2025-09-02T15:30:45Z
   ⚡ Status: active
   🎯 Delivery URL: https://api.example.com/webhook
   📡 Events (2): push, pull_request
   🔒 Secret: configured and verified

💡 Test webhook: gz git webhook test --id 12345678
💡 View deliveries: gz git webhook deliveries --id 12345678

stderr: (empty)
Exit Code: 0
```

**Webhook Endpoint Validation Failed**:
```text
🔗 Creating webhook for repository: myorg/myrepo

📋 Webhook Configuration:
   🌐 Endpoint URL: https://invalid.example.com/webhook
   📡 Events: push, pull_request
   🔒 Secret: ●●●●●●●● (configured)

🔍 Validating webhook endpoint...
   ❌ URL inaccessible: https://invalid.example.com/webhook
   • Error: connection timeout after 10s
   • Status: DNS resolution failed
   
⚠️  SSL Certificate Issues:
   ❌ Certificate expired: 2025-08-15 (17 days ago)
   ❌ Hostname mismatch: cert for *.old-example.com
   
💡 Endpoint validation failed. Continue anyway? [y/N]: n

🚫 Webhook creation cancelled due to endpoint validation failure.

💡 Fix endpoint issues:
   - Check URL accessibility: curl -I https://invalid.example.com/webhook
   - Verify SSL certificate: openssl s_client -connect invalid.example.com:443
   - Test webhook handler: gz git webhook test --url https://invalid.example.com/webhook

stderr: endpoint validation failed
Exit Code: 1
```

**Repository Permission Error**:
```text
🔗 Creating webhook for repository: myorg/myrepo

📂 Target Repository: myorg/myrepo

❌ Insufficient repository permissions:
   • Current access: write (required: admin)
   • Webhook management requires admin permissions
   • Repository: myorg/myrepo

💡 Required permissions:
   - Repository admin access to create webhooks
   - Organization webhook permissions (if applicable)

⚠️  Contact repository administrator to:
   1. Grant admin access to your account
   2. Create webhook manually through GitHub/GitLab UI
   3. Use organization-level webhooks if available

🚫 Webhook creation failed due to insufficient permissions.

stderr: insufficient permissions
Exit Code: 2
```

**Duplicate Webhook Detection**:
```text
🔗 Creating webhook for repository: myorg/myrepo

📂 Target Repository: myorg/myrepo

⚠️  Existing webhook detected:
   🆔 Webhook ID: 87654321
   🌐 URL: https://api.example.com/webhook
   📡 Events: push, pull_request, issues
   📅 Created: 2025-08-15T10:20:30Z
   ⚡ Status: active

🤔 Webhook with similar configuration already exists.

Options:
   [1] Cancel creation (recommended)
   [2] Create duplicate webhook
   [3] Update existing webhook
   [4] Delete existing and create new

Select option [1-4]: 3

🔄 Updating existing webhook (ID: 87654321)...

📋 Configuration Changes:
   📡 Events: push, pull_request, issues → push, pull_request
   🔒 Secret: ●●●●●●●● → ●●●●●●●● (updated)
   📄 Content Type: application/json (unchanged)

✅ Webhook updated successfully!

📊 Updated Webhook Details:
   🆔 Webhook ID: 87654321
   📅 Modified: 2025-09-02T15:30:45Z
   ⚡ Status: active
   🎯 Delivery URL: https://api.example.com/webhook
   📡 Events (2): push, pull_request
   🔒 Secret: updated and verified

stderr: (empty)
Exit Code: 0
```

### Side Effects

**Files Created**:
- `~/.gzh/git/webhooks/<repo>-webhooks.json` - Repository webhook registry
- `~/.gzh/git/webhook-creation.log` - Webhook creation audit log

**Files Modified**:
- Repository webhook configuration via API
- Webhook registry updated with new webhook details

**State Changes**:
- New webhook registered with repository
- Webhook endpoint configured for event delivery
- Secret key stored securely in repository settings

### Validation

**Automated Tests**:
```bash
# Test webhook creation (requires test repository with admin access)
result=$(gz git webhook create --repo "test-org/test-repo" --url "https://webhook.site/test" --events "push" 2>&1)
exit_code=$?

assert_contains "$result" "Creating webhook for repository"
# Exit code: 0 (success), 1 (validation failed), 2 (permission denied)

# Check webhook registry creation
assert_file_exists "$HOME/.gzh/git/webhooks/test-repo-webhooks.json"
registry_content=$(cat "$HOME/.gzh/git/webhooks/test-repo-webhooks.json")
assert_contains "$registry_content" '"webhook_id":'
assert_contains "$registry_content" '"endpoint_url":'
```

**Manual Verification**:
1. Create webhook with valid endpoint
2. Verify webhook appears in repository settings
3. Test webhook receives events correctly
4. Check duplicate webhook detection
5. Validate permission error handling
6. Confirm SSL validation works

### Edge Cases

**Event Configuration**:
- Invalid event names for platform
- Platform-specific event differences (GitHub vs GitLab)
- Wildcard event subscriptions
- Event filtering based on branch/path patterns

**URL and Security**:
- Non-HTTPS URLs (insecure webhooks)
- URLs with authentication parameters
- Internal/localhost URLs
- URL redirections and proxy handling

**Repository States**:
- Archived repositories (webhook creation restricted)
- Private repositories with limited access
- Organization-owned repositories
- Fork relationships and webhook inheritance

**API and Network Issues**:
- GitHub/GitLab API rate limiting
- Network connectivity issues during creation
- Webhook endpoint temporary unavailability
- SSL/TLS handshake failures

### Performance Expectations

**Response Time**:
- Simple webhook creation: < 5 seconds
- With endpoint validation: < 15 seconds
- Batch webhook creation: < 30 seconds per webhook
- SSL certificate validation: < 3 seconds

**Resource Usage**:
- Memory: < 30MB
- Network: API calls + endpoint validation
- CPU: Low impact JSON processing

**Validation Coverage**:
- URL accessibility and response time
- SSL certificate validity and hostname matching
- Webhook endpoint capability testing
- Event type validation for platform

## Notes

- Support for GitHub, GitLab, Gitea webhook creation
- Comprehensive endpoint validation before creation
- Duplicate webhook detection and management
- Secret key generation and secure storage
- Event filtering and custom payload formats
- Webhook testing capabilities integration
- Batch webhook creation for multiple repositories
- Template-based webhook configuration
- Integration with CI/CD pipeline setup
