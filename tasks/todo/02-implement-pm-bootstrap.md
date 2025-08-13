# TODO: 패키지 매니저 Bootstrap 기능 구현

- status: [ ]
- priority: high (P1)
- category: package-manager
- estimated_effort: 1.5시간
- depends_on: [01-fix-lint-issues.md]
- spec_reference: `cmd/pm/advanced.go:40`, `specs/package-manager.md`

## 📋 작업 개요

패키지 매니저들의 자동 설치 및 구성 기능을 구현하여 사용자가 개발 환경을 손쉽게 설정할 수 있도록 합니다. 현재 "not yet implemented" 상태인 bootstrap 명령어를 완전히 구현합니다.

## 🎯 구현 목표

### 핵심 기능
- [ ] **설치 상태 체크** - 어떤 패키지 매니저가 설치되어 있는지 확인
- [ ] **자동 설치** - 누락된 패키지 매니저 자동 설치
- [ ] **구성 설정** - 설치 후 기본 설정 적용
- [ ] **의존성 해결** - 패키지 매니저 간 의존성 관리

### 지원할 패키지 매니저
- [ ] **brew** (macOS) - Homebrew 설치
- [ ] **asdf** - 범용 버전 매니저
- [ ] **nvm** - Node.js 버전 매니저
- [ ] **rbenv** - Ruby 버전 매니저
- [ ] **pyenv** - Python 버전 매니저
- [ ] **sdkman** - JVM 관련 도구 매니저

## 🔧 기술적 구현

### 1. Bootstrap 상태 체크 구조체
```go
type BootstrapStatus struct {
    Manager     string          `json:"manager"`
    Installed   bool            `json:"installed"`
    Version     string          `json:"version,omitempty"`
    ConfigPath  string          `json:"config_path,omitempty"`
    Issues      []string        `json:"issues,omitempty"`
    Dependencies []string       `json:"dependencies,omitempty"`
}

type BootstrapReport struct {
    Platform    string            `json:"platform"`
    Summary     BootstrapSummary  `json:"summary"`
    Managers    []BootstrapStatus `json:"managers"`
    Timestamp   time.Time         `json:"timestamp"`
}

type BootstrapSummary struct {
    Total       int `json:"total"`
    Installed   int `json:"installed"`
    Missing     int `json:"missing"`
    Configured  int `json:"configured"`
}
```

### 2. Bootstrap 인터페이스
```go
type PackageManagerBootstrapper interface {
    CheckInstallation(ctx context.Context) (*BootstrapStatus, error)
    Install(ctx context.Context, force bool) error
    Configure(ctx context.Context) error
    GetDependencies() []string
    GetInstallScript() (string, error)
    Validate(ctx context.Context) error
}

type BootstrapManager struct {
    platform      string
    bootstrappers map[string]PackageManagerBootstrapper
    logger        logger.Logger
}
```

### 3. 플랫폼별 설치 로직
```go
type HomebrewBootstrapper struct {
    platform string
    logger   logger.Logger
}

func (h *HomebrewBootstrapper) Install(ctx context.Context, force bool) error {
    if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
        return fmt.Errorf("Homebrew not supported on %s", runtime.GOOS)
    }
    
    // macOS/Linux 설치 스크립트 실행
    script := `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`
    cmd := exec.CommandContext(ctx, "bash", "-c", script)
    
    return cmd.Run()
}

func (h *HomebrewBootstrapper) Configure(ctx context.Context) error {
    // PATH 설정, shell profile 업데이트
    return h.updateShellProfile()
}
```

### 4. 의존성 해결 시스템
```go
type DependencyResolver struct {
    graph map[string][]string
}

func (dr *DependencyResolver) ResolveDependencies(managers []string) ([]string, error) {
    // 의존성 순서에 따른 설치 순서 결정
    // 예: brew -> asdf -> nvm (brew가 asdf 설치에 필요할 수 있음)
}
```

## 📁 파일 구조

### 새로 생성할 파일
- `internal/pm/bootstrap/manager.go` - Bootstrap 매니저 구현
- `internal/pm/bootstrap/homebrew.go` - Homebrew 설치 로직
- `internal/pm/bootstrap/asdf.go` - asdf 설치 로직
- `internal/pm/bootstrap/version_managers.go` - nvm, rbenv, pyenv 설치
- `internal/pm/bootstrap/dependencies.go` - 의존성 해결 로직
- `internal/pm/bootstrap/shell_integration.go` - 쉘 프로파일 연동

### 수정할 파일
- `cmd/pm/advanced.go` - bootstrap 명령어 실제 구현
- `cmd/pm/pm.go` - 도움말 업데이트

## 🎯 명령어 구조

### 현재 명령어 확장
```bash
# 설치 상태 확인
gz pm bootstrap --check
gz pm bootstrap --check --json

# 모든 매니저 설치
gz pm bootstrap --install

# 특정 매니저들만 설치
gz pm bootstrap --install brew,asdf,nvm

# 강제 재설치
gz pm bootstrap --install --force

# 구성만 재설정 (설치 없이)
gz pm bootstrap --configure
```

### 출력 예시
```
📦 Package Manager Bootstrap Status

Platform: darwin (macOS 14.5)

Manager Status:
  ✅ brew      v4.1.14    /opt/homebrew/bin/brew
  ❌ asdf      missing    Will install via brew
  ✅ nvm       v0.39.0    ~/.nvm/nvm.sh
  ❌ rbenv     missing    Will install via brew
  ❌ pyenv     missing    Will install via brew
  ❌ sdkman    missing    Will install via curl

Summary: 2/6 installed, 4 missing

Recommended installation order:
  1. asdf (depends on: brew)
  2. rbenv (depends on: brew)  
  3. pyenv (depends on: brew)
  4. sdkman (independent)
```

## 🧪 테스트 요구사항

### 1. 단위 테스트
```go
func TestBootstrapManager_CheckStatus(t *testing.T) {
    // 각 플랫폼별 상태 체크 테스트
}

func TestDependencyResolver(t *testing.T) {
    // 의존성 해결 로직 테스트
}
```

### 2. 통합 테스트
```bash
# Docker 환경에서 전체 설치 과정 테스트
go test ./internal/pm/bootstrap -tags=integration
```

### 3. 플랫폼별 테스트
- [ ] macOS - Homebrew, asdf, nvm 등
- [ ] Linux - 패키지 매니저별 설치 확인
- [ ] Windows - 미지원 플랫폼 에러 처리

## ✅ 완료 기준

### 기능 완성도
- [ ] 6개 이상 패키지 매니저 지원
- [ ] 플랫폼별 적절한 설치 방법 구현
- [ ] 의존성 자동 해결
- [ ] 설치 후 자동 구성

### 사용자 경험
- [ ] 명확한 진행 상황 표시
- [ ] 에러 발생 시 복구 방법 안내
- [ ] JSON 출력으로 자동화 지원
- [ ] 설치 시간 예상치 제공

### 안정성
- [ ] 부분 실패 시 rollback 지원
- [ ] 중단된 설치 재개 가능
- [ ] 네트워크 오류 처리
- [ ] 권한 문제 해결 안내

## 🚀 커밋 메시지 가이드

```
feat(claude-opus): 패키지 매니저 bootstrap 기능 구현

- 6개 패키지 매니저 자동 설치 지원 (brew, asdf, nvm, rbenv, pyenv, sdkman)
- 플랫폼별 최적화된 설치 로직 구현
- 의존성 자동 해결 및 설치 순서 최적화
- 설치 상태 체크 및 JSON 출력 지원
- 부분 실패 시 복구 가이드 제공

Closes: cmd/pm/advanced.go:40 "bootstrap command not yet implemented"

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 💡 구현 힌트

1. **점진적 구현**: brew 먼저 구현 후 다른 매니저들 순차 추가
2. **에러 처리**: 네트워크, 권한, 플랫폼 호환성 에러 세심하게 처리
3. **진행 상황**: 설치 과정의 각 단계를 사용자에게 표시
4. **테스트**: Docker 환경에서 전체 설치 과정 자동화 테스트

## 🔗 관련 작업

이 작업이 완료되면 다음 작업들과 연계됩니다:
- `03-implement-pm-upgrade-managers.md` - 설치된 매니저들의 업그레이드
- `04-implement-pm-sync-versions.md` - 버전 동기화
- 기존 `cache.go` - 설치된 매니저들의 캐시 관리

## ⚠️ 주의사항

- 시스템에 변경사항을 가하므로 사용자 동의 필수
- 네트워크 연결 필요한 작업임을 명시
- 기존 설치와의 충돌 방지
- sudo 권한이 필요한 경우 명확한 안내