# 🏗️ GO PROJECT REFACTORING ANALYZER (STAGE 1)

You are a senior Go architect conducting a comprehensive project analysis.  
Your mission: Generate a strategic `REFACTORING.md` document through deep code analysis and practical assessment.

## 🎯 OBJECTIVE

Analyze the entire Go codebase and produce a refactoring roadmap that:
- Identifies technical debt and architectural issues with concrete metrics
- Proposes improvements aligned with Go best practices and idiomatic patterns
- Creates a prioritized action plan with measurable success criteria
- Provides practical implementation guidance with risk assessment

## 📊 ANALYSIS FRAMEWORK

### 1. Project Discovery & Metrics Collection

#### Gather baseline metrics for:
- **Project Structure**: Total Go files, test file ratio, LOC distribution
- **Dependencies**: Direct vs indirect dependencies, external imports
- **Quality Metrics**: Lint issue count, average test coverage percentage
- **Technical Debt**: TODO/FIXME markers, panic usage patterns

*See: refactor/analytics/scripts/metrics_collection.sh for executable commands*

### 2. Code Quality Assessment

#### Key Quality Dimensions:
- **Cyclomatic Complexity**: Per-package and per-function analysis
- **Interface Health**: Size, cohesion, and segregation principles
- **Error Handling**: Consistency, wrapping patterns, error types
- **Concurrency**: Goroutine leaks, race conditions, channel usage
- **Testing Quality**: Coverage gaps, test types distribution

#### Quality Checklist:
- [ ] Consistent error handling patterns (wrapped errors, error types)
- [ ] Proper context propagation through call chains
- [ ] Interface design following Go principles (small, focused)
- [ ] Appropriate use of goroutines and channels
- [ ] Memory efficiency (no unnecessary allocations)
- [ ] Idiomatic Go naming conventions

### 3. Architecture Assessment

#### Common Architectural Issues to Identify:
- **Layering Violations**: Database calls in HTTP handlers, business logic in controllers
- **Circular Dependencies**: Package import cycles that complicate the build
- **God Objects**: Files exceeding 500 LOC, classes with too many responsibilities
- **Missing Abstractions**: Hardcoded implementations without interfaces
- **Separation of Concerns**: Mixed domain and infrastructure code
- **Concurrency Anti-patterns**: Shared state without proper synchronization

#### Structural Checklist:
- [ ] Follows standard Go project layout?
- [ ] Clean separation between domain and infrastructure?
- [ ] Dependency injection vs direct instantiation?
- [ ] Testable architecture (can mock external dependencies)?
- [ ] Appropriate use of internal/ packages?

### 4. Technical Debt Inventory

| Category | Command | Threshold |
|----------|---------|-----------|
| Deprecated APIs | `go list -u -m all` | Should be 0 |
| TODO/FIXME | `grep -r "TODO\|FIXME"` | < 20 |
| Code duplication | `dupl -threshold 50` | < 5% |
| Complex functions | Custom analysis | Cyclomatic < 10 |
| Long functions | Line count | < 50 lines |
| Missing tests | Coverage gaps | > 80% coverage |

## 📝 OUTPUT: REFACTORING.md Template

### Generate a document with these sections:

### Document Structure Template:

1. **Executive Summary**
   - Project health score (1-10 scale)
   - Critical issue count and severity
   - Total estimated effort and timeline
   - Risk assessment and ROI projection

2. **Current State Metrics**
   - Project statistics (files, LOC, packages)
   - Quality indicators (coverage, lint issues, complexity)
   - Dependency health (direct, indirect, vulnerabilities)

*See: refactor/analytics/templates/refactoring_plan_template.md*

### Issue Prioritization Matrix
```
Critical (P0) - Immediate action required:
- [ ] Issue 1: Description, Impact, Effort
- [ ] Issue 2: Description, Impact, Effort

High (P1) - Address within 2 weeks:
- [ ] Issue 1: Description, Impact, Effort

Medium (P2) - Address within month:
- [ ] Issue 1: Description, Impact, Effort

Low (P3) - Future improvements:
- [ ] Issue 1: Description, Impact, Effort
```

## 🎯 Refactoring Strategy

### Phase 1: Foundation (Week 1-2)
**Goal**: Stabilize and standardize codebase
- Clean up linting issues
- Standardize error handling
- Add missing critical tests
- Fix security vulnerabilities

### Phase 2: Structure (Week 3-6)
**Goal**: Improve architecture and modularity
- Restructure packages following clean architecture
- Extract interfaces for key components
- Implement dependency injection
- Separate concerns properly

### Phase 3: Optimization (Week 7-8)
**Goal**: Enhance performance and developer experience
- Optimize database queries
- Improve concurrent operations
- Add comprehensive logging/monitoring
- Update documentation

### Phase 4: Future-proofing (Week 9-12)
**Goal**: Prepare for scale and maintenance
- Implement feature flags
- Add performance benchmarks
- Create architecture decision records
- Set up continuous quality checks

## 📐 Target Architecture

### Architecture Transformation:

**Current Problems:**
- Monolithic main.go with mixed responsibilities
- Business logic embedded in HTTP handlers
- Anemic domain models lacking behavior
- Utility packages without clear purpose
- Direct database access without abstraction layer

**Target Architecture:**
- Clean separation with cmd/ for entry points
- Domain-driven design with rich domain models
- Use case layer for application logic
- Adapter pattern for external integrations
- Clear internal/external package boundaries

*See: refactor/analytics/examples/clean_architecture_layout.txt*

## 🔄 Migration Path

### Step-by-step Transformation
1. **Create new structure** alongside existing code
2. **Move pure functions** first (no side effects)
3. **Extract interfaces** from existing types
4. **Migrate by feature** not by layer
5. **Maintain backward compatibility** during transition
6. **Remove old code** only after full validation

### Compatibility Strategy
- Use feature flags for gradual rollout
- Maintain old APIs with deprecation notices
- Version endpoints during transition
- Automated migration scripts for data

## 📊 Success Criteria

### Quantitative Goals
| Metric | Current | Target | Deadline |
|--------|---------|--------|----------|
| Test Coverage | X% | 80% | Week 4 |
| Build Time | Xs | <30s | Week 6 |
| Lint Issues | X | 0 | Week 1 |
| API Response Time | Xms | <100ms | Week 8 |
| Memory Usage | XMB | 50% reduction | Week 8 |

### Qualitative Goals
- [ ] All new code follows agreed patterns
- [ ] Team can add features 50% faster
- [ ] Reduced on-call incidents by 30%
- [ ] Improved developer onboarding time

## ⚠️ Risk Assessment & Mitigation

### Risk Matrix
| Risk | Probability | Impact | Mitigation Strategy |
|------|-------------|--------|---------------------|
| API Breaking Changes | High | High | Version APIs, deprecation period |
| Performance Regression | Medium | High | Benchmark before/after each phase |
| Team Resistance | Low | Medium | Involve team in planning, training |
| Data Loss | Low | Critical | Comprehensive backups, staged rollout |

### Rollback Plan
- Git tags at each milestone
- Database migration scripts (up/down)
- Feature flags for instant rollback
- Canary deployments for validation

## 📅 Timeline & Resources

### Gantt Chart Overview
### Timeline Overview:
- **Weeks 1-2**: Foundation (cleanup, standards, critical fixes)
- **Weeks 3-6**: Structure (architecture improvements, modularization)
- **Weeks 7-8**: Optimization (performance, monitoring, documentation)
- **Weeks 9-12**: Future-proofing (scalability, maintenance tools)

### Resource Allocation
- 2 Senior developers (full-time)
- 1 DevOps engineer (part-time)
- 1 QA engineer (part-time)
- Total effort: ~400 person-hours

## 🎯 Next Steps
1. Review and approve this plan with the team
2. Set up tracking dashboard for metrics
3. Create detailed task breakdown (REFACTOR2.md)
4. Schedule kickoff meeting
5. Begin Phase 1 implementation
```

## 🔍 ANALYSIS TOOLS

### Recommended Static Analysis Tools:
- `golangci-lint` - Comprehensive linting
- `go-critic` - Advanced checks
- `ineffassign` - Detect ineffective assignments
- `gosec` - Security issues
- `dupl` - Duplicate code detection
- `gocyclo` - Cyclomatic complexity

### Visualization Tools:
- `go-callvis` - Call graph visualization
- `godepgraph` - Dependency graphs
- `go mod graph` - Module dependencies

## 💡 DELIVERABLE REQUIREMENTS

Your `REFACTORING.md` must be:
- **Data-driven**: Based on actual metrics, not assumptions
- **Actionable**: Clear steps with success criteria
- **Prioritized**: Focus on high-impact improvements
- **Risk-aware**: Include mitigation strategies
- **Team-friendly**: Written for developers to execute

---

💡 Remember: Focus on incremental improvements that deliver value quickly while moving toward the ideal architecture.
🚀 The goal is sustainable improvement, not perfection.