# TODO: synclone 고급 설정 생성 및 상태 관리 구현

- status: [ ]
- priority: medium
- category: synclone
- estimated_effort: 3-4 days
- depends_on: []
- spec_reference: `/specs/synclone.md` lines 140-227

## 📋 작업 개요

synclone의 고급 설정 생성 기능과 완전한 상태 관리 시스템을 구현하여 사용자 편의성과 운영 안정성을 크게 향상시킵니다.

## 🎯 구현 목표

### 고급 설정 생성 기능
- [ ] `gz synclone config generate discover` - 기존 저장소에서 설정 자동 생성
- [ ] `gz synclone config generate template` - 템플릿 기반 설정 생성
- [ ] `gz synclone config generate github` - GitHub 조직 기반 설정 생성
- [ ] 설정 마이그레이션 및 업그레이드 도구

### 완전한 Resume 및 상태 관리
- [ ] 중단된 작업의 정확한 재개 기능
- [ ] 상태 분석 및 정리 도구
- [ ] 작업 이력 및 성능 메트릭
- [ ] 자동 정리 및 최적화 기능

## 🔧 기술적 요구사항

### 1. 설정 생성: Discover 기능

#### 명령어 구조
```bash
gz synclone config generate discover --path ~/repos    # 디렉토리 스캔
gz synclone config generate discover --path ~/repos --output config.yaml
gz synclone config generate discover --path ~/repos --merge-existing
gz synclone config generate discover --recursive --depth 3
```

#### 구현 세부사항
```go
type RepoDiscoverer struct {
    BasePath     string
    MaxDepth     int
    IgnorePatterns []string
    FollowSymlinks bool
}

type DiscoveredRepo struct {
    Path        string `yaml:"path"`
    RemoteURL   string `yaml:"remote_url"`
    Provider    string `yaml:"provider"`
    Org         string `yaml:"org"`
    RepoName    string `yaml:"repo_name"`
    Branch      string `yaml:"branch"`
    LastCommit  string `yaml:"last_commit"`
    Size        int64  `yaml:"size_bytes"`
}

func (rd *RepoDiscoverer) DiscoverRepos() ([]DiscoveredRepo, error) {
    // 디렉토리 재귀 탐색
    // .git 디렉토리 감지
    // remote URL 파싱하여 provider/org/repo 추출
    // 브랜치 및 커밋 정보 수집
}
```

### 2. 설정 생성: Template 기능

#### 템플릿 시스템
```bash
gz synclone config generate template --template enterprise
gz synclone config generate template --template minimal
gz synclone config generate template --template multi-org
gz synclone config generate template --list-templates
```

#### 템플릿 정의
```yaml
# templates/enterprise.yaml
name: "Enterprise Configuration"
description: "Multi-organization setup with security and compliance features"
template:
  version: "1.0.0"
  global:
    clone_base_dir: "${HOME}/enterprise-repos"
    default_strategy: reset
    concurrency:
      clone_workers: 5
      update_workers: 10
  
  providers:
    github:
      organizations:
        - name: "{{.CompanyOrg}}"
          clone_dir: "${HOME}/enterprise-repos/{{.CompanyOrg}}"
          visibility: private
          exclude:
            - ".*-archive$"
            - ".*-deprecated$"
          auth:
            token: "${GITHUB_ENTERPRISE_TOKEN}"
    
  sync_mode:
    cleanup_orphans: true
    conflict_resolution: "remote-overwrite"

variables:
  - name: "CompanyOrg"
    description: "Your company's GitHub organization name"
    required: true
    type: "string"
```

### 3. 완전한 Resume 기능

#### 상태 추적 개선
```go
type OperationState struct {
    ID            string                 `json:"id"`
    StartTime     time.Time             `json:"start_time"`
    LastUpdate    time.Time             `json:"last_update"`
    Status        OperationStatus       `json:"status"`
    Config        *Config               `json:"config"`
    Progress      OperationProgress     `json:"progress"`
    Repositories  map[string]RepoState  `json:"repositories"`
    Errors        []OperationError      `json:"errors"`
    Metrics       OperationMetrics      `json:"metrics"`
}

type RepoState struct {
    Name         string    `json:"name"`
    Status       string    `json:"status"` // pending, cloning, completed, failed
    AttemptCount int       `json:"attempt_count"`
    LastError    string    `json:"last_error,omitempty"`
    StartTime    time.Time `json:"start_time,omitempty"`
    EndTime      time.Time `json:"end_time,omitempty"`
    BytesCloned  int64     `json:"bytes_cloned"`
}
```

#### 지능적 Resume 로직
```go
func (r *ResumableCloner) ResumeOperation(stateID string) error {
    // 상태 파일 로드 및 검증
    state, err := r.LoadState(stateID)
    if err != nil {
        return fmt.Errorf("failed to load state: %w", err)
    }
    
    // 환경 변화 감지 (네트워크, 크리덴셜 등)
    if err := r.ValidateEnvironment(state); err != nil {
        return fmt.Errorf("environment validation failed: %w", err)
    }
    
    // 부분 완료된 저장소 상태 확인
    pendingRepos := r.IdentifyPendingRepos(state)
    
    // 실패한 저장소 재시도 전략 결정
    retryRepos := r.CalculateRetryStrategy(state)
    
    // Resume 실행
    return r.ExecuteResume(pendingRepos, retryRepos, state)
}
```

### 4. 고급 상태 관리

#### 상태 분석 도구
```bash
gz synclone state analyze <state-id>        # 상태 분석
gz synclone state analyze --all             # 모든 상태 분석
gz synclone state optimize                  # 상태 파일 최적화
gz synclone state repair <state-id>         # 손상된 상태 복구
```

#### 자동 정리 시스템
```go
type StateManager struct {
    StateDir     string
    RetentionPolicy RetentionPolicy
}

type RetentionPolicy struct {
    MaxAge          time.Duration
    MaxCompletedOps int
    MaxFailedOps    int
    AutoCleanup     bool
}

func (sm *StateManager) RunCleanup() error {
    // 오래된 상태 파일 정리
    // 중복된 상태 파일 병합
    // 손상된 파일 복구 또는 삭제
    // 메트릭 업데이트
}
```

## 📁 파일 구조

### 새로 생성할 파일
- `cmd/synclone/config_generate_discover.go` - Discover 기능
- `cmd/synclone/config_generate_template.go` - Template 기능
- `internal/synclone/discovery/repo_discoverer.go` - 저장소 자동 발견
- `internal/synclone/template/template_engine.go` - 템플릿 엔진
- `internal/synclone/template/builtin_templates.go` - 내장 템플릿
- `internal/synclone/state/advanced_manager.go` - 고급 상태 관리
- `internal/synclone/state/resume_engine.go` - Resume 엔진
- `internal/synclone/state/analyzer.go` - 상태 분석기
- `pkg/synclone/templates/` - 템플릿 디렉토리

### 수정할 파일
- `cmd/synclone/config_generate.go` - discover, template 명령어 추가
- `cmd/synclone/synclone_state.go` - 고급 상태 관리 명령어 추가

## 🧪 테스트 요구사항

### 단위 테스트
- [ ] Repository discovery 로직 테스트
- [ ] 템플릿 엔진 및 변수 치환 테스트
- [ ] Resume 로직 및 상태 복구 테스트
- [ ] 상태 분석 및 정리 기능 테스트

### 통합 테스트
- [ ] 실제 저장소 디렉토리에서 discovery 테스트
- [ ] 다양한 템플릿 생성 시나리오 테스트
- [ ] 복잡한 resume 시나리오 테스트

### E2E 테스트
- [ ] 중단 후 재개 전체 워크플로우 테스트
- [ ] 대용량 조직 clone 중단/재개 테스트

## 📊 완료 기준

### 기능 완성도
- [ ] 모든 고급 설정 생성 기능 구현
- [ ] 완전한 resume 기능 동작
- [ ] 상태 분석 및 정리 도구 완성

### 신뢰성
- [ ] 중단/재개 과정에서 데이터 무결성 보장
- [ ] 네트워크 오류, 권한 오류 등 다양한 실패 상황 처리
- [ ] 상태 파일 손상 복구 기능

### 성능
- [ ] 대용량 저장소 발견 성능 최적화
- [ ] Resume 시 불필요한 재작업 최소화

## 🔗 관련 작업

이 작업은 기존 synclone 기능을 확장하므로 독립적으로 진행 가능합니다.

## 💡 구현 힌트

1. **점진적 발견**: 대용량 디렉토리 스캔 시 점진적으로 결과 표시
2. **템플릿 상속**: 기본 템플릿을 상속하는 사용자 정의 템플릿 지원
3. **상태 압축**: 오래된 상태 파일의 압축 저장으로 공간 절약
4. **병렬 검증**: Resume 시 저장소 상태 병렬 검증으로 속도 향상

## ⚠️ 주의사항

- 대용량 디렉토리 스캔 시 시스템 부하 고려
- 상태 파일의 하위 호환성 유지
- Resume 과정에서의 부분 실패 상황 처리
- 템플릿 보안 (사용자 입력 검증 및 제한)