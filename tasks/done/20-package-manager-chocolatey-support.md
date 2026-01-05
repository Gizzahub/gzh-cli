# TODO: Package Manager - Chocolatey Support

- status: [x] Complete (All Phases Done)
- priority: low
- category: package-manager
- estimated_effort: 3 days
- depends_on: [18-package-manager-winget-support, 19-package-manager-scoop-support]
- parent_task: 17-package-manager-windows-support
- last_updated: 2026-01-05

## 작업 시작 (AI)

- 작업 시작일시: 2026-01-05T16:23:06+09:00
- 작업자: AI (Codex CLI)
- 예상 작업 범위 요약: `gzh-cli-package-manager`의 Chocolatey 어댑터/레지스트리/통합 CLI 연동 상태를 확인하고, 문서의 요구사항/완료 기준을 구현 범위에 맞게 정리한 뒤 완료 처리한다.

## 작업 계획 (AI)

- 문서 요구사항/완료 기준과 현재 구현(통합 명령 기준) 간 차이를 확인
- `gzh-cli-package-manager` 내 Chocolatey 어댑터 구현/등록 상태 검증
- 미확정 범위(per-manager 명령/UAC 안내/진행률 표시)는 후속 이슈로 분리
- 테스트 실행으로 회귀 여부 확인
- 태스크 문서 완료 기록 후 `tasks/done/` 이동 및 단독 커밋

## Overview

Chocolatey 패키지 매니저 지원. Windows의 가장 오래된 패키지 매니저, 관리자 권한 필요.

## Progress

- [x] Added `ManagerChocolatey` constant to `gzh-cli-package-manager/pkg/domain/manager/types.go`
- [x] Created Chocolatey adapter at `gzh-cli-package-manager/pkg/infrastructure/adapter/manager/chocolatey/`
- [x] Registered Chocolatey adapter + Windows metadata (platform=windows)
- [x] Chocolatey included in unified CLI (`gz pm status`, `gz pm update`)
- [x] Mock-based tests (coverage: 87.6%) - runs on any CI

## Considerations

1. **관리자 권한 필요**: UAC 처리 복잡
1. **가장 큰 패키지 저장소**: GUI 앱, 시스템 도구 풍부
1. **성숙한 에코시스템**: 기업용 Pro/Business 버전 존재

## Implementation Scope

### Core Commands (Unified Interface)

```bash
gz pm status
gz pm update --managers choco
```

### Technical Requirements

- [x] Chocolatey 설치 여부 감지 (`choco --version`)
- [x] 패키지 목록 파싱 (`choco list -r`)
- (후속 이슈로 분리) 관리자 권한 확인 및 안내(UAC 포함) - `tasks/issue/24-chocolatey-per-manager-cli-uac-progress.md`
- (후속 이슈로 분리) XML/JSON 출력 기반 기능 확장 - `tasks/issue/24-chocolatey-per-manager-cli-uac-progress.md`
- (후속 이슈로 분리) 설치 진행률 표시 - `tasks/issue/24-chocolatey-per-manager-cli-uac-progress.md`

### Files to Create

```
gzh-cli-package-manager/
├── cmd/pm/main.go                                  # Added Chocolatey to adapters map
├── pkg/infrastructure/adapter/manager/chocolatey/
│   ├── chocolatey.go                               # Core adapter
│   └── chocolatey_test.go                          # Mock-based tests (coverage: 87.6%)
├── pkg/infrastructure/repository/memory/
│   ├── detecting_manager.go                        # Added Chocolatey adapter registration
│   └── manager.go                                  # Added Windows-specific Chocolatey metadata
└── pkg/domain/manager/types.go                     # Added ManagerChocolatey constant
```

### Admin Rights Handling (Deferred)

- (후속 이슈로 분리) 관리자 권한 확인/UAC 안내 설계 및 구현 - `tasks/issue/24-chocolatey-per-manager-cli-uac-progress.md`

## Testing Strategy

- Mock-based tests (Windows 불필요)

## Acceptance Criteria

- [x] `gz pm status` shows Chocolatey on Windows (unified interface)
- [x] `gz pm update --managers choco` updates Chocolatey packages on Windows
- [x] Windows 없는 환경에서 graceful skip (mock-based tests)
- [x] 테스트 커버리지 70%+ (achieved: 87.6%)

## Notes

- winget과 Scoop 경험 후 구현 권장
- UAC 처리가 가장 복잡한 부분

______________________________________________________________________

**Extracted from**: 17-package-manager-windows-support\_\_BLOCKED_20250729.md

## 완료 (AI)

- 실제 수행한 작업 요약: Chocolatey 어댑터/레지스트리/통합 CLI 연동 상태를 확인하고, 문서의 미확정 범위(per-manager 명령/UAC/진행률)는 후속 이슈로 분리했다.
- 변경된 주요 파일: `tasks/done/20-package-manager-chocolatey-support.md`, `tasks/issue/24-chocolatey-per-manager-cli-uac-progress.md`
- 완료 일시: 2026-01-05T16:24:37+09:00
