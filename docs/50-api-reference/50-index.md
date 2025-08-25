# 📋 API Reference

Complete reference documentation for all gzh-cli commands, APIs, and interfaces.

## 📖 Table of Contents

- [Command Reference](#command-reference)
- [API Documentation](#api-documentation)
- [Integration Guides](#integration-guides)
- [Quick Reference](#quick-reference)

## 🚀 Command Reference

### Core Commands

- **[Complete Command Reference](50-command-reference.md)** - All commands with options, examples, and troubleshooting

### Command Categories

#### Repository Management

```bash
# Multi-platform synchronization
gz synclone github --org kubernetes
gz synclone gitlab --group mygroup

# Single repository operations
gz git repo clone-or-update https://github.com/user/repo.git
gz git repo list --org myorg

# Repository configuration
gz repo-config audit --org myorg
gz repo-config backup --output backup.json
```

#### Code Quality & Development

```bash
# Code quality management
gz quality run
gz quality install --languages go,python
gz quality format --fix

# IDE monitoring
gz ide monitor
gz ide status
gz ide fix-sync

# Performance profiling
gz profile start --type cpu
gz profile server --port 6060
gz profile analyze profile.pprof
```

#### Environment Management

```bash
# Development environment
gz dev-env aws --profile production
gz dev-env docker context list
gz dev-env ssh keys list

# Network environment
gz net-env auto-switch
gz net-env vpn connect office
gz net-env proxy set --http proxy.company.com:8080

# Package managers
gz pm update --all
gz pm doctor --check-conflicts
```

#### System Tools

```bash
# System diagnostics
gz doctor
gz config validate
gz version --detailed
```

## 🔧 API Documentation

### Internal APIs

- **[Debug API](51-debug-api.md)** - Debugging, logging, and profiling APIs

### Provider APIs

#### Git Platform Providers

```go
// Provider interface for Git platforms
type Provider interface {
    GetRepositories(ctx context.Context, opts GetRepositoriesOptions) ([]Repository, error)
    CloneRepository(ctx context.Context, repo Repository, opts CloneOptions) error
    AuthenticateUser(ctx context.Context) (*User, error)
}

// Supported providers
// - GitHub (github.com, GitHub Enterprise)
// - GitLab (gitlab.com, self-hosted)
// - Gitea (self-hosted)
// - Gogs (self-hosted)
```

#### Configuration API

```go
// Configuration loader interface
type Loader interface {
    LoadConfig(ctx context.Context) (*Config, error)
    LoadConfigFromFile(ctx context.Context, filename string) (*Config, error)
    ValidateConfig(ctx context.Context, config *Config) error
}

// Configuration validator
type Validator interface {
    Validate(ctx context.Context, config *Config) (*ValidationResult, error)
    ValidateProvider(ctx context.Context, provider ProviderConfig) error
}
```

### Output Formats API

All commands support multiple output formats:

```go
// Output formatter interface
type OutputFormatter interface {
    FormatOutput(data interface{}) error
    SetFormat(format OutputFormat) error
    SetWriter(writer io.Writer) error
}

// Supported formats
type OutputFormat string
const (
    FormatTable  OutputFormat = "table"
    FormatJSON   OutputFormat = "json"
    FormatYAML   OutputFormat = "yaml"
    FormatCSV    OutputFormat = "csv"
    FormatHTML   OutputFormat = "html"
    FormatSARIF  OutputFormat = "sarif"
)
```

## 🔗 Integration Guides

### CI/CD Integration

#### GitHub Actions

```yaml
- name: Repository Quality Check
  run: |
    gz synclone github --org myorg --dry-run
    gz quality run --output sarif --output-file quality.sarif

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: quality.sarif
```

#### GitLab CI

```yaml
quality_check:
  script:
    - gz quality run --output json --output-file quality-report.json
  artifacts:
    reports:
      codequality: quality-report.json
```

#### Jenkins Pipeline

```groovy
stage('Quality Check') {
    steps {
        sh 'gz quality run --output junit --output-file quality-results.xml'
        publishTestResults testResultsPattern: 'quality-results.xml'
    }
}
```

### Monitoring Integration

#### Prometheus Metrics

```bash
# Enable Prometheus metrics
gz profile server --prometheus-endpoint /metrics

# Custom metrics export
gz stats --format prometheus | curl -X POST \
  --data-binary @- http://pushgateway:9091/metrics/job/gzh-cli
```

#### Grafana Dashboards

```bash
# Export metrics for Grafana
gz stats --format json | jq '.performance' > metrics.json

# Real-time monitoring
gz profile server --web-ui --port 8080
```

## ⚡ Quick Reference

### Command Syntax

```bash
# Basic syntax
gz <command> [subcommand] [options] [arguments]

# Global options (available for all commands)
--config string     Configuration file path
--output string     Output format (table|json|yaml|csv)
--verbose          Verbose logging
--debug           Debug mode
--help            Show help
--version         Show version
```

### Common Options

#### Output Options

```bash
--output table     # Human-readable table (default)
--output json      # JSON format for automation
--output yaml      # YAML format
--output csv       # CSV format for analysis
--output-file FILE # Save output to file
```

#### Concurrency Options

```bash
--concurrent N     # Number of concurrent operations
--timeout DURATION # Operation timeout
--retry N         # Number of retry attempts
```

#### Authentication Options

```bash
--token STRING    # API token
--auth-file FILE  # Authentication file
--no-auth        # Skip authentication
```

### Environment Variables

#### Authentication

```bash
# Git platform tokens
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"
export GITEA_TOKEN="your_gitea_token"
export GOGS_TOKEN="your_gogs_token"

# Configuration
export GZH_CONFIG_PATH="~/.config/gzh-manager/gzh.yaml"
export GZH_LOG_LEVEL="info"
export GZH_OUTPUT_FORMAT="table"
```

#### Network Configuration

```bash
# Proxy settings
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="http://proxy.company.com:8080"
export NO_PROXY="localhost,127.0.0.1,.company.com"
```

### Configuration Quick Reference

#### Minimal Configuration

```yaml
global:
  clone_base_dir: "$HOME/repos"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/myorg"
```

#### Multi-Platform Configuration

```yaml
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset
  concurrent_jobs: 5

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "company"
        clone_dir: "$HOME/repos/github/company"

  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "team"
        clone_dir: "$HOME/repos/gitlab/team"
```

### Exit Codes

| Exit Code | Meaning               |
| --------- | --------------------- |
| 0         | Success               |
| 1         | General error         |
| 2         | Configuration error   |
| 3         | Authentication error  |
| 4         | Network error         |
| 5         | File system error     |
| 10        | Validation error      |
| 20        | Quality check failure |

### File Locations

#### Configuration Files

```bash
~/.config/gzh-manager/gzh.yaml           # Main configuration
~/.config/gzh-manager/profiles/          # Configuration profiles
~/.config/gzh-manager/schemas/           # Configuration schemas
```

#### Data Directories

```bash
~/.config/gzh-manager/backups/          # Configuration backups
~/.config/gzh-manager/logs/             # Application logs
~/.config/gzh-manager/cache/            # Cache files
~/.config/gzh-manager/ide-backups/      # IDE configuration backups
```

#### Temporary Files

```bash
/tmp/gzh-cli/                           # Temporary files
/tmp/gzh-cli/profiles/                  # Temporary profiles
```

## 🆘 Getting Help

### Built-in Help

```bash
# General help
gz --help
gz help

# Command-specific help
gz synclone --help
gz quality run --help

# List all commands
gz help commands

# Show examples
gz examples
gz synclone examples
```

### Diagnostic Commands

```bash
# System diagnostics
gz doctor

# Configuration validation
gz config validate

# Version information
gz version --detailed

# Debug information
gz debug info
```

### Troubleshooting Resources

- **[Configuration Guide](../40-configuration/40-configuration-guide.md)** - Complete configuration reference
- **[Troubleshooting Guide](../90-maintenance/90-troubleshooting.md)** - Common issues and solutions
- **[Architecture Overview](../20-architecture/20-system-overview.md)** - System design and components

______________________________________________________________________

**Total Commands**: 50+ commands across 8 major categories
**Output Formats**: table, json, yaml, csv, html, sarif
**Supported Platforms**: GitHub, GitLab, Gitea, Gogs
**CI/CD Integration**: GitHub Actions, GitLab CI, Jenkins, Azure DevOps
