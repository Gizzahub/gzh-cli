# dev-env Command Reference

Development environment configuration management for AWS, Docker, Kubernetes, and SSH settings.

## Synopsis

```bash
gz dev-env <action> [flags]
gz dev-env <action> --config <config-file>
```

## Description

The `dev-env` command manages development environment configurations across cloud platforms, containerization tools, and remote access settings.

## Supported Environments

- **AWS** - Profiles, credentials, regions
- **Docker** - Contexts, registries, configurations
- **Kubernetes** - Contexts, namespaces, clusters
- **SSH** - Host configurations, key management
- **GCP** - Projects, service accounts, configurations

## Actions

### `gz dev-env sync`

Synchronize development environment configurations.

```bash
gz dev-env sync [environment] [flags]
```

**Arguments:**
- `environment` - Specific environment to sync: aws, docker, kubernetes, ssh, gcp

**Flags:**
- `--all` - Sync all environments (default: true)
- `--dry-run` - Show what would be synced without executing
- `--backup` - Create backup before syncing (default: true)

**Examples:**
```bash
# Sync all environments
gz dev-env sync --all

# Sync specific environment
gz dev-env sync aws

# Dry run to preview changes
gz dev-env sync --dry-run
```

### `gz dev-env list`

List available development environments and their status.

```bash
gz dev-env list [flags]
```

**Flags:**
- `--environments` - Filter by specific environments
- `--status` - Filter by status: active, inactive, error
- `--output` - Output format: table, json, yaml

**Examples:**
```bash
# List all environments
gz dev-env list

# Show only active environments
gz dev-env list --status active

# JSON output
gz dev-env list --output json
```

### `gz dev-env switch`

Switch between environment configurations.

```bash
gz dev-env switch <environment> <profile> [flags]
```

**Arguments:**
- `environment` - Environment type: aws, docker, kubernetes
- `profile` - Profile/context name

**Flags:**
- `--persist` - Make switch persistent across sessions

**Examples:**
```bash
# Switch AWS profile
gz dev-env switch aws production

# Switch Kubernetes context
gz dev-env switch kubernetes staging-cluster

# Switch Docker context
gz dev-env switch docker remote-docker
```

### `gz dev-env configure`

Configure development environment settings.

```bash
gz dev-env configure <environment> [flags]
```

**Flags:**
- `--interactive` - Interactive configuration mode
- `--template` - Use configuration template
- `--import` - Import from existing configuration

**Examples:**
```bash
# Interactive AWS configuration
gz dev-env configure aws --interactive

# Configure using template
gz dev-env configure docker --template development

# Import existing configuration
gz dev-env configure kubernetes --import ~/.kube/config
```

### `gz dev-env backup`

Backup development environment configurations.

```bash
gz dev-env backup [flags]
```

**Flags:**
- `--environments` - Specific environments to backup
- `--output-dir` - Backup directory
- `--compress` - Compress backup files

### `gz dev-env restore`

Restore development environment configurations.

```bash
gz dev-env restore --backup <backup-path> [flags]
```

**Flags:**
- `--backup` - Backup file path (required)
- `--environments` - Specific environments to restore
- `--merge` - Merge with existing configurations

## Configuration

```yaml
version: "1.0"

# AWS configuration
aws:
  profiles:
    - name: "development"
      region: "us-west-2"
      output: "json"
    - name: "production"
      region: "us-east-1"
      output: "json"
  default_profile: "development"

# Docker configuration
docker:
  contexts:
    - name: "local"
      endpoint: "unix:///var/run/docker.sock"
    - name: "remote"
      endpoint: "tcp://docker.company.com:2376"
  registries:
    - name: "docker.io"
      username: "${DOCKER_USERNAME}"
    - name: "company-registry.com"
      username: "${COMPANY_DOCKER_USER}"

# Kubernetes configuration
kubernetes:
  contexts:
    - name: "development"
      cluster: "dev-cluster"
      namespace: "default"
    - name: "staging"
      cluster: "staging-cluster"
      namespace: "staging"
  default_context: "development"

# SSH configuration
ssh:
  hosts:
    - name: "bastion"
      hostname: "bastion.company.com"
      user: "deploy"
      key: "~/.ssh/company_rsa"
    - name: "dev-server"
      hostname: "dev.company.com"
      user: "developer"
      proxy_jump: "bastion"
```

## Environment-Specific Examples

### AWS Management

```bash
# List AWS profiles
gz dev-env list aws

# Switch to production profile
gz dev-env switch aws production

# Configure new profile
gz dev-env configure aws --interactive

# Sync AWS credentials
gz dev-env sync aws
```

### Docker Environment

```bash
# List Docker contexts
gz dev-env list docker

# Switch to remote Docker
gz dev-env switch docker remote

# Configure registry authentication
gz dev-env configure docker --registry company-registry.com
```

### Kubernetes Management

```bash
# List Kubernetes contexts
gz dev-env list kubernetes

# Switch to staging cluster
gz dev-env switch kubernetes staging

# Import kubeconfig
gz dev-env configure kubernetes --import ~/.kube/staging-config
```

### SSH Configuration

```bash
# List SSH hosts
gz dev-env list ssh

# Add new SSH host
gz dev-env configure ssh --add-host production-server

# Generate SSH key pair
gz dev-env configure ssh --generate-key company-key
```

## Integration Examples

### Multi-Environment Development

```bash
# Morning setup - sync all environments
gz dev-env sync --all

# Switch to development environment
gz dev-env switch aws development
gz dev-env switch kubernetes dev-cluster
gz dev-env switch docker local

# Verify configurations
gz dev-env list --status active
```

### Environment Transitions

```bash
# Switch from development to staging
gz dev-env switch aws staging
gz dev-env switch kubernetes staging-cluster

# Deploy to staging
kubectl apply -f deployment.yaml

# Switch back to development
gz dev-env switch aws development
gz dev-env switch kubernetes dev-cluster
```

## Backup and Restore

```bash
# Create backup before major changes
gz dev-env backup --output-dir ./backups

# Make configuration changes...

# Restore if needed
gz dev-env restore --backup ./backups/dev-env-backup-20250804.tar.gz
```

## Related Commands

- [`gz pm`](pm.md) - Package manager management
- [`gz net-env`](net-env.md) - Network environment transitions

## See Also

- [Development Environment Examples](../../examples/dev-env/)
- [AWS Profiles Guide](../03-core-features/development-environment/aws-profiles.md)
- [GCP Projects Guide](../03-core-features/development-environment/gcp-projects.md)
