# Phase 1: Mock Repository Factory

## 📋 개요

**목표**: 다양한 Git 저장소 상황을 시뮬레이션하는 테스트용 저장소 생성  
**소요 시간**: 3일  
**우선순위**: 높음 (즉시 구현 필요)

## 🎯 구현 범위

- **위치**: `scripts/testing/synclone/`
- **목적**: synclone 커맨드가 다양한 상황에서 올바르게 동작하는지 검증하기 위한 테스트 저장소 팩토리

## 📝 테스트 케이스 (15개)

### 1. 기본 저장소 유형 (3개)
- **빈 저장소** (fresh repository)
  - 초기화만 된 상태
  - README.md만 있는 상태
- **커밋이 있는 표준 저장소**
  - 여러 커밋 히스토리
  - 표준적인 파일 구조
- **대용량 저장소** (100MB+ 파일 포함)
  - 바이너리 파일 포함
  - 클론 시간 테스트용

### 2. 브랜치 상황 (3개)
- **단일 main 브랜치**
  - 가장 기본적인 구조
- **다중 브랜치** (main, develop, feature/*)
  - 실제 개발 환경과 유사
- **기본 브랜치가 master인 경우**
  - 레거시 저장소 시뮬레이션

### 3. 충돌 시나리오 (6개)
- **로컬 변경사항이 있는 저장소**
  - Uncommitted changes
  - Staged changes
- **리모트와 로컬이 diverged 상태**
  - 서로 다른 커밋 추가
- **머지 충돌 상태**
  - 동일 파일의 다른 라인 수정
  - 동일 파일의 같은 라인 수정
- **Untracked files 존재**
- **Stash 항목이 있는 상태**

### 4. 특수 상황 (3개)
- **Git LFS 파일이 있는 저장소**
  - 대용량 파일 관리 테스트
- **Submodule이 있는 저장소**
  - 복잡한 의존성 구조
- **네트워크 오류 시뮬레이션**
  - 클론 중단 상황

## 🗂️ 구현 파일 구조

```
scripts/testing/synclone/
├── setup-test-repos.sh          # 메인 테스트 저장소 생성 스크립트
├── scenarios/                   # 시나리오별 설정 스크립트
│   ├── basic-repos.sh           # 기본 저장소 생성 (3개)
│   ├── branch-repos.sh          # 브랜치 상황 생성 (3개)
│   ├── conflict-repos.sh        # 충돌 상황 생성 (6개)
│   └── special-repos.sh         # 특수 상황 생성 (3개)
├── templates/                   # 저장소 템플릿 파일
│   ├── README-template.md       # 기본 README 템플릿
│   ├── large-file-generator.sh  # 대용량 파일 생성
│   └── .gitlfs-template         # Git LFS 설정 템플릿
├── utils/                       # 유틸리티 스크립트
│   ├── git-helpers.sh           # Git 조작 헬퍼 함수
│   ├── validation-helpers.sh    # 저장소 상태 검증 함수
│   └── cleanup-helpers.sh       # 정리 헬퍼 함수
├── cleanup-test-repos.sh        # 테스트 저장소 정리 스크립트
├── run-scenario-tests.sh        # 시나리오 테스트 실행기
└── README.md                    # Phase 1 사용 가이드
```

## 📋 구체적인 구현 계획

### Day 1: 기초 인프라 구축
- [ ] `scripts/testing/synclone/` 디렉터리 생성
- [ ] 기본 파일 구조 생성
- [ ] `setup-test-repos.sh` 메인 스크립트 작성
- [ ] `git-helpers.sh` 유틸리티 함수 작성

### Day 2: 시나리오 구현
- [ ] `basic-repos.sh` - 기본 저장소 3개 시나리오
- [ ] `branch-repos.sh` - 브랜치 상황 3개 시나리오  
- [ ] `conflict-repos.sh` - 충돌 상황 6개 시나리오

### Day 3: 특수 케이스 및 검증
- [ ] `special-repos.sh` - 특수 상황 3개 시나리오
- [ ] `validation-helpers.sh` - 검증 함수 작성
- [ ] `run-scenario-tests.sh` - 통합 테스트 실행기
- [ ] 전체 시나리오 테스트 및 검증

## ✅ 성공 기준

- [ ] **15개 테스트 저장소**가 자동 생성됨
- [ ] 각 저장소는 **예상된 상태를 정확히 반영**
- [ ] 생성된 저장소로 synclone 명령어 실행 시 **예상 결과 도출**
- [ ] **정리 스크립트**로 테스트 환경 완전 초기화 가능
- [ ] **재실행 가능**한 멱등성 보장

## 🧪 사용 예시

```bash
# 모든 테스트 저장소 생성
./scripts/testing/synclone/setup-test-repos.sh

# 특정 시나리오만 생성
./scripts/testing/synclone/scenarios/basic-repos.sh

# 시나리오 테스트 실행
./scripts/testing/synclone/run-scenario-tests.sh

# 테스트 저장소 정리
./scripts/testing/synclone/cleanup-test-repos.sh
```

## 🔧 기술 요구사항

### 필요 도구
- **Git 2.0+**: 기본 Git 명령어
- **Bash 4.0+**: 스크립트 실행 환경
- **jq**: JSON 파싱 (설정 파일 처리)
- **Git LFS**: LFS 파일 테스트용 (선택사항)

### 환경 변수
```bash
export GZ_TEST_REPOS_BASE="/tmp/gz-test-repos"  # 테스트 저장소 기본 경로
export GZ_TEST_CLEANUP_AUTO="true"              # 자동 정리 활성화
export GZ_TEST_VERBOSE="false"                  # 상세 로그 출력
```

## 📊 검증 방법

### 자동 검증 항목
1. **저장소 상태 확인**
   - 브랜치 개수 및 이름
   - 커밋 히스토리 길이
   - 파일 개수 및 크기

2. **Git 상태 검증**
   - Working directory 상태
   - Staging area 상태
   - Remote tracking 상태

3. **특수 기능 확인**
   - Git LFS 파일 인식
   - Submodule 초기화 상태
   - 네트워크 오류 시뮬레이션 동작

## 🚀 다음 단계 연계

Phase 1 완료 후:
- **Phase 2**: 생성된 Mock Repository를 활용한 매트릭스 테스트
- **기존 E2E 테스트 통합**: `test/e2e/scenarios/synclone_e2e_test.go`와 연계
- **CI/CD 통합**: 자동화된 테스트 파이프라인에 포함

---

**작성일**: 2025-08-28  
**예상 완료**: 2025-08-31  
**담당자**: Development Team  
**문서 버전**: 1.0.0