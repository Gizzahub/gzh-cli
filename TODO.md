# TODO

## 중요!!
테스트 할 때 Gizzahub에다 하지 않기 ^^
동작 잘 하는거 확인했다 ^^b al

---

## 🔧 gzh.yaml 스키마 기반 설정 시스템 구현

### 📋 핵심 설정 파서 구현
[x] gzh.yaml 스키마 정의 문서 작성 (YAML/JSON Schema 형식)
[x] Config, Provider, GitTarget 구조체 구현 (`pkg/config/schema.go`)
[x] YAML 파서 및 환경변수 치환 기능 구현 (`os.ExpandEnv` 활용)
[x] 설정 파일 탐색 로직 구현 (우선순위: `./gzh.yaml` → `~/.config/gzh.yaml`)
[x] 설정 파일 유효성 검증 기능 구현 (필수 필드, enum 값 검증)

### 🔄 bulk-clone 명령어 gzh.yaml 통합
[x] 기존 bulk-clone 설정과 gzh.yaml 스키마 호환성 분석
[x] gzh.yaml 기반 bulk-clone 실행 옵션 추가 (`--use-gzh-config`)
[x] provider별 조직/그룹 일괄 클론 기능 구현
[x] visibility 필터링 로직 구현 (public/private/all)
[x] 정규식 기반 리포지토리 필터링 구현 (`match` 필드)
[x] flatten 옵션에 따른 디렉토리 구조 생성 로직 구현

### 🧪 테스트 및 문서화
[x] gzh.yaml 파서 단위 테스트 작성
[x] 다양한 설정 시나리오별 통합 테스트 작성
[x] gzh.yaml 사용 가이드 및 예제 문서 작성
[x] 마이그레이션 가이드 작성 (기존 bulk-clone.yaml → gzh.yaml)

💡 **추가 제안 기능:**
[x] `gz config validate` - gzh.yaml 유효성 검사 명령어
[x] `gz config init` - 대화형 gzh.yaml 생성 도구
[x] 설정 프로필 기능 (dev/prod 환경별 설정 분리)

---

## 🚀 GitHub Organization & Repository 관리 기능

### 📋 기본 설계 및 API 연동
[x] GitHub 리포지토리 설정 관리 요구사항 상세 정리
[x] `gz repo-config` 명령어 구조 설계 (list/apply/validate 서브커맨드)
[x] GitHub API 클라이언트 래퍼 구현 (`pkg/github/repo_config.go`)
[x] 리포지토리 설정 스키마 정의 (YAML 형식)
[ ] API Rate Limiting 처리 로직 구현 (재시도, 대기 시간 계산)

### ⚙️ 리포지토리 설정 관리 구현
[ ] 리포지토리 현재 설정 조회 기능 구현 (`repos.get` API)
[ ] 리포지토리 설정 일괄 업데이트 기능 구현 (`repos.update` API)
[ ] 조직 내 모든 리포지토리 대상 일괄 적용 기능
[ ] 설정 변경 이력 추적 및 롤백 기능 설계
[ ] Dry-run 모드 구현 (변경사항 미리보기)

### 🔐 보안 및 권한 관리
[ ] 필요한 GitHub 토큰 권한 문서화 (repos, admin:org)
[ ] 토큰 권한 자동 검증 기능 구현
[ ] 민감한 설정 변경 시 확인 프롬프트 추가
[ ] 설정 변경 로그 기록 기능 구현

### 📊 정책 템플릿 시스템
[ ] 기본 정책 템플릿 작성 (보안 강화, 오픈소스, 엔터프라이즈)
[ ] 정책 템플릿 상속 및 오버라이드 기능 구현
[ ] 리포지토리별 예외 처리 기능 구현
[ ] 정책 준수 여부 감사 리포트 생성 기능

### 🧪 테스트 및 문서화
[ ] GitHub API 모킹을 활용한 단위 테스트 작성
[ ] 실제 테스트 조직을 활용한 통합 테스트 시나리오 작성
[ ] 사용자 가이드 및 정책 템플릿 예제 문서 작성
[ ] Terraform 대안 비교 문서 작성

💡 **추가 제안 기능:**
- `gz repo-config diff` - 현재 설정과 목표 설정 비교
- `gz repo-config audit` - 조직 전체 정책 준수 리포트
- 웹훅 설정 관리 기능
- 브랜치 보호 규칙 일괄 관리
- GitHub Actions 권한 정책 관리

---

## 🔍 기존 기능 개선사항

### 📦 bulk-clone 성능 개선
[ ] 병렬 클론 옵션 추가 (goroutine 활용)
[ ] 중단된 작업 재개 기능 구현 (상태 저장)
[ ] 프로그레스 바 세분화 (리포지토리별 진행률)

### 🔧 설정 시스템 통합
[ ] 모든 명령어에 대한 통합 설정 파일 체계 설계
[ ] 설정 우선순위 문서화 (CLI 플래그 > 환경변수 > 설정파일)
[ ] 설정 마이그레이션 도구 구현
