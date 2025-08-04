# IDE Management Guide

The `gz ide` command provides comprehensive monitoring and management for JetBrains IDE settings, helping developers detect and fix synchronization issues in real-time.

## Overview

JetBrains IDEs store their configuration in platform-specific directories. When these settings become corrupted or out of sync, it can lead to lost preferences, plugin issues, and reduced productivity. The `gz ide` command monitors these settings in real-time and provides automated fixes.

## Supported IDEs

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

## Command Reference

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
```

### Fix Synchronization Issues

Automatically detect and fix common synchronization problems:

```bash
# Fix sync issues (interactive)
gz ide fix-sync

# Preview changes without applying
gz ide fix-sync --dry-run

# Fix specific product
gz ide fix-sync --product PyCharm2023.3

# Force fix without confirmation
gz ide fix-sync --force
```

### List Installed IDEs

Display all detected JetBrains installations:

```bash
# List all IDEs
gz ide list

# Output in JSON format
gz ide list --format json

# Include version details
gz ide list --detailed
```

## Configuration

### Configuration File

Create `~/.config/gzh-manager/ide.yaml`:

```yaml
ide:
  monitoring:
    enabled: true
    interval: 1s
    filter_temp_files: true

  products:
    - name: "IntelliJIdea"
      enabled: true
      custom_path: "/custom/path/to/config"

    - name: "PyCharm"
      enabled: true
      backup_before_fix: true

  sync:
    backup_enabled: true
    backup_retention: 7  # days
    auto_fix: false

  notifications:
    desktop: true
    sound: false
```

### Environment Variables

```bash
# Custom JetBrains configuration path
export JETBRAINS_CONFIG_PATH="/custom/jetbrains/config"

# Monitoring interval
export IDE_MONITOR_INTERVAL="2s"

# Enable auto-fix
export IDE_AUTO_FIX=true

# Disable specific products
export IDE_DISABLE_PRODUCTS="AndroidStudio,Rider"
```

## Common Use Cases

### Detecting Configuration Corruption

The monitor can detect various types of configuration issues:

1. **XML Parse Errors**: Malformed configuration files
2. **Permission Issues**: Incorrect file permissions
3. **Lock File Conflicts**: Multiple IDE instances
4. **Plugin Conflicts**: Incompatible plugin configurations
5. **Cache Corruption**: Invalid cache states

### Example Monitoring Session

```bash
$ gz ide monitor
🔍 Starting JetBrains IDE monitor...
📁 Monitoring 3 IDE configurations

[2025-08-04 10:15:23] ✅ IntelliJIdea2023.2: Configuration healthy
[2025-08-04 10:15:24] ⚠️  PyCharm2023.3: Detected XML parse error in keymap.xml
[2025-08-04 10:15:24] 🔧 PyCharm2023.3: Backup created at ~/.config/JetBrains/backups/
[2025-08-04 10:15:25] ✅ PyCharm2023.3: Configuration fixed
[2025-08-04 10:15:26] 📝 GoLand2023.1: Settings change detected (editor.xml)

Press Ctrl+C to stop monitoring...
```

### Fixing Sync Issues

When synchronization issues are detected:

```bash
$ gz ide fix-sync
🔍 Scanning for IDE synchronization issues...

Found issues in 2 products:

1. IntelliJIdea2023.2:
   - Corrupted: workspace.xml
   - Missing: ide.general.xml
   - Permission issue: options/

2. PyCharm2023.3:
   - Lock file exists (cleanup required)
   - Cache invalidation needed

? Select products to fix: [Space to select, Enter to confirm]
> [x] IntelliJIdea2023.2
  [x] PyCharm2023.3

🔧 Fixing selected products...
✅ All issues resolved!
```

## Advanced Features

### Custom Configuration Paths

For non-standard installations:

```bash
# Add custom IDE path
gz ide config add-path --product "IntelliJIdea" --path "/opt/idea/config"

# Monitor custom installation
gz ide monitor --custom-path "/portable/apps/pycharm/config"
```

### Backup Management

The tool automatically creates backups before making changes:

```bash
# List all backups
gz ide backup list

# Restore specific backup
gz ide backup restore --id backup-20250804-101523

# Clean old backups
gz ide backup clean --older-than 30d
```

### Integration with Other Commands

The IDE command integrates with other gz features:

```bash
# Monitor IDE while doing quality checks
gz ide monitor & gz quality run --watch

# Include IDE config in development environment backup
gz dev-env backup --include-ide
```

## Troubleshooting

### Common Issues

1. **"IDE not detected"**
   - Check if IDE is installed in standard location
   - Use `--custom-path` for non-standard installations
   - Verify IDE version is supported

2. **"Permission denied"**
   - Ensure proper permissions on config directory
   - Run with appropriate user (not root)
   - Check file ownership

3. **"Monitoring not detecting changes"**
   - Increase monitoring interval
   - Check if temp file filtering is too aggressive
   - Verify fsnotify is working on your system

### Debug Mode

For detailed debugging information:

```bash
# Enable debug logging
gz ide monitor --debug

# Test file system events
gz ide test-monitor --path ~/.config/JetBrains

# Validate configuration
gz ide validate-config
```

## Best Practices

1. **Regular Monitoring**: Run monitor during development sessions
2. **Backup Before Major Updates**: Use before IDE version upgrades
3. **Clean Cache Periodically**: Prevent accumulation of stale data
4. **Use Dry Run**: Always preview changes with `--dry-run`
5. **Version Control**: Consider versioning IDE settings

## Related Documentation

- [Development Environment Guide](development-environment/)
- [Configuration Guide](../04-configuration/configuration-guide.md)
- [Troubleshooting Guide](../06-development/debugging-guide.md)
