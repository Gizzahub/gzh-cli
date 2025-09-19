# AGENTS.md - pm (패키지 매니저 통합)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**pm**은 다양한 패키지 매니저(Homebrew, asdf, npm, pip 등)를 통합 관리하는 모듈입니다.

### 핵심 기능

- 다중 패키지 매니저 지원 (brew, apt, pip, npm, cargo 등)
- 설정 파일 기반 패키지 관리 (`~/.gzh/pm/`)
- 패키지 상태 모니터링 (status)
- 설치/업데이트/내보내기 (install/update/export)
- 캐시 관리 (cache)

## ⚠️ 개발 시 주의사항

### 1. 패키지 매니저별 차이점 처리

```go
// ✅ 패키지 매니저별 추상화
type PackageManager interface {
    Install(packages []string) error
    Update(packages []string) error
    List() ([]Package, error)
    IsInstalled(pkg string) bool
}

// Homebrew 구현
type HomebrewManager struct{}
func (h *HomebrewManager) Install(packages []string) error {
    return exec.Command("brew", append([]string{"install"}, packages...)...).Run()
}

// npm 구현
type NPMManager struct{}
func (n *NPMManager) Install(packages []string) error {
    return exec.Command("npm", append([]string{"install", "-g"}, packages...)...).Run()
}
```

### 2. 설정 파일 검증

```go
// ✅ 설정 파일 유효성 검사
func (c *ConfigManager) ValidateConfig(configPath string) error {
    config, err := c.loadConfig(configPath)
    if err != nil {
        return err
    }

    // 패키지 매니저 존재 확인
    for manager := range config.Managers {
        if !c.isManagerAvailable(manager) {
            return fmt.Errorf("package manager %s not found", manager)
        }
    }

    return nil
}
```

### 3. 충돌 방지

```go
// ✅ 패키지 충돌 감지
func (p *PackageManager) DetectConflicts(packages []string) ([]Conflict, error) {
    conflicts := []Conflict{}

    for _, pkg := range packages {
        installedBy := p.getInstalledBy(pkg)
        if len(installedBy) > 1 {
            conflicts = append(conflicts, Conflict{
                Package:     pkg,
                Managers:    installedBy,
                Suggestion:  "Remove duplicate installations",
            })
        }
    }

    return conflicts, nil
}
```

## 🧪 테스트 요구사항

- **플랫폼별 테스트**: macOS, Linux, Windows 각각
- **패키지 매니저 조합**: 여러 매니저 동시 사용 시나리오
- **네트워크 장애**: 패키지 다운로드 실패 처리
- **권한 문제**: sudo 권한 필요한 설치 처리

**핵심**: 시스템 패키지를 직접 조작하므로 충돌 방지와 롤백 기능이 중요합니다.
