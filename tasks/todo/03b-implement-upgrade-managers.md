# TODO: 개별 패키지 매니저 업그레이더 구현

---
status: [ ]
priority: high
severity: medium
file_type: service_layer
estimated_effort: 45분
source: 03-implement-pm-upgrade-managers.md (분할됨)
depends_on: [03a-implement-upgrade-interfaces.md]
spec_reference: `cmd/pm/advanced.go:71`, `specs/package-manager.md`
---

## 📋 작업 개요

6개 패키지 매니저(brew, asdf, nvm, rbenv, pyenv, sdkman)에 대한 구체적인 업그레이드 로직을 구현합니다. 각각의 고유한 업그레이드 방식을 지원합니다.

## 🎯 구현 목표

### Step 1: Homebrew 업그레이더 구현
```go
type HomebrewUpgrader struct {
    logger logger.Logger
}

func (h *HomebrewUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
    // brew --version으로 현재 버전 확인
    // GitHub API로 최신 릴리즈 정보 확인
}

func (h *HomebrewUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
    // brew update && brew upgrade 실행
}
```

### Step 2: 버전 매니저 업그레이더 구현 (asdf, nvm, rbenv, pyenv, sdkman)
각 도구의 고유한 업데이트 메커니즘을 구현합니다.

- **asdf**: `asdf update` (Git pull 방식)
- **nvm**: 최신 설치 스크립트 다운로드 및 실행
- **rbenv**: macOS는 brew, Linux는 git pull
- **pyenv**: macOS는 brew, Linux는 pyenv-installer
- **sdkman**: 자체 업데이트 스크립트 실행

## 📁 파일 구조

### 생성할 파일
- `internal/pm/upgrade/homebrew.go` - Homebrew 업그레이드 로직
- `internal/pm/upgrade/asdf.go` - asdf 업그레이드 로직
- `internal/pm/upgrade/version_managers.go` - nvm, rbenv, pyenv, sdkman 업그레이드

## ✅ 완료 기준

- [ ] 6개 패키지 매니저 업그레이드 지원
- [ ] 버전 확인 및 비교 정확성

## 🚀 커밋 메시지

```
feat(claude-opus): 개별 패키지 매니저 업그레이더 구현

- Homebrew, asdf, nvm, rbenv, pyenv, sdkman 업그레이드 로직
- 플랫폼별 최적화된 업데이트 방식 지원
- 현재/최신 버전 확인 및 비교 기능

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```
