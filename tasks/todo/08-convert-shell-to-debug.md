# Task: Convert shell Command to Hidden Debug Feature

## Objective
shell 명령어를 제거하고 --debug-shell 플래그나 환경 변수로 활성화되는 숨겨진 디버그 기능으로 전환한다.

## Requirements
- [ ] 현재 shell 명령어 기능 분석
- [ ] 디버그 모드 활성화 방법 설계
- [ ] 일반 사용자에게 노출되지 않도록 처리
- [ ] 개발자를 위한 문서화

## Steps

### 1. Analyze Current shell Command
- [ ] cmd/shell/ 구조 및 기능 분석
- [ ] REPL 기능 범위 파악
- [ ] 현재 사용 사례 확인
- [ ] 디버깅에 유용한 기능 식별

### 2. Design Debug Mode Activation
```bash
# 방법 1: 전역 플래그
gz --debug-shell

# 방법 2: 환경 변수
GZH_DEBUG_SHELL=1 gz

# 방법 3: 숨겨진 명령어 (help에 표시 안됨)
gz debug shell

# 방법 4: 개발 빌드에서만 활성화
# build tag: -tags debug
```

### 3. Implementation Options

#### Option A: Global Flag
```go
// cmd/root.go
var debugShell bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&debugShell, "debug-shell", false, "")
    rootCmd.PersistentFlags().MarkHidden("debug-shell")
}

func Execute() {
    if debugShell || os.Getenv("GZH_DEBUG_SHELL") == "1" {
        runDebugShell()
        return
    }
    // 정상 실행
}
```

#### Option B: Hidden Command
```go
// cmd/debug/debug.go
// +build debug

var debugCmd = &cobra.Command{
    Use:    "debug",
    Hidden: true,
    Short:  "Debug utilities (hidden)",
}

var shellCmd = &cobra.Command{
    Use:   "shell",
    Short: "Start interactive debug shell",
    Run:   runDebugShell,
}
```

### 4. Debug Shell Features
- [ ] 현재 설정 검사
- [ ] 내부 상태 조회
- [ ] 명령어 직접 실행
- [ ] 로그 레벨 동적 변경
- [ ] 성능 프로파일링 시작/중지

### 5. Enhanced Debug Capabilities
```go
// pkg/debug/shell.go
type DebugShell struct {
    commands map[string]DebugCommand
}

func (s *DebugShell) RegisterCommands() {
    s.Register("config", ShowConfig)
    s.Register("env", ShowEnvironment)
    s.Register("cache", InspectCache)
    s.Register("profile", StartProfiling)
    s.Register("trace", EnableTracing)
    s.Register("exec", ExecuteInternal)
}
```

### 6. Security Considerations
- [ ] 프로덕션 빌드에서 제외 옵션
- [ ] 민감한 정보 노출 방지
- [ ] 실행 권한 제한
- [ ] 감사 로그 남기기

### 7. Developer Documentation
```markdown
# Debug Shell Usage (Internal)

## Activation
- Development: `go run -tags debug main.go debug shell`
- Binary: `GZH_DEBUG_SHELL=1 ./gz`
- Flag: `./gz --debug-shell` (hidden flag)

## Available Commands
- `config dump` - Show all configurations
- `env` - Display environment variables
- `cache clear` - Clear internal caches
- `profile cpu start/stop` - CPU profiling
- `trace on/off` - Enable/disable tracing

## Security
- Never enable in production builds
- Requires explicit activation
- No sensitive data exposure
```

### 8. Build Configuration
```makefile
# Makefile
.PHONY: build-debug
build-debug:
	go build -tags debug -o gz-debug ./cmd/gz

.PHONY: build
build:
	go build -o gz ./cmd/gz
```

## Expected Output
- 업데이트된 `cmd/root.go` (디버그 플래그 추가)
- `pkg/debug/shell.go` (리팩토링된 shell 로직)
- `docs/development/debug-shell.md` (개발자 문서)
- 업데이트된 빌드 설정

## Verification Criteria
- [ ] 일반 사용자는 shell 명령어를 볼 수 없음
- [ ] 개발자는 디버그 모드 활성화 가능
- [ ] 프로덕션 빌드에 디버그 코드 미포함
- [ ] 디버그 기능이 정상 작동
- [ ] 보안 취약점 없음

## Notes
- 디버그 모드는 개발 목적으로만 사용
- 사용자 facing 문서에는 포함하지 않음
- CI/CD 빌드는 디버그 태그 제외
- 필요시 별도 디버그 바이너리 제공