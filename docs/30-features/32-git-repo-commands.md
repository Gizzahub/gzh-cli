# Git Repository Management Guide (`gz git repo`)

Git 호스팅 플랫폼의 리포지터리를 효율적으로 관리하는 통합 CLI 도구입니다.

## 🚀 빠른 참조 (Quick Reference)

```bash
# 가장 자주 사용하는 명령어들
gz git repo clone-or-update <repo-url>                    # 스마트 클론/업데이트
gz git repo pull-all ~/workspace --parallel 5            # 재귀적 일괄 업데이트
gz git repo list --provider github --org myorg           # 리포지터리 목록
gz git repo create --provider github --org myorg --name repo  # 새 리포지터리 생성

# 고급 사용법
gz git repo pull-all --include-pattern ".*api.*" --dry-run    # 패턴 필터링
gz git repo sync --from github:org/repo --to gitlab:org/repo # 플랫폼 간 동기화
```

## 목차

1. [개요](#%EA%B0%9C%EC%9A%94)
1. [빠른 시작](#%EB%B9%A0%EB%A5%B8-%EC%8B%9C%EC%9E%91)
1. [명령어 참조](#%EB%AA%85%EB%A0%B9%EC%96%B4-%EC%B0%B8%EC%A1%B0)
1. [사용 예제](#%EC%82%AC%EC%9A%A9-%EC%98%88%EC%A0%9C)
1. [고급 기능](#%EA%B3%A0%EA%B8%89-%EA%B8%B0%EB%8A%A5)
1. [모범 사례](#%EB%AA%A8%EB%B2%94-%EC%82%AC%EB%A1%80)
1. [문제 해결](#%EB%AC%B8%EC%A0%9C-%ED%95%B4%EA%B2%B0)

## 개요

`gz git repo` 명령어는 다양한 Git 호스팅 플랫폼(GitHub, GitLab, Gitea, Gogs)에서 리포지터리를 관리하는 통합 도구입니다.

### 주요 기능

- **📦 리포지터리 클로닝 및 업데이트**: 단일/대량 클로닝과 스마트 업데이트
- **🔧 리포지터리 관리**: 생성, 삭제, 아카이브, 검색
- **🔄 크로스 플랫폼 동기화**: 플랫폼 간 리포지터리 동기화 및 마이그레이션
- **⚡ 성능 최적화**: 병렬 처리, 재개 기능, 스마트 캐싱
- **🛡️ 안전 기능**: Dry-run, 충돌 감지, 백업 옵션

### 지원 플랫폼

| 플랫폼 | 클론 | 생성 | 삭제 | 동기화 | 상태      |
| ------ | ---- | ---- | ---- | ------ | --------- |
| GitHub | ✅   | ✅   | ✅   | ✅     | 완전 지원 |
| GitLab | ✅   | ✅   | ✅   | ✅     | 완전 지원 |
| Gitea  | ✅   | 🔄   | 🔄   | 🔄     | 개발 중   |
| Gogs   | ✅   | 🔄   | 🔄   | 🔄     | 개발 중   |

## 빠른 시작

### 전제 조건

1. **Git 설치**

   ```bash
   git --version  # 확인
   ```

1. **인증 토큰 설정**

   ```bash
   # GitHub
   export GITHUB_TOKEN="your-github-token"

   # GitLab
   export GITLAB_TOKEN="your-gitlab-token"
   ```

1. **gzh-cli 설치**

   ```bash
   go install github.com/gizzahub/gzh-cli/cmd/gz@latest
   ```

### 기본 사용법 (5분 가이드)

```bash
# 1. 단일 리포지터리 클론
gz git repo clone-or-update https://github.com/user/repo.git

# 2. 특정 전략으로 업데이트
gz git repo clone-or-update --strategy reset https://github.com/user/repo.git

# 3. 조직의 모든 리포지터리 나열
gz git repo list --provider github --org myorg

# 4. 새 리포지터리 생성
gz git repo create --provider github --org myorg --name my-new-repo

# 5. 하위 디렉토리의 모든 Git 리포지터리 일괄 업데이트
gz git repo pull-all --parallel 5 --verbose
```

## 명령어 참조

### 1. `clone` - 대량 리포지터리 클로닝

조직이나 그룹의 여러 리포지터리를 병렬로 클론합니다.

```bash
gz git repo clone [flags]
```

**주요 플래그:**

- `--provider`: Git 플랫폼 (github, gitlab, gitea, gogs)
- `--org`: 조직/그룹 이름
- `--target`: 대상 디렉토리
- `--parallel`: 병렬 워커 수 (기본: 5)
- `--strategy`: 클론 전략 (reset, pull, fetch)
- `--match`: 리포지터리 이름 패턴
- `--resume`: 중단된 작업 재개

**예제:**

```bash
# GitHub 조직의 모든 리포지터리 클론
gz git repo clone --provider github --org myorg --target ./repos

# 패턴 매칭과 병렬 처리
gz git repo clone --provider gitlab --org mygroup --match "api-*" --parallel 10

# 중단된 클론 작업 재개
gz git repo clone --provider github --org myorg --resume
```

### 2. `clone-or-update` - 스마트 단일 리포지터리 관리

단일 리포지터리를 클론하거나 기존 리포지터리를 업데이트합니다.

```bash
gz git repo clone-or-update <repository-url> [target-path] [flags]
```

**전략 옵션:**

- `rebase` (기본): 로컬 변경사항을 원격 변경사항 위에 리베이스
- `reset`: 하드 리셋으로 원격 상태와 일치 (로컬 변경사항 삭제)
- `clone`: 기존 디렉토리 제거 후 새로 클론
- `skip`: 기존 리포지터리 변경하지 않음
- `pull`: 표준 git pull (병합)
- `fetch`: 원격 변경사항만 가져오기

**예제:**

```bash
# 자동 디렉토리 이름으로 클론
gz git repo clone-or-update https://github.com/user/repo.git

# 명시적 경로와 전략 지정
gz git repo clone-or-update --strategy reset https://github.com/user/repo.git ./my-repo

# 특정 브랜치와 얕은 클론
gz git repo clone-or-update --branch develop --depth 1 https://github.com/user/repo.git
```

### 3. `pull-all` - 재귀적 일괄 업데이트 ⭐ NEW

하위 디렉토리의 모든 Git 리포지터리를 안전하게 일괄 업데이트합니다.

```bash
gz git repo pull-all [directory] [flags]
```

**안전 기능:**

- 로컬 변경사항이 없는 경우에만 자동 업데이트
- 충돌 예상 시 수동 처리 알림
- 병합 상태 및 스태시 감지
- 모든 스캔된 리포지터리 결과 표시

**주요 플래그:**

- `--parallel`: 병렬 워커 수 (기본: 5)
- `--max-depth`: 최대 스캔 깊이 (기본: 10)
- `--dry-run`: 시뮬레이션만 실행
- `--json`: JSON 형식 출력
- `--include-pattern`: 포함할 리포지터리 패턴
- `--exclude-pattern`: 제외할 리포지터리 패턴
- `--no-fetch`: 원격 변경사항 가져오지 않음

**예제:**

```bash
# 현재 디렉토리부터 모든 Git 리포지터리 업데이트
gz git repo pull-all

# 특정 디렉토리와 병렬 처리
gz git repo pull-all /home/user/projects --parallel 10 --verbose

# 패턴 필터링
gz git repo pull-all --include-pattern ".*api.*" --exclude-pattern ".*test.*"

# JSON 형식으로 결과 출력
gz git repo pull-all --json > update-results.json
```

### 4. `list` - 리포지터리 목록 조회

고급 필터링과 정렬 옵션으로 리포지터리를 나열합니다.

```bash
gz git repo list [flags]
```

**필터링 옵션:**

- `--provider`: Git 플랫폼
- `--org`: 조직/그룹 이름
- `--visibility`: public, private, internal
- `--language`: 프로그래밍 언어
- `--min-stars`: 최소 스타 수
- `--max-stars`: 최대 스타 수
- `--match`: 이름 패턴
- `--archived-only`: 아카이브된 리포지터리만
- `--no-archived`: 아카이브된 리포지터리 제외

**출력 옵션:**

- `--format`: 출력 형식 (table, json, yaml, csv)
- `--sort`: 정렬 기준 (name, created, updated, stars)
- `--order`: 정렬 순서 (asc, desc)
- `--limit`: 결과 수 제한

**예제:**

```bash
# 기본 리포지터리 목록
gz git repo list --provider github --org myorg

# Go 언어 리포지터리만 필터링
gz git repo list --provider github --org myorg --language go --format json

# 스타 수로 정렬
gz git repo list --provider github --org myorg --sort stars --order desc --limit 10

# CSV 형식으로 내보내기
gz git repo list --provider github --org myorg --format csv > repos.csv
```

### 5. `create` - 리포지터리 생성

다양한 옵션으로 새 리포지터리를 생성합니다.

```bash
gz git repo create [flags]
```

**필수 플래그:**

- `--provider`: Git 플랫폼
- `--org`: 조직/그룹 이름
- `--name`: 리포지터리 이름

**설정 옵션:**

- `--description`: 설명
- `--private`: 비공개 리포지터리
- `--template`: 템플릿 리포지터리
- `--auto-init`: README.md 자동 생성
- `--gitignore-template`: .gitignore 템플릿
- `--license`: 라이선스
- `--default-branch`: 기본 브랜치 이름

**기능 옵션:**

- `--issues`: 이슈 활성화
- `--wiki`: 위키 활성화
- `--projects`: 프로젝트 활성화

**예제:**

```bash
# 기본 공개 리포지터리 생성
gz git repo create --provider github --org myorg --name my-new-repo

# 완전한 설정으로 생성
gz git repo create \
  --provider github --org myorg --name my-api \
  --description "My REST API" --private \
  --auto-init --gitignore-template Go --license MIT \
  --issues --wiki
```

### 6. `delete` - 리포지터리 삭제

안전한 방법으로 리포지터리를 삭제합니다.

```bash
gz git repo delete [flags]
```

**안전 기능:**

- 확인 프롬프트
- Dry-run 옵션
- 패턴 매칭 지원
- 삭제 전 백업 옵션

**예제:**

```bash
# 단일 리포지터리 삭제
gz git repo delete --provider github --org myorg --repo old-project

# 패턴으로 여러 리포지터리 삭제 (주의!)
gz git repo delete --provider github --org myorg --pattern "test-*" --dry-run
```

### 7. `archive` - 리포지터리 아카이브

리포지터리를 아카이브 상태로 변경합니다.

```bash
gz git repo archive [flags]
```

### 8. `sync` - 플랫폼 간 동기화

Git 플랫폼 간 리포지터리를 동기화합니다.

```bash
gz git repo sync [flags]
```

**동기화 옵션:**

- `--from`: 소스 플랫폼 (provider:org/repo)
- `--to`: 대상 플랫폼 (provider:org/repo)
- `--create-missing`: 누락된 리포지터리 생성
- `--include-code`: 코드 동기화
- `--include-issues`: 이슈 동기화
- `--include-wiki`: 위키 동기화
- `--include-releases`: 릴리스 동기화

**예제:**

```bash
# 단일 리포지터리 동기화
gz git repo sync --from github:myorg/repo --to gitlab:mygroup/repo

# 조직 전체 동기화
gz git repo sync --from github:myorg --to gitea:myorg --create-missing

# 특정 기능만 동기화
gz git repo sync --from github:org/repo --to gitlab:group/repo \
  --include-issues --include-wiki --include-releases
```

### 9. `migrate` - 리포지터리 마이그레이션

완전한 플랫폼 마이그레이션을 수행합니다.

```bash
gz git repo migrate [flags]
```

*주의: 현재 개발 중인 기능입니다.*

### 10. `search` - 고급 리포지터리 검색

고급 검색 기능으로 리포지터리를 찾습니다.

```bash
gz git repo search [flags]
```

*주의: 현재 개발 중인 기능입니다.*

## 사용 예제

### 시나리오 1: 새 개발 환경 설정

```bash
# 1. 작업 디렉토리 생성
mkdir ~/workspace && cd ~/workspace

# 2. 주요 프로젝트 클론
gz git repo clone-or-update https://github.com/myorg/main-project.git
gz git repo clone-or-update https://github.com/myorg/api-server.git

# 3. 모든 조직 리포지터리 클론
gz git repo clone --provider github --org myorg --target ./myorg --parallel 8

# 4. 정기적 업데이트 스크립트
gz git repo pull-all ~/workspace --parallel 5 --verbose
```

### 시나리오 2: 코드 리뷰 및 품질 관리

```bash
# 1. 특정 언어 프로젝트만 클론
gz git repo list --provider github --org myorg --language go --format json \
  | jq -r '.[].clone_url' \
  | xargs -I {} gz git repo clone-or-update {}

# 2. 모든 프로젝트 최신 상태로 업데이트
gz git repo pull-all --include-pattern ".*go.*" --verbose

# 3. 결과를 JSON으로 저장하여 분석
gz git repo pull-all --json > update-report.json
```

### 시나리오 3: 플랫폼 마이그레이션

```bash
# 1. 소스 플랫폼 리포지터리 목록 확인
gz git repo list --provider github --org old-org --format json > source-repos.json

# 2. 대상 플랫폼에 조직 생성 후 동기화
gz git repo sync --from github:old-org --to gitlab:new-org --create-missing

# 3. 동기화 결과 확인
gz git repo list --provider gitlab --org new-org --format table
```

## 고급 기능

### 1. 병렬 처리 최적화

```bash
# CPU 코어 수에 따른 최적 워커 수 설정
WORKERS=$(nproc)
gz git repo pull-all --parallel $WORKERS

# 네트워크 대역폭 고려한 조정
gz git repo clone --provider github --org myorg --parallel 3 --strategy fetch
```

### 2. 패턴 기반 필터링

```bash
# 정규식을 이용한 고급 필터링
gz git repo pull-all \
  --include-pattern "^.*(api|service|backend).*$" \
  --exclude-pattern "^.*(test|demo|example).*$"

# 언어별 프로젝트 분리
gz git repo list --provider github --org myorg --language go | \
  jq -r '.[].name' | \
  xargs -I {} gz git repo clone-or-update https://github.com/myorg/{}.git ./go-projects/{}
```

### 3. 자동화 및 스크립팅

```bash
#!/bin/bash
# daily-update.sh - 일일 업데이트 스크립트

# 업데이트 실행
gz git repo pull-all ~/workspace --json > /tmp/update-$(date +%Y%m%d).json

# 실패한 리포지터리 추출
jq -r '.[] | select(.status == "error") | .path' /tmp/update-$(date +%Y%m%d).json > /tmp/failed-repos.txt

# 알림 전송 (Slack, 이메일 등)
if [ -s /tmp/failed-repos.txt ]; then
    echo "Failed repositories:" $(cat /tmp/failed-repos.txt)
fi
```

### 4. 설정 파일 기반 관리

```yaml
# ~/.config/gzh/repo-config.yaml
default:
  parallel: 5
  strategy: rebase

profiles:
  production:
    strategy: reset
    parallel: 2

  development:
    strategy: rebase
    parallel: 10
    include_pattern: ".*dev.*"
```

## 모범 사례

### 1. 안전한 리포지터리 관리

```bash
# 항상 dry-run으로 먼저 테스트
gz git repo pull-all --dry-run

# 중요한 작업 전 백업
gz git repo list --provider github --org myorg --format json > backup-$(date +%Y%m%d).json

# 단계적 업데이트 (소규모 그룹부터)
gz git repo pull-all ./critical-projects --parallel 2 --verbose
gz git repo pull-all ./dev-projects --parallel 8 --verbose
```

### 2. 성능 최적화

```bash
# 네트워크 제한 환경에서 fetch 전용 사용
gz git repo pull-all --no-fetch --strategy fetch

# 대용량 리포지터리 shallow clone
gz git repo clone-or-update --depth 1 https://github.com/large/repo.git

# 점진적 병렬 처리 증가
for workers in 2 4 8; do
    echo "Testing with $workers workers"
    time gz git repo pull-all --parallel $workers --dry-run
done
```

### 3. 모니터링 및 로깅

```bash
# 상세 로깅으로 문제 진단
gz git repo pull-all --verbose 2>&1 | tee update.log

# JSON 출력으로 분석 데이터 수집
gz git repo pull-all --json | jq -r '.[] | "\(.path): \(.status)"'

# 성능 메트릭 수집
time gz git repo pull-all --json > results.json
```

## 문제 해결

### 일반적인 문제

#### 1. 인증 실패

```bash
# 토큰 확인
echo $GITHUB_TOKEN

# 토큰 권한 확인 (repo, admin:org 필요)
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# SSH 키 설정 확인
ssh -T git@github.com
```

#### 2. 네트워크 타임아웃

```bash
# Git 전역 타임아웃 설정
git config --global http.timeout 300

# 병렬 워커 수 감소
gz git repo pull-all --parallel 2

# 재시도 메커니즘 사용
for i in {1..3}; do
    gz git repo pull-all && break
    echo "Retry $i failed, waiting..."
    sleep 10
done
```

#### 3. 디스크 공간 부족

```bash
# 얕은 클론 사용
gz git repo clone-or-update --depth 1 <repo-url>

# 불필요한 히스토리 정리
find . -name ".git" -type d -exec git -C {} gc --aggressive \;

# LFS 파일 정리
find . -name ".git" -type d -exec git -C {} lfs prune \;
```

#### 4. 충돌 해결

```bash
# 충돌이 있는 리포지터리 식별
gz git repo pull-all --json | jq -r '.[] | select(.status == "conflicts") | .path'

# 수동 해결 후 재시도
cd conflicted-repo
git status
git add .
git rebase --continue
cd ..
gz git repo pull-all ./conflicted-repo
```

### 디버깅 도구

```bash
# 상세 Git 로그 활성화
export GIT_TRACE=1
export GIT_CURL_VERBOSE=1

# gzh-cli 디버그 모드
gz --debug git repo pull-all

# 특정 리포지터리 문제 진단
gz git repo clone-or-update --verbose https://problematic-repo.git
```

### 성능 튜닝

```bash
# 시스템 리소스 모니터링
top -p $(pgrep -f "gz git repo")

# 네트워크 사용량 확인
nethogs

# 디스크 I/O 모니터링
iotop -o
```

## 관련 문서

- [Git Repository Configuration Management](./31-repository-management.md)
- [Synclone User Guide](./30-synclone.md)
- Authentication Setup Guide
- Performance Optimization
- [Git Repo Examples Configuration](../../examples/git-repo-examples.yaml)

## 지원 및 피드백

- **Issues**: [GitHub Issues](https://github.com/gizzahub/gzh-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gizzahub/gzh-cli/discussions)
- **Documentation**: [프로젝트 위키](https://github.com/gizzahub/gzh-cli/wiki)

______________________________________________________________________

*마지막 업데이트: 2025년 1월*
