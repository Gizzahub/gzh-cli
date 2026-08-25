# Security Scanning

This project implements security scanning using [gosec](https://github.com/securego/gosec) and the gosec analyzer integrated into golangci-lint.

## Overview

Security scanning is performed at multiple levels:

1. **Integrated with golangci-lint**: Runs as part of normal linting process
1. **Standalone gosec**: Dedicated security analysis with detailed reporting
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
| G307      | Insecure `os.Create` permissions  | MEDIUM   |
| G401      | Weak crypto (DES, RC4, MD5, SHA1) | HIGH     |
| G402      | Bad TLS settings                  | HIGH     |
| G403      | Weak RSA keys (\<2048 bits)       | HIGH     |
| G404      | Insecure random source            | MEDIUM   |
| G501-G505 | Crypto import blocklist           | HIGH     |
| G601      | Implicit memory aliasing          | MEDIUM   |
| G602      | Slice bounds checking             | MEDIUM   |

### Configuration Files

- **`.golangci.yml`**: Integrates gosec with the regular lint policy
- **`.gosec.json`**: Standalone gosec rule configuration
- **`.make/tools.mk`**: Pins the repository-local gosec version and strict rule set

## Usage

### Basic Security Scan

```bash
# Run integrated security scanning with golangci-lint
make lint

# Run the standalone gosec policy
make security-code

# Generate the SARIF report uploaded by CI
make security-json
```

### Advanced Usage

```bash
# Install the pinned repository-local scanner
make install-gosec

# Run gosec with the repository config
./bin/tools/gosec -conf=.gosec.json ./...

# Run with specific rules only
./bin/tools/gosec -include=G101,G102,G104 ./...

# Run with verbose output
./bin/tools/gosec -verbose=text ./...

# Generate different output formats
./bin/tools/gosec -fmt=json ./...        # JSON format
./bin/tools/gosec -fmt=yaml ./...        # YAML format
./bin/tools/gosec -fmt=csv ./...         # CSV format
./bin/tools/gosec -fmt=junit-xml ./...   # JUnit XML
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

## Troubleshooting

### Common Issues

1. **gosec not found**

   ```bash
   make install-gosec
   ```

1. **Too many false positives**

   - Use `#nosec` comments for legitimate cases
   - Update exclude rules in `.gosec.json`
   - Adjust confidence/severity thresholds

1. **Performance issues**

   - Use `--exclude-dir` to skip large directories
   - Run on specific packages: `./bin/tools/gosec ./pkg/...`
   - Tune concurrency with `-concurrency`

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
./bin/tools/gosec -fmt=json ./... | jq '.Issues | length'
./bin/tools/gosec -fmt=json ./... | jq '.Issues | group_by(.severity) | map({severity: .[0].severity, count: length})'
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

- [gosec Documentation](https://github.com/securego/gosec)
- [Go Security Checklist](https://github.com/Checkmarx/Go-SCP)
- [OWASP Go Security Guide](https://owasp.org/www-project-go-secure-coding-practices-guide/)
- [Go Vulnerability Database](https://vuln.go.dev/)
