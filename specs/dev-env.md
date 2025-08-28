<!-- 🚫 AI_MODIFY_PROHIBITED -->

<!-- This file should not be modified by AI agents -->

# Development Environment Management Specification

## Overview

The `dev-env` command provides development environment management capabilities, focusing on cloud provider configurations, container platforms, and Kubernetes environments. It enables developers to save, restore, and switch between different development environment configurations through both individual service management and unified environment switching.

## Commands

### Core Approach: Hybrid Management

The dev-env command provides both individual service management (for fine-grained control) and unified environment switching (for streamlined workflows):

#### Individual Service Management (Primary)
- `gz dev-env kubeconfig` - Kubernetes configuration management
- `gz dev-env docker` - Docker environment and context management
- `gz dev-env aws` - AWS configuration management
- `gz dev-env aws-credentials` - AWS credentials management
- `gz dev-env aws-profile` - AWS profile management with SSO support
- `gz dev-env gcloud` - Google Cloud configuration management
- `gz dev-env gcloud-credentials` - Google Cloud credentials management
- `gz dev-env gcp-project` - GCP project management
- `gz dev-env azure-subscription` - Azure subscription management
- `gz dev-env ssh` - SSH configuration management

#### Unified Environment Management (Secondary)
- `gz dev-env tui` - Interactive TUI dashboard for environment management
- `gz dev-env switch-all` - Atomic environment switching across all services
- `gz dev-env status` - Display current status of all development services

### Individual Service Commands

#### Kubernetes Configuration (`gz dev-env kubeconfig`)

**Purpose**: Manage Kubernetes configuration files and contexts

**Features**:
- Save and restore kubeconfig files
- Switch between different cluster contexts
- Manage multiple environment configurations

**Usage**:

```bash
gz dev-env kubeconfig save --name my-cluster    # Save current kubeconfig
gz dev-env kubeconfig load --name my-cluster    # Load saved kubeconfig
gz dev-env kubeconfig list                      # List saved configurations
gz dev-env kubeconfig delete --name old-cluster # Delete saved configuration
```

#### Docker Environment (`gz dev-env docker`)

**Purpose**: Manage Docker contexts and configurations

**Features**:
- Save and restore Docker contexts
- Switch between different Docker environments
- Manage Docker registry configurations

**Usage**:

```bash
gz dev-env docker save --name production        # Save current Docker config
gz dev-env docker load --name production        # Load Docker configuration
gz dev-env docker list                          # List saved Docker configs
gz dev-env docker context switch remote-prod    # Switch Docker context
```

#### AWS Management (`gz dev-env aws`)

**Purpose**: Manage AWS configurations and settings

**Features**:
- Save and restore AWS configurations
- Manage multiple AWS environment setups
- Region and account switching

**Usage**:

```bash
gz dev-env aws save --name production           # Save AWS configuration
gz dev-env aws load --name production           # Load AWS configuration
gz dev-env aws list                             # List saved AWS configs
gz dev-env aws region set us-west-2             # Set AWS region
```

#### AWS Credentials (`gz dev-env aws-credentials`)

**Purpose**: Secure management of AWS credentials

**Features**:
- Encrypted storage of AWS credentials
- Multiple credential profile management
- Credential rotation and validation

**Usage**:

```bash
gz dev-env aws-credentials save --name prod     # Save AWS credentials
gz dev-env aws-credentials load --name prod     # Load AWS credentials
gz dev-env aws-credentials list                 # List credential profiles
gz dev-env aws-credentials validate --name prod # Validate credentials
```

#### AWS Profile Management (`gz dev-env aws-profile`)

**Purpose**: Advanced AWS profile management with SSO support

**Features**:
- AWS SSO integration
- Profile switching and management
- Session management and renewal

**Usage**:

```bash
gz dev-env aws-profile list                     # List AWS profiles
gz dev-env aws-profile switch production        # Switch to profile
gz dev-env aws-profile login production         # Login with SSO
gz dev-env aws-profile logout                   # Logout from current profile
```

#### Google Cloud Management (`gz dev-env gcloud`)

**Purpose**: Manage Google Cloud SDK configurations

**Features**:
- Save and restore gcloud configurations
- Multiple gcloud configuration management
- Account and authentication handling

**Usage**:

```bash
gz dev-env gcloud save --name production        # Save gcloud config
gz dev-env gcloud load --name production        # Load gcloud config
gz dev-env gcloud list                          # List saved configurations
gz dev-env gcloud auth login                    # Authenticate with Google Cloud
```

#### Google Cloud Credentials (`gz dev-env gcloud-credentials`)

**Purpose**: Secure Google Cloud credentials management

**Features**:
- Service account key management
- Application default credentials handling
- Credential validation and rotation

**Usage**:

```bash
gz dev-env gcloud-credentials save --name prod  # Save GC credentials
gz dev-env gcloud-credentials load --name prod  # Load GC credentials
gz dev-env gcloud-credentials list              # List credential profiles
gz dev-env gcloud-credentials validate          # Validate current credentials
```

#### GCP Project Management (`gz dev-env gcp-project`)

**Purpose**: Manage GCP projects and configurations

**Features**:
- Project switching and management
- gcloud configuration per project
- Service account management
- Project validation

**Usage**:

```bash
gz dev-env gcp-project list                     # List available projects
gz dev-env gcp-project switch my-project-id     # Switch to project
gz dev-env gcp-project show                     # Show current project details
gz dev-env gcp-project config create --name prod --project my-prod-project # Create config
gz dev-env gcp-project validate                 # Validate project setup
```

#### Azure Subscription Management (`gz dev-env azure-subscription`)

**Purpose**: Manage Azure subscriptions and configurations

**Features**:
- Azure subscription switching
- Multi-tenant support
- Azure CLI integration
- Subscription validation

**Usage**:

```bash
gz dev-env azure-subscription list              # List subscriptions
gz dev-env azure-subscription switch my-sub-id  # Switch subscription
gz dev-env azure-subscription show              # Show current subscription
gz dev-env azure-subscription login             # Login to Azure
gz dev-env azure-subscription validate          # Validate subscription
```

#### SSH Configuration (`gz dev-env ssh`)

**Purpose**: Manage SSH configurations and keys

**Features**:
- SSH config file management
- SSH key backup and restore
- Multiple environment SSH setups

**Usage**:

```bash
gz dev-env ssh save --name production           # Save SSH configuration
gz dev-env ssh load --name production           # Load SSH configuration
gz dev-env ssh list                             # List saved SSH configs
gz dev-env ssh key generate --name new-key      # Generate SSH key pair
```

### Unified Environment Commands

#### Interactive TUI Dashboard (`gz dev-env tui`)

**Purpose**: Provides an interactive terminal interface for managing all development environments

**Features**:
- Real-time service status monitoring
- Interactive service management
- Environment switching capabilities
- Service logs and details view
- Keyboard shortcuts and navigation

**Usage**:

```bash
gz dev-env tui                                  # Launch interactive TUI dashboard
```

**TUI Navigation**:
- `↑/k, ↓/j` - Navigate up/down
- `←/h, →/l` - Navigate left/right
- `Enter` - Select/confirm action
- `Esc` - Go back to previous view
- `q` - Quit dashboard

#### Atomic Environment Switching (`gz dev-env switch-all`)

**Purpose**: Switch multiple cloud services and development environments atomically

**Features**:
- Atomic switching across all configured services
- Rollback on failure
- Dependency resolution and ordering
- Progress tracking
- Pre/post switch hooks

**Usage**:

```bash
gz dev-env switch-all --env production          # Switch all to production
gz dev-env switch-all --env dev --dry-run       # Preview changes
gz dev-env switch-all --from-file env.yaml      # Switch using environment file
gz dev-env switch-all --rollback                # Rollback last switch
```

#### Development Environment Status (`gz dev-env status`)

**Purpose**: Display comprehensive status of all development environment services

**Features**:
- Unified status view of all services
- Color-coded status indicators
- Credential expiration warnings
- Service health checks
- Multiple output formats

**Usage**:

```bash
gz dev-env status                               # Show all services status
gz dev-env status --service aws                 # Show specific service status
gz dev-env status --format json                 # Output as JSON
gz dev-env status --check-health                # Include health checks
gz dev-env status --watch                       # Real-time status updates
```

## Configuration

### Environment Configuration Files

Environment configurations are stored in:

- `~/.config/gzh-manager/dev-env/` - User-specific configurations
- `~/.config/gzh-manager/dev-env/environments/` - Environment presets
- Environment variable: `GZH_DEV_ENV_CONFIG`

### Configuration Structure

```yaml
# Example environment preset: ~/.config/gzh-manager/dev-env/environments/production.yaml
name: "Production Environment"
description: "Production cloud environment configuration"

services:
  gcp:
    project: "my-company-prod"
    region: "us-central1"
    account: "prod-service@company.iam.gserviceaccount.com"

  aws:
    profile: "production"
    region: "us-west-2"
    account: "123456789012"

  azure:
    subscription: "prod-subscription-id"
    tenant: "company-tenant-id"

  kubernetes:
    context: "prod-cluster"
    namespace: "default"

  docker:
    context: "prod-remote"

  ssh:
    config: "production"
```

### Individual Service Configuration

Each service maintains its own configuration format:

```yaml
# ~/.config/gzh-manager/dev-env/kubeconfig/
my-cluster.yaml              # Saved kubeconfig file

# ~/.config/gzh-manager/dev-env/docker/
production.json              # Docker context configuration

# ~/.config/gzh-manager/dev-env/aws/
production-config            # AWS configuration
production-credentials       # AWS credentials (encrypted)
```

## Examples

### Individual Service Management

```bash
# Set up production GCP environment
gz dev-env gcp-project switch my-prod-project
gz dev-env gcp-project config create --name prod --project my-prod-project

# Configure AWS for production
gz dev-env aws-profile switch production
gz dev-env aws-profile login production

# Switch Kubernetes context
gz dev-env kubeconfig load --name prod-cluster

# Configure Docker for remote environment
gz dev-env docker load --name production
```

### Environment Switching Workflow

```bash
# Individual approach (fine-grained control)
gz dev-env gcp-project switch prod-project
gz dev-env aws-profile switch production
gz dev-env kubeconfig load --name prod-cluster
gz dev-env docker context switch prod-remote

# Unified approach (streamlined workflow)
gz dev-env switch-all --env production

# Check current status
gz dev-env status --format table
```

### Interactive Management

```bash
# Launch TUI for visual management
gz dev-env tui

# Monitor all services in real-time
gz dev-env status --watch
```

### Service-Specific Operations

```bash
# AWS profile management with SSO
gz dev-env aws-profile list
gz dev-env aws-profile login production
gz dev-env aws-credentials validate --name production

# GCP project management
gz dev-env gcp-project list
gz dev-env gcp-project show
gz dev-env gcp-project service-account list

# Azure subscription management
gz dev-env azure-subscription show
gz dev-env azure-subscription validate
```

## Integration Points

- **Repository Management**: Coordinates with `synclone` for cloud-specific repository access
- **Network Environment**: Integrates with `net-env` for cloud provider network configurations
- **Package Management**: Works with `pm` for cloud CLI tool updates
- **IDE Integration**: Synchronizes with `ide` command for cloud development configurations

## Security Considerations

- **Credential Encryption**: All sensitive credentials are encrypted at rest
- **Permission Management**: Follows principle of least privilege for service accounts
- **Audit Logging**: Complete logging of all environment switches and credential usage
- **Credential Rotation**: Automated alerts for credential expiration and rotation
- **Secure Storage**: Credentials stored in protected directories with appropriate permissions
- **Session Management**: Automatic cleanup of temporary credentials and sessions

## Platform Support

- **Linux**: Full support for all cloud CLI tools and SSH management
- **macOS**: Complete support for all services and TUI interface
- **Windows**: Basic support for cloud services, limited SSH management
- **Cloud Environments**: Native integration with AWS, GCP, and Azure services
- **Container Environments**: Docker and Kubernetes integration across platforms

## Future Enhancements

The following advanced features may be considered for future releases:

### Advanced Integration Features
- **Environment Validation**: Comprehensive configuration and permission validation
- **Configuration Sync**: Bidirectional sync between local and cloud configurations
- **Advanced TUI Features**: Enhanced visual dashboard with metrics and logs
- **Environment Templates**: Ready-to-use templates for common environment setups
- **Quick Switch Presets**: Saved environment states for instant switching

### Automation and Orchestration
- **Pre/Post Hooks**: Custom scripts executed during environment switches
- **Dependency Resolution**: Automatic ordering and dependency handling during switches
- **Rollback Mechanisms**: Comprehensive rollback for failed environment transitions
- **Health Monitoring**: Continuous monitoring and alerting for environment health

### Enhanced Security
- **Multi-Factor Authentication**: Enhanced MFA support for cloud services
- **Credential Vaults**: Integration with external credential management systems
- **Policy Enforcement**: Configuration compliance and security policy enforcement
- **Audit Reports**: Detailed audit trails and compliance reporting

## Summary

The `dev-env` command provides flexible development environment management through both individual service control and unified environment operations:

### Current Capabilities

- **Individual Service Management**: Fine-grained control over each cloud service and development tool
- **Unified Environment Operations**: Streamlined switching and monitoring across all services
- **Interactive Management**: TUI dashboard for visual environment management
- **Secure Configuration**: Encrypted storage and secure handling of credentials
- **Multi-Platform Support**: Cross-platform support for major cloud providers

### Design Philosophy

- **Flexibility**: Choose between individual service control or unified operations
- **Security**: Comprehensive security measures for credential and configuration management
- **Reliability**: Atomic operations with rollback capabilities for environment switching
- **Usability**: Both command-line and interactive interfaces for different workflows
- **Extensibility**: Foundation for advanced environment management features

The command serves developers who need both the precision of individual service management and the efficiency of unified environment operations, providing a comprehensive solution for modern multi-cloud development workflows.
