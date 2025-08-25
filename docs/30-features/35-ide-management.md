# 💻 IDE Management Guide

The `gz ide` command provides comprehensive monitoring and management for JetBrains IDE settings, helping developers detect and fix synchronization issues in real-time.

## 📋 Table of Contents

- [Overview](#overview)
- [Supported IDEs](#supported-ides)
- [Command Reference](#command-reference)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)

## 🎯 Overview

JetBrains IDEs store their configuration in platform-specific directories. When these settings become corrupted or out of sync, it can lead to lost preferences, plugin issues, and reduced productivity. The `gz ide` command monitors these settings in real-time and provides automated fixes.

### Key Features

- **Real-time monitoring** of IDE configuration changes
- **Automatic detection** of synchronization issues
- **Cross-platform support** for all major operating systems
- **Multiple IDE support** with unified management
- **Backup and restore** functionality for IDE settings

## 🔧 Supported IDEs

### Full Support

- **IntelliJ IDEA** (Community & Ultimate)
- **PyCharm** (Community & Professional)
- **WebStorm**
- **PhpStorm**
- **RubyMine**
- **CLion**
- **GoLand**
- **DataGrip**
- **Android Studio**
- **Rider**

### Platform Support

- **Linux**: Full support with automatic path detection
- **macOS**: Full support including Library paths
- **Windows**: Full support for both standard and Store installations

## 📖 Command Reference

### Monitor IDE Settings

Real-time monitoring of IDE configuration changes:

```bash
# Monitor all JetBrains IDEs
gz ide monitor

# Monitor specific IDE
gz ide monitor --product IntelliJIdea2023.2

# Monitor with custom interval
gz ide monitor --interval 500ms

# Monitor in verbose mode
gz ide monitor --verbose

# Background monitoring
gz ide monitor --daemon &
```

### Fix Synchronization Issues

Detect and fix common IDE synchronization problems:

```bash
# Check for sync issues
gz ide check-sync

# Fix detected issues automatically
gz ide fix-sync

# Fix specific IDE
gz ide fix-sync --product GoLand2023.2

# Dry run to see what would be fixed
gz ide fix-sync --dry-run
```

### IDE Status and Information

```bash
# List all detected IDEs
gz ide list

# Show detailed IDE information
gz ide status

# Check specific IDE status
gz ide status --product IntelliJIdea2023.2

# Export IDE configuration
gz ide export --output ide-config.json
```

### Settings Backup and Restore

```bash
# Backup IDE settings
gz ide backup --output ./ide-backup

# Restore from backup
gz ide restore --input ./ide-backup

# Backup specific IDE
gz ide backup --product PyCharm2023.2 --output ./pycharm-backup

# List available backups
gz ide backup list
```

## ⚙️ Configuration

### Basic Configuration

Add IDE management settings to your `~/.config/gzh-manager/gzh.yaml`:

```yaml
commands:
  ide:
    # Enable automatic monitoring
    auto_monitor: true

    # Monitoring interval
    monitor_interval: "1s"

    # Enable automatic fixes
    auto_fix: false

    # Backup settings
    backup:
      enabled: true
      location: "$HOME/.config/gzh-manager/ide-backups"
      retention_days: 30

    # IDE-specific settings
    products:
      IntelliJIdea:
        priority: high
        auto_fix: true
        custom_config_path: "/custom/path"

      GoLand:
        priority: high
        backup_frequency: "daily"
```

### Advanced Configuration

```yaml
commands:
  ide:
    # Detection settings
    detection:
      scan_paths:
        - "$HOME/.config/JetBrains"
        - "$HOME/Library/Application Support/JetBrains"
        - "$APPDATA/JetBrains"

      # Custom IDE detection
      custom_ides:
        - name: "CustomIDE"
          config_path: "/path/to/custom/ide"
          product_info: "/path/to/product-info.json"

    # Monitoring settings
    monitoring:
      events:
        - config_change
        - plugin_install
        - theme_change
        - keymap_change

      # File watch patterns
      watch_patterns:
        - "*.xml"
        - "*.json"
        - "options/*.xml"

    # Fix strategies
    fixes:
      config_corruption:
        enabled: true
        backup_before_fix: true

      plugin_conflicts:
        enabled: true
        safe_mode: true

      keymap_issues:
        enabled: false  # Require manual approval
```

## 🔍 Monitoring Features

### Real-time Change Detection

The IDE monitor tracks changes to:

- **Configuration files** (settings, preferences)
- **Plugin installations** and updates
- **Theme and appearance** changes
- **Keymap modifications**
- **Project-specific settings**

### Notification System

```bash
# Enable desktop notifications
gz ide monitor --notify

# Custom notification commands
gz ide monitor --on-change "echo 'IDE config changed'"

# Integration with external tools
gz ide monitor --webhook "http://localhost:8080/ide-events"
```

### Change History

```bash
# View recent changes
gz ide history

# Show changes for specific timeframe
gz ide history --since "2 hours ago"

# Export change history
gz ide history --output json > ide-changes.json
```

## 🛠️ Troubleshooting

### Common Issues

#### IDE Not Detected

```bash
# Force rescan for IDEs
gz ide scan --force

# Add custom IDE path
gz ide add-path "/custom/jetbrains/path"

# Check detection logs
gz ide scan --verbose
```

#### Settings Corruption

```bash
# Detect corruption
gz ide check-corruption

# Restore from backup
gz ide restore --latest

# Reset to defaults (careful!)
gz ide reset --product IntelliJIdea2023.2 --confirm
```

#### Monitoring Issues

```bash
# Check monitoring status
gz ide monitor status

# Restart monitoring
gz ide monitor restart

# Check file permissions
gz ide check-permissions
```

### Diagnostic Information

```bash
# Complete diagnostic report
gz ide diagnose

# Platform-specific checks
gz ide diagnose --platform

# Export diagnostics for support
gz ide diagnose --output diagnostics.json
```

## 🚀 Advanced Usage

### Integration with CI/CD

```bash
# Check IDE configuration in CI
gz ide validate --config-only

# Export configuration for team sharing
gz ide export --template team-config.yaml

# Validate team configuration
gz ide validate --template team-config.yaml
```

### Scripting and Automation

```bash
# Monitor with custom actions
gz ide monitor --on-error "./fix-ide-script.sh"

# Batch operations
gz ide batch-fix --all-products

# Scheduled backups
gz ide backup --schedule "0 2 * * *"  # Daily at 2 AM
```

### Multi-Developer Environments

```yaml
# Team-wide IDE settings
commands:
  ide:
    team_sync:
      enabled: true
      shared_config: "team://ide-config"

      # Sync specific settings
      sync_items:
        - code_style
        - live_templates
        - file_templates
        - inspection_profiles

      # Exclude personal settings
      exclude_items:
        - recent_projects
        - statistics
        - evaluation
```

## 📊 Monitoring Dashboard

### Status Overview

```bash
# Dashboard view
gz ide dashboard

# Web interface (if available)
gz ide server --port 8080
```

### Metrics and Analytics

```bash
# Usage statistics
gz ide stats

# Performance metrics
gz ide metrics --export

# Health score
gz ide health-score
```

## 🔗 Integration Examples

### With Development Workflow

```bash
# Pre-commit hook
gz ide check-sync --fail-on-issues

# Project setup script
gz ide setup-project --template golang

# Environment validation
gz ide validate --required-plugins "Go,Git"
```

### With Other gzh-cli Commands

```bash
# Combine with repository management
gz git repo clone-or-update repo.git && gz ide setup-project

# Quality checks
gz quality run && gz ide check-sync

# Development environment setup
gz dev-env setup && gz ide monitor --daemon
```

## 📋 Output Formats

All IDE commands support multiple output formats:

```bash
# JSON output for scripting
gz ide status --output json

# YAML for configuration
gz ide export --output yaml

# Table format (default)
gz ide list --output table

# CSV for analysis
gz ide history --output csv
```

## 🆘 Getting Help

### Command Help

```bash
# General help
gz ide --help

# Specific command help
gz ide monitor --help

# Examples and tutorials
gz ide examples
```

### Support Resources

- **Debug mode**: Add `--debug` to any command
- **Verbose logging**: Use `--verbose` for detailed output
- **Log files**: Check `~/.config/gzh-manager/logs/ide.log`

______________________________________________________________________

**Supported Platforms**: Linux, macOS, Windows
**JetBrains Products**: All major IDEs supported
**Monitoring**: Real-time with configurable intervals
**Integration**: Full gzh-cli ecosystem support
