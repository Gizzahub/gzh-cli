# 엔터프라이즈 기능 백로그

이 디렉토리는 gzh-cli에서 제거된 엔터프라이즈 기능들의 기능 명세를 백로그 형태로 보관합니다.

## 제거 이유

CLI 도구의 핵심 목적에 집중하고 과도한 복잡성을 제거하기 위해 다음 기능들을 백로그로 이동했습니다:

- 대부분의 사용자가 필요로 하지 않는 고급 엔터프라이즈 기능
- 별도의 전문 도구로 더 잘 해결되는 기능들
- 유지보수 복잡도를 크게 증가시키는 기능들

## 백로그 구조

### infrastructure/

인프라 관리 및 자동화 도구들

- **terraform-management.md**: Terraform IaC 통합 관리 기능
- **cloudformation-management.md**: AWS CloudFormation 스택 관리
- **ansible-management.md**: Ansible 배포 자동화 기능

### ci-cd/

CI/CD 파이프라인 관리 및 자동화 도구들

- **jenkins-integration.md**: Jenkins 파이프라인 관리 및 자동화
- **github-actions-management.md**: GitHub Actions 워크플로우 관리
- **gitlab-ci-management.md**: GitLab CI/CD 파이프라인 관리

### monitoring/

고급 모니터링, 로깅, 추적 및 관찰성 도구들

- **observability-platform.md**: 종합적인 모니터링 및 관찰성 플랫폼

### platform/

플랫폼 생태계 및 확장성 기능들

- **plugin-ecosystem.md**: 확장 가능한 플러그인 아키텍처 및 마켓플레이스
- **template-marketplace.md**: 프로젝트 템플릿 및 코드 생성기 마켓플레이스

### web-services/

웹 서비스, API 및 대시보드 기능들

- **dashboard-platform.md**: 실시간 모니터링 웹 대시보드 플랫폼

## 복원 가이드

이 기능들은 필요시 다음과 같이 복원할 수 있습니다:

1. 해당 기능의 명세서를 검토
2. 현재 아키텍처와의 호환성 확인
3. 의존성 및 복잡도 영향 분석
4. 단계적 구현 계획 수립

## 대안 도구 권장사항

제거된 각 기능에 대해 더 전문적인 대안 도구들을 권장사항으로 포함했습니다.
