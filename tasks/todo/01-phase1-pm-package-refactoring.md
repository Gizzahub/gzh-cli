# Phase 1: PM 패키지 구조 리팩토링

## 개요
- **목표**: cmd/pm 패키지 파일들을 기능별 디렉터리로 이동하여 탐색성 개선
- **우선순위**: HIGH
- **예상 소요시간**: 2시간
- **담당자**: Backend
- **복잡도**: 낮음

## 선행 작업
- [x] 기존 빈 디렉터리들이 이미 생성됨 확인
- [ ] 현재 브랜치 상태 커밋 완료
- [ ] `refactor-start` 태그 생성

## 세부 작업 목록

### 1. 현재 상태 확인 및 백업
- [ ] **현재 디렉터리 구조 확인** (`cmd/pm/`)
  - 기존 파일 목록 문서화
  - 빈 디렉터리 상태 확인 (advanced/, cache/, doctor/, export/, install/, status/, update/)
  - 완료 기준: `ls -la cmd/pm/` 결과 확인 완료
  - 주의사항: 파일과 디렉터리 중복 존재 확인

- [ ] **Git 백업 지점 생성** (`git tag refactor-phase1-start`)
  - refactor-phase1-pm 브랜치 생성
  - 현재 상태 커밋
  - 완료 기준: 태그 및 브랜치 생성 확료
  - 주의사항: 작업 전 반드시 백업 지점 생성

- [ ] **빌드 상태 사전 검증** (`make build`)
  - `go build ./cmd/pm` 성공 확인
  - `go build ./...` 전체 빌드 성공 확인
  - 완료 기준: 컴파일 에러 없음
  - 주의사항: 리팩토링 전 정상 상태 확인 필수

### 2. 파일 이동 실행
- [ ] **파일 이동 매핑 실행** (`cmd/pm/`)
  ```bash
  mv cmd/pm/advanced.go cmd/pm/advanced/
  mv cmd/pm/cache.go cmd/pm/cache/
  mv cmd/pm/doctor.go cmd/pm/doctor/
  mv cmd/pm/export.go cmd/pm/export/
  mv cmd/pm/install.go cmd/pm/install/
  mv cmd/pm/status.go cmd/pm/status/
  mv cmd/pm/update.go cmd/pm/update/
  ```
  - 완료 기준: 모든 기능별 파일이 해당 디렉터리로 이동 완료
  - 주의사항: 루트 pm.go는 이동하지 않음 (커맨드 조립 역할)

- [ ] **Package 선언 확인** (`package pm`)
  - 모든 이동된 파일의 package 선언이 `package pm` 유지 확인
  - 완료 기준: 모든 파일에서 package 선언 일관성 확인
  - 주의사항: Go의 동일 패키지 분산 디렉터리 기능 활용

### 3. 빌드 검증 및 테스트
- [ ] **1차 빌드 검증** (`go build ./cmd/pm`)
  - PM 패키지 개별 빌드 성공 확인
  - import 에러 수정 (필요시)
  - 완료 기준: 컴파일 에러 없음
  - 주의사항: 에러 발생시 즉시 롤백 고려

- [ ] **전체 빌드 검증** (`go build ./...`)
  - 전체 프로젝트 빌드 성공 확인
  - 완료 기준: 모든 패키지 컴파일 성공
  - 주의사항: 의존성 문제 없음 확인

- [ ] **기능 테스트 실행** (`./gz pm`)
  ```bash
  ./gz pm --help          # 기본 도움말
  ./gz pm status          # 상태 확인
  ./gz pm install --help  # 도움말 확인
  ./gz pm cache list      # 캐시 목록
  ./gz pm doctor          # 진단 실행
  ```
  - 완료 기준: 모든 pm 하위 명령어 정상 동작
  - 주의사항: 기능 손실 없음 확인

### 4. 테스트 스위트 실행
- [ ] **PM 패키지 단위 테스트** (`go test ./cmd/pm -v`)
  - PM 패키지 테스트 통과 확인
  - 완료 기준: 모든 테스트 PASS
  - 주의사항: 테스트 실패시 원인 분석 후 수정

- [ ] **관련 테스트 실행** (`go test ./cmd/pm/... -v`)
  - 서브디렉터리 포함 전체 테스트
  - 완료 기준: 테스트 에러 없음
  - 주의사항: 테스트 경로 문제 확인

### 5. 코드 품질 검사
- [ ] **코드 포맷팅** (`make fmt`)
  - gofumpt, gci 포맷팅 실행
  - 완료 기준: 포맷팅 이슈 없음
  - 주의사항: 포맷 변경사항 커밋에 포함

- [ ] **린팅 검사** (`make lint`)
  - golangci-lint 검사 통과
  - 완료 기준: 린팅 에러 없음
  - 주의사항: 구조 변경으로 인한 새로운 린팅 이슈 해결

### 6. 최종 정리 및 커밋
- [ ] **최종 구조 확인**
  ```
  cmd/pm/
  ├── pm.go                    # 루트 (유지)
  ├── advanced/
  │   └── advanced.go          # 이동됨
  ├── cache/
  │   └── cache.go            # 이동됨
  ├── doctor/
  │   └── doctor.go           # 이동됨
  ├── export/
  │   └── export.go           # 이동됨
  ├── install/
  │   └── install.go          # 이동됨
  ├── status/
  │   └── status.go           # 이동됨
  └── update/
      └── update.go           # 이동됨
  ```
  - 완료 기준: 예상 구조와 일치
  - 주의사항: 파일 누락 없음 확인

- [ ] **Git 커밋** (`refactor(pm): reorganize files into feature directories`)
  - 의미있는 커밋 메시지 작성
  - 완료 기준: 커밋 완료 및 phase-1-completed 태그 생성
  - 주의사항: 커밋 메시지에 변경 사항 상세 기록

## 완료 검증 체크리스트

### 필수 검증
- [ ] `go build ./cmd/pm` 성공
- [ ] `go build ./...` 성공  
- [ ] `go test ./cmd/pm` 성공
- [ ] `./gz pm status` 실행 성공
- [ ] `./gz pm --help` 출력 정상

### 구조 검증
- [ ] 모든 기능별 파일이 해당 디렉터리로 이동
- [ ] 루트 `pm.go`는 그대로 유지
- [ ] package 선언이 모두 `package pm`
- [ ] import 경로 변경 없음

### 기능 검증  
- [ ] pm install 명령어 정상 동작
- [ ] pm status 명령어 정상 동작
- [ ] pm cache 명령어 정상 동작
- [ ] pm doctor 명령어 정상 동작
- [ ] pm export 명령어 정상 동작
- [ ] pm update 명령어 정상 동작
- [ ] pm advanced 명령어 정상 동작

## 롤백 계획

### 문제 발생 시 즉시 롤백
```bash
# 변경사항 되돌리기
git checkout -- cmd/pm/

# 또는 태그로 롤백
git checkout refactor-phase1-start
```

### 부분 롤백
- 특정 파일만 문제가 있다면 해당 파일만 원위치로 복구
- 단계별로 이동했다면 문제 파일부터 역순으로 복구

## 예상 문제 및 해결책

### 문제 1: 빌드 에러
- **원인**: package 선언 불일치
- **해결**: 모든 파일의 package가 `pm`인지 확인

### 문제 2: import 에러  
- **원인**: 상대 경로 import 문제
- **해결**: 절대 경로로 import 수정

### 문제 3: 테스트 실패
- **원인**: 테스트 파일 경로 문제
- **해결**: 테스트 파일도 함께 이동 확인

## 성공 기준
1. **기능 보존**: 모든 pm 하위 명령어가 정상 동작
2. **구조 개선**: 기능별 파일 그룹핑으로 탐색성 향상
3. **빌드 성공**: 컴파일 에러 없음
4. **테스트 통과**: 기존 테스트 모두 통과

## 관련 파일
- `cmd/pm/pm.go` (루트 유지)
- `cmd/pm/advanced/advanced.go` (이동됨)  
- `cmd/pm/cache/cache.go` (이동됨)
- `cmd/pm/doctor/doctor.go` (이동됨)
- `cmd/pm/export/export.go` (이동됨)
- `cmd/pm/install/install.go` (이동됨)
- `cmd/pm/status/status.go` (이동됨)
- `cmd/pm/update/update.go` (이동됨)

## 다음 단계
Phase 1 완료 후 → [02-phase2-repo-config-refactoring.md](./02-phase2-repo-config-refactoring.md)