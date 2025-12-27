# Architecture Guide - gzh-cli

## Integration Libraries Pattern

**Core Principle**: gzh-cli is a wrapper-based integration CLI. Core logic lives in external libraries.

```
gzh-cli (Main CLI - gz binary)
├── gzh-cli-core        → Core utilities
├── gzh-cli-git         → Git operations
├── gzh-cli-gitforge    → GitHub/GitLab integration
├── gzh-cli-quality     → Code quality tools
├── gzh-cli-shellforge  → Shell config builder
├── gzh-cli-package-manager → Package manager
├── gzh-cli-os-env      → OS environment
├── gzh-cli-net-env     → Network environment
├── gzh-cli-dev-env     → Development environment
└── gzh-cli-template    → Template utilities
```

## Integration Library Wrappers

| Library | Wrapper Location | Lines | Purpose |
|---------|-----------------|-------|---------|
| gzh-cli-git | `cmd/git/repo/*_wrapper.go` | 473 | Local Git operations |
| gzh-cli-quality | `cmd/quality_wrapper.go` | 45 | Code quality tools |
| gzh-cli-package-manager | `cmd/pm_wrapper.go` | 65 | Package manager integration |
| gzh-cli-shellforge | `cmd/shellforge_wrapper.go` | 71 | Shell config builder |

**When to Modify**:

- **Core logic change** → Modify external library repository
- **CLI integration change** → Modify wrapper file in gzh-cli

## Key Patterns

### 1. Interface Abstraction

```go
// internal/git/interfaces.go
type Client interface {
    Clone(ctx context.Context, options CloneOptions) error
    Pull(ctx context.Context, options PullOptions) error
}
```

### 2. Provider Registry

```go
// pkg/git/provider/
providerRegistry := provider.NewRegistry()
providerRegistry.Register("github", github.NewProvider())
```

### 3. Strategy Pattern (Git Operations)

- `rebase`: Rebase local changes on remote
- `reset`: Hard reset to match remote
- `clone`: Fresh clone (remove existing)
- `pull`: Standard git pull (merge)
- `fetch`: Update refs only

## Extensions & Lifecycle System (Phase 3)

### Extension System

**Location**: `internal/extensions/`

User-extensible aliases, workflows, and external command integration.

**Three Alias Types**:

1. **Simple Alias**: Command shortcut

```yaml
update-all:
  command: "pm update --all"
  description: "Update all package managers"
```

2. **Multi-Step Workflow**: Sequential execution

```yaml
full-sync:
  description: "Complete sync workflow"
  steps:
    - "synclone run"
    - "pm update --all"
    - "git repo pull-all"
```

3. **Parameterized Alias**: Variable substitution

```yaml
clone-and-setup:
  command: "git repo clone-or-update ${url}"
  params:
    - name: url
      description: "Repository URL"
      required: true
```

### Lifecycle Management

**Location**: `cmd/registry/lifecycle.go`

Commands progress through stages with automatic filtering and warnings.

**Four Stages**:

| Stage | Behavior | Enablement | Warning |
|-------|----------|------------|---------|
| **Experimental** | Disabled by default | `GZ_EXPERIMENTAL=1` or `--experimental` | ⚠️ May change or be removed |
| **Beta** | Enabled, shows info | Always enabled | ℹ️ In testing, report issues |
| **Stable** | Default, no warnings | Always enabled | None |
| **Deprecated** | Shows warning | Always enabled | ⚠️ Will be removed |

**Adding Metadata**:

```go
func (p *cmdProvider) Metadata() registry.CommandMetadata {
    return registry.CommandMetadata{
        Name:         "my-command",
        Category:     registry.CategoryUtility,
        Version:      "1.0.0",
        Lifecycle:    registry.LifecycleStable,
        Dependencies: []string{"git", "docker"},
        Tags:         []string{"utility", "tool"},
        Priority:     50,
    }
}
```

## Configuration Architecture

### Configuration Hierarchy

1. `$GZH_CONFIG_PATH` (env var)
1. `./gzh.yaml` (current dir)
1. `~/.config/gzh-manager/gzh.yaml` (user config)
1. `/etc/gzh-manager/gzh.yaml` (system config)

### Configuration Structure

```yaml
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/github/myorg"
```

**Features**:

- Environment variable expansion: `${GITHUB_TOKEN}`
- JSON Schema validation
- Priority system: CLI flags > env vars > config files > defaults

## Performance Benchmarks

### Target Metrics

- **Startup time**: \<50ms
- **Binary size**: ~33MB
- **Memory**: Minimal footprint
- **Command response**: \<100ms for most commands

### Profiling

```bash
# Runtime stats
gz profile stats

# CPU profiling
gz profile cpu --duration 30s

# Memory profiling
gz profile memory

# Web interface
gz profile server --port 6060
```
