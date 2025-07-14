# REFACTORING_CHECKLIST.md

## 🎯 Go Project Refactoring Checklist

### 1. 📋 Code Quality

#### Package Naming & Structure
- [x] Rename packages to remove underscores (e.g., `bulk_clone` → `bulkclone`)
  - 📌 **Why**: Go convention prefers camelCase or single words for package names
  - 🧠 **How**: Use `gofmt -r` or manual rename, update all imports
  - 📁 **Files**: All packages under `cmd/` and `pkg/` directories

- [x] Move `helpers/` to `internal/helpers/`
  - 📌 **Why**: Helper packages should be internal to prevent external dependencies
  - 🧠 **How**: `git mv helpers/ internal/helpers/`, update imports
  - 📁 **Files**: `helpers/git_helper.go`, all files importing helpers

- [x] Fix import aliases to follow Go conventions
  - 📌 **Why**: Import aliases should not use underscores
  - 🧠 **How**: Update imports like `net_env` to `netenv`
  - 📁 **Files**: All `.go` files with aliased imports

#### Dead Code Removal
- [x] Complete or remove Gogs implementation
  - 📌 **Why**: Incomplete implementation with TODO comments
  - 🧠 **How**: Either implement fully or remove the feature
  - 📁 **Files**: `cmd/bulk-clone/bulk_clone_gogs.go`, `pkg/gogs/`

- [x] Remove or implement TODO/FIXME items
  - 📌 **Why**: 7 files contain incomplete implementations
  - 🧠 **How**: Review each TODO, implement or create issues
  - 📁 **Files**: Use `grep -r "TODO\|FIXME" .` to find all instances

#### Code Formatting & Linting
- [x] Enable additional golangci-lint linters
  - 📌 **Why**: Many useful linters are disabled (gosec, dupl, gocyclo)
  - 🧠 **How**: Update `.golangci.yml`, fix issues incrementally
  - 📁 **Files**: `.golangci.yml`, then run `make lint`

- [x] Fix all existing linter warnings
  - 📌 **Why**: Clean code base improves maintainability
  - 🧠 **How**: Run `golangci-lint run --fix` for auto-fixes
  - 📁 **Files**: All `.go` files with warnings

### 2. 📦 Code Structure

#### Interface Design
- [x] Create service interfaces for all major components
  - 📌 **Why**: Only 3 interfaces exist, limiting testability
  - 🧠 **How**: Extract interfaces from concrete types
  - 📁 **Files**: Create `interfaces.go` in each package

- [x] Define file system abstraction interface
  - 📌 **Why**: Direct file operations make testing difficult
  - 🧠 **How**: Create `type FileSystem interface` with methods
  - 📁 **Files**: `internal/filesystem/interfaces.go`

- [x] Create HTTP client interface
  - 📌 **Why**: Direct HTTP calls are hard to mock
  - 🧠 **How**: Define `type HTTPClient interface`
  - 📁 **Files**: `internal/httpclient/interfaces.go`

#### Dependency Injection
- [x] Implement constructor functions with dependencies
  - 📌 **Why**: Current code uses global state and direct instantiation
  - 🧠 **How**: Add `New*` functions accepting interfaces
  - 📁 **Files**: All service implementations

- [x] Remove direct environment variable access from packages
  - 📌 **Why**: Tight coupling to environment
  - 🧠 **How**: Pass config through constructors
  - 📁 **Files**: All files using `os.Getenv()`

- [x] Create factory pattern for provider instantiation
  - 📌 **Why**: Commands directly create provider instances
  - 🧠 **How**: Implement `ProviderFactory` interface
  - 📁 **Files**: `pkg/*/factory.go`

### 3. 🔧 Interface Design & Dependency Management

#### API Surface Reduction
- [x] Make internal types unexported
  - 📌 **Why**: Large public API surface increases maintenance
  - 🧠 **How**: Lowercase first letter of internal types
  - 📁 **Files**: Review all exported types in `pkg/`

- [x] Create facade interfaces for complex operations
  - 📌 **Why**: Simplify API usage and hide implementation
  - 🧠 **How**: Define high-level operation interfaces
  - 📁 **Files**: `pkg/*/facade.go`

#### Package Boundaries
- [x] Define clear package responsibilities
  - 📌 **Why**: Some packages have mixed concerns
  - 🧠 **How**: Document package purpose in `doc.go`
  - 📁 **Files**: Add `doc.go` to each package

- [x] Reduce inter-package dependencies
  - 📌 **Why**: High coupling between packages
  - 🧠 **How**: Use interfaces at package boundaries
  - 📁 **Files**: Review imports in each package

### 4. 🔄 Concurrency & Goroutine Safety

#### Context Propagation
- [x] Add context.Context to all long-running operations
  - 📌 **Why**: No cancellation support currently
  - 🧠 **How**: Add `ctx context.Context` as first parameter
  - 📁 **Files**: All functions doing I/O or network calls

- [x] Implement graceful shutdown
  - 📌 **Why**: No cleanup on interrupt signals
  - 🧠 **How**: Use `signal.Notify` and context cancellation
  - 📁 **Files**: `cmd/root.go`, all command files

#### Structured Concurrency
- [x] Replace raw goroutines with errgroup
  - 📌 **Why**: Better error handling and synchronization
  - 🧠 **How**: Use `golang.org/x/sync/errgroup`
  - 📁 **Files**: `cmd/net-env/`, any concurrent operations

- [x] Implement worker pool for bulk operations
  - 📌 **Why**: Unbounded concurrency can overwhelm resources
  - 🧠 **How**: Use semaphore or channel-based pool
  - 📁 **Files**: `pkg/github/`, `pkg/gitlab/` bulk operations

### 5. ⚙️ Configuration & Environment Separation

#### Unified Configuration
- [x] Merge bulk-clone.yaml and gzh.yaml formats
  - 📌 **Why**: Duplicate configuration approaches
  - 🧠 **How**: Create single schema, migration tool
  - 📁 **Files**: `pkg/config/`, configuration schemas

- [x] Implement central configuration service
  - 📌 **Why**: Configuration loading duplicated across commands
  - 🧠 **How**: Create `ConfigService` with Viper
  - 📁 **Files**: `internal/config/service.go`

- [x] Add configuration validation at startup
  - 📌 **Why**: Runtime failures from bad config
  - 🧠 **How**: Use validator tags and custom rules
  - 📁 **Files**: All config struct definitions

#### Environment Management
- [x] Create environment abstraction layer
  - 📌 **Why**: Direct os.Getenv calls throughout
  - 🧠 **How**: Define `Environment` interface
  - 📁 **Files**: `internal/env/environment.go`

- [x] Implement configuration hot-reloading
  - 📌 **Why**: Restart required for config changes
  - 🧠 **How**: Use fsnotify with Viper
  - 📁 **Files**: Configuration service implementation

### 6. 🧪 Testing

#### Test Infrastructure
- [x] Create test fixtures and builders
  - 📌 **Why**: Test data creation is repetitive
  - 🧠 **How**: Implement builder pattern for test objects
  - 📁 **Files**: `internal/testutil/builders/`

- [x] Implement comprehensive mocking strategy
  - 📌 **Why**: Limited mocks make testing difficult
  - 🧠 **How**: Use gomock or testify/mock
  - 📁 **Files**: `*_test.go` files, create `mocks/` directories

- [x] Add table-driven tests
  - 📌 **Why**: Many tests could be more concise
  - 🧠 **How**: Convert to `[]struct{...}` test cases
  - 📁 **Files**: All test files with repetitive tests

#### Integration Testing
- [x] Create Docker-based integration test suite
  - 📌 **Why**: Tests require real tokens/services
  - 🧠 **How**: Use testcontainers-go
  - 📁 **Files**: `test/integration/`

- [x] Add E2E test scenarios
  - 📌 **Why**: No full workflow testing
  - 🧠 **How**: Script common user workflows
  - 📁 **Files**: `test/e2e/`

### 7. 🛠 Tooling & Automation

#### Build & CI
- [x] Add pre-commit hooks
  - 📌 **Why**: Catch issues before commit
  - 🧠 **How**: Use pre-commit framework
  - 📁 **Files**: `.pre-commit-config.yaml`

- [x] Enable security scanning (gosec)
  - 📌 **Why**: Security linter is disabled
  - 🧠 **How**: Enable in `.golangci.yml`
  - 📁 **Files**: Fix security issues found

- [x] Implement automated release process
  - 📌 **Why**: Manual release process
  - 🧠 **How**: Configure goreleaser
  - 📁 **Files**: `.goreleaser.yml`

#### Development Tools
- [x] Add development container configuration
  - 📌 **Why**: Consistent dev environment
  - 🧠 **How**: Create `.devcontainer/`
  - 📁 **Files**: `.devcontainer/devcontainer.json`

- [x] Create debugging configurations
  - 📌 **Why**: No standard debug setup
  - 🧠 **How**: Add VS Code/GoLand configs
  - 📁 **Files**: `.vscode/launch.json`

### 8. 📚 Documentation

#### Code Documentation
- [x] Add package-level documentation
  - 📌 **Why**: Missing package descriptions
  - 🧠 **How**: Create `doc.go` with package docs
  - 📁 **Files**: One per package

- [ ] Document all exported types and functions
  - 📌 **Why**: Limited godoc coverage
  - 🧠 **How**: Add comments per godoc standards
  - 📁 **Files**: All exported symbols

#### User Documentation
- [ ] Create comprehensive examples
  - 📌 **Why**: Limited usage examples
  - 🧠 **How**: Add `_example_test.go` files
  - 📁 **Files**: One per major package

- [ ] Add architecture documentation
  - 📌 **Why**: No high-level design docs
  - 🧠 **How**: Create `docs/architecture.md`
  - 📁 **Files**: `docs/` directory

## 🗓 Refactoring Execution Plan

### Phase 1: Foundation (Week 1-2)
1. Fix package naming conventions
2. Enable additional linters
3. Create core interfaces
4. Set up test infrastructure

### Phase 2: Structure (Week 3-4)
1. Implement dependency injection
2. Add context propagation
3. Create service abstractions
4. Unify configuration management

### Phase 3: Quality (Week 5-6)
1. Add comprehensive tests
2. Implement mocking strategy
3. Set up integration tests
4. Add pre-commit hooks

### Phase 4: Polish (Week 7-8)
1. Complete documentation
2. Add remaining tooling
3. Performance optimization
4. Security hardening

## 🧪 Testing Scope

### Unit Test Coverage Goals
- **Target**: 80% coverage minimum
- **Priority Packages**:
  - `pkg/github/` - Critical business logic
  - `pkg/config/` - Configuration management
  - `internal/git/` - Git operations
  - All command implementations

### Integration Test Requirements
- Git operations with test repositories
- Configuration loading scenarios
- Multi-provider workflows
- Network failure scenarios
- Concurrent operation stress tests

### Manual Testing Checklist
- [ ] Cross-platform compatibility (Linux/macOS/Windows)
- [ ] Large repository handling (1000+ repos)
- [ ] Network interruption recovery
- [ ] Token expiration handling
- [ ] Configuration migration from old format

## 📊 Success Metrics

- Zero golangci-lint warnings
- 80%+ test coverage
- All packages have interfaces
- No direct environment variable access
- All long operations support cancellation
- Documentation for all exported APIs
- Automated release pipeline
- Integration test suite passing

---

**Note**: Execute items in order within each phase. Mark items complete as you progress. Create separate PRs for each major change to facilitate code review.