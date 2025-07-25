# Task: Remove Deprecated gen-config Command

## Objective
deprecated된 gen-config 명령어를 완전히 제거하고 synclone config generate로 통합을 완료한다.

## Requirements
- [x] gen-config 명령어 코드 제거
- [x] 모든 문서에서 gen-config 참조 제거/업데이트
- [x] 테스트 코드 업데이트
- [x] README 업데이트
- [x] 별칭에서는 유지 (하위 호환성)

## Steps

### 1. Remove gen-config Command Code
- [x] cmd/gen-config/ 디렉토리 제거
- [x] cmd/root.go에서 gen-config 명령어 제거
- [x] 관련 import 제거

### 2. Update Documentation
- [x] README.md에서 gen-config 참조 제거
- [x] CLAUDE.md 업데이트
- [x] test/e2e/README.md 업데이트
- [x] 기타 문서에서 참조 제거

### 3. Update Test Code
- [x] gen-config 관련 테스트 찾기
- [x] 테스트를 synclone config generate로 변경
- [x] 불필요한 테스트 제거

### 4. Keep Aliases for Compatibility
- [x] scripts/aliases.bash 확인 (이미 있음)
- [x] scripts/aliases.fish 확인 (이미 있음)
- [x] 별칭은 유지하여 하위 호환성 보장

### 5. Update Examples
- [x] examples/ 디렉토리 확인
- [x] gen-config 사용 예제를 synclone config generate로 변경

## Expected Output
- cmd/gen-config/ 디렉토리 삭제됨
- 모든 문서가 synclone config generate 사용
- 테스트가 정상 동작
- 별칭을 통한 하위 호환성 유지

## Verification Criteria
- [x] 빌드가 성공적으로 완료
- [x] 테스트가 모두 통과
- [x] gz gen-config 실행 시 별칭을 통해 동작
- [x] 문서에 gen-config 직접 참조 없음

## Notes
- 별칭은 최소 6개월간 유지
- 사용자에게 충분한 마이그레이션 시간 제공
- deprecation 경고는 별칭에서 계속 표시