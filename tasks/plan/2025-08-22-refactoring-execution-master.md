# 코드 구조 리팩토링 실행 계획 (마스터)

## 개요
4개의 주요 cmd 패키지(ide, net-env, pm, repo-config)에 대한 단계적 리팩토링 실행 계획

**생성일**: 2025-08-22
**예상 총 소요시간**: 15시간
**실행 방식**: Phase별 순차 실행

## 배경
현재 다음 패키지들의 코드 구조 개선이 필요:
- `cmd/pm`: 파일별 기능이 명확하나 탐색성 개선 필요
- `cmd/repo-config`: 파일/라인 수가 많아 가독성 저하
- `cmd/ide`: 무거운 로직과 커맨드 조립이 혼재
- `cmd/net-env`: 명령 수와 테스트가 많고 환경 의존성 높음

## 실행 우선순위 및 일정

### Phase 1: pm 패키지 (2시간)
- **복잡도**: 낮음
- **현재 상태**: 디렉터리 구조 부분 생성됨
- **작업**: 파일 이동 및 정리

### Phase 2: repo-config 패키지 (3시간)
- **복잡도**: 중간
- **현재 상태**: 디렉터리 구조 부분 생성됨
- **작업**: 파일 이동 및 의존성 정리

### Phase 3: ide 패키지 (4시간)
- **복잡도**: 높음
- **현재 상태**: 평면 구조
- **작업**: internal 추출 + 서브패키지화

### Phase 4: net-env 패키지 (6시간)
- **복잡도**: 가장 높음
- **현재 상태**: 평면 구조, 많은 파일
- **작업**: 전면적 구조 재편

## 성공 기준

### 기능 보존
- [ ] 모든 기존 명령어 정상 동작
- [ ] 빌드 성공: `go build ./...`
- [ ] 테스트 통과: `go test ./...`
- [ ] CI/CD 파이프라인 통과

### 구조 개선
- [ ] 기능별 디렉터리 분리
- [ ] 관련 파일들의 논리적 그룹핑
- [ ] 의존성 정리 및 순환 참조 제거
- [ ] 코드 탐색성 향상

## 리스크 관리

### 주요 리스크
1. **빌드 실패**: 의존성 문제로 인한 컴파일 에러
2. **테스트 실패**: 환경 의존성으로 인한 테스트 실패
3. **기능 손실**: 파일 이동 중 누락
4. **순환 참조**: 패키지 분리 중 의존성 문제

### 완화 전략
- **단계별 검증**: 각 Phase마다 빌드/테스트 확인
- **Git 브랜치 전략**: Phase별 별도 브랜치 생성
- **체크리스트 활용**: 상세 검증 항목으로 누락 방지
- **롤백 계획**: 각 단계별 커밋으로 쉬운 되돌리기

## 롤백 계획

### Git 전략
```bash
# 각 Phase별 브랜치 생성
git checkout -b refactor-phase1-pm
git checkout -b refactor-phase2-repo-config
git checkout -b refactor-phase3-ide
git checkout -b refactor-phase4-net-env
```

### 백업 지점
- **시작 전**: `refactor-start` 태그 생성
- **각 Phase 완료**: `phase-N-completed` 태그 생성
- **문제 발생 시**: 해당 태그로 즉시 롤백

## 관련 문서

### 상세 실행 계획
- [Phase 1: PM 실행 계획](./2025-08-22-phase1-pm-execution.md)
- [Phase 2: repo-config 실행 계획](./2025-08-22-phase2-repo-config-execution.md)
- [Phase 3: IDE 실행 계획](./2025-08-22-phase3-ide-execution.md)
- [Phase 4: net-env 실행 계획](./2025-08-22-phase4-net-env-execution.md)

### 검증 도구
- [전체 체크리스트](./2025-08-22-refactoring-checklist.md)

### 원본 제안서
- [IDE internal 분리 제안](../issue/2025-08-22-ide-internal-depth-plan.md)
- [net-env depth 구조 제안](../issue/2025-08-22-net-env-depth-plan.md)
- [PM depth 구조 제안](../issue/2025-08-22-pm-depth-plan.md)
- [repo-config depth 구조 제안](../issue/2025-08-22-repo-config-depth-plan.md)

## 실행 지침

### 실행 전 준비
1. 현재 작업 내용 커밋
2. `refactor-start` 태그 생성
3. 별도 브랜치에서 작업

### 실행 중 원칙
- 한 번에 하나의 Phase만 진행
- 각 단계마다 빌드/테스트 확인
- 문제 발생 시 즉시 중단하고 롤백

### 완료 후 정리
- 성공한 브랜치를 main에 머지
- 실패한 브랜치는 폐기
- 문서 업데이트 및 이슈 클로즈

---

**다음 단계**: [Phase 1 PM 실행 계획](./2025-08-22-phase1-pm-execution.md) 검토 후 실행
