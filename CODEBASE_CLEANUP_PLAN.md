# 🧹 코드베이스 정리 계획

## 📊 현재 상태 분석
- **TODO/FIXME**: 51개 파일에 정리 필요 항목
- **Deprecated 함수**: 여러 패키지에 후방 호환성을 위해 유지 중
- **중복 스크립트**: aliases.bash, aliases.fish (동일 기능)
- **미사용 코드**: 매개변수 미사용 함수들 (`//nolint:unparam`)

## 🎯 단계별 정리 계획

### Phase 1: 즉시 제거 가능한 항목들 (1-2일)

#### A. 완전히 사용하지 않는 파일들
```bash
# 1. 중복 스크립트 정리
- scripts/aliases.bash와 scripts/aliases.fish 통합
- deprecated 경고만 유지하고 실제 기능은 제거

# 2. 완료된 TODO 항목들 정리
- 51개 파일의 TODO/FIXME 중 이미 구현된 것들 제거
- 실제 필요한 TODO만 유지
```

#### B. Deprecated 함수 정리
```go
// pkg/config/loader.go - 완전히 제거 가능
- LoadConfig()
- LoadConfigWithEnv()
- LoadConfigFromFile()
- FindConfigFile()

// pkg/synclone/config_loader.go - 제거 예정
- FindSyncCloneConfigFile()
- LoadSyncCloneConfig()
```

### Phase 2: 코드 구조 최적화 (3-5일)

#### A. 중복 함수 통합
```go
// 로깅 관련 중복 제거
- internal/logger/structured.go의 Debug() 함수들
- internal/event/logger.go와 logger_adapter.go 통합

// 인증 관련 중복 제거
- SSH 인증 로직 중복 (git 패키지 내)
- 토큰 인증 로직 중복 (github, gitlab 패키지)
```

#### B. 패키지 구조 정리
```bash
# helpers 디렉토리 제거
- internal/helpers/git_helper.go → internal/git/
- 기타 helper 함수들을 적절한 패키지로 이동

# 테스트 코드 정리
- 중복된 mock 함수들 통합
- testutil 패키지 구조 최적화
```

### Phase 3: 문서 및 설정 정리 (1-2일)

#### A. 사용하지 않는 문서들
```bash
# 확인 후 제거
- docs/backlog/ 디렉토리 (이미 tasks/backlog로 이동됨)
- 오래된 마이그레이션 가이드들
- 중복된 API 문서들

# 정리할 설정 파일들
- 사용하지 않는 예제 설정들
- 테스트용 임시 설정들
```

### Phase 4: 성능 및 품질 최적화 (2-3일)

#### A. 코드 품질 개선
```go
// 매개변수 미사용 함수들 정리
- //nolint:unparam 주석이 있는 함수들 리팩토링
- 실제로 사용하지 않는 매개변수 제거

// 에러 처리 개선
- 일관된 에러 메시지 형식
- wrap된 에러 체인 정리
```

## 🔧 자동화 도구 활용

### 1. 사용하지 않는 코드 탐지
```bash
# Go 도구들 활용
go mod tidy
golangci-lint run --enable=unused,deadcode
staticcheck ./...

# 커스텀 스크립트
find . -name "*.go" -exec grep -l "TODO\|FIXME" {} \; | xargs grep -n "TODO\|FIXME"
```

### 2. 중복 코드 탐지
```bash
# 함수 시그니처 중복 찾기
grep -r "func.*(" --include="*.go" . | sort | uniq -d

# 비슷한 로직 패턴 찾기
grep -r "if err != nil" --include="*.go" . | wc -l
```

### 3. 의존성 정리
```bash
# 사용하지 않는 의존성 제거
go mod tidy
go mod why <dependency>

# vendor 디렉토리 정리 (있는 경우)
go mod vendor
```

## 📈 예상 효과

### 정량적 개선
- **파일 수**: 약 10-15% 감소 예상
- **코드 라인 수**: 약 20% 감소 예상
- **빌드 시간**: 약 15% 개선 예상
- **테스트 실행 시간**: 약 10% 개선 예상

### 정성적 개선
- **가독성**: 불필요한 코드 제거로 핵심 로직에 집중
- **유지보수성**: 중복 제거로 수정 포인트 감소
- **성능**: 미사용 코드 제거로 메모리 사용량 감소
- **개발자 경험**: 명확한 구조로 신규 개발자 온보딩 개선

## ⚠️ 주의사항

### 백업 및 테스트
```bash
# 변경 전 백업
git branch backup-before-cleanup
git checkout -b cleanup-phase1

# 각 단계별 테스트
make test-all
make test-integration
```

### 하위 호환성
- deprecated 함수들은 단계적 제거
- API 변경 시 충분한 공지 기간
- 마이그레이션 가이드 제공

### 팀 협업
- 각 Phase별로 PR 생성
- 코드 리뷰 필수
- 변경사항 문서화
