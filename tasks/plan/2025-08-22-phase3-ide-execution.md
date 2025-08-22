# Phase 3: IDE 패키지 리팩토링 실행 계획

## 개요
**목표**: cmd/ide 패키지의 internal 추출 및 서브패키지화
**소요시간**: 약 4시간
**복잡도**: 높음
**우선순위**: 3순위

## 현재 상태 분석

### 현재 디렉터리 구조
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

### 의존성 분석

#### 공용 요소 (internal로 추출 대상)
- `IDE` struct: 모든 파일에서 사용
- `IDEDetector` interface: detector, scan에서 사용
- `NewIDEDetector()`: 팩토리 함수
- `joinStrings()`: 유틸리티 함수
- 기타 공용 헬퍼 함수들

#### 기능별 그룹
- **open**: open.go, open_test.go
- **status**: status.go, status_test.go
- **scan**: scan.go
- **detector**: detector.go, detector_test.go (일부는 internal로)

## 실행 계획

### 1단계: 의존성 분석 및 추출 대상 식별 (30분)

#### 공용 심볼 확인
```bash
# IDE struct 사용 현황
grep -r "type IDE struct" cmd/ide/
grep -r "IDE{" cmd/ide/

# IDEDetector interface 사용 현황
grep -r "IDEDetector" cmd/ide/

# NewIDEDetector 호출 현황
grep -r "NewIDEDetector" cmd/ide/

# 기타 공용 함수 확인
grep -r "joinStrings" cmd/ide/
grep -r "func.*(" cmd/ide/ | grep -v "test"
```

#### 파일별 의존성 매핑
각 파일이 어떤 공용 요소를 사용하는지 문서화

### 2단계: internal/idecore 생성 (60분)

#### 디렉터리 생성
```bash
mkdir -p internal/idecore
```

#### 공용 타입 추출
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

#### 공용 함수 추출
```go
// internal/idecore/detector.go
package idecore

// NewIDEDetector creates a new IDE detector
func NewIDEDetector() IDEDetector {
    // 구현부 이전
}

// internal/idecore/utils.go
package idecore

// joinStrings joins strings with separator
func joinStrings(strs []string, sep string) string {
    // 구현부 이전
}
```

#### 테스트 파일 생성
```go
// internal/idecore/types_test.go
// internal/idecore/detector_test.go
// internal/idecore/utils_test.go
```

### 3단계: cmd/ide 파일들의 internal 의존성 변경 (45분)

#### import 추가
각 파일에 `"github.com/Gizzahub/gzh-cli/internal/idecore"` import 추가

#### 타입 참조 변경
- `IDE` → `idecore.IDE`
- `IDEDetector` → `idecore.IDEDetector`
- `NewIDEDetector()` → `idecore.NewIDEDetector()`
- `joinStrings()` → `idecore.joinStrings()`

#### 1차 빌드 테스트
```bash
go build ./internal/idecore
go build ./cmd/ide
```

### 4단계: 서브패키지 생성 (90분)

#### open 패키지 생성
```bash
mkdir -p cmd/ide/open
mv cmd/ide/open.go cmd/ide/open/
mv cmd/ide/open_test.go cmd/ide/open/
```

```go
// cmd/ide/open/open.go
package open

import "github.com/Gizzahub/gzh-cli/internal/idecore"

// NewCmd creates the open command
func NewCmd() *cobra.Command {
    // 기존 newOpenCmd 내용을 여기로 이동
}
```

#### status 패키지 생성
```bash
mkdir -p cmd/ide/status
mv cmd/ide/status.go cmd/ide/status/
mv cmd/ide/status_test.go cmd/ide/status/
```

```go
// cmd/ide/status/status.go
package status

import "github.com/Gizzahub/gzh-cli/internal/idecore"

// NewCmd creates the status command
func NewCmd() *cobra.Command {
    // 기존 newStatusCmd 내용을 여기로 이동
}
```

#### scan 패키지 생성
```bash
mkdir -p cmd/ide/scan
mv cmd/ide/scan.go cmd/ide/scan/
```

```go
// cmd/ide/scan/scan.go
package scan

import "github.com/Gizzahub/gzh-cli/internal/idecore"

// NewCmd creates the scan command
func NewCmd() *cobra.Command {
    // 기존 newScanCmd 내용을 여기로 이동
}
```

### 5단계: 루트 커맨드 조립 수정 (30분)

#### ide.go 수정
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
        Short: "JetBrains IDE monitoring and management",
        // ...
    }

    // 서브커맨드 추가
    cmd.AddCommand(open.NewCmd())
    cmd.AddCommand(status.NewCmd())
    cmd.AddCommand(scan.NewCmd())

    return cmd
}
```

### 6단계: detector 정리 (30분)

#### detector.go 분할
- 공용 부분은 이미 internal/idecore로 이동
- IDE별 특화 감지 로직만 남김
- 또는 detector 전체를 internal로 이동 고려

### 7단계: 검증 및 테스트 (45분)

#### 빌드 검증
```bash
# 각 패키지별 빌드
go build ./internal/idecore
go build ./cmd/ide
go build ./cmd/ide/open
go build ./cmd/ide/status
go build ./cmd/ide/scan

# 전체 빌드
go build ./...
```

#### 기능 테스트
```bash
# 기본 명령어
./gz ide --help

# 서브커맨드들
./gz ide open --help
./gz ide status --help
./gz ide scan --help

# 실제 기능 테스트
./gz ide status
./gz ide scan
```

#### 단위 테스트
```bash
# 각 패키지 테스트
go test ./internal/idecore -v
go test ./cmd/ide -v
go test ./cmd/ide/open -v
go test ./cmd/ide/status -v
go test ./cmd/ide/scan -v

# 전체 IDE 관련 테스트
go test ./internal/idecore ./cmd/ide/... -v
```

### 8단계: 최종 정리 및 커밋 (30분)

#### 최종 구조 확인
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
├── detector.go             # IDE별 특화 로직 (옵션)
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

#### Git 커밋
```bash
git add internal/idecore cmd/ide/
git commit -m "refactor(ide): extract internal core and create subpackages

Phase 1: Extract shared components to internal/idecore
- Move IDE types and interfaces to internal/idecore
- Extract IDEDetector interface and factory
- Move shared utility functions
- Add comprehensive tests for core components

Phase 2: Create feature-based subpackages
- cmd/ide/open: IDE opening functionality
- cmd/ide/status: IDE status checking
- cmd/ide/scan: IDE scanning features
- Maintain detector logic in cmd/ide (IDE-specific)

Benefits:
- Improved code organization and navigation
- Clear separation of concerns
- Reusable core components
- Independent testing of features

Files moved:
- Shared types → internal/idecore/types.go
- Detector interface → internal/idecore/detector.go
- Utilities → internal/idecore/utils.go
- open.go → cmd/ide/open/
- status.go → cmd/ide/status/
- scan.go → cmd/ide/scan/"
```

## 검증 체크리스트

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
**증상**: idecore가 cmd/ide를 참조하거나 그 역
**해결**: 의존성 방향을 일방향으로 고정 (cmd/ide → internal/idecore)

### 문제 2: 너무 많은 코드 이동
**증상**: 빌드 에러가 너무 많이 발생
**해결**: 단계적 이동 - 먼저 타입만, 그 다음 함수들

### 문제 3: 테스트 의존성 문제
**증상**: 테스트가 이동된 코드를 찾지 못함
**해결**: 테스트 파일도 함께 이동하고 import 수정

### 문제 4: detector 로직 분할 어려움
**증상**: detector.go를 어디에 둘지 애매함
**해결**: 인터페이스는 internal, 구현체는 cmd에 유지

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

## 다음 단계
Phase 3 완료 후 → [Phase 4: net-env 실행 계획](./2025-08-22-phase4-net-env-execution.md)
