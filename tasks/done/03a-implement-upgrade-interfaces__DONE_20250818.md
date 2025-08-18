# TODO: 업그레이드 인터페이스 및 핵심 타입 정의

---
status: [x] COMPLETED
priority: high
severity: medium
file_type: service_layer
estimated_effort: 30분
actual_effort: 25분
source: 03-implement-pm-upgrade-managers.md (분할됨)
depends_on: [02-implement-pm-bootstrap.md]
spec_reference: `cmd/pm/advanced.go:71`, `specs/package-manager.md`
completed_date: 2025-08-18
commit_hash: be49fdd
---

## 📋 작업 개요

패키지 매니저 업그레이드 시스템의 핵심 인터페이스와 데이터 구조를 정의합니다. 이는 후속 구현의 기반이 되는 중요한 아키텍처 작업입니다.

## 🎯 구현 목표

### Step 1: 핵심 데이터 구조 정의
업그레이드 상태와 보고서를 위한 구조체들을 정의합니다.

```go
type UpgradeStatus struct {
    Manager         string    `json:"manager"`
    CurrentVersion  string    `json:"current_version"`
    LatestVersion   string    `json:"latest_version"`
    UpdateAvailable bool     `json:"update_available"`
    UpdateMethod    string    `json:"update_method"`
    ReleaseDate     time.Time `json:"release_date,omitempty"`
    ChangelogURL    string    `json:"changelog_url,omitempty"`
    Size            int64     `json:"size,omitempty"`
}

type UpgradeReport struct {
    Platform      string          `json:"platform"`
    TotalManagers int             `json:"total_managers"`
    UpdatesNeeded int             `json:"updates_needed"`
    Managers      []UpgradeStatus `json:"managers"`
    Timestamp     time.Time       `json:"timestamp"`
}
```

### Step 2: 업그레이드 인터페이스 설계
모든 패키지 매니저 업그레이더가 구현해야 할 공통 인터페이스를 정의합니다.

```go
type PackageManagerUpgrader interface {
    CheckUpdate(ctx context.Context) (*UpgradeStatus, error)
    Upgrade(ctx context.Context, options UpgradeOptions) error
    Backup(ctx context.Context) (string, error)
    Rollback(ctx context.Context, backupPath string) error
    GetUpdateMethod() string
    ValidateUpgrade(ctx context.Context) error
}

type UpgradeOptions struct {
    Force           bool
    PreRelease      bool
    BackupEnabled   bool
    SkipValidation  bool
    Timeout         time.Duration
}

type UpgradeManager struct {
    upgraders map[string]PackageManagerUpgrader
    logger    logger.Logger
    backupDir string
}
```

## 📁 파일 구조

### 생성할 파일
- `internal/pm/upgrade/types.go` - 핵심 데이터 구조 및 인터페이스 정의

## ✅ 완료 기준

- [x] UpgradeStatus, UpgradeReport 구조체 완성
- [x] PackageManagerUpgrader 인터페이스 정의 완료
- [x] UpgradeOptions 및 UpgradeManager 타입 구현
- [x] internal/pm/upgrade/types.go 파일 생성

## 📝 실제 구현 내용

- `internal/pm/upgrade/types.go` 파일 생성 완료
- 모든 필수 인터페이스 및 구조체 정의
- logger.CommonLogger 인터페이스 사용으로 기존 로깅 시스템과 통합
- 향후 확장을 위한 유연한 구조 설계

## 🚀 커밋 메시지

```
feat(claude-opus): 패키지 매니저 업그레이드 인터페이스 정의

- 업그레이드 상태 및 보고서 구조체 정의
- PackageManagerUpgrader 공통 인터페이스 구현
- 백업/롤백 기능을 위한 옵션 구조 설계

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```
