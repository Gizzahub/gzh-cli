# GitHub Token Permissions Documentation

This document outlines the required GitHub token permissions for various gzh-manager-go operations.

## Overview

The gzh-manager-go tool requires different levels of GitHub API access depending on the operations being performed. The tool supports both GitHub.com and GitHub Enterprise Server (GHES) instances.

## Required Permissions by Feature

### Basic Repository Operations

#### Repository Information & Cloning
**Operations**: `gz bulk-clone`, repository listing, basic repository information
**Required Scopes**:
- `repo` (for private repositories)
- `public_repo` (for public repositories only)

**API Endpoints Used**:
- `GET /orgs/{org}/repos` - List organization repositories
- `GET /repos/{owner}/{repo}` - Get repository information

#### Repository Configuration Management
**Operations**: Repository settings retrieval and updates
**Required Scopes**:
- `repo` - Full repository access
- `admin:org` - Organization administration (for organization-level operations)

**API Endpoints Used**:
- `GET /repos/{owner}/{repo}` - Get repository details
- `PATCH /repos/{owner}/{repo}` - Update repository settings
- `GET /repos/{owner}/{repo}/branches/{branch}/protection` - Get branch protection
- `PUT /repos/{owner}/{repo}/branches/{branch}/protection` - Update branch protection

### Advanced Repository Management

#### Team and User Permissions
**Operations**: Managing repository collaborators and team access
**Required Scopes**:
- `repo` - Repository access
- `admin:org` - Organization administration
- `read:org` - Read organization membership (minimum)

**API Endpoints Used**:
- `GET /repos/{owner}/{repo}/teams` - List repository teams
- `GET /repos/{owner}/{repo}/collaborators` - List repository collaborators
- `PUT /repos/{owner}/{repo}/teams/{team_slug}` - Add team to repository
- `PUT /repos/{owner}/{repo}/collaborators/{username}` - Add collaborator

#### Organization-wide Operations
**Operations**: Bulk operations across all organization repositories
**Required Scopes**:
- `repo` - Full repository access
- `admin:org` - Organization administration
- `read:org` - Read organization information

**API Endpoints Used**:
- `GET /orgs/{org}/repos` - List all organization repositories
- Multiple repository endpoints for bulk operations

## Permission Levels Explained

### Token Types

#### Personal Access Tokens (Classic)
Classic personal access tokens with the following scopes:

**Minimum Required**:
```
repo                    # Full control of private repositories
admin:org               # Full control of orgs and teams
read:org               # Read org and team membership
```

**Optional but Recommended**:
```
admin:repo_hook        # Admin access to repository hooks
read:repo_hook         # Read access to repository hooks
admin:org_hook         # Admin access to organization hooks
```

#### Fine-grained Personal Access Tokens
For organizations that support fine-grained tokens:

**Repository Permissions**:
- `Contents`: Read (for repository information)
- `Metadata`: Read (for basic repository data)
- `Administration`: Write (for repository settings)
- `Pull requests`: Write (for branch protection rules)

**Organization Permissions**:
- `Members`: Read (for team management)
- `Administration`: Read (for organization information)

### GitHub Apps
When using GitHub Apps, the following permissions are required:

**Repository Permissions**:
- `Repository administration`: Write
- `Contents`: Read
- `Metadata`: Read
- `Pull requests`: Write

**Organization Permissions**:
- `Members`: Read
- `Administration`: Read

## Token Verification

### Automatic Token Validation
The tool automatically validates token permissions before performing operations:

```go
// Example validation check
func (c *RepoConfigClient) ValidatePermissions(ctx context.Context) error {
    // Check if token has required repository access
    // Check if token has organization access
    // Return detailed error if permissions are insufficient
}
```

### Manual Token Testing
You can test your token permissions using the GitHub API directly:

```bash
# Test repository access
curl -H "Authorization: token YOUR_TOKEN" \
  https://api.github.com/repos/OWNER/REPO

# Test organization access  
curl -H "Authorization: token YOUR_TOKEN" \
  https://api.github.com/orgs/ORG/repos
```

## Common Permission Issues

### "Resource not accessible by integration"
**Cause**: Token lacks required permissions for the operation
**Solution**: Grant additional scopes or permissions to the token

### "Not Found" for existing resources
**Cause**: Token lacks read access to the resource
**Solution**: Ensure `repo` scope for private repositories, `read:org` for organization resources

### Rate Limiting
**Impact**: All operations are subject to GitHub's rate limits
**Mitigation**: The tool implements automatic rate limiting and retry logic

## Security Best Practices

### Token Management
1. **Use minimum required permissions**: Only grant scopes necessary for your operations
2. **Token rotation**: Regularly rotate tokens for security
3. **Environment variables**: Store tokens in environment variables, not code
4. **Scope limitation**: Use fine-grained tokens when available

### Repository Access
1. **Private repositories**: Require `repo` scope, which grants broad access
2. **Organization repos**: Consider using GitHub Apps for better permission control
3. **Audit logs**: Monitor token usage through GitHub's audit logs

## Environment Configuration

### Setting Up Tokens
```bash
# Set GitHub token for the tool
export GITHUB_TOKEN="ghp_your_token_here"
export GH_TOKEN="ghp_your_token_here"  # Alternative

# For GitHub Enterprise Server
export GITHUB_API_URL="https://github.company.com/api/v3"
```

### Configuration Validation
```bash
# Test token permissions
gz repo-config validate-token

# Test organization access
gz bulk-clone --dry-run --org your-org
```

## Troubleshooting

### Common Error Messages

#### 401 Unauthorized
```
GitHub API error (401): Bad credentials
```
**Solution**: Check token validity and ensure it's correctly set

#### 403 Forbidden - Insufficient Permissions
```
GitHub API error (403): Resource not accessible by integration
```
**Solution**: Grant required scopes to the token

#### 403 Forbidden - Rate Limited
```
GitHub API error (403): API rate limit exceeded
```
**Solution**: Wait for rate limit reset or use token with higher rate limits

#### 404 Not Found
```
GitHub API error (404): Not Found
```
**Possible Causes**:
- Resource doesn't exist
- Token lacks read permissions
- Organization/repository is private and token lacks access

### Permission Testing Script
```bash
#!/bin/bash
# Test GitHub token permissions

TOKEN="${GITHUB_TOKEN}"
ORG="${1:-your-org}"

echo "Testing GitHub token permissions..."

# Test basic API access
echo "1. Testing basic API access..."
curl -s -H "Authorization: token $TOKEN" \
  https://api.github.com/user | jq -r '.login' || echo "Failed"

# Test organization access
echo "2. Testing organization access..."
curl -s -H "Authorization: token $TOKEN" \
  "https://api.github.com/orgs/$ORG" | jq -r '.login' || echo "Failed"

# Test repository listing
echo "3. Testing repository listing..."
curl -s -H "Authorization: token $TOKEN" \
  "https://api.github.com/orgs/$ORG/repos?per_page=1" | jq -r '.[0].name' || echo "Failed"

echo "Permission testing complete."
```

## References

- [GitHub REST API Authentication](https://docs.github.com/en/rest/overview/authenticating-to-the-rest-api)
- [GitHub Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [GitHub Apps Permissions](https://docs.github.com/en/developers/apps/building-github-apps/setting-permissions-for-github-apps)
- [GitHub Rate Limiting](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)
