# Synclone 테스트 자료

이 디렉토리는 synclone 기능을 완전히 테스트하기 위한 포괄적인 테스트 자료를 포함하고 있습니다.

## 📁 파일 구조

```
specs/synclone/
├── README.md                 # 이 파일
├── test-commands.md          # 모든 synclone 명령어 목록
├── test-scenarios.md         # 상황별 테스트 시나리오
├── test-data.md             # 샘플 데이터 및 리포지터리 목록
├── test-runner.sh           # 자동화된 테스트 실행 스크립트
└── test-configs/            # 테스트용 설정 파일들
    ├── basic-test.yaml      # 기본 기능 테스트
    ├── filtering-test.yaml  # 필터링 기능 테스트
    ├── performance-test.yaml # 성능 테스트
    ├── multi-provider-test.yaml # 다중 제공자 테스트
    ├── enterprise-test.yaml # 엔터프라이즈 테스트
    ├── ci-cd-test.yaml     # CI/CD 테스트
    └── error-handling-test.yaml # 에러 처리 테스트
```

## 🚀 빠른 시작

### 1. 기본 테스트 실행

```bash
# 환경 변수 설정 (선택적)
export GITHUB_TOKEN="your-github-token"

# 기본 테스트 실행
./test-runner.sh basic

# 상세 로그와 함께 실행
./test-runner.sh --verbose basic
```

### 2. 자동화된 전체 테스트

```bash
# 모든 테스트 실행
./test-runner.sh all

# 드라이런으로 테스트 계획 확인
./test-runner.sh --dry-run all

# 병렬 테스트 (빠른 실행)
./test-runner.sh --parallel all
```

### 3. 개별 기능 테스트

```bash
# 필터링 기능 테스트
./test-runner.sh filtering

# 성능 테스트
./test-runner.sh performance

# 에러 처리 테스트
./test-runner.sh error
```

## 📋 테스트 카테고리

### 1. 기본 기능 테스트 (`basic`)
- 소규모 조직 클로닝
- 기본 전략 (reset, pull, fetch)
- 간단한 설정 파일 사용
- **예상 시간**: 2-5분
- **대상 조직**: golangci, spf13

### 2. 필터링 기능 테스트 (`filtering`)
- 언어별 필터링
- 패턴 매칭 (include/exclude)
- 토픽 기반 필터링
- 스타 수 및 크기 제한
- **예상 시간**: 5-10분
- **대상 조직**: kubernetes, prometheus, urfave

### 3. 성능 테스트 (`performance`)
- 병렬 처리 검증
- 대규모 조직 클로닝
- 최적화 기능 테스트
- **예상 시간**: 10-15분
- **대상 조직**: prometheus, grafana, hashicorp

### 4. CI/CD 테스트 (`ci`)
- 빠른 fetch 전략
- 제한된 리소스 환경
- 타임아웃 테스트
- **예상 시간**: 5-10분
- **대상 조직**: prometheus, helm, containerd

### 5. 에러 처리 테스트 (`error`)
- 존재하지 않는 조직 (404 에러)
- 네트워크 타임아웃
- 권한 없는 접근 (403 에러)
- **예상 시간**: 2-5분
- **특수 시나리오**: 의도적 에러 유발

## 🔧 테스트 설정

### 환경 변수
```bash
# 필수 (API 제한 회피용)
export GITHUB_TOKEN="your-github-token"

# 선택적
export GITLAB_TOKEN="your-gitlab-token"
export GITEA_TOKEN="your-gitea-token"

# 바이너리 경로 커스터마이징
export GZ_BINARY="./gz"  # 로컬 빌드된 바이너리 사용
```

### 테스트 커스터마이징
```bash
# 타임아웃 조정
./test-runner.sh --timeout 1200 performance

# 정리 없이 테스트 (결과 확인용)
./test-runner.sh --no-cleanup basic

# 드라이런으로 계획 확인
./test-runner.sh --dry-run all
```

## 📊 결과 확인

### 테스트 결과 위치
```
specs/synclone/test-results/
├── logs/                    # 상세 로그 파일들
├── results.log             # 테스트 결과 요약
└── [테스트별 디렉토리들]    # 실제 클론된 리포지터리들
```

### 수동 결과 확인
```bash
# 클론된 리포지터리 수 확인
find ./test-results -name ".git" -type d | wc -l

# gzh.yaml 파일 확인
find ./test-results -name "gzh.yaml"

# 로그 확인
tail -f ./test-results/logs/test.log
```

## 🎯 테스트 전략

### 단계적 테스트 접근법
1. **기본 테스트** → 핵심 기능 확인
2. **필터링 테스트** → 고급 기능 확인
3. **성능 테스트** → 대규모 환경 확인
4. **에러 테스트** → 예외 상황 확인

### 테스트 데이터 선택 기준
- **소형 조직** (5-20개 리포): 빠른 기능 검증
- **중형 조직** (20-100개 리포): 병렬 처리 검증
- **대형 조직** (100+ 리포): 성능 및 최적화 검증

## 🛠️ 수동 테스트

자동화 스크립트 외에도 개별 명령어로 수동 테스트 가능:

```bash
# 간단한 수동 테스트
gz synclone github -o golangci --target ./manual-test

# 필터링 수동 테스트
gz synclone github -o kubernetes \
  --include "^kubectl.*" \
  --language Go \
  --min-stars 100 \
  --target ./manual-filter-test

# 성능 수동 테스트
gz synclone github -o prometheus \
  --parallel 10 \
  --optimized \
  --cache \
  --target ./manual-perf-test
```

## 🔍 문제 해결

### 일반적인 문제들

#### 1. "gz 바이너리를 찾을 수 없습니다"
```bash
# gz 바이너리 빌드
cd ../../  # gzh-cli 루트로 이동
make build

# PATH에 추가하거나 환경변수 설정
export GZ_BINARY="./gz"
```

#### 2. "GitHub API 연결 실패"
```bash
# 네트워크 연결 확인
curl -I https://api.github.com

# 토큰 설정 확인
echo $GITHUB_TOKEN
```

#### 3. "테스트 타임아웃"
```bash
# 타임아웃 증가
./test-runner.sh --timeout 1800 performance

# 또는 네트워크가 느린 경우 단계적 실행
./test-runner.sh basic
./test-runner.sh filtering
```

#### 4. "디스크 공간 부족"
```bash
# 테스트 후 즉시 정리
./test-runner.sh basic && rm -rf ./test-results/test-*

# 또는 로그만 보존하고 정리
find ./test-results -type d -name "test-*" -exec rm -rf {} +
```

## 📈 성능 참고치

일반적인 성능 참고치 (GitHub API 토큰 사용시):

| 테스트 타입 | 조직 크기 | 예상 시간 | 예상 리포지터리 수 |
|------------|----------|----------|------------------|
| basic      | 소형     | 2-5분    | 10-30개          |
| filtering  | 중형     | 5-10분   | 20-50개          |
| performance| 대형     | 10-15분  | 50-150개         |
| ci         | 선별     | 5-10분   | 30-80개          |
| error      | 혼합     | 2-5분    | 0-20개           |

## 🤝 기여하기

새로운 테스트 시나리오나 설정 파일을 추가하려면:

1. 적절한 카테고리의 설정 파일 생성
2. `test-runner.sh`에 새 테스트 함수 추가
3. `README.md` 업데이트
4. 테스트 실행하여 검증

이 테스트 자료를 통해 synclone의 모든 기능을 체계적으로 검증할 수 있습니다!
