# TODO: Package Manager - Scoop Support

- status: [ ]
- priority: medium
- category: package-manager
- estimated_effort: 2 days
- depends_on: [18-package-manager-winget-support]
- parent_task: 17-package-manager-windows-support

## Overview

Scoop 패키지 매니저 지원. 사용자 공간 설치, 관리자 권한 불필요.

## Why Scoop?

1. **관리자 권한 불필요**: 사용자 디렉토리에 설치
2. **개발자 친화적**: CLI 도구, 개발 환경 특화
3. **Portable 앱 특화**: 깔끔한 설치/제거
4. **Bucket 시스템**: 커뮤니티 패키지 풍부

## Implementation Scope

### Core Commands

```bash
gz pm scoop search <query>
gz pm scoop install <package>
gz pm scoop list
gz pm scoop update *
gz pm scoop cleanup *
gz pm scoop uninstall <package>
gz pm scoop bucket list
gz pm scoop bucket add extras
```

### Technical Requirements

- [ ] Scoop 설치 여부 감지 (`~/scoop/` 또는 SCOOP 환경변수)
- [ ] JSON 출력 파싱
- [ ] Bucket 관리 (add, remove, known)
- [ ] App manifest 파싱

### Files to Create

```
gzh-cli-package-manager/
├── pkg/scoop/
│   ├── scoop.go           # Core implementation
│   ├── bucket.go          # Bucket management
│   ├── manifest.go        # Manifest parser
│   └── scoop_test.go
```

## Testing Strategy

- winget 구현 패턴 재사용
- Mock-based tests
- JSON manifest 파싱 테스트

## Acceptance Criteria

- [ ] `gz pm scoop list` 동작
- [ ] `gz pm scoop bucket list` 동작
- [ ] Bucket add/remove 동작
- [ ] 테스트 커버리지 70%+

---

**Extracted from**: 17-package-manager-windows-support__BLOCKED_20250729.md
