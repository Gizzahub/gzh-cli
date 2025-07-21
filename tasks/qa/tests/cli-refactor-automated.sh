#!/bin/bash
# CLI 중심 재편 기능 자동화 테스트
# 자동화 가능한 62.5% 시나리오 테스트

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test result tracking
declare -a FAILED_TEST_NAMES=()

# Helper functions
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_exit_code="${3:-0}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "\n🧪 Testing: $test_name"

    if eval "$test_command"; then
        if [ "$expected_exit_code" -eq 0 ]; then
            echo -e "${GREEN}✓ PASSED${NC}: $test_name"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}✗ FAILED${NC}: $test_name (expected to fail, but succeeded)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_TEST_NAMES+=("$test_name")
        fi
    else
        local exit_code=$?
        if [ "$exit_code" -eq "$expected_exit_code" ]; then
            echo -e "${GREEN}✓ PASSED${NC}: $test_name (correctly failed with code $exit_code)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}✗ FAILED${NC}: $test_name (exit code: $exit_code, expected: $expected_exit_code)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_TEST_NAMES+=("$test_name")
        fi
    fi
}

# Create temporary test directory
TEST_DIR=$(mktemp -d)
echo "📁 Test directory: $TEST_DIR"

# Create test configuration file
cat > "$TEST_DIR/test-config.yaml" << 'EOF'
version: 1.0
network_profiles:
  office:
    dns_servers:
      - 10.0.0.1
      - 10.0.0.2
    proxy:
      http: "http://proxy.company.com:8080"
      https: "https://proxy.company.com:8080"
      no_proxy: "localhost,127.0.0.1,10.0.0.0/8"
    vpn:
      name: "office-vpn"
      server: "vpn.company.com"
    hosts:
      - "10.0.1.100 internal.company.com"
      - "10.0.1.101 gitlab.company.com"

  home:
    dns_servers:
      - 1.1.1.1
      - 8.8.8.8
    proxy: none
    vpn: none
    hosts: []

  mobile:
    dns_servers:
      - 9.9.9.9
      - 149.112.112.112
    proxy: none
    vpn:
      name: "personal-vpn"
      server: "vpn.personal.com"
    hosts: []

hooks:
  pre_switch:
    - "/usr/local/bin/pre-switch-hook.sh"
  post_switch:
    - "/usr/local/bin/post-switch-hook.sh"
EOF

# Create mock hook scripts
cat > "$TEST_DIR/mock-hook.sh" << 'EOF'
#!/bin/bash
echo "Hook executed: $0 with profile: $1"
exit 0
EOF
chmod +x "$TEST_DIR/mock-hook.sh"

echo -e "\n🚀 Starting CLI Refactor Functional Tests"

# Test 1: Build check
run_test "Build gz binary" "go build -o $TEST_DIR/gz ./cmd"

# Use the test binary for remaining tests
GZ="$TEST_DIR/gz"

# Test 2: Basic command structure
run_test "net-env command exists" "$GZ net-env --help"

# Test 3: Status command
run_test "net-env status command" "$GZ net-env status"

# Test 4: Switch command with help
run_test "net-env switch help" "$GZ net-env switch --help"

# Test 5: List profiles from config
run_test "List available profiles" "$GZ net-env --config $TEST_DIR/test-config.yaml list-profiles"

# Test 6: Validate config file
run_test "Validate configuration file" "$GZ net-env --config $TEST_DIR/test-config.yaml validate"

# Test 7: Switch to non-existent profile (should fail)
run_test "Switch to non-existent profile" "$GZ net-env --config $TEST_DIR/test-config.yaml switch --profile nonexistent" 1

# Test 8: Dry-run switch
run_test "Dry-run profile switch" "$GZ net-env --config $TEST_DIR/test-config.yaml switch --profile office --dry-run"

# Test 9: Check for daemon processes
run_test "No daemon processes running" "! pgrep -f 'gz.*daemon|gzh-manager.*daemon'"

# Test 10: Check systemd service doesn't exist
run_test "No systemd service" "! systemctl list-unit-files | grep -q gzh-manager"

# Test 11: YAML schema validation
cat > "$TEST_DIR/invalid-config.yaml" << 'EOF'
version: 2.0
invalid_field: true
daemon_mode: enabled  # This should not be allowed
EOF

run_test "Invalid config rejection" "$GZ net-env --config $TEST_DIR/invalid-config.yaml validate" 1

# Test 12: Hook execution simulation
run_test "Hook system test" "$GZ net-env --config $TEST_DIR/test-config.yaml switch --profile home --hook $TEST_DIR/mock-hook.sh --dry-run"

# Test 13: Multiple config file precedence
export GZH_NET_CONFIG="$TEST_DIR/test-config.yaml"
run_test "Environment variable config" "$GZ net-env list-profiles"

# Test 14: Command output format
run_test "JSON output format" "$GZ net-env status --output json | jq ."

# Test 15: Verbose mode
run_test "Verbose output" "$GZ net-env --verbose status"

# Test 16: Permission error simulation
touch "$TEST_DIR/readonly-config.yaml"
chmod 000 "$TEST_DIR/readonly-config.yaml"
run_test "Permission error handling" "$GZ net-env --config $TEST_DIR/readonly-config.yaml status" 1
chmod 644 "$TEST_DIR/readonly-config.yaml"

# Test 17: Profile completion
run_test "Profile name completion" "$GZ net-env completion bash | grep -q 'profile'"

# Test 18: Resource usage check
run_test "Memory usage baseline" "ps aux | grep gz | awk '{print \$6}' | head -1 | xargs -I {} test {} -lt 50000"

# Test 19: Config migration check
cat > "$TEST_DIR/old-config.yaml" << 'EOF'
version: 0.9
daemon:
  enabled: true
  interval: 30s
profiles:
  office:
    dns: ["10.0.0.1"]
EOF

run_test "Old config format warning" "$GZ net-env --config $TEST_DIR/old-config.yaml validate 2>&1 | grep -i 'deprecated\\|warning'"

# Test 20: Error message clarity
run_test "Clear error messages" "$GZ net-env switch --profile 2>&1 | grep -E 'Usage|required|missing'"

echo -e "\n📊 Test Summary"
echo "================"
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ ${#FAILED_TEST_NAMES[@]} -gt 0 ]; then
    echo -e "\n${RED}Failed Tests:${NC}"
    for test_name in "${FAILED_TEST_NAMES[@]}"; do
        echo -e "  - $test_name"
    done
fi

# Cleanup
rm -rf "$TEST_DIR"

# Exit with appropriate code
if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✨ All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed${NC}"
    exit 1
fi
