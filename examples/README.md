# gzh-manager-go Examples

This directory contains example configurations and usage patterns for the gz CLI tool, organized by functionality.

## 📁 Directory Structure

### Core Features

#### 🔄 synclone/
Repository synchronization configurations:
- **`synclone-simple.yaml`** - Minimal configuration for getting started
- **`synclone-example.yaml`** - Comprehensive configuration with all options
- **`synclone.yml`** - Advanced multi-platform setup
- **`gzh.yml`** - Legacy configuration format (v0.1)

#### ✨ quality/
Code quality management configurations:
- **`quality-simple.yaml`** - Basic formatter and linter setup
- **`quality-example.yaml`** - Full multi-language configuration

#### 💻 ide/
JetBrains IDE monitoring configurations:
- **`ide-monitor.yaml`** - IDE sync monitoring and backup settings

#### 📦 pm/
Package manager configurations:
- **`asdf.yml`** - asdf version manager packages
- **`brew.yml`** - Homebrew packages and casks
- **`npm.yml`** - Node.js global packages
- **`pip.yml`** - Python packages
- **`gem.yml`** - Ruby gems
- **`sdkman.yml`** - SDKMAN Java/Kotlin/Scala tools
- **`global.yml`** - Combined package manager setup

### Git Platform Management

#### 🐙 github/
GitHub-specific configurations and schemas:
- **`org-settings.yaml`** - Organization settings template
- **`repo-settings.yaml`** - Repository settings template
- **`schema.org-settings.yaml`** - Schema for organization settings
- **`schema.repo-settings.yaml`** - Schema for repository settings

#### 🪝 webhooks/
Webhook and event configurations:
- **`webhook-policy-example.yaml`** - Webhook policy configuration
- **`org-webhook-config-example.yaml`** - Organization-wide webhook settings

### Environment Management

#### 🌐 network/
Network environment configurations:
- **`vpn-hierarchy-example.yaml`** - VPN hierarchy and network rules

#### 🔧 dev-env/
Development environment configurations:
- AWS, Docker, Kubernetes configuration examples

### Other Resources

#### 🤖 automation/
Automation rules and templates:
- **`automation-rule-example.yaml`** - GitHub automation rules
- **`automation-rule-templates.yaml`** - Reusable automation templates

#### 🔧 misc/
Miscellaneous resources:
- **`clone-workflow.sh`** - Example automation script
- **`Dockerfile.example`** - Docker deployment example

## 🚀 Quick Start

### 1. Repository Synchronization (synclone)

```bash
# Simple GitHub organization clone
gz synclone github --org kubernetes

# Using configuration file
cp examples/synclone/synclone-simple.yaml ~/.config/gzh-manager/synclone.yaml
gz synclone --config ~/.config/gzh-manager/synclone.yaml
```

### 2. Code Quality Management

```bash
# Install quality tools
gz quality install

# Run formatters and linters
gz quality run

# Use custom configuration
gz quality run --config examples/quality/quality-simple.yaml
```

### 3. IDE Monitoring

```bash
# Start monitoring JetBrains IDEs
gz ide monitor

# Use configuration for advanced features
gz ide monitor --config examples/ide/ide-monitor.yaml
```

### 4. Git Repository Management

```bash
# Smart clone or update
gz git repo clone-or-update https://github.com/user/repo.git

# Configure repository settings
gz git config apply --config examples/github/repo-settings.yaml
```

### 5. Package Manager Updates

```bash
# Update all package managers
gz pm update --all

# Update specific managers
gz pm update --managers brew,npm --config examples/pm/global.yml
```

## 📋 Configuration Precedence

The gz tool loads configuration in the following order (highest to lowest priority):

1. Command-line flags
2. Environment variables
3. Config file specified with --config flag
4. Config file in current directory
5. User config (~/.config/gzh-manager/)
6. System config (/etc/gzh-manager/)

## 🔑 Environment Variables

### Authentication Tokens

```bash
# Git platforms
export GITHUB_TOKEN="ghp_..."
export GITLAB_TOKEN="glpat-..."
export GITEA_TOKEN="..."
export GOGS_TOKEN="..."
```

### Configuration Paths

```bash
# Override default config locations
export GZ_CONFIG_DIR="~/my-configs"
export GZ_SYNCLONE_CONFIG="~/my-synclone.yaml"
export GZ_QUALITY_CONFIG="~/my-quality.yaml"
```

## 📖 Examples by Use Case

### Personal Development Setup

Complete development environment setup:
```bash
# 1. Clone your repositories
gz synclone github --org your-username

# 2. Set up code quality tools
gz quality install
gz quality run

# 3. Monitor IDE settings
gz ide monitor

# 4. Keep packages updated
gz pm update --all
```

### Team/Enterprise Environment

For teams with standards and policies:
```bash
# 1. Clone organization repos with filters
gz synclone --config examples/synclone/synclone-example.yaml

# 2. Apply repository policies
gz git config apply --config examples/github/org-settings.yaml

# 3. Set up webhooks
gz git webhook bulk create --config examples/webhooks/org-webhook-config-example.yaml

# 4. Enable quality checks
gz quality run --config examples/quality/quality-example.yaml
```

### CI/CD Integration

Automated pipeline setup:
```bash
# Set tokens from CI secrets
export GITHUB_TOKEN="${CI_GITHUB_TOKEN}"

# Run quality checks
gz quality run --fail-on-error

# Generate reports
gz quality analyze --output-format sarif > quality-report.sarif
```

### Multi-Platform Repository Management

Working with multiple Git platforms:
```bash
# Use comprehensive synclone config
gz synclone --config examples/synclone/synclone-example.yaml

# This handles GitHub, GitLab, Gitea, and Gogs simultaneously
```

## ✅ Validation

Validate configurations before use:

```bash
# Validate synclone configuration
gz synclone validate --config my-synclone.yaml

# Check quality configuration
gz quality check-config --config my-quality.yaml

# Diagnose system setup
gz doctor
```

## 📚 Schema Documentation

Configuration schemas are available in multiple formats:

- **Synclone**: `docs/04-configuration/schemas/synclone-schema.yaml`
- **Quality**: `docs/04-configuration/schemas/quality-schema.yaml`
- **IDE**: `docs/04-configuration/schemas/ide-schema.yaml`
- **Repo Config**: `docs/04-configuration/schemas/repo-config-schema.yaml`

## 💡 Best Practices

1. **Start Simple**: Use minimal configs first (e.g., `synclone-simple.yaml`)
2. **Test First**: Always use `--dry-run` when available
3. **Use Version Control**: Keep your configurations in git
4. **Environment-Specific**: Use different configs for work/personal/CI
5. **Validate Changes**: Run validation after editing configs
6. **Check Logs**: Use `--debug` for troubleshooting

## 🆘 Getting Help

```bash
# Command help
gz help <command>
gz <command> --help

# List all commands
gz help

# Diagnose issues
gz doctor

# Check version
gz version
```

## 📖 Related Documentation

- [Getting Started Guide](../docs/01-getting-started/)
- [Architecture Overview](../docs/02-architecture/overview.md)
- [Configuration Guide](../docs/04-configuration/)
- [Command Reference](../docs/05-command-reference/)
