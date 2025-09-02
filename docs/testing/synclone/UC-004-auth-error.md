# Command: gz synclone github (Authentication Error)

## Scenario: Handle GitHub authentication failure

### Input

**Command**:
```bash
gz synclone github -o private-org --token invalid_token_12345
```

**Prerequisites**:
- [ ] gzh-cli binary installed
- [ ] Network connectivity to api.github.com
- [ ] Invalid or expired GitHub token provided
- [ ] Target organization exists but requires authentication

### Expected Output

**Authentication Error Case**:
```
12:35:10 INFO  [component=gzh-cli org=private-org] Starting GitHub synclone operation
12:35:10 INFO  [component=gzh-cli org=private-org] Starting synclone workflow: fetching repository list from GitHub
🔍 Fetching repository list from GitHub organization: private-org
12:35:11 ERROR [component=gzh-cli org=private-org op=github-synclone] Non-retryable error encountered: [auth] GitHub authentication failed: failed to get repositories: 401 Unauthorized
Error: [auth] GitHub authentication failed: failed to get repositories: 401 Unauthorized

stderr: (empty)
Exit Code: 1
```

**Critical Requirement**: 
- ❌ **NO Usage block should be displayed**
- ❌ **NO command help should be shown**

### Side Effects

**Files Created**: None (operation fails before file creation)
**Files Modified**: None  
**State Changes**: None (clean failure with no partial state)

### Validation

**Automated Tests**:
```bash
# Test authentication error
result=$(gz synclone github -o private-org --token invalid_token_12345 2>&1)
exit_code=$?

# Critical assertions - NO Usage block
assert_not_contains "$result" "Usage:"
assert_not_contains "$result" "Flags:"
assert_not_contains "$result" "Global Flags:"

# Positive assertions - Error message content
assert_contains "$result" "[auth] GitHub authentication failed"
assert_contains "$result" "401 Unauthorized"
assert_exit_code 1

# Verify no side effects
assert_not_directory_exists "./private-org"
assert_not_file_exists "./private-org/gzh.yaml"
```

**Alternative Test with Environment Variable**:
```bash
# Test with invalid environment token
export GITHUB_TOKEN="invalid_token_12345"
result=$(gz synclone github -o private-org 2>&1)
exit_code=$?

assert_not_contains "$result" "Usage:"
assert_contains "$result" "[auth] GitHub authentication failed"
assert_exit_code 1
unset GITHUB_TOKEN
```

**Manual Verification**:
1. Try with obviously invalid token (random string)
2. Try with expired token
3. Try with token lacking required scopes
4. **Verify NO Usage block is displayed after error**
5. Confirm error message is clear and actionable

### Common Authentication Scenarios

**Invalid Token Format**:
```bash
gz synclone github -o myorg --token "not-a-real-token"
# Expected: 401 Unauthorized error
```

**Expired Token**:
```bash
gz synclone github -o myorg --token "ghp_expired_token"
# Expected: 401 Unauthorized error
```

**Insufficient Scope**:
```bash
gz synclone github -o myorg --token "ghp_limited_scope_token"
# Expected: 403 Forbidden (if token valid but lacks repo scope)
```

**Private Organization**:
```bash
gz synclone github -o private-enterprise-org
# Expected: 401 Unauthorized (unauthenticated access to private org)
```

### Error Message Requirements

**Must Include**:
- Clear authentication failure indicator
- Specific HTTP status code (401 Unauthorized)
- Error type classification [auth]
- Context about which operation failed

**Must NOT Include**:
- Usage block with command help
- List of available flags
- Verbose troubleshooting steps

**Should Include** (future enhancement):
- Hint about checking token validity
- Link to GitHub token creation guide
- Suggestion to verify token scopes

### Performance Expectations

**Response Time**:
- Error detection: < 3 seconds
- Error message display: Immediate
- Process termination: Clean and fast

**Resource Usage**:
- Memory: Minimal (error path)
- No file I/O operations
- Network: Only failed authentication attempt

## Notes

- Authentication errors should fail fast
- No retry attempts should be made for auth failures
- Token validation happens on first API call
- Error suppresses Usage display (fixed in recent update)
- Future enhancement could include token validation hints