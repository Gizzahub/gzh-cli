# AGENTS.md - repo-config (GitHub 저장소 설정 관리)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**repo-config**는 GitHub 조직의 저장소 설정을 대규모로 관리하고 컴플라이언스를 보장하는 인프라 관리 모듈입니다.

### 핵심 기능

- 대규모 GitHub 저장소 설정 관리
- 보안 정책 및 브랜치 보호 규칙 적용
- 템플릿 기반 설정 관리
- 컴플라이언스 감사 및 리포팅
- 실시간 대시보드 모니터링
- CVSS 기반 위험도 평가

## 🔐 개발 시 핵심 주의사항

### 1. GitHub API 속도 제한 관리

```go
// ✅ API 속도 제한 대응
type GitHubAPIClient struct {
    client      *github.Client
    rateLimiter *rate.Limiter
    retryPolicy *RetryPolicy
}

func (c *GitHubAPIClient) MakeAPICall(ctx context.Context, fn func() error) error {
    // 속도 제한 준수
    if err := c.rateLimiter.Wait(ctx); err != nil {
        return err
    }

    // 지수 백오프로 재시도
    return c.retryPolicy.Execute(func() error {
        if err := fn(); err != nil {
            if isRateLimitError(err) {
                time.Sleep(time.Minute) // 1분 대기
                return err
            }
            return err
        }
        return nil
    })
}
```

### 2. 대량 작업 안전성

```go
// ✅ 안전한 대량 저장소 처리
func (r *RepoManager) ApplyConfigBatch(repos []Repository, config Config) error {
    // 드라이런 모드 지원
    if r.dryRun {
        return r.validateConfigApplication(repos, config)
    }

    // 배치 크기 제한
    batchSize := 10
    for i := 0; i < len(repos); i += batchSize {
        end := i + batchSize
        if end > len(repos) {
            end = len(repos)
        }

        batch := repos[i:end]
        if err := r.processBatch(batch, config); err != nil {
            return fmt.Errorf("batch %d failed: %w", i/batchSize, err)
        }

        // 배치 간 쿨다운
        time.Sleep(2 * time.Second)
    }
}
```

### 3. 설정 백업 및 롤백

```go
// ✅ 안전한 설정 변경
func (r *RepoManager) ApplyConfigWithBackup(repo string, config Config) error {
    // 현재 설정 백업
    currentConfig, err := r.captureCurrentConfig(repo)
    if err != nil {
        return fmt.Errorf("failed to backup config: %w", err)
    }

    // 백업 저장
    backupID := r.saveBackup(repo, currentConfig)

    // 설정 적용
    if err := r.applyConfig(repo, config); err != nil {
        // 실패 시 롤백
        r.restoreFromBackup(repo, backupID)
        return fmt.Errorf("config application failed: %w", err)
    }

    return nil
}
```

## 🏗️ 템플릿 시스템

### 설정 템플릿 관리

```yaml
# ✅ 계층적 템플릿 구조
templates:
  base:
    branch_protection:
      required_status_checks:
        strict: true
      enforce_admins: true

  security:
    extends: base
    security:
      secret_scanning: enabled
      dependency_vulnerability_alerts: enabled

  enterprise:
    extends: security
    additional_settings:
      delete_branch_on_merge: true
      squash_merge_commit_title: "COMMIT_OR_PR_TITLE"
```

### 템플릿 검증

```go
// ✅ 템플릿 유효성 검사
func (t *TemplateManager) ValidateTemplate(template Template) error {
    // 순환 참조 체크
    if err := t.checkCircularDependency(template); err != nil {
        return fmt.Errorf("circular dependency detected: %w", err)
    }

    // 필수 필드 검사
    if err := t.validateRequiredFields(template); err != nil {
        return fmt.Errorf("missing required fields: %w", err)
    }

    // GitHub API 호환성 검사
    if err := t.validateGitHubCompatibility(template); err != nil {
        return fmt.Errorf("GitHub API incompatible: %w", err)
    }

    return nil
}
```

## 📊 컴플라이언스 감사

### 감사 규칙 엔진

```go
// ✅ 유연한 감사 시스템
type ComplianceRule struct {
    Name        string
    Category    string
    Severity    string // critical, high, medium, low
    CheckFunc   func(repo Repository) ComplianceResult
    FixFunc     func(repo Repository) error
}

type ComplianceEngine struct {
    rules []ComplianceRule
}

func (c *ComplianceEngine) RunAudit(repos []Repository) AuditReport {
    report := AuditReport{
        Timestamp: time.Now(),
        Results:   make(map[string][]ComplianceResult),
    }

    for _, repo := range repos {
        for _, rule := range c.rules {
            result := rule.CheckFunc(repo)
            report.Results[repo.FullName] = append(report.Results[repo.FullName], result)
        }
    }

    return report
}
```

## 🧪 테스트 요구사항

### 대규모 시나리오 테스트

```bash
# 대량 저장소 처리 테스트
go test ./cmd/repo-config -v -run TestMassRepositoryProcessing

# API 속도 제한 테스트
go test ./cmd/repo-config -v -run TestRateLimitHandling

# 컴플라이언스 감사 테스트
go test ./cmd/repo-config -v -run TestComplianceAudit

# 템플릿 시스템 테스트
go test ./cmd/repo-config -v -run TestTemplateSystem
```

### GitHub 통합 테스트

- **다양한 저장소 크기**: 소규모부터 수천개 저장소까지
- **권한 수준별**: 관리자, 쓰기, 읽기 권한으로 테스트
- **네트워크 장애**: GitHub API 연결 실패 시나리오
- **설정 충돌**: 기존 설정과 새 설정 간 충돌 처리

## 📈 성능 최적화

### API 호출 최적화

```go
// ✅ GraphQL 배치 쿼리 활용
func (c *GitHubClient) FetchRepositoriesBatch(org string, limit int) ([]Repository, error) {
    // REST API 대신 GraphQL 사용하여 한 번에 많은 데이터 조회
    query := `
    query($org: String!, $limit: Int!) {
        organization(login: $org) {
            repositories(first: $limit) {
                nodes {
                    name
                    description
                    isPrivate
                    branchProtectionRules(first: 10) { ... }
                }
            }
        }
    }`

    return c.executeGraphQLQuery(query, map[string]interface{}{
        "org":   org,
        "limit": limit,
    })
}
```

### 병렬 처리 최적화

- **워커 풀 크기**: GitHub API 속도 제한 고려하여 조절
- **배치 처리**: 관련 저장소를 그룹핑하여 효율성 증대
- **캐싱**: 반복 조회하는 메타데이터 캐싱

## 🔧 디버깅 가이드

### 일반적인 문제 해결

```bash
# 설정 차이 확인
gz repo-config diff --org myorg --show-details

# 드라이런으로 변경사항 미리보기
gz repo-config apply --dry-run --org myorg

# 특정 저장소 상세 검증
gz repo-config validate --repo myorg/myrepo --verbose

# 컴플라이언스 리포트 생성
gz repo-config audit --org myorg --format json
```

### 주요 문제 패턴

1. **API 속도 제한**: `--parallel` 값 조정 및 대기 시간 증가
1. **권한 부족**: 조직 관리자 권한 및 토큰 스코프 확인
1. **설정 충돌**: 기존 설정과 템플릿 간 우선순위 정리
1. **대량 작업 실패**: 배치 크기 줄이고 재시도 정책 조정

## 🚨 위험 관리

### 프로덕션 저장소 보호

```go
// ✅ 프로덕션 저장소 보호 장치
func (r *RepoManager) isProductionRepo(repo Repository) bool {
    productionPatterns := []string{
        "^prod-",
        "^production-",
        "-prod$",
        "-production$",
    }

    for _, pattern := range productionPatterns {
        if matched, _ := regexp.MatchString(pattern, repo.Name); matched {
            return true
        }
    }
    return false
}

func (r *RepoManager) requiresAdditionalConfirmation(repo Repository) bool {
    return r.isProductionRepo(repo) || repo.IsPublic || repo.HasActiveIssues
}
```

**핵심**: repo-config는 조직의 모든 저장소에 영향을 줄 수 있으므로, 안전한 배치 처리와 충분한 백업/롤백 기능이 필수입니다.
