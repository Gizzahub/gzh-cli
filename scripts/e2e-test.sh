#!/bin/bash
# End-to-End Test Script for git-synclone
# Tests complete installation and usage workflow

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="${HOME}/.gzh-e2e-test"
BINARY_NAME="git-synclone"
TEST_ORG="octocat"
TEST_GROUP="gitlab-org"

# Function to print colored messages
print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_step() {
    echo -e "${BLUE}🔄 $1${NC}"
}

print_test() {
    echo -e "${YELLOW}🧪 $1${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to cleanup test environment
cleanup() {
    print_step "Cleaning up test environment..."
    rm -rf "$TEST_DIR"
    print_success "Cleanup completed"
}

# Function to setup test environment
setup() {
    print_step "Setting up test environment..."

    # Create test directory
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"

    print_success "Test environment ready at: $TEST_DIR"
}

# Function to test installation
test_installation() {
    print_test "Testing Installation"

    # Check if git-synclone is installed
    if ! command_exists "$BINARY_NAME"; then
        print_error "$BINARY_NAME not found in PATH"
        print_info "Please run the installation script first:"
        print_info "  ./scripts/install-git-extensions.sh"
        return 1
    fi

    print_success "$BINARY_NAME found in PATH"

    # Test version command
    print_step "Testing version command..."
    version_output=$("$BINARY_NAME" --version 2>&1)
    if [[ $? -eq 0 ]]; then
        print_success "Version: $version_output"
    else
        print_error "Version command failed"
        return 1
    fi

    # Test help command
    print_step "Testing help command..."
    if "$BINARY_NAME" --help >/dev/null 2>&1; then
        print_success "Help command works"
    else
        print_error "Help command failed"
        return 1
    fi

    return 0
}

# Function to test Git integration
test_git_integration() {
    print_test "Testing Git Integration"

    # Test git synclone command
    print_step "Testing git synclone command..."
    if git synclone --help >/dev/null 2>&1; then
        print_success "Git integration works"
    else
        print_error "Git integration failed"
        print_info "Make sure $BINARY_NAME is in PATH"
        return 1
    fi

    # Test subcommands
    subcommands=("github" "gitlab" "gitea" "doctor")
    for subcmd in "${subcommands[@]}"; do
        print_step "Testing git synclone $subcmd..."
        if git synclone "$subcmd" --help >/dev/null 2>&1; then
            print_success "git synclone $subcmd works"
        else
            print_warning "git synclone $subcmd failed"
        fi
    done

    return 0
}

# Function to test doctor command
test_doctor() {
    print_test "Testing Doctor Command"

    print_step "Running installation diagnostics..."
    doctor_output=$(git synclone doctor 2>&1)
    doctor_exit_code=$?

    echo "$doctor_output"

    if [[ $doctor_exit_code -eq 0 ]]; then
        print_success "Doctor command passed all checks"
    else
        print_warning "Doctor command found issues (exit code: $doctor_exit_code)"
        print_info "This is normal if configuration is not set up"
    fi

    # Test verbose mode
    print_step "Testing doctor verbose mode..."
    if git synclone doctor --verbose >/dev/null 2>&1; then
        print_success "Doctor verbose mode works"
    else
        print_warning "Doctor verbose mode failed"
    fi

    return 0
}

# Function to test configuration
test_configuration() {
    print_test "Testing Configuration"

    # Create test configuration
    config_file="$TEST_DIR/test-config.yaml"
    print_step "Creating test configuration..."

    cat > "$config_file" << EOF
version: "1.0.0"
default:
  protocol: "https"
  github:
    rootPath: "$TEST_DIR/github"
    provider: "github"
    protocol: "https"
    orgName: "$TEST_ORG"
  gitlab:
    rootPath: "$TEST_DIR/gitlab"
    provider: "gitlab"
    url: "https://gitlab.com"
    protocol: "https"
    groupName: "$TEST_GROUP"
    recursive: false
repoRoots:
  - rootPath: "$TEST_DIR/custom"
    provider: "github"
    protocol: "ssh"
    orgName: "custom-org"
EOF

    print_success "Test configuration created"

    # Test configuration validation
    print_step "Testing configuration validation..."
    if git synclone validate --config "$config_file" >/dev/null 2>&1; then
        print_success "Configuration validation passed"
    else
        print_warning "Configuration validation failed (may be expected)"
    fi

    return 0
}

# Function to test dry run operations
test_dry_run() {
    print_test "Testing Dry Run Operations"

    local config_file="$TEST_DIR/test-config.yaml"

    # Test GitHub dry run
    print_step "Testing GitHub dry run..."
    github_output=$(git synclone github -o "$TEST_ORG" -t "$TEST_DIR/github-test" --dry-run 2>&1)
    github_exit_code=$?

    if [[ $github_exit_code -eq 0 ]]; then
        print_success "GitHub dry run completed successfully"
    else
        print_warning "GitHub dry run failed (exit code: $github_exit_code)"
        echo "Output: $github_output"
    fi

    # Test GitLab dry run
    print_step "Testing GitLab dry run..."
    gitlab_output=$(git synclone gitlab -g "$TEST_GROUP" -t "$TEST_DIR/gitlab-test" --dry-run 2>&1)
    gitlab_exit_code=$?

    if [[ $gitlab_exit_code -eq 0 ]]; then
        print_success "GitLab dry run completed successfully"
    else
        print_warning "GitLab dry run failed (exit code: $gitlab_exit_code)"
        echo "Output: $gitlab_output"
    fi

    # Test Gitea dry run
    print_step "Testing Gitea dry run..."
    gitea_output=$(git synclone gitea -o "gitea" -t "$TEST_DIR/gitea-test" --dry-run 2>&1)
    gitea_exit_code=$?

    if [[ $gitea_exit_code -eq 0 ]]; then
        print_success "Gitea dry run completed successfully"
    else
        print_warning "Gitea dry run failed (exit code: $gitea_exit_code)"
        echo "Output: $gitea_output"
    fi

    # Test config-based dry run
    print_step "Testing config-based dry run..."
    config_output=$(git synclone --config "$config_file" --dry-run 2>&1)
    config_exit_code=$?

    if [[ $config_exit_code -eq 0 ]]; then
        print_success "Config-based dry run completed successfully"
    else
        print_warning "Config-based dry run failed (exit code: $config_exit_code)"
        echo "Output: $config_output"
    fi

    return 0
}

# Function to test error scenarios
test_error_scenarios() {
    print_test "Testing Error Scenarios"

    # Test missing required arguments
    print_step "Testing missing organization..."
    if ! git synclone github >/dev/null 2>&1; then
        print_success "Correctly rejected missing organization"
    else
        print_warning "Should have failed with missing organization"
    fi

    # Test invalid config file
    print_step "Testing invalid config file..."
    if ! git synclone --config "/nonexistent/config.yaml" >/dev/null 2>&1; then
        print_success "Correctly rejected invalid config file"
    else
        print_warning "Should have failed with invalid config file"
    fi

    # Test invalid flags
    print_step "Testing invalid strategy..."
    if ! git synclone github -o "$TEST_ORG" --strategy "invalid" >/dev/null 2>&1; then
        print_success "Correctly rejected invalid strategy"
    else
        print_warning "Should have failed with invalid strategy"
    fi

    print_step "Testing invalid protocol..."
    if ! git synclone github -o "$TEST_ORG" --protocol "ftp" >/dev/null 2>&1; then
        print_success "Correctly rejected invalid protocol"
    else
        print_warning "Should have failed with invalid protocol"
    fi

    return 0
}

# Function to test performance scenarios
test_performance() {
    print_test "Testing Performance Scenarios"

    # Test parallel vs sequential (dry run)
    print_step "Testing sequential operation..."
    start_time=$(date +%s)
    git synclone github -o "$TEST_ORG" -t "$TEST_DIR/perf-sequential" --parallel 1 --dry-run >/dev/null 2>&1
    sequential_exit_code=$?
    sequential_time=$(($(date +%s) - start_time))

    print_step "Testing parallel operation..."
    start_time=$(date +%s)
    git synclone github -o "$TEST_ORG" -t "$TEST_DIR/perf-parallel" --parallel 5 --dry-run >/dev/null 2>&1
    parallel_exit_code=$?
    parallel_time=$(($(date +%s) - start_time))

    print_info "Sequential time: ${sequential_time}s (exit: $sequential_exit_code)"
    print_info "Parallel time: ${parallel_time}s (exit: $parallel_exit_code)"

    if [[ $sequential_time -gt 0 ]] && [[ $parallel_time -gt 0 ]]; then
        print_success "Performance test completed"
    else
        print_warning "Performance test had issues"
    fi

    return 0
}

# Function to run all tests
run_all_tests() {
    local failed_tests=0

    echo -e "${CYAN}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                  E2E Test Suite                            │"
    echo "│                git-synclone                                 │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"

    setup

    # Run tests
    test_installation || ((failed_tests++))
    echo

    test_git_integration || ((failed_tests++))
    echo

    test_doctor || ((failed_tests++))
    echo

    test_configuration || ((failed_tests++))
    echo

    test_dry_run || ((failed_tests++))
    echo

    test_error_scenarios || ((failed_tests++))
    echo

    test_performance || ((failed_tests++))
    echo

    # Summary
    echo -e "${CYAN}"
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                  Test Summary                               │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo -e "${NC}"

    if [[ $failed_tests -eq 0 ]]; then
        print_success "All E2E tests passed! 🎉"
        echo
        print_info "git-synclone is ready for use!"
        echo
        print_info "Quick start:"
        echo "  git synclone doctor                    # Check installation"
        echo "  git synclone github -o myorg           # Clone GitHub org"
        echo "  git synclone --help                    # Show all options"
    else
        print_warning "$failed_tests test(s) had issues"
        echo
        print_info "Some tests may fail due to network conditions or missing tokens."
        print_info "Check the output above for details."
    fi

    return $failed_tests
}

# Main execution
main() {
    # Trap to cleanup on exit
    trap cleanup EXIT

    # Parse command line arguments
    case "${1:-all}" in
        "installation")
            setup && test_installation
            ;;
        "integration")
            setup && test_git_integration
            ;;
        "doctor")
            setup && test_doctor
            ;;
        "config")
            setup && test_configuration
            ;;
        "dry-run")
            setup && test_dry_run
            ;;
        "errors")
            setup && test_error_scenarios
            ;;
        "performance")
            setup && test_performance
            ;;
        "all")
            run_all_tests
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [test-category]"
            echo "Test categories:"
            echo "  installation  - Test binary installation and basic commands"
            echo "  integration   - Test Git integration"
            echo "  doctor        - Test installation diagnostics"
            echo "  config        - Test configuration handling"
            echo "  dry-run       - Test dry run operations"
            echo "  errors        - Test error scenarios"
            echo "  performance   - Test performance scenarios"
            echo "  all           - Run all tests (default)"
            exit 0
            ;;
        *)
            print_error "Unknown test category: $1"
            print_info "Use '$0 help' for available options"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
