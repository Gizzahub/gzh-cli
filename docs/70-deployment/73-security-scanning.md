# Security Scanning

This project implements security scanning using [gosec](https://github.com/securego/gosec) and the gosec analyzer integrated into golangci-lint.

## Overview

Security scanning is performed at multiple levels:

1. **Integrated with golangci-lint**: Runs as part of normal linting process
1. **Standalone gosec**: Dedicated security analysis with detailed reporting
1. **CI/CD integration**: Automated security checks in pipeline

## Configuration

### gosec Rules Enabled

The standalone scan uses the default rule set shipped by the pinned gosec release. It does not
maintain a separate include allowlist, so rules added to that pinned release remain visible. Any
accepted exclusion must be explicit and documented rather than omitted implicitly.

### Configuration Files

- **`.golangci.yml`**: Integrates gosec with the regular lint policy
- **`.gosec.json`**: Standalone gosec rule configuration
- **`.make/tools.mk`**: Pins the repository-local gosec version and scan thresholds

## Usage

### Basic Security Scan

```bash
# Run integrated security scanning with golangci-lint
make lint

# Run the standalone gosec policy
make security-code

# Generate the SARIF report uploaded by CI
# Requires jq so the report structure can be validated before publication.
make security-json
```

### Advanced Usage

```bash
# Install the pinned repository-local scanner
make install-gosec

# Prefer make security-code. This direct form is policy-equivalent only while
# every pinned flag below stays identical to GOSEC_SCAN_FLAGS.
./bin/tools/gosec -conf=.gosec.json -exclude-generated -exclude-dir=vendor -exclude-dir=node_modules -exclude-dir=.git -exclude-dir=tmp -tests -confidence=medium -severity=medium -no-fail ./...

# Policy-equivalent text diagnostics with the same pinned scan scope.
./bin/tools/gosec -conf=.gosec.json -exclude-generated -exclude-dir=vendor -exclude-dir=node_modules -exclude-dir=.git -exclude-dir=tmp -tests -confidence=medium -severity=medium -no-fail -verbose=text ./...

# Advisory partial diagnosis only: this filters rules and is not the repository policy scan.
./bin/tools/gosec -include=G101,G102,G104 ./...

# Advisory output-format demonstrations only: they omit the pinned scan flags
# and must not be used for policy or CI evidence. Use make security-json for
# the validated SARIF report.
./bin/tools/gosec -fmt=json ./...
./bin/tools/gosec -fmt=yaml ./...
./bin/tools/gosec -fmt=csv ./...
./bin/tools/gosec -fmt=junit-xml ./...
```

### Accepted Risks (the replacement for `#nosec`)

The pinned standalone scanner is gosec v2.28.0 (`GOSEC_VERSION` in
`.make/tools.mk`). Its repository configuration sets `global.nosec` to `false`,
so `#nosec` comments are inert for the standalone scan. `#nosec` is therefore
not a suppression mechanism in this repository: do not add one, and do not read
an existing one as evidence that a finding has been handled. The legacy `#nosec`
comments that used to be in the tree have been removed for that reason.

The only mechanism that suppresses a standalone result is a `gosec:disable`
directive naming a registered accepted-risk identifier:

```go
//gosec:disable G304 -- AR-2026-003 each test reads only paths it created under t.TempDir.
file, err := os.Open(configPath)
```

#### Trusted base

| File | Contents |
| ---- | -------- |
| `security/policy.yaml` | Who may approve, which evidence formats count, and the review cadence |
| `security/accepted-risks.yaml` | One immutable `AR-YYYY-NNN` record per suppressed site |
| `security/internal/acceptedrisk` | The fail-closed validator over both files and the source directives |

`security/policy.yaml` matches an approver on the **immutable GitHub numeric
user id**, never on the login, which is renameable. `type: Bot` and any
automation or agent identity can never approve. The single accepted evidence
format is `signed-commit`: a 40-character lowercase hex commit id whose
signature verifies. A GitHub Issue or pull request URL is deliberately **not**
accepted, because both remain editable after the fact.

#### Procedure

1. **Remediate first.** A suppression is the last option, not the first. Fix the
   finding, or restructure the code so the pattern no longer occurs.
1. **Register the risk.** Add a record to `security/accepted-risks.yaml` with a
   new, never-reused `AR-YYYY-NNN` identifier and every required field: `rule`,
   `path` (repository-relative), `symbol`, `threat`, `compensating_control`,
   `owner`, `approver` (`id` and `login`), `created_at`, `last_reviewed_at`,
   `evidence`, and `test_evidence` naming the tests that hold the compensating
   control in place. Leave `evidence.sha` as the sentinel
   `pending-owner-approval` until step 3; the validator rejects that sentinel, so
   an unapproved record cannot pass.
1. **Obtain owner approval.** The policy owner approves by making a signed
   commit and recording its full SHA in `evidence.sha`. Never write a SHA you
   did not obtain from a real signed approval commit.
1. **Link the suppression.** Exactly one directive references the record, and
   the record covers exactly one site. Its rule and file path must match the
   record; a second directive sharing an identifier is a violation, as is a
   record no directive references.
1. **Review and sunset.** `review_by` is derived as `last_reviewed_at + 90 days`
   and is never stored. The hard sunset is `created_at + 180 days`. Past the hard
   sunset the identifier is **not** renewable: bumping `last_reviewed_at` is
   rejected, and the risk needs a fresh analysis, a fresh approval and a new
   `AR-YYYY-NNN` identifier.

#### What the validator rejects

`go test ./security/internal/acceptedrisk/...` fails closed on an approver who
is absent from `security/policy.yaml`, an approver that is a bot or agent, a
registry login that has drifted from the policy, missing or malformed approval
evidence, a SHA that is not full lowercase hex or whose signature does not
verify, a duplicate `AR-YYYY-NNN`, a directive referencing an unregistered
identifier, a malformed directive, a record no directive references, an overdue
review, a passed hard sunset, and a renewal attempted after the hard sunset.

#### Current state

Every record in `security/accepted-risks.yaml` carries
`sha: pending-owner-approval`. The owner has approved the authority SSOT, the
evidence format and the cadence, but has not yet approved any individual
accepted risk, so the registry does not validate yet and that is the intended
state. See the
[gosec suppression inventory](../90-maintenance/94-gosec-suppression-inventory.md)
for the per-site listing.

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

### Report File Naming

`gosec-report.json` contains a SARIF 2.1.0 document, not gosec's native JSON
output. `make security-json` validates gosec's native JSON internally, then
moves the SARIF report to `gosec-report.json` and discards the native JSON.
The `.json` extension is misleading, but the name is retained deliberately:
the CI upload step above consumes the file as `sarif_file:`
(`.github/workflows/main.yml`), and renaming it would require updating that
workflow and this document together, which is out of scope here.

## Troubleshooting

### Common Issues

1. **gosec not found**

   ```bash
   make install-gosec
   ```

1. **Too many false positives**

   - Do not use `#nosec`: it is inert under `global.nosec: false`.
   - First remediate the finding or document why it is not a finding.
   - A `gosec:disable` directive requires a record in
     `security/accepted-risks.yaml` approved by the owner declared in
     `security/policy.yaml` with a verifying signed commit. Never issue an AR ID
     and mark it approved yourself.
   - Do not broaden `.gosec.json` exclusions or lower thresholds to silence a
     result without a separately approved policy change.
   - See [Accepted Risks](#accepted-risks-the-replacement-for-nosec) for the
     full procedure.

1. **Performance issues**

   - Use `--exclude-dir` to skip large directories
   - Advisory partial diagnosis only: run a specific package with
     `./bin/tools/gosec ./pkg/...`; do not treat it as a policy scan.
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
# Policy-equivalent local metrics. Prefer make security-json for CI SARIF.
./bin/tools/gosec -conf=.gosec.json -exclude-generated -exclude-dir=vendor -exclude-dir=node_modules -exclude-dir=.git -exclude-dir=tmp -tests -confidence=medium -severity=medium -no-fail -fmt=json ./... | jq '.Issues | length'
./bin/tools/gosec -conf=.gosec.json -exclude-generated -exclude-dir=vendor -exclude-dir=node_modules -exclude-dir=.git -exclude-dir=tmp -tests -confidence=medium -severity=medium -no-fail -fmt=json ./... | jq '.Issues | group_by(.severity) | map({severity: .[0].severity, count: length})'
```

## Best Practices

1. **Run security scans regularly** - Include in CI/CD pipeline
1. **Address HIGH severity issues first** - Prioritize by impact
1. **Review gosec output manually** - Don't rely solely on automation
1. **Keep security tools updated** - Regular updates catch new vulnerabilities
1. **Train team on secure coding** - Prevention is better than detection
1. **Document exceptions** - Register them in `security/accepted-risks.yaml` under an owner-approved `AR-YYYY-NNN`; `#nosec` suppresses nothing
1. **Validate user input** - Never trust external data
1. **Use principle of least privilege** - Minimal file permissions and access
1. **Regular security reviews** - Periodic manual code reviews for security
1. **Monitor for new vulnerabilities** - Subscribe to security advisories

## References

- [gosec Documentation](https://github.com/securego/gosec)
- [Go Security Checklist](https://github.com/Checkmarx/Go-SCP)
- [OWASP Go Security Guide](https://owasp.org/www-project-go-secure-coding-practices-guide/)
- [Go Vulnerability Database](https://vuln.go.dev/)
