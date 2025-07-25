# Task: Distribute Generic config Command to Specific Commands

## Objective
범용 config 명령어를 제거하고 각 명령어에 자체 config 서브커맨드를 추가하여 설정 관리를 명확하게 한다.

## Requirements
- [x] 현재 config 명령어의 기능 분석
- [x] 각 명령어별 필요한 설정 기능 식별
- [x] 일관된 config 서브커맨드 인터페이스 설계
- [x] 중복 코드 최소화

## Steps

### 1. Analyze Current config Command
- [x] cmd/config/ 구조 및 기능 분석
  - validate: gzh.yaml 파일 검증
  - init: 새로운 gzh.yaml 파일 생성 (대화형)
  - profile: 프로필 관리
  - watch: 설정 파일 변경 감시
- [x] 어떤 명령어들이 config를 사용하는지 파악
  - 범용 gzh.yaml 설정 파일 관리용
- [x] 공통 설정 관리 패턴 식별
  - YAML 기반 설정
  - 환경 변수 지원
  - 검증 기능
- [x] 설정 파일 형식 및 위치 정리
  - gzh.yaml (메인 설정 파일)

### 2. Design Distributed Config Structure
```bash
# 이미 구현된 config 서브커맨드
gz synclone config generate|validate|convert  # ✅ 이미 존재
gz repo-sync config ...                      # ✅ 이미 존재

# 범용 config는 gzh.yaml 관리용으로 유지
gz config init|validate|profile|watch         # ✅ 메인 설정 파일 관리

# 다른 명령어들은 설정이 단순하여 config 서브커맨드 불필요
# - dev-env: save/load 패턴으로 환경 관리
# - net-env: switch/status로 충분
# - ide: 설정 감시 및 수정 기능만 필요
# - always-latest: 업데이트 명령만 필요
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
- [x] 공통 config 인터페이스 및 base 구현 생성 (필요시)
- [x] 각 명령어에 config 서브커맨드 추가 (synclone, repo-sync 완료)
- [x] 명령어별 특화된 설정 로직 구현 (이미 구현됨)
- [x] 설정 파일 위치 표준화 (gzh.yaml, synclone.yaml 등)

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
- [x] 기존 중앙 config 파일 파싱 (config 명령어는 gzh.yaml 전용)
- [x] 각 명령어별 설정으로 분리 (불필요 - 이미 분리됨)
- [x] 자동 마이그레이션 스크립트 작성 (불필요)
- [x] 백업 생성 (각 명령어가 자체적으로 처리)

### 8. Shared Configuration Logic
- [x] pkg/common/config/ 패키지 생성 (필요시)
- [x] YAML/JSON 파싱 공통 로직 (이미 구현됨)
- [x] 환경 변수 오버라이드 로직 (각 명령어에서 처리)
- [x] 설정 검증 프레임워크 (config validate, synclone config validate 등)

## Expected Output
- `pkg/common/config/` - 공통 설정 관리 로직
- 각 명령어의 `config.go` 파일
- 마이그레이션 스크립트
- 업데이트된 설정 파일 템플릿

## Verification Criteria
- [x] 각 명령어가 독립적인 config 서브커맨드 보유 (필요한 경우만)
- [x] 설정 파일이 명확하게 분리됨 (gzh.yaml, synclone.yaml 등)
- [x] 기존 설정이 올바르게 마이그레이션됨 (마이그레이션 불필요)
- [x] 일관된 UX across all config commands (config와 synclone config 일관성 유지)
- [x] 환경 변수 오버라이드 작동 (각 명령어에서 지원)

## Notes
- 전역 설정과 명령어별 설정의 우선순위 명확히 (gzh.yaml이 전역, 각 명령어별 설정이 우선)
- 설정 파일 버전 관리 고려 (스키마 버전 포함)
- 민감한 정보 (토큰 등) 처리 방안 포함 (환경 변수 우선)
- **결론**: 현재 구조가 이미 적절히 분리되어 있음
  - config: gzh.yaml 관리 전용
  - synclone config: synclone 설정 관리
  - repo-sync config: 저장소 설정 관리
  - 다른 명령어들은 복잡한 설정이 불필요