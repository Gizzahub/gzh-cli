# Migration Scripts

This directory contains scripts for migrating from legacy command structures and maintaining backward compatibility.

## 🔄 Migration Scripts

### `migrate-gz.sh`
**Purpose**: Helper script for migrating to new gz command structure

**Usage**:
```bash
./scripts/migration/migrate-gz.sh
```

**Features**:
- Shows command mapping between old and new formats
- Searches for old commands in common configuration files
- Creates basic compatibility aliases

---

### `rollback-gz.sh`
**Purpose**: Rollback helper for gz command migration

**Usage**:
```bash
./scripts/migration/rollback-gz.sh
```

**What it does**:
- Removes compatibility aliases file
- Provides rollback instructions

---

## 🔗 Backward Compatibility Scripts

### `deprecated-aliases.sh` & `deprecated-aliases.fish`
**Purpose**: Provides backward compatibility aliases with deprecation warnings

**Features**:
- Shows deprecation warnings for old commands
- Provides migration guidance
- Maintains functionality during transition period

**Shell Support**:
- `deprecated-aliases.sh` - Bash/Zsh compatibility
- `deprecated-aliases.fish` - Fish shell compatibility

---

### `install-aliases.sh` & `uninstall-aliases.sh`
**Purpose**: Install/remove backward compatibility aliases system-wide

**Install Usage**:
```bash
./scripts/migration/install-aliases.sh
```

**Uninstall Usage**:
```bash
./scripts/migration/uninstall-aliases.sh
```

**Features**:
- Auto-detects shell type (bash, zsh, fish)
- Adds source lines to appropriate shell configuration files
- Creates deprecation schedule documentation

---

## ⚠️ Deprecation Schedule

These scripts are part of a planned deprecation cycle:

- **v2.0.0**: Deprecated commands show warnings
- **v3.0.0**: Commands will be removed (estimated: 2025-01-01)

## 🎯 Usage Recommendations

1. **For Immediate Migration**: Use `migrate-gz.sh` to identify and update old commands
2. **For Gradual Transition**: Use `install-aliases.sh` to maintain compatibility during migration
3. **For Complete Removal**: Use `uninstall-aliases.sh` when migration is complete

---

**Note**: These are temporary migration tools and will be removed in future versions once the transition period is complete.