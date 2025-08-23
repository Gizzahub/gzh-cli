# Phase 3: IDE internal 추출 및 서브패키지화

## 개요
- **목표**: cmd/ide 패키지의 internal 추출 및 서브패키지 생성으로 구조 현대화
- **우선순위**: HIGH
- **예상 소요시간**: 4시간
- **담당자**: Backend
- **복잡도**: 높음 (internal 추출 + 서브패키지화)

## 선행 작업
- [ ] Phase 2 (repo-config 패키지 리팩토링) 완료
- [ ] refactor-phase3-ide 브랜치 생성
- [ ] 공용 심볼 의존성 분석 완료

## 세부 작업 목록

### 1. 의존성 분석 및 추출 대상 식별
- [ ] **IDE struct 사용 현황 분석** (`cmd/ide/`)
  ```bash
  grep -r "type IDE struct" cmd/ide/
  grep -r "IDE{" cmd/ide/
  ```
  - 완료 기준: IDE 타입 사용 파일 목록 정리
  - 주의사항: 모든 파일에서 사용되는 핵심 타입

- [ ] **IDEDetector interface 사용 분석** (`cmd/ide/`)
  ```bash
  grep -r "IDEDetector" cmd/ide/
  grep -r "NewIDEDetector" cmd/ide/
  ```
  - 완료 기준: IDEDetector 의존성 매핑 완료
  - 주의사항: detector, scan 파일에서 중요 사용

- [ ] **공용 함수 식별** (`cmd/ide/`)
  ```bash
  grep -r "joinStrings" cmd/ide/
  grep -r "func.*(" cmd/ide/ | grep -v "test"
  ```
  - 완료 기준: 공용 헬퍼 함수 목록 정리
  - 주의사항: internal로 이동할 함수들 식별

### 2. Git 백업 및 브랜치 준비
- [ ] **백업 지점 생성** (`git tag refactor-phase3-start`)
  - refactor-phase3-ide 브랜치 생성 및 체크아웃
  - 현재 상태 커밋
  - 완료 기준: 브랜치 및 태그 생성 완료
  - 주의사항: Phase 2 완료 상태에서 시작

- [ ] **현재 상태 문서화**
  ```
  cmd/ide/
  ├── ide.go           # 메인 커맨드 + 공용 타입
  ├── detector.go      # IDE 감지 로직
  ├── detector_test.go # 감지 테스트
  ├── open.go          # IDE 열기 기능
  ├── open_test.go     # 열기 테스트
  ├── scan.go          # IDE 스캔 기능
  ├── status.go        # IDE 상태 확인
  ├── status_test.go   # 상태 테스트
  └── doc.go           # 패키지 문서
  ```
  - 완료 기준: 현재 파일 구조 파악 완료
  - 주의사항: 평면 구조에서 계층 구조로 전환

### 3. internal/idecore 생성 및 공용 타입 추출
- [ ] **internal/idecore 디렉터리 생성**
  ```bash
  mkdir -p internal/idecore
  ```
  - 완료 기준: 디렉터리 생성 완료
  - 주의사항: Go module 내 internal 패키지 규칙 준수

- [ ] **공용 타입 추출** (`internal/idecore/types.go`)
  ```go
  // internal/idecore/types.go
  package idecore

  // IDE represents an IDE installation
  type IDE struct {
      Name            string
      Path            string
      Version         string
      InstallationMethod string
      // ... 기타 필드들
  }

  // IDEDetector interface for IDE detection
  type IDEDetector interface {
      DetectIDEs() ([]IDE, error)
      // ... 기타 메서드들
  }
  ```
  - 완료 기준: IDE 관련 핵심 타입 idecore로 이동
  - 주의사항: 기존 구조체 필드 누락 없이 이동

- [ ] **detector 팩토리 추출** (`internal/idecore/detector.go`)
  ```go
  // internal/idecore/detector.go
  package idecore

  // NewIDEDetector creates a new IDE detector
  func NewIDEDetector() IDEDetector {
      // 기존 구현부 이전
  }
  ```
  - 완료 기준: IDEDetector 팩토리 함수 이동
  - 주의사항: 구현 로직 손실 없이 이동

- [ ] **공용 유틸리티 추출** (`internal/idecore/utils.go`)
  ```go
  // internal/idecore/utils.go
  package idecore

  // joinStrings joins strings with separator
  func joinStrings(strs []string, sep string) string {
      // 기존 구현부 이전
  }
  ```
  - 완료 기준: 공용 헬퍼 함수들 이동
  - 주의사항: 함수 시그니처 변경 없이 이동

### 4. cmd/ide 파일들의 internal 의존성 변경
- [ ] **ide.go import 수정** (`cmd/ide/ide.go`)
  ```go
  import (
      "github.com/Gizzahub/gzh-cli/internal/idecore"
      // 기타 imports
  )
  ```
  - 완료 기준: idecore import 추가
  - 주의사항: 기존 import와 충돌 없음

- [ ] **타입 참조 변경** (모든 관련 파일)
  - `IDE` → `idecore.IDE`
  - `IDEDetector` → `idecore.IDEDetector`  
  - `NewIDEDetector()` → `idecore.NewIDEDetector()`
  - `joinStrings()` → `idecore.joinStrings()`
  - 완료 기준: 모든 파일에서 타입 참조 업데이트
  - 주의사항: 누락된 참조 없음 확인

- [ ] **1차 빌드 테스트** (`go build ./internal/idecore && go build ./cmd/ide`)
  - internal과 cmd/ide 패키지 빌드 성공
  - 완료 기준: 컴파일 에러 없음
  - 주의사항: 순환 참조 발생 안함

### 5. 서브패키지 생성 (open, status, scan)
- [ ] **open 패키지 생성** (`cmd/ide/open/`)
  ```bash
  mkdir -p cmd/ide/open
  mv cmd/ide/open.go cmd/ide/open/
  mv cmd/ide/open_test.go cmd/ide/open/
  ```
  - 완료 기준: open 관련 파일 이동 완료
  - 주의사항: 테스트 파일과 소스 파일 함께 이동

- [ ] **open 패키지 수정** (`cmd/ide/open/open.go`)
  ```go
  // cmd/ide/open/open.go
  package open

  import "github.com/Gizzahub/gzh-cli/internal/idecore"

  // NewCmd creates the open command
  func NewCmd() *cobra.Command {
      // 기존 newOpenCmd 내용을 여기로 이동
  }
  ```
  - 완료 기준: 독립 패키지로 변환 완료
  - 주의사항: NewCmd 함수 export하여 루트에서 호출 가능

- [ ] **status 패키지 생성** (`cmd/ide/status/`)
  ```bash
  mkdir -p cmd/ide/status
  mv cmd/ide/status.go cmd/ide/status/
  mv cmd/ide/status_test.go cmd/ide/status/
  ```
  - 완료 기준: status 관련 파일 이동 완료
  - 주의사항: 테스트 파일과 소스 파일 함께 이동

- [ ] **status 패키지 수정** (`cmd/ide/status/status.go`)
  ```go
  // cmd/ide/status/status.go
  package status

  import "github.com/Gizzahub/gzh-cli/internal/idecore"

  // NewCmd creates the status command
  func NewCmd() *cobra.Command {
      // 기존 newStatusCmd 내용을 여기로 이동
  }
  ```
  - 완료 기준: 독립 패키지로 변환 완료
  - 주의사항: idecore 의존성 올바르게 설정

- [ ] **scan 패키지 생성** (`cmd/ide/scan/`)
  ```bash
  mkdir -p cmd/ide/scan
  mv cmd/ide/scan.go cmd/ide/scan/
  ```
  - 완료 기준: scan.go 이동 완료
  - 주의사항: scan 기능은 테스트 파일이 없을 수 있음

- [ ] **scan 패키지 수정** (`cmd/ide/scan/scan.go`)
  ```go
  // cmd/ide/scan/scan.go
  package scan

  import "github.com/Gizzahub/gzh-cli/internal/idecore"

  // NewCmd creates the scan command
  func NewCmd() *cobra.Command {
      // 기존 newScanCmd 내용을 여기로 이동
  }
  ```
  - 완료 기준: 독립 패키지로 변환 완료
  - 주의사항: idecore 의존성 올바르게 설정

### 6. 루트 커맨드 조립 수정
- [ ] **ide.go 대폭 수정** (`cmd/ide/ide.go`)
  ```go
  // cmd/ide/ide.go
  package ide

  import (
      "github.com/spf13/cobra"
      "github.com/Gizzahub/gzh-cli/cmd/ide/open"
      "github.com/Gizzahub/gzh-cli/cmd/ide/scan"
      "github.com/Gizzahub/gzh-cli/cmd/ide/status"
      "github.com/Gizzahub/gzh-cli/internal/idecore"
  )

  // NewIDECmd creates the ide command
  func NewIDECmd() *cobra.Command {
      cmd := &cobra.Command{
          Use:   "ide",
          Short: "IDE monitoring and management",
          // ...
      }

      // 서브커맨드 추가
      cmd.AddCommand(open.NewCmd())
      cmd.AddCommand(status.NewCmd())
      cmd.AddCommand(scan.NewCmd())

      return cmd
  }
  ```
  - 완료 기준: 서브패키지 조립 구조 완성
  - 주의사항: 기존 커맨드 누락 없음

### 7. detector 파일 정리
- [ ] **detector.go 분할 결정** (`cmd/ide/detector.go`)
  - 공용 부분은 이미 internal/idecore로 이동 완료
  - IDE별 특화 감지 로직만 남김 또는 전체를 internal로 이동
  - 완료 기준: detector 로직 최적 위치 결정
  - 주의사항: 인터페이스는 internal, 구현체는 적절한 위치

### 8. 종합 빌드 검증 및 테스트
- [ ] **각 패키지별 빌드** 
  ```bash
  go build ./internal/idecore
  go build ./cmd/ide
  go build ./cmd/ide/open
  go build ./cmd/ide/status
  go build ./cmd/ide/scan
  ```
  - 완료 기준: 모든 패키지 빌드 성공
  - 주의사항: 순환 의존성 없음 확인

- [ ] **전체 프로젝트 빌드** (`go build ./...`)
  - 전체 프로젝트 빌드 성공 확인
  - 완료 기준: 컴파일 에러 없음
  - 주의사항: 다른 패키지에 영향 없음 확인

### 9. 기능 테스트 실행
- [ ] **기본 명령어 테스트** (`./gz ide`)
  ```bash
  ./gz ide --help           # 기본 도움말
  ./gz ide open --help      # open 서브커맨드
  ./gz ide status --help    # status 서브커맨드  
  ./gz ide scan --help      # scan 서브커맨드
  ```
  - 완료 기준: 모든 IDE 명령어 정상 출력
  - 주의사항: 서브커맨드 구조 올바름 확인

- [ ] **실제 기능 테스트** (환경 허용 범위)
  ```bash
  ./gz ide status           # IDE 상태 확인
  ./gz ide scan             # IDE 스캔
  ```
  - 완료 기준: 기본 기능 정상 동작
  - 주의사항: 환경에 따라 결과 다를 수 있음

### 10. 테스트 스위트 실행
- [ ] **internal 패키지 테스트** (`go test ./internal/idecore -v`)
  - idecore 단위 테스트 통과
  - 완료 기준: 모든 테스트 PASS
  - 주의사항: 추출된 코드의 테스트도 함께 이동

- [ ] **각 서브패키지 테스트**
  ```bash
  go test ./cmd/ide -v
  go test ./cmd/ide/open -v
  go test ./cmd/ide/status -v
  go test ./cmd/ide/scan -v
  ```
  - 완료 기준: 모든 IDE 관련 테스트 통과
  - 주의사항: 테스트 파일 경로 문제 없음

- [ ] **전체 IDE 테스트** (`go test ./internal/idecore ./cmd/ide/... -v`)
  - IDE 관련 전체 테스트 통과
  - 완료 기준: 테스트 스위트 안정성 확보
  - 주의사항: 테스트 실패시 원인 분석 후 수정

### 11. 코드 품질 검사
- [ ] **코드 포맷팅** (`make fmt`)
  - gofumpt, gci 포맷팅 실행
  - 완료 기준: 포맷팅 이슈 없음
  - 주의사항: internal 패키지도 포함하여 포맷팅

- [ ] **린팅 검사** (`make lint`)
  - golangci-lint 검사 통과
  - 완료 기준: 린팅 에러 없음
  - 주의사항: 새로운 패키지 구조로 인한 이슈 해결

### 12. 최종 정리 및 커밋
- [ ] **최종 구조 확인**
  ```
  internal/idecore/            # 새로 생성
  ├── types.go                 # IDE 타입 정의
  ├── detector.go             # 감지 인터페이스/구현
  ├── utils.go                # 공용 유틸리티
  ├── types_test.go           # 타입 테스트
  ├── detector_test.go        # 감지 테스트
  └── utils_test.go           # 유틸리티 테스트

  cmd/ide/                    # 수정됨
  ├── ide.go                  # 루트 커맨드 (수정)
  ├── detector.go             # IDE별 특화 로직 (남은 부분)
  ├── detector_test.go        # 감지 테스트 (남은 부분)
  ├── doc.go                  # 패키지 문서 (유지)
  ├── open/                   # 새로 생성
  │   ├── open.go            # 이동됨
  │   └── open_test.go       # 이동됨
  ├── status/                # 새로 생성
  │   ├── status.go          # 이동됨
  │   └── status_test.go     # 이동됨
  └── scan/                  # 새로 생성
      └── scan.go            # 이동됨
  ```
  - 완료 기준: 예상 구조와 일치
  - 주의사항: internal과 cmd 패키지 모두 정리

- [ ] **Git 커밋** (`refactor(ide): extract internal core and create subpackages`)
  - 상세한 커밋 메시지 작성
  - 완료 기준: 커밋 완료 및 phase-3-completed 태그 생성
  - 주의사항: 두 가지 주요 변경(internal 추출 + 서브패키지화) 명시

## 완료 검증 체크리스트

### 빌드 검증
- [ ] `go build ./internal/idecore` 성공
- [ ] `go build ./cmd/ide` 성공  
- [ ] `go build ./cmd/ide/open` 성공
- [ ] `go build ./cmd/ide/status` 성공
- [ ] `go build ./cmd/ide/scan` 성공
- [ ] `go build ./...` 성공

### 기능 검증
- [ ] `./gz ide --help` 정상 출력
- [ ] `./gz ide open --help` 정상 출력
- [ ] `./gz ide status --help` 정상 출력
- [ ] `./gz ide scan --help` 정상 출력
- [ ] `./gz ide status` 실행 성공
- [ ] `./gz ide scan` 실행 성공

### 테스트 검증  
- [ ] `go test ./internal/idecore` 성공
- [ ] `go test ./cmd/ide` 성공
- [ ] `go test ./cmd/ide/open` 성공
- [ ] `go test ./cmd/ide/status` 성공
- [ ] `go test ./cmd/ide/scan` 성공

### 구조 검증
- [ ] 공용 타입들이 internal/idecore로 이동
- [ ] 각 기능별 서브패키지 생성 완료
- [ ] 루트 커맨드가 서브패키지들을 올바르게 조립
- [ ] import 경로가 모두 올바름
- [ ] 순환 참조 없음

## 예상 문제 및 해결책

### 문제 1: 순환 참조
- **증상**: idecore가 cmd/ide를 참조하거나 그 역
- **해결**: 의존성 방향을 일방향으로 고정 (cmd/ide → internal/idecore)

### 문제 2: 너무 많은 코드 이동
- **증상**: 빌드 에러가 너무 많이 발생
- **해결**: 단계적 이동 - 먼저 타입만, 그 다음 함수들

### 문제 3: 테스트 의존성 문제
- **증상**: 테스트가 이동된 코드를 찾지 못함
- **해결**: 테스트 파일도 함께 이동하고 import 수정

### 문제 4: detector 로직 분할 어려움
- **증상**: detector.go를 어디에 둘지 애매함
- **해결**: 인터페이스는 internal, 구현체는 cmd에 유지

## 롤백 계획

### 전체 롤백
```bash
# 모든 변경사항 되돌리기
rm -rf internal/idecore
git checkout -- cmd/ide/
```

### 단계별 롤백  
```bash
# 1단계만 롤백 (internal 생성 취소)
rm -rf internal/idecore
git checkout -- cmd/ide/ide.go

# 2단계만 롤백 (서브패키지 생성 취소)
rm -rf cmd/ide/open cmd/ide/status cmd/ide/scan
git checkout -- cmd/ide/
```

## 성공 기준
1. **구조 개선**: 공용 코드 internal 분리 + 기능별 서브패키지
2. **의존성 정리**: 순환 참조 없는 깔끔한 의존성 구조
3. **기능 보존**: 모든 IDE 명령어 정상 동작
4. **테스트 강화**: internal 코어 컴포넌트 독립 테스트 가능
5. **재사용성**: idecore가 다른 패키지에서도 사용 가능

## 관련 파일
- `internal/idecore/types.go` (새로 생성)
- `internal/idecore/detector.go` (새로 생성)
- `internal/idecore/utils.go` (새로 생성)
- `cmd/ide/ide.go` (대폭 수정)
- `cmd/ide/open/open.go` (이동됨)
- `cmd/ide/status/status.go` (이동됨)
- `cmd/ide/scan/scan.go` (이동됨)
- 관련 테스트 파일들

## 다음 단계
Phase 3 완료 후 → [04-phase4-net-env-comprehensive-restructuring.md](./04-phase4-net-env-comprehensive-restructuring.md)