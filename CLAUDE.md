# CLAUDE.md

This file provides LLM-optimized guidance for Claude Code when working with this repository.

______________________________________________________________________

## Quick Start (30s scan)

**Binary**: `gz` (not `gzh-cli`)
**Architecture**: Integration Libraries Pattern (wrapper-based integration)
**Go Version**: 1.23+
**Main Branch**: `master` (for PRs)

Core principle: Interface-driven design with direct constructors. External libraries via thin wrappers.

______________________________________________________________________

## Top 10 Commands

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `make bootstrap` | One-time setup | First time only |
| `make fmt && make lint` | Format + lint | Before every commit |
| `make test` | All tests | Pre-commit validation |
| `make build` | Build `gz` binary | After changes |
| `make install` | Install to $GOPATH/bin | Local testing |
| `make cover` | Coverage report | Check coverage |
| `make generate-mocks` | Generate test mocks | After interface changes |
| `go test ./cmd/{module} -v` | Test specific module | Focused testing |
| `make clean` | Clean artifacts | Fresh start |
| `make dev` | Quick dev cycle | Rapid iteration |

______________________________________________________________________

## Absolute Rules (DO/DON'T)

### DO

- ✅ Read `cmd/AGENTS_COMMON.md` before ANY modification
- ✅ Read `cmd/{module}/AGENTS.md` for module-specific rules
- ✅ Run `make fmt && make lint` before every commit
- ✅ Use Korean comments for new code
- ✅ Maintain 80%+ test coverage for core logic
- ✅ Check wrapper vs library decision (see Context Docs)

### DON'T

- ❌ Use binary name `gzh-cli` (correct: `gz`)
- ❌ Over-engineer (see `cmd/AGENTS_COMMON.md`)
- ❌ Skip reading AGENTS.md files
- ❌ Commit without fmt + lint + test
- ❌ Modify core logic in wrappers (use external libraries)

______________________________________________________________________

## Directory Structure

```
.
├── cmd/                    # CLI commands
│   ├── root.go            # Main CLI entry
│   ├── *_wrapper.go       # Integration library wrappers
│   ├── git/               # Git platform integration
│   ├── ide/               # IDE management
│   └── */AGENTS.md        # Module-specific guides (READ THESE!)
├── internal/              # Private abstractions
│   ├── git/               # Git operations (interfaces)
│   ├── logger/            # Logging abstractions
│   └── cli/               # Command builder utilities
├── pkg/                   # Public APIs
│   ├── github/            # GitHub API integration
│   ├── gitlab/            # GitLab API integration
│   └── gitea/             # Gitea API integration
├── docs/
│   └── .claude-context/   # Context docs (see below)
└── .make/                 # Modular Makefile (7 modules)
```

______________________________________________________________________

## Context Documentation

| Guide | Purpose |
|-------|---------|
| [Architecture Guide](docs/.claude-context/architecture-guide.md) | Integration pattern, extensions, lifecycle |
| [Testing Guide](docs/.claude-context/testing-guide.md) | Test organization, mocking, coverage |
| [Build Guide](docs/.claude-context/build-guide.md) | Build workflow, troubleshooting |
| [Common Tasks](docs/.claude-context/common-tasks.md) | Adding commands, modifying wrappers |

**CRITICAL**: Read module-specific guides:

- `cmd/AGENTS_COMMON.md` - Project-wide conventions
- `cmd/{module}/AGENTS.md` - Module-specific rules (15 files)

______________________________________________________________________

## Common Mistakes (Top 3)

1. **Not reading AGENTS.md before modifying code**

   - ⚠️ Will miss critical module-specific rules
   - ✅ Always check: `cmd/AGENTS_COMMON.md` + `cmd/{module}/AGENTS.md`

1. **Modifying core logic in wrapper files**

   - ⚠️ Breaks separation of concerns
   - ✅ Core logic → External library, Integration → Wrapper

1. **Skipping `make fmt && make lint` before commit**

   - ⚠️ CI will fail
   - ✅ Run before every commit

______________________________________________________________________

## Integration Libraries

**IMPORTANT**: Core features live in external libraries. Wrappers provide CLI integration.

| Library | Wrapper | Purpose |
|---------|---------|---------|
| gzh-cli-git | `cmd/git/repo/*_wrapper.go` | Local Git operations |
| gzh-cli-quality | `cmd/quality_wrapper.go` | Code quality tools |
| gzh-cli-package-manager | `cmd/pm_wrapper.go` | Package manager |
| gzh-cli-shellforge | `cmd/shellforge_wrapper.go` | Shell config builder |
| gzh-cli-dev-env | `cmd/dev_env_wrapper.go` | Dev environment management |

### gzh-cli-dev-env Integration

The dev-env command uses a **hybrid integration** approach:

- **Current mode**: Old `cmd/dev-env/` provides most subcommands, library provides `status`, `tui`, `switch-all`
- **Future mode**: Build with `-tags devenv_external` for full wrapper mode

```bash
# Current (default) - hybrid mode
make build

# Future migration - full wrapper mode
go build -tags devenv_external
```

**Wrapper files**:

- `cmd/dev_env_wrapper.go` - Full wrapper (enabled with build tag)
- `cmd/dev_env_wrapper_stub.go` - Documentation for migration

**Decision**: Core logic change → External library. CLI integration → Wrapper.

______________________________________________________________________

## Git Commit Format

```
{type}({scope}): {description}

{body}

Model: claude-{model}
Co-Authored-By: Claude <noreply@anthropic.com>
```

**Types**: feat, fix, docs, refactor, test, chore
**Scope**: REQUIRED (e.g., git, ide, quality, docs)

______________________________________________________________________

## Command Categories

### Repository Operations

- `gz git repo clone-or-update` - Smart single repo management
- `gz git repo pull-all` - Bulk recursive update
- `gz synclone` - Mass organization cloning

### Development Environment

- `gz dev-env` - Environment management
- `gz pm` - Package manager updates
- `gz ide` - IDE monitoring

### Code Quality

- `gz quality` - Multi-language quality tools
- `gz profile` - Go pprof performance profiling

### Network & Shell

- `gz net-env` - Network environment transitions
- `gz shellforge` - Modular shell config builder

______________________________________________________________________

## FAQ

**Q: Should I modify wrapper or external library?**
A: **Core logic** → External library. **CLI integration** → Wrapper.

**Q: Where to add new Git platform (e.g., Bitbucket)?**
A: `pkg/bitbucket/` for API, register in provider registry.

**Q: How to handle secrets in tests?**
A: Use environment variables, skip tests if not available.

**Q: Where are AGENTS.md files?**
A: `cmd/AGENTS_COMMON.md` + `cmd/{module}/AGENTS.md` (15 modules).

______________________________________________________________________

## Performance Targets

- **Startup time**: \<50ms
- **Binary size**: ~33MB
- **Command response**: \<100ms for most commands

______________________________________________________________________

**Last Updated**: 2025-12-26
**Previous**: 165 lines → **Current**: ~185 lines (added dev-env wrapper docs)
