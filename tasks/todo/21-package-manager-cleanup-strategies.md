# TODO: Package Manager - Advanced Cleanup Strategies

- status: [ ]
- priority: low
- category: package-manager
- estimated_effort: 3-4 days
- depends_on: []
- parent_task: 17-package-manager-windows-support

## Overview

고급 패키지 클린업 전략 구현. Windows 특화 기능이 아니므로 독립 구현 가능.

## Features

### 1. Quarantine Mode

관리되지 않는 패키지를 삭제 대신 격리.

```bash
gz pm clean --strategy quarantine
gz pm quarantine list
gz pm quarantine restore <package>
gz pm quarantine purge --older-than 30d
```

### 2. Dependency Analysis

사용하지 않는 의존성 식별 및 정리.

```bash
gz pm deps analyze
gz pm deps orphans
gz pm deps tree <package>
gz pm clean --orphans
```

### 3. Version Cleanup

오래된 버전 정리.

```bash
gz pm versions list <package>
gz pm clean --old-versions
gz pm clean --keep-latest 2
```

### 4. Cache Management

패키지 캐시 최적화.

```bash
gz pm cache status
gz pm cache clean
gz pm cache clean --older-than 7d
```

## Implementation Scope

### Files to Create

```
gzh-cli-package-manager/
├── pkg/cleanup/
│   ├── quarantine.go      # Quarantine management
│   ├── dependency.go      # Dependency analysis
│   ├── versions.go        # Version cleanup
│   ├── cache.go           # Cache management
│   └── *_test.go
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

## Testing Strategy

- 크로스 플랫폼 테스트 (Linux, macOS, Windows)
- 파일 시스템 모킹
- 의존성 그래프 시나리오 테스트

## Acceptance Criteria

- [ ] Quarantine 저장/복원 동작
- [ ] Orphan 패키지 식별 정확도 90%+
- [ ] 캐시 정리 안전하게 동작
- [ ] 테스트 커버리지 80%+

## Notes

- Windows PM(winget, scoop, chocolatey)과 독립적으로 구현 가능
- 기존 Linux/macOS PM에도 적용 가능

______________________________________________________________________

**Extracted from**: 17-package-manager-windows-support\_\_BLOCKED_20250729.md
