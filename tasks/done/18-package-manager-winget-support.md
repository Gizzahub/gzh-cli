# TODO: Package Manager - winget Support

- status: [x] Complete (All Phases Done)
- priority: medium
- category: package-manager
- estimated_effort: 2-3 days
- depends_on: []
- parent_task: 17-package-manager-windows-support
- last_updated: 2025-12-27

## 작업 시작 (AI)

- 작업 시작일시: 2026-01-05T16:14:38+09:00
- 작업자: AI (Codex CLI)
- 예상 작업 범위 요약: 기존 구현(winget 어댑터/레지스트리/CLI 래핑) 상태를 검증하고, 태스크 문서/완료 처리(필요 시 후속 이슈 분리)를 정리한다.

## 작업 계획 (AI)

- 요구사항/완료 기준 재확인(문서 내 미완료 체크 항목 포함)
- `gzh-cli-package-manager` 내 winget 어댑터/레지스트리/CLI 통합 상태 확인
- 현재 CLI가 제공하는 범위(통합 명령 vs per-manager 명령)를 기준으로 스코프 정합성 정리
- 테스트 실행으로 회귀 여부 확인
- 태스크 문서에 실제 작업/변경 파일/완료 일시 기록 후 `tasks/done/` 이동 및 단독 커밋

## Overview

winget (Windows Package Manager) 지원 추가. Microsoft 공식 패키지 매니저로 Windows 10 1709+ 기본 포함.

## Progress

### Phase 1: Core Adapter (Complete - 2025-12-27)

- [x] Added `ManagerWinget` constant to `pkg/domain/manager/types.go`
- [x] Created winget adapter at `pkg/infrastructure/adapter/manager/winget/`
- [x] Implemented all Adapter interface methods:
  - Detect, GetVersion, GetBinaryPath, GetConfigPath
  - ListPackages (JSON + text fallback parsing)
  - CheckHealth, Update
- [x] Mock-based tests (94.6% coverage) - runs on any CI
- [x] Fixed golangci-lint v2 compatibility in `.golangci.yml`

### Phase 2: Registry Integration (Complete - 2025-12-27)

- [x] Register winget adapter in `detecting_manager.go`
- [x] Add winget metadata to base `manager.go` (Windows-only)
- [x] Update adapter registration test

### Phase 3: CLI Integration (Complete - 2025-12-27)

- [x] Add winget adapter to `main.go` adapters map
- [x] CLI commands work via unified interface (`gz pm status`, `gz pm update`)
- [x] All tests pass
- (후속 이슈로 분리) Windows 환경 Integration tests (requires Windows environment) - `tasks/issue/22-winget-per-manager-commands-and-sources.md`

### Files Created/Modified

```
gzh-cli-package-manager/
├── cmd/pm/main.go                # Added winget to adapters map
├── pkg/infrastructure/adapter/manager/winget/
│   ├── winget.go           # Core adapter (hexagonal architecture)
│   └── winget_test.go      # Mock-based tests (94.6% coverage)
├── pkg/infrastructure/repository/memory/
│   ├── detecting_manager.go      # Added winget adapter registration
│   ├── detecting_manager_test.go # Updated expected adapter count
│   └── manager.go                # Added Windows-specific winget metadata
├── pkg/domain/manager/types.go   # Added ManagerWinget constant
└── .golangci.yml                 # Fixed golangci-lint v2 compatibility
```

## Why winget First?

1. **낮은 진입 장벽**: Windows 10/11 기본 포함
1. **관리자 권한 선택적**: 일부 패키지만 필요
1. **간단한 CLI**: JSON 출력 지원
1. **활발한 개발**: Microsoft 공식 지원

## Implementation Scope

### Core Commands

```bash
gz pm winget search <query>
gz pm winget install <package>
gz pm winget list
gz pm winget upgrade <package>
gz pm winget upgrade --all
gz pm winget uninstall <package>
```

### Technical Requirements

- [x] winget 설치 여부 감지 (`Detect` method)
- [x] JSON 출력 파싱 (with text fallback)
- (후속 이슈로 분리) 패키지 소스 관리 (winget, msstore) - `tasks/issue/22-winget-per-manager-commands-and-sources.md`
- (후속 이슈로 분리) UAC 필요 시 안내 메시지 - `tasks/issue/22-winget-per-manager-commands-and-sources.md`

### Files to Create

```
gzh-cli-package-manager/
├── pkg/winget/
│   ├── winget.go          # Core implementation
│   ├── parser.go          # JSON output parser
│   └── winget_test.go     # Tests (mock-based)
```

### Files to Modify

- `gzh-cli-package-manager/pkg/manager/registry.go` - Register winget manager

## Testing Strategy

- Mock-based tests (CI에서 Windows 불필요)
- 실제 winget 있으면 integration test
- Build tag: `//go:build windows`

## Acceptance Criteria

- [x] `gz pm status` shows winget on Windows (unified interface)
- [x] `gz pm update` updates winget packages on Windows
- (후속 이슈로 분리) `gz pm winget search golang` 동작 (per-manager commands) - `tasks/issue/22-winget-per-manager-commands-and-sources.md`
- [x] Windows 없는 환경에서 graceful skip (mock-based tests)
- [x] 테스트 커버리지 70%+ (achieved: 94.6%)

## Notes

- 다른 Windows PM(Chocolatey, Scoop)보다 먼저 구현 권장
- winget 경험을 바탕으로 다른 PM 패턴 정립

______________________________________________________________________

**Extracted from**: 17-package-manager-windows-support\_\_BLOCKED_20250729.md

## 완료 (AI)

- 실제 수행한 작업 요약: 기존 winget 지원(어댑터/레지스트리/통합 CLI) 구현 상태를 확인하고, 미확정 범위(per-manager 명령/소스/UAC/Windows IT)는 후속 이슈로 분리했다.
- 변경된 주요 파일: `tasks/done/18-package-manager-winget-support.md`, `tasks/issue/22-winget-per-manager-commands-and-sources.md`
- 완료 일시: 2026-01-05T16:18:02+09:00
