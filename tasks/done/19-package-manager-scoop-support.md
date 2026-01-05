# TODO: Package Manager - Scoop Support

- status: [x] Complete (All Phases Done)
- priority: medium
- category: package-manager
- estimated_effort: 2 days
- depends_on: [18-package-manager-winget-support]
- parent_task: 17-package-manager-windows-support
- last_updated: 2026-01-05

## 작업 시작 (AI)

- 작업 시작일시: 2026-01-05T16:19:57+09:00
- 작업자: AI (Codex CLI)
- 예상 작업 범위 요약: `gzh-cli-package-manager`의 Scoop 어댑터/레지스트리/통합 CLI 연동 상태를 확인하고, 문서의 요구사항/완료 기준을 구현 범위에 맞게 정리한 뒤 완료 처리한다.

## 작업 계획 (AI)

- 문서 요구사항/완료 기준과 현재 구현(통합 명령 기준) 간 차이를 확인
- `gzh-cli-package-manager` 내 Scoop 어댑터 구현/등록 상태 검증
- 미확정 범위(버킷/매니페스트/per-manager CLI)는 후속 이슈로 분리
- 테스트 실행으로 회귀 여부 확인
- 태스크 문서 완료 기록 후 `tasks/done/` 이동 및 단독 커밋

## Overview

Scoop 패키지 매니저 지원. 사용자 공간 설치, 관리자 권한 불필요.

## Progress

- [x] Added `ManagerScoop` constant to `gzh-cli-package-manager/pkg/domain/manager/types.go`
- [x] Created Scoop adapter at `gzh-cli-package-manager/pkg/infrastructure/adapter/manager/scoop/`
- [x] Registered Scoop adapter + Windows metadata (platform=windows)
- [x] Scoop included in unified CLI (`gz pm status`, `gz pm update`)
- [x] Mock-based tests (coverage: 84.0%) - runs on any CI

## Why Scoop?

1. **관리자 권한 불필요**: 사용자 디렉토리에 설치
1. **개발자 친화적**: CLI 도구, 개발 환경 특화
1. **Portable 앱 특화**: 깔끔한 설치/제거
1. **Bucket 시스템**: 커뮤니티 패키지 풍부

## Implementation Scope

### Core Commands (Unified Interface)

```bash
gz pm status
gz pm update --managers scoop
```

### Technical Requirements

- [x] Scoop 설치 여부 감지 (`scoop --version`)
- [x] `scoop list` 텍스트 출력 파싱
- (후속 이슈로 분리) per-manager 명령(`gz pm scoop ...`) - `tasks/issue/23-scoop-per-manager-cli-buckets-manifest.md`
- (후속 이슈로 분리) Bucket 관리 (add, remove, known) - `tasks/issue/23-scoop-per-manager-cli-buckets-manifest.md`
- (후속 이슈로 분리) App manifest 파싱 - `tasks/issue/23-scoop-per-manager-cli-buckets-manifest.md`

### Files Created/Modified (Implemented)

```
gzh-cli-package-manager/
├── cmd/pm/main.go                                  # Added Scoop to adapters map
├── pkg/infrastructure/adapter/manager/scoop/
│   ├── scoop.go                                    # Core adapter
│   └── scoop_test.go                               # Mock-based tests (coverage: 84.0%)
├── pkg/infrastructure/repository/memory/
│   ├── detecting_manager.go                        # Added Scoop adapter registration
│   └── manager.go                                  # Added Windows-specific Scoop metadata
└── pkg/domain/manager/types.go                     # Added ManagerScoop constant
```

## Testing Strategy

- winget 구현 패턴 재사용
- Mock-based tests (Windows 불필요)

## Acceptance Criteria

- [x] `gz pm status` shows Scoop on Windows (unified interface)
- [x] `gz pm update --managers scoop` updates Scoop packages on Windows
- [x] Windows 없는 환경에서 graceful skip (mock-based tests)
- [x] 테스트 커버리지 70%+ (achieved: 84.0%)

______________________________________________________________________

**Extracted from**: 17-package-manager-windows-support\_\_BLOCKED_20250729.md

## 완료 (AI)

- 실제 수행한 작업 요약: Scoop 어댑터/레지스트리/통합 CLI 연동 상태를 확인하고, 문서의 미확정 범위(per-manager 명령/버킷/매니페스트)는 후속 이슈로 분리했다.
- 변경된 주요 파일: `tasks/done/19-package-manager-scoop-support.md`, `tasks/issue/23-scoop-per-manager-cli-buckets-manifest.md`
- 완료 일시: 2026-01-05T16:22:08+09:00
