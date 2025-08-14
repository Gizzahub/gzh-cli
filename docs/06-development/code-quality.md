# Code Quality Standards

This document describes the code quality standards and linting configuration for the gzh-cli project.

## Overview

We use `golangci-lint` as our primary code quality tool, configured with a comprehensive set of linters to ensure:

- Code correctness and bug prevention
- Security best practices
- Performance optimization
- Consistent code style
- Maintainability

## Quick Start

```bash
# Run all linters
make lint

# Format code
make fmt

# Run full verification (format, lint, test, coverage)
make verify

# Check for specific issues
make todo    # Show all TODO comments
make fixme   # Show all FIXME comments
make vuln    # Check for vulnerabilities
```

## Linter Categories

### Core Linters

- **errcheck**: Ensures all errors are handled
- **govet**: Reports suspicious constructs
- **staticcheck**: Advanced static analysis
- **unused**: Finds unused code
- **ineffassign**: Detects ineffectual assignments

### Code Quality

- **revive**: Extensible linter with many rules
- **gocritic**: Opinionated linter with many checks
- **goconst**: Finds repeated strings that could be constants
- **unconvert**: Removes unnecessary type conversions
- **unparam**: Reports unused function parameters
- **misspell**: Fixes common misspellings
- **prealloc**: Suggests slice pre-allocation
- **nakedret**: Finds naked returns in long functions

### Security

- **gosec**: Security-focused linter
- **G104**: Unchecked errors (handled by errcheck)
- **G204**: Command execution (allowed for CLI tools)
- **G304**: File path from user input (allowed for CLI tools)

### Complexity

- **gocyclo**: Cyclomatic complexity (max: 15)
- **funlen**: Function length (max: 100 lines, 50 statements)
- **gocognit**: Cognitive complexity (max: 20)
- **dupl**: Code duplication (threshold: 100 tokens)
- **nestif**: Nested if statements (max: 4)
- **maintidx**: Maintainability index (min: 20)

### Style

- **lll**: Line length limit (180 characters)
- **whitespace**: Whitespace issues
- **wsl**: Whitespace linter with strict rules
- **gofumpt**: Stricter gofmt
- **goimports**: Import formatting and grouping
- **godot**: Check comment punctuation
- **godox**: TODO/FIXME comment tracker
- **goheader**: License header checker

### Performance

- **bodyclose**: HTTP response body closure
- **rowserrcheck**: SQL rows.Err() checking
- **sqlclosecheck**: SQL rows/stmt closure

### Bug Prevention

- **nilerr**: Finds code returning nil instead of error
- **nilnil**: Finds functions returning (nil, nil)
- **noctx**: HTTP requests without context
- **copyloopvar**: Loop variable copying issues
- **musttag**: Struct tag validation for marshaling
- **contextcheck**: Context propagation
- **errorlint**: Error wrapping best practices

### Dependencies

- **gomodguard**: Blocks specific modules
- **depguard**: Dependency constraints

### Testing

- **testpackage**: Separate test packages
- **tparallel**: Parallel test detection
- **thelper**: Test helper function checks
- **testableexamples**: Example test validation

## Configuration Details

### Line Length

- Maximum: 180 characters
- Rationale: Modern monitors can display longer lines comfortably

### Complexity Thresholds

- Cyclomatic complexity: 15
- Cognitive complexity: 20
- Function length: 100 lines, 50 statements
- Maintainability index: 20

### Variable Naming

- Minimum length: 2 characters
- Allowed short names: err, ok, id, i, j, k, v, tt, tc, t, mu, fn, op

### Excluded Patterns

- Test files have relaxed rules for:
  - Function length
  - Complexity
  - Variable naming
  - Whitespace
- Generated files are excluded entirely
- Integration tests are excluded (outdated API usage)

## Custom Rules

### Error Handling

- All errors must be checked except in specific cases:
  - Print functions
  - File.Close() in deferred calls
  - Flush/Sync operations
  - Remove operations in cleanup

### Context Usage

- Production code must propagate context
- context.Background() only allowed in:
  - main.go
  - Test files

### Dependencies

- Prohibited packages:
  - `github.com/pkg/errors` (use stdlib errors)
  - `io/ioutil` (deprecated)

### TODO Comments

- Tracked by godox linter
- Must be addressed or converted to issues
- Keywords: TODO, FIXME, BUG, HACK, OPTIMIZE, XXX

## Pre-commit Integration

The linting configuration is integrated with pre-commit hooks:

```bash
# Install pre-commit hooks
make pre-commit-install

# Run manually
make pre-commit
```

## Continuous Integration

All pull requests are automatically checked for:

1. Formatting issues
2. Linting violations
3. Test coverage
4. Security vulnerabilities

## Suppressing False Positives

When you need to suppress a false positive:

```go
// For a specific line
//nolint:errcheck // This error is intentionally ignored because...

// For a function
//nolint:funlen,gocognit // This function is necessarily complex because...
func complexFunction() {
    // ...
}
```

Always provide a reason when suppressing linters.

## Adding New Linters

When adding a new linter:

1. Add it to `.golangci.yml`
2. Configure appropriate settings
3. Add exclusions for false positives
4. Update this documentation
5. Run on the entire codebase and fix issues

## Resources

- [golangci-lint documentation](https://golangci-lint.run/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
