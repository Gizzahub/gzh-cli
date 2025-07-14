#!/bin/bash

# Docker Integration Test Runner
# This script runs Docker-based integration tests with proper setup and cleanup

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TIMEOUT=${TIMEOUT:-"30m"}
VERBOSE=${VERBOSE:-"true"}
PARALLEL=${PARALLEL:-"1"}
SHORT_MODE=${SHORT_MODE:-"false"}

# Test directories
DOCKER_TEST_DIR="./test/integration/docker"
TESTCONTAINERS_DIR="./test/integration/testcontainers"

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

check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        error "Docker daemon is not running or not accessible"
        exit 1
    fi
    
    # Check Go
    if ! command -v go &> /dev/null; then
        error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check available memory
    if command -v free &> /dev/null; then
        AVAILABLE_MB=$(free -m | awk 'NR==2{printf "%.0f", $7}')
        if [ "$AVAILABLE_MB" -lt 4096 ]; then
            warning "Available memory ($AVAILABLE_MB MB) is less than recommended 4GB"
        fi
    fi
    
    success "Prerequisites check passed"
}

cleanup_containers() {
    log "Cleaning up test containers..."
    
    # Remove testcontainers
    docker ps -a --filter "label=org.testcontainers=true" -q | xargs -r docker rm -f || true
    
    # Clean up networks
    docker network ls --filter "label=org.testcontainers=true" -q | xargs -r docker network rm || true
    
    # Prune unused containers and networks
    docker container prune -f || true
    docker network prune -f || true
    
    success "Container cleanup completed"
}

pull_images() {
    log "Pre-pulling container images..."
    
    local images=(
        "gitlab/gitlab-ce:16.11.0-ce.0"
        "gitea/gitea:1.21.10"
        "redis:7.2-alpine"
    )
    
    for image in "${images[@]}"; do
        log "Pulling $image..."
        if docker pull "$image"; then
            success "Pulled $image"
        else
            warning "Failed to pull $image, will retry during test"
        fi
    done
}

run_tests() {
    local test_path="$1"
    local test_name="$2"
    
    log "Running $test_name tests..."
    
    local go_test_args=("-timeout" "$TIMEOUT")
    
    if [ "$VERBOSE" = "true" ]; then
        go_test_args+=("-v")
    fi
    
    if [ "$SHORT_MODE" = "true" ]; then
        go_test_args+=("-short")
    fi
    
    if [ "$PARALLEL" != "1" ]; then
        go_test_args+=("-parallel" "$PARALLEL")
    fi
    
    go_test_args+=("$test_path")
    
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
Docker Integration Test Runner

Usage: $0 [OPTIONS] [COMMAND]

Commands:
    all                 Run all Docker integration tests (default)
    testcontainers      Run only testcontainer tests
    docker              Run only Docker integration tests
    gitlab              Run only GitLab integration tests
    gitea               Run only Gitea integration tests
    redis               Run only Redis integration tests
    multi               Run only multi-provider integration tests
    
Options:
    -t, --timeout TIMEOUT    Test timeout (default: 30m)
    -v, --verbose           Enable verbose output (default: true)
    -s, --short             Run in short mode (skip Docker tests)
    -p, --parallel N        Number of parallel tests (default: 1)
    --no-pull              Skip pre-pulling container images
    --no-cleanup           Skip cleanup before tests
    -h, --help             Show this help message

Environment Variables:
    TIMEOUT        Test timeout (default: 30m)
    VERBOSE        Enable verbose output (default: true)
    SHORT_MODE     Run in short mode (default: false)
    PARALLEL       Number of parallel tests (default: 1)

Examples:
    $0                              # Run all tests
    $0 docker                       # Run Docker integration tests only
    $0 -s all                      # Run all tests in short mode (skip Docker)
    $0 -t 45m --parallel 2 all     # Run with 45m timeout and 2 parallel tests
    $0 gitlab                      # Run only GitLab tests
    
EOF
}

main() {
    local command="all"
    local skip_pull=false
    local skip_cleanup=false
    
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
            --no-pull)
                skip_pull=true
                shift
                ;;
            --no-cleanup)
                skip_cleanup=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            all|testcontainers|docker|gitlab|gitea|redis|multi)
                command="$1"
                shift
                ;;
            *)
                error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    log "Starting Docker integration tests..."
    log "Configuration: timeout=$TIMEOUT, verbose=$VERBOSE, short=$SHORT_MODE, parallel=$PARALLEL"
    
    # Setup
    check_prerequisites
    
    if [ "$skip_cleanup" = "false" ]; then
        cleanup_containers
    fi
    
    if [ "$skip_pull" = "false" ] && [ "$SHORT_MODE" = "false" ]; then
        pull_images
    fi
    
    # Trap cleanup on exit
    trap cleanup_containers EXIT
    
    local test_failed=false
    
    # Run tests based on command
    case $command in
        all)
            if [ "$SHORT_MODE" = "false" ]; then
                run_tests "$TESTCONTAINERS_DIR/..." "Testcontainers" || test_failed=true
                run_tests "$DOCKER_TEST_DIR/..." "Docker Integration" || test_failed=true
            else
                log "Skipping Docker tests in short mode"
            fi
            ;;
        testcontainers)
            run_tests "$TESTCONTAINERS_DIR/..." "Testcontainers" || test_failed=true
            ;;
        docker)
            run_tests "$DOCKER_TEST_DIR/..." "Docker Integration" || test_failed=true
            ;;
        gitlab)
            run_tests "$DOCKER_TEST_DIR" -run "TestBulkClone_GitLab" "GitLab Integration" || test_failed=true
            ;;
        gitea)
            run_tests "$DOCKER_TEST_DIR" -run "TestBulkClone_Gitea" "Gitea Integration" || test_failed=true
            ;;
        redis)
            run_tests "$DOCKER_TEST_DIR" -run "TestBulkClone_Redis" "Redis Integration" || test_failed=true
            ;;
        multi)
            run_tests "$DOCKER_TEST_DIR" -run "TestMultiProvider" "Multi-Provider Integration" || test_failed=true
            ;;
    esac
    
    # Final cleanup
    cleanup_containers
    
    # Report results
    if [ "$test_failed" = "true" ]; then
        error "Some tests failed"
        exit 1
    else
        success "All tests passed!"
        exit 0
    fi
}

# Run main function with all arguments
main "$@"