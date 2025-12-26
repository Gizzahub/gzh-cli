# TODO: Package Manager - winget Support

- status: [ ]
- priority: medium
- category: package-manager
- estimated_effort: 2-3 days
- depends_on: []
- parent_task: 17-package-manager-windows-support

## Overview

winget (Windows Package Manager) 지원 추가. Microsoft 공식 패키지 매니저로 Windows 10 1709+ 기본 포함.

## Why winget First?

1. **낮은 진입 장벽**: Windows 10/11 기본 포함
2. **관리자 권한 선택적**: 일부 패키지만 필요
3. **간단한 CLI**: JSON 출력 지원
4. **활발한 개발**: Microsoft 공식 지원

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

- [ ] winget 설치 여부 감지
- [ ] JSON 출력 파싱 (`--output json`)
- [ ] 패키지 소스 관리 (winget, msstore)
- [ ] UAC 필요 시 안내 메시지

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

- [ ] `gz pm winget list` 동작
- [ ] `gz pm winget search golang` 동작
- [ ] `gz pm winget install --dry-run` 동작
- [ ] Windows 없는 환경에서 graceful skip
- [ ] 테스트 커버리지 70%+

## Notes

- 다른 Windows PM(Chocolatey, Scoop)보다 먼저 구현 권장
- winget 경험을 바탕으로 다른 PM 패턴 정립

---

**Extracted from**: 17-package-manager-windows-support__BLOCKED_20250729.md
