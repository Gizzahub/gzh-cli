# gzh.yaml Quick Reference

## Minimal Configuration

```yaml
version: "1.0.0"
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "your-org"
```

## Complete Configuration Options

```yaml
version: "1.0.0"                      # Required
default_provider: github              # Optional: github|gitlab|gitea|gogs

providers:
  github:                            # Provider: github|gitlab|gitea|gogs
    token: "${TOKEN}"                # Required: Auth token
    orgs:                            # For GitHub/Gitea/Gogs
      - name: "org-name"             # Required: Organization name
        visibility: "all"            # Optional: public|private|all (default: all)
        clone_dir: "./path"          # Optional: Target directory
        match: "^pattern.*"          # Optional: Regex filter
        exclude: ["pattern-*"]       # Optional: Exclude patterns
        strategy: "reset"            # Optional: reset|pull|fetch (default: reset)
        flatten: false               # Optional: true|false (default: false)
        
  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:                          # For GitLab only
      - name: "group-name"
        recursive: true              # Optional: Include subgroups (default: false)
        # ... same options as above
```

## Common Patterns

### Multiple Organizations
```yaml
providers:
  github:
    token: "${GITHUB_TOKEN}"
    orgs:
      - name: "work-org"
        clone_dir: "./work"
      - name: "personal-org"  
        clone_dir: "./personal"
```

### Filtered Repositories
```yaml
orgs:
  - name: "my-org"
    match: "^api-.*"               # Only API repositories
    exclude: ["*-archive", "*-backup"]  # Exclude patterns
```

### Directory Structures
```yaml
# Nested: ./github/org-name/repo-name
flatten: false

# Flat: ./github/repo-name  
flatten: true
```

### Strategies
```yaml
strategy: "reset"    # Hard reset (default)
strategy: "pull"     # Git pull if exists
strategy: "fetch"    # Git fetch only
```

### Environment Variables
```yaml
token: "${GITHUB_TOKEN}"                    # Required variable
clone_dir: "${HOME}/repos"                  # Path variable
token: "${OPTIONAL_TOKEN:default-value}"   # With default value
```

## Command Usage

```bash
# Use gzh.yaml from current directory
gzh bulk-clone --use-gzh-config

# Use specific config file
gzh bulk-clone --config-file /path/to/gzh.yaml

# Validate configuration
gzh config validate

# Dry run (preview only)
gzh bulk-clone --dry-run --use-gzh-config
```

## File Locations (Search Order)

1. `--config-file` flag
2. `./gzh.yaml`
3. `~/.config/gzh.yaml`
4. `$GZH_CONFIG_PATH`

## Common Error Solutions

| Error | Solution |
|-------|----------|
| `missing required field: token` | Set environment variable: `export GITHUB_TOKEN="..."` |
| `invalid regex pattern` | Test regex: `echo "test" \| grep -E "pattern"` |
| `configuration file not found` | Check file path and permissions |
| `permission denied` | Use accessible directory: `${HOME}/repos` |

## Examples

- [Simple](../samples/gzh-simple.yaml) - Basic single organization
- [Multi-provider](../samples/gzh-multi-provider.yaml) - GitHub + GitLab + Gitea + Gogs
- [Development](../samples/gzh-development.yaml) - Development workflow
- [Enterprise](../samples/gzh-enterprise.yaml) - Large organization setup