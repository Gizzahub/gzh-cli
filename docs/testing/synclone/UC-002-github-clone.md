# Command: gz synclone github

## Scenario: Successfully clone all repositories from GitHub organization

### Input

**Command**:
```bash
gz synclone github -o ScriptonBasestar
```

**Prerequisites**:
- [ ] gzh-cli binary installed (`gz --version` works)
- [ ] Network connectivity to api.github.com
- [ ] GITHUB_TOKEN environment variable set (recommended for rate limits)
- [ ] Write permissions in current directory

### Expected Output

**Success Case**:
```
12:30:01 INFO  [component=gzh-cli org=ScriptonBasestar] Starting GitHub synclone operation
12:30:01 INFO  [component=gzh-cli org=ScriptonBasestar] Starting synclone workflow: fetching repository list from GitHub
🔍 Fetching repository list from GitHub organization: ScriptonBasestar
📋 Found 37 repositories in organization ScriptonBasestar
📝 Generated gzh.yaml with 37 repositories
⚙️ Starting repository synchronization with strategy: reset
All Target 37 >>>>>>>>>>>>>>>>>>>>
sb-wp-review_infobox
sb-wp-template-skeleton
...
sb-libs-py
All Target <<<<<<<<<<<<<<<<<<<
Clone or Reset sb-wp-review_infobox [████████████████████████████████] 37/37
12:30:45 INFO  [component=gzh-cli org=ScriptonBasestar] GitHub synclone operation completed successfully duration=44.2s

stderr: (empty)
Exit Code: 0
```

### Side Effects

**Files Created**:
- `./ScriptonBasestar/`: Target directory for organization
- `./ScriptonBasestar/gzh.yaml`: Metadata file with repository information
- `./ScriptonBasestar/{repo-name}/`: Directory for each repository with full git clone

**Files Modified**: None (first run)

**State Changes**:
- Creates organization-specific directory structure
- Initializes git repositories with remote tracking
- Saves organizational metadata for future syncs

### Validation

**Automated Tests**:
```bash
# Test successful clone
export GITHUB_TOKEN="your_token_here"
result=$(gz synclone github -o ScriptonBasestar 2>&1)
exit_code=$?

# Assertions
assert_contains "$result" "📋 Found"
assert_contains "$result" "repositories in organization ScriptonBasestar"
assert_contains "$result" "📝 Generated gzh.yaml"
assert_exit_code 0

# Verify side effects
assert_directory_exists "./ScriptonBasestar"
assert_file_exists "./ScriptonBasestar/gzh.yaml"
assert_file_contains "./ScriptonBasestar/gzh.yaml" "organization: ScriptonBasestar"
assert_file_contains "./ScriptonBasestar/gzh.yaml" "provider: github"

# Verify repository structure
repo_count=$(ls -1 ./ScriptonBasestar | grep -v gzh.yaml | wc -l)
assert_greater_than "$repo_count" 30

# Verify git repositories
for repo_dir in ./ScriptonBasestar/*/; do
    assert_directory_exists "$repo_dir/.git"
    cd "$repo_dir"
    remote_url=$(git remote get-url origin)
    assert_contains "$remote_url" "github.com/ScriptonBasestar"
    cd - > /dev/null
done
```

**Manual Verification**:
1. Run command and observe progress indicators
2. Check that target directory `./ScriptonBasestar` is created
3. Verify `gzh.yaml` contains correct organization metadata
4. Confirm each repository directory contains a valid git clone
5. Validate that remote URLs point to ScriptonBasestar organization

### Edge Cases

**Large Organization (>100 repos)**:
```bash
gz synclone github -o kubernetes
# Should handle pagination automatically
# Should show progress for all repositories (not just first 30)
```

**Empty Organization**:
```bash
gz synclone github -o empty-test-org-12345
# Expected: "📋 Found 0 repositories in organization empty-test-org-12345"
# Should create target directory but no repository subdirectories
```

**Special Characters in Repository Names**:
- Repository names with hyphens: `my-awesome-repo`
- Repository names with underscores: `my_awesome_repo`
- Repository names with dots: `my.awesome.repo`

### Performance Expectations

**Response Time**:
- Small orgs (<10 repos): < 30 seconds
- Medium orgs (10-50 repos): < 2 minutes
- Large orgs (50-100 repos): < 5 minutes

**Resource Usage**:
- Memory: < 500MB (configurable with --memory-limit)
- Concurrent operations: 10 parallel clones (configurable with --parallel)
- Network: Respects GitHub API rate limits

## Notes

- Default strategy is "reset" which performs `git reset --hard HEAD && git pull`
- Creates organization-specific directory automatically (./ScriptonBasestar)
- Subsequent runs will use existing gzh.yaml if available
- Supports resumable operations with `--resume` flag
- Optimized mode available with `--optimized` for large organizations