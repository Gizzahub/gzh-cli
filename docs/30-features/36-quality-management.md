# ✨ Code Quality Management

The `gz quality` command provides comprehensive code quality management with multi-language support, automated formatting, linting, security scanning, and CI/CD integration.

## 📋 Table of Contents

- [Overview](#overview)
- [Supported Languages](#supported-languages)
- [Command Reference](#command-reference)
- [Configuration](#configuration)
- [Quality Tools](#quality-tools)
- [CI/CD Integration](#cicd-integration)
- [Custom Rules](#custom-rules)

## 🎯 Overview

Code quality is essential for maintainable, secure, and performant software. The `gz quality` command unifies various quality tools under a single interface, providing consistent quality checks across different programming languages and projects.

### Key Features

- **Multi-Language Support** - Go, Python, JavaScript, Rust, Java, C/C++, and more
- **Unified Interface** - Single command for all quality tools
- **Automatic Tool Installation** - Install missing tools automatically
- **Configurable Rules** - Custom quality rules and standards
- **CI/CD Integration** - SARIF output for security scanning
- **Performance Monitoring** - Track quality metrics over time
- **Team Standards** - Shared quality configurations

## 🛠️ Supported Languages

### Full Support

| Language | Formatters | Linters | Security Scanners |
| ------------------------- | --------------------- | -------------------------- | ------------------ |
| **Go** | gofmt, gofumpt, gci | golangci-lint, staticcheck | gosec, govulncheck |
| **Python** | black, autopep8, yapf | pylint, flake8, mypy | bandit, safety |
| **JavaScript/TypeScript** | prettier, eslint | eslint, tslint | eslint-security |
| **Rust** | rustfmt | clippy | cargo-audit |
| **Java** | google-java-format | checkstyle, spotbugs | spotbugs-security |
| **C/C++** | clang-format | clang-tidy, cppcheck | cppcheck |

### Basic Support

- **Shell/Bash** - shellcheck, shfmt
- **YAML** - yamllint, prettier
- **JSON** - jq, prettier
- **Markdown** - markdownlint, prettier
- **Dockerfile** - hadolint

## 📖 Command Reference

### Quality Checks

Run comprehensive quality checks:

```bash
# Run all quality checks
gz quality run

# Run specific check types
gz quality run --checks format,lint,security

# Run for specific languages
gz quality run --languages go,python

# Run with auto-fix
gz quality run --fix

# Run with specific severity
gz quality run --severity error,warning
```

### Formatting

Code formatting operations:

```bash
# Format all supported files
gz quality format

# Format specific languages
gz quality format --languages go,python

# Check formatting without changes
gz quality format --check

# Format specific files
gz quality format --files src/main.go,scripts/deploy.py

# Format with specific tool
gz quality format --tool gofumpt --languages go
```

### Linting

Code linting and analysis:

```bash
# Run linters for all languages
gz quality lint

# Run specific linters
gz quality lint --tools golangci-lint,pylint

# Lint with custom configuration
gz quality lint --config .quality.yaml

# Lint with JSON output
gz quality lint --output json

# Lint specific files only
gz quality lint --files src/ tests/
```

### Security Scanning

Security vulnerability scanning:

```bash
# Run security scans
gz quality security

# Generate SARIF output for CI/CD
gz quality security --output sarif --output-file security.sarif

# Run dependency vulnerability checks
gz quality security --deps

# Scan specific paths
gz quality security --paths src/,scripts/

# High severity issues only
gz quality security --severity high,critical
```

### Tool Management

Manage quality tools:

```bash
# List available tools
gz quality tools list

# Install missing tools
gz quality tools install

# Install specific tools
gz quality tools install --tools golangci-lint,black

# Update all tools
gz quality tools update

# Check tool versions
gz quality tools versions

# Validate tool installation
gz quality tools validate
```

## ⚙️ Configuration

### Basic Configuration

Add quality management settings to your `~/.config/gzh-manager/gzh.yaml`:

```yaml
commands:
  quality:
    # Enable automatic tool installation
    auto_install_tools: true

    # Default languages to check
    default_languages: ["go", "python", "javascript"]

    # Default checks to run
    default_checks: ["format", "lint", "security"]

    # Output settings
    output:
      format: table
      verbose: false

    # Tool installation
    tools:
      install_location: "$HOME/.local/bin"
      auto_update: false

    # Global exclusions
    exclude:
      - "vendor/"
      - "node_modules/"
      - "*.pb.go"
      - "*.generated.*"
```

### Language-Specific Configuration

```yaml
commands:
  quality:
    languages:
      go:
        formatters:
          - name: gofumpt
            args: ["-extra"]
          - name: gci
            args: ["--local", "github.com/myorg"]

        linters:
          - name: golangci-lint
            config: ".golangci.yml"
            timeout: "5m"

        security:
          - name: gosec
            config: "gosec.json"
          - name: govulncheck

      python:
        formatters:
          - name: black
            args: ["--line-length", "88"]
          - name: isort

        linters:
          - name: pylint
            config: ".pylintrc"
          - name: mypy
            config: "mypy.ini"

        security:
          - name: bandit
          - name: safety

      javascript:
        formatters:
          - name: prettier
            config: ".prettierrc"

        linters:
          - name: eslint
            config: ".eslintrc.js"

        security:
          - name: eslint
            config: ".eslintrc.security.js"
```

### Project-Specific Configuration

Create `.quality.yaml` in your project root:

```yaml
# Project quality configuration
version: "1.0"

# Project-specific settings
project:
  name: "my-awesome-project"
  languages: ["go", "python", "javascript"]

# File inclusion/exclusion
files:
  include:
    - "src/**/*"
    - "scripts/**/*"
  exclude:
    - "**/*_test.go"  # Exclude from security scans
    - "vendor/"
    - "node_modules/"

# Quality standards
standards:
  # Fail on these severity levels
  fail_on: ["error", "critical"]

  # Maximum number of warnings allowed
  max_warnings: 10

  # Required code coverage (if applicable)
  min_coverage: 80

# Custom rules
rules:
  # Custom Go rules
  go:
    max_line_length: 120
    require_comments: true

  # Custom Python rules
  python:
    max_complexity: 10
    require_docstrings: true

# Tool-specific overrides
tools:
  golangci-lint:
    timeout: "10m"
    issues:
      max-issues-per-linter: 0
      max-same-issues: 0

  pylint:
    score_threshold: 8.0

  prettier:
    tab_width: 2
    use_tabs: false
```

## 🔧 Quality Tools

### Go Language Tools

```bash
# Install Go quality tools
gz quality tools install --languages go

# Use specific Go formatter
gz quality format --tool gofumpt --args "-extra"

# Run Go-specific linting
gz quality lint --tools golangci-lint --config .golangci.yml

# Go security scanning
gz quality security --tools gosec,govulncheck
```

### Python Tools

```bash
# Install Python quality tools
gz quality tools install --languages python

# Format with black and isort
gz quality format --languages python --tools black,isort

# Comprehensive Python linting
gz quality lint --languages python --tools pylint,mypy,flake8

# Python security scanning
gz quality security --tools bandit,safety
```

### JavaScript/TypeScript Tools

```bash
# Install JS/TS quality tools
gz quality tools install --languages javascript,typescript

# Format with Prettier
gz quality format --languages javascript --tools prettier

# Lint with ESLint
gz quality lint --languages javascript --tools eslint

# TypeScript type checking
gz quality lint --languages typescript --tools tsc
```

## 🚀 CI/CD Integration

### GitHub Actions Integration

```yaml
name: Code Quality
on: [push, pull_request]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install gzh-cli
        run: |
          # Install gzh-cli binary

      - name: Run Quality Checks
        run: |
          gz quality run --output sarif --output-file quality.sarif

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: quality.sarif

      - name: Quality Gate
        run: |
          gz quality validate --fail-on error,critical
```

### GitLab CI Integration

```yaml
code_quality:
  stage: test
  script:
    - gz quality run --output json --output-file quality-report.json
  artifacts:
    reports:
      codequality: quality-report.json
    expire_in: 1 week
```

### Jenkins Pipeline Integration

```groovy
pipeline {
    agent any
    stages {
        stage('Code Quality') {
            steps {
                sh 'gz quality run --output junit --output-file quality-results.xml'
                publishTestResults testResultsPattern: 'quality-results.xml'
            }
        }
    }
}
```

## 📊 Quality Metrics and Reporting

### Metrics Collection

```bash
# Generate quality metrics
gz quality metrics

# Historical quality trends
gz quality metrics --trend --days 30

# Export metrics for analysis
gz quality metrics --output csv > quality-metrics.csv

# Quality score calculation
gz quality score --weights format:0.2,lint:0.5,security:0.3
```

### Report Generation

```bash
# Generate comprehensive quality report
gz quality report --output html > quality-report.html

# Generate team quality dashboard
gz quality dashboard --team --output team-quality.html

# Create quality badge
gz quality badge --metric score --output quality-badge.svg

# Export quality data for external tools
gz quality export --format prometheus
```

### Quality Gates

```bash
# Set up quality gates
gz quality gate create --name "release-gate" \
  --max-errors 0 \
  --max-warnings 5 \
  --min-score 8.5

# Validate against quality gate
gz quality gate validate --gate release-gate

# List available gates
gz quality gate list

# Export gate results
gz quality gate validate --output json
```

## 🎨 Custom Rules and Standards

### Custom Rule Definition

Create custom quality rules in `.quality-rules.yaml`:

```yaml
# Custom quality rules
rules:
  # Custom Go rules
  go_custom:
    - id: "no-global-vars"
      message: "Global variables are not allowed"
      pattern: "^var\\s+\\w+\\s*="
      severity: "error"

    - id: "require-error-handling"
      message: "Error handling is required"
      pattern: "\\w+\\s*,\\s*err\\s*:=.*\\n(?!.*if.*err)"
      severity: "warning"

  # Custom Python rules
  python_custom:
    - id: "no-print-statements"
      message: "Print statements should not be used in production"
      pattern: "print\\s*\\("
      severity: "warning"
      exclude_files: ["test_*.py", "*_test.py"]

  # Custom JavaScript rules
  javascript_custom:
    - id: "no-console-log"
      message: "Console.log should not be used in production"
      pattern: "console\\.log\\s*\\("
      severity: "info"
      exclude_dirs: ["test/", "spec/"]
```

### Team Standards

```yaml
# Team coding standards
team_standards:
  # Code style requirements
  style:
    max_line_length: 100
    indentation: spaces
    indent_size: 2

  # Documentation requirements
  documentation:
    require_function_docs: true
    require_class_docs: true
    require_module_docs: true

  # Testing requirements
  testing:
    min_coverage: 80
    require_tests: true
    test_naming_pattern: "*_test.*"

  # Security requirements
  security:
    no_hardcoded_secrets: true
    require_input_validation: true
    max_security_warnings: 0
```

## 🔄 Integration with Other Commands

### Repository Quality on Sync

```bash
# Run quality checks after repository sync
gz synclone github --org myorg --post-sync "gz quality run"

# Quality-aware repository management
gz git repo clone-or-update repo.git && gz quality run
```

### Development Workflow Integration

```bash
# Pre-commit quality checks
gz quality run --fast --fix

# IDE integration
gz ide monitor &
gz quality monitor --on-change
```

## 📋 Output Formats

Quality commands support multiple output formats:

```bash
# SARIF for security tools
gz quality security --output sarif

# JUnit XML for CI/CD
gz quality run --output junit

# JSON for automation
gz quality metrics --output json

# CSV for analysis
gz quality report --output csv

# HTML for human review
gz quality report --output html
```

## 🆘 Troubleshooting

### Tool Installation Issues

```bash
# Check tool availability
gz quality tools check

# Force reinstall tools
gz quality tools install --force

# Debug tool installation
gz quality tools install --debug --verbose

# Check system requirements
gz quality tools system-check
```

### Quality Check Problems

```bash
# Debug quality checks
gz quality run --debug --verbose

# Check file permissions
gz quality check-permissions

# Validate configuration
gz quality config validate

# Reset quality configuration
gz quality config reset
```

### Performance Issues

```bash
# Run quality checks in parallel
gz quality run --parallel

# Exclude large files/directories
gz quality run --exclude "vendor/,node_modules/"

# Use fast mode for pre-commit
gz quality run --fast

# Profile quality tool performance
gz quality profile --output performance.json
```

______________________________________________________________________

**Supported Languages**: 15+ languages with full tool integration
**Output Formats**: SARIF, JUnit, JSON, CSV, HTML
**CI/CD**: GitHub Actions, GitLab CI, Jenkins, Azure DevOps
**Security**: Comprehensive vulnerability scanning and SAST tools
