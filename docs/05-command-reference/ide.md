# ide Command Reference

JetBrains IDE monitoring and management for settings synchronization and configuration issues.

## Synopsis

```bash
gz ide <action> [flags]
gz ide <action> --config <config-file>
```

## Description

The `ide` command monitors JetBrains IDE configurations, detects synchronization issues, and provides automated fixes for common problems.

## Supported IDEs

- IntelliJ IDEA
- GoLand
- PyCharm
- WebStorm
- PhpStorm
- RubyMine
- CLion
- Rider
- DataGrip
- Android Studio
- Fleet

## Actions

### `gz ide monitor`

Start real-time monitoring of IDE configurations.

```bash
gz ide monitor [flags]
```

**Flags:**
- `--interval` - Check interval (default: 1s)
- `--ides` - Specific IDEs to monitor (comma-separated)
- `--auto-fix` - Automatically fix detected issues (default: false)
- `--notify` - Enable desktop notifications (default: true)
- `--daemon` - Run as background daemon

**Examples:**
```bash
# Monitor all IDEs
gz ide monitor

# Monitor specific IDEs
gz ide monitor --ides goland,intellij

# Monitor with auto-fix
gz ide monitor --auto-fix

# Run as daemon
gz ide monitor --daemon
```

### `gz ide check`

Check current IDE configuration status.

```bash
gz ide check [flags]
```

**Flags:**
- `--ides` - Specific IDEs to check
- `--output` - Output format: table, json, yaml
- `--detailed` - Show detailed configuration info

**Examples:**
```bash
# Check all IDEs
gz ide check

# Check specific IDE
gz ide check --ides goland

# Detailed output
gz ide check --detailed --output json
```

### `gz ide fix-sync`

Fix IDE synchronization issues.

```bash
gz ide fix-sync [flags]
```

**Flags:**
- `--ides` - Specific IDEs to fix
- `--strategy` - Fix strategy: restart-sync, clear-cache, reset-config
- `--backup` - Create backup before fixing (default: true)

**Examples:**
```bash
# Fix sync issues for all IDEs
gz ide fix-sync

# Fix specific IDE
gz ide fix-sync --ides goland

# Clear cache strategy
gz ide fix-sync --strategy clear-cache
```

### `gz ide backup`

Backup IDE configurations.

```bash
gz ide backup [flags]
```

**Flags:**
- `--ides` - IDEs to backup
- `--output-dir` - Backup directory (default: ~/.config/gzh-manager/ide-backups)
- `--compress` - Compress backups (default: true)

### `gz ide restore`

Restore IDE configurations from backup.

```bash
gz ide restore --backup <backup-path> [flags]
```

**Flags:**
- `--backup` - Backup file path (required)
- `--ides` - IDEs to restore
- `--confirm` - Skip confirmation prompt

### `gz ide list`

List detected IDEs and their status.

```bash
gz ide list [flags]
```

**Flags:**
- `--output` - Output format: table, json, yaml
- `--status` - Filter by status: running, stopped, syncing, error

## Configuration

```yaml
version: "1.0"

monitoring:
  enabled: true
  interval: "1s"
  auto_fix: false
  notify: true

ides:
  goland:
    enabled: true
    sync:
      enabled: true
      provider: "jetbrains_account"
  intellij:
    enabled: true

backup:
  enabled: true
  interval: "30m"
  retention: "7d"
```

## Examples

### Basic Monitoring

```bash
# Start monitoring
gz ide monitor

# Check status
gz ide check

# Fix any issues
gz ide fix-sync
```

### Automated Setup

```bash
# Monitor with auto-fix enabled
gz ide monitor --auto-fix --daemon

# Create initial backup
gz ide backup
```

## Related Commands

- [`gz quality`](quality.md) - Code quality management

## See Also

- [IDE Management Guide](../03-core-features/ide-management.md)
- [IDE Configuration Schema](../04-configuration/schemas/ide-schema.yaml)
- [IDE Examples](../../examples/ide/)
