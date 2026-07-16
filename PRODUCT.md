# Product Goals (No-PRD)

**Project**: gzh-cli (`gz` binary)
**Doc Type**: Goals + Constraints + Quality Gates
**Status**: Active
**Last Updated**: 2026-07-16

______________________________________________________________________

## Product Intent

gzh-cli is the **thin assembling surface** of the gzh-cli library federation. The
`gz` binary is the single entry point that:

- integrates independent `gzh-cli-*` libraries through thin wrappers,
- makes bulk-first workflows (parallel, dry-run, progress) the default across
  git / quality / package-manager / dev·net·os·shell environments,
- and adds convenience over existing tools without reimplementing them.

This document sits below [SOUL.md](SOUL.md) (philosophy) and is the repo-level
contract. It replaces a full PRD.

| 제공하는 것 (Is)                          | 되지 않을 것 (Is Not)                       |
| ----------------------------------------- | ------------------------------------------- |
| 라이브러리 연합의 단일 진입점 (`gz`)      | 독립 실행 로직을 품은 모놀리식 앱           |
| 얇은 wrapper를 통한 CLI 통합              | 핵심 로직을 wrapper에 구현                  |
| bulk-first·dry-run 표준 장비              | git/brew/asdf 등 하위 도구 재구현           |
| 여러 도구를 감싸는 편의 계층              | GUI·웹·IDE 플러그인·고급 TUI                |

______________________________________________________________________

## Goals (Measurable Targets)

G1. **Thin assembler (wrapper discipline)**

- Target: `cmd/*_wrapper.go`에 핵심 비즈니스 로직 0줄 — 라이브러리 위임만

G2. **Startup and response latency**

- Target: `gz` 기동 < 50ms, 대다수 명령 응답 < 100ms

G3. **Binary footprint**

- Target: 단일 정적 바이너리 ~33MB 이하 유지

G4. **Library-first parity**

- Target: 모든 신규 기능은 `gzh-cli-*` 라이브러리로 먼저 존재하고, CLI 없이
  import만으로 동일 기능 사용 가능 (100%)

G5. **Bulk-first defaults**

- Target: 다중 리포·다중 환경 명령은 병렬 실행·dry-run·진행 표시를 기본 제공

______________________________________________________________________

## Non-Goals (Explicitly Out of Scope)

- No GUI, 웹 인터페이스, IDE 플러그인, 고급 TUI
- No git/brew/asdf 등 하위 도구의 재구현 (감싸기만)
- No wrapper 내부 핵심 비즈니스 로직
- No 모노레포화 — feature 라이브러리는 독립 배포·독립 사용 유지
- No CI/CD 오케스트레이션 자체 (파이프라인용 명령·출력만 제공)

______________________________________________________________________

## Guardrails and Technical Constraints

**Architecture**

- Integration Libraries Pattern: `cmd/*_wrapper.go`는 얇게, 핵심은 `gzh-cli-*`
- 모든 작업은 `context.Context`로 취소/타임아웃을 받는다

**Dependency Boundaries**

- gzh-cli는 **유일한 조립자** — 모든 feature 라이브러리를 의존할 수 있는 단 하나의 리포
- `replace` 지시자는 `make local-dev` 전용, 커밋 전 `make local-dev-disable` 필수

**Compatibility**

- Go 1.25+ (`go.mod` go 1.25.7; devbox 툴체인 1.26)

**Safety**

- 파괴적 작업은 명시적 플래그 또는 dry-run을 요구한다; 기본값은 안전한 쪽

**Documentation**

- [SOUL.md](SOUL.md)(철학) → 본 문서(계약) 계층을 유지; 명령 레퍼런스는 실제 플래그와 일치

______________________________________________________________________

## Quality Gates (Release Readiness)

**Build and Lint**

- `make fmt && make lint && make build` pass with no warnings

**Testing**

- `make test` pass; 핵심 로직 커버리지 >= 80%

**Performance**

- 기동 < 50ms, 대다수 명령 응답 < 100ms

**Integration**

- 모든 feature 라이브러리를 릴리스 태그로 통합·테스트; `make local-dev-disable` 후 커밋

**Docs**

- SOUL.md 승인 게이트와 정합; CLI 레퍼런스가 실제 명령·플래그와 일치

______________________________________________________________________

## Decision Rules

- 새 기능은 SOUL.md **4-게이트**(틈 · 라이브러리 · 대량/전환 · 날카로움)를 모두 통과해야 한다
- 최소 하나의 goal에 매핑되거나 명시적으로 승인되어야 한다
- Guardrails를 위반하는 변경은 문서화된 예외를 요구한다
- Quality Gates 미충족 시 릴리스는 차단된다
- 신념이 충돌하면 SOUL.md 우선표(깊이>넓이, 안전>속도, 라이브러리>CLI…)를 따른다

______________________________________________________________________

**End of Document**
