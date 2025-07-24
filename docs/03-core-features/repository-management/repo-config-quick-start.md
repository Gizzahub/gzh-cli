# Repository Configuration Quick Start Guide

Get started with `gz repo-config` in 5 minutes!

## Prerequisites

1. Install gzh-manager:

   ```bash
   go install github.com/gizzahub/gzh-manager-go@latest
   ```

2. Set your GitHub token:
   ```bash
   export GITHUB_TOKEN="ghp_your_token_here"
   ```

## 1. Basic Configuration (1 minute)

Create a simple configuration to standardize your repositories:

```yaml
# repo-config.yaml
version: "1.0.0"
organization: "your-org"

templates:
  standard:
    description: "Standard repository settings"
    settings:
      has_issues: true
      has_wiki: false
      delete_branch_on_merge: true
    security:
      vulnerability_alerts: true

repositories:
  - name: "*"
    template: "standard"
```

Apply it:

```bash
gz repo-config apply --config repo-config.yaml --dry-run
```

## 2. Add Security (2 minutes)

Enhance with security policies:

```yaml
# repo-config.yaml
version: "1.0.0"
organization: "your-org"

templates:
  standard:
    description: "Standard repository settings"
    settings:
      has_issues: true
      has_wiki: false
      delete_branch_on_merge: true
    security:
      vulnerability_alerts: true
      branch_protection:
        main:
          required_reviews: 2
          enforce_admins: true

policies:
  security:
    description: "Basic security requirements"
    rules:
      branch_protection:
        type: "branch_protection"
        value: true
        enforcement: "required"
        message: "Main branch must be protected"

repositories:
  - name: "*"
    template: "standard"
    policies: ["security"]
```

Check compliance:

```bash
gz repo-config audit --config repo-config.yaml
```

## 3. Different Repository Types (3 minutes)

Configure different templates for different repository types:

```yaml
# repo-config.yaml
version: "1.0.0"
organization: "your-org"

templates:
  backend:
    description: "Backend service configuration"
    settings:
      private: true
      has_issues: true
    security:
      secret_scanning: true
      vulnerability_alerts: true
      branch_protection:
        main:
          required_reviews: 2
          required_status_checks:
            - "ci/build"
            - "ci/test"

  frontend:
    description: "Frontend application configuration"
    settings:
      has_issues: true
      has_pages: true
    security:
      vulnerability_alerts: true
      branch_protection:
        main:
          required_reviews: 1

  documentation:
    description: "Documentation repository"
    settings:
      private: false
      has_issues: true
      has_wiki: true
      has_pages: true

patterns:
  - pattern: "*-api"
    template: "backend"
  - pattern: "*-service"
    template: "backend"
  - pattern: "*-ui"
    template: "frontend"
  - pattern: "*-web"
    template: "frontend"
  - pattern: "*-docs"
    template: "documentation"

# Specific repository overrides
repositories:
  - name: "public-api-docs"
    template: "documentation"
    settings:
      private: false # Override to make public
```

## 4. Add Required Files (4 minutes)

Ensure all repositories have necessary files:

```yaml
# repo-config.yaml
version: "1.0.0"
organization: "your-org"

templates:
  standard:
    description: "Standard repository settings"
    settings:
      has_issues: true
      delete_branch_on_merge: true
    required_files:
      - path: "README.md"
        content: |
          # Repository Name

          ## Description
          Add your description here

          ## Getting Started

          ## Contributing
          Please read CONTRIBUTING.md

      - path: ".github/CODEOWNERS"
        content: |
          # Global owners
          * @your-org/dev-team

      - path: ".gitignore"
        content: |
          # IDE
          .idea/
          .vscode/
          *.swp

          # OS
          .DS_Store
          Thumbs.db

          # Build
          dist/
          build/
          *.log

repositories:
  - name: "*"
    template: "standard"
```

## 5. Complete Example (5 minutes)

Here's a complete configuration for a typical organization:

```yaml
# repo-config.yaml
version: "1.0.0"
organization: "awesome-corp"

# Define reusable templates
templates:
  base:
    description: "Base settings for all repositories"
    settings:
      has_issues: true
      delete_branch_on_merge: true
    security:
      vulnerability_alerts: true

  private-service:
    base: "base"
    description: "Private microservice"
    settings:
      private: true
    security:
      secret_scanning: true
      branch_protection:
        main:
          required_reviews: 2
          enforce_admins: true
          required_status_checks:
            - "ci/build"
            - "ci/test"

  public-library:
    base: "base"
    description: "Public library or SDK"
    settings:
      private: false
      has_wiki: true
    required_files:
      - path: "LICENSE"
        content: |
          MIT License
          Copyright (c) 2024 Awesome Corp
      - path: "CONTRIBUTING.md"
        content: |
          # Contributing
          We welcome contributions!

# Define policies
policies:
  production:
    description: "Production-ready requirements"
    rules:
      must_be_private:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "Production services must be private"

      security_scanning:
        type: "security_feature"
        value: true
        enforcement: "required"
        message: "Security scanning is mandatory"

# Apply templates based on patterns
patterns:
  - pattern: "*-service"
    template: "private-service"
    policies: ["production"]

  - pattern: "*-sdk"
    template: "public-library"

  - pattern: "*-client"
    template: "public-library"

# Handle exceptions
repositories:
  - name: "status-page"
    template: "private-service"
    exceptions:
      - policy: "production"
        rule: "must_be_private"
        reason: "Public status page for customers"
        approved_by: "cto@awesome-corp.com"
        expires_at: "2025-01-01"
```

## Common Commands

### Check what would change:

```bash
gz repo-config apply --config repo-config.yaml --dry-run
```

### Apply configuration:

```bash
gz repo-config apply --config repo-config.yaml
```

### Check compliance:

```bash
gz repo-config audit --config repo-config.yaml
```

### Generate HTML report:

```bash
gz repo-config audit --config repo-config.yaml --format html --output report.html
```

### Apply to specific repositories:

```bash
gz repo-config apply --config repo-config.yaml --repos "repo1,repo2"
```

## Next Steps

1. **Customize Templates**: Modify the templates to match your organization's standards
2. **Add More Policies**: Define policies for compliance, security, and quality
3. **Set Up CI/CD**: Add compliance checks to your CI/CD pipeline
4. **Monitor Compliance**: Schedule regular audits

## Tips

- Start with a few repositories to test your configuration
- Use `--dry-run` to preview changes before applying
- Keep your configuration in version control
- Document why exceptions exist
- Review and update policies regularly

## Need Help?

- Run `gz repo-config --help` for all options
- Check the [full user guide](./repo-config-user-guide.md)
- See [policy examples](./repo-config-policy-examples.md)
- View [sample configurations](../examples/)
