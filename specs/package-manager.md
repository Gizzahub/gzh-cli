<!-- 🚫 AI_MODIFY_PROHIBITED -->
<!-- This file should not be modified by AI agents -->

# Unified Package Manager Specification

## Overview

The unified package manager feature provides centralized management for multiple package managers through configuration files. It enables developers to maintain consistent development environments across machines by managing packages from various package managers including system-level (apt, brew, port), language-specific (pip, gem, npm), and version managers (asdf, rbenv, sdkman).

## Supported Package Managers

### System Package Managers
- **brew** (Homebrew) - macOS/Linux package manager
- **port** (MacPorts) - macOS package manager  
- **apt** (APT) - Debian/Ubuntu package manager
- **yum** (YUM) - RHEL/CentOS package manager
- **dnf** (DNF) - Fedora package manager
- **pacman** (Pacman) - Arch Linux package manager

### Version Managers
- **asdf** - Multi-language version manager
- **sdkman** - JVM ecosystem version manager
- **rbenv** - Ruby version manager
- **pyenv** - Python version manager
- **nvm** - Node.js version manager

### Language Package Managers
- **pip/uv** - Python package manager
- **gem** - Ruby package manager
- **npm/pnpm/yarn** - Node.js package managers
- **cargo** - Rust package manager
- **go** - Go modules
- **composer** - PHP package manager

## Commands

### Core Commands

- `gz pm` - Main package manager command
- `gz pm status` - Show status of all configured package managers
- `gz pm install` - Install packages from configuration files
- `gz pm update` - Update packages based on version strategy
- `gz pm sync` - Synchronize installed packages with configuration
- `gz pm export` - Export current installations to configuration files
- `gz pm validate` - Validate configuration files
- `gz pm clean` - Clean unused packages based on strategy
- `gz pm bootstrap` - Install and configure package managers
- `gz pm upgrade-managers` - Upgrade package managers themselves
- `gz pm sync-versions` - Synchronize version manager and package manager versions

### Package Manager Specific Commands

- `gz pm brew` - Manage Homebrew packages
- `gz pm asdf` - Manage asdf plugins and versions
- `gz pm pip` - Manage Python packages
- `gz pm npm` - Manage Node.js packages
- `gz pm [manager]` - Manage specific package manager

### Command Options

```bash
# Install packages from all configured managers
gz pm install

# Install from specific package manager
gz pm install --manager brew

# Install with specific strategy
gz pm install --strategy strict

# Export current installations
gz pm export --all
gz pm export --manager brew

# Update packages
gz pm update --all
gz pm update --manager asdf --strategy latest

# Sync with configuration (add missing, optionally remove extra)
gz pm sync --cleanup
gz pm sync --preserve-extra

# Clean unused packages
gz pm clean --dry-run
gz pm clean --force

# Bootstrap package managers
gz pm bootstrap --check
gz pm bootstrap --install brew,nvm,rbenv

# Upgrade package managers
gz pm upgrade-managers --all
gz pm upgrade-managers --manager brew

# Synchronize versions
gz pm sync-versions --check
gz pm sync-versions --fix
```

## Configuration

### Configuration File Locations

All package manager configurations are stored in:
```
~/.gzh/pm/
├── asdf.yml
├── brew.yml
├── sdkman.yml
├── port.yml
├── apt.yml
├── pip.yml
├── gem.yml
├── npm.yml
└── global.yml    # Global settings
```

### Global Configuration (`~/.gzh/pm/global.yml`)

```yaml
# Global Package Manager Configuration
version: "1.0.0"

# Default settings for all package managers
defaults:
  strategy: "preserve"          # preserve, strict, latest
  auto_update: false
  backup_before_changes: true
  parallel_operations: true
  max_workers: 4

# Package manager enablement
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

# Global version strategies
version_strategies:
  default: "stable"            # latest, stable, fixed
  development: "latest"
  production: "fixed"

# Cleanup policies
cleanup:
  remove_orphans: false        # Remove packages not in config
  remove_unused_deps: true     # Remove unused dependencies
  keep_cache: false           # Keep package caches
  dry_run_default: true       # Default to dry-run for safety

# Logging
logging:
  level: "info"
  file: "~/.gzh/logs/pm.log"
  format: "json"
```

### Homebrew Configuration (`~/.gzh/pm/brew.yml`)

```yaml
# Homebrew Package Configuration
version: "1.0.0"
generated_at: "2025-01-25T10:00:00Z"
platform: "darwin"  # darwin, linux

# Homebrew settings
settings:
  HOMEBREW_PREFIX: "/opt/homebrew"
  HOMEBREW_CASK_OPTS: "--appdir=/Applications"
  HOMEBREW_NO_ANALYTICS: "1"

# Version management strategy
strategy:
  default: "latest"        # latest, stable, fixed
  update_frequency: "weekly"
  auto_cleanup: true
  
# Cleanup policy
cleanup_policy: "preserve"  # preserve, remove, quarantine

# Taps (third-party repositories)
taps:
  - homebrew/core
  - homebrew/cask
  - homebrew/services
  - hashicorp/tap

# Formulae (command-line packages)
formulae:
  # Development tools
  - name: git
    version: "latest"
    options: []
    
  - name: vim
    version: "latest"
    options: ["--with-override-system-vi"]
    
  - name: tmux
    version: "stable"
    
  - name: docker
    version: "latest"
    start_service: true
    
  # Programming languages
  - name: go
    version: "1.21.5"      # Fixed version
    
  - name: rust
    version: "stable"
    
  - name: node
    version: "latest"
    link: true
    
  # Development tools with fixed versions
  - name: terraform
    tap: "hashicorp/tap"
    version: "1.6.0"
    
  - name: postgresql@15
    version: "15.5"
    start_service: true
    restart_service: "changed"

# Casks (GUI applications)
casks:
  # Browsers
  - name: google-chrome
    version: "latest"
    
  - name: firefox
    version: "latest"
    
  # Development tools
  - name: visual-studio-code
    version: "latest"
    
  - name: iterm2
    version: "stable"
    
  - name: docker
    version: "latest"
    
  # Productivity
  - name: slack
    version: "latest"
    quarantine: false
    
  - name: notion
    version: "latest"

# Services
services:
  - name: postgresql@15
    status: "started"      # started, stopped, restarted
    run_at: "login"       # login, boot
    
  - name: redis
    status: "started"
    run_at: "login"

# Post-install hooks
hooks:
  post_install:
    - "brew cleanup"
    - "brew doctor"
  post_update:
    - "brew cleanup --prune=7"
```

### ASDF Configuration (`~/.gzh/pm/asdf.yml`)

```yaml
# ASDF Version Manager Configuration
version: "1.0.0"
generated_at: "2025-01-25T10:00:00Z"

# ASDF settings
settings:
  legacy_version_file: true
  plugin_repository_last_check_duration: "7 days"
  
# Version management strategy
strategy:
  default: "stable"
  update_plugins: true
  
# Cleanup policy
cleanup_policy: "preserve"

# Plugins and versions
plugins:
  # Node.js
  nodejs:
    repository: "https://github.com/asdf-vm/asdf-nodejs.git"
    versions:
      - version: "20.11.0"
        global: true
      - version: "18.19.0"
        install: true
    post_install:
      - "npm install -g yarn pnpm"
      
  # Python
  python:
    repository: "https://github.com/asdf-vm/asdf-python.git"
    versions:
      - version: "3.12.1"
        global: true
      - version: "3.11.7"
        install: true
      - version: "3.10.13"
        install: true
    build_deps:
      darwin:
        - "openssl"
        - "readline"
        - "sqlite3"
        
  # Ruby
  ruby:
    repository: "https://github.com/asdf-vm/asdf-ruby.git"
    versions:
      - version: "3.3.0"
        global: true
      - version: "3.2.2"
        install: true
    environment:
      RUBY_CONFIGURE_OPTS: "--with-openssl-dir=/opt/homebrew/opt/openssl@3"
      
  # Go
  golang:
    repository: "https://github.com/asdf-vm/asdf-golang.git"
    versions:
      - version: "1.21.5"
        global: true
      - version: "1.20.12"
        install: false
        
  # Java
  java:
    repository: "https://github.com/halcyon/asdf-java.git"
    versions:
      - version: "openjdk-21"
        global: true
      - version: "openjdk-17.0.9"
        install: true
      - version: "openjdk-11.0.21"
        install: true

# Global tool versions (creates ~/.tool-versions)
global_versions:
  nodejs: "20.11.0"
  python: "3.12.1"
  ruby: "3.3.0"
  golang: "1.21.5"
  java: "openjdk-21"

# Hooks
hooks:
  post_plugin_add:
    - "asdf plugin update --all"
  post_install:
    - "asdf reshim"
```

### Python/pip Configuration (`~/.gzh/pm/pip.yml`)

```yaml
# Python Package Configuration
version: "1.0.0"
generated_at: "2025-01-25T10:00:00Z"

# Package manager settings
settings:
  package_manager: "pip"      # pip, uv
  use_venv: false            # Use global packages
  index_url: "https://pypi.org/simple"
  trusted_host: []
  
# Version strategy
strategy:
  default: "compatible"      # latest, compatible, fixed
  
# Cleanup policy
cleanup_policy: "preserve"

# Global packages
packages:
  # Development tools
  - name: poetry
    version: "latest"
    
  - name: pipenv
    version: "latest"
    
  - name: black
    version: "~=23.0"
    
  - name: flake8
    version: ">=6.0.0"
    
  - name: mypy
    version: "~=1.7.0"
    
  - name: pytest
    version: ">=7.4.0"
    extras: ["cov"]
    
  # Data science tools
  - name: jupyter
    version: "latest"
    
  - name: ipython
    version: ">=8.0.0"
    
  # CLI tools
  - name: httpie
    version: "latest"
    
  - name: awscli
    version: "~=2.15.0"
    
  # Fixed versions for compatibility
  - name: ansible
    version: "==8.7.0"
    
  - name: docker-compose
    version: "==2.23.0"

# Development dependencies (not installed globally)
dev_packages:
  - name: pre-commit
    version: ">=3.5.0"
    
  - name: tox
    version: ">=4.0.0"

# Requirements files to sync
requirements_files:
  - path: "~/.gzh/requirements/global.txt"
    strategy: "compatible"
  - path: "~/.gzh/requirements/tools.txt"
    strategy: "latest"
```

### SDKMAN Configuration (`~/.gzh/pm/sdkman.yml`)

```yaml
# SDKMAN JVM Tools Configuration
version: "1.0.0"
generated_at: "2025-01-25T10:00:00Z"

# SDKMAN settings
settings:
  auto_answer: true
  auto_selfupdate: true
  colour_enable: true
  
# Version strategy
strategy:
  default: "stable"
  
# Cleanup policy
cleanup_policy: "remove"    # Remove old versions

# Candidates (JVM tools)
candidates:
  # Java versions
  java:
    versions:
      - version: "21-tem"
        default: true
        install: true
      - version: "17.0.9-tem"
        install: true
      - version: "11.0.21-tem"
        install: true
        
  # Build tools
  gradle:
    versions:
      - version: "8.5"
        default: true
        install: true
      - version: "7.6.3"
        install: false
        
  maven:
    versions:
      - version: "3.9.6"
        default: true
        install: true
        
  # Kotlin
  kotlin:
    versions:
      - version: "1.9.22"
        default: true
        install: true
        
  # Scala
  scala:
    versions:
      - version: "3.3.1"
        default: true
        install: true
      - version: "2.13.12"
        install: true
        
  # Other JVM languages
  groovy:
    versions:
      - version: "4.0.18"
        default: true
        install: true
        
  # Build tools
  sbt:
    versions:
      - version: "1.9.8"
        default: true
        install: true
        
  # Application servers
  springboot:
    versions:
      - version: "3.2.1"
        default: true
        install: true

# Environment variables
environment:
  JAVA_HOME: "$SDKMAN_DIR/candidates/java/current"
  GRADLE_HOME: "$SDKMAN_DIR/candidates/gradle/current"
  MAVEN_HOME: "$SDKMAN_DIR/candidates/maven/current"
```

### NPM Configuration (`~/.gzh/pm/npm.yml`)

```yaml
# Node.js Package Configuration
version: "1.0.0"
generated_at: "2025-01-25T10:00:00Z"

# Package manager settings
settings:
  package_manager: "npm"     # npm, pnpm, yarn
  registry: "https://registry.npmjs.org/"
  prefix: "~/.npm-global"
  
# Version strategy
strategy:
  default: "latest"
  
# Cleanup policy
cleanup_policy: "preserve"

# Global packages
packages:
  # Package managers
  - name: pnpm
    version: "latest"
    
  - name: yarn
    version: "latest"
    
  # Build tools
  - name: typescript
    version: "^5.3.0"
    
  - name: ts-node
    version: "latest"
    
  - name: nodemon
    version: "latest"
    
  # Linting and formatting
  - name: eslint
    version: "^8.56.0"
    
  - name: prettier
    version: "^3.1.0"
    
  # Testing
  - name: jest
    version: "^29.7.0"
    
  # CLI tools
  - name: "@angular/cli"
    version: "^17.0.0"
    
  - name: "@vue/cli"
    version: "^5.0.8"
    
  - name: create-react-app
    version: "latest"
    
  - name: "@nestjs/cli"
    version: "^10.2.0"
    
  - name: "vercel"
    version: "latest"
    
  - name: "netlify-cli"
    version: "latest"
    
  # Development tools
  - name: "serve"
    version: "latest"
    
  - name: "concurrently"
    version: "latest"
    
  - name: "npm-check-updates"
    version: "latest"

# NPM configuration
npm_config:
  init-author-name: "Your Name"
  init-author-email: "your.email@example.com"
  init-license: "MIT"
  save-exact: false
```

## Package Manager Bootstrap

### Overview

The bootstrap feature manages the installation and upgrade of package managers themselves. This ensures that all required package managers are available and up-to-date before managing packages.

### Bootstrap Commands

#### `gz pm bootstrap`

Installs missing package managers based on platform and configuration.

```bash
# Check which package managers need installation
gz pm bootstrap --check

# Install all missing package managers
gz pm bootstrap --install

# Install specific package managers
gz pm bootstrap --install brew,nvm,rbenv

# Force reinstall
gz pm bootstrap --force --install brew
```

### Bootstrap Configuration

Add to `~/.gzh/pm/global.yml`:

```yaml
# Package manager bootstrap configuration
bootstrap:
  auto_install: true              # Automatically install missing managers
  check_on_startup: true          # Check manager availability on each run
  
  # Package manager installation sources
  managers:
    # Homebrew
    brew:
      darwin:
        install_script: "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh"
        check_command: "brew --version"
        min_version: "4.0.0"
      linux:
        install_script: "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh"
        check_command: "brew --version"
        
    # NVM (Node Version Manager)
    nvm:
      install_script: "https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh"
      check_command: "nvm --version"
      min_version: "0.39.0"
      post_install:
        - "source ~/.nvm/nvm.sh"
        
    # rbenv
    rbenv:
      darwin:
        install_method: "brew"
        brew_formula: "rbenv ruby-build"
      linux:
        install_script: "https://github.com/rbenv/rbenv-installer/raw/HEAD/bin/rbenv-installer"
      check_command: "rbenv --version"
      
    # pyenv
    pyenv:
      darwin:
        install_method: "brew"
        brew_formula: "pyenv"
      linux:
        install_script: "https://github.com/pyenv/pyenv-installer/raw/master/bin/pyenv-installer"
      check_command: "pyenv --version"
      
    # SDKMAN
    sdkman:
      install_script: "https://get.sdkman.io"
      check_command: "sdk version"
      environment:
        SDKMAN_DIR: "$HOME/.sdkman"
        
    # pip
    pip:
      check_command: "pip --version"
      install_with: "python"
      upgrade_command: "python -m pip install --upgrade pip"
      
    # npm (usually comes with Node.js)
    npm:
      check_command: "npm --version"
      install_with: "node"
      upgrade_command: "npm install -g npm@latest"
```

### Package Manager Upgrade

#### `gz pm upgrade-managers`

Upgrades package managers to their latest versions.

```bash
# Upgrade all package managers
gz pm upgrade-managers --all

# Upgrade specific manager
gz pm upgrade-managers --manager brew
gz pm upgrade-managers --manager pip

# Check available upgrades
gz pm upgrade-managers --check
```

## Version Manager Coordination

### Overview

Handles the complexity of version managers and their associated package managers, ensuring compatibility and smooth transitions between versions.

### Version Synchronization Issues

#### Node.js / npm Coordination

When using nvm to manage Node.js versions, npm versions can become mismatched:
- Each Node.js version comes with a bundled npm version
- Global npm packages are tied to specific Node.js versions
- Switching Node.js versions can lead to missing packages or version conflicts

#### Ruby / gem Coordination

When using rbenv to manage Ruby versions:
- Each Ruby version has its own gem installation directory
- Gems installed for one Ruby version are not available in another
- Switching Ruby versions requires reinstalling all gems

### Configuration for Version Coordination

Add to `~/.gzh/pm/global.yml`:

```yaml
# Version manager coordination
version_coordination:
  # Node.js and npm coordination
  node_npm_sync:
    enabled: true
    strategy: "bundled"         # bundled, latest, compatible, fixed
    npm_version_matrix:
      "20.x": "10.x"           # Node 20 uses npm 10
      "18.x": "9.x"            # Node 18 uses npm 9
      "16.x": "8.x"            # Node 16 uses npm 8
    global_packages_backup: true
    
  # Ruby and gem coordination  
  ruby_gem_migration:
    enabled: true
    strategy: "copy"           # copy, reinstall, selective, manual
    backup_before_switch: true
    gem_sets:                  # Define gem sets
      - name: "development"
        gems: ["bundler", "rails", "rspec", "rubocop"]
      - name: "deployment"
        gems: ["capistrano", "mina"]
        
  # Python and pip coordination
  python_pip_migration:
    enabled: true
    strategy: "venv"           # venv, reinstall, requirements
    auto_create_venv: true
    requirements_file: "~/.gzh/requirements/python-global.txt"
    
  # Java coordination (multiple JDKs)
  java_coordination:
    enabled: true
    default_jdk: "21"
    project_jdk_detection: true  # Detect from .java-version or pom.xml
```

### Version Sync Commands

#### `gz pm sync-versions`

Ensures version managers and package managers are in sync.

```bash
# Check for version mismatches
gz pm sync-versions --check

# Fix version mismatches
gz pm sync-versions --fix

# Sync specific language
gz pm sync-versions --language node
gz pm sync-versions --language ruby --strategy reinstall
```

## Language-Specific Package Migration

### Overview

Provides intelligent migration of packages when switching between language versions, minimizing the need to manually reinstall packages.

### Migration Strategies

#### Ruby Migration

When upgrading Ruby versions via rbenv:

```yaml
# Ruby-specific migration configuration
ruby_migration:
  default_strategy: "intelligent"    # intelligent, copy-all, reinstall, manual
  
  strategies:
    intelligent:
      # Analyze Gemfile.lock for version compatibility
      check_compatibility: true
      # Skip gems with native extensions for different Ruby versions
      skip_incompatible: true
      # Reinstall gems with native extensions
      rebuild_native: true
      
    copy-all:
      # Copy all gems from source to target Ruby version
      include_bundled: false
      verify_after_copy: true
      
    reinstall:
      # Use Gemfile or gem list to reinstall
      source: "gemfile"             # gemfile, gem-list, backup
      parallel_install: true
      
  # Gem backup configuration
  backup:
    enabled: true
    location: "~/.gzh/backups/ruby-gems"
    format: "tarball"              # tarball, directory
    keep_last: 3
```

#### Node.js Migration

When upgrading Node.js versions via nvm:

```yaml
# Node.js-specific migration configuration
node_migration:
  default_strategy: "intelligent"
  
  strategies:
    intelligent:
      # Check package.json for version constraints
      respect_constraints: true
      # Update packages to compatible versions
      auto_update: true
      # Handle peer dependencies
      resolve_peers: true
      
    preserve:
      # Keep exact versions where possible
      use_lockfile: true
      # Fall back to compatible versions
      fallback_strategy: "compatible"
      
  # Global packages to always install
  always_install:
    - "npm"
    - "yarn"
    - "pnpm"
    - "typescript"
    - "ts-node"
    
  # npm configuration preservation
  preserve_npm_config: true
  npm_config_items:
    - "registry"
    - "prefix"
    - "init-*"
```

#### Python Migration

When upgrading Python versions via pyenv:

```yaml
# Python-specific migration configuration  
python_migration:
  default_strategy: "venv-based"
  
  strategies:
    venv-based:
      # Create new venv for each Python version
      auto_create: true
      # Copy requirements from old venv
      copy_requirements: true
      # Update packages to compatible versions
      update_on_create: true
      
    global-packages:
      # Maintain a global package list
      track_global: true
      requirements_file: "~/.gzh/requirements/python-global.txt"
      
  # Always install these packages globally
  essential_packages:
    - "pip"
    - "setuptools"
    - "wheel"
    - "virtualenv"
```


## Version Management Strategies

### Version Specification

- **latest**: Always update to the latest version
- **stable**: Update to the latest stable version (exclude pre-release)
- **fixed**: Keep the exact specified version
- **compatible**: Update within compatible version range (e.g., ~1.2.3, ^2.0.0)

### Examples

```yaml
# Latest version
- name: docker
  version: "latest"

# Stable version
- name: postgresql
  version: "stable"

# Fixed version
- name: terraform
  version: "1.6.0"

# Compatible version (pip/npm style)
- name: black
  version: "~=23.0"    # >=23.0, <24.0

- name: typescript
  version: "^5.3.0"    # >=5.3.0, <6.0.0
```

## Cleanup Strategies

### Strategy Options

1. **preserve** (default): Keep all installed packages, only add missing ones
2. **remove**: Remove packages not in configuration
3. **quarantine**: Move unmanaged packages to quarantine list
4. **strict**: Remove anything not explicitly defined

### Per-Package Manager Behavior

```yaml
# In global.yml or per-manager config
cleanup:
  strategy: "preserve"
  exceptions:
    - "build-essential"    # Never remove
    - "git"               # Never remove
  quarantine_path: "~/.gzh/pm/quarantine/"
```

## Synchronization Process

### Install Process
1. Read configuration files
2. Detect installed package managers
3. Compare configured vs installed packages
4. Install missing packages respecting version strategy
5. Optionally cleanup based on cleanup strategy

### Export Process
1. Detect all package managers
2. List installed packages with versions
3. Generate/update configuration files
4. Organize by package manager
5. Add metadata (timestamp, platform)

### Update Process
1. Read configuration files
2. Check for updates based on version strategy
3. Update packages in dependency order
4. Run post-update hooks
5. Log all changes

## Multi-Manager Coordination

### Dependency Resolution
- Detect cross-manager dependencies
- Install in correct order (system → version managers → language packages)
- Handle conflicts between managers

### Example Workflow

```bash
# First time setup: export current state
gz pm export --all

# Install on new machine
gz pm install --all

# Regular updates
gz pm update --strategy stable

# Sync and cleanup
gz pm sync --cleanup --dry-run
gz pm sync --cleanup

# Update specific manager
gz pm update --manager brew
```

## Platform Support

### Operating System Detection
- macOS: brew, port, asdf
- Linux: apt/yum/dnf/pacman, brew, asdf
- Windows: chocolatey, scoop (future)

### Architecture Support
- Intel (x86_64)
- Apple Silicon (arm64)
- Linux ARM

## Error Handling

### Common Scenarios
- Package manager not installed
- Version conflicts
- Missing dependencies
- Network failures
- Permission issues

### Recovery Options
- Rollback to previous state
- Skip failed packages
- Retry with different strategy
- Generate error report

## Integration Points

- **Development Environment**: Coordinates with dev-env for tool installation
- **Network Management**: Respects proxy settings from net-env
- **Version Control**: Can track configuration files in git
- **CI/CD**: Export configurations for reproducible builds

## Security Considerations

- Verify package signatures when possible
- Use official package repositories
- Audit dependencies for vulnerabilities
- Store sensitive data in environment variables
- Regular security updates based on advisories

## Performance Optimization

- Parallel package installation where supported
- Cache package downloads
- Incremental updates
- Minimal dependency resolution
- Background update checks

## Future Enhancements

- Windows package manager support (chocolatey, scoop)
- Container-based isolation
- Package vulnerability scanning
- Automated dependency updates with PR creation
- Cloud sync for configurations
- Package usage analytics