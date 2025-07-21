# 🔧 GO PROJECT REFACTORING CHECKLIST GENERATOR (STAGE 2)

You are a Go refactoring specialist creating an executable checklist from the strategic plan.  
Your mission: Transform `REFACTORING.md` into a detailed, actionable `REFACTORING_CHECKLIST.md`.

## 🎯 OBJECTIVE

Convert high-level refactoring goals into:
- Atomic, independently executable tasks (2-4 hour chunks)
- Dependency-ordered execution plan with clear prerequisites
- Risk-assessed implementation steps with rollback procedures
- Automated validation criteria with specific commands

## 📋 CHECKLIST GENERATION FRAMEWORK

### 1. Task Decomposition Rules

#### Task Requirements:
- **Atomicity**: Each task completable in 2-4 hours, no broken intermediate states
- **Measurability**: Clear metrics, specific file changes, pass/fail criteria
- **Reversibility**: Git commits per task, rollback procedures, data migration safety

*See: refactor/analytics/templates/task_decomposition_guide.yaml*

### 2. Task Categorization & Risk Assessment

#### Task Categories:
- **🧹 Cleanup**: Low risk, immediate value (formatting, linting, dead code)
- **🏗️ Structure**: Medium risk, foundational changes (refactoring, modularization)
- **🔄 Behavior**: High risk, functional changes (logic modifications, API changes)
- **🚀 Enhancement**: Optional improvements (performance, monitoring, documentation)

#### Risk Assessment Framework:
- **Risk Levels**: none, low, medium, high
- **Risk Factors**: Breaking changes, data migration, external dependencies
- **Mitigation**: Feature flags, staged rollout, comprehensive testing

### 3. Dependency Management

#### Dependency Graph Pattern:
1. **Cleanup tasks** must complete before structural changes
2. **Structural refactoring** enables behavior modifications
3. **Behavior changes** should stabilize before enhancements
4. **Enhancements** are optional and can be deferred

*Visual dependency graphs should be generated based on actual tasks*

## 📝 OUTPUT: REFACTORING_CHECKLIST.md Template

### Checklist Structure Template:

1. **Pre-flight Checklist**
   - Test status verification
   - Backup procedures
   - Team communication
   - Monitoring setup
   - Rollback readiness

2. **Task Template Format**
   - Task ID and description
   - Time estimate and risk level
   - Dependencies and affected files
   - Execution steps (reference scripts)
   - Validation criteria
   - Commit message template
   - Rollback procedure

*See: refactor/analytics/templates/refactoring_checklist_template.md*

---

### CL002: Fix Auto-fixable Lint Issues
- **Time**: 1 hour
- **Risk**: Low
- **Dependencies**: [CL001]
- **Files**: Various (see lint report)

#### Pre-check:
- Generate baseline lint report
- Create refactoring branch

#### Execution:
- Run auto-fix for fixable issues
- Identify manual fixes needed
- Apply fixes incrementally

#### Validation:
- Compare before/after metrics
- Ensure all tests pass
- Verify no regressions

*See: refactor/analytics/scripts/lint_cleanup_workflow.sh*

---

## 🏗️ Phase 2: Structural Refactoring (Medium Risk)

### ST001: Extract Business Logic from main.go
- **Time**: 2 hours
- **Risk**: Medium
- **Dependencies**: [CL001, CL002]
- **Breaking Changes**: None (internal restructure)

#### Structural Refactoring Pattern:

**Problem**: Monolithic main.go with mixed concerns
- Database setup mixed with routing
- Business logic in handler functions
- No separation of concerns

**Solution**: Extract into clean architecture
- Minimal main.go for bootstrapping
- Separate app initialization layer
- Handler layer with dependency injection
- Repository pattern for data access

*See: refactor/analytics/examples/main_extraction_pattern.go*

#### Migration Steps:
1. Create new directory structure
2. Move main.go to cmd/api/
3. Extract components using IDE refactoring tools
4. Update import paths systematically
5. Verify no circular dependencies

*See: refactor/analytics/scripts/structure_migration.sh*

#### Tests Required:
- Application lifecycle validation
- Graceful shutdown testing  
- Dependency injection verification
- Integration test coverage

*See: refactor/analytics/examples/app_lifecycle_test.go*

---

### ST002: Implement Repository Pattern
- **Time**: 3 hours
- **Risk**: Medium
- **Dependencies**: [ST001]
- **Breaking Changes**: Internal API changes

#### Repository Pattern Implementation:

**Interface Design Principles**:
- Context-aware for cancellation and timeouts
- Clear separation of commands and queries
- Filter objects for flexible querying
- Error handling at repository boundary

**Implementation Guidelines**:
- One repository per aggregate root
- No business logic in repositories
- Use prepared statements for SQL
- Transaction support through context

*See: refactor/analytics/examples/repository_pattern.go*

func NewUserRepository(db *sql.DB) repository.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
    query := `SELECT id, email, name, created_at FROM users WHERE id = $1`
    
    var user entity.User
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID, &user.Email, &user.Name, &user.CreatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    
    return &user, err
}
```

---

## 🔄 Phase 3: Behavioral Changes (High Risk)

### BH001: Add Context Propagation
- **Time**: 4 hours  
- **Risk**: High (API changes)
- **Dependencies**: [ST001, ST002]
- **Breaking Changes**: YES - All handler signatures change

#### Migration Strategy:
1. Add context parameter (maintain compatibility)
2. Deprecate old methods
3. Remove deprecated methods (next major version)

#### Pattern Application:
1. Add new methods with context parameter
2. Delegate old methods to new ones
3. Mark old methods as deprecated
4. Update all callers incrementally
5. Remove deprecated methods after migration

*See: refactor/analytics/examples/context_migration_example.go*

#### Validation Tests:

Comprehensive test suite for context behavior including:
- Context cancellation handling
- Timeout enforcement
- Context value propagation
- Graceful degradation for legacy code

*See: refactor/analytics/examples/context_validation_test.go*

---

## 🚀 Phase 4: Enhancements (Optional)

### EN001: Implement Caching Layer
- **Time**: 2 hours
- **Risk**: Low (additive change)
- **Dependencies**: [BH001]
- **Performance Target**: 50% reduction in DB queries

#### Implementation:

Cache-aside pattern with Redis integration:
- Automatic cache population on miss
- TTL-based expiration
- Cache invalidation on updates
- Batch operations support
- Cache warming capabilities

*See: refactor/analytics/examples/caching_layer_example.go*

---

## 📊 Progress Tracking

### Summary Dashboard

Real-time progress tracking with phase breakdown:
- Task completion metrics
- Time tracking per phase
- Quality metrics monitoring
- Overall progress visualization

*See: refactor/analytics/templates/progress_tracking_dashboard.yaml*

### Task Dependencies Graph
```
CL001 ─┬─> CL002 ─┬─> ST001 ─┬─> ST002 ─┬─> BH001 ─┬─> EN001
       │          │          │          │          │
       └─> CL003  └─> CL004  └─> ST003  └─> BH002  └─> EN002
```

## 🛡️ Validation Gates

### After Each Task:
1. **Build**: `go build ./...` must succeed
2. **Test**: `go test ./...` must pass
3. **Lint**: No new lint issues introduced
4. **Benchmark**: No performance regression > 10%
5. **Coverage**: Coverage must not decrease

### Automated Validation Script:
Automated validation ensures quality standards:
- Build verification
- Test suite execution
- Coverage comparison
- Lint issue tracking
- Benchmark regression detection

*See: refactor/analytics/scripts/validate_task_checklist.sh*

## 🎯 Completion Criteria

### Per Task:
- [ ] Code changes implemented
- [ ] Tests written and passing
- [ ] Documentation updated
- [ ] Peer review completed
- [ ] Merged to main branch

### Per Phase:
- [ ] All tasks in phase completed
- [ ] Integration tests passing
- [ ] Performance benchmarks met
- [ ] Team sign-off received

### Overall:
- [ ] All 25 tasks completed
- [ ] Final integration test suite passes
- [ ] Performance improvements verified
- [ ] Documentation fully updated
- [ ] Team trained on new patterns
```

## 💡 CHECKLIST GENERATION BEST PRACTICES

1. **Granularity**: If a task takes > 4 hours, split it
2. **Dependencies**: Explicitly state what must complete first
3. **Validation**: Automate as much validation as possible
4. **Documentation**: Update docs in the same task that changes code
5. **Rollback**: Every task must be reversible

---

🚀 Generate a checklist that enables:
- Parallel work where possible
- Safe incremental progress  
- Clear success measurement
- Rapid rollback if needed