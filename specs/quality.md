# Code Quality Management Specification

## Overview

The `quality` command provides comprehensive code formatting and linting capabilities for multiple programming languages. It integrates various quality tools into a unified interface, supporting automatic tool detection, project analysis, configuration management, and standardized execution across different programming languages and frameworks.

## Commands

### Core Commands

- `gz quality run` - Execute all formatting and linting tools (default)
- `gz quality check` - Run linting only (no changes, validation only)
- `gz quality init` - Generate project configuration files automatically

### Tool Management Commands

- `gz quality tool` - Execute individual quality tools directly
- `gz quality analyze` - Analyze project and show recommended tools
- `gz quality install` - Install quality tools
- `gz quality upgrade` - Upgrade quality tools
- `gz quality version` - Show quality tool versions
- `gz quality list` - List available quality tools

### Run All Tools (`gz quality run`)

**Purpose**: Execute all configured formatting and linting tools for the current project

**Features**:
- Multi-language support with automatic detection
- Parallel tool execution for performance
- Configurable tool selection
- Format-only and lint-only modes
- Changed files and staged files filtering
- Comprehensive reporting

**Usage**:
```bash
gz quality run                                          # Run all tools on all files
gz quality run --format-only                            # Run formatting tools only
gz quality run --lint-only                              # Run linting tools only
gz quality run --changed                                # Process only changed files
gz quality run --staged                                 # Process only staged files
gz quality run --language go                            # Process only Go files
```

**Parameters**:
- `--format-only`: Execute only formatting tools
- `--lint-only`: Execute only linting tools  
- `--changed`: Process only changed files (git diff)
- `--staged`: Process only staged files (git diff --cached)
- `--language`: Filter by specific programming language
- `--parallel`: Enable parallel tool execution (default: true)
- `--timeout`: Set timeout for tool execution (default: 300s)

### Check Only (`gz quality check`)

**Purpose**: Run linting and validation without making any changes to files

**Features**:
- Read-only validation mode
- Comprehensive error reporting
- Exit code indication for CI/CD
- Severity-based filtering
- Multiple output formats

**Usage**:
```bash
gz quality check                                        # Check all files
gz quality check --severity error                       # Show only errors
gz quality check --format json                          # JSON output for CI/CD
gz quality check --changed                              # Check only changed files
```

**Parameters**:
- `--severity`: Filter by severity level (error, warning, info, all)
- `--format`: Output format (text, json, junit)
- `--changed`: Check only changed files
- `--staged`: Check only staged files
- `--fail-on-warning`: Exit with error code on warnings

### Initialize Configuration (`gz quality init`)

**Purpose**: Generate project-specific configuration files for quality tools

**Features**:
- Automatic project type detection
- Language-specific configurations
- Template-based configuration generation
- Integration with existing project structure
- Customizable configuration templates

**Usage**:
```bash
gz quality init                                         # Initialize for detected languages
gz quality init --language go                           # Initialize for specific language
gz quality init --template strict                       # Use strict configuration template
gz quality init --force                                 # Overwrite existing configurations
```

**Parameters**:
- `--language`: Target specific programming language
- `--template`: Configuration template (default, strict, permissive)
- `--force`: Overwrite existing configuration files
- `--dry-run`: Preview configuration changes without creating files

**Supported Languages and Tools**:
- **Go**: gofumpt, golangci-lint, goimports, gci
- **Python**: ruff (format + lint), black, isort, flake8, mypy
- **JavaScript/TypeScript**: prettier, eslint, dprint
- **Rust**: rustfmt, clippy
- **Java**: google-java-format, checkstyle, spotbugs
- **C/C++**: clang-format, clang-tidy
- **CSS/SCSS**: stylelint, prettier
- **HTML**: djlint, prettier
- **YAML**: yamllint, prettier
- **JSON**: prettier, jq
- **Markdown**: markdownlint, prettier
- **Shell**: shfmt, shellcheck

## Tool Management

### Direct Tool Execution (`gz quality tool`)

**Purpose**: Execute individual quality tools with direct parameter passing

**Features**:
- Direct tool parameter forwarding
- Tool-specific help and documentation
- Custom tool configuration
- Integration with project settings

**Usage**:
```bash
gz quality tool gofumpt --help                          # Show tool-specific help
gz quality tool ruff --changed                          # Run ruff on changed files
gz quality tool prettier --staged                       # Run prettier on staged files
gz quality tool eslint src/ --fix                       # Run eslint with specific parameters
```

**Common Tool Patterns**:
- `--changed`: Process only changed files (git diff)
- `--staged`: Process only staged files (git diff --cached)
- `--help`: Show tool-specific help
- Tool-specific parameters are forwarded directly

**Supported Tools**:
```bash
# Go tools
gz quality tool gofumpt [files...]                      # Go formatter
gz quality tool golangci-lint run [path]                # Go linter
gz quality tool goimports [files...]                    # Go import formatter
gz quality tool gci [files...]                          # Go import organizer

# Python tools
gz quality tool ruff format [files...]                  # Python formatter
gz quality tool ruff check [files...]                   # Python linter
gz quality tool black [files...]                        # Python formatter
gz quality tool isort [files...]                        # Python import sorter

# JavaScript/TypeScript tools
gz quality tool prettier [files...]                     # Multi-language formatter
gz quality tool eslint [files...]                       # JavaScript linter
gz quality tool dprint fmt [files...]                   # Multi-language formatter

# Rust tools
gz quality tool rustfmt [files...]                      # Rust formatter
gz quality tool clippy                                  # Rust linter

# Other tools
gz quality tool shfmt [files...]                        # Shell formatter
gz quality tool shellcheck [files...]                   # Shell linter
gz quality tool markdownlint [files...]                 # Markdown linter
```

### Project Analysis (`gz quality analyze`)

**Purpose**: Analyze project structure and recommend appropriate quality tools

**Features**:
- Automatic language detection
- File type analysis
- Configuration detection
- Tool recommendation engine
- Dependency analysis

**Usage**:
```bash
gz quality analyze                                       # Analyze current project
gz quality analyze --detailed                           # Show detailed analysis
gz quality analyze --recommend-only                     # Show only recommendations
gz quality analyze --format json                        # JSON output
```

**Analysis Output**:
```
📊 Project Quality Analysis
===========================

🔍 Detected Languages:
  - Go (85 files, 15,420 lines)
  - JavaScript (12 files, 2,341 lines)  
  - YAML (8 files, 456 lines)

🛠️  Recommended Tools:
  - gofumpt (Go formatting)
  - golangci-lint (Go linting)
  - prettier (JavaScript/YAML formatting)
  - eslint (JavaScript linting)

⚙️  Existing Configurations:
  - .golangci.yml (golangci-lint)
  - .eslintrc.json (eslint)
  - Missing: .prettierrc.json

💡 Suggestions:
  - Run 'gz quality init' to generate missing configurations
  - Consider adding pre-commit hooks
  - Update golangci-lint to latest version
```

### Tool Installation (`gz quality install`)

**Purpose**: Install quality tools with version management

**Features**:
- Automatic tool installation
- Version specification support
- System-wide and project-local installation
- Dependency management
- Installation verification

**Usage**:
```bash
gz quality install                                      # Install recommended tools
gz quality install gofumpt                              # Install specific tool
gz quality install --all                                # Install all supported tools
gz quality install --version v1.2.3 golangci-lint      # Install specific version
```

**Parameters**:
- `--all`: Install all available quality tools
- `--version`: Specify tool version to install
- `--global`: Install system-wide (default: project-local)
- `--force`: Force reinstallation

### Tool Upgrade (`gz quality upgrade`)

**Purpose**: Upgrade quality tools to latest versions

**Features**:
- Selective tool upgrading
- Version compatibility checking
- Backup of previous versions
- Upgrade verification

**Usage**:
```bash
gz quality upgrade                                       # Upgrade all installed tools
gz quality upgrade gofumpt                              # Upgrade specific tool
gz quality upgrade --check                              # Check for available upgrades
```

### Version Information (`gz quality version`)

**Purpose**: Display version information for all quality tools

**Features**:
- Comprehensive version listing
- Update availability checking
- Version compatibility matrix
- Installation status reporting

**Usage**:
```bash
gz quality version                                       # Show all tool versions
gz quality version --check-updates                      # Check for updates
gz quality version --format json                        # JSON output
```

**Version Output**:
```
🔧 Quality Tools Versions
=========================

✅ Go Tools:
  - gofumpt: v0.5.0 (latest)
  - golangci-lint: v1.54.2 (v1.55.0 available)
  - goimports: v0.14.0 (latest)

✅ Python Tools:
  - ruff: v0.1.4 (latest)
  - black: v23.9.1 (latest)

⚠️  JavaScript Tools:
  - prettier: v3.0.3 (v3.1.0 available)
  - eslint: v8.50.0 (latest)

💡 Run 'gz quality upgrade' to update outdated tools
```

### List Available Tools (`gz quality list`)

**Purpose**: Display all available quality tools with their capabilities

**Features**:
- Comprehensive tool listing
- Language categorization
- Tool capability description
- Installation status indication

**Usage**:
```bash
gz quality list                                         # List all available tools
gz quality list --language go                           # List tools for specific language
gz quality list --installed-only                        # Show only installed tools
gz quality list --format json                           # JSON output
```

## Configuration

### Project Configuration

Quality tools can be configured through project-specific configuration files:

**Global Configuration** (`quality.yaml`):
```yaml
quality:
  tools:
    enabled: ["gofumpt", "golangci-lint", "prettier", "eslint"]
    disabled: []
  
  execution:
    parallel: true
    timeout: 300
    fail_fast: false
  
  filters:
    exclude_patterns:
      - "vendor/"
      - "node_modules/"
      - "*.generated.go"
    
  languages:
    go:
      tools: ["gofumpt", "goimports", "golangci-lint"]
    javascript:
      tools: ["prettier", "eslint"]
    python:
      tools: ["ruff"]
```

**Tool-Specific Configurations**:
- `.golangci.yml` - golangci-lint configuration
- `.prettierrc.json` - Prettier formatting rules
- `.eslintrc.json` - ESLint linting rules
- `ruff.toml` - Ruff configuration
- `.yamllint.yml` - YAML linting rules

### Environment Variables

- `QUALITY_CONFIG_PATH`: Override default configuration file location
- `QUALITY_TOOLS_PATH`: Custom path for tool installations
- `QUALITY_PARALLEL`: Enable/disable parallel execution
- `QUALITY_TIMEOUT`: Default timeout for tool execution
- `QUALITY_DEBUG`: Enable debug logging

### Configuration Templates

**Default Template**: Balanced configuration with moderate rules
**Strict Template**: Strict rules for production code quality
**Permissive Template**: Relaxed rules for development/prototyping

## Output Formats

### Text Output (Default)

```
🚀 Running Quality Tools
========================

✅ gofumpt: Formatted 15 files
✅ golangci-lint: No issues found
⚠️  prettier: Fixed 3 formatting issues
❌ eslint: 2 errors, 5 warnings found

📊 Summary:
  - 4 tools executed
  - 18 files processed
  - 3 formatting fixes applied
  - 2 errors, 5 warnings found
```

### JSON Output

```json
{
  "timestamp": "2025-01-04T10:30:00Z",
  "tools_executed": 4,
  "files_processed": 18,
  "execution_time": "12.5s",
  "results": [
    {
      "tool": "gofumpt",
      "status": "success",
      "files_changed": 15,
      "execution_time": "2.1s"
    },
    {
      "tool": "eslint", 
      "status": "error",
      "errors": 2,
      "warnings": 5,
      "execution_time": "3.8s"
    }
  ],
  "summary": {
    "success": true,
    "total_errors": 2,
    "total_warnings": 5,
    "files_formatted": 18
  }
}
```

### JUnit XML Output (for CI/CD)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="quality-check" tests="4" errors="2" warnings="5">
  <testsuite name="gofumpt" tests="1" errors="0">
    <testcase name="formatting" status="passed"/>
  </testsuite>
  <testsuite name="eslint" tests="1" errors="2">
    <testcase name="linting" status="failed">
      <error message="2 errors, 5 warnings found"/>
    </testcase>
  </testsuite>
</testsuites>
```

## Integration

### Git Integration

**Pre-commit Hooks**:
```bash
# Install pre-commit hooks
gz quality init --hooks

# Manual pre-commit check
gz quality check --staged
```

**Git Workflow Integration**:
- Automatic detection of changed files (`--changed`, `--staged`)
- Integration with git hooks for automated quality checks
- Support for git worktrees and submodules

### CI/CD Integration

**GitHub Actions Example**:
```yaml
- name: Code Quality Check
  run: |
    gz quality check --format junit --fail-on-warning
    
- name: Upload Quality Report
  uses: mikepenz/action-junit-report@v3
  with:
    report_paths: 'quality-report.xml'
```

**Jenkins Pipeline Example**:
```groovy
stage('Quality Check') {
    steps {
        sh 'gz quality check --format junit'
        publishTestResults testResultsPattern: 'quality-report.xml'
    }
}
```

### IDE Integration

- **VS Code**: Integration through tasks and settings
- **JetBrains IDEs**: External tool configuration
- **Vim/Neovim**: Integration through autocmd and plugins
- **Language Server Protocol**: Quality tool integration with LSP

## Examples

### Basic Quality Management

```bash
# Initialize project with quality tools
gz quality init

# Run all quality tools
gz quality run

# Check code quality without changes
gz quality check

# Analyze project and get recommendations
gz quality analyze
```

### Targeted Quality Operations

```bash
# Format only changed files
gz quality run --format-only --changed

# Lint only staged files
gz quality check --staged

# Run specific tool on specific files
gz quality tool prettier src/components/

# Check only Go files
gz quality check --language go
```

### Tool Management Workflow

```bash
# List available tools
gz quality list

# Install recommended tools
gz quality install

# Check tool versions
gz quality version --check-updates

# Upgrade outdated tools  
gz quality upgrade

# Analyze project structure
gz quality analyze --detailed
```

### CI/CD Integration

```bash
# Pre-commit quality check
gz quality check --staged --fail-on-warning

# Full project quality validation
gz quality check --format junit --output quality-report.xml

# Format check for PR validation
gz quality run --format-only --dry-run --changed
```

## Error Handling

### Common Issues

- **Tool not installed**: Automatic installation prompt or error guidance
- **Configuration errors**: Validation and fix suggestions
- **File permission issues**: Clear error messages and resolution steps
- **Tool execution failures**: Detailed error reporting with context
- **Timeout errors**: Configurable timeout handling

### Error Recovery

- **Missing tools**: Automatic installation or installation guidance
- **Configuration issues**: Configuration repair and regeneration
- **Execution failures**: Fallback mechanisms and alternative approaches
- **Performance issues**: Parallel execution optimization and timeout handling

## Performance Considerations

### Execution Optimization

- **Parallel tool execution**: Multiple tools run simultaneously
- **Incremental processing**: Process only changed files when possible
- **Caching**: Tool results caching for repeated executions
- **Smart file filtering**: Efficient file selection based on patterns

### Resource Management

- **Memory usage**: Bounded memory usage for large projects
- **CPU utilization**: Configurable parallel execution limits
- **Disk I/O**: Efficient file processing and temporary file management
- **Network usage**: Minimal network requirements (tool installation only)

## Security Considerations

### Tool Security

- **Tool verification**: Checksum verification for tool installations
- **Sandboxed execution**: Isolated tool execution environment
- **Permission management**: Minimal required permissions for tools
- **Configuration validation**: Safe configuration file processing

### Data Privacy

- **Local execution**: All processing happens locally
- **No data collection**: No telemetry or data collection
- **Configuration privacy**: Sensitive configuration protection
- **File access**: Minimal file system access required

## Best Practices

### Quality Management

- **Consistent configuration**: Use standardized configurations across projects
- **Incremental adoption**: Gradually introduce quality tools
- **Team alignment**: Establish team-wide quality standards
- **Regular updates**: Keep tools and configurations up to date

### Performance Optimization

- **Selective execution**: Use `--changed` and `--staged` flags when appropriate
- **Parallel execution**: Enable parallel processing for large projects
- **Smart filtering**: Configure appropriate file exclusion patterns
- **Tool selection**: Use only necessary tools for each project

### CI/CD Integration

- **Fast feedback**: Use quality checks in early CI/CD stages
- **Comprehensive validation**: Run full quality checks on main branches
- **Fail-fast approach**: Stop builds on critical quality issues
- **Report integration**: Use structured output formats for reporting

## Future Enhancements

### Planned Features

- **Custom tool integration**: Support for project-specific quality tools
- **Quality metrics**: Historical quality tracking and reporting
- **Team dashboards**: Centralized quality monitoring across projects
- **AI-powered suggestions**: Intelligent quality improvement recommendations

### Advanced Capabilities

- **Code complexity analysis**: Advanced code quality metrics
- **Security scanning**: Integration with security analysis tools
- **Performance analysis**: Code performance impact assessment
- **Documentation quality**: Documentation completeness and quality checking

## Implementation Status

- ✅ **Core quality execution**: Multi-tool formatting and linting
- ✅ **Project analysis**: Language detection and tool recommendation
- ✅ **Tool management**: Installation, upgrade, and version management
- ✅ **Configuration management**: Project-specific configuration generation
- ✅ **Git integration**: Changed/staged file processing
- ✅ **CI/CD integration**: Multiple output formats for automation
- 🚧 **Custom tool support**: User-defined tool integration (planned)
- 📋 **Quality metrics**: Historical tracking and reporting (planned)
- 📋 **Team collaboration**: Shared quality standards management (planned)