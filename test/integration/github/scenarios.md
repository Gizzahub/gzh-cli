# GitHub Integration Test Scenarios

This document describes the integration test scenarios for GitHub repository configuration management.

## Test Scenarios Overview

### 1. Basic Configuration Management

**Purpose**: Verify basic repository configuration operations work correctly.

**Test Steps**:

1. List all repositories in the test organization
1. Get configuration for a specific repository
1. Update repository settings (description, topics, visibility)
1. Verify changes were applied correctly

**Expected Results**:

- All API calls succeed without errors
- Repository settings are updated as expected
- Configuration retrieval matches applied settings

### 2. Template Application

**Purpose**: Test applying configuration templates to repositories.

**Test Steps**:

1. Create a configuration with templates
1. Apply template to matching repositories (dry-run)
1. Review proposed changes
1. Apply template (actual execution)
1. Verify all repositories have correct settings

**Expected Results**:

- Dry-run shows accurate preview of changes
- Templates are applied only to matching repositories
- Non-matching repositories remain unchanged

### 3. Policy Compliance Audit

**Purpose**: Ensure policy compliance checking works correctly.

**Test Steps**:

1. Define security and compliance policies
1. Run compliance audit across organization
1. Identify non-compliant repositories
1. Generate compliance report
1. Test policy exceptions

**Expected Results**:

- Audit correctly identifies violations
- Exceptions are properly handled
- Report includes all necessary details

### 4. Branch Protection Management

**Purpose**: Test branch protection rule configuration.

**Test Steps**:

1. Get current branch protection settings
1. Update protection rules (reviews, status checks)
1. Test enforcement settings
1. Verify protection is active
1. Remove protection rules

**Expected Results**:

- Branch protection can be configured
- Rules are enforced as specified
- Changes can be reverted

### 5. Bulk Operations

**Purpose**: Verify bulk operations work efficiently and correctly.

**Test Steps**:

1. Select multiple repositories using patterns
1. Apply configuration to all selected repos
1. Monitor progress and handle errors
1. Verify all repositories updated
1. Generate operation summary

**Expected Results**:

- Operations complete within reasonable time
- Errors are handled gracefully
- Summary shows accurate statistics

### 6. Rate Limit Handling

**Purpose**: Ensure proper rate limit management.

**Test Steps**:

1. Check current rate limit status
1. Make concurrent API requests
1. Observe rate limit handling
1. Test retry logic
1. Verify backoff behavior

**Expected Results**:

- Rate limits are respected
- Requests are retried appropriately
- No requests fail due to rate limiting

### 7. Error Recovery

**Purpose**: Test error handling and recovery mechanisms.

**Test Steps**:

1. Test with invalid repository names
1. Test with insufficient permissions
1. Test network timeout scenarios
1. Test partial failure recovery
1. Verify error reporting

**Expected Results**:

- Errors are caught and reported clearly
- Partial failures don't affect other operations
- System can recover from transient errors

### 8. Template Inheritance

**Purpose**: Verify template inheritance works correctly.

**Test Steps**:

1. Create base template
1. Create derived templates
1. Apply inherited template
1. Verify merged settings
1. Test override behavior

**Expected Results**:

- Inheritance chain is resolved correctly
- Overrides work as expected
- No circular dependencies

### 9. Webhook Configuration

**Purpose**: Test webhook management capabilities.

**Test Steps**:

1. List existing webhooks
1. Create new webhook configuration
1. Update webhook settings
1. Test webhook activation
1. Remove test webhooks

**Expected Results**:

- Webhooks can be managed programmatically
- Settings are applied correctly
- Cleanup removes all test webhooks

### 10. Security Settings

**Purpose**: Verify security configuration management.

**Test Steps**:

1. Enable vulnerability alerts
1. Configure secret scanning
1. Set up security policies
1. Test security advisories
1. Verify all settings active

**Expected Results**:

- Security features can be enabled
- Settings persist correctly
- Policies are enforced

## Test Data Requirements

### Required Test Repositories

- `test-repo-basic`: Simple public repository
- `test-repo-private`: Private repository
- `test-repo-protected`: Repository with branch protection
- `test-repo-archived`: Archived repository (for filtering)
- `integration-test-*`: Multiple repos for bulk operations

### Required Permissions

- Organization admin access
- Repository admin permissions
- Webhook management rights
- Security settings access

## Performance Benchmarks

### Expected Performance Metrics

- Single repository update: < 2 seconds
- Bulk update (10 repos): < 20 seconds
- Compliance audit (50 repos): < 30 seconds
- Full organization scan: < 2 minutes

### Concurrency Limits

- Maximum concurrent API requests: 10
- Rate limit buffer: 100 requests
- Retry attempts: 3
- Backoff multiplier: 2x

## Troubleshooting Common Issues

### Authentication Failures

- Verify token has all required scopes
- Check token expiration
- Ensure organization access

### Permission Errors

- Confirm admin access to test org
- Check repository-specific permissions
- Verify API scope requirements

### Rate Limiting

- Monitor remaining quota
- Implement proper delays
- Use conditional requests

### Network Issues

- Check proxy settings
- Verify GitHub API accessibility
- Test with curl/wget first

## CI/CD Integration

### GitHub Actions Configuration

```yaml
integration-tests:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: "1.22"
    - name: Run Integration Tests
      env:
        GITHUB_TOKEN: ${{ secrets.INTEGRATION_TEST_TOKEN }}
        GITHUB_TEST_ORG: ${{ vars.INTEGRATION_TEST_ORG }}
      run: |
        ./test/integration/run_integration_tests.sh all
```

### Required Secrets

- `INTEGRATION_TEST_TOKEN`: PAT with required scopes
- `INTEGRATION_TEST_ORG`: Test organization name

## Success Criteria

All integration tests pass when:

1. ✅ All API operations complete successfully
1. ✅ Configuration changes are applied correctly
1. ✅ Policy compliance is accurately reported
1. ✅ Error handling works as designed
1. ✅ Performance meets benchmarks
1. ✅ No data corruption or loss
1. ✅ Clean test execution without side effects
