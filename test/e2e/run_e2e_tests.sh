#!/bin/bash

# End-to-End Test Runner
# This script runs E2E tests for the gz CLI application

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TIMEOUT=${TIMEOUT:-"20m"}
VERBOSE=${VERBOSE:-"true"}
PARALLEL=${PARALLEL:-"1"}
SHORT_MODE=${SHORT_MODE:-"false"}
CLEANUP=${CLEANUP:-"true"}

# Test directories
E2E_TEST_DIR="./test/e2e/scenarios"
PROJECT_ROOT=""

# Functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

find_project_root() {
    local dir="$PWD"
    while [[ "$dir" != "/" ]]; do
        if [[ -f "$dir/go.mod" ]]; then
            PROJECT_ROOT="$dir"
            return 0
        fi
        dir=$(dirname "$dir")
    done

    error "Could not find project root (go.mod not found)"
    exit 1
}

check_prerequisites() {
    log "Checking prerequisites..."

    # Check Go
    if ! command -v go &> /dev/null; then
        error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check if we're in the right directory
    find_project_root

    # Check if binary exists or can be built
    if [[ ! -f "$PROJECT_ROOT/gz" ]]; then
        log "Binary not found, building..."
        if ! build_binary; then
            error "Failed to build binary"
            exit 1
        fi
    fi

    success "Prerequisites check passed"
}

build_binary() {
    log "Building gz binary..."
    cd "$PROJECT_ROOT"

    if make build; then
        success "Binary built successfully"
        return 0
    else
        error "Failed to build binary"
        return 1
    fi
}

cleanup_test_artifacts() {
    if [[ "$CLEANUP" == "true" ]]; then
        log "Cleaning up test artifacts..."

        # Remove temporary test directories
        find /tmp -name "gz-e2e-test-*" -type d -exec rm -rf {} + 2>/dev/null || true

        # Remove any stray gz binaries in temp directories
        find /tmp -name "gz" -type f -delete 2>/dev/null || true

        success "Cleanup completed"
    fi
}

run_tests() {
    local test_pattern="$1"
    local test_name="$2"

    log "Running $test_name tests..."

    local go_test_args=("-timeout" "$TIMEOUT")

    if [[ "$VERBOSE" == "true" ]]; then
        go_test_args+=("-v")
    fi

    if [[ "$SHORT_MODE" == "true" ]]; then
        go_test_args+=("-short")
    fi

    if [[ "$PARALLEL" != "1" ]]; then
        go_test_args+=("-parallel" "$PARALLEL")
    fi

    if [[ -n "$test_pattern" ]]; then
        go_test_args+=("-run" "$test_pattern")
    fi

    go_test_args+=("$E2E_TEST_DIR")

    cd "$PROJECT_ROOT"

    if go test "${go_test_args[@]}"; then
        success "$test_name tests passed"
        return 0
    else
        error "$test_name tests failed"
        return 1
    fi
}

show_usage() {
    cat << EOF
End-to-End Test Runner

Usage: $0 [OPTIONS] [COMMAND]

Commands:
    all                 Run all E2E tests (default)
    bulk-clone          Run bulk clone E2E tests
    config              Run configuration E2E tests
    ide                 Run IDE E2E tests
    help                Show help for specific test scenarios

Options:
    -t, --timeout TIMEOUT    Test timeout (default: 20m)
    -v, --verbose           Enable verbose output (default: true)
    -s, --short             Run in short mode (skip slow tests)
    -p, --parallel N        Number of parallel tests (default: 1)
    --no-cleanup           Skip cleanup of test artifacts
    --build                Force rebuild of binary
    -h, --help             Show this help message

Environment Variables:
    TIMEOUT        Test timeout (default: 20m)
    VERBOSE        Enable verbose output (default: true)
    SHORT_MODE     Run in short mode (default: false)
    PARALLEL       Number of parallel tests (default: 1)
    CLEANUP        Cleanup test artifacts (default: true)

Examples:
    $0                              # Run all E2E tests
    $0 bulk-clone                   # Run bulk clone tests only
    $0 -s all                      # Run all tests in short mode
    $0 -t 30m --parallel 2 all     # Run with 30m timeout and 2 parallel tests
    $0 config                      # Run configuration tests only

Test Scenarios:
    bulk-clone    - Repository bulk cloning workflows
    config        - Configuration management and validation
    ide           - JetBrains IDE integration and monitoring

EOF
}

show_test_help() {
    cat << EOF
E2E Test Scenarios Help

Bulk Clone Tests (bulk-clone):
  - Configuration generation and validation
  - Multi-provider support (GitHub, GitLab, Gitea)
  - Clone strategies (reset, pull, fetch)
  - Pattern matching and filtering
  - Error handling and edge cases

Configuration Tests (config):
  - Configuration initialization
  - Validation and schema checking
  - Profile management
  - Environment variable overrides
  - Migration and backup

IDE Tests (ide):
  - JetBrains IDE detection
  - Settings monitoring and synchronization
  - Fix problematic configurations
  - Multi-IDE support
  - Backup and restore

Running Individual Test Functions:
    go test ./test/e2e/scenarios -v -run TestBulkClone_ConfigGeneration_E2E
    go test ./test/e2e/scenarios -v -run TestConfig_Init_E2E
    go test ./test/e2e/scenarios -v -run TestIDE_List_E2E

Debug Mode:
    Set environment variables for debugging:
    export GZ_DEBUG=true
    export GZ_TRACE=true

EOF
}

main() {
    local command="all"
    local force_build=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE="true"
                shift
                ;;
            -s|--short)
                SHORT_MODE="true"
                shift
                ;;
            -p|--parallel)
                PARALLEL="$2"
                shift 2
                ;;
            --no-cleanup)
                CLEANUP="false"
                shift
                ;;
            --build)
                force_build=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            all|bulk-clone|config|ide)
                command="$1"
                shift
                ;;
            help)
                show_test_help
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    log "Starting E2E tests..."
    log "Configuration: timeout=$TIMEOUT, verbose=$VERBOSE, short=$SHORT_MODE, parallel=$PARALLEL"

    # Setup
    check_prerequisites

    if [[ "$force_build" == "true" ]]; then
        build_binary
    fi

    # Trap cleanup on exit
    trap cleanup_test_artifacts EXIT

    local test_failed=false

    # Run tests based on command
    case $command in
        all)
            if [[ "$SHORT_MODE" == "false" ]]; then
                run_tests "TestBulkClone" "Bulk Clone" || test_failed=true
                run_tests "TestConfig" "Configuration" || test_failed=true
                run_tests "TestIDE" "IDE" || test_failed=true
            else
                log "Running quick E2E tests in short mode"
                run_tests "TestConfig_Init_E2E|TestBulkClone_ConfigValidation_E2E|TestIDE_List_E2E" "Quick E2E" || test_failed=true
            fi
            ;;
        bulk-clone)
            run_tests "TestBulkClone" "Bulk Clone" || test_failed=true
            ;;
        config)
            run_tests "TestConfig" "Configuration" || test_failed=true
            ;;
        ide)
            run_tests "TestIDE" "IDE" || test_failed=true
            ;;
    esac

    # Final cleanup
    cleanup_test_artifacts

    # Report results
    if [[ "$test_failed" == "true" ]]; then
        error "Some E2E tests failed"
        exit 1
    else
        success "All E2E tests passed!"
        exit 0
    fi
}

# Run main function with all arguments
main "$@"
