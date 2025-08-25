# 🚀 Quick Start Guide

Get up and running with gzh-cli in under 10 minutes.

## 🎯 First Commands

### 1. System Check

```bash
# Verify installation and diagnose issues
gz doctor

# Expected output shows system status
```

### 2. Single Repository Management

```bash
# Clone or update a single repository
gz git repo clone-or-update https://github.com/octocat/Hello-World.git

# With custom target path
gz git repo clone-or-update https://github.com/user/repo.git ~/my-projects/repo

# With specific branch
gz git repo clone-or-update https://github.com/user/repo.git -b develop
```

### 3. Multi-Repository Sync

```bash
# GitHub organization sync
gz synclone github --org your-organization

# GitLab group sync
gz synclone gitlab --group your-group

# Multiple platforms with config file
gz synclone --config ~/.config/gzh-manager/synclone.yaml
```

### 4. Code Quality

```bash
# Run quality checks on current directory
gz quality run

# Install quality tools if missing
gz quality install

# Format code only
gz quality format
```

### 5. Development Environment

```bash
# Monitor JetBrains IDE settings
gz ide monitor

# AWS profile management
gz dev-env aws --profile default

# Package manager updates
gz pm update --all
```

## 📋 Essential Configuration

### Minimal Setup

Create `~/.config/gzh-manager/gzh.yaml`:

```yaml
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "your-org"
        clone_dir: "$HOME/repos/github/your-org"

  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "your-group"
        clone_dir: "$HOME/repos/gitlab/your-group"
```

### Token Setup

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc)
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"

# Reload shell configuration
source ~/.bashrc  # or ~/.zshrc
```

## 🎮 Interactive Usage

### Real-time Repository Sync

```bash
# Start with your organization
gz synclone github --org myorg

# Example output:
# ✅ Cloning myorg/project1 → /home/user/repos/github/myorg/project1
# 🔄 Updating myorg/project2 → /home/user/repos/github/myorg/project2
# ⚡ Completed: 15 repos, 12 updated, 3 new
```

### Quality Management Workflow

```bash
# Check code quality
gz quality run

# Fix formatting issues
gz quality format

# Install missing tools
gz quality install --language go,python,javascript
```

### IDE Integration

```bash
# Start IDE monitoring (runs in background)
gz ide monitor &

# Check IDE sync status
gz ide status

# Fix sync issues
gz ide fix-sync
```

## 🔧 Output Formats

gzh-cli supports multiple output formats for integration:

```bash
# JSON output for scripting
gz synclone github --org myorg --output json

# YAML output
gz git repo list --output yaml

# CSV for data analysis
gz quality run --output csv

# Table format (default)
gz pm update --output table
```

## 📊 Common Workflows

### Daily Developer Workflow

```bash
# Morning sync - update all repositories
gz synclone --config ~/.config/gzh-manager/synclone.yaml

# Work on code...

# End of day - quality check
gz quality run
```

### Project Setup Workflow

```bash
# Clone project and dependencies
gz git repo clone-or-update https://github.com/myorg/main-project.git

# Setup development environment
gz dev-env setup

# Start IDE monitoring
gz ide monitor
```

### CI/CD Integration

```bash
# Quality check with SARIF output for security scanning
gz quality run --output sarif --output-file quality.sarif

# Repository compliance check
gz git config audit --org myorg --output json
```

## 🎯 Learning Paths

### 🚶 Beginner (15 minutes)

1. Try single repository clone: `gz git repo clone-or-update`
1. Run quality check: `gz quality run`
1. Check system status: `gz doctor`

### 🏃 Intermediate (30 minutes)

1. Set up configuration file with your organization
1. Sync entire organization: `gz synclone`
1. Explore output formats: `--output json`

### 🏊 Advanced (60 minutes)

1. Configure multiple Git platforms
1. Set up IDE monitoring and dev environment
1. Integrate with CI/CD pipelines

## 🆘 Quick Troubleshooting

### Command Not Found

```bash
# Check installation
which gz
echo $PATH | grep $GOPATH/bin

# Reinstall if needed
cd gzh-cli && make install
```

### Authentication Issues

```bash
# Test GitHub token
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# Test GitLab token
curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" https://gitlab.com/api/v4/user
```

### Configuration Problems

```bash
# Validate configuration
gz config validate

# Check diagnostics
gz doctor

# Reset configuration
rm ~/.config/gzh-manager/gzh.yaml
# Then recreate with minimal setup above
```

## 📖 Next Steps

### Essential Reading

- **[Configuration Guide](../40-configuration/40-configuration-guide.md)** - Complete configuration reference
- **[Command Reference](../50-api-reference/50-command-reference.md)** - All available commands
- **[Features Overview](../30-features/)** - Detailed feature documentation

### Migration from Other Tools

- **[bulk-clone Migration](migration-guides/bulk-clone-to-gzh.md)** - If migrating from bulk-clone
- **[daemon Migration](migration-guides/daemon-to-cli.md)** - If migrating from daemon-based tools

### Advanced Topics

- **[Architecture](../20-architecture/20-system-overview.md)** - Understanding system design
- **[Development](../60-development/)** - Contributing to gzh-cli
- **[Enterprise Features](../80-integrations/enterprise/)** - Enterprise deployment

______________________________________________________________________

**Quick Command**: `gz --help` for command overview
**Validation**: `gz doctor` for system diagnostics
**Configuration**: `gz config validate` for config check
