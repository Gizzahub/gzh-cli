# Repository Configuration API

This document describes the GitHub repository configuration API implementation in gzh-manager.

## Overview

The repository configuration API provides comprehensive access to GitHub repository settings, including:
- Basic repository metadata
- Feature settings (issues, wiki, projects)
- Security settings (branch protection, vulnerability alerts)
- Team and user permissions
- Merge options and policies

## API Client Methods

### GetRepository

Retrieves basic repository information.

```go
repo, err := client.GetRepository(ctx, "owner", "repo")
```

Returns:
- Repository name, description, homepage
- Visibility (public/private)
- Default branch
- Feature flags (has_issues, has_wiki, etc.)
- Merge settings

### GetRepositoryConfiguration

Retrieves comprehensive repository configuration including branch protection and permissions.

```go
config, err := client.GetRepositoryConfiguration(ctx, "owner", "repo")
```

Returns a `RepositoryConfig` structure containing:
- Basic repository settings
- Branch protection rules for default branch
- Team and user permissions
- All feature settings

### GetBranchProtection

Retrieves branch protection rules for a specific branch.

```go
protection, err := client.GetBranchProtection(ctx, "owner", "repo", "main")
```

Returns:
- Required status checks
- Pull request review requirements
- Push restrictions
- Admin enforcement settings

### GetRepositoryPermissions

Retrieves team and user permissions for a repository.

```go
teamPerms, userPerms, err := client.GetRepositoryPermissions(ctx, "owner", "repo")
```

Returns:
- Map of team slugs to permission levels
- Map of usernames to permission levels

## Data Structures

### RepositoryConfig

Complete repository configuration:

```go
type RepositoryConfig struct {
    Name             string
    Description      string
    Homepage         string
    Private          bool
    Archived         bool
    Topics           []string
    Settings         RepoConfigSettings
    BranchProtection map[string]BranchProtectionConfig
    Permissions      PermissionsConfig
}
```

### RepoConfigSettings

Repository feature settings:

```go
type RepoConfigSettings struct {
    HasIssues           bool
    HasProjects         bool
    HasWiki             bool
    HasDownloads        bool
    AllowSquashMerge    bool
    AllowMergeCommit    bool
    AllowRebaseMerge    bool
    DeleteBranchOnMerge bool
    DefaultBranch       string
}
```

### BranchProtectionConfig

Branch protection configuration:

```go
type BranchProtectionConfig struct {
    RequiredReviews               int
    DismissStaleReviews           bool
    RequireCodeOwnerReviews       bool
    RequiredStatusChecks          []string
    StrictStatusChecks            bool
    EnforceAdmins                 bool
    RestrictPushes                bool
    AllowedUsers                  []string
    AllowedTeams                  []string
    RequireConversationResolution bool
    AllowForcePushes              bool
    AllowDeletions                bool
}
```

## Usage in Commands

### List Command

The `gz repo-config list` command uses these APIs to:

1. List all repositories in an organization
2. Retrieve configuration for each repository (optional)
3. Check compliance against defined templates
4. Display results in various formats

Example:
```bash
# List repositories with basic info
gz repo-config list --org myorg

# Include detailed configuration
gz repo-config list --org myorg --show-config

# Filter by pattern
gz repo-config list --org myorg --filter "^api-.*"

# Output as JSON
gz repo-config list --org myorg --format json
```

## Error Handling

The API handles various error scenarios:

- **404 Not Found**: Repository or branch protection doesn't exist
- **403 Forbidden**: Insufficient permissions (gracefully handled)
- **429 Too Many Requests**: Automatic retry with backoff
- **5xx Server Errors**: Automatic retry with exponential backoff

## Performance Considerations

1. **Pagination**: Repository listing supports pagination for large organizations
2. **Rate Limiting**: Integrated rate limiter prevents API limit exhaustion
3. **Parallel Requests**: Can be configured for parallel operations
4. **Caching**: Results can be cached to reduce API calls

## Security

Required GitHub token permissions:
- `repo`: Full repository access
- `read:org`: Read organization data (for team permissions)
- `admin:repo_hook`: For webhook configuration (future)

## Example Integration

```go
// Create client with token
client := github.NewRepoConfigClient(token)

// List all repositories
repos, err := client.ListRepositories(ctx, "myorg", nil)
if err != nil {
    return err
}

// Get detailed config for each
for _, repo := range repos {
    config, err := client.GetRepositoryConfiguration(ctx, "myorg", repo.Name)
    if err != nil {
        log.Printf("Warning: %v", err)
        continue
    }
    
    // Process configuration
    fmt.Printf("Repository: %s\n", config.Name)
    fmt.Printf("  Template: %s\n", detectTemplate(repo, templateConfig))
    fmt.Printf("  Branch Protection: %v\n", len(config.BranchProtection) > 0)
    fmt.Printf("  Teams: %d\n", len(config.Permissions.Teams))
}
```

## Future Enhancements

1. **Webhook Configuration**: Retrieve webhook settings
2. **Deploy Keys**: List and manage deploy keys
3. **Secrets**: Organization and repository secrets (with proper permissions)
4. **Actions Permissions**: GitHub Actions settings and permissions
5. **Environments**: Deployment environment configurations