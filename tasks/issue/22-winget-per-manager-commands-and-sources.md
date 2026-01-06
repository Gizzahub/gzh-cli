skill with Clean Architecture patterns
- FastAPI/Django/Flask integra# ISSUE: winget per-manager commands & sources

- status: open
- priority: medium
- category: package-manager
- created_at: 2026-01-05T16:14:38+09:00
- derived_from: tasks/done/18-package-manager-winget-support.md

## Background

`gzh-cli-package-manager`에 winget 어댑터가 통합되어 `gz pm status`, `gz pm update` 등 **통합(unified) 명령** 기준으로는 winget 지원이 완료되었다.

다만, 문서 초안에 포함되어 있던 winget 전용(per-manager) 명령과 고급 기능(소스/권한 안내)은 별도 설계가 필요하여 후속 이슈로 분리한다.

## Scope

- winget 전용 CLI(예: `gz pm winget ...`) 제공 여부/형태 결정
- 다음 기능 구현 및 테스트(가능한 범위에서 mock 기반):
  - `search`, `install`, `uninstall`, `upgrade`, `upgrade --all`, `list`
  - 패키지 소스 관리(winget/msstore 등) 지원 여부
  - UAC/권한 필요 시 사용자 안내 메시지/에러 처리 정책
- Windows 환경에서의 통합 테스트 전략 수립(선택)

## Acceptance Criteria

- 스펙 결정 문서(명령 구조/출력 형식/에러 정책) 정리
- `gz pm winget <subcommand>` 제공 시: 기본 서브커맨드 최소 1개 이상 동작 및 테스트 포함
- 구현 범위가 확정되면 `tasks/todo/`로 태스크 전환

