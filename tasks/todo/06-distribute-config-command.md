# Task: Distribute Generic config Command to Specific Commands

## Objective
범용 config 명령어를 제거하고 각 명령어에 자체 config 서브커맨드를 추가하여 설정 관리를 명확하게 한다.

## Requirements
- [ ] 현재 config 명령어의 기능 분석
- [ ] 각 명령어별 필요한 설정 기능 식별
- [ ] 일관된 config 서브커맨드 인터페이스 설계
- [ ] 중복 코드 최소화

## Steps

### 1. Analyze Current config Command
- [ ] cmd/config/ 구조 및 기능 분석
- [ ] 어떤 명령어들이 config를 사용하는지 파악
- [ ] 공통 설정 관리 패턴 식별
- [ ] 설정 파일 형식 및 위치 정리

### 2. Design Distributed Config Structure
```bash
# 각 명령어별 config 서브커맨드
gz synclone config init|get|set|list|edit
gz dev-env config init|get|set|list|edit
gz net-env config init|get|set|list|edit
gz repo-sync config init|get|set|list|edit
gz ide config init|get|set|list|edit
gz always-latest config init|get|set|list|edit
gz docker config init|get|set|list|edit
```

### 3. Common Config Interface
```go
// pkg/common/config/interface.go
type ConfigManager interface {
    Init() error
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    List() (map[string]interface{}, error)
    Edit() error
    Validate() error
    Export(writer io.Writer) error
    Import(reader io.Reader) error
}
```

### 4. Implementation Strategy
- [ ] 공통 config 인터페이스 및 base 구현 생성
- [ ] 각 명령어에 config 서브커맨드 추가
- [ ] 명령어별 특화된 설정 로직 구현
- [ ] 설정 파일 위치 표준화

### 5. Config File Organization
```
~/.config/gzh-manager/
├── global.yaml          # 전역 설정
├── synclone.yaml       # synclone 전용 설정
├── dev-env.yaml        # dev-env 전용 설정
├── net-env.yaml        # net-env 전용 설정
├── repo-sync.yaml      # repo-sync 전용 설정
├── ide.yaml            # ide 전용 설정
├── always-latest.yaml  # always-latest 전용 설정
└── docker.yaml         # docker 전용 설정
```

### 6. Code Changes for Each Command
```go
// 예: cmd/synclone/config.go
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Manage synclone configuration",
}

var configInitCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize configuration with defaults",
    Run: func(cmd *cobra.Command, args []string) {
        // synclone 특화 설정 초기화
    },
}

var configGetCmd = &cobra.Command{
    Use:   "get [key]",
    Short: "Get configuration value",
    Run: func(cmd *cobra.Command, args []string) {
        // 설정 값 조회
    },
}
```

### 7. Migration from Central Config
- [ ] 기존 중앙 config 파일 파싱
- [ ] 각 명령어별 설정으로 분리
- [ ] 자동 마이그레이션 스크립트 작성
- [ ] 백업 생성

### 8. Shared Configuration Logic
- [ ] pkg/common/config/ 패키지 생성
- [ ] YAML/JSON 파싱 공통 로직
- [ ] 환경 변수 오버라이드 로직
- [ ] 설정 검증 프레임워크

## Expected Output
- `pkg/common/config/` - 공통 설정 관리 로직
- 각 명령어의 `config.go` 파일
- 마이그레이션 스크립트
- 업데이트된 설정 파일 템플릿

## Verification Criteria
- [ ] 각 명령어가 독립적인 config 서브커맨드 보유
- [ ] 설정 파일이 명확하게 분리됨
- [ ] 기존 설정이 올바르게 마이그레이션됨
- [ ] 일관된 UX across all config commands
- [ ] 환경 변수 오버라이드 작동

## Notes
- 전역 설정과 명령어별 설정의 우선순위 명확히
- 설정 파일 버전 관리 고려
- 민감한 정보 (토큰 등) 처리 방안 포함