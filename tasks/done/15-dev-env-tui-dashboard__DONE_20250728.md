# TODO: dev-env TUI 대시보드 구현

- status: [ ]
- priority: medium
- category: dev-env
- estimated_effort: 5-6 days
- depends_on: ["11-dev-env-switch-all-command.md", "12-dev-env-unified-status.md"]
- spec_reference: `/specs/dev-env.md` lines 34-49

## 📋 작업 개요

`gz dev-env` 명령어로 실행되는 대화형 터미널 UI를 구현하여 모든 개발 환경을 시각적으로 관리할 수 있는 대시보드를 제공합니다.

## 🎯 구현 목표

### 핵심 TUI 기능
- [ ] 실시간 서비스 상태 대시보드
- [ ] 계층적 메뉴 네비게이션 (Service → Action → Target)
- [ ] 키보드 단축키로 빠른 작업 수행
- [ ] 검색 및 필터링 기능
- [ ] 컬러풀하고 직관적인 인터페이스

### 주요 화면 구성
- [ ] 메인 대시보드 - 모든 서비스 상태 개요
- [ ] 서비스별 상세 화면 - AWS, GCP, Azure, Docker, K8s 등
- [ ] 환경 전환 화면 - switch-all 기능 통합
- [ ] 설정 화면 - 환경 프로필 관리
- [ ] 로그 화면 - 작업 이력 및 에러 로그

## 🔧 기술적 요구사항

### TUI 라이브러리 선택
- **bubbletea** (Charm 에코시스템) 사용 권장
- **lipgloss** for 스타일링
- **bubbles** for UI 컴포넌트

### 메인 대시보드 레이아웃
```
┌─ GZH Development Environment Manager ─────────────────────────────────┐
│ Current Environment: production                     Updated: 14:35:22  │
├───────────────────────────────────────────────────────────────────────┤
│ Service    │ Status      │ Current              │ Credentials    │ ⚡  │
├────────────┼─────────────┼──────────────────────┼────────────────┼────┤
│ AWS        │ ✅ Active   │ prod-profile (us-w-2) │ ⚠️ Expires 2h   │ →  │
│ GCP        │ ✅ Active   │ my-prod-project      │ ✅ Valid (30d)  │ →  │
│ Azure      │ ❌ Inactive │ -                    │ ❌ Expired     │ →  │
│ Docker     │ ✅ Active   │ prod-context         │ -              │ →  │
│ Kubernetes │ ✅ Active   │ prod-cluster/default │ ✅ Valid       │ →  │
│ SSH        │ ✅ Active   │ production           │ ✅ Key loaded  │ →  │
├───────────────────────────────────────────────────────────────────────┤
│ Quick Actions:                                                        │
│ [1] Switch Environment  [2] Refresh Status  [3] View Logs  [q] Quit   │
│ [s] Search  [f] Filter  [h] Help  [Enter] Service Details             │
└───────────────────────────────────────────────────────────────────────┘
```

### 키보드 인터랙션
```go
type KeyMap struct {
    Up         key.Binding
    Down       key.Binding
    Left       key.Binding
    Right      key.Binding
    Enter      key.Binding
    Back       key.Binding
    Quit       key.Binding
    Help       key.Binding
    Refresh    key.Binding
    Search     key.Binding
    Filter     key.Binding
    SwitchEnv  key.Binding
    QuickAction key.Binding
}

var DefaultKeyMap = KeyMap{
    Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
    Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
    Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
    Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
    SwitchEnv:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "switch env")),
    Refresh:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
}
```

### TUI 애플리케이션 구조
```go
type Model struct {
    state        AppState
    keymap       KeyMap
    help         help.Model
    table        table.Model
    statusBar    StatusBar
    currentView  ViewType
    services     []ServiceStatus
    lastUpdate   time.Time

    // 각 화면별 모델
    dashboardModel  *DashboardModel
    serviceModel    *ServiceDetailModel
    switchModel     *EnvironmentSwitchModel
    settingsModel   *SettingsModel
    logsModel       *LogsModel
}

type AppState int
const (
    StateLoading AppState = iota
    StateDashboard
    StateServiceDetail
    StateEnvironmentSwitch
    StateSettings
    StateLogs
    StateError
)
```

### 실시간 상태 업데이트
```go
type StatusUpdater struct {
    interval     time.Duration
    statusChan   chan []ServiceStatus
    errorChan    chan error
    stopChan     chan struct{}
}

func (su *StatusUpdater) Start() tea.Cmd {
    return tea.Tick(su.interval, func(t time.Time) tea.Msg {
        return TickMsg{Time: t}
    })
}

// 백그라운드에서 주기적으로 상태 업데이트
func (su *StatusUpdater) updateStatus() tea.Cmd {
    return func() tea.Msg {
        statuses, err := collectAllServiceStatus()
        if err != nil {
            return ErrorMsg{Error: err}
        }
        return StatusUpdateMsg{Statuses: statuses}
    }
}
```

## 📁 파일 구조

### 새로 생성할 파일
- `cmd/dev-env/tui.go` - TUI 명령어 엔트리 포인트
- `internal/devenv/tui/model.go` - 메인 TUI 모델
- `internal/devenv/tui/dashboard.go` - 메인 대시보드 화면
- `internal/devenv/tui/service_detail.go` - 서비스 상세 화면
- `internal/devenv/tui/environment_switch.go` - 환경 전환 화면
- `internal/devenv/tui/settings.go` - 설정 화면
- `internal/devenv/tui/logs.go` - 로그 화면
- `internal/devenv/tui/components/` - 재사용 가능한 TUI 컴포넌트
- `internal/devenv/tui/styles.go` - 스타일 정의
- `internal/devenv/tui/keymap.go` - 키보드 단축키
- `internal/devenv/tui/messages.go` - TUI 메시지 타입

### 수정할 파일
- `cmd/dev-env/dev_env.go` - TUI 모드 추가
- `go.mod` - bubbletea 관련 의존성 추가

## 🎨 UI/UX 설계

### 컬러 스킴
```go
var (
    ColorPrimary    = lipgloss.Color("#00ADD8")  // Go 블루
    ColorSecondary  = lipgloss.Color("#5E81AC")  // 차분한 블루
    ColorSuccess    = lipgloss.Color("#A3BE8C")  // 녹색
    ColorWarning    = lipgloss.Color("#EBCB8B")  // 노란색
    ColorError      = lipgloss.Color("#BF616A")  // 빨간색
    ColorText       = lipgloss.Color("#D8DEE9")  // 밝은 회색
    ColorSubtle     = lipgloss.Color("#4C566A")  // 어두운 회색
)

var (
    StyleTitle = lipgloss.NewStyle().
        Foreground(ColorPrimary).
        Bold(true).
        Padding(0, 1)

    StyleStatus = lipgloss.NewStyle().
        Padding(0, 1).
        Margin(0, 1)

    StyleActive = StyleStatus.Copy().
        Foreground(ColorSuccess)

    StyleInactive = StyleStatus.Copy().
        Foreground(ColorSubtle)
)
```

### 애니메이션 및 전환
- [ ] 화면 전환 시 부드러운 슬라이드 애니메이션
- [ ] 로딩 상태에서 스피너 애니메이션
- [ ] 상태 변경 시 페이드 인/아웃 효과

## 🧪 테스트 요구사항

### 단위 테스트
- [ ] 각 TUI 컴포넌트 로직 테스트
- [ ] 키보드 인터랙션 처리 테스트
- [ ] 상태 업데이트 로직 테스트

### 통합 테스트
- [ ] 전체 TUI 플로우 테스트
- [ ] 실시간 업데이트 기능 테스트

### 사용성 테스트
- [ ] 다양한 터미널 크기에서 레이아웃 테스트
- [ ] 컬러/모노크롬 환경 호환성 테스트
- [ ] 키보드 접근성 테스트

## 📊 완료 기준

### 기능 완성도
- [ ] 모든 주요 화면 구현 및 네비게이션
- [ ] 실시간 상태 업데이트 정상 동작
- [ ] switch-all 기능 TUI 통합
- [ ] 설정 관리 및 프로필 편집 기능

### 사용자 경험
- [ ] 직관적인 네비게이션 및 키보드 단축키
- [ ] 반응형 레이아웃 (다양한 터미널 크기 지원)
- [ ] 명확한 도움말 및 가이드
- [ ] 부드러운 애니메이션 및 피드백

### 성능
- [ ] 상태 업데이트 지연 시간 1초 이내
- [ ] 대용량 로그 표시 시 메모리 효율성
- [ ] 부드러운 스크롤링 및 인터랙션

## 🔗 관련 작업

이 작업은 다음 TODO에 의존합니다:
- `11-dev-env-switch-all-command.md` - switch-all 기능을 TUI에 통합
- `12-dev-env-unified-status.md` - 상태 정보를 TUI에 표시

## 💡 구현 힌트

1. **점진적 개발**: 먼저 기본 대시보드만 구현 후 점진적으로 기능 추가
2. **컴포넌트 재사용**: 공통 UI 컴포넌트를 만들어 일관성 유지
3. **상태 관리**: Elm Architecture 패턴으로 상태 관리 단순화
4. **성능 최적화**: 불필요한 렌더링 최소화 및 지연 로딩 활용

## ⚠️ 주의사항

- 터미널 크기 변경 시 레이아웃 깨짐 방지
- 다양한 터미널 에뮬레이터 호환성 확인
- 컬러 지원하지 않는 환경에서의 대체 표시
- 메모리 사용량 모니터링 (장시간 실행 시)
