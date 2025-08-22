# Phase 4: net-env 패키지 리팩토링 실행 계획

## 개요
**목표**: cmd/net-env 패키지의 전면적 구조 재편 및 서브패키지화
**소요시간**: 약 6시간
**복잡도**: 가장 높음
**우선순위**: 4순위 (가장 복잡하므로 마지막)

## 현재 상태 분석

### 현재 파일 현황 (43개 파일)
```
cmd/net-env/
├── actions.go                    # 네트워크 액션
├── actions_test.go               # 액션 테스트
├── cloud.go                      # 클라우드 네트워킹
├── cloud_test.go                 # 클라우드 테스트
├── container_detection.go        # 컨테이너 감지
├── container_detection_cmd.go    # 컨테이너 감지 명령
├── container_detection_test.go   # 컨테이너 감지 테스트
├── docker_container_network_test.go # Docker 네트워크 테스트
├── docker_network.go             # Docker 네트워킹
├── docker_network_cmd.go         # Docker 네트워크 명령
├── docker_network_test.go        # Docker 네트워크 테스트
├── kubernetes_network.go         # K8s 네트워킹
├── kubernetes_network_cmd.go     # K8s 네트워크 명령
├── kubernetes_network_simple.go  # K8s 간단 네트워크
├── kubernetes_network_test.go    # K8s 네트워크 테스트
├── kubernetes_service_mesh.go    # K8s 서비스 메시
├── kubernetes_service_mesh_cmd.go # K8s 서비스 메시 명령
├── kubernetes_service_mesh_test.go # K8s 서비스 메시 테스트
├── monitor_unified.go            # 통합 모니터링
├── net_env.go                    # 메인 커맨드
├── net_env_test.go               # 메인 테스트
├── network_analysis_cmd.go       # 네트워크 분석 명령
├── network_metrics_cmd.go        # 네트워크 메트릭 명령
├── network_topology.go           # 네트워크 토폴로지
├── network_topology_cmd.go       # 네트워크 토폴로지 명령
├── network_topology_test.go      # 네트워크 토폴로지 테스트
├── optimal_routing_cmd.go        # 최적 라우팅 명령
├── optimized_managers.go         # 최적화된 매니저
├── performance.go                # 성능 관련
├── performance_test.go           # 성능 테스트
├── profile_unified.go            # 통합 프로필
├── quick_unified.go              # 빠른 통합
├── status.go                     # 상태 확인
├── status_test.go                # 상태 테스트
├── status_unified.go             # 통합 상태
├── switch.go                     # 네트워크 전환
├── switch_test.go                # 전환 테스트
├── switch_unified.go             # 통합 전환
├── tui.go                        # TUI 인터페이스
├── vpn_failover_cmd.go           # VPN 페일오버 명령
├── vpn_hierarchy_cmd.go          # VPN 계층 명령
├── vpn_profile_cmd.go            # VPN 프로필 명령
└── doc.go                        # 패키지 문서
```

### 복잡성 요인
1. **파일 수 많음**: 43개 파일로 가장 많음
2. **환경 의존성**: Docker, Kubernetes, VPN 등 외부 환경에 의존
3. **테스트 환경**: CI에서만 재현되는 실패 가능성
4. **명명 규칙**: unified, cmd 등 일관성 없는 접미사

## 기능별 분류 및 매핑

### 1. actions 그룹
```
actions/
├── actions.go
├── actions_test.go
└── optimized_managers.go      # 액션 관련 매니저
```

### 2. cloud 그룹
```
cloud/
├── cloud.go
└── cloud_test.go
```

### 3. container 그룹
```
container/
├── container_detection.go
├── container_detection_cmd.go
├── container_detection_test.go
├── docker_network.go
├── docker_network_cmd.go
├── docker_network_test.go
├── docker_container_network_test.go
├── kubernetes_network.go
├── kubernetes_network_cmd.go
├── kubernetes_network_simple.go
├── kubernetes_network_test.go
├── kubernetes_service_mesh.go
├── kubernetes_service_mesh_cmd.go
└── kubernetes_service_mesh_test.go
```

### 4. profile 그룹
```
profile/
├── profile_unified.go
└── quick_unified.go           # 빠른 프로필 관련
```

### 5. status 그룹
```
status/
├── status.go
├── status_test.go
└── status_unified.go
```

### 6. switch 그룹
```
switch/
├── switch.go
├── switch_test.go
└── switch_unified.go
```

### 7. vpn 그룹
```
vpn/
├── vpn_failover_cmd.go
├── vpn_hierarchy_cmd.go
└── vpn_profile_cmd.go
```

### 8. analysis 그룹
```
analysis/
├── network_analysis_cmd.go
├── network_topology.go
├── network_topology_cmd.go
├── network_topology_test.go
├── optimal_routing_cmd.go
├── performance.go
└── performance_test.go
```

### 9. metrics 그룹
```
metrics/
├── network_metrics_cmd.go
└── monitor_unified.go
```

### 10. tui 그룹
```
tui/
└── tui.go
```

### 11. 루트 유지
```
net_env.go         # 메인 커맨드 조립
net_env_test.go    # 메인 테스트
doc.go             # 패키지 문서
```

## 실행 계획

### 1단계: 파일 매핑 및 의존성 분석 (60분)

#### 공용 요소 식별
```bash
# 공용 타입/구조체 확인
grep -r "type.*struct" cmd/net-env/ | head -20

# 공용 함수 확인
grep -r "func [A-Z]" cmd/net-env/ | head -20

# 플래그/옵션 공유 확인
grep -r "flag\|Flag" cmd/net-env/ | head -10

# 로거 사용 현황
grep -r "logger\|log\." cmd/net-env/ | head -10
```

#### 의존성 매트릭스 작성
각 파일이 다른 파일의 어떤 함수/타입을 사용하는지 매핑

### 2단계: internal/netenv 생성 (60분)

#### 공용 컴포넌트 추출
```bash
mkdir -p internal/netenv
```

```go
// internal/netenv/types.go
package netenv

// 공용 구조체들
type NetworkConfig struct {
    // 공용 네트워크 설정
}

type CommonOptions struct {
    // 공용 옵션들
}

// internal/netenv/logger.go
package netenv

// 공용 로거 설정

// internal/netenv/utils.go
package netenv

// 공용 유틸리티 함수들

// internal/netenv/flags.go
package netenv

// 공용 플래그 정의
```

### 3단계: 서브패키지 단계별 생성 (180분)

#### 3.1 actions 패키지 (20분)
```bash
mkdir -p cmd/net-env/actions
mv cmd/net-env/actions.go cmd/net-env/actions/
mv cmd/net-env/actions_test.go cmd/net-env/actions/
mv cmd/net-env/optimized_managers.go cmd/net-env/actions/
```

```go
// cmd/net-env/actions/actions.go
package actions

import "github.com/Gizzahub/gzh-cli/internal/netenv"

func NewCmd() *cobra.Command {
    // actions 관련 커맨드 조립
}
```

#### 3.2 cloud 패키지 (15분)
```bash
mkdir -p cmd/net-env/cloud
mv cmd/net-env/cloud.go cmd/net-env/cloud/
mv cmd/net-env/cloud_test.go cmd/net-env/cloud/
```

#### 3.3 container 패키지 (40분) - 가장 복잡
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

#### 3.4 나머지 패키지들 (105분)
각 그룹별로 15분씩 배정하여 순차적으로 이동

### 4단계: 루트 커맨드 조립 수정 (45분)

#### net_env.go 대폭 수정
```go
// cmd/net-env/net_env.go
package netenv

import (
    "github.com/spf13/cobra"
    "github.com/Gizzahub/gzh-cli/cmd/net-env/actions"
    "github.com/Gizzahub/gzh-cli/cmd/net-env/cloud"
    "github.com/Gizzahub/gzh-cli/cmd/net-env/container"
    // ... 기타 서브패키지들
    "github.com/Gizzahub/gzh-cli/internal/netenv"
)

func NewNetEnvCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "net-env",
        Short: "Network environment management",
        // ...
    }

    // 서브커맨드들 추가
    cmd.AddCommand(actions.NewCmd())
    cmd.AddCommand(cloud.NewCmd())
    cmd.AddCommand(container.NewCmd())
    // ... 기타 서브커맨드들

    return cmd
}
```

### 5단계: 점진적 빌드 및 검증 (90분)

#### 단계별 빌드 검증
```bash
# 1. internal 패키지 빌드
go build ./internal/netenv

# 2. 각 서브패키지별 빌드
go build ./cmd/net-env/actions
go build ./cmd/net-env/cloud
go build ./cmd/net-env/container
# ... 기타

# 3. 전체 net-env 빌드
go build ./cmd/net-env

# 4. 전체 프로젝트 빌드
go build ./...
```

#### 의존성 에러 수정
각 단계에서 발생하는 import 에러, 타입 에러 등을 수정

### 6단계: 기능 테스트 (60분)

#### 기본 명령어 테스트
```bash
# 메인 도움말
./gz net-env --help

# 각 서브커맨드 도움말 (환경에 따라 스킵될 수 있음)
./gz net-env actions --help
./gz net-env cloud --help
./gz net-env status --help
./gz net-env switch --help
```

#### 안전한 기능 테스트
```bash
# 환경 의존성이 적은 명령들 위주
./gz net-env status --dry-run
./gz net-env profile list
```

### 7단계: 테스트 스위트 (45분)

#### 단위 테스트
```bash
# internal 테스트
go test ./internal/netenv -v

# 각 서브패키지 테스트 (일부는 스킵될 수 있음)
go test ./cmd/net-env/actions -v
go test ./cmd/net-env/cloud -v
go test ./cmd/net-env/status -v

# 전체 net-env 테스트
go test ./cmd/net-env/... -v
```

#### 환경 의존성 테스트 처리
Docker, Kubernetes 등이 없는 환경에서는 테스트가 스킵되도록 확인

### 8단계: 최종 정리 및 커밋 (30분)

#### 최종 구조 확인
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
├── actions/                     # 새로 생성
│   ├── actions.go
│   ├── actions_test.go
│   └── optimized_managers.go
├── cloud/                       # 새로 생성
│   ├── cloud.go
│   └── cloud_test.go
├── container/                   # 새로 생성 (가장 큰 그룹)
│   ├── container_detection.go
│   ├── docker_network.go
│   ├── kubernetes_network.go
│   └── ... (14개 파일)
├── profile/                     # 새로 생성
│   ├── profile_unified.go
│   └── quick_unified.go
├── status/                      # 새로 생성
│   ├── status.go
│   ├── status_test.go
│   └── status_unified.go
├── switch/                      # 새로 생성
│   ├── switch.go
│   ├── switch_test.go
│   └── switch_unified.go
├── vpn/                         # 새로 생성
│   ├── vpn_failover_cmd.go
│   ├── vpn_hierarchy_cmd.go
│   └── vpn_profile_cmd.go
├── analysis/                    # 새로 생성
│   ├── network_analysis_cmd.go
│   ├── network_topology.go
│   └── ... (7개 파일)
├── metrics/                     # 새로 생성
│   ├── network_metrics_cmd.go
│   └── monitor_unified.go
└── tui/                         # 새로 생성
    └── tui.go
```

#### Git 커밋
```bash
git add internal/netenv cmd/net-env/
git commit -m "refactor(net-env): complete restructuring with subpackages

This is the most comprehensive refactoring of the 4 phases:

Phase 1: Extract shared components to internal/netenv
- Move common types, utilities, and configurations
- Establish shared logger and flag handling
- Create foundation for subpackage organization

Phase 2: Create 10 feature-based subpackages
- actions/: Network action management (3 files)
- cloud/: Cloud networking features (2 files)
- container/: Docker/K8s networking (14 files, largest group)
- profile/: Network profile management (2 files)
- status/: Network status checking (3 files)
- switch/: Network switching functionality (3 files)
- vpn/: VPN management commands (3 files)
- analysis/: Network analysis and topology (7 files)
- metrics/: Network metrics and monitoring (2 files)
- tui/: Terminal UI interface (1 file)

Benefits:
- Dramatically improved code navigation (43 → 10 logical groups)
- Clear functional boundaries and responsibilities
- Reduced cognitive load for developers
- Better test organization and isolation
- Reusable internal components

Challenges addressed:
- Environment dependency isolation in tests
- Complex inter-file dependencies resolved
- Consistent command structure across subpackages
- Maintained backward compatibility

Total files reorganized: 43 files → 10 subpackages + internal"
```

## 검증 체크리스트

### 빌드 검증
- [ ] `go build ./internal/netenv` 성공
- [ ] `go build ./cmd/net-env` 성공
- [ ] 모든 서브패키지 빌드 성공
- [ ] `go build ./...` 성공

### 기능 검증 (환경 가능한 범위)
- [ ] `./gz net-env --help` 정상 출력
- [ ] 각 서브커맨드 도움말 정상 출력
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
**증상**: Docker, K8s 없는 환경에서 테스트 실패
**해결**: `testing.Short()` 또는 환경 변수로 스킵 처리

### 문제 2: 복잡한 의존성 체인
**증상**: 파일들 간의 복잡한 상호 참조로 빌드 실패
**해결**: 단계적 이동, interface를 통한 의존성 역전

### 문제 3: 명령어 구조 불일치
**증상**: 서브패키지마다 다른 명령어 생성 패턴
**해결**: 공통 인터페이스 정의 후 일관된 패턴 적용

### 문제 4: 과도한 abstraction
**증상**: internal 패키지가 너무 복잡해짐
**해결**: 정말 공용인 것만 추출, 나머지는 각 패키지에 유지

## 롤백 계획

### 전체 롤백
```bash
# 모든 변경사항 되돌리기 (위험도 높음)
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
문제가 있는 서브패키지는 원래 위치로 되돌리고, 성공한 부분만 유지

## 성공 기준
1. **가독성 혁신**: 43개 파일 → 10개 논리 그룹으로 획기적 개선
2. **기능 보존**: 모든 net-env 명령어 정상 동작 (환경 허용 범위)
3. **테스트 안정성**: 환경 의존성 적절히 처리된 안정적 테스트 스위트
4. **확장성**: 새로운 네트워킹 기능 추가 시 명확한 위치 제공
5. **유지보수성**: 개발자가 원하는 기능을 빠르게 찾을 수 있음

## 프로젝트 완료

Phase 4 완료 시 전체 리팩토링 프로젝트 완료:
- ✅ Phase 1: PM 패키지 리팩토링
- ✅ Phase 2: repo-config 패키지 리팩토링
- ✅ Phase 3: IDE 패키지 리팩토링
- ✅ Phase 4: net-env 패키지 리팩토링

**최종 결과**: 4개 주요 패키지의 코드 구조 현대화 완료 🎉
