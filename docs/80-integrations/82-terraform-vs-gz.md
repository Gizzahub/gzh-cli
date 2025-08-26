# Practical Examples: Terraform vs gz repo-config

This document provides side-by-side comparisons of common GitHub management tasks using both Terraform and gz repo-config.

## Common Scenarios

### 1. Standardize 50 Microservice Repositories

**Terraform Approach:**

```hcl
# Define locals for common settings
locals {
  microservice_repos = [
    "user-service",
    "payment-service",
    "order-service",
    # ... 47 more
  ]

  common_topics = ["microservice", "backend", "production"]
}

# Create module for standard configuration
module "microservice_repo" {
  source = "./modules/github-repo"

  for_each = toset(local.microservice_repos)

  name        = each.key
  visibility  = "private"
  topics      = local.common_topics

  has_issues   = true
  has_wiki     = false
  has_projects = false

  branch_protection = {
    pattern = "main"
    required_reviews = 2
    required_checks = ["ci/build", "ci/test", "security/scan"]
  }
}

# Requires running for each repo
# terraform plan
# terraform apply
# Time: ~30 minutes for 50 repos (with rate limiting)
```

**gz repo-config Approach:**

```yaml
version: "1.0.0"
organization: "my-company"

templates:
  microservice:
    description: "Standard microservice configuration"
    settings:
      private: true
      has_issues: true
      has_wiki: false
      has_projects: false
    topics: ["microservice", "backend", "production"]
    security:
      branch_protection:
        main:
          required_reviews: 2
          required_status_checks: ["ci/build", "ci/test", "security/scan"]

patterns:
  - pattern: "*-service"
    template: "microservice"
# Single command
# gz repo-config apply --config microservice-standard.yaml
# Time: ~5 minutes for 50 repos (parallel operations)
```

### 2. Enforce Security Policy Across Organization

**Terraform Approach:**

```hcl
# No built-in policy enforcement
# Must use external tools like OPA or Sentinel

# Example with Sentinel (Terraform Cloud/Enterprise only)
policy "github-security" {
  enforcement_level = "hard-mandatory"

  rule {
    condition = all github_repository.* as r {
      r.visibility == "private" and
      r.vulnerability_alerts_enabled == true
    }

    error_message = "All repositories must be private with vulnerability alerts"
  }
}

# Or use data sources to check compliance
data "github_repositories" "all" {
  query = "org:my-company"
}

locals {
  non_compliant = [
    for repo in data.github_repositories.all.names :
    repo if !data.github_repository.check[repo].private
  ]
}

output "non_compliant_repos" {
  value = local.non_compliant
}
```

**gz repo-config Approach:**

```yaml
version: "1.0.0"
organization: "my-company"

policies:
  security-baseline:
    description: "Mandatory security requirements"
    rules:
      must_be_private:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "All repositories must be private"

      vulnerability_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Vulnerability alerts must be enabled"

      branch_protection:
        type: "branch_protection"
        value: true
        enforcement: "required"
        message: "Default branch must be protected"

patterns:
  - pattern: "*"
    policies: ["security-baseline"]
# Check compliance
# gz repo-config audit --config security-policy.yaml

# Output:
# Compliance Report:
# - Total repositories: 150
# - Compliant: 142 (94.7%)
# - Non-compliant: 8 (5.3%)
#
# Violations:
# - api-gateway: missing branch protection
# - legacy-service: repository is public
# ...
```

### 3. Different Configurations by Repository Type

**Terraform Approach:**

```hcl
# Define different modules
module "backend_repos" {
  source = "./modules/backend-repo"

  for_each = {
    "user-service"    = { language = "go" }
    "payment-service" = { language = "java" }
    "order-service"   = { language = "python" }
  }

  name     = each.key
  language = each.value.language
}

module "frontend_repos" {
  source = "./modules/frontend-repo"

  for_each = toset(["web-app", "mobile-app", "admin-portal"])

  name = each.key
  has_pages = true
}

module "library_repos" {
  source = "./modules/library-repo"

  for_each = toset(["common-utils", "auth-lib", "logging-lib"])

  name = each.key
  visibility = "public"
}

# Each module has its own configuration
# Total files: 4 main + 3 modules = 7 files
# Lines of code: ~300-400
```

**gz repo-config Approach:**

```yaml
version: "1.0.0"
organization: "my-company"

templates:
  backend:
    settings:
      private: true
      has_issues: true
    security:
      secret_scanning: true
      branch_protection:
        main:
          required_reviews: 2
          required_status_checks: ["ci/build", "ci/test"]

  frontend:
    settings:
      has_issues: true
      has_pages: true
    security:
      branch_protection:
        main:
          required_reviews: 1
          required_status_checks: ["ci/build", "ci/test", "lighthouse"]

  library:
    settings:
      private: false
      has_issues: true
      has_wiki: true
    required_files:
      - path: "LICENSE"
        content: "MIT License..."
      - path: "CONTRIBUTING.md"
        content: "# Contributing..."

patterns:
  - pattern: "*-service"
    template: "backend"
  - pattern: "*-app"
    template: "frontend"
  - pattern: "*-lib"
    template: "library"
# Single file: 50 lines
# One command to apply all
```

### 4. Handle Exceptions

**Terraform Approach:**

```hcl
# Must handle each exception individually
resource "github_repository" "public_docs" {
  name       = "public-docs"
  visibility = "public"  # Exception to private-only policy

  # No built-in way to document why this is an exception
}

# Or use conditional logic
resource "github_repository" "standard" {
  for_each = toset(var.repository_names)

  name       = each.key
  visibility = contains(var.public_exceptions, each.key) ? "public" : "private"
}

variable "public_exceptions" {
  default = ["public-docs", "oss-toolkit"]
}
```

**gz repo-config Approach:**

```yaml
policies:
  private-only:
    rules:
      must_be_private:
        type: "visibility"
        value: "private"
        enforcement: "required"

repositories:
  - name: "public-docs"
    template: "standard"
    settings:
      private: false # Override
    exceptions:
      - policy: "private-only"
        rule: "must_be_private"
        reason: "Public documentation site required for customers"
        approved_by: "security-team@company.com"
        expires_at: "2024-12-31"
# Exceptions are documented and auditable
# gz repo-config audit shows exceptions
```

### 5. Gradual Rollout

**Terraform Approach:**

```hcl
# Use workspace or variables for gradual rollout
variable "enable_new_settings" {
  default = false
}

variable "pilot_repos" {
  default = ["test-service", "staging-service"]
}

resource "github_repository" "managed" {
  for_each = var.repository_names

  name = each.key

  # Complex conditional logic
  delete_branch_on_merge = var.enable_new_settings || contains(var.pilot_repos, each.key)

  dynamic "branch_protection" {
    for_each = var.enable_new_settings || contains(var.pilot_repos, each.key) ? [1] : []

    content {
      # New protection rules
    }
  }
}
```

**gz repo-config Approach:**

```yaml
# Phase 1: Pilot repos only
repositories:
  - name: "test-service"
    template: "new-standard"
  - name: "staging-service"
    template: "new-standard"

# Phase 2: Apply to pattern
patterns:
  - pattern: "*-dev"
    template: "new-standard"

# Phase 3: Organization-wide
patterns:
  - pattern: "*"
    template: "new-standard"

# Easy to see what's applied where
# No complex conditionals
```

### 6. Compliance Reporting

**Terraform Approach:**

```hcl
# No built-in compliance reporting
# Must build custom solution

# Example: Output non-compliant repos
data "github_repositories" "all" {
  query = "org:my-company"
}

data "github_repository" "details" {
  for_each = toset(data.github_repositories.all.names)
  name     = each.key
}

locals {
  compliance_report = {
    for name, repo in data.github_repository.details : name => {
      compliant = repo.private && repo.vulnerability_alerts_enabled
      issues = concat(
        repo.private ? [] : ["not private"],
        repo.vulnerability_alerts_enabled ? [] : ["vulnerability alerts disabled"]
      )
    }
  }

  non_compliant = {
    for name, status in local.compliance_report :
    name => status.issues if !status.compliant
  }
}

output "compliance_summary" {
  value = {
    total = length(local.compliance_report)
    compliant = length([for r in local.compliance_report : r if r.compliant])
    violations = local.non_compliant
  }
}
```

**gz repo-config Approach:**

```yaml
# Built-in compliance reporting
policies:
  production-ready:
    rules:
      private:
        type: "visibility"
        value: "private"
        enforcement: "required"
      security_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"
      branch_protection:
        type: "branch_protection"
        value: true
        enforcement: "required"
# Single command for full report
# gz repo-config audit --format html --output report.html

# Multiple output formats
# gz repo-config audit --format json | jq '.summary'
# gz repo-config audit --format csv > compliance.csv
```

## Performance Comparison

### Bulk Operations (100 repositories)

**Terraform:**

```bash
# Sequential API calls
# Time: ~45-60 minutes
# API calls: ~300-400 (multiple per repo)
# Rate limit issues: Likely
```

**gz repo-config:**

```bash
# Parallel operations with built-in rate limiting
# Time: ~5-10 minutes
# API calls: ~150-200 (optimized)
# Rate limit issues: Handled automatically
```

## Developer Experience

### Terraform Workflow:

```bash
# 1. Write HCL code
# 2. Initialize
terraform init

# 3. Plan changes
terraform plan -out=plan.tfplan

# 4. Review plan (often lengthy)
# 5. Apply changes
terraform apply plan.tfplan

# 6. Handle state conflicts
terraform state lock
terraform state unlock

# 7. Import existing resources
terraform import github_repository.example org/example
```

### gz repo-config Workflow:

```bash
# 1. Write YAML config
# 2. Validate
gz repo-config validate

# 3. Preview changes
gz repo-config apply --dry-run

# 4. Apply
gz repo-config apply

# No state management needed
# Automatic discovery of existing repos
```

## Summary

| Use Case | Better Tool | Why |
| ----------------------------- | -------------- | ----------------------------- |
| Create new repos | Terraform | Manages full lifecycle |
| Configure 100+ existing repos | gz repo-config | Bulk operations, no state |
| Enforce security policies | gz repo-config | Built-in policy engine |
| Mixed infrastructure | Terraform | Single tool for all resources |
| Compliance auditing | gz repo-config | Native audit features |
| Team collaboration | Both work | Different trade-offs |
| Simple standardization | gz repo-config | Easier to learn and use |
| Complex dependencies | Terraform | Better dependency management |
