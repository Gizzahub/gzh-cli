# Gizzahub Manager (gzh-cli)

**통합 개발 환경 CLI 도구**

![Test Status](https://github.com/gizzahub/gzh-cli/actions/workflows/test.yml/badge.svg)
![Lint Status](https://github.com/gizzahub/gzh-cli/actions/workflows/lint.yml/badge.svg)
![GoDoc](https://pkg.go.dev/badge/github.com/gizzahub/gzh-cli.svg)
![Code Coverage](https://codecov.io/gh/Gizzahub/gzh-cli/branch/main/graph/badge.svg)
![Latest Release](https://img.shields.io/github/v/release/Gizzahub/gzh-cli)

---

## 개요

`gzh-cli` (바이너리: `gz`)는 개발자를 위한 종합 CLI 도구로, Git 플랫폼 통합 관리, IDE 모니터링, 코드 품질 관리, 개발 환경 설정을 하나의 명령어로 통합합니다.

**핵심 가치**:
- 🔗 **다중 플랫폼 통합**: GitHub, GitLab, Gitea, Gogs를 하나의 인터페이스로
- 🛠️ **개발 워크플로우 자동화**: IDE, 코드 품질, 패키지 매니저 통합 관리
- 📦 **확장 가능한 아키텍처**: Integration Libraries Pattern으로 모듈화

---

## 빠른 시작

### 설치

```bash
# Go로 설치 (권장)
go install github.com/Gizzahub/gzh-cli/cmd/gz@latest

# 소스에서 빌드
git clone https://github.com/Gizzahub/gzh-cli.git
cd gzh-cli
make bootstrap  # 빌드 의존성 설치
make build      # gz 바이너리 생성
make install    # $GOPATH/bin에 설치
```

### 첫 명령어

```bash
# 시스템 상태 진단
gz doctor

# IDE 스캔 및 관리
gz ide scan
gz ide status

# Git 리포지토리 관리
gz git repo clone-or-update https://github.com/user/repo.git
gz git repo pull-all ~/workspace --parallel 5

# 코드 품질 검사
gz quality run
```

### 다음 단계
- 📚 [전체 문서](docs/00-overview/00-index.md)
- 🚀 [설치 가이드](docs/10-getting-started/10-installation.md)
- ⚙️ [설정 가이드](docs/40-configuration/40-configuration-guide.md)

---

## 주요 기능

| 기능 | 설명 | 상세 문서 |
|-----|------|---------|
| **Git 플랫폼 통합** | GitHub/GitLab/Gitea/Gogs 리포지토리 관리, 크로스 플랫폼 동기화 | [📖 Docs](docs/30-features/31-repository-management.md) |
| **IDE 관리** | JetBrains/VS Code 스캔, 상태 모니터링, 프로젝트 열기 | [📖 Docs](docs/30-features/35-ide-management.md) |
| **코드 품질** | 다중 언어 린팅/포매팅 (Go, Python, JS, Rust 등) | [📖 Docs](docs/30-features/36-quality-management.md) |
| **성능 프로파일링** | Go pprof 기반 CPU/메모리 프로파일링 | [📖 Docs](docs/30-features/37-performance-profiling.md) |
| **패키지 매니저** | asdf, Homebrew, SDKMAN, npm, pip 통합 업데이트 | [📖 Docs](docs/30-features/) |
| **쉘 설정 빌더** | .zshrc/.bashrc 모듈화 및 의존성 관리 | [📖 Docs](docs/30-features/) |
| **개발 환경 관리** | AWS, Docker, Kubernetes, SSH 설정 통합 | [📖 Docs](docs/30-features/33-development-environment.md) |
| **네트워크 환경** | WiFi, VPN, DNS, 프록시 자동 전환 | [📖 Docs](docs/30-features/34-network-management.md) |

### 명령어 구조

```bash
gz [command] [subcommand] [flags]

# 주요 명령어
git         # Git 플랫폼 통합 (repo, webhook, event)
ide         # IDE 모니터링 및 관리
quality     # 코드 품질 도구 (포매팅 + 린팅)
profile     # 성능 프로파일링
pm          # 패키지 매니저 관리
shellforge  # 쉘 설정 빌더
synclone    # 대량 리포지토리 클론
dev-env     # 개발 환경 관리
net-env     # 네트워크 환경 관리
repo-config # GitHub 리포지토리 설정
doctor      # 시스템 진단
```

**전체 명령어**: [`gz --help`](docs/50-api-reference/50-command-reference.md)

---

## 🧩 하위 프로젝트 (Subprojects)

gzh-cli는 핵심 기능을 독립 라이브러리로 분리하여 개발합니다. 각 라이브러리는 독립적으로 사용 가능합니다.

| 프로젝트 | 목적 | 독립 사용 | 문서 |
|---------|------|---------|------|
| [gzh-cli-git][git-repo] | 로컬 Git 작업 관리 (clone, pull, push) | ✅ | [📖][git-doc] |
| [gzh-cli-quality][quality-repo] | 다중 언어 코드 품질 도구 | ✅ | [📖][quality-doc] |
| [gzh-cli-package-manager][pm-repo] | 패키지 매니저 통합 관리 | ✅ | [📖][pm-doc] |
| [gzh-cli-shellforge][shell-repo] | 모듈형 쉘 설정 빌더 | ✅ | [📖][shell-doc] |

**통합 아키텍처**: [Integration Libraries Pattern](docs/integration/00-SUBPROJECTS_GUIDE.md)

**코드 감소 효과**: 6,702줄 (92.0% 감소율)

[git-repo]: https://github.com/gizzahub/gzh-cli-git
[git-doc]: https://github.com/gizzahub/gzh-cli-git#readme
[quality-repo]: https://github.com/Gizzahub/gzh-cli-quality
[quality-doc]: https://github.com/Gizzahub/gzh-cli-quality#readme
[pm-repo]: https://github.com/gizzahub/gzh-cli-package-manager
[pm-doc]: https://github.com/gizzahub/gzh-cli-package-manager#readme
[shell-repo]: https://github.com/gizzahub/gzh-cli-shellforge
[shell-doc]: https://github.com/gizzahub/gzh-cli-shellforge#readme

---

## 사용 예제

### Git 리포지토리 관리

```bash
# 스마트 클론/업데이트 (6가지 전략)
gz git repo clone-or-update https://github.com/user/repo.git
gz git repo clone-or-update https://github.com/user/repo.git --strategy rebase --branch develop

# 재귀적 일괄 업데이트 (하위 디렉토리 모든 Git 리포지토리)
gz git repo pull-all ~/workspace --parallel 10 --verbose

# 크로스 플랫폼 동기화
gz git repo sync --from github:org/repo --to gitlab:group/repo
```

### IDE 관리

```bash
# IDE 스캔 및 감지
gz ide scan                  # 24시간 캐시
gz ide scan --refresh        # 캐시 무시

# IDE 상태 확인
gz ide status
gz ide status --running      # 실행 중인 IDE만

# IDE로 프로젝트 열기
gz ide open /path/to/project
gz ide open . --ide goland
```

### 코드 품질

```bash
# 전체 품질 검사 및 수정
gz quality run

# 변경된 파일만 처리
gz quality run --changed

# 린팅만 (수정 없이 검사)
gz quality check
```

### 대량 리포지토리 클론

```bash
# GitHub 조직 전체 클론
gz synclone github --orgName myorg --targetPath ~/repos --token $GITHUB_TOKEN

# GitLab 그룹 클론
gz synclone gitlab --groupName mygroup --targetPath ~/repos --token $GITLAB_TOKEN

# 설정 파일로 실행
gz synclone --config synclone.yaml
```

---

## 문서

### 사용자 가이드
- 📚 [문서 전체 인덱스](docs/00-overview/00-index.md)
- 🚀 [설치 가이드](docs/10-getting-started/10-installation.md)
- 📖 [빠른 시작](docs/10-getting-started/11-quick-start.md)
- ⚙️ [설정 가이드](docs/40-configuration/40-configuration-guide.md)
- 📋 [명령어 레퍼런스](docs/50-api-reference/50-command-reference.md)

### 개발자 가이드
- 🏗️ [아키텍처](docs/20-architecture/)
- 💻 [개발 환경 설정](docs/60-development/60-index.md)
- 🧪 [테스트 가이드](docs/60-development/)
- 🔧 [기여 가이드](docs/CONTRIBUTING.md)

### 추가 리소스
- 🔍 [문제 해결](docs/90-maintenance/90-troubleshooting.md)
- 📈 [성능 모니터링](docs/30-features/37-performance-profiling.md)
- 🔐 [보안 가이드](docs/70-deployment/75-security-guidelines.md)

---

## 설정

### 기본 설정 파일

설정 파일 위치 (우선순위 순):
1. `$GZH_CONFIG_PATH` (환경 변수)
2. `./gzh.yaml` (현재 디렉토리)
3. `~/.config/gzh-manager/gzh.yaml` (사용자 설정)
4. `/etc/gzh-manager/gzh.yaml` (시스템 설정)

### 설정 예제

```yaml
global:
  clone_base_dir: "$HOME/repos"
  default_strategy: reset

providers:
  github:
    token: "${GITHUB_TOKEN}"
    organizations:
      - name: "myorg"
        clone_dir: "$HOME/repos/github/myorg"

  gitlab:
    token: "${GITLAB_TOKEN}"
    groups:
      - name: "mygroup"
        clone_dir: "$HOME/repos/gitlab/mygroup"
```

**상세 설정**: [Configuration Guide](docs/40-configuration/40-configuration-guide.md)

---

## 아키텍처

### Integration Libraries Pattern

gzh-cli는 공통 기능을 외부 라이브러리로 분리하여 **단일 정보 소스(Single Source of Truth)**를 확립합니다.

```
gzh-cli (통합 CLI)
├── cmd/*_wrapper.go (45-473줄) - 얇은 래퍼
└── 외부 라이브러리 통합
    ├── gzh-cli-git (로컬 Git 작업)
    ├── gzh-cli-quality (코드 품질)
    ├── gzh-cli-package-manager (패키지 관리)
    └── gzh-cli-shellforge (쉘 설정)
```

**이점**:
- ✅ 코드 중복 제거 (92% 감소)
- ✅ 독립 사용 가능
- ✅ 단일 정보 소스
- ✅ 유지보수 간소화

**상세 아키텍처**: [Integration Documentation](docs/integration/README.md)

---

## 개발

### 빌드 및 테스트

```bash
# 개발 환경 설정
make bootstrap      # 빌드 의존성 설치 (최초 1회)

# 빌드
make build          # gz 바이너리 생성

# 코드 품질 (커밋 전 필수)
make fmt            # 코드 포매팅
make lint           # 린팅 검사
make test           # 테스트 실행

# 전체 품질 검사
make lint-all       # 포맷 + 린트 + pre-commit
```

### Pre-commit 훅

```bash
# 설치 (최초 1회)
make pre-commit-install

# 수동 실행
make pre-commit
make pre-push
```

### 모듈별 테스트

```bash
# 특정 패키지 테스트
go test ./cmd/git/repo -v
go test ./cmd/ide -v
go test ./pkg/github -v

# 특정 테스트 함수
go test ./cmd/git -run "TestCloneOrUpdate" -v
```

---

## 기여하기

### 기여 프로세스

1. **이슈 확인**: [Issues](https://github.com/Gizzahub/gzh-cli/issues)
2. **Fork & 브랜치**: `feature/your-feature` or `fix/issue-number`
3. **구현**: 코드 작성 + 테스트
4. **품질 검사**: `make lint-all` 통과
5. **PR 제출**: 상세 설명 포함

### 품질 기준

- ✅ `make test` 통과
- ✅ `make lint` 통과
- ✅ `make fmt` 적용
- ✅ 문서 업데이트 (필요시)
- ✅ 커밋 메시지 규칙 준수

**자세한 내용**: [Contributing Guide](docs/CONTRIBUTING.md)

---

## 시스템 요구사항

- **Go**: 1.23.0+
- **Git**: 2.0+
- **OS**: Linux, macOS, Windows (WSL 권장)

---

## 라이선스

MIT License - [LICENSE](LICENSE) 파일 참조

---

## 링크

- **GitHub**: [Gizzahub/gzh-cli](https://github.com/Gizzahub/gzh-cli)
- **문서**: [docs/](docs/)
- **이슈**: [Issues](https://github.com/Gizzahub/gzh-cli/issues)
- **기술 스택**: [TECH_STACK.md](TECH_STACK.md)
- **변경 이력**: [CHANGELOG.md](CHANGELOG.md)

---

**Made with ❤️ by the Gizzahub Team**
