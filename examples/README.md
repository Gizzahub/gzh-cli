# gzh-manager-go Examples

This directory contains example configurations and usage patterns for the gz CLI tool.

## Configuration Examples

### Basic Usage

- **`bulk-clone-simple.yaml`** - Minimal configuration for cloning a single organization
- **`gzh-simple.yaml`** - Basic unified configuration example

### Advanced Configurations

- **`bulk-clone-example.yaml`** - Comprehensive bulk clone configuration with all options
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

1. Copy the appropriate example file:
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

3. Run the command:
   ```bash
   gz bulk-clone
   ```

## Configuration Precedence

The gz tool loads configuration in the following order (highest to lowest priority):

1. Command-line flags
2. Environment variables (GZH_* prefix)
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

Use `bulk-clone.home.yaml` as a starting point for managing personal projects:

```bash
gz bulk-clone --config examples/bulk-clone.home.yaml
```

### Enterprise Deployment

For enterprise environments with multiple teams and platforms:

```bash
gz bulk-clone --config examples/gzh-enterprise.yaml
```

### CI/CD Integration

For automated environments, use environment variables:

```bash
export GZH_GITHUB_TOKEN="${GITHUB_TOKEN}"
export GZH_DRY_RUN="true"
gz bulk-clone --config examples/bulk-clone-simple.yaml
```

### Multi-Platform Setup

Clone from multiple providers simultaneously:

```bash
gz bulk-clone --config examples/gzh-multi-provider.yaml
```

## Validation

Validate your configuration before use:

```bash
gz bulk-clone validate --config my-config.yaml
```

## Schema Documentation

- JSON Schema: `docs/bulk-clone-schema.json`
- YAML Schema: `docs/bulk-clone-schema.yaml`

## Tips

1. **Start Simple**: Begin with `bulk-clone-simple.yaml` and add features as needed
2. **Use Dry Run**: Always test with `--dry-run` flag first
3. **Check Logs**: Enable debug logging with `--log-level debug`
4. **Validate First**: Use `gz bulk-clone validate` to check configuration
5. **Version Control**: Keep your configurations in version control

## Need Help?

- Run `gz help bulk-clone` for command documentation
- Check `docs/` directory for detailed guides
- Run `gz doctor` to diagnose configuration issues