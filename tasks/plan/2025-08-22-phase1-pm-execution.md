# Phase 1: PM 패키지 리팩토링 실행 계획

## 개요
**목표**: cmd/pm 패키지의 파일들을 기능별 디렉터리로 이동
**소요시간**: 약 2시간
**복잡도**: 낮음
**우선순위**: 1순위 (가장 간단)

## 현재 상태 분석

### 디렉터리 구조
```
cmd/pm/
├── pm.go              # 루트 커맨드 조립
├── advanced.go        # 고급 기능
├── advanced/          # (빈 디렉터리)
├── cache.go           # 캐시 관리
├── cache/             # (빈 디렉터리)
├── doctor.go          # 진단 기능
├── doctor/            # (빈 디렉터리)
├── export.go          # 설정 내보내기
├── export/            # (빈 디렉터리)
├── install.go         # 패키지 설치
├── install/           # (빈 디렉터리)
├── status.go          # 상태 확인
├── status/            # (빈 디렉터리)
├── update.go          # 업데이트 관리
└── update/            # (빈 디렉터리)
```

### 문제점
- 디렉터리는 생성되었으나 파일들이 이동되지 않음
- 파일과 디렉터리가 중복으로 존재
- package 선언이 분산될 가능성

## 실행 계획

### 1단계: 현재 상태 확인 (5분)
```bash
# 파일 존재 확인
ls -la /home/archmagece/myopen/Gizzahub/gzh-cli/cmd/pm/

# 빌드 상태 확인
cd /home/archmagece/myopen/Gizzahub/gzh-cli
go build ./cmd/pm
```

### 2단계: 파일 이동 전략 결정 (10분)

#### 방식 A: 파일 유지 + package 통일
- 각 파일을 해당 디렉터리로 이동
- 모든 파일의 package를 `pm`으로 유지
- Go의 동일 패키지 분산 디렉터리 기능 활용

#### 방식 B: 서브패키지화
- 각 디렉터리를 독립 패키지로 분리
- package 이름을 기능별로 변경
- 루트에서 import하여 조립

**선택**: 방식 A (원본 계획서 방침에 따라)

### 3단계: 파일 이동 실행 (30분)

#### 이동 매핑
```bash
# 파일 이동
mv cmd/pm/advanced.go cmd/pm/advanced/
mv cmd/pm/cache.go cmd/pm/cache/
mv cmd/pm/doctor.go cmd/pm/doctor/
mv cmd/pm/export.go cmd/pm/export/
mv cmd/pm/install.go cmd/pm/install/
mv cmd/pm/status.go cmd/pm/status/
mv cmd/pm/update.go cmd/pm/update/
```

#### 주의사항
- 루트 `pm.go`는 이동하지 않음 (커맨드 조립 역할)
- 각 파일의 package 선언은 `package pm` 유지
- import 경로는 변경하지 않음

### 4단계: 검증 및 테스트 (45분)

#### 빌드 검증
```bash
# 빌드 테스트
go build ./cmd/pm

# 전체 빌드 테스트
go build ./...
```

#### 기능 테스트
```bash
# 주요 명령어 테스트
./gz pm status          # 상태 확인
./gz pm install --help  # 도움말 확인
./gz pm cache list      # 캐시 목록
./gz pm doctor          # 진단 실행
```

#### 단위 테스트
```bash
# PM 패키지 테스트
go test ./cmd/pm -v

# 관련 테스트들
go test ./cmd/pm/... -v
```

### 5단계: 정리 및 커밋 (30분)

#### 최종 구조 확인
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

#### Git 커밋
```bash
git add cmd/pm/
git commit -m "refactor(pm): reorganize files into feature directories

- Move feature-specific files to subdirectories
- Maintain package pm namespace across directories
- Improve code navigation and organization

Files moved:
- advanced.go → advanced/
- cache.go → cache/
- doctor.go → doctor/
- export.go → export/
- install.go → install/
- status.go → status/
- update.go → update/"
```

## 검증 체크리스트

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

### 문제 발생 시
```bash
# 변경사항 되돌리기
git checkout -- cmd/pm/

# 또는 커밋 전이라면
git restore cmd/pm/
```

### 부분 롤백
- 특정 파일만 문제가 있다면 해당 파일만 원위치로 복구
- 단계별로 이동했다면 문제 파일부터 역순으로 복구

## 예상 문제 및 해결책

### 문제 1: 빌드 에러
**원인**: package 선언 불일치
**해결**: 모든 파일의 package가 `pm`인지 확인

### 문제 2: import 에러
**원인**: 상대 경로 import 문제
**해결**: 절대 경로로 import 수정

### 문제 3: 테스트 실패
**원인**: 테스트 파일 경로 문제
**해결**: 테스트 파일도 함께 이동 확인

## 성공 기준
1. **기능 보존**: 모든 pm 하위 명령어가 정상 동작
2. **구조 개선**: 기능별 파일 그룹핑으로 탐색성 향상
3. **빌드 성공**: 컴파일 에러 없음
4. **테스트 통과**: 기존 테스트 모두 통과

## 다음 단계
Phase 1 완료 후 → [Phase 2: repo-config 실행 계획](./2025-08-22-phase2-repo-config-execution.md)
