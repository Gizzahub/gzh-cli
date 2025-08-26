# GitHub Actions 권한 정책 스키마

이 문서는 GitHub Actions 권한 정책 스키마의 구조와 사용법을 설명합니다.

## 개요

GitHub Actions 권한 정책 스키마는 조직과 리포지토리에서 GitHub Actions의 권한과 보안 설정을 체계적으로 관리하기 위한 포괄적인 데이터 구조입니다.

## 정책 구조

### 기본 정책 (ActionsPolicy)

```go
type ActionsPolicy struct {
    ID                     string                    // 정책 고유 식별자
    Name                   string                    // 정책 이름
    Description            string                    // 정책 설명
    Organization           string                    // 대상 조직
    Repository             string                    // 대상 리포지토리 (선택사항)
    PermissionLevel        ActionsPermissionLevel    // Actions 권한 수준
    AllowedActions         []string                  // 허용된 Actions 목록
    AllowedActionsPatterns []string                  // 허용된 Actions 패턴
    WorkflowPermissions    WorkflowPermissions       // 워크플로우 권한 설정
    SecuritySettings       ActionsSecuritySettings   // 보안 설정
    SecretsPolicy          SecretsPolicy             // 시크릿 정책
    Variables              map[string]string         // 환경 변수
    Environments           []EnvironmentPolicy       // 환경별 정책
    Runners                RunnerPolicy              // 러너 정책
    CreatedAt              time.Time                 // 생성 시간
    UpdatedAt              time.Time                 // 수정 시간
    CreatedBy              string                    // 생성자
    UpdatedBy              string                    // 수정자
    Version                int                       // 정책 버전
    Enabled                bool                      // 정책 활성화 여부
    Tags                   []string                  // 정책 태그
}
```

## 권한 수준 (ActionsPermissionLevel)

| 값 | 설명 |
| ------------ | --------------------- |
| `disabled` | Actions 완전 비활성화 |
| `all` | 모든 Actions 허용 |
| `local_only` | 로컬 Actions만 허용 |
| `selected` | 선택된 Actions만 허용 |

## 워크플로우 권한 (WorkflowPermissions)

### 기본 권한 (DefaultPermissions)

| 값 | 설명 |
| ------------ | ------------------- |
| `read` | 읽기 권한만 부여 |
| `write` | 읽기/쓰기 권한 부여 |
| `restricted` | 최소 권한만 부여 |

### 토큰 권한 (ActionsTokenPermission)

| 값 | 설명 |
| ------- | --------- |
| `none` | 권한 없음 |
| `read` | 읽기 권한 |
| `write` | 쓰기 권한 |

### 세부 권한 설정

```go
type WorkflowPermissions struct {
    DefaultPermissions       DefaultPermissions        // 기본 권한 수준
    CanApproveOwnChanges    bool                      // 자신의 변경사항 승인 허용
    ActionsReadPermission   ActionsTokenPermission    // Actions 읽기 권한
    ContentsPermission      ActionsTokenPermission    // 컨텐츠 권한
    MetadataPermission      ActionsTokenPermission    // 메타데이터 권한
    PackagesPermission      ActionsTokenPermission    // 패키지 권한
    PullRequestsPermission  ActionsTokenPermission    // PR 권한
    IssuesPermission        ActionsTokenPermission    // 이슈 권한
    DeploymentsPermission   ActionsTokenPermission    // 배포 권한
    ChecksPermission        ActionsTokenPermission    // 체크 권한
    StatusesPermission      ActionsTokenPermission    // 상태 권한
    SecurityEventsPermission ActionsTokenPermission   // 보안 이벤트 권한
    IdTokenPermission       ActionsTokenPermission    // ID 토큰 권한
    AttestationsPermission  ActionsTokenPermission    // 증명 권한
    CustomPermissions       map[string]ActionsTokenPermission // 사용자 정의 권한
}
```

## 보안 설정 (ActionsSecuritySettings)

```go
type ActionsSecuritySettings struct {
    RequireCodeScanningApproval     bool                     // 코드 스캐닝 승인 필요
    RequireSecretScanningApproval   bool                     // 시크릿 스캐닝 승인 필요
    AllowForkPRs                   bool                     // 포크 PR 허용
    RequireApprovalForForkPRs      bool                     // 포크 PR 승인 필요
    AllowPrivateRepoForkRun        bool                     // 프라이빗 리포 포크 실행 허용
    RequireApprovalForPrivateFork  bool                     // 프라이빗 포크 승인 필요
    RestrictedActionsPatterns      []string                 // 제한된 Actions 패턴
    AllowGitHubOwnedActions        bool                     // GitHub 소유 Actions 허용
    AllowVerifiedPartnerActions    bool                     // 검증된 파트너 Actions 허용
    AllowMarketplaceActions        ActionsMarketplacePolicy // 마켓플레이스 Actions 정책
    RequireSignedCommits           bool                     // 서명된 커밋 필요
    EnforceAdminsOnBranches        bool                     // 브랜치에서 관리자 강제 적용
    OIDCCustomClaims               map[string]string        // OIDC 사용자 정의 클레임
}
```

### 마켓플레이스 정책 (ActionsMarketplacePolicy)

| 값 | 설명 |
| --------------- | ------------------------------ |
| `disabled` | 마켓플레이스 Actions 비활성화 |
| `verified_only` | 검증된 Actions만 허용 |
| `all` | 모든 마켓플레이스 Actions 허용 |
| `selected` | 선택된 Actions만 허용 |

## 시크릿 정책 (SecretsPolicy)

```go
type SecretsPolicy struct {
    AllowedSecrets                []string             // 허용된 시크릿 목록
    RestrictedSecrets             []string             // 제한된 시크릿 목록
    RequireApprovalForNewSecrets  bool                 // 새 시크릿 승인 필요
    SecretVisibility              SecretVisibility     // 시크릿 가시성
    AllowSecretsInheritance       bool                 // 시크릿 상속 허용
    SecretNamingPatterns          []string             // 시크릿 네이밍 패턴
    MaxSecretCount                int                  // 최대 시크릿 수
    SecretRotationPolicy          SecretRotationPolicy // 시크릿 교체 정책
}
```

### 시크릿 가시성 (SecretVisibility)

| 값 | 설명 |
| ---------- | ------------------------------- |
| `all` | 모든 리포지토리에서 접근 가능 |
| `private` | 프라이빗 리포지토리만 접근 가능 |
| `selected` | 선택된 리포지토리만 접근 가능 |

### 시크릿 교체 정책 (SecretRotationPolicy)

```go
type SecretRotationPolicy struct {
    Enabled                 bool          // 자동 교체 활성화
    RotationInterval        time.Duration // 교체 주기
    RequireRotationWarning  bool          // 교체 경고 필요
    WarningDays             int           // 경고 일수
    AutoRotateSecrets       []string      // 자동 교체 시크릿 목록
}
```

## 환경 정책 (EnvironmentPolicy)

```go
type EnvironmentPolicy struct {
    Name                     string                   // 환경 이름
    RequiredReviewers        []string                 // 필수 리뷰어
    RequiredReviewerTeams    []string                 // 필수 리뷰어 팀
    WaitTimer                time.Duration            // 대기 시간
    BranchPolicyType         EnvironmentBranchPolicy  // 브랜치 정책 유형
    ProtectedBranches        []string                 // 보호된 브랜치
    BranchPatterns           []string                 // 브랜치 패턴
    RequireDeploymentBranch  bool                     // 배포 브랜치 필요
    PreventSelfReview        bool                     // 자가 리뷰 방지
    Secrets                  []string                 // 환경 시크릿
    Variables                map[string]string        // 환경 변수
}
```

### 환경 브랜치 정책 (EnvironmentBranchPolicy)

| 값 | 설명 |
| ----------- | ----------------------------- |
| `all` | 모든 브랜치에서 배포 허용 |
| `protected` | 보호된 브랜치에서만 배포 허용 |
| `selected` | 선택된 브랜치에서만 배포 허용 |
| `none` | 브랜치 제한 없음 |

## 러너 정책 (RunnerPolicy)

```go
type RunnerPolicy struct {
    AllowedRunnerTypes       []RunnerType             // 허용된 러너 유형
    RequireSelfHostedLabels  []string                 // 필수 셀프 호스티드 라벨
    RestrictedRunnerLabels   []string                 // 제한된 러너 라벨
    MaxConcurrentJobs        int                      // 최대 동시 작업 수
    MaxJobExecutionTime      time.Duration            // 최대 작업 실행 시간
    RunnerGroups             []string                 // 러너 그룹
    RequireRunnerApproval    bool                     // 러너 승인 필요
    SelfHostedRunnerPolicy   SelfHostedRunnerPolicy   // 셀프 호스티드 러너 정책
}
```

### 러너 유형 (RunnerType)

| 값 | 설명 |
| --------------- | -------------------- |
| `github_hosted` | GitHub 호스티드 러너 |
| `self_hosted` | 셀프 호스티드 러너 |
| `organization` | 조직 러너 |
| `repository` | 리포지토리 러너 |

### 셀프 호스티드 러너 정책 (SelfHostedRunnerPolicy)

```go
type SelfHostedRunnerPolicy struct {
    RequireRunnerRegistration   bool          // 러너 등록 필요
    AllowedOperatingSystems     []string      // 허용된 운영체제
    RequiredSecurityPatches     bool          // 보안 패치 필요
    DisallowPublicRepositories  bool          // 공개 리포지토리 비허용
    RequireEncryptedStorage     bool          // 암호화된 스토리지 필요
    RunnerTimeout               time.Duration // 러너 타임아웃
    MaxRunners                  int           // 최대 러너 수
}
```

## 정책 위반 (ActionsPolicyViolation)

```go
type ActionsPolicyViolation struct {
    ID            string                         // 위반 식별자
    PolicyID      string                         // 정책 식별자
    ViolationType ActionsPolicyViolationType     // 위반 유형
    Severity      PolicyViolationSeverity        // 심각도
    Resource      string                         // 리소스
    Description   string                         // 위반 설명
    Details       map[string]interface{}         // 상세 정보
    DetectedAt    time.Time                      // 탐지 시간
    ResolvedAt    *time.Time                     // 해결 시간
    Status        PolicyViolationStatus          // 위반 상태
}
```

### 위반 유형 (ActionsPolicyViolationType)

| 값 | 설명 |
| ---------------------------- | ------------------------- |
| `unauthorized_action` | 허가되지 않은 Action 사용 |
| `excessive_permissions` | 과도한 권한 사용 |
| `secret_misuse` | 시크릿 남용 |
| `runner_policy_breach` | 러너 정책 위반 |
| `environment_breach` | 환경 정책 위반 |
| `workflow_permission_breach` | 워크플로우 권한 위반 |
| `security_settings_breach` | 보안 설정 위반 |

### 위반 심각도 (PolicyViolationSeverity)

| 값 | 설명 |
| ---------- | ------ |
| `low` | 낮음 |
| `medium` | 보통 |
| `high` | 높음 |
| `critical` | 치명적 |

### 위반 상태 (PolicyViolationStatus)

| 값 | 설명 |
| ------------- | ------- |
| `open` | 열림 |
| `in_progress` | 진행 중 |
| `resolved` | 해결됨 |
| `ignored` | 무시됨 |

## 사용 예제

### 기본 정책 생성

```go
// 기본 정책 템플릿 가져오기
defaultPolicy := GetDefaultActionsPolicy()
defaultPolicy.ID = "org-default-policy"
defaultPolicy.Organization = "myorg"
defaultPolicy.CreatedBy = "admin"

// 정책 관리자 생성
manager := NewActionsPolicyManager(logger, apiClient)

// 정책 생성
err := manager.CreatePolicy(ctx, defaultPolicy)
if err != nil {
    log.Fatal(err)
}
```

### 사용자 정의 정책 생성

```go
customPolicy := &ActionsPolicy{
    ID:           "custom-strict-policy",
    Name:         "Strict Security Policy",
    Description:  "엄격한 보안 정책",
    Organization: "secureorg",
    PermissionLevel: ActionsPermissionSelectedActions,
    AllowedActions: []string{
        "actions/checkout@v4",
        "actions/setup-go@v4",
    },
    WorkflowPermissions: WorkflowPermissions{
        DefaultPermissions:     DefaultPermissionsRestricted,
        CanApproveOwnChanges:   false,
        ContentsPermission:     TokenPermissionRead,
        PullRequestsPermission: TokenPermissionNone,
    },
    SecuritySettings: ActionsSecuritySettings{
        RequireCodeScanningApproval:   true,
        RequireSecretScanningApproval: true,
        AllowForkPRs:                 false,
        AllowGitHubOwnedActions:      true,
        AllowVerifiedPartnerActions:  false,
        AllowMarketplaceActions:      MarketplacePolicyDisabled,
        RequireSignedCommits:         true,
    },
    SecretsPolicy: SecretsPolicy{
        RequireApprovalForNewSecrets: true,
        SecretVisibility:            SecretVisibilityPrivate,
        MaxSecretCount:              10,
        SecretRotationPolicy: SecretRotationPolicy{
            Enabled:          true,
            RotationInterval: 30 * 24 * time.Hour, // 30일
            WarningDays:      7,
        },
    },
    Environments: []EnvironmentPolicy{
        {
            Name:                   "production",
            RequiredReviewers:      []string{"admin", "security-team"},
            WaitTimer:              30 * time.Minute,
            BranchPolicyType:       EnvironmentBranchPolicyProtected,
            ProtectedBranches:      []string{"main", "release/*"},
            PreventSelfReview:      true,
        },
    },
    Runners: RunnerPolicy{
        AllowedRunnerTypes:    []RunnerType{RunnerTypeGitHubHosted},
        MaxConcurrentJobs:     2,
        MaxJobExecutionTime:   2 * time.Hour,
        RequireRunnerApproval: true,
    },
    Enabled: true,
}

err := manager.CreatePolicy(ctx, customPolicy)
if err != nil {
    log.Fatal(err)
}
```

### 정책 조회 및 관리

```go
// 정책 조회
policy, err := manager.GetPolicy(ctx, "custom-strict-policy")
if err != nil {
    log.Fatal(err)
}

// 조직의 모든 정책 조회
policies, err := manager.ListPolicies(ctx, "myorg")
if err != nil {
    log.Fatal(err)
}

// 정책 업데이트
policy.Description = "업데이트된 엄격한 보안 정책"
policy.UpdatedBy = "admin"
err = manager.UpdatePolicy(ctx, policy.ID, policy)
if err != nil {
    log.Fatal(err)
}

// 정책 삭제
err = manager.DeletePolicy(ctx, "old-policy-id")
if err != nil {
    log.Fatal(err)
}
```

## 정책 검증

정책 관리자는 자동으로 정책의 유효성을 검증합니다:

- 필수 필드 검증
- 권한 수준 유효성 검사
- 환경 설정 검증
- 러너 정책 검증
- 시크릿 정책 일관성 검사

## 모범 사례

### 1. 점진적 권한 부여

- 최소 권한 원칙으로 시작
- 필요에 따라 단계적으로 권한 확장
- 정기적인 권한 검토 및 정리

### 2. 환경별 정책 분리

- 개발, 스테이징, 프로덕션 환경별 다른 정책 적용
- 프로덕션 환경에 더 엄격한 정책 적용
- 환경별 시크릿 및 변수 관리

### 3. 모니터링 및 감사

- 정책 위반 모니터링
- 정기적인 정책 준수 감사
- 위반 사항에 대한 자동 알림 설정

### 4. 문서화 및 교육

- 정책 변경 사항 문서화
- 개발팀 대상 정책 교육
- 정책 위반 시 대응 절차 수립

## 제한사항

- 정책은 조직 또는 리포지토리 수준에서만 적용 가능
- 일부 GitHub Enterprise 기능은 추가 라이센스가 필요할 수 있음
- 정책 변경 사항은 즉시 적용되지 않을 수 있음

## 참고 자료

- [GitHub Actions 보안 가이드](https://docs.github.com/en/actions/security-guides)
- [워크플로우 권한 관리](https://docs.github.com/en/actions/using-jobs/assigning-permissions-to-jobs)
- [시크릿 관리 모범 사례](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [셀프 호스티드 러너 보안](https://docs.github.com/en/actions/hosting-your-own-runners/about-self-hosted-runners#self-hosted-runner-security)
