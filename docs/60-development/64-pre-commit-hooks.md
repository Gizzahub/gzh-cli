# Pre-commit Hooks

This project uses [pre-commit](https://pre-commit.com/) hooks to ensure code quality and consistency before commits are made.

## Installation

### Prerequisites

1. **Python and pip**: Pre-commit requires Python 3.6+
1. **Go tooling**: Various Go tools are used by the hooks

### Install Pre-commit

```bash
# Option 1: Using pip
pip install pre-commit

# Option 2: Using Homebrew (macOS)
brew install pre-commit

# Option 3: Using conda
conda install -c conda-forge pre-commit
```

### Install Go Tools

Install the formatter and pinned linter used by the local hooks:

```bash
make format-install-tools install-golangci-lint
```

### Install Hooks

Once pre-commit and the Go tools are installed, set up the hooks:

```bash
# Install pre-commit hooks for this repository
make pre-commit-install

# Or install the configured pre-commit stage manually:
pre-commit install --install-hooks
```

## Usage

### Automatic Execution

Once installed, the hooks will run automatically:

- **pre-commit**: Runs the configured file checks and fast Go checks before each commit

### Manual Execution

```bash
# Run all hooks on all files
make pre-commit

# Run specific hooks
pre-commit run fast-format-check --all-files
pre-commit run fast-lint-check --all-files

# Run hooks on specific files
pre-commit run --files cmd/bulk-clone/*.go
```

### Update Hooks

```bash
# Update hook versions
make pre-commit-update

# Or manually:
pre-commit autoupdate
```

## Configured Hooks

### File Validation

- **trailing-whitespace**: Removes trailing whitespace
- **end-of-file-fixer**: Ensures files end with a newline
- **check-yaml**: Validates YAML syntax
- **check-json**: Validates JSON syntax
- **check-toml**: Validates TOML syntax
- **check-xml**: Validates XML syntax
- **check-added-large-files**: Prevents large files (>500KB)
- **check-case-conflict**: Prevents case-sensitive filename conflicts
- **check-merge-conflict**: Detects merge conflict markers
- **check-shebang-scripts-are-executable**: Validates executable script metadata
- **check-executables-have-shebangs**: Requires executable text files to declare an interpreter
- **mixed-line-ending**: Normalizes line endings to LF
- **pretty-format-json**: Normalizes JSON formatting

### Go Code Quality

- **fast-format-check**: Runs `make fmt-diff` for changed Go files
- **fast-lint-check**: Runs `make lint-diff` for changed Go files
- **go-mod-tidy**: Ensures go.mod and go.sum are tidy

## Configuration

### Pre-commit Configuration

The configuration is in `.pre-commit-config.yaml`. Key settings:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
  - repo: local
    hooks:
      - id: fast-format-check
      - id: fast-lint-check
      - id: go-mod-tidy
```

### Emergency Bypass

```bash
# Bypass all hooks (use sparingly)
git commit --no-verify -m "emergency fix"

# Bypass specific hooks
SKIP=fast-lint-check git commit -m "fix: urgent patch"
```

### Permanent Exclusions

Edit `.pre-commit-config.yaml` to exclude files or disable hooks:

```yaml
- id: fast-lint-check
  exclude: ^(path/to/exclude|another/path)
```

## Troubleshooting

### Common Issues

1. **Hook fails with "command not found"**

   ```bash
   # Ensure Go tools are installed and in PATH
   go install mvdan.cc/gofumpt@latest
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

1. **Slow hook execution**

   ```bash
   # Skip the lint hook during development
   SKIP=fast-lint-check git commit -m "wip: development"
   ```

1. **Pre-commit not found**

   ```bash
   # Install pre-commit
   pip install pre-commit
   # Or
   brew install pre-commit
   ```

1. **Permission denied errors**

   ```bash
   # Ensure hooks are executable
   chmod +x .git/hooks/*
   ```

### Performance Tips

- Use `SKIP` environment variable for development commits
- Consider using `--files` flag for partial runs during development

## Best Practices

1. **Install hooks early**: Set up pre-commit hooks when starting development
1. **Regular updates**: Keep hooks updated with `make pre-commit-update`
1. **Gradual adoption**: Enable hooks incrementally for existing projects
1. **Team consistency**: Ensure all team members use the same hook configuration
1. **CI validation**: Use the same hooks in CI for consistency

## Integration with Development Workflow

```bash
# Daily development workflow
git add .
git commit -m "feat(module): add new feature"  # Hooks run automatically

# Before pushing, run the configured hooks across the repository
make pre-commit
git push

# Periodic maintenance
make pre-commit-update  # Update hook versions
make pre-commit         # Run all hooks manually
```

The pre-commit hooks are designed to catch issues early and maintain code quality standards across the entire development team.
