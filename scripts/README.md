# Scripts Directory

This directory contains utility scripts for the gzh-cli project. These scripts automate various development, build, and maintenance tasks.

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

______________________________________________________________________

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

______________________________________________________________________

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

______________________________________________________________________

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

______________________________________________________________________

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
make cover

# Git hooks will automatically run:
# - Code formatting (gofumpt, gci)
# - Linting (golangci-lint)
# - Tests and coverage checks
```

### Debugging Issues

```bash
# Debug specific command
./scripts/debug/debug-cli.sh your-command --your-flags

# Debug failing tests
./scripts/debug/debug-test.sh ./path/to/package TestFunctionName
```

______________________________________________________________________

## 📁 Script Categories

### ✅ Essential Scripts (Always Keep)

- `setup-git-hooks.sh` - Development workflow automation
- `completions.sh` - User experience improvement
- `manpages.sh` - Documentation generation

### ⚠️ Development Tools (Keep for Development)

- `debug/` - Debugging utilities (3 files)

### ✅ Cleaned Up (Removed)

- ~~`aliases-unified.sh`~~ - Broken wrapper script
- ~~`e2e-test.sh`~~ - Obsolete test for old `git-synclone` binary
- ~~`migration/`~~ - Legacy command migration and backward compatibility scripts (6 files)
- ~~`install-git-extensions.sh` / `uninstall-git-extensions.sh`~~ - Git extension management
- ~~`cmd/git-synclone/`~~ - Legacy git-synclone binary (replaced by gz)
- ~~`homebrew/gzh-git-extensions.rb`~~ - Homebrew formula for git-synclone
- ~~`debian/`~~ - Debian packaging for git-synclone
- ~~`install.sh`~~ - Online installer for git-synclone
- ~~`migrate-config.sh`~~ - One-time bulk-clone.yaml to gzh.yaml migration (`bulk-clone` command removed)
- ~~`verify-cleanup.sh`~~ - One-time check that `internal/legacy` and `internal/api` were gone
- ~~`simple-benchmark.sh`~~ - Startup/size check, duplicate of gzh-cli-template's wired `benchmark.sh`
- ~~`benchmark-performance.sh`~~ - Baseline benchmarking, superseded by `make bench` / `make bench-compare`
- ~~`pre-commit-lint.sh`~~ - Pre-commit lint hook that no installer ever wired
- ~~`test-git-repo-e2e.sh`~~ - One-off manual E2E fixture run
- ~~`testing/synclone/setup-test-repos/` / `cleanup-test-repos/`~~ - Manual synclone test fixture commands
- ~~`check-coverage.sh`~~ - Coverage threshold check, superseded by `make cover` / `make cover-report`
- ~~`add-build-tags.sh`~~ - One-time insertion of integration/e2e build tags

______________________________________________________________________

## 🔧 Script Maintenance

### Adding New Scripts

1. Create the script with executable permissions:

   ```bash
   chmod +x scripts/new-script.sh
   ```

1. Add documentation to this README

1. Include usage examples and error handling

1. Test the script in different environments

### Best Practices

- Use `set -e` for error handling
- Provide clear usage messages
- Include help options (`-h, --help`)
- Use meaningful exit codes
- Add logging/output for user feedback

______________________________________________________________________

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

______________________________________________________________________

## 📚 Additional Resources

- [Pre-commit Documentation](https://pre-commit.com/)
- [Delve Debugger Guide](debug/README.md)
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [gzh-manager Configuration Guide](../docs/)

______________________________________________________________________

**Note**: These scripts are designed specifically for the gzh-cli project. Modify paths and configurations as needed for your environment.
