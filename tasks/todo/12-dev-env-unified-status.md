# TODO: dev-env 통합 상태 표시 구현

- status: [x]
- priority: high
- category: dev-env
- estimated_effort: 2-3 days
- depends_on: []
- spec_reference: `/specs/dev-env.md` lines 69-86

## 📋 작업 개요

`gz dev-env status` 명령어를 구현하여 모든 개발 환경 서비스의 현재 상태를 통합적으로 표시하는 기능을 제공합니다.

## 🎯 구현 목표

### 핵심 기능
- [ ] 모든 서비스 상태 통합 표시 (AWS, GCP, Azure, Docker, Kubernetes, SSH)
- [ ] 컬러 코딩된 상태 인디케이터
- [ ] 크리덴셜 만료 경고 시스템
- [ ] 서비스별 상태 검증 (health check)
- [ ] 다양한 출력 형식 지원 (table, json, yaml)

### 상태 정보 항목
- [ ] 현재 활성 프로필/컨텍스트
- [ ] 크리덴셜 상태 및 만료 시간
- [ ] 서비스 연결 상태
- [ ] 권한 및 접근성 확인
- [ ] 마지막 사용 시간

## 🔧 기술적 요구사항

### 명령어 구조
```bash
gz dev-env status                    # 모든 서비스 상태 표시
gz dev-env status --service aws      # 특정 서비스만 표시
gz dev-env status --format json      # JSON 형식으로 출력
gz dev-env status --check-health     # 상세한 헬스 체크 포함
gz dev-env status --watch           # 실시간 상태 갱신
```

### 출력 예시
```
Development Environment Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Service    │ Status      │ Current              │ Credentials    │ Last Used
───────────┼─────────────┼──────────────────────┼────────────────┼───────────
AWS        │ ✅ Active   │ prod-profile (us-w-2) │ ⚠️ Expires 2h   │ 5 min ago
GCP        │ ✅ Active   │ my-prod-project      │ ✅ Valid (30d)  │ 1 hour ago
Azure      │ ❌ Inactive │ -                    │ ❌ Expired     │ 2 days ago
Docker     │ ✅ Active   │ prod-context         │ -              │ 10 min ago
Kubernetes │ ✅ Active   │ prod-cluster/default │ ✅ Valid       │ 5 min ago
SSH        │ ✅ Active   │ production           │ ✅ Key loaded  │ 30 min ago

Health Status: ⚠️ Warning (Azure credentials expired)
Active Environments: 5/6
```

### 구현 세부사항

#### 1. 서비스 상태 인터페이스
```go
type ServiceStatus struct {
    Name           string            `json:"name"`
    Status         StatusType        `json:"status"`
    Current        CurrentConfig     `json:"current"`
    Credentials    CredentialStatus  `json:"credentials"`
    LastUsed       time.Time         `json:"last_used"`
    HealthCheck    HealthStatus      `json:"health_check,omitempty"`
    Details        map[string]string `json:"details,omitempty"`
}

type StatusType string
const (
    StatusActive   StatusType = "active"
    StatusInactive StatusType = "inactive"
    StatusError    StatusType = "error"
    StatusUnknown  StatusType = "unknown"
)

type CredentialStatus struct {
    Valid      bool      `json:"valid"`
    ExpiresAt  time.Time `json:"expires_at,omitempty"`
    Type       string    `json:"type"`
    Warning    string    `json:"warning,omitempty"`
}
```

#### 2. 서비스별 상태 체크 구현
```go
type ServiceChecker interface {
    Name() string
    CheckStatus(ctx context.Context) (*ServiceStatus, error)
    CheckHealth(ctx context.Context) (*HealthStatus, error)
}

// AWS 상태 체크
func (a *AWSChecker) CheckStatus(ctx context.Context) (*ServiceStatus, error) {
    // 현재 프로필 확인
    // 크리덴셜 유효성 검사
    // STS GetCallerIdentity 호출로 접근성 확인
    // 세션 토큰 만료 시간 확인
}
```

#### 3. 통합 상태 수집기
```go
type StatusCollector struct {
    checkers []ServiceChecker
    timeout  time.Duration
}

func (sc *StatusCollector) CollectAll(ctx context.Context, options StatusOptions) ([]ServiceStatus, error) {
    // 병렬로 모든 서비스 상태 수집
    // 타임아웃 처리
    // 에러 상황에서도 가능한 정보 수집
}
```

#### 4. 출력 포맷터
```go
type StatusFormatter interface {
    Format(statuses []ServiceStatus) (string, error)
}

type TableFormatter struct{}
type JSONFormatter struct{}
type YAMLFormatter struct{}
```

## 📁 파일 구조

### 새로 생성할 파일
- `cmd/dev-env/status.go` - 메인 status 명령어
- `internal/devenv/status/collector.go` - 상태 수집 로직
- `internal/devenv/status/checker.go` - 서비스별 체크 인터페이스
- `internal/devenv/status/aws_checker.go` - AWS 상태 체크
- `internal/devenv/status/gcp_checker.go` - GCP 상태 체크
- `internal/devenv/status/azure_checker.go` - Azure 상태 체크
- `internal/devenv/status/docker_checker.go` - Docker 상태 체크
- `internal/devenv/status/k8s_checker.go` - Kubernetes 상태 체크
- `internal/devenv/status/ssh_checker.go` - SSH 상태 체크
- `internal/devenv/status/formatter.go` - 출력 포맷터

### 수정할 파일
- `cmd/dev-env/dev_env.go` - status 명령어 추가

## 🧪 테스트 요구사항

### 단위 테스트
- [ ] 각 서비스별 상태 체크 로직 테스트
- [ ] 크리덴셜 만료 감지 테스트
- [ ] 출력 포맷터 테스트
- [ ] 병렬 상태 수집 테스트

### 통합 테스트
- [ ] 모든 서비스 상태 수집 통합 테스트
- [ ] 타임아웃 및 에러 처리 테스트
- [ ] 다양한 출력 형식 검증

### E2E 테스트
- [ ] 실제 서비스 연동 상태 확인 (토큰 필요)
- [ ] Watch 모드 동작 검증

## 📊 완료 기준

### 기능 완성도
- [ ] 모든 서비스 상태 정확히 표시
- [ ] 크리덴셜 만료 경고 정상 동작
- [ ] 모든 출력 형식 지원
- [ ] Watch 모드 실시간 갱신

### 성능 요구사항
- [ ] 전체 상태 수집 시간 5초 이내
- [ ] 병렬 처리로 효율성 확보
- [ ] 네트워크 오류 시 적절한 타임아웃

### 사용자 경험
- [ ] 직관적인 상태 표시
- [ ] 컬러 코딩으로 가독성 향상
- [ ] 명확한 경고 메시지

## 🔗 관련 작업

이 작업은 다음 TODO와 연관됩니다:
- `11-dev-env-switch-all-command.md` - switch-all 실행 전 상태 확인
- `15-dev-env-tui-dashboard.md` - TUI에서 상태 정보 표시

## 💡 구현 힌트

1. **기존 개별 명령어 활용**: 각 서비스의 기존 상태 확인 로직 재사용
2. **캐싱 전략**: 빈번한 상태 체크를 위한 적절한 캐싱 구현
3. **비동기 처리**: 서비스별 상태 수집을 goroutine으로 병렬 처리
4. **에러 처리**: 일부 서비스 실패 시에도 다른 서비스 정보 표시

## ⚠️ 주의사항

- API rate limiting을 고려한 적절한 간격으로 상태 체크
- 크리덴셜 정보를 로그나 출력에 노출하지 않도록 주의
- 네트워크 연결이 불안정한 환경에서의 동작 고려
- 서비스별 특성에 맞는 상태 판단 기준 적용