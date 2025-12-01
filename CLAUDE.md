# CLAUDE.md

This file provides LLM-optimized guidance for Claude Code when working with this repository.

---

## Project Context (Quick Overview)

**Binary**: `gz` (not `gzh-cli`)
**Architecture**: Integration Libraries Pattern (wrapper-based integration)
**Go Version**: 1.23+
**Main Branch**: `master` (for PRs)

### Core Principles
- **Interface-driven design**: Heavy use of Go interfaces for abstraction
- **Direct constructors**: No DI containers, simple factory pattern
- **Modular commands**: Each command is self-contained under `cmd/`
- **Integration pattern**: External libraries via thin wrappers

---

## Development Workflow (LLM Task Guide)

### Before Code Modification
1. **Read module's AGENTS.md**: `cmd/{module}/AGENTS.md` for module-specific rules
2. **Check common rules**: `cmd/AGENTS_COMMON.md` for project-wide conventions
3. **Review existing patterns**: Understand current implementation before changes

### Code Modification Process
```bash
# 1. Read relevant AGENTS.md
# 2. Write code + tests
# 3. Quality checks (CRITICAL)
make fmt && make lint && make test
# 4. Commit with proper message format
```

### Module-Specific Guides
- [Common Guidelines](cmd/AGENTS_COMMON.md) - **Read first**
- [git module](cmd/git/AGENTS.md)
- [ide module](cmd/ide/AGENTS.md)
- [quality module](cmd/quality/AGENTS.md)
- See `cmd/*/AGENTS.md` for other modules (15 files total)

---

## Essential Commands Reference

### Development Workflow
```bash
# One-time setup
make bootstrap

# Before every commit (CRITICAL)
make fmt && make lint && make test

# Build & install
make build
make install
```

### Testing
```bash
# All tests
make test

# Specific package
go test ./cmd/{module} -v

# Specific test function
go test ./cmd/git -run "TestCloneOrUpdate" -v

# Coverage
make cover
```

### Code Quality (Pre-commit REQUIRED)
```bash
make fmt        # gofumpt + gci
make lint       # golangci-lint with auto-fix
make lint-all   # format + lint + pre-commit
```

### Mocking
```bash
make generate-mocks    # Generate mocks using gomock
make regenerate-mocks  # Clean + regenerate
```

---

## Architecture Patterns (LLM Decision Guide)

### Key Patterns to Use

#### 1. Interface Abstraction
```go
// internal/git/interfaces.go
type Client interface {
    Clone(ctx context.Context, options CloneOptions) error
    Pull(ctx context.Context, options PullOptions) error
}
```

#### 2. Provider Registry
```go
// pkg/git/provider/
providerRegistry := provider.NewRegistry()
providerRegistry.Register("github", github.NewProvider())
```

#### 3. Strategy Pattern (Git Operations)
- `rebase`: Rebase local changes on remote
- `reset`: Hard reset to match remote
- `clone`: Fresh clone (remove existing)
- `pull`: Standard git pull (merge)
- `fetch`: Update refs only

### Integration Libraries (External Dependencies)

**IMPORTANT**: These features are implemented in external libraries, not gzh-cli.

| Library | Wrapper Location | Lines | Purpose |
|---------|-----------------|-------|---------|
| [gzh-cli-git][git] | `cmd/git/repo/*_wrapper.go` | 473 | Local Git operations |
| [gzh-cli-quality][quality] | `cmd/quality_wrapper.go` | 45 | Code quality tools |
| [gzh-cli-package-manager][pm] | `cmd/pm_wrapper.go` | 65 | Package manager integration |
| [gzh-cli-shellforge][shell] | `cmd/shellforge_wrapper.go` | 71 | Shell config builder |

[git]: https://github.com/gizzahub/gzh-cli-git
[quality]: https://github.com/Gizzahub/gzh-cli-quality
[pm]: https://github.com/gizzahub/gzh-cli-package-manager
[shell]: https://github.com/gizzahub/gzh-cli-shellforge

**When to Modify**:
- **Core logic change** → Modify external library repository
- **CLI integration change** → Modify wrapper file in gzh-cli

**Local Development**:
```go
// go.mod uses replace directives for local dev
replace github.com/gizzahub/gzh-cli-git => ../gzh-cli-git
```

---

## Project Structure (Quick Reference)

```
.
├── cmd/                    # CLI commands
│   ├── root.go            # Main CLI entry
│   ├── *_wrapper.go       # Integration library wrappers
│   ├── git/               # Git platform integration
│   ├── ide/               # IDE management
│   ├── quality/           # Code quality (DEPRECATED, use wrapper)
│   └── */AGENTS.md        # Module-specific guides
├── internal/              # Private abstractions
│   ├── git/               # Git operations (interfaces, constructors)
│   ├── logger/            # Logging abstractions
│   └── cli/               # Command builder utilities
├── pkg/                   # Public APIs
│   ├── github/            # GitHub API integration
│   ├── gitlab/            # GitLab API integration
│   └── gitea/             # Gitea API integration
├── docs/                  # User documentation
└── .make/                 # Modular Makefile (7 modules)
```

---

## Configuration Architecture

### Configuration Hierarchy
1. `$GZH_CONFIG_PATH` (env var)
2. `./gzh.yaml` (current dir)
3. `~/.config/gzh-manager/gzh.yaml` (user config)
4. `/etc/gzh-manager/gzh.yaml` (system config)

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

---

## Important Rules (LLM Must Follow)

### Critical Requirements
- ✅ **Always run** `make fmt && make lint` before commit
- ✅ **Korean comments** for new code
- ✅ **Read AGENTS.md** before modifying any module
- ✅ **Test coverage**: 80%+ for core logic
- ✅ **Commit scope**: Mandatory in commit messages

### Code Style
- **Binary name**: `gz` (never `gzh-cli`)
- **No over-engineering**: See `cmd/AGENTS_COMMON.md`
- **Interface-driven**: Use interfaces for testability
- **Error handling**: Structured errors with context

### Testing
- **Unit tests**: `*_test.go` alongside source
- **Mocking**: Use `gomock` for external dependencies
- **Integration tests**: `test/integration/` with Docker
- **Environment-specific**: Check for tokens (GITHUB_TOKEN, etc.)

### Commit Format
```
{type}({scope}): {description}

{body}

Model: claude-{model}
Co-Authored-By: Claude <noreply@anthropic.com>
```

**Types**: feat, fix, docs, refactor, test, chore
**Scope**: REQUIRED (e.g., git, ide, quality, docs)

---

## Command Categories (Quick Lookup)

### Repository Operations
- `gz git repo clone-or-update` - Smart single repo management
- `gz git repo pull-all` - Bulk recursive update
- `gz synclone` - Mass organization cloning
- `gz repo-config` - GitHub repo configuration

### Development Environment
- `gz dev-env` - Environment management
- `gz pm` - Package manager updates
- `gz ide` - IDE monitoring (JetBrains, VS Code)
- `gz doctor` - System diagnostics

### Code Quality & Performance
- `gz quality` - Multi-language quality tools
- `gz profile` - Go pprof performance profiling

### Network & Shell
- `gz net-env` - Network environment transitions
- `gz shellforge` - Modular shell config builder

---

## Common Tasks (LLM Quick Guide)

### Adding a New Command
1. Check if feature belongs in external library or gzh-cli
2. Create `cmd/{command}/` directory
3. Add `cmd/{command}/AGENTS.md` with module rules
4. Implement using Cobra framework
5. Register in `cmd/root.go`
6. Add tests: `cmd/{command}/*_test.go`
7. Update docs: `docs/30-features/`

### Modifying Integration Library Command
1. **Check wrapper**: `cmd/*_wrapper.go` or `cmd/{module}/*_wrapper.go`
2. **Core logic**: Modify in external library repository
3. **Integration**: Modify wrapper if needed
4. **Local test**: Use `replace` directive in go.mod

### Adding Tests
```bash
# Create test file
touch cmd/{module}/{feature}_test.go

# Run tests
go test ./cmd/{module} -v

# Check coverage
go test ./cmd/{module} -cover
```

---

## Troubleshooting (LLM Quick Fix)

### Build Issues
```bash
make clean
make bootstrap
make build
```

### Lint Failures
```bash
make fmt        # Fix formatting
make lint       # Auto-fix linting issues
```

### Test Failures
```bash
# Run specific test with verbose
go test ./cmd/{module} -run "TestName" -v

# Check for race conditions
go test ./cmd/{module} -race
```

### Import Cycle
- **Cause**: Circular dependencies between packages
- **Fix**: Move shared types to `internal/` or `pkg/`

---

## FAQ for LLMs

**Q: Should I modify wrapper or external library?**
A: **Core logic** → External library. **CLI integration** → Wrapper.

**Q: Where to add new Git platform (e.g., Bitbucket)?**
A: `pkg/bitbucket/` for API, register in provider registry.

**Q: How to handle secrets in tests?**
A: Use environment variables, skip tests if not available.

**Q: Can I add dependencies?**
A: Yes, but prefer standard library. Run `go mod tidy` after.

**Q: Where are integration tests?**
A: `test/integration/` with Docker containers.

---

## Performance Benchmarks

### Target Metrics
- **Startup time**: <50ms
- **Binary size**: ~33MB
- **Memory**: Minimal footprint
- **Command response**: <100ms for most commands

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

---

**Last Updated**: 2025-12-01
**Line Count**: ~290 lines (38% reduction from 474 lines)
**Optimization**: LLM-focused, removed user-facing content
