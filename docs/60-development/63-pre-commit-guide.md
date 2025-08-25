# Pre-commit Hook Guide

This guide explains how to use and configure pre-commit hooks in the gzh-cli project.

## Overview

Pre-commit hooks help maintain code quality by automatically checking and fixing issues before commits. Our configuration includes:

- Code formatting (gofumpt, gci)
- Linting (golangci-lint)
- Security scanning (gosec, detect-secrets)
- Test execution
- Commit message validation
- File size checks
- TODO format validation

## Installation

### Prerequisites

1. Install pre-commit:

   ```bash
   # Using pip
   pip install pre-commit

   # Using Homebrew (macOS)
   brew install pre-commit

   # Using apt (Ubuntu/Debian)
   apt install pre-commit
   ```

1. Install Go tools:

   ```bash
   make bootstrap
   ```

### Setup

Run the setup script to install all hooks:

```bash
make pre-commit-install
# or
./scripts/setup-git-hooks.sh
```

This will:

- Install pre-commit hooks
- Install commit-msg hooks for conventional commits
- Install pre-push hooks for additional checks
- Set up a prepare-commit-msg hook for branch-based issue prefixes

## Hook Types

### Pre-commit Hooks

Run before each commit:

1. **File Checks**

   - Trailing whitespace removal
   - End-of-file fixing
   - YAML/JSON/TOML/XML validation
   - Large file detection (>500KB)
   - Case conflict detection
   - Merge conflict markers
   - Private key detection

1. **Go Formatting**

   - `gofumpt`: Stricter gofmt
   - `gci`: Import grouping and sorting

1. **Go Quality**

   - `go vet`: Static analysis
   - `golangci-lint`: Comprehensive linting
   - `go build`: Compilation check
   - `go test -short`: Quick tests

1. **Security**

   - `gosec`: Security vulnerability scanning
   - `detect-secrets`: Secret detection

1. **Custom Checks**

   - TODO format validation
   - File size warnings
   - Package documentation checks

### Commit-msg Hooks

Validates commit messages follow conventional format:

```
type(scope): description

[optional body]

[optional footer]
```

Examples:

- `feat(synclone): add retry logic for failed clones`
- `fix(config): resolve env var precedence issue`
- `docs(readme): update installation instructions`
- `refactor(internal): move packages from pkg to internal`

### Pre-push Hooks

Run before pushing to remote:

- Full test suite (`go test`)
- Coverage checks
- Comprehensive linting (`make lint-all`)

### Prepare-commit-msg Hook

Automatically adds issue numbers from branch names:

- Branch: `feature/ISSUE-123-add-retry-logic`
- Commit message: `[ISSUE-123] <your message>`

## Usage

### Running Hooks Manually

```bash
# Run all hooks on all files
pre-commit run --all-files

# Run specific hook
pre-commit run golangci-lint-mod

# Run hooks for specific files
pre-commit run --files cmd/synclone/*.go

# Run push hooks
pre-commit run --hook-stage pre-push
```

### Bypassing Hooks

Use sparingly when necessary:

```bash
# Skip pre-commit hooks
git commit --no-verify

# Skip pre-push hooks
git push --no-verify
```

### Updating Hooks

```bash
# Update hook versions
pre-commit autoupdate

# Update baseline for secrets
detect-secrets scan --baseline .secrets.baseline
```

## Configuration

### Adding New Hooks

Edit `.pre-commit-config.yaml`:

```yaml
- repo: local
  hooks:
    - id: my-custom-check
      name: My Custom Check
      entry: ./scripts/my-check.sh
      language: script
      types: [go]
```

### Excluding Files

Add patterns to exclude files:

```yaml
- id: prettier
  exclude: |
    (?x)(
      ^test/.*\.yaml$|
      ^vendor/.*
    )
```

### Hook-specific Configuration

Most hooks respect tool-specific config files:

- `golangci-lint`: `.golangci.yml`
- `prettier`: `.prettierrc`
- `hadolint`: `.hadolint.yaml`

## Troubleshooting

### Common Issues

1. **Hook Installation Fails**

   ```bash
   # Ensure git is initialized
   git init

   # Check pre-commit version
   pre-commit --version
   ```

1. **Go Tools Not Found**

   ```bash
   # Install required tools
   make bootstrap

   # Ensure GOPATH/bin is in PATH
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

1. **Secrets Detected**

   ```bash
   # Update baseline if false positive
   detect-secrets scan --baseline .secrets.baseline

   # Or add inline comment
   secret = "not-a-real-secret"  # pragma: allowlist secret
   ```

1. **Large File Blocked**

   ```bash
   # Add to .gitattributes if needed
   *.bin filter=lfs diff=lfs merge=lfs -text

   # Or increase limit in .pre-commit-config.yaml
   args: [--maxkb=1000]
   ```

### Performance Tips

1. **Stage Specific Files**

   ```bash
   # Only check staged files
   git add specific-file.go
   git commit
   ```

1. **Skip Expensive Hooks**

   ```bash
   # Skip specific hooks
   SKIP=go-test-mod,gosec git commit
   ```

1. **Use Quick Mode**

   ```bash
   # For development iterations
   make dev-fast  # Format and quick tests only
   ```

## CI Integration

Pre-commit hooks are also run in CI:

- GitHub Actions runs the same checks
- Ensures consistency between local and CI environments
- Auto-fixes are applied in PRs when possible

## Best Practices

1. **Fix Issues Immediately**

   - Don't accumulate linting debt
   - Address issues before committing

1. **Keep Hooks Fast**

   - Use `-short` flag for tests in pre-commit
   - Move expensive checks to pre-push

1. **Document Suppressions**

   - Always explain why when using `--no-verify`
   - Add comments when disabling linters

1. **Regular Updates**

   - Run `pre-commit autoupdate` monthly
   - Update tool versions in sync

1. **Team Alignment**

   - Ensure all team members use hooks
   - Document project-specific conventions
   - Share hook configuration updates

## Resources

- [Pre-commit documentation](https://pre-commit.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Project Code Quality Standards](./code-quality.md)
- [Testing Strategy](./testing-strategy.md)
