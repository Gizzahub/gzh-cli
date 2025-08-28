# Phase 2: 테스트 매트릭스 자동화

## 📋 개요

**목표**: 모든 전략/제공자/가시성 조합을 자동으로 테스트  
**소요 시간**: 5일  
**우선순위**: 높음  
**전제 조건**: Phase 1 완료 (Mock Repository Factory)

## 🎯 구현 범위

- **기반**: Phase 1의 Mock Repository를 활용
- **목적**: synclone의 모든 조합을 체계적으로 검증
- **확장**: 기존 E2E 테스트 패턴을 자동화로 발전

## 📊 테스트 매트릭스 (75개 조합)

### 핵심 매트릭스
```
전략 (5개) × 제공자 (3개) × 가시성 (5개) = 75개 조합

전략: reset, pull, fetch, rebase, clone
제공자: github, gitlab, gitea  
가시성: public, private, all, internal, none
```

### 조합 예시
```yaml
# 예시 1: GitHub + reset + public
provider: github
strategy: reset  
visibility: public
expected_result: success

# 예시 2: GitLab + rebase + private  
provider: gitlab
strategy: rebase
visibility: private
expected_result: auth_required

# 예시 3: Gitea + clone + internal
provider: gitea
strategy: clone
visibility: internal
expected_result: permission_check
```

## 📝 추가 테스트 케이스 (20개)

### 1. 설정 파일 검증 (5개)
- **유효한 YAML 설정**
  - 표준 형식 준수
  - 모든 필수 필드 포함
- **잘못된 YAML 구문**
  - 인덴테이션 오류
  - 구문 오류
- **환경 변수 치환 테스트**
  - `${GITHUB_TOKEN}` 형식
  - 존재하지 않는 환경 변수
- **스키마 검증 실패**
  - 잘못된 필드명
  - 타입 불일치

### 2. 에러 처리 (8개)
- **존재하지 않는 조직/그룹**
  - 404 응답 처리
  - 적절한 에러 메시지
- **권한 없는 저장소 접근**
  - 403 응답 처리
  - 인증 요구 안내
- **네트워크 연결 실패**
  - 타임아웃 처리
  - 재시도 로직
- **API Rate Limit 초과**
  - Rate limit 감지
  - 백오프 전략
- **디스크 공간 부족**
  - 공간 체크
  - 부분 클론 정리
- **Git 명령어 실패**
  - Git 오류 파싱
  - 사용자 친화적 메시지
- **설정 파일 접근 권한 없음**
  - 파일 권한 체크
  - 대안 경로 제안
- **중복 실행 방지**
  - 락 파일 처리
  - 프로세스 충돌 방지

### 3. 성능 테스트 (7개)
- **다중 저장소 동시 처리**
  - 동시성 레벨 테스트 (1, 5, 10, 20개)
- **대용량 저장소 처리**
  - 1GB+ 저장소 클론 시간
  - 메모리 사용량 모니터링
- **메모리 사용량 추적**
  - 메모리 누수 감지
  - 피크 메모리 사용량
- **네트워크 대역폭 효율성**
  - 불필요한 다운로드 방지
  - 증분 업데이트 확인

## 🗂️ 구현 파일 구조

```
scripts/testing/synclone/
├── matrix/                      # 매트릭스 테스트 관련
│   ├── matrix-test.sh           # 메인 매트릭스 테스트 실행기
│   ├── matrix-config.yaml       # 매트릭스 설정 정의
│   ├── combination-generator.sh # 조합 자동 생성기
│   └── results-analyzer.sh      # 결과 분석기
├── templates/                   # 설정 파일 템플릿
│   ├── github-template.yaml     # GitHub 설정 템플릿
│   ├── gitlab-template.yaml     # GitLab 설정 템플릿
│   ├── gitea-template.yaml      # Gitea 설정 템플릿
│   └── base-template.yaml       # 기본 템플릿
├── validators/                  # 결과 검증 도구
│   ├── validate-clone.sh        # 클론 결과 검증
│   ├── validate-strategy.sh     # 전략 실행 결과 검증
│   ├── validate-config.sh       # 설정 파일 검증
│   └── validate-performance.sh  # 성능 검증
├── error-scenarios/             # 에러 시나리오 테스트
│   ├── network-errors.sh        # 네트워크 오류 시뮬레이션
│   ├── auth-errors.sh           # 인증 오류 테스트
│   ├── permission-errors.sh     # 권한 오류 테스트
│   └── resource-errors.sh       # 리소스 오류 테스트
├── performance/                 # 성능 테스트
│   ├── load-test.sh             # 부하 테스트
│   ├── memory-test.sh           # 메모리 테스트
│   ├── concurrent-test.sh       # 동시성 테스트
│   └── benchmark-runner.sh      # 벤치마크 실행기
├── reports/                     # 테스트 결과 보고서
│   ├── generate-report.sh       # 보고서 생성기
│   ├── html-template.html       # HTML 보고서 템플릿
│   └── json-schema.json         # JSON 결과 스키마
└── phase2-runner.sh             # Phase 2 통합 실행기
```

## 📋 구체적인 구현 계획

### Day 1: 매트릭스 인프라 구축
- [ ] 매트릭스 설정 정의 (`matrix-config.yaml`)
- [ ] 조합 자동 생성기 (`combination-generator.sh`)
- [ ] 기본 템플릿 시스템 구축
- [ ] 설정 파일 동적 생성 로직

### Day 2: 핵심 매트릭스 테스트 (75개 조합)
- [ ] `matrix-test.sh` 메인 실행기 작성
- [ ] 각 제공자별 템플릿 완성
- [ ] 전략별 검증 로직 구현
- [ ] 가시성별 테스트 케이스 구현

### Day 3: 에러 처리 테스트 (8개)
- [ ] 네트워크 오류 시뮬레이션
- [ ] 인증/권한 오류 테스트
- [ ] API Rate Limit 처리 테스트
- [ ] 리소스 부족 상황 테스트

### Day 4: 성능 테스트 (7개)
- [ ] 동시성 테스트 구현
- [ ] 메모리 사용량 모니터링
- [ ] 대용량 저장소 처리 테스트
- [ ] 성능 벤치마크 기준 설정

### Day 5: 통합 및 검증
- [ ] 결과 분석기 구현
- [ ] HTML/JSON 보고서 생성
- [ ] 전체 95개 케이스 검증
- [ ] CI/CD 통합 준비

## ✅ 성공 기준

- [ ] **75개 매트릭스 조합**이 자동으로 테스트됨
- [ ] **20개 추가 케이스** (설정, 에러, 성능) 검증
- [ ] 각 조합의 **성공/실패가 명확히 기록**됨
- [ ] 실패 케이스에 대한 **상세 로그 제공**
- [ ] 전체 테스트 실행 시간 **< 30분**
- [ ] **95% 이상 성공률** 달성 (정상 케이스)
- [ ] **HTML 보고서** 자동 생성

## 🧪 사용 예시

```bash
# 전체 매트릭스 테스트 실행
./scripts/testing/synclone/phase2-runner.sh

# 특정 제공자만 테스트
./scripts/testing/synclone/matrix/matrix-test.sh --provider github

# 성능 테스트만 실행  
./scripts/testing/synclone/performance/load-test.sh

# 에러 시나리오만 테스트
./scripts/testing/synclone/error-scenarios/network-errors.sh

# 결과 보고서 생성
./scripts/testing/synclone/reports/generate-report.sh
```

## 📊 예상 테스트 결과

### 매트릭스 조합 분석
```
성공 예상: 60/75 (80%)
- GitHub: 24/25 (인증 오류 1건 예상)
- GitLab: 23/25 (설정 이슈 2건 예상)  
- Gitea: 13/25 (호환성 이슈 12건 예상)

실패 분석: 15/75 (20%)
- 인증 관련: 5건
- 권한 관련: 4건
- 호환성 관련: 6건
```

### 성능 벤치마크 목표
```
단일 저장소 클론: < 30초
10개 동시 클론: < 2분
메모리 사용량: < 500MB
CPU 사용률: < 80%
```

## 🔧 기술 요구사항

### 확장된 도구
- **기존 요구사항**: Git, Bash, jq (Phase 1에서 계속)
- **새로 추가**:
  - `yq`: YAML 파싱 및 조작
  - `curl`: API 호출 테스트
  - `timeout`: 시간 제한 명령어
  - `parallel`: 병렬 처리 (GNU parallel)

### 환경 변수 확장
```bash
# Phase 1 변수 계속 사용
export GZ_TEST_REPOS_BASE="/tmp/gz-test-repos"

# Phase 2 추가 변수
export GZ_TEST_MATRIX_TIMEOUT="1800"          # 매트릭스 테스트 타임아웃 (30분)
export GZ_TEST_PARALLEL_JOBS="5"              # 동시 실행 작업 수
export GZ_TEST_PERFORMANCE_MODE="true"        # 성능 측정 활성화
export GZ_TEST_REPORT_FORMAT="html,json"      # 보고서 형식
```

## 🚀 다음 단계 연계

Phase 2 완료 후:
- **Phase 3**: Docker 격리 환경에서 실제 Git 서버 연동 테스트
- **CI/CD 통합**: GitHub Actions에서 자동 실행
- **성능 기준선 설정**: 향후 성능 회귀 감지를 위한 baseline 생성

---

**작성일**: 2025-08-28  
**예상 시작**: Phase 1 완료 후 (2025-09-01)  
**예상 완료**: 2025-09-06  
**담당자**: Development Team  
**문서 버전**: 1.0.0