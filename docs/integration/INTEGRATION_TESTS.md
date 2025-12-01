# Integration Tests Restoration Plan

## Status: ⚠️ Deferred

Integration tests for Git repo commands are currently disabled pending provider interface refactoring.

## Background

During Phase 3 (Git integration), wrapper files were created to delegate functionality to gzh-cli-git library. However, existing integration tests depend on the old implementation and provider interfaces.

## Disabled Tests

Location: `cmd/git/repo/repo_integration_test.go`

**Total**: 9 integration tests disabled with TODO comments

```go
// TODO: 통합 테스트는 provider 인터페이스 리팩토링 후 재구현 필요
```

### Test Coverage Needed

1. **clone-or-update wrapper** (`repo_clone_or_update_wrapper.go`)

   - Test all 6 update strategies (rebase, reset, clone, skip, pull, fetch)
   - Test branch specification
   - Test depth configuration
   - Test error handling

1. **bulk-update wrapper** (`repo_bulk_update_wrapper.go`)

   - Test recursive repository scanning
   - Test parallel processing
   - Test pattern filtering (include/exclude)
   - Test dry-run mode
   - Test various output formats (table, JSON)

## Dependencies

### Prerequisite Work

1. **Provider Interface Refactoring**

   - Current: `pkg/git/provider` interface
   - Issue: Tightly coupled with old implementation
   - Needed: Clean interface compatible with wrapper pattern

1. **Test Infrastructure**

   - Mock gzh-cli-git library responses
   - Test repositories setup/teardown
   - Parallel test execution support

## Proposed Approach

### Option A: Mock Library Calls

```go
// Use gomock to mock gzh-cli-git client
mockClient := repository_mocks.NewMockClient(ctrl)
mockClient.EXPECT().
    CloneOrUpdate(gomock.Any(), gomock.Any()).
    Return(&repository.CloneOrUpdateResult{...}, nil)
```

**Pros**:

- Fast execution
- No external dependencies
- Full control over test scenarios

**Cons**:

- Doesn't test actual library integration
- Requires mock maintenance

### Option B: Integration with Real Library

```go
// Use actual gzh-cli-git library with test repositories
client := repository.NewClient()
result, err := client.CloneOrUpdate(ctx, opts)
```

**Pros**:

- Tests real integration
- Catches library API changes
- Validates end-to-end functionality

**Cons**:

- Slower execution
- Requires test repository setup
- More complex teardown

### Option C: Hybrid Approach (Recommended)

```go
// Unit tests: Mock library
// Integration tests: Real library with testcontainers
```

**Pros**:

- Best of both worlds
- Fast feedback + comprehensive coverage
- Clear test separation

**Cons**:

- More initial setup work

## Implementation Plan

### Phase 1: Provider Interface Cleanup (1-2 hours)

1. Review current `pkg/git/provider` interface
1. Identify wrapper-specific requirements
1. Create adapter layer if needed
1. Document interface contracts

### Phase 2: Mock Setup (2-3 hours)

1. Generate mocks for gzh-cli-git interfaces
   ```bash
   make generate-mocks
   ```
1. Create test helpers for common scenarios
1. Write unit tests for wrappers

### Phase 3: Integration Tests (3-4 hours)

1. Set up test repository infrastructure
1. Implement testcontainers for Git server
1. Write integration tests for each wrapper
1. Verify all 6 update strategies work

### Phase 4: CI Integration (1 hour)

1. Add integration tests to CI pipeline
1. Configure parallel test execution
1. Set up test reporting

**Total Estimated Time**: 7-10 hours

## Test Structure

```
test/
├── unit/
│   └── cmd/
│       └── git/
│           └── repo/
│               ├── clone_or_update_wrapper_test.go  # Mock-based
│               └── bulk_update_wrapper_test.go      # Mock-based
└── integration/
    └── git/
        └── repo/
            ├── clone_or_update_integration_test.go  # Real library
            └── bulk_update_integration_test.go      # Real library
```

## Success Criteria

- [ ] All 9 disabled integration tests re-enabled
- [ ] 100% coverage of wrapper functions
- [ ] All 6 update strategies tested
- [ ] Bulk update parallel processing tested
- [ ] Pattern filtering tested
- [ ] Error scenarios covered
- [ ] CI pipeline includes integration tests
- [ ] Test execution time < 5 minutes (unit + integration)

## Priority

**Medium Priority (P2)**

- Not blocking current functionality
- Wrappers work correctly (manually verified)
- Can be addressed in next sprint

## Risks

1. **API Changes**: gzh-cli-git may change APIs, breaking tests

   - Mitigation: Version pin library in go.mod

1. **Test Complexity**: Integration tests may be flaky

   - Mitigation: Use testcontainers for isolation

1. **Resource Requirements**: Parallel tests need more resources

   - Mitigation: Configure test parallelism based on CI resources

## Related Work

- [ ] Provider interface refactoring
- [ ] gzh-cli-git API stabilization
- [ ] Wrapper unit test coverage
- [ ] CI/CD pipeline improvements

## References

- Current disabled tests: `cmd/git/repo/repo_integration_test.go`
- Wrapper implementations:
  - `cmd/git/repo/repo_clone_or_update_wrapper.go`
  - `cmd/git/repo/repo_bulk_update_wrapper.go`
- Library documentation: gzh-cli-git/README.md

______________________________________________________________________

**Created**: 2025-12-01
**Priority**: P2 (Medium)
**Estimated Effort**: 7-10 hours
**Status**: Awaiting provider interface refactoring
**Assignee**: TBD
**Model**: claude-sonnet-4-5-20250929
