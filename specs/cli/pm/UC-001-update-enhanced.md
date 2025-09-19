# Command: gz pm update

## Scenario: Update all package managers and their packages

### Input

**Command**:

```bash
gz pm update --all
```

**Additional Command Variants**:

```bash
gz pm update --manager brew                    # Single manager
gz pm update --managers brew,asdf,npm         # Multiple specific managers
gz pm update --all --strategy latest          # Update strategy
gz pm update --all --dry-run                  # Preview changes
gz pm update --all --output json             # JSON output format
gz pm update --all --check-duplicates        # Check for duplicate binaries
gz pm update --manager pip --pip-allow-conda # Allow pip in conda environment
```

**Prerequisites**:

- [ ] Package managers installed (asdf, brew, npm, etc.)
- [ ] Network connectivity for package downloads
- [ ] Admin permissions for system-wide package managers (apt, pacman)
- [ ] Sufficient disk space for package downloads and cache

### Expected Output

**Success Case - All Managers Updated**:

```text
🔄 Updating all package managers...

📋 Manager Overview:
MANAGER      SUPPORTED  INSTALLED  NOTE
------------ ---------- ---------- --------------------
brew         ✅         ✅
asdf         ✅         ✅
sdkman       ✅         ✅
npm          ✅         ✅
pip          ✅         ✅
apt          🚫         ⛔         Linux 전용
pacman       🚫         ⛔         Arch/Manjaro 계열 전용

🧪 Duplicate Installation Check:
Found 2 potential conflicts:
  • node: /usr/local/bin/node (brew), ~/.asdf/shims/node (asdf)
  • python3: /usr/bin/python3 (system), ~/.asdf/shims/python3 (asdf)

═══════════ 🚀 [1/5] brew — Updating ═══════════
🍺 Updating Homebrew...
✅ brew update: Updated 23 formulae
✅ brew upgrade: Upgraded 5 packages
   • node: 20.11.0 → 20.11.1 (24.8MB)
   • git: 2.43.0 → 2.43.1 (8.4MB)
   • python@3.11: 3.11.7 → 3.11.8 (15.2MB)
   • jq: 1.6 → 1.7 (1.1MB)
   • tree: 2.1.0 → 2.1.1 (156KB)
✅ brew cleanup: Freed 245MB disk space

═══════════ 🚀 [2/5] asdf — Updating ═══════════
🔄 Updating asdf plugins...
✅ asdf plugin update --all: 8 plugins updated
✅ asdf update: Updated to v0.14.0

Checking nodejs for updates...
✅ nodejs: 20.11.0 → 20.11.1 installed
✅ Post-action: npm cache clean --force

Checking golang for updates...
💡 golang: 1.21.5 already latest, skipping

Checking python for updates...
✅ python: 3.11.7 → 3.11.8 installed
✅ Post-action: pip install --upgrade pip

═══════════ 🚀 [3/5] sdkman — Updating ═══════════
☕ Updating SDKMAN...
✅ sdk selfupdate: Updated SDKMAN to 5.18.2
✅ sdk update: Refreshed candidate metadata
💡 Available updates:
   • java: 21.0.1-oracle → 21.0.2-oracle (use: sdk install java 21.0.2-oracle)
   • maven: 3.9.5 → 3.9.6 (use: sdk install maven 3.9.6)

═══════════ 🚀 [4/5] npm — Updating ═══════════
🧩 Updating npm global packages...
✅ npm update -g: 12 global packages updated
   • @angular/cli: 17.0.7 → 17.0.8
   • typescript: 5.3.2 → 5.3.3
   • prettier: 3.1.0 → 3.1.1
   • eslint: 8.55.0 → 8.56.0

═══════════ 🚀 [5/5] pip — Updating ═══════════
🐍 Updating pip packages...
✅ pip install --upgrade pip: Updated to 24.0
Checking for outdated packages...
✅ Updated 6 packages:
   • requests: 2.31.0 → 2.32.0
   • numpy: 1.24.3 → 1.24.4
   • pandas: 2.1.3 → 2.1.4
   • matplotlib: 3.7.2 → 3.7.3

🎉 Package manager updates completed successfully!

📊 Summary:
   • Total managers processed: 5
   • Successfully updated: 5
   • Packages upgraded: 27
   • Total download size: 52.1MB
   • Disk space freed: 245MB
   • Conflicts detected: 2 (non-blocking)

💡 Recommended actions:
   • Update asdf language versions: asdf install golang 1.21.6
   • Update SDKMAN candidates: sdk install java 21.0.2-oracle
   • Consider switching node to single manager to avoid conflicts

⏰ Update completed in 3m 42s

stderr: (empty)
Exit Code: 0
```

**Partial Success with Detailed Failures**:

```text
🔄 Updating all package managers...

═══════════ 🚀 [1/4] brew — Updating ═══════════
🍺 Updating Homebrew...
✅ brew update: Updated 15 formulae
❌ brew upgrade: Failed to upgrade 2/7 packages
   ✅ git: 2.43.0 → 2.43.1 (success)
   ✅ jq: 1.6 → 1.7 (success)
   ❌ postgresql: Version conflict detected
      • Current: 14.9 (via Homebrew)
      • Available: 16.1 (breaking changes)
      • Fix: brew unlink postgresql@14 && brew install postgresql@16
   ❌ docker: Insufficient disk space (need 1.2GB, available: 800MB)
      • Fix: brew cleanup or free disk space

═══════════ ⚠️ [2/4] sdkman — SKIP ═══════════
❌ Network error: Cannot reach SDKMAN servers
   • DNS resolution failed for get.sdkman.io
   • Timeout after 30 seconds
   • Check network connectivity and firewall settings
   • Retry: gz pm update --manager sdkman

═══════════ 🚀 [3/4] npm — Updating ═══════════
🧩 Updating npm global packages...
✅ npm update -g: 8 global packages updated

═══════════ ⚠️ [4/4] pip — SKIP ═══════════
⚠️  Conda environment detected: /opt/miniconda3/envs/myproject
   • pip updates in conda environments can cause dependency conflicts
   • Use conda/mamba for package management instead
   • Override with: gz pm update --manager pip --pip-allow-conda

⚠️  Package manager updates partially completed.

📊 Summary:
   • Total managers processed: 4
   • Successfully updated: 2
   • Failed: 1 (network issues)
   • Skipped: 1 (environment conflict)
   • Packages upgraded: 10
   • Manual fixes required: 2

🔧 Required manual fixes:
   1. PostgreSQL version conflict: brew unlink postgresql@14 && brew install postgresql@16
   2. Docker disk space: Free 400MB+ or run brew cleanup
   3. Network connectivity: Check SDKMAN server access

💡 Retry failed updates: gz pm update --managers sdkman

stderr: partial update completed with issues
Exit Code: 1
```

### Side Effects

**Files Created**:

- `~/.gzh/pm-update.log` - Detailed update log with timestamps
- `~/.gzh/pm/state/update-<timestamp>.json` - Update session results
- `~/.gzh/pm/cache/` - Package manager cache files
- `/tmp/gz-pm-*.tmp` - Temporary download and processing files

**Files Modified**:

- Package manager databases updated (brew, apt, pacman, etc.)
- Installed packages upgraded to new versions
- Package manager configuration files updated
- System PATH potentially modified (for newly installed tools)

**State Changes**:

- Package databases refreshed with latest available versions
- Outdated packages upgraded to newer versions
- Package caches cleaned and optimized
- Environment variables updated for new tool versions

### Validation

**Automated Tests**:

```bash
# Test basic update functionality
result=$(gz pm update --all --dry-run 2>&1)
exit_code=$?

assert_contains "$result" "Updating package managers"
assert_contains "$result" "dry run"

# Test specific manager update
result=$(gz pm update --manager brew --dry-run 2>&1)
assert_contains "$result" "brew"
assert_exit_code 0

# Test JSON output format
result=$(gz pm update --all --dry-run --output json 2>&1)
json_valid=$(echo "$result" | jq . >/dev/null 2>&1 && echo "valid" || echo "invalid")
assert_equals "$json_valid" "valid"

# Check log file creation
assert_file_exists "$HOME/.gzh/pm-update.log"
log_content=$(cat "$HOME/.gzh/pm-update.log")
assert_contains "$log_content" "update session"
```

### Edge Cases

**System Resource Issues**:

- Insufficient disk space during package downloads
- Network connectivity issues (DNS, firewall, proxy)
- Slow network connections with timeout handling
- Package download corruption and retry mechanisms

**Version Management Conflicts**:

- Multiple versions of same tool across different managers
- Version pinning conflicts (e.g., .node-version vs global asdf)
- Breaking changes in major version updates
- Dependency conflicts between package managers

**Environment Conflicts**:

- Conda/Mamba vs pip package management
- System Python vs asdf/pyenv Python
- Global npm vs local node_modules conflicts
- Docker container vs host package management

**Permission and Security**:

- System package managers requiring sudo (apt, pacman, yum)
- Package signature verification failures
- Firewall blocking package repositories
- Corporate proxy authentication issues

### Performance Expectations

**Response Time**:

- Manager detection: < 5 seconds
- Single manager update: 30 seconds - 5 minutes
- All managers update: 2-15 minutes (varies by packages)
- Dry-run analysis: < 30 seconds

**Resource Usage**:

- Memory: 100-500MB (varies by package count)
- CPU: Moderate during downloads, high during compilation
- Network: 10MB - 2GB+ (varies by update size)
- Disk: Temporary space for downloads, permanent for packages

## Notes

- Comprehensive multi-platform package manager support
- Intelligent conflict detection and resolution guidance
- Environment-aware updates (conda, virtual environments)
- Detailed logging and audit trail for troubleshooting
- Integration with system permissions and security
- Rollback capabilities for failed updates
- Performance monitoring and resource management
- Extensible architecture for adding new package managers
