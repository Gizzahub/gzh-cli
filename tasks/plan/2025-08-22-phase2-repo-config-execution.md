# Phase 2: repo-config 패키지 리팩토링 실행 계획

## 개요
**목표**: cmd/repo-config 패키지의 파일들을 기능별 디렉터리로 이동
**소요시간**: 약 3시간
**복잡도**: 중간
**우선순위**: 2순위

## 현재 상태 분석

### 디렉터리 구조
```
cmd/repo-config/
├── repo_config.go         # 루트 커맨드 조립
├── client_factory.go      # 공용 팩토리
├── utils.go              # 공용 유틸리티
├── doc.go                # 패키지 문서
├── apply.go              # 설정 적용
├── apply/                # (빈 디렉터리)
├── audit.go              # 감사 기능
├── audit/                # (빈 디렉터리)
├── dashboard.go          # 대시보드
├── dashboard/            # (빈 디렉터리)
├── diff.go               # 차이점 비교
├── diff_test.go          # 차이점 테스트
├── diff/                 # (빈 디렉터리)
├── list.go               # 목록 조회
├── list/                 # (빈 디렉터리)
├── risk.go               # 리스크 분석
├── risk/                 # (빈 디렉터리)
├── template.go           # 템플릿 관리
├── template/             # (빈 디렉터리)
├── templates/            # 템플릿 파일들
├── validate.go           # 검증 기능
├── validate/             # (빈 디렉터리)
├── webhook.go            # 웹훅 관리
├── webhook/              # (빈 디렉터리)
├── integration_test.go   # 통합 테스트
└── repo_config_test.go   # 메인 테스트
```

### 의존성 분석

#### 공용 파일 (루트 유지)
- `repo_config.go`: 메인 커맨드 조립
- `client_factory.go`: GitHub/GitLab 클라이언트 팩토리
- `utils.go`: 공용 헬퍼 함수들
- `doc.go`: 패키지 문서
- `integration_test.go`: 통합 테스트
- `repo_config_test.go`: 메인 테스트

#### GlobalFlags 의존성
원본 계획서에 따르면 `GlobalFlags`와 관련 헬퍼들이 많은 파일에서 공유됨. 1차에서는 파일 이동만 하고 2차에서 `internal/repoconfig`로 추출 예정.

## 실행 계획

### 1단계: 의존성 분석 및 확인 (20분)

#### GlobalFlags 사용 현황 확인
```bash
# GlobalFlags 사용 파일 검색
grep -r "GlobalFlags" cmd/repo-config/
grep -r "addGlobalFlags" cmd/repo-config/
```

#### 공용 함수 사용 현황 확인
```bash
# 공용 함수 호출 관계 확인
grep -r "client_factory" cmd/repo-config/
grep -r "utils\." cmd/repo-config/
```

### 2단계: 파일 이동 실행 (60분)

#### 기능별 파일 매핑
```bash
# apply 기능
mv cmd/repo-config/apply.go cmd/repo-config/apply/

# audit 기능
mv cmd/repo-config/audit.go cmd/repo-config/audit/

# dashboard 기능
mv cmd/repo-config/dashboard.go cmd/repo-config/dashboard/

# diff 기능 (테스트 파일 포함)
mv cmd/repo-config/diff.go cmd/repo-config/diff/
mv cmd/repo-config/diff_test.go cmd/repo-config/diff/

# list 기능
mv cmd/repo-config/list.go cmd/repo-config/list/

# risk 기능
mv cmd/repo-config/risk.go cmd/repo-config/risk/

# template 기능
mv cmd/repo-config/template.go cmd/repo-config/template/
# templates/ 디렉터리는 이미 올바른 위치

# validate 기능
mv cmd/repo-config/validate.go cmd/repo-config/validate/

# webhook 기능
mv cmd/repo-config/webhook.go cmd/repo-config/webhook/
```

#### Package 선언 확인
각 이동된 파일의 package 선언이 `package repoconfig`인지 확인

### 3단계: 빌드 및 1차 검증 (30분)

#### 빌드 테스트
```bash
# repo-config 패키지 빌드
go build ./cmd/repo-config

# 전체 프로젝트 빌드
go build ./...
```

#### Import 에러 수정
파일 이동으로 인한 내부 참조 문제가 있다면 수정

### 4단계: 기능 테스트 (45분)

#### 주요 명령어 테스트
```bash
# 기본 도움말
./gz repo-config --help

# 각 서브커맨드 도움말
./gz repo-config apply --help
./gz repo-config audit --help
./gz repo-config dashboard --help
./gz repo-config diff --help
./gz repo-config list --help
./gz repo-config risk --help
./gz repo-config template --help
./gz repo-config validate --help
./gz repo-config webhook --help
```

#### 간단한 기능 테스트
```bash
# 검증 기능 (의존성이 적은 기능부터)
./gz repo-config validate --help

# 템플릿 기능
./gz repo-config template list

# 리스트 기능
./gz repo-config list --help
```

### 5단계: 단위 테스트 (30분)

#### 테스트 실행
```bash
# repo-config 패키지 전체 테스트
go test ./cmd/repo-config -v

# 개별 기능별 테스트
go test ./cmd/repo-config/diff -v

# 통합 테스트 (환경에 따라 스킵될 수 있음)
go test ./cmd/repo-config -run Integration -v
```

### 6단계: 정리 및 커밋 (15분)

#### 최종 구조 확인
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

#### Git 커밋
```bash
git add cmd/repo-config/
git commit -m "refactor(repo-config): reorganize files into feature directories

- Move feature-specific files to subdirectories
- Maintain package repoconfig namespace across directories
- Keep shared utilities and factories in root
- Preserve templates/ directory structure

Phase 1 changes (file organization only):
- apply.go → apply/
- audit.go → audit/
- dashboard.go → dashboard/
- diff.go, diff_test.go → diff/
- list.go → list/
- risk.go → risk/
- template.go → template/
- validate.go → validate/
- webhook.go → webhook/

Shared files retained in root:
- repo_config.go (command assembly)
- client_factory.go (shared factory)
- utils.go (shared utilities)
- *_test.go (integration tests)"
```

## 검증 체크리스트

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
**증상**: 이동된 파일에서 GlobalFlags를 찾을 수 없음
**해결**: 일시적으로 각 파일에서 개별 import 추가하거나 상대 경로 조정

### 문제 2: 공용 함수 참조 에러
**증상**: utils.go나 client_factory.go의 함수를 찾을 수 없음
**해결**: 같은 패키지 내에서는 직접 호출 가능하지만, 경로 문제시 상대 import 확인

### 문제 3: 테스트 파일 문제
**증상**: diff_test.go가 diff.go를 찾을 수 없음
**해결**: 테스트 파일과 소스 파일이 같은 디렉터리에 있는지 확인

### 문제 4: templates 경로 문제
**증상**: template.go에서 templates/ 디렉터리를 찾을 수 없음
**해결**: 상대 경로를 `../templates/`로 수정하거나 절대 경로 사용

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

## 2차 계획 준비

Phase 2 완료 후, 다음 단계인 `internal/repoconfig` 추출을 위한 준비:
1. GlobalFlags 의존성 매핑 완료
2. 공용 헬퍼 함수 목록 정리
3. 순환 의존성 가능성 분석

## 성공 기준
1. **구조 개선**: 기능별 파일 분리로 탐색성 향상
2. **기능 보존**: 모든 repo-config 명령어 정상 동작
3. **빌드 성공**: 컴파일 에러 없음
4. **테스트 통과**: 기존 테스트 모두 통과
5. **준비 완료**: 2차 internal 추출을 위한 기반 마련

## 다음 단계
Phase 2 완료 후 → [Phase 3: IDE 실행 계획](./2025-08-22-phase3-ide-execution.md)
