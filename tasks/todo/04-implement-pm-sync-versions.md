# TODO: 패키지 매니저 버전 동기화 기능 구현

- status: [ ]
- priority: medium (P2)
- category: package-manager
- estimated_effort: 1시간
- depends_on: [02-implement-pm-bootstrap.md, 03-implement-pm-upgrade-managers.md]
- spec_reference: `cmd/pm/advanced.go:106`, `specs/package-manager.md`

## 📋 작업 개요

버전 매니저(nvm, rbenv, pyenv)와 그들이 관리하는 패키지 매니저(npm, gem, pip) 간의 버전 동기화 기능을 구현합니다. 현재 "not yet implemented" 상태인 sync-versions 명령어를 완전히 구현하여 일관된 개발 환경을 유지할 수 있도록 합니다.

## 🎯 구현 목표

### 핵심 기능
- [ ] **버전 불일치 감지** - 버전 매니저와 패키지 매니저 간 버전 차이 확인
- [ ] **자동 동기화** - 불일치 해결을 위한 자동 버전 조정
- [ ] **동기화 정책** - 어떤 버전을 기준으로 할지 정책 설정
- [ ] **충돌 해결** - 여러 버전이 설치된 경우 우선순위 결정

### 지원할 버전 매니저 쌍
- [ ] **nvm ↔ npm** - Node.js 버전과 npm 버전 동기화
- [ ] **rbenv ↔ gem** - Ruby 버전과 gem 버전 동기화
- [ ] **pyenv ↔ pip** - Python 버전과 pip 버전 동기화
- [ ] **asdf ↔ multiple** - asdf가 관리하는 모든 도구들과 패키지 매니저들

## 🔧 기술적 구현

### 1. 버전 동기화 상태 구조체
```go
type VersionSyncStatus struct {
    VersionManager    string          `json:"version_manager"`
    PackageManager    string          `json:"package_manager"`
    VMVersion         string          `json:"vm_version"`
    PMVersion         string          `json:"pm_version"`
    ExpectedPMVersion string          `json:"expected_pm_version"`
    InSync            bool            `json:"in_sync"`
    SyncAction        string          `json:"sync_action"`
    Issues            []string        `json:"issues,omitempty"`
}

type SyncReport struct {
    Platform       string              `json:"platform"`
    TotalPairs     int                 `json:"total_pairs"`
    InSyncCount    int                 `json:"in_sync_count"`
    OutOfSyncCount int                 `json:"out_of_sync_count"`
    SyncStatuses   []VersionSyncStatus `json:"sync_statuses"`
    Timestamp      time.Time           `json:"timestamp"`
}

type SyncPolicy struct {
    Strategy        string `json:"strategy"`         // "vm_priority", "pm_priority", "latest"
    AutoFix         bool   `json:"auto_fix"`
    BackupEnabled   bool   `json:"backup_enabled"`
    PromptUser      bool   `json:"prompt_user"`
}
```

### 2. 동기화 인터페이스
```go
type VersionSynchronizer interface {
    CheckSync(ctx context.Context) (*VersionSyncStatus, error)
    Synchronize(ctx context.Context, policy SyncPolicy) error
    GetExpectedVersion(ctx context.Context, vmVersion string) (string, error)
    ValidateSync(ctx context.Context) error
}

type SyncManager struct {
    synchronizers map[string]VersionSynchronizer
    policy        SyncPolicy
    logger        logger.Logger
}
```

### 3. 개별 동기화 구현
```go
// NVM ↔ NPM 동기화
type NvmNpmSynchronizer struct {
    logger logger.Logger
}

func (nns *NvmNpmSynchronizer) CheckSync(ctx context.Context) (*VersionSyncStatus, error) {
    // nvm current로 현재 Node.js 버전 확인
    nodeVersion, err := nns.getCurrentNodeVersion(ctx)
    if err != nil {
        return nil, err
    }

    // npm --version으로 현재 npm 버전 확인
    npmVersion, err := nns.getCurrentNpmVersion(ctx)
    if err != nil {
        return nil, err
    }

    // Node.js 버전에 기본 포함된 npm 버전 확인
    expectedNpmVersion, err := nns.getExpectedNpmVersion(ctx, nodeVersion)
    if err != nil {
        return nil, err
    }

    inSync := nns.compareVersions(npmVersion, expectedNpmVersion)

    return &VersionSyncStatus{
        VersionManager:    "nvm",
        PackageManager:    "npm",
        VMVersion:         nodeVersion,
        PMVersion:         npmVersion,
        ExpectedPMVersion: expectedNpmVersion,
        InSync:            inSync,
        SyncAction:        nns.determineSyncAction(npmVersion, expectedNpmVersion),
    }, nil
}

func (nns *NvmNpmSynchronizer) Synchronize(ctx context.Context, policy SyncPolicy) error {
    status, err := nns.CheckSync(ctx)
    if err != nil {
        return err
    }

    if status.InSync {
        return nil // 이미 동기화됨
    }

    switch policy.Strategy {
    case "vm_priority":
        // Node.js 버전에 맞는 npm 설치
        return nns.installMatchingNpm(ctx, status.VMVersion)
    case "pm_priority":
        // npm 버전에 맞는 Node.js 설치
        return nns.installMatchingNode(ctx, status.PMVersion)
    case "latest":
        // 둘 다 최신 버전으로 업데이트
        return nns.upgradeToLatest(ctx)
    }

    return nil
}

// rbenv ↔ gem 동기화
type RbenvGemSynchronizer struct {
    logger logger.Logger
}

func (rgs *RbenvGemSynchronizer) CheckSync(ctx context.Context) (*VersionSyncStatus, error) {
    // rbenv version으로 현재 Ruby 버전 확인
    rubyVersion, err := rgs.getCurrentRubyVersion(ctx)
    if err != nil {
        return nil, err
    }

    // gem --version으로 현재 gem 버전 확인
    gemVersion, err := rgs.getCurrentGemVersion(ctx)
    if err != nil {
        return nil, err
    }

    // Ruby 버전에 기본 포함된 gem 버전 확인
    expectedGemVersion, err := rgs.getExpectedGemVersion(ctx, rubyVersion)
    if err != nil {
        return nil, err
    }

    return &VersionSyncStatus{
        VersionManager:    "rbenv",
        PackageManager:    "gem",
        VMVersion:         rubyVersion,
        PMVersion:         gemVersion,
        ExpectedPMVersion: expectedGemVersion,
        InSync:            rgs.compareVersions(gemVersion, expectedGemVersion),
    }, nil
}

// pyenv ↔ pip 동기화
type PyenvPipSynchronizer struct {
    logger logger.Logger
}

func (pps *PyenvPipSynchronizer) CheckSync(ctx context.Context) (*VersionSyncStatus, error) {
    // pyenv version으로 현재 Python 버전 확인
    pythonVersion, err := pps.getCurrentPythonVersion(ctx)
    if err != nil {
        return nil, err
    }

    // pip --version으로 현재 pip 버전 확인
    pipVersion, err := pps.getCurrentPipVersion(ctx)
    if err != nil {
        return nil, err
    }

    // Python 버전에 기본 포함된 pip 버전 확인
    expectedPipVersion, err := pps.getExpectedPipVersion(ctx, pythonVersion)
    if err != nil {
        return nil, err
    }

    return &VersionSyncStatus{
        VersionManager:    "pyenv",
        PackageManager:    "pip",
        VMVersion:         pythonVersion,
        PMVersion:         pipVersion,
        ExpectedPMVersion: expectedPipVersion,
        InSync:            pps.compareVersions(pipVersion, expectedPipVersion),
    }, nil
}
```

### 4. 동기화 정책 시스템
```go
type PolicyEngine struct {
    defaultPolicy SyncPolicy
    customPolicies map[string]SyncPolicy
}

func (pe *PolicyEngine) GetPolicy(managerPair string) SyncPolicy {
    if policy, exists := pe.customPolicies[managerPair]; exists {
        return policy
    }
    return pe.defaultPolicy
}

func (pe *PolicyEngine) ApplyPolicy(ctx context.Context, status *VersionSyncStatus, policy SyncPolicy) error {
    if policy.PromptUser {
        return pe.promptUserForAction(status)
    }

    if policy.AutoFix {
        return pe.autoFixSync(ctx, status, policy)
    }

    return nil
}
```

## 📁 파일 구조

### 새로 생성할 파일
- `internal/pm/sync/manager.go` - 동기화 매니저 구현
- `internal/pm/sync/nvm_npm.go` - nvm-npm 동기화 로직
- `internal/pm/sync/rbenv_gem.go` - rbenv-gem 동기화 로직
- `internal/pm/sync/pyenv_pip.go` - pyenv-pip 동기화 로직
- `internal/pm/sync/asdf_multi.go` - asdf 다중 도구 동기화
- `internal/pm/sync/policy.go` - 동기화 정책 엔진
- `internal/pm/sync/version_resolver.go` - 버전 호환성 해결

### 수정할 파일
- `cmd/pm/advanced.go` - sync-versions 명령어 실제 구현

## 🎯 명령어 구조

### 현재 명령어 확장
```bash
# 동기화 상태 확인
gz pm sync-versions --check
gz pm sync-versions --check --json

# 불일치 자동 수정
gz pm sync-versions --fix

# 특정 매니저 쌍만 확인
gz pm sync-versions --check --pair nvm-npm
gz pm sync-versions --fix --pair rbenv-gem

# 동기화 정책 지정
gz pm sync-versions --fix --strategy vm_priority
gz pm sync-versions --fix --strategy pm_priority
gz pm sync-versions --fix --strategy latest

# 백업과 함께 동기화
gz pm sync-versions --fix --backup
```

### 출력 예시
```
🔄 Package Manager Version Synchronization Status

Checking version synchronization...

Version Manager Pairs:
  ✅ nvm (v0.39.0) ↔ npm        Node v18.17.0 ↔ npm v9.6.7     (in sync)
  ❌ rbenv (v1.2.0) ↔ gem      Ruby v3.1.0 ↔ gem v3.4.1       (out of sync)
     Expected gem version: v3.3.7 (bundled with Ruby 3.1.0)
     Action needed: downgrade gem or upgrade Ruby

  ✅ pyenv (v2.3.9) ↔ pip      Python v3.11.0 ↔ pip v22.3     (in sync)

  ❌ asdf (v0.13.1) ↔ nodejs   Node v16.20.0 ↔ npm v8.19.4    (out of sync)
     Action needed: upgrade Node to v18+ or downgrade npm

Summary: 2/4 pairs synchronized, 2 need attention

Synchronization strategies:
  --strategy vm_priority    Update package managers to match version managers
  --strategy pm_priority    Update version managers to match package managers
  --strategy latest         Update both to latest compatible versions

Fix synchronization issues? [y/N]:
```

## 🧪 테스트 요구사항

### 1. 단위 테스트
```go
func TestVersionSynchronizer_CheckSync(t *testing.T) {
    // 각 동기화 쌍의 상태 확인 테스트
}

func TestPolicyEngine_ApplyPolicy(t *testing.T) {
    // 동기화 정책 적용 테스트
}

func TestVersionResolver_GetExpectedVersion(t *testing.T) {
    // 예상 버전 계산 테스트
}
```

### 2. 통합 테스트
```bash
# 실제 환경에서 동기화 테스트
go test ./internal/pm/sync -tags=integration
```

### 3. 시나리오 테스트
- [ ] 다중 Node.js 버전 설치 환경
- [ ] gem 수동 업그레이드 후 불일치 상황
- [ ] pip 가상환경과 시스템 pip 충돌

## ✅ 완료 기준

### 기능 완성도
- [ ] 4개 주요 매니저 쌍 동기화 지원
- [ ] 정확한 버전 호환성 감지
- [ ] 다양한 동기화 전략 구현
- [ ] 안전한 백업/복원 메커니즘

### 사용자 경험
- [ ] 명확한 동기화 상태 표시
- [ ] 권장 조치 방법 안내
- [ ] 동기화 과정 진행 상황 표시
- [ ] 문제 해결 가이드 제공

### 안정성
- [ ] 기존 환경 백업 보장
- [ ] 동기화 실패 시 롤백 가능
- [ ] 여러 버전 공존 환경 지원
- [ ] 가상환경과의 충돌 방지

## 🚀 커밋 메시지 가이드

```
feat(claude-opus): 패키지 매니저 버전 동기화 기능 구현

- nvm↔npm, rbenv↔gem, pyenv↔pip, asdf↔multi 동기화 지원
- 3가지 동기화 전략 구현 (vm_priority, pm_priority, latest)
- 버전 호환성 자동 감지 및 권장 조치 안내
- 안전한 백업/롤백 시스템 포함
- 다중 버전 환경 및 가상환경 지원

Closes: cmd/pm/advanced.go:106 "sync-versions command not yet implemented"

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 💡 구현 힌트

1. **버전 호환성 DB**: 각 언어별 버전 매트릭스 구축
2. **점진적 동기화**: 한 번에 모든 쌍이 아닌 단계별 동기화
3. **사용자 확인**: 중요한 버전 변경은 사용자 동의 필수
4. **가상환경 고려**: pyenv의 virtualenv, rbenv의 gemset 등 고려

## 🔗 관련 작업

이 작업은 다음과 연계됩니다:
- `02-implement-pm-bootstrap.md` - 설치된 매니저들 간 동기화
- `03-implement-pm-upgrade-managers.md` - 업그레이드 후 동기화 확인
- 기존 `status.go` - 현재 버전 정보 활용

## ⚠️ 주의사항

- 버전 동기화는 기존 환경을 변경할 수 있으므로 신중하게 처리
- 가상환경이나 프로젝트별 설정과 충돌하지 않도록 주의
- 버전 매니저별로 동작 방식이 다름에 주의
- 일부 패키지 매니저는 독립적으로 설치/업그레이드될 수 있음
