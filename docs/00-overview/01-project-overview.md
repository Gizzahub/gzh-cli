# 📚 gzh-cli 프로젝트 개요

> **gzh-cli**는 개발자를 위한 종합적인 CLI 도구로, 개발 환경과 Git 저장소를 통합적으로 관리합니다.

## 🚀 핵심 기능

### 리포지토리 관리

- **다중 플랫폼 동기화**: GitHub/GitLab/Gitea/Gogs 조직 전체 저장소 일괄 클론 및 동기화
- **크로스 플랫폼 동기화**: `gz git repo sync`로 GitHub ↔ GitLab ↔ Gitea 간 저장소 동기화
- **스마트 전략**: 다양한 동기화 전략 (reset, pull, fetch, rebase, clone)
- **Git 통합 관리**: 저장소 생성/삭제/아카이브, 웹훅, 이벤트 통합 인터페이스

### 개발 환경 관리

- **클라우드 프로필**: AWS/GCP/Azure 클라우드 설정 관리
- **향상된 IDE 관리**:
  - **IDE 스캔**: 시스템에 설치된 모든 IDE 자동 감지 (JetBrains, VS Code, 기타 에디터)
  - **상태 모니터링**: IDE 프로세스, 메모리, 프로젝트 상태 실시간 확인
  - **프로젝트 열기**: 감지된 IDE로 프로젝트 직접 열기
- **패키지 관리**: asdf, Homebrew, SDKMAN, npm, pip 등 통합 관리
- **네트워크 관리**: WiFi 프로필, VPN, 프록시 자동 전환

### 코드 품질 및 성능

- **코드 품질 관리**: 다중 언어 포매팅/린팅 도구 통합 실행 (테스트 커버리지 34.4%)
- **성능 프로파일링**: Go pprof 기반 성능 분석 도구 (테스트 커버리지 36.6%)
- **진단 도구**: 시스템 상태 및 설정 문제 진단 (테스트 커버리지 10.3%)

## 🎯 주요 명령어

```bash
# 리포지토리 동기화
gz synclone github --org myorg

# 단일 저장소 관리
gz git repo clone-or-update https://github.com/user/repo.git

# 크로스 플랫폼 동기화
gz git repo sync --from github:org/repo --to gitlab:group/repo

# IDE 관리
gz ide scan          # 설치된 IDE 스캔
gz ide status        # IDE 상태 확인
gz ide open .        # 현재 디렉토리를 IDE로 열기
gz ide monitor       # JetBrains IDE 모니터링

# 코드 품질 검사
gz quality run

# 개발 환경 관리
gz dev-env aws setup

# 성능 프로파일링
gz profile server
```

## 📖 문서 구조

### 필수 문서

- **[설치 가이드](../10-getting-started/)** - 설치 및 초기 설정
- **[설정 가이드](../40-configuration/40-configuration-guide.md)** - 통합 설정 시스템
- **[명령어 참조](../50-api-reference/50-command-reference.md)** - 완전한 명령어 문서

### 기능별 가이드

- **[Synclone](../30-features/30-synclone.md)** - 다중 플랫폼 저장소 동기화
- **[리포지토리 관리](../30-features/31-repository-management.md)** - 저장소 관리 기능
- **[개발 환경](../30-features/33-development-environment.md)** - 개발 환경 설정
- **[네트워크 관리](../30-features/34-network-management.md)** - 네트워크 환경 관리

### 개발자 리소스

- **[아키텍처](../20-architecture/20-system-overview.md)** - 시스템 설계 및 패턴
- **[개발 가이드](../60-development/)** - 개발 지침 및 표준
- **[배포 가이드](../70-deployment/)** - 릴리스 및 배포 프로세스

## 🔧 지원 플랫폼

### Git 플랫폼

- GitHub (github.com, GitHub Enterprise)
- GitLab (gitlab.com, self-hosted)
- Gitea (self-hosted)
- Gogs (self-hosted)

### 운영체제

- Linux (Ubuntu, CentOS, Arch, 등)
- macOS (Intel, Apple Silicon)
- Windows (WSL 권장)

### 개발 도구

- **JetBrains IDEs**: IntelliJ IDEA, GoLand, WebStorm, PyCharm, 등
- **VS Code 계열**: Visual Studio Code, VS Code Insiders, Cursor, VSCodium
- **기타 에디터**: Sublime Text, Vim, Neovim, Emacs
- **패키지 매니저**: asdf, Homebrew, SDKMAN, npm, pip, cargo

## 🚀 시작하기

### 새로운 사용자

1. **설치**: [설치 가이드](../10-getting-started/10-installation.md) 참조
1. **첫 설정**: [빠른 시작](../10-getting-started/11-quick-start.md) 가이드 따라하기
1. **설정**: [설정 가이드](../40-configuration/40-configuration-guide.md)로 환경 구성
1. **사용**: [명령어 참조](../50-api-reference/50-command-reference.md)에서 필요한 명령어 찾기

### 기존 사용자

- **업그레이드**: [마이그레이션 가이드](../10-getting-started/migration-guides/) 확인
- **새 기능**: [릴리스 노트](../70-deployment/release-notes/) 확인
- **문제 해결**: [트러블슈팅](../90-maintenance/90-troubleshooting.md) 가이드

## 💡 도움말 및 지원

- **명령어 도움말**: `gz --help` 또는 `gz <command> --help`
- **설정 검증**: `gz config validate`
- **시스템 진단**: `gz doctor`
- **성능 분석**: `gz profile`

______________________________________________________________________

**프로젝트**: gzh-cli
**바이너리 이름**: `gz`
**최근 업데이트**: 2025-08-22
**문서 버전**: 2.1.0
**테스트 커버리지**: Git (91.7%), Shell (69.2%), IDE (40.4%), Profile (36.6%), Quality (34.4%), Doctor (10.3%)
