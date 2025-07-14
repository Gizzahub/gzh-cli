# Pre-commit Hooks

This project uses [pre-commit](https://pre-commit.com/) hooks to ensure code quality and consistency before commits are made.

## Installation

### Prerequisites

1. **Python and pip**: Pre-commit requires Python 3.6+
2. **Go tooling**: Various Go tools are used by the hooks

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

The following Go tools are required for the hooks:

```bash
# Install required Go tools
go install mvdan.cc/gofumpt@latest
go install github.com/daixiang0/gci@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Install Hooks

Once pre-commit and the Go tools are installed, set up the hooks:

```bash
# Install pre-commit hooks for this repository
make pre-commit-install

# Or manually:
pre-commit install --install-hooks
pre-commit install --hook-type commit-msg
pre-commit install --hook-type pre-push
```

## Usage

### Automatic Execution

Once installed, the hooks will run automatically:

- **pre-commit**: Runs before each commit
- **commit-msg**: Validates commit message format
- **pre-push**: Runs before pushing to remote

### Manual Execution

```bash
# Run all hooks on all files
make pre-commit

# Run specific hooks
pre-commit run --hook-stage manual gosec
pre-commit run --hook-stage manual go-vuln-check

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
- **check-symlinks**: Validates symlinks
- **detect-private-key**: Detects private keys
- **mixed-line-ending**: Normalizes line endings to LF

### Go Code Quality
- **go-build-mod**: Verifies code compiles
- **go-test-mod**: Runs unit tests with race detection
- **go-vet-mod**: Runs `go vet` static analysis
- **go-staticcheck-mod**: Runs staticcheck linter
- **go-fumpt**: Formats Go code (stricter than gofmt)
- **golangci-lint-mod**: Runs comprehensive linting
- **gci**: Formats Go imports
- **gosec**: Security vulnerability scanning
- **go-mod-tidy**: Ensures go.mod and go.sum are tidy
- **go-vuln-check**: Checks for known vulnerabilities

### Additional Linting
- **prettier**: Formats YAML, JSON, and Markdown files
- **hadolint-docker**: Lints Dockerfiles
- **shellcheck**: Lints shell scripts
- **conventional-pre-commit**: Enforces conventional commit messages
- **detect-secrets**: Detects hardcoded secrets

## Configuration

### Pre-commit Configuration

The configuration is in `.pre-commit-config.yaml`. Key settings:

```yaml
default_install_hook_types: [pre-commit, commit-msg, pre-push]
default_stages: [pre-commit]

ci:
  skip: [go-test-mod, go-vuln-check, gosec]  # Skip slow hooks in CI
```

### Secrets Detection

The `.secrets.baseline` file contains the baseline for secrets detection. To update:

```bash
detect-secrets scan --baseline .secrets.baseline
```

### Exclusions

Some hooks exclude certain paths:

- **prettier**: Excludes test and sample configuration files
- **detect-secrets**: Excludes test files and documentation

## Bypassing Hooks

### Emergency Bypass

```bash
# Bypass all hooks (use sparingly)
git commit --no-verify -m "emergency fix"

# Bypass specific hooks
SKIP=gosec,go-vuln-check git commit -m "fix: urgent patch"
```

### Permanent Exclusions

Edit `.pre-commit-config.yaml` to exclude files or disable hooks:

```yaml
- id: prettier
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

2. **Slow hook execution**
   ```bash
   # Skip slow hooks during development
   SKIP=go-test-mod,gosec git commit -m "wip: development"
   ```

3. **Pre-commit not found**
   ```bash
   # Install pre-commit
   pip install pre-commit
   # Or
   brew install pre-commit
   ```

4. **Permission denied errors**
   ```bash
   # Ensure hooks are executable
   chmod +x .git/hooks/*
   ```

### Performance Tips

- Use `SKIP` environment variable for development commits
- Run `pre-commit run --hook-stage manual` for expensive checks
- Consider using `--files` flag for partial runs during development

### CI Integration

The configuration includes CI-specific settings for pre-commit.ci:

- Automatically creates PRs for hook updates
- Skips resource-intensive hooks in CI environment
- Provides auto-fix capabilities for formatting issues

## Best Practices

1. **Install hooks early**: Set up pre-commit hooks when starting development
2. **Regular updates**: Keep hooks updated with `make pre-commit-update`
3. **Gradual adoption**: Enable hooks incrementally for existing projects
4. **Team consistency**: Ensure all team members use the same hook configuration
5. **CI validation**: Use the same hooks in CI for consistency

## Integration with Development Workflow

```bash
# Daily development workflow
git add .
git commit -m "feat(module): add new feature"  # Hooks run automatically

# Before pushing
git push  # pre-push hooks run

# Periodic maintenance
make pre-commit-update  # Update hook versions
make pre-commit         # Run all hooks manually
```

The pre-commit hooks are designed to catch issues early and maintain code quality standards across the entire development team.