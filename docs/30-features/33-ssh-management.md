# SSH Configuration Management

## Overview

The `gz dev-env ssh` command provides comprehensive SSH configuration management capabilities with advanced features for development environment setup and key deployment. It automatically handles SSH config parsing, key management, and remote key installation using modern Go-based implementations.

## Key Features

### 🔍 Intelligent SSH Configuration Parsing

- **Automatic Include Parsing**: Recursively parses and backs up all files referenced by `Include` directives
- **Glob Pattern Support**: Handles complex include patterns like `config.d/*` and `~/.ssh/configs/*.conf`
- **IdentityFile Detection**: Automatically detects and backs up private keys referenced by `IdentityFile` directives
- **Public Key Pairing**: Automatically includes corresponding `.pub` files for each private key
- **Directory Structure Preservation**: Maintains relative paths and directory relationships

### 🔐 Advanced Key Management

- **Comprehensive Backup**: Saves SSH config files, include files, private keys, and public keys
- **Permission Preservation**: Maintains proper file permissions (600 for private keys, 644 for public keys)
- **Metadata Tracking**: Stores timestamps, descriptions, and file listings in JSON metadata
- **Multiple Configurations**: Manage different SSH setups for different environments

### 🚀 Modern Key Installation

- **SFTP-Based Installation**: Pure Go implementation using SFTP for remote key installation
- **Shell Independence**: No dependency on remote shell commands or grep/echo utilities
- **Robust Error Handling**: Detailed error reporting with verbose logging support
- **Duplicate Prevention**: Automatic deduplication of SSH keys
- **Permission Management**: Automatically sets proper permissions on remote directories and files

## Architecture

### SSH Configuration Parser

```go
// Core parsing logic
type SSHConfigParser struct {
    BaseDir      string
    ExpandHome   bool
}

type ParsedSSHConfig struct {
    MainConfig      string              // ~/.ssh/config content
    IncludeFiles    map[string]string   // Include files content
    PrivateKeys     []string            // Private key paths
    PublicKeys      []string            // Public key paths
}
```

### SFTP-Based Key Installation

The key installation system uses Go's SSH and SFTP libraries for reliable remote operations:

1. **SSH Connection**: Establishes secure connection with multiple authentication methods
2. **SFTP Session**: Creates SFTP client for file operations
3. **File Processing**: Reads/writes `authorized_keys` using native Go file operations
4. **Permission Setting**: Uses SFTP chmod for proper file permissions

## Commands Reference

### Save SSH Configuration

```bash
gz dev-env ssh save --name <config-name> [options]
```

**Options:**
- `--name`: Configuration name (required)
- `--description`: Optional description for the configuration
- `--include-keys`: Include private keys in backup (default: true)
- `--base-dir`: Custom SSH directory path (default: ~/.ssh)

**Examples:**

```bash
# Save complete SSH setup with all includes and keys
gz dev-env ssh save --name production --description "Production environment setup"

# Save configuration without private keys (config only)
gz dev-env ssh save --name minimal --include-keys=false

# Save from custom SSH directory
gz dev-env ssh save --name custom --base-dir /custom/ssh/path
```

**What Gets Saved:**

```
~/.gz/ssh-configs/production/
├── config              # Main ~/.ssh/config file
├── includes/           # All Include directive files
│   ├── config.d/       
│   │   ├── work.conf
│   │   └── personal.conf
│   └── special.config
├── keys/               # All IdentityFile keys + public keys
│   ├── id_rsa
│   ├── id_rsa.pub
│   ├── work_ed25519
│   └── work_ed25519.pub
└── metadata.json       # Metadata with file listings and timestamps
```

### Load SSH Configuration

```bash
gz dev-env ssh load --name <config-name> [options]
```

**Options:**
- `--name`: Configuration name to load (required)
- `--backup`: Create backup of current SSH config before loading
- `--force`: Overwrite existing files without confirmation

**Examples:**

```bash
# Load saved configuration (with backup)
gz dev-env ssh load --name production --backup

# Force load without confirmation
gz dev-env ssh load --name staging --force
```

### List Configurations

```bash
gz dev-env ssh list [options]
```

**Options:**
- `--all`: Show detailed information for all configurations
- `--name`: Show details for specific configuration

**Examples:**

```bash
# List all saved configurations
gz dev-env ssh list

# Show detailed information
gz dev-env ssh list --all

# Show specific configuration details
gz dev-env ssh list --name production
```

### Install SSH Keys to Remote Server

#### Advanced Installation (Recommended)

```bash
gz dev-env ssh install-key --host <hostname> --user <username> [options]
```

**Options:**
- `--host`: Remote server hostname or IP (required)
- `--user`: SSH username (required)  
- `--public-key`: Path to public key file
- `--private-key`: Path to private key for authentication
- `--config`: Install all keys from saved configuration
- `--password`: SSH password (prompted if not provided)
- `--port`: SSH port (default: 22)
- `--force`: Install even if key already exists
- `--dry-run`: Show what would be done without making changes

**Examples:**

```bash
# Install specific public key
gz dev-env ssh install-key --host server.com --user admin --public-key ~/.ssh/id_rsa.pub

# Install with verbose output
gz dev-env ssh install-key --host server.com --user admin --public-key ~/.ssh/id_rsa.pub --verbose

# Install all keys from saved configuration  
gz dev-env ssh install-key --config production --host server.com --user admin

# Dry run to preview changes
gz dev-env ssh install-key --config production --host server.com --user admin --dry-run
```

#### Simple Installation (Fallback)

```bash
gz dev-env ssh install-key-simple --host <hostname> --user <username> --public-key <key-path>
```

Uses system SSH command for compatibility when the advanced installer encounters issues.

### List Available Keys

```bash
gz dev-env ssh list-keys [options]
```

**Options:**
- `--config`: Show keys from specific configuration

**Examples:**

```bash
# List keys from all configurations
gz dev-env ssh list-keys

# List keys from specific configuration
gz dev-env ssh list-keys --config production
```

## Logging and Output

### Standard Mode (Default)

Provides clean, minimal output focusing on essential information:

```bash
Password for user@host: 
✅ SSH key installed successfully
```

### Verbose Mode (`--verbose` or `-v`)

Shows detailed step-by-step information for debugging:

```bash
Password for user@host:
Password read successfully (length: 8)
Attempting to connect to 192.168.1.22:22...
SSH connection established successfully
Checking if key already exists on remote server...
Creating SFTP client for key existence check...
Reading .ssh/authorized_keys file via SFTP...
Key not found in authorized_keys
Installing key to remote server...
Creating SFTP client for key installation...
Ensuring .ssh directory exists...
Setting .ssh directory permissions to 0700...
Processing keys...
Writing updated authorized_keys file...
Setting authorized_keys permissions to 0600...
✅ SSH key installed successfully
Total keys in authorized_keys: 1
```

## File Structure and Metadata

### Configuration Storage

SSH configurations are stored in `~/.gz/ssh-configs/` with the following structure:

```
~/.gz/ssh-configs/
├── production/
│   ├── config              # Main SSH config
│   ├── includes/           # Include directive files
│   │   └── config.d/
│   │       ├── work.conf
│   │       └── personal.conf  
│   ├── keys/               # SSH keys
│   │   ├── id_rsa
│   │   ├── id_rsa.pub
│   │   ├── work_key
│   │   └── work_key.pub
│   └── metadata.json       # Configuration metadata
└── staging/
    └── ...
```

### Metadata Format

```json
{
  "name": "production",
  "description": "Production environment SSH setup",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z",
  "include_keys": true,
  "base_dir": "/home/user/.ssh",
  "main_config": "/home/user/.ssh/config",
  "include_files": [
    "/home/user/.ssh/config.d/work.conf",
    "/home/user/.ssh/config.d/personal.conf"
  ],
  "private_keys": [
    "/home/user/.ssh/id_rsa",
    "/home/user/.ssh/work_key"
  ],
  "public_keys": [
    "/home/user/.ssh/id_rsa.pub", 
    "/home/user/.ssh/work_key.pub"
  ]
}
```

## Use Cases

### Development Environment Setup

```bash
# Save your complete SSH setup
gz dev-env ssh save --name dev-machine --description "My development setup"

# On new machine, restore everything
gz dev-env ssh load --name dev-machine
```

### Project-Specific SSH Configurations

```bash
# Save project-specific SSH setup
gz dev-env ssh save --name project-alpha --description "Alpha project SSH config"

# Switch between projects
gz dev-env ssh load --name project-alpha
gz dev-env ssh load --name project-beta
```

### Server Deployment

```bash
# Install keys to multiple servers from saved configuration
gz dev-env ssh install-key --config production --host web1.example.com --user deploy
gz dev-env ssh install-key --config production --host web2.example.com --user deploy
gz dev-env ssh install-key --config production --host db.example.com --user admin
```

### Backup and Migration

```bash
# Create backup before making changes
gz dev-env ssh save --name backup-$(date +%Y%m%d) --description "Backup before changes"

# Migrate SSH setup to new machine
gz dev-env ssh list --all              # Review saved configurations
gz dev-env ssh load --name production  # Restore on new machine
```

## Security Considerations

### File Permissions

The system automatically maintains proper SSH file permissions:

- **SSH Directory**: `~/.ssh` set to 700 (owner read/write/execute only)
- **Private Keys**: Set to 600 (owner read/write only)
- **Public Keys**: Set to 644 (owner read/write, others read)
- **Config Files**: Set to 644 (owner read/write, others read)
- **Remote authorized_keys**: Set to 600 (owner read/write only)

### Key Storage

- Private keys are stored in the configuration backup with original permissions
- Keys are never transmitted in plain text (only via secure SSH/SFTP connections)
- Backup directories have restricted access permissions

### Remote Installation

- Uses secure SSH/SFTP protocols for all remote operations
- Supports multiple authentication methods (key-based, password, keyboard-interactive)
- Validates remote file operations before making changes
- Automatic deduplication prevents key accumulation

## Troubleshooting

### Common Issues

#### SSH Connection Failures

**Problem**: Cannot connect to remote server

```bash
# Use verbose mode for detailed connection information
gz dev-env ssh install-key --host server.com --user admin --public-key ~/.ssh/id_rsa.pub --verbose
```

**Solutions**:
- Verify hostname and port are correct
- Check if SSH service is running on remote server
- Ensure firewall allows SSH connections
- Try simple installation method: `install-key-simple`

#### Include File Parsing Issues

**Problem**: Include files not being parsed correctly

**Solutions**:
- Check include directive syntax in SSH config
- Verify include file paths exist and are readable
- Use absolute paths in Include directives
- Check for circular includes

#### Permission Errors

**Problem**: Permission denied errors during installation

**Solutions**:
- Verify SSH user has permission to modify `~/.ssh/authorized_keys`
- Check if `~/.ssh` directory exists and has proper permissions
- Try installing as a different user with appropriate permissions

### Debug Commands

```bash
# Verbose output for all operations
gz dev-env ssh save --name debug --verbose
gz dev-env ssh load --name debug --verbose
gz dev-env ssh install-key --host server.com --user admin --public-key ~/.ssh/id_rsa.pub --verbose

# Dry run for testing
gz dev-env ssh install-key --config production --host server.com --user admin --dry-run

# List detailed configuration information
gz dev-env ssh list --all
```

## Integration

### With Other Dev-Env Commands

The SSH management integrates with other `gz dev-env` commands:

```bash
# Complete environment switch including SSH
gz dev-env switch-all --environment production  # (includes SSH config)

# Status check including SSH
gz dev-env status  # Shows current SSH configuration status
```

### With Repository Operations

```bash
# Set up SSH for repository access
gz dev-env ssh load --name git-access
gz git repo clone-or-update git@github.com:user/repo.git
```

### Automation Scripts

```bash
#!/bin/bash
# Deployment setup script

echo "Setting up SSH for deployment..."
gz dev-env ssh load --name deployment --force

echo "Installing keys to servers..."
for server in web1 web2 db; do
    gz dev-env ssh install-key --config deployment --host $server.example.com --user deploy
done

echo "SSH setup complete!"
```

This comprehensive SSH management system provides robust, secure, and convenient SSH configuration handling for modern development workflows.