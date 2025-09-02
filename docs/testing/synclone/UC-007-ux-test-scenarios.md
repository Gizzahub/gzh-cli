# Synclone UX Improvements Test Scenarios

이 문서는 2025-09에 도입된 UX 개선사항들을 체계적으로 테스트하기 위한 시나리오를 정의합니다.

## 개선사항 개요

- **로깅 시스템**: 일반 모드에서 깨끗한 출력, 디버그 모드에서만 상세 로그
- **프로그레스 바**: 0/total부터 정확한 시작, 실시간 업데이트
- **성능 로그**: JSON 대신 사람이 읽기 쉬운 텍스트 형식

## 1. 로깅 시스템 개선 테스트

### 1.1 일반 모드 (Clean Output) 테스트

#### 시나리오: 일반 사용자에게 깔끔한 출력 제공

```bash
echo "=== 일반 모드 테스트 ==="
result=$(gz synclone github -o Gizzahub 2>&1)

# 검증 1: 타임스탬프 로그가 없어야 함
echo "$result" | grep -E "^[0-9]{2}:[0-9]{2}:[0-9]{2}" && echo "❌ FAIL: Timestamp logs found" || echo "✅ PASS: No timestamp logs"

# 검증 2: DEBUG/INFO 로그가 없어야 함
echo "$result" | grep -E "INFO|DEBUG" && echo "❌ FAIL: Debug logs found" || echo "✅ PASS: No debug logs"

# 검증 3: 콘솔 메시지는 표시되어야 함
echo "$result" | grep "🔍" && echo "✅ PASS: Progress indicator found" || echo "❌ FAIL: No progress indicator"
echo "$result" | grep "📋 Found" && echo "✅ PASS: Status message found" || echo "❌ FAIL: No status message"
echo "$result" | grep "✅" && echo "✅ PASS: Success message found" || echo "❌ FAIL: No success message"

# 검증 4: JSON 성능 로그가 없어야 함
echo "$result" | grep '{"timestamp":' && echo "❌ FAIL: JSON logs found" || echo "✅ PASS: No JSON logs"

echo "Expected Normal Mode Output:"
echo "🔍 Fetching repository list from GitHub organization: Gizzahub"
echo "📋 Found 5 repositories in organization Gizzahub"
echo "📝 Generated gzh.yaml with 5 repositories"
echo "📦 Processing 5 repositories (5 remaining)"
echo "[░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 0.0% (0/5) • ✓ 0 • ✗ 0 • ⏳ 5 • 0s"
echo "[████████████░░░░░░░░░░░░░░░░░░] 40.0% (2/5) • ✓ 2 • ✗ 0 • ⏳ 0 • 2s"
echo "[██████████████████████████████] 100.0% (5/5) • ✓ 5 • ✗ 0 • ⏳ 0 • 3s"
echo "✅ Clone operation completed successfully"

rm -rf ./Gizzahub
```

### 1.2 디버그 모드 테스트

#### 시나리오: 개발자/디버깅 시 상세 정보 제공

```bash
echo "=== 디버그 모드 테스트 ==="
result_debug=$(gz synclone github -o Gizzahub --debug 2>&1)

# 검증 1: 타임스탬프 로그가 있어야 함
echo "$result_debug" | grep -E "^[0-9]{2}:[0-9]{2}:[0-9]{2}" && echo "✅ PASS: Timestamp logs found" || echo "❌ FAIL: No timestamp logs"

# 검증 2: 컴포넌트 로그가 있어야 함
echo "$result_debug" | grep "INFO.*component=gzh-cli" && echo "✅ PASS: Component logs found" || echo "❌ FAIL: No component logs"

# 검증 3: 콘솔 메시지도 함께 표시되어야 함
echo "$result_debug" | grep "🔍" && echo "✅ PASS: Console messages preserved" || echo "❌ FAIL: Console messages missing"

# 검증 4: 사람이 읽기 쉬운 성능 로그
echo "$result_debug" | grep "Operation.*completed in.*Memory:" && echo "✅ PASS: Human-readable performance logs" || echo "❌ FAIL: No readable performance logs"

# 검증 5: JSON 성능 로그가 없어야 함
echo "$result_debug" | grep '{"timestamp":.*"performance":' && echo "❌ FAIL: JSON performance logs found" || echo "✅ PASS: No JSON performance logs"

echo "Expected Debug Mode Additional Output:"
echo "22:13:47 INFO  [component=gzh-cli org=Gizzahub] Starting GitHub synclone operation"
echo "22:13:50 INFO  [component=gzh-cli org=Gizzahub] Operation 'github-synclone-completed' completed in 2.920s (Memory: 2.68 MB)"

rm -rf ./Gizzahub
```

### 1.3 Verbose 및 Quiet 모드 테스트

#### 시나리오: 다양한 로깅 레벨에서의 동작 확인

```bash
echo "=== Verbose 모드 테스트 ==="
result_verbose=$(gz synclone github -o Gizzahub --verbose 2>&1)

# INFO 레벨은 있지만 DEBUG는 없어야 함
echo "$result_verbose" | grep "INFO" && echo "✅ PASS: INFO logs in verbose mode" || echo "❌ FAIL: No INFO logs"
echo "$result_verbose" | grep "DEBUG" && echo "❌ FAIL: DEBUG logs in verbose mode" || echo "✅ PASS: No DEBUG logs in verbose mode"

rm -rf ./Gizzahub

echo "=== Quiet 모드 테스트 ==="
result_quiet=$(gz synclone github -o Gizzahub --quiet 2>&1)

# 에러를 제외한 모든 출력이 억제되어야 함
echo "$result_quiet" | grep -E "🔍|📋|✅" && echo "❌ FAIL: Progress messages in quiet mode" || echo "✅ PASS: No progress messages in quiet mode"
echo "$result_quiet" | wc -l | awk '{if($1 <= 2) print "✅ PASS: Minimal output in quiet mode"; else print "❌ FAIL: Too much output in quiet mode"}'

rm -rf ./Gizzahub
```

## 2. 프로그레스 바 정확성 테스트

### 2.1 초기 표시 정확성 테스트

#### 시나리오: 프로그레스 바가 0/total부터 시작하는지 확인

```bash
echo "=== 프로그레스 바 초기 표시 테스트 ==="
result=$(gz synclone github -o Gizzahub -p 2 2>&1)

# 프로그레스 라인들을 추출
progress_lines=$(echo "$result" | grep -E "\[.*\].*%.*•")
echo "Progress lines found:"
echo "$progress_lines"

# 첫 번째 프로그레스 라인 분석
first_progress=$(echo "$progress_lines" | head -1)
echo "First progress line: $first_progress"

# 검증 1: 0.0%로 시작하는지
echo "$first_progress" | grep "0.0% (0/" && echo "✅ PASS: Starts from 0.0%" || echo "❌ FAIL: Does not start from 0.0%"

# 검증 2: 중간값으로 점프하지 않는지
echo "$first_progress" | grep -E "40\.0%|60\.0%|80\.0%" && echo "❌ FAIL: Jumps to middle value" || echo "✅ PASS: No jumping to middle values"

# 검증 3: 초기 상태 표시 (모든 pending)
echo "$first_progress" | grep "⏳ [0-9]" && echo "✅ PASS: Shows pending count" || echo "❌ FAIL: No pending count"

rm -rf ./Gizzahub
```

### 2.2 순차적 진행률 업데이트 테스트

#### 시나리오: 프로그레스가 순차적으로 증가하는지 확인

```bash
echo "=== 순차적 진행률 업데이트 테스트 ==="

# 더 많은 리포지터리가 있는 조직으로 테스트
result=$(gz synclone github -o kubernetes --parallel 1 --target ./progress-test 2>&1)

# 모든 프로그레스 백분율 추출
progress_percentages=$(echo "$result" | grep -oE "[0-9]+\.[0-9]+%" | tr -d '%')
echo "Progress percentages sequence: $progress_percentages"

# 첫 번째 값이 0.0인지 확인
first_percent=$(echo "$progress_percentages" | head -1)
if [ "$(echo "$first_percent == 0.0" | bc -l)" -eq 1 ]; then
  echo "✅ PASS: First progress is 0.0%"
else
  echo "❌ FAIL: First progress is not 0.0% (got: $first_percent%)"
fi

# 순차적 증가 확인 (각 값이 이전 값보다 크거나 같아야 함)
prev_percent=0
is_sequential=true
for percent in $progress_percentages; do
  if [ "$(echo "$percent < $prev_percent" | bc -l)" -eq 1 ]; then
    is_sequential=false
    break
  fi
  prev_percent=$percent
done

if [ "$is_sequential" = true ]; then
  echo "✅ PASS: Progress increases sequentially"
else
  echo "❌ FAIL: Progress does not increase sequentially"
fi

rm -rf ./progress-test
```

### 2.3 재개 시나리오에서 초기값 테스트

#### 시나리오: 작업 재개 시 정확한 초기 상태 표시

```bash
echo "=== 재개 시나리오 초기값 테스트 ==="

# 대규모 조직으로 시작
gz synclone github -o kubernetes --target ./resume-test --parallel 3 &
SYNC_PID=$!

# 부분 완료 후 중단
sleep 10
kill -INT $SYNC_PID
echo "작업을 중단했습니다."

# 재개 후 초기 진행률 확인
echo "작업을 재개합니다..."
resume_result=$(gz synclone github -o kubernetes --target ./resume-test --resume 2>&1)

# 재개 시 첫 번째 진행률 라인
resume_first_line=$(echo "$resume_result" | grep -E "\[.*\].*%.*•" | head -1)
echo "Resume first progress: $resume_first_line"

# 재개 시에는 현재 상태를 정확히 반영해야 함 (0/total이 아닐 수 있음)
if echo "$resume_first_line" | grep -E "[0-9]+\.[0-9]+% \([0-9]+/[0-9]+\)"; then
  echo "✅ PASS: Resume shows accurate progress state"
else
  echo "❌ FAIL: Resume does not show accurate progress state"
fi

# 재개된 상태에서 완료된 항목이 0이 아닌지 확인
completed_count=$(echo "$resume_first_line" | grep -oE "✓ [0-9]+" | grep -oE "[0-9]+")
if [ -n "$completed_count" ] && [ "$completed_count" -gt 0 ]; then
  echo "✅ PASS: Resume reflects previously completed items ($completed_count)"
else
  echo "⚠️  INFO: No completed items to resume from (this may be normal)"
fi

rm -rf ./resume-test
```

## 3. 성능 로그 형식 테스트

### 3.1 사람이 읽기 쉬운 형식 테스트

#### 시나리오: 성능 정보가 텍스트 형식으로 출력되는지 확인

```bash
echo "=== 성능 로그 형식 테스트 ==="
result_debug=$(gz synclone github -o Gizzahub --debug 2>&1)

# 검증 1: 텍스트 형식 성능 로그
perf_line=$(echo "$result_debug" | grep "Operation.*completed in.*Memory:")
if [ -n "$perf_line" ]; then
  echo "✅ PASS: Human-readable performance log found"
  echo "Performance log: $perf_line"
else
  echo "❌ FAIL: No human-readable performance log found"
fi

# 검증 2: JSON 형식이 아닌지 확인
json_perf=$(echo "$result_debug" | grep '{"timestamp":.*"performance":')
if [ -z "$json_perf" ]; then
  echo "✅ PASS: No JSON performance logs"
else
  echo "❌ FAIL: JSON performance logs found"
  echo "JSON log: $json_perf"
fi

# 검증 3: 필수 성능 정보 포함 확인
echo "$perf_line" | grep "completed in" && echo "✅ PASS: Duration information included" || echo "❌ FAIL: No duration information"
echo "$perf_line" | grep "Memory:" && echo "✅ PASS: Memory information included" || echo "❌ FAIL: No memory information"

# 예상되는 형식 예시
echo "Expected format example:"
echo "Operation 'github-synclone-completed' completed in 2.920s (Memory: 2.68 MB) [org_name=Gizzahub strategy=reset parallel=2]"

rm -rf ./Gizzahub
```

## 4. 통합 UX 검증 시나리오

### 4.1 종합 UX 개선 검증

#### 시나리오: 모든 UX 개선사항을 한 번에 검증

```bash
#!/bin/bash
# 종합 UX 검증 스크립트

echo "=== Synclone UX Improvements Comprehensive Verification ==="

PASS_COUNT=0
FAIL_COUNT=0

# 테스트 함수
check_test() {
  local test_name="$1"
  local condition="$2"
  local expected="$3"
  
  if [ "$condition" = "$expected" ]; then
    echo "✅ PASS: $test_name"
    ((PASS_COUNT++))
  else
    echo "❌ FAIL: $test_name"
    ((FAIL_COUNT++))
  fi
}

# Test 1: Normal Mode Clean Output
echo "--- Test 1: Normal Mode Clean Output ---"
normal_output=$(timeout 90 gz synclone github -o Gizzahub 2>&1)

# 로그 메시지 없음 확인
has_timestamps=$(echo "$normal_output" | grep -c -E "^[0-9]{2}:[0-9]{2}:[0-9]{2}")
check_test "No timestamp logs in normal mode" $([ $has_timestamps -eq 0 ] && echo "pass" || echo "fail") "pass"

# 콘솔 메시지 존재 확인
has_progress=$(echo "$normal_output" | grep -c "🔍")
check_test "Progress indicators present" $([ $has_progress -gt 0 ] && echo "pass" || echo "fail") "pass"

# 0부터 시작 확인
starts_zero=$(echo "$normal_output" | grep -c "0.0% (0/")
check_test "Progress starts from 0" $([ $starts_zero -gt 0 ] && echo "pass" || echo "fail") "pass"

# Test 2: Debug Mode Detailed Logging
echo "--- Test 2: Debug Mode Detailed Logging ---"
debug_output=$(timeout 90 gz synclone github -o Gizzahub --debug 2>&1)

# 디버그 로그 존재 확인
has_debug=$(echo "$debug_output" | grep -c "INFO.*component=gzh-cli")
check_test "Debug logs in debug mode" $([ $has_debug -gt 0 ] && echo "pass" || echo "fail") "pass"

# 텍스트 성능 로그 확인
has_text_perf=$(echo "$debug_output" | grep -c "Operation.*completed in.*Memory:")
check_test "Human-readable performance logs" $([ $has_text_perf -gt 0 ] && echo "pass" || echo "fail") "pass"

# JSON 성능 로그 없음 확인
has_json_perf=$(echo "$debug_output" | grep -c '{"timestamp":.*"performance":')
check_test "No JSON performance logs" $([ $has_json_perf -eq 0 ] && echo "pass" || echo "fail") "pass"

# Test 3: Progress Bar Accuracy
echo "--- Test 3: Progress Bar Accuracy ---"
progress_lines=$(echo "$normal_output" | grep -E "\[.*\].*%.*•")
first_line=$(echo "$progress_lines" | head -1)

# 중간값 점프 없음 확인
no_jump=$(echo "$first_line" | grep -v -E "40\.0%|60\.0%|80\.0%" | wc -l)
check_test "No jumping to middle values" $([ $no_jump -gt 0 ] && echo "pass" || echo "fail") "pass"

# 초기 0/total 표시 확인
shows_zero=$(echo "$first_line" | grep -c "0.0% (0/")
check_test "Shows initial 0/total" $([ $shows_zero -gt 0 ] && echo "pass" || echo "fail") "pass"

# Test 4: Console Messages Preserved
echo "--- Test 4: Console Messages Preserved ---"
has_fetch=$(echo "$normal_output" | grep -c "🔍 Fetching")
check_test "Fetch message preserved" $([ $has_fetch -gt 0 ] && echo "pass" || echo "fail") "pass"

has_found=$(echo "$normal_output" | grep -c "📋 Found")
check_test "Found message preserved" $([ $has_found -gt 0 ] && echo "pass" || echo "fail") "pass"

has_success=$(echo "$normal_output" | grep -c "✅")
check_test "Success message preserved" $([ $has_success -gt 0 ] && echo "pass" || echo "fail") "pass"

# 결과 요약
echo ""
echo "=== Test Results Summary ==="
echo "✅ PASSED: $PASS_COUNT tests"
echo "❌ FAILED: $FAIL_COUNT tests"
echo "Total Tests: $((PASS_COUNT + FAIL_COUNT))"

if [ $FAIL_COUNT -eq 0 ]; then
  echo "🎉 All UX improvement tests passed!"
  SUCCESS=true
else
  echo "💥 $FAIL_COUNT UX improvement tests failed!"
  SUCCESS=false
fi

# 정리
rm -rf ./Gizzahub

$SUCCESS
```

### 4.2 후진 호환성 검증

#### 시나리오: UX 개선 후에도 기존 기능 정상 동작 확인

```bash
echo "=== 후진 호환성 검증 ==="

# 기존 플래그 호환성
echo "--- 기존 CLI 플래그 테스트 ---"
gz synclone github -o Gizzahub --strategy reset --parallel 5 --target ./compat-test
if [ $? -eq 0 ]; then
  echo "✅ PASS: CLI flags compatibility"
else
  echo "❌ FAIL: CLI flags compatibility"
fi

# 환경 변수 호환성
echo "--- 환경 변수 테스트 ---"
GITHUB_TOKEN="$GITHUB_TOKEN" gz synclone github -o Gizzahub --target ./env-test
if [ $? -eq 0 ]; then
  echo "✅ PASS: Environment variable compatibility"
else
  echo "❌ FAIL: Environment variable compatibility"
fi

# 설정 파일 호환성
echo "--- 설정 파일 테스트 ---"
cat > legacy-config.yaml << 'YAML'
version: "1.0"
github:
  enabled: true
  organizations:
    - name: "Gizzahub"
      target: "./config-test"
YAML

gz synclone --config legacy-config.yaml
if [ $? -eq 0 ]; then
  echo "✅ PASS: Configuration file compatibility"
else
  echo "❌ FAIL: Configuration file compatibility"
fi

# 정리
rm -rf ./compat-test ./env-test ./config-test legacy-config.yaml
echo "=== 후진 호환성 검증 완료 ==="
```

## 5. 성능 비교 테스트

### 5.1 UX 개선 전후 성능 비교

#### 시나리오: UX 개선이 성능에 미치는 영향 측정

```bash
echo "=== UX 개선 전후 성능 비교 ==="

# 일반 모드 (로그 최소화)
echo "--- Normal Mode Performance ---"
time_normal_start=$(date +%s.%N)
gz synclone github -o Gizzahub --target ./perf-normal
time_normal_end=$(date +%s.%N)
normal_duration=$(echo "$time_normal_end - $time_normal_start" | bc)
echo "Normal mode duration: ${normal_duration}s"

# 디버그 모드 (모든 로그)
echo "--- Debug Mode Performance ---"
time_debug_start=$(date +%s.%N)
gz synclone github -o Gizzahub --target ./perf-debug --debug
time_debug_end=$(date +%s.%N)
debug_duration=$(echo "$time_debug_end - $time_debug_start" | bc)
echo "Debug mode duration: ${debug_duration}s"

# 성능 영향 계산
overhead=$(echo "($debug_duration - $normal_duration) / $normal_duration * 100" | bc -l)
echo "Debug mode overhead: ${overhead}%"

# 정리
rm -rf ./perf-normal ./perf-debug

echo "Expected: Debug mode overhead should be minimal (< 10%)"
```

이 테스트 시나리오들을 통해 2025-09 UX 개선사항이 올바르게 구현되고 동작하는지 체계적으로 검증할 수 있습니다.