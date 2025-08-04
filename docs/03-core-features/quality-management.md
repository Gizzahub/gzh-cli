# Code Quality Management Guide

The `gz quality` command provides a unified interface for managing code quality across multiple programming languages, integrating various formatting and linting tools into a single, cohesive workflow.

## Overview

Modern projects often use multiple programming languages, each with its own set of quality tools. The `gz quality` command eliminates the complexity of managing these tools individually by providing a unified interface that automatically detects your project's languages and runs the appropriate tools.

## Supported Languages and Tools

### Programming Languages

#### Go
- **gofumpt**: Stricter Go formatter
- **golangci-lint**: Comprehensive Go linter
- **goimports**: Import statement formatter
- **gci**: Import grouping and sorting

#### Python
- **ruff**: Fast Python formatter and linter (recommended)
- **black**: Opinionated Python formatter
- **isort**: Import statement sorter
- **flake8**: Style guide enforcement
- **mypy**: Static type checker

#### JavaScript/TypeScript
- **prettier**: Opinionated code formatter
- **eslint**: Pluggable linting utility
- **dprint**: Fast formatter alternative

#### Rust
- **rustfmt**: Official Rust formatter
- **clippy**: Official Rust linter

#### Java
- **google-java-format**: Google's Java formatter
- **checkstyle**: Code style checker
- **spotbugs**: Bug pattern detector

#### C/C++
- **clang-format**: LLVM formatter
- **clang-tidy**: LLVM linter and static analyzer

### Other Formats
- **YAML**: yamllint, prettier
- **JSON**: jq, prettier
- **Markdown**: markdownlint, prettier
- **Shell**: shellcheck, shfmt
- **Dockerfile**: hadolint
- **Terraform**: terraform fmt, tflint

## Command Reference

### Run Quality Checks

Execute all applicable quality tools:

```bash
# Run all formatters and linters
gz quality run

# Run on specific files
gz quality run --files src/main.go,src/utils.py

# Run on changed files only
gz quality run --changed

# Run on staged files only
gz quality run --staged

# Run specific tools only
gz quality run --tools gofumpt,black,prettier
```

### Check Mode (Lint Only)

Run linters without making changes:

```bash
# Check all files
gz quality check

# Check with specific severity
gz quality check --severity error

# Output in different formats
gz quality check --format json
gz quality check --format junit
gz quality check --format sarif
```

### Initialize Project

Generate configuration files for quality tools:

```bash
# Interactive initialization
gz quality init

# Generate specific configs
gz quality init --tools golangci-lint,prettier,ruff

# Use recommended settings
gz quality init --preset recommended

# Use strict settings
gz quality init --preset strict
```

### Analyze Project

Get recommendations for your project:

```bash
# Analyze project structure
gz quality analyze

# Get detailed report
gz quality analyze --detailed

# Include performance metrics
gz quality analyze --benchmark
```

### Tool Management

Install and manage quality tools:

```bash
# Install all recommended tools
gz quality install

# Install specific tools
gz quality install gofumpt ruff prettier

# Upgrade tools to latest versions
gz quality upgrade

# List installed tools and versions
gz quality version

# Run specific tool directly
gz quality tool prettier --write "**/*.{js,ts,json}"
```

## Configuration

### Configuration File

Create `~/.config/gzh-manager/quality.yaml`:

```yaml
quality:
  # Tool configuration
  tools:
    enabled:
      - gofumpt
      - golangci-lint
      - ruff
      - prettier
      - eslint
    disabled: []

  # Tool-specific settings
  settings:
    golangci-lint:
      config: .golangci.yml
      timeout: 5m

    prettier:
      config: .prettierrc
      ignore: .prettierignore

    ruff:
      line-length: 88
      select: ["E", "F", "W"]

  # Execution settings
  execution:
    parallel: true
    max_workers: 4
    timeout: 300
    fail_fast: false

  # File filtering
  filters:
    include_patterns:
      - "**/*.go"
      - "**/*.py"
      - "**/*.js"
      - "**/*.ts"
    exclude_patterns:
      - "vendor/**"
      - "node_modules/**"
      - "*.generated.*"
      - "*.pb.go"

  # Output settings
  output:
    format: "colored"  # colored, plain, json, junit
    verbose: false
    show_diff: true
    group_by_file: true
```

### Environment Variables

```bash
# Enable/disable specific tools
export QUALITY_ENABLE_TOOLS="gofumpt,black,prettier"
export QUALITY_DISABLE_TOOLS="mypy"

# Execution settings
export QUALITY_PARALLEL=true
export QUALITY_TIMEOUT=600
export QUALITY_FAIL_FAST=false

# Output format
export QUALITY_OUTPUT_FORMAT=json
export QUALITY_VERBOSE=true

# CI/CD mode
export QUALITY_CI_MODE=true
```

## Usage Examples

### Basic Workflow

```bash
# 1. Analyze your project
$ gz quality analyze
📊 Project Analysis:
  - Go: 45% (120 files)
  - Python: 30% (80 files)
  - JavaScript: 20% (53 files)
  - YAML: 5% (13 files)

Recommended tools:
  ✓ gofumpt, golangci-lint (Go)
  ✓ ruff (Python)
  ✓ prettier, eslint (JavaScript)

# 2. Initialize configurations
$ gz quality init
🔧 Generating configuration files...
  ✓ .golangci.yml
  ✓ .ruff.toml
  ✓ .prettierrc
  ✓ .eslintrc.json

# 3. Install tools
$ gz quality install
📦 Installing quality tools...
  ✓ gofumpt v0.6.0
  ✓ golangci-lint v1.59.0
  ✓ ruff v0.5.0
  ✓ prettier v3.3.0
  ✓ eslint v9.0.0

# 4. Run quality checks
$ gz quality run
🔍 Running quality checks...

Go files (120):
  ✓ gofumpt: 118 files formatted
  ✓ goimports: imports organized
  ⚠️ golangci-lint: 3 warnings in 2 files

Python files (80):
  ✓ ruff: 78 files formatted
  ✓ ruff: no linting issues

JavaScript files (53):
  ✓ prettier: 52 files formatted
  ⚠️ eslint: 5 warnings in 3 files

Summary: 248 files formatted, 8 warnings
```

### CI/CD Integration

```bash
# Check mode for CI
gz quality check --severity error --format junit > quality-report.xml

# GitHub Actions example
- name: Code Quality Check
  run: |
    gz quality check --format sarif > results.sarif

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: results.sarif
```

### Pre-commit Hook

```bash
# Check staged files before commit
gz quality run --staged --fail-fast

# As a git hook (.git/hooks/pre-commit)
#!/bin/bash
gz quality run --staged --fail-fast || exit 1
```

## Advanced Features

### Custom Tool Configuration

```bash
# Add custom tool
gz quality config add-tool \
  --name "custom-linter" \
  --command "custom-lint {files}" \
  --extensions ".custom,.special"

# Override tool command
gz quality config set-command \
  --tool prettier \
  --command "prettier --single-quote --no-semi"
```

### Performance Optimization

```bash
# Benchmark tool performance
gz quality benchmark

# Run with profiling
gz quality run --profile

# Optimize for large codebases
gz quality run --batch-size 50 --cache-results
```

### Integration with Other Commands

```bash
# Run quality checks after syncing repos
gz synclone github --org myorg && gz quality run

# Monitor quality in real-time
gz quality run --watch

# Include in development workflow
gz dev-env setup && gz quality init && gz quality install
```

## Language-Specific Guides

### Go Projects

```bash
# Recommended setup for Go
gz quality init --preset go-strict
gz quality install gofumpt golangci-lint gci

# Common Go quality workflow
gz quality run --tools gofumpt,gci  # Format first
gz quality check --tools golangci-lint  # Then lint
```

### Python Projects

```bash
# Modern Python setup (ruff is fastest)
gz quality init --tools ruff
gz quality install ruff

# Legacy Python setup
gz quality init --tools black,isort,flake8,mypy
gz quality install black isort flake8 mypy
```

### Frontend Projects

```bash
# JavaScript/TypeScript setup
gz quality init --preset frontend
gz quality install prettier eslint

# With framework-specific configs
gz quality init --preset react  # or vue, angular
```

## Troubleshooting

### Common Issues

1. **"Tool not found"**
   ```bash
   # Install missing tool
   gz quality install <tool-name>

   # Or install all required tools
   gz quality install
   ```

2. **"Configuration conflicts"**
   ```bash
   # Validate configurations
   gz quality validate

   # Show tool precedence
   gz quality config show --precedence
   ```

3. **"Slow performance"**
   ```bash
   # Enable caching
   gz quality run --cache

   # Limit parallelism
   gz quality run --max-workers 2

   # Process in batches
   gz quality run --batch-size 20
   ```

### Debug Mode

```bash
# Verbose output
gz quality run --debug

# Show tool commands
gz quality run --dry-run --show-commands

# Test specific tool
gz quality test-tool gofumpt --file main.go
```

## Best Practices

1. **Start with Analysis**: Use `gz quality analyze` to understand your project
2. **Use Presets**: Start with presets and customize as needed
3. **CI/CD Integration**: Use check mode in CI to prevent regressions
4. **Incremental Adoption**: Enable tools gradually for large projects
5. **Cache Results**: Use caching for better performance in large codebases
6. **Regular Updates**: Keep tools updated with `gz quality upgrade`

## Related Documentation

- [Pre-commit Hooks](../06-development/pre-commit-hooks.md)
- [CI/CD Integration](../07-deployment/)
- [Configuration Guide](../04-configuration/configuration-guide.md)
- [Development Workflow](../06-development/)
