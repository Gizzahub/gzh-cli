# 🔄 Migration Guides

Comprehensive guides for migrating from legacy systems, upgrading between gzh-cli versions, and transitioning from other tools.

## 📋 Table of Contents

- [Version Upgrades](#version-upgrades)
- [Legacy System Migration](#legacy-system-migration)
- [Tool Migration](#tool-migration)
- [Configuration Migration](#configuration-migration)
- [Breaking Changes](#breaking-changes)

## 🆙 Version Upgrades

### Upgrading from v1.x to v2.x

#### Major Changes
- **Configuration Format**: Unified `gzh.yaml` replaces multiple config files
- **Command Structure**: Reorganized command hierarchy with new subcommands
- **Provider System**: New multi-platform provider architecture
- **Output Formats**: Enhanced output formatting with additional formats

#### Migration Steps

1. **Backup Current Configuration**
   ```bash
   # Backup existing configurations
   cp ~/.config/gzh-manager/synclone.yaml ~/.config/gzh-manager/synclone.yaml.backup
   cp ~/.config/gzh-manager/repo-config.yaml ~/.config/gzh-manager/repo-config.yaml.backup
   ```

2. **Install New Version**
   ```bash
   # Update to latest version
   curl -L https://github.com/gizzahub/gzh-cli/releases/latest/download/install.sh | bash
   ```

3. **Migrate Configuration**
   ```bash
   # Use built-in migration tool
   gz config migrate --from-version v1.x --to-version v2.x

   # Validate new configuration
   gz config validate
   ```

4. **Update Commands**
   ```bash
   # Old v1.x commands → New v2.x commands
   gzh synclone       → gz synclone
   gzh repo-config    → gz repo-config
   gzh quality        → gz quality
   ```

#### Configuration Migration

**Old Format (v1.x)**:
```yaml
# ~/.config/gzh-manager/synclone.yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "~/repos/github/myorg"
```

**New Format (v2.x)**:
```yaml
# ~/.config/gzh-manager/gzh.yaml
global:
  clone_base_dir: "~/repos"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "~/repos/github/myorg"
        strategy: "reset"
```

### Upgrading from v0.x to v1.x

#### Major Changes
- **Command Renaming**: `gzh-cli` → `gz`
- **New Features**: Quality management, IDE integration
- **Enhanced Logging**: Structured logging with multiple levels

#### Migration Steps

1. **Uninstall Old Version**
   ```bash
   # Remove old binary
   rm /usr/local/bin/gzh-cli

   # Clean old configurations if needed
   rm -rf ~/.gzh-cli/
   ```

2. **Install v1.x**
   ```bash
   # Install new version
   make install
   ```

3. **Recreate Configuration**
   ```bash
   # Create new configuration structure
   mkdir -p ~/.config/gzh-manager
   gz config init
   ```

## 🏗️ Legacy System Migration

### From Manual Git Operations

If you're currently managing repositories manually with git commands:

#### Assessment
```bash
# Assess current repository structure
find ~/repos -name ".git" -type d | wc -l
find ~/repos -name ".git" -type d | head -20
```

#### Migration Strategy
1. **Inventory Current Repositories**
   ```bash
   # Create inventory of existing repos
   find ~/repos -name ".git" -type d -exec dirname {} \; > current-repos.txt
   ```

2. **Configure gzh-cli**
   ```yaml
   # gzh.yaml
   global:
     clone_base_dir: "~/repos"

   providers:
     github:
       token: "${GITHUB_TOKEN}"
       organizations:
         - name: "myorg"
           strategy: "reset"  # Use reset to align with remote
   ```

3. **Sync with gzh-cli**
   ```bash
   # Dry-run to see what would happen
   gz synclone github --org myorg --dry-run

   # Perform actual sync
   gz synclone github --org myorg
   ```

### From Other Git Management Tools

#### From GitHub CLI (gh)
```bash
# Export repository list from gh
gh repo list myorg --limit 1000 --json name,sshUrl > gh-repos.json

# Configure gzh-cli to manage same repositories
gz synclone github --org myorg
```

#### From GitLab CLI (glab)
```bash
# List current GitLab projects
glab repo list --group mygroup

# Configure gzh-cli for GitLab
gz synclone gitlab --org mygroup
```

## 🔧 Tool Migration

### From Terraform to gzh-cli (Repository Management)

While Terraform and gzh-cli serve different purposes, you can migrate repository management:

#### Current Terraform State
```hcl
# terraform/repositories.tf
resource "github_repository" "repos" {
  for_each = var.repositories

  name        = each.key
  description = each.value.description
  private     = each.value.private
}
```

#### Equivalent gzh-cli Management
```yaml
# gzh.yaml
providers:
  github:
    organizations:
      - name: "myorg"
        repositories:
          include_patterns:
            - ".*"  # All repositories
        webhook_management: true
        policy_enforcement: true
```

#### Migration Process
1. **Extract Repository List from Terraform**
   ```bash
   terraform show -json | jq '.values.root_module.resources[] | select(.type=="github_repository") | .values.name'
   ```

2. **Configure gzh-cli**
   ```bash
   # Initialize gzh-cli management
   gz synclone github --org myorg

   # Apply repository policies
   gz repo-config audit --org myorg
   ```

3. **Gradual Migration**
   - Keep Terraform for repository creation
   - Use gzh-cli for ongoing management
   - Migrate policies to gzh-cli over time

### From Custom Scripts to gzh-cli

#### Common Script Patterns

**Repository Sync Script**:
```bash
#!/bin/bash
# Old custom script
for repo in $(curl -s "https://api.github.com/orgs/myorg/repos" | jq -r '.[].clone_url'); do
  git clone "$repo" || git -C "$(basename "$repo" .git)" pull
done
```

**Equivalent gzh-cli Command**:
```bash
gz synclone github --org myorg --strategy pull
```

**Quality Check Script**:
```bash
#!/bin/bash
# Old custom script
find . -name "*.go" -exec golangci-lint run {} \;
find . -name "*.py" -exec black --check {} \;
```

**Equivalent gzh-cli Command**:
```bash
gz quality run --languages go,python
```

## ⚙️ Configuration Migration

### Environment Variables

#### Legacy Environment Variables
```bash
# Old variable names
export GZH_CLI_TOKEN="${GITHUB_TOKEN}"
export GZH_CLI_BASE_DIR="${HOME}/repos"
```

#### New Environment Variables
```bash
# New variable names
export GITHUB_TOKEN="${GITHUB_TOKEN}"
export GZH_CONFIG_PATH="${HOME}/.config/gzh-manager/gzh.yaml"
```

### Configuration File Migration

#### Automated Migration Tool
```bash
# Built-in migration tool
gz config migrate \
  --from-file ~/.gzh-cli/config.yaml \
  --to-file ~/.config/gzh-manager/gzh.yaml \
  --format yaml

# Validate migrated configuration
gz config validate --file ~/.config/gzh-manager/gzh.yaml
```

#### Manual Migration
```bash
# Convert old format manually
cat > ~/.config/gzh-manager/gzh.yaml << 'EOF'
global:
  clone_base_dir: "${HOME}/repos"
  concurrent_jobs: 5
  timeout: "30m"

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        strategy: "reset"
        clone_dir: "${HOME}/repos/github/myorg"
EOF
```

## ⚠️ Breaking Changes

### Version 2.0.0
- **Configuration Format**: Unified YAML structure
- **Command Names**: Standardized command hierarchy
- **Output Format**: New table format as default
- **API Changes**: Provider interface modifications

### Version 1.5.0
- **Strategy Names**: `merge` → `pull`, `force` → `reset`
- **Flag Names**: `--workers` → `--concurrent-jobs`
- **Config Paths**: Multiple files → single `gzh.yaml`

### Version 1.0.0
- **Binary Name**: `gzh-cli` → `gz`
- **Config Location**: `~/.gzh-cli/` → `~/.config/gzh-manager/`
- **Log Format**: Plain text → structured JSON

## 🚨 Troubleshooting Migration Issues

### Common Issues

#### Configuration Not Found
```bash
Error: configuration file not found
```

**Solution**:
```bash
# Check configuration paths
gz config show-paths

# Create default configuration
gz config init

# Specify custom path
gz --config /path/to/config.yaml synclone
```

#### Permission Errors
```bash
Error: insufficient permissions for repository access
```

**Solution**:
```bash
# Verify token permissions
curl -H "Authorization: token ${GITHUB_TOKEN}" https://api.github.com/user

# Update token with required scopes
# Required: repo, admin:org, admin:repo_hook
```

#### Command Not Found
```bash
bash: gz: command not found
```

**Solution**:
```bash
# Check installation
which gz

# Add to PATH
export PATH=$PATH:$GOPATH/bin

# Reinstall if necessary
make install
```

### Migration Validation

#### Post-Migration Checklist
- [ ] Configuration validates successfully
- [ ] All repositories are accessible
- [ ] Commands execute without errors
- [ ] Output format is as expected
- [ ] Performance is acceptable

#### Validation Commands
```bash
# Validate configuration
gz config validate

# Test provider connectivity
gz config test-auth --all

# Verify repository access
gz git repo list --org myorg --limit 5

# Test core functionality
gz synclone github --org myorg --dry-run
```

## 📞 Migration Support

### Self-Help Resources
```bash
# Built-in help
gz --help
gz config migrate --help

# Documentation
gz docs open

# Examples
gz examples list
gz examples show migration
```

### Community Support
- **GitHub Discussions**: Community Q&A
- **GitHub Issues**: Bug reports and feature requests
- **Documentation**: Comprehensive guides and examples

### Professional Support
- **Enterprise Support**: Priority support for enterprise customers
- **Migration Services**: Professional migration assistance
- **Training**: Team training and onboarding

---

**Related Documentation**: [Configuration Guide](../40-configuration/40-configuration-guide.md) | [Troubleshooting](../90-maintenance/90-troubleshooting.md) | [Command Reference](../50-api-reference/50-command-reference.md)
**Migration Tools**: [Configuration Migration](../40-configuration/40-configuration-guide.md#migration) | [Version Upgrade Guide](../60-development/60-index.md#upgrading)
**Support**: For complex migrations, consider professional migration services or enterprise support options.
