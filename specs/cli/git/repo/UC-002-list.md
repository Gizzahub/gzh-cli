# Command: gz git repo list

## Scenario: List repositories from Git platform organization

### Input

**Command**:

```bash
gz git repo list --org myorg --platform github
```

**Prerequisites**:

- [ ] Valid authentication token (optional, for private repos)
- [ ] Organization exists
- [ ] Network connectivity

### Expected Output

**Success Case**:

```text
📋 Repositories in organization: myorg (platform: github)

NAME                    VISIBILITY  LANGUAGE    STARS  UPDATED
awesome-project         public      Go          45     2 days ago
secret-sauce           private     TypeScript   0      1 week ago  
documentation          public      Markdown     12     3 days ago
legacy-app             public      Java         8      2 months ago

Total: 4 repositories (2 public, 2 private)

stderr: (empty)
Exit Code: 0
```

**Empty Organization**:

```text
📋 Repositories in organization: empty-org (platform: github)

No repositories found.

stderr: (empty)
Exit Code: 0
```

**Rate Limit Error**:

```text
🚫 GitHub API Rate Limit Exceeded!
   Rate Limit: 60 requests/hour
   Remaining: 0
   Reset Time: Tue, 02 Sep 2025 12:45:43 KST
   Wait Time: 14 minutes 22 seconds

💡 Solution: Set GITHUB_TOKEN environment variable to bypass rate limits
   export GITHUB_TOKEN="your_github_personal_access_token"

Error: [rate_limit] GitHub API rate limit exceeded

stderr: (empty)
Exit Code: 1
```

### Side Effects

**Files Created**: None
**Files Modified**: None
**State Changes**: None (read-only operation)

### Validation

**Automated Tests**:

```bash
# Test successful listing
export GITHUB_TOKEN="your_token"
result=$(gz git repo list --org ScriptonBasestar --platform github 2>&1)
exit_code=$?

assert_contains "$result" "Repositories in organization: ScriptonBasestar"
assert_contains "$result" "Total:"
assert_exit_code 0

# Test with output format
result=$(gz git repo list --org ScriptonBasestar --platform github --output json 2>&1)
assert_contains "$result" '"name":'
assert_contains "$result" '"visibility":'
```

**Manual Verification**:

1. Run command with known organization
1. Verify repository count matches web interface
1. Check that private repositories are shown only with valid token
1. Confirm repository details are accurate

### Edge Cases

**Large Organizations**:

- Organizations with >100 repositories (pagination)
- Should handle API pagination automatically
- Progress indication for large lists

**Mixed Visibility**:

- Public repositories visible without token
- Private repositories require valid authentication
- Clear indication of visibility status

**Different Output Formats**:

- Default table format for human reading
- JSON format for programmatic use
- CSV format for data export

### Performance Expectations

**Response Time**:

- Small orgs (\<10 repos): < 3 seconds
- Large orgs (100+ repos): < 10 seconds with pagination
- Very large orgs (1000+ repos): Progress indication

**Resource Usage**:

- Memory: < 100MB for large organizations
- Network: Efficient API pagination

## Notes

- Supports multiple output formats: table (default), json, csv
- Automatically handles API pagination for large organizations
- Shows repository metadata: language, stars, last update
- Private repository access requires authentication
- Cross-platform support for GitHub, GitLab, Gitea, Gogs
