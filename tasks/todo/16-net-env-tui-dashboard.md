# TODO: net-env TUI 대시보드 구현

- status: [ ]
- priority: medium
- category: net-env
- estimated_effort: 4-5 days
- depends_on: ["13-net-env-simplified-interface.md"]
- spec_reference: `/specs/net-env.md` lines 39-66

## 📋 작업 개요

`gz net-env` 명령어로 실행되는 대화형 네트워크 환경 관리 TUI를 구현하여 네트워크 설정을 시각적으로 관리할 수 있는 대시보드를 제공합니다.

## 🎯 구현 목표

### 핵심 TUI 기능
- [ ] 실시간 네트워크 상태 시각화
- [ ] 빠른 프로필 전환 인터페이스
- [ ] VPN 연결 관리 및 모니터링
- [ ] DNS 및 프록시 설정 관리
- [ ] 네트워크 헬스 모니터링
- [ ] 키보드 단축키로 빠른 네트워크 작업

### TUI 화면 구성
- [ ] 메인 대시보드 - 네트워크 상태 개요
- [ ] 프로필 관리 화면
- [ ] VPN 연결 상태 화면
- [ ] 네트워크 모니터링 화면
- [ ] 설정 화면

## 🔧 기술적 요구사항

### 메인 대시보드 레이아웃
```
┌─ GZH Network Environment Manager ─────────────────────────────────────────┐
│ Current Profile: office                     Network: Corporate WiFi        │
├─────────────────────────────────────────────────────────────────────────┤
│ Component    │ Status      │ Details                    │ Health         │
├──────────────┼─────────────┼────────────────────────────┼────────────────┤
│ WiFi         │ ● Connected │ Corporate WiFi (5GHz)      │ Excellent     │
│ VPN          │ ● Active    │ corp-vpn (10.0.0.1)        │ 15ms latency  │
│ DNS          │ ● Custom    │ 10.0.0.1, 10.0.0.2         │ <5ms response │
│ Proxy        │ ● Enabled   │ proxy.corp.com:8080        │ Connected     │
│ Docker       │ ○ Default   │ office context             │ -             │
├─────────────────────────────────────────────────────────────────────────┤
│ Quick Actions: [s]witch [v]pn [d]ns [p]roxy [r]efresh [?]help [Q]uit    │
└─────────────────────────────────────────────────────────────────────────┘
```

### 실시간 네트워크 모니터링
```go
type NetworkMonitor struct {
    interval        time.Duration
    statusChan      chan NetworkStatus
    metricsHistory  []NetworkMetrics
    alerts          []NetworkAlert
}

type NetworkStatus struct {
    WiFi        WiFiStatus    `json:"wifi"`
    VPN         VPNStatus     `json:"vpn"`
    DNS         DNSStatus     `json:"dns"`
    Proxy       ProxyStatus   `json:"proxy"`
    Docker      DockerStatus  `json:"docker"`
    Connectivity ConnectivityStatus `json:"connectivity"`
    Timestamp   time.Time     `json:"timestamp"`
}

type WiFiStatus struct {
    SSID         string  `json:"ssid"`
    SignalStrength int   `json:"signal_strength"`
    Frequency    string  `json:"frequency"`
    Security     string  `json:"security"`
    Connected    bool    `json:"connected"`
}

type VPNStatus struct {
    Name         string        `json:"name"`
    Connected    bool          `json:"connected"`
    ServerIP     string        `json:"server_ip"`
    Latency      time.Duration `json:"latency"`
    BytesUp      int64         `json:"bytes_up"`
    BytesDown    int64         `json:"bytes_down"`
}
```

### 프로필 전환 인터페이스
```
┌─ Switch Network Profile ───────────────────────────────────────────┐
│                                                                    │
│ Available Profiles:                                                │
│                                                                    │
│ > office        Corporate network with VPN and proxy              │
│   home          Home network configuration                        │
│   cafe          Public WiFi with VPN protection                   │
│   mobile        Mobile hotspot configuration                      │
│                                                                    │
│ Profile Details (office):                                         │
│ ┌────────────────────────────────────────────────────────────────┐ │
│ │ WiFi:  Auto-detect Corporate WiFi                             │ │
│ │ VPN:   corp-vpn.company.com                                   │ │
│ │ DNS:   10.0.0.1, 10.0.0.2                                     │ │
│ │ Proxy: proxy.corp.com:8080                                    │ │
│ │ Docker: office context                                        │ │
│ └────────────────────────────────────────────────────────────────┘ │
│                                                                    │
│ [Enter] Apply Profile  [e] Edit  [n] New  [d] Delete  [Esc] Back  │
└────────────────────────────────────────────────────────────────────┘
```

### VPN 관리 화면
```
┌─ VPN Connection Manager ───────────────────────────────────────────┐
│                                                                    │
│ Active Connection:                                                 │
│ ┌────────────────────────────────────────────────────────────────┐ │
│ │ corp-vpn                           ● Connected (00:45:12)      │ │
│ │ Server: vpn.company.com                   Latency: 15ms        │ │
│ │ IP: 10.0.0.100                           Speed: ↑2.1MB ↓5.4MB  │ │
│ └────────────────────────────────────────────────────────────────┘ │
│                                                                    │
│ Available VPN Connections:                                         │
│                                                                    │
│ > corp-vpn      Company VPN (Active)                              │
│   backup-vpn    Backup VPN server                                 │
│   client-vpn    Client network access                             │
│                                                                    │
│ Connection Log:                                                    │
│ ┌────────────────────────────────────────────────────────────────┐ │
│ │ 14:30:15 corp-vpn connected successfully                      │ │
│ │ 14:25:02 Attempting connection to corp-vpn                    │ │
│ │ 14:24:58 backup-vpn disconnected                              │ │
│ └────────────────────────────────────────────────────────────────┘ │
│                                                                    │
│ [c] Connect  [d] Disconnect  [r] Reconnect  [l] Logs  [Esc] Back  │
└────────────────────────────────────────────────────────────────────┘
```

### TUI 애플리케이션 구조
```go
type NetEnvTUIModel struct {
    state           AppState
    keymap          KeyMap
    networkStatus   NetworkStatus
    profiles        []NetworkProfile
    selectedProfile int
    
    // 화면별 모델
    dashboardModel  *DashboardModel
    profileModel    *ProfileModel  
    vpnModel        *VPNModel
    monitorModel    *MonitorModel
    settingsModel   *SettingsModel
    
    // 상태 관리
    monitor         *NetworkMonitor
    lastUpdate      time.Time
    updateInterval  time.Duration
}

type AppState int
const (
    StateDashboard AppState = iota
    StateProfileSwitch
    StateVPNManager
    StateMonitoring
    StateSettings
    StateError
)
```

### 네트워크 자동 감지
```go
type NetworkDetector struct {
    detectionRules []DetectionRule
    profiles       map[string]*NetworkProfile
}

type DetectionRule struct {
    Name        string            `yaml:"name"`
    Conditions  []Condition       `yaml:"conditions"`
    Profile     string            `yaml:"profile"`
    Priority    int               `yaml:"priority"`
}

type Condition struct {
    Type     string `yaml:"type"`     // wifi_ssid, ip_range, gateway
    Value    string `yaml:"value"`
    Operator string `yaml:"operator"` // equals, contains, matches
}

func (nd *NetworkDetector) DetectCurrentProfile() (*NetworkProfile, error) {
    // 현재 네트워크 환경 스캔
    // 규칙 기반 프로필 매칭
    // 가장 높은 우선순위 프로필 반환
}
```

## 📁 파일 구조

### 새로 생성할 파일
- `cmd/net-env/tui.go` - TUI 명령어 엔트리 포인트
- `internal/netenv/tui/model.go` - 메인 TUI 모델
- `internal/netenv/tui/dashboard.go` - 메인 대시보드
- `internal/netenv/tui/profile_switch.go` - 프로필 전환 화면
- `internal/netenv/tui/vpn_manager.go` - VPN 관리 화면
- `internal/netenv/tui/monitor.go` - 네트워크 모니터링 화면
- `internal/netenv/tui/settings.go` - 설정 화면
- `internal/netenv/monitor/network_monitor.go` - 네트워크 모니터링
- `internal/netenv/detector/auto_detector.go` - 자동 감지
- `internal/netenv/tui/styles.go` - TUI 스타일
- `internal/netenv/tui/keymap.go` - 키보드 단축키

### 수정할 파일
- `cmd/net-env/net_env.go` - TUI 모드 추가

## 🎨 UI/UX 설계

### 실시간 상태 업데이트
- [ ] 네트워크 상태 1초마다 갱신
- [ ] VPN 연결 상태 실시간 모니터링
- [ ] 대역폭 사용량 그래프 (간단한 ASCII 차트)
- [ ] 연결 품질 시각적 표시

### 키보드 단축키
```go
var NetEnvKeyMap = KeyMap{
    SwitchProfile: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "switch profile")),
    VPNToggle:     key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "vpn toggle")),
    DNSSettings:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "dns settings")),
    ProxyToggle:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "proxy toggle")),
    Refresh:       key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
    Monitor:       key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "monitor")),
}
```

### 시각적 피드백
- [ ] 연결 상태별 컬러 인디케이터
- [ ] 로딩 상태 스피너
- [ ] 성공/실패 애니메이션
- [ ] 신호 강도 바 그래프

## 🧪 테스트 요구사항

### 단위 테스트
- [ ] 네트워크 상태 감지 로직 테스트
- [ ] 프로필 전환 로직 테스트
- [ ] TUI 컴포넌트 테스트

### 통합 테스트
- [ ] 실제 네트워크 환경에서 감지 테스트
- [ ] VPN 연결/해제 테스트
- [ ] 프로필 자동 전환 테스트

### 사용성 테스트
- [ ] 다양한 터미널 환경 호환성
- [ ] 키보드 접근성 테스트
- [ ] 반응 속도 테스트

## 📊 완료 기준

### 기능 완성도
- [ ] 모든 네트워크 컴포넌트 상태 표시
- [ ] 프로필 기반 네트워크 전환
- [ ] VPN 연결 관리
- [ ] 실시간 모니터링

### 사용자 경험
- [ ] 직관적인 네비게이션
- [ ] 빠른 키보드 단축키
- [ ] 명확한 상태 표시
- [ ] 반응형 레이아웃

### 성능
- [ ] 상태 업데이트 1초 이내
- [ ] 부드러운 화면 전환
- [ ] 메모리 효율적 모니터링

## 🔗 관련 작업

이 작업은 다음 TODO에 의존합니다:
- `13-net-env-simplified-interface.md` - 간소화된 인터페이스를 TUI에 통합

## 💡 구현 힌트

1. **모듈식 설계**: 각 네트워크 컴포넌트를 독립적인 모듈로 구현
2. **캐싱 전략**: 네트워크 상태 정보 캐싱으로 응답 속도 향상
3. **에러 복구**: 네트워크 오류 시 자동 재연결 및 사용자 알림
4. **설정 백업**: 프로필 전환 전 현재 설정 백업

## ⚠️ 주의사항

- 네트워크 권한 요구사항 (관리자 권한 필요한 작업)
- 플랫폼별 네트워크 API 차이점 처리
- VPN 연결 실패 시 graceful degradation
- 민감한 네트워크 정보 보안 처리