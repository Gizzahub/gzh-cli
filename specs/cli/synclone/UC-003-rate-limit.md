# Command: gz synclone github (Rate Limit Error)

## Scenario: Handle GitHub API rate limit exceeded

### Input

**Command**:

```bash
gz synclone github -o microsoft
```

**Prerequisites**:

- [ ] gzh-cli binary installed
- [ ] Network connectivity to api.github.com
- [ ] NO GITHUB_TOKEN environment variable set (to trigger rate limiting)
- [ ] Previous API calls have exhausted the rate limit

### Expected Output

**Rate Limit Error Case**:

```
12:31:20 INFO  [component=gzh-cli org=microsoft] Starting GitHub synclone operation
12:31:20 INFO  [component=gzh-cli org=microsoft] Starting synclone workflow: fetching repository list from GitHub
🔍 Fetching repository list from GitHub organization: microsoft

🚫 GitHub API Rate Limit Exceeded!
   Rate Limit: 60 requests/hour
   Remaining: 0
   Reset Time: Tue, 02 Sep 2025 12:45:43 KST
   Wait Time: 14 minutes 22 seconds

💡 Solution: Set GITHUB_TOKEN environment variable to bypass rate limits
   export GITHUB_TOKEN="your_github_personal_access_token"

12:31:20 ERROR [component=gzh-cli org=microsoft op=github-synclone] Non-retryable error encountered: [rate_limit] GitHub API rate limit exceeded: GitHub API rate limit exceeded
Error: [rate_limit] GitHub API rate limit exceeded: GitHub API rate limit exceeded

stderr: (empty)
Exit Code: 1
```

**Critical Requirement**:

- ❌ **NO Usage block should be displayed**
- ❌ **NO command help should be shown**

### Side Effects

**Files Created**: None (operation fails before file creation)
**Files Modified**: None\
**State Changes**: None (clean failure with no partial state)

### Validation

**Automated Tests**:

```bash
# Test rate limit error (requires exhausted API quota)
unset GITHUB_TOKEN  # Ensure no token
result=$(gz synclone github -o microsoft 2>&1)
exit_code=$?

# Critical assertions - NO Usage block
assert_not_contains "$result" "Usage:"
assert_not_contains "$result" "Flags:"
assert_not_contains "$result" "Global Flags:"

# Positive assertions - Error message content
assert_contains "$result" "🚫 GitHub API Rate Limit Exceeded!"
assert_contains "$result" "Rate Limit: 60 requests/hour"
assert_contains "$result" "Remaining: 0"
assert_contains "$result" "💡 Solution: Set GITHUB_TOKEN"
assert_contains "$result" "[rate_limit] GitHub API rate limit exceeded"
assert_exit_code 1

# Verify no side effects
assert_not_directory_exists "./microsoft"
assert_not_file_exists "./microsoft/gzh.yaml"
```

**Manual Verification**:

1. Ensure no GITHUB_TOKEN is set in environment
1. Run command on large organization (microsoft, google, etc.)
1. **Verify NO Usage block is displayed after error**
1. Confirm error message provides helpful guidance
1. Check that no partial files/directories are created

### Error Message Requirements

**Must Include**:

- 🚫 Clear rate limit exceeded indicator
- Specific rate limit details (60 requests/hour for unauthenticated)
- Remaining requests count (should be 0)
- Reset time with timezone
- Wait time in human-readable format
- 💡 Solution with exact command to set token

**Must NOT Include**:

- Usage block with command help
- List of available flags
- Global flags documentation

### Performance Expectations

**Response Time**:

- Error detection: < 5 seconds
- Error message display: Immediate
- Process termination: Clean and fast

**Resource Usage**:

- Memory: Minimal (error path)
- No file I/O operations
- Network: Only failed API call

## Notes

- This error handling was specifically fixed to suppress Usage display
- Rate limit applies to unauthenticated requests (60/hour)
- Authenticated requests have higher limits (5000/hour)
- Error recovery system should NOT retry rate limit errors
- Command should exit gracefully without creating partial state
