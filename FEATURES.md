# Features

This document describes the implemented functionality of gzh-manager-go (gz CLI tool).

## 🚀 최근 완료된 주요 기능들

### GitHub Organization & Repository 관리 고도화

- **정책 템플릿 시스템**: 보안 강화, 오픈소스, 엔터프라이즈용 정책 템플릿 미리 제공
- **정책 준수 감사**: 조직 전체 정책 준수 여부 자동 검사 및 상세 리포트 생성
- **예외 처리**: 리포지토리별 정책 예외 처리 및 문서화 지원
- **상속 및 오버라이드**: 템플릿 상속 구조로 유연한 정책 관리 가능

### 성능 개선 사항

- **병렬 처리**: 최대 5개 리포지토리 동시 클론 지원으로 안정적인 처리 보장
- **중단된 작업 재개**: 상태 저장 시스템으로 중단된 작업을 이어서 진행 가능
- **프로그레스 바 세분화**: 리포지토리별 진행률 표시로 세밀한 진행 상황 파악
- **고급 클론 전략**: reset, pull, fetch 전략으로 기존 리포지토리 효율적 관리

### 통합 설정 시스템 완성

- **gzh.yaml 통합 설정**: 모든 도구의 설정을 하나의 파일로 통합 관리
- **설정 마이그레이션 도구**: 기존 bulk-clone.yaml을 gzh.yaml로 자동 변환
- **대화형 설정 생성**: `gz config init` 명령으로 안내식 설정 파일 생성
- **설정 우선순위 체계**: CLI 플래그 > 환경변수 > 설정파일 > 기본값 순서 확립

### 네트워크 환경 관리 시스템

- **수동 네트워크 전환**: 명령어를 통한 VPN/DNS/프록시 설정 전환
- **포괄적인 네트워크 액션**: VPN, DNS, 프록시, 호스트 파일 관리 통합
- **안전한 설정 변경**: 모든 변경사항 자동 백업 및 롤백 기능
- **크로스 플랫폼 지원**: Linux, macOS, Windows에서 동작

## Repository Management

### Bulk Repository Cloning

- **Multi-platform Git hosting support**: Clone entire organizations from GitHub, GitLab, Gitea, and Gogs
- **Flexible cloning strategies**: Choose between reset, pull, or fetch strategies for existing repositories
- **Protocol flexibility**: Support for both HTTPS and SSH protocols with automatic authentication
- **Private repository support**: Token-based authentication for accessing private repositories
- **Configuration-driven**: YAML configuration files with environment-specific overrides (home, work, etc.)
- **Kustomize-style configuration**: Layer multiple configuration files for different environments
- **gzh.yaml integration**: Native support for gzh.yaml configuration format with `--use-gzh-config` option
- **Provider-based organization cloning**: Configure multiple organizations and groups across different Git hosting platforms
- **Visibility filtering**: Filter repositories by visibility (public, private, all) per organization
- **Regex-based filtering**: Use regular expressions to match specific repository names with the `match` field
- **Flexible directory structure**: Flatten option to control directory hierarchy and organization

### SSH Configuration Management

- **Automated SSH config generation**: Create SSH configurations for Git repositories
- **Multi-service support**: Generate configs for GitHub, GitLab, Gitea, and Gogs
- **Key management**: Automatic SSH key association and configuration

### GitHub Organization & Repository Management

- **Repository configuration management**: Comprehensive GitHub repository settings control through `gz repo-config` command
- **Bulk operations**: Apply configuration changes across entire organizations or selected repositories
- **Schema-driven configuration**: YAML-based repository settings with validation and templating
- **API integration**: Full GitHub API client with rate limiting and retry logic (concurrent operations limited to 5)
- **Security features**: Token permission validation and confirmation prompts for sensitive operations
- **Change tracking**: Configuration change history logging with rollback capabilities
- **Dry-run mode**: Preview changes before applying them to repositories
- **Multi-command interface**: List current settings, apply configurations, and validate schemas
- **Organization-wide operations**: Manage repository settings across all repositories in an organization
- **Automated validation**: Ensure token permissions match required operations before execution

## Package Management

### Always-Latest Package Updates

- **Multi-package manager support**: Automated updates for asdf, Homebrew, SDKMAN, MacPorts, APT, and rbenv
- **Flexible update strategies**:
  - Minor latest: Update to latest minor version within the same major version
  - Major latest: Update to the absolute latest version
- **Bulk package operations**: Update multiple packages and tools simultaneously
- **Cross-platform compatibility**: Works across Linux, macOS, and Windows where applicable

## Development Environment Management

### Configuration Backup and Restore

- **Cloud service configurations**: Save and restore AWS, Google Cloud (gcloud) configurations and credentials
- **Container configurations**: Docker configuration management
- **Kubernetes integration**: kubeconfig backup and restore for cluster management
- **SSH configuration**: Complete SSH config save/load functionality
- **Metadata tracking**: Track save dates, descriptions, and source paths for all configurations
- **Safe operations**: Automatic backups before loading configurations

## Network Environment Management

### System Service Monitoring

- **Comprehensive daemon monitoring**: Monitor and manage system services (daemons) with real-time status updates
- **Network service filtering**: Identify and monitor network-related services specifically
- **Service dependency tracking**: Understand service relationships and dependencies
- **Live monitoring**: Real-time service status updates with configurable intervals
- **Cross-platform support**: Works with systemctl, service managers across different operating systems

### Network Profile Management

- **Network profile switching**: Manually switch between different network configurations
- **Command-driven actions**: Execute network-specific actions via CLI commands
- **YAML-based action configuration**: Define network-specific actions using flexible configuration files
- **Profile management**: Save and apply different network profiles for various environments
- **Dry-run testing**: Test configurations safely without executing actual commands

### 네트워크 설정 액션

- **VPN 연결 관리**: OpenVPN, WireGuard, NetworkManager를 통한 VPN 연결/해제 자동화
- **DNS 설정 전환**: resolvectl, NetworkManager를 사용하여 네트워크 환경에 맞는 DNS 서버 자동 설정
- **프록시 관리**: HTTP/HTTPS/SOCKS 프록시 설정 및 환경 변수를 통한 시스템 전체 적용
- **호스트 파일 관리**: 시스템 호스트 파일에 엔트리 추가/제거, 자동 백업 기능 제공
- **통합 자동화**: WiFi 네트워크 변경 시 네트워크 설정을 자동으로 실행하는 완전 자동화 시스템
- **안전 기능**: 자동 백업, 드라이런 모드, 시스템 변경 전 검증 기능

### 네트워크 환경 전환

- **원활한 환경 전환**: 네트워크 간 이동 시 (집, 사무실, 공공 WiFi) 수동으로 시스템 설정 전환
- **프로필 기반 설정**: 각 네트워크별 VPN, DNS, 프록시, 호스트 설정을 프로필로 관리
- **명령어 기반 전환**: CLI 명령어를 통해 네트워크 프로필과 설정을 전환
- **롤백 기능**: 안전한 설정 변경을 위한 자동 백업 및 복원 기능

### 완료된 네트워크 환경 관리 기능

- **✅ 데몬 모니터링**: 시스템 서비스 상태 실시간 모니터링 및 관리
- **✅ 수동 네트워크 액션 시스템**: 명령어를 통한 네트워크 프로필 전환
- **✅ 네트워크 설정 관리**: VPN, DNS, 프록시, 호스트 파일 변경 기능 완료

## Configuration Management

### 통합 설정 시스템 (gzh.yaml)

- **✅ 통합 설정 포맷**: 도구별 설정을 gzh.yaml 파일로 관리하는 스키마 정의 (부분 지원)
- **✅ 스키마 검증**: JSON/YAML 스키마 검증 기능과 내장된 필드 검증 및 열거형 검사 완료
- **✅ 설정 파일 계층 구조**: 자동 발견 기능과 우선순위 (./gzh.yaml → ~/.config/gzh.yaml → 시스템 전체) 완료
- **✅ 환경 변수 치환**: os.ExpandEnv를 사용한 동적 설정 값 지원으로 유연한 배포 가능
- **✅ 설정 프로필**: 개발/운영 환경별 프로필 기반 설정 지원
- **✅ 대화형 설정**: `gz config init` 명령을 통한 안내식 설정 파일 생성
- **✅ 검증 도구**: `gz config validate` 명령으로 설정 파일 검증 및 문제 해결
- **✅ 마이그레이션 지원**: 기존 bulk-clone.yaml 형식에서 통합 gzh.yaml로 자동 마이그레이션 도구 완료
- **✅ 공급자 기반 구조**: GitHub, GitLab, Gitea, Gogs 등 다양한 Git 호스팅 공급자별 체계적인 설정 구조

### YAML Configuration System

- **Hierarchical configurations**: Layer multiple YAML files for different environments and contexts
- **Example configurations**: Built-in templates and examples for all major features
- **Configuration validation**: Syntax checking and validation for all configuration files
- **Environment-specific overrides**: Separate configurations for home, work, and other environments

### CLI Interface

- **Comprehensive help system**: Detailed help documentation for all commands and options
- **Consistent command structure**: Logical command hierarchy across all functionality
- **Rich output formatting**: Color-coded, emoji-enhanced output for better user experience
- **Verbose and dry-run modes**: Detailed logging and safe testing options across all commands

## IDE and Development Tools

### JetBrains IDE Settings Management

- **Cross-platform IDE detection**: Automatic detection of JetBrains products on Linux, macOS, and Windows
- **Real-time settings monitoring**: Track configuration changes across all JetBrains IDE installations using fsnotify
- **Settings synchronization fixes**: Detect and repair common sync issues, particularly with filetypes.xml corruption
- **Multi-IDE support**: Compatible with IntelliJ IDEA, PyCharm, WebStorm, PhpStorm, RubyMine, CLion, GoLand, DataGrip, Android Studio, and Rider
- **Smart file filtering**: Ignore temporary files and focus on meaningful configuration changes
- **Installation discovery**: List all detected JetBrains IDE installations with detailed information
- **Backup and recovery**: Automatic backup creation before applying sync fixes

## Event Management and Automation

### GitHub Event and Webhook Management

- **GitHub event handler**: Process and respond to GitHub webhook events
- **Webhook management**: Create, update, delete, and list webhooks via CLI
- **Bulk webhook operations**: Manage webhooks across multiple repositories simultaneously
- **Event filtering**: Configure which events trigger webhook notifications
- **Webhook validation**: Verify webhook signatures and payloads for security
- **Event routing**: Route events to different handlers based on event type
- **Automation rules**: Define custom automation rules for event processing

## Cross-Platform Support

- **Operating system compatibility**: Linux, macOS, and Windows support where applicable
- **Multiple backend support**: Fallback mechanisms for different system tools and package managers
- **Flexible authentication**: Support for various authentication methods across different services
