# 🔗 Integrations Guide

Comprehensive guide for integrating gzh-cli with external tools, services, and enterprise systems.

## 📋 Table of Contents

- [Infrastructure Tools](#infrastructure-tools)
- [Enterprise Features](#enterprise-features)
- [CI/CD Integrations](#cicd-integrations)
- [Monitoring & Observability](#monitoring--observability)
- [Third-Party Services](#third-party-services)

## 📚 Integration Documentation

### Infrastructure Integration
- **[Terraform Comparison](81-terraform-comparison.md)** - Comparing gzh-cli with Terraform for infrastructure management
- **[Terraform vs gzh-cli Examples](82-terraform-vs-gz.md)** - Side-by-side examples and migration patterns
- **[Webhook Management](83-webhook-management.md)** - Git platform webhook configuration and management

### Enterprise Features
- **[Enterprise Directory](enterprise/)** - Enterprise-specific integrations and policies
  - **Actions Policy Enforcement** - GitHub Actions policy management
  - **Actions Policy Schema** - Policy configuration schema

## 🚀 Quick Integration Guide

### CI/CD Platform Integration

#### GitHub Actions
```yaml
name: gzh-cli Integration
on: [push, pull_request]

jobs:
  quality-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install gzh-cli
        run: |
          curl -L https://github.com/gizzahub/gzh-cli/releases/latest/download/gz-linux-amd64 -o gz
          chmod +x gz
          sudo mv gz /usr/local/bin/

      - name: Validate Configuration
        run: gz config validate

      - name: Run Quality Checks
        run: |
          gz quality run --output sarif --output-file quality.sarif

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: quality.sarif

      - name: Repository Compliance
        run: |
          gz repo-config audit --org ${{ github.repository_owner }} \
            --output json --output-file compliance.json
```

#### GitLab CI
```yaml
stages:
  - validate
  - quality
  - compliance

variables:
  GZH_VERSION: "latest"

.gzh-install: &gzh-install
  before_script:
    - curl -L "https://github.com/gizzahub/gzh-cli/releases/latest/download/gz-linux-amd64" -o gz
    - chmod +x gz
    - mv gz /usr/local/bin/

validate-config:
  stage: validate
  <<: *gzh-install
  script:
    - gz config validate

quality-check:
  stage: quality
  <<: *gzh-install
  script:
    - gz quality run --output json --output-file quality-report.json
  artifacts:
    reports:
      codequality: quality-report.json

compliance-check:
  stage: compliance
  <<: *gzh-install
  script:
    - gz repo-config audit --org $CI_PROJECT_NAMESPACE --output json
  only:
    - main
```

#### Jenkins Pipeline
```groovy
pipeline {
    agent any

    environment {
        GZH_CONFIG_PATH = credentials('gzh-config')
        GITHUB_TOKEN = credentials('github-token')
    }

    stages {
        stage('Install gzh-cli') {
            steps {
                sh '''
                    curl -L https://github.com/gizzahub/gzh-cli/releases/latest/download/gz-linux-amd64 -o gz
                    chmod +x gz
                    sudo mv gz /usr/local/bin/
                '''
            }
        }

        stage('Quality Check') {
            steps {
                sh 'gz quality run --output junit --output-file quality-results.xml'
                publishTestResults testResultsPattern: 'quality-results.xml'
            }
        }

        stage('Repository Sync') {
            steps {
                sh 'gz synclone github --org myorg --dry-run'
            }
        }
    }
}
```

### Container Integration

#### Docker Compose
```yaml
version: '3.8'

services:
  gzh-cli:
    image: gizzahub/gzh-cli:latest
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - GZH_CONFIG_PATH=/config/gzh.yaml
    volumes:
      - ./config:/config
      - ./repos:/repos
      - ~/.ssh:/root/.ssh:ro
    command: ["synclone", "github", "--org", "myorg"]

  gzh-quality:
    image: gizzahub/gzh-cli:latest
    volumes:
      - ./src:/src
    working_dir: /src
    command: ["quality", "run", "--output", "json"]
```

#### Kubernetes
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: gzh-sync
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: gzh-cli
            image: gizzahub/gzh-cli:latest
            env:
            - name: GITHUB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: github-credentials
                  key: token
            - name: GZH_CONFIG_PATH
              value: "/config/gzh.yaml"
            volumeMounts:
            - name: config
              mountPath: /config
            - name: repos
              mountPath: /repos
            command: ["gz", "synclone", "--config", "/config/gzh.yaml"]
          volumes:
          - name: config
            configMap:
              name: gzh-config
          - name: repos
            persistentVolumeClaim:
              claimName: repos-pvc
          restartPolicy: OnFailure
```

## 🎯 Infrastructure Integration

### Terraform Integration

While gzh-cli and Terraform serve different purposes, they can complement each other:

```hcl
# terraform/main.tf
resource "github_repository" "repos" {
  for_each = var.repositories

  name        = each.key
  description = each.value.description
  private     = each.value.private

  # Configure repository for gzh-cli management
  topics = concat(each.value.topics, ["gzh-managed"])
}

# Use gzh-cli for post-creation management
resource "null_resource" "gzh_sync" {
  depends_on = [github_repository.repos]

  provisioner "local-exec" {
    command = "gz synclone github --org ${var.github_org} --filter topics:gzh-managed"
  }
}
```

### Infrastructure as Code Integration

```yaml
# ansible-playbook.yml
- name: Setup Development Environment
  hosts: developers
  tasks:
    - name: Install gzh-cli
      get_url:
        url: "https://github.com/gizzahub/gzh-cli/releases/latest/download/gz-linux-amd64"
        dest: /usr/local/bin/gz
        mode: '0755'

    - name: Deploy gzh-cli configuration
      template:
        src: gzh.yaml.j2
        dest: /etc/gzh-cli/gzh.yaml
        mode: '0600'

    - name: Initialize repositories
      command: gz synclone --config /etc/gzh-cli/gzh.yaml
      become_user: "{{ item }}"
      loop: "{{ developers }}"
```

## 📊 Monitoring & Observability

### Prometheus Integration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'gzh-cli'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

```bash
# Enable Prometheus metrics in gzh-cli
gz profile server --prometheus-endpoint /metrics --port 9090
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "gzh-cli Metrics",
    "panels": [
      {
        "title": "Repository Sync Operations",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(gzh_synclone_operations_total[5m])",
            "legendFormat": "Sync Operations/sec"
          }
        ]
      },
      {
        "title": "Quality Check Results",
        "type": "stat",
        "targets": [
          {
            "expr": "gzh_quality_check_failures_total",
            "legendFormat": "Failed Checks"
          }
        ]
      }
    ]
  }
}
```

### Logging Integration

#### ELK Stack
```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  paths:
    - /var/log/gzh-cli/*.log
  fields:
    service: gzh-cli
  fields_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "gzh-cli-%{+yyyy.MM.dd}"
```

#### Fluentd
```conf
<source>
  @type tail
  path /var/log/gzh-cli/*.log
  pos_file /var/log/fluentd/gzh-cli.log.pos
  tag gzh-cli
  format json
</source>

<match gzh-cli>
  @type elasticsearch
  host elasticsearch
  port 9200
  index_name gzh-cli
  type_name _doc
</match>
```

## 🔐 Security Integrations

### SAST Integration

```yaml
# GitHub Actions - Security Analysis
- name: CodeQL Analysis
  uses: github/codeql-action/analyze@v2
  with:
    languages: go

- name: gzh-cli Security Scan
  run: |
    gz quality security --output sarif --output-file security.sarif

- name: Upload Security SARIF
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: security.sarif
```

### Vulnerability Scanning

```bash
# Integrate with vulnerability databases
gz quality security --deps --output json | \
  jq '.vulnerabilities[] | select(.severity == "critical")'

# SARIF output for security tools
gz quality security --output sarif --output-file security.sarif
```

## 🏢 Enterprise Integrations

### LDAP/Active Directory
```yaml
# Enterprise authentication
auth:
  ldap:
    enabled: true
    server: "ldap://company.com:389"
    bind_dn: "cn=gzh-service,ou=services,dc=company,dc=com"
    user_base: "ou=users,dc=company,dc=com"
    group_base: "ou=groups,dc=company,dc=com"
```

### SAML/SSO Integration
```yaml
# SAML configuration
auth:
  saml:
    enabled: true
    idp_url: "https://sso.company.com/saml"
    sp_cert: "/etc/ssl/certs/gzh-cli.crt"
    sp_key: "/etc/ssl/private/gzh-cli.key"
```

### Enterprise Policy Management
```yaml
# Enterprise policies
policies:
  repository_access:
    require_mfa: true
    allowed_domains: ["company.com"]
    blocked_repositories: ["*/secret-*"]

  quality_standards:
    minimum_coverage: 80
    required_checks: ["security", "lint", "test"]
    fail_on_warnings: true
```

## 📱 API Integrations

### REST API Integration

```python
# Python integration example
import requests

class GzhCliAPI:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url

    def sync_repositories(self, org):
        response = requests.post(f"{self.base_url}/api/synclone",
                               json={"org": org, "provider": "github"})
        return response.json()

    def get_quality_report(self, project_path):
        response = requests.get(f"{self.base_url}/api/quality",
                              params={"path": project_path})
        return response.json()

# Usage
api = GzhCliAPI()
result = api.sync_repositories("myorg")
quality = api.get_quality_report("/path/to/project")
```

### Webhook Integration

```javascript
// Express.js webhook handler
const express = require('express');
const app = express();

app.post('/webhook/gzh-cli', (req, res) => {
    const { event, repository, quality_score } = req.body;

    switch(event) {
        case 'repository_synced':
            console.log(`Repository ${repository} synced successfully`);
            break;
        case 'quality_check_completed':
            if (quality_score < 8.0) {
                // Trigger alerts
                notifyTeam(`Quality score ${quality_score} below threshold`);
            }
            break;
    }

    res.status(200).send('OK');
});
```

## 🔧 Custom Integrations

### Plugin System (Future)

```go
// Plugin interface
type Plugin interface {
    Name() string
    Execute(ctx context.Context, args []string) error
    Configure(config map[string]interface{}) error
}

// Custom plugin example
type CustomIntegration struct {
    config map[string]interface{}
}

func (c *CustomIntegration) Execute(ctx context.Context, args []string) error {
    // Custom integration logic
    return nil
}
```

### Extension Points

```yaml
# Configuration for custom extensions
extensions:
  pre_sync_hooks:
    - name: "backup"
      command: "backup-script.sh"
    - name: "notify"
      command: "notify-team.sh"

  post_sync_hooks:
    - name: "update_docs"
      command: "update-documentation.sh"

  quality_plugins:
    - name: "custom_linter"
      path: "/usr/local/bin/custom-linter"
      config: "custom-linter.yaml"
```

---

**Integration Types**: CI/CD, Infrastructure, Monitoring, Security, Enterprise
**Supported Platforms**: GitHub Actions, GitLab CI, Jenkins, Kubernetes
**API Integration**: REST API, Webhooks, Plugin system
**Enterprise**: LDAP, SAML, Policy management
