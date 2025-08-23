# Phase 4: net-env 종합 구조 재편

## 개요
- **목표**: cmd/net-env 패키지의 전면적 구조 재편 및 10개 서브패키지 생성 (43개 파일 → 논리적 그룹핑)
- **우선순위**: HIGH
- **예상 소요시간**: 6시간
- **담당자**: Backend
- **복잡도**: 가장 높음 (환경 의존성 + 대량 파일 이동)

## 선행 작업
- [ ] Phase 3 (IDE internal 추출) 완료
- [ ] refactor-phase4-net-env 브랜치 생성
- [ ] 환경 의존성 테스트 전략 수립

## 세부 작업 목록

### 1. 현재 상태 분석 및 파일 매핑
- [ ] **43개 파일 현황 조사** (`cmd/net-env/`)
  ```bash
  find cmd/net-env/ -name "*.go" | wc -l  # 파일 수 확인
  ls -la cmd/net-env/                     # 전체 파일 목록
  ```
  - 완료 기준: 모든 파일 분류 및 매핑 완료
  - 주의사항: 가장 많은 파일 수, 신중한 분류 필요

- [ ] **공용 요소 식별 및 분석**
  ```bash
  grep -r "type.*struct" cmd/net-env/ | head -20    # 공용 구조체
  grep -r "func [A-Z]" cmd/net-env/ | head -20      # 공용 함수
  grep -r "flag\|Flag" cmd/net-env/ | head -10      # 플래그 공유
  grep -r "logger\|log\." cmd/net-env/ | head -10   # 로거 사용
  ```
  - 완료 기준: internal/netenv 추출 대상 식별 완료
  - 주의사항: 복잡한 상호 의존성 체인 파악

### 2. Git 백업 및 브랜치 준비
- [ ] **백업 지점 생성** (`git tag refactor-phase4-start`)
  - refactor-phase4-net-env 브랜치 생성 및 체크아웃
  - 현재 상태 커밋
  - 완료 기준: 브랜치 및 태그 생성 완료
  - 주의사항: 가장 복잡한 Phase이므로 안전장치 철저히

- [ ] **의존성 매트릭스 작성** (파일별 상호 참조)
  - 각 파일이 다른 파일의 어떤 함수/타입을 사용하는지 매핑
  - 완료 기준: 이동 순서 결정을 위한 의존성 문서화
  - 주의사항: 순환 의존성 가능성 사전 점검

### 3. internal/netenv 생성 및 공용 컴포넌트 추출
- [ ] **internal/netenv 디렉터리 생성**
  ```bash
  mkdir -p internal/netenv
  ```
  - 완료 기준: 디렉터리 생성 완료
  - 주의사항: net-env 전용 internal 패키지

- [ ] **공용 타입 추출** (`internal/netenv/types.go`)
  ```go
  // internal/netenv/types.go
  package netenv

  // NetworkConfig represents network configuration
  type NetworkConfig struct {
      // 공용 네트워크 설정 필드들
  }

  // CommonOptions represents shared command options
  type CommonOptions struct {
      // 공용 옵션 필드들
  }
  ```
  - 완료 기준: 공용 구조체 정의 완료
  - 주의사항: 모든 서브패키지에서 공유되는 타입만 추출

- [ ] **공용 로거 설정 추출** (`internal/netenv/logger.go`)
  ```go
  // internal/netenv/logger.go
  package netenv

  // SetupLogger configures logging for net-env
  func SetupLogger() {
      // 공용 로거 설정 로직
  }
  ```
  - 완료 기준: 로거 설정 통합 완료
  - 주의사항: 모든 서브패키지에서 일관된 로깅

- [ ] **공용 플래그 정의 추출** (`internal/netenv/flags.go`)
  ```go
  // internal/netenv/flags.go  
  package netenv

  // CommonFlags represents shared flags
  type CommonFlags struct {
      // 공용 플래그 필드들
  }
  ```
  - 완료 기준: 공용 플래그 체계 정리
  - 주의사항: 서브커맨드 간 일관성 확보

- [ ] **공용 유틸리티 추출** (`internal/netenv/utils.go`)
  ```go
  // internal/netenv/utils.go
  package netenv

  // 공용 헬퍼 함수들
  ```
  - 완료 기준: 유틸리티 함수 통합 완료
  - 주의사항: 정말 공용인 것만 추출

### 4. 서브패키지 단계별 생성 (10개 그룹)
- [ ] **actions 패키지 생성** (`cmd/net-env/actions/`)
  ```bash
  mkdir -p cmd/net-env/actions
  mv cmd/net-env/actions.go cmd/net-env/actions/
  mv cmd/net-env/actions_test.go cmd/net-env/actions/
  mv cmd/net-env/optimized_managers.go cmd/net-env/actions/
  ```
  - 완료 기준: actions 그룹 파일 이동 완료 (3개 파일)
  - 주의사항: 관련 매니저 파일도 함께 그룹핑

- [ ] **cloud 패키지 생성** (`cmd/net-env/cloud/`)
  ```bash
  mkdir -p cmd/net-env/cloud
  mv cmd/net-env/cloud.go cmd/net-env/cloud/
  mv cmd/net-env/cloud_test.go cmd/net-env/cloud/
  ```
  - 완료 기준: cloud 그룹 파일 이동 완료 (2개 파일)
  - 주의사항: 클라우드 네트워킹 관련 기능

- [ ] **container 패키지 생성 (가장 큰 그룹)** (`cmd/net-env/container/`)
  ```bash
  mkdir -p cmd/net-env/container
  # Docker 관련
  mv cmd/net-env/docker_network*.go cmd/net-env/container/
  mv cmd/net-env/docker_container_network_test.go cmd/net-env/container/
  # Kubernetes 관련
  mv cmd/net-env/kubernetes_*.go cmd/net-env/container/
  # 컨테이너 감지
  mv cmd/net-env/container_detection*.go cmd/net-env/container/
  ```
  - 완료 기준: container 그룹 파일 이동 완료 (약 14개 파일)
  - 주의사항: Docker, Kubernetes, 컨테이너 감지 모두 포함하는 큰 그룹

- [ ] **profile 패키지 생성** (`cmd/net-env/profile/`)
  ```bash
  mkdir -p cmd/net-env/profile
  mv cmd/net-env/profile_unified.go cmd/net-env/profile/
  mv cmd/net-env/quick_unified.go cmd/net-env/profile/
  ```
  - 완료 기준: profile 그룹 파일 이동 완료 (2개 파일)
  - 주의사항: 네트워크 프로필 관리 기능

- [ ] **status 패키지 생성** (`cmd/net-env/status/`)
  ```bash
  mkdir -p cmd/net-env/status
  mv cmd/net-env/status.go cmd/net-env/status/
  mv cmd/net-env/status_test.go cmd/net-env/status/
  mv cmd/net-env/status_unified.go cmd/net-env/status/
  ```
  - 완료 기준: status 그룹 파일 이동 완료 (3개 파일)
  - 주의사항: 네트워크 상태 확인 관련 기능

- [ ] **switch 패키지 생성** (`cmd/net-env/switch/`)
  ```bash
  mkdir -p cmd/net-env/switch
  mv cmd/net-env/switch.go cmd/net-env/switch/
  mv cmd/net-env/switch_test.go cmd/net-env/switch/
  mv cmd/net-env/switch_unified.go cmd/net-env/switch/
  ```
  - 완료 기준: switch 그룹 파일 이동 완료 (3개 파일)
  - 주의사항: 네트워크 전환 기능

- [ ] **vpn 패키지 생성** (`cmd/net-env/vpn/`)
  ```bash
  mkdir -p cmd/net-env/vpn
  mv cmd/net-env/vpn_failover_cmd.go cmd/net-env/vpn/
  mv cmd/net-env/vpn_hierarchy_cmd.go cmd/net-env/vpn/
  mv cmd/net-env/vpn_profile_cmd.go cmd/net-env/vpn/
  ```
  - 완료 기준: vpn 그룹 파일 이동 완료 (3개 파일)
  - 주의사항: VPN 관련 모든 명령어

- [ ] **analysis 패키지 생성** (`cmd/net-env/analysis/`)
  ```bash
  mkdir -p cmd/net-env/analysis
  mv cmd/net-env/network_analysis_cmd.go cmd/net-env/analysis/
  mv cmd/net-env/network_topology.go cmd/net-env/analysis/
  mv cmd/net-env/network_topology_cmd.go cmd/net-env/analysis/
  mv cmd/net-env/network_topology_test.go cmd/net-env/analysis/
  mv cmd/net-env/optimal_routing_cmd.go cmd/net-env/analysis/
  mv cmd/net-env/performance.go cmd/net-env/analysis/
  mv cmd/net-env/performance_test.go cmd/net-env/analysis/
  ```
  - 완료 기준: analysis 그룹 파일 이동 완료 (7개 파일)
  - 주의사항: 네트워크 분석 및 토폴로지, 성능 관련

- [ ] **metrics 패키지 생성** (`cmd/net-env/metrics/`)
  ```bash
  mkdir -p cmd/net-env/metrics
  mv cmd/net-env/network_metrics_cmd.go cmd/net-env/metrics/
  mv cmd/net-env/monitor_unified.go cmd/net-env/metrics/
  ```
  - 완료 기준: metrics 그룹 파일 이동 완료 (2개 파일)
  - 주의사항: 네트워크 메트릭 및 모니터링

- [ ] **tui 패키지 생성** (`cmd/net-env/tui/`)
  ```bash
  mkdir -p cmd/net-env/tui
  mv cmd/net-env/tui.go cmd/net-env/tui/
  ```
  - 완료 기준: tui 그룹 파일 이동 완료 (1개 파일)
  - 주의사항: Terminal UI 인터페이스

### 5. 각 서브패키지의 NewCmd 함수 생성
- [ ] **actions NewCmd 구현** (`cmd/net-env/actions/actions.go`)
  ```go
  // cmd/net-env/actions/actions.go
  package actions

  import "github.com/Gizzahub/gzh-cli/internal/netenv"

  func NewCmd() *cobra.Command {
      // actions 관련 커맨드 조립
  }
  ```
  - 완료 기준: actions 커맨드 생성 함수 구현
  - 주의사항: 기존 로직 누락 없이 이동

- [ ] **cloud NewCmd 구현** (`cmd/net-env/cloud/cloud.go`)
  - 완료 기준: cloud 커맨드 생성 함수 구현
  - 주의사항: internal/netenv 의존성 올바르게 설정

- [ ] **container NewCmd 구현** (`cmd/net-env/container/container.go`)
  - 완료 기준: container 커맨드 생성 함수 구현
  - 주의사항: 가장 복잡한 그룹, Docker/K8s 명령어 통합

- [ ] **나머지 7개 패키지 NewCmd 구현**
  - profile, status, switch, vpn, analysis, metrics, tui
  - 완료 기준: 모든 서브패키지에 NewCmd 함수 구현
  - 주의사항: 일관된 커맨드 생성 패턴 적용

### 6. 루트 커맨드 조립 대폭 수정
- [ ] **net_env.go 전면 개편** (`cmd/net-env/net_env.go`)
  ```go
  // cmd/net-env/net_env.go
  package netenv

  import (
      "github.com/spf13/cobra"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/actions"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/cloud"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/container"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/profile"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/status"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/switch"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/vpn"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/analysis"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/metrics"
      "github.com/Gizzahub/gzh-cli/cmd/net-env/tui"
      "github.com/Gizzahub/gzh-cli/internal/netenv"
  )

  func NewNetEnvCmd() *cobra.Command {
      cmd := &cobra.Command{
          Use:   "net-env",
          Short: "Network environment management",
          // ...
      }

      // 10개 서브커맨드 추가
      cmd.AddCommand(actions.NewCmd())
      cmd.AddCommand(cloud.NewCmd())
      cmd.AddCommand(container.NewCmd())
      cmd.AddCommand(profile.NewCmd())
      cmd.AddCommand(status.NewCmd())
      cmd.AddCommand(switch.NewCmd())
      cmd.AddCommand(vpn.NewCmd())
      cmd.AddCommand(analysis.NewCmd())
      cmd.AddCommand(metrics.NewCmd())
      cmd.AddCommand(tui.NewCmd())

      return cmd
  }
  ```
  - 완료 기준: 10개 서브패키지 조립 완료
  - 주의사항: import 경로 모두 올바르고 누락 없음

### 7. 점진적 빌드 검증 (중간 체크포인트)
- [ ] **internal 패키지 빌드** (`go build ./internal/netenv`)
  - internal 패키지 단독 빌드 성공
  - 완료 기준: netenv 공용 컴포넌트 빌드 성공
  - 주의사항: 순환 의존성 발생 안함

- [ ] **각 서브패키지별 순차 빌드**
  ```bash
  go build ./cmd/net-env/actions      # 간단한 것부터
  go build ./cmd/net-env/cloud
  go build ./cmd/net-env/profile
  go build ./cmd/net-env/status
  go build ./cmd/net-env/switch
  go build ./cmd/net-env/vpn
  go build ./cmd/net-env/analysis
  go build ./cmd/net-env/metrics
  go build ./cmd/net-env/tui
  go build ./cmd/net-env/container    # 가장 복잡한 것은 마지막
  ```
  - 완료 기준: 각 서브패키지 개별 빌드 성공
  - 주의사항: 에러 발생시 해당 패키지 즉시 수정

- [ ] **전체 net-env 빌드** (`go build ./cmd/net-env`)
  - net-env 전체 패키지 빌드 성공
  - 완료 기준: 루트 커맨드 조립 성공
  - 주의사항: 서브패키지 import 에러 없음

- [ ] **전체 프로젝트 빌드** (`go build ./...`)
  - 전체 프로젝트 빌드 성공
  - 완료 기준: 다른 패키지에 영향 없음 확인
  - 주의사항: 전체적인 안정성 보장

### 8. 기능 테스트 (환경 의존성 고려)
- [ ] **메인 도움말 테스트** (`./gz net-env --help`)
  - net-env 기본 도움말 정상 출력
  - 완료 기준: 10개 서브커맨드 모두 표시
  - 주의사항: 누락된 서브커맨드 없음

- [ ] **각 서브커맨드 도움말 테스트**
  ```bash
  ./gz net-env actions --help       # 환경 의존성 낮음
  ./gz net-env cloud --help
  ./gz net-env profile --help
  ./gz net-env status --help
  ./gz net-env switch --help
  ./gz net-env vpn --help          # 환경 의존성 있을 수 있음
  ./gz net-env analysis --help
  ./gz net-env metrics --help
  ./gz net-env tui --help
  ./gz net-env container --help    # Docker/K8s 의존성 있음
  ```
  - 완료 기준: 모든 서브커맨드 도움말 정상 출력
  - 주의사항: 환경 의존적 명령어는 에러 처리 확인

- [ ] **안전한 기능 테스트** (환경 의존성 최소)
  ```bash
  ./gz net-env status --dry-run     # dry-run 모드
  ./gz net-env profile list         # 프로필 목록 (파일 기반)
  ```
  - 완료 기준: 기본적인 상태 명령어 실행 가능
  - 주의사항: 실제 네트워크 변경은 피하기

### 9. 테스트 스위트 실행 (환경 의존성 처리)
- [ ] **internal 패키지 테스트** (`go test ./internal/netenv -v`)
  - netenv 공용 컴포넌트 테스트 통과
  - 완료 기준: internal 패키지 테스트 성공
  - 주의사항: 환경 독립적 테스트 위주

- [ ] **환경 독립적 서브패키지 테스트**
  ```bash
  go test ./cmd/net-env/actions -v
  go test ./cmd/net-env/cloud -v
  go test ./cmd/net-env/status -v
  ```
  - 완료 기준: 환경 의존성 낮은 패키지 테스트 통과
  - 주의사항: Docker, K8s 없는 환경에서도 실행 가능

- [ ] **환경 의존적 테스트 처리**
  ```bash
  go test ./cmd/net-env/container -v    # Docker, K8s 의존성
  go test ./cmd/net-env/vpn -v          # VPN 설정 의존성
  ```
  - 완료 기준: 환경 의존적 테스트는 적절히 스킵 처리
  - 주의사항: `testing.Short()` 또는 환경 변수로 스킵 로직

- [ ] **전체 net-env 테스트** (`go test ./cmd/net-env/... -v`)
  - net-env 전체 테스트 스위트 실행
  - 완료 기준: 실행 가능한 테스트 모두 통과
  - 주의사항: 환경 의존성으로 스킵된 테스트 확인

### 10. 코드 품질 검사
- [ ] **코드 포맷팅** (`make fmt`)
  - gofumpt, gci 포맷팅 실행
  - 완료 기준: 모든 net-env 관련 파일 포맷팅 완료
  - 주의사항: internal과 10개 서브패키지 모두 포함

- [ ] **린팅 검사** (`make lint`)
  - golangci-lint 검사 통과
  - 완료 기준: 린팅 에러 없음
  - 주의사항: 대량 구조 변경으로 인한 새로운 이슈 해결

### 11. 최종 정리 및 커밋
- [ ] **최종 구조 확인** (예상 구조와 비교)
  ```
  internal/netenv/                 # 새로 생성
  ├── types.go                     # 공용 타입
  ├── logger.go                    # 공용 로거
  ├── utils.go                     # 공용 유틸리티
  ├── flags.go                     # 공용 플래그
  └── ...

  cmd/net-env/                     # 대폭 수정
  ├── net_env.go                   # 루트 커맨드 (수정)
  ├── net_env_test.go             # 메인 테스트 (유지)
  ├── doc.go                       # 패키지 문서 (유지)
  ├── actions/                     # (3개 파일)
  ├── cloud/                       # (2개 파일)
  ├── container/                   # (14개 파일, 가장 큰 그룹)
  ├── profile/                     # (2개 파일)
  ├── status/                      # (3개 파일)
  ├── switch/                      # (3개 파일)
  ├── vpn/                         # (3개 파일)
  ├── analysis/                    # (7개 파일)
  ├── metrics/                     # (2개 파일)
  └── tui/                         # (1개 파일)
  ```
  - 완료 기준: 43개 파일이 10개 논리 그룹으로 완벽 분리
  - 주의사항: 모든 파일이 올바른 위치에 있음

- [ ] **Git 커밋** (`refactor(net-env): complete restructuring with subpackages`)
  - 가장 상세한 커밋 메시지 작성 (가장 큰 변경사항)
  - 완료 기준: 커밋 완료 및 phase-4-completed 태그 생성
  - 주의사항: 43개 파일 재구성의 의미와 효과 명시

## 완료 검증 체크리스트

### 빌드 검증
- [ ] `go build ./internal/netenv` 성공
- [ ] `go build ./cmd/net-env` 성공
- [ ] 모든 서브패키지 빌드 성공 (10개)
- [ ] `go build ./...` 성공

### 기능 검증 (환경 허용 범위)
- [ ] `./gz net-env --help` 정상 출력
- [ ] 각 서브커맨드 도움말 정상 출력 (10개)
- [ ] 기본적인 상태 명령어 실행 가능
- [ ] 환경 의존성 명령어는 적절히 에러 처리

### 테스트 검증
- [ ] `go test ./internal/netenv` 성공
- [ ] 환경 독립적 테스트들 모두 통과
- [ ] 환경 의존적 테스트들 적절히 스킵
- [ ] 전체 테스트 스위트 안정성 확보

### 구조 검증
- [ ] 43개 파일이 논리적으로 10개 그룹으로 분리
- [ ] 공용 컴포넌트가 internal/netenv로 추출
- [ ] 각 서브패키지가 독립적으로 빌드 가능
- [ ] 순환 의존성 없음
- [ ] 명령어 구조 일관성 유지

## 예상 문제 및 해결책

### 문제 1: 환경 의존성 테스트 실패
- **증상**: Docker, K8s 없는 환경에서 테스트 실패
- **해결**: `testing.Short()` 또는 환경 변수로 스킵 처리

### 문제 2: 복잡한 의존성 체인
- **증상**: 파일들 간의 복잡한 상호 참조로 빌드 실패
- **해결**: 단계적 이동, interface를 통한 의존성 역전

### 문제 3: 명령어 구조 불일치
- **증상**: 서브패키지마다 다른 명령어 생성 패턴
- **해결**: 공통 인터페이스 정의 후 일관된 패턴 적용

### 문제 4: 과도한 abstraction
- **증상**: internal 패키지가 너무 복잡해짐
- **해결**: 정말 공용인 것만 추출, 나머지는 각 패키지에 유지

## 롤백 계획

### 전체 롤백 (위험도 높음)
```bash
# 모든 변경사항 되돌리기 - 신중히 사용
rm -rf internal/netenv
git checkout -- cmd/net-env/
```

### 단계별 롤백
```bash
# 특정 서브패키지만 롤백
rm -rf cmd/net-env/container
git checkout -- cmd/net-env/container_*.go cmd/net-env/docker_*.go cmd/net-env/kubernetes_*.go

# internal만 롤백
rm -rf internal/netenv
git checkout -- cmd/net-env/net_env.go
```

### 점진적 복구
- 문제가 있는 서브패키지는 원래 위치로 되돌리기
- 성공한 부분만 유지하여 부분적 개선 효과 확보

## 성공 기준
1. **가독성 혁신**: 43개 파일 → 10개 논리 그룹으로 획기적 개선
2. **기능 보존**: 모든 net-env 명령어 정상 동작 (환경 허용 범위)
3. **테스트 안정성**: 환경 의존성 적절히 처리된 안정적 테스트 스위트
4. **확장성**: 새로운 네트워킹 기능 추가 시 명확한 위치 제공
5. **유지보수성**: 개발자가 원하는 기능을 빠르게 찾을 수 있음

## 관련 파일
### 새로 생성
- `internal/netenv/` 패키지 전체
- `cmd/net-env/actions/` (3개 파일)
- `cmd/net-env/cloud/` (2개 파일) 
- `cmd/net-env/container/` (14개 파일)
- `cmd/net-env/profile/` (2개 파일)
- `cmd/net-env/status/` (3개 파일)
- `cmd/net-env/switch/` (3개 파일)
- `cmd/net-env/vpn/` (3개 파일)
- `cmd/net-env/analysis/` (7개 파일)
- `cmd/net-env/metrics/` (2개 파일)
- `cmd/net-env/tui/` (1개 파일)

### 수정됨
- `cmd/net-env/net_env.go` (대폭 수정)

### 유지됨
- `cmd/net-env/net_env_test.go` (메인 테스트)
- `cmd/net-env/doc.go` (패키지 문서)

## 프로젝트 완료
Phase 4 완료 시 **전체 리팩토링 프로젝트 완료** 🎉:
- ✅ Phase 1: PM 패키지 리팩토링 (2시간)
- ✅ Phase 2: repo-config 패키지 리팩토링 (3시간)
- ✅ Phase 3: IDE 패키지 리팩토링 (4시간)
- ✅ Phase 4: net-env 패키지 리팩토링 (6시간)

**최종 결과**: 4개 주요 패키지의 코드 구조 현대화 완료 (총 15시간)

## 다음 단계
Phase 4 완료 후 → 전체 리팩토링 프로젝트 완료 및 성과 정리