# Command Reference

Complete reference documentation for all `gz` commands and their options.

## Overview

The `gz` CLI provides 9 main commands for managing development environments, repositories, and code quality:

```
gz <command> [subcommand] [flags]
```

## Commands

### Core Features
- [**synclone**](synclone.md) - Multi-platform repository synchronization
- [**git**](git.md) - Unified Git operations and platform management
- [**quality**](quality.md) - Multi-language code quality management
- [**ide**](ide.md) - JetBrains IDE monitoring and management
- [**profile**](profile.md) - Performance profiling and analysis

### Environment Management
- [**dev-env**](dev-env.md) - Development environment configuration
- [**net-env**](net-env.md) - Network environment transitions
- [**pm**](pm.md) - Package manager updates and management

### Repository Management
- [**repo-config**](repo-config.md) - GitHub repository configuration management

## Global Flags

All commands support these global flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Configuration file path | Auto-detected |
| `--debug` | Enable debug logging | `false` |
| `--help` | Show help information | - |
| `--log-level` | Set log level (debug, info, warn, error) | `info` |
| `--quiet` | Suppress non-error output | `false` |
| `--version` | Show version information | - |

## Configuration

Each command can be configured through:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration files** (YAML/JSON)
4. **Default values** (lowest priority)

### Configuration File Locations

Commands look for configuration in this order:

1. `--config` flag value
2. Current directory: `./<command>.yaml`
3. User config: `~/.config/gzh-manager/<command>.yaml`
4. System config: `/etc/gzh-manager/<command>.yaml`

### Environment Variables

Most commands support environment variable overrides:

```bash
# Authentication
export GITHUB_TOKEN="ghp_..."
export GITLAB_TOKEN="glpat-..."
export GITEA_TOKEN="..."

# Global settings
export GZ_DEBUG="true"
export GZ_LOG_LEVEL="debug"
export GZ_CONFIG_DIR="~/.config/gzh-manager"
```

## Quick Reference

### Most Common Commands

```bash
# Repository operations
gz synclone github --org kubernetes
gz git repo clone-or-update https://github.com/user/repo.git

# Code quality
gz quality install
gz quality run

# IDE monitoring
gz ide monitor

# Performance analysis
gz profile start --type cpu
gz profile analyze profile.pprof

# Environment management
gz pm update --all
gz dev-env sync
```

### Getting Help

```bash
# Show all commands
gz help

# Command-specific help
gz help <command>
gz <command> --help

# Subcommand help
gz <command> <subcommand> --help

# Examples
gz help synclone
gz synclone --help
gz synclone github --help
```

## Exit Codes

All commands use standard exit codes:

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Misuse of command (invalid arguments) |
| `3` | Authentication error |
| `4` | Network error |
| `5` | Configuration error |
| `6` | File system error |

## Version Information

Check version and build information:

```bash
gz version

# Example output:
# gz version 1.0.0
# Built with Go 1.24.0
# Commit: abc123def
# Built: 2025-08-04T10:15:30Z
```

## Troubleshooting

### Common Issues

1. **Command not found**
   ```bash
   # Check installation
   which gz

   # Check PATH
   echo $PATH

   # Reinstall if needed
   make install
   ```

2. **Configuration errors**
   ```bash
   # Validate configuration
   gz <command> validate --config your-config.yaml

   # Show configuration sources
   gz <command> config show
   ```

3. **Authentication issues**
   ```bash
   # Check tokens
   gz doctor auth

   # Test connectivity
   gz <command> test-connection
   ```

4. **Debug mode**
   ```bash
   # Enable debug logging
   gz --debug <command>

   # Set log level
   gz --log-level debug <command>
   ```

## Related Documentation

- [Getting Started Guide](../01-getting-started/)
- [Configuration Guide](../04-configuration/)
- [Examples](../../examples/)
- [Architecture Overview](../02-architecture/overview.md)
