# GZ Command Structure Migration Guide

## Overview

This guide helps you migrate from the old gz command structure to the new consolidated structure.

## Command Changes

The following commands have been reorganized for better consistency:

| Old Command | New Command | Status |
|-------------|-------------|---------|
| `gz gen-config` | `gz synclone config generate` | Deprecated with warning |
| `gz repo-config` | `gz repo-sync config` | Deprecated with warning |
| `gz event` | `gz repo-sync event` | Deprecated with warning |
| `gz webhook` | `gz repo-sync webhook` | Deprecated with warning |
| `gz ssh-config` | `gz dev-env ssh` | Still works, planned deprecation |
| `gz always-latest` | `gz pm` | Deprecated with warning |

## Migration Methods

### Method 1: Automatic Migration Script

Run the migration helper script:

```bash
./scripts/migrate-gz.sh
```

This script will:
- Show you the command changes
- Search for old commands in your shell configuration files
- Create compatibility aliases (optional)

### Method 2: Manual Migration

1. **Update your scripts**: Replace old commands with new ones
2. **Update aliases**: If you have custom aliases, update them
3. **Update CI/CD**: Update any automation scripts

### Method 3: Use Compatibility Mode

The old commands still work but show deprecation warnings. This gives you time to migrate gradually.

## Backward Compatibility

### Using Aliases

Source the compatibility aliases file:

```bash
source ~/.config/gzh-manager/aliases.sh
```

This provides wrapper functions that redirect old commands to new ones.

### Deprecation Warnings

When you use an old command, you'll see:

```
Warning: 'gen-config' is deprecated. Use 'gz synclone config generate' instead.
```

## Examples

### Before
```bash
# Generate configuration from GitHub org
gz gen-config github myorg

# Manage repository configurations
gz repo-config apply --org myorg

# Setup webhook server
gz webhook create --org myorg

# Update package managers
gz always-latest brew
gz always-latest asdf --strategy major
```

### After
```bash
# Generate configuration from GitHub org
gz synclone config generate github myorg

# Manage repository configurations
gz repo-sync config apply --org myorg

# Setup webhook server
gz repo-sync webhook create --org myorg

# Update package managers
gz pm update --manager brew
gz pm update --manager asdf --strategy major
```

## Rollback

If you need to rollback:

```bash
./scripts/rollback-gz.sh
```

This removes the compatibility aliases. The deprecated commands will continue to work with warnings.

## FAQ

**Q: Will my old scripts break?**
A: No, the old commands still work but show deprecation warnings.

**Q: How long will the old commands be supported?**
A: They will be supported for at least 6 months after the new structure is released.

**Q: Can I use both old and new commands?**
A: Yes, but we recommend migrating to the new commands to avoid confusion.

## Getting Help

- Run `gz help` for general help
- Run `gz [command] --help` for command-specific help
- Check the deprecation warnings for migration hints