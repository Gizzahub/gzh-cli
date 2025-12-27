# Common Tasks - gzh-cli

## Adding a New Command

### Decision: External Library or gzh-cli?

1. **Check if feature belongs in external library**

   - Reusable logic? → External library
   - CLI-specific integration? → gzh-cli

1. **Create command directory**

   ```bash
   mkdir -p cmd/{command}/
   ```

1. **Add module rules**

   ```bash
   touch cmd/{command}/AGENTS.md
   ```

1. **Implement using Cobra**

   ```go
   // cmd/{command}/root.go
   var rootCmd = &cobra.Command{
       Use:   "command",
       Short: "Description",
       RunE:  runCommand,
   }
   ```

1. **Register in root command**

   ```go
   // cmd/root.go
   rootCmd.AddCommand(command.NewCommand())
   ```

1. **Add tests**

   ```bash
   touch cmd/{command}/*_test.go
   ```

1. **Update documentation**

   - Add to `docs/30-features/`

## Modifying Integration Library Command

### When to modify wrapper vs library

1. **Check wrapper file**

   - `cmd/*_wrapper.go` or `cmd/{module}/*_wrapper.go`

1. **Core logic changes**

   - Modify in external library repository
   - Example: gzh-cli-git, gzh-cli-quality

1. **Integration changes**

   - Modify wrapper if needed
   - CLI flags, output formatting

1. **Local testing**

   - Use `replace` directive in go.mod

   ```go
   replace github.com/gizzahub/gzh-cli-git => ../gzh-cli-git
   ```

## Adding Tests

```bash
# Create test file
touch cmd/{module}/{feature}_test.go

# Run tests
go test ./cmd/{module} -v

# Check coverage
go test ./cmd/{module} -cover

# Specific test
go test ./cmd/{module} -run "TestName" -v

# Race detection
go test ./cmd/{module} -race
```

## Adding New Git Platform (e.g., Bitbucket)

1. **Create API package**

   ```bash
   mkdir -p pkg/bitbucket/
   ```

1. **Implement platform interface**

   ```go
   type Client interface {
       ListRepos(ctx context.Context, org string) ([]Repo, error)
       // ... other methods
   }
   ```

1. **Register in provider registry**

   ```go
   // internal/git/provider/registry.go
   registry.Register("bitbucket", bitbucket.NewProvider())
   ```

1. **Add tests**

   - Unit tests with mocked API
   - Integration tests if possible

## Handling Secrets in Tests

```go
func TestGitHubAPI(t *testing.T) {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        t.Skip("GITHUB_TOKEN not set")
    }
    // Test with real API
}
```

## Working with Module-Specific Guides

### Before modifying any module

1. Read `cmd/AGENTS_COMMON.md` - project-wide conventions
1. Read `cmd/{module}/AGENTS.md` - module-specific rules
1. Check existing patterns in the module
1. Follow the established code style

### Module guides available

- Common Guidelines: `cmd/AGENTS_COMMON.md`
- Git module: `cmd/git/AGENTS.md`
- IDE module: `cmd/ide/AGENTS.md`
- Quality module: `cmd/quality/AGENTS.md`
- 15 modules total - see `cmd/*/AGENTS.md`

## Code Style & Conventions

### Binary Naming

- **Correct**: `gz` (never `gzh-cli`)
- Commands use `gz` prefix

### Interface-Driven Design

- Heavy use of Go interfaces
- Testability through dependency injection
- Direct constructors (no DI containers)

### Error Handling

- Structured errors with context
- Use wrapped errors: `fmt.Errorf("context: %w", err)`

### Korean Comments

- New code should use Korean comments
- Maintain consistency with existing code
