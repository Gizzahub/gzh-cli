# Output Formats and Backup Features

This document covers new features implemented in gzh-cli including advanced output formatting options and backup/restore functionality.

## Table of Contents

1. [Output Format Features](#output-format-features)
2. [Backup and Restore System](#backup-and-restore-system)
3. [Network Metrics and Analysis](#network-metrics-and-analysis)
4. [Advanced Repository Operations](#advanced-repository-operations)
5. [Quality Tool Enhancements](#quality-tool-enhancements)

## Output Format Features

### Supported Output Formats

gzh-cli now supports multiple output formats across most commands for better integration with scripts, automation tools, and data processing pipelines.

#### Available Formats

- **table** - Human-readable table format (default)
- **json** - JSON format for programmatic use
- **yaml** - YAML format for configuration export
- **csv** - CSV format for spreadsheet integration
- **html** - HTML format for web dashboards (select commands)
- **sarif** - SARIF format for security scanning integration

#### Commands with Enhanced Output Support

##### Repository Operations

```bash
# List repositories in multiple formats
gz git repo list --org myorg --output json
gz git repo list --org myorg --output csv --output-file repos.csv
gz git repo list --org myorg --output yaml

# Repository configuration audit
gz git config audit --org myorg --output json --output-file audit.json
gz git config audit --org myorg --output csv --output-file compliance.csv
gz git config audit --org myorg --output html --output-file report.html
```

##### Package Manager Status

```bash
# Package manager status in various formats
gz pm status --output json
gz pm status --output table
gz pm doctor --output yaml
```

##### Quality Analysis

```bash
# Code quality reports
gz quality run --output sarif --output-file quality.sarif
gz quality run --output json --output-file quality.json
gz quality analyze --output html --output-file quality-report.html
```

##### Network Environment

```bash
# Network metrics and topology
gz net-env status --output json
gz net-env metrics --output csv --output-file network-metrics.csv
gz net-env topology --output yaml
```

#### Output Formatting Examples

##### JSON Output
```bash
gz git repo list --org myorg --output json
```
```json
{
  "repositories": [
    {
      "name": "api-service",
      "full_name": "myorg/api-service",
      "description": "Core API service",
      "private": false,
      "fork": false,
      "archived": false,
      "language": "Go",
      "stars": 42,
      "forks": 8,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-08-19T15:45:30Z"
    }
  ],
  "total_count": 1,
  "generated_at": "2025-08-19T12:00:00Z"
}
```

##### CSV Output
```bash
gz git repo list --org myorg --output csv
```
```csv
name,full_name,description,private,language,stars,forks,created_at,updated_at
api-service,myorg/api-service,Core API service,false,Go,42,8,2024-01-15T10:30:00Z,2024-08-19T15:45:30Z
web-frontend,myorg/web-frontend,Frontend application,false,TypeScript,28,5,2024-02-01T09:15:00Z,2024-08-18T14:20:15Z
```

##### YAML Output
```bash
gz git config audit --org myorg --output yaml
```
```yaml
audit_results:
  organization: myorg
  framework: SOC2
  repositories:
    - name: api-service
      compliance_score: 85
      issues:
        - severity: medium
          category: branch_protection
          message: "Require pull request reviews not enforced"
        - severity: low
          category: vulnerability_alerts
          message: "Dependabot alerts not enabled"
  summary:
    total_repositories: 1
    compliant_repositories: 0
    average_score: 85
    critical_issues: 0
    high_issues: 0
    medium_issues: 1
    low_issues: 1
generated_at: "2025-08-19T12:00:00Z"
```

## Backup and Restore System

### Development Environment Configuration Backup

The dev-env command now includes comprehensive backup and restore functionality for various development tools and cloud provider configurations.

#### Supported Services

- **AWS** - AWS CLI configuration, credentials, profiles
- **GCP** - gcloud configuration, service accounts, projects
- **Azure** - Azure CLI configuration, subscriptions
- **Docker** - Docker configuration, contexts, registries
- **Kubernetes** - kubeconfig files, contexts, cluster configurations
- **SSH** - SSH keys, configurations, known hosts

#### Basic Usage

##### Save Configurations

```bash
# Save AWS configuration
gz dev-env aws save --name production-aws --description "Production AWS setup"

# Save GCP configuration
gz dev-env gcp save --name staging-gcp --description "Staging GCP project"

# Save Docker configuration
gz dev-env docker save --name local-dev --description "Local development Docker setup"

# Save Kubernetes configuration
gz dev-env k8s save --name cluster-config --description "Production cluster config"
```

##### Load Configurations

```bash
# Load saved AWS configuration
gz dev-env aws load --name production-aws

# Load with force override
gz dev-env aws load --name staging-aws --force

# List available configurations
gz dev-env aws list

# Load specific configuration file
gz dev-env gcp load --name production-gcp --config-path ~/.gcp/custom-config
```

#### Advanced Backup Features

##### Metadata and Organization

Each saved configuration includes metadata:

```json
{
  "description": "Production AWS setup for microservices",
  "saved_at": "2025-08-19T12:00:00Z",
  "source_path": "/home/user/.aws/config",
  "checksum": "sha256:abc123...",
  "size": 2048,
  "version": "1.0"
}
```

##### List and Manage Backups

```bash
# List all saved configurations
gz dev-env aws list

# List with detailed information
gz dev-env aws list --all

# Output as JSON for scripting
gz dev-env aws list --output json
```

Example output:
```
NAME              DESCRIPTION                    SAVED AT              SIZE
production-aws    Production AWS setup          2025-08-19 12:00:00   2.1KB
staging-aws       Staging environment           2025-08-18 15:30:00   1.8KB
dev-local         Local development setup       2025-08-17 09:15:00   1.2KB
```

##### Configuration Storage

Configurations are stored in organized directories:

```
~/.gz/
├── aws-configs/
│   ├── production-aws/
│   │   ├── config
│   │   ├── credentials
│   │   └── metadata.json
│   └── staging-aws/
│       ├── config
│       ├── credentials
│       └── metadata.json
├── gcp-configs/
│   └── staging-gcp/
│       ├── config
│       └── metadata.json
└── docker-configs/
    └── local-dev/
        ├── config.json
        └── metadata.json
```

### Package Manager State Backup

The package manager commands include backup functionality for environment states:

```bash
# Backup current package manager state
gz pm backup --name before-upgrade --description "State before major upgrade"

# Restore previous state
gz pm restore --name before-upgrade

# List backup states
gz pm backup list

# Create automatic backup before updates
gz pm update --all --auto-backup
```

## Network Metrics and Analysis

### Network Environment Monitoring

Enhanced network environment commands with comprehensive metrics and analysis:

#### Network Topology Analysis

```bash
# Analyze network topology
gz net-env topology --output json --output-file network-topology.json

# Container network analysis
gz net-env container-detection --output yaml

# Kubernetes service mesh analysis
gz net-env k8s service-mesh --output json
```

#### Network Metrics Collection

```bash
# Collect network metrics
gz net-env metrics --period 24h --output csv --output-file metrics.csv

# Real-time network monitoring
gz net-env monitor --follow --output json

# Network performance analysis
gz net-env analyze --baseline baseline.json --output html
```

#### VPN and Connection Management

```bash
# VPN profile management
gz net-env vpn profile list --output json
gz net-env vpn failover --config failover-config.yaml

# Optimal routing analysis
gz net-env routing optimize --output yaml
```

## Advanced Repository Operations

### Enhanced Repository Management

#### Repository Creation and Deletion

```bash
# Create repository with templates
gz git repo create --org myorg --name new-service --template go-microservice

# Delete repository with backup
gz git repo delete --org myorg --name old-service --backup --confirm
```

#### Webhook Management

```bash
# Bulk webhook operations
gz git webhook bulk create --org myorg --config webhooks.yaml --output json

# Webhook monitoring
gz git webhook monitor --org myorg --follow --output json
```

#### Event Processing

```bash
# Real-time event monitoring
gz git event list --org myorg --follow --output json

# Event metrics and analytics
gz git event metrics --org myorg --period 7d --output dashboard
```

## Quality Tool Enhancements

### Advanced Quality Analysis

#### Multi-format Reporting

```bash
# Generate comprehensive quality reports
gz quality run --output sarif --output-file security.sarif
gz quality analyze --output html --output-file quality-dashboard.html

# Quality metrics tracking
gz quality metrics --baseline baseline.json --output json
```

#### Tool Configuration Management

```bash
# Quality tool configuration
gz quality init --languages go,python,javascript --template strict

# Tool installation with versioning
gz quality install --version golangci-lint@1.54.2,black@23.7.0 --output json
```

## Integration Examples

### CI/CD Pipeline Integration

#### GitHub Actions

```yaml
name: Quality and Compliance Check
on: [push, pull_request]

jobs:
  quality-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install gz
        run: |
          # Install gz binary
          
      - name: Quality Analysis
        run: |
          gz quality run --output sarif --output-file quality.sarif
          
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: quality.sarif
          
      - name: Repository Compliance Check
        run: |
          gz git config audit --org ${{ github.repository_owner }} \
            --framework SOC2 --output json --output-file compliance.json
```

#### GitLab CI

```yaml
quality-check:
  stage: test
  script:
    - gz quality run --output json --output-file quality.json
    - gz git config audit --org $CI_PROJECT_NAMESPACE --output csv --output-file audit.csv
  artifacts:
    reports:
      junit: quality.json
    paths:
      - audit.csv
    expire_in: 1 week
```

### Monitoring and Dashboards

#### Prometheus Metrics Export

```bash
# Export metrics for Prometheus
gz net-env metrics --output prometheus --port 9090
gz git event metrics --output prometheus --endpoint /metrics
```

#### Dashboard Generation

```bash
# Generate HTML dashboards
gz quality analyze --output dashboard --template corporate
gz git config audit --org myorg --output dashboard --refresh-interval 5m
```

## Configuration

### Output Format Configuration

Global output format preferences can be set in the configuration file:

```yaml
# gzh.yaml
global:
  default_output_format: json
  output_settings:
    json:
      pretty: true
      timestamp: true
    csv:
      headers: true
      delimiter: ","
    yaml:
      indent: 2
      
# Command-specific defaults
commands:
  git:
    config:
      audit:
        default_output: csv
        default_output_file: "audit-{{.Date}}.csv"
  quality:
    run:
      default_output: sarif
      default_output_file: "quality-{{.Date}}.sarif"
```

### Backup Configuration

```yaml
# gzh.yaml
dev_env:
  backup:
    enabled: true
    auto_backup: true
    backup_location: "$HOME/.gz/backups"
    retention_days: 30
    compress: true
    
  aws:
    default_profile: default
    backup_profiles: true
    backup_credentials: true
    
  docker:
    backup_contexts: true
    backup_registries: true
```

## Best Practices

### Output Format Selection

1. **table** - Interactive use, human readability
2. **json** - API integration, programmatic processing
3. **yaml** - Configuration export, GitOps workflows
4. **csv** - Data analysis, spreadsheet import
5. **html** - Dashboards, reporting, sharing
6. **sarif** - Security scanning, CI/CD integration

### Backup Strategy

1. **Regular Backups**: Use auto-backup features for critical configurations
2. **Descriptive Names**: Use meaningful names and descriptions for backups
3. **Version Control**: Consider version controlling backup configurations
4. **Testing**: Regularly test restore procedures
5. **Retention**: Set appropriate retention policies for backup storage

### Integration Patterns

1. **CI/CD Integration**: Use appropriate output formats for automation
2. **Monitoring**: Leverage JSON/CSV outputs for metrics collection
3. **Reporting**: Use HTML outputs for stakeholder communication
4. **Compliance**: Use structured formats for audit trails

## Troubleshooting

### Output Format Issues

```bash
# Debug output formatting
gz git repo list --org myorg --output json --debug

# Validate output file permissions
gz quality run --output-file results.json --verbose

# Check available formats for a command
gz git config audit --help | grep -A 10 "output formats"
```

### Backup and Restore Issues

```bash
# List available backups
gz dev-env aws list --all

# Validate backup integrity
gz dev-env aws validate --name production-aws

# Debug restore operation
gz dev-env aws load --name production-aws --debug --dry-run
```

## See Also

- [Complete Command Reference](../40-api-reference/40-command-reference.md)
- [Configuration Guide](../30-configuration/30-configuration-guide.md)
- [Repository Management](21-repository-management.md)
- [Network Environment Management](../03-core-features/network-management/)