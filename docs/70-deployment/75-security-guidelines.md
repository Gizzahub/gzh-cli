# Security Policy

## Overview

The gzh-cli project takes security seriously and implements comprehensive security measures across development, deployment, and runtime operations. This document outlines our security practices, vulnerability reporting procedures, and integrated security tooling.

## Supported Versions

We actively maintain security updates for the following versions:

| Version | Supported | Go Version | Security Features |
| ------- | ------------------ | ---------- | ------------------- |
| 1.x.x | :white_check_mark: | Go 1.22.0+ | Full security suite |
| 0.x.x | :x: | Go 1.21+ | Legacy (archived) |

## Security Architecture

### Core Security Principles

1. **Secure by Default**: All features implement security best practices by default
1. **Least Privilege**: Commands operate with minimal required permissions
1. **Defense in Depth**: Multiple layers of security controls
1. **Zero Trust**: Validate all inputs and authenticate all operations
1. **Privacy First**: No unnecessary data collection or transmission

### Authentication & Authorization

#### Token-Based Authentication

All Git platform integrations use secure token-based authentication:

```bash
# Environment variable (recommended)
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"
export GITEA_TOKEN="xxxxxxxxxxxx"

# Secure token storage
gz config set-token github --token-file ~/.config/gzh-manager/github.token
```

#### Supported Authentication Methods

- **GitHub**: Personal Access Tokens (PAT), GitHub Apps
- **GitLab**: Personal Access Tokens, Deploy Tokens
- **Gitea**: Access Tokens, OAuth2
- **Enterprise**: SAML, LDAP integration (planned)

#### Token Security Features

- Automatic token validation and expiration checking
- Secure in-memory token handling (no disk storage)
- Token scope validation (minimum required permissions)
- Rate limit aware token rotation

### Code Security Integration

#### Static Analysis Tools

Comprehensive static analysis integrated into development workflow:

**Go Security Tools**:

```bash
# Integrated via golangci-lint
gosec                    # Security-focused static analysis
goconst                  # Hardcoded constant detection
ineffassign              # Ineffectual assignment detection
unconvert                # Unnecessary type conversions
```

**Multi-Language Security** (via `gz quality`):

```bash
# Python security
bandit                   # Security linter for Python
safety                   # Dependency vulnerability checker

# JavaScript/TypeScript security
eslint-plugin-security   # Security rules for ESLint
audit                    # npm/yarn security audit

# General
secretlint               # Secret detection across languages
gitleaks                 # Git history secret scanning
```

#### Dependency Security

**Go Modules Security**:

```bash
# Automated dependency vulnerability scanning
go list -json -deps ./... | gz quality scan-deps
govulncheck ./...        # Official Go vulnerability scanner

# Dependency analysis
go mod graph | gz quality analyze-deps
```

**Supply Chain Security**:

- SLSA (Supply-chain Levels for Software Artifacts) compliance
- Signed releases with GPG verification
- Reproducible builds with checksums
- Dependency license compliance checking

### Runtime Security

#### Secure File Operations

All file operations implement security best practices:

```go
// Example: Secure configuration file handling
func LoadConfig(path string) (*Config, error) {
    // Validate file path (prevent directory traversal)
    cleanPath := filepath.Clean(path)
    if !strings.HasPrefix(cleanPath, allowedConfigDir) {
        return nil, ErrInvalidConfigPath
    }

    // Check file permissions (owner read/write only)
    if err := validateFilePermissions(cleanPath, 0600); err != nil {
        return nil, err
    }

    // Secure file reading with size limits
    content, err := secureReadFile(cleanPath, maxConfigSize)
    if err != nil {
        return nil, err
    }

    return parseConfig(content)
}
```

#### Input Validation

Comprehensive input validation for all user inputs:

- **URL Validation**: Strict URL parsing with allowlist validation
- **Path Validation**: Directory traversal prevention
- **Token Validation**: Format and scope verification
- **Configuration Validation**: Schema-based validation with sanitization

#### Secure Communication

All network communications use secure protocols:

- **HTTPS Only**: All HTTP communications use TLS 1.3
- **Certificate Validation**: Strict certificate chain validation
- **API Rate Limiting**: Built-in rate limiting with exponential backoff
- **Request Signing**: HMAC-SHA256 signing for webhook payloads

### Data Protection

#### Sensitive Data Handling

**In-Memory Security**:

```go
// Secure string handling for tokens
type SecureString struct {
    data []byte
    mu   sync.RWMutex
}

func (s *SecureString) Clear() {
    s.mu.Lock()
    defer s.mu.Unlock()

    // Zero out memory
    for i := range s.data {
        s.data[i] = 0
    }
    s.data = nil
}
```

**Configuration File Security**:

- Configuration files use 0600 permissions (owner read/write only)
- Sensitive data encrypted at rest using AES-256-GCM
- Configuration validation prevents injection attacks
- Secure defaults for all security-related settings

#### Logging Security

Secure logging implementation prevents information disclosure:

```go
// Security-aware logging
func (l *SecureLogger) Log(level Level, msg string, fields ...Field) {
    // Sanitize sensitive data from logs
    sanitizedFields := make([]Field, 0, len(fields))
    for _, field := range fields {
        if isSensitiveField(field.Key) {
            sanitizedFields = append(sanitizedFields, Field{
                Key:   field.Key,
                Value: "[REDACTED]",
            })
        } else {
            sanitizedFields = append(sanitizedFields, field)
        }
    }

    l.logger.Log(level, msg, sanitizedFields...)
}
```

### Infrastructure Security

#### Container Security

When deployed in containerized environments:

```dockerfile
# Multi-stage build for minimal attack surface
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache ca-certificates git
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o gz

# Minimal runtime container
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/gz /gz

# Non-root user
USER 65534:65534
ENTRYPOINT ["/gz"]
```

#### Kubernetes Security

When deployed on Kubernetes:

```yaml
# Security context
securityContext:
  runAsNonRoot: true
  runAsUser: 65534
  runAsGroup: 65534
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL

# Network policies
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gzh-manager-network-policy
spec:
  podSelector:
    matchLabels:
      app: gzh-manager
  policyTypes:
  - Ingress
  - Egress
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443  # HTTPS only
```

## Vulnerability Management

### Vulnerability Scanning

Automated vulnerability scanning integrated into CI/CD:

```bash
# Development workflow
make security-scan       # Run all security scans
make vuln-check         # Go vulnerability database check
make deps-audit         # Dependency vulnerability audit
make secrets-scan       # Secret detection scan

# CI/CD Integration
gz quality run --security-only --format sarif > security-results.sarif
```

### Vulnerability Response Process

1. **Detection**: Automated scanning identifies vulnerabilities
1. **Assessment**: Security team evaluates impact and severity
1. **Prioritization**: Vulnerabilities classified using CVSS v3.1
1. **Remediation**: Patches developed and tested
1. **Deployment**: Security updates released with advisory
1. **Verification**: Post-deployment verification of fixes

### Security Advisory Process

**Severity Levels**:

- **Critical**: CVSS 9.0-10.0 (24-hour response)
- **High**: CVSS 7.0-8.9 (7-day response)
- **Medium**: CVSS 4.0-6.9 (30-day response)
- **Low**: CVSS 0.1-3.9 (90-day response)

## Security Tooling Integration

### Integrated Security Commands

The `gz quality` command provides comprehensive security tooling:

```bash
# Security-focused quality checks
gz quality run --security-only
gz quality check --severity high
gz quality analyze --security-report
gz quality scan --secrets --dependencies --code

# Security tool management
gz quality install gosec bandit
gz quality upgrade --security-tools
gz quality version --security-tools
```

### IDE Security Integration

The `gz ide` command includes security-aware features:

```bash
# Monitor for security-related configuration changes
gz ide monitor --security-alerts

# Detect insecure IDE configurations
gz ide fix-sync --security-check

# Security-focused IDE auditing
gz ide list --security-audit
```

### Performance Profiling Security

The `gz profile` command implements secure profiling:

```bash
# Secure profiling with authentication
gz profile server --auth-token $PROFILE_TOKEN --bind-localhost

# CPU profiling with data protection
gz profile cpu --duration 30s --secure-output

# Memory profiling with sanitization
gz profile memory --sanitize-addresses
```

## Compliance & Standards

### Compliance Frameworks

The project supports multiple compliance frameworks:

**SOC 2 Type II**:

- Comprehensive logging and audit trails
- Access control and authentication
- Data protection and encryption
- Incident response procedures

**ISO 27001**:

- Information security management system
- Risk assessment and management
- Security policy enforcement
- Continuous monitoring and improvement

**NIST Cybersecurity Framework**:

- Identify: Asset inventory and risk assessment
- Protect: Access controls and data protection
- Detect: Continuous monitoring and alerting
- Respond: Incident response procedures
- Recover: Business continuity planning

### Security Attestations

**SLSA (Supply-chain Levels for Software Artifacts)**:

- Level 3 compliance for build process
- Signed provenance for all releases
- Build environment isolation
- Non-falsifiable build metadata

**FIPS 140-2 Compliance** (Enterprise):

- FIPS-validated cryptographic modules
- Secure key management
- Hardware security module integration

## Incident Response

### Security Incident Reporting

**Internal Reporting**:

```bash
# Automated incident detection
gz doctor --security-audit --report-incidents
gz profile monitor --security-alerts

# Manual incident reporting
gz security report --type [vulnerability|breach|policy-violation]
```

**External Reporting**:

- Email: <security@gizzahub.com>
- GPG Key: Available at <https://gizzahub.com/security.asc>
- Response SLA: 24 hours for critical, 48 hours for others

### Incident Response Process

1. **Detection & Analysis**

   - Automated monitoring alerts
   - Manual vulnerability reports
   - Threat intelligence integration

1. **Containment & Eradication**

   - Immediate threat isolation
   - Root cause analysis
   - Vulnerability patching

1. **Recovery & Post-Incident**

   - System restoration
   - Monitoring enhancement
   - Lessons learned documentation

## Security Best Practices

### For Users

**Token Management**:

```bash
# Use environment variables (recommended)
export GITHUB_TOKEN="$(cat ~/.config/gzh-manager/github.token)"

# Rotate tokens regularly
gz config rotate-token github --expiry 90d

# Validate token permissions
gz config validate-token --scope repo:read
```

**Secure Configuration**:

```yaml
# ~/.config/gzh-manager/config.yaml
security:
  strict_tls: true
  validate_certificates: true
  max_redirects: 3
  timeout: 30s

logging:
  level: info
  sanitize_tokens: true
  max_log_size: 100MB
```

### For Developers

**Secure Development Practices**:

```bash
# Pre-commit security checks
make pre-commit-install    # Install security hooks
make security-test         # Run security test suite
make vuln-scan            # Vulnerability scanning

# Code review checklist
gz quality check --security-checklist
```

**Security Testing**:

```go
// Example: Security-focused unit test
func TestTokenSanitization(t *testing.T) {
    logger := NewSecureLogger()
    token := "ghp_sensitive_token_123"

    logger.Info("Processing request", "token", token)

    // Verify token is not in logs
    logs := logger.GetLogs()
    assert.NotContains(t, logs, token)
    assert.Contains(t, logs, "[REDACTED]")
}
```

## Security Configuration

### Environment Variables

```bash
# Security-related environment variables
export GZH_SECURITY_STRICT_MODE=true
export GZH_TLS_VERIFY=true
export GZH_LOG_LEVEL=info
export GZH_MAX_FILE_SIZE=100MB
export GZH_TIMEOUT=30s

# Enterprise security features
export GZH_FIPS_MODE=true
export GZH_HSM_ENABLED=true
export GZH_AUDIT_LOG_PATH=/var/log/gzh-manager/audit.log
```

### Configuration File Security

```yaml
# ~/.config/gzh-manager/security.yaml
security:
  # Strict security mode
  strict_mode: true

  # TLS configuration
  tls:
    min_version: "1.3"
    verify_certificates: true
    ca_bundle_path: "/etc/ssl/certs/ca-certificates.crt"

  # Authentication
  auth:
    token_validation: true
    session_timeout: "24h"
    max_login_attempts: 3

  # File operations
  files:
    max_size: "100MB"
    allowed_extensions: [".yaml", ".yml", ".json"]
    secure_permissions: true

  # Network security
  network:
    allowed_hosts: []  # Empty = allow all HTTPS
    max_redirects: 3
    timeout: "30s"
    user_agent: "gzh-manager/1.0.0"

  # Logging security
  logging:
    sanitize_sensitive_data: true
    max_log_size: "100MB"
    retention_days: 90
    audit_enabled: true
```

## Contact & Support

### Security Team

- **Security Email**: <security@gizzahub.com>
- **GPG Key ID**: Available at <https://gizzahub.com/security.asc>
- **Security Advisories**: <https://github.com/gizzahub/gzh-cli/security/advisories>

### Response Times

- **Critical Security Issues**: 24 hours
- **High Severity Issues**: 48 hours
- **Medium/Low Severity Issues**: 5 business days
- **General Security Questions**: 5 business days

### Bug Bounty Program

We operate a responsible disclosure program:

- **Scope**: Latest stable release of gzh-cli
- **Rewards**: Recognition and potential monetary rewards
- **Process**: Report via <security@gizzahub.com>
- **Timeline**: 90-day coordinated disclosure

______________________________________________________________________

**Last Updated**: January 2025
**Version**: 1.0.0
**Next Review**: July 2025

For the most current security information, visit: <https://gizzahub.com/security>
