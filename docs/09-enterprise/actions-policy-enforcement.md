# GitHub Actions 정책 적용 및 검증 시스템

이 문서는 GitHub Actions 정책 적용 및 검증 시스템의 기능과 사용법을 설명합니다.

## 개요

Actions 정책 적용 및 검증 시스템은 GitHub Actions 정책을 실제 리포지토리에 적용하고, 정책 준수 여부를 검증하는 포괄적인 솔루션입니다. 자동화된 정책 적용, 실시간 검증, 위반 사항 모니터링 기능을 제공합니다.

## 주요 기능

### 🔧 정책 적용 (Policy Enforcement)
- GitHub API를 통한 실제 설정 변경
- 단계별 적용 및 롤백 지원
- 배치 처리를 통한 대량 리포지토리 관리
- 적용 결과 추적 및 로깅

### 🔍 정책 검증 (Policy Validation)
- 실시간 정책 준수 검증
- 다양한 검증 규칙 엔진
- 위반 사항 심각도 분류
- 자동 개선 제안 생성

### 📊 규정 준수 모니터링
- 조직/리포지토리별 준수 현황 추적
- 정책 위반 추세 분석
- 자동 알림 및 보고서 생성
- 대시보드를 통한 시각화

### 🚨 위반 사항 관리
- 정책 위반 자동 탐지
- 위반 유형별 분류 및 우선순위 설정
- 위반 사항 해결 과정 추적
- 반복 위반 패턴 분석

## 시스템 구성

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Tool      │    │ Policy Enforcer │    │ Validation Rules│
│                 │    │                 │    │                 │
│ - Create Policy │    │ - Apply Changes │    │ - Permission    │
│ - Enforce       │    │ - Validate      │    │ - Security      │
│ - Monitor       │    │ - Track Results │    │ - Secrets       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │ Policy Manager  │
                    │                 │
                    │ - CRUD Operations│
                    │ - Version Control│
                    │ - Tag Management │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │  GitHub API     │
                    │                 │
                    │ - Repository    │
                    │ - Actions       │
                    │ - Security      │
                    └─────────────────┘
```

## 핵심 구성 요소

### 1. ActionsPolicyEnforcer

정책 적용과 검증을 담당하는 핵심 컴포넌트입니다.

```go
type ActionsPolicyEnforcer struct {
    logger         Logger
    apiClient      APIClient
    policyManager  *ActionsPolicyManager
    validationRules []PolicyValidationRule
}
```

**주요 메서드:**
- `EnforcePolicy()`: 정책을 리포지토리에 적용
- `ValidatePolicy()`: 정책 준수 여부 검증
- `GetRepositoryActionsState()`: 현재 설정 상태 조회

### 2. 검증 규칙 (Validation Rules)

각 정책 영역별로 특화된 검증 규칙을 제공합니다.

#### PermissionLevelValidationRule
Actions 권한 수준 검증
- 권한 상승 탐지
- 정책 불일치 확인
- 보안 위험도 평가

#### WorkflowPermissionsValidationRule
워크플로우 토큰 권한 검증
- 기본 권한 수준 확인
- 개별 권한 범위 검증
- 과도한 권한 탐지

#### SecuritySettingsValidationRule
보안 설정 검증
- 포크 PR 정책 확인
- 마켓플레이스 Actions 정책 검증
- 중요 보안 설정 위반 탐지

#### AllowedActionsValidationRule
허용된 Actions 검증
- 승인되지 않은 Actions 탐지
- 패턴 매칭을 통한 허용 여부 확인
- 워크플로우 히스토리 분석

#### SecretPolicyValidationRule
시크릿 정책 검증
- 시크릿 수량 제한 확인
- 네이밍 패턴 준수 검증
- 제한된 시크릿 탐지

#### RunnerPolicyValidationRule
러너 정책 검증
- 허용된 러너 유형 확인
- 셀프 호스티드 러너 제한 검증
- 필수 라벨 확인

### 3. 정책 위반 (Policy Violations)

```go
type ActionsPolicyViolation struct {
    ID            string
    PolicyID      string
    ViolationType ActionsPolicyViolationType
    Severity      PolicyViolationSeverity
    Resource      string
    Description   string
    DetectedAt    time.Time
    Status        PolicyViolationStatus
}
```

**위반 유형:**
- `unauthorized_action`: 승인되지 않은 Action 사용
- `excessive_permissions`: 과도한 권한 사용
- `secret_misuse`: 시크릿 남용
- `runner_policy_breach`: 러너 정책 위반
- `environment_breach`: 환경 정책 위반
- `workflow_permission_breach`: 워크플로우 권한 위반
- `security_settings_breach`: 보안 설정 위반

**심각도 분류:**
- `low`: 낮음 - 모니터링 필요
- `medium`: 보통 - 개선 권장
- `high`: 높음 - 조속한 해결 필요
- `critical`: 치명적 - 즉시 해결 필요

## CLI 도구 사용법

### 정책 생성

```bash
# 기본 정책 생성
actions-policy create "default-policy" --org myorg --template default

# 엄격한 보안 정책 생성
actions-policy create "strict-policy" --org myorg --template strict \
  --description "High security policy for production"

# 사용자 정의 정책 생성
actions-policy create "custom-policy" --org myorg --repo myrepo \
  --template permissive --tags security,compliance
```

### 정책 적용

```bash
# 정책 적용
actions-policy enforce policy-123 myorg myrepo

# 드라이 런 (검증만 수행)
actions-policy enforce policy-123 myorg myrepo --dry-run

# 강제 적용 (검증 실패 시에도 적용)
actions-policy enforce policy-123 myorg myrepo --force

# 타임아웃 설정
actions-policy enforce policy-123 myorg myrepo --timeout 600
```

### 정책 검증

```bash
# 기본 검증
actions-policy validate policy-123 myorg myrepo

# 상세 검증 결과
actions-policy validate policy-123 myorg myrepo --detailed

# 특정 심각도만 확인
actions-policy validate policy-123 myorg myrepo --severity critical

# JSON 형식으로 출력
actions-policy validate policy-123 myorg myrepo --format json
```

### 정책 목록 및 조회

```bash
# 전체 정책 목록
actions-policy list

# 조직별 필터링
actions-policy list --org myorg

# 활성화된 정책만
actions-policy list --enabled-only

# 태그별 필터링
actions-policy list --tags security,compliance

# 정책 상세 정보
actions-policy show policy-123

# JSON 형식으로 출력
actions-policy show policy-123 --format json
```

### 규정 준수 모니터링

```bash
# 일회성 모니터링
actions-policy monitor myorg

# 지속적 모니터링
actions-policy monitor myorg --continuous --interval 10m

# 웹훅 알림 설정
actions-policy monitor myorg --webhook-url https://hooks.example.com/alerts
```

## API 사용 예제

### 정책 적용

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gizzahub/gzh-manager-go/pkg/github"
)

func main() {
    // 컴포넌트 초기화
    logger := &consoleLogger{}
    apiClient := github.NewGitHubClient("your-token", logger)
    policyManager := github.NewActionsPolicyManager(logger, apiClient)
    enforcer := github.NewActionsPolicyEnforcer(logger, apiClient, policyManager)

    ctx := context.Background()

    // 정책 생성
    policy := github.GetDefaultActionsPolicy()
    policy.ID = "example-policy"
    policy.Organization = "myorg"
    policy.Name = "Example Policy"

    err := policyManager.CreatePolicy(ctx, policy)
    if err != nil {
        log.Fatal(err)
    }

    // 정책 적용
    result, err := enforcer.EnforcePolicy(ctx, "example-policy", "myorg", "myrepo")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Enforcement successful: %t\n", result.Success)
    fmt.Printf("Applied changes: %d\n", len(result.AppliedChanges))
    fmt.Printf("Violations: %d\n", len(result.Violations))
}
```

### 사용자 정의 검증 규칙

```go
type CustomValidationRule struct{}

func (r *CustomValidationRule) GetRuleID() string {
    return "custom_security_check"
}

func (r *CustomValidationRule) GetDescription() string {
    return "Custom security validation rule"
}

func (r *CustomValidationRule) Validate(ctx context.Context, policy *github.ActionsPolicy, currentState *github.RepositoryActionsState) (*github.PolicyValidationResult, error) {
    result := &github.PolicyValidationResult{
        RuleID: r.GetRuleID(),
    }

    // 사용자 정의 검증 로직
    if customSecurityCheck(policy, currentState) {
        result.Passed = true
        result.Message = "Custom security check passed"
        result.Severity = github.ViolationSeverityLow
    } else {
        result.Passed = false
        result.Message = "Custom security check failed"
        result.Severity = github.ViolationSeverityHigh
        result.Suggestions = []string{
            "Update configuration to meet custom security requirements",
        }
    }

    return result, nil
}

// 검증 규칙 추가
enforcer.AddValidationRule(&CustomValidationRule{})
```

### 배치 정책 적용

```go
func enforceOrgPolicy(ctx context.Context, enforcer *github.ActionsPolicyEnforcer, policyID, org string) error {
    // 조직의 모든 리포지토리 조회
    repos, err := apiClient.ListOrganizationRepositories(ctx, org)
    if err != nil {
        return err
    }

    results := make(chan *github.PolicyEnforcementResult, len(repos))
    errors := make(chan error, len(repos))

    // 병렬 처리
    for _, repo := range repos {
        go func(repoName string) {
            result, err := enforcer.EnforcePolicy(ctx, policyID, org, repoName)
            if err != nil {
                errors <- err
                return
            }
            results <- result
        }(repo.Name)
    }

    // 결과 수집
    successCount := 0
    failCount := 0

    for i := 0; i < len(repos); i++ {
        select {
        case result := <-results:
            if result.Success {
                successCount++
            } else {
                failCount++
            }
        case err := <-errors:
            log.Printf("Error enforcing policy: %v", err)
            failCount++
        }
    }

    fmt.Printf("Policy enforcement completed: %d success, %d failed\n", successCount, failCount)
    return nil
}
```

## 모니터링 및 알림

### 정책 위반 모니터링

```go
func monitorCompliance(ctx context.Context, enforcer *github.ActionsPolicyEnforcer, org string) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            violations := checkOrgCompliance(ctx, enforcer, org)
            if len(violations) > 0 {
                sendAlerts(violations)
            }
        case <-ctx.Done():
            return
        }
    }
}

func checkOrgCompliance(ctx context.Context, enforcer *github.ActionsPolicyEnforcer, org string) []github.ActionsPolicyViolation {
    var allViolations []github.ActionsPolicyViolation

    // 조직의 정책들을 조회하여 각 리포지토리에 대해 검증
    // 실제 구현에서는 병렬 처리 및 에러 핸들링 추가

    return allViolations
}

func sendAlerts(violations []github.ActionsPolicyViolation) {
    critical := 0
    high := 0

    for _, v := range violations {
        switch v.Severity {
        case github.ViolationSeverityCritical:
            critical++
        case github.ViolationSeverityHigh:
            high++
        }
    }

    if critical > 0 {
        // 즉시 알림 발송
        sendCriticalAlert(critical, violations)
    }

    if high > 0 {
        // 일반 알림 발송
        sendHighPriorityAlert(high, violations)
    }
}
```

### 대시보드 데이터 생성

```go
type ComplianceDashboard struct {
    Organization     string                           `json:"organization"`
    TotalPolicies    int                             `json:"total_policies"`
    ActivePolicies   int                             `json:"active_policies"`
    TotalRepos       int                             `json:"total_repositories"`
    CompliantRepos   int                             `json:"compliant_repositories"`
    ViolationsByType map[string]int                  `json:"violations_by_type"`
    TrendData        []ComplianceTrendPoint          `json:"trend_data"`
    LastUpdated      time.Time                       `json:"last_updated"`
}

func generateComplianceDashboard(ctx context.Context, org string) (*ComplianceDashboard, error) {
    dashboard := &ComplianceDashboard{
        Organization:     org,
        ViolationsByType: make(map[string]int),
        LastUpdated:      time.Now(),
    }

    // 정책 수집
    policies, err := policyManager.ListPolicies(ctx, org)
    if err != nil {
        return nil, err
    }

    dashboard.TotalPolicies = len(policies)

    activePolicies := 0
    for _, policy := range policies {
        if policy.Enabled {
            activePolicies++
        }
    }
    dashboard.ActivePolicies = activePolicies

    // 리포지토리 규정 준수 상태 수집
    repos, err := apiClient.ListOrganizationRepositories(ctx, org)
    if err != nil {
        return nil, err
    }

    dashboard.TotalRepos = len(repos)

    compliantCount := 0
    for _, repo := range repos {
        isCompliant := checkRepositoryCompliance(ctx, repo.Name, policies)
        if isCompliant {
            compliantCount++
        }
    }
    dashboard.CompliantRepos = compliantCount

    return dashboard, nil
}
```

## 모범 사례

### 1. 단계별 정책 적용
```go
// 1단계: 검증만 수행
result, err := enforcer.ValidatePolicy(ctx, policy, currentState)

// 2단계: 위험도가 낮은 변경사항만 적용
if canSafelyApply(result) {
    enforcer.EnforcePolicy(ctx, policyID, org, repo)
}

// 3단계: 전체 정책 적용
enforcer.EnforcePolicy(ctx, policyID, org, repo)
```

### 2. 정책 버전 관리
```go
// 정책 업데이트 시 버전 증가
policy.Version++
policy.UpdatedAt = time.Now()
policy.UpdatedBy = "admin"

// 이전 버전과의 호환성 검증
if err := validateBackwardCompatibility(oldPolicy, policy); err != nil {
    return err
}
```

### 3. 점진적 배포
```go
// 소수의 리포지토리에서 테스트
testRepos := []string{"test-repo-1", "test-repo-2"}
for _, repo := range testRepos {
    result, err := enforcer.EnforcePolicy(ctx, policyID, org, repo)
    if err != nil || !result.Success {
        return fmt.Errorf("test deployment failed")
    }
}

// 전체 조직에 배포
enforceOrgPolicy(ctx, enforcer, policyID, org)
```

### 4. 예외 처리
```go
type PolicyException struct {
    Repository  string    `json:"repository"`
    PolicyID    string    `json:"policy_id"`
    Reason      string    `json:"reason"`
    ExpiresAt   time.Time `json:"expires_at"`
    ApprovedBy  string    `json:"approved_by"`
}

func isExempt(repo, policyID string) bool {
    // 예외 승인 여부 확인
    return checkException(repo, policyID)
}
```

## 성능 최적화

### 1. 배치 처리
- 동시성 제어를 통한 병렬 처리
- API 레이트 리밋 고려
- 에러 복구 및 재시도 로직

### 2. 캐싱
- 정책 정보 캐싱
- 리포지토리 상태 캐싱
- 검증 결과 캐싱

### 3. 증분 업데이트
- 변경된 항목만 업데이트
- 델타 기반 적용
- 최적화된 API 호출

## 보안 고려사항

### 1. 권한 관리
- 최소 권한 원칙 적용
- 정책별 접근 제어
- 감사 로그 유지

### 2. 민감 정보 보호
- 시크릿 정보 암호화
- 로그에서 민감 정보 제거
- 안전한 토큰 관리

### 3. 무결성 검증
- 정책 변경 추적
- 변경 사항 승인 프로세스
- 롤백 기능 제공

## 문제 해결

### 일반적인 문제

1. **정책 적용 실패**
   - GitHub API 권한 확인
   - 네트워크 연결 상태 확인
   - 리포지토리 설정 권한 확인

2. **검증 오류**
   - 정책 정의 검토
   - 검증 규칙 로직 확인
   - 리포지토리 상태 정보 확인

3. **성능 문제**
   - API 호출 최적화
   - 배치 크기 조정
   - 동시성 설정 튜닝

### 디버깅

```bash
# 상세 로그 활성화
actions-policy enforce policy-123 myorg myrepo --verbose

# 드라이 런으로 문제 파악
actions-policy enforce policy-123 myorg myrepo --dry-run --detailed

# 개별 검증 규칙 테스트
actions-policy validate policy-123 myorg myrepo --severity high --detailed
```

## 확장성

### 1. 사용자 정의 규칙
새로운 검증 규칙을 쉽게 추가할 수 있는 플러그인 시스템

### 2. 다양한 백엔드 지원
- GitHub Enterprise Server
- GitHub.com
- 기타 Git 플랫폼

### 3. 통합 지원
- CI/CD 파이프라인 통합
- 모니터링 시스템 연동
- 알림 채널 확장

이 시스템을 통해 GitHub Actions의 보안과 규정 준수를 체계적으로 관리할 수 있습니다.
