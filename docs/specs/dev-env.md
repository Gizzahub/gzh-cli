# Development Environment Management Specification

## Overview

The `dev-env` command provides comprehensive development environment management capabilities, focusing on cloud provider configurations, container platforms, and Kubernetes environments. It enables developers to switch between different environment contexts seamlessly and manage multiple development setups.

## Commands

### Core Commands

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

## Examples

### Switching Between Development Environments

```bash
# Switch to development environment
gz dev-env aws switch --profile development
gz dev-env gcp switch --project my-dev-project
gz dev-env docker switch --context development
gz dev-env kubernetes switch --context dev-cluster

# Validate all configurations
gz dev-env aws validate
gz dev-env gcp validate
gz dev-env docker validate
gz dev-env kubernetes validate
```

### Setting Up New Environment

```bash
# Setup new staging environment
gz dev-env aws save --name staging
gz dev-env gcp save --name staging
gz dev-env docker save --name staging
gz dev-env kubernetes save --name staging
```

### Environment Automation

```bash
# Switch entire environment with single command (future feature)
gz dev-env switch-all --environment production

# Backup current environment settings
gz dev-env backup --output /backup/dev-env-backup.yaml

# Restore environment from backup
gz dev-env restore --input /backup/dev-env-backup.yaml
```

## Integration Points

- **Network Environment**: Coordinates with `net-env` for network-specific configurations
- **Repository Management**: Integrates with `synclone` for environment-specific repository access
- **Configuration Generation**: Works with `gen-config` for environment-specific configuration files
- **IDE Settings**: Synchronizes with `ide` command for environment-specific IDE configurations

## Security Considerations

- Credential encryption at rest
- Secure credential rotation workflows
- MFA support for sensitive operations
- Audit logging for environment changes
- Role-based access control integration
