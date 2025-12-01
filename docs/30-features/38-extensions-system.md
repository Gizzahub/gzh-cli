# Extensions System

gz provides a powerful extension system that allows users to customize and extend functionality without modifying source code.

## Overview

The extension system supports three main features:

1. **Simple Aliases** - Short names for frequently used commands
1. **Multi-Step Workflows** - Sequential execution of multiple commands
1. **Parameterized Aliases** - Reusable commands with variable substitution
1. **External Commands** - Integration of external tools as subcommands

## Configuration File

**Location**: `~/.config/gzh-manager/extensions.yaml`

The configuration file is automatically loaded at startup. If the file doesn't exist, gz continues without errors.

## Simple Aliases

Create shortcuts for commonly used commands.

### Example

```yaml
aliases:
  update-all:
    command: "pm update --all"
    description: "Update all package managers"

  pull-all:
    command: "git repo pull-all"
    description: "Pull all repositories recursively"
```

### Usage

```bash
gz update-all
gz pull-all
```

## Multi-Step Workflows

Execute multiple commands in sequence with progress tracking.

### Example

```yaml
aliases:
  full-sync:
    description: "Complete synchronization workflow"
    steps:
      - "synclone run"
      - "pm update --all"
      - "git repo pull-all"

  deploy:
    description: "Build and deploy workflow"
    steps:
      - "quality check"
      - "git repo pull-all"
      - "pm update --all"
```

### Behavior

- **Sequential Execution**: Each step runs only if the previous step succeeds
- **Progress Tracking**: Clear indicators for each step
- **Fail-Fast**: Execution stops on first error
- **Error Reporting**: Detailed error messages with step number

### Output

```
🔄 Step 1/3: synclone run
... command output ...
✅ Step 1/3 completed

🔄 Step 2/3: pm update --all
... command output ...
✅ Step 2/3 completed

🔄 Step 3/3: git repo pull-all
... command output ...
✅ Step 3/3 completed

🎉 All steps completed successfully!
```

## Parameterized Aliases

Create reusable commands with variable substitution.

### Example

```yaml
aliases:
  clone-and-setup:
    command: "git repo clone-or-update ${url} && dev-env bootstrap"
    description: "Clone repository and setup development environment"
    params:
      - name: url
        description: "Repository URL to clone"
        required: true

  check-path:
    command: "quality check ${path}"
    description: "Run quality checks on specific path"
    params:
      - name: path
        description: "Path to check (default: current directory)"
        required: false
```

### Variable Syntax

Both syntaxes are supported:

- `${variable}` - Recommended
- `$variable` - Also works

### Parameter Types

**Required Parameters**:

```yaml
params:
  - name: url
    description: "Repository URL"
    required: true
```

Shown as `<url>` in help text. Command fails if not provided.

**Optional Parameters**:

```yaml
params:
  - name: path
    description: "Path to check"
    required: false
```

Shown as `[path]` in help text. Can be omitted.

### Usage

```bash
# Required parameter
gz clone-and-setup https://github.com/user/repo.git

# Optional parameter
gz check-path ./src
gz check-path  # Uses default

# Extra arguments after parameters
gz clone-and-setup https://github.com/user/repo.git --branch main
```

## External Commands

Integrate external tools as gz subcommands.

### Example

```yaml
external:
  - name: terraform
    command: /usr/local/bin/terraform
    description: "Terraform infrastructure management"
    passthrough: true

  - name: custom-lint
    command: /opt/tools/custom-lint
    description: "Custom linting tool"
    args:
      - "--config"
      - "$HOME/.config/lint.yaml"
```

### Configuration Options

- **name**: Subcommand name (e.g., `gz terraform`)
- **command**: Path to external executable
- **description**: Help text description
- **passthrough**: Pass all arguments directly (disables flag parsing)
- **args**: Default arguments prepended to user arguments

### Behavior

- Command availability is checked at startup
- Missing commands show warning but don't block gz
- External commands run with user's environment
- stdin/stdout/stderr are directly connected

### Usage

```bash
gz terraform plan
gz terraform apply

gz custom-lint src/
```

## Experimental Features

Commands can be marked as experimental and are disabled by default.

### Enabling Experimental Features

**Via Environment Variable**:

```bash
export GZ_EXPERIMENTAL=1
gz experimental-command
```

**Via Flag**:

```bash
gz --experimental experimental-command
```

### Lifecycle Stages

Commands progress through lifecycle stages:

- **Stable**: Production-ready (no warnings)
- **Beta**: Feature-complete, testing phase (info message)
- **Experimental**: Early development (requires enablement)
- **Deprecated**: Will be removed (warning message)

## Advanced Patterns

### Combining Features

```yaml
aliases:
  # Workflow with parameterized steps
  deploy-to:
    description: "Deploy to specific environment"
    command: "deploy-workflow ${env}"
    params:
      - name: env
        description: "Environment (dev/staging/prod)"
        required: true

  deploy-workflow:
    description: "Internal deployment workflow"
    steps:
      - "quality check"
      - "git repo pull-all"
      - "external-deploy ${env}"
```

### Environment Variables in Commands

```yaml
aliases:
  backup:
    command: "synclone run --output ${BACKUP_DIR}/repos-backup"
    description: "Backup all repositories"
```

Usage:

```bash
export BACKUP_DIR=/mnt/backup
gz backup
```

## Best Practices

### Naming Conventions

- Use kebab-case for alias names
- Keep names short and memorable
- Use descriptive names for workflows
- Prefix internal workflows with underscore

### Command Organization

```yaml
aliases:
  # User-facing commands
  sync:
    description: "Quick sync"
    steps:
      - "_sync-repos"
      - "_sync-packages"

  # Internal workflows (prefix with _)
  _sync-repos:
    command: "git repo pull-all"
    description: "Internal: sync repositories"

  _sync-packages:
    command: "pm update --all"
    description: "Internal: update packages"
```

### Error Handling

- Workflows stop on first error
- Check command prerequisites in first step
- Use meaningful descriptions for debugging
- Test workflows in safe environment first

### Security Considerations

- Validate external command paths
- Avoid embedding secrets in config
- Use environment variables for credentials
- Review external commands before adding
- Keep extensions.yaml in version control (without secrets)

## Troubleshooting

### Alias Not Found

Check if extensions.yaml is in correct location:

```bash
ls ~/.config/gzh-manager/extensions.yaml
```

### Workflow Step Fails

Run steps individually to identify issue:

```bash
gz step-that-failed --help
```

### Parameter Not Substituted

Verify parameter name matches exactly:

```yaml
params:
  - name: url  # Must match ${url} in command
```

### External Command Not Found

Verify command path:

```bash
which terraform
# Update command path in extensions.yaml
```

## Examples

See [examples/extensions.yaml](../../examples/extensions.yaml) for complete examples.

## See Also

- [Command Lifecycle Management](39-lifecycle-management.md)
- [Configuration Guide](../40-configuration/40-configuration-guide.md)
- [Architecture Documentation](../20-architecture/24-improvement-strategies.md)
