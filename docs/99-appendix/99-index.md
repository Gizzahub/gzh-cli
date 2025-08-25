# 📚 Appendix

Supplementary documentation, reference materials, and additional resources for gzh-cli.

## 📋 Table of Contents

- [Glossary](#glossary)
- [Enterprise Features](#enterprise-features)
- [Reference Materials](#reference-materials)
- [Support Resources](#support-resources)

## 📚 Reference Documentation

### Core Reference

- **[99-glossary.md](99-glossary.md)** - Complete glossary of terms and concepts
- **[99-references.md](99-references.md)** - External references and resources
- **[99-changelog-archive.md](99-changelog-archive.md)** - Historical changelog archive
- **[99-migration-guides.md](99-migration-guides.md)** - Migration guides from legacy systems

### Enterprise Documentation

- **[enterprise/](enterprise/)** - Enterprise-specific features and documentation
  - **Actions Policy Schema** - GitHub Actions policy configuration schema
  - **Actions Policy Enforcement** - Policy application and validation system
  - **Webhook Management** - Advanced webhook management patterns

## 🔍 Quick Reference

### Essential Commands

```bash
# System diagnostics
gz doctor                           # Run comprehensive system check
gz version --detailed               # Show detailed version information
gz config validate                  # Validate configuration

# Repository operations
gz git repo clone-or-update URL     # Smart repository management
gz synclone github --org myorg      # Synchronize organization repositories
gz quality run                      # Run code quality checks

# Development environment
gz dev-env status                   # Check development environment
gz pm update                        # Update package managers
gz ide monitor                      # Monitor IDE settings
```

### Configuration Hierarchy

1. Command-line flags (highest priority)
1. Environment variables
1. Configuration files:
   - `./gzh.yaml` (current directory)
   - `~/.config/gzh-manager/gzh.yaml` (user config)
   - `/etc/gzh-manager/gzh.yaml` (system config)
1. Built-in defaults (lowest priority)

### Supported Platforms

- **Git Providers**: GitHub, GitLab, Gitea, Gogs
- **Operating Systems**: Linux, macOS, Windows
- **Package Managers**: asdf, Homebrew, SDKMAN, npm, pip, cargo, go modules
- **IDEs**: JetBrains IDEs (IntelliJ, WebStorm, GoLand, etc.)

## 🏢 Enterprise Features

### Policy Management

- GitHub Actions policy enforcement
- Repository configuration compliance
- Webhook management at scale
- Security policy automation

### Integration Points

- CI/CD pipeline integration
- Monitoring and observability
- Enterprise authentication (LDAP, SAML)
- Audit logging and compliance reporting

## 📖 Additional Resources

### Community

- **GitHub Repository**: https://github.com/gizzahub/gzh-cli
- **Issue Tracker**: Report bugs and feature requests
- **Discussions**: Community support and questions
- **Wiki**: Community-contributed documentation

### Documentation Standards

- All documentation follows the unified numbering system (10-unit increments)
- Files use kebab-case naming convention with numeric prefixes
- Content is organized by complexity and usage frequency
- Cross-references use relative paths with file extensions

### License and Attribution

- **License**: See [LICENSE](../../LICENSE) file
- **Third-party Licenses**: See [THIRD_PARTY_LICENSES.md](99-third-party-licenses.md)
- **Contributors**: See [CONTRIBUTORS.md](99-contributors.md)

______________________________________________________________________

**Quick Navigation**: [Overview](../00-overview/) | [Getting Started](../10-getting-started/) | [Features](../30-features/) | [Configuration](../40-configuration/) | [API Reference](../50-api-reference/)
**Support Resources**: [Troubleshooting](../90-maintenance/90-troubleshooting.md) | [GitHub Issues](https://github.com/gizzahub/gzh-cli/issues) | [Documentation](../00-overview/00-index.md)
**Enterprise**: [Policy Management](enterprise/) | [Integrations](../80-integrations/) | [Security](99-security-compliance.md)
