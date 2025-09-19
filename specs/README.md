# Specifications Directory

## 📋 목적과 성격

이 디렉토리는 gzh-cli 프로젝트의 \*\*공식 기능 명세서(Specifications)\*\*를 포함합니다.

**SDD (Specification-Driven Development)** 방법론을 따라 모든 CLI 명령어의 입출력 계약이 구현 전에 정의됩니다.

### 🎯 specs/의 핵심 원칙

1. **구현 기준점**: 모든 코드는 이 명세를 기반으로 개발되어야 합니다
1. **우선순위**: specs → 소스코드 → docs 순서로 권위를 가집니다
1. **테스트 가능**: 각 명세는 자동화된 검증이 가능해야 합니다
1. **진실의 원천(Source of Truth)**: 기능에 대한 논쟁이 있을 때 최종 기준이 됩니다

## 📁 디렉토리 구조

```bash
specs/
├── README.md                    # 이 문서
├── cli/                         # CLI 명령어 스펙 (SDD)
│   ├── README.md               # CLI 스펙 개요
│   ├── template.md             # 스펙 템플릿
│   ├── synclone/               # synclone 명령어
│   │   ├── UC-001-help.md
│   │   ├── UC-002-github-clone.md
│   │   ├── UC-003-rate-limit.md
│   │   ├── UC-004-auth-error.md
│   │   └── UC-005-pagination.md
│   ├── git/                    # git 명령어 그룹
│   │   ├── repo/
│   │   │   ├── UC-001-create.md
│   │   │   ├── UC-002-list.md
│   │   │   └── UC-003-clone-or-update.md
│   │   ├── webhook/
│   │   │   ├── UC-001-create.md
│   │   │   ├── UC-002-list.md
│   │   │   └── UC-003-delete.md
│   │   └── event/
│   │       └── UC-001-process.md
│   ├── quality/                # 코드 품질 명령어
│   │   └── UC-001-run.md
│   ├── doctor/                 # 시스템 진단
│   │   └── UC-001-diagnose.md
│   ├── pm/                     # 패키지 매니저
│   │   └── UC-001-update.md
│   ├── ide/                    # IDE 관리
│   │   └── UC-001-scan.md
│   ├── dev-env/                # 개발 환경
│   │   ├── UC-001-switch.md
│   │   ├── UC-002-status.md
│   │   └── UC-003-tui.md
│   ├── net-env/                # 네트워크 환경
│   │   ├── UC-001-switch.md
│   │   ├── UC-002-status.md
│   │   └── UC-003-profiles.md
│   ├── repo-config/            # 저장소 설정
│   │   ├── UC-001-apply.md
│   │   └── UC-002-validate.md
│   ├── profile/                # 프로파일링
│   │   └── UC-001-run.md
│   └── version/                # 버전 정보
│       └── UC-001-show.md
├── core/                        # 핵심 기능 명세
│   ├── git.md                  # Git 통합 아키텍처
│   ├── synclone.md             # Synclone 핵심 로직
│   └── synclone-git-extension.md
├── patterns/                    # 설계 패턴
│   ├── common.md               # 공통 패턴
│   └── compatibility-rules.md  # 호환성 규칙
└── testing/                     # 테스트 시나리오
    └── synclone/               # Synclone 테스트 케이스
        ├── README.md
        ├── test-scenarios.md
        ├── test-commands.md
        ├── test-data.md
        └── test-runner.sh
```

## 🔍 명세서 종류

### 1. CLI 명령어 스펙 (specs/cli/)

**SDD (Specification-Driven Development)** 방법론에 따른 CLI 명령어 입출력 계약 정의:

#### 완성된 명령어 스펙:

- **synclone/**: GitHub/GitLab 조직 대량 클론 (5개 UC)
- **git/repo/**: 저장소 생성, 목록, 클론/업데이트 (3개 UC)
- **quality/**: 코드 품질 검사 (1개 UC)
- **doctor/**: 시스템 진단 (1개 UC)
- **pm/**: 패키지 매니저 업데이트 (1개 UC)
- **ide/**: IDE 스캔 (1개 UC)
- **version/**: 버전 정보 (1개 UC)
- **dev-env/**: 개발 환경 전환, 상태 확인, TUI (3개 UC) ✅
- **net-env/**: 네트워크 환경 전환, 상태 확인, 프로필 관리 (3개 UC) ✅
- **repo-config/**: 저장소 설정 적용, 검증 (2개 UC) ✅
- **git/webhook/**: 웹훅 생성, 목록, 삭제 (3개 UC) ✅
- **git/event/**: Git 이벤트 처리 (1개 UC) ✅
- **profile/**: 성능 프로파일링 (1개 UC) ✅

#### 작성 예정 명령어 스펙:

현재 모든 주요 명령어 스펙이 완료되었습니다.

### 2. 핵심 기능 명세 (specs/core/)

시스템의 핵심 기능과 아키텍처 설계:

- **git.md**: Git 통합 명령어 아키텍처
- **synclone.md**: 저장소 동기화 및 클론 핵심 로직
- **synclone-git-extension.md**: Git extension 패턴

### 3. 설계 패턴 (specs/patterns/)

공통 설계 패턴과 호환성 규칙:

- **common.md**: 공통 패턴 및 구조
- **compatibility-rules.md**: 플랫폼 간 호환성 규칙

### 4. 테스트 시나리오 (specs/testing/)

실제 테스트 케이스와 시나리오:

- **synclone/test-scenarios.md**: 다양한 테스트 시나리오
- **synclone/test-commands.md**: 테스트 명령어 모음
- **synclone/test-data.md**: 테스트 데이터 정의
- **synclone/test-runner.sh**: 테스트 실행 스크립트

## 📊 CLI 명세 현황

### ✅ 완료된 명령어 (24개 UC)

| 명령어 | UC 수 | 상태 | 비고 |
|--------|-------|------|------|
| synclone | 5 | ✅ | Rate limit, 페이지네이션 포함 |
| git repo | 3 | ✅ | 생성, 목록, 클론/업데이트 |
| quality | 1 | ✅ | 다중 언어 품질 검사 |
| doctor | 1 | ✅ | 시스템 종합 진단 |
| pm | 1 | ✅ | 패키지 매니저 업데이트 |
| ide | 1 | ✅ | IDE 스캔 및 감지 |
| version | 1 | ✅ | 버전 정보 표시 |
| dev-env | 3 | ✅ | 환경 전환, 상태, TUI |
| net-env | 3 | ✅ | 네트워크 전환, 상태, 프로필 |
| repo-config | 2 | ✅ | 설정 적용, 검증 |
| git webhook | 3 | ✅ | 웹훅 생성, 목록, 삭제 |
| git event | 1 | ✅ | Git 이벤트 처리 |
| profile | 1 | ✅ | Go 애플리케이션 프로파일링 |

### 🚧 작성 예정 명령어

모든 주요 CLI 명령어 스펙이 완료되었습니다.

## 🚀 이전된 기능 명세서

아래 기능 명세서들은 **docs/30-features/** 로 이동되었습니다:

- ✅ dev-env-specification.md - 개발 환경 관리
- ✅ net-env-specification.md - 네트워크 환경 관리
- ✅ ide-specification.md - IDE 모니터링 및 관리
- ✅ quality-specification.md - 코드 품질 도구 통합
- ✅ profile-specification.md - 성능 프로파일링
- ✅ actions-policy-specification.md - GitHub Actions 정책 관리
- ✅ repo-config-specification.md - 저장소 설정 관리
- ✅ shell-specification.md - 인터랙티브 디버깅 셸
- ✅ package-manager-specification.md - 패키지 매니저 통합
- ✅ man-specification.md - 매뉴얼 페이지 생성

## 📖 관련 문서

- [CLI Specification Strategy](../docs/60-development/68-cli-specification-strategy.md) - SDD 방법론
- [Testing Strategy](../docs/60-development/67-testing-strategy.md) - 테스트 전략
- [Feature Specifications](../docs/30-features/) - 기능별 상세 명세서

## 🎯 향후 계획

### ✅ 전체 CLI 스펙 완료

- ✅ Phase 1: dev-env, net-env 명령어 (6개 UC)
- ✅ Phase 2: repo-config, git webhook, git event, profile 명령어 (7개 UC)
- ✅ 기존 명령어: synclone, git repo, quality, doctor, pm, ide, version (13개 UC)

**총 24개 UC 스펙 완료** - 모든 주요 CLI 명령어 커버

### Phase 3 (다음 단계)

- 자동화된 CLI 계약 테스트 구현
- 성능 테스트 시나리오 추가
- 엣지 케이스 및 에러 시나리오 보완
- 크로스 플랫폼 호환성 테스트 강화

______________________________________________________________________

**주의**: specs/ 디렉토리의 내용이 구현의 기준이 됩니다. 변경 시 관련 코드와 문서도 함께 업데이트해야 합니다.
