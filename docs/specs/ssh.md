# SSH Configuration Management Specification

## Overview

The `ssh-config` command provides comprehensive SSH configuration management for Git operations across multiple hosting services. It helps generate and manage SSH configurations for GitHub, GitLab, Gitea, and other Git services, managing SSH keys for different organizations and services.

## Commands

### Core Commands

- `gz ssh-config generate` - Generate SSH configuration for Git operations
- `gz ssh-config validate` - Validate SSH configuration files

### SSH Configuration Generation (`gz ssh-config generate`)

**Purpose**: Generate SSH configurations for Git operations

**Features**:
- Generate ~/.ssh/config entries for different Git hosting services
- Manage SSH keys for different organizations and services
- Support for multiple Git hosting providers
- Custom SSH configuration templates
- Environment-specific SSH configurations

**Usage**:
```bash
gz ssh-config generate                    # Generate SSH config from default config
gz ssh-config generate --config gzh.yaml  # Generate from specific config
gz ssh-config generate --output ~/.ssh/config # Output to specific file
gz ssh-config generate --dry-run         # Preview without applying
```

### SSH Configuration Validation (`gz ssh-config validate`)

**Purpose**: Validate SSH configuration files and connectivity

**Features**:
- Validate SSH configuration syntax
- Test SSH connectivity to Git hosting services
- Verify SSH key permissions and formats
- Check SSH agent configuration
- Validate host-specific configurations

**Usage**:
```bash
gz ssh-config validate                    # Validate default SSH config
gz ssh-config validate --config gzh.yaml  # Validate from specific config
gz ssh-config validate --test-connection  # Test SSH connectivity
gz ssh-config validate --verbose         # Show detailed validation info
```

## Configuration

### Global Configuration

SSH configurations are managed through:
- `~/.config/gzh-manager/ssh-config.yaml` - User-specific SSH settings
- `/etc/gzh-manager/ssh-config.yaml` - System-wide SSH settings
- Environment variable: `GZH_SSH_CONFIG`
- Integration with `gzh.yaml` configuration files

### Configuration Structure

```yaml
# SSH Configuration
ssh_config:
  # Global SSH settings
  global:
    identity_file: "~/.ssh/id_rsa"
    user: "git"
    port: 22
    strict_host_key_checking: "yes"
    user_known_hosts_file: "~/.ssh/known_hosts"

  # Host-specific configurations
  hosts:
    # GitHub configuration
    github.com:
      hostname: "github.com"
      user: "git"
      identity_file: "~/.ssh/github_rsa"
      preferred_authentications: "publickey"

    # GitLab configuration
    gitlab.com:
      hostname: "gitlab.com"
      user: "git"
      identity_file: "~/.ssh/gitlab_rsa"
      preferred_authentications: "publickey"

    # Corporate GitHub Enterprise
    github.company.com:
      hostname: "github.company.com"
      user: "git"
      identity_file: "~/.ssh/company_github_rsa"
      port: 22
      strict_host_key_checking: "yes"

    # Corporate GitLab instance
    gitlab.company.com:
      hostname: "gitlab.company.com"
      user: "git"
      identity_file: "~/.ssh/company_gitlab_rsa"
      port: 22
      proxy_command: "ssh -W %h:%p proxy.company.com"

    # Gitea instance
    gitea.company.com:
      hostname: "gitea.company.com"
      user: "git"
      identity_file: "~/.ssh/company_gitea_rsa"
      port: 22

  # SSH Key Management
  keys:
    # Default key
    default:
      path: "~/.ssh/id_rsa"
      type: "rsa"
      bits: 4096

    # GitHub key
    github:
      path: "~/.ssh/github_rsa"
      type: "ed25519"

    # GitLab key
    gitlab:
      path: "~/.ssh/gitlab_rsa"
      type: "ed25519"

    # Corporate keys
    corporate:
      github:
        path: "~/.ssh/company_github_rsa"
        type: "rsa"
        bits: 4096
      gitlab:
        path: "~/.ssh/company_gitlab_rsa"
        type: "rsa"
        bits: 4096

  # SSH Agent Configuration
  agent:
    enabled: true
    auto_add: true
    confirm: false
    lifetime: "1h"

  # Proxy Configuration
  proxy:
    # Corporate proxy for external Git services
    corporate:
      enabled: false
      host: "proxy.company.com"
      port: 8080
      user: "${PROXY_USER}"
      password: "${PROXY_PASSWORD}"

  # Validation Settings
  validation:
    test_connections: true
    check_key_permissions: true
    verify_host_keys: true
    timeout: "30s"
```

### Environment Variables

- `GZH_SSH_CONFIG` - Path to SSH configuration file
- `SSH_AUTH_SOCK` - SSH agent socket path
- `SSH_AGENT_PID` - SSH agent process ID
- `GZH_SSH_KEY_PATH` - Override default SSH key path
- `GZH_SSH_USER` - Override default SSH user

## Examples

### Basic SSH Configuration Generation

```bash
# Generate SSH config from gzh.yaml
gz ssh-config generate --config gzh.yaml

# Generate SSH config and output to file
gz ssh-config generate --output ~/.ssh/config

# Generate SSH config with dry-run
gz ssh-config generate --dry-run

# Generate SSH config for specific provider
gz ssh-config generate --provider github
```

### SSH Configuration Validation

```bash
# Validate SSH configuration
gz ssh-config validate

# Validate and test connections
gz ssh-config validate --test-connection

# Validate with verbose output
gz ssh-config validate --verbose

# Validate specific configuration file
gz ssh-config validate --config ssh-config.yaml
```

### SSH Key Management

```bash
# Generate new SSH key for GitHub
ssh-keygen -t ed25519 -C "your_email@example.com" -f ~/.ssh/github_rsa

# Add SSH key to agent
ssh-add ~/.ssh/github_rsa

# Test SSH connection to GitHub
ssh -T git@github.com

# Test SSH connection to GitLab
ssh -T git@gitlab.com
```

### Corporate Environment Configuration

```bash
# Generate SSH config for corporate environment
gz ssh-config generate --config corporate-gzh.yaml

# Validate corporate SSH configuration
gz ssh-config validate --config corporate-gzh.yaml --test-connection

# Check SSH agent status
ssh-add -l

# Add corporate SSH keys to agent
ssh-add ~/.ssh/company_github_rsa
ssh-add ~/.ssh/company_gitlab_rsa
```

## Integration Points

- **Repository Management**: Integrates with `synclone` for Git authentication
- **Development Environment**: Coordinates with `dev-env` for SSH configuration management
- **Network Management**: Works with `net-env` for proxy-aware SSH configurations
- **Configuration Generation**: Generates SSH configurations through `gen-config`

## Security Considerations

- **Key Permissions**: Ensure SSH keys have proper file permissions (600)
- **Agent Security**: Secure SSH agent configuration and key management
- **Host Key Verification**: Verify host keys to prevent man-in-the-middle attacks
- **Key Encryption**: Use passphrase-protected SSH keys for sensitive environments
- **Key Rotation**: Regular SSH key rotation policies
- **Audit Logging**: Complete logging of SSH configuration changes and access

## Platform Support

- **Linux**: Full SSH support with OpenSSH
- **macOS**: Full SSH support with OpenSSH
- **Windows**: Support through Windows Subsystem for Linux (WSL) or Git for Windows
- **Container Environments**: SSH configuration in Docker containers
- **Cloud Platforms**: SSH access to cloud instances and services
