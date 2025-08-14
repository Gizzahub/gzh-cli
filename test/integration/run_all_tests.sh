#!/bin/bash
set -euo pipefail

# run_all_tests.sh - Comprehensive integration test runner
# This script runs all modernized integration tests for gzh-cli

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
TIMEOUT="${TEST_TIMEOUT:-45m}"
VERBOSE="${VERBOSE:-false}"
COVERAGE="${COVERAGE:-false}"
PARALLEL="${PARALLEL:-true}"
DRY_RUN="${DRY_RUN:-false}"

# Test categories
RUN_BULK_CLONE="${RUN_BULK_CLONE:-true}"
RUN_NET_ENV="${RUN_NET_ENV:-true}"
RUN_CLI="${RUN_CLI:-true}"

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"
}

info() {
    echo -e "${CYAN}[INFO]${NC} $*"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

# Print usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Run comprehensive integration tests for gzh-cli

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    -c, --coverage          Enable test coverage reporting
    -t, --timeout DURATION Set test timeout (default: 45m)
    -s, --sequential        Run tests sequentially instead of parallel
    --skip-env-check        Skip environment variable checks
    --dry-run              Show what would be executed without running tests
    --bulk-clone-only      Run only bulk-clone tests
    --net-env-only         Run only net-env tests
    --cli-only             Run only CLI integration tests
    --skip-build           Skip building the gz binary

ENVIRONMENT VARIABLES:
    GITHUB_TOKEN           GitHub personal access token
    GITLAB_TOKEN           GitLab personal access token
    GITEA_TOKEN            Gitea personal access token
    TEST_TIMEOUT           Test timeout duration (default: 45m)
    VERBOSE                Enable verbose output (true/false)
    COVERAGE               Enable coverage reporting (true/false)
    RUN_BULK_CLONE         Run bulk-clone tests (true/false)
    RUN_NET_ENV            Run net-env tests (true/false)
    RUN_CLI                Run CLI tests (true/false)

EXAMPLES:
    # Run all integration tests
    $0

    # Run with verbose output and coverage
    $0 --verbose --coverage

    # Run only bulk-clone tests
    $0 --bulk-clone-only

    # Run with custom timeout
    $0 --timeout 60m

    # Dry run to see what would be executed
    $0 --dry-run

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -c|--coverage)
                COVERAGE=true
                shift
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -s|--sequential)
                PARALLEL=false
                shift
                ;;
            --skip-env-check)
                SKIP_ENV_CHECK=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --bulk-clone-only)
                RUN_BULK_CLONE=true
                RUN_NET_ENV=false
                RUN_CLI=false
                shift
                ;;
            --net-env-only)
                RUN_BULK_CLONE=false
                RUN_NET_ENV=true
                RUN_CLI=false
                shift
                ;;
            --cli-only)
                RUN_BULK_CLONE=false
                RUN_NET_ENV=false
                RUN_CLI=true
                shift
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        error "Go is not installed or not in PATH"
        return 1
    fi

    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    info "Go version: $go_version"

    # Check if we're in the correct directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        error "Not in a Go module directory. Please run from the project root."
        return 1
    fi

    # Check if required test files exist
    local required_files=(
        "$SCRIPT_DIR/bulk_clone_modern_test.go"
        "$SCRIPT_DIR/net-env/net_env_integration_test.go"
        "$SCRIPT_DIR/run_modern_tests.sh"
    )

    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            error "Required test file not found: $file"
            return 1
        fi
    done

    info "Prerequisites check passed"
}

# Check environment variables
check_environment() {
    if [[ "${SKIP_ENV_CHECK:-false}" == "true" ]]; then
        warn "Skipping environment variable checks"
        return 0
    fi

    log "Checking environment variables..."

    local tokens_found=0

    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        info "GitHub token found"
        tokens_found=$((tokens_found + 1))
    else
        warn "GITHUB_TOKEN not set - GitHub integration tests will be skipped"
    fi

    if [[ -n "${GITLAB_TOKEN:-}" ]]; then
        info "GitLab token found"
        tokens_found=$((tokens_found + 1))
    else
        warn "GITLAB_TOKEN not set - GitLab integration tests will be skipped"
    fi

    if [[ -n "${GITEA_TOKEN:-}" ]]; then
        info "Gitea token found"
        tokens_found=$((tokens_found + 1))
    else
        warn "GITEA_TOKEN not set - Gitea integration tests will be skipped"
    fi

    if [[ $tokens_found -eq 0 ]]; then
        warn "No authentication tokens found - some tests will be skipped"
        warn "Set GITHUB_TOKEN, GITLAB_TOKEN, or GITEA_TOKEN to enable full testing"
    else
        info "Found $tokens_found authentication token(s)"
    fi
}

# Build the gz binary
build_binary() {
    if [[ "${SKIP_BUILD:-false}" == "true" ]]; then
        warn "Skipping binary build"
        return 0
    fi

    log "Building gz binary..."

    cd "$PROJECT_ROOT"

    if [[ "$DRY_RUN" == "true" ]]; then
        info "DRY RUN - would execute: make build"
        return 0
    fi

    if make build; then
        success "Binary build completed"
    else
        warn "Binary build failed - some CLI tests may be skipped"
    fi
}

# Build test flags
build_test_flags() {
    local flags=()

    flags+=("-timeout" "$TIMEOUT")

    if [[ "$VERBOSE" == "true" ]]; then
        flags+=("-v")
    fi

    if [[ "$COVERAGE" == "true" ]]; then
        flags+=("-cover" "-coverprofile=coverage.out")
    fi

    if [[ "$PARALLEL" == "false" ]]; then
        flags+=("-p" "1")
    fi

    # Add race detection
    flags+=("-race")

    # Add test tags for integration tests
    flags+=("-tags" "integration")

    echo "${flags[@]}"
}

# Run bulk-clone integration tests
run_bulk_clone_tests() {
    if [[ "$RUN_BULK_CLONE" != "true" ]]; then
        info "Skipping bulk-clone tests"
        return 0
    fi

    log "Running bulk-clone integration tests..."

    local test_flags
    test_flags=$(build_test_flags)

    cd "$PROJECT_ROOT"

    local test_cmd="go test $test_flags ./test/integration/bulk_clone_modern_test.go"

    if [[ "$DRY_RUN" == "true" ]]; then
        info "DRY RUN - would execute: $test_cmd"
        return 0
    fi

    info "Executing: $test_cmd"

    if eval "$test_cmd"; then
        success "Bulk-clone integration tests passed"
        return 0
    else
        error "Bulk-clone integration tests failed"
        return 1
    fi
}

# Run net-env integration tests
run_net_env_tests() {
    if [[ "$RUN_NET_ENV" != "true" ]]; then
        info "Skipping net-env tests"
        return 0
    fi

    log "Running net-env integration tests..."

    local test_flags
    test_flags=$(build_test_flags)

    cd "$PROJECT_ROOT"

    local test_cmd="go test $test_flags ./test/integration/net-env/..."

    if [[ "$DRY_RUN" == "true" ]]; then
        info "DRY RUN - would execute: $test_cmd"
        return 0
    fi

    info "Executing: $test_cmd"

    if eval "$test_cmd"; then
        success "Net-env integration tests passed"
        return 0
    else
        error "Net-env integration tests failed"
        return 1
    fi
}

# Run CLI integration tests
run_cli_tests() {
    if [[ "$RUN_CLI" != "true" ]]; then
        info "Skipping CLI tests"
        return 0
    fi

    log "Running CLI integration tests..."

    # Check if gz binary exists
    if [[ ! -f "$PROJECT_ROOT/gz" ]] && [[ ! -f "$(which gz 2>/dev/null)" ]]; then
        warn "gz binary not found - building first"
        build_binary
    fi

    # Run CLI tests through the existing net-env integration test
    # which includes CLI testing scenarios
    run_net_env_tests
}

# Generate comprehensive coverage report
generate_coverage() {
    if [[ "$COVERAGE" != "true" ]] || [[ ! -f "coverage.out" ]]; then
        return 0
    fi

    log "Generating comprehensive coverage report..."

    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html

    # Generate JSON coverage report for CI
    go tool cover -func=coverage.out -o coverage.txt

    # Display coverage summary
    local coverage_percent
    coverage_percent=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}')

    info "Coverage report generated: coverage.html"
    info "Coverage summary: coverage.txt"
    info "Total coverage: $coverage_percent"

    # Check coverage threshold (optional)
    local threshold="${COVERAGE_THRESHOLD:-50.0}"
    local coverage_num
    coverage_num=$(echo "$coverage_percent" | sed 's/%//')

    if (( $(echo "$coverage_num >= $threshold" | bc -l) )); then
        success "Coverage $coverage_percent meets threshold of $threshold%"
    else
        warn "Coverage $coverage_percent below threshold of $threshold%"
    fi
}

# Run performance benchmarks
run_benchmarks() {
    if [[ "${RUN_BENCHMARKS:-false}" != "true" ]]; then
        return 0
    fi

    log "Running integration benchmarks..."

    cd "$PROJECT_ROOT"

    local bench_cmd="go test -bench=. -benchmem ./test/integration/..."

    if [[ "$DRY_RUN" == "true" ]]; then
        info "DRY RUN - would execute: $bench_cmd"
        return 0
    fi

    info "Executing: $bench_cmd"

    if eval "$bench_cmd"; then
        success "Benchmarks completed"
    else
        warn "Some benchmarks failed"
    fi
}

# Cleanup function
cleanup() {
    log "Cleaning up..."

    # Remove temporary files if they exist
    if [[ -f "coverage.out" ]] && [[ "$COVERAGE" != "true" ]]; then
        rm -f coverage.out
    fi

    # Clean up any temporary test artifacts
    find "$SCRIPT_DIR" -name "*.tmp" -delete 2>/dev/null || true
}

# Print test summary
print_summary() {
    log "Integration Test Summary:"
    info "==============================================="
    info "Project: gzh-cli"
    info "Total duration: $1"
    info "Tests run:"

    if [[ "$RUN_BULK_CLONE" == "true" ]]; then
        info "  ✓ Bulk-clone integration tests"
    fi

    if [[ "$RUN_NET_ENV" == "true" ]]; then
        info "  ✓ Net-env integration tests"
    fi

    if [[ "$RUN_CLI" == "true" ]]; then
        info "  ✓ CLI integration tests"
    fi

    if [[ "$COVERAGE" == "true" ]]; then
        info "  ✓ Coverage reporting enabled"
    fi

    info "==============================================="
}

# Main execution
main() {
    local start_time
    start_time=$(date +%s)

    parse_args "$@"

    log "Starting comprehensive integration test runner..."
    info "Project root: $PROJECT_ROOT"
    info "Script directory: $SCRIPT_DIR"
    info "Timeout: $TIMEOUT"
    info "Verbose: $VERBOSE"
    info "Coverage: $COVERAGE"
    info "Parallel: $PARALLEL"
    info "Dry run: $DRY_RUN"

    # Set up cleanup trap
    trap cleanup EXIT

    # Run all checks and tests
    check_prerequisites || exit 1
    check_environment
    build_binary

    local test_failures=0

    # Run test suites
    if ! run_bulk_clone_tests; then
        test_failures=$((test_failures + 1))
    fi

    if ! run_net_env_tests; then
        test_failures=$((test_failures + 1))
    fi

    if ! run_cli_tests; then
        test_failures=$((test_failures + 1))
    fi

    # Generate reports
    generate_coverage
    run_benchmarks

    # Calculate duration
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local duration_str
    duration_str=$(printf "%02d:%02d" $((duration/60)) $((duration%60)))

    # Print summary
    print_summary "$duration_str"

    if [[ $test_failures -eq 0 ]]; then
        success "All integration tests completed successfully!"
        exit 0
    else
        error "$test_failures test suite(s) failed"
        exit 1
    fi
}

# Execute main function if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
