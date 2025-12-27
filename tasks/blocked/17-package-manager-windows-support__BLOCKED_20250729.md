# TODO: 패키지 매니저 Windows 지원 및 고급 기능 구현

- status: [>] BLOCKED → Split into sub-tasks
- priority: low
- category: package-manager
- estimated_effort: 4-5 days (original) → 10-12 days (detailed)
- depends_on: []
- spec_reference: `/specs/package-manager.md` lines 70-71, 325-327

## Sub-Tasks (2025-12-26 분할)

이 태스크는 아래 하위 태스크로 분할되었습니다:

| Task | Priority | Effort | Status |
|------|----------|--------|--------|
| [18-winget-support](../18-package-manager-winget-support.md) | medium | 2-3d | [ ] |
| [19-scoop-support](../19-package-manager-scoop-support.md) | medium | 2d | [ ] |
| [20-chocolatey-support](../20-package-manager-chocolatey-support.md) | low | 3d | [ ] |
| [21-cleanup-strategies](../21-package-manager-cleanup-strategies.md) | low | 3-4d | [ ] |

**권장 순서**: winget → Scoop → Chocolatey (의존성 순서)
**독립 태스크**: cleanup-strategies (언제든 진행 가능)

______________________________________________________________________

## 📋 작업 개요

패키지 매니저의 Windows 지원을 추가하고, 고급 클린업 전략 및 추가 기능을 구현하여 크로스 플랫폼 완전 지원을 달성합니다.

## 🎯 구현 목표

### Windows 패키지 매니저 지원

- [>] **Chocolatey** 패키지 매니저 지원 # 대규모 작업으로 인한 연기 - 핵심 TUI 기능 완료 후 별도 계획 필요
- [ ] **Scoop** 패키지 매니저 지원
- [ ] **winget** (Windows Package Manager) 지원
- [ ] Windows 전용 설정 및 경로 처리

### 고급 클린업 전략

- [ ] **Quarantine 모드** - 관리되지 않는 패키지를 격리
- [ ] **의존성 분석** - 사용하지 않는 의존성 정리
- [ ] **버전 정리** - 오래된 버전 정리
- [ ] **캐시 관리** - 패키지 캐시 최적화

### 추가 기능

- [ ] 패키지 보안 스캔
- [ ] 라이선스 호환성 체크
- [ ] 업데이트 일정 관리
- [ ] 패키지 사용량 분석

## 🔧 기술적 요구사항

### 1. Windows 패키지 매니저 구현

#### Chocolatey 지원

```bash
gz pm chocolatey install git
gz pm chocolatey list --local-only
gz pm chocolatey upgrade all
gz pm chocolatey uninstall git
```

```go
type ChocolateyManager struct {
    execPath    string
    configPath  string
    sources     []string
}

func (c *ChocolateyManager) Install(packages []string) error {
    // choco install 명령어 실행
    // 관리자 권한 확인
    // 설치 진행상황 추적
}

func (c *ChocolateyManager) ListInstalled() ([]Package, error) {
    // choco list --local-only 실행
    // XML 출력 파싱
    // 패키지 정보 구조체로 변환
}
```

#### Scoop 지원

```bash
gz pm scoop install git
gz pm scoop bucket add extras
gz pm scoop update *
gz pm scoop cleanup *
```

```go
type ScoopManager struct {
    scoopPath   string
    bucketsPath string
    appsPath    string
}

func (s *ScoopManager) AddBucket(bucket, repo string) error {
    // scoop bucket add 실행
    // Git 저장소 클론
    // Bucket 정보 업데이트
}
```

#### winget 지원

```bash
gz pm winget install Microsoft.PowerToys
gz pm winget search --name "Visual Studio Code"
gz pm winget upgrade --all
```

### 2. 고급 클린업 전략

#### Quarantine 모드 구현

```yaml
# ~/.gzh/pm/global.yml
cleanup:
  quarantine:
    enabled: true
    quarantine_dir: "~/.gzh/pm/quarantine"
    auto_quarantine: false
    retention_days: 30

  strategies:
    - name: "quarantine"
      description: "Move unmanaged packages to quarantine directory"
      destructive: false

    - name: "remove"
      description: "Remove unmanaged packages permanently"
      destructive: true
      confirmation_required: true
```

```go
type QuarantineManager struct {
    quarantineDir   string
    retentionDays   int
    metadata        map[string]QuarantineMetadata
}

type QuarantineMetadata struct {
    OriginalPath    string    `json:"original_path"`
    QuarantineTime  time.Time `json:"quarantine_time"`
    Reason          string    `json:"reason"`
    Manager         string    `json:"manager"`
    Size            int64     `json:"size"`
    Dependencies    []string  `json:"dependencies"`
}

func (qm *QuarantineManager) QuarantinePackage(pkg Package, reason string) error {
    // 패키지를 격리 디렉토리로 이동
    // 메타데이터 저장
    // 의존성 체크
    // 복구 스크립트 생성
}
```

#### 의존성 분석 시스템

```go
type DependencyAnalyzer struct {
    managers    []PackageManager
    depGraph    *DependencyGraph
    orphanPolicy OrphanPolicy
}

type DependencyGraph struct {
    nodes map[string]*PackageNode
    edges map[string][]string
}

type PackageNode struct {
    Name         string
    Version      string
    Manager      string
    InstallTime  time.Time
    LastUsed     time.Time
    Dependencies []string
    Dependents   []string
    UserInstalled bool
}

func (da *DependencyAnalyzer) FindOrphans() ([]Package, error) {
    // 의존성 그래프 구축
    // 리프 노드 중 사용자가 직접 설치하지 않은 패키지 식별
    // 마지막 사용 시간 기반 필터링
}
```

### 3. 플랫폼별 설정 관리

#### Windows 전용 설정

```yaml
# ~/.gzh/pm/global.yml
platform_specific:
  windows:
    chocolatey:
      install_missing: true
      use_system_python: false
      proxy_settings: "inherit"

    scoop:
      global_installs: false
      enable_long_paths: true

    winget:
      source_priorities:
        - "winget"
        - "msstore"

  execution:
    require_admin: true
    uac_bypass: false
    execution_policy: "RemoteSigned"
```

#### 경로 및 권한 처리

```go
type WindowsPackageManager struct {
    requiresAdmin   bool
    executionPolicy string
    pathResolver    *WindowsPathResolver
}

type WindowsPathResolver struct {
    programFiles    string
    programFilesX86 string
    localAppData    string
    roamingAppData  string
}

func (wpm *WindowsPackageManager) CheckAdminRights() (bool, error) {
    // Windows API 호출로 관리자 권한 확인
    // UAC 상태 확인
}

func (wpm *WindowsPackageManager) ElevateIfNeeded() error {
    // 필요 시 관리자 권한으로 재실행
    // UAC 프롬프트 처리
}
```

### 4. 보안 및 라이선스 기능

#### 패키지 보안 스캔

```go
type SecurityScanner struct {
    vulnerabilityDB VulnerabilityDB
    scanners        []PackageScanner
}

type VulnerabilityScan struct {
    Package         Package           `json:"package"`
    Vulnerabilities []Vulnerability   `json:"vulnerabilities"`
    RiskLevel       RiskLevel         `json:"risk_level"`
    Recommendations []string          `json:"recommendations"`
}

func (ss *SecurityScanner) ScanPackage(pkg Package) (*VulnerabilityScan, error) {
    // 알려진 취약점 DB와 대조
    // 패키지 서명 확인
    // 의심스러운 권한 체크
}
```

#### 라이선스 호환성 체크

```go
type LicenseChecker struct {
    compatibilityMatrix map[string][]string
    projectLicense      string
}

func (lc *LicenseChecker) CheckCompatibility(packages []Package) (*LicenseReport, error) {
    // 각 패키지의 라이선스 정보 수집
    // 프로젝트 라이선스와 호환성 확인
    // 충돌하는 라이선스 리포트
}
```

## 📁 파일 구조

### 새로 생성할 파일

- `cmd/pm/chocolatey.go` - Chocolatey 패키지 매니저 명령어
- `cmd/pm/scoop.go` - Scoop 패키지 매니저 명령어
- `cmd/pm/winget.go` - winget 패키지 매니저 명령어
- `internal/pm/windows/` - Windows 전용 패키지 매니저 구현
- `internal/pm/cleanup/quarantine.go` - Quarantine 관리
- `internal/pm/analysis/dependency.go` - 의존성 분석
- `internal/pm/security/scanner.go` - 보안 스캔
- `internal/pm/license/checker.go` - 라이선스 체크
- `pkg/pm/windows/` - Windows 패키지 매니저 공용 라이브러리

### 수정할 파일

- `cmd/pm/pm.go` - Windows 패키지 매니저 명령어 추가
- `cmd/pm/clean.go` - 고급 클린업 전략 추가
- `internal/pm/config/global.go` - Windows 설정 지원

## 🧪 테스트 요구사항

### Windows 환경 테스트

- [ ] Windows 10/11 환경에서 패키지 매니저 테스트
- [ ] 관리자 권한 필요한 작업 테스트
- [ ] UAC 상호작용 테스트

### 클린업 전략 테스트

- [ ] Quarantine 모드 동작 테스트
- [ ] 의존성 분석 정확성 테스트
- [ ] 복구 기능 테스트

### 크로스 플랫폼 테스트

- [ ] Linux, macOS, Windows 동일 설정 파일 호환성
- [ ] 플랫폼별 설정 오버라이드 테스트

## 📊 완료 기준

### 기능 완성도

- [ ] 3개 Windows 패키지 매니저 완전 지원
- [ ] 모든 고급 클린업 전략 구현
- [ ] 보안 및 라이선스 체크 기능

### Windows 지원

- [ ] Windows 10/11 완전 호환
- [ ] PowerShell/CMD 양쪽 지원
- [ ] UAC 및 관리자 권한 적절한 처리

### 사용자 경험

- [ ] 플랫폼 간 일관된 명령어 구조
- [ ] Windows 사용자를 위한 명확한 가이드
- [ ] 에러 상황에서 도움말 제공

## 🔗 관련 작업

이 작업은 기존 패키지 매니저 기능을 확장하므로 독립적으로 진행 가능합니다.

## 💡 구현 힌트

1. **점진적 구현**: 먼저 Chocolatey만 구현하고 순차적으로 확장
1. **관리자 권한 처리**: 필요할 때만 권한 상승 요청
1. **에러 처리**: Windows 특유의 에러 상황 고려
1. **성능 최적화**: Windows에서 느릴 수 있는 명령어 실행 최적화

## ⚠️ 주의사항

- Windows Defender 및 안티바이러스 소프트웨어와의 충돌 가능성
- UAC 설정에 따른 동작 차이
- Windows 업데이트 시 패키지 매니저 동작 변경 가능성
- 32bit/64bit 아키텍처 고려
- Windows 경로 길이 제한 및 특수 문자 처리

## 📋 Windows 패키지 매니저 비교

| 기능 | Chocolatey | Scoop | winget |
| ----------- | ---------- | -------- | ------ |
| 관리자 권한 | 필요 | 불필요 | 선택적 |
| GUI 앱 | 지원 | 제한적 | 지원 |
| 시스템 도구 | 지원 | 지원 | 지원 |
| 포터블 앱 | 제한적 | 특화 | 제한적 |
| 개발 도구 | 완전지원 | 완전지원 | 지원 |

이 정보를 바탕으로 각 패키지 매니저의 특성에 맞는 구현을 진행해야 합니다.
