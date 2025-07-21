#!/bin/bash

# Integration Test Runner Script
# This script runs integration tests with proper environment setup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "$YELLOW" "Checking prerequisites..."

    # Check for required environment variables
    if [[ -z "$GITHUB_TOKEN" ]]; then
        print_status "$RED" "Error: GITHUB_TOKEN environment variable is not set"
        echo "Please set your GitHub personal access token:"
        echo "  export GITHUB_TOKEN='your-token-here'"
        return 1
    fi

    if [[ -z "$GITHUB_TEST_ORG" ]]; then
        print_status "$RED" "Error: GITHUB_TEST_ORG environment variable is not set"
        echo "Please set your test organization name:"
        echo "  export GITHUB_TEST_ORG='your-test-org'"
        return 1
    fi

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_status "$RED" "Error: Go is not installed"
        return 1
    fi

    print_status "$GREEN" "✓ All prerequisites met"
    return 0
}

# Function to run tests with coverage
run_tests_with_coverage() {
    local test_path=$1
    local coverage_file="${PROJECT_ROOT}/coverage/integration.out"

    mkdir -p "${PROJECT_ROOT}/coverage"

    print_status "$YELLOW" "Running integration tests with coverage..."

    cd "$PROJECT_ROOT"
    go test -v -coverprofile="$coverage_file" -covermode=atomic "$test_path" || return 1

    # Generate HTML coverage report
    go tool cover -html="$coverage_file" -o "${PROJECT_ROOT}/coverage/integration.html"

    print_status "$GREEN" "✓ Coverage report generated: coverage/integration.html"
}

# Function to run specific test
run_specific_test() {
    local test_name=$1
    local test_path="./test/integration/..."

    print_status "$YELLOW" "Running test: $test_name"

    cd "$PROJECT_ROOT"
    go test -v "$test_path" -run "$test_name" || return 1
}

# Function to run all integration tests
run_all_tests() {
    local test_path="./test/integration/..."

    print_status "$YELLOW" "Running all integration tests..."

    cd "$PROJECT_ROOT"
    go test -v -timeout 30m "$test_path" || return 1
}

# Function to run tests with race detection
run_tests_with_race() {
    local test_path="./test/integration/..."

    print_status "$YELLOW" "Running integration tests with race detection..."

    cd "$PROJECT_ROOT"
    go test -v -race "$test_path" || return 1
}

# Function to clean test artifacts
clean_test_artifacts() {
    print_status "$YELLOW" "Cleaning test artifacts..."

    # Remove temporary test directories
    rm -rf /tmp/repo-config-integration-*
    rm -rf /tmp/github-integration-test-*

    print_status "$GREEN" "✓ Test artifacts cleaned"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  all              Run all integration tests"
    echo "  specific <name>  Run specific test by name"
    echo "  coverage         Run tests with coverage report"
    echo "  race             Run tests with race detection"
    echo "  clean            Clean test artifacts"
    echo "  help             Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  GITHUB_TOKEN     GitHub personal access token (required)"
    echo "  GITHUB_TEST_ORG  GitHub test organization name (required)"
    echo ""
    echo "Examples:"
    echo "  $0 all"
    echo "  $0 specific TestIntegration_RepoConfig_EndToEnd"
    echo "  $0 coverage"
}

# Main execution
main() {
    local command=${1:-all}

    case "$command" in
        all)
            check_prerequisites || exit 1
            run_all_tests || exit 1
            print_status "$GREEN" "✓ All integration tests passed"
            ;;
        specific)
            if [[ -z "$2" ]]; then
                print_status "$RED" "Error: Test name required"
                echo "Usage: $0 specific <test-name>"
                exit 1
            fi
            check_prerequisites || exit 1
            run_specific_test "$2" || exit 1
            print_status "$GREEN" "✓ Test passed: $2"
            ;;
        coverage)
            check_prerequisites || exit 1
            run_tests_with_coverage "./test/integration/..." || exit 1
            ;;
        race)
            check_prerequisites || exit 1
            run_tests_with_race || exit 1
            print_status "$GREEN" "✓ No race conditions detected"
            ;;
        clean)
            clean_test_artifacts
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            print_status "$RED" "Error: Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
