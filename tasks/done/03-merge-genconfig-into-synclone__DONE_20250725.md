# Task: Merge gen-config Functionality into synclone Command

## Objective
gen-config 명령어의 기능을 synclone 명령어의 서브커맨드로 통합하여 관련 기능을 논리적으로 그룹화한다.

## Requirements
- [x] gen-config의 모든 기능을 synclone에 통합
- [x] 기존 gen-config 사용자를 위한 알리아스 제공
- [x] 설정 파일 생성 로직 재사용
- [x] 테스트 커버리지 유지

## Steps

### 1. Analyze gen-config Command
- [x] cmd/gen-config/gen_config.go 분석
- [x] pkg/gen-config/ 패키지 기능 파악
- [x] 현재 플래그 및 옵션 목록화
- [x] 테스트 케이스 검토

### 2. Design New Command Structure
```bash
# 현재 구조
gz gen-config --output bulk-clone.yaml --source ./repos

# 새로운 구조
gz synclone config generate --output bulk-clone.yaml --source ./repos
gz synclone config validate --file bulk-clone.yaml
gz synclone config convert --from v1 --to v2 --file config.yaml
```

### 3. Implementation Tasks
- [x] synclone에 config 서브커맨드 추가
- [x] config 아래에 generate, validate, convert 서브커맨드 추가
- [x] pkg/gen-config 로직을 pkg/synclone/config로 이동
- [x] 기존 gen-config 명령어를 synclone config로 리다이렉트

### 4. Code Changes
```go
// cmd/synclone/config.go
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Manage synclone configuration files",
}

var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Generate configuration from existing repositories",
    // gen-config 로직 이동
}
```

### 5. Backward Compatibility
- [x] gen-config 명령어를 deprecated로 표시
- [x] gen-config 실행 시 synclone config generate로 자동 리다이렉트
- [x] 경고 메시지 출력: "gen-config is deprecated, use 'gz synclone config generate' instead"

### 6. Update Tests
- [x] 기존 gen-config 테스트를 synclone config 테스트로 이동
- [x] 리다이렉션 테스트 추가
- [x] 통합 테스트 업데이트

## Expected Output
- `cmd/synclone/config.go` - 새로운 config 서브커맨드
- `cmd/synclone/config_generate.go` - generate 서브커맨드
- `cmd/synclone/config_validate.go` - validate 서브커맨드
- `pkg/synclone/config/` - 이동된 gen-config 로직
- 업데이트된 테스트 파일들

## Verification Criteria
- [x] `gz synclone config generate`가 기존 gen-config와 동일하게 작동
- [x] 기존 gen-config 명령어가 deprecation 경고와 함께 작동
- [x] 모든 테스트가 통과
- [x] 문서가 새로운 명령어 구조를 반영

## Notes
- 기존 사용자의 스크립트가 깨지지 않도록 주의
- 충분한 deprecation 기간 제공
- 마이그레이션 가이드에 명확한 전환 방법 포함