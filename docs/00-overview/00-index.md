# Documentation Index

Complete documentation map for the gzh-cli project, organized for easy navigation and understanding.

## Table of Contents

1. [Quick Start](#quick-start)
1. [Documentation Structure](#documentation-structure)
1. [Navigation Guide](#navigation-guide)
1. [Learning Paths](#learning-paths)
1. [Documentation Management](#documentation-management)
1. [Reference Materials](#reference-materials)

## Quick Start

### New Users

1. **Installation**: See [Getting Started](../10-getting-started/)
1. **First Commands**: Try `gz ide scan`, `gz git repo sync`, `gz quality run`
1. **Configuration**: Set up [Configuration](../40-configuration/40-configuration-guide.md)
1. **Commands**: Browse [Complete Command Reference](../50-api-reference/50-command-reference.md)

### Returning Users

- **📋 Command Reference**: [Complete Command Reference](../50-api-reference/50-command-reference.md)
- **⚙️ Configuration**: [Configuration Guide](../40-configuration/40-configuration-guide.md)
- **🔧 Troubleshooting**: See command-specific troubleshooting sections

## Documentation Structure

### 📂 Unified Documentation Organization

```
docs/
├── 00-overview/           # Project overview and navigation
├── 10-getting-started/    # Installation and initial setup
├── 20-architecture/       # System design and architecture
├── 30-features/          # Core feature documentation
├── 40-configuration/     # Configuration management
├── 50-api-reference/     # Complete command documentation
├── 60-development/       # Development guidelines
├── 70-deployment/        # Deployment and release management
├── 80-integrations/      # External integrations
├── 90-maintenance/       # Maintenance and troubleshooting
└── 99-appendix/          # Additional resources
```

## Navigation Guide

### 🎯 By Use Case

#### Repository Management

- **Quick Start**: [Repository Management](../30-features/31-repository-management.md)
- **Multi-Platform Sync**: [Synclone Guide](../30-features/30-synclone.md)
- **Cross-Platform Sync**: `gz git repo sync` for GitHub ↔ GitLab ↔ Gitea synchronization
- **Command Details**: [git repo commands](../50-api-reference/50-command-reference.md#git-repo) | [git webhook](../50-api-reference/50-command-reference.md#git-webhook) | [synclone commands](../50-api-reference/50-command-reference.md#synclone)

#### Development Environment

- **Enhanced IDE Management**: [IDE Features](../30-features/35-ide-management.md) - Now with `scan`, `status`, `open` commands
- **IDE Detection**: Support for JetBrains, VS Code family, and popular editors
- **Environment Setup**: [Development Environment](../30-features/33-development-environment.md)
- **Network Management**: [Network Features](../30-features/34-network-management.md)
- **Command Details**: [ide commands](../50-api-reference/50-command-reference.md#ide) | [dev-env commands](../50-api-reference/50-command-reference.md#dev-env) | [net-env commands](../50-api-reference/50-command-reference.md#net-env)

#### Code Quality

- **Quality Management**: [Quality Features](../30-features/36-quality-management.md) - Test coverage improved to 34.4%
- **Command Details**: [quality commands](../50-api-reference/50-command-reference.md#quality)

#### Performance and Monitoring

- **Performance Profiling**: [Profiling Features](../30-features/37-performance-profiling.md) - Test coverage improved to 36.6%
- **System Health**: Enhanced doctor package with 10.3% test coverage
- **Command Details**: [profile commands](../50-api-reference/50-command-reference.md#profile) | [doctor commands](../50-api-reference/50-command-reference.md#doctor)

#### Output Formats and Backup

- **New Features**: [Output Formats & Backup](../30-features/32-output-formats-backup.md)
- **Format Examples**: JSON, YAML, CSV, HTML, SARIF output formats
- **Backup Features**: Development environment configuration backup/restore

### 🎯 By User Type

#### End Users

- **Getting Started**: [Installation](../10-getting-started/) → [Configuration](../40-configuration/40-configuration-guide.md) → [Commands](../50-api-reference/50-command-reference.md)
- **Daily Usage**: [Command Reference](../50-api-reference/50-command-reference.md) | [Troubleshooting](../90-maintenance/90-troubleshooting.md)

#### System Administrators

- **Enterprise Setup**: [Enterprise Features](../80-integrations/enterprise/)
- **Configuration Management**: [Configuration Guide](../40-configuration/40-configuration-guide.md)
- **Repository Policies**: [Repository Management](../30-features/31-repository-management.md)

#### Developers

- **Architecture**: [System Overview](../20-architecture/20-system-overview.md)
- **Development**: [Development Guidelines](../60-development/)
- **Testing**: [Testing Strategy](../60-development/testing-strategy.md)
- **API Integration**: [Command Reference](../50-api-reference/50-command-reference.md)

#### DevOps Engineers

- **CI/CD Integration**: [Output Formats](../30-features/32-output-formats-backup.md#cicd-pipeline-integration)
- **Automation**: [Configuration Management](../40-configuration/40-configuration-guide.md)
- **Monitoring**: [Performance Features](../30-features/37-performance-profiling.md)

## Learning Paths

### 🚀 Path 1: Basic Usage (30 minutes)

1. [Install gzh-cli](../10-getting-started/10-installation.md) (5 min)
1. [Try first commands](../10-getting-started/11-quick-start.md) (10 min)
1. [Set up configuration](../40-configuration/40-configuration-guide.md) (15 min)

### 🔧 Path 2: Repository Management (45 minutes)

1. [Understanding synclone](../30-features/30-synclone.md) (15 min)
1. [Configuration setup](../40-configuration/40-configuration-guide.md#synclone) (15 min)
1. [Advanced repository management](../30-features/31-repository-management.md) (15 min)

### 💻 Path 3: Development Environment (60 minutes)

1. [Development environment basics](../30-features/33-development-environment.md) (20 min)
1. [IDE management](../30-features/35-ide-management.md) (20 min)
1. [Network environment management](../30-features/34-network-management.md) (20 min)

### 🏗️ Path 4: Architecture and Development (90 minutes)

1. [System architecture](../20-architecture/20-system-overview.md) (30 min)
1. [Development setup](../60-development/60-development-setup.md) (30 min)
1. [Contributing guidelines](../60-development/61-contributing.md) (30 min)

## Documentation Management

### Documentation Categories

#### 1. Core Project Documents (Protected)

**Location**: Project root directory
**Protection**: AI modification prohibited
**Files**:

- README.md - Project overview and quick start
- TECH_STACK.md - Technology stack and architecture
- FEATURES.md - Feature list and capabilities
- USAGE.md - Detailed usage instructions
- CHANGELOG.md - Version history and changes
- SECURITY.md - Security policy and practices
- LICENSE - Project license
- CLAUDE.md - AI agent instructions

**Management Rules**:

- These files contain `<!-- 🚫 AI_MODIFY_PROHIBITED -->` header
- AI agents should NOT modify these files
- Only human maintainers should update these documents
- Changes require careful review and approval

#### 2. Auto-generated API Documentation

**Location**: `/api-docs/`
**Protection**: No manual edits allowed
**Content**: API reference documentation generated from code

**Management Rules**:

- Generated automatically by documentation tools
- Manual edits will be overwritten
- Source code comments should be updated instead
- Regenerate using appropriate build commands

#### 3. Core Design Specifications (Protected)

**Location**: `/specs/`
**Protection**: AI modification prohibited
**Files**:

- common.md - Common functionality specifications
- dev-env.md - Development environment management specs
- net-env.md - Network environment management specs
- package-manager.md - Package manager integration specs
- synclone.md - Repository synchronization specs

**Management Rules**:

- These files contain `<!-- 🚫 AI_MODIFY_PROHIBITED -->` header
- Human-written design documents
- AI agents should NOT modify these files

#### 4. General Documentation (Editable)

**Location**: `/docs/`
**Protection**: Standard editing allowed
**Content**: Feature guides, tutorials, API documentation, development guides

**Management Rules**:

- AI agents CAN modify these files for improvements
- Follow project documentation standards
- Maintain consistency with existing style
- Test examples and verify accuracy

### Protection Mechanisms

1. **.claudeignore file**: Lists protected files and directories
1. **File headers**: Protected files contain `<!-- 🚫 AI_MODIFY_PROHIBITED -->` header
1. **Directory structure**: Clear separation between protected and editable documentation

## Reference Materials

### Essential Documentation

- **[Complete Command Reference](../50-api-reference/50-command-reference.md)** - All commands and options
- **[Configuration Guide](../40-configuration/40-configuration-guide.md)** - Complete configuration reference
- **[Architecture Overview](../20-architecture/20-system-overview.md)** - System design and patterns

### Development Resources

- **[Development Guidelines](../60-development/)** - Coding standards and practices
- **[Testing Strategy](../60-development/testing-strategy.md)** - Testing approach and tools
- **[Deployment Guide](../70-deployment/)** - Release and deployment processes

### Support and Community

- **[Troubleshooting](../90-maintenance/90-troubleshooting.md)** - Common issues and solutions
- **[FAQ](../90-maintenance/91-faq.md)** - Frequently asked questions
- **[Contributing](../60-development/61-contributing.md)** - How to contribute to the project

______________________________________________________________________

**Last Updated**: 2025-08-22
**Documentation Version**: 2.1.0
**CLI Version**: Latest (Git package test coverage: 91.7%)

For questions or feedback about this documentation, please see the [Contributing Guide](../60-development/61-contributing.md).
