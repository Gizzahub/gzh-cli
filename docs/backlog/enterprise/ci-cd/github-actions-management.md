# GitHub Actions 관리 기능

## 개요

GitHub Actions 워크플로우 관리 및 자동화 기능

## 제거된 기능

### 1. 워크플로우 생성 및 관리

- **명령어**: `gz github-actions create`, `gz github-actions sync`
- **기능**: GitHub Actions 워크플로우 자동 생성 및 동기화
- **특징**:
  - 프로젝트 타입별 템플릿
  - 다중 환경 지원
  - 의존성 캐싱 최적화
  - 매트릭스 빌드 설정

### 2. 시크릿 관리

- **명령어**: `gz github-actions secrets set`, `gz github-actions secrets sync`
- **기능**: GitHub 저장소 및 조직 시크릿 관리
- **특징**:
  - 환경별 시크릿 분리
  - 암호화된 값 설정
  - 대량 시크릿 업데이트
  - 권한 기반 접근

### 3. 러너 관리

- **명령어**: `gz github-actions runner deploy`, `gz github-actions runner scale`
- **기능**: 셀프 호스팅 러너 관리
- **특징**:
  - 자동 러너 등록
  - 클라우드 기반 스케일링
  - 라벨 기반 작업 분배
  - 보안 그룹 관리

### 4. 워크플로우 모니터링

- **명령어**: `gz github-actions status`, `gz github-actions logs`
- **기능**: 워크플로우 실행 상태 모니터링
- **특징**:
  - 실시간 상태 추적
  - 로그 수집 및 분석
  - 실패 알림
  - 성능 메트릭

## 사용 예시 (제거 전)

```bash
# 새 워크플로우 생성
gz github-actions create --type nodejs \
  --repo owner/repository \
  --environments "dev,staging,prod"

# 시크릿 동기화
gz github-actions secrets sync --config secrets.yaml \
  --repo owner/repository

# 셀프 호스팅 러너 배포
gz github-actions runner deploy --count 3 \
  --labels "linux,docker" --instance-type t3.medium

# 워크플로우 상태 확인
gz github-actions status --repo owner/repository \
  --workflow-name ci.yml
```

## 설정 파일 형식

```yaml
github_actions:
  repositories:
    - name: myapp-backend
      owner: company
      workflows:
        - name: ci.yml
          triggers:
            - push: [master, main, develop]
            - pull_request: [master, main]
            - schedule: "0 2 * * *"

          jobs:
            test:
              runs_on: ubuntu-latest
              strategy:
                matrix:
                  node_version: [16, 18, 20]
              steps:
                - checkout
                - setup_node: ${{ matrix.node_version }}
                - install_dependencies
                - run_tests
                - upload_coverage

            build:
              needs: test
              runs_on: ubuntu-latest
              steps:
                - checkout
                - setup_node: 18
                - build_application
                - upload_artifacts

            deploy:
              needs: build
              runs_on: ubuntu-latest
              environment: production
              if: github.ref == 'refs/heads/main'
              steps:
                - download_artifacts
                - deploy_to_aws

  secrets:
    organization_level:
      AWS_ACCESS_KEY_ID: ${{ secrets.ORG_AWS_ACCESS_KEY }}
      AWS_SECRET_ACCESS_KEY: ${{ secrets.ORG_AWS_SECRET_KEY }}

    repository_level:
      - repo: myapp-backend
        secrets:
          DATABASE_URL: postgres://user:pass@host:5432/db
          API_KEY: secret-api-key

  environments:
    - name: development
      protection_rules:
        required_reviewers: 0
        wait_timer: 0

    - name: staging
      protection_rules:
        required_reviewers: 1
        wait_timer: 5

    - name: production
      protection_rules:
        required_reviewers: 2
        wait_timer: 30
        deployment_branches:
          - main

  runners:
    self_hosted:
      - name: linux-runners
        count: 3
        labels: [self-hosted, linux, docker]
        instance_type: t3.large
        auto_scaling:
          min: 1
          max: 10
          target_utilization: 80

    github_hosted:
      default: ubuntu-latest
      matrix:
        - ubuntu-latest
        - windows-latest
        - macos-latest

  notifications:
    slack:
      webhook_url: https://hooks.slack.com/...
      channels:
        success: "#deployments"
        failure: "#alerts"

    email:
      on_failure: [devops@company.com]
      on_success: [team@company.com]
```

## 워크플로우 템플릿

### 1. Node.js CI/CD

```yaml
name: Node.js CI/CD

on:
  push:
    branches: [master, main, develop]
  pull_request:
    branches: [master, main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [16, 18, 20]

    steps:
      - uses: actions/checkout@v4
      - name: Use Node.js ${{ matrix.node-version }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: "npm"

      - run: npm ci
      - run: npm run build --if-present
      - run: npm test

      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          token: ${{ secrets.CODECOV_TOKEN }}

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: production

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: "18"
          cache: "npm"

      - run: npm ci
      - run: npm run build

      - name: Deploy to AWS
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        run: npm run deploy
```

### 2. Docker 이미지 빌드/배포

```yaml
name: Docker Build and Deploy

on:
  push:
    branches: [master, main]
    tags: ["v*"]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
```

### 3. 멀티플랫폼 테스트

```yaml
name: Multi-platform Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        go-version: [1.23, 1.24]

    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}

      - name: Build
        run: go build -v ./...

      - name: Test
        run: go test -v ./...
```

## 고급 기능

### 1. 매트릭스 빌드

- 다중 버전/환경 테스트
- 병렬 실행 최적화
- 조건부 매트릭스 설정
- 실패 시 빠른 중단

### 2. 환경 보호 규칙

- 배포 승인 워크플로우
- 시간 지연 설정
- 브랜치 보호 규칙
- 환경별 시크릿 분리

### 3. 재사용 가능한 워크플로우

- 조직 차원 워크플로우 공유
- 매개변수화된 워크플로우
- 버전 관리
- 상속 및 오버라이드

### 4. 아티팩트 관리

- 빌드 결과물 보관
- 크로스 잡 아티팩트 공유
- 자동 정리
- 외부 저장소 연동

## 통합 기능

### 1. 패키지 레지스트리

- GitHub Packages 연동
- npm, Maven, Docker 지원
- 자동 버전 태깅
- 의존성 보안 스캔

### 2. 보안 스캔

- CodeQL 정적 분석
- Dependabot 취약점 스캔
- 시크릿 스캔
- 라이선스 검사

### 3. 이슈 및 PR 연동

- 자동 라벨링
- 브랜치 상태 체크
- 자동 머지
- 릴리스 노트 생성

## 권장 대안 도구

1. **GitHub Actions 직접 사용**: GitHub 네이티브 CI/CD
2. **GitLab CI/CD**: GitLab 통합 파이프라인
3. **Azure DevOps**: Microsoft CI/CD 플랫폼
4. **CircleCI**: 클라우드 CI/CD 서비스
5. **Jenkins**: 오픈소스 자동화 서버
6. **Travis CI**: GitHub 통합 CI 서비스
7. **Buildkite**: 하이브리드 CI/CD 플랫폼

## 복원 시 고려사항

- GitHub API 권한 및 토큰 관리
- 워크플로우 파일 버전 관리
- 시크릿 보안 및 순환 정책
- 러너 비용 최적화
- 병렬 실행 제한 관리
- 아티팩트 저장 용량 관리
- 로그 보존 정책
