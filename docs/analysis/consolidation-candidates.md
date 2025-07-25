# Command Consolidation Candidates

## Summary
Consolidate from 18 commands to 10 commands, improving user experience while maintaining all functionality.

## Consolidation Plan

### 1. Repository Management Consolidation
**Target Command**: `repo-sync`

**Commands to Merge**:
- `repo-config` → `repo-sync config`
- `webhook` → `repo-sync webhook`
- `event` → `repo-sync event`

**Rationale**:
- All three commands interact with GitHub API
- Webhook and event handling are closely related
- Repository configuration is part of synchronization workflow

### 2. Configuration Generation Consolidation
**Target Command**: `synclone`

**Commands to Merge**:
- `gen-config` → `synclone config generate`

**Rationale**:
- gen-config generates configuration for bulk-clone operations
- Logically belongs with the synclone command that uses these configs
- Reduces confusion between generic `config` and `gen-config`

### 3. SSH Configuration Consolidation
**Target Command**: `dev-env`

**Commands to Merge**:
- `ssh-config` → `dev-env ssh`

**Rationale**:
- SSH configuration is part of development environment setup
- dev-env already manages various environment configurations
- Natural fit for development workflow

### 4. Configuration Command Distribution
**Commands to Distribute**:
- `config` → Each command gets its own `config` subcommand

**Implementation**:
- `synclone config`
- `dev-env config`
- `net-env config`
- `repo-sync config`
- etc.

**Rationale**:
- Generic config command creates ambiguity
- Command-specific configuration is clearer
- Follows principle of locality

### 5. Doctor Command Distribution
**Commands to Distribute**:
- `doctor` → Each command gets `validate` subcommand

**Implementation**:
- `synclone validate`
- `dev-env validate`
- `net-env validate`
- Global `validate --all` for system-wide checks

**Rationale**:
- Command-specific validation is more focused
- Reduces top-level command clutter
- Allows targeted troubleshooting

### 6. Shell Command Conversion
**Target**: Hidden debug feature

**Implementation**:
- `shell` → `--debug-shell` flag or `GZH_DEBUG_SHELL=1`

**Rationale**:
- Rarely used in production
- Developer-only feature
- Reduces visible command count

### 7. Migrate Command Handling
**Target**: Standalone script or built-in migration

**Options**:
1. Move to `docs/migration/migrate.sh`
2. Keep as hidden command
3. Auto-detect and migrate on first run

**Rationale**:
- One-time operation for most users
- Not a core functionality

## Final Command Structure

### Before (18 commands):
```
always-latest, config, dev-env, docker, doctor, event, gen-config, 
ide, migrate, net-env, repo-config, repo-sync, shell, ssh-config, 
synclone, version, webhook, help
```

### After (10 commands):
```
Core Commands:
  synclone      # Includes gen-config functionality
  dev-env       # Includes ssh-config
  net-env       
  repo-sync     # Includes webhook, event, repo-config

Tool Commands:
  ide
  always-latest
  docker

Utility Commands:
  validate      # Replaces doctor
  version
  help
```

## Benefits

1. **Improved Discoverability**: Related functions grouped together
2. **Reduced Complexity**: 44% fewer top-level commands
3. **Logical Organization**: Commands grouped by workflow
4. **Maintained Functionality**: All features preserved
5. **Better UX**: Clearer command hierarchy

## Migration Impact

### High Impact (Breaking Changes):
- gen-config users
- repo-config, webhook, event users
- ssh-config users
- config command users
- doctor users

### Low Impact:
- shell users (developers only)
- migrate users (one-time use)

### No Impact:
- synclone, dev-env, net-env, ide, always-latest, docker users

## Implementation Priority

1. **Phase 1**: Merge simple commands
   - gen-config → synclone
   - ssh-config → dev-env

2. **Phase 2**: Consolidate GitHub commands
   - repo-config, webhook, event → repo-sync

3. **Phase 3**: Distribute generic commands
   - config → individual commands
   - doctor → validate subcommands

4. **Phase 4**: Clean up
   - Hide/remove shell
   - Handle migrate