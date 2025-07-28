# TODO: dev-env switch-all 통합 환경 스위칭 구현

- status: [ ]
- priority: high
- category: dev-env
- estimated_effort: 3-4 days
- depends_on: []
- spec_reference: `/specs/dev-env.md` lines 51-67

## 📋 작업 개요

`gz dev-env switch-all` 명령어를 구현하여 여러 클라우드 서비스와 개발 환경을 한 번에 전환할 수 있는 통합 환경 스위칭 기능을 제공합니다.

## 🎯 구현 목표

### 핵심 기능
- [x] Atomic 환경 스위칭 (모든 서비스가 성공하거나 모두 롤백)
- [x] 의존성 해결 및 순서 보장 (예: AWS 프로필 → Kubernetes 컨텍스트)
- [x] 실패 시 자동 롤백 기능
- [x] 환경별 설정 파일 지원
- [x] 진행률 추적 및 상세 출력

### 지원 환경 전환
- [x] AWS 계정/프로필/리전
- [x] GCP 프로젝트/계정
- [x] Azure 구독/테넌트  
- [x] Docker 컨텍스트
- [x] Kubernetes 클러스터/네임스페이스
- [x] SSH 설정

## 🔧 기술적 요구사항

### 명령어 구조
```bash
gz dev-env switch-all --env production     # 사전 정의된 환경으로 전환
gz dev-env switch-all --env dev --dry-run  # 미리보기 모드
gz dev-env switch-all --from-file env.yaml # 환경 파일 사용
gz dev-env switch-all --interactive        # 대화형 환경 선택
```

### 환경 설정 파일 구조
```yaml
# ~/.gzh/dev-env/environments/production.yaml
name: production
description: "Production environment configuration"

services:
  aws:
    profile: prod-profile
    region: us-west-2
    account_id: "123456789012"
  
  gcp:
    project: my-prod-project
    account: prod@company.com
    region: us-central1
  
  kubernetes:
    context: prod-cluster
    namespace: default
  
  docker:
    context: prod-docker
  
  ssh:
    config: production

dependencies:
  - aws -> kubernetes  # AWS 프로필 설정 후 Kubernetes 컨텍스트 전환
  - gcp -> kubernetes  # GCP 설정 후 Kubernetes 전환 가능

pre_hooks:
  - command: "echo 'Switching to production environment'"
  
post_hooks:
  - command: "kubectl get nodes"
  - command: "aws sts get-caller-identity"
```

### 구현 세부사항

#### 1. 환경 정의 및 검증
```go
type Environment struct {
    Name         string            `yaml:"name"`
    Description  string            `yaml:"description"`
    Services     map[string]interface{} `yaml:"services"`
    Dependencies []string          `yaml:"dependencies"`
    PreHooks     []Hook           `yaml:"pre_hooks"`
    PostHooks    []Hook           `yaml:"post_hooks"`
}

type Hook struct {
    Command string `yaml:"command"`
    Timeout string `yaml:"timeout"`
    OnError string `yaml:"on_error"` // continue, fail, rollback
}
```

#### 2. 의존성 해결 알고리즘
- [ ] 의존성 그래프 생성 및 순환 참조 검증
- [ ] 토폴로지 정렬을 통한 실행 순서 결정
- [ ] 병렬 실행 가능한 서비스 그룹 식별

#### 3. Atomic 스위칭 구현
- [ ] 현재 상태 백업 (모든 서비스의 현재 설정 저장)
- [ ] 순차적 서비스 전환 및 검증
- [ ] 실패 시 백업 상태로 롤백
- [ ] 부분 성공 상태 처리

#### 4. 진행률 추적
```go
type SwitchProgress struct {
    TotalServices    int
    CompletedServices int
    CurrentService   string
    Status          string
    StartTime       time.Time
    EstimatedEnd    time.Time
}
```

## 📁 파일 구조

### 새로 생성할 파일
- `cmd/dev-env/switch_all.go` - 메인 명령어 구현
- `internal/devenv/environment.go` - 환경 정의 및 관리
- `internal/devenv/switcher.go` - 스위칭 로직 구현
- `internal/devenv/dependency.go` - 의존성 해결
- `internal/devenv/rollback.go` - 롤백 기능
- `pkg/devenv/config.go` - 환경 설정 파일 처리

### 수정할 파일
- `cmd/dev-env/dev_env.go` - switch-all 명령어 추가
- `cmd/dev-env/validate.go` - 환경 검증 로직 확장

## 🧪 테스트 요구사항

### 단위 테스트
- [ ] 환경 설정 파일 파싱 테스트
- [ ] 의존성 해결 알고리즘 테스트
- [ ] 각 서비스별 스위칭 로직 테스트
- [ ] 롤백 기능 테스트

### 통합 테스트
- [ ] 전체 환경 스위칭 시나리오 테스트
- [ ] 실패 및 롤백 시나리오 테스트
- [ ] 의존성이 있는 복잡한 환경 테스트

### E2E 테스트
- [ ] 실제 클라우드 서비스와의 통합 테스트 (토큰 필요)
- [ ] Dry-run 모드 검증

## 📊 완료 기준

### 기능 완성도
- [ ] 모든 명령어 옵션이 스펙과 일치
- [ ] 환경 설정 파일 형식 완전 지원
- [ ] Atomic 스위칭 및 롤백 기능 동작
- [ ] 의존성 해결 정상 작동

### 코드 품질
- [ ] 테스트 커버리지 80% 이상
- [ ] 모든 에러 케이스 처리
- [ ] 명확한 로깅 및 진행률 표시
- [ ] 코드 리뷰 통과

### 문서화
- [ ] 명령어 help 텍스트 완성
- [ ] 환경 설정 파일 예제 제공
- [ ] 사용자 가이드 작성

## 🔗 관련 작업

이 작업은 다음 TODO와 연관됩니다:
- `12-dev-env-unified-status.md` - 통합 상태 표시와 함께 사용
- `15-dev-env-tui-dashboard.md` - TUI에서 switch-all 기능 활용

## 💡 구현 힌트

1. **기존 개별 서비스 명령어 활용**: 각 서비스의 기존 switch 로직을 재사용
2. **상태 저장**: 스위칭 전 현재 상태를 JSON으로 저장하여 롤백 지원
3. **검증 단계**: 실제 전환 전에 모든 설정의 유효성 검증
4. **병렬 처리**: 의존성이 없는 서비스들은 병렬로 처리하여 성능 향상

## ⚠️ 주의사항

- 실패한 상태에서의 부분 롤백 처리 주의
- 네트워크 오류 등으로 인한 중간 실패 상황 고려
- 각 서비스의 rate limiting 및 API 제한 사항 고려
- 크리덴셜 만료 상황에서의 graceful degradation