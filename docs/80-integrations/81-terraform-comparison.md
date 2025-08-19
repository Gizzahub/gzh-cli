# Terraform vs gz repo-config: GitHub Repository Management Comparison

This document compares Terraform's GitHub provider with gzh-manager's `gz repo-config` feature for managing GitHub repository configurations.

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Feature Comparison](#feature-comparison)
3. [Architecture Comparison](#architecture-comparison)
4. [Use Case Analysis](#use-case-analysis)
5. [Migration Guide](#migration-guide)
6. [Recommendations](#recommendations)

## Executive Summary

Both Terraform and `gz repo-config` can manage GitHub repository configurations as code, but they serve different use cases and have distinct strengths:

- **Terraform**: General-purpose infrastructure-as-code tool with GitHub provider, best for mixed infrastructure management
- **gz repo-config**: Purpose-built for GitHub repository management, optimized for bulk operations and compliance

### Key Differences

| Aspect                  | Terraform                                | gz repo-config                      |
| ----------------------- | ---------------------------------------- | ----------------------------------- |
| **Primary Purpose**     | General infrastructure management        | GitHub repository management        |
| **Learning Curve**      | Steeper (HCL syntax, Terraform concepts) | Gentler (YAML, focused feature set) |
| **State Management**    | Required (backend configuration)         | Stateless (queries GitHub directly) |
| **Bulk Operations**     | Resource-by-resource                     | Optimized for bulk operations       |
| **Policy Enforcement**  | External tools needed                    | Built-in policy engine              |
| **Compliance Auditing** | Third-party tools                        | Built-in audit reports              |

## Feature Comparison

### Repository Configuration

#### Terraform GitHub Provider

```hcl
resource "github_repository" "example" {
  name        = "example-repo"
  description = "Example repository"
  visibility  = "private"

  has_issues   = true
  has_wiki     = false
  has_projects = false

  allow_merge_commit     = true
  allow_squash_merge     = true
  allow_rebase_merge     = false
  delete_branch_on_merge = true

  topics = ["terraform", "example"]
}

resource "github_branch_protection" "example" {
  repository_id = github_repository.example.node_id
  pattern       = "main"

  required_status_checks {
    strict   = true
    contexts = ["ci/build", "ci/test"]
  }

  required_pull_request_reviews {
    dismiss_stale_reviews      = true
    require_code_owner_reviews = true
    required_approving_review_count = 2
  }
}
```

#### gz repo-config

```yaml
version: "1.0.0"
organization: "my-org"

templates:
  standard:
    description: "Standard repository configuration"
    settings:
      private: true
      has_issues: true
      has_wiki: false
      has_projects: false
      allow_merge_commit: true
      allow_squash_merge: true
      allow_rebase_merge: false
      delete_branch_on_merge: true
    topics: ["managed", "standard"]
    security:
      branch_protection:
        main:
          required_reviews: 2
          dismiss_stale_reviews: true
          require_code_owner_reviews: true
          required_status_checks: ["ci/build", "ci/test"]
          strict_status_checks: true

repositories:
  - name: "example-repo"
    template: "standard"
    description: "Example repository"
    topics: ["example"] # Merged with template topics
```

### Feature Support Matrix

| Feature                            | Terraform     | gz repo-config          |
| ---------------------------------- | ------------- | ----------------------- |
| **Basic Settings**                 |
| Repository creation                | ✅            | ❌ (configuration only) |
| Visibility control                 | ✅            | ✅                      |
| Feature flags (issues, wiki, etc.) | ✅            | ✅                      |
| Default branch                     | ✅            | ✅                      |
| Topics                             | ✅            | ✅                      |
| **Security**                       |
| Branch protection                  | ✅            | ✅                      |
| Required status checks             | ✅            | ✅                      |
| Review requirements                | ✅            | ✅                      |
| Push restrictions                  | ✅            | ✅                      |
| Vulnerability alerts               | ✅            | ✅                      |
| Secret scanning                    | ✅            | ✅                      |
| **Advanced Features**              |
| Webhooks                           | ✅            | ✅                      |
| Deploy keys                        | ✅            | ❌ (planned)            |
| Environments                       | ✅            | ❌ (planned)            |
| GitHub Actions permissions         | ✅            | ✅                      |
| **Management Features**            |
| Templates/Modules                  | ✅ (modules)  | ✅ (templates)          |
| Bulk operations                    | ❌ (foreach)  | ✅ (native)             |
| Policy enforcement                 | ❌ (external) | ✅ (built-in)           |
| Compliance auditing                | ❌            | ✅                      |
| Dry run                            | ✅ (plan)     | ✅                      |
| State management                   | ✅ (required) | ❌ (stateless)          |
| Import existing                    | ✅            | ✅ (automatic)          |

## Architecture Comparison

### Terraform Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   HCL Code  │────▶│  Terraform   │────▶│   GitHub    │
│   (.tf)     │     │    Core      │     │     API     │
└─────────────┘     └──────────────┘     └─────────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │ State File   │
                    │ (backend)    │
                    └──────────────┘
```

**Characteristics:**

- Requires state management
- Tracks resource lifecycle
- Supports cross-provider dependencies
- Plan/apply workflow

### gz repo-config Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  YAML Config│────▶│ gz repo-config│────▶│   GitHub    │
│   (.yaml)   │     │    Engine    │     │     API     │
└─────────────┘     └──────────────┘     └─────────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │ Policy Engine│
                    │ (built-in)   │
                    └──────────────┘
```

**Characteristics:**

- Stateless operation
- Direct GitHub API queries
- Built-in policy engine
- Audit-first approach

## Use Case Analysis

### When to Use Terraform

1. **Mixed Infrastructure Management**

   ```hcl
   # Managing GitHub alongside other providers
   resource "github_repository" "app" {
     name = "my-app"
   }

   resource "aws_ecr_repository" "app" {
     name = "my-app"
   }

   resource "kubernetes_namespace" "app" {
     metadata {
       name = "my-app"
     }
   }
   ```

2. **Complex Dependencies**

   ```hcl
   # Repository depends on team creation
   resource "github_team" "developers" {
     name = "developers"
   }

   resource "github_repository" "app" {
     name = "my-app"
   }

   resource "github_team_repository" "developers_app" {
     team_id    = github_team.developers.id
     repository = github_repository.app.name
     permission = "push"
   }
   ```

3. **Infrastructure Provisioning**
   - Creating new repositories
   - Setting up complete GitHub organization structure
   - Managing teams and memberships

### When to Use gz repo-config

1. **Bulk Repository Management**

   ```yaml
   # Apply settings to hundreds of repos
   patterns:
     - pattern: "*-service"
       template: "microservice"
     - pattern: "*-lib"
       template: "library"
     - pattern: "frontend-*"
       template: "frontend"
   ```

2. **Policy Enforcement**

   ```yaml
   policies:
     security-baseline:
       rules:
         vulnerability_alerts:
           type: "security_feature"
           value: true
           enforcement: "required"
         branch_protection:
           type: "branch_protection"
           value: true
           enforcement: "required"
   ```

3. **Compliance Auditing**

   ```bash
   # Generate compliance reports
   gz repo-config audit --format html --output compliance-report.html
   ```

4. **Template-Based Standardization**
   ```yaml
   templates:
     production:
       base: "standard"
       security:
         secret_scanning: true
         branch_protection:
           main:
             required_reviews: 3
   ```

## Migration Guide

### From Terraform to gz repo-config

1. **Export Current State**

   ```bash
   # List all managed repositories
   terraform state list | grep github_repository

   # Generate configuration
   gz repo-config generate --from-github --org my-org
   ```

2. **Convert HCL to YAML**

   Terraform:

   ```hcl
   resource "github_repository" "example" {
     name       = "example"
     visibility = "private"
     has_issues = true
   }
   ```

   gz repo-config:

   ```yaml
   repositories:
     - name: "example"
       settings:
         private: true
         has_issues: true
   ```

3. **Apply Configuration**

   ```bash
   # Dry run first
   gz repo-config apply --dry-run

   # Apply changes
   gz repo-config apply
   ```

### From gz repo-config to Terraform

1. **Generate Terraform Code**

   ```bash
   # For each repository in config
   for repo in $(gz repo-config list); do
     cat > ${repo}.tf <<EOF
   resource "github_repository" "${repo}" {
     name = "${repo}"
     # ... settings
   }
   EOF
   done
   ```

2. **Import Existing Resources**
   ```bash
   terraform import github_repository.example my-org/example
   ```

## Recommendations

### Choose Terraform When:

1. **You need full infrastructure lifecycle management**
   - Creating and destroying repositories
   - Managing organization structure
   - Cross-provider dependencies

2. **You have existing Terraform infrastructure**
   - Consistent tooling across infrastructure
   - Shared modules and workflows
   - Team expertise in Terraform

3. **You need advanced state management**
   - Tracking infrastructure changes
   - Managing resource dependencies
   - Supporting multiple environments

### Choose gz repo-config When:

1. **You manage many existing repositories**
   - Bulk configuration updates
   - Template-based standardization
   - Pattern-based configuration

2. **Compliance and security are priorities**
   - Built-in policy engine
   - Automated compliance auditing
   - Exception management

3. **You want simplicity**
   - No state management
   - Simple YAML configuration
   - Focused on GitHub only

### Hybrid Approach

You can use both tools together:

```yaml
# Terraform for infrastructure
resource "github_repository" "new_service" {
  name = "payment-service"
  # Basic creation only
}

# gz repo-config for configuration management
repositories:
  - name: "payment-service"
    template: "production-service"
    policies: ["security", "compliance"]
```

## Cost Comparison

| Aspect                  | Terraform                 | gz repo-config |
| ----------------------- | ------------------------- | -------------- |
| **Tool Cost**           | Free (OSS) / Paid (Cloud) | Free (OSS)     |
| **State Storage**       | Required (S3, etc.)       | None           |
| **Learning Investment** | High                      | Low            |
| **Maintenance**         | State management overhead | Minimal        |

## Conclusion

Both tools are valuable for different scenarios:

- **Terraform**: Best for infrastructure provisioning and when GitHub is part of a larger infrastructure ecosystem
- **gz repo-config**: Best for managing existing repositories, enforcing policies, and compliance auditing

For most GitHub-focused teams managing existing repositories, `gz repo-config` provides a simpler, more focused solution. For teams already using Terraform for broader infrastructure, the GitHub provider maintains consistency.

Consider your specific needs:

- Repository lifecycle (creation vs configuration)
- Scale (few vs many repositories)
- Compliance requirements
- Team expertise
- Existing tooling

The tools can also complement each other, with Terraform handling provisioning and gz repo-config managing ongoing configuration and compliance.
