---
status: suspended
reason: 과밀 파일 (22개 체크박스 항목) - 논리적 단위로 분할됨
split_into:
  - 03a-implement-upgrade-interfaces.md
  - 03b-implement-upgrade-managers.md
  - 03c-implement-upgrade-commands.md
  - 03d-implement-upgrade-tests.md
---

# TODO: 패키지 매니저 업그레이드 기능 구현 (원본)

- status: [ ]
- priority: medium (P2)
- category: package-manager
- estimated_effort: 1시간
- depends_on: [02-implement-pm-bootstrap.md]
- spec_reference: `cmd/pm/advanced.go:71`, `specs/package-manager.md`

## 📋 작업 개요

패키지 매니저 도구들 자체의 버전을 업그레이드하는 기능을 구현합니다. 현재 "not yet implemented" 상태인 upgrade-managers 명령어를 완전히 구현하여 사용자가 최신 도구들을 유지할 수 있도록 합니다.

## 🎯 구현 목표

### 핵심 기능
- [ ] **업그레이드 가능 여부 확인** - 최신 버전 대비 현재 버전 체크
- [ ] **개별 매니저 업그레이드** - 특정 패키지 매니저만 선택적 업그레이드
- [ ] **일괄 업그레이드** - 모든 패키지 매니저 한번에 업그레이드
- [ ] **백업 및 롤백** - 업그레이드 실패 시 이전 버전으로 복원

### 지원할 업그레이드 방식
- [ ] **Self-update** - 도구 자체의 self-update 기능 활용
- [ ] **Package manager** - 상위 패키지 매니저를 통한 업그레이드
- [ ] **Manual download** - 직접 다운로드 및 설치

## 🔧 기술적 구현

### 1. 업그레이드 상태 구조체
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

### 2. 업그레이드 인터페이스
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

### 3. 개별 매니저 업그레이더 구현
```go
// Homebrew 업그레이드
type HomebrewUpgrader struct {
    logger logger.Logger
}

func (h *HomebrewUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
    // brew --version으로 현재 버전 확인
    currentVersion, err := h.getCurrentVersion(ctx)
    if err != nil {
        return nil, err
    }

    // GitHub API로 최신 릴리즈 정보 확인
    latestVersion, err := h.getLatestVersion(ctx)
    if err != nil {
        return nil, err
    }

    return &UpgradeStatus{
        Manager:          "brew",
        CurrentVersion:   currentVersion,
        LatestVersion:    latestVersion,
        UpdateAvailable:  h.compareVersions(currentVersion, latestVersion),
        UpdateMethod:     "self-update",
    }, nil
}

func (h *HomebrewUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
    // 백업 생성 (필요한 경우)
    if options.BackupEnabled {
        backupPath, err := h.Backup(ctx)
        if err != nil {
            return fmt.Errorf("backup failed: %w", err)
        }
        h.logger.Info("Backup created: %s", backupPath)
    }

    // brew update && brew upgrade
    cmd := exec.CommandContext(ctx, "brew", "update")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("brew update failed: %w", err)
    }

    // brew 자체 업그레이드는 update에 포함됨
    return nil
}

// asdf 업그레이드
type AsdfUpgrader struct {
    logger logger.Logger
}

func (a *AsdfUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
    // asdf update (Git pull)
    cmd := exec.CommandContext(ctx, "asdf", "update")
    return cmd.Run()
}

// nvm 업그레이드
type NvmUpgrader struct {
    logger logger.Logger
}

func (n *NvmUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
    // curl을 통한 최신 설치 스크립트 실행
    script := "curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"
    cmd := exec.CommandContext(ctx, "bash", "-c", script)
    return cmd.Run()
}
```

### 4. 버전 비교 시스템
```go
type VersionComparator struct{}

func (vc *VersionComparator) Compare(v1, v2 string) int {
    // Semantic versioning 비교
    // v1 < v2: -1, v1 == v2: 0, v1 > v2: 1
}

func (vc *VersionComparator) IsNewerVersion(current, latest string) bool {
    return vc.Compare(current, latest) < 0
}
```

## 📁 파일 구조

### 새로 생성할 파일
- `internal/pm/upgrade/manager.go` - 업그레이드 매니저 구현
- `internal/pm/upgrade/homebrew.go` - Homebrew 업그레이드 로직
- `internal/pm/upgrade/asdf.go` - asdf 업그레이드 로직
- `internal/pm/upgrade/version_managers.go` - nvm, rbenv, pyenv 업그레이드
- `internal/pm/upgrade/version_comparator.go` - 버전 비교 유틸리티
- `internal/pm/upgrade/backup.go` - 백업 및 롤백 로직

### 수정할 파일
- `cmd/pm/advanced.go` - upgrade-managers 명령어 실제 구현

## 🎯 명령어 구조

### 현재 명령어 확장
```bash
# 업그레이드 가능 여부 확인
gz pm upgrade-managers --check
gz pm upgrade-managers --check --json

# 모든 매니저 업그레이드
gz pm upgrade-managers --all

# 특정 매니저만 업그레이드
gz pm upgrade-managers --manager brew
gz pm upgrade-managers --manager asdf,nvm

# 백업과 함께 업그레이드
gz pm upgrade-managers --all --backup

# 강제 업그레이드 (버전 확인 무시)
gz pm upgrade-managers --all --force

# 프리릴리즈 포함
gz pm upgrade-managers --check --pre-release
```

### 출력 예시
```
🔄 Package Manager Upgrade Status

Checking for updates...

Available Updates:
  📦 brew      v4.1.14 → v4.2.0    (released 2 days ago)
  📦 asdf      v0.12.0 → v0.13.1   (released 1 week ago)
  📦 nvm       v0.39.0 → v0.39.2   (released 3 days ago)
  ✅ rbenv     v1.2.0 (up to date)
  ✅ pyenv     v2.3.9 (up to date)
  ❌ sdkman    (not installed)

Summary: 3 updates available, 2 up to date, 1 not installed

Estimated download size: 15.2 MB
Estimated time: 2-3 minutes

Continue with upgrades? [y/N]:
```

## 🧪 테스트 요구사항

### 1. 단위 테스트
```go
func TestVersionComparator(t *testing.T) {
    // 버전 비교 로직 테스트
}

func TestUpgradeManager_CheckUpdates(t *testing.T) {
    // 업데이트 확인 테스트 (API 모킹)
}

func TestBackupAndRollback(t *testing.T) {
    // 백업 및 롤백 테스트
}
```

### 2. 통합 테스트
```bash
# 실제 패키지 매니저와의 통합 테스트
go test ./internal/pm/upgrade -tags=integration
```

### 3. 시뮬레이션 테스트
- [ ] 업그레이드 실패 시나리오
- [ ] 네트워크 오류 처리
- [ ] 부분 업그레이드 완료 상황

## ✅ 완료 기준

### 기능 완성도
- [ ] 6개 패키지 매니저 업그레이드 지원
- [ ] 버전 확인 및 비교 정확성
- [ ] 백업/롤백 기능 안정성
- [ ] 에러 상황 적절한 처리

### 사용자 경험
- [ ] 업그레이드 진행 상황 시각화
- [ ] 예상 소요 시간 및 다운로드 크기 표시
- [ ] 업그레이드 후 변경사항 요약
- [ ] 실패 시 복구 방법 안내

### 안정성
- [ ] 중요 데이터 백업 보장
- [ ] 업그레이드 중 중단 시 복구 가능
- [ ] 호환되지 않는 버전 감지
- [ ] 의존성 충돌 방지

## 🚀 커밋 메시지 가이드

```
feat(claude-opus): 패키지 매니저 업그레이드 기능 구현

- 6개 패키지 매니저 자동 업그레이드 지원
- 버전 비교 및 최신 릴리즈 확인 기능
- 백업/롤백 시스템으로 안전한 업그레이드
- 개별 및 일괄 업그레이드 옵션 제공
- 진행 상황 시각화 및 예상 시간 표시

Closes: cmd/pm/advanced.go:71 "upgrade-managers command not yet implemented"

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 💡 구현 힌트

1. **안전 우선**: 모든 업그레이드는 백업과 함께 수행
2. **점진적 업그레이드**: 한 번에 모든 매니저가 아닌 단계적 업그레이드 옵션
3. **외부 API**: GitHub API 등을 활용한 최신 버전 정보 확인
4. **사용자 확인**: 중요한 업그레이드는 사용자 동의 필요

## 🔗 관련 작업

이 작업은 다음과 연계됩니다:
- `02-implement-pm-bootstrap.md` - 설치된 매니저들의 업그레이드
- Bootstrap에서 설치한 매니저들을 최신 상태로 유지

## ⚠️ 주의사항

- 업그레이드는 되돌릴 수 없는 작업일 수 있으므로 신중하게 처리
- 패키지 매니저별로 업그레이드 방식이 다름에 주의
- 업그레이드 후 기존 패키지들과의 호환성 확인 필요
- 네트워크 연결 및 충분한 디스크 공간 필요
