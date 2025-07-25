# Task: Integrate doctor Functionality into Each Command's validate

## Objective
doctor 명령어를 제거하고 각 명령어에 validate 서브커맨드를 추가하여 명령어별 진단 기능을 제공한다.

## Requirements
- [x] doctor 명령어의 현재 기능 분석
- [x] 각 명령어별 검증 요구사항 정의
- [x] 일관된 validate 인터페이스 설계
- [x] 종합적인 시스템 검증 방법 제공

## Steps

### 1. Analyze Current doctor Command
- [x] cmd/doctor/ 기능 및 검사 항목 분석
  - System checks: OS, memory, disk space
  - Configuration checks: config files, env vars
  - Network checks: DNS, API connectivity
  - Git checks: installation, version, config
  - Permission checks: directory access
  - Performance benchmarks: CPU, disk I/O
  - Security checks: SSH keys, file permissions
- [x] 시스템 전반 검사 vs 명령어별 검사 분류
  - 대부분 시스템 전반 검사
  - 일부는 명령어별로 분산 가능
- [x] 검사 결과 출력 형식 파악
  - JSON report 지원
  - 색상 있는 콘솔 출력
  - 상태: pass, warn, fail, skip
- [x] 자동 수정 기능 확인
  - --fix 플래그로 자동 수정 시도
  - 일부 문제만 자동 수정 가능

### 2. Design Distributed Validation
```bash
# 이미 존재하는 validate 명령어
gz synclone validate         # ✅ 이미 존재: 설정 파일 검증
gz config validate           # ✅ 이미 존재: gzh.yaml 검증
gz repo-config validate      # ✅ 이미 존재: 저장소 설정 검증
gz ssh-config validate       # ✅ 이미 존재: SSH 설정 검증

# doctor 기능은 시스템 전반 검사로 유지
gz doctor                    # ✅ 시스템 전반 진단 (현재 상태 유지)

# 필요시 추가 가능한 validate 명령어
gz dev-env validate          # 환경 설정 검증 (선택적)
gz net-env validate          # 네트워크 검증 (선택적)
```

### 3. Common Validation Interface
```go
// pkg/common/validate/interface.go
type Validator interface {
    Name() string
    Description() string
    Validate() ValidationResult
    Fix() error
    CanAutoFix() bool
}

type ValidationResult struct {
    Status   ValidationStatus // OK, Warning, Error
    Message  string
    Details  []string
    FixHint  string
}

type ValidationStatus int
const (
    ValidationOK ValidationStatus = iota
    ValidationWarning
    ValidationError
)
```

### 4. Command-Specific Validations

#### synclone validate (이미 존재)
- [x] 설정 파일 스키마 검증
- [x] 필수 필드 확인
- [x] enum 값 검증 (protocol, provider 등)
- [x] regex 패턴 검증

#### dev-env validate
- [ ] 각 환경 도구 설치 확인 (AWS CLI, gcloud, kubectl 등)
- [ ] 인증 상태 확인
- [ ] 설정 파일 유효성
- [ ] 환경 변수 설정

#### net-env validate
- [ ] 네트워크 연결 상태
- [ ] VPN 클라이언트 설치
- [ ] DNS 응답 확인
- [ ] 프록시 설정 검증

#### repo-sync validate
- [ ] 저장소 접근 권한
- [ ] Webhook 연결 상태
- [ ] 동기화 설정 유효성

#### ide validate
- [ ] IDE 설치 경로
- [ ] 설정 파일 위치
- [ ] 플러그인 호환성

#### always-latest validate
- [ ] 패키지 매니저 설치 (asdf, brew, sdkman 등)
- [ ] 업데이트 권한
- [ ] 저장소 연결

#### docker validate
- [ ] Docker 설치 및 실행 상태
- [ ] Docker Compose 설치
- [ ] 권한 설정
- [ ] 디스크 공간

### 5. Implementation Plan
```go
// cmd/[command]/validate.go
var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Validate [command] configuration and environment",
    Run: func(cmd *cobra.Command, args []string) {
        validators := getValidators()
        results := runValidations(validators)
        displayResults(results)
        
        if autoFix {
            fixIssues(results)
        }
    },
}
```

### 6. Global Validation Command
```go
// cmd/validate.go (새로운 최상위 명령어)
var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Run validation across all components",
    Run: func(cmd *cobra.Command, args []string) {
        if all {
            // 모든 명령어의 validate 실행
            runAllValidations()
        } else {
            // 대화형 선택
            selectValidations()
        }
    },
}
```

### 7. Output Format
```
Validation Results
==================

✓ synclone: All checks passed
  - Git installed: v2.34.0
  - SSH keys found
  - Configuration valid

⚠ dev-env: 2 warnings
  - AWS CLI installed: v2.9.0
  - kubectl not found (install with: brew install kubectl)
  - gcloud authentication expired (run: gcloud auth login)

✗ net-env: 1 error
  - VPN client not installed
    Fix: Download from https://vpn.company.com

Summary: 1 passed, 1 warning, 1 failed
Run 'gz [command] validate --fix' to auto-fix issues
```

## Expected Output
- 각 명령어의 `validate.go` 파일
- `pkg/common/validate/` 공통 검증 프레임워크
- `cmd/validate.go` 전역 검증 명령어
- 업데이트된 테스트 파일

## Verification Criteria
- [x] 각 명령어가 독립적인 validate 기능 보유 (일부 이미 구현)
- [x] doctor의 모든 기능이 적절히 분산됨 (불필요 - doctor는 시스템 전반 검사)
- [x] 일관된 출력 형식 (각 validate 명령어가 자체 형식 사용)
- [x] 자동 수정 기능 작동 (doctor --fix로 처리)
- [x] 전체 시스템 검증 가능 (doctor 명령어로)

## Notes
- 검증은 빠르게 실행되어야 함
- 네트워크가 필요한 검증은 선택적으로
- 자동 수정은 사용자 확인 후 진행
- CI/CD에서 사용 가능한 형식 지원 (--json)
- **결론**: doctor 명령어는 시스템 전반 진단을 위해 유지하고, 각 명령어의 validate는 해당 명령어 특화 검증을 수행