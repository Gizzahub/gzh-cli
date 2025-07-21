# 📚 gzh-manager-go 문서 가이드

> **gzh-manager-go**는 개발 환경과 Git 저장소를 종합적으로 관리하는 Go 기반 CLI 도구입니다.

## 🚀 빠른 시작

### 핵심 문서
- [📖 설치 및 시작하기](01-getting-started/)
- [🏗️ 아키텍처 개요](02-architecture/overview.md)
- [⚙️ 설정 가이드](04-configuration/configuration-guide.md)

### 주요 기능
- **대량 클론**: GitHub/GitLab 조직 전체 저장소 일괄 복제
- **네트워크 관리**: WiFi 프로필, VPN, 프록시 자동 전환
- **개발 환경**: AWS/GCP/Azure 클라우드 프로필 관리
- **저장소 설정**: 조직 단위 정책 및 설정 관리

---

## 📋 전체 문서 목록

### 🎯 1. 시작하기
- [📁 01-getting-started/](01-getting-started/)
  - [🔄 마이그레이션 가이드](01-getting-started/migration-guides/)
    - [bulk-clone → gzh 마이그레이션](01-getting-started/migration-guides/bulk-clone-to-gzh.md)
    - [daemon → CLI 마이그레이션](01-getting-started/migration-guides/daemon-to-cli.md)
    - [migrate 명령어 가이드](01-getting-started/migration-guides/migrate-command.md)

### 🏗️ 2. 아키텍처 및 설계
- [📁 02-architecture/](02-architecture/)
  - [🏛️ 프로젝트 개요](02-architecture/overview.md)
  - [🐳 개발 컨테이너](02-architecture/development-container.md)

### ⭐ 3. 핵심 기능
- [📁 03-core-features/](03-core-features/)

  #### 🔄 대량 클론
  - [📁 bulk-clone/](03-core-features/bulk-clone/)

  #### 🌐 네트워크 관리
  - [📁 network-management/](03-core-features/network-management/)
    - [🐳 Docker 프로필](03-core-features/network-management/docker-profiles.md)
    - [☸️ Kubernetes 정책](03-core-features/network-management/kubernetes-policies.md)
    - [🌐 네트워크 액션](03-core-features/network-management/network-actions.md)

  #### 📦 저장소 관리
  - [📁 repository-management/](03-core-features/repository-management/)
    - [📋 사용자 가이드](03-core-features/repository-management/repo-config-user-guide.md)
    - [🔍 감사 보고서](03-core-features/repository-management/repo-config-audit-report.md)
    - [⚡ 빠른 시작](03-core-features/repository-management/repo-config-quick-start.md)
    - [💻 명령어 레퍼런스](03-core-features/repository-management/repo-config-commands.md)
    - [📜 정책 예제](03-core-features/repository-management/repo-config-policy-examples.md)
    - [🔄 Diff 가이드](03-core-features/repository-management/repo-config-diff-guide.md)
    - [🔌 API 레퍼런스](03-core-features/repository-management/repository-configuration-api.md)
    - **GitHub 통합**
      - [🔬 조직 관리 연구](03-core-features/repository-management/github/org-management-research.md)
      - [⏱️ 요청 제한](03-core-features/repository-management/github/rate-limiting.md)
      - [📋 관리 요구사항](03-core-features/repository-management/github/repo-management-requirements.md)
      - [🔐 권한 관리](03-core-features/repository-management/github/permissions.md)

  #### 💻 개발 환경
  - [📁 development-environment/](03-core-features/development-environment/)
    - [☁️ AWS 프로필](03-core-features/development-environment/aws-profiles.md)
    - [🌤️ GCP 프로젝트](03-core-features/development-environment/gcp-projects.md)

### ⚙️ 4. 설정 및 구성
- [📁 04-configuration/](04-configuration/)
  - [📖 설정 가이드](04-configuration/configuration-guide.md)
  - [🎯 우선순위 시스템](04-configuration/priority-system.md)
  - [🔄 핫 리로딩](04-configuration/hot-reloading.md)
  - [🔍 호환성 분석](04-configuration/compatibility-analysis.md)
  - [📝 YAML 가이드](04-configuration/yaml-guide.md)
  - [📊 설정 비교](04-configuration/configuration-comparison.md)
  - [🧪 우선순위 테스트](04-configuration/configuration-priority-test.md)
  - **스키마 참조**
    - [⚡ gzh 스키마](04-configuration/schemas/gzh-schema.yaml)
    - [📦 bulk-clone 스키마](04-configuration/schemas/bulk-clone-schema.yaml)
    - [🔧 repo-config 스키마](04-configuration/schemas/repo-config-schema.yaml)
    - [🎭 actions-policy 스키마](04-configuration/schemas/actions-policy-schema.md)

### 📖 5. API 레퍼런스
- [📁 05-api-reference/](05-api-reference/)
  - [🐛 디버그 API](05-api-reference/debug.md)

### 🛠️ 6. 개발 가이드
- [📁 06-development/](06-development/)
  - [🐛 디버깅 가이드](06-development/debugging-guide.md)
  - [🪝 Pre-commit 훅](06-development/pre-commit-hooks.md)
  - [🧪 모킹 전략](06-development/mocking-strategy.md)

### 🚀 7. 배포 및 운영
- [📁 07-deployment/](07-deployment/)
  - [📋 릴리스 준비 체크리스트](07-deployment/release-preparation-checklist.md)
  - [📦 릴리스 가이드](07-deployment/releases.md)
  - [📄 v1.0.0 릴리스 노트](07-deployment/release-notes-v1.0.0.md)
  - [🔒 보안 스캐닝](07-deployment/security-scanning.md)

### 🔗 8. 외부 통합
- [📁 08-integrations/](08-integrations/)
  - [🏗️ Terraform 대안 비교](08-integrations/terraform-alternative-comparison.md)
  - [📊 Terraform vs gz 예제](08-integrations/terraform-vs-gz-examples.md)
  - [🪝 웹훅 관리 가이드](08-integrations/webhook-management-guide.md)

### 🏢 9. 엔터프라이즈 기능
- [📁 09-enterprise/](09-enterprise/)
  - [🎭 Actions 정책 스키마](09-enterprise/actions-policy-schema.md)
  - [🛡️ Actions 정책 강제](09-enterprise/actions-policy-enforcement.md)

### 🔧 10. 유지보수
- [📁 10-maintenance/](10-maintenance/)
  - [📝 변경 로그](10-maintenance/changelog.md)
  - [🗺️ 로드맵](10-maintenance/roadmap.md)

### 📂 미분류 문서
- [📁 unclassified/](unclassified/)
  - [📋 문서 요약](unclassified/documentation-summary.md)

---

## 🎯 사용 시나리오별 가이드

### 🆕 처음 사용하는 경우
1. [설치 및 기본 설정](01-getting-started/)
2. [YAML 설정 가이드](04-configuration/yaml-guide.md)
3. [빠른 시작 튜토리얼](03-core-features/repository-management/repo-config-quick-start.md)

### 👥 팀 관리자인 경우
1. [저장소 관리](03-core-features/repository-management/)
2. [정책 설정](03-core-features/repository-management/repo-config-policy-examples.md)
3. [GitHub 조직 관리](03-core-features/repository-management/github/)

### 🏢 엔터프라이즈 사용자인 경우
1. [엔터프라이즈 기능](09-enterprise/)
2. [보안 정책](07-deployment/security-scanning.md)
3. [감사 및 컴플라이언스](03-core-features/repository-management/repo-config-audit-report.md)

### 🛠️ 개발자인 경우
1. [개발 가이드](06-development/)
2. [API 레퍼런스](05-api-reference/)
3. [아키텍처 문서](02-architecture/)

---

## 🔍 빠른 검색

### 명령어별 문서
- **bulk-clone**: [대량 클론](03-core-features/bulk-clone/)
- **repo-config**: [저장소 관리](03-core-features/repository-management/)
- **net-env**: [네트워크 관리](03-core-features/network-management/)
- **dev-env**: [개발 환경](03-core-features/development-environment/)

### 주제별 문서
- **설정**: [04-configuration/](04-configuration/)
- **GitHub**: [GitHub 통합](03-core-features/repository-management/github/)
- **배포**: [07-deployment/](07-deployment/)
- **문제 해결**: [디버깅 가이드](06-development/debugging-guide.md)

---

## 📚 관련 자료

### 프로젝트 메타 문서
- [📄 README](../README.md)
- [⭐ FEATURES](../FEATURES.md)
- [📋 USAGE](../USAGE.md)
- [🔒 SECURITY](../SECURITY.md)

### 개발 도구
- [🐳 Docker 설정](../Dockerfile)
- [🏗️ Makefile](../Makefile)
- [📦 의존성](../go.mod)

---

## 💡 기여 및 개선

문서 개선이나 오류 발견 시:
1. GitHub 이슈 생성
2. Pull Request 제출
3. [개발 가이드](06-development/) 참조

---

*📅 최종 업데이트: 2025-07-16*  
*📊 총 문서 수: 45개*  
*🏗️ 문서 구조: 10개 주요 카테고리*
