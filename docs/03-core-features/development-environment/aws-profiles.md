# AWS Profile Management Guide

This guide describes how to use the AWS profile management feature in gzh-manager-go to manage multiple AWS accounts, SSO integration, and automatic credential renewal.

## Overview

The `gz dev-env aws-profile` command provides comprehensive AWS profile management capabilities:

- **List and switch** between AWS profiles
- **SSO login** and token management
- **Multi-account** profile switching
- **Automatic credential renewal**
- **Profile validation** and health checks
- **Interactive selection** mode

## Features

### 1. Profile Management

#### List All Profiles
```bash
# List all AWS profiles in table format
gz dev-env aws-profile list

# List profiles in JSON format
gz dev-env aws-profile list --output json
```

Output example:
```
+------------+-----------+------------+--------------------------------+----------------+--------+
| PROFILE    | REGION    | TYPE       | SSO URL                        | ACCOUNT ID     | ACTIVE |
+------------+-----------+------------+--------------------------------+----------------+--------+
| default    | us-east-1 | Standard   |                                |                |        |
| production | us-west-2 | SSO        | https://mycompany.awsapps.com  | 123456789012   | ✓      |
| staging    | eu-west-1 | AssumeRole |                                |                |        |
+------------+-----------+------------+--------------------------------+----------------+--------+
```

#### Switch Profiles
```bash
# Switch to a specific profile
gz dev-env aws-profile switch production

# Interactive profile selection
gz dev-env aws-profile switch --interactive
```

The switch command:
- Updates the `AWS_PROFILE` environment variable
- Updates shell configuration files (.bashrc, .zshrc, .profile)
- Validates credentials if available

### 2. SSO Integration

#### SSO Login
```bash
# Login to AWS SSO for a profile
gz dev-env aws-profile login production
```

The SSO login process:
1. Opens a browser window for authentication
2. Displays a user code for verification
3. Waits for authentication completion
4. Saves tokens for future use

#### SSO Configuration

AWS profiles with SSO should be configured in `~/.aws/config`:

```ini
[profile production]
sso_start_url = https://mycompany.awsapps.com/start
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = AdministratorAccess
region = us-west-2
output = json
```

### 3. Profile Types

The tool supports various AWS profile types:

#### Standard Profile
```ini
[default]
region = us-east-1
output = json
```

#### SSO Profile
```ini
[profile production]
sso_start_url = https://mycompany.awsapps.com/start
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = AdministratorAccess
region = us-west-2
```

#### AssumeRole Profile
```ini
[profile staging]
role_arn = arn:aws:iam::123456789012:role/StagingRole
source_profile = default
region = eu-west-1
mfa_serial = arn:aws:iam::123456789012:mfa/user
external_id = unique-external-id
```

#### Credential Process Profile
```ini
[profile custom]
credential_process = /usr/local/bin/aws-credential-helper
region = us-east-1
```

### 4. Profile Information

#### Show Profile Details
```bash
# Display detailed profile information
gz dev-env aws-profile show production
```

Output example:
```
Profile: production
Region: us-west-2
Output Format: json

SSO Configuration:
  Start URL: https://mycompany.awsapps.com/start
  Role Name: AdministratorAccess
  Account ID: 123456789012
  SSO Region: us-east-1

Validating credentials...
Status: ✅ Valid
```

### 5. Credential Validation

#### Validate Single Profile
```bash
# Validate current profile
gz dev-env aws-profile validate

# Validate specific profile
gz dev-env aws-profile validate production
```

#### Validate All Profiles
```bash
# Validate all configured profiles
gz dev-env aws-profile validate --all
```

Output example:
```
+------------+------------+------------+---------------------------+
| PROFILE    | TYPE       | STATUS     | DETAILS                   |
+------------+------------+------------+---------------------------+
| default    | Standard   | ✅ Valid   | Credentials are valid     |
| production | SSO        | ❌ Invalid | SSO token expired         |
| staging    | AssumeRole | ✅ Valid   | Credentials are valid     |
+------------+------------+------------+---------------------------+
```

## Use Cases

### 1. Daily Development Workflow

```bash
# Morning: Switch to development profile
gz dev-env aws-profile switch development

# Login if using SSO
gz dev-env aws-profile login development

# Work on development tasks...

# Afternoon: Switch to production for deployment
gz dev-env aws-profile switch production

# Validate credentials before deployment
gz dev-env aws-profile validate production
```

### 2. Multi-Account Management

```bash
# List all available profiles
gz dev-env aws-profile list

# Interactive selection for quick switching
gz dev-env aws-profile switch -i

# Validate all profiles weekly
gz dev-env aws-profile validate --all
```

### 3. CI/CD Integration

```bash
# Validate profile before running pipeline
if gz dev-env aws-profile validate ci-profile; then
    echo "Credentials valid, proceeding with deployment"
    # Run deployment
else
    echo "Invalid credentials, aborting"
    exit 1
fi
```

## Environment Variables

The tool respects and updates these environment variables:

- `AWS_PROFILE` - Current active AWS profile
- `AWS_REGION` - Default AWS region (from profile)
- `AWS_DEFAULT_REGION` - Alternative region variable

## Shell Integration

The tool automatically updates shell configuration files:

- `~/.bashrc` - Bash configuration
- `~/.zshrc` - Zsh configuration
- `~/.profile` - Generic shell profile

To apply changes immediately:
```bash
# For bash
source ~/.bashrc

# For zsh
source ~/.zshrc
```

## Security Considerations

1. **SSO Tokens**: Stored securely in `~/.aws/sso/cache/`
2. **Credentials**: Never displayed in plain text
3. **MFA Support**: Full support for MFA-protected roles
4. **Token Expiry**: Automatic detection of expired tokens

## Troubleshooting

### SSO Login Issues

If SSO login fails:
1. Check your SSO start URL is correct
2. Ensure you have access to the SSO portal
3. Verify your browser can open the authentication page
4. Check network connectivity

### Profile Not Found

If a profile is not found:
1. Verify the profile exists in `~/.aws/config`
2. Check the profile name spelling
3. Ensure proper formatting in the config file

### Credential Validation Failures

If credentials fail validation:
1. Re-login for SSO profiles: `gz dev-env aws-profile login <profile>`
2. Check MFA token for AssumeRole profiles
3. Verify IAM permissions are correct
4. Check for expired temporary credentials

## Best Practices

1. **Use SSO** for enterprise environments
2. **Validate credentials** before critical operations
3. **Use interactive mode** when unsure of profile names
4. **Set up MFA** for sensitive profiles
5. **Regular validation** of all profiles
6. **Use descriptive profile names** (e.g., `prod-readonly`, `dev-admin`)

## Configuration Examples

### Enterprise SSO Setup
```ini
[profile prod-readonly]
sso_start_url = https://company.awsapps.com/start
sso_region = us-east-1
sso_account_id = 111111111111
sso_role_name = ReadOnlyAccess
region = us-west-2

[profile prod-admin]
sso_start_url = https://company.awsapps.com/start
sso_region = us-east-1
sso_account_id = 111111111111
sso_role_name = AdministratorAccess
region = us-west-2
```

### Multi-Environment Setup
```ini
[profile dev]
sso_start_url = https://company.awsapps.com/start
sso_account_id = 222222222222
sso_role_name = DeveloperAccess
region = us-east-1

[profile staging]
role_arn = arn:aws:iam::333333333333:role/DeploymentRole
source_profile = dev
region = us-east-1

[profile production]
role_arn = arn:aws:iam::444444444444:role/ProductionRole
source_profile = dev
mfa_serial = arn:aws:iam::222222222222:mfa/developer
region = us-east-1
```

## Integration with Other Tools

The AWS profile management integrates seamlessly with:

- **AWS CLI**: Uses standard AWS configuration files
- **AWS SDKs**: All SDKs respect the AWS_PROFILE variable
- **Terraform**: Works with AWS provider configuration
- **kubectl**: For EKS cluster access
- **Other gzh-manager commands**: Consistent profile usage

## Future Enhancements

Planned features for future releases:

1. **Automatic token refresh** for SSO profiles
2. **Profile groups** for batch operations
3. **Credential rotation** reminders
4. **Integration with AWS Vault**
5. **Profile usage analytics**
6. **Export/import** profile configurations
