# ISSUE: Scoop per-manager commands, bucket, manifest

- status: open
- priority: medium
- category: package-manager
- created_at: 2026-01-05T16:19:57+09:00
- derived_from: tasks/done/19-package-manager-scoop-support.md

## Background

`gzh-cli-package-manager`에 Scoop 어댑터가 통합되어 `gz pm status`, `gz pm update` 등 **통합(unified) 명령** 기준의 Scoop 지원은 완료되었다.

하지만 초기 문서에 포함되어 있던 Scoop 전용(per-manager) 명령(`gz pm scoop ...`)과 버킷/매니페스트 기반의 확장 기능은 별도 설계가 필요하여 후속 이슈로 분리한다.

## Scope

- Scoop 전용 CLI 제공 여부/형태 결정
- 다음 기능 구현 및 테스트(가능한 범위에서 mock 기반):
  - `search`, `install`, `uninstall`, `update`, `cleanup`, `list`
  - `bucket` 관리(list/add/remove/known)
  - (선택) manifest 파싱/표시 정책
- Windows 환경에서의 통합 테스트 전략 수립(선택)

## Acceptance Criteria

- 스펙 결정 문서(명령 구조/출력 형식/에러 정책) 정리
- `gz pm scoop <subcommand>` 제공 시: 기본 서브커맨드 최소 1개 이상 동작 및 테스트 포함
- 구현 범위가 확정되면 `tasks/todo/`로 태스크 전환

