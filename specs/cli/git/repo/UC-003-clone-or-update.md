# Command: gz git repo clone-or-update

## Scenario: Smart clone or update repository with strategy

### Input

**Command**:

```bash
gz git repo clone-or-update https://github.com/user/repo.git --strategy rebase
```

**Prerequisites**:

- [ ] Git installed and configured
- [ ] Network connectivity to Git platform
- [ ] Authentication for private repositories
- [ ] Write permissions in target directory

### Expected Output

**Success Case (Clone)**:

```text
📍 Repository: https://github.com/user/repo.git
📂 Target: ./repo
🔧 Strategy: rebase

🔍 Repository not found locally, cloning...
Cloning into 'repo'...
remote: Counting objects: 245, done.
remote: Compressing objects: 100% (156/156), done.
remote: Total 245 (delta 89), reused 178 (delta 45)
Receiving objects: 100% (245/245), 45.23 KiB | 2.26 MiB/s, done.
Resolving deltas: 100% (89/89), done.

✅ Repository cloned successfully to ./repo

stderr: (empty)
Exit Code: 0
```

**Success Case (Update)**:

```text
📍 Repository: https://github.com/user/repo.git
📂 Target: ./repo
🔧 Strategy: rebase

🔍 Repository found locally, updating...
🔄 Fetching latest changes...
🚀 Rebasing local changes on remote...
Current branch main is up to date.

✅ Repository updated successfully

stderr: (empty)
Exit Code: 0
```

**Rebase Conflict Error**:

```text
📍 Repository: https://github.com/user/repo.git
📂 Target: ./repo
🔧 Strategy: rebase

🔍 Repository found locally, updating...
🔄 Fetching latest changes...
🚀 Rebasing local changes on remote...

❌ Rebase conflict detected
🔧 Manual resolution required:
   1. Resolve conflicts in affected files
   2. Run: git add <resolved-files>
   3. Run: git rebase --continue

📂 Working directory: ./repo

stderr: rebase conflicts detected
Exit Code: 1
```

**Authentication Error**:

```text
📍 Repository: https://github.com/private-org/private-repo.git
📂 Target: ./private-repo
🔧 Strategy: rebase

❌ Authentication failed
🔑 Private repository requires authentication
💡 Solutions:
   - Set GITHUB_TOKEN environment variable
   - Use SSH key authentication
   - Configure git credentials

stderr: authentication required
Exit Code: 1
```

### Side Effects

**Files Created**:

- `./repo/` - Git repository directory (clone case)
- `.git/` - Git metadata directory

**Files Modified**:

- Repository files updated according to strategy

**State Changes**:

- Local repository synchronized with remote
- Git history updated based on strategy

### Validation

**Automated Tests**:

```bash
# Test clone
result=$(gz git repo clone-or-update https://github.com/octocat/Hello-World.git 2>&1)
exit_code=$?

assert_contains "$result" "Repository cloned successfully"
assert_exit_code 0
assert_directory_exists "./Hello-World"
assert_directory_exists "./Hello-World/.git"

# Test update (run again)
result=$(gz git repo clone-or-update https://github.com/octocat/Hello-World.git --strategy reset 2>&1)
assert_contains "$result" "Repository updated successfully"
assert_exit_code 0

# Cleanup
rm -rf ./Hello-World
```

**Manual Verification**:

1. Clone public repository - verify files are present
1. Make local changes and run update with different strategies
1. Test with private repository requiring authentication
1. Verify conflict resolution workflow

### Edge Cases

**Different Strategies**:

- `rebase` (default): Rebase local changes on remote
- `reset`: Hard reset to match remote (discards local changes)
- `pull`: Standard merge
- `fetch`: Update refs only
- `skip`: Leave unchanged

**Repository States**:

- Clean working directory
- Uncommitted changes
- Untracked files
- Conflicting changes
- Detached HEAD

**URL Formats**:

- HTTPS: `https://github.com/user/repo.git`
- SSH: `git@github.com:user/repo.git`
- Short format: `github.com/user/repo`

### Performance Expectations

**Response Time**:

- Small repos (\<10MB): < 30 seconds
- Large repos (>100MB): Progress indication
- Update operations: < 10 seconds typically

**Resource Usage**:

- Disk space: Repository size + Git metadata
- Network: Transfer based on changes
- Memory: < 200MB for large repositories

## Notes

- Auto-detects repository name from URL for target directory
- Supports all major Git hosting platforms
- Handles authentication automatically via environment or SSH
- Strategy selection allows different update behaviors
- Conflict resolution guidance provided for rebase failures
