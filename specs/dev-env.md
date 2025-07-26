<!-- 🚫 AI_MODIFY_PROHIBITED -->
<!-- This file should not be modified by AI agents -->

# Development Environment Management Specification

## Overview

The `dev-env` command provides comprehensive development environment management capabilities, focusing on cloud provider configurations, container platforms, and Kubernetes environments. It enables developers to switch between different environment contexts seamlessly and manage multiple development setups.

## Commands

### Core Commands

- `gz dev-env` - Interactive TUI mode for environment management
- `gz dev-env switch-all` - Unified environment switching across all services
- `gz dev-env status` - Display current environment status for all services
- `gz dev-env validate` - Validate all environment configurations
- `gz dev-env sync` - Synchronize configurations with actual state
- `gz dev-env quick` - Quick switch to saved environment presets
- `gz dev-env aws` - AWS profile and configuration management
- `gz dev-env gcp` - Google Cloud Platform project management
- `gz dev-env docker` - Docker environment and profile switching
- `gz dev-env kubernetes` - Kubernetes context and namespace management
- `gz dev-env azure` - Azure subscription management
- `gz dev-env ssh` - SSH configuration management
- `gz dev-env kubeconfig` - Kubernetes configuration management
- `gz dev-env aws-credentials` - AWS credentials management
- `gz dev-env aws-profile` - AWS profile management with SSO support
- `gz dev-env gcloud` - Google Cloud configuration management
- `gz dev-env gcloud-credentials` - Google Cloud credentials management
- `gz dev-env gcp-project` - GCP project management
- `gz dev-env azure-subscription` - Azure subscription management

### Interactive TUI Mode (`gz dev-env`)

**Purpose**: Provides an interactive terminal UI for managing all development environments

**Features**:
- Visual dashboard showing current environment status
- Hierarchical menu navigation (Service → Action → Target)
- Real-time status updates
- Keyboard shortcuts for common operations
- Search and filter capabilities

**Usage**:
```bash
gz dev-env                    # Launch interactive TUI
gz dev-env --mode compact     # Compact view for smaller terminals
```

### Unified Environment Switching (`gz dev-env switch-all`)

**Purpose**: Switch multiple services to a predefined environment configuration with a single command

**Features**:
- Atomic environment switching (all or nothing)
- Dependency resolution and ordering
- Pre/post switch hooks
- Rollback on failure
- Progress tracking with detailed output

**Usage**:
```bash
gz dev-env switch-all --env production     # Switch all services to production
gz dev-env switch-all --env dev --dry-run  # Preview changes without applying
gz dev-env switch-all --from-file env.yaml # Switch using environment file
```

### Environment Status (`gz dev-env status`)

**Purpose**: Display comprehensive status of all configured services

**Features**:
- Unified status table
- Color-coded status indicators
- Credential expiration warnings
- Service health checks
- Export to various formats (json, yaml, table)

**Usage**:
```bash
gz dev-env status                    # Show all services status
gz dev-env status --service aws      # Show specific service status
gz dev-env status --format json      # Output as JSON
gz dev-env status --check-health     # Include health checks
```

### Environment Validation (`gz dev-env validate`)

**Purpose**: Validate all environment configurations and credentials

**Features**:
- Configuration syntax validation
- Credential validity checks
- Permission verification
- Resource accessibility tests
- Detailed error reporting

**Usage**:
```bash
gz dev-env validate                  # Validate all configurations
gz dev-env validate --service gcp    # Validate specific service
gz dev-env validate --fix            # Attempt to fix issues
```

### Configuration Sync (`gz dev-env sync`)

**Purpose**: Synchronize local configurations with actual service states

**Features**:
- Detect configuration drift
- Update local configs from services
- Backup before sync
- Selective sync options
- Conflict resolution

**Usage**:
```bash
gz dev-env sync                      # Sync all configurations
gz dev-env sync --service kubernetes # Sync specific service
gz dev-env sync --direction pull     # Pull configs from services
gz dev-env sync --direction push     # Push configs to services
```

### Quick Switch (`gz dev-env quick`)

**Purpose**: Quickly switch between frequently used environment presets

**Features**:
- Save current state as preset
- Instant environment switching
- Recent environments history
- Preset management (list, delete, rename)
- Auto-complete for preset names

**Usage**:
```bash
gz dev-env quick save dev            # Save current state as 'dev'
gz dev-env quick dev                 # Switch to 'dev' preset
gz dev-env quick list                # List all presets
gz dev-env quick delete old-env      # Delete a preset
gz dev-env quick last                # Switch to last used environment
```

### AWS Management (`gz dev-env aws`)

**Purpose**: Manage AWS profiles, credentials, and environment configurations

**Features**:
- Switch between AWS profiles
- Manage multiple AWS account configurations
- Handle credential rotation and validation
- Support for MFA and SSO configurations
- Environment-specific AWS settings

**Usage**:
```bash
gz dev-env aws list                    # List available AWS profiles
gz dev-env aws switch --profile prod   # Switch to production profile
gz dev-env aws validate               # Validate current credentials
gz dev-env aws save --name production # Save current AWS config
gz dev-env aws load --name production # Load saved AWS config
```

### GCP Management (`gz dev-env gcp`)

**Purpose**: Manage Google Cloud Platform projects and service account configurations

**Features**:
- Switch between GCP projects
- Manage service account credentials
- Handle gcloud configuration contexts
- Support for multiple billing accounts
- Environment-specific GCP settings

**Usage**:
```bash
gz dev-env gcp list                    # List available GCP projects
gz dev-env gcp switch --project my-app # Switch to specific project
gz dev-env gcp validate               # Validate current configuration
gz dev-env gcp save --name production # Save current GCP config
gz dev-env gcp load --name production # Load saved GCP config
```

### Docker Management (`gz dev-env docker`)

**Purpose**: Manage Docker environments, registries, and container configurations

**Features**:
- Switch between Docker contexts
- Manage container registry configurations
- Handle Docker daemon settings
- Support for multiple Docker environments
- Container image management

**Usage**:
```bash
gz dev-env docker list                 # List Docker contexts
gz dev-env docker switch --context prod # Switch Docker context
gz dev-env docker save --name production # Save current Docker config
gz dev-env docker load --name production # Load saved Docker config
```

### Kubernetes Management (`gz dev-env kubernetes`)

**Purpose**: Manage Kubernetes clusters, contexts, and namespace configurations

**Features**:
- Switch between Kubernetes contexts
- Manage namespace configurations
- Handle kubeconfig files
- Support for multiple clusters
- RBAC and policy management

**Usage**:
```bash
gz dev-env kubernetes list             # List available contexts
gz dev-env kubernetes switch --context staging # Switch context
gz dev-env kubernetes save --name production # Save current kubeconfig
gz dev-env kubernetes load --name production # Load saved kubeconfig
```

### Azure Management (`gz dev-env azure`)

**Purpose**: Manage Azure subscriptions and tenant configurations

**Features**:
- Switch between Azure subscriptions
- Manage multi-tenant configurations
- Handle Azure CLI settings
- Support for service principal authentication
- Environment-specific Azure settings

**Usage**:
```bash
gz dev-env azure list                  # List available Azure subscriptions
gz dev-env azure switch --subscription my-sub # Switch to specific subscription
gz dev-env azure validate             # Validate current configuration
gz dev-env azure save --name production # Save current Azure config
gz dev-env azure load --name production # Load saved Azure config
```

### SSH Management (`gz dev-env ssh`)

**Purpose**: Manage SSH configurations for Git operations

**Features**:
- Save and load SSH configurations
- Manage SSH keys for different services
- Handle SSH config file generation
- Support for multiple SSH profiles

**Usage**:
```bash
gz dev-env ssh save --name production  # Save current SSH config
gz dev-env ssh load --name production  # Load saved SSH config
gz dev-env ssh list                   # List saved SSH configurations
```

## Configuration

### Global Configuration

Development environment configurations are stored in:
- `~/.config/gzh-manager/dev-env.yaml` - User-specific settings
- `/etc/gzh-manager/dev-env.yaml` - System-wide settings
- Environment variable: `GZH_DEV_ENV_CONFIG`

### Configuration Structure

```yaml
# Development Environment Configuration
version: "1.0.0"

# Environment profiles for unified switching
environments:
  development:
    description: "Development environment with local services"
    aws_profile: "dev-account"
    gcp_project: "my-dev-project"
    kubernetes_context: "docker-desktop"
    docker_context: "default"
    azure_subscription: "dev-subscription"
    hooks:
      pre_switch:
        - "echo 'Switching to development environment...'"
      post_switch:
        - "echo 'Development environment activated'"
    
  staging:
    description: "Staging environment with shared resources"
    aws_profile: "staging-account"
    gcp_project: "my-staging-project"
    kubernetes_context: "staging-cluster"
    docker_context: "staging"
    azure_subscription: "staging-subscription"
    dependencies:
      - "vpn:staging"  # Ensure VPN is connected
    
  production:
    description: "Production environment with strict access controls"
    aws_profile: "prod-account"
    gcp_project: "my-prod-project"
    kubernetes_context: "prod-cluster"
    docker_context: "production"
    azure_subscription: "prod-subscription"
    require_confirmation: true
    hooks:
      pre_switch:
        - "az login --tenant prod-tenant"
        - "aws sso login --profile prod-account"
        - "gcloud auth login"
      post_switch:
        - "kubectl config set-context --current --namespace=production"

# Quick switch presets
quick_presets:
  dev:
    environment: "development"
    description: "Quick switch to dev"
  prod:
    environment: "production"
    description: "Quick switch to production (requires confirmation)"
  
# Auto-detection rules
auto_detection:
  enabled: true
  rules:
    - path_pattern: "*/mycompany/backend/*"
      environment: "development"
    - path_pattern: "*/production-repos/*"
      environment: "production"
    - git_remote: "git@github.com:mycompany-prod/*"
      environment: "production"

# Service-specific configurations
dev_env:
  # AWS Configuration
  aws:
    default_profile: "development"
    profiles:
      development:
        region: "us-west-2"
        output: "json"
        mfa_device: "arn:aws:iam::123456789012:mfa/user"
      production:
        region: "us-east-1"
        output: "table"
        role_arn: "arn:aws:iam::987654321098:role/ProdAccess"

  # GCP Configuration
  gcp:
    default_project: "my-dev-project"
    projects:
      my-dev-project:
        zone: "us-central1-a"
        region: "us-central1"
        service_account: "dev-sa@my-project.iam.gserviceaccount.com"
      my-prod-project:
        zone: "us-east1-a"
        region: "us-east1"
        service_account: "prod-sa@my-project.iam.gserviceaccount.com"

  # Docker Configuration
  docker:
    default_context: "development"
    contexts:
      development:
        host: "unix:///var/run/docker.sock"
        registries:
          - url: "docker.io"
            username: "myuser"
          - url: "gcr.io/my-project"
            auth_method: "gcloud"
      production:
        host: "tcp://prod-docker:2376"
        tls_verify: true
        cert_path: "/path/to/certs"

  # Kubernetes Configuration
  kubernetes:
    default_context: "dev-cluster"
    contexts:
      dev-cluster:
        cluster: "development"
        namespace: "default"
        user: "dev-user"
      staging-cluster:
        cluster: "staging"
        namespace: "staging"
        user: "staging-user"
```

### Environment Variables

- `GZH_DEV_ENV_CONFIG` - Path to configuration file
- `GZH_AWS_PROFILE` - Override default AWS profile
- `GZH_GCP_PROJECT` - Override default GCP project
- `GZH_DOCKER_CONTEXT` - Override default Docker context
- `GZH_KUBE_CONTEXT` - Override default Kubernetes context
- `GZH_DEV_ENV` - Override default environment
- `GZH_DEV_ENV_AUTO_DETECT` - Enable/disable auto-detection (true/false)
- `GZH_DEV_ENV_HOOKS_ENABLED` - Enable/disable hooks execution (true/false)

## Examples

### Interactive TUI Mode

```bash
# Launch interactive environment manager
gz dev-env

# Use compact mode for smaller terminals
gz dev-env --mode compact

# Start with specific service selected
gz dev-env --service aws
```

### Unified Environment Switching

```bash
# Switch all services to production environment
gz dev-env switch-all --env production

# Preview changes before applying
gz dev-env switch-all --env staging --dry-run

# Force switch without confirmation
gz dev-env switch-all --env development --force

# Switch with custom timeout
gz dev-env switch-all --env production --timeout 5m
```

### Quick Environment Switching

```bash
# Save current state as a preset
gz dev-env quick save myproject

# Switch to saved preset
gz dev-env quick myproject

# Switch to last used environment
gz dev-env quick last

# List all presets
gz dev-env quick list

# Delete old preset
gz dev-env quick delete old-project
```

### Environment Status and Validation

```bash
# Show comprehensive status
gz dev-env status

# Check specific service status
gz dev-env status --service kubernetes

# Export status as JSON
gz dev-env status --format json > env-status.json

# Validate all configurations
gz dev-env validate

# Validate and attempt fixes
gz dev-env validate --fix

# Validate specific service
gz dev-env validate --service aws
```

### Configuration Synchronization

```bash
# Sync all configurations with actual state
gz dev-env sync

# Pull configurations from services
gz dev-env sync --direction pull

# Push local configs to services
gz dev-env sync --direction push --confirm

# Sync specific service only
gz dev-env sync --service gcp
```

### Individual Service Management

```bash
# Switch to development environment (traditional way)
gz dev-env aws switch --profile development
gz dev-env gcp switch --project my-dev-project
gz dev-env docker switch --context development
gz dev-env kubernetes switch --context dev-cluster

# Interactive service switching
gz dev-env aws-profile switch --interactive
gz dev-env gcp-project switch --interactive
gz dev-env azure-subscription switch --interactive
```

### Environment Templates

```bash
# Create template from current environment
gz dev-env template create --name microservice --from current

# Apply template to new project
gz dev-env template apply --name microservice --project new-api

# List available templates
gz dev-env template list

# Export template for sharing
gz dev-env template export --name microservice > microservice-env.yaml

# Import shared template
gz dev-env template import --file team-template.yaml
```

### Auto-Detection and Smart Defaults

```bash
# Enable auto-detection for current directory
cd ~/projects/backend-api
gz dev-env auto-detect enable

# Check which environment would be auto-selected
gz dev-env auto-detect check

# Override auto-detection temporarily
GZH_DEV_ENV_AUTO_DETECT=false gz dev-env status
```

### Hook Management

```bash
# Test hooks without switching
gz dev-env hooks test --env production

# Skip hooks during switch
gz dev-env switch-all --env staging --skip-hooks

# Run only specific hook
gz dev-env hooks run --env production --hook pre_switch
```

## Integration Points

- **Network Environment**: Coordinates with `net-env` for network-specific configurations
- **Repository Management**: Integrates with `synclone` for environment-specific repository access
- **Configuration Generation**: Works with configuration commands for environment-specific configuration files
- **IDE Settings**: Synchronizes with `ide` command for environment-specific IDE configurations

## TUI Interface Design

### Main Dashboard

The interactive TUI provides a comprehensive view of all environment configurations:

```
┌─ GZH Development Environment Manager ─────────────────────────────────────┐
│ Current Environment: development                    [↑↓] Navigate [Enter] Select │
├─────────────────────────────────────────────────────────────────────────┤
│ Service          │ Status    │ Current Config         │ Last Updated     │
├──────────────────┼───────────┼────────────────────────┼──────────────────┤
│ AWS              │ ● Active  │ dev-account           │ 2 hours ago      │
│ GCP              │ ● Active  │ my-dev-project        │ 1 day ago        │
│ Kubernetes       │ ○ Warning │ docker-desktop        │ 3 days ago       │
│ Docker           │ ● Active  │ default               │ Just now         │
│ Azure            │ ○ Expired │ dev-subscription      │ 5 days ago       │
├─────────────────────────────────────────────────────────────────────────┤
│ Quick Actions: [s]witch-all [v]alidate [y]nc [q]uick [?]help [Q]uit     │
└─────────────────────────────────────────────────────────────────────────┘
```

### Service Detail View

```
┌─ AWS Configuration ───────────────────────────────────────────────────────┐
│ Profile: dev-account                                                      │
│ Region: us-west-2                                                         │
│ Status: ● Active (SSO session valid for 8h)                             │
├─────────────────────────────────────────────────────────────────────────┤
│ Available Actions:                                                        │
│   [1] Switch Profile          [4] Validate Credentials                   │
│   [2] Login (SSO)            [5] View Configuration                     │
│   [3] Save Current State     [6] Back to Dashboard                      │
├─────────────────────────────────────────────────────────────────────────┤
│ Recent Profiles:                                                          │
│   • dev-account (current)                                                 │
│   • staging-account                                                       │
│   • prod-account                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Hook System

### Hook Configuration

Hooks allow automatic execution of commands during environment transitions:

```yaml
environments:
  production:
    hooks:
      pre_switch:
        - command: "aws sso login --profile prod-account"
          description: "Authenticate with AWS SSO"
          timeout: "5m"
          on_failure: "abort"  # abort | continue | retry
        
        - command: "kubectl config use-context prod-cluster"
          description: "Switch Kubernetes context"
          on_failure: "retry"
          max_retries: 3
      
      post_switch:
        - command: "notify-slack 'Switched to production environment'"
          description: "Send notification"
          on_failure: "continue"
      
      on_error:
        - command: "gz dev-env switch-all --env development"
          description: "Rollback to development"
```

### Hook Types

- **pre_switch**: Execute before switching environments
- **post_switch**: Execute after successful switch
- **on_error**: Execute if switch fails
- **validate**: Execute during validation phase

## Smart Defaults and Auto-Detection

### Project-Level Configuration

Create `.gzh-env` file in project root:

```yaml
# .gzh-env
default_environment: "development"
auto_switch: true
required_services:
  - aws
  - kubernetes
  - docker
```

### Auto-Detection Rules

```yaml
auto_detection:
  enabled: true
  priority: # Order of precedence
    - explicit_override    # Command line flags
    - environment_var      # GZH_DEV_ENV variable
    - project_file        # .gzh-env file
    - git_remote          # Git remote URL patterns
    - path_pattern        # Directory path patterns
    - last_used           # Last used in this directory
  
  rules:
    - name: "Production repos"
      git_remote_pattern: ".*production.*"
      environment: "production"
      require_confirmation: true
    
    - name: "Dev workspace"
      path_pattern: "~/dev/.*"
      environment: "development"
    
    - name: "Client projects"
      path_pattern: "~/clients/([^/]+)/.*"
      environment: "client-${1}"  # Dynamic environment
```

## Environment Templates

### Template Structure

```yaml
# Template: microservice-dev
template:
  name: "microservice-dev"
  description: "Standard microservice development setup"
  version: "1.0.0"
  
  services:
    aws:
      profile_template: "${project}-dev"
      region: "us-west-2"
    
    kubernetes:
      context_template: "${project}-dev-cluster"
      namespace: "${project}"
    
    docker:
      context: "default"
      registry: "docker.io/${org}/${project}"
  
  variables:
    - name: "project"
      description: "Project name"
      required: true
    - name: "org"
      description: "Organization name"
      default: "mycompany"
```

### Template Commands

```bash
# Create template from current state
gz dev-env template create --name webapp --description "Web application template"

# Apply template with variables
gz dev-env template apply --name microservice-dev \
  --var project=user-service \
  --var org=acme-corp

# Share templates
gz dev-env template export --name webapp > webapp-template.yaml
gz dev-env template import --url https://templates.example.com/webapp.yaml
```

## Security Considerations

- Credential encryption at rest using OS keychain
- Secure credential rotation workflows with automatic expiry detection
- MFA support for sensitive operations
- Comprehensive audit logging for all environment changes
- Role-based access control integration
- Hook script sandboxing and timeout enforcement
- Template validation and sanitization
- Confirmation prompts for production environments

## Summary of Enhancements

The enhanced `dev-env` command provides a comprehensive solution for managing development environments with focus on:

### User Experience
- **Interactive TUI**: Visual dashboard for easy environment management
- **Unified Switching**: Single command to switch entire development stack
- **Quick Presets**: Fast switching between frequently used environments
- **Smart Defaults**: Automatic environment detection based on context

### Automation
- **Hook System**: Automated tasks during environment transitions
- **Template System**: Reusable environment configurations
- **Validation**: Automatic configuration and credential validation
- **Synchronization**: Keep local configs in sync with actual state

### Safety and Security
- **Atomic Operations**: All-or-nothing environment switches
- **Rollback Support**: Automatic rollback on failures
- **Production Safeguards**: Confirmation prompts and access controls
- **Audit Trail**: Comprehensive logging of all environment changes

These enhancements make `gz dev-env` a powerful tool for developers working with multiple environments, reducing context switching overhead while maintaining security and consistency.
