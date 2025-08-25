# ⚙️ First Configuration

Set up your initial gzh-cli configuration for optimal workflow.

## 🎯 Configuration Strategy

### Understanding Configuration Priority

gzh-cli uses a priority system for configuration:

1. **Command-line flags** (highest priority)
1. **Environment variables**
1. **Configuration files**
1. **Default values** (lowest priority)

### Configuration Locations

gzh-cli searches for configuration files in this order:

1. `$GZH_CONFIG_PATH` (if set)
1. `./gzh.yaml` (current directory)
1. `~/.config/gzh-manager/gzh.yaml` (user config)
1. `/etc/gzh-manager/gzh.yaml` (system config)

## 📝 Basic Configuration

### Step 1: Create Configuration Directory

```bash
# Create user configuration directory
mkdir -p ~/.config/gzh-manager

# Create project-specific config (optional)
mkdir -p ~/.config/gzh-manager/projects
```

### Step 2: Authentication Setup

#### GitHub Configuration

```bash
# Set environment variable (recommended)
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# Add to shell profile for persistence
echo 'export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"' >> ~/.bashrc
```

#### GitLab Configuration

```bash
# GitLab.com
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"

# Self-hosted GitLab
export GITLAB_API_URL="https://gitlab.company.com/api/v4"
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"
```

#### Gitea/Gogs Configuration

```bash
# Gitea
export GITEA_TOKEN="your_gitea_token"
export GITEA_API_URL="https://gitea.company.com/api/v1"

# Gogs
export GOGS_TOKEN="your_gogs_token"
export GOGS_API_URL="https://gogs.company.com/api/v1"
```

### Step 3: Basic Configuration File

Create `~/.config/gzh-manager/gzh.yaml`:

```yaml
# Global settings
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset
  log_level: info
  output_format: table

# Git platform providers
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "your-org"
        clone_dir: "$HOME/repos/github/your-org"
        visibility: all
        include_forks: false
        include_archived: false

  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.com/api/v4"
    groups:
      - name: "your-group"
        clone_dir: "$HOME/repos/gitlab/your-group"
        visibility: all

# Command-specific settings
commands:
  synclone:
    concurrent_jobs: 5
    retry_attempts: 3

  quality:
    auto_install_tools: true
    default_languages: ["go", "python", "javascript"]

  git:
    default_branch: "main"
    auto_setup_upstream: true
```

## 🔧 Advanced Configuration

### Multiple Organizations/Groups

```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "work-org"
        clone_dir: "$HOME/repos/work/github"
        visibility: private
        strategy: reset

      - name: "personal-org"
        clone_dir: "$HOME/repos/personal/github"
        visibility: all
        strategy: pull
        include_forks: true

      - name: "open-source"
        clone_dir: "$HOME/repos/oss/github"
        visibility: public
        filters:
          topics: ["golang", "cli"]
          min_stars: 10
```

### Repository-Specific Settings

```yaml
providers:
  github:
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/github/myorg"
        repositories:
          - name: "critical-service"
            strategy: rebase
            branch: "develop"
            clone_dir: "$HOME/projects/critical-service"

          - name: "documentation"
            strategy: pull
            exclude: true  # Skip this repository
```

### Network and Performance Settings

```yaml
global:
  # Performance settings
  concurrent_jobs: 10
  timeout: "30m"
  retry_attempts: 3
  retry_delay: "5s"

  # Network settings
  proxy:
    http: "http://proxy.company.com:8080"
    https: "http://proxy.company.com:8080"
    no_proxy: "localhost,127.0.0.1,.company.com"

# Rate limiting for API calls
rate_limiting:
  github:
    requests_per_hour: 4500  # Stay under 5000 limit

  gitlab:
    requests_per_minute: 300
```

## 🧪 Configuration Testing

### Validate Configuration

```bash
# Check configuration syntax and values
gz config validate

# Show effective configuration
gz config show

# Test authentication
gz config test-auth
```

### Test Repository Access

```bash
# Test GitHub organization access
gz synclone github --org your-org --dry-run

# Test single repository clone
gz git repo clone-or-update https://github.com/your-org/test-repo.git --dry-run

# List accessible repositories
gz git repo list --org your-org --limit 5
```

## 🔄 Configuration Templates

### Minimal Single-Platform Setup

```yaml
# ~/.config/gzh-manager/gzh.yaml
global:
  clone_base_dir: "$HOME/repos"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "your-org"
        clone_dir: "$HOME/repos/your-org"
```

### Multi-Platform Enterprise Setup

```yaml
global:
  clone_base_dir: "$HOME/work/repos"
  default_strategy: reset
  concurrent_jobs: 8

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "company"
        clone_dir: "$HOME/work/repos/github"

  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.company.com/api/v4"
    groups:
      - name: "platform"
        clone_dir: "$HOME/work/repos/gitlab/platform"

      - name: "services"
        clone_dir: "$HOME/work/repos/gitlab/services"

  gitea:
    token: "${GITEA_TOKEN}"
    api_url: "https://gitea.company.com/api/v1"
    organizations:
      - name: "infrastructure"
        clone_dir: "$HOME/work/repos/gitea/infra"
```

### Development-Focused Setup

```yaml
global:
  clone_base_dir: "$HOME/dev"
  default_strategy: rebase  # Preserve local changes

providers:
  github:
    organizations:
      - name: "mycompany"
        clone_dir: "$HOME/dev/work"
        visibility: private

      - name: "myusername"
        clone_dir: "$HOME/dev/personal"
        visibility: all
        include_forks: true

commands:
  synclone:
    concurrent_jobs: 3  # Gentle on system

  quality:
    auto_install_tools: true
    run_on_sync: true  # Auto-check quality after sync

  ide:
    auto_monitor: true
    sync_settings: true
```

## 🎨 Output Format Configuration

### Default Output Formats

```yaml
# Set default output formats for commands
output_formats:
  default: table

  # Command-specific defaults
  synclone: table
  quality: json
  git: yaml
  profile: human
```

### CI/CD Integration Format

```yaml
# Optimized for automation
global:
  output_format: json
  log_level: error  # Reduce noise

commands:
  quality:
    output_file: "quality-report.sarif"
    exit_on_failure: true
```

## 📋 Configuration Best Practices

### Security

- ✅ Use environment variables for tokens
- ✅ Set proper file permissions: `chmod 600 ~/.config/gzh-manager/gzh.yaml`
- ❌ Never commit tokens to configuration files

### Performance

- ✅ Adjust `concurrent_jobs` based on your system
- ✅ Use appropriate strategies for different repositories
- ✅ Filter repositories to avoid cloning unnecessary ones

### Organization

- ✅ Use separate clone directories for different organizations
- ✅ Group similar repositories together
- ✅ Use descriptive names for configurations

## 🆘 Configuration Troubleshooting

### Common Issues

#### Token Authentication Fails

```bash
# Test token manually
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Check token scopes (GitHub)
gz config test-auth --provider github --verbose
```

#### Configuration Not Found

```bash
# Check search paths
gz config show-paths

# Verify file exists and has correct name
ls -la ~/.config/gzh-manager/gzh.yaml
```

#### Permission Denied

```bash
# Fix file permissions
chmod 600 ~/.config/gzh-manager/gzh.yaml
chmod 700 ~/.config/gzh-manager/
```

#### Invalid YAML Syntax

```bash
# Validate YAML syntax
gz config validate

# Use a YAML validator
python3 -c "import yaml; yaml.safe_load(open('~/.config/gzh-manager/gzh.yaml'))"
```

## 📖 Next Steps

After setting up your configuration:

1. **[Try Quick Start Commands](11-quick-start.md)** - Test your setup
1. **[Complete Configuration Guide](../40-configuration/40-configuration-guide.md)** - Advanced options
1. **[Feature Guides](../30-features/)** - Detailed feature documentation

______________________________________________________________________

**Configuration File**: `~/.config/gzh-manager/gzh.yaml`
**Validation**: `gz config validate`
**Testing**: `gz config test-auth`
