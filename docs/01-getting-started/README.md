# 🚀 시작하기

gzh-manager-go (`gz`) CLI 도구 사용을 위한 기본 가이드입니다.

## 📋 이 섹션의 내용

### 🔄 마이그레이션 가이드

기존 도구에서 gzh-manager-go로 전환하는 방법을 안내합니다.

- [📦 bulk-clone → synclone 마이그레이션](migration-guides/bulk-clone-to-gzh.md)
  - 기존 bulk-clone 설정을 synclone.yaml로 변환
  - 향상된 기능 및 새로운 명령어 활용법

- [🔄 daemon → CLI 마이그레이션](migration-guides/daemon-to-cli.md)
  - 데몬 기반에서 CLI 기반으로 전환
  - 설정 및 워크플로우 변경사항


## 🎯 주요 기능 소개

### 📦 리포지토리 동기화 (synclone)

```bash
# GitHub 조직 전체 클론
gz synclone github --org my-organization

# GitLab 그룹 동기화
gz synclone gitlab --group my-group

# 여러 플랫폼 동시 클론
gz synclone --config synclone.yaml
```

### 🔧 Git 통합 관리

```bash
# 스마트 클론/업데이트
gz git repo clone-or-update https://github.com/user/repo.git

# 저장소 설정 관리
gz git config audit --org myorg
```

### ✨ 코드 품질 관리

```bash
# 다중 언어 포매팅/린팅
gz quality run

# 품질 도구 설치
gz quality install
```

### 💻 IDE 모니터링

```bash
# JetBrains IDE 설정 모니터링
gz ide monitor

# 동기화 문제 해결
gz ide fix-sync
```

### 🌐 네트워크 환경 관리

```bash
# WiFi 프로필 자동 전환
gz net-env auto-switch

# VPN 연결 관리
gz net-env vpn connect office
```

### 💻 개발 환경 설정

```bash
# AWS 프로필 관리
gz dev-env aws --profile production

# 클라우드 환경 동기화
gz dev-env sync --all
```

### 📊 성능 프로파일링

```bash
# CPU 프로파일링 시작
gz profile start --type cpu

# 프로파일 분석
gz profile analyze cpu-profile.pprof
```

### 📦 패키지 매니저 업데이트

```bash
# 모든 패키지 매니저 업데이트
gz pm update --all

# 특정 매니저만 업데이트
gz pm update --managers homebrew,asdf
```

## 🚀 빠른 시작

### 1. 설치

```bash
# 소스에서 빌드 (Go 1.24.0+ 필요)
git clone https://github.com/yourusername/gzh-manager-go.git
cd gzh-manager-go
make bootstrap  # 빌드 도구 설치
make build     # gz 바이너리 생성
make install   # $GOPATH/bin에 설치
```

### 2. 기본 설정

```bash
# 토큰 설정 (필요한 플랫폼만)
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITLAB_TOKEN="glpat-xxxxxxxxxxxx"

# 설정 디렉토리 생성
mkdir -p ~/.config/gzh-manager
```

### 3. 첫 번째 사용

```bash
# 단일 저장소 클론/업데이트
gz git repo clone-or-update https://github.com/user/repo.git

# GitHub 조직 동기화
gz synclone github --org your-organization

# 코드 품질 체크
gz quality run

# IDE 모니터링 시작
gz ide monitor
```

## 📚 다음 단계

### 초보자 경로

1. [리포지토리 동기화 가이드](../03-core-features/synclone-guide.md)
2. [코드 품질 관리](../03-core-features/quality-management.md)
3. [YAML 설정 가이드](../04-configuration/yaml-guide.md)

### 중급 사용자 경로

1. [Git 통합 명령어](../03-core-features/git-unified-command.md)
2. [IDE 모니터링 설정](../03-core-features/ide-management.md)
3. [네트워크 환경 관리](../03-core-features/network-management/)

### 고급 사용자 경로

1. [아키텍처 이해](../02-architecture/overview.md)
2. [성능 프로파일링](../03-core-features/performance-profiling.md)
3. [엔터프라이즈 기능](../09-enterprise/)

## 💡 도움말

### 자주 묻는 질문

- **Q: 어떤 Git 플랫폼을 지원하나요?**
  - A: GitHub, GitLab, Gitea, Gogs를 지원합니다.

- **Q: bulk-clone 명령어는 어디로 갔나요?**
  - A: `gz synclone`으로 개선되었습니다. [마이그레이션 가이드](migration-guides/bulk-clone-to-gzh.md)를 참조하세요.

- **Q: Go 버전 요구사항은?**
  - A: Go 1.24.0 이상이 필요합니다.

- **Q: 프록시 환경에서 사용할 수 있나요?**
  - A: 네, [네트워크 관리](../03-core-features/network-management/) 문서를 참조하세요.

- **Q: 어떤 코드 품질 도구를 지원하나요?**
  - A: Go, Python, JavaScript, Rust, Java, C/C++ 등 다양한 언어의 포매터와 린터를 지원합니다.

### 추가 리소스

- [📖 전체 문서 목록](../INDEX.md)
- [🐛 문제 해결](../06-development/debugging-guide.md)
- [🔧 설정 참조](../04-configuration/)

---

_💡 팁: 명령어에 `--help` 플래그를 사용하면 상세한 도움말을 볼 수 있습니다._

_📅 최종 업데이트: 2025-08-04_
_🔧 Go 버전: 1.24.0+ (toolchain: go1.24.5)_
