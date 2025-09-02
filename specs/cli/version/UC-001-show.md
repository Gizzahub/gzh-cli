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
```text
gzh-cli version 2.1.0 (commit: a1b2c3d4, built: 2025-09-02T12:30:00Z)

stderr: (empty)
Exit Code: 0
```

**Development Build**:
```text
gzh-cli version dev (commit: local-dev, built: 2025-09-02T15:45:00Z)

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
# Test version command
result=$(gz version 2>&1)
exit_code=$?

assert_contains "$result" "gzh-cli version"
assert_matches "$result" "version [0-9]+\.[0-9]+\.[0-9]+|dev"
assert_contains "$result" "commit:"
assert_contains "$result" "built:"
assert_exit_code 0
```

**Manual Verification**:
1. Run command and verify output format
2. Check version number is meaningful
3. Verify commit hash is present
4. Confirm build timestamp is reasonable

### Edge Cases

**Missing Build Information**:
```text
gzh-cli version unknown (commit: unknown, built: unknown)
```

**Custom Build Tags**:
```text
gzh-cli version 2.1.0-beta.3 (commit: a1b2c3d4, built: 2025-09-02T12:30:00Z)
```

### Performance Expectations

**Response Time**: < 100ms (instant)
**Resource Usage**: Minimal

## Notes

- Version follows semantic versioning (semver)
- Build information embedded at compile time
- Used for debugging and support purposes
- Hidden in main help to reduce clutter
