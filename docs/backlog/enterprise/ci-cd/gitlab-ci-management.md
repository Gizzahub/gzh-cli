# GitLab CI/CD 관리 기능

## 개요

GitLab CI/CD 파이프라인 관리 및 자동화 기능

## 제거된 기능

### 1. GitLab CI 파이프라인 관리

- **명령어**: `gz gitlab-ci create`, `gz gitlab-ci deploy`
- **기능**: .gitlab-ci.yml 생성 및 파이프라인 관리
- **특징**:
  - 스테이지별 작업 정의
  - 조건부 파이프라인 실행
  - 병렬 및 의존성 작업
  - 수동 배포 승인

### 2. 러너 관리

- **명령어**: `gz gitlab-ci runner register`, `gz gitlab-ci runner scale`
- **기능**: GitLab Runner 등록 및 관리
- **특징**:
  - 공유/전용 러너 설정
  - Docker/Kubernetes 실행자
  - 자동 스케일링
  - 태그 기반 작업 분배

### 3. 변수 및 시크릿 관리

- **명령어**: `gz gitlab-ci variables set`, `gz gitlab-ci variables sync`
- **기능**: GitLab CI/CD 변수 및 시크릿 관리
- **특징**:
  - 환경별 변수 분리
  - 마스킹된 변수
  - 파일 타입 변수
  - 그룹/프로젝트 레벨 변수

### 4. 환경 및 배포 관리

- **명령어**: `gz gitlab-ci environment create`, `gz gitlab-ci deploy`
- **기능**: GitLab 환경 관리 및 배포 추적
- **특징**:
  - 환경별 배포 기록
  - 롤백 기능
  - 배포 승인 워크플로우
  - 환경 상태 모니터링

## 사용 예시 (제거 전)

```bash
# 새 파이프라인 생성
gz gitlab-ci create --type nodejs \
  --project company/myapp \
  --environments "dev,staging,prod"

# 러너 등록
gz gitlab-ci runner register --token $RUNNER_TOKEN \
  --executor docker --tags "linux,docker"

# 변수 설정
gz gitlab-ci variables set --project company/myapp \
  --key API_KEY --value secret --masked

# 환경 생성
gz gitlab-ci environment create --project company/myapp \
  --name production --url https://app.company.com
```

## 설정 파일 형식

```yaml
gitlab_ci:
  projects:
    - name: myapp-backend
      path: company/myapp-backend

      pipeline:
        stages:
          - build
          - test
          - security
          - deploy

        variables:
          DOCKER_DRIVER: overlay2
          POSTGRES_DB: testdb

        before_script:
          - apt-get update -qq
          - apt-get install -qq git

        jobs:
          build:
            stage: build
            image: node:18
            script:
              - npm ci
              - npm run build
            artifacts:
              paths:
                - dist/
              expire_in: 1 week
            cache:
              paths:
                - node_modules/

          test:
            stage: test
            image: node:18
            script:
              - npm ci
              - npm test
            coverage: '/Coverage: \d+\.\d+%/'
            artifacts:
              reports:
                junit: test-results.xml
                coverage_report:
                  coverage_format: cobertura
                  path: coverage/cobertura-coverage.xml

          security_scan:
            stage: security
            image: owasp/zap2docker-stable
            script:
              - zap-baseline.py -t $CI_ENVIRONMENT_URL
            artifacts:
              reports:
                sast: gl-sast-report.json

          deploy_staging:
            stage: deploy
            image: alpine:latest
            script:
              - apk add --no-cache curl
              - curl -X POST $STAGING_WEBHOOK
            environment:
              name: staging
              url: https://staging.company.com
            only:
              - develop

          deploy_production:
            stage: deploy
            image: alpine:latest
            script:
              - apk add --no-cache curl
              - curl -X POST $PRODUCTION_WEBHOOK
            environment:
              name: production
              url: https://app.company.com
            when: manual
            only:
              - main

  runners:
    shared:
      - name: docker-runner-01
        executor: docker
        tags: [linux, docker]
        concurrent: 4

    group_runners:
      - name: kubernetes-runner
        executor: kubernetes
        tags: [k8s, production]
        concurrent: 10

  variables:
    global:
      DOCKER_REGISTRY: registry.company.com

    project_level:
      - project: company/myapp-backend
        variables:
          DATABASE_URL: postgres://localhost/myapp
          API_ENDPOINT: https://api.company.com

    group_level:
      - group: company
        variables:
          AWS_DEFAULT_REGION: us-west-2

  environments:
    - name: development
      project: company/myapp-backend
      url: https://dev.company.com
      auto_stop_in: 1 week

    - name: staging
      project: company/myapp-backend
      url: https://staging.company.com
      deployment_tier: staging

    - name: production
      project: company/myapp-backend
      url: https://app.company.com
      deployment_tier: production
      protected: true

  schedules:
    - description: "Nightly tests"
      ref: main
      cron: "0 2 * * *"
      active: true

  notifications:
    email:
      on_failure: [devops@company.com]
    slack:
      webhook_url: https://hooks.slack.com/...
```

## 고급 기능

### 1. 멀티 프로젝트 파이프라인

- 프로젝트 간 의존성 관리
- 크로스 프로젝트 아티팩트 공유
- 통합 배포 워크플로우
- 마이크로서비스 조정

### 2. 동적 파이프라인

- 조건부 작업 실행
- 매트릭스 빌드
- include/extends 활용
- 런타임 파이프라인 생성

### 3. 보안 스캔 통합

- SAST/DAST 스캔
- 의존성 스캔
- 라이선스 스캔
- 컨테이너 스캔

### 4. 아티팩트 관리

- 패키지 레지스트리 연동
- 빌드 아티팩트 보관
- 테스트 보고서 수집
- 배포 패키지 관리

## 파이프라인 템플릿

### 1. Node.js 애플리케이션

```yaml
stages:
  - build
  - test
  - deploy

variables:
  NODE_VERSION: "18"

cache:
  paths:
    - node_modules/

build:
  stage: build
  image: node:$NODE_VERSION
  script:
    - npm ci
    - npm run build
  artifacts:
    paths:
      - dist/
    expire_in: 1 week

test:
  stage: test
  image: node:$NODE_VERSION
  script:
    - npm ci
    - npm test
  artifacts:
    reports:
      junit: test-results.xml
      coverage_report:
        coverage_format: cobertura
        path: coverage/cobertura-coverage.xml

deploy:
  stage: deploy
  image: alpine:latest
  script:
    - echo "Deploying to production"
    - ./deploy.sh
  environment:
    name: production
    url: https://app.example.com
  only:
    - main
  when: manual
```

### 2. Docker 이미지 빌드

```yaml
stages:
  - build
  - test
  - release

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"

services:
  - docker:dind

build:
  stage: build
  image: docker:latest
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

test:
  stage: test
  image: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  script:
    - ./run-tests.sh

release:
  stage: release
  image: docker:latest
  script:
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA $CI_REGISTRY_IMAGE:latest
    - docker push $CI_REGISTRY_IMAGE:latest
  only:
    - main
```

### 3. Kubernetes 배포

```yaml
stages:
  - build
  - test
  - deploy

deploy:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl config use-context $KUBE_CONTEXT
    - kubectl set image deployment/myapp myapp=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - kubectl rollout status deployment/myapp
  environment:
    name: production
    kubernetes:
      namespace: production
  only:
    - main
  when: manual
```

## 통합 기능

### 1. GitLab 페이지

- 정적 사이트 호스팅
- 문서 자동 배포
- 아티팩트 기반 배포
- 커스텀 도메인 지원

### 2. 모니터링 연동

- Prometheus 메트릭
- 애플리케이션 성능 모니터링
- 에러 추적
- 로그 수집

### 3. 이슈 추적 연동

- 자동 이슈 닫기
- MR과 이슈 연결
- 릴리스 노트 생성
- 변경 로그 추적

## 권장 대안 도구

1. **GitLab CI/CD 직접 사용**: GitLab 네이티브 파이프라인
2. **GitHub Actions**: GitHub 통합 CI/CD
3. **Jenkins**: 오픈소스 자동화 서버
4. **Azure DevOps**: Microsoft CI/CD 플랫폼
5. **CircleCI**: 클라우드 CI/CD 서비스
6. **Travis CI**: GitHub/GitLab 연동 CI
7. **Drone**: 클라우드 네이티브 CI/CD

## 복원 시 고려사항

- GitLab API 권한 및 토큰 관리
- 러너 설치 및 등록
- 변수 보안 및 암호화
- 환경별 배포 전략
- 아티팩트 저장소 설정
- 모니터링 및 알림 연동
- 백업 및 복구 계획
