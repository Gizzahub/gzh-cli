# Scripts Directory

This directory contains utility scripts for the gzh-manager-go project. These scripts automate various development, build, and maintenance tasks.

## 📋 Available Scripts

### 🔧 Development & Build Scripts

#### `completions.sh`
**Purpose**: Generate shell completion files for the gzh-manager CLI

**Usage**:
```bash
./scripts/completions.sh
```

**Output**: Creates completion files in the `completions/` directory:
- `completions/gzh-manager.bash` - Bash completion
- `completions/gzh-manager.zsh` - Zsh completion
- `completions/gzh-manager.fish` - Fish completion

**Installation**:
```bash
# Bash
sudo cp completions/gzh-manager.bash /etc/bash_completion.d/

# Zsh (add to ~/.zshrc)
source /path/to/completions/gzh-manager.zsh

# Fish
cp completions/gzh-manager.fish ~/.config/fish/completions/
```

---

#### `manpages.sh`
**Purpose**: Generate manual pages for the gzh-manager CLI

**Usage**:
```bash
./scripts/manpages.sh
```

**Output**: Creates `manpages/gzh-manager.1.gz` - compressed man page

**Installation**:
```bash
sudo cp manpages/gzh-manager.1.gz /usr/share/man/man1/
man gzh-manager  # View the manual
```

---

### 🏗️ Build & Quality Scripts

#### `setup-git-hooks.sh`
**Purpose**: Set up Git hooks for pre-commit, commit-msg, and pre-push checks

**Usage**:
```bash
./scripts/setup-git-hooks.sh
```

**Features**:
- Installs pre-commit hooks for formatting and linting
- Sets up commit message validation (conventional commits)
- Configures pre-push hooks for tests and coverage
- Creates custom prepare-commit-msg hook for branch-based commits

**Requirements**:
- `pre-commit` must be installed (`pip install pre-commit`)

---

#### `check-coverage.sh`
**Purpose**: Check test coverage and enforce thresholds

**Usage**:
```bash
# Basic coverage check
./scripts/check-coverage.sh

# Generate detailed report
./scripts/check-coverage.sh --report
```

**Environment Variables**:
```bash
export COVERAGE_THRESHOLD=70           # Total coverage threshold (default: 70%)
export PACKAGE_COVERAGE_THRESHOLD=60  # Package coverage threshold (default: 60%)
```

**Exit Codes**:
- `0`: Coverage meets thresholds
- `1`: Coverage below thresholds

---

#### `add-build-tags.sh`
**Purpose**: Add build tags to integration and e2e test files

**Usage**:
```bash
./scripts/add-build-tags.sh
```

**What it does**:
- Adds `//go:build integration` tags to `test/integration/*_test.go`
- Adds `//go:build e2e` tags to `test/e2e/*_test.go`
- Skips files that already have build tags

**Running tagged tests**:
```bash
go test -tags=integration ./test/integration/...
go test -tags=e2e ./test/e2e/...
```

---

### 📦 Migration & Configuration Scripts

#### `migrate-config.sh`
**Purpose**: Migrate from bulk-clone.yaml to gzh.yaml configuration format

**Usage**:
```bash
# Basic migration
./scripts/migrate-config.sh

# Custom input/output files
./scripts/migrate-config.sh -i old-config.yaml -o new-config.yaml

# Preview migration without creating files
./scripts/migrate-config.sh --dry-run

# Create backups before migration
./scripts/migrate-config.sh --backup
```

**Options**:
- `-i, --input FILE`: Input bulk-clone.yaml file (default: bulk-clone.yaml)
- `-o, --output FILE`: Output gzh.yaml file (default: gzh.yaml)
- `--dry-run`: Preview migration without creating files
- `--backup`: Create backup of existing files
- `-h, --help`: Show help message

**Migration Process**:
1. Analyzes existing bulk-clone.yaml configuration
2. Extracts organizations and ignore patterns
3. Generates gzh.yaml template with migration notes
4. Provides next steps for manual configuration

**Requirements**:
- `yq` (optional, for better migration support)

---

#### `migration/` Directory
Contains legacy command migration and backward compatibility scripts. See [`migration/README.md`](migration/README.md) for detailed documentation.

**Available migration scripts**:
- `migrate-gz.sh` - Command structure migration helper
- `deprecated-aliases.sh` / `deprecated-aliases.fish` - Backward compatibility aliases
- `install-aliases.sh` / `uninstall-aliases.sh` - Alias installation/removal

**Quick usage**:
```bash
# Migrate old commands
./scripts/migration/migrate-gz.sh

# Install backward compatibility
./scripts/migration/install-aliases.sh

# Remove compatibility aliases
./scripts/migration/uninstall-aliases.sh
```

---

### 🐛 Debug Scripts

#### `debug/` Directory
Contains specialized debugging scripts for development. See [`debug/README.md`](debug/README.md) for detailed documentation.

**Available debug scripts**:
- `debug-cli.sh` - Debug CLI application with any command
- `debug-test.sh` - Debug Go tests in specific packages
- `debug-attach.sh` - Attach debugger to running processes

**Quick usage**:
```bash
# Debug CLI commands
./scripts/debug/debug-cli.sh bulk-clone --config examples/gzh-simple.yaml --dry-run

# Debug specific tests
./scripts/debug/debug-test.sh ./cmd/bulk-clone TestBulkClone

# Attach to running process
./scripts/debug/debug-attach.sh
```

---

## 🚀 Common Workflows

### Initial Project Setup
```bash
# Set up development environment
./scripts/setup-git-hooks.sh
./scripts/completions.sh
./scripts/manpages.sh
```

### Before Committing
```bash
# Check coverage
./scripts/check-coverage.sh

# Git hooks will automatically run:
# - Code formatting (gofumpt, gci)
# - Linting (golangci-lint)
# - Tests and coverage checks
```

### Migrating from Old Configuration
```bash
# Preview migration
./scripts/migrate-config.sh --dry-run

# Perform migration with backup
./scripts/migrate-config.sh --backup
```

### Debugging Issues
```bash
# Debug specific command
./scripts/debug/debug-cli.sh your-command --your-flags

# Debug failing tests
./scripts/debug/debug-test.sh ./path/to/package TestFunctionName
```

---

## 📁 Script Categories

### ✅ Essential Scripts (Always Keep)
- `setup-git-hooks.sh` - Development workflow automation
- `completions.sh` - User experience improvement
- `migrate-config.sh` - Configuration migration tool
- `manpages.sh` - Documentation generation
- `pre-commit-lint.sh` - Pre-commit linting
- `test-git-repo-e2e.sh` - Git repository E2E testing

### ⚠️ Development Tools (Keep for Development)
- `debug/` - Debugging utilities (3 files)
- `install-git-extensions.sh` / `uninstall-git-extensions.sh` - Git extension management

### 🔄 Migration Tools (Temporary)
- `migration/` - Legacy command migration and backward compatibility (6 files)
  - Will be removed after v3.0.0 transition (estimated: 2025-01-01)

### ✅ Cleaned Up (Removed)
- ~~`aliases-unified.sh`~~ - Broken wrapper script
- ~~`e2e-test.sh`~~ - Obsolete test for old `git-synclone` binary

---

## 🔧 Script Maintenance

### Adding New Scripts
1. Create the script with executable permissions:
   ```bash
   chmod +x scripts/new-script.sh
   ```

2. Add documentation to this README

3. Include usage examples and error handling

4. Test the script in different environments

### Best Practices
- Use `set -e` for error handling
- Provide clear usage messages
- Include help options (`-h, --help`)
- Use meaningful exit codes
- Add logging/output for user feedback

---

## 🐛 Troubleshooting

### Common Issues

#### Permission Denied
```bash
chmod +x scripts/script-name.sh
```

#### Missing Dependencies
Most scripts will check for required tools and provide installation instructions.

#### Pre-commit Issues
```bash
# Reinstall hooks
./scripts/setup-git-hooks.sh

# Run manually
pre-commit run --all-files
```

#### Coverage Script Fails
```bash
# Check if bc is installed (required for threshold comparison)
which bc || sudo apt-get install bc  # Ubuntu/Debian
```

---

## 📚 Additional Resources

- [Pre-commit Documentation](https://pre-commit.com/)
- [Delve Debugger Guide](debug/README.md)
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [gzh-manager Configuration Guide](../docs/)

---

**Note**: These scripts are designed specifically for the gzh-manager-go project. Modify paths and configurations as needed for your environment.
