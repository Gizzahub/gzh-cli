# AGENTS.md - dev-env (클라우드 환경 관리)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**dev-env**는 다중 클라우드 환경(GCP, AWS, Azure)과 개발 도구 환경을 관리하는 복잡한 모듈입니다.

### 핵심 기능
- 클라우드 환경 관리 (GCP 프로젝트, AWS 프로필, Azure 구독)
- 서비스 계정과 인증 키 관리
- 개발 환경 전환 및 상태 동기화
- 복잡한 TUI 인터페이스
- SSH 키 및 Kubeconfig 관리

## 🔒 보안 고려사항 (Critical)

### 1. 인증 정보 보호
```go
// ✅ 안전한 인증 정보 처리
func (m *GCPProjectManager) switchProject(projectID string) error {
    // 로그에 민감 정보 출력 금지
    logger.Info("Switching to project", "project_id", projectID) // ✅
    logger.Debug("Auth token: " + token) // ❌ 절대 금지

    // 메모리에서 즉시 정리
    defer func() {
        token = ""
        serviceAccountKey = nil
    }()
}
```

### 2. 서비스 계정 키 관리
```go
// ✅ 안전한 키 저장소 사용
keyPath := filepath.Join(homeDir, ".config", "gzh", "keys")
if err := os.Chmod(keyPath, 0700); err != nil { // 소유자만 접근
    return fmt.Errorf("failed to set key directory permissions: %w", err)
}
```
- **권한 최소화**: 필요한 최소 권한으로 서비스 계정 생성
- **키 순환**: 주기적인 서비스 계정 키 갱신 알림
- **암호화 저장**: 민감한 설정은 암호화하여 저장

## ⚠️ 개발 시 핵심 주의사항

### 1. 환경 전환 안전성
```go
// ✅ 안전한 환경 전환
func (s *ServiceSwitcher) SwitchEnvironment(target string) error {
    // 현재 상태 백업
    if err := s.backupCurrentState(); err != nil {
        return fmt.Errorf("failed to backup current state: %w", err)
    }

    // 롤백 가능한 전환
    if err := s.applyNewEnvironment(target); err != nil {
        s.rollbackToPreviousState() // 실패 시 자동 롤백
        return fmt.Errorf("environment switch failed: %w", err)
    }
}
```

### 2. 클라우드 API 오류 처리
```go
// ✅ 클라우드별 오류 처리
switch cloudProvider {
case "gcp":
    if isQuotaExceeded(err) {
        return errors.WithSuggestion(err, "GCP 할당량 초과. 프로젝트 관리자에게 문의하세요.")
    }
case "aws":
    if isCredentialsExpired(err) {
        return errors.WithSuggestion(err, "AWS 자격증명 만료. `aws sso login` 실행이 필요합니다.")
    }
}
```

### 3. TUI 상태 관리
```go
// ✅ TUI 안정성
type TUIState struct {
    mutex       sync.RWMutex
    currentView string
    data        map[string]interface{}

    // 에러 복구를 위한 상태 백업
    previousState *TUIState
}

func (t *TUIState) SafeUpdate(fn func() error) error {
    t.mutex.Lock()
    defer t.mutex.Unlock()

    // 상태 백업
    backup := t.clone()

    if err := fn(); err != nil {
        t.restore(backup) // 에러 시 복구
        return err
    }
    return nil
}
```

## 🧪 테스트 요구사항

### 클라우드 환경별 테스트
```bash
# GCP 환경 테스트
GCP_PROJECT_ID=test-project go test ./cmd/dev-env -run TestGCP

# AWS 환경 테스트
AWS_PROFILE=test-profile go test ./cmd/dev-env -run TestAWS

# Azure 환경 테스트
AZURE_SUBSCRIPTION_ID=test-sub go test ./cmd/dev-env -run TestAzure
```

### 필수 테스트 시나리오
- **인증 실패**: 잘못된 자격증명 처리
- **네트워크 장애**: 클라우드 API 연결 실패
- **권한 부족**: 부적절한 권한으로 접근 시도
- **환경 충돌**: 동시에 여러 환경 전환 시도
- **TUI 크래시**: 예상치 못한 터미널 크기 변경

## 📊 모니터링 요구사항

### 환경 상태 추적
- **활성 클라우드 환경**: 현재 사용 중인 프로젝트/계정
- **인증 상태**: 토큰 만료 시간 추적
- **리소스 사용량**: 클라우드 리소스 비용 모니터링 알림
- **보안 이벤트**: 비정상적인 접근 패턴 감지

### 성능 메트릭
- **환경 전환 시간**: 프로젝트/계정 전환 소요 시간
- **API 응답 시간**: 클라우드 서비스별 지연 시간
- **TUI 반응성**: 사용자 입력에 대한 응답 시간

## 🔧 디버깅 가이드

### 환경 전환 문제 진단
```bash
# 현재 환경 상태 확인
gz dev-env status --verbose

# 환경 전환 드라이런
gz dev-env switch gcp-dev --dry-run

# 인증 상태 검증
gz dev-env validate --all-providers
```

### 일반적인 문제와 해결방안
1. **GCP 프로젝트 접근 불가**: `gcloud auth list`로 인증 상태 확인
2. **AWS 프로필 오류**: `aws configure list-profiles` 확인
3. **Kubeconfig 충돌**: 백업된 설정 파일 복구
4. **TUI 깨짐**: 터미널 크기 및 인코딩 확인

## 🚨 위험 상황 대응

### 환경 오염 방지
- **프로덕션 환경 보호**: 프로덕션 리소스 접근 시 추가 확인
- **자동 백업**: 중요한 설정 변경 전 자동 백업
- **변경 추적**: 환경 변경 이력 로그 유지

**핵심**: dev-env는 프로덕션 리소스에 직접 영향을 줄 수 있으므로, 모든 변경사항은 보안과 안정성을 최우선으로 검토해야 합니다.
