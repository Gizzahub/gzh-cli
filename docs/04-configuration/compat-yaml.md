# Compatibility Filters Configuration (`~/.gzh/pm/compat.yml`)

## Purpose
- Configure custom compatibility filters to apply environment variables, warnings, and post actions per manager/plugin.
- Override or extend built-in filters.

## Schema
```yaml
filters:
  - manager: string               # e.g., "asdf"
    plugin: string                # e.g., "rust", "nodejs", "python"
    kind: string                  # "advisory" | "conflict"
    level: string                 # optional log level (info|warn|error)
    warning: string               # optional
    env:                          # optional key=value map
      KEY: "VALUE"
    when:                         # optional conditions
      os: ["linux", "darwin"]
      arch: ["amd64", "arm64"]
      has_command: ["rustup", "corepack"]
      version_range:
        manager: ">=0.13"        # reserved for future use
    match_env:                    # optional environment checks
      path_contains: ["~/.asdf/shims"]
    post:                         # optional post actions
      - command: ["bash","-lc","corepack enable"]
        description: "Enable corepack"
        ignore_error: true
```

## Notes
- User filters are merged after built-in filters; later values override earlier ones on key conflicts.
- `conflict` kind is used by strict mode to fail fast.
- `post` actions run after successful installs or can be previewed in dry-run.
