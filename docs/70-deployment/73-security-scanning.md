# Security Scanning

This project implements comprehensive security scanning using [gosec](https://github.com/securecodewarrior/gosec) and integrated security linters in golangci-lint.

## Overview

Security scanning is performed at multiple levels:

1. **Integrated with golangci-lint**: Runs as part of normal linting process
1. **Standalone gosec**: Dedicated security analysis with detailed reporting
1. **Pre-commit hooks**: Catches security issues before commits
1. **CI/CD integration**: Automated security checks in pipeline

## Configuration

### gosec Rules Enabled

| Rule      | Description                       | Severity |
| --------- | --------------------------------- | -------- |
| G101      | Hardcoded credentials             | HIGH     |
| G102      | Bind to all interfaces            | MEDIUM   |
| G103      | Unsafe blocks                     | HIGH     |
| G104      | Unchecked errors                  | MEDIUM   |
| G106      | SSH InsecureIgnoreHostKey         | HIGH     |
| G107      | URL injection                     | MEDIUM   |
| G108      | Profiling endpoint exposed        | MEDIUM   |
| G109      | Integer overflow                  | MEDIUM   |
| G110      | DoS via decompression             | HIGH     |
| G201      | SQL injection (format)            | HIGH     |
| G202      | SQL injection (concat)            | HIGH     |
| G203      | Unescaped HTML templates          | MEDIUM   |
| G204      | Command injection                 | HIGH     |
| G301      | Poor directory permissions        | MEDIUM   |
| G302      | Poor file permissions (chmod)     | MEDIUM   |
| G303      | Predictable tempfile              | MEDIUM   |
| G304      | File path injection               | MEDIUM   |
| G305      | ZIP/TAR traversal                 | HIGH     |
| G306      | Poor file permissions (write)     | MEDIUM   |
| G307      | Deferred error not checked        | LOW      |
| G401      | Weak crypto (DES, RC4, MD5, SHA1) | HIGH     |
| G402      | Bad TLS settings                  | HIGH     |
| G403      | Weak RSA keys (\<2048 bits)       | HIGH     |
| G404      | Insecure random source            | MEDIUM   |
| G501-G505 | Crypto import blocklist           | HIGH     |
| G601      | Implicit memory aliasing          | MEDIUM   |
| G602      | Slice bounds checking             | MEDIUM   |

### Configuration Files

- **`.golang-ci.yml`**: Integrates gosec with golangci-lint
- **`.gosec.yaml`**: Standalone gosec configuration
- **`.pre-commit-config.yaml`**: Pre-commit security hooks

## Usage

### Basic Security Scan

```bash
# Run integrated security scanning with golangci-lint
make lint

# Run standalone gosec analysis
make security

# Generate JSON report
make security-json
```

### Advanced Usage

```bash
# Run gosec with custom config
gosec -config=.gosec.yaml ./...

# Run with specific rules only
gosec -include=G101,G102,G104 ./...

# Run with verbose output
gosec -verbose ./...

# Generate different output formats
gosec -fmt=json ./...        # JSON format
gosec -fmt=yaml ./...        # YAML format
gosec -fmt=csv ./...         # CSV format
gosec -fmt=junit-xml ./...   # JUnit XML
```

### Excluding False Positives

Use `#nosec` comments to exclude specific lines:

```go
// Legitimate use of command execution
cmd := exec.Command("git", "status") // #nosec G204

// Controlled file path usage
file, err := os.Open(configPath) // #nosec G304

// Deliberate weak crypto for compatibility
hash := md5.Sum(data) // #nosec G401
```

## Security Guidelines

### Credential Management

❌ **Don't do this:**

```go
const apiKey = "sk-1234567890abcdef"  // G101: Hardcoded credential
token := "github_pat_" + userInput    // G101: Potential credential leak
```

✅ **Do this:**

```go
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    return errors.New("API_KEY environment variable required")
}
```

### File Operations

❌ **Don't do this:**

```go
// G304: File path from user input
file, err := os.Open(userProvidedPath)

// G306: Overly permissive file permissions
os.WriteFile("config.json", data, 0777)
```

✅ **Do this:**

```go
// Validate and sanitize file paths
cleanPath := filepath.Clean(userProvidedPath)
if !strings.HasPrefix(cleanPath, "/safe/directory/") {
    return errors.New("invalid file path")
}

// Use appropriate file permissions
os.WriteFile("config.json", data, 0600)
```

### Command Execution

❌ **Don't do this:**

```go
// G204: Command injection vulnerability
cmd := exec.Command("sh", "-c", userInput)
```

✅ **Do this:**

```go
// Use allowlist of safe commands
allowedCommands := map[string]bool{
    "git": true, "go": true, "docker": true,
}

if !allowedCommands[command] {
    return errors.New("command not allowed")
}

// Use explicit arguments
cmd := exec.Command(command, "--help")
```

### Cryptography

❌ **Don't do this:**

```go
// G401: Weak hash function
import "crypto/md5"
hash := md5.Sum(data)

// G402: Insecure TLS config
tls.Config{InsecureSkipVerify: true}
```

✅ **Do this:**

```go
// Use strong hash functions
import "crypto/sha256"
hash := sha256.Sum256(data)

// Secure TLS configuration
tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
    },
}
```

### Error Handling

❌ **Don't do this:**

```go
// G104: Unchecked error
file.Close()
json.Unmarshal(data, &result)
```

✅ **Do this:**

```go
// Always check errors
if err := file.Close(); err != nil {
    log.Printf("Failed to close file: %v", err)
}

if err := json.Unmarshal(data, &result); err != nil {
    return fmt.Errorf("failed to unmarshal JSON: %w", err)
}
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Security Scan
  run: |
    make security-json
    # Upload results to security dashboard
    if [ -f gosec-report.json ]; then
      echo "Security issues found:"
      cat gosec-report.json
    fi
```

### Pre-commit Integration

Security scanning is automatically integrated into pre-commit hooks:

```yaml
- id: gosec
  name: Security scan Go code with gosec
  entry: gosec
  args: [-fmt=json, -out=gosec-report.json, -stdout, -verbose=text, ./...]
  language: system
  files: \.go$
```

## Troubleshooting

### Common Issues

1. **gosec not found**

   ```bash
   go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
   ```

1. **Too many false positives**

   - Use `#nosec` comments for legitimate cases
   - Update exclude rules in `.gosec.yaml`
   - Adjust confidence/severity thresholds

1. **Performance issues**

   - Use `--exclude-dir` to skip large directories
   - Run on specific packages: `gosec ./pkg/...`
   - Use parallel execution with `--parallel`

### Reporting Security Issues

If you discover a security vulnerability:

1. **Do not** create a public issue
1. Email security concerns to the maintainers
1. Include details about the vulnerability
1. Provide steps to reproduce if possible

## Metrics and Reporting

### Security Metrics

Track these metrics over time:

- Number of security issues by severity
- Time to fix security issues
- Coverage of security scanning
- False positive rate

### Integration with Metrics

```bash
# Generate metrics for reporting
gosec -fmt=json ./... | jq '.Issues | length'
gosec -fmt=json ./... | jq '.Issues | group_by(.severity) | map({severity: .[0].severity, count: length})'
```

## Best Practices

1. **Run security scans regularly** - Include in CI/CD pipeline
1. **Address HIGH severity issues first** - Prioritize by impact
1. **Review gosec output manually** - Don't rely solely on automation
1. **Keep security tools updated** - Regular updates catch new vulnerabilities
1. **Train team on secure coding** - Prevention is better than detection
1. **Document exceptions** - Use clear `#nosec` comments with explanations
1. **Validate user input** - Never trust external data
1. **Use principle of least privilege** - Minimal file permissions and access
1. **Regular security reviews** - Periodic manual code reviews for security
1. **Monitor for new vulnerabilities** - Subscribe to security advisories

## References

- [gosec Documentation](https://github.com/securecodewarrior/gosec)
- [Go Security Checklist](https://github.com/Checkmarx/Go-SCP)
- [OWASP Go Security Guide](https://owasp.org/www-project-go-secure-coding-practices-guide/)
- [Go Vulnerability Database](https://vuln.go.dev/)
