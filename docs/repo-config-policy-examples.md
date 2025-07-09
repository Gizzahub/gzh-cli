# Repository Configuration Policy Templates

This document provides ready-to-use policy templates for common scenarios.

## Table of Contents

1. [Security Policies](#security-policies)
2. [Compliance Policies](#compliance-policies)
3. [Open Source Policies](#open-source-policies)
4. [Enterprise Policies](#enterprise-policies)
5. [Development Workflow Policies](#development-workflow-policies)
6. [Complete Examples](#complete-examples)

## Security Policies

### Basic Security Policy

Minimum security requirements for all repositories.

```yaml
policies:
  basic-security:
    description: "Basic security requirements"
    rules:
      vulnerability_alerts:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Vulnerability alerts must be enabled for security scanning"
      
      default_branch_protection:
        type: "branch_protection"
        value: true
        enforcement: "required"
        message: "Default branch must be protected"
```

### Enhanced Security Policy

Stricter security requirements for sensitive repositories.

```yaml
policies:
  enhanced-security:
    description: "Enhanced security for sensitive repositories"
    rules:
      must_be_private:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "Sensitive repositories must be private"
      
      vulnerability_alerts:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Vulnerability alerts are mandatory"
      
      secret_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Secret scanning must be enabled"
      
      automated_security_fixes:
        type: "security_feature"
        value: true
        enforcement: "recommended"
        message: "Automated security fixes are recommended"
      
      branch_protection_reviews:
        type: "branch_protection_setting"
        value: 
          required_reviews: 2
          dismiss_stale_reviews: true
        enforcement: "required"
        message: "Main branch must require 2 reviews"
```

### Zero-Trust Security Policy

Maximum security for critical infrastructure repositories.

```yaml
policies:
  zero-trust:
    description: "Zero-trust security model"
    rules:
      private_only:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "All repositories must be private"
      
      all_security_features:
        type: "security_features"
        value:
          vulnerability_alerts: true
          automated_security_fixes: true
          secret_scanning: true
          secret_scanning_push_protection: true
          dependency_graph: true
          security_advisories: true
        enforcement: "required"
        message: "All security features must be enabled"
      
      strict_branch_protection:
        type: "branch_protection"
        value:
          required_reviews: 3
          dismiss_stale_reviews: true
          require_code_owner_reviews: true
          enforce_admins: true
          require_conversation_resolution: true
          required_signatures: true
        enforcement: "required"
        message: "Strict branch protection required"
      
      no_force_push:
        type: "branch_protection_setting"
        value:
          allow_force_pushes: false
          allow_deletions: false
        enforcement: "required"
        message: "Force pushes and deletions are prohibited"
```

## Compliance Policies

### SOC2 Compliance

Policy for SOC2 compliance requirements.

```yaml
policies:
  soc2-compliance:
    description: "SOC2 compliance requirements"
    rules:
      access_logging:
        type: "audit_log"
        value: true
        enforcement: "required"
        message: "Audit logging must be enabled for SOC2"
      
      code_review_required:
        type: "branch_protection_setting"
        value:
          required_reviews: 2
          require_code_owner_reviews: true
        enforcement: "required"
        message: "Code review is mandatory for SOC2"
      
      vulnerability_management:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Vulnerability scanning required for SOC2"
      
      signed_commits:
        type: "branch_protection_setting"
        value:
          required_signatures: true
        enforcement: "recommended"
        message: "Signed commits recommended for SOC2"
```

### GDPR Compliance

Policy for GDPR data protection requirements.

```yaml
policies:
  gdpr-compliance:
    description: "GDPR data protection compliance"
    rules:
      private_repos:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "GDPR data must be in private repositories"
      
      access_control:
        type: "permissions"
        value:
          max_permission: "push"
          require_2fa: true
        enforcement: "required"
        message: "Strict access control for GDPR compliance"
      
      data_retention:
        type: "retention_policy"
        value:
          max_days: 365
          require_approval_for_extension: true
        enforcement: "required"
        message: "Data retention limits for GDPR"
      
      audit_trail:
        type: "audit_log"
        value: true
        enforcement: "required"
        message: "Audit trail required for GDPR"
```

### HIPAA Compliance

Policy for HIPAA healthcare data protection.

```yaml
policies:
  hipaa-compliance:
    description: "HIPAA compliance for healthcare data"
    rules:
      encryption_at_rest:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "PHI must be in private repositories"
      
      access_controls:
        type: "permissions"
        value:
          require_2fa: true
          max_permission: "push"
          review_required_for_admin: true
        enforcement: "required"
        message: "HIPAA requires strict access controls"
      
      audit_logging:
        type: "audit_log"
        value: true
        enforcement: "required"
        message: "HIPAA requires comprehensive audit logging"
      
      vulnerability_scanning:
        type: "security_features"
        value:
          vulnerability_alerts: true
          secret_scanning: true
          secret_scanning_push_protection: true
        enforcement: "required"
        message: "HIPAA requires vulnerability management"
```

## Open Source Policies

### Basic Open Source

Minimum requirements for open source projects.

```yaml
policies:
  open-source-basic:
    description: "Basic open source project requirements"
    rules:
      must_be_public:
        type: "visibility"
        value: "public"
        enforcement: "required"
        message: "Open source projects must be public"
      
      has_license:
        type: "file_exists"
        value: "LICENSE"
        enforcement: "required"
        message: "Open source projects must have a license"
      
      has_readme:
        type: "file_exists"
        value: "README.md"
        enforcement: "required"
        message: "README is required for open source projects"
      
      community_features:
        type: "settings"
        value:
          has_issues: true
          has_discussions: true
        enforcement: "required"
        message: "Community features must be enabled"
```

### Community-Driven Open Source

Requirements for community-driven projects.

```yaml
policies:
  open-source-community:
    description: "Community-driven open source requirements"
    rules:
      public_visibility:
        type: "visibility"
        value: "public"
        enforcement: "required"
        message: "Community projects must be public"
      
      required_files:
        type: "files_exist"
        value:
          - "LICENSE"
          - "README.md"
          - "CONTRIBUTING.md"
          - "CODE_OF_CONDUCT.md"
        enforcement: "required"
        message: "Community files are required"
      
      community_settings:
        type: "settings"
        value:
          has_issues: true
          has_discussions: true
          has_wiki: true
          has_projects: true
        enforcement: "required"
        message: "All community features must be enabled"
      
      branch_protection:
        type: "branch_protection"
        value:
          required_reviews: 1
          dismiss_stale_reviews: true
        enforcement: "recommended"
        message: "Code review recommended for quality"
```

### Enterprise Open Source

Requirements for enterprise-backed open source.

```yaml
policies:
  enterprise-open-source:
    description: "Enterprise open source standards"
    rules:
      visibility:
        type: "visibility"
        value: "public"
        enforcement: "required"
        message: "Enterprise OSS must be public"
      
      legal_files:
        type: "files_exist"
        value:
          - "LICENSE"
          - "NOTICE"
          - "CONTRIBUTING.md"
          - "CODE_OF_CONDUCT.md"
          - "SECURITY.md"
        enforcement: "required"
        message: "Legal and community files required"
      
      quality_controls:
        type: "branch_protection"
        value:
          required_reviews: 2
          require_code_owner_reviews: true
          required_status_checks:
            - "ci/build"
            - "ci/test"
            - "license/cla"
        enforcement: "required"
        message: "Quality controls are mandatory"
      
      security_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Security scanning required for enterprise OSS"
```

## Enterprise Policies

### Standard Enterprise

Standard requirements for enterprise repositories.

```yaml
policies:
  enterprise-standard:
    description: "Standard enterprise repository requirements"
    rules:
      default_private:
        type: "visibility"
        value: "private"
        enforcement: "recommended"
        message: "Enterprise repos should be private by default"
      
      branch_protection:
        type: "branch_protection"
        value:
          required_reviews: 2
          dismiss_stale_reviews: true
          require_code_owner_reviews: true
        enforcement: "required"
        message: "Code review is mandatory"
      
      ci_integration:
        type: "status_checks"
        value:
          - "continuous-integration"
          - "security-scan"
        enforcement: "required"
        message: "CI/CD integration is required"
      
      documentation:
        type: "file_exists"
        value: "README.md"
        enforcement: "required"
        message: "Documentation is mandatory"
```

### Enterprise Production

Requirements for production repositories.

```yaml
policies:
  enterprise-production:
    description: "Production repository requirements"
    rules:
      private_only:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "Production repos must be private"
      
      strict_protection:
        type: "branch_protection"
        value:
          required_reviews: 3
          dismiss_stale_reviews: true
          require_code_owner_reviews: true
          enforce_admins: false
          restrict_push_access:
            - "release-team"
        enforcement: "required"
        message: "Strict protection for production"
      
      deployment_gates:
        type: "status_checks"
        value:
          - "build"
          - "test"
          - "security-scan"
          - "performance-test"
          - "approval/production"
        enforcement: "required"
        message: "All deployment gates must pass"
      
      change_management:
        type: "pull_request_settings"
        value:
          require_conversation_resolution: true
          delete_branch_on_merge: true
        enforcement: "required"
        message: "Change management controls required"
```

## Development Workflow Policies

### Agile Development

Policy supporting agile development practices.

```yaml
policies:
  agile-development:
    description: "Agile development workflow"
    rules:
      feature_branches:
        type: "branch_protection"
        value:
          pattern: "feature/*"
          required_reviews: 1
          allow_force_pushes: true
        enforcement: "recommended"
        message: "Feature branches should have light protection"
      
      main_protection:
        type: "branch_protection"
        value:
          required_reviews: 2
          required_status_checks:
            - "ci/build"
            - "ci/test"
        enforcement: "required"
        message: "Main branch must be protected"
      
      project_management:
        type: "settings"
        value:
          has_issues: true
          has_projects: true
          has_wiki: true
        enforcement: "required"
        message: "Project management features required"
```

### GitFlow Workflow

Policy enforcing GitFlow branching model.

```yaml
policies:
  gitflow-workflow:
    description: "GitFlow branching model"
    rules:
      branch_structure:
        type: "branch_rules"
        value:
          protected_branches:
            - "main"
            - "develop"
            - "release/*"
            - "hotfix/*"
        enforcement: "required"
        message: "GitFlow branches must exist and be protected"
      
      main_branch:
        type: "branch_protection"
        value:
          branch: "main"
          required_reviews: 3
          enforce_admins: true
          allow_force_pushes: false
        enforcement: "required"
        message: "Main branch requires strict protection"
      
      develop_branch:
        type: "branch_protection"
        value:
          branch: "develop"
          required_reviews: 2
          required_status_checks:
            - "ci/build"
            - "ci/test"
        enforcement: "required"
        message: "Develop branch requires protection"
      
      release_branches:
        type: "branch_protection"
        value:
          pattern: "release/*"
          required_reviews: 2
          restrict_push_access:
            - "release-managers"
        enforcement: "required"
        message: "Release branches require approval"
```

## Complete Examples

### Startup Configuration

Complete configuration for a startup.

```yaml
version: "1.0.0"
organization: "awesome-startup"

templates:
  default:
    description: "Default startup repository"
    settings:
      has_issues: true
      has_wiki: true
      delete_branch_on_merge: true
    security:
      vulnerability_alerts: true
      branch_protection:
        main:
          required_reviews: 1
          required_status_checks:
            - "ci/test"

  production:
    base: "default"
    description: "Production services"
    settings:
      private: true
    security:
      secret_scanning: true
      branch_protection:
        main:
          required_reviews: 2
          enforce_admins: true

policies:
  startup-baseline:
    description: "Baseline requirements"
    rules:
      has_readme:
        type: "file_exists"
        value: "README.md"
        enforcement: "required"
        message: "All repos need documentation"
      
      basic_protection:
        type: "branch_protection"
        value: true
        enforcement: "required"
        message: "Main branch must be protected"

patterns:
  - pattern: "*-service"
    template: "production"
    policies: ["startup-baseline"]
  - pattern: "*-prototype"
    template: "default"
    policies: ["startup-baseline"]
```

### Enterprise Configuration

Complete configuration for an enterprise.

```yaml
version: "1.0.0"
organization: "bigcorp"

templates:
  base:
    description: "Base configuration"
    settings:
      private: true
      has_issues: true
      delete_branch_on_merge: true
    security:
      vulnerability_alerts: true
      secret_scanning: true

  microservice:
    base: "base"
    description: "Microservice template"
    security:
      branch_protection:
        main:
          required_reviews: 2
          required_status_checks:
            - "build"
            - "test"
            - "sonarqube"
    webhooks:
      - name: "jenkins"
        url: "${JENKINS_URL}/github-webhook/"
        events: ["push", "pull_request"]

  frontend:
    base: "base"
    description: "Frontend application"
    settings:
      has_pages: true
    security:
      branch_protection:
        main:
          required_reviews: 2
          required_status_checks:
            - "build"
            - "test"
            - "lighthouse"

  data-service:
    base: "base"
    description: "Data service with strict security"
    security:
      secret_scanning_push_protection: true
      branch_protection:
        main:
          required_reviews: 3
          require_code_owner_reviews: true
          enforce_admins: false
          restrict_push_access:
            - "data-team"

policies:
  security-baseline:
    description: "Security baseline for all repos"
    rules:
      private_default:
        type: "visibility"
        value: "private"
        enforcement: "recommended"
        message: "Repositories should be private"
      
      vulnerability_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Vulnerability scanning is mandatory"

  compliance:
    description: "Compliance requirements"
    rules:
      code_review:
        type: "branch_protection_setting"
        value:
          required_reviews: 2
        enforcement: "required"
        message: "Code review is required for compliance"
      
      audit_log:
        type: "audit_log"
        value: true
        enforcement: "required"
        message: "Audit logging required"

  production:
    description: "Production environment requirements"
    rules:
      strict_protection:
        type: "branch_protection"
        value:
          required_reviews: 3
          enforce_admins: false
        enforcement: "required"
        message: "Production requires strict controls"
      
      deployment_approvals:
        type: "environment_protection"
        value:
          required_reviewers: 2
          deployment_branch_policy: "protected"
        enforcement: "required"
        message: "Production deployments need approval"

patterns:
  - pattern: "*-service"
    template: "microservice"
    policies: ["security-baseline", "compliance"]
  
  - pattern: "*-ui"
    template: "frontend"
    policies: ["security-baseline"]
  
  - pattern: "*-data"
    template: "data-service"
    policies: ["security-baseline", "compliance", "production"]
  
  - pattern: "prod-*"
    policies: ["production"]

repositories:
  - name: "customer-data-service"
    template: "data-service"
    policies: ["security-baseline", "compliance", "production"]
    exceptions:
      - policy: "compliance"
        rule: "audit_log"
        reason: "Legacy system - migration planned Q2 2024"
        approved_by: "cto@bigcorp.com"
        expires_at: "2024-06-30"
```

## Usage Tips

### 1. Start Simple

Begin with basic policies and gradually add complexity:

```yaml
# Start with this
policies:
  basic:
    rules:
      has_readme:
        type: "file_exists"
        value: "README.md"
        enforcement: "required"

# Then add more rules as needed
```

### 2. Use Template Inheritance

Build complex templates from simple ones:

```yaml
templates:
  base:
    settings:
      has_issues: true
  
  secure:
    base: "base"
    settings:
      private: true
  
  production:
    base: "secure"
    security:
      secret_scanning: true
```

### 3. Document Exceptions

Always document why exceptions exist:

```yaml
exceptions:
  - policy: "must-be-private"
    rule: "visibility"
    reason: "Public API documentation"
    approved_by: "security-team"
    expires_at: "2024-12-31"
```

### 4. Regular Reviews

Schedule regular policy reviews:
- Monthly: Review exceptions
- Quarterly: Update policies
- Annually: Major policy revision

### 5. Gradual Enforcement

Start with `recommended` and move to `required`:

```yaml
# Phase 1: Recommended
enforcement: "recommended"

# Phase 2: Required with exceptions
enforcement: "required"
# Allow exceptions for transition

# Phase 3: Strict enforcement
enforcement: "required"
# No exceptions
```