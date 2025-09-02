# Command: gz quality

## Scenario: Run comprehensive code quality checks

### Input

**Command**:
```bash
gz quality --fix
```

**Prerequisites**:

- [ ] Project with source code files
- [ ] Quality tools installed (linters, formatters)
- [ ] Write permissions for auto-fix operations

### Expected Output

**Success Case (Issues Fixed)**:
```text
🔍 Scanning project for quality issues...

📂 Go Files (12 files)
✅ gofmt: 2 files formatted
✅ goimports: 1 import organized  
⚠️  golangci-lint: 3 issues found, 2 auto-fixed
   ❌ unused variable 'result' in main.go:45

📂 TypeScript Files (8 files)
✅ prettier: 3 files formatted
✅ eslint: 5 issues found, 5 auto-fixed

📂 Python Files (4 files)
✅ black: 1 file formatted
✅ isort: imports organized
✅ flake8: no issues found

📋 Summary
   Total files: 24
   Issues found: 9
   Auto-fixed: 8
   Manual fix required: 1

⚠️  1 issue requires manual attention. See details above.

stderr: (empty)
Exit Code: 1
```

**Success Case (Clean Code)**:
```text
🔍 Scanning project for quality issues...

📂 Go Files (5 files)
✅ gofmt: all files properly formatted
✅ goimports: all imports organized
✅ golangci-lint: no issues found

📂 Markdown Files (3 files)
✅ markdownlint: all files properly formatted

📋 Summary
   Total files: 8
   Issues found: 0
   Auto-fixed: 0

🎉 Code quality excellent! No issues found.

stderr: (empty)
Exit Code: 0
```

**Tool Missing Error**:
```text
🔍 Scanning project for quality issues...

❌ Required tools missing:
   - golangci-lint: not found in PATH
   💡 Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   - prettier: not found
   💡 Install: npm install -g prettier

🚫 Cannot proceed without required quality tools.

stderr: missing required tools
Exit Code: 2
```

### Side Effects

**Files Created**:
- `.gzh/quality-report.json` - Detailed quality report
- `.gzh/quality-cache/` - Tool cache directory

**Files Modified**:
- Source files (when using --fix flag)
- Configuration files (auto-generated if missing)

**State Changes**:
- Code formatted according to project standards
- Import statements organized
- Linting issues resolved where possible

### Validation

**Automated Tests**:
```bash
# Test quality check on clean code
cd test-project-clean
result=$(gz quality 2>&1)
exit_code=$?

assert_contains "$result" "Code quality excellent"
assert_exit_code 0

# Test quality check with issues
cd test-project-issues  
result=$(gz quality --fix 2>&1)
assert_contains "$result" "Issues found:"
assert_contains "$result" "Auto-fixed:"
# Exit code 0 (all fixed) or 1 (manual fixes needed)
```

**Manual Verification**:
1. Run on project with known quality issues
2. Verify auto-fixes are applied correctly
3. Check that manual issues are clearly reported
4. Confirm quality report is generated

### Edge Cases

**Large Codebases**:
- Progress indication for large file counts
- Parallel processing of files
- Memory management for large projects

**Mixed Languages**:
- Go, TypeScript, Python, Rust, Java support
- Language-specific tool configuration
- Cross-language consistency rules

**Configuration Files**:
- Auto-detection of existing configs (.golangci.yml, .eslintrc, etc.)
- Generation of missing configuration files
- Respect for project-specific settings

**Git Integration**:
- Check only modified files with --git-diff
- Pre-commit hook integration
- Ignore patterns from .gitignore

### Performance Expectations

**Response Time**:
- Small projects (<50 files): < 10 seconds
- Medium projects (50-500 files): < 60 seconds
- Large projects (>500 files): Progress indication

**Resource Usage**:
- Memory: < 500MB for large projects
- CPU: Parallel processing when possible
- Disk: Temporary cache for incremental checks

## Notes

- Supports multiple programming languages simultaneously
- Auto-fix capabilities for formatting and simple issues
- Incremental checking for improved performance
- Integration with popular development tools and IDEs
- Configurable quality standards per project
