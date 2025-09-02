# Command: gz doctor

## Scenario: Comprehensive system health diagnosis

### Input

**Command**:
```bash
gz doctor
```

**Prerequisites**:

- [ ] gzh-cli binary installed
- [ ] System permissions for diagnostic checks

### Expected Output

**Success Case (Healthy System)**:
```text
🔍 Running system diagnostics...

✅ System Health Check
   OS: Linux 6.6.87.2-microsoft-standard-WSL2
   Architecture: x86_64
   Shell: /bin/bash

✅ Git Configuration
   Git version: 2.34.1
   User name: John Doe
   User email: john@example.com
   Global config: ~/.gitconfig

✅ Development Tools
   Go: 1.21.5 (/usr/local/go/bin/go)
   Node.js: 20.11.1 (/usr/bin/node)
   Python: 3.11.0 (/usr/bin/python3)
   Docker: 24.0.7 (running)

✅ Network Connectivity
   GitHub API: reachable (200 OK)
   GitLab API: reachable (200 OK)
   Package registries: accessible

✅ Authentication
   GITHUB_TOKEN: configured
   SSH keys: 2 keys loaded
   GPG signing: configured

🎉 All systems operational!

stderr: (empty)
Exit Code: 0
```

**Warning Case (Issues Found)**:
```text
🔍 Running system diagnostics...

✅ System Health Check
   OS: macOS 14.2.1
   Architecture: arm64
   Shell: /bin/zsh

⚠️  Git Configuration
   Git version: 2.34.1
   User name: John Doe
   ❌ User email: not configured
   💡 Fix: git config --global user.email "your@email.com"

✅ Development Tools
   Go: 1.21.5 (/opt/homebrew/bin/go)
   ❌ Node.js: not found
   💡 Install: brew install node
   Python: 3.11.0 (/opt/homebrew/bin/python3)
   ❌ Docker: not running

⚠️  Authentication
   ❌ GITHUB_TOKEN: not configured
   💡 Set: export GITHUB_TOKEN="your_token"
   SSH keys: 1 key loaded
   ✅ GPG signing: configured

⚠️  Issues found. See recommendations above.

stderr: (empty)
Exit Code: 1
```

**Critical Error Case**:
```text
🔍 Running system diagnostics...

❌ System Health Check
   OS: Windows 11
   Architecture: x86_64
   Shell: PowerShell 7.3.0

❌ Git Configuration
   ❌ Git: not found in PATH
   💡 Install Git: https://git-scm.com/download/windows

❌ Network Connectivity
   ❌ GitHub API: connection timeout
   💡 Check network connection and firewall settings

❌ Critical issues detected! Please resolve before using gzh-cli.

stderr: (empty)  
Exit Code: 2
```

### Side Effects

**Files Created**: 
- `~/.gzh/doctor-report.json` - Detailed diagnostic report

**Files Modified**: None
**State Changes**: None (read-only diagnostic)

### Validation

**Automated Tests**:
```bash
# Test basic doctor run
result=$(gz doctor 2>&1)
exit_code=$?

assert_contains "$result" "Running system diagnostics"
assert_contains "$result" "System Health Check"
# Exit code can be 0, 1, or 2 depending on system state

# Test report generation
assert_file_exists "$HOME/.gzh/doctor-report.json"
report_content=$(cat "$HOME/.gzh/doctor-report.json")
assert_contains "$report_content" '"timestamp":'
assert_contains "$report_content" '"system":'
```

**Manual Verification**:
1. Run on healthy system - should show all green checks
2. Run without required tools - should show warnings
3. Run without network - should show connectivity errors
4. Verify recommendations are actionable

### Edge Cases

**Missing Dependencies**:
- Git not installed or not in PATH
- Required development tools missing
- Docker not running or not accessible

**Network Issues**:
- Offline environment (no internet)
- Corporate firewall blocking APIs
- DNS resolution problems
- Proxy configuration issues

**Permission Problems**:
- Cannot access configuration files
- Cannot write diagnostic report
- Restricted system information access

**Platform Differences**:
- Windows PowerShell vs Command Prompt vs WSL
- macOS with Homebrew vs system tools
- Linux distributions with different package managers
- Container environments

### Performance Expectations

**Response Time**:
- Complete diagnostic: < 10 seconds
- Network checks: < 5 seconds per endpoint
- Tool detection: < 2 seconds

**Resource Usage**:
- Memory: < 100MB
- Network: Minimal (API health checks only)
- Disk: < 1MB for diagnostic report

## Notes

- Exit codes indicate severity: 0=healthy, 1=warnings, 2=critical
- Diagnostic report saved for troubleshooting
- Platform-specific recommendations provided
- Network connectivity tests are optional (can run offline)
- Integration with system package managers for recommendations
