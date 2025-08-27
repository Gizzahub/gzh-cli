# Gizzahub Manager

**Comprehensive CLI Tool**

![Test Status](https://github.com/gizzahub/gzh-cli/actions/workflows/test.yml/badge.svg)
![Lint Status](https://github.com/gizzahub/gzh-cli/actions/workflows/lint.yml/badge.svg)
![GoDoc](https://pkg.go.dev/badge/github.com/gizzahub/gzh-cli.svg)
![Code Coverage](https://codecov.io/gh/Gizzahub/gzh-cli/branch/main/graph/badge.svg)
![Latest Release](https://img.shields.io/github/v/release/Gizzahub/gzh-cli)
![Docker Pulls](https://img.shields.io/docker/pulls/Gizzahub/gzh-cli)
![Total Downloads](https://img.shields.io/github/downloads/Gizzahub/gzh-cli/total.svg)

## Table of Contents

<!--ts-->

- [Usage](#usage)
- [Features](#features)
- [Installation](#installation)
- [Command Reference](#command-reference)
- [Configuration](#configuration)
- [Performance Monitoring](#performance-monitoring)
- [Development](#development)
- [Contributing](#contributing)

<!--te-->

## Usage

## 핵심 기능 개요

`gzh-cli` (바이너리명: `gz`)는 개발자를 위한 종합적인 CLI 도구로, 다음과 같은 주요 기능을 제공합니다:

### 🏗️ 개발 환경 통합 관리

- **Git 플랫폼 통합**: GitHub, GitLab, Gitea, Gogs를 하나의 인터페이스로 관리
- **향상된 IDE 관리**: JetBrains/VS Code 통합 스캔, 상태 모니터링, 프로젝트 열기 지원
- **코드 품질 관리**: 다중 언어 포매팅/린팅 도구의 통합 실행 및 관리 (테스트 커버리지 34.4%↑)
- **성능 프로파일링**: Go pprof 기반의 간편한 성능 분석 도구 (테스트 커버리지 36.6%↑)
- **개발 환경 설정**: AWS, Docker, Kubernetes, SSH 설정 관리
- **네트워크 환경 전환**: WiFi, VPN, DNS, 프록시 설정 자동 전환

### 📦 리포지토리 관리

- **대량 클론 도구**: GitHub, GitLab, Gitea, Gogs에서 전체 조직의 리포지토리를 일괄 클론
- **크로스 플랫폼 동기화**: 서로 다른 Git 플랫폼 간 리포지토리 동기화 (코드, 이슈, 위키, 릴리스)
- **고급 클론 전략**: reset, pull, fetch, rebase 모드 지원으로 기존 리포지토리 동기화 방식 제어
- **재개 가능한 작업**: 중단된 클론 작업을 이어서 진행할 수 있는 상태 관리 시스템
- **병렬 처리**: 최대 50개의 동시 클론 작업으로 대규모 조직 처리 성능 향상
- **스마트 URL 파싱**: HTTPS, SSH, git:// 등 다양한 Git URL 형식 지원

### 🏢 GitHub 조직 관리

- **리포지토리 설정 관리**: 조직 전체 리포지토리의 설정을 템플릿 기반으로 일괄 관리
- **웹훅 관리**: GitHub 웹훅의 생성, 수정, 삭제 및 모니터링
- **이벤트 처리**: GitHub 이벤트 수신 및 자동화된 응답 처리
- **보안 정책 적용**: 조직 차원의 보안 정책 일괄 적용 및 감사

### 🛠️ 개발 도구 통합

- **패키지 매니저 업데이트**: asdf, Homebrew, SDKMAN, npm, pip 등 다양한 패키지 매니저 통합 관리
- **IDE 설정 동기화**: JetBrains 제품군의 설정 충돌 감지 및 자동 복구
- **코드 품질 자동화**: Go, Python, JavaScript, Rust 등 다중 언어 품질 도구 통합
- **성능 모니터링**: 애플리케이션 성능 프로파일링 및 벤치마킹

## 빠른 시작

### 설치

```bash
# Go를 통한 설치 (권장)
go install github.com/Gizzahub/gzh-cli/cmd/gz@latest

# 또는 소스에서 빌드
git clone https://github.com/Gizzahub/gzh-cli.git
cd gzh-cli
make build
make install
```

### 기본 사용법

```bash
# 시스템 상태 진단 (숨겨진 명령어)
gz doctor

# 설정 파일 검증
gz synclone validate --config examples/synclone.yaml

# GitHub 조직의 저장소 클론
gz synclone github --orgName myorg --targetPath ~/repos/myorg --token $GITHUB_TOKEN

# IDE 시스템 스캔 및 상태 확인
gz ide scan          # 설치된 IDE 스캔
gz ide status        # IDE 상태 확인
gz ide open project-path  # IDE로 프로젝트 열기

# JetBrains IDE 설정 모니터링
gz ide monitor

# 코드 품질 검사 및 포매팅
gz quality run

# 성능 프로파일링
gz profile stats
gz profile cpu --duration 30s

# 리포지토리 설정 감사
gz repo-config audit --org myorg --framework SOC2
```

## CLI 명령어 구조

```bash
$ gz --help
gz는 개발자를 위한 종합 CLI 도구입니다.

개발 환경 설정, Git 플랫폼 관리, IDE 모니터링, 네트워크 환경 전환 등
다양한 개발 워크플로우를 통합적으로 관리할 수 있습니다.

Utility Commands: doctor, version

Usage:
  gz [flags]
  gz [command]

Available Commands:
  dev-env     Manage development environment configurations
  git         🔗 통합 Git 플랫폼 관리 도구 (repo, webhook, event)
  ide         Monitor and manage IDE configuration changes
  net-env     Manage network environment transitions
  pm          Manage development tools and package managers
  profile     Performance profiling using standard Go pprof
  quality     통합 코드 품질 도구 (포매팅 + 린팅)
  repo-config GitHub repository configuration management
  synclone    Synchronize and clone repositories from multiple Git hosting services

Flags:
      --debug     Enable debug logging (shows all log levels)
  -h, --help      help for gz
  -q, --quiet     Suppress all logs except critical errors
  -v, --verbose   Enable verbose logging

Use "gz [command] --help" for more information about a command.
```

Each command module under `cmd/<module>` includes an `AGENTS.md` file with
module-specific coding conventions, required tests, and review steps. Always
consult these guidelines before modifying command implementations.

## Features

## 🔗 Git 플랫폼 통합 관리 (`gz git`)

통합된 Git 명령어 인터페이스로 다양한 Git 호스팅 플랫폼을 하나의 명령어로 관리합니다.

### 주요 기능

- **리포지토리 라이프사이클**: 생성, 삭제, 아카이브, 클론 및 업데이트
- **크로스 플랫폼 동기화**: GitHub ↔ GitLab ↔ Gitea 간 리포지토리 동기화
- **웹훅 관리**: GitHub, GitLab 웹훅 통합 관리
- **이벤트 처리**: Git 플랫폼 이벤트 수신 및 처리
- **설정 관리**: 다중 플랫폼 설정 통합

```bash
# 리포지토리 스마트 클론/업데이트
gz git repo clone-or-update https://github.com/user/repo.git
gz git repo clone-or-update https://github.com/user/repo.git --strategy rebase

# 리포지토리 생성/삭제
gz git repo create --name myrepo --org myorg --private
gz git repo delete --name myrepo --org myorg --confirm

# 크로스 플랫폼 동기화
gz git repo sync --from github:org/repo --to gitlab:group/repo
gz git repo sync --from github:org --to gitea:org --create-missing

# 웹훅 관리
gz git webhook list --org myorg
gz git webhook create --org myorg --repo myrepo --url https://api.example.com/webhook

# 이벤트 서버
gz git event server --port 8080
```

## 🖥️ IDE 모니터링 및 관리 (`gz ide`)

다양한 IDE의 설정을 관리하고 프로젝트를 열 수 있는 통합 IDE 관리 도구입니다.

### 지원하는 IDE

**JetBrains 제품군:**

- IntelliJ IDEA (Community, Ultimate)
- PyCharm (Community, Professional)
- WebStorm, PhpStorm, RubyMine
- CLion, GoLand, DataGrip
- Android Studio, Rider

**VS Code 계열:**

- Visual Studio Code
- VS Code Insiders
- Cursor
- VSCodium

**기타 에디터:**

- Sublime Text, Vim, Neovim, Emacs

### 주요 기능

- **IDE 스캔**: 시스템에 설치된 모든 IDE 자동 감지 (캐시 지원)
- **상태 모니터링**: IDE 프로세스, 메모리, 프로젝트 상태 실시간 확인
- **프로젝트 열기**: 감지된 IDE로 프로젝트 직접 열기
- **실시간 모니터링**: 설정 파일 변경 감지
- **동기화 수정**: 설정 충돌 자동 해결
- **크로스플랫폼 지원**: Linux, macOS, Windows
- **백업 및 복구**: 설정 변경 전 자동 백업

```bash
# IDE 스캔 (24시간 캐시)
gz ide scan
gz ide scan --refresh  # 캐시 무시하고 새로 스캔
gz ide scan --verbose  # 상세 정보 표시

# IDE 상태 확인
gz ide status          # 모든 IDE 상태
gz ide status --running  # 실행 중인 IDE만
gz ide status --format json  # JSON 출력

# IDE로 프로젝트 열기
gz ide open /path/to/project
gz ide open . --ide goland  # 특정 IDE로 열기

# JetBrains IDE 모니터링
gz ide monitor
gz ide monitor --product IntelliJIdea2023.2

# 동기화 문제 수정
gz ide fix-sync

# 설치된 IDE 목록 (레거시)
gz ide list
```

## 🔧 코드 품질 관리 (`gz quality`)

다중 언어를 지원하는 통합 코드 품질 관리 도구입니다.

### 지원 언어 및 도구

- **Go**: gofumpt, golangci-lint, goimports, gci
- **Python**: ruff (format + lint), black, isort, flake8, mypy
- **JavaScript/TypeScript**: prettier, eslint, dprint
- **Rust**: rustfmt, clippy
- **Java**: google-java-format, checkstyle, spotbugs
- **C/C++**: clang-format, clang-tidy
- **기타**: YAML, JSON, Markdown, Shell 스크립트 지원

### 주요 기능

- **통합 실행**: 모든 품질 도구를 하나의 명령어로 실행
- **선택적 처리**: 변경된 파일 또는 스테이징된 파일만 처리
- **도구 관리**: 품질 도구 설치, 업그레이드, 버전 관리
- **프로젝트 분석**: 프로젝트에 적합한 도구 자동 추천
- **CI/CD 통합**: JSON, JUnit XML 출력 형식 지원

```bash
# 모든 품질 도구 실행
gz quality run

# 변경된 파일만 처리
gz quality run --changed

# 린팅만 실행 (변경 없이 검사)
gz quality check

# 프로젝트 분석 및 도구 추천
gz quality analyze

# 품질 도구 설치
gz quality install

# 특정 도구 직접 실행
gz quality tool prettier --staged
```

## 📊 성능 프로파일링 (`gz profile`)

Go의 표준 pprof를 기반으로 한 간편한 성능 분석 도구입니다.

### 주요 기능

- **HTTP 서버**: pprof 웹 인터페이스 제공
- **CPU 프로파일링**: 지정된 시간 동안 CPU 사용량 분석
- **메모리 프로파일링**: 힙 메모리 사용량 분석
- **런타임 통계**: 실시간 메모리 및 GC 통계

```bash
# 런타임 통계 확인
gz profile stats

# pprof HTTP 서버 시작
gz profile server --port 6060

# CPU 프로파일링 (30초)
gz profile cpu --duration 30s

# 메모리 프로파일링
gz profile memory
```

## 🌐 네트워크 환경 관리 (`gz net-env`)

네트워크 환경 변화를 감지하고 자동으로 설정을 전환하는 도구입니다.

### 주요 기능

- **WiFi 변화 감지**: 네트워크 변경 자동 감지
- **프록시 설정**: 환경별 프록시 자동 전환
- **DNS 관리**: 환경별 DNS 서버 설정
- **VPN 통합**: VPN 연결 상태 관리

## 🔄 패키지 매니저 통합 (`gz pm`)

다양한 패키지 매니저를 통합 관리하는 도구입니다.

### 지원하는 패키지 매니저

- **언어별**: asdf, nvm, pyenv, rbenv
- **시스템**: Homebrew (macOS), apt (Ubuntu), yum (CentOS)
- **개발도구**: npm, pip, cargo, go modules
- **클라우드**: SDKMAN, kubectl, helm

### 주요 기능

- **일괄 업데이트**: 모든 패키지 매니저 동시 업데이트
- **선택적 업데이트**: 특정 도구만 업데이트
- **상태 확인**: 설치된 도구 및 버전 확인
- **의존성 관리**: 의존성 충돌 감지 및 해결

## 📦 대량 리포지토리 클론 (`gz synclone`)

다중 Git 플랫폼에서 대량의 리포지토리를 효율적으로 관리하는 도구입니다.

### 지원하는 플랫폼

- **GitHub**: 조직, 개인 리포지토리
- **GitLab**: 그룹, 프로젝트
- **Gitea**: 조직, 개인 리포지토리
- **Gogs**: 조직, 개인 리포지토리 (계획 중)

### 주요 기능

- **병렬 클론**: 최대 50개 동시 작업
- **재개 기능**: 중단된 작업 이어서 진행
- **다양한 전략**: reset, pull, fetch, rebase
- **상태 관리**: 클론 진행 상황 추적 및 저장

## Installation

## 시스템 요구사항

- **Go**: 1.22 이상
- **Git**: 2.0 이상
- **OS**: Linux, macOS, Windows

## 설치 방법

### 1. Go Install (권장)

```bash
go install github.com/Gizzahub/gzh-cli/cmd/gz@latest
```

### 2. 소스에서 빌드

```bash
git clone https://github.com/Gizzahub/gzh-cli.git
cd gzh-cli
make bootstrap  # 빌드 의존성 설치
make build      # gz 실행 파일 생성
make install    # $GOPATH/bin에 설치
```

### 3. 사전 컴파일된 바이너리

[Releases](https://github.com/Gizzahub/gzh-cli/releases) 페이지에서 플랫폼별 바이너리를 다운로드하세요.

## 설치 확인

```bash
gz --version
gz doctor  # 시스템 상태 진단 (숨겨진 명령어)
```

## Command Reference

## 전역 플래그

모든 명령어에서 사용할 수 있는 공통 플래그입니다:

```bash
--verbose, -v    # 상세 로그 출력
--debug          # 디버그 로그 출력 (모든 로그 레벨)
--quiet, -q      # 오류 외 모든 로그 숨김
--help, -h       # 도움말 표시
```

## 주요 명령어별 세부 사용법

### `gz synclone` - 리포지토리 대량 클론

```bash
# GitHub 조직 전체 클론
gz synclone github --orgName myorg --targetPath ~/repos --token $GITHUB_TOKEN

# GitLab 그룹 클론
gz synclone gitlab --groupName mygroup --targetPath ~/repos --token $GITLAB_TOKEN

# 설정 파일로 실행
gz synclone --config examples/synclone.yaml

# 작업 재개
gz synclone --resume

# 설정 검증
gz synclone validate --config synclone.yaml
```

### `gz git` - Git 플랫폼 통합

```bash
# 리포지토리 클론 또는 업데이트
gz git repo clone-or-update https://github.com/user/repo.git
gz git repo clone-or-update https://github.com/user/repo.git --branch develop --strategy rebase

# 리포지토리 생성/삭제
gz git repo create --name myrepo --org myorg --private
gz git repo delete --name myrepo --org myorg --confirm

# 크로스 플랫폼 동기화
gz git repo sync --from github:org/repo --to gitlab:group/repo
gz git repo sync --from github:org --to gitea:org --create-missing

# 웹훅 관리
gz git webhook list --org myorg
gz git webhook create --org myorg --repo myrepo --url https://example.com/hook

# 이벤트 서버 시작
gz git event server --port 8080
```

### `gz quality` - 코드 품질 관리

```bash
# 전체 품질 검사 및 수정
gz quality run

# 린팅만 (수정 없이 검사)
gz quality check --severity error

# 변경된 파일만 처리
gz quality run --changed

# 프로젝트 초기 설정
gz quality init

# 도구 관리
gz quality install gofumpt
gz quality upgrade
gz quality version
```

### `gz ide` - IDE 관리

```bash
# IDE 스캔 및 감지
gz ide scan                  # 설치된 IDE 스캔 (24시간 캐시)
gz ide scan --refresh        # 캐시 무시하고 새로 스캔

# IDE 상태 확인
gz ide status                # 모든 IDE 상태
gz ide status --running      # 실행 중인 IDE만

# IDE로 프로젝트 열기
gz ide open /path/to/project
gz ide open . --ide goland   # 특정 IDE로 열기

# 실시간 모니터링 (JetBrains)
gz ide monitor
gz ide monitor --product IntelliJIdea2023.2

# 동기화 문제 수정
gz ide fix-sync --dry-run    # 미리보기
gz ide fix-sync

# IDE 목록 확인
gz ide list --format json
```

### `gz profile` - 성능 프로파일링

```bash
# 기본 통계
gz profile stats

# HTTP 서버 시작
gz profile server --port 6060

# CPU 프로파일링
gz profile cpu --duration 60s

# 메모리 프로파일링
gz profile memory
```

### `gz dev-env` - 개발 환경 관리

```bash
# AWS 설정 관리
gz dev-env aws configure
gz dev-env aws status

# Docker 환경 설정
gz dev-env docker setup
gz dev-env docker status

# Kubernetes 설정
gz dev-env k8s configure
gz dev-env k8s status
```

### `gz pm` - 패키지 매니저 관리

```bash
# 전체 업데이트
gz pm update

# 특정 매니저 업데이트
gz pm update --manager homebrew

# 상태 확인
gz pm status
gz pm list
```

## Configuration

## 설정 파일 계층 구조

설정 파일은 다음 순서로 우선순위를 가집니다:

1. 환경 변수: `GZH_CONFIG_PATH`
1. 현재 디렉토리: `./synclone.yaml` 또는 `./synclone.yml`
1. 사용자 설정: `~/.config/gzh-manager/synclone.yaml`
1. 시스템 설정: `/etc/gzh-manager/synclone.yaml`

## 주요 설정 파일

### synclone.yaml - 리포지토리 클론 설정

```yaml
# 기본 설정
parallel_limit: 10
timeout: 300
resume_enabled: true

# GitHub 설정
github:
  token: "${GITHUB_TOKEN}"
  organizations:
    - name: "myorg"
      target_path: "~/repos/myorg"
      strategy: "reset"

# GitLab 설정
gitlab:
  token: "${GITLAB_TOKEN}"
  groups:
    - name: "mygroup"
      target_path: "~/repos/gitlab"
      strategy: "pull"
```

### quality.yaml - 코드 품질 설정

```yaml
quality:
  tools:
    enabled: ["gofumpt", "golangci-lint", "prettier", "eslint"]
    disabled: []

  execution:
    parallel: true
    timeout: 300
    fail_fast: false

  filters:
    exclude_patterns:
      - "vendor/"
      - "node_modules/"
      - "*.generated.go"
```

### ide.yaml - IDE 설정

```yaml
ide:
  monitoring:
    enabled: true
    interval: 1s
    filter_temp_files: true

  products:
    - name: "IntelliJIdea"
      enabled: true
      custom_path: "/custom/path/to/config"

  sync:
    backup_enabled: true
    backup_retention: 7  # days
```

## 환경 변수

```bash
# 인증 토큰
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"
export GITEA_TOKEN="xxxxxxxxxxxx"

# 설정 경로
export GZH_CONFIG_PATH="/path/to/config.yaml"

# 디버그 모드
export GZH_DEBUG_SHELL=1  # 디버그 셸 활성화

# IDE 관련
export JETBRAINS_CONFIG_PATH="/custom/jetbrains/config"
export IDE_MONITOR_INTERVAL="1s"

# 품질 도구 관련
export QUALITY_PARALLEL=true
export QUALITY_TIMEOUT=300
```

## Performance Monitoring

## 성능 벤치마킹

프로젝트에는 자동화된 성능 모니터링 시스템이 포함되어 있습니다:

### 빠른 성능 체크

```bash
# 기본 성능 체크 (startup time, binary size, memory)
./scripts/simple-benchmark.sh
```

### 상세 성능 분석

```bash
# 베이스라인 생성
./scripts/benchmark-performance.sh --baseline > baseline.json

# 베이스라인과 비교
./scripts/benchmark-performance.sh --compare baseline.json

# 사람이 읽기 쉬운 형태로 출력
./scripts/benchmark-performance.sh --format human
```

### 성능 메트릭

- **시작 시간**: 50ms 이하 목표
- **바이너리 크기**: ~33MB
- **메모리 사용량**: 최소한으로 유지
- **명령어 응답 시간**: 대부분 100ms 이하

### 성능 프로파일링

```bash
# 런타임 통계 확인
gz profile stats

# CPU 프로파일링 (30초간)
gz profile cpu --duration 30s

# 메모리 프로파일링
gz profile memory

# pprof 웹 인터페이스 시작
gz profile server --port 6060
# http://localhost:6060/debug/pprof/ 접속
```

## Development

## 개발 환경 설정

### 필수 도구 설치

```bash
# 빌드 의존성 설치 (한 번만 실행)
make bootstrap

# 개발 도구 확인
make check-tools
```

### 빌드 및 테스트

```bash
# 빌드
make build

# 테스트
make test
make test-coverage

# 코드 품질 검사 (커밋 전 필수)
make fmt        # 코드 포매팅
make lint       # 린팅 검사
make lint-all   # 전체 품질 검사

# 특정 패키지 테스트
go test ./cmd/ide -v
go test ./cmd/quality -v
go test ./pkg/github -v
```

### Pre-commit 훅 설정

```bash
# pre-commit 훅 설치 (한 번만 실행)
make pre-commit-install

# 수동으로 pre-commit 실행
make pre-commit

# pre-push 훅 실행
make pre-push
```

### 코드 생성

```bash
# Mock 파일 생성
make generate-mocks

# Mock 파일 정리 및 재생성
make clean-mocks
make regenerate-mocks
```

## 아키텍처 개요

### 프로젝트 구조

```
.
├── cmd/                    # CLI 명령어 구현
│   ├── root.go            # 메인 CLI 진입점
│   ├── git/               # Git 통합 명령어
│   ├── ide/               # IDE 모니터링
│   ├── quality/           # 코드 품질 도구
│   ├── profile/           # 성능 프로파일링
│   ├── synclone/          # 대량 리포지토리 클론
│   ├── dev-env/           # 개발 환경 관리
│   ├── net-env/           # 네트워크 환경 관리
│   ├── pm/                # 패키지 매니저 관리
│   └── repo-config/       # 리포지토리 설정 관리
├── internal/              # 내부 패키지
│   ├── git/               # Git 조작 추상화
│   ├── logger/            # 로깅 추상화
│   ├── simpleprof/        # 간단한 프로파일링
│   └── testlib/           # 테스트 유틸리티
├── pkg/                   # 공개 패키지
│   ├── github/            # GitHub API 통합
│   ├── gitlab/            # GitLab API 통합
│   ├── gitea/             # Gitea API 통합
│   └── synclone/          # 클론 설정 및 검증
├── scripts/               # 유틸리티 스크립트
│   ├── simple-benchmark.sh      # 빠른 성능 체크
│   └── benchmark-performance.sh # 상세 성능 분석
├── specs/                 # 기능 명세서
├── examples/              # 설정 파일 예제
└── docs/                  # 문서
```

### 핵심 설계 원칙

1. **간단한 아키텍처**: CLI 도구에 적합한 직접적인 구현
1. **서비스별 구현**: 각 Git 플랫폼별 전용 패키지
1. **설정 기반 설계**: YAML 설정과 스키마 검증
1. **크로스플랫폼 지원**: Linux, macOS, Windows 네이티브 지원
1. **원자적 작업**: 백업 및 롤백 기능을 가진 안전한 실행
1. **표준 도구 통합**: Go의 표준 pprof 등 표준 도구 활용

## 기여 가이드라인

### 새 기능 추가

1. `specs/`에서 관련 명세 확인 또는 작성
1. 명세에 따라 구현
1. 테스트 작성
1. 문서 업데이트
1. PR 제출

### 코드 스타일

- `make fmt`로 포매팅 (gofumpt + gci 사용)
- `make lint`로 린팅 통과 필수
- 테스트 커버리지 유지
- 의미 있는 커밋 메시지 작성

### 테스트 작성

```bash
# 새 테스트 작성 시
go test ./path/to/package -v

# 특정 테스트 함수 실행
go test ./cmd/git -run "TestExtractRepoNameFromURL" -v

# 커버리지 포함 테스트
make test-coverage
```

## Contributing

## 기여 방법

1. **이슈 확인**: 기존 이슈를 확인하거나 새 이슈 생성
1. **Fork**: 리포지토리 포크
1. **브랜치 생성**: `feature/your-feature-name` 또는 `fix/issue-number`
1. **구현**: 명세 기반 구현 및 테스트 작성
1. **품질 검사**: `make lint-all` 실행
1. **PR 제출**: 상세한 설명과 함께 Pull Request 생성

## 품질 기준

### 필수 체크리스트

- [ ] 모든 테스트 통과 (`make test`)
- [ ] 린팅 통과 (`make lint`)
- [ ] 포매팅 적용 (`make fmt`)
- [ ] 문서 업데이트 (필요시)
- [ ] 성능 회귀 없음 (`./scripts/simple-benchmark.sh`)

### 커밋 메시지 형식

```
<type>(<scope>): <description>

<body>

<footer>
```

예시:

```
feat(ide): add JetBrains settings sync monitoring

- Implement real-time file system monitoring
- Add automatic backup before sync fixes
- Support cross-platform path detection

Closes #123
```

## 라이선스

이 프로젝트는 MIT 라이선스 하에 배포됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

______________________________________________________________________

**개발 중인 기능들**:

- 🚧 **Manual Page Generation** (`gz man`): Unix 매뉴얼 페이지 자동 생성 (코드 존재, 비활성화)
- 🚧 **Interactive Shell** (`gz shell`): 디버깅용 인터랙티브 셸 (디버그 모드에서만 활성화)
- 🚧 **Actions Policy Management** (`gz actions-policy`): GitHub Actions 정책 관리 (코드 존재, 비활성화)

이 도구는 지속적으로 발전하고 있으며, 개발자 워크플로우를 개선하기 위한 새로운 기능들이 계속 추가되고 있습니다.
