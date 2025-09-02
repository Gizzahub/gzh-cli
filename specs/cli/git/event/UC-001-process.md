# Command: gz git event process

## Scenario: Process GitHub/GitLab events from webhooks

### Input

**Command**:
```bash
gz git event process --source webhook --file /tmp/github-payload.json
```

**Prerequisites**:

- [ ] Event payload file or webhook data available
- [ ] Event processing configuration set up
- [ ] Required integrations configured (CI/CD, notifications, etc.)

### Expected Output

**Push Event Processing**:
```text
⚡ Processing Git event from webhook

📄 Event Source: /tmp/github-payload.json
   🌐 Platform: GitHub
   📡 Event Type: push
   📅 Timestamp: 2025-09-02T15:30:45Z

📂 Repository Context:
   🔗 Repository: myorg/myrepo
   🌳 Branch: main
   👤 Pusher: developer-user
   📊 Commits: 3 new commits

📋 Event Details:
   🔀 Before: a1b2c3d4e5f6789...
   🔀 After: f6e5d4c3b2a1098...
   📝 Commits (3):
      • f6e5d4c: feat(api): add user authentication endpoint
      • c3b2a10: fix(auth): resolve JWT token validation
      • b2a1098: docs: update API documentation

🔄 Processing Actions:

✅ CI/CD Integration:
   🚀 Triggered GitHub Actions workflow: .github/workflows/ci.yml
   🏗️  Build #1234 started: https://github.com/myorg/myrepo/actions/runs/1234
   📊 Estimated completion: 8 minutes

✅ Code Quality Checks:
   🔍 SonarCloud analysis queued
   🛡️  Security scan initiated
   📏 Code coverage analysis scheduled

✅ Notifications:
   💬 Slack notification sent: #dev-team channel
   📧 Email digest queued for stakeholders
   📱 Mobile push notification: 3 commits to main branch

✅ Documentation:
   📚 API documentation build triggered
   🔄 Changelog updated automatically
   📖 Release notes draft generated

📊 Processing Summary:
   ⏱️  Processing time: 2.3 seconds
   🎯 Actions triggered: 8
   ✅ Successful: 8, ❌ Failed: 0
   🔗 Integrations: GitHub Actions, SonarCloud, Slack, Email

🎉 Git event processed successfully!

💡 Monitor progress:
   • CI/CD: https://github.com/myorg/myrepo/actions
   • Code Quality: https://sonarcloud.io/project/myorg_myrepo
   • Team notifications: Slack #dev-team channel

stderr: (empty)
Exit Code: 0
```

**Pull Request Event Processing**:
```text
⚡ Processing Git event from webhook

📄 Event Source: webhook payload
   🌐 Platform: GitHub
   📡 Event Type: pull_request
   🎬 Action: opened
   📅 Timestamp: 2025-09-02T15:30:45Z

📂 Repository Context:
   🔗 Repository: myorg/myrepo
   🔀 Pull Request: #123
   📝 Title: "Add user profile management feature"
   👤 Author: feature-developer
   🌳 Branch: feature/user-profile → main
   📊 Changes: +234 -12 lines across 8 files

📋 Pull Request Details:
   🏷️  Labels: enhancement, feature, needs-review
   👥 Reviewers: @senior-dev, @team-lead
   🔗 URL: https://github.com/myorg/myrepo/pull/123

🔄 Processing Actions:

✅ Automated Checks:
   🤖 PR size analysis: Medium (246 lines changed)
   🔍 File pattern analysis: Frontend + Backend changes detected
   📋 Required reviewers assigned based on CODEOWNERS
   🏷️  Labels auto-applied: enhancement, frontend, backend

✅ CI/CD Pipeline:
   🚀 PR validation workflow triggered
   🧪 Test suite execution: unit, integration, e2e
   🏗️  Preview environment deployment initiated
   📊 Build #1235: https://github.com/myorg/myrepo/actions/runs/1235

✅ Code Quality:
   🔍 SonarCloud PR analysis
   🛡️  Security vulnerability scan
   📊 Code coverage diff calculation
   🔧 Linting and formatting checks

✅ Team Notifications:
   💬 Slack: Posted to #code-review channel
   👥 Reviewer notifications sent
   📧 Stakeholder summary email queued
   🔔 GitHub notifications triggered

✅ Documentation:
   📚 Affected API endpoints identified
   📝 Breaking changes analysis
   🔄 Documentation impact assessment

📊 Processing Summary:
   ⏱️  Processing time: 3.1 seconds
   🎯 Actions triggered: 12
   ✅ Successful: 12, ❌ Failed: 0
   🔄 Status: All checks in progress

🎉 Pull request event processed successfully!

💡 Next steps:
   • Wait for CI/CD pipeline completion
   • Review automated analysis results
   • Assign additional reviewers if needed

stderr: (empty)
Exit Code: 0
```

**Event Processing Failed**:
```text
⚡ Processing Git event from webhook

📄 Event Source: /tmp/malformed-payload.json

❌ Event parsing failed:
   • Error: Invalid JSON structure
   • Line 15: Unexpected token '}' at position 342
   • Expected: Valid GitHub/GitLab webhook payload format

🔍 Payload Analysis:
   📄 File size: 1.2KB
   🔍 Content type: text/plain (expected: application/json)
   📋 Structure: Malformed JSON object

💡 Troubleshooting:
   • Validate JSON syntax: cat /tmp/malformed-payload.json | jq .
   • Check webhook configuration: gz git webhook list --repo myorg/myrepo
   • Test webhook endpoint: gz git webhook test --id 12345678

❌ Example valid payload structure:
```json
{
  "action": "opened",
  "pull_request": { ... },
  "repository": { ... },
  "sender": { ... }
}
```

🚫 Git event processing failed due to invalid payload.

stderr: invalid JSON payload
Exit Code: 1
```

**Unknown Event Type**:
```text
⚡ Processing Git event from webhook

📄 Event Source: webhook payload
   🌐 Platform: GitHub
   📡 Event Type: custom_organization_event
   📅 Timestamp: 2025-09-02T15:30:45Z

⚠️  Unknown event type: custom_organization_event

📋 Supported Event Types:
   📌 Repository events: push, pull_request, issues, release
   👥 Organization events: repository, member, team
   🔐 Security events: security_advisory, secret_scanning_alert
   ⭐ Community events: fork, star, watch

🔍 Event Analysis:
   📄 Payload structure: valid JSON
   🌐 Platform: GitHub (recognized)
   📊 Size: 2.3KB
   🎯 Headers: X-GitHub-Event: custom_organization_event

💡 Options:
   1. Skip processing: gz git event process --skip-unknown
   2. Log for analysis: gz git event process --log-unknown
   3. Add custom handler: gz git event handlers add custom_organization_event

⚠️  Event logged but not processed. No actions triggered.

📊 Processing Summary:
   ⏱️  Processing time: 0.5 seconds
   🎯 Actions triggered: 0
   📋 Status: Unknown event type (logged)

stderr: unknown event type
Exit Code: 1
```

### Side Effects

**Files Created**:
- `~/.gzh/git/events/processed-<timestamp>.json` - Processing log
- `~/.gzh/git/events/actions-<event-id>.log` - Action execution log
- `/tmp/gz-event-<id>.json` - Temporary event processing data

**Files Modified**:
- Integration configuration files (if actions modify settings)
- Notification queue files
- CI/CD pipeline trigger logs

**State Changes**:
- External integrations triggered (CI/CD, notifications)
- Repository state updated (labels, assignments, etc.)
- Event processing metrics updated

### Validation

**Automated Tests**:
```bash
# Test event processing with sample payload
echo '{"action": "opened", "pull_request": {"number": 1}, "repository": {"full_name": "test/repo"}}' > /tmp/test-event.json
result=$(gz git event process --source webhook --file /tmp/test-event.json 2>&1)
exit_code=$?

assert_contains "$result" "Processing Git event from webhook"
# Exit code: 0 (success), 1 (parsing/unknown event), 2 (configuration error)

# Check processing log creation
assert_file_exists "$HOME/.gzh/git/events/processed-*.json"
log_content=$(cat "$HOME/.gzh/git/events/processed-"*".json" | head -1)
assert_contains "$log_content" '"event_type":'
assert_contains "$log_content" '"actions_triggered":'
```

**Manual Verification**:
1. Process push event and verify CI/CD triggers
2. Test pull request event processing
3. Check unknown event type handling
4. Verify malformed payload error handling
5. Test integration action execution
6. Confirm notification delivery

### Edge Cases

**Payload Variations**:
- Large payloads (>1MB) with many commits
- Minimal payloads missing optional fields
- Legacy webhook format compatibility
- Platform-specific payload differences

**Integration Failures**:
- CI/CD system unavailable
- Notification service rate limits
- Authentication token expiration
- Network connectivity issues

**Event Timing**:
- Out-of-order event delivery
- Duplicate event processing
- Very old events (timestamp validation)
- High-frequency event bursts

**Configuration Issues**:
- Missing integration configurations
- Invalid webhook signatures
- Misconfigured action handlers
- Permission errors for integrations

### Performance Expectations

**Response Time**:
- Simple events: < 2 seconds
- Complex events with many integrations: < 10 seconds
- Bulk event processing: < 30 seconds per batch
- Error handling: < 1 second

**Resource Usage**:
- Memory: < 100MB for large payloads
- CPU: Low to moderate during JSON processing
- Network: Varies by integration count

**Throughput**:
- Single event: immediate processing
- Batch processing: 10-20 events per minute
- Concurrent processing: 3-5 events simultaneously
- Queue management for high-volume scenarios

## Notes

- Support for GitHub, GitLab, and Gitea webhook events
- Extensible action system for custom integrations
- Event deduplication and replay protection
- Comprehensive logging and audit trails
- Integration with popular CI/CD and communication tools
- Custom event handler development support
- Batch processing for high-volume scenarios
- Real-time processing with queue fallback
