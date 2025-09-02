# CLI Specification Template

This template provides a standardized format for writing CLI command specifications following SDD (Specification-Driven Development) principles.

## Basic Template

```markdown
# Command: gz [subcommand] [options]

## Scenario: [Brief description of what this specification covers]

### Input

**Command**:
```bash
gz command --flag value argument
```

**Prerequisites**:
- [ ] Required environment variables set
- [ ] Required files/directories exist
- [ ] Network connectivity (if needed)
- [ ] Authentication tokens (if needed)

### Expected Output

**Success Case**:
```
stdout: [Expected output patterns]
stderr: [Error messages, if any]
Exit Code: 0
```

**Error Cases**:
```
stdout: [Error output patterns]
stderr: [Specific error messages]
Exit Code: [Non-zero error code]
```

### Side Effects

**Files Created**:
- `path/to/file`: Description of file contents
- `directory/`: Description of directory structure

**Files Modified**:
- `existing/file`: What changes are made

**State Changes**:
- Configuration updates
- Cache modifications
- Network operations performed

### Validation

**Automated Tests**:
```bash
# Test commands to verify behavior
result=$(gz command --flag value argument)
exit_code=$?

# Assertions
assert_contains "$result" "expected string"
assert_exit_code 0
assert_directory_exists "./expected/path"
assert_file_contains "./file" "expected content"
```

**Manual Verification**:
1. Step-by-step verification process
2. Visual confirmation points
3. Expected vs actual comparison

### Edge Cases

**Boundary Conditions**:
- Empty inputs
- Maximum values
- Special characters
- Unicode handling

**Error Conditions**:
- Network failures
- Authentication errors
- Permission issues
- Resource constraints

### Performance Expectations

**Response Time**:
- Normal case: < X seconds
- Large datasets: < Y seconds
- Network operations: < Z seconds

**Resource Usage**:
- Memory: < N MB
- CPU: < M% for P seconds
- Disk I/O: < Q operations

## Notes

Additional context, implementation details, or future considerations.
```

## Quick Reference Template

For simple commands, use this condensed format:

```markdown
# gz [command]

**Input**: `gz command [args]`
**Output**: Expected output pattern
**Exit Code**: 0 (success) / 1 (error)
**Creates**: Files/directories created
**Test**: `assert_contains "$(gz command)" "expected"`
```

## Example Specifications

### Success Scenario

```markdown
# Command: gz version

## Scenario: Display version information

### Input

**Command**:
```bash
gz version
```

**Prerequisites**:
- [ ] gzh-cli binary installed

### Expected Output

**Success Case**:
```
stdout: gzh-cli version X.Y.Z (commit: abcd1234, built: 2025-01-01T00:00:00Z)
stderr: (empty)
Exit Code: 0
```

### Side Effects

**Files Created**: None
**Files Modified**: None
**State Changes**: None

### Validation

**Automated Tests**:
```bash
result=$(gz version)
exit_code=$?

assert_contains "$result" "gzh-cli version"
assert_matches "$result" "version [0-9]+\.[0-9]+\.[0-9]+"
assert_exit_code 0
```
```

### Error Scenario

```markdown
# Command: gz synclone github -o nonexistent-org

## Scenario: Handle organization not found error

### Input

**Command**:
```bash
gz synclone github -o nonexistent-org-12345
```

**Prerequisites**:
- [ ] Valid GitHub token (optional, for better rate limits)

### Expected Output

**Error Case**:
```
stdout: 🔍 Fetching repository list from GitHub organization: nonexistent-org-12345
stderr: (empty)
Exit Code: 1
```

**Final Output**:
```
Error: failed to fetch repository list: failed to get repositories: 404 Not Found
```

### Side Effects

**Files Created**: None
**Files Modified**: None  
**State Changes**: None (no partial state)

### Validation

**Automated Tests**:
```bash
result=$(gz synclone github -o nonexistent-org-12345 2>&1)
exit_code=$?

assert_contains "$result" "failed to fetch repository list"
assert_contains "$result" "404 Not Found"
assert_exit_code 1
assert_not_directory_exists "./nonexistent-org-12345"
```
```

## Specification Checklist

Use this checklist to ensure your specification is complete:

### Input Coverage
- [ ] All required arguments specified
- [ ] All optional flags documented
- [ ] Flag aliases covered
- [ ] Environment variable dependencies listed
- [ ] Prerequisites clearly defined

### Output Coverage
- [ ] Success output patterns defined
- [ ] Error output patterns defined
- [ ] Exit codes specified
- [ ] Progress indicators described
- [ ] Quiet mode behavior covered

### Side Effects Coverage
- [ ] All file operations documented
- [ ] Directory structure changes listed
- [ ] Network operations specified
- [ ] Configuration changes noted
- [ ] Cache operations described

### Edge Cases Coverage
- [ ] Empty inputs handled
- [ ] Invalid inputs covered
- [ ] Network failures addressed
- [ ] Permission errors specified
- [ ] Resource exhaustion handled

### Testing Coverage
- [ ] Success case tests provided
- [ ] Error case tests provided
- [ ] Boundary condition tests included
- [ ] Performance tests specified
- [ ] Manual verification steps listed

### Documentation Quality
- [ ] Clear, unambiguous language used
- [ ] Examples are realistic and testable
- [ ] Prerequisites are complete
- [ ] Validation commands are correct
- [ ] Related specifications linked

## Common Patterns

### Authentication Patterns

```markdown
**Prerequisites**:
- [ ] GITHUB_TOKEN environment variable set, OR
- [ ] --token flag provided, OR  
- [ ] Interactive authentication completed

**Error Cases**:
```
# No authentication
stdout: ⚠️ Warning: No GitHub token provided. API rate limits may apply.

# Invalid authentication  
stdout: 🚫 GitHub authentication failed
stderr: (empty)
Exit Code: 1
```
```

### Progress Patterns

```markdown
**Success Output**:
```
🔍 Fetching repository list from GitHub organization: myorg
📋 Found 25 repositories in organization myorg
🚀 Using optimized streaming API for large-scale operations
⚙️ Starting repository synchronization with strategy: reset
[████████████████████████████████] 25/25 repositories processed
✅ Synchronization completed successfully
```
```

### Rate Limit Patterns

```markdown
**Rate Limit Error**:
```
🚫 GitHub API Rate Limit Exceeded!
   Rate Limit: 60 requests/hour
   Remaining: 0
   Reset Time: Tue, 02 Sep 2025 12:45:43 KST
   Wait Time: 14 minutes 22 seconds

💡 Solution: Set GITHUB_TOKEN environment variable to bypass rate limits
   export GITHUB_TOKEN="your_github_personal_access_token"

Error: [rate_limit] GitHub API rate limit exceeded: GitHub API rate limit exceeded
```

**Important**: No Usage block should be displayed for rate limit errors.
```

## Integration with Testing Framework

### Automated Test Generation

Specifications can be converted to automated tests:

```bash
# Generate tests from specifications
make generate-spec-tests

# Run specification-based tests
make test-specs

# Validate all specifications
make validate-specs
```

### CI/CD Integration

```yaml
# Specification validation in CI
- name: Validate CLI Specifications
  run: |
    make build
    make test-specs
    make validate-specs
```

---

**Usage**: Copy this template for each new command specification. Customize sections based on command complexity and requirements.

**Related**: See [CLI Specification Strategy](../60-development/68-cli-specification-strategy.md) for detailed methodology.