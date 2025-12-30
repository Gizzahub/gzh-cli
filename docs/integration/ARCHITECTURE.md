# gzh-cli Architecture

## Overview

gzh-cli uses an "Integration Libraries Pattern" where common functionality is extracted into specialized external libraries and integrated as dependencies, establishing single sources of truth while reducing code duplication.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        gzh-cli (Binary: gz)                     │
│                     Comprehensive CLI Tool                       │
└─────────────────────────────────────────────────────────────────┘
                                 │
                ┌────────────────┼────────────────┐
                │                │                │
                ▼                ▼                ▼
     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
     │   Wrapper    │  │   Wrapper    │  │   Wrapper    │
     │  Commands    │  │  Commands    │  │  Commands    │
     │              │  │              │  │              │
     │  45 lines    │  │  65 lines    │  │  473 lines   │
     └──────────────┘  └──────────────┘  └──────────────┘
             │                 │                 │
             ▼                 ▼                 ▼
     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
     │ gzh-cli-     │  │ gzh-cli-     │  │  gzh-cli-    │
     │  quality     │  │  package-    │  │    git       │
     │              │  │  manager     │  │              │
     │ Code Quality │  │ PM Updates   │  │ Git Ops      │
     │ 3,469 lines  │  │ 2,388 lines  │  │ 845 lines    │
     │ saved (98.7%)│  │ saved (97.3%)│  │ saved (64.2%)│
     └──────────────┘  └──────────────┘  └──────────────┘
```

**Total Code Reduction**: 6,702 lines (92.0% reduction rate)

## Integration Libraries Pattern

### Pattern Components

```
┌─────────────────────────────────────────────────────────────────┐
│                      gzh-cli (Main Binary)                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  cmd/quality_wrapper.go (45 lines)                             │
│  ┌────────────────────────────────────────────────┐            │
│  │ func NewQualityCmd() *cobra.Command {          │            │
│  │     cmd := quality.NewRootCmd()  // Delegate   │            │
│  │     return cmd                                  │            │
│  │ }                                               │            │
│  └────────────────────────────────────────────────┘            │
│                          │                                      │
│                          │ (Thin wrapper delegates)             │
│                          ▼                                      │
└─────────────────────────────────────────────────────────────────┘
                           │
                           │ go.mod dependency
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              github.com/gizzahub/gzh-cli-quality                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  • Full implementation (3,514 lines)                            │
│  • Standalone binary (quality)                                 │
│  • Dual usage: CLI tool + library                              │
│  • Single source of truth                                      │
│                                                                 │
│  Supported Tools:                                               │
│  ├─ Go: gofumpt, golangci-lint, gci                            │
│  ├─ Python: ruff, black, isort, mypy                           │
│  ├─ JS/TS: prettier, eslint, dprint                            │
│  └─ Rust, Java, C/C++, Shell, YAML, etc.                       │
└─────────────────────────────────────────────────────────────────┘
```

### Benefits

1. **Single Source of Truth**: Fixes and features implemented once
1. **Independent Development**: Each library evolves independently
1. **Dual Usage**: Works standalone AND integrated
1. **Reduced Maintenance**: Less code to maintain in gzh-cli
1. **Clear Separation**: Explicit functionality boundaries

## Full System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         gzh-cli Binary (gz)                     │
└─────────────────────────────────────────────────────────────────┘
                                 │
        ┌────────────────────────┼────────────────────────┐
        │                        │                        │
        ▼                        ▼                        ▼
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│   Integrated  │      │   gzh-cli     │      │    Native     │
│   Libraries   │      │   Specific    │      │   Features    │
└───────────────┘      └───────────────┘      └───────────────┘
        │                        │                        │
        │                        │                        │
        ▼                        ▼                        ▼
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│ • quality     │      │ • synclone    │      │ • ide         │
│ • pm          │      │ • repo-config │      │ • profile     │
│ • git (local) │      │ • actions     │      │ • dev-env     │
│               │      │   policy      │      │ • net-env     │
└───────────────┘      └───────────────┘      └───────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│ External      │      │ Git Platform  │      │ System        │
│ Libraries     │      │ APIs          │      │ Integration   │
│               │      │               │      │               │
│ • gzh-cli-    │      │ • GitHub      │      │ • JetBrains   │
│   quality     │      │ • GitLab      │      │ • VS Code     │
│ • gzh-cli-    │      │ • Gitea       │      │ • Docker      │
│   package-    │      │ • Gogs        │      │ • Kubernetes  │
│   manager     │      │               │      │               │
│ • gzh-cli-gitforge │      │               │      │               │
└───────────────┘      └───────────────┘      └───────────────┘
```

## Command Structure

### Integrated Commands (Wrapper Pattern)

```
gz quality          → gzh-cli-quality (45-line wrapper)
  ├─ run           → Multi-language code quality
  ├─ check         → Lint without changes
  ├─ analyze       → Project analysis
  └─ install       → Tool installation

gz pm              → gzh-cli-package-manager (65-line wrapper)
  ├─ update        → Update all package managers
  ├─ status        → Check package manager status
  └─ list          → List installed tools

gz git repo        → gzh-cli-gitforge (473-line wrapper)
  ├─ clone-or-update  → Smart clone/update (rebase, reset, clone, skip, pull, fetch)
  └─ pull-all         → Bulk recursive update
```

### Native gzh-cli Commands

```
gz ide             → IDE monitoring and management
  ├─ scan          → Detect installed IDEs
  ├─ status        → Check IDE status
  ├─ open          → Open project in IDE
  └─ monitor       → Real-time settings monitoring

gz synclone        → Multi-platform repository cloning
  ├─ github        → Clone GitHub organizations
  ├─ gitlab        → Clone GitLab groups
  ├─ gitea         → Clone Gitea organizations
  └─ validate      → Validate configuration

gz profile         → Performance profiling (Go pprof)
  ├─ stats         → Runtime statistics
  ├─ server        → HTTP pprof server
  ├─ cpu           → CPU profiling
  └─ memory        → Memory profiling

gz dev-env         → Development environment management
gz net-env         → Network environment transitions
gz repo-config     → GitHub repository configuration
```

## Directory Structure with Integration

```
gzh-cli/
├── cmd/                           # Command implementations
│   ├── root.go                    # Main CLI entry point
│   │
│   ├── quality_wrapper.go         # → gzh-cli-quality (45 lines)
│   ├── pm_wrapper.go              # → gzh-cli-package-manager (65 lines)
│   │
│   ├── git/                       # Git platform integration
│   │   ├── root.go                # Git command root
│   │   ├── webhook.go             # Webhook management (native)
│   │   ├── event.go               # Event processing (native)
│   │   └── repo/                  # Repository operations
│   │       ├── repo_clone_or_update_wrapper.go  # → gzh-cli-gitforge
│   │       ├── repo_bulk_update_wrapper.go      # → gzh-cli-gitforge
│   │       ├── repo_list.go       # List repos (native - uses platform APIs)
│   │       ├── repo_create.go     # Create repos (native - uses platform APIs)
│   │       └── repo_sync.go       # Sync repos (native - cross-platform)
│   │
│   ├── ide/                       # IDE monitoring (native - gzh-cli specific)
│   │   ├── root.go
│   │   ├── scan.go                # IDE detection
│   │   ├── status.go              # IDE status
│   │   ├── open.go                # Open projects
│   │   └── monitor.go             # Settings monitoring
│   │
│   ├── synclone/                  # Multi-platform cloning (native)
│   ├── profile/                   # Performance profiling (native)
│   ├── dev-env/                   # Dev environment management (native)
│   ├── net-env/                   # Network environment (native)
│   └── repo-config/               # Repo configuration (native)
│
├── internal/                      # Internal packages
│   ├── git/                       # Git operations abstraction
│   ├── logger/                    # Logging abstraction
│   ├── simpleprof/                # Simple profiling
│   └── testlib/                   # Test utilities
│
├── pkg/                           # Public packages
│   ├── github/                    # GitHub API integration
│   ├── gitlab/                    # GitLab API integration
│   ├── gitea/                     # Gitea API integration
│   └── synclone/                  # Clone configuration
│
└── docs/
    └── integration/               # Integration documentation
        ├── README.md              # Integration overview
        ├── ARCHITECTURE.md        # This file
        ├── integration-summary.md # Phase 1-3 summary
        └── git-migration-final-status.md
```

## Integration vs Native Decision Tree

```
                    New Feature Request
                            │
                            ▼
              ┌─────────────────────────┐
              │  Does it duplicate      │
              │  existing code >50%?    │
              └─────────────────────────┘
                     │            │
                   Yes           No
                     │            │
                     ▼            ▼
         ┌──────────────────┐   Keep in
         │  Is it standalone │   gzh-cli
         │  functionality?   │   (native)
         └──────────────────┘
                  │       │
                Yes      No
                  │       │
                  ▼       ▼
         Extract to    Keep in
         external lib  gzh-cli
         (integrate)   (native)

Examples:

Integrated:
✅ Code quality tools (98.7% duplication with standalone tool)
✅ Package manager updates (97.3% duplication)
✅ Git local operations (64.2% duplication)

Native:
❌ IDE monitoring (gzh-cli specific workflow)
❌ synclone (multi-platform orchestration)
❌ Git platform APIs (GitHub/GitLab/Gitea integration)
❌ Network environment switching (system integration)
```

## Architecture Principles

### What to Integrate (External Libraries)

✅ **High code duplication** (>50%)

- Code quality tools: 98.7% overlap
- Package manager updates: 97.3% overlap
- Git local operations: 64.2% overlap

✅ **Clear single responsibility**

- Each library has one clear purpose
- Well-defined scope and boundaries

✅ **Standalone functionality**

- Can be used independently as CLI tool
- Dual usage: standalone + library

✅ **Stable interfaces**

- Minimal breaking changes
- Versioned releases

### What to Keep Native (gzh-cli Specific)

❌ **Low duplication** (\<50%)

- Unique workflows and orchestration
- gzh-cli specific integrations

❌ **Different purposes/goals**

- Platform-specific features (GitHub/GitLab/Gitea APIs)
- Cross-platform synchronization

❌ **High coupling with gzh-cli internals**

- IDE monitoring (JetBrains/VS Code integration)
- Development environment management
- Network environment transitions

❌ **Platform-specific integrations**

- GitHub API operations (list, create, sync)
- GitLab API operations
- Gitea API operations

## Data Flow Examples

### Integrated Command Flow (gz quality run)

```
User Command
    │
    ▼
┌─────────────────────────────────────────┐
│  gz quality run                         │
│  (User executes command)                │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│  cmd/quality_wrapper.go                 │
│  NewQualityCmd()                        │
│  (45-line wrapper)                      │
└─────────────────────────────────────────┘
    │
    │ Delegates to
    ▼
┌─────────────────────────────────────────┐
│  gzh-cli-quality library                │
│  quality.NewRootCmd()                   │
│  (Full implementation: 3,514 lines)     │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│  Execute quality tools:                 │
│  • gofumpt (Go formatting)              │
│  • golangci-lint (Go linting)           │
│  • prettier (JS/TS formatting)          │
│  • eslint (JS/TS linting)               │
│  • ruff (Python)                        │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│  Return results to user                 │
│  (formatted output)                     │
└─────────────────────────────────────────┘
```

### Native Command Flow (gz ide scan)

```
User Command
    │
    ▼
┌─────────────────────────────────────────┐
│  gz ide scan                            │
│  (User executes command)                │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│  cmd/ide/scan.go                        │
│  (Native implementation in gzh-cli)     │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│  Scan system for installed IDEs:        │
│  • JetBrains products                   │
│  • VS Code variants                     │
│  • Other editors                        │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│  Cache results (24-hour TTL)            │
│  Return formatted output                │
└─────────────────────────────────────────┘
```

## Migration History

### Phase 1: Package Manager Integration (2025-11-28)

- Integrated: `gzh-cli-package-manager`
- Code reduced: 2,388 lines (97.3%)
- Wrapper: `cmd/pm_wrapper.go` (65 lines)

### Phase 2: Quality Integration (2025-11-29)

- Integrated: `gzh-cli-quality`
- Code reduced: 3,469 lines (98.7%)
- Wrapper: `cmd/quality_wrapper.go` (45 lines)

### Phase 3: Git Integration (2025-11-30)

- Integrated: `gzh-cli-gitforge`
- Code reduced: 845 lines (64.2%)
- Wrappers: `cmd/git/repo/*_wrapper.go` (473 lines)
- Retained: Git platform API operations (list, sync, create, webhook, event)

### Total Impact

- **Lines reduced**: 6,702
- **Reduction rate**: 92.0%
- **Wrappers total**: 583 lines
- **Net reduction**: 6,119 lines

## Future Architecture

### Planned Improvements

1. **Publish Stable Versions**

   - Remove `replace` directives from go.mod
   - Use versioned releases (v0.1.0, v0.2.0, etc.)
   - Semantic versioning for breaking changes

1. **Wrapper Tests**

   - Add unit tests for wrapper functions
   - Test integration with external libraries
   - Verify command registration

1. **Architecture Documentation**

   - Create visual diagrams (Mermaid, PlantUML)
   - Document decision records (ADRs)
   - Update architecture as it evolves

1. **CI/CD Integration**

   - Automated testing of integrated libraries
   - Version compatibility checks
   - Integration test suite

## Related Documentation

- [Integration Overview](README.md)
- [Integration Summary](integration-summary.md)
- [Git Migration Status](git-migration-final-status.md)
- [Integration Tests](INTEGRATION_TESTS.md)

______________________________________________________________________

**Last Updated**: 2025-12-01
**Status**: Phase 1-3 Complete, Architecture Documented
**Model**: claude-sonnet-4-5-20250929
