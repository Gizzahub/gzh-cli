# 📖 Glossary

Complete glossary of terms, concepts, and acronyms used throughout gzh-cli documentation and codebase.

## A

**Actions Policy**
: A comprehensive configuration system for managing GitHub Actions permissions, security settings, and compliance rules at the organization or repository level.

**API Client**
: The interface component that handles communication with Git platform APIs (GitHub, GitLab, Gitea, Gogs) using authentication tokens and rate limiting.

**Audit**
: The process of examining repository configurations, quality metrics, and compliance with established policies to identify violations or improvements.

## B

**BaseCommand Pattern**
: The architectural pattern used in gzh-cli where all commands inherit from a common base interface, providing consistent error handling, configuration loading, and logging.

**Bulk Operations**
: Commands that operate on multiple repositories simultaneously, such as `synclone` for organization-wide repository synchronization or mass webhook management.

## C

**Clone Strategy**
: The method used when acquiring or updating repositories. Options include `rebase`, `reset`, `clone`, `skip`, `pull`, and `fetch`, each with different behavior for handling local changes.

**Configuration Hierarchy**
: The priority order for configuration sources: CLI flags → environment variables → config files → defaults.

**Concurrent Jobs**
: The number of parallel operations gzh-cli can perform simultaneously, configurable to balance performance with API rate limits.

## D

**Dry Run**
: A mode where commands simulate their actions without making actual changes, allowing users to preview results before execution.

**Development Environment (dev-env)**
: Development environment management through both individual service control (fine-grained management of AWS, GCP, Azure, Docker, Kubernetes, SSH configurations) and unified environment operations (TUI dashboard, atomic switching, status monitoring).

## E

**Enforcer**
: The component responsible for applying policies to repositories and validating compliance with defined rules.

**Environment Variable Expansion**
: The feature that allows configuration files to reference environment variables using `${VARIABLE_NAME}` syntax.

## F

**Formatter**
: A component that standardizes output display across different formats (table, JSON, YAML, CSV, HTML) for consistent user experience.

## G

**Git Provider**
: Any Git hosting platform supported by gzh-cli, including GitHub, GitLab, Gitea, and Gogs.

**gzh.yaml**
: The unified configuration file format that centralizes settings for all gzh-cli commands and providers.

## H

**Hook**
: Shell commands configured to execute in response to specific tool events, such as before or after command execution.

## I

**IDE Integration**
: Features for monitoring and managing JetBrains IDE settings, detecting configuration drift, and maintaining consistent development environments.

**Interface-Driven Design**
: The architectural principle where gzh-cli defines clear interfaces before implementations, enabling testability and extensibility.

## J

**JSON Schema**
: The validation framework used to ensure configuration files conform to expected structure and data types.

## L

**Linter**
: A code analysis tool that checks for style violations, potential bugs, and adherence to coding standards across multiple programming languages.

## M

**Mock Generation**
: The automated creation of test doubles using `gomock` for interface testing and development.

**Multi-Platform Support**
: The capability to operate across different Git hosting platforms and operating systems with consistent behavior.

## N

**Network Environment (net-env)**
: Network environment management through interactive TUI dashboard, status monitoring, profile management, network actions, and cloud integration. Advanced features like automatic WiFi detection and complex VPN management are planned for future releases.

## O

**Organization**
: A Git platform entity that contains multiple repositories and can be managed as a unit through bulk operations.

**Output Format**
: The presentation style for command results, supporting table, JSON, YAML, CSV, and HTML formats.

## P

**Package Manager (PM)**
: Tools for managing software dependencies and runtime environments, including asdf, Homebrew, SDKMAN, npm, pip, and others.

**Policy Validation**
: The process of checking whether repositories comply with defined organizational policies and security requirements.

**Provider Registry**
: The system that manages and instantiates different Git platform implementations at runtime.

## Q

**Quality Management**
: The integrated system for running code quality checks, security scans, and compliance validations across multiple programming languages.

## R

**Rate Limiting**
: The mechanism for respecting API quotas and preventing excessive requests to Git platform services.

**Repository Configuration (repo-config)**
: Settings and policies applied to individual repositories, including webhooks, branch protection, and Actions policies.

## S

**Schema Validation**
: The process of verifying that configuration files match expected structure and contain valid data types and values.

**Synclone**
: The primary bulk operation for synchronizing repositories from Git platform organizations to local filesystem with various strategies.

**Strategy Pattern**
: The design approach used for Git operations where different update strategies (rebase, reset, pull, etc.) can be selected based on requirements.

## T

**Token Management**
: The secure handling of authentication credentials for Git platform APIs, supporting environment variable storage and rotation.

**Troubleshooting**
: The systematic approach to diagnosing and resolving issues with gzh-cli operations, configurations, or integrations.

## V

**Validation Rules**
: Specific checks performed during policy enforcement to ensure compliance with organizational standards and security requirements.

**Version Control**
: The management of policy and configuration versions, enabling rollback and change tracking.

## W

**Webhook Management**
: The comprehensive system for creating, updating, and managing Git platform webhooks at individual repository or organization scale.

**Worker Pool**
: The concurrency management system that controls parallel execution of bulk operations while respecting resource constraints.

**Workflow Permissions**
: GitHub Actions token permissions that control what actions workflows can perform within repositories.

## Acronyms and Abbreviations

**API**
: Application Programming Interface

**CI/CD**
: Continuous Integration/Continuous Deployment

**CLI**
: Command Line Interface

**CRUD**
: Create, Read, Update, Delete

**HTTPS**
: HyperText Transfer Protocol Secure

**IDE**
: Integrated Development Environment

**JSON**
: JavaScript Object Notation

**LDAP**
: Lightweight Directory Access Protocol

**MFA**
: Multi-Factor Authentication

**OIDC**
: OpenID Connect

**PR**
: Pull Request

**SAML**
: Security Assertion Markup Language

**SARIF**
: Static Analysis Results Interchange Format

**SSH**
: Secure Shell

**SSL**
: Secure Sockets Layer

**TLS**
: Transport Layer Security

**URL**
: Uniform Resource Locator

**VPN**
: Virtual Private Network

**YAML**
: Yet Another Markup Language / YAML Ain't Markup Language

## Command Reference Quick Lookup

**Core Commands**

- `gz doctor` - System diagnostics
- `gz version` - Version information
- `gz config` - Configuration management

**Repository Operations**

- `gz git` - Git operations and repository management
- `gz synclone` - Bulk repository synchronization
- `gz repo-config` - Repository configuration management

**Development Tools**

- `gz quality` - Code quality management
- `gz dev-env` - Development environment management (individual services + unified operations)
- `gz ide` - IDE integration and monitoring
- `gz pm` - Package manager updates

**Network and System**

- `gz net-env` - Network environment management (TUI, status, profiles, actions, cloud)
- `gz profile` - Performance profiling
- `gz shell` - Interactive debugging shell (debug mode only)

## Configuration Keywords

**Global Settings**

- `clone_base_dir` - Base directory for repository clones
- `concurrent_jobs` - Number of parallel operations
- `default_strategy` - Default Git operation strategy
- `timeout` - Operation timeout duration

**Provider Configuration**

- `token` - Authentication token
- `base_url` - Custom API endpoint
- `organizations` - Organization-specific settings
- `repositories` - Repository-specific settings

**Quality Settings**

- `enabled_checks` - Active quality checks
- `ignore_patterns` - Files/directories to skip
- `severity_threshold` - Minimum severity level

______________________________________________________________________

**Related Documentation**: [Configuration Guide](../40-configuration/40-configuration-guide.md) | [Command Reference](../50-api-reference/50-command-reference.md) | [Troubleshooting](../90-maintenance/90-troubleshooting.md)
**For Developers**: [Architecture](../20-architecture/20-system-overview.md) | [Development Guide](../60-development/60-index.md)
**Enterprise**: [Policy Management](enterprise/) | [Security Compliance](99-security-compliance.md)
