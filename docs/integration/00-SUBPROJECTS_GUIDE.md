# 하위 프로젝트 통합 가이드

**작성일**: 2025-12-01
**목적**: gzh-cli 하위 라이브러리 사용 가이드

---

## 개요

gzh-cli는 **Integration Libraries Pattern**을 사용하여 핵심 기능을 독립 라이브러리로 분리합니다. 이를 통해:

- ✅ **단일 정보 소스** (Single Source of Truth) 확립
- ✅ **독립 사용 가능**: 각 라이브러리를 단독으로 사용 가능
- ✅ **코드 중복 제거**: 92% 코드 감소 (6,702줄)
- ✅ **유지보수 간소화**: 버그 수정과 기능 추가를 한 곳에서 관리

---

## 통합된 하위 프로젝트

### 1. gzh-cli-git

**목적**: 로컬 Git 리포지토리 작업 관리

**독립 사용**: ✅ 가능

#### 설치

```bash
# 독립 설치
go install github.com/gizzahub/gzh-cli-git/cmd/gzh-git@latest

# gzh-cli 통합 (이미 포함됨)
go install github.com/Gizzahub/gzh-cli/cmd/gz@latest
```

#### 주요 기능

- **스마트 클론/업데이트**: 6가지 전략 (rebase, reset, clone, skip, pull, fetch)
- **재귀적 일괄 업데이트**: 하위 디렉토리의 모든 Git 리포지토리 일괄 관리
- **안전 검증**: 충돌 감지, dry-run, 자동 백업

#### 명령어 비교

| 기능 | 독립 실행 | gzh-cli 통합 |
|-----|---------|-------------|
| 스마트 클론 | `gzh-git clone https://github.com/user/repo.git` | `gz git repo clone-or-update https://github.com/user/repo.git` |
| 일괄 업데이트 | `gzh-git pull-all ~/workspace` | `gz git repo pull-all ~/workspace` |
| 설정 파일 | `git-config.yaml` | `gzh.yaml` (통합 설정) |
| 인증 | `GIT_TOKEN` 환경 변수 | gzh-cli 토큰 공유 |

#### 사용 예제

**독립 사용**:
```bash
# 환경 변수 설정
export GIT_TOKEN="your_token"

# 스마트 클론
gzh-git clone https://github.com/user/repo.git --strategy rebase

# 일괄 업데이트
gzh-git pull-all ~/workspace --parallel 10
```

**gzh-cli 통합**:
```bash
# gzh.yaml 설정 사용
gz git repo clone-or-update https://github.com/user/repo.git --strategy rebase

# 통합 설정으로 일괄 업데이트
gz git repo pull-all ~/workspace --parallel 10
```

#### 문서

- **프로젝트**: [gzh-cli-git](https://github.com/gizzahub/gzh-cli-git)
- **README**: [gzh-cli-git README](https://github.com/gizzahub/gzh-cli-git#readme)
- **gzh-cli 통합 문서**: [Git Repository Management](../30-features/31-repository-management.md)

---

### 2. gzh-cli-quality

**목적**: 다중 언어 코드 품질 도구 통합

**독립 사용**: ✅ 가능

#### 설치

```bash
# 독립 설치
go install github.com/Gizzahub/gzh-cli-quality/cmd/gzh-quality@latest

# gzh-cli 통합 (이미 포함됨)
go install github.com/Gizzahub/gzh-cli/cmd/gz@latest
```

#### 주요 기능

- **다중 언어 지원**: Go, Python, JavaScript/TypeScript, Rust, Java, C/C++
- **린팅 + 포매팅**: 통합 실행 및 관리
- **CI/CD 통합**: JSON, JUnit XML 출력 지원
- **도구 관리**: 자동 설치, 버전 관리

#### 명령어 비교

| 기능 | 독립 실행 | gzh-cli 통합 |
|-----|---------|-------------|
| 전체 품질 검사 | `gzh-quality run` | `gz quality run` |
| 린팅만 | `gzh-quality check` | `gz quality check` |
| 변경 파일만 | `gzh-quality run --changed` | `gz quality run --changed` |
| 설정 파일 | `quality.yaml` | `gzh.yaml` (통합 설정) |

#### 지원 도구

| 언어 | 포매터 | 린터 |
|-----|--------|------|
| Go | gofumpt, gci | golangci-lint |
| Python | black, ruff | ruff, mypy, flake8 |
| JavaScript/TypeScript | prettier, dprint | eslint |
| Rust | rustfmt | clippy |
| Java | google-java-format | checkstyle, spotbugs |
| C/C++ | clang-format | clang-tidy |

#### 사용 예제

**독립 사용**:
```bash
# 전체 품질 검사
gzh-quality run

# 변경된 파일만
gzh-quality run --changed

# 프로젝트 분석
gzh-quality analyze
```

**gzh-cli 통합**:
```bash
# 통합 설정으로 실행
gz quality run

# gzh.yaml 설정 참조
gz quality run --changed
```

#### 문서

- **프로젝트**: [gzh-cli-quality](https://github.com/Gizzahub/gzh-cli-quality)
- **README**: [gzh-cli-quality README](https://github.com/Gizzahub/gzh-cli-quality#readme)
- **gzh-cli 통합 문서**: [Code Quality Management](../30-features/36-quality-management.md)

---

### 3. gzh-cli-package-manager

**목적**: 다중 패키지 매니저 통합 관리

**독립 사용**: ✅ 가능

#### 설치

```bash
# 독립 설치
go install github.com/gizzahub/gzh-cli-package-manager/cmd/gzh-pm@latest

# gzh-cli 통합 (이미 포함됨)
go install github.com/Gizzahub/gzh-cli/cmd/gz@latest
```

#### 주요 기능

- **다중 매니저 지원**: asdf, Homebrew, SDKMAN, npm, pip, cargo, go modules
- **일괄 업데이트**: 모든 패키지 매니저 동시 업데이트
- **선택적 업데이트**: 특정 도구만 업데이트
- **상태 확인**: 설치된 도구 및 버전 확인

#### 명령어 비교

| 기능 | 독립 실행 | gzh-cli 통합 |
|-----|---------|-------------|
| 전체 업데이트 | `gzh-pm update` | `gz pm update` |
| 특정 매니저 | `gzh-pm update --manager homebrew` | `gz pm update --manager homebrew` |
| 상태 확인 | `gzh-pm status` | `gz pm status` |
| 설정 파일 | `pm-config.yaml` | `gzh.yaml` (통합 설정) |

#### 지원 패키지 매니저

| 카테고리 | 매니저 |
|---------|--------|
| 언어 버전 관리 | asdf, nvm, pyenv, rbenv, rustup |
| 시스템 패키지 | Homebrew (macOS), apt (Ubuntu), yum (CentOS) |
| 언어별 | npm, pip, cargo, go modules |
| 개발 도구 | SDKMAN, kubectl, helm |

#### 사용 예제

**독립 사용**:
```bash
# 전체 업데이트
gzh-pm update

# 특정 매니저만
gzh-pm update --manager homebrew

# 상태 확인
gzh-pm status
```

**gzh-cli 통합**:
```bash
# 통합 설정으로 업데이트
gz pm update

# gzh.yaml 설정 참조
gz pm update --manager homebrew
```

#### 문서

- **프로젝트**: [gzh-cli-package-manager](https://github.com/gizzahub/gzh-cli-package-manager)
- **README**: [gzh-cli-package-manager README](https://github.com/gizzahub/gzh-cli-package-manager#readme)

---

### 4. gzh-cli-shellforge

**목적**: 모듈형 쉘 설정 빌더

**독립 사용**: ✅ 가능

#### 설치

```bash
# 독립 설치
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest

# gzh-cli 통합 (이미 포함됨)
go install github.com/Gizzahub/gzh-cli/cmd/gz@latest
```

#### 주요 기능

- **모듈화**: 모놀리식 `.zshrc`/`.bashrc`를 모듈로 분리
- **의존성 관리**: 위상 정렬로 자동 해결
- **OS별 필터링**: macOS/Linux 플랫폼별 선택적 포함
- **백업/복원**: Git 기반 버전 관리
- **템플릿**: 6가지 내장 템플릿 제공

#### 명령어 비교

| 기능 | 독립 실행 | gzh-cli 통합 |
|-----|---------|-------------|
| 빌드 | `shellforge build --manifest manifest.yaml` | `gz shellforge build --manifest manifest.yaml` |
| 검증 | `shellforge validate --manifest manifest.yaml` | `gz shellforge validate --manifest manifest.yaml` |
| 백업 | `shellforge backup --file ~/.zshrc` | `gz shellforge backup --file ~/.zshrc` |
| 템플릿 생성 | `shellforge template generate --type path` | `gz shellforge template generate --type path` |

#### 사용 예제

**독립 사용**:
```bash
# 설정 빌드
shellforge build --manifest manifest.yaml --output ~/.zshrc

# 검증
shellforge validate --manifest manifest.yaml

# 백업
shellforge backup --file ~/.zshrc --backup-dir ~/.shellforge/backups
```

**gzh-cli 통합**:
```bash
# 통합 명령어
gz shellforge build --manifest manifest.yaml --output ~/.zshrc
gz shellforge validate --manifest manifest.yaml
gz shellforge backup --file ~/.zshrc
```

#### 문서

- **프로젝트**: [gzh-cli-shellforge](https://github.com/gizzahub/gzh-cli-shellforge)
- **README**: [gzh-cli-shellforge README](https://github.com/gizzahub/gzh-cli-shellforge#readme)

---

## 통합 아키텍처

### Wrapper Pattern

gzh-cli는 **얇은 래퍼(Thin Wrapper)**를 통해 하위 라이브러리를 통합합니다.

#### 래퍼 구조

```go
// Example: cmd/quality_wrapper.go (45줄)
package cmd

import (
    "github.com/Gizzahub/gzh-cli-quality/cmd"
    "github.com/spf13/cobra"
)

func NewQualityCmd(appCtx *app.AppContext) *cobra.Command {
    // 하위 라이브러리에 위임
    return quality.NewRootCmd()
}

// Registry pattern 지원
func RegisterQualityCmd(appCtx *app.AppContext) {
    registry.Register(qualityCmdProvider{appCtx: appCtx})
}
```

#### 코드 감소 효과

| 라이브러리 | 래퍼 크기 | 원래 코드 | 감소량 | 감소율 |
|-----------|---------|---------|--------|--------|
| gzh-cli-quality | 45줄 | 3,514줄 | 3,469줄 | 98.7% |
| gzh-cli-package-manager | 65줄 | 2,453줄 | 2,388줄 | 97.3% |
| gzh-cli-git | 473줄 | 1,318줄 | 845줄 | 64.2% |
| gzh-cli-shellforge | 71줄 | - | - | (신규) |
| **전체** | **654줄** | **7,285줄** | **6,702줄** | **92.0%** |

### 개발 환경

#### 로컬 개발

```bash
# go.mod에 replace 지시문 사용
# go.mod:
replace github.com/gizzahub/gzh-cli-git => ../gzh-cli-git
replace github.com/gizzahub/gzh-cli-package-manager => ../gzh-cli-package-manager
replace github.com/gizzahub/gzh-cli-shellforge => ../gzh-cli-shellforge

# 빌드 (로컬 하위 프로젝트 자동 참조)
cd gzh-cli
make build
```

#### 의존성 구조

```
gzh-cli
├── go.mod (의존성 선언)
│   ├── github.com/gizzahub/gzh-cli-git
│   ├── github.com/Gizzahub/gzh-cli-quality
│   ├── github.com/gizzahub/gzh-cli-package-manager
│   └── github.com/gizzahub/gzh-cli-shellforge
└── cmd/
    ├── quality_wrapper.go (45줄) → gzh-cli-quality
    ├── pm_wrapper.go (65줄) → gzh-cli-package-manager
    ├── shellforge_wrapper.go (71줄) → gzh-cli-shellforge
    └── git/repo/
        ├── repo_clone_or_update_wrapper.go → gzh-cli-git
        └── repo_bulk_update_wrapper.go → gzh-cli-git
```

---

## FAQ

### Q: 하위 프로젝트를 독립적으로 사용할 수 있나요?

**A**: 네, 모든 하위 프로젝트는 독립 실행 가능합니다. 각 프로젝트는 자체 CLI 바이너리를 제공합니다.

```bash
# 독립 설치 예시
go install github.com/gizzahub/gzh-cli-git/cmd/gzh-git@latest
go install github.com/Gizzahub/gzh-cli-quality/cmd/gzh-quality@latest
```

### Q: gzh-cli 없이 특정 기능만 사용하면 되나요?

**A**: 가능합니다. 예를 들어 Git 기능만 필요하면 `gzh-cli-git`만 설치하면 됩니다.

**차이점**:
- **독립 사용**: 단일 기능, 개별 설정 파일
- **gzh-cli 통합**: 통합 설정, 다중 플랫폼 API, 추가 기능

### Q: 하위 프로젝트 버전은 어떻게 관리되나요?

**A**: 각 프로젝트는 독립적인 버전을 가지며, gzh-cli의 `go.mod`에서 의존성 버전을 명시합니다.

```go
// go.mod
require (
    github.com/Gizzahub/gzh-cli-quality v0.1.2
    github.com/gizzahub/gzh-cli-git v0.0.0-...
)
```

### Q: 하위 프로젝트에 기여하려면 어떻게 하나요?

**A**: 각 하위 프로젝트 리포지토리에서 직접 기여할 수 있습니다. 기여 후 gzh-cli에서 버전을 업데이트하세요.

```bash
# 1. 하위 프로젝트에서 작업
cd gzh-cli-git
git commit -m "feat: add new feature"
git push

# 2. gzh-cli에서 의존성 업데이트
cd gzh-cli
go get github.com/gizzahub/gzh-cli-git@latest
go mod tidy
```

### Q: 왜 Integration Libraries Pattern을 사용하나요?

**A**: 코드 중복을 제거하고 단일 정보 소스를 확립하기 위해서입니다.

**이점**:
- ✅ 버그 수정 한 번에 모든 곳 적용
- ✅ 기능 추가 중복 작업 제거
- ✅ 테스트 및 검증 한 곳에서 관리
- ✅ 독립 사용 및 통합 사용 모두 지원

### Q: 기존 코드는 어떻게 되나요?

**A**: 하위 프로젝트로 마이그레이션되었으며, gzh-cli에는 얇은 래퍼만 남아있습니다.

**Before (통합 전)**:
```
gzh-cli/cmd/quality/*.go (3,514줄)
```

**After (통합 후)**:
```
gzh-cli/cmd/quality_wrapper.go (45줄) → gzh-cli-quality 라이브러리
```

---

## 다음 단계

### 사용자

1. **gzh-cli 설치**: [Installation Guide](../10-getting-started/10-installation.md)
2. **기능 탐색**: [Features Overview](../30-features/)
3. **설정**: [Configuration Guide](../40-configuration/40-configuration-guide.md)

### 개발자

1. **아키텍처 이해**: [Integration Architecture](./ARCHITECTURE.md)
2. **로컬 개발 설정**: [Development Guide](../60-development/60-index.md)
3. **기여 가이드**: [Contributing](../CONTRIBUTING.md)

---

**마지막 업데이트**: 2025-12-01
**통합 완료**: Phase 1-3 (git, quality, package-manager, shellforge)
**코드 감소**: 6,702줄 (92.0% 감소율)
