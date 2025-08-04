# pm Command Reference

Package manager updates and management across multiple platforms and languages.

## Synopsis

```bash
gz pm <action> [flags]
gz pm <action> --config <config-file>
```

## Description

The `pm` command manages and updates packages across multiple package managers, including system-level, language-specific, and development tools.

## Supported Package Managers

- **System**: Homebrew (macOS/Linux), apt, yum, pacman
- **Version Managers**: asdf, SDKMAN
- **Language-Specific**: npm, pip, gem, cargo, go modules
- **Development**: Docker, kubectl

## Actions

### `gz pm update`

Update packages across all or specific package managers.

```bash
gz pm update [flags]
```

**Flags:**
- `--managers` - Specific managers to update (comma-separated)
- `--all` - Update all detected managers (default: true)
- `--dry-run` - Show what would be updated without executing
- `--parallel` - Run updates in parallel (default: false)
- `--skip-system` - Skip system package managers
- `--skip-user` - Skip user-level package managers

**Examples:**
```bash
# Update all package managers
gz pm update --all

# Update specific managers
gz pm update --managers homebrew,npm,pip

# Dry run to see pending updates
gz pm update --dry-run

# Parallel updates (faster but more resource intensive)
gz pm update --parallel
```

### `gz pm list`

List package managers and their status.

```bash
gz pm list [flags]
```

**Flags:**
- `--managers` - Filter by specific managers
- `--status` - Filter by status: available, installed, outdated
- `--output` - Output format: table, json, yaml

**Examples:**
```bash
# List all package managers
gz pm list

# Show only installed managers
gz pm list --status installed

# JSON output
gz pm list --output json
```

### `gz pm install`

Install packages or package managers.

```bash
gz pm install <manager> [packages...] [flags]
```

**Arguments:**
- `manager` - Package manager name (required)
- `packages` - Package names to install (optional)

**Flags:**
- `--global` - Install packages globally
- `--version` - Specific version to install
- `--force` - Force installation/reinstallation

**Examples:**
```bash
# Install a package manager
gz pm install asdf

# Install packages via specific manager
gz pm install npm typescript eslint prettier

# Install specific version
gz pm install python --version 3.11.0 --manager asdf

# Global installation
gz pm install npm --global typescript
```

### `gz pm check`

Check for available updates without installing.

```bash
gz pm check [flags]
```

**Flags:**
- `--managers` - Check specific managers
- `--output` - Output format: table, json, summary

**Examples:**
```bash
# Check all managers for updates
gz pm check

# Check specific managers
gz pm check --managers homebrew,npm

# Summary output
gz pm check --output summary
```

### `gz pm clean`

Clean package manager caches and temporary files.

```bash
gz pm clean [flags]
```

**Flags:**
- `--managers` - Clean specific managers
- `--cache` - Clean only caches (default: true)
- `--temp` - Clean temporary files
- `--all` - Clean everything including unused packages

**Examples:**
```bash
# Clean all caches
gz pm clean

# Clean specific manager caches
gz pm clean --managers npm,pip

# Deep clean including unused packages
gz pm clean --all
```

### `gz pm doctor`

Diagnose package manager health and configuration.

```bash
gz pm doctor [flags]
```

**Flags:**
- `--managers` - Check specific managers
- `--fix` - Attempt to fix detected issues
- `--output` - Output format: text, json

**Examples:**
```bash
# Check all package managers
gz pm doctor

# Fix detected issues
gz pm doctor --fix

# Check specific manager
gz pm doctor --managers homebrew
```

## Configuration

### Global Configuration

```yaml
version: "1.0"

# Global settings
parallel: false
auto_cleanup: true
update_interval: "24h"

# Manager-specific settings
managers:
  homebrew:
    enabled: true
    auto_update: true
    cleanup_after_update: true
    cask_upgrade: true

  npm:
    enabled: true
    global_packages:
      - typescript
      - eslint
      - prettier
    check_security: true

  pip:
    enabled: true
    upgrade_pip: true
    global_packages:
      - black
      - ruff
      - mypy

  asdf:
    enabled: true
    plugins:
      - nodejs
      - python
      - golang
    auto_install_missing: true

  sdkman:
    enabled: true
    candidates:
      - java
      - kotlin
      - gradle
    auto_env: true
```

### Manager-Specific Configurations

```yaml
# Homebrew configuration
homebrew:
  formulae:
    - git
    - jq
    - ripgrep
    - fd
  casks:
    - visual-studio-code
    - docker
    - postman
  taps:
    - homebrew/cask-fonts

# npm configuration
npm:
  global_packages:
    - "@angular/cli"
    - "create-react-app"
    - "typescript"
    - "eslint"
  registries:
    default: "https://registry.npmjs.org/"

# pip configuration
pip:
  global_packages:
    - black
    - isort
    - ruff
    - mypy
    - pytest
  index_url: "https://pypi.org/simple"
```

## Environment Variables

```bash
# Package manager paths
export HOMEBREW_PREFIX="/opt/homebrew"
export ASDF_DIR="$HOME/.asdf"
export SDKMAN_DIR="$HOME/.sdkman"

# Configuration
export GZ_PM_CONFIG="~/.config/gzh-manager/pm.yaml"
export GZ_PM_PARALLEL="false"
export GZ_PM_AUTO_CLEANUP="true"

# Manager-specific
export HOMEBREW_NO_AUTO_UPDATE="1"
export NPM_CONFIG_FUND="false"
export PIP_REQUIRE_VIRTUALENV="false"
```

## Package Manager Details

### Homebrew

```bash
# Update Homebrew and all packages
gz pm update --managers homebrew

# Install specific formulae
gz pm install homebrew git jq ripgrep

# Install casks
gz pm install homebrew --cask visual-studio-code docker
```

### asdf

```bash
# Install asdf plugins and versions
gz pm install asdf nodejs python golang

# Update all asdf packages
gz pm update --managers asdf

# Install specific version
gz pm install asdf nodejs --version 18.17.0
```

### npm

```bash
# Update npm and global packages
gz pm update --managers npm

# Install global packages
gz pm install npm --global typescript eslint prettier

# Check for security vulnerabilities
gz pm check --managers npm --security
```

### pip

```bash
# Update pip and packages
gz pm update --managers pip

# Install packages
gz pm install pip black ruff mypy

# Upgrade pip itself
gz pm update --managers pip --upgrade-pip
```

## Examples

### Daily Development Setup

```bash
# Morning routine - update everything
gz pm update --all

# Check for issues
gz pm doctor

# Clean up caches
gz pm clean --cache
```

### New Machine Setup

```bash
# Install essential package managers
gz pm install homebrew
gz pm install asdf

# Install development tools
gz pm install homebrew git jq ripgrep fd
gz pm install npm --global typescript eslint prettier
gz pm install pip black ruff mypy pytest

# Verify everything is working
gz pm doctor
```

### CI/CD Integration

```bash
# Install dependencies in CI
gz pm install --managers npm,pip --cache-dir /cache

# Check for security issues
gz pm check --security --fail-on-vulnerabilities

# Update lockfiles
gz pm update --lock-files-only
```

### Project-Specific Setup

```bash
# Install project dependencies based on config
gz pm install --config project-pm.yaml

# Update development dependencies
gz pm update --dev-only

# Clean unused packages
gz pm clean --unused
```

## Troubleshooting

### Common Issues

1. **Permission Errors**
   ```bash
   # Fix permissions
   gz pm doctor --fix --managers homebrew

   # Use sudo for system packages
   sudo gz pm update --managers apt
   ```

2. **Network Issues**
   ```bash
   # Use alternative registries
   gz pm update --registry https://registry.npm.taobao.org

   # Check connectivity
   gz pm doctor --check-network
   ```

3. **Version Conflicts**
   ```bash
   # Check for conflicts
   gz pm doctor --check-conflicts

   # Force resolution
   gz pm install package --force --version latest
   ```

### Debug Mode

```bash
# Enable debug logging
gz pm update --debug

# Verbose output
gz pm update --verbose

# Dry run with details
gz pm update --dry-run --verbose
```

## Related Commands

- [`gz quality install`](quality.md#install) - Install code quality tools
- [`gz dev-env`](dev-env.md) - Development environment management

## See Also

- [Package Manager Examples](../../examples/pm/)
- [Development Environment Guide](../03-core-features/development-environment/)
