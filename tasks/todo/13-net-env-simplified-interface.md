# TODO: net-env 간소화된 인터페이스 구현

- status: [x]
- priority: high
- category: net-env
- estimated_effort: 4-5 days
- depends_on: []
- spec_reference: `/specs/net-env.md` lines 12-23

## 📋 작업 개요

net-env 명령어의 복잡한 레거시 구조를 5개 핵심 명령어로 간소화하고, 기존 기능들을 새로운 사용자 친화적 인터페이스로 통합합니다.

## 🎯 구현 목표

### 새로운 핵심 명령어 구조
- [ ] `gz net-env status` - 네트워크 상태 통합 표시
- [ ] `gz net-env switch` - 네트워크 프로필 전환  
- [ ] `gz net-env profile` - 네트워크 프로필 관리
- [ ] `gz net-env quick` - 빠른 네트워크 작업
- [ ] `gz net-env monitor` - 네트워크 모니터링

### 레거시 명령어 통합
- [ ] 기존 40개+ 명령어를 새로운 5개 구조로 매핑
- [ ] 하위 호환성 유지 (기존 명령어는 deprecated로 표시)
- [ ] 복잡한 기능을 직관적인 옵션으로 단순화

## 🔧 기술적 요구사항

### 1. 네트워크 상태 (`gz net-env status`)

#### 명령어 구조
```bash
gz net-env status                  # 현재 네트워크 상태 표시
gz net-env status --verbose       # 상세 네트워크 정보
gz net-env status --json          # JSON 형식 출력
gz net-env status --health        # 헬스 체크 포함
gz net-env status --watch         # 실시간 상태 업데이트
```

#### 출력 예시
```
Network Environment Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Profile: office (auto-detected)
Network: Corporate WiFi (5GHz, -45 dBm)
Security: WPA2-Enterprise

Components:
  WiFi      ✓ Connected     Corporate WiFi
  VPN       ✓ Active        corp-vpn (15ms)
  DNS       ✓ Custom        10.0.0.1, 10.0.0.2
  Proxy     ✓ Enabled       proxy.corp.com:8080
  Docker    ✓ Configured    office context

Network Health: Excellent
Last Profile Switch: 2 hours ago
```

### 2. 네트워크 프로필 전환 (`gz net-env switch`)

#### 명령어 구조
```bash
gz net-env switch                 # 자동 감지 및 프로필 제안
gz net-env switch office          # office 프로필로 전환
gz net-env switch --interactive   # 대화형 프로필 선택
gz net-env switch --list         # 사용 가능한 프로필 목록
gz net-env switch --preview office # 변경 사항 미리보기
gz net-env switch --last         # 마지막 사용 프로필로 전환
```

### 3. 프로필 관리 (`gz net-env profile`)

#### 명령어 구조
```bash
gz net-env profile list           # 프로필 목록
gz net-env profile create home    # 새 프로필 생성
gz net-env profile edit office    # 프로필 편집
gz net-env profile delete old     # 프로필 삭제
gz net-env profile export office  # 프로필 내보내기
gz net-env profile import file.yaml # 프로필 가져오기
```

### 4. 빠른 작업 (`gz net-env quick`)

#### 명령어 구조
```bash
gz net-env quick vpn on           # VPN 빠른 연결
gz net-env quick vpn off          # VPN 빠른 해제
gz net-env quick dns reset        # DNS 초기화
gz net-env quick proxy toggle     # 프록시 토글
gz net-env quick wifi scan        # WiFi 스캔
```

### 5. 네트워크 모니터링 (`gz net-env monitor`)

#### 명령어 구조
```bash
gz net-env monitor                # 실시간 네트워크 모니터링
gz net-env monitor --changes      # 네트워크 변경 감지
gz net-env monitor --performance  # 성능 모니터링
gz net-env monitor --log file.log # 로그 파일 저장
```

## 📁 구현 구조

### 새로 생성할 파일
- `cmd/net-env/status_unified.go` - 통합 상태 표시
- `cmd/net-env/switch_unified.go` - 통합 프로필 전환
- `cmd/net-env/profile_unified.go` - 프로필 관리
- `cmd/net-env/quick_unified.go` - 빠른 작업
- `cmd/net-env/monitor_unified.go` - 모니터링
- `internal/netenv/profile_manager.go` - 프로필 관리 로직
- `internal/netenv/network_detector.go` - 네트워크 자동 감지
- `internal/netenv/component_manager.go` - 네트워크 컴포넌트 관리

### 수정할 파일
- `cmd/net-env/net_env.go` - 새로운 명령어 구조 적용

### 레거시 명령어 매핑

#### 기존 → 새 명령어 매핑 테이블
```go
var legacyCommandMapping = map[string]string{
    // 상태 관련
    "actions":               "status",
    "container-detection":   "status --verbose",
    "network-topology":      "status --topology",
    
    // 전환 관련
    "switch":               "switch",
    
    // VPN 관련
    "vpn-hierarchy":        "quick vpn",
    "vpn-profile":          "profile",
    "vpn-failover":         "quick vpn failover",
    
    // 모니터링 관련
    "network-metrics":      "monitor --performance",
    "network-analysis":     "monitor --analysis",
    "optimal-routing":      "monitor --routing",
    
    // Docker/Kubernetes
    "docker-network":       "profile docker",
    "kubernetes-network":   "profile kubernetes",
}
```

## 🔄 레거시 호환성 처리

### Deprecated 경고 시스템
```go
func showDeprecationWarning(oldCmd, newCmd string) {
    fmt.Printf("⚠️  Warning: 'gz net-env %s' is deprecated. Use 'gz net-env %s' instead.\n", oldCmd, newCmd)
    fmt.Printf("   The old command will be removed in a future version.\n\n")
}
```

### 명령어 위임 패턴
```go
func newLegacyActionsCmd() *cobra.Command {
    return &cobra.Command{
        Use:        "actions",
        Short:     "Legacy network actions (deprecated)",
        Hidden:    true,
        RunE: func(cmd *cobra.Command, args []string) error {
            showDeprecationWarning("actions", "status")
            // 새로운 status 명령어로 위임
            return newStatusUnifiedCmd().Execute()
        },
    }
}
```

## 🧪 테스트 요구사항

### 단위 테스트
- [ ] 각 새 명령어별 단위 테스트
- [ ] 프로필 관리 로직 테스트
- [ ] 네트워크 자동 감지 테스트
- [ ] 레거시 명령어 매핑 테스트

### 통합 테스트
- [ ] 전체 워크플로우 테스트
- [ ] 프로필 전환 시나리오 테스트
- [ ] 레거시 호환성 테스트

### E2E 테스트
- [ ] 실제 네트워크 환경에서 테스트
- [ ] 다양한 네트워크 프로필 전환 테스트

## 📊 완료 기준

### 기능 완성도
- [ ] 5개 핵심 명령어 모두 구현
- [ ] 모든 레거시 기능이 새 인터페이스로 매핑
- [ ] 프로필 기반 네트워크 관리 완전 구현

### 사용자 경험
- [ ] 직관적인 명령어 구조
- [ ] 명확한 도움말 및 예제
- [ ] 부드러운 레거시 마이그레이션 경험

### 성능
- [ ] 상태 확인 응답 시간 3초 이내
- [ ] 프로필 전환 시간 5초 이내

## 🔗 관련 작업

이 작업은 다음 TODO와 연관됩니다:
- `16-net-env-tui-dashboard.md` - TUI에서 간소화된 인터페이스 활용

## 💡 구현 힌트

1. **점진적 마이그레이션**: 기존 명령어는 유지하면서 새 명령어 추가
2. **자동 감지 로직**: WiFi SSID, IP 대역 등을 기반으로 환경 자동 감지
3. **프로필 템플릿**: 일반적인 네트워크 환경 템플릿 제공
4. **설정 마이그레이션**: 기존 복잡한 설정을 새 프로필 형식으로 자동 변환

## ⚠️ 주의사항

- 기존 사용자의 워크플로우 중단 최소화
- 복잡한 레거시 기능의 정확한 매핑 보장
- 네트워크 권한 및 보안 설정 주의
- 플랫폼별 네트워크 API 차이점 고려