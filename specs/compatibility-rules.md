# Compatibility Rules (Filter Chain)

## Overview
- Integrates a filter chain to apply environment variables, warnings, and post actions per manager/plugin.
- Purpose: Avoid common conflicts (e.g., rustup vs asdf rust) and recommend best practices by default.

## Built-in Filters
- asdf + rust
  - Env: `RUSTUP_INIT_SKIP_PATH_CHECK=yes`, `RUSTUP_INIT_YES=1`
  - Kind: conflict
  - Rationale: rustup installer aborts if Rust is already on PATH via asdf shims
- asdf + nodejs
  - Env: `COREPACK_ENABLE=1`
  - Post: `corepack enable`, `corepack prepare pnpm@latest --activate`
  - Kind: advisory
- asdf + python
  - Env: `PIP_REQUIRE_VIRTUALENV=1`
  - Kind: advisory
- asdf + golang
  - Env (if unset): `GOBIN=$HOME/go/bin`
  - Warning if `$GOBIN` or `$GOPATH/bin` not in PATH
  - Kind: advisory

## Modes
- `--compat auto` (default): apply filters, print warnings, run post actions
- `--compat strict`: if any conflict filter matches, abort with error
- `--compat off`: disable all filters

## User Configuration
- File: `~/.gzh/pm/compat.yml`
- Merge order: user filters override built-ins for Env and behavior
- Schema extensions supported: `when`, `match_env`, `level`

### Example `compat.yml`
```yaml
filters:
  - manager: "asdf"
    plugin: "rust"
    kind: "conflict"
    warning: "rustup 권장: PATH 충돌 주의"
    env:
      RUSTUP_INIT_SKIP_PATH_CHECK: "yes"
      RUSTUP_INIT_YES: "1"

  - manager: "asdf"
    plugin: "nodejs"
    kind: "advisory"
    warning: "corepack 권장 활성화"
    env:
      COREPACK_ENABLE: "1"
    post:
      - command: ["bash","-lc","corepack enable"]
        description: "Enable corepack"
        ignore_error: true
      - command: ["bash","-lc","corepack prepare pnpm@latest --activate"]
        description: "Corepack prepare pnpm"
        ignore_error: true
```

## Future Work
- Add `when.version_range` evaluation (manager version constraints)
- Extend to more managers (apt/brew/sdkman/npm/pip)
