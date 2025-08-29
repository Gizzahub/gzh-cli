# Synclone Test Scenarios

이 문서는 synclone 기능을 다양한 실제 상황에서 테스트하기 위한 시나리오별 명령어를 정리한 것입니다.

## 1. 기본 기능 테스트 시나리오

### 1.1 첫 설치 및 기본 사용

#### 시나리오: 새 사용자가 처음으로 synclone을 사용하는 경우

```bash
# Step 1: 도움말 확인
gz synclone --help
gz synclone github --help

# Step 2: 간단한 공개 조직 클로닝 (토큰 없음)
gz synclone github -o golangci --target ./test-golangci
echo "Expected: 소수의 공개 리포지터리만 클로닝됨"

# Step 3: 결과 확인
ls -la ./test-golangci/
ls ./test-golangci/ | wc -l
echo "Expected: gzh.yaml 파일이 생성되어야 함"

# Step 4: 정리
rm -rf ./test-golangci/
```

### 1.2 토큰 기반 인증 테스트

#### 시나리오: GitHub 토큰을 사용한 프라이빗 리포지터리 접근

```bash
# Step 1: 토큰 없이 프라이빗 조직 시도 (실패해야 함)
gz synclone github -o your-private-org --target ./test-private-fail
echo "Expected: 인증 오류 발생"

# Step 2: 토큰 설정 후 재시도
export GITHUB_TOKEN="your-github-token"
gz synclone github -o your-private-org --target ./test-private-success --include-private
echo "Expected: 프라이빗 리포지터리도 클로닝됨"

# Step 3: 토큰 파라미터로 직접 전달
gz synclone github -o your-private-org --token $GITHUB_TOKEN --target ./test-token-param --include-private

# Step 4: 정리
rm -rf ./test-private-*
```

## 2. 전략별 테스트 시나리오

### 2.1 Reset 전략 테스트

#### 시나리오: 로컬 변경사항이 있는 리포지터리에서 reset 전략 사용

```bash
# Step 1: 초기 클로닝
gz synclone github -o prometheus --target ./test-reset --strategy reset

# Step 2: 일부 파일 수정
cd ./test-reset
# 첫 번째 리포지터리 찾아서 수정
FIRST_REPO=$(ls | head -1)
echo "# Local change" >> "$FIRST_REPO/README.md"
cd ..

# Step 3: Reset 전략으로 재동기화
gz synclone github -o prometheus --target ./test-reset --strategy reset
echo "Expected: 로컬 변경사항이 모두 사라져야 함"

# Step 4: 확인
cd ./test-reset/$FIRST_REPO
tail README.md
echo "Expected: '# Local change' 가 없어야 함"
cd ../..

# Step 5: 정리
rm -rf ./test-reset
```

### 2.2 Pull 전략 테스트

#### 시나리오: 로컬 변경사항을 보존하면서 업데이트

```bash
# Step 1: 초기 클로닝
gz synclone github -o containerd --target ./test-pull --strategy pull

# Step 2: 로컬 브랜치 생성 및 커밋
cd ./test-pull
FIRST_REPO=$(ls | head -1)
cd "$FIRST_REPO"
git checkout -b local-changes
echo "# Local feature" >> LOCAL_FEATURE.md
git add LOCAL_FEATURE.md
git commit -m "Add local feature"
cd ../..

# Step 3: Pull 전략으로 재동기화
gz synclone github -o containerd --target ./test-pull --strategy pull
echo "Expected: 로컬 변경사항이 보존되어야 함"

# Step 4: 확인
cd ./test-pull/$FIRST_REPO
ls LOCAL_FEATURE.md
git log --oneline | head -3
cd ../..

# Step 5: 정리
rm -rf ./test-pull
```

### 2.3 Fetch 전략 테스트

#### 시나리오: 참조만 업데이트하고 작업 디렉토리는 건드리지 않음

```bash
# Step 1: 초기 클로닝
gz synclone github -o etcd-io --target ./test-fetch --strategy fetch

# Step 2: 워킹 디렉토리 상태 확인
cd ./test-fetch
FIRST_REPO=$(ls | head -1)
cd "$FIRST_REPO"
INITIAL_COMMIT=$(git rev-parse HEAD)
echo "Initial HEAD: $INITIAL_COMMIT"
cd ../..

# Step 3: Fetch 전략으로 재동기화
gz synclone github -o etcd-io --target ./test-fetch --strategy fetch
echo "Expected: 참조는 업데이트되지만 HEAD는 동일해야 함"

# Step 4: 확인
cd ./test-fetch/$FIRST_REPO
AFTER_COMMIT=$(git rev-parse HEAD)
echo "After HEAD: $AFTER_COMMIT"
echo "Remote refs updated but working directory unchanged"
cd ../..

# Step 5: 정리
rm -rf ./test-fetch
```

## 3. 병렬 처리 테스트 시나리오

### 3.1 기본 병렬 처리 테스트

#### 시나리오: 다양한 병렬 처리 설정으로 성능 비교

```bash
# Step 1: 순차 처리 (parallel=1)
time gz synclone github -o cncf --target ./test-sequential --parallel 1
SEQUENTIAL_TIME=$?

# Step 2: 중간 병렬 처리 (parallel=5)
rm -rf ./test-sequential
time gz synclone github -o cncf --target ./test-parallel-5 --parallel 5
PARALLEL_5_TIME=$?

# Step 3: 높은 병렬 처리 (parallel=10)
rm -rf ./test-parallel-5
time gz synclone github -o cncf --target ./test-parallel-10 --parallel 10
PARALLEL_10_TIME=$?

echo "Performance comparison:"
echo "Sequential (1): $SEQUENTIAL_TIME"
echo "Parallel (5): $PARALLEL_5_TIME"
echo "Parallel (10): $PARALLEL_10_TIME"

# Step 4: 정리
rm -rf ./test-parallel-*
```

### 3.2 재시도 및 복구 테스트

#### 시나리오: 네트워크 오류 상황에서 재시도 기능 테스트

```bash
# Step 1: 낮은 재시도 설정으로 테스트 (빠른 실패)
gz synclone github -o kubernetes --target ./test-retry-low --max-retries 1 --parallel 3
echo "Expected: 일부 리포지터리 실패 가능"

# Step 2: 높은 재시도 설정으로 테스트 (안정적)
gz synclone github -o kubernetes --target ./test-retry-high --max-retries 5 --parallel 3
echo "Expected: 대부분 리포지터리 성공"

# Step 3: Resume 기능 테스트
gz synclone github -o kubernetes --target ./test-resume --resume
echo "Expected: 실패한 리포지터리만 재시도"

# Step 4: 정리
rm -rf ./test-retry-* ./test-resume
```

## 4. 필터링 테스트 시나리오

### 4.1 패턴 매칭 테스트

#### 시나리오: 특정 패턴의 리포지터리만 클로닝

```bash
# Step 1: kubectl 관련 리포지터리만 클로닝
gz synclone github -o kubernetes --target ./test-kubectl --include "^kubectl.*"
echo "kubectl 관련 리포지터리만 클로닝되어야 함"
ls ./test-kubectl/ | grep kubectl

# Step 2: 아카이브된 리포지터리 제외
gz synclone github -o kubernetes --target ./test-no-archive --exclude ".*-archive$|.*-deprecated$"
echo "아카이브된 리포지터리가 제외되어야 함"

# Step 3: 여러 패턴 조합
gz synclone github -o kubernetes \
  --target ./test-combined \
  --include "^kube.*|^kubectl.*" \
  --exclude ".*-test$|.*-example$"
echo "kube/kubectl로 시작하지만 test/example로 끝나지 않는 리포지터리"

# Step 4: 정리
rm -rf ./test-kubectl ./test-no-archive ./test-combined
```

### 4.2 언어 및 토픽 필터링

#### 시나리오: 특정 언어나 토픽의 리포지터리만 클로닝

```bash
# Step 1: Go 언어 리포지터리만
gz synclone github -o hashicorp --target ./test-go-only --language Go
echo "Go 언어 리포지터리만 클로닝되어야 함"

# Step 2: 특정 토픽 포함
gz synclone github -o kubernetes --target ./test-cli-topic --topics "cli,tools"
echo "CLI나 tools 토픽이 있는 리포지터리만"

# Step 3: 특정 토픽 제외
gz synclone github -o kubernetes --target ./test-no-docs --exclude-topics "documentation,tutorial"
echo "문서나 튜토리얼 토픽이 없는 리포지터리만"

# Step 4: 정리
rm -rf ./test-go-only ./test-cli-topic ./test-no-docs
```

### 4.3 스타 수 및 크기 필터링

#### 시나리오: 인기도나 크기 기준으로 리포지터리 필터링

```bash
# Step 1: 최소 스타 수 필터링
gz synclone github -o kubernetes --target ./test-popular --min-stars 1000
echo "1000개 이상의 스타를 가진 인기 리포지터리만"
ls ./test-popular/ | wc -l

# Step 2: 스타 수 범위 제한
gz synclone github -o prometheus --target ./test-mid-popular --min-stars 100 --max-stars 1000
echo "100-1000 스타 범위의 리포지터리만"

# Step 3: 크기 제한
gz synclone github -o kubernetes --target ./test-small --size-limit 10000
echo "10MB 미만의 작은 리포지터리만"

# Step 4: 정리
rm -rf ./test-popular ./test-mid-popular ./test-small
```

## 5. 최적화 기능 테스트 시나리오

### 5.1 스트리밍 및 캐싱 테스트

#### 시나리오: 대규모 조직에서 최적화 기능 비교

```bash
# Step 1: 기본 모드 (성능 기준선)
time gz synclone github -o kubernetes --target ./test-basic
BASIC_TIME=$?

# Step 2: 최적화 모드
rm -rf ./test-basic
time gz synclone github -o kubernetes --target ./test-optimized --optimized
OPTIMIZED_TIME=$?

# Step 3: 스트리밍 모드
rm -rf ./test-optimized
time gz synclone github -o kubernetes --target ./test-streaming --streaming --memory-limit 512MB
STREAMING_TIME=$?

# Step 4: 캐싱 모드
rm -rf ./test-streaming
time gz synclone github -o kubernetes --target ./test-cached --cache
CACHED_TIME=$?

echo "Performance comparison:"
echo "Basic: $BASIC_TIME"
echo "Optimized: $OPTIMIZED_TIME"
echo "Streaming: $STREAMING_TIME"
echo "Cached: $CACHED_TIME"

# Step 5: 정리
rm -rf ./test-*
```

### 5.2 메모리 제한 테스트

#### 시나리오: 제한된 메모리 환경에서 동작 확인

```bash
# Step 1: 낮은 메모리 제한
gz synclone github -o kubernetes \
  --target ./test-low-memory \
  --memory-limit 128MB \
  --parallel 2 \
  --streaming
echo "저메모리 환경에서 정상 동작해야 함"

# Step 2: 중간 메모리 제한
gz synclone github -o kubernetes \
  --target ./test-mid-memory \
  --memory-limit 512MB \
  --parallel 5

# Step 3: 메모리 제한 없음 (기본)
gz synclone github -o kubernetes \
  --target ./test-no-limit \
  --parallel 10

# Step 4: 정리
rm -rf ./test-*memory ./test-no-limit
```

## 6. 설정 파일 테스트 시나리오

### 6.1 다중 조직 설정 테스트

#### 시나리오: 여러 조직을 하나의 설정으로 관리

```bash
# Step 1: 다중 조직 설정 파일 생성
cat > test-multi-org.yaml << 'YAML'
version: "1.0"
target: "./multi-org-test"
strategy: "reset"
parallel: 5

github:
  enabled: true
  organizations:
    - name: "prometheus"
      target: "./multi-org-test/prometheus"
      strategy: "pull"
    - name: "grafana"
      target: "./multi-org-test/grafana"
      include_archived: false
    - name: "jaegertracing"
      target: "./multi-org-test/jaeger"
      filters:
        languages: ["Go", "JavaScript"]
YAML

# Step 2: 설정 파일 검증
gz synclone config validate --config test-multi-org.yaml
echo "Expected: 설정 파일이 유효해야 함"

# Step 3: 설정 기반 클로닝 실행
gz synclone --config test-multi-org.yaml
echo "Expected: 3개 조직이 각각 다른 전략으로 클로닝됨"

# Step 4: 결과 확인
ls -la ./multi-org-test/
ls ./multi-org-test/prometheus/ | wc -l
ls ./multi-org-test/grafana/ | wc -l
ls ./multi-org-test/jaeger/ | wc -l

# Step 5: 정리
rm -rf ./multi-org-test test-multi-org.yaml
```

### 6.2 설정 변환 테스트

#### 시나리오: 다양한 설정 포맷 간 변환

```bash
# Step 1: 기본 YAML 설정 생성
gz synclone config generate init
echo "Expected: synclone.yaml 파일 생성"

# Step 2: YAML to JSON 변환
gz synclone config convert --from synclone.yaml --to synclone.json
echo "Expected: synclone.json 파일 생성"

# Step 3: JSON 설정으로 검증
gz synclone config validate --config synclone.json
echo "Expected: JSON 설정도 유효해야 함"

# Step 4: gzh 포맷으로 변환
gz synclone config convert --from synclone.yaml --format gzh
echo "Expected: gzh.yaml 포맷으로 변환"

# Step 5: 정리
rm -f synclone.yaml synclone.json gzh.yaml
```

## 7. 에러 처리 및 복구 테스트 시나리오

### 7.1 네트워크 오류 시나리오

#### 시나리오: 네트워크 중단 및 복구 상황

```bash
# Step 1: 높은 병렬성으로 시작 (네트워크 부하 유발)
gz synclone github -o kubernetes --target ./test-network --parallel 20 &
SYNCLONE_PID=$!

# Step 2: 잠시 후 중단 (Ctrl+C 시뮬레이션)
sleep 10
kill -INT $SYNCLONE_PID
echo "네트워크 오류로 중단된 상황 시뮬레이션"

# Step 3: Resume으로 복구
gz synclone github -o kubernetes --target ./test-network --resume
echo "Expected: 중단된 지점부터 재개되어야 함"

# Step 4: 정리
rm -rf ./test-network
```

### 7.2 권한 오류 시나리오

#### 시나리오: 잘못된 토큰이나 권한 없는 조직 접근

```bash
# Step 1: 잘못된 토큰으로 프라이빗 접근 시도
export GITHUB_TOKEN="invalid-token"
gz synclone github -o your-private-org --target ./test-invalid-token --include-private
echo "Expected: 인증 오류 발생"

# Step 2: 존재하지 않는 조직
gz synclone github -o nonexistent-org-12345 --target ./test-nonexistent
echo "Expected: 조직을 찾을 수 없다는 오류"

# Step 3: 권한 없는 프라이빗 조직
export GITHUB_TOKEN="valid-but-limited-token"
gz synclone github -o super-secret-org --target ./test-no-access --include-private
echo "Expected: 접근 권한 없다는 오류"

# Step 4: 토큰 복원
unset GITHUB_TOKEN

# Step 5: 정리
rm -rf ./test-invalid-token ./test-nonexistent ./test-no-access
```

## 8. 상태 관리 테스트 시나리오

### 8.1 상태 추적 및 관리

#### 시나리오: 작업 상태 추적 및 관리 기능

```bash
# Step 1: 작업 시작 전 상태 확인
gz synclone state list
echo "Expected: 비어있거나 이전 작업들이 표시됨"

# Step 2: 새 작업 시작
gz synclone github -o prometheus --target ./test-state --parallel 3

# Step 3: 작업 후 상태 확인
gz synclone state list
echo "Expected: 새로운 작업이 completed 상태로 표시됨"

# Step 4: 마지막 작업 상세 조회
gz synclone state show --last
echo "Expected: 작업 상세 정보 표시"

# Step 5: 오래된 상태 정리
gz synclone state clean --age 1d
echo "Expected: 1일 이상된 상태 파일 정리"

# Step 6: 정리
rm -rf ./test-state
```

## 9. 통합 시나리오

### 9.1 프로덕션 환경 시뮬레이션

#### 시나리오: 실제 프로덕션 환경에서의 사용 패턴

```bash
# Step 1: 환경 설정
export GITHUB_TOKEN="your-production-token"
export GZH_SYNCLONE_CONFIG="production-synclone.yaml"

# Step 2: 프로덕션 설정 파일 생성
cat > production-synclone.yaml << 'YAML'
version: "1.0"
target: "./production-repos"
strategy: "reset"
parallel: 15

github:
  enabled: true
  organizations:
    - name: "kubernetes"
      target: "./production-repos/k8s"
      strategy: "reset"
      filters:
        languages: ["Go"]
        min_stars: 100
        exclude_topics: ["archived", "deprecated"]
    - name: "prometheus"
      target: "./production-repos/monitoring"
      strategy: "pull"
      include_archived: false
    - name: "grafana"
      target: "./production-repos/grafana"
      filters:
        updated_after: "2023-01-01"
YAML

# Step 3: 설정 검증
gz synclone config validate --config production-synclone.yaml --strict

# Step 4: 토큰 유효성 검사
gz synclone validate --check-tokens --config production-synclone.yaml

# Step 5: 실제 클로닝 실행
time gz synclone --config production-synclone.yaml \
  --max-retries 5 \
  --cache \
  --cleanup-orphans

# Step 6: 결과 검증
echo "=== 클로닝 결과 ==="
find ./production-repos -name "gzh.yaml" | wc -l
echo "Expected: 3개의 gzh.yaml 파일 (각 조직별)"

find ./production-repos -type d -name ".git" | wc -l
echo "git 리포지터리 개수"

# Step 7: 상태 확인
gz synclone state show --last --format json > last-operation.json
echo "마지막 작업 결과를 JSON으로 저장"

# Step 8: 정리
rm -rf ./production-repos production-synclone.yaml last-operation.json
```

### 9.2 CI/CD 환경 시뮬레이션

#### 시나리오: 자동화된 CI/CD 파이프라인에서의 사용

```bash
# Step 1: CI 환경 변수 시뮬레이션
export CI=true
export GITHUB_TOKEN="ci-token"
export BUILD_NUMBER="123"

# Step 2: CI용 설정 (빠른 실행)
cat > ci-synclone.yaml << 'YAML'
version: "1.0"
target: "./ci-repos"
strategy: "fetch"  # CI에서는 fetch만 수행
parallel: 10

github:
  enabled: true
  organizations:
    - name: "your-org"
      target: "./ci-repos/main"
      filters:
        languages: ["Go", "JavaScript", "Python"]
        updated_after: "2024-01-01"
YAML

# Step 3: 검증 (실패시 CI 중단)
gz synclone config validate --config ci-synclone.yaml --strict || exit 1

# Step 4: 빠른 클로닝 (fetch 전략)
timeout 600 gz synclone --config ci-synclone.yaml \
  --max-retries 3 \
  --memory-limit 1GB \
  --streaming

# Step 5: CI 결과 검증
if [ $? -eq 0 ]; then
  echo "✅ CI synclone completed successfully"
  echo "BUILD_NUMBER: $BUILD_NUMBER" > ./ci-repos/build-info.txt
else
  echo "❌ CI synclone failed"
  exit 1
fi

# Step 6: 아티팩트 준비 (CI에서 저장할 정보)
gz synclone state show --last > ci-synclone-report.txt

# Step 7: 정리
rm -rf ./ci-repos ci-synclone.yaml ci-synclone-report.txt
```

이러한 테스트 시나리오들은 synclone의 모든 기능을 다양한 실제 상황에서 체계적으로 검증하기 위한 것입니다.
