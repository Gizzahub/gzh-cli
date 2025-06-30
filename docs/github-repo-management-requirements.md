# GitHub Repository Management Requirements

This document outlines the detailed requirements for implementing GitHub repository configuration management functionality in gzh-manager.

## Overview

The GitHub repository management feature allows users to manage repository settings across entire organizations in a standardized, automated way. This provides infrastructure-as-code capabilities for GitHub repository configurations.

## Core Requirements

### 1. Repository Configuration Management

#### 1.1 Supported Settings
- **Basic Settings**
  - Repository name and description
  - Homepage URL and topics
  - Visibility (public/private)
  - Default branch configuration
  - Features enable/disable (Issues, Wiki, Projects, etc.)

- **Security Settings**
  - Branch protection rules
  - Required status checks
  - Required reviews (count, dismiss stale reviews, require code owner reviews)
  - Restrict pushes to specific users/teams
  - Allow force pushes and deletions

- **Collaboration Settings**
  - Issue and PR templates
  - Auto-merge configuration
  - Delete head branches after merge
  - Squash merge options

- **Advanced Settings**
  - Webhook configurations
  - Deploy keys management
  - Secrets management (organization and repository level)
  - GitHub Actions permissions and settings

#### 1.2 Configuration Schema

**Note: The configuration schema has been implemented in `pkg/config/repo_config_schema.go` with comprehensive support for all features listed below.**

```yaml
# Repository configuration schema (v1.0.0)
version: "1.0.0"
organization: "my-organization"

# Default settings for all repositories
defaults:
  template: "standard"
  settings:
    private: true
    has_issues: true

# Reusable templates
templates:
  standard:
    description: "Standard repository configuration"
    settings:
      has_issues: true
      has_wiki: false
    security:
      branch_protection:
        main:
          required_reviews: 1

repositories:
  # Apply to specific repositories
  specific:
    - name: "api-server"
      description: "Main API server"
      homepage: "https://api.example.com"
      topics: ["api", "microservice", "go"]
      template: "microservice"
      features:
        issues: true
        wiki: false
        projects: true
      security:
        branch_protection:
          main:
            required_reviews: 2
            dismiss_stale_reviews: true
            require_code_owner_reviews: true
            required_status_checks:
              - "ci/build"
              - "ci/test"
      
  # Apply to repositories matching patterns
  patterns:
    - match: "^service-.*"
      template: "microservice"
    - match: "^lib-.*"
      template: "library"
      
  # Apply to all repositories
  default:
    template: "default"
```

### 2. Template System

#### 2.1 Policy Templates
- **Security Template**: High security standards for production repositories
- **Open Source Template**: Standard configuration for public repositories
- **Enterprise Template**: Corporate policies and compliance requirements
- **Library Template**: Configuration for shared libraries
- **Microservice Template**: Standard microservice configuration

#### 2.2 Template Inheritance
- Base templates with common settings
- Environment-specific overrides (dev/staging/prod)
- Organization-specific customizations
- Repository-specific exceptions

#### 2.3 Template Structure
```yaml
templates:
  security:
    base: "enterprise"
    description: "High security template"
    settings:
      features:
        issues: true
        wiki: false
        projects: false
      security:
        branch_protection:
          main:
            required_reviews: 2
            dismiss_stale_reviews: true
            require_code_owner_reviews: true
            restrict_pushes: true
            allowed_users: []
            allowed_teams: ["security-team"]
        webhooks:
          - url: "https://security.example.com/webhook"
            events: ["push", "pull_request"]
            secret: "${SECURITY_WEBHOOK_SECRET}"
```

### 3. Command Interface

#### 3.1 Main Commands
```bash
# Repository configuration management
gz repo-config list                    # List repositories with current settings
gz repo-config apply                   # Apply configuration to repositories
gz repo-config validate               # Validate configuration files
gz repo-config diff                   # Show differences between current and target state
gz repo-config audit                  # Generate compliance audit report

# Template management
gz repo-config template list          # List available templates
gz repo-config template show <name>   # Show template details
gz repo-config template validate <name> # Validate template

# Dry-run and safety features
gz repo-config apply --dry-run        # Preview changes without applying
gz repo-config apply --interactive    # Interactive confirmation for each change
gz repo-config apply --filter "pattern" # Apply to repositories matching pattern
```

#### 3.2 Command Options
- `--org <organization>`: Target specific organization
- `--config <file>`: Use specific configuration file
- `--template <name>`: Apply specific template
- `--dry-run`: Preview changes without applying
- `--interactive`: Interactive mode with confirmations
- `--force`: Skip confirmations (use with caution)
- `--parallel <num>`: Number of parallel operations
- `--timeout <duration>`: API timeout duration

### 4. Safety and Validation

#### 4.1 Pre-flight Checks
- **Permission Validation**: Verify required GitHub permissions
- **Configuration Validation**: Validate YAML syntax and schema
- **Template Validation**: Ensure templates are valid and complete
- **Conflict Detection**: Identify potential conflicts with existing settings
- **Dependency Validation**: Check for required teams, users, or webhooks

#### 4.2 Change Management
- **Change Preview**: Show detailed diff of proposed changes
- **Rollback Capability**: Ability to revert to previous configurations
- **Change History**: Track all configuration changes with timestamps
- **Approval Workflow**: Optional approval process for sensitive changes

#### 4.3 Error Handling
- **Graceful Failures**: Continue processing other repositories if one fails
- **Detailed Error Reporting**: Clear error messages with suggested fixes
- **Retry Logic**: Automatic retry for transient failures
- **Rate Limiting**: Respect GitHub API rate limits

### 5. Authentication and Permissions

#### 5.1 Required GitHub Permissions
- **Repository Administration**: `admin:repo_hook`, `repo`
- **Organization Management**: `admin:org` (for organization-level settings)
- **Team Management**: `admin:org` (for team-based restrictions)

#### 5.2 Authentication Methods
- **Personal Access Tokens**: For individual use
- **GitHub Apps**: For organization-wide deployment
- **Organization Tokens**: For enterprise use

#### 5.3 Security Considerations
- **Token Security**: Secure storage and rotation of access tokens
- **Audit Logging**: Log all configuration changes for security audits
- **Least Privilege**: Request only necessary permissions
- **Multi-factor Authentication**: Require MFA for sensitive operations

### 6. Configuration File Management

#### 6.1 File Structure
```
.gzh/
├── repo-config.yaml          # Main configuration file
├── templates/                 # Template definitions
│   ├── security.yaml
│   ├── opensource.yaml
│   └── enterprise.yaml
├── policies/                  # Policy definitions
│   ├── branch-protection.yaml
│   └── webhook-config.yaml
└── overrides/                 # Environment-specific overrides
    ├── dev.yaml
    ├── staging.yaml
    └── prod.yaml
```

#### 6.2 Configuration Discovery
- Search paths: `./.gzh/`, `~/.config/gzh/`, `/etc/gzh/`
- Environment variables: `GZH_REPO_CONFIG_PATH`
- Command-line flags: `--config`

### 7. Integration and Automation

#### 7.1 CI/CD Integration
- **GitHub Actions**: Workflow for automated configuration management
- **GitOps**: Git-based configuration management workflow
- **Webhook Integration**: Automatic configuration updates on repository creation

#### 7.2 Monitoring and Alerting
- **Configuration Drift Detection**: Alert when repositories drift from policy
- **Compliance Reporting**: Regular compliance status reports
- **Change Notifications**: Notify on configuration changes

### 8. Reporting and Analytics

#### 8.1 Audit Reports
- **Compliance Status**: Which repositories comply with policies
- **Security Posture**: Security configuration analysis
- **Configuration Coverage**: Which repositories have managed configurations

#### 8.2 Metrics and Dashboards
- **Repository Count**: Total managed repositories
- **Policy Compliance**: Percentage of compliant repositories
- **Change Frequency**: Configuration change frequency
- **Error Rates**: Failed configuration attempts

### 9. Use Cases and Scenarios

#### 9.1 Initial Setup
1. **Organization Onboarding**: Apply standard configurations to all repositories
2. **Policy Enforcement**: Ensure all repositories meet security requirements
3. **Template Deployment**: Roll out new organizational standards

#### 9.2 Ongoing Management
1. **New Repository Setup**: Automatically configure new repositories
2. **Policy Updates**: Update security policies across all repositories
3. **Compliance Auditing**: Regular compliance checks and reporting

#### 9.3 Special Scenarios
1. **Repository Migration**: Migrate repositories between organizations
2. **Security Incident Response**: Quickly apply security patches
3. **Compliance Requirements**: Meet regulatory compliance requirements

### 10. Implementation Phases

#### Phase 1: Core Functionality
- Basic repository listing and configuration reading
- Simple configuration application (non-security settings)
- Template system foundation
- CLI interface implementation

#### Phase 2: Security Features
- Branch protection rule management
- Webhook configuration
- Security policy templates
- Permission validation

#### Phase 3: Advanced Features
- Complex template inheritance
- GitOps integration
- Compliance reporting
- Automated drift detection

#### Phase 4: Enterprise Features
- Advanced audit logging
- Multi-organization support
- Custom policy engines
- Integration APIs

### 11. Success Criteria

#### 11.1 Functional Requirements
- [ ] Successfully manage repository configurations for 100+ repositories
- [ ] Apply security policies consistently across organization
- [ ] Reduce manual repository configuration time by 90%
- [ ] Zero configuration drift for critical security settings

#### 11.2 Non-Functional Requirements
- [ ] API rate limit compliance (< 5000 requests/hour)
- [ ] Configuration application within 5 minutes for 100 repositories
- [ ] 99.9% reliability for configuration operations
- [ ] Zero data loss during configuration updates

### 12. Risks and Mitigations

#### 12.1 Technical Risks
- **API Rate Limiting**: Implement intelligent rate limiting and queuing
- **Large Organization Scale**: Optimize for batch operations and parallel processing
- **GitHub API Changes**: Abstract API interactions and version compatibility

#### 12.2 Security Risks
- **Unauthorized Access**: Implement proper authentication and authorization
- **Configuration Errors**: Extensive validation and dry-run capabilities
- **Data Exposure**: Secure handling of sensitive configuration data

#### 12.3 Operational Risks
- **Service Disruption**: Gradual rollout and rollback capabilities
- **User Adoption**: Comprehensive documentation and training
- **Maintenance Overhead**: Automated testing and monitoring

### 13. Alternative Solutions Comparison

#### 13.1 Terraform GitHub Provider
**Pros**: Mature, widely adopted, infrastructure-as-code standard
**Cons**: Complex for simple use cases, requires Terraform knowledge

#### 13.2 GitHub REST API Scripts
**Pros**: Simple, direct control
**Cons**: No standardization, difficult to maintain, error-prone

#### 13.3 GitHub CLI with Scripts
**Pros**: Official tool, comprehensive
**Cons**: Limited batch operations, scripting complexity

#### 13.4 gzh-manager Advantages
- **Integrated**: Part of existing Git management workflow
- **Specialized**: Purpose-built for repository configuration
- **User-Friendly**: Simple YAML configuration format
- **Safe**: Built-in safety features and validation

This requirements document serves as the foundation for implementing comprehensive GitHub repository configuration management in gzh-manager.