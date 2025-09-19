# Command: gz ide scan

## Scenario: Scan system for installed IDEs

### Input

**Command**:

```bash
gz ide scan
```

**Prerequisites**:

- [ ] System access to application directories
- [ ] Read permissions for IDE installation paths

### Expected Output

**Success Case (IDEs Found)**:

```text
🔍 Scanning system for installed IDEs...

📂 JetBrains IDEs
✅ IntelliJ IDEA Ultimate 2023.3.2
   📍 /Applications/IntelliJ IDEA.app
   🔧 Config: ~/.config/JetBrains/IntelliJIdea2023.3

✅ WebStorm 2023.3.1
   📍 /Applications/WebStorm.app
   🔧 Config: ~/.config/JetBrains/WebStorm2023.3

📂 Visual Studio Code Family
✅ Visual Studio Code 1.85.1
   📍 /Applications/Visual Studio Code.app
   🔧 Config: ~/Library/Application Support/Code

✅ Cursor 0.21.2
   📍 /Applications/Cursor.app
   🔧 Config: ~/Library/Application Support/Cursor

📂 Text Editors
✅ Sublime Text 4 (Build 4169)
   📍 /Applications/Sublime Text.app

📋 Summary
   Total IDEs found: 5
   JetBrains: 2
   VS Code family: 2  
   Text editors: 1

stderr: (empty)
Exit Code: 0
```

**No IDEs Found**:

```text
🔍 Scanning system for installed IDEs...

❌ No IDEs detected on this system.

💡 Supported IDEs:
   - JetBrains: IntelliJ, WebStorm, PyCharm, GoLand, etc.
   - VS Code: Visual Studio Code, Cursor, VSCodium
   - Text Editors: Sublime Text, Atom, Vim/Neovim

🚫 Consider installing an IDE for better development experience.

stderr: (empty)  
Exit Code: 1
```

**Permission Error**:

```text
🔍 Scanning system for installed IDEs...

⚠️  Permission issues detected:
   ❌ Cannot access /Applications (permission denied)
   ❌ Cannot read ~/.config (permission denied)

💡 Solutions:
   - Run with appropriate permissions
   - Check file system access permissions
   - On macOS: Grant Full Disk Access in System Preferences

🔧 Partial scan completed with limited results.

stderr: permission issues detected
Exit Code: 1
```

### Side Effects

**Files Created**:

- `~/.gzh/ide-registry.json` - IDE detection cache
- `~/.gzh/ide-scan.log` - Detailed scan log

**Files Modified**: None
**State Changes**: IDE registry cache updated

### Validation

**Automated Tests**:

```bash
# Test IDE scan
result=$(gz ide scan 2>&1)
exit_code=$?

assert_contains "$result" "Scanning system for installed IDEs"
# Exit code varies: 0 (found), 1 (none found or issues)

# Check cache file creation
assert_file_exists "$HOME/.gzh/ide-registry.json"
registry_content=$(cat "$HOME/.gzh/ide-registry.json")
assert_contains "$registry_content" '"scan_timestamp":'
```

**Manual Verification**:

1. Run on system with known IDEs installed
1. Verify detected IDEs match actual installations
1. Check configuration paths are correct
1. Confirm cache file contains accurate information

### Edge Cases

**Multiple Versions**:

- Multiple versions of same IDE (e.g., IntelliJ 2023.2 and 2023.3)
- Beta/EAP versions alongside stable
- Different installation methods (App Store, direct download, package manager)

**Custom Installation Paths**:

- Non-standard installation directories
- Portable installations
- Network/shared installations

**Platform Differences**:

- macOS: Applications folder and Library configs
- Windows: Program Files and AppData configs
- Linux: Various package manager locations and ~/.config

**Corrupted Installations**:

- IDE binary present but config missing
- Incomplete installations
- Broken symlinks or aliases

### Performance Expectations

**Response Time**:

- Fast scan (common paths): < 3 seconds
- Deep scan (all paths): < 10 seconds
- Large systems: < 30 seconds with progress indication

**Resource Usage**:

- Memory: < 100MB
- Disk I/O: Read-only filesystem scanning
- CPU: Low impact scanning

## Notes

- Cross-platform IDE detection (macOS, Windows, Linux)
- Caches results for performance improvement
- Detects both mainstream and alternative IDEs
- Configuration path discovery for settings management
- Plugin and extension detection (future enhancement)
