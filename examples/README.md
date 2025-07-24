# gzh-manager-go Examples

This directory contains example configurations and usage patterns for the gz CLI tool.

## Configuration Examples

### Basic Usage

- **`synclone-simple.yaml`** - Minimal configuration for synclone operations
- **`bulk-clone-simple.yaml`** - Legacy: Minimal configuration for cloning a single organization
- **`gzh-simple.yaml`** - Basic unified configuration example

### Advanced Configurations

- **`synclone-example.yaml`** - Comprehensive synclone configuration with all options
- **`synclone.yml`** - Advanced synclone configuration features
- **`bulk-clone-example.yaml`** - Legacy: Comprehensive bulk clone configuration with all options
- **`gzh-unified-example.yaml`** - Complete unified configuration showing all features
- **`gzh-multi-provider.yaml`** - Multi-provider setup (GitHub, GitLab, Gitea)

### Environment-Specific

- **`bulk-clone.home.yaml`** - Personal/home environment configuration
- **`bulk-clone.work.yaml`** - Work environment configuration
- **`gzh-development.yaml`** - Development environment setup
- **`gzh-enterprise.yaml`** - Enterprise deployment configuration

### Feature-Specific

- **`automation-rule-example.yaml`** - GitHub automation rules configuration
- **`automation-rule-templates.yaml`** - Reusable automation templates
- **`webhook-policy-example.yaml`** - Webhook policy configuration
- **`org-webhook-config-example.yaml`** - Organization-wide webhook settings
- **`vpn-hierarchy-example.yaml`** - VPN hierarchy and network configuration

## Quick Start

### Synclone (Recommended)

1. Copy the synclone example file:

   ```bash
   cp examples/synclone-simple.yaml ~/.config/gzh-manager/synclone.yaml
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

### Legacy Bulk-Clone (Deprecated)

**Note**: The bulk-clone command has been replaced by synclone. Please use synclone for new projects.

1. Copy the bulk-clone example file:

   ```bash
   cp examples/bulk-clone-simple.yaml ~/.config/gzh-manager/bulk-clone.yaml
   ```

2. Edit the configuration with your settings:

   ```bash
   # Set your GitHub token
   export GZH_GITHUB_TOKEN="your-token-here"

   # Or add to the config file
   vim ~/.config/gzh-manager/bulk-clone.yaml
   ```

3. **Command no longer available** - use `gz synclone` instead

## Configuration Precedence

The gz tool loads configuration in the following order (highest to lowest priority):

### For Synclone (Recommended)
1. Command-line flags
2. Environment variables (GZH\_\* prefix)
3. Config file in current directory (./synclone.yaml)
4. User config (~/.config/gzh-manager/synclone.yaml)
5. System config (/etc/gzh-manager/synclone.yaml)

### For Legacy Bulk-Clone (Deprecated)
1. Command-line flags
2. Environment variables (GZH\_\* prefix)
3. Config file in current directory (./bulk-clone.yaml)
4. User config (~/.config/gzh-manager/bulk-clone.yaml)
5. System config (/etc/gzh-manager/bulk-clone.yaml)

## Environment Variables

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

## Examples by Use Case

### Personal Repository Management

Use `synclone-simple.yaml` as a starting point for managing personal projects:

```bash
gz synclone github -o your-username -t ~/personal --config examples/synclone-simple.yaml
```

### Enterprise Deployment

For enterprise environments with multiple teams and platforms:

```bash
gz synclone github -o company-org -t ~/work --config examples/synclone-example.yaml
```

### CI/CD Integration

For automated environments, use environment variables:

```bash
export GITHUB_TOKEN="${GITHUB_TOKEN}"
export GZH_DRY_RUN="true"
gz synclone github -o your-org -t ~/repos --config examples/synclone-simple.yaml
```

### Multi-Platform Setup

Use synclone with different providers:

```bash
# GitHub organization
gz synclone github -o your-org -t ~/repos/github

# GitLab group (when implemented)
# gz synclone gitlab -g your-group -t ~/repos/gitlab
```

## Validation

Validate your configuration before use:

```bash
# For synclone (recommended)
gz synclone validate --config my-config.yaml

# For legacy bulk-clone (deprecated)  
gz bulk-clone validate --config my-config.yaml
```

## Schema Documentation

- Synclone Schema: `docs/synclone-schema.json` (planned)
- Legacy Bulk-Clone Schema: `docs/bulk-clone-schema.json`
- Legacy YAML Schema: `docs/bulk-clone-schema.yaml`

## Tips

1. **Start Simple**: Begin with `synclone-simple.yaml` and add features as needed
2. **Use Dry Run**: Always test with `--dry-run` flag first
3. **Check Logs**: Enable debug logging with `--log-level debug`
4. **Generate gzh.yaml**: synclone automatically generates repository manifests
5. **Version Control**: Keep your configurations in version control

## Need Help?

- Run `gz help synclone` for command documentation
- Check `docs/` directory for detailed guides
- Run `gz doctor` to diagnose configuration issues
