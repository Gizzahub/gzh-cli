# Webhook Management Guide

이 가이드는 gzh-cli의 웹훅 관리 기능에 대한 종합적인 설명을 제공합니다.

## 목차

- [개요](#%EA%B0%9C%EC%9A%94)
- [기본 웹훅 관리](#%EA%B8%B0%EB%B3%B8-%EC%9B%B9%ED%9B%85-%EA%B4%80%EB%A6%AC)
- [대량 웹훅 작업](#%EB%8C%80%EB%9F%89-%EC%9B%B9%ED%9B%85-%EC%9E%91%EC%97%85)
- [이벤트 기반 자동화](#%EC%9D%B4%EB%B2%A4%ED%8A%B8-%EA%B8%B0%EB%B0%98-%EC%9E%90%EB%8F%99%ED%99%94)
- [설정 파일 참조](#%EC%84%A4%EC%A0%95-%ED%8C%8C%EC%9D%BC-%EC%B0%B8%EC%A1%B0)
- [고급 사용법](#%EA%B3%A0%EA%B8%89-%EC%82%AC%EC%9A%A9%EB%B2%95)
- [문제 해결](#%EB%AC%B8%EC%A0%9C-%ED%95%B4%EA%B2%B0)

## 개요

gzh-cli는 GitHub 리포지토리의 웹훅을 효율적으로 관리하는 포괄적인 도구 세트를 제공합니다:

- **개별 웹훅 관리**: 특정 리포지토리의 웹훅 CRUD 작업
- **대량 웹훅 작업**: 조직 전체 리포지토리에 웹훅 일괄 적용
- **이벤트 기반 자동화**: GitHub 이벤트에 따른 자동화 규칙 엔진

### 주요 특징

- ✅ **완전한 CRUD 지원**: 웹훅 생성, 조회, 수정, 삭제
- ✅ **병렬 처리**: 대량 작업 시 최대 50개 동시 처리
- ✅ **패턴 매칭**: 리포지토리 이름 패턴으로 대상 선택
- ✅ **실시간 자동화**: GitHub 이벤트 기반 액션 실행
- ✅ **설정 파일 지원**: YAML 기반 구성 관리
- ✅ **Dry-run 모드**: 실제 실행 전 미리보기

## 기본 웹훅 관리

### 웹훅 목록 조회

```bash
# 특정 리포지토리의 모든 웹훅 조회
gz repo-config webhook list --org myorg --repo myrepo

# JSON 형식으로 출력
gz repo-config webhook list --org myorg --repo myrepo --output json

# 테이블 형식으로 출력 (기본값)
gz repo-config webhook list --org myorg --repo myrepo --output table
```

### 웹훅 생성

```bash
# 기본 웹훅 생성
gz repo-config webhook create \
  --org myorg \
  --repo myrepo \
  --url https://example.com/webhook \
  --events push,pull_request \
  --secret mysecret

# 비활성 상태로 웹훅 생성
gz repo-config webhook create \
  --org myorg \
  --repo myrepo \
  --url https://example.com/webhook \
  --events push \
  --active=false

# 폼 인코딩 웹훅 생성
gz repo-config webhook create \
  --org myorg \
  --repo myrepo \
  --url https://example.com/webhook \
  --events push \
  --content-type form
```

### 웹훅 수정

```bash
# 웹훅 이벤트 변경
gz repo-config webhook update \
  --org myorg \
  --repo myrepo \
  --id 12345 \
  --events push,issues,pull_request

# 웹훅 URL 변경
gz repo-config webhook update \
  --org myorg \
  --repo myrepo \
  --id 12345 \
  --url https://new-endpoint.com/webhook

# 웹훅 비활성화
gz repo-config webhook update \
  --org myorg \
  --repo myrepo \
  --id 12345 \
  --active=false
```

### 웹훅 조회 및 삭제

```bash
# 특정 웹훅 조회
gz repo-config webhook get --org myorg --repo myrepo --id 12345

# 웹훅 삭제
gz repo-config webhook delete --org myorg --repo myrepo --id 12345
```

## 대량 웹훅 작업

### 설정 파일 생성

먼저 대량 웹훅 설정을 위한 YAML 파일을 생성합니다:

```yaml
# webhook-bulk-config.yaml
version: "1.0"

# 정의할 웹훅들
webhooks:
  # CI/CD 웹훅
  - url: https://ci.example.com/github/webhook
    events:
      - push
      - pull_request
    active: true
    content_type: json
    secret: ${WEBHOOK_SECRET}

  # 이슈 추적 웹훅
  - url: https://tracker.example.com/github/webhook
    events:
      - issues
      - issue_comment
      - pull_request_review
    active: true
    content_type: json

# 대상 리포지토리 지정
targets:
  all: true # 모든 리포지토리에 적용
  exclude:
    - test-repo
    - archived-repo

# 작업 옵션
options:
  skip_existing: false
  max_workers: 5
  continue_on_error: true
```

### 대량 웹훅 명령어

```bash
# 모든 리포지토리에 웹훅 생성
gz repo-config webhook bulk create \
  --org myorg \
  --config webhook-bulk-config.yaml

# 특정 패턴 리포지토리에만 적용
gz repo-config webhook bulk create \
  --org myorg \
  --config webhook-bulk-config.yaml \
  --pattern "^(api-|service-)"

# Dry-run으로 미리보기
gz repo-config webhook bulk create \
  --org myorg \
  --config webhook-bulk-config.yaml \
  --dry-run

# 기존 웹훅과 동기화
gz repo-config webhook bulk sync \
  --org myorg \
  --config webhook-bulk-config.yaml

# 대량 웹훅 조회
gz repo-config webhook bulk list \
  --org myorg \
  --all

# 대량 웹훅 삭제
gz repo-config webhook bulk delete \
  --org myorg \
  --url https://old-endpoint.com/webhook \
  --confirm
```

### 타겟 지정 옵션

```yaml
targets:
  # 모든 리포지토리
  all: true

  # 특정 리포지토리들
  repositories:
    - my-app
    - my-api
    - my-lib

  # 패턴 매칭
  pattern: "^(api-|service-)"

  # 제외할 리포지토리들
  exclude:
    - test-repo
    - archived-repo
    - legacy-app
```

## 이벤트 기반 자동화

### 자동화 엔진 설정

자동화 규칙을 정의하는 YAML 파일을 생성합니다:

```yaml
# webhook-automation-rules.yaml
version: "1.0"

global:
  enabled: true
  default_timeout: "30s"
  max_concurrency: 10
  notification_urls:
    slack: "${SLACK_WEBHOOK_URL}"

rules:
  # PR 크기별 자동 라벨링
  - id: "auto-label-pr-size"
    name: "Auto-label Pull Request Size"
    enabled: true
    priority: 100
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "pull_request.opened"
    actions:
      - type: "add_label"
        parameters:
          labels:
            - "needs-review"

  # 첫 기여자 환영 메시지
  - id: "welcome-first-time-contributor"
    name: "Welcome First-Time Contributors"
    enabled: true
    priority: 90
    conditions:
      - type: "event_type"
        operator: "in"
        value: ["pull_request.opened", "issues.opened"]
    actions:
      - type: "create_comment"
        parameters:
          body: |
            Welcome @{{sender.login}}! 👋
            Thank you for your contribution!
```

### 자동화 엔진 실행

```bash
# 웹훅 서버 시작 (포트 8080)
gz repo-config webhook automation server \
  --config webhook-automation-rules.yaml \
  --port 8080

# 특정 포트에서 서버 시작
gz repo-config webhook automation server \
  --config webhook-automation-rules.yaml \
  --port 9000 \
  --host 0.0.0.0

# 백그라운드에서 실행
gz repo-config webhook automation server \
  --config webhook-automation-rules.yaml \
  --daemon

# 설정 검증
gz repo-config webhook automation validate \
  --config webhook-automation-rules.yaml

# 테스트 이벤트 실행
gz repo-config webhook automation test \
  --config webhook-automation-rules.yaml \
  --event-type pull_request.opened

# 예제 설정 생성
gz repo-config webhook automation example > automation-rules.yaml
```

### 지원되는 액션 타입

#### 1. 라벨 관리

```yaml
- type: "add_label"
  parameters:
    labels:
      - "bug"
      - "enhancement"
```

#### 2. 이슈 생성

```yaml
- type: "create_issue"
  parameters:
    title: "New issue: {{payload.title}}"
    body: "Description: {{payload.description}}"
    labels:
      - "auto-created"
    assignees:
      - "maintainer"
```

#### 3. 댓글 생성

```yaml
- type: "create_comment"
  parameters:
    body: |
      Thank you for your contribution!
      A maintainer will review this soon.
```

#### 4. PR 머지

```yaml
- type: "merge_pr"
  parameters:
    merge_method: "squash" # merge, squash, rebase
    commit_title: "Auto-merge: {{payload.title}}"
```

#### 5. 알림 전송

```yaml
- type: "notification"
  parameters:
    type: "slack" # slack, discord, teams
    message: "New PR opened: {{payload.title}}"
    async: true
```

#### 6. 워크플로우 실행

```yaml
- type: "run_workflow"
  parameters:
    workflow_id: "ci.yml"
    ref: "main"
    inputs:
      environment: "production"
```

### 조건 표현식

#### 이벤트 타입

```yaml
conditions:
  - type: "event_type"
    operator: "equals"
    value: "pull_request.opened"

  - type: "event_type"
    operator: "in"
    value: ["push", "pull_request"]

  - type: "event_type"
    operator: "matches"
    value: "workflow_run.*"
```

#### 페이로드 조건

```yaml
conditions:
  - type: "payload"
    field: "pull_request.base.ref"
    operator: "equals"
    value: "main"

  - type: "sender"
    field: "login"
    operator: "equals"
    value: "dependabot[bot]"
```

#### 복합 조건

```yaml
conditions:
  - type: "event_type"
    operator: "equals"
    value: "pull_request.opened"
  - type: "payload"
    field: "pull_request.draft"
    operator: "equals"
    value: false
```

## 설정 파일 참조

### 대량 웹훅 설정 스키마

```yaml
version: "1.0" # 필수

webhooks: # 필수
  - url: string # 필수
    events: [string] # 필수
    active: boolean # 선택 (기본값: true)
    content_type: string # 선택 (기본값: json)
    secret: string # 선택

targets: # 필수
  all: boolean # 선택
  repositories: [string] # 선택
  pattern: string # 선택
  exclude: [string] # 선택

options: # 선택
  skip_existing: boolean # 기본값: false
  max_workers: integer # 기본값: 5
  continue_on_error: boolean # 기본값: true
```

### 자동화 규칙 설정 스키마

```yaml
version: "1.0" # 필수

global: # 선택
  enabled: boolean # 기본값: true
  default_timeout: string # 기본값: "30s"
  max_concurrency: integer # 기본값: 10
  notification_urls:
    slack: string
    discord: string
    teams: string

rules: # 필수
  - id: string # 필수
    name: string # 필수
    description: string # 선택
    enabled: boolean # 기본값: true
    priority: integer # 기본값: 100
    conditions: [object] # 필수
    actions: [object] # 필수
```

## 고급 사용법

### 환경 변수 사용

설정 파일에서 환경 변수를 사용할 수 있습니다:

```yaml
webhooks:
  - url: ${WEBHOOK_URL}
    secret: ${WEBHOOK_SECRET}
    events:
      - push
```

```bash
export WEBHOOK_URL="https://example.com/webhook"
export WEBHOOK_SECRET="mysecret"
gz repo-config webhook bulk create --config config.yaml
```

### 템플릿 변수

자동화 규칙에서 GitHub 이벤트 데이터를 템플릿으로 사용:

```yaml
actions:
  - type: "create_comment"
    parameters:
      body: |
        Hello @{{sender.login}}!

        Repository: {{repo.name}}
        Event: {{event_type}}
        PR Number: {{payload.pull_request.number}}
        PR Title: {{payload.pull_request.title}}
```

### 조건부 실행

복잡한 조건을 통한 세밀한 제어:

```yaml
rules:
  - id: "security-only-main"
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "push"
      - type: "payload"
        field: "ref"
        operator: "equals"
        value: "refs/heads/main"
      - type: "payload"
        field: "repository.private"
        operator: "equals"
        value: true
    actions:
      - type: "run_workflow"
        parameters:
          workflow_id: "security-scan.yml"
```

### 병렬 처리 최적화

```bash
# 높은 병렬성으로 빠른 처리
gz repo-config webhook bulk create \
  --config config.yaml \
  --max-workers 20

# 안전한 순차 처리
gz repo-config webhook bulk create \
  --config config.yaml \
  --max-workers 1
```

## 문제 해결

### 일반적인 오류

#### 1. 인증 오류

```bash
Error: authentication failed
```

**해결방법**: GitHub 토큰이 올바른지 확인하고 적절한 권한이 있는지 확인

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

#### 2. 권한 부족

```bash
Error: insufficient permissions
```

**해결방법**: 토큰에 `admin:repo_hook` 스코프가 있는지 확인

#### 3. 웹훅 생성 실패

```bash
Error: webhook creation failed: validation failed
```

**해결방법**:

- URL이 유효한 HTTPS 엔드포인트인지 확인
- 이벤트 타입이 올바른지 확인
- secret가 너무 길지 않은지 확인

#### 4. 설정 파일 오류

```bash
Error: invalid configuration: missing required field
```

**해결방법**: 설정 파일 검증 실행

```bash
gz repo-config webhook automation validate --config config.yaml
```

### 디버깅 옵션

```bash
# 상세 로그 출력
gz repo-config webhook bulk create \
  --config config.yaml \
  --verbose

# 디버그 모드
gz repo-config webhook automation server \
  --config config.yaml \
  --debug

# Dry-run으로 테스트
gz repo-config webhook bulk create \
  --config config.yaml \
  --dry-run
```

### 성능 튜닝

#### 대량 작업 최적화

```yaml
options:
  max_workers: 10 # API 제한에 맞게 조정
  continue_on_error: true # 일부 실패해도 계속 진행
  skip_existing: true # 중복 생성 방지
```

#### 자동화 엔진 최적화

```yaml
global:
  max_concurrency: 20 # 동시 처리할 이벤트 수
  default_timeout: "60s" # 액션 타임아웃
```

### 모니터링

#### 웹훅 상태 확인

```bash
# 조직의 모든 웹훅 상태 조회
gz repo-config webhook bulk list --org myorg --all --output json

# 특정 URL을 가진 웹훅 검색
gz repo-config webhook bulk list --org myorg --url "example.com"
```

#### 자동화 로그 확인

자동화 서버는 표준 출력으로 처리 로그를 출력합니다:

```bash
gz repo-config webhook automation server --config config.yaml 2>&1 | tee automation.log
```

## 예제 시나리오

### 시나리오 1: CI/CD 웹훅 일괄 설정

모든 리포지토리에 CI/CD 웹훅을 설정:

```yaml
# ci-webhooks.yaml
version: "1.0"
webhooks:
  - url: https://ci.company.com/github/webhook
    events: [push, pull_request]
    active: true
    secret: ${CI_WEBHOOK_SECRET}

targets:
  all: true
  exclude:
    - archived-*
    - test-*
```

```bash
gz repo-config webhook bulk create --org myorg --config ci-webhooks.yaml
```

### 시나리오 2: 자동 PR 리뷰 시스템

PR이 열릴 때 자동으로 리뷰어 할당:

```yaml
# auto-review.yaml
rules:
  - id: "assign-reviewers"
    name: "Auto-assign Reviewers"
    conditions:
      - type: "event_type"
        operator: "equals"
        value: "pull_request.opened"
    actions:
      - type: "add_label"
        parameters:
          labels: ["needs-review"]
      - type: "create_comment"
        parameters:
          body: "🔍 This PR has been automatically assigned for review."
```

```bash
gz repo-config webhook automation server --config auto-review.yaml --port 8080
```

### 시나리오 3: 보안 이벤트 알림

보안 취약점 발견 시 즉시 알림:

```yaml
# security-alerts.yaml
rules:
  - id: "security-alert"
    name: "Security Vulnerability Alert"
    priority: 100
    conditions:
      - type: "event_type"
        operator: "matches"
        value: "security_advisory.*"
    actions:
      - type: "create_issue"
        parameters:
          title: "🔒 URGENT: Security Alert - {{payload.security_advisory.summary}}"
          labels: ["security", "urgent"]
      - type: "notification"
        parameters:
          type: "slack"
          message: "🚨 Security vulnerability detected!"
```

## 결론

gzh-cli의 웹훅 관리 기능은 GitHub 리포지토리의 자동화를 위한 강력하고 유연한 도구입니다. 개별 웹훅 관리부터 조직 전체의 대량 작업, 그리고 이벤트 기반 자동화까지 포괄적인 기능을 제공하여 개발 워크플로우를 크게 개선할 수 있습니다.

더 자세한 정보는 다음 문서를 참조하세요:

- [API 참조](webhook-api-reference.md)
- [고급 설정](webhook-advanced-configuration.md)
- [문제 해결 가이드](webhook-troubleshooting.md)
