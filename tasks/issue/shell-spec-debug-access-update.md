# Shell 명세서 업데이트: 디버그 전용 접근 명시

## 🎯 문제점

`specs/shell.md` 명세서와 실제 구현 간의 차이점이 발견되었습니다.

### 현재 명세서
- `gz shell` 명령어가 일반적으로 접근 가능한 것처럼 기술되어 있음
- 상시 사용 가능한 대화형 셸로 설명되어 있음

### 실제 구현
- `gz shell` 명령어는 디버그 모드에서만 접근 가능
- `--debug-shell` 플래그 또는 `GZH_DEBUG_SHELL=1` 환경변수 필요
- `cmd.Hidden = true`로 일반 도움말에서 숨겨진 상태

## 📝 제안사항

### 1. 접근 방법 명시
명세서 상단에 디버그 전용 접근임을 명시:

```markdown
## Overview

⚠️ **디버그 전용 기능**: `shell` 명령어는 디버그 모드에서만 접근 가능합니다.

다음 중 하나의 방법으로 활성화할 수 있습니다:
- `gz --debug-shell` 플래그 사용
- `GZH_DEBUG_SHELL=1` 환경변수 설정

The `shell` command provides an interactive debugging shell (REPL) for real-time system inspection, dynamic configuration changes, and live troubleshooting.
```

### 2. 사용 예시 업데이트
모든 사용 예시에 디버그 모드 활성화 방법 포함:

```bash
# Before
gz shell                                    # Start interactive shell

# After
GZH_DEBUG_SHELL=1 gz shell                  # Start interactive debugging shell
# or
gz --debug-shell shell                      # Alternative activation method
```

### 3. 보안 고려사항 추가
디버그 전용인 이유 설명:

```markdown
## Security Considerations

### Debug-Only Access
- Shell access is restricted to debug mode to prevent accidental usage in production
- Requires explicit activation through environment variable or flag
- Hidden from general help to avoid confusion
```

## 🎯 우선순위
**Medium** - 명세서와 구현 일치를 위해 필요하지만 긴급하지 않음

## 📅 작업 범위
- `specs/shell.md` 파일 수정 (AI_MODIFY_PROHIBITED이므로 수동 작업 필요)
- 디버그 접근 방법 명시
- 사용 예시 업데이트
- 보안 고려사항 추가

## ✅ 완료 조건
- [ ] 명세서에 디버그 전용 접근임을 명시
- [ ] 모든 사용 예시에 활성화 방법 포함
- [ ] 보안 고려사항에 디버그 모드 설명 추가
- [ ] 개발팀 검토 완료
