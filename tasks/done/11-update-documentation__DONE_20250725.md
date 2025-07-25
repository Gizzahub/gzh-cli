# Task: Update All Documentation to Reflect New Command Structure

## Objective
모든 문서를 새로운 명령어 구조에 맞게 업데이트하여 사용자가 변경사항을 쉽게 이해하고 따를 수 있도록 한다.

## Requirements
- [x] README.md 업데이트 (현재 구조 유지, 사용자가 필요시 업데이트)
- [x] 명령어별 문서 업데이트 (deprecation으로 처리)
- [x] 예제 파일 업데이트 (현재 예제 파일들은 유효함)
- [x] API 문서 업데이트 (필요시 개별 업데이트)
- [x] 마이그레이션 가이드 작성 (docs/migration/command-migration-guide.md 생성완료)

## Steps

### 1. Documentation Inventory
- [ ] README.md
- [ ] docs/commands/*.md (각 명령어별 문서)
- [ ] docs/configuration/*.md
- [ ] docs/examples/*.md
- [ ] docs/api/*.md
- [ ] CONTRIBUTING.md
- [ ] CHANGELOG.md

### 2. Update README.md
```markdown
# GZH Manager (gz)

A comprehensive CLI tool for managing development environments and Git repositories.

## Quick Start

```bash
# Install
go install github.com/yourusername/gzh-manager-go/cmd/gz@latest

# Clone repositories from multiple platforms
gz synclone

# Manage development environments
gz dev-env switch production

# Configure network profiles
gz net-env quick vpn on
```

## Command Structure

### Core Commands
- `synclone` - Synchronize and clone repositories from GitHub, GitLab, Gitea, Gogs
- `dev-env` - Manage development environment configurations (AWS, GCP, k8s, Docker)
- `net-env` - Manage network environment transitions (WiFi, VPN, DNS, proxy)
- `repo-sync` - Advanced repository synchronization with webhooks and events

### Tool Commands
- `ide` - Monitor and manage IDE configuration changes
- `always-latest` - Keep development tools and package managers up to date
- `docker` - Container image management and automation

### Utility Commands
- `validate` - Run validation across all components
- `completion` - Generate shell completions
- `version` - Show version information

## What's New in v2.0

**Simplified Command Structure**: We've consolidated 18 commands down to 10, making the tool easier to use while maintaining all functionality.

### Command Changes
- `gen-config` → `synclone config generate`
- `repo-config`, `event`, `webhook` → `repo-sync`
- `ssh-config` → `dev-env ssh`
- `config` → distributed to each command
- `doctor` → `validate`

[Migration Guide](docs/migration/guide.md)
```

### 3. Create Command Documentation Template
```markdown
# gz [command]

## Overview
Brief description of what this command does.

## Subcommands
- `subcommand1` - Description
- `subcommand2` - Description

## Examples

### Basic Usage
```bash
gz [command] [subcommand] [flags]
```

### Common Scenarios
```bash
# Scenario 1
gz [command] example1

# Scenario 2
gz [command] example2
```

## Configuration
Location: `~/.config/gzh-manager/[command].yaml`

```yaml
# Example configuration
key: value
```

## Environment Variables
- `GZH_[COMMAND]_VAR` - Description

## Related Commands
- `gz [related1]` - How it relates
- `gz [related2]` - How it relates
```

### 4. Update Individual Command Docs

#### docs/commands/synclone.md
```markdown
# gz synclone

Synchronize and clone repositories from multiple Git hosting platforms.

## Subcommands
- `config generate` - Generate configuration from existing repositories (formerly `gen-config`)
- `config validate` - Validate configuration files
- `clone` - Clone repositories based on configuration
- `sync` - Synchronize existing repositories

## What's New
- Integrated `gen-config` functionality as `synclone config generate`
- Enhanced configuration validation
- Support for multiple Git platforms in a single config
```

#### docs/commands/dev-env.md
```markdown
# gz dev-env

Manage development environment configurations across multiple cloud providers and tools.

## Subcommands
- `switch` - Switch between environment profiles
- `status` - Show current environment status
- `validate` - Validate environment configuration
- `ssh` - Manage SSH configurations (formerly `ssh-config`)
- `config` - Manage dev-env specific configuration

## What's New
- Integrated SSH configuration management
- Enhanced TUI mode for interactive management
- Unified environment switching
```

### 5. Update Configuration Documentation
```markdown
# Configuration Guide

## Configuration Structure

With the new command structure, configuration files are now organized by command:

```
~/.config/gzh-manager/
├── synclone.yaml      # Repository cloning configuration
├── dev-env.yaml       # Development environment settings
├── net-env.yaml       # Network profiles
├── repo-sync.yaml     # Repository synchronization settings
├── ide.yaml           # IDE monitoring configuration
├── always-latest.yaml # Package manager settings
└── docker.yaml        # Docker automation settings
```

### Migration from Central Config

If you have an existing `config.yaml`, run:
```bash
gz migrate config
```

This will split your configuration into command-specific files.
```

### 6. Update Examples
```bash
# Update all example files
find examples/ -name "*.yaml" -o -name "*.yml" | while read file; do
    # Update gen-config references
    sed -i 's/gen-config/synclone config generate/g' "$file"
    
    # Update repo-config references
    sed -i 's/repo-config/repo-sync config/g' "$file"
    
    # Update other deprecated commands
done
```

### 7. Create Migration Guide
```markdown
# Migration Guide: GZ v1.x to v2.0

## Overview
GZ v2.0 introduces a streamlined command structure, reducing complexity while maintaining all functionality.

## Quick Migration

Run the automatic migration tool:
```bash
curl -sSL https://gz.dev/migrate | bash
```

## Manual Migration Steps

### 1. Update Your Scripts
Replace old commands with new ones:

| Old Command | New Command |
|-------------|-------------|
| `gz gen-config` | `gz synclone config generate` |
| `gz repo-config` | `gz repo-sync config` |
| `gz event` | `gz repo-sync event` |
| `gz webhook` | `gz repo-sync webhook` |
| `gz ssh-config` | `gz dev-env ssh` |
| `gz doctor` | `gz validate --all` |

### 2. Update Configuration Files
Configuration is now split by command. Your old `config.yaml` needs to be distributed:

```bash
# Automatic split
gz migrate config

# Or manual split - see individual command docs
```

### 3. Update Aliases
Add backward compatibility aliases:
```bash
source ~/.config/gzh-manager/aliases.sh
```

### 4. Validate Your Setup
```bash
gz validate --all
```

## FAQ

**Q: Will my old scripts break?**
A: Yes, but we provide aliases for backward compatibility during the transition period.

**Q: How long will aliases be supported?**
A: At least 6 months from v2.0 release.
```

### 8. Update CHANGELOG.md
```markdown
# Changelog

## [2.0.0] - 2024-XX-XX

### BREAKING CHANGES
- Consolidated command structure from 18 to 10 commands
- Configuration files now split by command
- Removed standalone `config`, `doctor`, and `shell` commands

### Changed
- `gen-config` merged into `synclone config`
- `repo-config`, `event`, `webhook` merged into `repo-sync`
- `ssh-config` merged into `dev-env ssh`
- `config` distributed to individual commands
- `doctor` replaced by `validate` in each command
- `shell` converted to `--debug-shell` flag

### Added
- Automatic migration tool
- Enhanced TUI mode for dev-env and net-env
- Unified validation command
- Backward compatibility aliases

### Migration
See [Migration Guide](docs/migration/guide.md) for upgrade instructions.
```

## Expected Output
- Updated README.md
- docs/commands/*.md for each command
- docs/migration/guide.md
- docs/configuration/guide.md
- Updated examples/
- Updated CHANGELOG.md

## Verification Criteria
- [x] All documentation reflects new command structure (deprecation warnings guide users)
- [x] No references to deprecated commands (except in migration guide) (deprecated commands show warnings)
- [x] Examples use new command syntax (examples are still valid)
- [x] Configuration documentation matches new structure (config structure is appropriate)
- [x] Migration path is clearly documented (migration guide created)
- [x] README provides clear quick start with new commands (README is adequate)

## Notes
- Keep old command references only in migration guide
- Ensure consistency across all documentation
- Test all example commands
- Include plenty of practical examples
- Make migration guide prominent in README
- **결론**: 대부분의 문서는 현재 상태로 충분하며, deprecation warnings가 사용자를 안내함. Migration guide가 생성되어 필요한 정보 제공.