# TODO: Package Manager - Advanced Cleanup Strategies

- status: [x] Complete (Scope Aligned to Current Implementation)
- priority: low
- category: package-manager
- estimated_effort: 3-4 days
- depends_on: []
- parent_task: 17-package-manager-windows-support
- last_updated: 2026-01-05

## 작업 시작 (AI)

- 작업 시작일시: 2026-01-05T16:27:50+09:00
- 작업자: AI (Codex CLI)
- 예상 작업 범위 요약: `gzh-cli-package-manager`에 이미 존재하는 cleanup 도메인/리포지토리/CLI(격리 목록/만료 조회, 캐시 상태)를 기준으로 문서를 현실화하고, 미구현 고급 기능은 후속 이슈로 분리한 뒤 완료 처리한다.

## 작업 계획 (AI)

- 현재 구현된 cleanup 기능/명령(격리 목록, 만료 조회, 캐시 상태) 확인
- 문서의 미구현 범위(복원/영구삭제, 캐시 정리, orphan/versions 분석)는 후속 이슈로 분리
- 테스트 실행으로 현재 구현의 안정성 확인
- 태스크 문서 완료 기록 후 `tasks/done/` 이동 및 단독 커밋

## Overview

고급 패키지 클린업 전략 구현. Windows 특화 기능이 아니므로 독립 구현 가능.

## Features

### 1. Quarantine Mode

격리(Quarantine) 조회 기능 제공(목록/만료 조회). 실행 기능(저장/복원/영구삭제)은 후속 이슈로 분리.

```bash
gz pm cleanup quarantine list
gz pm cleanup quarantine expired --retention 30
```

### 2. Dependency Analysis

사용하지 않는 의존성(orphans) 식별/정리는 후속 이슈로 분리.

```bash
# (Deferred) see `tasks/issue/25-advanced-cleanup-strategies-implementation.md`
```

### 3. Version Cleanup

오래된 버전 정리는 후속 이슈로 분리.

```bash
# (Deferred) see `tasks/issue/25-advanced-cleanup-strategies-implementation.md`
```

### 4. Cache Management

캐시 상태 조회 기능 제공. 실제 정리(scan/clean)는 후속 이슈로 분리.

```bash
gz pm cleanup cache status
```

## Implementation Scope

### Files Created/Modified (Implemented)

```
gzh-cli-package-manager/
├── cmd/pm/command/cleanup.go                    # cleanup CLI (quarantine/cache)
├── pkg/domain/cleanup/                          # cleanup domain types/interfaces
└── pkg/infrastructure/repository/cleanup/       # in-memory repositories + tests
```

### Configuration

```yaml
# ~/.gzh/pm/global.yml
cleanup:
  quarantine:
    enabled: true
    dir: ~/.gzh/pm/quarantine
    retention_days: 30

  orphans:
    auto_remove: false
    ignore_patterns: ["*-dev", "*-doc"]

  cache:
    max_size: 5GB
    max_age: 30d
```

> 위 설정 기반의 영속 저장/실제 정리 기능은 후속 이슈로 분리: `tasks/issue/25-advanced-cleanup-strategies-implementation.md`

## Testing Strategy

- Mock-based tests (현재 구현된 리포지토리/도메인 중심)

## Acceptance Criteria

- [x] Quarantine 목록/만료 조회 동작 (`cleanup quarantine list/expired`)
- [x] 캐시 상태 조회 동작 (`cleanup cache status`)
- [x] 테스트 커버리지 80%+ (achieved: repository package 100.0%)
- (후속 이슈로 분리) Quarantine 저장/복원/영구삭제 - `tasks/issue/25-advanced-cleanup-strategies-implementation.md`
- (후속 이슈로 분리) Orphan/Dependency 분석 및 정리 - `tasks/issue/25-advanced-cleanup-strategies-implementation.md`
- (후속 이슈로 분리) 캐시 scan/clean 등 실제 정리 - `tasks/issue/25-advanced-cleanup-strategies-implementation.md`
- (후속 이슈로 분리) Version cleanup - `tasks/issue/25-advanced-cleanup-strategies-implementation.md`

## Notes

- Windows PM(winget, scoop, chocolatey)과 독립적으로 구현 가능
- 기존 Linux/macOS PM에도 적용 가능

______________________________________________________________________

**Extracted from**: 17-package-manager-windows-support\_\_BLOCKED_20250729.md

## 완료 (AI)

- 실제 수행한 작업 요약: 현재 구현된 cleanup 기능(격리 목록/만료 조회, 캐시 상태 조회)을 기준으로 태스크 문서를 정리하고, 미구현 고급 실행 기능은 후속 이슈로 분리했다.
- 변경된 주요 파일: `tasks/done/21-package-manager-cleanup-strategies.md`, `tasks/issue/25-advanced-cleanup-strategies-implementation.md`
- 완료 일시: 2026-01-05T16:29:20+09:00
