# 🔧 REFACTORING_CHECKLIST.md

> **Structured Go Refactoring Checklist for gzh-manager-go**  
> Based on current codebase analysis and Go best practices

---

## 📋 1. Code Quality & Formatting

### Basic Code Hygiene
- [x] **Remove unused imports and variables across all packages**
  - 📌 **Why**: Current codebase shows some packages with potentially unused imports
  - 🧠 **How**: Run `goimports -w .` and `golangci-lint run --enable=unused`
  - 📁 **Files**: All `*.go` files, particularly `cmd/` and `pkg/` directories

- [x] **Standardize naming conventions (eliminate mixed camelCase/snake_case)**
  - 📌 **Why**: Some struct fields and variables use inconsistent naming
  - 🧠 **How**: Use `golangci-lint run --enable=stylecheck,golint` to identify issues
  - 📁 **Files**: `pkg/github/interfaces.go`, `internal/` packages

- [x] **Ensure all exported functions/types have godoc comments**
  - 📌 **Why**: Many public interfaces lack proper documentation
  - 🧠 **How**: Add docstrings following Go conventions for all exported symbols
  - 📁 **Files**: `pkg/github/interfaces.go`, `pkg/config/`, `pkg/synclone/`

- [x] **Fix inconsistent error message formatting**
  - 📌 **Why**: Error messages should be lowercase and avoid redundant "error:" prefixes
  - 🧠 **How**: Use `errors.New("something failed")` not `errors.New("Error: something failed")`
  - 📁 **Files**: All error returns in `cmd/` and `pkg/` packages

## 📦 2. Code Structure & Architecture

### Package Organization
- [x] **Move helper functions from `helpers/` to appropriate internal packages**
  - 📌 **Why**: `helpers/` directory indicates poor separation of concerns
  - 🧠 **How**: Move Git-related helpers to `internal/git/`, platform helpers to `internal/platform/`
  - 📁 **Files**: `helpers/git_helper.go` → `internal/git/helpers.go`

- [x] **Consolidate duplicate configuration loading logic**
  - 📌 **Why**: Multiple config loaders exist in `pkg/config/` and `internal/config/`
  - 🧠 **How**: Create single config factory with dependency injection
  - 📁 **Files**: `pkg/config/loader.go`, `internal/config/`, `cmd/synclone/config.go`

- [x] **Extract common CLI patterns from command implementations**
  - 📌 **Why**: Commands have duplicated flag handling and validation logic
  - 🧠 **How**: Create `internal/cli/` package with common command builders
  - 📁 **Files**: `cmd/synclone/synclone.go`, `cmd/repo-config/`, all command files

- [x] **Separate business logic from CLI handlers**
  - 📌 **Why**: Command files contain too much business logic (violation of SRP)
  - 🧠 **How**: Extract logic to service layer in `internal/services/`
  - 📁 **Files**: `cmd/*/` → move logic to `internal/services/synclone.go`, etc.

### Main Function Isolation
- [x] **Reduce main.go responsibilities to absolute minimum**
  - 📌 **Why**: Current `main.go` handles signal management, should only bootstrap
  - 🧠 **How**: Move signal handling to `internal/app/` package
  - 📁 **Files**: `main.go` → extract to `internal/app/runner.go`

## 🔌 3. Interface Design & Dependency Management

### Interface Consistency
- [ ] **Standardize context.Context usage across all interfaces**
  - 📌 **Why**: Some interfaces don't consistently use context for cancellation
  - 🧠 **How**: Ensure all I/O operations accept `ctx context.Context` as first parameter
  - 📁 **Files**: `pkg/github/interfaces.go`, `pkg/gitlab/interfaces.go`

- [ ] **Create unified Git platform interface**
  - 📌 **Why**: GitHub, GitLab, Gitea have similar but inconsistent interfaces
  - 🧠 **How**: Define common `GitPlatform` interface in `pkg/git/provider/`
  - 📁 **Files**: Create `pkg/git/provider/interface.go`, refactor platform packages

- [ ] **Implement proper dependency injection container**
  - 📌 **Why**: Hard-coded dependencies make testing difficult
  - 🧠 **How**: Use `wire` or create simple DI container in `internal/container/`
  - 📁 **Files**: Create `internal/container/container.go`, update `cmd/root.go`

- [ ] **Add interface compliance verification**
  - 📌 **Why**: No compile-time verification that structs implement interfaces
  - 🧠 **How**: Add `var _ Interface = (*Implementation)(nil)` in implementation files
  - 📁 **Files**: All implementation files in `pkg/github/`, `pkg/gitlab/`, etc.

## 🔄 4. Concurrency & Goroutine Safety

### Thread Safety
- [ ] **Audit shared state for race conditions**
  - 📌 **Why**: Worker pools and rate limiters may have race conditions
  - 🧠 **How**: Run `go test -race ./...` and add proper synchronization
  - 📁 **Files**: `internal/workerpool/pool.go`, rate limiter implementations

- [ ] **Implement proper goroutine lifecycle management**
  - 📌 **Why**: No clear pattern for goroutine cleanup on context cancellation
  - 🧠 **How**: Use `errgroup.Group` for managed goroutine pools
  - 📁 **Files**: `internal/workerpool/`, `pkg/github/bulk_operations.go`

- [ ] **Add timeout handling for all external API calls**
  - 📌 **Why**: HTTP clients may hang indefinitely without timeouts
  - 🧠 **How**: Set context deadlines for all HTTP operations
  - 📁 **Files**: `pkg/github/http_adapter.go`, `pkg/gitlab/http_adapter.go`

- [ ] **Review channel usage patterns for potential deadlocks**
  - 📌 **Why**: Unbuffered channels used without proper goroutine coordination
  - 🧠 **How**: Use buffered channels or ensure proper sender/receiver matching
  - 📁 **Files**: `internal/workerpool/pool.go`, progress reporting code

## ⚙️ 5. Configuration & Environment Separation

### Configuration Management
- [ ] **Validate configuration schema at startup**
  - 📌 **Why**: Invalid configs cause runtime failures instead of early validation
  - 🧠 **How**: Implement JSON schema validation using `github.com/xeipuuv/gojsonschema`
  - 📁 **Files**: `pkg/config/validator.go`, add schema files to `docs/schemas/`

- [ ] **Implement configuration hot-reloading**
  - 📌 **Why**: Long-running operations need config updates without restart
  - 🧠 **How**: Use `fsnotify` to watch config files and reload safely
  - 📁 **Files**: `pkg/config/hot_reload.go`, integrate in service layer

- [ ] **Separate environment-specific configuration**
  - 📌 **Why**: Development/production configs mixed with business logic
  - 🧠 **How**: Create `configs/` directory with environment-specific files
  - 📁 **Files**: Create `configs/dev.yaml`, `configs/prod.yaml`

- [ ] **Add configuration migration system**
  - 📌 **Why**: Config format changes break existing setups
  - 🧠 **How**: Implement version-aware config migration in `pkg/config/migration.go`
  - 📁 **Files**: `pkg/config/migration.go`, version detection logic

### Environment Variables
- [ ] **Standardize environment variable naming**
  - 📌 **Why**: Inconsistent env var prefixes (GZH_, GITHUB_, etc.)
  - 🧠 **How**: Use consistent `GZH_` prefix, document in `internal/env/`
  - 📁 **Files**: `internal/env/keys.go`, update all env var usage

## 🧪 6. Testing Strategy

### Test Coverage
- [ ] **Achieve >80% test coverage for core business logic**
  - 📌 **Why**: Critical functionality lacks adequate test coverage
  - 🧠 **How**: Write unit tests for `internal/` and `pkg/` packages
  - 📁 **Files**: Add `*_test.go` files for all packages, focus on `pkg/github/`, `pkg/config/`

- [ ] **Implement table-driven tests for repetitive test cases**
  - 📌 **Why**: Many functions have similar test patterns that could be parameterized
  - 🧠 **How**: Convert to `[]struct{name, input, expected}` pattern
  - 📁 **Files**: Configuration validation tests, URL building tests

- [ ] **Create integration test suite with real Git repositories**
  - 📌 **Why**: Current tests mock too much, missing real integration issues
  - 🧠 **How**: Use `testcontainers-go` to spin up real Git servers
  - 📁 **Files**: `test/integration/` directory, Docker compose files

- [ ] **Add benchmark tests for performance-critical paths**
  - 📌 **Why**: No performance regression detection for bulk operations
  - 🧠 **How**: Create `BenchmarkXxx` functions for clone operations, rate limiting
  - 📁 **Files**: `pkg/github/bulk_operations_test.go`, worker pool benchmarks

### Mock Strategy
- [ ] **Generate mocks using gomock for all interfaces**
  - 📌 **Why**: Manual mocks are inconsistent and hard to maintain
  - 🧠 **How**: Add `//go:generate mockgen` directives, create `make generate-mocks`
  - 📁 **Files**: `pkg/github/interfaces.go`, create `pkg/github/mocks/`

- [ ] **Create test fixtures for common scenarios**
  - 📌 **Why**: Tests recreate the same test data repeatedly
  - 🧠 **How**: Create `internal/testutil/fixtures/` with standard test data
  - 📁 **Files**: `internal/testutil/fixtures/github.go`, configuration fixtures

## 🔧 7. Tooling & Automation

### Build System
- [ ] **Optimize Makefile for development workflow**
  - 📌 **Why**: Current Makefile is complex but lacks common dev tasks
  - 🧠 **How**: Add `make dev`, `make watch`, `make clean-deps` targets
  - 📁 **Files**: `Makefile`, consider splitting into `Makefile.dev.mk`

- [ ] **Add pre-commit hooks for consistent code quality**
  - 📌 **Why**: Manual linting leads to inconsistent code quality
  - 🧠 **How**: Setup `pre-commit` with `gofmt`, `golangci-lint`, `go mod tidy`
  - 📁 **Files**: `.pre-commit-config.yaml`, `scripts/pre-commit.sh`

- [ ] **Implement reproducible builds**
  - 📌 **Why**: Build artifacts vary between environments
  - 🧠 **How**: Pin all tool versions, use `go.mod` vendor directory
  - 📁 **Files**: `tools/tools.go`, `go.mod` with specific versions

- [ ] **Add security scanning to CI pipeline**
  - 📌 **Why**: No automated security vulnerability detection
  - 🧠 **How**: Integrate `govulncheck`, `nancy`, `semgrep` in GitHub Actions
  - 📁 **Files**: `.github/workflows/security.yml`

### Linting Configuration
- [ ] **Tune golangci-lint configuration for project needs**
  - 📌 **Why**: Default linting rules may not fit project requirements
  - 🧠 **How**: Create `.golangci.yml` with appropriate rules enabled/disabled
  - 📁 **Files**: `.golangci.yml`, adjust for false positives

## 📚 8. Documentation & API Reference

### Code Documentation
- [ ] **Generate and publish godoc documentation**
  - 📌 **Why**: Public packages need searchable documentation
  - 🧠 **How**: Setup `godoc` server, ensure pkg.go.dev indexing
  - 📁 **Files**: All public packages, add package-level documentation

- [ ] **Create architecture decision records (ADRs)**
  - 📌 **Why**: Design decisions are not documented for future reference
  - 🧠 **How**: Create `docs/adr/` directory with numbered ADR files
  - 📁 **Files**: `docs/adr/001-cli-framework-choice.md`, etc.

- [ ] **Add example configurations for common use cases**
  - 📌 **Why**: Complex configuration format needs more examples
  - 🧠 **How**: Expand `examples/` directory with real-world scenarios
  - 📁 **Files**: `examples/enterprise/`, `examples/simple/`, `examples/advanced/`

- [ ] **Document configuration schema with examples**
  - 📌 **Why**: YAML configuration lacks inline documentation
  - 🧠 **How**: Add comprehensive comments to example configs, JSON schema docs
  - 📁 **Files**: `docs/config-reference.md`, annotated example files

---

## 🎯 3. Refactoring Execution Plan

### Phase 1: Foundation (Week 1-2)
1. Fix basic code quality issues (formatting, unused imports, naming)
2. Add missing godoc comments and error message standardization
3. Setup proper linting configuration and pre-commit hooks
4. Implement interface compliance verification

### Phase 2: Architecture (Week 3-4)
5. Extract business logic from CLI handlers to service layer
6. Consolidate configuration loading logic
7. Create unified Git platform interface
8. Implement dependency injection container

### Phase 3: Reliability (Week 5-6)
9. Add comprehensive test coverage with mocks and fixtures
10. Implement proper goroutine lifecycle management
11. Add timeout handling and race condition fixes
12. Create integration test suite

### Phase 4: Production Readiness (Week 7-8)
13. Implement configuration validation and hot-reloading
14. Add security scanning and reproducible builds
15. Create comprehensive documentation and examples
16. Performance benchmarks and optimization

---

## 🧪 4. Testing Scope

### Unit Testing Requirements
- **Core Logic**: `pkg/config/`, `pkg/github/`, `pkg/synclone/` - 90%+ coverage
- **Utilities**: `internal/git/`, `internal/workerpool/` - 80%+ coverage
- **Commands**: `cmd/*/` handlers - 70%+ coverage (focus on business logic)

### Integration Testing Scope
- **Configuration Loading**: Test all config file formats and precedence
- **Git Operations**: Clone, pull, fetch operations against real repositories
- **Platform APIs**: GitHub, GitLab, Gitea API integration with rate limiting
- **CLI Workflows**: End-to-end command execution with various flag combinations

### Manual Testing Requirements
- **Cross-platform Compatibility**: Test on Linux, macOS, Windows
- **Large Repository Sets**: Test with 100+ repository organizations
- **Network Failure Scenarios**: Test with intermittent network issues
- **Configuration Edge Cases**: Test with malformed configs, missing tokens

---

## 🔍 5. Risk Assessment

### High Risk Refactoring Items
- **Goroutine Safety**: Changing concurrency patterns may introduce deadlocks
- **Configuration Breaking Changes**: Config schema changes affect all users
- **Interface Changes**: Public API modifications break downstream usage

### Medium Risk Items
- **Package Restructuring**: Import path changes require careful coordination
- **CLI Flag Changes**: Command-line interface modifications affect scripts
- **Error Message Changes**: May break log parsing scripts

### Low Risk Items
- **Code Formatting**: No functional impact
- **Documentation**: Only improves developer experience
- **Test Coverage**: Improves reliability without changing behavior

---

**📅 Estimated Timeline**: 8 weeks for complete refactoring  
**👥 Recommended Team Size**: 2-3 developers  
**🎯 Success Metrics**: >80% test coverage, <100ms average command startup time, zero security vulnerabilities