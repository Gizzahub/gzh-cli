# 🎯 즉시 실행 가능한 정리 액션

## 🔧 자동 수정 가능한 항목들

### 1. 코드 포맷팅 및 스타일 (5분)
```bash
# 자동 포맷팅
gofumpt -w .
gci write .

# Import 정리
go mod tidy
```

### 2. JSON 태그 네이밍 수정 (10분)
```bash
# camelCase로 변경
sed -i 's/json:"stack_trace"/json:"stackTrace"/g' internal/errors/standard_errors.go
sed -i 's/json:"status_code"/json:"statusCode"/g' internal/httpclient/interfaces.go
# ... 기타 태그들
```

### 3. 사용하지 않는 매개변수 정리 (15분)
```bash
# _ 로 변경하여 unused 경고 제거
# 예: func example(ctx context.Context) -> func example(_ context.Context)
```

## 🗄️ 파일 정리 (30분)

### 제거 가능한 파일들
```bash
# 1. 중복 스크립트 (백업 후 제거)
mv scripts/aliases.bash scripts/backup/
mv scripts/aliases.fish scripts/backup/

# 2. 테스트 임시 파일들
find . -name "*_test_temp*" -delete
find . -name "*.tmp" -delete

# 3. 사용하지 않는 예제 파일들
# (신중히 검토 후 제거)
```

### 정리 대상 디렉토리들
```
- docs/backlog/ (tasks/backlog으로 이동됨)
- internal/helpers/ (각 적절한 패키지로 이동)
- 중복된 테스트 fixture들
```

## 📦 패키지 구조 최적화 (1시간)

### 1. helpers 패키지 해체
```bash
# git helper -> internal/git/
mv internal/helpers/git_helper.go internal/git/helpers.go

# platform helper -> internal/platform/
mkdir -p internal/platform
mv internal/helpers/platform_* internal/platform/
```

### 2. 중복 인터페이스 통합
```go
// 여러 패키지에서 중복되는 Logger 인터페이스 통합
// CommonLogger 인터페이스로 표준화
```

## 🔒 보안 및 성능 수정 (30분)

### 1. HTTP 서버 Timeout 설정
```go
// internal/profiling/profiler.go:364
p.server = &http.Server{
    Addr:         fmt.Sprintf(":%d", p.config.HTTPPort),
    Handler:      mux,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

### 2. Integer Overflow 수정
```go
// 안전한 타입 변환 함수 사용
func safeUint64ToInt64(val uint64) int64 {
    if val > math.MaxInt64 {
        return math.MaxInt64
    }
    return int64(val)
}
```

## 📈 즉시 얻을 수 있는 효과

### 정량적 개선
- **빌드 시간**: 10-15% 단축
- **린트 에러**: 80개 → 20개 이하로 감소
- **코드 라인**: 1,000-1,500줄 감소
- **파일 수**: 50-100개 파일 감소

### 정성적 개선
- **가독성**: 불필요한 코드 제거
- **유지보수성**: 중복 제거로 수정 포인트 감소
- **개발자 경험**: 명확한 구조
- **CI/CD**: 빌드 및 테스트 속도 향상

## 🚨 주의사항

### 백업 필수
```bash
# 작업 전 브랜치 생성
git checkout -b cleanup-immediate-fixes
git add .
git commit -m "Before immediate cleanup"
```

### 테스트 실행
```bash
# 각 수정 후 테스트
make test
make test-integration
make lint
```

### 점진적 적용
1. 한 번에 모든 변경사항 적용하지 말고
2. 카테고리별로 나누어 진행
3. 각 단계별로 커밋 생성
4. PR 리뷰 후 머지

## 📋 체크리스트

- [ ] 백업 브랜치 생성
- [ ] 자동 포맷팅 실행
- [ ] JSON 태그 수정
- [ ] 사용하지 않는 매개변수 정리
- [ ] 중복 파일 제거
- [ ] helpers 패키지 해체
- [ ] 보안 이슈 수정
- [ ] 전체 테스트 실행
- [ ] 린트 검사 통과
- [ ] PR 생성 및 리뷰 요청
