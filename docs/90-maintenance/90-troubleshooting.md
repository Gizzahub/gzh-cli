# 🛠️ Troubleshooting Guide

Comprehensive troubleshooting guide for common gzh-cli issues, error resolution, and diagnostic procedures.

## 📋 Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Common Issues](#common-issues)
- [Error Categories](#error-categories)
- [Diagnostic Commands](#diagnostic-commands)
- [Performance Issues](#performance-issues)
- [Configuration Problems](#configuration-problems)
- [Getting Support](#getting-support)

## 🩺 Quick Diagnostics

### System Health Check

```bash
# Run comprehensive system diagnostics
gz doctor

# Quick health check
gz doctor --quick

# Detailed system information
gz doctor --detailed --output json > diagnostics.json
```

### Configuration Validation

```bash
# Validate configuration
gz config validate

# Show effective configuration
gz config show

# Test provider authentication
gz config test-auth --all
```

### Version Information

```bash
# Show version details
gz version --detailed

# Check for updates
gz version --check-updates

# Build information
gz version --build-info
```

## 🚨 Common Issues

### Installation Issues

#### Binary Not Found
```bash
# Issue: "gz: command not found"
# Check if binary is in PATH
which gz
echo $PATH

# Solution: Add to PATH or reinstall
export PATH=$PATH:$GOPATH/bin
# Or reinstall
make install
```

#### Permission Denied
```bash
# Issue: Permission denied when running gz
# Check file permissions
ls -la $(which gz)

# Solution: Fix permissions
chmod +x $(which gz)
# Or reinstall with proper permissions
sudo make install
```

#### Incompatible Architecture
```bash
# Issue: "cannot execute binary file: Exec format error"
# Check system architecture
uname -m

# Solution: Download correct architecture
# For ARM64: gz-linux-arm64
# For x86_64: gz-linux-amd64
```

### Authentication Issues

#### Invalid Token
```bash
# Issue: "authentication failed" or "401 Unauthorized"
# Check token validity
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Solution: Update token
export GITHUB_TOKEN="new_token_here"
# Or update in configuration file
```

#### Token Scope Issues
```bash
# Issue: "insufficient permissions" or "403 Forbidden"
# Check token scopes (GitHub)
curl -H "Authorization: token $GITHUB_TOKEN" -I https://api.github.com/user \
  | grep -i "x-oauth-scopes"

# Required scopes: repo, admin:org, admin:repo_hook
```

#### API Rate Limiting
```bash
# Issue: "rate limit exceeded"
# Check rate limit status
gz config test-auth --provider github --verbose

# Solution: Wait or use multiple tokens
# Configure rate limiting in config
providers:
  github:
    rate_limiting:
      requests_per_hour: 4500
```

### Repository Issues

#### Clone Failures
```bash
# Issue: Repository clone fails
# Check repository access
git clone https://github.com/owner/repo.git /tmp/test

# Check disk space
df -h

# Solution: Fix permissions or space
sudo chown -R $USER:$USER ~/repos
# Or change clone directory
```

#### Permission Denied on Repository
```bash
# Issue: "Permission denied (publickey)" or "Repository not found"
# Test SSH connectivity
ssh -T git@github.com

# Solution: Add SSH key or use HTTPS
ssh-add ~/.ssh/id_rsa
# Or configure HTTPS in git
git config --global credential.helper store
```

#### Repository Already Exists
```bash
# Issue: "directory already exists" errors
# Check existing repositories
ls -la ~/repos/

# Solution: Use different strategy
gz synclone github --org myorg --strategy reset
# Or clean existing directories
```

### Quality Check Issues

#### Tools Not Found
```bash
# Issue: "golangci-lint not found" or similar
# Check if tools are installed
which golangci-lint
which black
which prettier

# Solution: Install missing tools
gz quality tools install
# Or install specific tools
gz quality tools install --tools golangci-lint,black
```

#### Quality Check Failures
```bash
# Issue: Quality checks fail with errors
# Run individual checks
gz quality lint --verbose
gz quality format --check
gz quality security

# Solution: Fix issues or adjust configuration
gz quality format --fix
# Or exclude problematic files
```

### Network Issues

#### Connection Timeouts
```bash
# Issue: "connection timed out" or "network unreachable"
# Test connectivity
ping github.com
curl -I https://api.github.com

# Solution: Check proxy/firewall settings
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="http://proxy.company.com:8080"
```

#### SSL Certificate Issues
```bash
# Issue: "x509: certificate signed by unknown authority"
# Test SSL connection
openssl s_client -connect api.github.com:443

# Solution: Update CA certificates or skip verification
sudo update-ca-certificates
# Or configure git to skip SSL (not recommended)
git config --global http.sslVerify false
```

#### DNS Resolution Issues
```bash
# Issue: "no such host" or DNS resolution failures
# Test DNS resolution
nslookup github.com
dig github.com

# Solution: Configure DNS servers
echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf
```

## 📊 Error Categories

### Exit Codes

| Exit Code | Category | Description |
|-----------|----------|-------------|
| 0 | Success | Command completed successfully |
| 1 | General Error | General command failure |
| 2 | Configuration Error | Invalid configuration |
| 3 | Authentication Error | Authentication failure |
| 4 | Network Error | Network connectivity issues |
| 5 | File System Error | File/directory access issues |
| 10 | Validation Error | Input validation failure |
| 20 | Quality Check Failure | Code quality issues found |

### Error Messages

#### Configuration Errors
```
Error: configuration file not found
Error: invalid YAML syntax
Error: missing required field 'providers'
Error: invalid provider configuration
```

#### Authentication Errors
```
Error: authentication failed for provider 'github'
Error: token has insufficient permissions
Error: rate limit exceeded
```

#### Repository Errors
```
Error: repository not found or access denied
Error: failed to clone repository
Error: directory already exists
Error: insufficient disk space
```

#### Quality Errors
```
Error: linter not found
Error: quality check failed
Error: security vulnerability detected
Error: code formatting issues found
```

## 🔍 Diagnostic Commands

### Verbose Logging

```bash
# Enable verbose output
gz --verbose synclone github --org myorg

# Debug mode
gz --debug synclone github --org myorg

# Trace mode (very detailed)
GOTRACEBACK=all gz --debug synclone github --org myorg
```

### Configuration Debugging

```bash
# Show configuration loading process
gz config debug --trace-loading

# Show effective configuration
gz config show --expand-vars

# Validate specific configuration
gz config validate --file custom-config.yaml
```

### Network Debugging

```bash
# Test provider connectivity
gz config test-auth --provider github --verbose

# Network diagnostics
gz doctor --network

# Trace HTTP requests
GODEBUG=http2debug=1 gz synclone github --org myorg
```

### Performance Profiling

```bash
# Profile command execution
gz profile start --type cpu &
gz synclone github --org myorg
gz profile stop --analyze

# Memory usage analysis
GODEBUG=gctrace=1 gz synclone github --org myorg
```

## ⚡ Performance Issues

### Slow Repository Operations

#### Diagnosis
```bash
# Check concurrent job settings
gz config show | grep concurrent_jobs

# Monitor system resources
top -p $(pgrep gz)
iotop -p $(pgrep gz)
```

#### Solutions
```bash
# Reduce concurrent jobs
global:
  concurrent_jobs: 3

# Increase timeouts
global:
  timeout: "60m"

# Use faster strategy
providers:
  github:
    organizations:
      - name: "myorg"
        strategy: fetch  # Faster than reset
```

### High Memory Usage

#### Diagnosis
```bash
# Monitor memory usage
gz profile start --type memory &
gz synclone github --org large-org
gz profile stop --analyze

# Check system memory
free -h
```

#### Solutions
```bash
# Process repositories in smaller batches
gz synclone github --org myorg --limit 50

# Increase garbage collection frequency
GOGC=20 gz synclone github --org myorg

# Use memory-efficient settings
global:
  concurrent_jobs: 2
  batch_size: 10
```

### Network Performance

#### Diagnosis
```bash
# Test network speed
gz net-env test speed

# Monitor network usage
iftop
nethogs
```

#### Solutions
```bash
# Enable compression
providers:
  github:
    compression: true

# Use connection pooling
global:
  http_settings:
    max_idle_connections: 100
    idle_timeout: "30s"
```

## 🔧 Configuration Problems

### File Not Found

```bash
# Issue: Configuration file not found
# Check search paths
gz config show-paths

# Solution: Create configuration or specify path
mkdir -p ~/.config/gzh-manager
cp examples/gzh.yaml ~/.config/gzh-manager/gzh.yaml
# Or specify custom path
gz --config /path/to/config.yaml synclone
```

### Invalid YAML Syntax

```bash
# Issue: YAML parsing errors
# Validate YAML syntax
gz config validate

# Solution: Fix YAML issues
# Common issues:
# - Incorrect indentation (use 2 spaces)
# - Missing quotes around special characters
# - Duplicate keys
```

### Environment Variable Issues

```bash
# Issue: Environment variables not expanded
# Check variable expansion
gz config show --expand-vars

# Solution: Verify variables are set
echo $GITHUB_TOKEN
env | grep -E "(GITHUB|GITLAB)_TOKEN"

# Use proper syntax in config
providers:
  github:
    token: "${GITHUB_TOKEN}"  # Correct
    # token: "$GITHUB_TOKEN"  # Incorrect
```

## 📞 Getting Support

### Self-Help Resources

```bash
# Built-in help
gz --help
gz synclone --help

# Documentation
gz docs open  # Opens documentation website

# Examples
gz examples list
gz examples show synclone
```

### Diagnostic Information

```bash
# Generate diagnostic report
gz doctor --output json > diagnostic-report.json

# System information
gz version --detailed
uname -a
go version

# Configuration summary
gz config show --masked  # Hides sensitive data
```

### Log Files

```bash
# Default log locations
tail -f ~/.config/gzh-manager/logs/gzh.log

# Enable debug logging
gz --debug synclone 2>&1 | tee debug.log

# Rotate large log files
logrotate /etc/logrotate.d/gzh-cli
```

### Support Channels

#### GitHub Issues
- **Bug Reports**: Include diagnostic report and reproduction steps
- **Feature Requests**: Describe use case and expected behavior
- **Questions**: Use discussion board for general questions

#### Community Support
- **Documentation**: Check official documentation first
- **Examples**: Review example configurations
- **Known Issues**: Check GitHub issues for known problems

#### Enterprise Support
- **Priority Support**: Available for enterprise customers
- **Custom Integrations**: Professional services available
- **Training**: Team training and onboarding

### Creating Effective Bug Reports

#### Required Information
1. **Environment Details**
   ```bash
   gz version --detailed
   uname -a
   ```

2. **Configuration** (masked)
   ```bash
   gz config show --masked
   ```

3. **Reproduction Steps**
   - Exact commands run
   - Expected vs actual behavior
   - Error messages

4. **Diagnostic Output**
   ```bash
   gz doctor --output json
   ```

#### Example Bug Report Template
```markdown
## Bug Description
Brief description of the issue

## Environment
- gzh-cli version: (gz version)
- OS: (uname -a)
- Go version: (go version)

## Reproduction Steps
1. Run command: `gz synclone github --org myorg`
2. Expected: Repositories sync successfully
3. Actual: Error message appears

## Error Output
```
Error: authentication failed
```

## Configuration
```yaml
# Masked configuration
providers:
  github:
    token: "***masked***"
```

## Diagnostic Information
```json
{diagnostic output}
```
```

---

**Quick Fix**: Try `gz doctor` for immediate diagnostics
**Common Issues**: Authentication, configuration, network connectivity
**Debug Mode**: Use `--debug --verbose` for detailed information
**Support**: GitHub issues, documentation, community discussions
