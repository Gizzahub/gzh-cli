# Command Reference

Complete reference documentation for all `gz` commands and their options.

## Overview

The `gz` CLI provides comprehensive commands for managing development environments, repositories, and code quality.

**📋 Complete Command Reference**: See [Complete Command Reference](../40-api-reference/40-command-reference.md) for comprehensive documentation of all commands, options, and examples.

## Quick Navigation

### Core Features
- **synclone** - Multi-platform repository synchronization
- **git** - Unified Git operations and platform management
- **quality** - Multi-language code quality management
- **ide** - JetBrains IDE monitoring and management
- **profile** - Performance profiling and analysis

### Environment Management
- **dev-env** - Development environment configuration
- **net-env** - Network environment transitions
- **pm** - Package manager updates and management

### Repository Management
- **repo-config** - GitHub repository configuration management

## Getting Help

For complete command documentation including all options, flags, examples, and troubleshooting information, see:

**📋 [Complete Command Reference](../40-api-reference/40-command-reference.md)**

### Quick Help

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

## Most Common Commands

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
gz profile stats
gz profile cpu --duration 30s

# Environment management
gz pm update --all
gz dev-env aws status
```

## Related Documentation

- **📋 [Complete Command Reference](../40-api-reference/40-command-reference.md)** - Full command documentation
- **📖 [Configuration Guide](../30-configuration/30-configuration-guide.md)** - Configuration system
- **🚀 [Getting Started Guide](../01-getting-started/)** - Installation and setup
- **🏗️ [Architecture Overview](../02-architecture/overview.md)** - System architecture
- **📁 [Examples](../../examples/)** - Configuration examples
