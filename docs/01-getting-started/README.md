# 🚀 시작하기

gzh-manager-go CLI 도구 사용을 위한 기본 가이드입니다.

## 📋 이 섹션의 내용

### 🔄 마이그레이션 가이드
기존 도구에서 gzh-manager-go로 전환하는 방법을 안내합니다.

- [📦 bulk-clone → gzh 마이그레이션](migration-guides/bulk-clone-to-gzh.md)
  - 기존 bulk-clone 설정을 gzh.yaml로 변환
  - 설정 호환성 및 새로운 기능 활용법

- [🔄 daemon → CLI 마이그레이션](migration-guides/daemon-to-cli.md)  
  - 데몬 기반에서 CLI 기반으로 전환
  - 설정 및 워크플로우 변경사항

- [🛠️ migrate 명령어 가이드](migration-guides/migrate-command.md)
  - `gz migrate` 명령어 상세 사용법
  - 자동 마이그레이션 도구 활용

## 🎯 주요 기능 소개

### 📦 대량 저장소 클론
```bash
# GitHub 조직 전체 클론
gz bulk-clone --org my-organization

# 여러 플랫폼 동시 클론
gz bulk-clone --config bulk-clone.yaml
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

### 🔧 저장소 설정 관리
```bash
# 조직 정책 일괄 적용
gz repo-config apply --org my-org --policy security.yaml

# 설정 차이점 확인
gz repo-config diff --org my-org
```

## 🚀 빠른 시작

### 1. 설치
```bash
# Homebrew (macOS/Linux)
brew install gzh-manager-go

# 직접 빌드
make build
make install
```

### 2. 기본 설정
```bash
# 설정 초기화
gz config init

# 토큰 설정
export GITHUB_TOKEN="your-token"
export GITLAB_TOKEN="your-token"
```

### 3. 첫 번째 클론
```bash
# 간단한 클론
gz bulk-clone --org your-username

# 설정 파일 사용
gz bulk-clone --config examples/bulk-clone-simple.yaml
```

## 📚 다음 단계

### 초보자 경로
1. [YAML 설정 가이드](../04-configuration/yaml-guide.md)
2. [저장소 관리 빠른 시작](../03-core-features/repository-management/repo-config-quick-start.md)
3. [기본 네트워크 설정](../03-core-features/network-management/)

### 고급 사용자 경로
1. [아키텍처 이해](../02-architecture/overview.md)
2. [고급 설정](../04-configuration/configuration-guide.md)
3. [엔터프라이즈 기능](../09-enterprise/)

## 💡 도움말

### 자주 묻는 질문
- **Q: 어떤 Git 플랫폼을 지원하나요?**
  - A: GitHub, GitLab, Gitea, Gogs를 지원합니다.

- **Q: 기존 설정을 어떻게 마이그레이션하나요?**
  - A: [마이그레이션 가이드](migration-guides/)를 참조하세요.

- **Q: 프록시 환경에서 사용할 수 있나요?**
  - A: 네, [네트워크 관리](../03-core-features/network-management/) 문서를 참조하세요.

### 추가 리소스
- [📖 전체 문서 목록](../INDEX.md)
- [🐛 문제 해결](../06-development/debugging-guide.md)
- [🔧 설정 참조](../04-configuration/)

---

*💡 팁: 명령어에 `--help` 플래그를 사용하면 상세한 도움말을 볼 수 있습니다.*