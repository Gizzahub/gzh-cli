# Command: gz synclone github (Large Organization Pagination)

## Scenario: Successfully handle organizations with many repositories using pagination

### Input

**Command**:

```bash
gz synclone github -o kubernetes
```

**Prerequisites**:

- [ ] gzh-cli binary installed
- [ ] Network connectivity to api.github.com
- [ ] GITHUB_TOKEN environment variable set (recommended for large operations)
- [ ] Target organization has >30 repositories (to test pagination)

### Expected Output

**Success Case with Pagination**:

```
12:40:01 INFO  [component=gzh-cli org=kubernetes] Starting GitHub synclone operation
12:40:01 INFO  [component=gzh-cli org=kubernetes] Starting synclone workflow: fetching repository list from GitHub
🔍 Fetching repository list from GitHub organization: kubernetes
📋 Found 79 repositories in organization kubernetes
📝 Generated gzh.yaml with 79 repositories
⚙️ Starting repository synchronization with strategy: reset
All Target 79 >>>>>>>>>>>>>>>>>>>>
kubernetes
dashboard
website
...
externaljwt
All Target <<<<<<<<<<<<<<<<<<<
Clone or Reset kubernetes [████████████████████████████████] 79/79
12:42:15 INFO  [component=gzh-cli org=kubernetes] GitHub synclone operation completed successfully duration=2m14s

stderr: (empty)
Exit Code: 0
```

### Side Effects

**Files Created**:

- `./kubernetes/`: Target directory for organization
- `./kubernetes/gzh.yaml`: Metadata file with all 79 repositories
- `./kubernetes/{repo-name}/`: Directory for each of the 79 repositories

**API Behavior**:

- Makes multiple API calls with pagination (100 repos per page)
- First call: `GET /orgs/kubernetes/repos?page=1&per_page=100`
- Continues until fewer than 100 repos returned
- Total repositories fetched: All repositories in organization

### Validation

**Automated Tests**:

```bash
# Test pagination with kubernetes organization (79 repos)
export GITHUB_TOKEN="your_token_here"
result=$(gz synclone github -o kubernetes 2>&1)
exit_code=$?

# Critical assertion - Pagination worked correctly
assert_contains "$result" "📋 Found 79 repositories"
assert_not_contains "$result" "📋 Found 30 repositories"  # Old broken behavior
assert_exit_code 0

# Verify all repositories were fetched
assert_directory_exists "./kubernetes"
assert_file_exists "./kubernetes/gzh.yaml"

# Count actual repositories created
repo_count=$(ls -1 ./kubernetes | grep -v gzh.yaml | wc -l)
assert_equals "$repo_count" "79"

# Verify gzh.yaml contains all repositories
yaml_repo_count=$(grep -c "name:" ./kubernetes/gzh.yaml)
assert_equals "$yaml_repo_count" "79"

# Verify specific repositories exist (from different pages)
assert_directory_exists "./kubernetes/kubernetes"        # First repo
assert_directory_exists "./kubernetes/dashboard"         # Early repo
assert_directory_exists "./kubernetes/externaljwt"       # Last repo
```

**Test with Different Organization Sizes**:

```bash
# Test with small organization (ScriptonBasestar - 37 repos)
result=$(gz synclone github -o ScriptonBasestar 2>&1)
assert_contains "$result" "📋 Found 37 repositories"

# Test with medium organization (should handle any size)
result=$(gz synclone github -o medium-org 2>&1)
assert_contains "$result" "📋 Found"
```

**Manual Verification**:

1. Choose organization with known repository count >30
1. Run synclone command with valid token
1. Verify reported count matches actual organization size
1. Check that repositories from "later pages" are present
1. Confirm gzh.yaml metadata includes all repositories

### Technical Implementation Details

**Pagination Logic**:

- Uses `perPage = 100` (GitHub's maximum per page)
- Loops through pages until `len(repos) < perPage`
- Accumulates all repositories in `allRepos` slice
- Returns complete list regardless of organization size

**API Calls Made**:

```
Page 1: GET /orgs/kubernetes/repos?page=1&per_page=100  (returns 79 repos)
       → len(repos) < 100, so pagination stops
       → Total: 79 repositories
```

**For larger organizations**:

```
Page 1: GET /orgs/microsoft/repos?page=1&per_page=100   (returns 100 repos)
Page 2: GET /orgs/microsoft/repos?page=2&per_page=100   (returns 100 repos)  
Page N: GET /orgs/microsoft/repos?page=N&per_page=100   (returns <100 repos)
       → Pagination stops, returns all repositories
```

### Edge Cases

**Exactly 100 Repositories**:

- Should make 2 API calls (100 + 0)
- Second call returns empty array, pagination stops

**Organizations with 30 Repositories**:

- Should still work correctly (single page)
- Must not be limited to 30 due to old default behavior

**Very Large Organizations** (>1000 repos):

- Consider using `--optimized` flag for better performance
- May hit rate limits without authentication token

### Performance Expectations

**Response Time**:

- 1-100 repos: < 10 seconds for API fetching
- 100-500 repos: < 30 seconds for API fetching
- 500+ repos: Consider --optimized mode

**API Usage**:

- Calls: ⌈total_repos / 100⌉ API calls
- Rate limits: Authenticated = 5000/hour, Unauthenticated = 60/hour
- Efficient: Uses maximum page size (100)

## Notes

- Fixed pagination bug that previously limited results to 30 repositories
- Implementation now correctly handles organizations of any size
- Each API call fetches up to 100 repositories (GitHub's maximum)
- Pagination continues until fewer than 100 repos are returned
- Total repository count in output should match actual organization size
