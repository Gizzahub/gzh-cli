# ISSUE: Advanced cleanup strategies implementation

- status: open
- priority: low
- category: package-manager
- created_at: 2026-01-05T16:27:50+09:00
- derived_from: tasks/done/21-package-manager-cleanup-strategies.md

## Background

현재 `gzh-cli-package-manager`에는 cleanup 도메인 타입/인터페이스와 인메모리 리포지토리, 그리고 일부 조회용 CLI(`cleanup quarantine list/expired`, `cleanup cache status`)가 존재한다.

하지만 실제 “고급 클린업 전략”으로 제시된 실행 기능(격리 저장/복원/영구삭제, 캐시 정리, orphan/versions 분석)은 아직 설계/구현이 필요하여 후속 이슈로 분리한다.

## Scope

- Quarantine 실행 기능
  - 패키지 “격리 저장”의 의미 정의(파일 백업 vs 매니저 uninstall + reinstall 가능한 메타데이터 저장)
  - `cleanup quarantine restore`, `cleanup quarantine purge` 구현
  - 영구 보관 디렉토리/retention 설정 및 영속 저장소(파일 기반) 도입 여부 결정
- Cache 관리 기능
  - `cleanup cache scan`, `cleanup cache clean` 구현(안전/드라이런/권한 고려)
  - 매니저별 캐시 경로/정리 방법 매핑(크로스 플랫폼)
- Dependency/Orphans & Version cleanup
  - `cleanup deps/orphans` 및 `versions` 기능 스펙 확정(매니저별 가능 범위 정의)
  - 정확도/안전성 기준 및 테스트 전략(플랫폼별) 수립

## Acceptance Criteria

- 스펙 문서(각 서브커맨드 의미/지원 매니저/에러 정책) 확정
- 최소 1개의 실행 기능(예: cache scan/clean 또는 quarantine purge)이 실제로 동작 + 테스트 포함
- 구현 범위가 확정되면 `tasks/todo/`로 태스크 전환(필요 시 분할)

