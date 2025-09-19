# AGENTS.md - net-env (네트워크 환경 관리)

> 📋 **공통 규칙**: [cmd/AGENTS_COMMON.md](../AGENTS_COMMON.md) 참조

## 🎯 모듈 특성

**net-env**는 네트워크 환경 전환, VPN 관리, DNS 설정을 통합하는 복잡한 네트워크 관리 모듈입니다.

### 핵심 기능

- 네트워크 프로필 기반 환경 전환
- VPN 연결 관리 및 자동화
- DNS 서버 동적 변경
- 프록시 설정 관리
- TUI 기반 실시간 대시보드
- 컨테이너 네트워크 관리

## 🌐 개발 시 핵심 주의사항

### 1. 네트워크 상태 안전성

```go
// ✅ 안전한 네트워크 전환
func (p *ProfileManager) SwitchProfile(name string) error {
    // 현재 네트워크 상태 백업
    currentState, err := p.captureCurrentState()
    if err != nil {
        return fmt.Errorf("failed to backup current network state: %w", err)
    }

    // 롤백 가능한 전환
    if err := p.applyProfile(name); err != nil {
        p.restoreState(currentState) // 실패 시 이전 상태로 복구
        return fmt.Errorf("profile switch failed: %w", err)
    }
}
```

### 2. VPN 연결 안정성

```go
// ✅ VPN 연결 상태 모니터링
func (v *VPNManager) ConnectWithMonitoring(profile string) error {
    // 연결 시도
    if err := v.connect(profile); err != nil {
        return err
    }

    // 연결 안정성 확인 (5초간 모니터링)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if !v.isConnected() {
                return fmt.Errorf("VPN connection unstable")
            }
        case <-ctx.Done():
            return nil // 안정적 연결 확인
        }
    }
}
```

### 3. 플랫폼별 네트워크 처리

```go
// ✅ 크로스 플랫폼 네트워크 관리
type NetworkManager interface {
    SetDNS(servers []string) error
    SetProxy(config ProxyConfig) error
    ConnectVPN(profile string) error
}

// Linux 구현
type LinuxNetworkManager struct{}
func (l *LinuxNetworkManager) SetDNS(servers []string) error {
    // systemd-resolved 또는 /etc/resolv.conf 사용
}

// macOS 구현
type MacOSNetworkManager struct{}
func (m *MacOSNetworkManager) SetDNS(servers []string) error {
    // scutil 명령어 사용
}

// Windows 구현
type WindowsNetworkManager struct{}
func (w *WindowsNetworkManager) SetDNS(servers []string) error {
    // netsh 명령어 사용
}
```

### 4. TUI 실시간 업데이트

```go
// ✅ 안전한 TUI 상태 관리
type NetworkTUI struct {
    model     tea.Model
    mutex     sync.RWMutex
    stopChan  chan struct{}
}

func (n *NetworkTUI) StartMonitoring() {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            n.mutex.Lock()
            n.refreshNetworkStatus()
            n.mutex.Unlock()
        case <-n.stopChan:
            return
        }
    }
}
```

## 🧪 테스트 요구사항

### 네트워크 시나리오 테스트

```bash
# VPN 연결 테스트
go test ./cmd/net-env -v -run TestVPNConnection

# DNS 전환 테스트
go test ./cmd/net-env -v -run TestDNSSwitching

# 네트워크 프로필 테스트
go test ./cmd/net-env -v -run TestProfileSwitching

# TUI 기능 테스트
go test ./cmd/net-env -v -run TestTUIFunctionality
```

### 필수 시뮬레이션 테스트

- **네트워크 연결 끊김**: VPN 연결 실패 시나리오
- **DNS 응답 지연**: 느린 DNS 서버 응답 처리
- **프록시 인증 실패**: 프록시 서버 인증 문제
- **권한 부족**: 네트워크 설정 변경 권한 없음

## 🔧 플랫폼별 고려사항

### Linux

- **NetworkManager 연동**: `nmcli` 명령어 활용
- **systemd-resolved**: DNS 설정 관리
- **iptables 규칙**: 방화벽 설정 충돌 방지

### macOS

- **scutil 활용**: DNS 서버 동적 변경
- **keychain 접근**: VPN 자격증명 안전 저장
- **네트워크 서비스 우선순위**: 여러 인터페이스 관리

### Windows

- **netsh 명령어**: 네트워크 어댑터 설정
- **WMI 쿼리**: 네트워크 상태 조회
- **UAC 권한**: 관리자 권한 필요 작업 처리

## 📊 모니터링 메트릭

### 네트워크 성능

```go
type NetworkMetrics struct {
    Latency     time.Duration `json:"latency"`
    Bandwidth   uint64        `json:"bandwidth_mbps"`
    PacketLoss  float64       `json:"packet_loss_percent"`
    DNSLatency  time.Duration `json:"dns_latency"`
    VPNStatus   string        `json:"vpn_status"`
    ActiveProfile string      `json:"active_profile"`
}
```

### 연결 안정성

- **VPN 연결 지속 시간**: 연결 끊김 빈도 추적
- **DNS 응답 시간**: 도메인 해석 성능 모니터링
- **프로필 전환 성공률**: 환경 전환 실패 비율

## 🚨 보안 고려사항

### VPN 자격증명 보호

```go
// ✅ 안전한 자격증명 처리
func (v *VPNCredentials) Store(profile string, creds *Credentials) error {
    // 시스템 keystore 활용
    keyring, err := keyring.Open(keyring.Config{
        ServiceName: "gzh-manager-vpn",
    })
    if err != nil {
        return err
    }

    // 암호화하여 저장
    encrypted, err := v.encrypt(creds)
    if err != nil {
        return err
    }

    return keyring.Set(profile, string(encrypted))
}
```

### DNS 보안

- **DNS over HTTPS**: 안전한 DNS 쿼리 지원
- **DNS 필터링**: 악성 도메인 차단 기능
- **로그 최소화**: DNS 쿼리 로그 보안 고려

## 🔧 디버깅 가이드

### 네트워크 문제 진단

```bash
# 현재 네트워크 상태 확인
gz net-env status --verbose

# TUI 대시보드 실행
gz net-env tui

# 특정 프로필 테스트
gz net-env profile test office --dry-run

# VPN 연결 디버그
gz net-env actions vpn connect office --debug
```

### 일반적인 문제와 해결

1. **VPN 연결 실패**: 자격증명 및 네트워크 연결 확인
1. **DNS 변경 안됨**: 시스템 권한 및 NetworkManager 상태 확인
1. **TUI 응답 없음**: 터미널 호환성 및 권한 확인
1. **프로필 전환 실패**: 설정 파일 구문 오류 검사

**핵심**: net-env는 시스템 네트워크 설정을 직접 변경하므로, 모든 변경사항은 롤백 가능하도록 설계하고 권한 및 보안을 철저히 고려해야 합니다.
