# Configuration Migration Command

The `gz migrate` command provides a comprehensive solution for migrating configuration files from legacy formats to the unified `gzh.yaml` format.

## Overview

The migrate command can:

- Convert legacy `bulk-clone.yaml` files to unified `gzh.yaml` format
- Handle batch migration of multiple configuration files
- Create automatic backups before migration
- Preview migrations with dry-run mode
- Provide detailed migration reports and validation

## Usage

### Basic Migration

```bash
# Migrate a specific file
gz migrate ./bulk-clone.yaml ./gzh.yaml

# Migrate with dry-run to preview changes
gz migrate ./bulk-clone.yaml ./gzh.yaml --dry-run

# Migrate with backup (default behavior)
gz migrate ./bulk-clone.yaml ./gzh.yaml --backup

# Force migration even if target exists
gz migrate ./bulk-clone.yaml ./gzh.yaml --force
```

### Automatic Detection

```bash
# Auto-detect source and target files
gz migrate --auto

# This will:
# 1. Look for legacy files in standard locations
# 2. Generate appropriate target filenames
# 3. Perform migration with default settings
```

### Batch Migration

```bash
# Migrate all legacy files in current directory
gz migrate --batch

# This will:
# 1. Find all bulk-clone.yaml and bulk-clone.yml files
# 2. Convert each to corresponding gzh.yaml format
# 3. Provide summary statistics
```

### Advanced Options

```bash
# Verbose output with detailed migration information
gz migrate ./bulk-clone.yaml ./gzh.yaml --verbose

# Disable automatic backup creation
gz migrate ./bulk-clone.yaml ./gzh.yaml --no-backup

# Preview migration without making changes
gz migrate ./bulk-clone.yaml ./gzh.yaml --dry-run --verbose
```

## Command Options

| Flag        | Description                                   | Default |
| ----------- | --------------------------------------------- | ------- |
| `--dry-run` | Preview migration without making changes      | `false` |
| `--backup`  | Create backup before migration                | `true`  |
| `--force`   | Force migration even if target exists         | `false` |
| `--verbose` | Enable verbose output                         | `false` |
| `--format`  | Output format (yaml, json)                    | `yaml`  |
| `--batch`   | Migrate all legacy files in current directory | `false` |
| `--auto`    | Auto-detect source and target files           | `false` |

## Migration Process

### 1. **Source Detection**

The command automatically detects the format of the source file:

- Legacy `bulk-clone.yaml` format (version 0.1)
- Unified `gzh.yaml` format (version 1.0.0+)

### 2. **Validation**

Before migration, the command validates:

- Source file exists and is readable
- Source file is in legacy format
- Target file doesn't exist (unless `--force` is used)
- Source file has valid syntax

### 3. **Backup Creation**

If `--backup` is enabled (default):

- Creates timestamped backup: `bulk-clone.backup.20240101-120000.yaml`
- Preserves original file permissions
- Verifies backup integrity

### 4. **Migration**

The migration process:

- Converts legacy configuration structure to unified format
- Maintains all configuration values and settings
- Adds new unified format fields with defaults
- Preserves comments where possible

### 5. **Validation**

After migration:

- Validates target file syntax
- Checks unified format compliance
- Verifies all required fields are present
- Provides warnings for potential issues

## Migration Examples

### Example 1: Basic File Migration

**Legacy Configuration (`bulk-clone.yaml`):**

```yaml
version: "0.1"
default:
  protocol: "https"
  github:
    root_path: "/home/user/repos"
    org_name: "myorg"
repo_roots:
  - provider: "github"
    root_path: "/home/user/repos/github"
    org_name: "myorg"
    protocol: "https"
ignore_names:
  - "test-.*"
  - ".*-archive"
```

**Migration Command:**

```bash
gz migrate ./bulk-clone.yaml ./gzh.yaml --verbose
```

**Generated Unified Configuration (`gzh.yaml`):**

```yaml
version: "1.0.0"
default_provider: "github"
global:
  clone_base_dir: "/home/user/repos"
  default_strategy: "reset"
  default_visibility: "all"
providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "/home/user/repos/github"
        visibility: "all"
        strategy: "reset"
        exclude:
          - "test-.*"
          - ".*-archive"
```

### Example 2: Batch Migration

**Command:**

```bash
gz migrate --batch --verbose
```

**Output:**

```
🔄 Running batch migration in current directory
Found 3 legacy configuration files:
  - bulk-clone.yaml
  - project-bulk-clone.yml
  - backup-bulk-clone.yaml

🔄 Migrating: bulk-clone.yaml → gzh.yaml
✅ Migration successful

🔄 Migrating: project-bulk-clone.yml → project-gzh.yaml
✅ Migration successful

🔄 Migrating: backup-bulk-clone.yaml → backup-gzh.yaml
✅ Migration successful

📊 Batch migration completed:
  ✅ Successful: 3
  ❌ Failed: 0
```

### Example 3: Dry-Run Migration

**Command:**

```bash
gz migrate ./bulk-clone.yaml ./gzh.yaml --dry-run --verbose
```

**Output:**

```
🔄 Migrating configuration: ./bulk-clone.yaml → ./gzh.yaml
🧪 Dry-run mode: previewing migration

📊 Migration Results:
  📁 Source: ./bulk-clone.yaml
  📁 Target: ./gzh.yaml
  ✅ Success: true
  ⚠️  Warnings:
    - Token not configured, using environment variable placeholder
    - Default visibility set to 'all' - consider restricting for security
  🔧 Required Actions:
    - Set GITHUB_TOKEN environment variable
    - Review and update token configuration
  📊 Migration Statistics:
    - Migrated targets: 1
    - Migration report:
      Converted 1 GitHub organization
      Added unified format metadata
      Applied default security settings
```

## Error Handling

### Common Errors

1. **Source file not found**

   ```
   Error: source file does not exist: ./bulk-clone.yaml
   ```

2. **Target file already exists**

   ```
   Error: target file already exists: ./gzh.yaml (use --force to overwrite)
   ```

3. **Invalid source format**

   ```
   Error: source file is not in legacy format: ./gzh.yaml
   ```

4. **Permission errors**
   ```
   Error: failed to create backup: permission denied
   ```

### Troubleshooting

1. **Check file permissions**:

   ```bash
   ls -la bulk-clone.yaml
   ```

2. **Validate source file syntax**:

   ```bash
   gz config validate ./bulk-clone.yaml
   ```

3. **Use verbose mode for detailed information**:

   ```bash
   gz migrate ./bulk-clone.yaml ./gzh.yaml --verbose
   ```

4. **Test with dry-run first**:
   ```bash
   gz migrate ./bulk-clone.yaml ./gzh.yaml --dry-run
   ```

## Integration with Other Commands

The migrate command works seamlessly with other gzh-manager commands:

### After Migration

```bash
# Validate the migrated configuration
gz config validate ./gzh.yaml

# Show effective configuration
gz config show --config ./gzh.yaml

# Use the migrated configuration
gz bulk-clone --config ./gzh.yaml --use-gzh-config
```

### Configuration Priority

The migrated configuration follows the same priority rules:

1. Command-line flags (highest)
2. Environment variables
3. Configuration files
4. Default values (lowest)

## Best Practices

1. **Always test with dry-run first**:

   ```bash
   gz migrate ./bulk-clone.yaml ./gzh.yaml --dry-run
   ```

2. **Keep backups enabled**:

   ```bash
   gz migrate ./bulk-clone.yaml ./gzh.yaml --backup
   ```

3. **Use batch migration for multiple files**:

   ```bash
   gz migrate --batch
   ```

4. **Validate after migration**:

   ```bash
   gz config validate ./gzh.yaml
   ```

5. **Review migration warnings**:
   ```bash
   gz migrate ./bulk-clone.yaml ./gzh.yaml --verbose
   ```

## Configuration File Search

The migration command searches for configuration files in the same order as other commands:

1. **Explicit paths**: Files specified in command arguments
2. **Current directory**: `./bulk-clone.yaml`, `./bulk-clone.yml`
3. **User config**: `~/.config/gzh-manager/bulk-clone.yaml`
4. **System config**: `/etc/gzh-manager/bulk-clone.yaml`

## See Also

- [Configuration System](configuration.md)
- [Configuration Priority](configuration-priority.md)
- [Bulk Clone Command](bulk-clone-command.md)
- [Configuration Validation](configuration-validation.md)
