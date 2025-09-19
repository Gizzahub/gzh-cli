# Synclone Test Commands

이 문서는 synclone 기능을 완전히 테스트하기 위한 모든 명령어 조합을 정리한 것입니다.

## 기본 명령어 테스트

### 1. 메인 synclone 명령어

```bash
# 도움말 확인
gz synclone --help
gz synclone -h

# 버전 확인
gz --version
```

### 2. GitHub 조직 클로닝

#### 기본 GitHub 클로닝

```bash
# 기본 클로닝 (현재 디렉토리)
gz synclone github -o kubernetes

# 조직명 축약 플래그 사용
gz synclone github --org kubernetes

# 특정 디렉토리로 클로닝
gz synclone github -o kubernetes -t ~/test-repos/k8s

# 풀 타겟 패스 플래그
gz synclone github -o kubernetes --targetPath ~/test-repos/k8s

# 단축 플래그 조합
gz synclone github -o kubernetes --target ~/test-repos/k8s
```

#### 전략 옵션

```bash
# Reset 전략 (기본값)
gz synclone github -o kubernetes --strategy reset

# Pull 전략 (기존 변경사항 유지 시도)
gz synclone github -o kubernetes --strategy pull

# Fetch 전략 (참조만 업데이트)
gz synclone github -o kubernetes --strategy fetch
```

#### 병렬 처리 옵션

```bash
# 병렬 처리 워커 수 지정
gz synclone github -o kubernetes --parallel 5
gz synclone github -o kubernetes -p 10

# 최대 재시도 횟수
gz synclone github -o kubernetes --max-retries 5

# 재개 기능
gz synclone github -o kubernetes --resume
```

#### 최적화 옵션

```bash
# 최적화된 스트리밍 모드
gz synclone github -o kubernetes --optimized

# 스트리밍 모드
gz synclone github -o kubernetes --streaming

# 메모리 제한
gz synclone github -o kubernetes --memory-limit 1GB

# 토큰 지정
gz synclone github -o kubernetes --token $GITHUB_TOKEN
```

#### 캐싱 옵션

```bash
# 로컬 캐시 활성화
gz synclone github -o kubernetes --cache

# Redis 캐시 활성화 (Redis 서버 필요)
gz synclone github -o kubernetes --redis --redis-addr localhost:6379
```

#### 필터링 옵션

```bash
# 패턴 매칭
gz synclone github -o kubernetes --include "^kubectl.*"
gz synclone github -o kubernetes --exclude ".*-archived$"

# 토픽 필터링
gz synclone github -o kubernetes --topics "go,cli"
gz synclone github -o kubernetes --exclude-topics "deprecated,archived"

# 언어 필터링
gz synclone github -o kubernetes --language Go

# 스타 수 필터링
gz synclone github -o kubernetes --min-stars 100
gz synclone github -o kubernetes --max-stars 1000

# 업데이트 날짜 필터링
gz synclone github -o kubernetes --updated-after 2024-01-01
gz synclone github -o kubernetes --updated-before 2024-12-31
```

#### 리포지터리 타입 필터링

```bash
# 아카이브된 리포지터리 포함
gz synclone github -o kubernetes --include-archived

# 포크 리포지터리 포함
gz synclone github -o kubernetes --include-forks

# 프라이빗 리포지터리 포함 (토큰 필요)
gz synclone github -o kubernetes --include-private

# 빈 리포지터리만
gz synclone github -o kubernetes --only-empty

# 크기 제한 (KB 단위)
gz synclone github -o kubernetes --size-limit 10000
```

#### 정리 옵션

```bash
# 고아 디렉토리 정리
gz synclone github -o kubernetes --cleanup-orphans
```

### 3. GitLab 그룹 클로닝

```bash
# 기본 GitLab 그룹 클로닝
gz synclone gitlab -g mygroup

# 특정 디렉토리로 클로닝
gz synclone gitlab -g mygroup -t ~/gitlab-repos

# 하위 그룹 포함
gz synclone gitlab -g mygroup --recursive

# 전략 지정
gz synclone gitlab -g mygroup --strategy pull
```

### 4. Gitea 조직 클로닝

```bash
# 기본 Gitea 조직 클로닝
gz synclone gitea -o myorg

# 커스텀 Gitea 인스턴스
gz synclone gitea -o myorg --api-url https://gitea.example.com

# 특정 디렉토리로 클로닝
gz synclone gitea -o myorg -t ~/gitea-repos
```

### 5. 설정 파일 사용

```bash
# 설정 파일 지정
gz synclone --config synclone.yaml

# 표준 위치의 설정 파일 사용
gz synclone --use-config

# gzh.yaml 포맷 사용
gz synclone --use-gzh-config

# 프로바이더 필터링
gz synclone --config synclone.yaml --provider github

# 전략 오버라이드
gz synclone --config synclone.yaml --strategy pull

# 병렬 처리 오버라이드
gz synclone --config synclone.yaml --parallel 15
```

## 설정 관리 명령어

### 1. 설정 생성

```bash
# 초기 설정 생성
gz synclone config generate init

# 템플릿에서 생성
gz synclone config generate template --template enterprise
gz synclone config generate template --template simple

# 기존 리포지터리에서 발견
gz synclone config generate discover --path ~/repos

# GitHub 전용 설정 생성
gz synclone config generate github --org mycompany

# GitLab 전용 설정 생성
gz synclone config generate gitlab --group mygroup
```

### 2. 설정 검증

```bash
# 기본 검증
gz synclone config validate

# 특정 파일 검증
gz synclone config validate --config synclone.yaml

# 엄격한 검증 (스키마 포함)
gz synclone config validate --strict

# 토큰 유효성 검사 포함
gz synclone validate --check-tokens --config synclone.yaml
```

### 3. 설정 변환

```bash
# YAML to JSON 변환
gz synclone config convert --from synclone.yaml --to synclone.json

# gzh 포맷으로 변환
gz synclone config convert --from synclone.yaml --format gzh

# 포맷 자동 감지
gz synclone config convert synclone.yaml synclone.json
```

## 상태 관리 명령어

### 1. 상태 목록 조회

```bash
# 모든 작업 목록
gz synclone state list

# 활성 작업만
gz synclone state list --active

# 실패한 작업만
gz synclone state list --failed

# 완료된 작업만
gz synclone state list --completed

# 날짜별 필터링
gz synclone state list --after 2024-01-01
gz synclone state list --before 2024-12-31
```

### 2. 상태 상세 조회

```bash
# ID로 조회
gz synclone state show <state-id>

# 마지막 작업 조회
gz synclone state show --last

# 실패한 작업의 상세 로그
gz synclone state show --last --logs

# JSON 포맷으로 조회
gz synclone state show --last --format json
```

### 3. 상태 정리

```bash
# 오래된 작업 정리 (7일 기준)
gz synclone state clean --age 7d

# 실패한 작업 정리
gz synclone state clean --failed

# 특정 작업 정리
gz synclone state clean --id <state-id>

# 모든 작업 정리 (확인 필요)
gz synclone state clean --all --force
```

## 검증 명령어

```bash
# 기본 검증
gz synclone validate

# 설정 파일 검증
gz synclone validate --config synclone.yaml

# 토큰 유효성 검사
gz synclone validate --check-tokens

# 네트워크 연결 확인
gz synclone validate --check-network

# 모든 검증 수행
gz synclone validate --all
```

## 조합 명령어 예제

### 1. 프로덕션 사용 시나리오

```bash
# 대규모 조직의 안전한 클로닝
gz synclone github -o kubernetes \
  --target ~/work/k8s \
  --strategy reset \
  --parallel 10 \
  --max-retries 3 \
  --optimized \
  --cache \
  --cleanup-orphans

# 필터링된 클로닝
gz synclone github -o kubernetes \
  --include "^kubectl.*|^kube-.*" \
  --exclude ".*-deprecated$|.*-archive$" \
  --language Go \
  --min-stars 50 \
  --updated-after 2023-01-01
```

### 2. 개발 환경 시나리오

```bash
# 빠른 개발용 클로닝
gz synclone github -o mycompany \
  --target ~/dev \
  --strategy pull \
  --parallel 5 \
  --include-private \
  --token $GITHUB_TOKEN

# 테스트용 클로닝
gz synclone github -o testorg \
  --target ./test-repos \
  --strategy fetch \
  --only-empty \
  --max-retries 1
```

### 3. 설정 기반 시나리오

```bash
# 엔터프라이즈 설정으로 클로닝
gz synclone --config enterprise.yaml \
  --parallel 20 \
  --resume

# 여러 프로바이더 동시 클로닝
gz synclone --config multi-provider.yaml \
  --provider github,gitlab \
  --cleanup-orphans
```

## 디버그 및 문제해결 명령어

```bash
# 상세 로그와 함께 실행
gz synclone github -o testorg --verbose

# 드라이런 모드 (아직 구현되지 않음)
gz synclone github -o testorg --dry-run

# 네트워크 이슈 대응
gz synclone github -o testorg \
  --max-retries 5 \
  --timeout 300s \
  --resume

# 메모리 제한 환경
gz synclone github -o testorg \
  --memory-limit 256MB \
  --parallel 3 \
  --streaming
```

## 환경 변수 조합

```bash
# 토큰 설정
export GITHUB_TOKEN="your-token"
export GITLAB_TOKEN="your-gitlab-token"
export GITEA_TOKEN="your-gitea-token"

# 설정 파일 경로
export GZH_SYNCLONE_CONFIG="~/my-synclone.yaml"

# 디버그 모드
export GZH_DEBUG=true

# 실행 후 명령어
gz synclone github -o kubernetes
```

이 명령어들은 synclone의 모든 기능을 체계적으로 테스트하기 위한 완전한 명령어 세트입니다.
