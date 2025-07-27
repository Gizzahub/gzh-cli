# Package Manager Specification Updates Needed

<!-- **Note**: The `package-manager.md` file is marked as AI_MODIFY_PROHIBITED, so these updates need to be applied manually. -->

## Required Updates to package-manager.md

### 1. Add Legacy Command Documentation

The following commands are implemented for backward compatibility but not documented in the specification. They should be added to the "Package Manager Specific Commands" section:

```markdown
### Package Manager Specific Commands

- `gz pm brew` - Manage Homebrew packages (legacy compatibility)
- `gz pm asdf` - Manage asdf plugins and versions (legacy compatibility)
- `gz pm sdkman` - Manage SDKMAN packages (legacy compatibility)
- `gz pm apt` - Manage APT packages (legacy compatibility)
- `gz pm port` - Manage MacPorts packages (legacy compatibility)
- `gz pm rbenv` - Manage rbenv versions (legacy compatibility)
```

### 2. Remove or Clarify Unimplemented Commands

The specification currently lists these commands that are NOT implemented:

- `gz pm pip` - This command doesn't exist in the implementation
- `gz pm npm` - This command doesn't exist in the implementation
- `gz pm [manager]` - This generic pattern is not implemented

**Recommendation**: Either:
1. Remove these from the specification, OR
2. Add a note that these are planned for future implementation

### 3. Add Implementation Notes Section

Add a new section explaining the legacy commands:

```markdown
## Implementation Notes

### Legacy Commands

For backward compatibility, the following package manager-specific commands are available:
- `gz pm brew` - Direct Homebrew management
- `gz pm asdf` - Direct asdf management
- `gz pm sdkman` - Direct SDKMAN management
- `gz pm apt` - Direct APT management
- `gz pm port` - Direct MacPorts management
- `gz pm rbenv` - Direct rbenv management

These commands provide direct access to specific package managers for users who prefer manager-specific workflows. However, the recommended approach is to use the unified commands (`install`, `update`, `sync`, etc.) with the `--manager` flag for consistency.

### Configuration-Driven Approach

While legacy commands exist, the primary design philosophy is configuration-driven management through YAML files in `~/.gzh/pm/`. The unified commands operate on these configurations to provide a consistent interface across all package managers.
```

### 4. Update Command Examples

The examples should clarify which commands actually exist:

```markdown
### Command Options

```bash
# Install packages from all configured managers
gz pm install

# Install from specific package manager
gz pm install --manager brew

# Legacy: Direct package manager access
gz pm brew install wget
gz pm asdf plugin add nodejs

# Export current installations
gz pm export --all
gz pm export --manager brew

# Update packages
gz pm update --all
gz pm update --manager asdf --strategy latest
```
```

### 5. Clarify Package Manager Coverage

Update the overview to be more accurate about implementation status:

```markdown
## Overview

The unified package manager feature provides centralized management for multiple package managers through configuration files. It enables developers to maintain consistent development environments across machines.

### Implementation Status

**Fully Integrated** (configuration-based management):
- System package managers: brew, apt, port
- Version managers: asdf, sdkman, rbenv
- All managers support unified commands: install, update, sync, export

**Legacy Direct Access** (backward compatibility):
- `gz pm brew` - Direct Homebrew commands
- `gz pm asdf` - Direct asdf commands
- `gz pm sdkman` - Direct SDKMAN commands
- `gz pm apt` - Direct APT commands
- `gz pm port` - Direct MacPorts commands
- `gz pm rbenv` - Direct rbenv commands

**Planned/Configuration-Only**:
- pip, npm, gem, cargo, go, composer - Managed through configuration files only
- No direct `gz pm pip` or `gz pm npm` commands implemented
```

## Summary of Changes

1. Document the 6 legacy commands that exist but aren't in the spec
2. Remove or mark as "planned" the pip/npm commands that don't exist
3. Clarify that the generic `[manager]` pattern isn't implemented
4. Add implementation notes explaining the dual approach (unified vs legacy)
5. Update examples to show actual available commands
