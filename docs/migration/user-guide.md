# GZH Manager v2.0 Migration Guide

## What's Changing?

GZH Manager v2.0 simplifies the command structure from 18 to 10 commands, making it easier to discover and use features.

## Quick Migration Reference

### Command Changes

| If you use... | Use this instead... |
|---------------|---------------------|
| `gz gen-config` | `gz synclone config generate` |
| `gz repo-config` | `gz repo-sync config` |
| `gz event` | `gz repo-sync event` |
| `gz webhook` | `gz repo-sync webhook` |
| `gz ssh-config` | `gz dev-env ssh` |
| `gz config` | `gz [command] config` |
| `gz doctor` | `gz validate --all` |
| `gz shell` | `gz --debug-shell` |

## Automatic Migration

### Option 1: One-Line Migration (Recommended)
```bash
curl -sSL https://gz.dev/migrate | bash
```

This will:
- ✅ Backup your current configuration
- ✅ Convert configuration files to new format
- ✅ Install command aliases for backward compatibility
- ✅ Update your shell configuration
- ✅ Validate the migration

### Option 2: Manual Migration Script
```bash
# Download migration script
wget https://github.com/gizzahub/gzh-manager-go/releases/latest/download/gz-migrate.sh

# Make it executable
chmod +x gz-migrate.sh

# Run migration
./gz-migrate.sh
```

### Option 3: Step-by-Step Manual Migration

#### Step 1: Update Your Scripts
Search and replace old commands in your scripts:

```bash
# Find all scripts using old commands
grep -r "gz gen-config\|gz repo-config\|gz event\|gz webhook\|gz ssh-config\|gz doctor" ~/scripts/

# Update each file
sed -i 's/gz gen-config/gz synclone config generate/g' your-script.sh
sed -i 's/gz repo-config/gz repo-sync config/g' your-script.sh
# ... repeat for other commands
```

#### Step 2: Update Configuration Files
The configuration structure has changed. Previously centralized configs are now command-specific:

**Old structure:**
```
~/.config/gzh-manager/
└── config.yaml  # All configuration in one file
```

**New structure:**
```
~/.config/gzh-manager/
├── synclone.yaml
├── dev-env.yaml
├── net-env.yaml
├── repo-sync.yaml
└── ...
```

Run the config splitter:
```bash
gz migrate config
```

#### Step 3: Install Backward Compatibility Aliases
Add aliases to your shell configuration:

```bash
# For bash/zsh
echo 'source ~/.config/gzh-manager/aliases.bash' >> ~/.bashrc

# For fish
echo 'source ~/.config/gzh-manager/aliases.fish' >> ~/.config/fish/config.fish
```

## Common Migration Scenarios

### Scenario 1: CI/CD Pipelines
If your CI/CD uses old commands, you have several options:

**Option A: Update the commands (Recommended)**
```yaml
# Old
- run: gz gen-config --output config.yaml
- run: gz doctor

# New
- run: gz synclone config generate --output config.yaml
- run: gz validate --all
```

**Option B: Suppress warnings temporarily**
```yaml
env:
  GZH_ALLOW_DEPRECATED: "1"
```

### Scenario 2: Shared Scripts
For scripts used by multiple team members:

1. Add version detection:
```bash
if gz version | grep -q "v2."; then
    # Use new commands
    gz synclone config generate
else
    # Use old commands
    gz gen-config
fi
```

2. Or use aliases during transition:
```bash
source ~/.config/gzh-manager/aliases.bash
# Old commands will work with warnings
```

### Scenario 3: Docker Images
Update your Dockerfiles:

```dockerfile
# Install aliases in container
RUN curl -sSL https://gz.dev/migrate | bash

# Or update commands directly
RUN sed -i 's/gz gen-config/gz synclone config generate/g' /scripts/*.sh
```

## Handling Specific Commands

### Configuration Management
**Old way:**
```bash
gz config get synclone.output
gz config set synclone.output ./repos
```

**New way:**
```bash
gz synclone config get output
gz synclone config set output ./repos
```

### SSH Configuration
**Old way:**
```bash
gz ssh-config generate
gz ssh-config update
```

**New way:**
```bash
gz dev-env ssh generate
gz dev-env ssh update
```

### System Validation
**Old way:**
```bash
gz doctor
gz doctor --fix
```

**New way:**
```bash
gz validate --all
gz validate --all --fix

# Or validate specific component
gz synclone validate
gz dev-env validate
```

## Troubleshooting

### Issue: "command not found" after update
**Solution:** Install backward compatibility aliases:
```bash
curl -sSL https://gz.dev/install-aliases | bash
source ~/.bashrc  # or ~/.zshrc
```

### Issue: Configuration not found
**Solution:** Run configuration migration:
```bash
gz migrate config
```

### Issue: Scripts failing in CI/CD
**Solution:** Set compatibility environment variable:
```bash
export GZH_ALLOW_DEPRECATED=1
```

### Issue: Warnings in production logs
**Solution:** Update to new commands or suppress warnings:
```bash
# Update commands (preferred)
sed -i 's/gz gen-config/gz synclone config generate/g' *.sh

# Or suppress warnings (temporary)
export GZH_SUPPRESS_DEPRECATION=1
```

## FAQ

**Q: How long will old commands continue to work?**
A: Old commands will work with deprecation warnings until v3.0 (approximately 6 months).

**Q: Can I rollback if something goes wrong?**
A: Yes, the migration script creates backups. Run `gz-migrate.sh --rollback`.

**Q: Will my existing configurations be lost?**
A: No, configurations are backed up and converted to the new format.

**Q: Do I need to migrate immediately?**
A: No, but we recommend migrating within 3 months to avoid issues when v3.0 is released.

**Q: Where can I get help?**
A: 
- GitHub Issues: https://github.com/gizzahub/gzh-manager-go/issues
- Documentation: https://gz.dev/docs/migration
- Community Discord: https://discord.gg/gizzahub

## Next Steps

1. **Run the migration**: Use the automatic migration script
2. **Test your workflows**: Ensure everything works with new commands
3. **Update documentation**: Update your team's documentation
4. **Remove aliases**: After confirming everything works, remove backward compatibility aliases

## Timeline

- **Now - 3 months**: Transition period with full backward compatibility
- **3 - 6 months**: Deprecation warnings become more prominent
- **6 months (v3.0)**: Old commands removed

Thank you for using GZH Manager! The new structure will make your workflow more efficient and intuitive.