# Command: gz pm update

## Scenario: Update all package managers and their packages

### Input

**Command**:

```bash
gz pm update
```

**Prerequisites**:

- [ ] Package managers installed (asdf, brew, npm, etc.)
- [ ] Network connectivity
- [ ] Admin permissions (for system-wide package managers)

### Expected Output

**Success Case**:

```text
🔄 Updating package managers...

📦 Homebrew
✅ brew update: Updated 23 formulae
✅ brew upgrade: Upgraded 5 packages
   - node: 20.11.0 -> 20.11.1
   - git: 2.43.0 -> 2.43.1
   - python@3.11: 3.11.7 -> 3.11.8

📦 asdf
✅ asdf update: Updated to latest version
✅ asdf plugin update --all: 8 plugins updated
✅ Available updates:
   - golang: 1.21.5 -> 1.21.6 (use: asdf install golang 1.21.6)
   - nodejs: 20.11.0 -> 20.11.1 (use: asdf install nodejs 20.11.1)

📦 Node.js (npm)
✅ npm update -g: 12 global packages updated
   - @angular/cli: 17.0.7 -> 17.0.8
   - typescript: 5.3.2 -> 5.3.3

📦 Python (pip)
✅ pip install --upgrade pip: Updated to 24.0
✅ pip list --outdated: 6 packages can be updated
   - requests: 2.31.0 -> 2.32.0
   - numpy: 1.24.3 -> 1.24.4

🎉 Package manager updates completed!
📋 Manual action needed:
   - Update asdf language versions (commands shown above)
   - Consider updating pip packages: pip install --upgrade <package>

stderr: (empty)
Exit Code: 0
```

**Partial Success (Some Failures)**:

```text
🔄 Updating package managers...

📦 Homebrew
✅ brew update: Updated 15 formulae
❌ brew upgrade: Failed to upgrade 2 packages
   - postgresql: version conflict (manual intervention needed)
   - docker: insufficient disk space

📦 SDKMAN
❌ Network error: Cannot reach SDKMAN servers
💡 Check network connection and try again later

📦 Node.js (npm)
✅ npm update -g: 8 global packages updated

⚠️  Some updates failed. See details above.
🔧 Manual fixes may be required for failed updates.

stderr: (empty)
Exit Code: 1
```

**No Package Managers Found**:

```text
🔍 Scanning for package managers...

❌ No supported package managers found!

💡 Supported package managers:
   - Homebrew (macOS/Linux): Install from https://brew.sh
   - asdf (Version manager): Install from https://asdf-vm.com
   - SDKMAN (Java ecosystem): Install from https://sdkman.io
   - Node.js npm: Installed with Node.js
   - Python pip: Installed with Python
   - Rust cargo: Installed with Rust

🚫 Nothing to update.

stderr: no package managers found
Exit Code: 2
```

### Side Effects

**Files Created**:

- `~/.gzh/pm-update.log` - Detailed update log
- Package manager cache files

**Files Modified**:

- Package manager databases updated
- Installed packages upgraded

**State Changes**:

- Package manager databases refreshed
- Available package versions updated
- Some packages upgraded automatically

### Validation

**Automated Tests**:

```bash
# Test update command (requires actual package managers)
result=$(gz pm update 2>&1)
exit_code=$?

# Should find at least one package manager in CI/test environment
assert_not_contains "$result" "No supported package managers found"
assert_contains "$result" "Updating package managers"

# Check log file creation
assert_file_exists "$HOME/.gzh/pm-update.log"
```

**Manual Verification**:

1. Run on system with multiple package managers
1. Verify each package manager is properly updated
1. Check that upgrade recommendations are actionable
1. Confirm failed updates are clearly reported

### Edge Cases

**Network Issues**:

- Offline environment handling
- Partial network connectivity
- Timeout handling for slow connections
- Proxy configuration support

**Permission Issues**:

- System package managers requiring sudo
- User-space vs system-wide installations
- Directory permission problems

**Disk Space Issues**:

- Insufficient space for package downloads
- Cache cleanup before updates
- Warning when approaching disk limits

**Version Conflicts**:

- Package dependency conflicts
- Breaking changes in major updates
- Rollback instructions when needed

### Performance Expectations

**Response Time**:

- Database updates: < 30 seconds per package manager
- Package discovery: < 10 seconds
- Full update cycle: 2-10 minutes depending on updates available

**Resource Usage**:

- Memory: < 200MB
- Network: Varies by number of package updates
- Disk: Temporary cache space for downloads

## Notes

- Supports major package managers across platforms
- Non-destructive updates (database refresh only)
- Manual confirmation for major package upgrades
- Detailed logging for troubleshooting
- Integration with system notification services
- Respects package manager-specific configurations
