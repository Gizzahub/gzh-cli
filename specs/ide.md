# IDE Management Specification (Updated)

## Overview

The `gz ide` command provides comprehensive IDE detection, monitoring, and management capabilities across multiple IDE families including JetBrains products, VS Code variants, and popular text editors.

## Purpose

Enhanced IDE management for:

- **Automatic IDE Detection**: Scan and detect all installed IDEs on the system
- **Status Monitoring**: Real-time monitoring of IDE processes and resource usage
- **Project Opening**: Open projects directly in detected IDEs
- **Settings Synchronization**: Monitor and fix JetBrains IDE settings sync issues
- **Cross-Platform Support**: Works on Linux, macOS, and Windows

## Command Structure

```
gz ide <subcommand> [options]
```

## Current Implementation (2025-08)

### Available Subcommands

| Subcommand | Purpose | Implementation Status |
| ---------- | ---------------------------------- | --------------------- |
| `scan` | Scan system for installed IDEs | ✅ Implemented (NEW) |
| `status` | Show IDE status and resource usage | ✅ Implemented (NEW) |
| `open` | Open project in detected IDE | ✅ Implemented (NEW) |
| `monitor` | Monitor JetBrains IDE settings | ✅ Implemented |
| `fix-sync` | Fix JetBrains sync issues | ✅ Implemented |
| `list` | List detected IDEs (legacy) | ✅ Implemented |

## Subcommand Specifications

### 1. IDE Scan (`gz ide scan`)

**Purpose**: Automatically detect all installed IDEs on the system.

**Description**: Scans common installation locations and package managers to find all installed IDEs. Results are cached for 24 hours to improve performance.

```bash
gz ide scan [--refresh] [--verbose]
```

**Options**:

- `--refresh` - Force refresh scan (ignore cache)
- `--verbose` - Show detailed scan information

**Detection Methods**:

1. **JetBrains Toolbox**: Parse Toolbox installations
1. **System Paths**: Check standard installation directories
1. **Package Managers**: Query brew, snap, flatpak
1. **AppImage**: Detect AppImage IDEs
1. **Custom Paths**: User-defined search paths

**Supported IDEs**:

**JetBrains Products**:

- IntelliJ IDEA (Community, Ultimate)
- PyCharm (Community, Professional)
- WebStorm
- PhpStorm
- RubyMine
- CLion
- GoLand
- DataGrip
- Rider
- Android Studio
- Fleet

**VS Code Family**:

- Visual Studio Code
- VS Code Insiders
- Cursor
- VSCodium

**Text Editors**:

- Sublime Text
- Vim
- Neovim
- Emacs

**Output Example**:

```
Found 8 IDEs:

JetBrains IDEs:
  ✓ IntelliJ IDEA Ultimate 2024.1 (/Applications/IntelliJ IDEA.app)
  ✓ GoLand 2024.1 (/Applications/GoLand.app)
  ✓ WebStorm 2024.1 (Toolbox: ~/.local/share/JetBrains/Toolbox/apps/WebStorm)

VS Code Family:
  ✓ Visual Studio Code 1.85.0 (/Applications/Visual Studio Code.app)
  ✓ Cursor 0.20.0 (/Applications/Cursor.app)

Text Editors:
  ✓ Sublime Text 4 (/Applications/Sublime Text.app)
  ✓ Neovim 0.9.5 (/opt/homebrew/bin/nvim)
  ✓ Vim 9.0 (/usr/bin/vim)
```

### 2. IDE Status (`gz ide status`)

**Purpose**: Display real-time status of IDEs including running processes and resource usage.

**Description**: Shows which IDEs are currently running, their resource consumption, and open projects.

```bash
gz ide status [--running] [--format <format>]
```

**Options**:

- `--running` - Show only running IDEs
- `--format` - Output format (table, json, yaml)
- `--watch` - Continuous monitoring mode

**Status Information**:

- IDE name and version
- Process status (running/stopped)
- Memory usage
- CPU usage
- Open projects
- Uptime

**Output Example**:

```
IDE Status Report:

┌─────────────────────┬─────────┬──────────┬──────────┬────────────────────┬──────────┐
│ IDE                 │ Status  │ Memory   │ CPU      │ Project            │ Uptime   │
├─────────────────────┼─────────┼──────────┼──────────┼────────────────────┼──────────┤
│ IntelliJ IDEA 2024.1│ Running │ 2.3 GB   │ 5.2%     │ ~/projects/myapp   │ 2h 15m   │
│ GoLand 2024.1       │ Running │ 1.8 GB   │ 3.1%     │ ~/go/src/mycli     │ 45m      │
│ VS Code             │ Running │ 512 MB   │ 2.0%     │ ~/docs             │ 3h 30m   │
│ WebStorm 2024.1     │ Stopped │ -        │ -        │ -                  │ -        │
│ Cursor              │ Stopped │ -        │ -        │ -                  │ -        │
└─────────────────────┴─────────┴──────────┴──────────┴────────────────────┴──────────┘

Total Running: 3 | Total Memory: 4.6 GB | Total CPU: 10.3%
```

### 3. Open Project (`gz ide open`)

**Purpose**: Open a project directory in a detected IDE.

**Description**: Intelligently opens projects in the appropriate IDE based on project type and user preferences.

```bash
gz ide open <project-path> [--ide <ide-name>]
```

**Options**:

- `--ide` - Specific IDE to use (optional)
- `--new-window` - Open in new window
- `--wait` - Wait for IDE to close

**Smart Detection**:

1. **Project Type Detection**:

   - Go projects → GoLand (if available) or VS Code
   - Python projects → PyCharm or VS Code
   - Java/Kotlin → IntelliJ IDEA
   - JavaScript/TypeScript → WebStorm or VS Code
   - Generic → User's default IDE

1. **Configuration Files**:

   - `.idea/` → JetBrains IDE
   - `.vscode/` → VS Code
   - `go.mod` → GoLand
   - `requirements.txt` → PyCharm
   - `package.json` → WebStorm

**Examples**:

```bash
# Open current directory with auto-detection
gz ide open .

# Open specific project with auto-detection
gz ide open ~/projects/myapp

# Open with specific IDE
gz ide open ~/projects/myapp --ide goland

# Open in new window
gz ide open ~/projects/another --new-window
```

### 4. Monitor Settings (`gz ide monitor`)

**Purpose**: Monitor JetBrains IDE configuration changes in real-time.

**Description**: Watches for configuration file changes and detects potential sync issues.

```bash
gz ide monitor [--product <product>] [--interval <duration>]
```

**Options**:

- `--product` - Specific JetBrains product to monitor
- `--interval` - Check interval (default: 1s)
- `--filter-temp` - Filter temporary file changes

**Monitored Items**:

- Configuration files (.xml)
- Plugin changes
- Color schemes
- Keymaps
- Code styles
- Live templates

### 5. Fix Sync Issues (`gz ide fix-sync`)

**Purpose**: Automatically resolve JetBrains settings synchronization conflicts.

```bash
gz ide fix-sync [--dry-run] [--backup]
```

**Options**:

- `--dry-run` - Preview changes without applying
- `--backup` - Create backup before fixing
- `--force` - Force resolution of conflicts

## Architecture Changes (2025-08)

### Enhanced Detection System

```go
type IDEDetector struct {
    cache        *IDECache      // 24-hour cache
    toolbox      *ToolboxParser // JetBrains Toolbox
    system       *SystemScanner // System paths
    packageMgr   *PackageQuery  // brew, snap, etc.
    appImage     *AppImageScan  // AppImage detection
}

type IDEInfo struct {
    Name            string
    Version         string
    Path            string
    InstallMethod   string // toolbox, system, brew, snap, appimage
    Family          string // jetbrains, vscode, editor
    Executable      string
    ConfigPath      string
    LastDetected    time.Time
}
```

### Caching Strategy

```yaml
ide:
  cache:
    enabled: true
    duration: 24h
    path: ~/.cache/gz/ide-scan.json

  scan:
    custom_paths:
      - ~/Applications
      - /opt/IDEs

    exclude_patterns:
      - "*.backup"
      - "*.old"
```

## Testing Coverage

- IDE package: 40.4% coverage (up from 33.5%)
- New test files:
  - `detector_test.go` - IDE detection tests
  - `status_test.go` - Status monitoring tests
  - `open_test.go` - Project opening tests

## Performance Improvements

### Scan Performance

- **Initial Scan**: ~2-5 seconds (depends on system)
- **Cached Results**: \<100ms
- **Parallel Detection**: Uses goroutines for concurrent scanning
- **Smart Caching**: Only rescan if system changes detected

### Resource Usage

- **Memory**: \<50MB during scan
- **CPU**: Minimal impact (background scanning)
- **Disk I/O**: Optimized with caching

## Platform-Specific Features

### macOS

- Scans `/Applications` and `~/Applications`
- Detects .app bundles
- JetBrains Toolbox in `~/Library/Application Support/JetBrains/Toolbox`

### Linux

- Scans `/usr/local/bin`, `/opt`, `~/.local/share`
- Snap and Flatpak integration
- AppImage detection in common directories
- JetBrains Toolbox in `~/.local/share/JetBrains/Toolbox`

### Windows

- Scans `Program Files`, `Program Files (x86)`
- Registry integration for installed programs
- JetBrains Toolbox in `%LOCALAPPDATA%\JetBrains\Toolbox`

## Configuration

### IDE Preferences

```yaml
ide:
  defaults:
    preferred_ide: goland        # Default IDE for 'open' command
    open_in_new_window: true    # Always open in new window

  project_associations:
    - pattern: "*.go"
      ide: goland
    - pattern: "*.py"
      ide: pycharm
    - pattern: "*.js,*.ts"
      ide: webstorm

  scan:
    cache_duration: 24h
    custom_paths:
      - /custom/ide/location

  monitoring:
    interval: 1s
    products:
      - IntelliJIdea
      - GoLand
      - WebStorm
```

## Future Enhancements

1. **Plugin Management**: Install/update IDE plugins
1. **Settings Sync**: Cross-IDE settings synchronization
1. **Project Templates**: Create projects with IDE-specific templates
1. **Remote Development**: Integration with remote development features
1. **AI Integration**: IDE-specific AI assistant configuration
1. **Performance Profiling**: IDE performance monitoring and optimization
1. **Extension Marketplace**: Custom extension management

## Migration Guide

### From Old Commands to New

**Old**:

```bash
gz ide list
gz ide monitor
```

**New**:

```bash
gz ide scan              # Replaces and enhances 'list'
gz ide status            # New: Shows running IDEs
gz ide open .            # New: Opens projects
gz ide monitor           # Still available, unchanged
```

## Troubleshooting

### Common Issues

1. **IDE Not Detected**:

   - Run `gz ide scan --refresh --verbose`
   - Check custom paths in configuration
   - Verify IDE installation path

1. **Cache Issues**:

   - Clear cache: `rm ~/.cache/gz/ide-scan.json`
   - Force refresh: `gz ide scan --refresh`

1. **Open Command Fails**:

   - Verify IDE is detected: `gz ide scan`
   - Check IDE executable permissions
   - Try specifying IDE explicitly: `--ide <name>`

## Documentation

- User Guide: `docs/30-features/35-ide-management.md`
- API Reference: `docs/50-api-reference/ide-commands.md`
- Configuration: `docs/40-configuration/ide-config.md`

## Security Considerations

1. **Path Validation**: Validate all file paths before execution
1. **Command Injection**: Sanitize all shell command arguments
1. **Configuration Access**: Respect file system permissions
1. **Cache Security**: Store cache in user-specific directory
1. **Process Monitoring**: Only show user's own processes
