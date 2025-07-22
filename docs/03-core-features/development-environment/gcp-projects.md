# GCP Project Management

## Overview

The GCP Project Management feature provides comprehensive tools for managing Google Cloud Platform projects, configurations, and service accounts. This functionality extends the existing development environment management capabilities with GCP-specific operations.

## Features

### Project Management

- **List Projects**: View all accessible GCP projects with details
- **Switch Projects**: Change active project context
- **Project Details**: View comprehensive project information
- **Project Validation**: Verify access, billing, APIs, and permissions

### Configuration Management

- **gcloud Configurations**: Manage multiple gcloud configurations
- **Configuration Switching**: Switch between different environments
- **Configuration Creation**: Create new configurations with specific settings
- **Active Configuration Tracking**: Monitor which configuration is currently active

### Service Account Management

- **List Service Accounts**: View all service accounts in a project
- **Create/Delete Service Accounts**: Manage service account lifecycle
- **Key Management**: Create, list, and delete service account keys
- **Service Account Activation**: Set active service account for authentication

## Commands

### Basic Project Operations

```bash
# List all available projects
gz dev-env gcp-project list

# List projects in JSON format
gz dev-env gcp-project list --output json

# Switch to a specific project
gz dev-env gcp-project switch my-project-id

# Interactive project selection
gz dev-env gcp-project switch --interactive

# Show current project details
gz dev-env gcp-project show

# Show specific project details
gz dev-env gcp-project show my-project-id --output json
```

### Project Validation

```bash
# Basic project validation
gz dev-env gcp-project validate

# Comprehensive validation with all checks
gz dev-env gcp-project validate my-project-id \
  --check-apis \
  --check-billing \
  --check-permissions
```

### Configuration Management

```bash
# List all gcloud configurations
gz dev-env gcp-project config list

# Create a new configuration
gz dev-env gcp-project config create \
  --name production \
  --project my-prod-project \
  --account prod@company.com \
  --region us-central1 \
  --zone us-central1-a

# Activate a configuration
gz dev-env gcp-project config activate production

# Delete a configuration
gz dev-env gcp-project config delete staging
```

### Service Account Management

```bash
# List service accounts
gz dev-env gcp-project service-account list

# Create a new service account
gz dev-env gcp-project service-account create \
  --name my-service \
  --display-name "My Service Account" \
  --description "Service account for automated tasks"

# Show service account details
gz dev-env gcp-project service-account show \
  my-service@project.iam.gserviceaccount.com

# Create a service account key
gz dev-env gcp-project service-account create-key \
  my-service@project.iam.gserviceaccount.com \
  --key-type json \
  --output-file ./my-service-key.json

# Activate service account
gz dev-env gcp-project service-account activate \
  my-service@project.iam.gserviceaccount.com \
  --key-file ./my-service-key.json

# Delete service account
gz dev-env gcp-project service-account delete \
  my-service@project.iam.gserviceaccount.com \
  --force
```

## Configuration File Support

The GCP project manager supports both JSON and INI format configuration files used by gcloud:

### JSON Format (Modern gcloud)

```json
{
  "core": {
    "project": "my-project-id",
    "account": "user@company.com"
  },
  "compute": {
    "region": "us-central1",
    "zone": "us-central1-a"
  }
}
```

### INI Format (Legacy gcloud)

```ini
[core]
project = my-project-id
account = user@company.com

[compute]
region = us-central1
zone = us-central1-a
```

## Integration with Existing AWS Profile Management

The GCP project management follows the same patterns as the existing AWS profile management:

- **Similar Command Structure**: Consistent CLI patterns and flag usage
- **Configuration Management**: Unified approach to managing cloud provider configurations
- **Interactive Selection**: Common user experience for cloud resource selection
- **Validation and Health Checks**: Standardized validation patterns

## Data Structures

### GCPProject

```go
type GCPProject struct {
    ID               string            `json:"id"`
    Name             string            `json:"name"`
    Number           string            `json:"number"`
    LifecycleState   string            `json:"lifecycle_state"`
    Account          string            `json:"account"`
    Region           string            `json:"region"`
    Zone             string            `json:"zone"`
    Configuration    string            `json:"configuration"`
    ServiceAccount   string            `json:"service_account,omitempty"`
    BillingAccount   string            `json:"billing_account,omitempty"`
    IsActive         bool              `json:"is_active"`
    LastUsed         *time.Time        `json:"last_used,omitempty"`
    Tags             map[string]string `json:"tags,omitempty"`
    EnabledAPIs      []string          `json:"enabled_apis,omitempty"`
    IAMPermissions   []string          `json:"iam_permissions,omitempty"`
}
```

### GCPConfiguration

```go
type GCPConfiguration struct {
    Name           string `json:"name"`
    Project        string `json:"project"`
    Account        string `json:"account"`
    Region         string `json:"region"`
    Zone           string `json:"zone"`
    IsActive       bool   `json:"is_active"`
    PropertiesPath string `json:"properties_path"`
}
```

### GCPServiceAccount

```go
type GCPServiceAccount struct {
    Email       string    `json:"email"`
    Name        string    `json:"name"`
    DisplayName string    `json:"displayName"`
    ProjectID   string    `json:"projectId"`
    UniqueID    string    `json:"uniqueId"`
    Description string    `json:"description"`
    Disabled    bool      `json:"disabled"`
    OAuth2ClientID string `json:"oauth2ClientId"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    KeyCount    int       `json:"keyCount"`
    IsActive    bool      `json:"isActive"`
}
```

## Prerequisites

- **gcloud CLI**: Must be installed and available in PATH
- **Authentication**: User must be authenticated with gcloud
- **Project Access**: Appropriate permissions for target GCP projects

## Security Considerations

1. **Service Account Keys**:
   - Generated key files have secure permissions (0600)
   - Users are warned about secure handling
   - Keys should not be committed to version control

2. **Authentication**:
   - Uses existing gcloud authentication
   - Supports both user and service account authentication
   - Credential validation before operations

3. **Project Access**:
   - Validates project access before operations
   - Checks billing and API enablement when requested
   - Verifies IAM permissions for sensitive operations

## Error Handling

The implementation includes comprehensive error handling for:

- **gcloud CLI Availability**: Graceful degradation when gcloud is not available
- **Authentication Issues**: Clear messages for authentication problems
- **Permission Errors**: Descriptive errors for insufficient permissions
- **Network Issues**: Proper handling of network connectivity problems
- **Configuration Errors**: Validation of configuration file formats

## Testing

The implementation includes comprehensive test coverage:

- **Unit Tests**: Core functionality and data structure validation
- **Integration Tests**: gcloud CLI integration (when available)
- **Mock Tests**: Simulated gcloud responses for consistent testing
- **Benchmark Tests**: Performance testing for large project lists

Run tests with:

```bash
go test ./cmd/dev-env -run "TestGCPProject" -v
```

## Future Enhancements

Potential future improvements:

1. **API Integration**: Direct GCP API integration as fallback to gcloud CLI
2. **Resource Management**: Extended resource management (VMs, storage, etc.)
3. **Quota Management**: Project quota monitoring and management
4. **Cost Tracking**: Project cost analysis and budgeting
5. **Multi-Project Operations**: Bulk operations across multiple projects
6. **Configuration Templates**: Predefined configuration templates for common setups
