# Task: Analyze Current Command Structure and Dependencies

## Objective
분석하여 현재 명령어 구조와 의존성을 파악하고, 통합 계획의 기초 자료를 만든다.

## Requirements
- [x] 모든 명령어의 현재 구조 분석
- [x] 각 명령어 간의 의존성 매핑
- [x] 공통 기능 식별
- [x] 중복 기능 목록화

## Steps

### 1. Command Structure Analysis
- [x] cmd/root.go 파일 분석하여 모든 등록된 명령어 목록 작성
- [x] 각 명령어별 하위 명령어 구조 문서화
- [x] 각 명령어의 주요 기능 요약

### 2. Dependency Mapping
- [x] 각 명령어가 사용하는 pkg/ 패키지 매핑
- [x] 각 명령어가 사용하는 internal/ 패키지 매핑
- [x] 공통으로 사용되는 헬퍼 함수 식별

### 3. Common Functionality Identification
- [x] 설정 파일 관리 기능 (config)
- [x] 검증 기능 (validate/doctor)
- [x] Git 관련 기능 (ssh, webhook, event)
- [x] 환경 관리 기능 (dev-env, net-env)

### 4. Duplicate Functionality List
- [x] gen-config vs synclone config generation
- [x] repo-config vs repo-sync configuration
- [x] event vs webhook
- [x] ssh-config vs dev-env ssh
- [x] doctor vs validate commands

## Expected Output
- `/docs/analysis/command-structure.md` - 현재 명령어 구조 분석 문서
- `/docs/analysis/dependency-map.md` - 의존성 매핑 문서
- `/docs/analysis/consolidation-candidates.md` - 통합 대상 명령어 목록

## Verification Criteria
- [x] 모든 현재 명령어가 문서화됨
- [x] 의존성 관계가 명확히 표시됨
- [x] 통합 가능한 명령어가 식별됨
- [x] 공통 기능이 카테고리별로 분류됨

## Notes
- 이 분석은 후속 작업의 기초가 되므로 철저히 수행
- 기존 사용자의 워크플로우 영향도 고려
- 백워드 호환성 요구사항 파악