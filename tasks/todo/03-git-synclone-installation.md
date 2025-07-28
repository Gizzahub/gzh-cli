# Task: Git Synclone Installation and Distribution

## 작업 목표
`git synclone` 명령어를 사용자가 쉽게 설치하고 사용할 수 있도록 설치 방법과 배포 전략을 구현합니다.

## 선행 조건
- [ ] 01-git-synclone-command-structure.md 완료
- [ ] 02-git-synclone-provider-integration.md 완료
- [ ] Git extension 설치 메커니즘 이해

## 구현 상세

### 1. 바이너리 이름 설정
```makefile
# Makefile 수정
GIT_SYNCLONE_BINARY = git-synclone
GIT_SYNCLONE_CMD = cmd/git-synclone

build-git-extensions:
	go build -o $(GIT_SYNCLONE_BINARY) $(GIT_SYNCLONE_CMD)/main.go
```

### 2. 설치 스크립트 작성
`scripts/install-git-extensions.sh`:
```bash
#!/bin/bash
set -e

# 색상 정의
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "Installing gzh Git extensions..."

# 1. 바이너리 빌드
make build-git-extensions

# 2. 설치 위치 결정
INSTALL_DIR="${HOME}/.local/bin"
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${RED}Warning: $INSTALL_DIR is not in PATH${NC}"
    echo "Add the following to your shell profile:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

# 3. 바이너리 복사
mkdir -p "$INSTALL_DIR"
cp git-synclone "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/git-synclone"

# 4. 설치 확인
if command -v git-synclone &> /dev/null; then
    echo -e "${GREEN}✓ git-synclone installed successfully${NC}"
    echo "Try: git synclone --help"
else
    echo -e "${RED}✗ Installation failed${NC}"
    exit 1
fi
```

### 3. Go Install 지원
```go
// cmd/git-synclone/main.go에 추가
// go install github.com/gizzahub/gzh-manager-go/cmd/git-synclone@latest
```

### 4. 패키지 매니저 통합

#### Homebrew Formula
`homebrew/gzh-git-extensions.rb`:
```ruby
class GzhGitExtensions < Formula
  desc "Git extensions for enhanced repository management"
  homepage "https://github.com/gizzahub/gzh-manager-go"
  url "https://github.com/gizzahub/gzh-manager-go/archive/v1.0.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", "git-synclone", "./cmd/git-synclone"
    bin.install "git-synclone"
  end

  test do
    system "git", "synclone", "--version"
  end
end
```

#### APT Package
`debian/control`:
```
Package: gzh-git-extensions
Version: 1.0.0
Architecture: amd64
Maintainer: Gizzahub Team
Description: Git extensions for repository management
 Provides git synclone command for enhanced repository cloning
Depends: git
```

### 5. 자동 설치 스크립트
`install.sh`:
```bash
#!/bin/bash
# 온라인 설치 스크립트
curl -sSL https://gizzahub.com/install-git-extensions.sh | bash
```

### 6. 설치 확인 및 진단
```bash
# cmd/git-synclone/doctor.go
func doctorCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "doctor",
        Short: "Check git-synclone installation",
        Run: func(cmd *cobra.Command, args []string) {
            checkInstallation()
            checkDependencies()
            checkConfiguration()
        },
    }
}
```

### 7. 언인스톨 지원
`scripts/uninstall-git-extensions.sh`:
```bash
#!/bin/bash
echo "Uninstalling gzh Git extensions..."

# 바이너리 제거
rm -f ~/.local/bin/git-synclone
rm -f /usr/local/bin/git-synclone

echo "Git extensions uninstalled"
```

## 구현 체크리스트
- [x] Makefile에 build-git-extensions 타겟 추가
- [x] 설치 스크립트 작성
- [x] go install 지원 확인
- [x] Homebrew formula 작성
- [x] APT 패키지 설정
- [x] 온라인 설치 스크립트
- [x] doctor 명령어 구현
- [x] 언인스톨 스크립트

## 테스트 요구사항
- [ ] 각 OS별 설치 테스트 (macOS, Linux, Windows)
- [ ] PATH 설정 확인
- [ ] Git 통합 확인 (`git synclone` 명령어 동작)
- [ ] 업그레이드 시나리오 테스트

## 검증 기준
- [ ] `go install` 명령으로 설치 가능
- [ ] 설치 후 `git synclone --help` 동작
- [ ] Homebrew로 설치 가능 (macOS)
- [ ] apt로 설치 가능 (Ubuntu/Debian)
- [ ] 설치 스크립트가 모든 플랫폼에서 동작

## 참고 문서
- Git documentation on custom commands
- Homebrew formula cookbook
- Debian packaging guide

## 완료 후 다음 단계
→ 04-git-synclone-testing.md