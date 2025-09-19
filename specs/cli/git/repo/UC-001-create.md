# Command: gz git repo create

## Scenario: Create new repository on Git platform

### Input

**Command**:

```bash
gz git repo create --name my-new-repo --org myorg --platform github
```

**Prerequisites**:

- [ ] Valid authentication token for target platform
- [ ] Organization exists and user has create permissions
- [ ] Network connectivity

### Expected Output

**Success Case**:

```text
🔧 Creating repository: my-new-repo in organization: myorg
📋 Platform: github
🚀 Repository created successfully
📍 URL: https://github.com/myorg/my-new-repo
✅ Repository ready for use

stderr: (empty)
Exit Code: 0
```

**Error Cases**:

**Repository Already Exists**:

```text
❌ Repository 'my-new-repo' already exists in organization 'myorg'
🔍 Existing URL: https://github.com/myorg/my-new-repo

stderr: (empty)
Exit Code: 1
```

**Authentication Error**:

```text
🔑 Authentication failed for platform: github
💡 Please check your GITHUB_TOKEN environment variable
   export GITHUB_TOKEN="your_github_personal_access_token"

stderr: (empty)
Exit Code: 1
```

### Side Effects

**Files Created**: None (remote operation only)
**Files Modified**: None
**State Changes**:

- New repository created on target platform
- Repository metadata configured

### Validation

**Automated Tests**:

```bash
# Test successful creation
result=$(gz git repo create --name test-repo-12345 --org test-org --platform github 2>&1)
exit_code=$?

assert_contains "$result" "Repository created successfully"
assert_contains "$result" "https://github.com/test-org/test-repo-12345"
assert_exit_code 0

# Verify repository exists
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/test-org/test-repo-12345 | grep '"name": "test-repo-12345"'

# Cleanup
gz git repo delete --name test-repo-12345 --org test-org --platform github --confirm
```

**Manual Verification**:

1. Run command with valid organization and authentication
1. Check that repository appears in organization on web interface
1. Verify repository is accessible with correct permissions
1. Confirm repository URL is correct and accessible

### Edge Cases

**Special Characters in Name**:

- Hyphens: `my-awesome-repo` (valid)
- Underscores: `my_awesome_repo` (valid)
- Dots: `my.awesome.repo` (valid for some platforms)
- Spaces: `my awesome repo` (should be rejected with clear error)

**Long Repository Names**:

- Very long names (>100 characters) should be handled gracefully
- Platform-specific limits should be enforced

**Network Issues**:

- Timeout handling with clear error messages
- Retry logic for temporary failures

### Performance Expectations

**Response Time**:

- Normal case: < 5 seconds
- Network delays: < 30 seconds with progress indication

**Resource Usage**:

- Memory: < 50MB
- Network: Single API call to create repository

## Notes

- Repository visibility defaults to public unless --private flag specified
- Template repository support available with --template flag
- Auto-initialization with README/gitignore supported
- Cross-platform repository creation enables multi-platform workflows
