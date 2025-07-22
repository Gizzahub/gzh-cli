# ⭐ 핵심 기능

gzh-manager-go의 주요 기능들을 카테고리별로 정리한 문서입니다.

## 📋 기능 카테고리

### 📦 대량 클론 (bulk-clone)

여러 Git 플랫폼에서 조직/그룹 단위로 저장소를 일괄 복제하는 기능입니다.

**지원 플랫폼:**

- GitHub (조직, 개인 저장소)
- GitLab (그룹, 프로젝트)
- Gitea (조직)
- Gogs (예정)

**주요 기능:**

- 멀티 플랫폼 동시 클론
- 증분 업데이트 (pull, fetch, reset)
- 병렬 처리로 성능 최적화
- 필터링 및 제외 규칙

[📁 bulk-clone 상세 문서](bulk-clone/)

### 🌐 네트워크 환경 관리 (network-management)

개발 환경에 따른 네트워크 설정을 자동으로 관리하는 기능입니다.

**주요 기능:**

- WiFi 환경 감지 및 자동 프로필 전환
- 프록시 설정 관리
- VPN 연결 자동화
- DNS 설정 변경
- Docker/Kubernetes 네트워크 프로필

[📁 network-management 상세 문서](network-management/)

### 📦 저장소 관리 (repository-management)

Git 저장소의 설정과 정책을 조직 단위로 관리하는 기능입니다.

**주요 기능:**

- 브랜치 보호 정책 일괄 적용
- 저장소 설정 감사 및 보고
- GitHub Actions 권한 정책
- 웹훅 설정 관리
- 의존성 관리 정책 (Dependabot)

[📁 repository-management 상세 문서](repository-management/)

### 💻 개발 환경 (development-environment)

클라우드 프로바이더와 개발 도구 설정을 관리하는 기능입니다.

**지원 플랫폼:**

- AWS (프로필, 리전, 자격증명)
- GCP (프로젝트, 서비스 계정)
- Azure (구독, 리소스 그룹)

**주요 기능:**

- 클라우드 프로필 자동 전환
- 환경별 설정 분리
- 자격증명 안전 관리
- 컨테이너 개발 환경 설정

[📁 development-environment 상세 문서](development-environment/)

## 🚀 일반적인 사용 시나리오

### 🏢 조직 관리자

```bash
# 1. 전체 조직 저장소 클론
gz bulk-clone --org company-org

# 2. 보안 정책 일괄 적용
gz repo-config apply --org company-org --policy security.yaml

# 3. 정책 준수 상태 감사
gz repo-config audit --org company-org --output report.html
```

### 👨‍💻 개발자

```bash
# 1. 개발 환경별 프로젝트 클론
gz bulk-clone --config dev-projects.yaml

# 2. 네트워크 환경 자동 전환
gz net-env auto-switch

# 3. 클라우드 환경 동기화
gz dev-env sync --profile development
```

### 🔧 DevOps 엔지니어

```bash
# 1. 인프라 저장소 일괄 관리
gz bulk-clone --org infrastructure --include "*-terraform" --include "*-k8s"

# 2. CI/CD 정책 강제
gz repo-config apply --policy actions-security.yaml

# 3. 네트워크 정책 배포
gz net-env deploy --k8s-namespace production
```

## 🎯 기능별 명령어 매핑

| 기능 영역       | 주요 명령어      | 설명               |
| --------------- | ---------------- | ------------------ |
| **대량 클론**   | `gz bulk-clone`  | 저장소 일괄 복제   |
| **저장소 관리** | `gz repo-config` | 저장소 설정 관리   |
| **네트워크**    | `gz net-env`     | 네트워크 환경 관리 |
| **개발 환경**   | `gz dev-env`     | 클라우드/개발 환경 |
| **설정**        | `gz config`      | 전역 설정 관리     |

## 🔗 통합 워크플로우

### 완전한 개발 환경 설정

```bash
# 1단계: 기본 설정
gz config init --template team-development

# 2단계: 저장소 클론
gz bulk-clone --config team-repos.yaml

# 3단계: 네트워크 환경 설정
gz net-env profile create office --dns 192.168.1.1 --proxy http://proxy:8080

# 4단계: 클라우드 프로필 설정
gz dev-env aws configure --profile dev --region ap-northeast-2

# 5단계: 정책 적용
gz repo-config apply --org my-team --policy team-standards.yaml
```

### CI/CD 파이프라인 통합

```yaml
# .github/workflows/repo-audit.yml
name: Repository Audit
on:
  schedule:
    - cron: "0 9 * * 1" # 매주 월요일 오전 9시

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install gzh-manager-go
        run: curl -sSL https://install.gzh.dev | sh
      - name: Run audit
        run: |
          gz repo-config audit --org ${{ github.repository_owner }} \
            --policy .github/policies/security.yaml \
            --output audit-report.html
      - name: Upload report
        uses: actions/upload-artifact@v3
        with:
          name: audit-report
          path: audit-report.html
```

## 📊 성능 및 스케일링

### 대규모 조직 지원

- **동시 처리**: 최대 100개 저장소 병렬 클론
- **메모리 최적화**: 스트리밍 API로 대용량 데이터 처리
- **재시도 메커니즘**: 네트워크 오류 자동 복구
- **진행률 추적**: 실시간 진행 상황 표시

### 성능 최적화 팁

```yaml
# gzh.yaml - 성능 최적화 설정
performance:
  concurrent_operations: 10
  timeout: "5m"
  retry_attempts: 3
  batch_size: 50

cache:
  enabled: true
  ttl: "1h"
  max_size: "500MB"
```

## 🛡️ 보안 고려사항

### 자격증명 관리

- 환경 변수 기반 토큰 관리
- 키체인/크리덴셜 매니저 통합
- 토큰 자동 갱신 지원
- 최소 권한 원칙 적용

### 감사 및 컴플라이언스

- 모든 작업 로그 기록
- 정책 위반 알림
- 접근 권한 추적
- SOC2/ISO27001 준수

## 📚 추가 리소스

### 설정 예제

- [개인 개발자용](../examples/gzh-simple.yaml)
- [팀 개발용](../examples/gzh-development.yaml)
- [엔터프라이즈용](../examples/gzh-enterprise.yaml)

### 통합 가이드

- [GitHub Actions 통합](../08-integrations/)
- [Terraform 대안](../08-integrations/terraform-alternative-comparison.md)

---

_💡 각 기능의 상세한 사용법은 해당 하위 디렉토리의 문서를 참조하세요._
