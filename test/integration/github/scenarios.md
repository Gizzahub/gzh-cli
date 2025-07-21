# GitHub Integration Test Scenarios

This document describes the integration test scenarios for GitHub repository configuration management.

## Test Scenarios Overview

### 1. Basic Configuration Management
**Purpose**: Verify basic repository configuration operations work correctly.

**Test Steps**:
1. List all repositories in the test organization
2. Get configuration for a specific repository
3. Update repository settings (description, topics, visibility)
4. Verify changes were applied correctly

**Expected Results**:
- All API calls succeed without errors
- Repository settings are updated as expected
- Configuration retrieval matches applied settings

### 2. Template Application
**Purpose**: Test applying configuration templates to repositories.

**Test Steps**:
1. Create a configuration with templates
2. Apply template to matching repositories (dry-run)
3. Review proposed changes
4. Apply template (actual execution)
5. Verify all repositories have correct settings

**Expected Results**:
- Dry-run shows accurate preview of changes
- Templates are applied only to matching repositories
- Non-matching repositories remain unchanged

### 3. Policy Compliance Audit
**Purpose**: Ensure policy compliance checking works correctly.

**Test Steps**:
1. Define security and compliance policies
2. Run compliance audit across organization
3. Identify non-compliant repositories
4. Generate compliance report
5. Test policy exceptions

**Expected Results**:
- Audit correctly identifies violations
- Exceptions are properly handled
- Report includes all necessary details

### 4. Branch Protection Management
**Purpose**: Test branch protection rule configuration.

**Test Steps**:
1. Get current branch protection settings
2. Update protection rules (reviews, status checks)
3. Test enforcement settings
4. Verify protection is active
5. Remove protection rules

**Expected Results**:
- Branch protection can be configured
- Rules are enforced as specified
- Changes can be reverted

### 5. Bulk Operations
**Purpose**: Verify bulk operations work efficiently and correctly.

**Test Steps**:
1. Select multiple repositories using patterns
2. Apply configuration to all selected repos
3. Monitor progress and handle errors
4. Verify all repositories updated
5. Generate operation summary

**Expected Results**:
- Operations complete within reasonable time
- Errors are handled gracefully
- Summary shows accurate statistics

### 6. Rate Limit Handling
**Purpose**: Ensure proper rate limit management.

**Test Steps**:
1. Check current rate limit status
2. Make concurrent API requests
3. Observe rate limit handling
4. Test retry logic
5. Verify backoff behavior

**Expected Results**:
- Rate limits are respected
- Requests are retried appropriately
- No requests fail due to rate limiting

### 7. Error Recovery
**Purpose**: Test error handling and recovery mechanisms.

**Test Steps**:
1. Test with invalid repository names
2. Test with insufficient permissions
3. Test network timeout scenarios
4. Test partial failure recovery
5. Verify error reporting

**Expected Results**:
- Errors are caught and reported clearly
- Partial failures don't affect other operations
- System can recover from transient errors

### 8. Template Inheritance
**Purpose**: Verify template inheritance works correctly.

**Test Steps**:
1. Create base template
2. Create derived templates
3. Apply inherited template
4. Verify merged settings
5. Test override behavior

**Expected Results**:
- Inheritance chain is resolved correctly
- Overrides work as expected
- No circular dependencies

### 9. Webhook Configuration
**Purpose**: Test webhook management capabilities.

**Test Steps**:
1. List existing webhooks
2. Create new webhook configuration
3. Update webhook settings
4. Test webhook activation
5. Remove test webhooks

**Expected Results**:
- Webhooks can be managed programmatically
- Settings are applied correctly
- Cleanup removes all test webhooks

### 10. Security Settings
**Purpose**: Verify security configuration management.

**Test Steps**:
1. Enable vulnerability alerts
2. Configure secret scanning
3. Set up security policies
4. Test security advisories
5. Verify all settings active

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
        go-version: '1.24'
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
2. ✅ Configuration changes are applied correctly
3. ✅ Policy compliance is accurately reported
4. ✅ Error handling works as designed
5. ✅ Performance meets benchmarks
6. ✅ No data corruption or loss
7. ✅ Clean test execution without side effects
