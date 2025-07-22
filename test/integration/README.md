# Integration Tests

This directory contains integration tests that interact with real external services.

## Prerequisites

### GitHub Integration Tests

To run GitHub integration tests, you need:

1. **GitHub Personal Access Token (PAT)**
   - Create a token at: https://github.com/settings/tokens
   - Required scopes:
     - `repo` (Full control of private repositories)
     - `admin:org` (Full control of orgs and teams, read and write org projects)
     - `admin:repo_hook` (Full control of repository hooks)
     - `delete_repo` (Delete repositories) - Optional, for cleanup tests

2. **Test Organization**
   - Create a test organization on GitHub
   - Or use an existing organization where you have admin rights
   - ⚠️ **WARNING**: Do not use production organizations!

3. **Environment Variables**
   ```bash
   export GITHUB_TOKEN="your-github-token"
   export GITHUB_TEST_ORG="your-test-org-name"
   ```

## Running Integration Tests

### Run All Integration Tests

```bash
go test ./test/integration/... -v
```

### Run Only GitHub Integration Tests

```bash
go test ./test/integration/github -v
```

### Run Specific Test

```bash
go test ./test/integration/github -v -run TestIntegration_RepoConfig_EndToEnd
```

### Skip Integration Tests

Integration tests are automatically skipped if required environment variables are not set.

## Test Organization Setup

### Recommended Test Organization Structure

1. Create a dedicated test organization (e.g., `mycompany-test`)
2. Create test repositories:
   ```
   test-repo-1 (public, for basic tests)
   test-repo-2 (private, for permission tests)
   test-repo-archived (archived, for filter tests)
   integration-test-repo-1 (for bulk operations)
   integration-test-repo-2 (for bulk operations)
   ```

### Safety Guidelines

1. **Never use production organizations or repositories**
2. **Use repositories that can be safely modified**
3. **Clean up test data after tests complete**
4. **Use distinctive names for test resources** (e.g., prefix with `test-` or `integration-`)

## Test Scenarios

### 1. Repository Configuration Management

- List repositories in organization
- Get repository configuration
- Update repository settings
- Apply configuration templates
- Bulk operations on multiple repositories

### 2. Policy Compliance

- Define security policies
- Run compliance audits
- Generate compliance reports
- Test policy exceptions

### 3. Branch Protection

- Get branch protection rules
- Update branch protection settings
- Test required status checks
- Test review requirements

### 4. Rate Limiting

- Test rate limit handling
- Concurrent request management
- Retry logic with backoff

### 5. Error Handling

- Invalid authentication
- Non-existent resources
- Permission errors
- Network failures

## Writing New Integration Tests

### Test Structure

```go
func TestIntegration_Feature_Scenario(t *testing.T) {
    // Skip if environment not configured
    skipIfNoTestOrg(t)

    // Setup
    ctx := context.Background()
    client := createTestClient()

    // Test
    t.Run("SubTest", func(t *testing.T) {
        // Test implementation
    })

    // Cleanup
    defer cleanup()
}
```

### Best Practices

1. **Idempotency**: Tests should be runnable multiple times
2. **Isolation**: Tests should not depend on other tests
3. **Cleanup**: Always clean up created resources
4. **Timeouts**: Use appropriate timeouts for API calls
5. **Logging**: Log important information for debugging
6. **Skip Logic**: Skip tests when prerequisites aren't met

## Continuous Integration

### GitHub Actions Setup

```yaml
- name: Run Integration Tests
  env:
    GITHUB_TOKEN: ${{ secrets.INTEGRATION_TEST_TOKEN }}
    GITHUB_TEST_ORG: ${{ secrets.INTEGRATION_TEST_ORG }}
  run: |
    go test ./test/integration/... -v -timeout 30m
```

### Security Notes

- Never commit tokens or sensitive data
- Use GitHub Secrets for CI/CD
- Rotate test tokens regularly
- Monitor test organization for unauthorized access

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify token has required scopes
   - Check token hasn't expired
   - Ensure environment variables are set

2. **Rate Limiting**
   - Tests may hit rate limits with small quotas
   - Consider using GitHub Apps for higher limits
   - Add delays between tests if needed

3. **Permission Errors**
   - Ensure token has admin access to test org
   - Some tests require specific permissions
   - Check organization settings

4. **Network Issues**
   - Tests require internet connectivity
   - Corporate proxies may interfere
   - Check firewall settings

## Test Data Management

### Creating Test Data

```bash
# Create test repositories
./scripts/setup-test-org.sh

# Populate with sample data
./scripts/populate-test-data.sh
```

### Cleaning Test Data

```bash
# Remove test artifacts
./scripts/cleanup-test-org.sh
```

## Performance Considerations

- Integration tests are slower than unit tests
- Run in parallel where possible
- Cache API responses when appropriate
- Use bulk operations to reduce API calls

## Future Enhancements

- [ ] GitLab integration tests
- [ ] Gitea integration tests
- [ ] Webhook testing
- [ ] GitHub Actions integration
- [ ] Performance benchmarks
- [ ] Load testing scenarios
