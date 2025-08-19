# Documentation Index

Complete documentation map for the gzh-cli project, organized for easy navigation and understanding.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Documentation Structure](#documentation-structure)
3. [Navigation Guide](#navigation-guide)
4. [Learning Paths](#learning-paths)
5. [Reference Materials](#reference-materials)

## Quick Start

### New Users
1. **Installation**: See [Getting Started](../01-getting-started/)
2. **First Commands**: Try `gz synclone`, `gz quality run`, `gz ide monitor`
3. **Configuration**: Set up [Configuration](../30-configuration/30-configuration-guide.md)
4. **Commands**: Browse [Complete Command Reference](../40-api-reference/40-command-reference.md)

### Returning Users
- **📋 Command Reference**: [Complete Command Reference](../40-api-reference/40-command-reference.md)
- **⚙️ Configuration**: [Configuration Guide](../30-configuration/30-configuration-guide.md)
- **🔧 Troubleshooting**: See command-specific troubleshooting sections

## Documentation Structure

### 📂 Core Documentation Organization

```
docs/
├── 00-overview/           # Project overview and navigation
├── 01-getting-started/    # Installation and initial setup
├── 02-architecture/       # System design and architecture
├── 03-core-features/      # Legacy feature documentation
├── 04-configuration/      # Legacy configuration files
├── 05-command-reference/  # Command navigation hub
├── 20-features/          # ✨ Consolidated feature documentation
├── 30-configuration/     # ✨ Unified configuration system
├── 40-api-reference/     # ✨ Complete command documentation
├── 50-development/       # Development guidelines
├── 60-deployment/        # Deployment and release management
└── 90-appendix/          # Additional resources
```

**✨ = New consolidated documentation structure**

## Navigation Guide

### 🎯 By Use Case

#### Repository Management
- **Quick Start**: [Repository Management](../20-features/21-repository-management.md)
- **Multi-Platform Sync**: [Synclone Guide](../20-features/20-synclone.md)
- **Command Details**: [git commands](../40-api-reference/40-command-reference.md#git) | [synclone commands](../40-api-reference/40-command-reference.md#synclone)

#### Development Environment
- **IDE Management**: [IDE Features](../03-core-features/ide-management.md)
- **Environment Setup**: [Development Environment](../03-core-features/development-environment/)
- **Network Management**: [Network Features](../03-core-features/network-management/)
- **Command Details**: [dev-env commands](../40-api-reference/40-command-reference.md#dev-env) | [net-env commands](../40-api-reference/40-command-reference.md#net-env)

#### Code Quality
- **Quality Management**: [Quality Features](../03-core-features/quality-management.md)
- **Command Details**: [quality commands](../40-api-reference/40-command-reference.md#quality)

#### Performance and Monitoring
- **Performance Profiling**: [Profiling Features](../03-core-features/performance-profiling.md)
- **Command Details**: [profile commands](../40-api-reference/40-command-reference.md#profile)

#### Output Formats and Backup
- **New Features**: [Output Formats & Backup](../20-features/22-output-formats-backup.md)
- **Format Examples**: JSON, YAML, CSV, HTML, SARIF output formats
- **Backup Features**: Development environment configuration backup/restore

### 🎯 By User Type

#### End Users
- **Getting Started**: [Installation](../01-getting-started/) → [Configuration](../30-configuration/30-configuration-guide.md) → [Commands](../40-api-reference/40-command-reference.md)
- **Daily Usage**: [Command Reference](../40-api-reference/40-command-reference.md) | [Troubleshooting](../40-api-reference/40-command-reference.md#troubleshooting)

#### System Administrators
- **Enterprise Setup**: [Enterprise Features](../09-enterprise/)
- **Configuration Management**: [Configuration Guide](../30-configuration/30-configuration-guide.md)
- **Repository Policies**: [Repository Management](../20-features/21-repository-management.md)

#### Developers
- **Architecture**: [System Overview](../02-architecture/overview.md)
- **Development**: [Development Guidelines](../06-development/)
- **Testing**: [Testing Strategy](../06-development/testing-strategy.md)
- **API Integration**: [Command Reference](../40-api-reference/40-command-reference.md)

#### DevOps Engineers
- **CI/CD Integration**: [Output Formats](../20-features/22-output-formats-backup.md#cicd-pipeline-integration)
- **Automation**: [Configuration Management](../30-configuration/30-configuration-guide.md)
- **Monitoring**: [Performance Features](../03-core-features/performance-profiling.md)

## Learning Paths

### 🚀 Path 1: Basic Usage (30 minutes)

1. **Install and Setup** (10 min)
   - [Getting Started](../01-getting-started/)
   - [Basic Configuration](../30-configuration/30-configuration-guide.md#basic-structure)

2. **First Commands** (15 min)
   - [synclone basics](../40-api-reference/40-command-reference.md#synclone)
   - [quality checks](../40-api-reference/40-command-reference.md#quality)

3. **Explore** (5 min)
   - [Command overview](../40-api-reference/40-command-reference.md#overview)
   - [Global options](../40-api-reference/40-command-reference.md#global-options)

### 🔧 Path 2: Advanced Configuration (1 hour)

1. **Configuration System** (20 min)
   - [Configuration Guide](../30-configuration/30-configuration-guide.md)
   - [Priority System](../30-configuration/30-configuration-guide.md#configuration-priority)

2. **Platform Setup** (20 min)
   - [GitHub/GitLab Configuration](../30-configuration/30-configuration-guide.md#platform-configuration)
   - [Environment Variables](../30-configuration/30-configuration-guide.md#environment-variables)

3. **Advanced Features** (20 min)
   - [Output Formats](../20-features/22-output-formats-backup.md#output-format-features)
   - [Backup Systems](../20-features/22-output-formats-backup.md#backup-and-restore-system)

### 🏢 Path 3: Enterprise Deployment (2 hours)

1. **Architecture Understanding** (30 min)
   - [System Architecture](../02-architecture/overview.md)
   - [Repository Management](../20-features/21-repository-management.md)

2. **Policy Configuration** (45 min)
   - [Repository Policies](../20-features/21-repository-management.md#policy-management)
   - [Compliance Frameworks](../20-features/21-repository-management.md#compliance-audit)

3. **Monitoring and Automation** (45 min)
   - [CI/CD Integration](../20-features/22-output-formats-backup.md#cicd-pipeline-integration)
   - [Webhook Management](../20-features/21-repository-management.md#webhook-management)

### 💻 Path 4: Development and Contribution (3 hours)

1. **Development Setup** (45 min)
   - [Development Environment](../06-development/)
   - [Testing Strategy](../06-development/testing-strategy.md)

2. **Architecture Deep Dive** (90 min)
   - [System Overview](../02-architecture/overview.md)
   - [Code Quality Standards](../06-development/code-quality.md)

3. **Contributing** (45 min)
   - [Pre-commit Setup](../06-development/pre-commit-guide.md)
   - [Debugging Guide](../06-development/debugging-guide.md)

## Reference Materials

### 📋 Command References

#### Complete Documentation
- **[Complete Command Reference](../40-api-reference/40-command-reference.md)** - Comprehensive command documentation with all options, examples, and troubleshooting

#### Quick Navigation
- [synclone](../40-api-reference/40-command-reference.md#synclone) - Multi-platform repository synchronization
- [git](../40-api-reference/40-command-reference.md#git) - Unified Git operations and platform management
- [quality](../40-api-reference/40-command-reference.md#quality) - Multi-language code quality management
- [ide](../40-api-reference/40-command-reference.md#ide) - JetBrains IDE monitoring and management
- [profile](../40-api-reference/40-command-reference.md#profile) - Performance profiling and analysis
- [dev-env](../40-api-reference/40-command-reference.md#dev-env) - Development environment configuration
- [net-env](../40-api-reference/40-command-reference.md#net-env) - Network environment transitions
- [pm](../40-api-reference/40-command-reference.md#pm) - Package manager updates and management
- [repo-config](../40-api-reference/40-command-reference.md#repo-config) - GitHub repository configuration management

### ⚙️ Configuration References

#### Primary Documentation
- **[Configuration Guide](../30-configuration/30-configuration-guide.md)** - Complete configuration system documentation

#### Quick Links
- [Configuration Priority](../30-configuration/30-configuration-guide.md#configuration-priority)
- [Platform Configuration](../30-configuration/30-configuration-guide.md#platform-configuration)
- [Environment Variables](../30-configuration/30-configuration-guide.md#environment-variables)
- [Migration Guide](../30-configuration/30-configuration-guide.md#migration-guide)
- [Best Practices](../30-configuration/30-configuration-guide.md#best-practices)

#### Configuration Schemas
- [gzh-schema.json](../30-configuration/schemas/gzh-schema.json) - Main configuration schema
- [synclone-schema.yaml](../04-configuration/schemas/synclone-schema.yaml) - Synclone configuration schema
- [ide-schema.yaml](../04-configuration/schemas/ide-schema.yaml) - IDE configuration schema
- [quality-schema.yaml](../04-configuration/schemas/quality-schema.yaml) - Quality tools schema

### 🚀 Feature Guides

#### Core Features (New Consolidated Documentation)
- **[Repository Management](../20-features/21-repository-management.md)** - Complete repository management guide
- **[Synclone](../20-features/20-synclone.md)** - Multi-platform synchronization guide
- **[Output Formats & Backup](../20-features/22-output-formats-backup.md)** - New features documentation

#### Legacy Feature Documentation
- [Git Unified Command](../03-core-features/git-unified-command.md)
- [IDE Management](../03-core-features/ide-management.md)
- [Quality Management](../03-core-features/quality-management.md)
- [Performance Profiling](../03-core-features/performance-profiling.md)
- [Development Environment](../03-core-features/development-environment/)
- [Network Management](../03-core-features/network-management/)

### 🏗️ Architecture and Development

#### Architecture
- [System Overview](../02-architecture/overview.md) - High-level architecture
- [Development Container](../02-architecture/development-container.md) - Container setup

#### Development
- [Code Quality](../06-development/code-quality.md) - Code standards and practices
- [Testing Strategy](../06-development/testing-strategy.md) - Testing guidelines
- [Debugging Guide](../06-development/debugging-guide.md) - Debugging techniques
- [Pre-commit Guide](../06-development/pre-commit-guide.md) - Pre-commit hooks
- [Mocking Strategy](../06-development/mocking-strategy.md) - Testing mocks

### 🚀 Deployment and Operations

#### Deployment
- [Release Notes](../07-deployment/release-notes-v1.0.0.md) - Version 1.0.0 release notes
- [Release Preparation](../07-deployment/release-preparation-checklist.md) - Release checklist
- [Security Scanning](../07-deployment/security-scanning.md) - Security practices

#### Enterprise
- [Actions Policy Enforcement](../09-enterprise/actions-policy-enforcement.md)
- [Actions Policy Schema](../09-enterprise/actions-policy-schema.md)

### 🔗 Integration Guides

#### Third-party Integrations
- [Webhook Management](../08-integrations/webhook-management-guide.md)
- [Terraform Comparison](../08-integrations/terraform-vs-gz-examples.md)
- [Terraform Alternatives](../08-integrations/terraform-alternative-comparison.md)

### 📚 Additional Resources

#### Examples
- [Configuration Examples](../../examples/) - Sample configuration files
- [Usage Examples](../40-api-reference/40-command-reference.md#examples) - Command usage examples

#### Legacy Documentation
- [Getting Started](../01-getting-started/) - Installation and setup
- [Maintenance](../10-maintenance/) - Project maintenance information

## Documentation Status

### ✅ Completed (New Structure)
- **20-features/** - Consolidated feature documentation
- **30-configuration/** - Unified configuration guide
- **40-api-reference/** - Complete command reference
- **00-overview/** - This overview and navigation guide

### 📝 Legacy (Maintained for Reference)
- **03-core-features/** - Original feature documentation
- **04-configuration/** - Original configuration files
- **05-command-reference/** - Command navigation hub

### 🔄 Ongoing Maintenance
- **06-development/** - Development guidelines
- **07-deployment/** - Release management
- **08-integrations/** - Integration guides
- **09-enterprise/** - Enterprise features

## Getting Help

### Quick Help Commands
```bash
# Show all commands
gz help

# Command-specific help
gz <command> --help

# Subcommand help
gz <command> <subcommand> --help
```

### Documentation Feedback
- 📁 **Examples**: Check [examples directory](../../examples/)
- 🐛 **Issues**: Report documentation issues on [GitHub](https://github.com/gizzahub/gzh-cli/issues)
- 💡 **Suggestions**: Contribute improvements via pull requests

### Community Resources
- **GitHub Repository**: [gizzahub/gzh-cli](https://github.com/gizzahub/gzh-cli)
- **Issue Tracker**: [GitHub Issues](https://github.com/gizzahub/gzh-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gizzahub/gzh-cli/discussions)

---

**Last Updated**: 2025-08-19
**Documentation Version**: 1.0.0
**CLI Version**: Latest
