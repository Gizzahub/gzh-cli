# gzh-manager-go Examples

This directory contains example configurations and usage patterns for the gz CLI tool, organized by functionality.

## 📁 Directory Structure

### 🔄 synclone/
Modern repository synchronization configurations (recommended approach):
- **`synclone-simple.yaml`** - Minimal configuration for synclone operations
- **`synclone-example.yaml`** - Comprehensive synclone configuration with all options
- **`synclone.yml`** - Advanced synclone configuration features

### 📦 bulk-clone/
Legacy bulk clone configurations (deprecated, use synclone instead):
- **`bulk-clone-simple.yaml`** - Minimal configuration for cloning a single organization
- **`bulk-clone-example.yaml`** - Comprehensive bulk clone configuration with all options
- **`bulk-clone.yml`** - Advanced bulk clone configuration features
- **`bulk-clone.home.yaml`** - Personal/home environment configuration
- **`bulk-clone.work.yaml`** - Work environment configuration

### 🎯 gzh-unified/
Unified GZH configurations combining multiple features:
- **`gzh-simple.yaml`** - Basic unified configuration example
- **`gzh-unified-example.yaml`** - Complete unified configuration showing all features
- **`gzh-multi-provider.yaml`** - Multi-provider setup (GitHub, GitLab, Gitea)
- **`gzh-development.yaml`** - Development environment setup
- **`gzh-enterprise.yaml`** - Enterprise deployment configuration

### 🤖 automation/
GitHub automation and rule configurations:
- **`automation-rule-example.yaml`** - GitHub automation rules configuration
- **`automation-rule-templates.yaml`** - Reusable automation templates

### 🪝 webhooks/
Webhook and organization configuration:
- **`webhook-policy-example.yaml`** - Webhook policy configuration
- **`org-webhook-config-example.yaml`** - Organization-wide webhook settings

### 🌐 network/
Network and VPN configuration:
- **`vpn-hierarchy-example.yaml`** - VPN hierarchy and network configuration

### 🐙 github/
GitHub-specific configurations and schemas:
- **`org-settings.yaml`** - Organization settings template
- **`repo-settings.yaml`** - Repository settings template
- **`schema.org-settings.yaml`** - Schema for organization settings
- **`schema.repo-settings.yaml`** - Schema for repository settings

### 🔧 misc/
Miscellaneous scripts and configurations:
- **`clone-workflow.sh`** - Example cloning workflow script
- **`Dockerfile.example`** - Example Docker configuration

## 🚀 Quick Start

### Synclone (Recommended)

1. Copy the synclone example file:
   ```bash
   cp examples/synclone/synclone-simple.yaml ~/.config/gzh-manager/synclone.yaml
   ```

2. Edit the configuration with your settings:
   ```bash
   # Set your GitHub token
   export GITHUB_TOKEN="your-token-here"

   # Or add to the config file
   vim ~/.config/gzh-manager/synclone.yaml
   ```

3. Run synclone:
   ```bash
   gz synclone github -o your-org -t ~/repos --cleanup-orphans
   ```

### Legacy Bulk-Clone

1. Copy the bulk-clone example file:
   ```bash
   cp examples/bulk-clone/bulk-clone-simple.yaml ~/.config/gzh-manager/bulk-clone.yaml
   ```

2. Edit the configuration with your settings:
   ```bash
   # Set your GitHub token
   export GZH_GITHUB_TOKEN="your-token-here"

   # Or add to the config file
   vim ~/.config/gzh-manager/bulk-clone.yaml
   ```

3. Run the command:
   ```bash
   gz bulk-clone
   ```

## 📋 Configuration Precedence

The gz tool loads configuration in the following order (highest to lowest priority):

1. Command-line flags
2. Environment variables (GZH\_\* prefix)
3. Config file in current directory (./bulk-clone.yaml)
4. User config (~/.config/gzh-manager/bulk-clone.yaml)
5. System config (/etc/gzh-manager/bulk-clone.yaml)

## 🔑 Environment Variables

All configurations support environment variable overrides:

```bash
# Authentication
export GZH_GITHUB_TOKEN="ghp_..."
export GZH_GITLAB_TOKEN="glpat-..."
export GZH_GITEA_TOKEN="..."

# Paths
export GZH_TARGET_PATH="~/repos"
export GZH_CONFIG_PATH="~/my-configs/gzh.yaml"

# Behavior
export GZH_CLONE_STRATEGY="reset"  # reset, pull, fetch
export GZH_CONCURRENCY="10"
```

## 📖 Examples by Use Case

### Personal Repository Management

Use synclone for modern repository management:
```bash
gz synclone github -o your-username -t ~/repos --config examples/synclone/synclone-simple.yaml
```

### Enterprise Deployment

For enterprise environments with multiple teams and platforms:
```bash
gz bulk-clone --config examples/gzh-unified/gzh-enterprise.yaml
```

### CI/CD Integration

For automated environments, use environment variables:
```bash
export GZH_GITHUB_TOKEN="${GITHUB_TOKEN}"
export GZH_DRY_RUN="true"
gz bulk-clone --config examples/bulk-clone/bulk-clone-simple.yaml
```

### Multi-Platform Setup

Clone from multiple providers simultaneously:
```bash
gz bulk-clone --config examples/gzh-unified/gzh-multi-provider.yaml
```

## ✅ Validation

Validate your configuration before use:

```bash
gz bulk-clone validate --config my-config.yaml
```

## 📚 Schema Documentation

- JSON Schema: `docs/bulk-clone-schema.json`
- YAML Schema: `docs/bulk-clone-schema.yaml`

## 💡 Tips

1. **Start Simple**: Begin with simple configurations and add features as needed
2. **Use Dry Run**: Always test with `--dry-run` flag first
3. **Check Logs**: Enable debug logging with `--log-level debug`
4. **Validate First**: Use `gz bulk-clone validate` to check configuration
5. **Version Control**: Keep your configurations in version control
6. **Use Synclone**: Prefer synclone over bulk-clone for new projects

## 🆘 Need Help?

- Run `gz help bulk-clone` for command documentation
- Check `docs/` directory for detailed guides
- Run `gz doctor` to diagnose configuration issues
- Visit the [GitHub repository](https://github.com/gizzahub/gzh-manager-go) for issues and discussions
