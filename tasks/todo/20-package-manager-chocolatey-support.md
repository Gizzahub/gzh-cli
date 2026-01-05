# TODO: Package Manager - Chocolatey Support

- status: [ ]
- priority: low
- category: package-manager
- estimated_effort: 3 days
- depends_on: [18-package-manager-winget-support, 19-package-manager-scoop-support]
- parent_task: 17-package-manager-windows-support

## Overview

Chocolatey 패키지 매니저 지원. Windows의 가장 오래된 패키지 매니저, 관리자 권한 필요.

## Considerations

1. **관리자 권한 필요**: UAC 처리 복잡
1. **가장 큰 패키지 저장소**: GUI 앱, 시스템 도구 풍부
1. **성숙한 에코시스템**: 기업용 Pro/Business 버전 존재

## Implementation Scope

### Core Commands

```bash
gz pm chocolatey search <query>
gz pm chocolatey install <package>
gz pm chocolatey list --local-only
gz pm chocolatey upgrade all
gz pm chocolatey uninstall <package>
```

### Technical Requirements

- [ ] Chocolatey 설치 여부 감지
- [ ] 관리자 권한 확인 및 안내
- [ ] XML/JSON 출력 파싱
- [ ] 설치 진행률 표시

### Files to Create

```
gzh-cli-package-manager/
├── pkg/chocolatey/
│   ├── chocolatey.go      # Core implementation
│   ├── admin.go           # Admin rights handling
│   ├── parser.go          # Output parser
│   └── chocolatey_test.go
```

### Admin Rights Handling

```go
type AdminChecker interface {
    IsAdmin() (bool, error)
    RequestElevation() error
}
```

## Testing Strategy

- Mock-based tests (관리자 권한 모킹)
- XML/JSON 파싱 테스트
- UAC 시나리오 테스트

## Acceptance Criteria

- [ ] 관리자 권한 없을 때 명확한 에러 메시지
- [ ] `gz pm chocolatey list` 동작
- [ ] 설치/제거 시 UAC 안내
- [ ] 테스트 커버리지 70%+

## Notes

- winget과 Scoop 경험 후 구현 권장
- UAC 처리가 가장 복잡한 부분

______________________________________________________________________

**Extracted from**: 17-package-manager-windows-support\_\_BLOCKED_20250729.md
