# Phase 2: repo-config 패키지 구조 리팩토링

## 개요
- **목표**: cmd/repo-config 패키지 파일들을 기능별 디렉터리로 이동 (1차: 파일 이동, 2차: internal 추출)
- **우선순위**: HIGH  
- **예상 소요시간**: 3시간
- **담당자**: Backend
- **복잡도**: 중간 (GlobalFlags 의존성 존재)

## 선행 작업
- [ ] Phase 1 (PM 패키지 리팩토링) 완료
- [ ] 현재 브랜치에서 refactor-phase2-repo-config 브랜치 생성
- [ ] GlobalFlags 의존성 분석 완료

## 세부 작업 목록

### 1. 의존성 분석 및 현재 상태 확인
- [ ] **GlobalFlags 사용 현황 분석** (`cmd/repo-config/`)
  ```bash
  grep -r "GlobalFlags" cmd/repo-config/
  grep -r "addGlobalFlags" cmd/repo-config/
  ```
  - 완료 기준: GlobalFlags 사용 파일 목록 정리 완료
  - 주의사항: 의존성 체인 파악하여 이동 순서 결정

- [ ] **공용 함수 의존성 확인** (`client_factory.go`, `utils.go`)
  ```bash
  grep -r "client_factory" cmd/repo-config/
  grep -r "utils\." cmd/repo-config/
  ```
  - 완료 기준: 공용 함수 호출 관계 매핑 완료
  - 주의사항: 순환 참조 가능성 사전 점검

- [ ] **현재 디렉터리 구조 확인**
  ```
  cmd/repo-config/
  ├── repo_config.go         # 루트 (유지)
  ├── client_factory.go      # 공용 (유지)
  ├── utils.go              # 공용 (유지)
  ├── doc.go                # 문서 (유지)
  ├── integration_test.go    # 통합테스트 (유지)
  ├── repo_config_test.go    # 메인테스트 (유지)
  └── [기능별 파일들...]     # 이동 대상
  ```
  - 완료 기준: 유지/이동 대상 파일 분류 완료
  - 주의사항: 빈 디렉터리들 존재 확인

### 2. Git 백업 및 브랜치 준비
- [ ] **백업 지점 생성** (`git tag refactor-phase2-start`)
  - refactor-phase2-repo-config 브랜치 생성 및 체크아웃
  - 현재 상태 커밋
  - 완료 기준: 브랜치 및 태그 생성 완료
  - 주의사항: Phase 1 완료 상태에서 시작

- [ ] **빌드 상태 사전 검증** (`go build ./cmd/repo-config`)
  - 전체 빌드 성공 확인
  - 완료 기준: 컴파일 에러 없음
  - 주의사항: 리팩토링 전 정상 상태 보장

### 3. 기능별 파일 이동 실행
- [ ] **apply 기능 이동** (`cmd/repo-config/apply/`)
  ```bash
  mv cmd/repo-config/apply.go cmd/repo-config/apply/
  ```
  - 완료 기준: apply.go 이동 완료
  - 주의사항: package 선언 `package repoconfig` 유지

- [ ] **audit 기능 이동** (`cmd/repo-config/audit/`)
  ```bash
  mv cmd/repo-config/audit.go cmd/repo-config/audit/
  ```
  - 완료 기준: audit.go 이동 완료
  - 주의사항: package 선언 확인

- [ ] **dashboard 기능 이동** (`cmd/repo-config/dashboard/`)
  ```bash
  mv cmd/repo-config/dashboard.go cmd/repo-config/dashboard/
  ```
  - 완료 기준: dashboard.go 이동 완료
  - 주의사항: package 선언 확인

- [ ] **diff 기능 이동 (테스트 포함)** (`cmd/repo-config/diff/`)
  ```bash
  mv cmd/repo-config/diff.go cmd/repo-config/diff/
  mv cmd/repo-config/diff_test.go cmd/repo-config/diff/
  ```
  - 완료 기준: diff 관련 파일 이동 완료
  - 주의사항: 테스트 파일과 소스 파일 함께 이동

- [ ] **list 기능 이동** (`cmd/repo-config/list/`)
  ```bash
  mv cmd/repo-config/list.go cmd/repo-config/list/
  ```
  - 완료 기준: list.go 이동 완료
  - 주의사항: package 선언 확인

- [ ] **risk 기능 이동** (`cmd/repo-config/risk/`)
  ```bash
  mv cmd/repo-config/risk.go cmd/repo-config/risk/
  ```
  - 완료 기준: risk.go 이동 완료
  - 주의사항: package 선언 확인

- [ ] **template 기능 이동** (`cmd/repo-config/template/`)
  ```bash
  mv cmd/repo-config/template.go cmd/repo-config/template/
  # templates/ 디렉터리는 이미 올바른 위치
  ```
  - 완료 기준: template.go 이동 완료
  - 주의사항: templates/ 디렉터리 경로 참조 확인

- [ ] **validate 기능 이동** (`cmd/repo-config/validate/`)
  ```bash
  mv cmd/repo-config/validate.go cmd/repo-config/validate/
  ```
  - 완료 기준: validate.go 이동 완료
  - 주의사항: package 선언 확인

- [ ] **webhook 기능 이동** (`cmd/repo-config/webhook/`)
  ```bash
  mv cmd/repo-config/webhook.go cmd/repo-config/webhook/
  ```
  - 완료 기준: webhook.go 이동 완료
  - 주의사항: package 선언 확인

### 4. 빌드 검증 및 에러 수정
- [ ] **1차 빌드 검증** (`go build ./cmd/repo-config`)
  - repo-config 패키지 빌드 성공 확인
  - import 에러 수정 (필요시)
  - 완료 기준: 컴파일 에러 없음
  - 주의사항: GlobalFlags 참조 에러 발생시 임시 수정

- [ ] **전체 빌드 검증** (`go build ./...`)
  - 전체 프로젝트 빌드 성공 확인
  - 완료 기준: 모든 패키지 컴파일 성공
  - 주의사항: 의존성 문제 즉시 해결

### 5. 기능 테스트 실행
- [ ] **기본 명령어 테스트** (`./gz repo-config`)
  ```bash
  ./gz repo-config --help            # 기본 도움말
  ./gz repo-config apply --help      # apply 도움말
  ./gz repo-config audit --help      # audit 도움말
  ./gz repo-config dashboard --help  # dashboard 도움말
  ./gz repo-config diff --help       # diff 도움말
  ./gz repo-config list --help       # list 도움말
  ./gz repo-config risk --help       # risk 도움말
  ./gz repo-config template --help   # template 도움말
  ./gz repo-config validate --help   # validate 도움말
  ./gz repo-config webhook --help    # webhook 도움말
  ```
  - 완료 기준: 모든 서브커맨드 도움말 정상 출력
  - 주의사항: 누락된 명령어 없음 확인

- [ ] **간단한 기능 테스트** (환경 의존성 최소)
  ```bash
  ./gz repo-config validate --help    # 검증 기능
  ./gz repo-config template list      # 템플릿 목록 (의존성 적음)
  ./gz repo-config list --help        # 리스트 기능
  ```
  - 완료 기준: 기본 기능 정상 동작
  - 주의사항: 환경 의존적 기능은 에러 처리 확인

### 6. 테스트 스위트 실행
- [ ] **repo-config 전체 테스트** (`go test ./cmd/repo-config -v`)
  - 메인 패키지 테스트 통과
  - 완료 기준: 모든 테스트 PASS
  - 주의사항: 통합 테스트는 환경에 따라 스킵될 수 있음

- [ ] **개별 기능 테스트** (`go test ./cmd/repo-config/diff -v`)
  - diff 기능 테스트 통과 (테스트 파일이 있는 기능)
  - 완료 기준: diff 테스트 성공
  - 주의사항: 테스트 파일과 소스 파일 경로 문제 없음

- [ ] **통합 테스트** (`go test ./cmd/repo-config -run Integration -v`)
  - 통합 테스트 실행 (환경에 따라 스킵)
  - 완료 기준: 실행 가능한 테스트 통과
  - 주의사항: 환경 의존성 테스트 적절히 처리

### 7. 코드 품질 검사
- [ ] **코드 포맷팅** (`make fmt`)
  - gofumpt, gci 포맷팅 실행
  - 완료 기준: 포맷팅 이슈 없음
  - 주의사항: 파일 이동으로 인한 import 정리

- [ ] **린팅 검사** (`make lint`)
  - golangci-lint 검사 통과
  - 완료 기준: 린팅 에러 없음
  - 주의사항: 구조 변경으로 인한 새로운 이슈 해결

### 8. 최종 정리 및 커밋
- [ ] **최종 구조 확인**
  ```
  cmd/repo-config/
  ├── repo_config.go           # 루트 (유지)
  ├── client_factory.go        # 공용 (유지)
  ├── utils.go                 # 공용 (유지)
  ├── doc.go                   # 문서 (유지)
  ├── integration_test.go      # 통합테스트 (유지)
  ├── repo_config_test.go      # 메인테스트 (유지)
  ├── apply/
  │   └── apply.go            # 이동됨
  ├── audit/
  │   └── audit.go            # 이동됨
  ├── dashboard/
  │   └── dashboard.go        # 이동됨
  ├── diff/
  │   ├── diff.go             # 이동됨
  │   └── diff_test.go        # 이동됨
  ├── list/
  │   └── list.go             # 이동됨
  ├── risk/
  │   └── risk.go             # 이동됨
  ├── template/
  │   └── template.go         # 이동됨
  ├── templates/              # 템플릿 파일들 (유지)
  ├── validate/
  │   └── validate.go         # 이동됨
  └── webhook/
      └── webhook.go          # 이동됨
  ```
  - 완료 기준: 예상 구조와 일치
  - 주의사항: 공용 파일들은 루트에 유지

- [ ] **Git 커밋** (`refactor(repo-config): reorganize files into feature directories`)
  - 의미있는 커밋 메시지 작성
  - 완료 기준: 커밋 완료 및 phase-2-completed 태그 생성
  - 주의사항: 파일 이동 내역 상세 기록

## 완료 검증 체크리스트

### 빌드 검증
- [ ] `go build ./cmd/repo-config` 성공
- [ ] `go build ./...` 성공
- [ ] 컴파일 에러 없음

### 기능 검증
- [ ] `./gz repo-config --help` 정상 출력
- [ ] `./gz repo-config apply --help` 정상 출력
- [ ] `./gz repo-config audit --help` 정상 출력
- [ ] `./gz repo-config dashboard --help` 정상 출력  
- [ ] `./gz repo-config diff --help` 정상 출력
- [ ] `./gz repo-config list --help` 정상 출력
- [ ] `./gz repo-config risk --help` 정상 출력
- [ ] `./gz repo-config template --help` 정상 출력
- [ ] `./gz repo-config validate --help` 정상 출력
- [ ] `./gz repo-config webhook --help` 정상 출력

### 테스트 검증
- [ ] `go test ./cmd/repo-config` 성공
- [ ] `go test ./cmd/repo-config/diff` 성공
- [ ] 기존 테스트 모두 통과

### 구조 검증
- [ ] 모든 기능별 파일이 해당 디렉터리로 이동
- [ ] 공용 파일들은 루트에 유지
- [ ] package 선언이 모두 `package repoconfig`
- [ ] templates/ 디렉터리 구조 유지

## 예상 문제 및 해결책

### 문제 1: GlobalFlags 참조 에러
- **증상**: 이동된 파일에서 GlobalFlags를 찾을 수 없음
- **해결**: 일시적으로 각 파일에서 상대 경로 조정 또는 개별 import 추가

### 문제 2: 공용 함수 참조 에러  
- **증상**: utils.go나 client_factory.go의 함수를 찾을 수 없음
- **해결**: 같은 패키지 내 직접 호출 가능, 경로 문제시 상대 import 확인

### 문제 3: templates 경로 문제
- **증상**: template.go에서 templates/ 디렉터리를 찾을 수 없음
- **해결**: 상대 경로를 `../templates/`로 수정하거나 절대 경로 사용

### 문제 4: 테스트 파일 문제
- **증상**: diff_test.go가 diff.go를 찾을 수 없음
- **해결**: 테스트 파일과 소스 파일이 같은 디렉터리에 있는지 확인

## 롤백 계획

### 즉시 롤백
```bash
# 모든 변경사항 되돌리기
git checkout -- cmd/repo-config/

# 스테이징된 변경사항 되돌리기  
git restore --staged cmd/repo-config/
git restore cmd/repo-config/
```

### 부분 롤백
특정 기능만 문제가 있는 경우:
```bash
# 예: diff 기능만 롤백
git checkout -- cmd/repo-config/diff/
mv cmd/repo-config/diff/diff.go cmd/repo-config/
mv cmd/repo-config/diff/diff_test.go cmd/repo-config/
```

## 성공 기준
1. **구조 개선**: 기능별 파일 분리로 탐색성 향상
2. **기능 보존**: 모든 repo-config 명령어 정상 동작
3. **빌드 성공**: 컴파일 에러 없음
4. **테스트 통과**: 기존 테스트 모두 통과
5. **준비 완료**: 2차 internal 추출을 위한 기반 마련

## 관련 파일
- `cmd/repo-config/repo_config.go` (루트, 유지)
- `cmd/repo-config/client_factory.go` (공용, 유지)
- `cmd/repo-config/utils.go` (공용, 유지)
- `cmd/repo-config/apply/apply.go` (이동됨)
- `cmd/repo-config/audit/audit.go` (이동됨)
- `cmd/repo-config/dashboard/dashboard.go` (이동됨)
- `cmd/repo-config/diff/diff.go` (이동됨)
- `cmd/repo-config/diff/diff_test.go` (이동됨)
- `cmd/repo-config/list/list.go` (이동됨)
- `cmd/repo-config/risk/risk.go` (이동됨)
- `cmd/repo-config/template/template.go` (이동됨)
- `cmd/repo-config/validate/validate.go` (이동됨)
- `cmd/repo-config/webhook/webhook.go` (이동됨)
- `cmd/repo-config/templates/` (템플릿 파일들, 유지)

## 다음 단계
Phase 2 완료 후 → [03-phase3-ide-internal-extraction.md](./03-phase3-ide-internal-extraction.md)