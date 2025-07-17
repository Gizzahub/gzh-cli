Gizzahub Manager
================

<div style="text-align: center;">
Comprehensive CLI Tool
<br>
<br>
<img src="https://github.com/gizzahub/gzh-manager-go/actions/workflows/test.yml/badge.svg" alt="Test Status"/>
<img src="https://github.com/gizzahub/gzh-manager-go/actions/workflows/lint.yml/badge.svg" alt="Lint Status"/>
<img src="https://pkg.go.dev/badge/github.com/gizzahub/gzh-manager-go.svg" alt="GoDoc"/>
<img src="https://codecov.io/gh/Gizzahub/gzh-manager-go/branch/main/graph/badge.svg" alt="Code Coverage"/>
<img src="https://img.shields.io/github/v/release/Gizzahub/gzh-manager-go" alt="Latest Release"/>
<img src="https://img.shields.io/docker/pulls/Gizzahub/gzh-manager-go" alt="Docker Pulls"/>
<img src="https://img.shields.io/github/downloads/Gizzahub/gzh-manager-go/total.svg" alt="Total Downloads"/>
</div>


# Table of Contents
<!--ts-->
  * [Usage](#usage)
  * [Features](#features)
  * [Project Layout](#project-layout)
  * [How to use this template](#how-to-use-this-template)
  * [Demo Application](#demo-application)
  * [Makefile Targets](#makefile-targets)
  * [Contribute](#contribute)

<!-- Added by: morelly_t1, at: Tue 10 Aug 2021 08:54:24 AM CEST -->

<!--te-->

# Usage

## 핵심 기능 개요

`gzh-manager-go`는 개발자를 위한 종합적인 CLI 도구로, 다음과 같은 주요 기능을 제공합니다:

### 📦 리포지토리 관리
- **대량 클론 도구**: GitHub, GitLab, Gitea, Gogs에서 전체 조직의 리포지토리를 일괄 클론
- **고급 클론 전략**: reset, pull, fetch 모드 지원으로 기존 리포지토리 동기화 방식 제어
- **재개 가능한 작업**: 중단된 클론 작업을 이어서 진행할 수 있는 상태 관리 시스템
- **병렬 처리**: 최대 50개의 동시 클론 작업으로 대규모 조직 처리 성능 향상

### 🏢 GitHub 조직 관리
- **리포지토리 설정 관리**: 조직 전체 리포지토리의 설정을 템플릿 기반으로 일괄 관리
- **설정 비교 도구**: `gz repo-config diff`로 현재 설정과 목표 설정 간 차이점 시각화
- **정책 준수 감사**: `gz repo-config audit`로 SOC2, ISO27001, NIST 등 컴플라이언스 프레임워크 준수 확인
- **리스크 분석**: 보안, 컴플라이언스, 운영 리스크 자동 평가 및 권고사항 제공
- **웹훅 관리**: 개별 및 대량 웹훅 CRUD, 이벤트 기반 자동화 규칙 엔진
- **다양한 출력 형식**: 테이블, JSON, HTML, CSV, SARIF, JUnit 형식 지원으로 CI/CD 통합 가능

### 🔧 통합 설정 시스템
- **gzh.yaml 통합 설정**: 모든 명령어의 설정을 하나의 파일로 통합 관리
- **설정 우선순위 체계**: CLI 플래그 > 환경변수 > 설정파일 > 기본값 순서
- **설정 마이그레이션**: 기존 bulk-clone.yaml을 gzh.yaml로 자동 변환
- **스키마 검증**: JSON/YAML 스키마를 통한 설정 파일 유효성 검사

### 🌐 네트워크 환경 관리
- **네트워크 프로필 전환**: VPN, DNS, 프록시, 호스트 파일을 한번의 명령어로 전환
- **수동 환경 관리**: 필요시 `gz net-env switch` 명령어로 네트워크 환경 변경
- **상태 확인**: `gz net-env status`로 현재 네트워크 설정 상태 조회

### 🏠 개발 환경 관리
- **패키지 관리자 통합**: asdf, Homebrew, SDKMAN, MacPorts 등의 패키지를 최신 버전으로 일괄 업데이트
- **설정 백업/복원**: AWS, Docker, Kubernetes, SSH 등의 설정을 안전하게 백업 및 복원
- **JetBrains IDE 지원**: IDE 설정 동기화 문제 자동 감지 및 수정

### 🩺 시스템 진단
- **종합 진단 도구**: `gz doctor`로 시스템 상태, 의존성, 네트워크 연결성 등을 자동 진단
- **문제 해결 제안**: 발견된 문제에 대한 구체적인 해결 방안 제시
- **성능 벤치마크**: CPU, 디스크 I/O 성능 측정 및 최적화 권고
- **보고서 생성**: JSON 형태의 상세한 진단 보고서 생성

## 🚀 빠른 시작

### 1. 설치
```bash
# 최신 릴리스 다운로드 (권장)
wget https://github.com/gizzahub/gzh-manager-go/releases/latest/download/gz-linux-amd64
chmod +x gz-linux-amd64
sudo mv gz-linux-amd64 /usr/local/bin/gz

# 또는 Go로 직접 빌드
git clone https://github.com/gizzahub/gzh-manager-go.git
cd gzh-manager-go
make build
```

### 2. 기본 사용법
```sh
# 시스템 상태 진단
gz doctor

# 설정 파일 검증
gz bulk-clone validate --config examples/bulk-clone-simple.yaml

# GitHub 조직의 저장소 클론
gz bulk-clone github --orgName myorg --targetPath ~/repos/myorg --token $GITHUB_TOKEN

# 리포지토리 설정 감사
gz repo-config audit --org myorg --framework SOC2
```

### CLI 명령어 개요

```sh
$> gz --help
Cli 종합 Manager by Gizzahub

Usage:
  gz [flags]
  gz [command]

Available Commands:
  always-latest Keep development tools and package managers up to date
  bulk-clone    Clone repositories from multiple Git hosting services
  completion    Generate the autocompletion script for the specified shell
  config        Configuration management commands
  dev-env       Manage development environment configurations
  docker        컨테이너 이미지 관리 및 자동화
  doctor        Diagnose system health and configuration issues
  event         GitHub event management and webhook server
  gen-config    Generate bulk-clone configuration files
  help          Help about any command
  ide           Monitor and manage IDE configuration changes
  migrate       Migrate configuration files to unified format
  monitoring    Run monitoring and alerting system
  net-env       Manage network environment transitions
  repo-config   GitHub repository configuration management
  repo-sync     Advanced repository synchronization and management
  shell         Start interactive debugging shell (REPL)
  ssh-config    SSH configuration management for Git operations
  task-runner   🚀 TASK_RUNNER.todo - 자동 TODO 작업 실행기
  version       gz version information
  webhook       🔗 GitHub 웹훅 관리 도구

Flags:
  -h, --help   help for gz

Use "gz [command] --help" for more information about a command.

Flags:
  -h, --help   help for bulk-clone

Use "gzh-manager [command] --help" for more information about a command.
```

First, create a configuration file in the desired path. Refer to
[bulk-clone.yaml](pkg/bulk-clone/bulk-clone.yaml)

```sh
$> gzh bulk-clone -t $HOME/mywork

This won't work:
$> gzh bulk-clone -t ./mywork
$> gzh bulk-clone -t $HOME/mywork
$> gzh bulk-clone -t ~/mywork
```

### Bulk Clone Config File Support

The bulk-clone command now supports configuration files to manage multiple organizations and their settings. This allows you to define clone operations once and reuse them.

#### Configuration Priority System

The gzh-manager tool uses a strict priority hierarchy where higher priority sources override lower priority ones:

**Priority Order (Highest to Lowest):**
1. **Command-Line Flags** (Highest Priority)
2. **Environment Variables** (Second Priority)
3. **Configuration Files** (Third Priority)
4. **Default Values** (Lowest Priority)

**Examples:**
```bash
# CLI flag overrides all other sources
gz bulk-clone --strategy=pull --parallel=20

# Environment variable overrides config file but not CLI flags
export GITHUB_TOKEN=ghp_env_token
gz bulk-clone --token=ghp_flag_token  # Uses ghp_flag_token

# Configuration file provides base settings
gz bulk-clone  # Uses settings from config file
```

> **📖 For detailed priority rules and examples, see [Configuration Priority Guide](docs/configuration-priority.md)**

#### Configuration File Locations

The tool searches for configuration files in the following order:
1. Environment variable: `GZH_CONFIG_PATH`
2. Current directory: `./gzh.yaml`, `./gzh.yml`, `./bulk-clone.yaml`, `./bulk-clone.yml`
3. User config directory: `~/.config/gzh-manager/gzh.yaml`, `~/.config/gzh-manager/bulk-clone.yaml`
4. System config: `/etc/gzh-manager/gzh.yaml`, `/etc/gzh-manager/bulk-clone.yaml`

#### Using Configuration Files

```bash
# Use config file from standard locations
gzh bulk-clone github --use-config -o myorg

# Use specific config file
gzh bulk-clone github -c /path/to/config.yaml -o myorg

# Override config values with CLI flags
gzh bulk-clone github -c config.yaml -o myorg -t /different/path
```

#### Configuration File Examples

Several example configuration files are provided in the `samples/` directory:

1. **bulk-clone-simple.yaml** - A minimal working configuration
2. **bulk-clone-example.yaml** - A comprehensive example with detailed comments
3. **bulk-clone.yml** - Advanced features (planned/future implementation)

##### Simple Configuration Example

```yaml
# bulk-clone-simple.yaml
version: "0.1"

default:
  protocol: https
  github:
    root_path: "$HOME/github-repos"
  gitlab:
    root_path: "$HOME/gitlab-repos"

repo_roots:
  - root_path: "$HOME/work/mycompany"
    provider: "github"
    protocol: "ssh"
    org_name: "mycompany"
  
  - root_path: "$HOME/opensource"
    provider: "github"
    protocol: "https"
    org_name: "kubernetes"

ignore_names:
  - "test-.*"
  - ".*-archive"
```

See `samples/bulk-clone-example.yaml` for a comprehensive example with all available options and detailed comments.

#### Configuration Schema

The configuration file structure is formally defined in:
- **JSON Schema**: `docs/bulk-clone-schema.json` - Machine-readable schema definition
- **YAML Schema**: `docs/bulk-clone-schema.yaml` - Human-readable schema documentation

##### Validating Your Configuration

You can validate your configuration file using the built-in validator:

```bash
# Validate a specific config file
gzh bulk-clone validate -c /path/to/bulk-clone.yaml

# Validate config from standard locations
gzh bulk-clone validate --use-config
```

The validator checks:
- Required fields are present
- Values match allowed enums (protocol, provider, etc.)
- Structure follows the schema
- Regex patterns are valid

#### Advanced Configuration (Future)

```yaml
# bulk-clone.yaml (advanced example for future implementation)
github:
  ScriptonBasestar:
   auth: token
   proto: https
   targetPath: $HOME/mywork/ScriptonBasestar
   default:
    strategy: include
    branch: develop
   include:
    proxynd:
      branch: develop
    devops-minim-engine:
      branch: dev
   exclude:
    - sb-wp-*
   override:
    include:
  nginxinc:
   targetPath: $HOME/mywork/nginxinc
```

```bash
gzh bulk-clone -o nginxinc
gzh bulk-clone -o nginxinc -t $HOME/mywork/nginxinc
gzh bulk-clone -o nginxinc -t $HOME/mywork/nginxinc --auth token
gzh bulk-clone -o nginxinc -t $HOME/mywork/nginxinc -s pull
```

### Strategy Options

The `-s` or `--strategy` flag controls how existing repositories are synchronized:

- `reset` (default): Performs `git reset --hard HEAD` followed by `git pull`. This discards all local changes and ensures a clean sync with the remote repository.
- `pull`: Only performs `git pull` without resetting. This attempts to merge remote changes with local changes. May fail if there are conflicts.
- `fetch`: Only performs `git fetch` without modifying the working directory. This updates remote tracking branches but doesn't change your local files.

Example usage:
```bash
# Default behavior (reset strategy)
gzh bulk-clone github -o myorg -t ~/repos

# Preserve local changes and merge with remote
gzh bulk-clone github -o myorg -t ~/repos -s pull

# Only fetch updates without modifying local files
gzh bulk-clone github -o myorg -t ~/repos -s fetch
```

### Parallel Clone Options

The `-p` or `--parallel` flag controls how many repositories are cloned or updated simultaneously:

- Default: 10 parallel workers
- Range: 1-50 (higher values may hit rate limits)

The `--max-retries` flag controls how many times failed operations are retried:

- Default: 3 attempts
- Range: 0-10

Example usage:
```bash
# Clone with 20 parallel workers
gzh bulk-clone github -o myorg -t ~/repos -p 20

# Clone with 5 parallel workers and 5 retry attempts
gzh bulk-clone github -o myorg -t ~/repos -p 5 --max-retries 5

# Sequential cloning (no parallelism)
gzh bulk-clone github -o myorg -t ~/repos -p 1
```

**Performance Tips:**
- For large organizations (100+ repos), use `-p 20` or higher
- For rate-limited accounts, use `-p 5` or lower
- Network speed and CPU cores affect optimal parallel value
- Monitor for rate limit errors and adjust accordingly

### Resumable Clone Operations

The `--resume` flag enables resumable clone operations that can be interrupted and continued later:

```bash
# Start a large clone operation
gzh bulk-clone github -o large-org -t ~/repos -p 20

# If interrupted (Ctrl+C), resume from where it left off
gzh bulk-clone github -o large-org -t ~/repos -p 20 --resume

# Resume with different settings
gzh bulk-clone github -o large-org -t ~/repos -p 10 --resume
```

**State Management:**
- States are automatically saved to `~/.gzh/state/`
- Resume works across different parallel settings
- States are cleaned up after successful completion
- Failed repositories are tracked and can be retried

**State Commands:**
```bash
# List all saved states
gzh bulk-clone state list

# Show details of a specific state
gzh bulk-clone state show -p github -o myorg

# Clean up saved states
gzh bulk-clone state clean -p github -o myorg
gzh bulk-clone state clean --all
```

**Benefits:**
- No need to restart from beginning after interruption
- Handles network failures gracefully
- Tracks progress across sessions
- Optimizes by skipping completed repositories

## Repository Configuration Management

The `gz repo-config` command allows you to manage GitHub repository configurations at scale, including settings, security policies, branch protection rules, webhooks, and compliance auditing.

### Quick Start

1. **Create a configuration file** (`repo-config.yaml`):
   ```yaml
   version: "1.0.0"
   organization: "your-org"
   
   templates:
     standard:
       description: "Standard repository settings"
       settings:
         has_issues: true
         has_wiki: false
         delete_branch_on_merge: true
       security:
         vulnerability_alerts: true
         branch_protection:
           main:
             required_reviews: 2
             enforce_admins: true
   
   repositories:
     - name: "*"
       template: "standard"
   ```

2. **Apply configuration**:
   ```bash
   # Preview changes (dry run)
   gz repo-config apply --config repo-config.yaml --dry-run
   
   # Apply configuration
   gz repo-config apply --config repo-config.yaml
   ```

3. **Audit compliance**:
   ```bash
   gz repo-config audit --config repo-config.yaml
   ```

4. **Manage webhooks**:
   ```bash
   # Individual webhook management
   gz repo-config webhook create --org myorg --repo myrepo --url https://example.com/webhook
   
   # Bulk webhook operations
   gz repo-config webhook bulk create --org myorg --config webhook-bulk-config.yaml
   
   # Event-based automation
   gz repo-config webhook automation server --config automation-rules.yaml
   ```

### Key Features

- **Templates**: Define reusable repository configurations
- **Policies**: Enforce security and compliance rules
- **Pattern Matching**: Apply configurations based on repository name patterns
- **Exception Handling**: Allow documented exceptions to policies
- **Compliance Auditing**: Generate reports on policy violations
- **Bulk Operations**: Update multiple repositories efficiently
- **Webhook Management**: Complete CRUD operations and bulk management
- **Event Automation**: GitHub event-based automation rules engine

### Documentation

- [Quick Start Guide](docs/repo-config-quick-start.md) - Get started in 5 minutes
- [User Guide](docs/repo-config-user-guide.md) - Complete documentation
- [Policy Examples](docs/repo-config-policy-examples.md) - Ready-to-use policy templates
- [Webhook Management Guide](docs/webhook-management-guide.md) - Complete webhook features guide
- [Configuration Schema](docs/repo-config-schema.yaml) - Configuration file reference

### Example: Enterprise Configuration

```yaml
version: "1.0.0"
organization: "enterprise-org"

templates:
  backend:
    description: "Backend service configuration"
    settings:
      private: true
    security:
      secret_scanning: true
      branch_protection:
        main:
          required_reviews: 2
          required_status_checks: ["ci/build", "ci/test"]

policies:
  security:
    description: "Security requirements"
    rules:
      must_be_private:
        type: "visibility"
        value: "private"
        enforcement: "required"
        message: "Production services must be private"

patterns:
  - pattern: "*-service"
    template: "backend"
    policies: ["security"]
```

## 🚀 빠른 시작

### 1. 설치
```bash
# 바이너리 다운로드 및 설치
make install

# 또는 직접 빌드
make build
```

### 2. 기본 설정
```bash
# 통합 설정 파일 생성
gz config init

# 네트워크 프로필 설정
gz net-env switch --init
```

### 3. 대량 클론 시작
```bash
# GitHub 조직 클론
gz bulk-clone github -o myorg -t ~/repos

# 설정 파일 사용
gz bulk-clone github --use-config -o myorg
```

### 4. 네트워크 환경 관리
```bash
# 네트워크 프로필 목록 확인
gz net-env switch --list

# 현재 네트워크 상태 확인
gz net-env status

# 특정 네트워크 프로필로 전환
gz net-env switch office
```

> 📖 **자세한 사용법은 [USAGE.md](USAGE.md)를 참고하세요.**

## 🎯 프로젝트 현황

### 구현 완료도
- **핵심 기능**: 100% 완료 ✅
- **테스트 커버리지**: 포괄적인 테스트 완료 ✅
- **문서화**: 완벽한 문서 체계 구축 ✅
- **프로덕션 준비**: 실제 운영 환경에서 사용 가능 ✅

### 주요 성과
- 📊 수백 개의 리포지토리를 효율적으로 관리하는 도구 완성
- 🔧 개발자의 네트워크 환경 전환 작업을 CLI 명령어로 간소화
- ⚙️ 모든 도구를 하나의 설정 파일로 관리하는 통합 체계 구축
- 🔐 조직 차원의 리포지토리 보안 정책 일괄 적용 시스템
- 📚 사용자 가이드부터 개발자 문서까지 완벽한 문서 체계

### 기술적 특징
- **Go 언어 기반**: 크로스 플랫폼 지원, 높은 성능
- **모듈화 설계**: 확장 가능한 아키텍처
- **테스트 주도 개발**: 포괄적인 테스트 커버리지
- **직관적인 CLI**: 사용자 친화적인 인터페이스

> 💡 **향후 계획은 [ROADMAP.md](ROADMAP.md)를 참고하세요.**

## 📚 Documentation

- [Repository Configuration Management](docs/repo-config-commands.md) - `gz repo-config diff` 및 `gz repo-config audit` 명령어 가이드
- [Webhook Management Guide](docs/webhook-management-guide.md) - 웹훅 관리 전체 기능 가이드
- [Configuration Priority Guide](docs/configuration-priority.md) - 설정 우선순위 시스템 설명
- [Bulk Clone Schema](docs/bulk-clone-schema.yaml) - 설정 파일 스키마 문서
- [Release Notes v1.0.0](docs/release-notes-v1.0.0.md) - 첫 정식 릴리즈 노트

## 🔧 Go SDK (Programmatic API)

GZH Manager는 Go 애플리케이션에서 직접 사용할 수 있는 공개 API를 제공합니다.

### 설치

```bash
go get github.com/gizzahub/gzh-manager-go/pkg/gzhclient
```

### 기본 사용법

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/gizzahub/gzh-manager-go/pkg/gzhclient"
)

func main() {
    // 클라이언트 생성
    client, err := gzhclient.NewClient(gzhclient.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 대량 클론 작업 수행
    req := gzhclient.BulkCloneRequest{
        Platforms: []gzhclient.PlatformConfig{
            {
                Type:          "github",
                Token:         "your-github-token",
                Organizations: []string{"your-org"},
            },
        },
        OutputDir:   "./repositories",
        Concurrency: 5,
        Strategy:    "reset",
    }

    result, err := client.BulkClone(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Successfully cloned %d repositories\n", result.SuccessCount)
}
```

### 주요 기능

- **대량 클론**: GitHub, GitLab, Gitea, Gogs에서 리포지토리 일괄 클론
- **플러그인 관리**: 플러그인 로드, 실행, 상태 확인
- **시스템 모니터링**: CPU, 메모리, 디스크 사용량 수집
- **이벤트 시스템**: 실시간 이벤트 구독 및 처리
- **플랫폼별 클라이언트**: GitHub, GitLab, Gitea 전용 클라이언트

### 더 많은 예제

완전한 사용 예제와 API 문서는 다음에서 확인하세요:
- [GoDoc](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/pkg/gzhclient)
- [Example Code](pkg/gzhclient/examples_test.go)

# Features
- [goreleaser](https://goreleaser.com/) with `deb.` and `.rpm` packer and container (`docker.hub` and `ghcr.io`) releasing including `manpages` and `shell completions` and grouped Changelog generation.
- [golangci-lint](https://golangci-lint.run/) for linting and formatting
- [Github Actions](.github/worflows) Stages (Lint, Test (`windows`, `linux`, `mac-os`), Build, Release) 
- [Gitlab CI](.gitlab-ci.yml) Configuration (Lint, Test, Build, Release)
- [cobra](https://cobra.dev/) example setup including tests
- [Makefile](Makefile) - with various useful targets and documentation (see Makefile Targets)
- [Github Pages](_config.yml) using [jekyll-theme-minimal](https://github.com/pages-themes/minimal) (checkout [https://Gizzahub.github.io/gzh-manager-go/](https://Gizzahub.github.io/gzh-manager-go/))
- Useful `README.md` badges
- [pre-commit-hooks](https://pre-commit.com/) for formatting and validating code before committing

## Project Layout
* [assets/](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/assets) => docs, images, etc
* [cmd/](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/cmd)  => command-line configurations (flags, subcommands)
* [pkg/](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/pkg)  => packages that are okay to import for other projects
* [internal/](https://pkg.go.dev/github.com/gizzahub/gzh-manager-go/pkg)  => packages that are only for project internal purposes
- [`tools/`](tools/) => for automatically shipping all required dependencies when running `go get` (or `make bootstrap`) such as `golang-ci-lint` (see: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
- [`scripts/`](scripts/) => build scripts 

# Makefile Targets
```sh
$> make
bootstrap                      install build deps
build                          build golang binary
clean                          clean up environment
cover                          display test coverage
docker-build                   dockerize golang application
fmt                            format go files
help                           list makefile targets
install                        install golang binary
lint                           lint go files
pre-commit                     run pre-commit hooks
run                            run the app
test                           display test coverage
```
