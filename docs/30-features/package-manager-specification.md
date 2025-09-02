# Unified Package Manager Specification

## Overview

The unified package manager feature provides centralized management for multiple package managers through configuration files. It enables developers to maintain consistent development environments across machines by managing packages from various package managers including system-level (apt, brew, port), language-specific (pip, gem, npm), and version managers (asdf, rbenv, sdkman).

## Implementation Status

### ✅ Fully Implemented

**Configuration-based unified commands** - All managers support: `install`, `update`, `sync`, `export`

Compatibility handling via filter chain:

- Modes: auto|strict|off

- Built-ins: asdf+rust (rustup), asdf+nodejs (corepack), asdf+python (venv), asdf+golang (GOBIN)

- User config: `~/.gzh/pm/compat.yml` (with `when`, `match_env`)

- System package managers: `brew`, `apt`, `port`

- Version managers: `asdf`, `sdkman`, `rbenv`

**Legacy direct access commands** - For backward compatibility:

- `gz pm brew` - Direct Homebrew commands
- `gz pm asdf` - Direct asdf commands
- `gz pm sdkman` - Direct SDKMAN commands
- `gz pm apt` - Direct APT commands
- `gz pm port` - Direct MacPorts commands
- `gz pm rbenv` - Direct rbenv commands
- `gz pm pip` - Direct pip commands (recently implemented)
- `gz pm npm` - Direct npm commands (recently implemented)

### 📋 Configuration-Only

These package managers are supported through configuration files but don't have direct commands:

- `gem` - Ruby packages
- `cargo` - Rust packages
- `go` - Go modules
- `composer` - PHP packages
- `yarn`, `pnpm` - Alternative Node.js package managers

### 🔮 Future Enhancement

- `gz pm [manager]` - Generic pattern for any package manager (see Future Considerations section)

## Supported Package Managers

### System Package Managers

- **brew** (Homebrew) - macOS/Linux
- **port** (MacPorts) - macOS
- **apt** (APT) - Debian/Ubuntu
- **yum** (YUM) - RHEL/CentOS
- **dnf** (DNF) - Fedora
- **pacman** (Pacman) - Arch Linux

### Version Managers

- **asdf** - Multi-language version manager
- **sdkman** - JVM ecosystem version manager
- **rbenv** - Ruby version manager
- **pyenv** - Python version manager
- **nvm** - Node.js version manager

### Language Package Managers

- **pip/uv** - Python
- **gem** - Ruby
- **npm/pnpm/yarn** - Node.js
- **cargo** - Rust
- **go** - Go modules
- **composer** - PHP

## Commands

### Core Unified Commands

```bash
gz pm                       # Show help
gz pm status               # Show status of all configured package managers
gz pm install              # Install packages from configuration files
gz pm update               # Update packages based on version strategy
gz pm sync                 # Synchronize installed packages with configuration
gz pm export               # Export current installations to configuration files
gz pm validate             # Validate configuration files
gz pm clean                # Clean unused packages based on strategy
gz pm bootstrap            # Install and configure package managers
gz pm upgrade-managers     # Upgrade package managers themselves
gz pm sync-versions        # Synchronize version manager and package manager versions
```

### Legacy Direct Access Commands

For backward compatibility and direct package manager access:

```bash
# System package managers
gz pm brew [subcommands]    # Direct Homebrew access
gz pm apt [subcommands]     # Direct APT access
gz pm port [subcommands]    # Direct MacPorts access

# Version managers
gz pm asdf [subcommands]    # Direct asdf access
gz pm sdkman [subcommands]  # Direct SDKMAN access
gz pm rbenv [subcommands]   # Direct rbenv access

# Language package managers
gz pm pip [subcommands]     # Direct pip access
gz pm npm [subcommands]     # Direct npm access
```

### Command Examples

```bash
# Unified commands (recommended)
gz pm install --manager brew
gz pm update --all --strategy stable
gz pm sync --cleanup
gz pm export --manager asdf

# Legacy direct access
gz pm brew install wget
gz pm asdf plugin add nodejs
gz pm pip install requests
gz pm npm install -g typescript

# Bootstrap and maintenance
gz pm bootstrap --check
gz pm bootstrap --install brew,nvm,rbenv
gz pm upgrade-managers --all
gz pm sync-versions --fix
```

## Configuration

### File Structure

```
~/.gzh/pm/
├── global.yml    # Global settings and manager registry
├── brew.yml      # Homebrew packages and casks
├── asdf.yml      # asdf plugins and versions
├── sdkman.yml    # SDKMAN candidates
├── pip.yml       # Python packages
├── npm.yml       # Node.js packages
├── apt.yml       # APT packages
├── port.yml      # MacPorts packages
└── ...           # Other package managers
```

### Global Configuration

`~/.gzh/pm/global.yml`:

```yaml
version: "1.0.0"

defaults:
  strategy: "preserve"          # preserve, strict, latest
  auto_update: false
  backup_before_changes: true
  parallel_operations: true
  max_workers: 4

managers:
  brew:
    enabled: true
    config_file: "brew.yml"
    priority: 1
  asdf:
    enabled: true
    config_file: "asdf.yml"
    priority: 2
  pip:
    enabled: true
    config_file: "pip.yml"
    priority: 3

version_strategies:
  default: "stable"
  development: "latest"
  production: "fixed"

cleanup:
  remove_orphans: false
  remove_unused_deps: true
  keep_cache: false
  dry_run_default: true
```

### Package Manager Configuration Examples

#### Homebrew (`brew.yml`)

```yaml
version: "1.0.0"
platform: "darwin"

settings:
  HOMEBREW_PREFIX: "/opt/homebrew"
  HOMEBREW_CASK_OPTS: "--appdir=/Applications"

strategy:
  default: "latest"
  update_frequency: "weekly"
  auto_cleanup: true

taps:
  - homebrew/core
  - homebrew/cask

formulae:
  - name: git
    version: "latest"
  - name: node
    version: "latest"
    link: true

casks:
  - name: visual-studio-code
    version: "latest"
  - name: docker
    version: "latest"
```

#### ASDF (`asdf.yml`)

```yaml
version: "1.0.0"

plugins:
  nodejs:
    repository: "https://github.com/asdf-vm/asdf-nodejs.git"
    versions:
      - version: "20.11.0"
        global: true
      - version: "18.19.0"
        install: true

  python:
    repository: "https://github.com/asdf-vm/asdf-python.git"
    versions:
      - version: "3.12.1"
        global: true

global_versions:
  nodejs: "20.11.0"
  python: "3.12.1"
```

## Advanced Features

### Package Manager Bootstrap

Automatically install missing package managers:

```yaml
# In global.yml
bootstrap:
  auto_install: true
  check_on_startup: true

  managers:
    brew:
      darwin:
        install_script: "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh"
        check_command: "brew --version"

    nvm:
      install_script: "https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh"
      check_command: "nvm --version"
```

### Version Manager Coordination

Handle version manager complexity:

```yaml
# In global.yml
version_coordination:
  node_npm_sync:
    enabled: true
    strategy: "bundled"
    npm_version_matrix:
      "20.x": "10.x"
      "18.x": "9.x"

  ruby_gem_migration:
    enabled: true
    strategy: "copy"
    backup_before_switch: true
```

### Version Strategies

- **latest**: Always update to the latest version
- **stable**: Latest stable version (exclude pre-release)
- **fixed**: Keep exact version
- **compatible**: Update within version range (e.g., ~1.2.3, ^2.0.0)

### Cleanup Strategies

- **preserve**: Keep all installed packages, only add missing
- **remove**: Remove packages not in configuration
- **quarantine**: Move unmanaged packages to quarantine
- **strict**: Remove anything not explicitly defined

## Multi-Manager Workflow

```bash
# First time setup
gz pm export --all              # Export current state

# New machine setup
gz pm bootstrap                 # Install package managers
gz pm install --all            # Install all packages

# Regular maintenance
gz pm update --strategy stable
gz pm sync --cleanup --dry-run
gz pm clean --force
```

## Platform Support

- **macOS**: brew, port, asdf
- **Linux**: apt/yum/dnf/pacman, brew, asdf
- **Windows**: chocolatey, scoop (future)

## Future Considerations: Generic Package Manager Pattern

### Overview

A generic `gz pm [manager]` pattern has been designed to allow dynamic access to any package manager without explicitly defining each one. This would enable support for new package managers without code changes.

### Proposed Command Structure

```bash
gz pm [manager] [subcommand] [args...]
```

### Technical Design

#### 1. Dynamic Command Registration

- Use cobra's `DisableFlagParsing` to capture all arguments
- Parse the first argument as the package manager name
- Pass remaining arguments to the package manager

#### 2. Package Manager Registry

```go
type PackageManager interface {
    Name() string
    IsInstalled() bool
    Execute(args []string) error
}

var registry = map[string]PackageManager{
    "brew": &BrewManager{},
    "apt":  &AptManager{},
    // etc.
}
```

#### 3. Implementation Example

```go
func newGenericPMCmd(ctx context.Context) *cobra.Command {
    return &cobra.Command{
        Use:                "pm [manager] [args...]",
        Short:              "Execute package manager commands",
        DisableFlagParsing: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) < 1 {
                return fmt.Errorf("specify a package manager")
            }

            manager := args[0]
            pmArgs := args[1:]

            // Check registry first
            if pm, ok := registry[manager]; ok {
                return pm.Execute(pmArgs)
            }

            // Fallback to system command
            return executeSystemCommand(ctx, manager, pmArgs)
        },
    }
}
```

#### 4. Fallback to System Command

If a manager is not in the registry, the system would attempt to execute it as a system command, providing flexibility for new or unknown package managers.

### Benefits

1. **Extensibility**: Support new package managers without code changes
1. **Flexibility**: Users can access any package manager command
1. **Consistency**: Unified interface for all package managers

### Challenges

1. **Command Discovery**: How to provide help/completion for dynamic commands
1. **Error Handling**: Distinguishing between "manager not found" vs "command failed"
1. **Security**: Preventing arbitrary command execution
1. **Configuration Integration**: How to integrate with unified config system

### Migration Path

1. Keep existing explicit commands for backward compatibility
1. Implement generic pattern alongside explicit commands
1. Eventually deprecate explicit commands in favor of generic pattern

### Current Decision

For now, we continue with explicit commands as they provide:

- Better documentation and help text
- Type safety and validation
- Clear command structure
- Better integration with configuration system

The generic pattern remains designed and ready for implementation when there's a clear need for supporting many more package managers dynamically.
