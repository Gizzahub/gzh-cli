#!/bin/bash
# scripts/test-git-repo-e2e.sh
# End-to-End tests for Git Repo functionality

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="${TEST_DIR:-/tmp/gzh-git-repo-e2e-test}"
BINARY="${BINARY:-./gz}"
LOG_FILE="${TEST_DIR}/e2e-test.log"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}✅ $1${NC}" | tee -a "$LOG_FILE"
    ((TESTS_PASSED++))
}

error() {
    echo -e "${RED}❌ $1${NC}" | tee -a "$LOG_FILE"
    ((TESTS_FAILED++))
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_exit_code="${3:-0}"
    
    ((TESTS_RUN++))
    log "Running test: $test_name"
    
    if [[ $expected_exit_code -eq 0 ]]; then
        if eval "$test_command" >> "$LOG_FILE" 2>&1; then
            success "$test_name"
            return 0
        else
            error "$test_name (command failed)"
            return 1
        fi
    else
        if eval "$test_command" >> "$LOG_FILE" 2>&1; then
            error "$test_name (expected failure but succeeded)"
            return 1
        else
            success "$test_name (expected failure)"
            return 0
        fi
    fi
}

setup_test_environment() {
    log "Setting up test environment..."
    
    # Clean up previous test runs
    rm -rf "$TEST_DIR"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
    
    # Initialize log file
    echo "Git Repo E2E Test Log - $(date)" > "$LOG_FILE"
    
    # Check if binary exists
    if [[ ! -f "$BINARY" ]]; then
        error "Binary not found: $BINARY"
        error "Please build the binary first: make build"
        exit 1
    fi
    
    success "Test environment ready"
}

cleanup_test_environment() {
    log "Cleaning up test environment..."
    cd /
    if [[ -d "$TEST_DIR" ]]; then
        rm -rf "$TEST_DIR"
    fi
    success "Test environment cleaned up"
}

# Test help and basic functionality
test_help_commands() {
    log "Testing help commands..."
    
    run_test "Git repo help" "$BINARY git repo --help"
    run_test "Git repo clone help" "$BINARY git repo clone --help"
    run_test "Git repo list help" "$BINARY git repo list --help"
    run_test "Git repo create help" "$BINARY git repo create --help"
    run_test "Git repo delete help" "$BINARY git repo delete --help"
    run_test "Git repo archive help" "$BINARY git repo archive --help"
    run_test "Git repo sync help" "$BINARY git repo sync --help"
    run_test "Git repo search help" "$BINARY git repo search --help"
}

# Test command validation
test_command_validation() {
    log "Testing command validation..."
    
    # Test missing required parameters
    run_test "Clone without provider" "$BINARY git repo clone --org testorg" 1
    run_test "Clone without org" "$BINARY git repo clone --provider github" 1
    run_test "List without provider" "$BINARY git repo list --org testorg" 1
    run_test "Create without required params" "$BINARY git repo create --provider github" 1
    run_test "Sync without from" "$BINARY git repo sync --to gitlab:org/repo" 1
    run_test "Sync without to" "$BINARY git repo sync --from github:org/repo" 1
    
    # Test invalid parameters
    run_test "Clone with invalid provider" "$BINARY git repo clone --provider invalid --org testorg" 1
    run_test "Clone with invalid parallel" "$BINARY git repo clone --provider github --org testorg --parallel 0" 1
    run_test "Sync with invalid format" "$BINARY git repo sync --from invalid-format --to gitlab:org/repo" 1
}

# Test dry-run functionality
test_dry_run_commands() {
    log "Testing dry-run functionality..."
    
    # These should succeed without making actual API calls
    run_test "Clone dry-run" "$BINARY git repo clone --provider github --org gizzahub --dry-run --limit 5"
    run_test "Sync dry-run" "$BINARY git repo sync --from github:gizzahub/test --to gitlab:test/test --dry-run"
}

# Test output formats
test_output_formats() {
    log "Testing output formats..."
    
    # Test different output formats (dry-run to avoid API calls)
    run_test "List with table format" "$BINARY git repo list --provider github --org gizzahub --format table --dry-run"
    run_test "List with JSON format" "$BINARY git repo list --provider github --org gizzahub --format json --dry-run"
    run_test "List with YAML format" "$BINARY git repo list --provider github --org gizzahub --format yaml --dry-run"
    
    # Test invalid format
    run_test "List with invalid format" "$BINARY git repo list --provider github --org gizzahub --format invalid" 1
}

# Test configuration and environment
test_configuration() {
    log "Testing configuration handling..."
    
    # Test with missing configuration
    run_test "Command without config" "$BINARY git repo list --provider github --org testorg --dry-run"
    
    # Test environment variable handling
    export GZH_CONFIG_PATH="/nonexistent/config.yaml"
    run_test "Command with invalid config path" "$BINARY git repo list --provider github --org testorg --dry-run"
    unset GZH_CONFIG_PATH
}

# Test filtering and pattern matching
test_filtering() {
    log "Testing filtering and pattern matching..."
    
    # Test pattern matching (dry-run)
    run_test "Clone with match pattern" "$BINARY git repo clone --provider github --org gizzahub --match 'api-*' --dry-run"
    run_test "Clone with exclude pattern" "$BINARY git repo clone --provider github --org gizzahub --exclude 'test-*' --dry-run"
    run_test "List with language filter" "$BINARY git repo list --provider github --org gizzahub --language Go --dry-run"
    run_test "List with visibility filter" "$BINARY git repo list --provider github --org gizzahub --visibility public --dry-run"
}

# Test parallel execution
test_parallel_execution() {
    log "Testing parallel execution..."
    
    # Test different parallel worker counts
    run_test "Clone with 1 worker" "$BINARY git repo clone --provider github --org gizzahub --parallel 1 --dry-run --limit 5"
    run_test "Clone with 3 workers" "$BINARY git repo clone --provider github --org gizzahub --parallel 3 --dry-run --limit 5"
    run_test "Clone with 10 workers" "$BINARY git repo clone --provider github --org gizzahub --parallel 10 --dry-run --limit 5"
    
    # Test sync with parallel workers
    run_test "Sync with parallel workers" "$BINARY git repo sync --from github:gizzahub --to gitlab:test --parallel 5 --dry-run"
}

# Test edge cases
test_edge_cases() {
    log "Testing edge cases..."
    
    # Test with very long parameters
    local long_org=$(printf 'a%.0s' {1..100})
    run_test "Command with long org name" "$BINARY git repo list --provider github --org '$long_org' --dry-run" 1
    
    # Test with special characters
    run_test "Command with special chars" "$BINARY git repo list --provider github --org 'test@org' --dry-run" 1
    
    # Test with empty parameters
    run_test "Command with empty org" "$BINARY git repo list --provider github --org '' --dry-run" 1
}

# Integration tests with real services (if tokens available)
test_integration() {
    log "Testing integration with real services..."
    
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        log "GitHub token found, testing GitHub integration"
        run_test "GitHub list repos" "$BINARY git repo list --provider github --org gizzahub --limit 5"
        success "GitHub integration test passed"
    else
        warning "GITHUB_TOKEN not set, skipping GitHub integration tests"
    fi
    
    if [[ -n "${GITLAB_TOKEN:-}" ]]; then
        log "GitLab token found, testing GitLab integration"
        run_test "GitLab list repos" "$BINARY git repo list --provider gitlab --org gizzahub --limit 5"
        success "GitLab integration test passed"
    else
        warning "GITLAB_TOKEN not set, skipping GitLab integration tests"
    fi
}

# Performance tests
test_performance() {
    log "Testing performance..."
    
    local start_time
    local end_time
    local duration
    
    # Test large repository listing
    start_time=$(date +%s)
    if run_test "Large repo list performance" "$BINARY git repo list --provider github --org microsoft --limit 100 --dry-run"; then
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        log "Large repo list took ${duration}s"
        
        if [[ $duration -gt 10 ]]; then
            warning "Large repo list took longer than expected (${duration}s > 10s)"
        else
            success "Large repo list performance acceptable (${duration}s)"
        fi
    fi
    
    # Test sync with many repositories
    start_time=$(date +%s)
    if run_test "Large sync performance" "$BINARY git repo sync --from github:microsoft --to gitlab:test --dry-run --limit 50"; then
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        log "Large sync test took ${duration}s"
        
        if [[ $duration -gt 5 ]]; then
            warning "Large sync test took longer than expected (${duration}s > 5s)"
        else
            success "Large sync performance acceptable (${duration}s)"
        fi
    fi
}

# Main test execution
main() {
    log "Starting Git Repo E2E Tests"
    log "Test directory: $TEST_DIR"
    log "Binary: $BINARY"
    
    setup_test_environment
    
    # Run test suites
    test_help_commands
    test_command_validation
    test_dry_run_commands
    test_output_formats
    test_configuration
    test_filtering
    test_parallel_execution
    test_edge_cases
    
    # Optional integration tests (require tokens)
    if [[ "${RUN_INTEGRATION_TESTS:-false}" == "true" ]]; then
        test_integration
    fi
    
    # Optional performance tests
    if [[ "${RUN_PERFORMANCE_TESTS:-false}" == "true" ]]; then
        test_performance
    fi
    
    # Print summary
    echo
    log "Test Summary:"
    log "============="
    log "Tests run: $TESTS_RUN"
    log "Passed: $TESTS_PASSED"
    log "Failed: $TESTS_FAILED"
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        success "All tests passed! 🎉"
        echo
        log "Log file: $LOG_FILE"
        cleanup_test_environment
        exit 0
    else
        error "Some tests failed!"
        echo
        log "Log file: $LOG_FILE"
        log "Check the log file for detailed error information"
        exit 1
    fi
}

# Trap signals for cleanup
trap cleanup_test_environment EXIT

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi