# 📚 gzh-cli 프로젝트 개요

> **gzh-cli**는 개발자를 위한 종합적인 CLI 도구로, 개발 환경과 Git 저장소를 통합적으로 관리합니다.

## 🚀 핵심 기능

### 리포지토리 관리
- **다중 플랫폼 동기화**: GitHub/GitLab/Gitea/Gogs 조직 전체 저장소 일괄 클론 및 동기화
- **스마트 전략**: 다양한 동기화 전략 (reset, pull, fetch, rebase, clone)
- **Git 통합 관리**: 저장소 설정, 웹훅, 이벤트 통합 인터페이스

### 개발 환경 관리
- **클라우드 프로필**: AWS/GCP/Azure 클라우드 설정 관리
- **IDE 모니터링**: JetBrains IDE 설정 실시간 감지 및 동기화
- **패키지 관리**: asdf, Homebrew, SDKMAN, npm, pip 등 통합 관리
- **네트워크 관리**: WiFi 프로필, VPN, 프록시 자동 전환

### 코드 품질 및 성능
- **코드 품질 관리**: 다중 언어 포매팅/린팅 도구 통합 실행
- **성능 프로파일링**: Go pprof 기반 성능 분석 도구
- **진단 도구**: 시스템 상태 및 설정 문제 진단

## 🎯 주요 명령어

```bash
# 리포지토리 동기화
gz synclone github --org myorg

# 단일 저장소 관리
gz git repo clone-or-update https://github.com/user/repo.git

# 코드 품질 검사
gz quality run

# IDE 모니터링
gz ide monitor

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
- JetBrains IDEs (IntelliJ, GoLand, WebStorm, 등)
- VS Code
- 다양한 패키지 매니저 및 도구체인

## 🚀 시작하기

### 새로운 사용자
1. **설치**: [설치 가이드](../10-getting-started/10-installation.md) 참조
2. **첫 설정**: [빠른 시작](../10-getting-started/11-quick-start.md) 가이드 따라하기
3. **설정**: [설정 가이드](../40-configuration/40-configuration-guide.md)로 환경 구성
4. **사용**: [명령어 참조](../50-api-reference/50-command-reference.md)에서 필요한 명령어 찾기

### 기존 사용자
- **업그레이드**: [마이그레이션 가이드](../10-getting-started/migration-guides/) 확인
- **새 기능**: [릴리스 노트](../70-deployment/release-notes/) 확인
- **문제 해결**: [트러블슈팅](../90-maintenance/90-troubleshooting.md) 가이드

## 💡 도움말 및 지원

- **명령어 도움말**: `gz --help` 또는 `gz <command> --help`
- **설정 검증**: `gz config validate`
- **시스템 진단**: `gz doctor`
- **성능 분석**: `gz profile`

---

**프로젝트**: gzh-cli
**바이너리 이름**: `gz`
**최근 업데이트**: 2025-08-19
**문서 버전**: 2.0.0
