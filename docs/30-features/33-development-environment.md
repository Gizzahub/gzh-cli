# 🛠️ Development Environment Management

The `gz dev-env` command provides development environment management through both individual service control and unified environment operations, including cloud profiles, containerization, SSH configurations, and development tools.

## 📋 Table of Contents

- [Overview](#overview)
- [Cloud Platform Management](#cloud-platform-management)
- [Container Management](#container-management)
- [SSH Configuration](#ssh-configuration)
- [Development Tools](#development-tools)
- [Environment Synchronization](#environment-synchronization)
- [Best Practices](#best-practices)

## 🎯 Overview

Modern development often requires managing multiple environments, cloud profiles, and development tools. The `gz dev-env` command streamlines this complexity by providing unified management across various platforms and tools.

### Key Features

**Individual Service Management**:

- **Multi-Cloud Support** - AWS, GCP, Azure profile management per service
- **Container Integration** - Docker and Kubernetes configuration management
- **SSH Management** - SSH keys, configs, and tunnels
- **Backup & Restore** - Individual service configuration backup and restore

**Unified Environment Operations**:

- **TUI Dashboard** - Interactive visual management interface
- **Atomic Switching** - Switch all services to target environment simultaneously
- **Status Monitoring** - Comprehensive status view across all services
- **Environment Sync** - Consistent environments across machines

## ☁️ Cloud Platform Management

### AWS Profile Management

Manage AWS profiles and credentials:

```bash
# Individual service management
gz dev-env aws-profile list
gz dev-env aws-profile switch production
gz dev-env aws-credentials save --name production
gz dev-env aws-profile login production

# GCP project management
gz dev-env gcp-project list
gz dev-env gcp-project switch my-project-id
gz dev-env gcp-project show

# Unified environment management
gz dev-env tui                        # Interactive dashboard
gz dev-env switch-all --env production # Switch all services
gz dev-env status                      # Show all services status
```

### AWS Advanced Operations

```bash
# Switch between profiles with session management
gz dev-env aws switch --profile development --session-duration 4h

# Multi-region profile setup
gz dev-env aws create --profile global \
  --regions us-east-1,us-west-2,eu-west-1

# Role-based profile creation
gz dev-env aws create --profile cross-account \
  --role-arn arn:aws:iam::123456789012:role/DevRole

# Profile with MFA
gz dev-env aws create --profile secure \
  --mfa-device arn:aws:iam::123456789012:mfa/username
```

### GCP Project Management

Manage Google Cloud Platform projects:

```bash
# List GCP projects
gz dev-env gcp list

# Set active project
gz dev-env gcp --project my-production-project

# Create service account configuration
gz dev-env gcp create-sa --project my-project --name dev-service

# Validate GCP configuration
gz dev-env gcp validate --project my-project

# Export project configuration
gz dev-env gcp export --project my-project --output gcp-config.json
```

### Azure Subscription Management

```bash
# List Azure subscriptions
gz dev-env azure list

# Set active subscription
gz dev-env azure --subscription production-sub

# Create resource group configuration
gz dev-env azure create-rg --subscription prod --name dev-resources

# Validate Azure configuration
gz dev-env azure validate --subscription production-sub
```

## 🐳 Container Management

### Docker Configuration

Manage Docker environments and registries:

```bash
# Configure Docker registry
gz dev-env docker registry add --name company-registry \
  --url registry.company.com --auth-type basic

# List Docker contexts
gz dev-env docker context list

# Create new Docker context
gz dev-env docker context create --name remote-docker \
  --host tcp://remote-host:2376

# Switch Docker context
gz dev-env docker context use remote-docker

# Registry authentication
gz dev-env docker login --registry registry.company.com
```

### Kubernetes Configuration

Manage Kubernetes contexts and configurations:

```bash
# List Kubernetes contexts
gz dev-env k8s context list

# Switch Kubernetes context
gz dev-env k8s context use production-cluster

# Create kubeconfig for cluster
gz dev-env k8s config create --cluster my-cluster \
  --server https://api.cluster.com --token $K8S_TOKEN

# Validate cluster connectivity
gz dev-env k8s validate --context production-cluster

# Export kubeconfig
gz dev-env k8s export --context all --output kubeconfig-backup.yaml
```

## 🔐 SSH Configuration

### SSH Key Management

```bash
# Generate SSH key pair
gz dev-env ssh keygen --name work-key --type ed25519

# List SSH keys
gz dev-env ssh keys list

# Add SSH key to agent
gz dev-env ssh keys add --name work-key

# Copy public key to clipboard
gz dev-env ssh keys copy --name work-key

# Deploy key to server
gz dev-env ssh keys deploy --name work-key --host server.company.com
```

### SSH Configuration Management

```bash
# Add SSH host configuration
gz dev-env ssh config add --host production \
  --hostname prod.company.com --user deploy --key work-key

# List SSH configurations
gz dev-env ssh config list

# Test SSH connection
gz dev-env ssh test --host production

# Create SSH tunnel
gz dev-env ssh tunnel --host production --local-port 3000 --remote-port 3000

# Export SSH config
gz dev-env ssh export --output ssh-config-backup
```

## 🔧 Development Tools

### Tool Installation and Management

```bash
# Install development tools
gz dev-env tools install --tool golang,nodejs,python

# Update all tools
gz dev-env tools update --all

# List installed tools
gz dev-env tools list

# Check tool versions
gz dev-env tools versions

# Install specific version
gz dev-env tools install --tool nodejs --version 18.17.0
```

### Environment Variables Management

```bash
# Set environment variable
gz dev-env env set DATABASE_URL postgres://localhost:5432/mydb

# List environment variables
gz dev-env env list

# Load environment from file
gz dev-env env load --file .env.production

# Export environment
gz dev-env env export --output environment.env

# Clear environment
gz dev-env env clear --pattern "AWS_*"
```

## 🔄 Environment Synchronization

### Cross-Machine Synchronization

```bash
# Backup current environment
gz dev-env backup --output dev-env-backup.tar.gz

# Restore environment on new machine
gz dev-env restore --input dev-env-backup.tar.gz

# Sync with remote configuration
gz dev-env sync --remote https://config.company.com/dev-env

# Compare environments
gz dev-env diff --remote other-machine
```

### Team Environment Sharing

```bash
# Create team environment template
gz dev-env template create --name golang-web-dev \
  --tools "golang,nodejs,docker" \
  --aws-profile development

# Apply team template
gz dev-env template apply --name golang-web-dev

# Share environment configuration
gz dev-env share --output team-config.yaml

# Validate team configuration
gz dev-env validate --template team-config.yaml
```

## ⚙️ Configuration

### Basic Configuration

Add development environment settings to your `~/.config/gzh-manager/gzh.yaml`:

```yaml
commands:
  dev_env:
    # Default cloud provider
    default_provider: aws

    # Backup settings
    backup:
      enabled: true
      location: "$HOME/.config/gzh-manager/dev-env-backups"
      retention_days: 30

    # Tool management
    tools:
      auto_update: false
      install_location: "$HOME/.local/bin"

    # SSH settings
    ssh:
      key_type: ed25519
      key_location: "$HOME/.ssh"
      config_location: "$HOME/.ssh/config"

    # Synchronization
    sync:
      enabled: true
      remote_url: "https://config.company.com/dev-env"
      auto_sync: false
```

### Advanced Configuration

```yaml
commands:
  dev_env:
    # Provider-specific settings
    aws:
      default_region: us-west-2
      default_output: json
      profiles_location: "$HOME/.aws"

    gcp:
      default_zone: us-central1-a
      service_accounts_location: "$HOME/.config/gcp"

    docker:
      default_registry: "registry.company.com"
      contexts_location: "$HOME/.docker/contexts"

    kubernetes:
      default_namespace: default
      kubeconfig_location: "$HOME/.kube/config"

    # Environment variables
    environment:
      global_vars:
        EDITOR: vim
        BROWSER: firefox

      provider_vars:
        aws:
          AWS_PAGER: ""
          AWS_CLI_AUTO_PROMPT: "on-partial"

    # Tool versions
    tool_versions:
      golang: "1.21.0"
      nodejs: "18.17.0"
      python: "3.11.0"
```

## 📊 Monitoring and Diagnostics

### Environment Health Checks

```bash
# Check overall environment health
gz dev-env health

# Check specific provider health
gz dev-env health --provider aws

# Validate all configurations
gz dev-env validate --all

# Generate diagnostic report
gz dev-env diagnose --output diagnostics.json
```

### Usage Analytics

```bash
# Show environment usage statistics
gz dev-env stats

# Provider usage breakdown
gz dev-env stats --provider aws

# Tool usage metrics
gz dev-env stats --tools

# Export usage data
gz dev-env stats --output csv > usage-stats.csv
```

## 🚀 Integration Examples

### CI/CD Pipeline Integration

```yaml
# GitHub Actions example
- name: Setup Development Environment
  run: |
    gz dev-env restore --input dev-env-template.tar.gz
    gz dev-env aws --profile ci-cd
    gz dev-env k8s context use staging-cluster
```

### Project Setup Automation

```bash
# Project initialization script
#!/bin/bash
gz dev-env template apply --name web-development
gz dev-env aws --profile development
gz dev-env docker context use local
gz dev-env k8s context use dev-cluster
gz dev-env env load --file .env.development
```

### Multi-Environment Workflow

```bash
# Development workflow
gz dev-env aws --profile development
gz dev-env k8s context use dev-cluster

# Staging deployment
gz dev-env aws --profile staging
gz dev-env k8s context use staging-cluster

# Production access (with additional verification)
gz dev-env aws --profile production --verify
gz dev-env k8s context use production-cluster --confirm
```

## 🛡️ Security Best Practices

### Credential Management

```bash
# Use temporary credentials
gz dev-env aws --profile production --session-duration 1h

# Rotate credentials regularly
gz dev-env aws rotate-credentials --profile development

# Audit credential usage
gz dev-env audit --provider aws --days 30
```

### Access Control

```bash
# Role-based access
gz dev-env aws assume-role --role DevRole --profile base

# Multi-factor authentication
gz dev-env aws --profile secure --require-mfa

# Session recording
gz dev-env session start --record --output session.log
```

## 📋 Output Formats

All dev-env commands support multiple output formats:

```bash
# JSON output for automation
gz dev-env aws list --output json

# YAML for configuration
gz dev-env export --output yaml

# Table format (default)
gz dev-env tools list --output table

# CSV for analysis
gz dev-env stats --output csv
```

## 🆘 Troubleshooting

### Common Issues

#### Profile Not Found

```bash
# List available profiles
gz dev-env aws list

# Recreate profile
gz dev-env aws create --profile missing-profile

# Import from AWS CLI
gz dev-env aws import --from-aws-cli
```

#### Connection Issues

```bash
# Test connectivity
gz dev-env validate --provider aws

# Debug network issues
gz dev-env debug --provider aws --verbose

# Reset configuration
gz dev-env reset --provider aws --confirm
```

#### Tool Installation Problems

```bash
# Check tool availability
gz dev-env tools check --tool golang

# Reinstall tool
gz dev-env tools install --tool golang --force

# Update tool sources
gz dev-env tools update-sources
```

______________________________________________________________________

**Supported Providers**: AWS, GCP, Azure, Docker, Kubernetes
**Tools Integration**: Package managers, development tools, CLI utilities
**Security**: MFA support, role-based access, credential rotation
**Sync**: Cross-machine synchronization, team templates
