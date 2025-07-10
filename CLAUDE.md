# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gzh-manager-go is a comprehensive CLI tool (binary: `gz`) for managing development environments and Git repositories across multiple platforms. It provides bulk operations for cloning organizations, package management, network environment transitions, IDE settings monitoring, and development environment configuration management.

## Essential Commands

### Development Setup
```bash
make bootstrap  # Install all build dependencies - run this first
```

### Building and Running
```bash
make build      # Creates 'gz' executable
make install    # Installs to GOPATH/bin
make run        # Run with version tag
```

### Code Quality - ALWAYS RUN BEFORE COMMITTING
```bash
make fmt        # Format code with gofumpt and gci
make lint       # Run golangci-lint checks
make test       # Run all tests with coverage
```

### Testing
```bash
make test       # Run unit tests with coverage
make cover      # Show coverage with race detection
go test ./cmd/bulk-clone -v        # Run specific package tests
go test ./cmd/ide -v               # Run IDE package tests
go test ./pkg/github -v            # Run GitHub integration tests
```

### Mocking Strategy
```bash
make generate-mocks    # Generate all interface mocks using gomock
make clean-mocks      # Remove all generated mock files
make regenerate-mocks # Clean and regenerate all mocks
```

## Architecture

### Command Structure
- **cmd/** - CLI commands using cobra framework
  - `root.go` - Main entry point with all command registrations
  - `bulk-clone/` - Multi-platform repository cloning (GitHub, GitLab, Gitea, Gogs)
  - `always-latest/` - Package manager updates (asdf, Homebrew, SDKMAN, etc.)
  - `dev-env/` - Development environment management (AWS, Docker, Kubernetes configs)
  - `net-env/` - Network environment transitions (WiFi monitoring, VPN, DNS, proxy)
  - `ide/` - JetBrains IDE settings monitoring and sync fixes
  - `gen-config/` - Configuration file generation and discovery
  - `ssh-config/` - SSH configuration management for Git services

### Core Packages
- **internal/** - Private packages
  - `convert/` - Data conversion utilities
  - `git/` - Git operations and helpers
  - `testlib/` - Testing utilities and environment checkers

- **pkg/** - Public packages (importable by other projects)
  - `bulk-clone/` - Configuration loading, schema validation, URL building
  - `github/` - GitHub API integration and organization cloning
  - `gitlab/` - GitLab API integration and group cloning
  - `gitea/` - Gitea API integration and organization cloning
  - `gogs/` - Gogs API integration (planned)
  - `gen-config/` - Directory-based configuration generation
  - `example/` - Example package structure

- **helpers/** - Utility functions
  - `git_helper.go` - Git repository operations and testing utilities

### Key Patterns
1. **Service-specific implementations**: Each Git platform (GitHub, GitLab, Gitea, Gogs) has dedicated packages following common interfaces
2. **Configuration-driven design**: Extensive YAML configuration support with schema validation (see `samples/` directory)
3. **Cobra CLI framework**: All commands use cobra with consistent flag patterns and help documentation
4. **Cross-platform support**: Native OS detection and platform-specific implementations (Linux, macOS, Windows)
5. **Environment variable integration**: Support for token authentication and configuration overrides
6. **Atomic operations**: Commands designed for safe execution with backup and rollback capabilities
7. **Comprehensive testing**: testify framework with mock services and environment-specific tests

## Configuration and Schema

### Configuration File Hierarchy
1. Environment variable: `GZH_CONFIG_PATH`
2. Current directory: `./bulk-clone.yaml` or `./bulk-clone.yml`
3. User config: `~/.config/gzh-manager/bulk-clone.yaml`
4. System config: `/etc/gzh-manager/bulk-clone.yaml`

### Schema Validation
- JSON Schema: `docs/bulk-clone-schema.json`
- YAML Schema: `docs/bulk-clone-schema.yaml`
- Built-in validator: `gz bulk-clone validate`

### Sample Configurations
- `samples/bulk-clone-simple.yaml` - Minimal working example
- `samples/bulk-clone-example.yaml` - Comprehensive with comments
- `samples/bulk-clone.yml` - Advanced features

## Testing Guidelines
- Test files use `*_test.go` convention
- Uses testify for assertions
- Environment-specific tests check for tokens (GITHUB_TOKEN, GITLAB_TOKEN)
- Integration tests mock external services when possible
- Cross-platform testing for path handling and OS-specific features

## Important Notes
- The binary is named 'gz' not 'gzh-manager-go'
- Always run `make fmt` before committing code
- Configuration files support both CLI flags and YAML config
- Supports multiple authentication methods per service

## Command Categories

### Repository Operations
- `gz bulk-clone` - Clone entire organizations from GitHub, GitLab, Gitea, Gogs
- `gz gen-config` - Generate configuration files from existing repositories
- `gz ssh-config` - Manage SSH configurations for Git services

### Development Environment
- `gz dev-env` - Manage AWS, Docker, Kubernetes configurations
- `gz always-latest` - Update package managers (asdf, Homebrew, SDKMAN)
- `gz ide` - Monitor JetBrains IDE settings and fix sync issues

### Network Management
- `gz net-env` - WiFi change detection, VPN/DNS/proxy management

## Repository Clone Strategies
- `reset` (default): Hard reset + pull (discards local changes)
- `pull`: Merge remote changes with local changes
- `fetch`: Update remote tracking without changing working directory

## Authentication
- Token-based authentication for private repositories
- Environment variable support (GITHUB_TOKEN, GITLAB_TOKEN, etc.)
- SSH key management and configuration

## TODO Runner Workflow

### 🚀 TASK_RUNNER.todo 프롬프트

> **역할**  
> /tasks/todo/ 디렉터리의 미완료 TODO 파일을 순차적으로 읽어 **작업 → 커밋 → 완료 파일 이동**을 자동화합니다.  
> 프롬프트 사용 시 `{DIR=/tasks/todo}` 나 `{DIR=/tasks/todo/my-feature}` 처럼 디렉터리를 지정할 수 있습니다.

#### ✅ 절차

1. **다음 `[ ]` 미완료 항목 하나 선택**   
   - 파일 이름 오름차순 → 항목 순서대로 탐색  
   - 한 번에 1 개 작업만 처리

2. **분석 & 의존성 파악**  
   - 관련 코드/문서 확인  
   - 선행 조건이 누락되면 먼저 TODO 항목으로 세분화하여 같은 파일 최상단에 삽입

3. **구현 & 테스트 & 문서화**  
   - 코드 / 설정 / 문서 수정  
   - 핵심 로직 테스트 코드 작성

4. **포맷 & 커밋**  
   - 포맷터 실행
   - 커밋 전 `[x]` 체크
   - 커밋 메시지  
     ```
     feat(모듈): 작업 요약 (50자 이내)
     ```

5. **파일 이동**  
   - 모든 항목이 `[x]` 되면 파일을 `/tasks/done/` 으로 이동  
     → 파일명 뒤에 `__DONE_YYYYMMDD.md` 자동 추가

#### ⚠️ 유의  
- 작업 불가(스펙 부재 등) → 해당 항목을 `[>]` 로 바꾸고 하단에 이유 서술  
- 긴 작업은 2–4 개 하위 TODO 로 나눈 뒤 **다음 루프**에서 처리