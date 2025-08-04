# quality Command Reference

Multi-language code quality management with integrated formatters, linters, and analysis tools.

## Synopsis

```bash
gz quality <action> [flags]
gz quality <action> --config <config-file>
```

## Description

The `quality` command provides comprehensive code quality management across multiple programming languages, including formatting, linting, and analysis capabilities.

## Supported Languages

- **Go** - gofumpt, gci, golangci-lint, staticcheck
- **Python** - black, isort, ruff, flake8, pylint, mypy
- **JavaScript/TypeScript** - prettier, eslint, tsc
- **Rust** - rustfmt, clippy
- **Java** - google-java-format, checkstyle, spotbugs, pmd
- **C/C++** - clang-format, clang-tidy, cppcheck

## Actions

### `gz quality run`

Run formatters and linters for detected languages.

```bash
gz quality run [path] [flags]
```

**Arguments:**
- `path` - Target directory or file (default: current directory)

**Flags:**
- `--languages` - Comma-separated list of languages to process
- `--tools` - Specific tools to run (comma-separated)
- `--exclude-tools` - Tools to exclude (comma-separated)
- `--auto-fix` - Automatically fix issues when possible (default: true)
- `--fail-on-error` - Exit with error code if issues found (default: true)
- `--parallel` - Run tools in parallel (default: true)
- `--output` - Output format: text, json, sarif, checkstyle
- `--output-file` - Save results to file

**Examples:**
```bash
# Run all quality checks
gz quality run

# Check specific directory
gz quality run ./src

# Run only Go tools
gz quality run --languages go

# Run specific tools
gz quality run --tools gofumpt,golangci-lint

# Generate SARIF report
gz quality run --output sarif --output-file quality.sarif

# Don't auto-fix issues
gz quality run --auto-fix=false
```

### `gz quality check`

Check code quality without making changes.

```bash
gz quality check [path] [flags]
```

**Flags:**
- Same as `run` but with `--auto-fix=false` by default

**Examples:**
```bash
# Check without fixing
gz quality check

# Check with specific output format
gz quality check --output json
```

### `gz quality fix`

Fix code quality issues automatically.

```bash
gz quality fix [path] [flags]
```

**Flags:**
- `--languages` - Languages to fix
- `--tools` - Specific tools to use for fixing
- `--backup` - Create backup files before fixing (default: false)

**Examples:**
```bash
# Fix all issues
gz quality fix

# Fix only formatting issues
gz quality fix --tools gofumpt,black,prettier

# Fix with backup
gz quality fix --backup
```

### `gz quality install`

Install or update quality tools.

```bash
gz quality install [tools] [flags]
```

**Arguments:**
- `tools` - Specific tools to install (optional, installs all if omitted)

**Flags:**
- `--languages` - Install tools for specific languages
- `--force` - Force reinstall even if already present
- `--version` - Install specific versions (format: tool@version)

**Examples:**
```bash
# Install all tools
gz quality install

# Install Go tools only
gz quality install --languages go

# Install specific tools
gz quality install golangci-lint black prettier

# Install specific versions
gz quality install --version golangci-lint@1.54.2,black@23.7.0

# Force reinstall
gz quality install --force
```

### `gz quality list`

List available and installed tools.

```bash
gz quality list [flags]
```

**Flags:**
- `--languages` - Filter by languages
- `--installed` - Show only installed tools
- `--available` - Show only available tools
- `--output` - Output format: table, json, yaml

**Examples:**
```bash
# List all tools
gz quality list

# Show installed tools
gz quality list --installed

# Show Go tools only
gz quality list --languages go

# JSON output
gz quality list --output json
```

### `gz quality init`

Initialize quality configuration for a project.

```bash
gz quality init [flags]
```

**Flags:**
- `--languages` - Languages to configure
- `--template` - Configuration template: minimal, standard, strict
- `--output` - Output file name (default: quality.yaml)
- `--interactive` - Interactive configuration

**Examples:**
```bash
# Initialize with auto-detection
gz quality init

# Initialize for specific languages
gz quality init --languages go,python,javascript

# Use strict template
gz quality init --template strict

# Interactive setup
gz quality init --interactive
```

### `gz quality analyze`

Analyze code quality metrics and trends.

```bash
gz quality analyze [path] [flags]
```

**Flags:**
- `--baseline` - Baseline report for comparison
- `--output` - Output format: text, json, html
- `--output-file` - Save analysis to file
- `--metrics` - Specific metrics to analyze

**Examples:**
```bash
# Basic analysis
gz quality analyze

# Compare with baseline
gz quality analyze --baseline baseline-report.json

# Generate HTML report
gz quality analyze --output html --output-file quality-report.html
```

### `gz quality validate`

Validate quality configuration file.

```bash
gz quality validate --config <config-file> [flags]
```

**Flags:**
- `--config` - Configuration file to validate (required)
- `--schema` - Schema file for validation

**Examples:**
```bash
# Validate configuration
gz quality validate --config quality.yaml

# Validate with custom schema
gz quality validate --config quality.yaml --schema custom-schema.json
```

## Configuration

### Configuration File Structure

```yaml
version: "1.0"

# Global settings
enabled: true
auto_fix: true
parallel: true
fail_on_error: false

ignore_patterns:
  - "vendor/"
  - "node_modules/"
  - ".git/"
  - "dist/"

# Language configurations
languages:
  go:
    enabled: true
    formatters:
      gofumpt:
        enabled: true
        extra: true
      gci:
        enabled: true
        sections:
          - "standard"
          - "default"
          - "prefix(github.com/yourorg)"
    linters:
      golangci_lint:
        enabled: true
        fix: true
        config: ".golangci.yml"

  python:
    enabled: true
    python_version: "3.11"
    formatters:
      black:
        enabled: true
        line_length: 88
      isort:
        enabled: true
        profile: "black"
    linters:
      ruff:
        enabled: true
        fix: true

  javascript:
    enabled: true
    package_manager: "npm"
    formatters:
      prettier:
        enabled: true
    linters:
      eslint:
        enabled: true
        fix: true

# CI/CD integration
ci:
  enabled: true
  report_format: "sarif"
  report_path: "quality-report.sarif"
  annotations: true
```

### Tool-Specific Configuration

Each tool can be configured with specific options:

```yaml
languages:
  go:
    linters:
      golangci_lint:
        enabled: true
        config: ".golangci.yml"
        fix: true
        timeout: "5m"
        args:
          - "--fast"
          - "--verbose"
        presets:
          - "bugs"
          - "performance"
          - "format"
```

## Environment Variables

```bash
# Tool behavior
export GZ_QUALITY_AUTO_FIX="true"
export GZ_QUALITY_PARALLEL="true"
export GZ_QUALITY_FAIL_ON_ERROR="false"

# Tool paths
export GZ_QUALITY_CONFIG="~/.config/gzh-manager/quality.yaml"
export GZ_QUALITY_CACHE_DIR="~/.cache/gzh-manager/quality"

# Language-specific
export GZ_QUALITY_GO_VERSION="1.21"
export GZ_QUALITY_PYTHON_VERSION="3.11"
export GZ_QUALITY_NODE_VERSION="18"
```

## Language-Specific Examples

### Go Projects

```bash
# Run Go quality checks
gz quality run --languages go

# Format with gofumpt and organize imports
gz quality run --tools gofumpt,gci

# Full Go linting
gz quality run --tools golangci-lint --auto-fix=false
```

### Python Projects

```bash
# Format Python code
gz quality run --languages python --tools black,isort

# Run all Python checks
gz quality run --languages python

# Type checking only
gz quality run --tools mypy
```

### JavaScript/TypeScript Projects

```bash
# Format and lint JavaScript
gz quality run --languages javascript

# TypeScript compilation check
gz quality run --languages typescript --tools tsc

# Prettier formatting only
gz quality run --tools prettier
```

### Multi-Language Projects

```bash
# Run all supported languages
gz quality run

# Specific language combination
gz quality run --languages go,python,javascript

# Web development stack
gz quality run --languages javascript,typescript --tools prettier,eslint
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Code Quality
on: [push, pull_request]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install gz
        run: |
          # Install gz binary

      - name: Run Quality Checks
        run: |
          gz quality run --output sarif --output-file quality.sarif

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: quality.sarif
```

### GitLab CI

```yaml
quality:
  stage: test
  script:
    - gz quality install
    - gz quality run --output json --output-file quality.json
  artifacts:
    reports:
      junit: quality.json
    expire_in: 1 week
```

## Performance Optimization

### Parallel Execution

```bash
# Enable parallel execution (default)
gz quality run --parallel

# Disable for resource-constrained environments
gz quality run --parallel=false
```

### Caching

```bash
# Enable tool caching (default)
gz quality run --cache

# Clear cache
gz quality run --clear-cache

# Custom cache directory
gz quality run --cache-dir /tmp/quality-cache
```

### Selective Execution

```bash
# Run only formatters
gz quality run --tools gofumpt,black,prettier

# Skip slow tools
gz quality run --exclude-tools spotbugs,mypy

# Quick checks only
gz quality run --profile quick
```

## Troubleshooting

### Common Issues

1. **Tool Not Found**
   ```bash
   # Install missing tools
   gz quality install

   # Check tool status
   gz quality list --installed
   ```

2. **Configuration Error**
   ```bash
   # Validate configuration
   gz quality validate --config quality.yaml

   # Use debug mode
   gz quality run --debug
   ```

3. **Permission Issues**
   ```bash
   # Fix file permissions
   chmod +x /path/to/tool

   # Run with verbose output
   gz quality run --verbose
   ```

### Debug Mode

```bash
# Enable debug logging
gz quality run --debug

# Show tool commands
gz quality run --verbose --debug

# Dry run to see what would be executed
gz quality run --dry-run
```

## Integration with Other Commands

### With Git Operations

```bash
# Quality check before commit
gz quality run && git commit -m "Add feature"

# Pre-commit hook integration
echo 'gz quality run --auto-fix' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### With CI/CD

```bash
# Generate reports for CI
gz quality run --output sarif --fail-on-error

# Quality gate in pipeline
gz quality check --fail-on-error --quiet
```

## Related Commands

- [`gz git repo clone-or-update`](git.md#repo-clone-or-update) - Repository management
- [`gz profile`](profile.md) - Performance profiling

## See Also

- [Code Quality Management Guide](../03-core-features/quality-management.md)
- [Configuration Schema](../04-configuration/schemas/quality-schema.yaml)
- [Quality Examples](../../examples/quality/)
