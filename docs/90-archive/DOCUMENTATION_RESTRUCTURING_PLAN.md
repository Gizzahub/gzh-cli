# gzh-cli 문서 재구성 계획서

**작성일**: 2025-12-01
**상태**: 초안 (Draft)
**목적**: 하위 프로젝트 분리에 맞춘 문서 체계 개편 및 사용자/LLM 문서 분리

______________________________________________________________________

## 📋 목차

1. [현황 분석](#%ED%98%84%ED%99%A9-%EB%B6%84%EC%84%9D)
1. [문제점 식별](#%EB%AC%B8%EC%A0%9C%EC%A0%90-%EC%8B%9D%EB%B3%84)
1. [문서 분리 전략](#%EB%AC%B8%EC%84%9C-%EB%B6%84%EB%A6%AC-%EC%A0%84%EB%9E%B5)
1. [재구성 계획](#%EC%9E%AC%EA%B5%AC%EC%84%B1-%EA%B3%84%ED%9A%8D)
1. [하위 프로젝트 연동](#%ED%95%98%EC%9C%84-%ED%94%84%EB%A1%9C%EC%A0%9D%ED%8A%B8-%EC%97%B0%EB%8F%99)
1. [실행 로드맵](#%EC%8B%A4%ED%96%89-%EB%A1%9C%EB%93%9C%EB%A7%B5)

______________________________________________________________________

## 현황 분석

### 📊 현재 문서 구조

#### 1. 문서 통계

- **전체 Markdown 파일**: 107개
- **AGENTS.md 파일** (LLM용): 15개
- **문서 디렉토리**: 14개 카테고리

#### 2. 문서 분류

##### A. 사용자용 문서 (Human-Facing Docs)

```
docs/
├── 00-overview/              # 프로젝트 개요
├── 10-getting-started/       # 시작 가이드
├── 20-architecture/          # 아키텍처 설명
├── 30-features/              # 기능 설명 (14개 파일)
├── 40-configuration/         # 설정 가이드
├── 50-api-reference/         # 명령어 레퍼런스
├── 70-deployment/            # 배포 가이드
├── 80-integrations/          # 통합 가이드
├── 90-maintenance/           # 유지보수
└── 99-appendix/              # 부록
```

##### B. LLM용 문서 (AI Agent Docs)

```
cmd/
├── AGENTS_COMMON.md          # 공통 개발 가이드
├── actions-policy/AGENTS.md  # 모듈별 가이드
├── dev-env/AGENTS.md
├── doctor/AGENTS.md
├── git/AGENTS.md
├── ide/AGENTS.md
└── [13개 모듈 AGENTS.md]
```

##### C. 통합/개발 문서

```
docs/
├── 60-development/           # 개발 가이드
├── integration/              # 하위 프로젝트 통합 문서
└── testing/                  # 테스트 문서
```

##### D. 루트 문서

```
프로젝트 루트/
├── README.md                 # 메인 문서 (사용자용)
├── CLAUDE.md                 # AI Agent용 메타 가이드
├── TECH_STACK.md            # 기술 스택
└── CHANGELOG.md             # 변경 이력
```

### 🔗 하위 프로젝트 현황

#### 통합된 하위 라이브러리 (go.mod 기준)

```go
// 외부 라이브러리로 분리된 기능들
github.com/gizzahub/gzh-cli-quality          // 코드 품질 도구
github.com/gizzahub/gzh-cli-package-manager  // 패키지 매니저
github.com/gizzahub/gzh-cli-gitforge              // 로컬 Git 작업
github.com/gizzahub/gzh-cli-shellforge       // 쉘 설정 빌더

// 로컬 개발 replace 지시문 (../상위 디렉토리)
replace github.com/gizzahub/gzh-cli-package-manager => ../gzh-cli-package-manager
replace github.com/gizzahub/gzh-cli-gitforge => ../gzh-cli-gitforge
replace github.com/gizzahub/gzh-cli-shellforge => ../gzh-cli-shellforge
```

#### 통합 패턴

- **Wrapper Pattern**: 얇은 래퍼로 외부 라이브러리 통합
- **코드 감소**: 6,702줄 (92.0% 감소율)
- **문서 위치**: `docs/integration/`

______________________________________________________________________

## 문제점 식별

### ❌ 현재 문제점

#### 1. 하위 프로젝트 문서 통합 부족

- ✅ **있는 것**: `docs/integration/` 디렉토리 (통합 개요)
- ❌ **없는 것**:
  - 각 하위 프로젝트의 상세 사용법 링크
  - 하위 프로젝트 설치 가이드
  - 독립 실행 vs 통합 실행 차이점 설명

#### 2. 문서 중복 및 일관성 부족

- `README.md`: 30,604 bytes (1,032줄) - **너무 긴 단일 파일**
- `CLAUDE.md`: 17,769 bytes (475줄) - 개발자용이지만 사용자 정보 혼재
- 기능 설명이 README와 `docs/30-features/`에 중복

#### 3. LLM 문서 분산

- ✅ **장점**: 모듈별 AGENTS.md로 분산 (15개 파일)
- ❌ **단점**:
  - 하위 프로젝트 AGENTS.md는 해당 리포지토리에만 존재
  - gzh-cli에서 참조 방법 불명확

#### 4. 사용자가 기능을 파악하기 어려움

- 문제: "이 도구가 뭘 할 수 있나?" → README 1,000줄 읽어야 함
- 필요: **기능별 Quick Reference** + **상세 문서 링크**

#### 5. 문서 발견성 (Discoverability) 부족

```
사용자 질문: "gzh-cli로 IDE 관리가 가능한가?"
현재 답변 경로: README.md (246줄) → docs/30-features/35-ide-management.md

이상적 경로: README.md 목차 → IDE 관리 → 상세 링크
```

______________________________________________________________________

## 문서 분리 전략

### 🎯 원칙

#### 원칙 1: 역할 기반 분리

```
사용자용 문서 (Human-Facing)
├── 목적: 도구 사용법, 기능 설명, 문제 해결
├── 언어: 사용자 친화적 한국어/영어
├── 형식: 단계별 가이드, 예제, 스크린샷
└── 위치: docs/ + README.md

LLM용 문서 (AI Agent)
├── 목적: 코드 작성, 테스트, 리팩토링 가이드
├── 언어: 기술 중심 영어/한국어
├── 형식: 코딩 컨벤션, API 명세, 테스트 규칙
└── 위치: CLAUDE.md + cmd/*/AGENTS.md
```

#### 원칙 2: 단일 정보 소스 (Single Source of Truth)

- **기능 설명**: `docs/30-features/*.md` (사용자용)
- **개발 가이드**: `cmd/*/AGENTS.md` (LLM용)
- **통합 개요**: `docs/integration/` (하위 프로젝트)

#### 원칙 3: 계층적 정보 구조

```
Level 1: README.md (Overview + Quick Links)
         ↓
Level 2: docs/00-overview/00-index.md (Navigation Hub)
         ↓
Level 3: docs/XX-category/*.md (Detailed Guides)
         ↓
Level 4: 하위 프로젝트 README.md (External Links)
```

______________________________________________________________________

## 재구성 계획

### 📂 Phase 1: README.md 슬림화 (현재 1,032줄 → 목표 300줄)

#### AS-IS (현재)

```markdown
# README.md (1,032줄)
- 핵심 기능 개요 (150줄)
- 빠른 시작 (100줄)
- CLI 명령어 구조 (100줄)
- 각 기능 상세 설명 (500줄) ← 중복
- 설치 방법 (50줄)
- 설정 예제 (100줄)
- 성능 모니터링 (32줄)
```

#### TO-BE (목표)

```markdown
# README.md (300줄 이하)

## 개요
- 프로젝트 소개 (3-5줄)
- 배지 (badges)
- 핵심 가치 제안 (What/Why/How - 10줄)

## 빠른 시작 (Quick Start)
- 설치 (5줄 + 링크)
- 첫 명령어 3개 (10줄)
- 다음 단계 링크

## 주요 기능 (Features Overview)
### 표 형식 요약
| 기능 | 한 줄 설명 | 상세 문서 |
|-----|----------|---------|
| Git 통합 | 다중 플랫폼 Git 관리 | [📖](docs/30-features/31-repository-management.md) |
| IDE 관리 | JetBrains/VS Code 통합 | [📖](docs/30-features/35-ide-management.md) |
| 코드 품질 | 다중 언어 린팅/포매팅 | [📖](docs/30-features/36-quality-management.md) |
| ...     | ...                    | ...                                                |

## 하위 프로젝트 (Subprojects)
- gzh-cli-gitforge → [링크]
- gzh-cli-quality → [링크]
- gzh-cli-package-manager → [링크]
- gzh-cli-shellforge → [링크]

## 문서 (Documentation)
- [📚 전체 문서](docs/00-overview/00-index.md)
- [🚀 시작 가이드](docs/10-getting-started/10-installation.md)
- [⚙️ 설정](docs/40-configuration/40-configuration-guide.md)
- [📋 명령어 레퍼런스](docs/50-api-reference/50-command-reference.md)

## 개발 참여
- [기여 가이드](docs/CONTRIBUTING.md)
- [개발 환경 설정](docs/60-development/60-index.md)

## 라이선스
```

**예상 효과**: 1,032줄 → 300줄 (70% 감소)

______________________________________________________________________

### 📂 Phase 2: CLAUDE.md 재구성 (LLM 전용 최적화)

#### AS-IS (현재)

```markdown
# CLAUDE.md (475줄)
- Project Overview (8줄)
- Essential Commands (90줄) ← 사용자용 내용
- Makefile Structure (84줄) ← 개발 가이드
- Architecture (148줄) ← 일부 중복
- Configuration and Schema (20줄)
- Testing Guidelines (15줄)
- Important Notes (25줄)
- Command Categories (50줄)
- Repository Clone Strategies (15줄)
- Authentication (5줄)
- Common Issues (15줄)
```

#### TO-BE (목표)

````markdown
# CLAUDE.md (300줄 이하, LLM 최적화)

## Project Context (Quick Overview)
- Binary name: gz
- Architecture: Integration Libraries Pattern
- Go version: 1.23+
- Key principles: Interface-driven, modular commands

## Development Workflow (LLM Task Guide)
### Code Modification Workflow
1. Read module's AGENTS.md first
2. Check existing patterns
3. Write code + tests
4. Run quality checks: `make fmt && make lint && make test`
5. Commit

### Command Structure
- cmd/*/AGENTS.md → Module-specific rules
- internal/ → Private abstractions
- pkg/ → Public APIs

## Essential Commands Reference (Quick Lookup)
```bash
# Development
make bootstrap      # One-time setup
make build         # Build binary
make test          # Run tests
make lint          # Lint checks

# Module Testing
go test ./cmd/{module} -v
go test ./cmd/git -run "TestSpecific" -v
````

## Architecture Patterns (LLM Decision Guide)

### When to Use Each Pattern

- Interface abstraction: `internal/git/interfaces.go`
- Provider registry: `pkg/git/provider/`
- Strategy pattern: Git operations (rebase/reset/clone/pull/fetch)

### Integration Libraries (External Dependencies)

| Library | Wrapper Location | Purpose |
|---------|-----------------|---------|
| gzh-cli-gitforge | cmd/git/repo/\*\_wrapper.go | Local Git ops |
| gzh-cli-quality | cmd/quality_wrapper.go | Code quality |
| gzh-cli-package-manager | cmd/pm_wrapper.go | Package mgmt |
| gzh-cli-shellforge | cmd/shellforge_wrapper.go | Shell configs |

## Module-Specific Guides (Links)

- [Common Guidelines](cmd/AGENTS_COMMON.md)
- [git module](cmd/git/AGENTS.md)
- [ide module](cmd/ide/AGENTS.md)
- [13 other modules]

## Important Rules

- Always run `make fmt && make lint` before commit
- Korean comments for new code
- Check cmd/AGENTS_COMMON.md for conventions
- Test coverage: 80%+ for core logic
- No over-engineering (see AGENTS_COMMON.md)

````

**예상 효과**: 475줄 → 300줄 (37% 감소) + LLM 최적화

---

### 📂 Phase 3: 하위 프로젝트 통합 문서 신규 작성

#### 새 문서: `docs/integration/00-SUBPROJECTS_GUIDE.md`

```markdown
# 하위 프로젝트 통합 가이드

## 개요
gzh-cli는 Integration Libraries Pattern을 사용하여 핵심 기능을 독립 라이브러리로 분리합니다.

## 통합된 하위 프로젝트

### 1. gzh-cli-gitforge
**목적**: 로컬 Git 리포지토리 작업
**독립 사용**: ✅ 가능
**설치**:
```bash
go install github.com/gizzahub/gzh-cli-gitforge/cmd/gzh-git@latest
````

**문서**: [gzh-cli-gitforge README](https://github.com/gizzahub/gzh-cli-gitforge)

**gzh-cli 통합 명령어**:

- `gz git repo clone-or-update` → gzh-cli-gitforge 사용
- `gz git repo pull-all` → gzh-cli-gitforge 사용

**차이점**:
| 기능 | 독립 실행 | gzh-cli 통합 |
|-----|---------|-------------|
| 명령어 | `gzh-git clone` | `gz git repo clone-or-update` |
| 설정 파일 | `git-config.yaml` | `gzh.yaml` (통합 설정) |
| 인증 | 별도 토큰 | gzh-cli 토큰 공유 |

______________________________________________________________________

### 2. gzh-cli-quality

**목적**: 다중 언어 코드 품질 도구
**독립 사용**: ✅ 가능
**설치**:

```bash
go install github.com/gizzahub/gzh-cli-quality/cmd/gzh-quality@latest
```

**문서**: [gzh-cli-quality README](https://github.com/gizzahub/gzh-cli-quality)

**gzh-cli 통합 명령어**:

- `gz quality run` → gzh-cli-quality 사용
- `gz quality check` → gzh-cli-quality 사용

______________________________________________________________________

### 3. gzh-cli-package-manager

**목적**: 다중 패키지 매니저 통합
**독립 사용**: ✅ 가능
**설치**:

```bash
go install github.com/gizzahub/gzh-cli-package-manager/cmd/gzh-pm@latest
```

**문서**: [gzh-cli-package-manager README](https://github.com/gizzahub/gzh-cli-package-manager)

**gzh-cli 통합 명령어**:

- `gz pm update` → gzh-cli-package-manager 사용

______________________________________________________________________

### 4. gzh-cli-shellforge

**목적**: 모듈형 쉘 설정 빌더
**독립 사용**: ✅ 가능
**설치**:

```bash
go install github.com/gizzahub/gzh-cli-shellforge/cmd/shellforge@latest
```

**문서**: [gzh-cli-shellforge README](https://github.com/gizzahub/gzh-cli-shellforge)

**gzh-cli 통합 명령어**:

- `gz shellforge build` → gzh-cli-shellforge 사용
- `gz shellforge validate` → gzh-cli-shellforge 사용

______________________________________________________________________

## 통합 아키텍처

### Wrapper Pattern

gzh-cli는 얇은 래퍼(Thin Wrapper)를 통해 하위 라이브러리를 통합합니다.

```go
// Example: cmd/quality_wrapper.go (45줄)
func NewQualityCmd(appCtx *app.AppContext) *cobra.Command {
    return quality.NewRootCmd() // Delegate to external library
}
```

### 개발 환경

```bash
# 로컬 개발 시 replace 지시문 사용
# go.mod:
replace github.com/gizzahub/gzh-cli-gitforge => ../gzh-cli-gitforge

# 빌드
cd gzh-cli
make build  # 자동으로 로컬 하위 프로젝트 참조
```

## FAQ

**Q: 하위 프로젝트를 독립적으로 사용할 수 있나요?**
A: 네, 모든 하위 프로젝트는 독립 실행 가능합니다.

**Q: gzh-cli 없이 gzh-cli-gitforge만 설치하면 되나요?**
A: Git 기능만 필요하면 가능합니다. 하지만 gzh-cli는 통합 설정, 다중 플랫폼 API 등 추가 기능을 제공합니다.

**Q: 하위 프로젝트 버전은 어떻게 관리되나요?**
A: 각 프로젝트는 독립적인 버전을 가지며, gzh-cli의 go.mod에서 의존성 버전을 명시합니다.

```

---

### 📂 Phase 4: docs/ 구조 최적화

#### 현재 문제점
```

docs/30-features/
├── 30-synclone.md # synclone 기능 설명
├── 31-repository-management.md # Git repo 기능 설명
├── 36-quality-management.md # quality 기능 설명
└── ...

→ 하위 프로젝트 사용법은 어디에?

````

#### 제안: 기능 문서와 하위 프로젝트 링크 통합

**예시: docs/30-features/36-quality-management.md 수정**

```markdown
# 코드 품질 관리 (Code Quality Management)

> **Note**: 이 기능은 [gzh-cli-quality](https://github.com/gizzahub/gzh-cli-quality) 라이브러리를 통합하여 제공됩니다.
> - **독립 사용**: `gzh-quality` 명령어로 독립 실행 가능
> - **통합 사용**: `gz quality` 명령어로 gzh-cli에서 실행
> - **상세 문서**: [gzh-cli-quality README](https://github.com/gizzahub/gzh-cli-quality)

## 개요
다중 언어를 지원하는 통합 코드 품질 관리 도구입니다.

[... 기존 내용 유지 ...]

## 추가 정보

### 독립 실행 (Standalone Usage)
하위 프로젝트를 직접 사용하려면:
```bash
# 설치
go install github.com/gizzahub/gzh-cli-quality/cmd/gzh-quality@latest

# 실행
gzh-quality run
````

### 통합 실행 (Integrated Usage)

gzh-cli를 통해 사용:

```bash
gz quality run
```

**차이점**: 통합 실행 시 gzh-cli의 통합 설정(`gzh.yaml`)을 사용합니다.

### 더 알아보기

- [gzh-cli-quality 전체 문서](https://github.com/gizzahub/gzh-cli-quality)
- [통합 아키텍처 설명](../integration/00-SUBPROJECTS_GUIDE.md)

````

---

## 하위 프로젝트 연동

### 🔗 연동 방안

#### 1. README.md 하위 프로젝트 섹션 추가

```markdown
## 🧩 하위 프로젝트 (Subprojects)

gzh-cli는 핵심 기능을 독립 라이브러리로 분리하여 개발합니다. 각 라이브러리는 독립적으로 사용 가능합니다.

| 프로젝트 | 목적 | 독립 사용 | 문서 |
|---------|------|---------|------|
| [gzh-cli-gitforge][git-repo] | 로컬 Git 작업 관리 | ✅ | [📖][git-doc] |
| [gzh-cli-quality][quality-repo] | 코드 품질 도구 | ✅ | [📖][quality-doc] |
| [gzh-cli-package-manager][pm-repo] | 패키지 매니저 통합 | ✅ | [📖][pm-doc] |
| [gzh-cli-shellforge][shell-repo] | 쉘 설정 빌더 | ✅ | [📖][shell-doc] |

**통합 아키텍처**: [Integration Libraries Pattern](docs/integration/00-SUBPROJECTS_GUIDE.md)

[git-repo]: https://github.com/gizzahub/gzh-cli-gitforge
[git-doc]: https://github.com/gizzahub/gzh-cli-gitforge#readme
[quality-repo]: https://github.com/gizzahub/gzh-cli-quality
[quality-doc]: https://github.com/gizzahub/gzh-cli-quality#readme
[pm-repo]: https://github.com/gizzahub/gzh-cli-package-manager
[pm-doc]: https://github.com/gizzahub/gzh-cli-package-manager#readme
[shell-repo]: https://github.com/gizzahub/gzh-cli-shellforge
[shell-doc]: https://github.com/gizzahub/gzh-cli-shellforge#readme
````

#### 2. 기능 문서에 하위 프로젝트 링크 추가

**템플릿**: 각 통합 기능 문서 상단에 추가

```markdown
> **🔗 Powered by**: [gzh-cli-{name}](https://github.com/gizzahub/gzh-cli-{name})
> - **독립 설치**: `go install github.com/gizzahub/gzh-cli-{name}/cmd/gzh-{name}@latest`
> - **상세 문서**: [{name} README](https://github.com/gizzahub/gzh-cli-{name}#readme)
```

#### 3. CLAUDE.md에 하위 프로젝트 참조 추가

```markdown
## Integration Libraries (External Dependencies)

When modifying commands that use integration libraries:

1. **Check wrapper file first**: `cmd/{module}_wrapper.go` or `cmd/{module}/repo/*_wrapper.go`
2. **External library source**: Actual implementation is in the external repository
3. **Local development**: Use `replace` directives in go.mod to test changes

| Library | Repository | Local Path (dev) |
|---------|-----------|------------------|
| gzh-cli-gitforge | [GitHub](https://github.com/gizzahub/gzh-cli-gitforge) | ../gzh-cli-gitforge |
| gzh-cli-quality | [GitHub](https://github.com/gizzahub/gzh-cli-quality) | (published) |
| gzh-cli-package-manager | [GitHub](https://github.com/gizzahub/gzh-cli-package-manager) | ../gzh-cli-package-manager |
| gzh-cli-shellforge | [GitHub](https://github.com/gizzahub/gzh-cli-shellforge) | ../gzh-cli-shellforge |

**Important**: If you need to modify functionality in a wrapper command:
- Core logic → Modify in the external library repository
- CLI integration → Modify wrapper file in gzh-cli
```

______________________________________________________________________

## 실행 로드맵

### 🗓️ 단계별 실행 계획

#### Phase 1: 문서 분석 및 계획 (✅ 완료)

- [x] 현재 문서 구조 분석
- [x] 하위 프로젝트 현황 파악
- [x] 문제점 식별
- [x] 재구성 계획 수립

#### Phase 2: 핵심 문서 슬림화 (1-2일)

- [ ] README.md 재작성 (1,032줄 → 300줄)
  - [ ] 주요 기능 표 형식 요약
  - [ ] 하위 프로젝트 섹션 추가
  - [ ] 상세 내용 docs/ 링크로 이동
- [ ] CLAUDE.md 최적화 (475줄 → 300줄)
  - [ ] LLM 전용 컨텍스트로 재구성
  - [ ] 하위 프로젝트 개발 가이드 추가

#### Phase 3: 하위 프로젝트 통합 문서 작성 (1일)

- [ ] `docs/integration/00-SUBPROJECTS_GUIDE.md` 작성
  - [ ] 각 하위 프로젝트 소개
  - [ ] 독립 vs 통합 사용법 비교
  - [ ] FAQ 작성
- [ ] 기능 문서에 하위 프로젝트 링크 추가
  - [ ] docs/30-features/36-quality-management.md
  - [ ] docs/30-features/31-repository-management.md (git)
  - [ ] 기타 통합 기능 문서

#### Phase 4: 문서 검증 및 테스트 (1일)

- [ ] 모든 링크 검증
- [ ] 문서 발견성 테스트 (시나리오 기반)
  - [ ] "IDE 관리 방법" 검색 시뮬레이션
  - [ ] "하위 프로젝트 독립 사용" 검색 시뮬레이션
- [ ] LLM 가독성 테스트 (Claude Code로 검증)

#### Phase 5: 배포 및 마이그레이션 (1일)

- [ ] 기존 문서 백업
- [ ] 새 문서 구조 적용
- [ ] CI/CD 문서 빌드 파이프라인 업데이트
- [ ] 팀 공유 및 피드백 수집

______________________________________________________________________

## 예상 효과

### 📈 정량적 개선

| 메트릭 | 현재 | 목표 | 개선율 |
|-------|------|------|--------|
| README.md 길이 | 1,032줄 | 300줄 | -70% |
| CLAUDE.md 길이 | 475줄 | 300줄 | -37% |
| 하위 프로젝트 문서 | 없음 | 1개 가이드 | +100% |
| 기능 설명 중복 | 높음 | 낮음 | -80% |
| 문서 발견 시간 | ~5분 | ~1분 | -80% |

### 🎯 정성적 개선

#### 사용자 (Human) 관점

- ✅ **빠른 기능 파악**: 표 형식 요약으로 1분 내 전체 기능 이해
- ✅ **명확한 문서 경로**: "어디서 찾지?" → 명확한 링크 구조
- ✅ **하위 프로젝트 이해**: 독립 사용 가능 여부 및 방법 명확

#### LLM (AI Agent) 관점

- ✅ **컨텍스트 최적화**: 불필요한 사용자 정보 제거
- ✅ **빠른 참조**: 명령어, 패턴, 모듈별 규칙 Quick Lookup
- ✅ **하위 프로젝트 인식**: 래퍼 vs 구현 구분 명확

#### 유지보수 관점

- ✅ **중복 제거**: 단일 정보 소스 확립
- ✅ **일관성 확보**: 역할별 문서 분리로 혼선 방지
- ✅ **확장성**: 새 하위 프로젝트 추가 시 명확한 패턴

______________________________________________________________________

## 추가 고려사항

### 🔍 검토 필요 사항

#### 1. 하위 프로젝트 README 품질

**확인 필요**:

- [ ] gzh-cli-gitforge의 README가 충분히 상세한가?
- [ ] gzh-cli-quality의 독립 사용 가이드가 있는가?
- [ ] 각 프로젝트의 문서가 gzh-cli와 일관성이 있는가?

**액션**:

- 하위 프로젝트 README 템플릿 작성 (선택)
- 최소 요구사항 체크리스트 작성

#### 2. 문서 빌드 자동화

**현재 상태**: 수동 Markdown 파일
**제안**:

- [ ] MkDocs 또는 Docusaurus 도입 고려 (선택)
- [ ] 링크 자동 검증 CI 추가
- [ ] 하위 프로젝트 문서 변경 감지 (webhook)

#### 3. 다국어 지원

**현재**: 영어/한국어 혼용
**제안**:

- 사용자 문서: 한국어 우선, 영어 번역 추가
- LLM 문서: 영어 우선 (국제 표준)

______________________________________________________________________

## 결론

### ✅ 핵심 개선 사항

1. **README.md 슬림화**: 1,000줄 → 300줄, 표 형식 요약
1. **CLAUDE.md 최적화**: LLM 전용 컨텍스트로 재구성
1. **하위 프로젝트 가시성**: 신규 가이드 문서 + 각 기능 문서 링크
1. **문서 발견성 개선**: 명확한 계층 구조 + Quick Links

### 🚀 다음 단계

**즉시 시작 가능**:

1. Phase 2 (README.md 슬림화) 시작
1. Phase 3 (하위 프로젝트 가이드 작성) 병행

**승인 필요**:

- [ ] 이 계획서 검토 및 승인
- [ ] 우선순위 조정 (필요시)

______________________________________________________________________

**작성**: Claude (claude-sonnet-4-5-20250929)
**리뷰 필요**: gzh-cli 메인테이너
**예상 작업 시간**: 3-5일 (단계별 실행)
