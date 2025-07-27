# Synclone Specification Updates Needed

<!-- **Note**: The `synclone.md` file is marked as AI_MODIFY_PROHIBITED, so these updates need to be applied manually. -->

## Required Updates to synclone.md

### 1. Add Missing Command Documentation

The following commands are implemented but not documented in the specification. They should be added after the "Configuration Validation" section:

```markdown
### Configuration Management (`gz synclone config`)

**Purpose**: Manage synclone configuration files with advanced tools

**Features**:
- Generate configurations from existing repository structures
- Validate configuration syntax and schema compliance
- Convert between different configuration formats
- Create configurations from templates

**Subcommands**:

#### Configuration Generation (`gz synclone config generate`)

Generate configuration files with various strategies:

- `gz synclone config generate init` - Create initial configuration
- `gz synclone config generate template` - Generate from templates
- `gz synclone config generate discover` - Discover and generate from existing repos
- `gz synclone config generate github` - Generate GitHub-specific configuration

**Usage**:
```bash
# Initialize new configuration
gz synclone config generate init

# Generate from template
gz synclone config generate template --template enterprise

# Discover existing repositories and create config
gz synclone config generate discover --path ~/repos

# Generate GitHub organization config
gz synclone config generate github --org mycompany
```

#### Configuration Validation (`gz synclone config validate`)

**Purpose**: Validate configuration file syntax and structure

**Usage**:
```bash
# Validate configuration file
gz synclone config validate --config synclone.yaml

# Validate with schema checking
gz synclone config validate --strict --config synclone.yaml
```

#### Configuration Conversion (`gz synclone config convert`)

**Purpose**: Convert between configuration formats

**Usage**:
```bash
# Convert YAML to JSON
gz synclone config convert --from synclone.yaml --to synclone.json

# Convert to gzh.yml format
gz synclone config convert --from synclone.yaml --format gzh
```

### State Management (`gz synclone state`)

**Purpose**: Manage synclone operation state and history

**Features**:
- Track clone operations and their status
- Resume interrupted operations
- Clean up partial clones
- View operation history

**Subcommands**:

#### List States (`gz synclone state list`)

**Purpose**: List all tracked synclone operations

**Usage**:
```bash
# List all states
gz synclone state list

# List only active operations
gz synclone state list --active

# List failed operations
gz synclone state list --failed
```

#### Show State Details (`gz synclone state show`)

**Purpose**: Display detailed information about a specific operation

**Usage**:
```bash
# Show state by ID
gz synclone state show <state-id>

# Show last operation
gz synclone state show --last
```

#### Clean State (`gz synclone state clean`)

**Purpose**: Clean up state files and incomplete operations

**Usage**:
```bash
# Clean completed operations older than 7 days
gz synclone state clean --age 7d

# Clean all failed operations
gz synclone state clean --failed

# Clean specific operation
gz synclone state clean --id <state-id>
```
```

### 2. Update Commands Section Structure

The Commands section should be reorganized to include all command categories:

```markdown
## Commands

### Core Commands

- `gz synclone` - Main command for synchronized repository cloning
- `gz synclone github` - Clone repositories from GitHub organizations
- `gz synclone gitlab` - Clone repositories from GitLab groups
- `gz synclone gitea` - Clone repositories from Gitea organizations
- `gz synclone validate` - Validate configuration files

### Configuration Commands

- `gz synclone config` - Configuration file management
- `gz synclone config generate` - Generate configuration files
- `gz synclone config validate` - Validate configuration syntax
- `gz synclone config convert` - Convert between formats

### State Management Commands

- `gz synclone state` - Manage operation state
- `gz synclone state list` - List tracked operations
- `gz synclone state show` - Show operation details
- `gz synclone state clean` - Clean up state files
```

### 3. Add State Management Section

Add a new section after "Performance Optimization":

```markdown
## State Management and Recovery

### Operation Tracking

Synclone tracks all clone operations to enable:
- Resume capability for interrupted operations
- Operation history and audit trails
- Cleanup of failed or partial clones
- Performance metrics and statistics

### State Storage

Operation states are stored in:
- `~/.config/gzh-manager/synclone/state/` - User state directory
- Individual state files per operation with metadata
- Automatic cleanup of old state files

### Resume Capability

When using `--resume`, synclone:
1. Checks for incomplete operations
2. Identifies repositories that need retry
3. Continues from the last successful repository
4. Maintains the same configuration and options

### State File Format

State files contain:
- Operation ID and timestamp
- Configuration snapshot
- Repository list and status
- Success/failure metrics
- Error logs for failed repositories
```

### 4. Update Examples Section

Add examples for the new commands:

```markdown
### Configuration Management Examples

```bash
# Generate initial configuration
gz synclone config generate init --output myconfig.yaml

# Discover repositories and create configuration
gz synclone config generate discover --path ~/projects

# Validate configuration before use
gz synclone config validate --config synclone.yaml

# Convert configuration format
gz synclone config convert --from synclone.yaml --to synclone.json
```

### State Management Examples

```bash
# View operation history
gz synclone state list

# Check status of last operation
gz synclone state show --last

# Clean up old operations
gz synclone state clean --age 30d

# Resume interrupted operation
gz synclone --resume
```
```

## Summary of Changes

1. Add complete documentation for `gz synclone config` command and subcommands
2. Add complete documentation for `gz synclone state` command and subcommands
3. Reorganize Commands section to include all command categories
4. Add State Management and Recovery section
5. Include examples for all new commands
6. Update the command count in overview if mentioned
