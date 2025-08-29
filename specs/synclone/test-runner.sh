#!/bin/bash

# 스크립트명: Synclone Automated Test Runner
# 용도: synclone 기능을 자동으로 테스트하는 종합 테스트 스크립트
# 사용법: test-runner.sh [옵션] [테스트_타입]
# 예시: test-runner.sh --verbose basic

set -euo pipefail

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 글로벌 변수
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_BASE_DIR="$SCRIPT_DIR/test-results"
LOG_DIR="$TEST_BASE_DIR/logs"
VERBOSE=false
DRY_RUN=false
CLEANUP=true
PARALLEL_TESTS=false
TEST_TIMEOUT=600  # 10분 기본 타임아웃
RESULTS_FILE="$TEST_BASE_DIR/test-results.json"

# 도움말 출력
show_help() {
    cat << 'EOFHELP'
Synclone Automated Test Runner

사용법: $0 [옵션] [테스트_타입]

테스트 타입:
    basic       기본 기능 테스트 (기본값)
    filtering   필터링 기능 테스트
    performance 성능 테스트
    multi       다중 제공자 테스트
    enterprise  엔터프라이즈 테스트
    ci          CI/CD 테스트
    error       에러 처리 테스트
    all         모든 테스트 실행

옵션:
    -v, --verbose       상세 로그 출력
    -d, --dry-run      실제 실행 없이 계획만 표시
    -n, --no-cleanup   테스트 후 정리 안함
    -p, --parallel     병렬 테스트 실행
    -t, --timeout SEC  테스트 타임아웃 (초, 기본값: 600)
    -h, --help         이 도움말 표시

환경 변수:
    GITHUB_TOKEN       GitHub API 토큰
    GITLAB_TOKEN       GitLab API 토큰 (선택적)
    GITEA_TOKEN        Gitea API 토큰 (선택적)
    GZ_BINARY         gz 바이너리 경로 (기본값: gz)

예시:
    $0 basic                    # 기본 테스트
    $0 --verbose performance    # 상세 로그와 성능 테스트
    $0 --dry-run all           # 모든 테스트 계획 확인
EOFHELP
}

# 로그 함수들
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] INFO: $1" >> "$LOG_DIR/test.log"
    fi
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: $1" >> "$LOG_DIR/test.log"
    fi
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] WARNING: $1" >> "$LOG_DIR/test.log"
    fi
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    if [ "$VERBOSE" = true ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >> "$LOG_DIR/test.log"
    fi
}

# 환경 준비
setup_test_environment() {
    log_info "테스트 환경 준비 중..."
    
    # 디렉토리 생성
    mkdir -p "$TEST_BASE_DIR" "$LOG_DIR"
    
    # gz 바이너리 확인
    GZ_BINARY="${GZ_BINARY:-gz}"
    if ! command -v "$GZ_BINARY" &> /dev/null; then
        log_error "gz 바이너리를 찾을 수 없습니다: $GZ_BINARY"
        log_info "gz 바이너리를 빌드하거나 PATH에 추가해주세요"
        exit 1
    fi
    
    # 버전 확인
    local version
    version=$($GZ_BINARY --version 2>/dev/null || echo "unknown")
    log_info "gz 버전: $version"
    
    # 토큰 확인
    if [ -z "${GITHUB_TOKEN:-}" ]; then
        log_warning "GITHUB_TOKEN이 설정되지 않았습니다. API 제한이 적용될 수 있습니다."
    else
        log_info "GitHub 토큰 확인됨"
    fi
    
    # 네트워크 연결 확인
    if ! curl -s --connect-timeout 5 https://api.github.com/user >/dev/null 2>&1; then
        log_warning "GitHub API 연결 확인 실패. 네트워크를 확인해주세요."
    fi
}

# 테스트 결과 저장 (단순화)
save_test_result() {
    local test_name="$1"
    local status="$2"
    local duration="$3"
    local details="$4"
    
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ): $test_name - $status ($duration seconds) - $details" >> "$TEST_BASE_DIR/results.log"
}

# 단일 테스트 실행
run_single_test() {
    local test_name="$1"
    local config_file="$2"
    local timeout="${3:-$TEST_TIMEOUT}"
    
    log_info "테스트 시작: $test_name"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY-RUN] $test_name 테스트를 실행할 예정"
        return 0
    fi
    
    local start_time
    start_time=$(date +%s)
    local test_log="$LOG_DIR/${test_name}-$(date +%s).log"
    local status="FAILED"
    local details=""
    
    # 설정 파일 검증
    if [ ! -f "$config_file" ]; then
        log_error "설정 파일 없음: $config_file"
        details="Configuration file not found"
        save_test_result "$test_name" "$status" 0 "$details"
        return 1
    fi
    
    # 설정 파일 유효성 검사
    log_info "설정 파일 검증 중: $config_file"
    if ! timeout 30 "$GZ_BINARY" synclone config validate --config "$config_file" > "$test_log" 2>&1; then
        log_error "설정 파일 검증 실패"
        details="Configuration validation failed"
        if [ "$VERBOSE" = true ]; then
            cat "$test_log"
        fi
        save_test_result "$test_name" "$status" 0 "$details"
        return 1
    fi
    
    # 실제 테스트 실행
    log_info "테스트 실행 중: $test_name (타임아웃: ${timeout}초)"
    if timeout "$timeout" "$GZ_BINARY" synclone --config "$config_file" >> "$test_log" 2>&1; then
        status="PASSED"
        details="Test completed successfully"
        log_success "$test_name 테스트 성공"
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            details="Test timed out after ${timeout} seconds"
            log_error "$test_name 테스트 타임아웃"
        else
            details="Test failed with exit code $exit_code"
            log_error "$test_name 테스트 실패 (exit code: $exit_code)"
        fi
        
        if [ "$VERBOSE" = true ]; then
            log_info "마지막 로그 출력:"
            tail -20 "$test_log"
        fi
    fi
    
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    save_test_result "$test_name" "$status" "$duration" "$details"
    
    # 결과 확인
    verify_test_results "$test_name" "$test_log"
    
    if [ "$status" == "PASSED" ]; then
        return 0
    else
        return 1
    fi
}

# 테스트 결과 검증
verify_test_results() {
    local test_name="$1"
    local test_log="$2"
    
    log_info "$test_name 결과 검증 중..."
    
    # gzh.yaml 파일 확인
    local gzh_files
    gzh_files=$(find "$TEST_BASE_DIR" -name "gzh.yaml" 2>/dev/null | wc -l)
    
    # 리포지터리 수 확인
    local repo_count
    repo_count=$(find "$TEST_BASE_DIR" -name ".git" -type d 2>/dev/null | wc -l)
    
    log_info "$test_name 통계:"
    log_info "  - gzh.yaml 파일: $gzh_files개"
    log_info "  - 클론된 리포지터리: $repo_count개"
    
    # 에러 로그 확인
    if [ -f "$test_log" ]; then
        local error_count
        error_count=$(grep -c "ERROR\|FATAL\|FAIL" "$test_log" 2>/dev/null || echo 0)
        if [ "$error_count" -gt 0 ]; then
            log_warning "$test_name에서 $error_count개의 에러 발견"
        fi
    fi
}

# 기본 기능 테스트
test_basic() {
    log_info "=== 기본 기능 테스트 시작 ==="
    run_single_test "basic" "$SCRIPT_DIR/test-configs/basic-test.yaml" 300
}

# 필터링 기능 테스트
test_filtering() {
    log_info "=== 필터링 기능 테스트 시작 ==="
    run_single_test "filtering" "$SCRIPT_DIR/test-configs/filtering-test.yaml" 600
}

# 성능 테스트
test_performance() {
    log_info "=== 성능 테스트 시작 ==="
    run_single_test "performance" "$SCRIPT_DIR/test-configs/performance-test.yaml" 900
}

# CI/CD 테스트
test_ci() {
    log_info "=== CI/CD 테스트 시작 ==="
    run_single_test "ci-cd" "$SCRIPT_DIR/test-configs/ci-cd-test.yaml" 600
}

# 에러 처리 테스트
test_error() {
    log_info "=== 에러 처리 테스트 시작 ==="
    run_single_test "error-handling" "$SCRIPT_DIR/test-configs/error-handling-test.yaml" 300
}

# 모든 테스트 실행
test_all() {
    log_info "=== 전체 테스트 수행 시작 ==="
    
    local tests=("basic" "filtering" "performance" "ci" "error")
    local failed_tests=()
    
    for test in "${tests[@]}"; do
        if ! "test_$test"; then
            failed_tests+=("$test")
        fi
    done
    
    if [ ${#failed_tests[@]} -eq 0 ]; then
        log_success "모든 테스트 통과!"
    else
        log_error "실패한 테스트: ${failed_tests[*]}"
        return 1
    fi
}

# 정리 함수
cleanup() {
    if [ "$CLEANUP" = true ] && [ "$DRY_RUN" = false ]; then
        log_info "테스트 정리 중..."
        
        # 테스트 디렉토리 정리 (로그는 보존)
        find "$TEST_BASE_DIR" -mindepth 1 -maxdepth 1 -type d ! -name "logs" -exec rm -rf {} + 2>/dev/null || true
        
        # 오래된 로그 정리 (7일 이상)
        find "$LOG_DIR" -type f -name "*.log" -mtime +7 -delete 2>/dev/null || true
        
        log_success "정리 완료"
    fi
}

# 메인 함수
main() {
    local test_type="basic"
    
    # 인자 처리
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -n|--no-cleanup)
                CLEANUP=false
                shift
                ;;
            -p|--parallel)
                PARALLEL_TESTS=true
                shift
                ;;
            -t|--timeout)
                TEST_TIMEOUT="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            basic|filtering|performance|multi|enterprise|ci|error|all)
                test_type="$1"
                shift
                ;;
            *)
                log_error "알 수 없는 옵션: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 시그널 핸들러 설정
    trap cleanup EXIT
    trap 'log_info "테스트 중단됨"; exit 130' INT TERM
    
    # 환경 준비
    setup_test_environment
    
    # 테스트 실행
    log_info "테스트 타입: $test_type"
    
    case $test_type in
        basic) test_basic ;;
        filtering) test_filtering ;;
        performance) test_performance ;;
        ci) test_ci ;;
        error) test_error ;;
        all) test_all ;;
        *)
            log_error "지원하지 않는 테스트 타입: $test_type"
            exit 1
            ;;
    esac
    
    log_success "테스트 러너 완료!"
}

# 스크립트 실행
main "$@"
