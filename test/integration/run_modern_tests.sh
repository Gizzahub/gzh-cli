#!/bin/bash
set -euo pipefail

# run_modern_tests.sh - Modern integration test runner for gzh-manager-go
# This script runs the updated integration tests that work with the current API

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TIMEOUT="${TEST_TIMEOUT:-30m}"
VERBOSE="${VERBOSE:-false}"
COVERAGE="${COVERAGE:-false}"
PARALLEL="${PARALLEL:-true}"

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"
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

Run modern integration tests for gzh-manager-go

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    -c, --coverage          Enable test coverage reporting
    -t, --timeout DURATION Set test timeout (default: 30m)
    -s, --sequential        Run tests sequentially instead of parallel
    --skip-env-check        Skip environment variable checks
    --dry-run              Show what would be executed without running tests

ENVIRONMENT VARIABLES:
    GITHUB_TOKEN           GitHub personal access token (for GitHub tests)
    GITLAB_TOKEN           GitLab personal access token (for GitLab tests)
    GITEA_TOKEN            Gitea personal access token (for Gitea tests)
    TEST_TIMEOUT           Test timeout duration (default: 30m)
    VERBOSE                Enable verbose output (true/false)
    COVERAGE               Enable coverage reporting (true/false)

EXAMPLES:
    # Run all integration tests
    $0

    # Run with verbose output and coverage
    $0 --verbose --coverage

    # Run with custom timeout
    $0 --timeout 45m

    # Run sequentially for debugging
    $0 --sequential --verbose

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
    log "Go version: $go_version"

    # Check if we're in the correct directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        error "Not in a Go module directory. Please run from the project root."
        return 1
    fi

    # Check if required test files exist
    if [[ ! -f "$SCRIPT_DIR/bulk_clone_modern_test.go" ]]; then
        error "Modern integration test file not found"
        return 1
    fi

    log "Prerequisites check passed"
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
        log "GitHub token found"
        tokens_found=$((tokens_found + 1))
    else
        warn "GITHUB_TOKEN not set - GitHub integration tests will be skipped"
    fi

    if [[ -n "${GITLAB_TOKEN:-}" ]]; then
        log "GitLab token found"
        tokens_found=$((tokens_found + 1))
    else
        warn "GITLAB_TOKEN not set - GitLab integration tests will be skipped"
    fi

    if [[ -n "${GITEA_TOKEN:-}" ]]; then
        log "Gitea token found"
        tokens_found=$((tokens_found + 1))
    else
        warn "GITEA_TOKEN not set - Gitea integration tests will be skipped"
    fi

    if [[ $tokens_found -eq 0 ]]; then
        warn "No authentication tokens found - some tests will be skipped"
        warn "Set GITHUB_TOKEN, GITLAB_TOKEN, or GITEA_TOKEN to enable full testing"
    else
        log "Found $tokens_found authentication token(s)"
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

# Run integration tests
run_tests() {
    log "Starting integration tests..."
    
    local test_flags
    test_flags=$(build_test_flags)
    
    # Change to project root
    cd "$PROJECT_ROOT"
    
    # Run the modern integration tests
    local test_cmd="go test $test_flags ./test/integration/bulk_clone_modern_test.go"
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log "DRY RUN - would execute: $test_cmd"
        return 0
    fi
    
    log "Executing: $test_cmd"
    
    if eval "$test_cmd"; then
        success "Integration tests completed successfully"
        return 0
    else
        error "Integration tests failed"
        return 1
    fi
}

# Generate coverage report
generate_coverage() {
    if [[ "$COVERAGE" != "true" ]] || [[ ! -f "coverage.out" ]]; then
        return 0
    fi
    
    log "Generating coverage report..."
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    
    # Display coverage summary
    local coverage_percent
    coverage_percent=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}')
    
    log "Coverage report generated: coverage.html"
    log "Total coverage: $coverage_percent"
}

# Cleanup function
cleanup() {
    log "Cleaning up..."
    
    # Remove temporary files if they exist
    if [[ -f "coverage.out" ]] && [[ "$COVERAGE" != "true" ]]; then
        rm -f coverage.out
    fi
}

# Main execution
main() {
    parse_args "$@"
    
    log "Starting modern integration test runner..."
    log "Project root: $PROJECT_ROOT"
    log "Script directory: $SCRIPT_DIR"
    log "Timeout: $TIMEOUT"
    log "Verbose: $VERBOSE"
    log "Coverage: $COVERAGE"
    log "Parallel: $PARALLEL"
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Run all checks and tests
    check_prerequisites || exit 1
    check_environment
    run_tests || exit 1
    generate_coverage
    
    success "All integration tests completed successfully!"
}

# Execute main function if script is run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi