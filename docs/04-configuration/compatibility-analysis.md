# Configuration Compatibility Analysis

## Overview

This document analyzes the compatibility between the existing `bulk-clone.yaml` configuration format and the new `gzh.yaml` schema format.

## Current bulk-clone.yaml Structure

### Schema Fields

```yaml
version: "0.1"
default:
  protocol: https
  github:
    root_path: "$HOME/repos/github"
    provider: "github"
    protocol: ""
    org_name: ""
  gitlab:
    root_path: "$HOME/repos/gitlab"
    provider: "gitlab"
    url: "https://gitlab.com"
    recursive: false
    protocol: ""
    group_name: ""

repo_roots:
  - root_path: "$HOME/work/mycompany"
    provider: "github"
    protocol: "ssh"
    org_name: "mycompany"
  - root_path: "$HOME/opensource"
    provider: "github"
    protocol: "https"
    org_name: "kubernetes"

ignore_names:
  - "test-.*"
  - ".*-archive"
```

## New gzh.yaml Structure

### Schema Fields

```yaml
version: "1.0.0"
default_provider: github
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "mycompany"
        visibility: all
        clone_dir: "$HOME/work/mycompany"
        strategy: reset
        match: "^project-.*"
        exclude:
          - "test-.*"
          - ".*-archive"
  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "mygroup"
        visibility: public
        recursive: true
        flatten: false
        clone_dir: "$HOME/opensource"
```

## Compatibility Mapping

### Direct Mappings

| bulk-clone.yaml          | gzh.yaml           | Notes                             |
| ------------------------ | ------------------ | --------------------------------- |
| `version`                | `version`          | Different format: "0.1" → "1.0.0" |
| `default.protocol`       | N/A                | Moved to provider-specific config |
| `repo_roots[].provider`  | Provider key       | "github" → `providers.github`     |
| `repo_roots[].org_name`  | `orgs[].name`      | Direct mapping                    |
| `repo_roots[].root_path` | `orgs[].clone_dir` | Renamed field                     |
| `ignore_names[]`         | `orgs[].exclude[]` | Moved to target level             |

### New Fields in gzh.yaml

| Field                | Purpose                      | Default  |
| -------------------- | ---------------------------- | -------- |
| `default_provider`   | Default Git provider         | "github" |
| `providers.*.token`  | Authentication token         | Required |
| `orgs[].visibility`  | Repository visibility filter | "all"    |
| `orgs[].strategy`    | Clone/update strategy        | "reset"  |
| `orgs[].match`       | Regex filter for repos       | None     |
| `groups[].recursive` | Include subgroups            | false    |
| `groups[].flatten`   | Flatten directory structure  | false    |

### Missing Fields in gzh.yaml

| bulk-clone.yaml Field   | Alternative in gzh.yaml                 |
| ----------------------- | --------------------------------------- |
| `repo_roots[].protocol` | Implicit in authentication method       |
| `default.github.url`    | Not needed for github.com               |
| `default.gitlab.url`    | Could be added as provider-level config |

## Migration Strategy

### 1. Automatic Migration

- Convert `version: "0.1"` → `version: "1.0.0"`
- Map `repo_roots[]` to appropriate provider sections
- Convert `org_name`/`group_name` to `name` field
- Map `root_path` to `clone_dir`
- Move `ignore_names` to individual `exclude` arrays

### 2. Manual Configuration Required

- **Authentication tokens**: Must be configured in new format
- **Visibility filters**: Need to be set based on requirements
- **Strategy selection**: Choose appropriate update strategy
- **Regex patterns**: Convert ignore patterns to match/exclude patterns

### 3. Enhanced Features Available

- **Token-based authentication**: More secure than protocol-based auth
- **Visibility filtering**: Better control over which repos to clone
- **Regex matching**: More flexible repository filtering
- **Strategy options**: Better control over update behavior
- **GitLab subgroups**: Enhanced GitLab support with recursive/flatten options

## Compatibility Matrix

| Feature              | bulk-clone.yaml | gzh.yaml    | Compatible    |
| -------------------- | --------------- | ----------- | ------------- |
| Basic org cloning    | ✅              | ✅          | ✅            |
| Multiple providers   | ✅              | ✅          | ✅            |
| Custom clone paths   | ✅              | ✅          | ✅            |
| Ignore patterns      | ✅              | ✅          | ✅ (enhanced) |
| Protocol selection   | ✅              | Implicit    | ⚠️            |
| Authentication       | Basic           | Token-based | ⚠️            |
| Visibility filtering | ❌              | ✅          | ➕            |
| Update strategies    | ❌              | ✅          | ➕            |
| Regex filtering      | ❌              | ✅          | ➕            |
| GitLab subgroups     | Basic           | Enhanced    | ➕            |

## Recommendations

### For Backward Compatibility

1. **Migration tool**: Create automatic migration from bulk-clone.yaml to gzh.yaml
2. **Hybrid support**: Support both formats during transition period
3. **Deprecation timeline**: Gradual phase-out of old format

### For New Implementations

1. **Use gzh.yaml**: More flexible and feature-rich
2. **Token authentication**: More secure and reliable
3. **Visibility filtering**: Better control over repository access
4. **Strategy selection**: Choose appropriate update behavior per use case

## Migration Tool Requirements

### Input Validation

- Validate existing bulk-clone.yaml format
- Check for required fields and valid values
- Warn about unsupported configurations

### Conversion Logic

- Map provider configurations correctly
- Convert ignore patterns to exclude lists
- Generate secure token placeholders
- Set appropriate default values

### Output Generation

- Generate valid gzh.yaml configuration
- Include migration comments
- Provide setup instructions for tokens
- Suggest optimal settings based on usage patterns
