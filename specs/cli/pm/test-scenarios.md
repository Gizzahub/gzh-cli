# PM Update Command Test Scenarios

## Test Categories

### 1. Basic Functionality Tests

#### Test 1.1: Simple Update Success

```bash
# Setup: System with brew and npm installed
gz pm update --all --dry-run
# Expected: Shows update plan for all detected managers
# Verify: Output contains manager overview, update actions
```

#### Test 1.2: Single Manager Update

```bash
# Setup: System with multiple managers
gz pm update --manager brew --dry-run
# Expected: Updates only Homebrew
# Verify: Other managers not processed
```

#### Test 1.3: Multiple Specific Managers

```bash
# Setup: System with brew, asdf, npm
gz pm update --managers brew,npm --dry-run
# Expected: Updates only specified managers
# Verify: asdf not processed, brew and npm processed
```

### 2. Output Format Tests

#### Test 2.1: Text Output (Default)

```bash
gz pm update --all --dry-run
# Expected: Human-readable text with emojis and formatting
# Verify: Contains sections, progress indicators, color coding
```

#### Test 2.2: JSON Output Format

```bash
gz pm update --all --dry-run --output json
# Expected: Valid JSON structure with update details
# Verify: Can be parsed with jq, contains all required fields
```

#### Test 2.3: Output Consistency

```bash
# Compare text vs JSON output for same operation
text_result=$(gz pm update --manager brew --dry-run)
json_result=$(gz pm update --manager brew --dry-run --output json)
# Verify: Same managers detected, same actions planned
```

### 3. Platform-Specific Tests

#### Test 3.1: macOS Package Managers

```bash
# Setup: macOS system
gz pm update --all --dry-run
# Expected: Detects brew, asdf, npm, pip, sdkman
# Verify: apt/pacman marked as unsupported
```

#### Test 3.2: Linux (Ubuntu) Package Managers

```bash
# Setup: Ubuntu system
gz pm update --all --dry-run
# Expected: Detects apt, asdf, npm, pip
# Verify: brew marked as supported but may not be installed
```

#### Test 3.3: Linux (Arch) Package Managers

```bash
# Setup: Arch Linux system
gz pm update --all --dry-run
# Expected: Detects pacman, yay, asdf, npm, pip
# Verify: apt marked as unsupported
```

### 4. Error Handling Tests

#### Test 4.1: Network Connectivity Issues

```bash
# Setup: Block network access to package repositories
# Simulate: DNS failure, firewall blocking, proxy issues
gz pm update --manager brew
# Expected: Clear error messages, retry suggestions
# Verify: Graceful degradation, other managers continue
```

#### Test 4.2: Permission Denied

```bash
# Setup: Run as non-privileged user
gz pm update --manager apt
# Expected: Permission error with clear fix instructions
# Verify: Suggests sudo usage or permission fixes
```

#### Test 4.3: Insufficient Disk Space

```bash
# Setup: System with very low disk space
gz pm update --manager brew
# Expected: Disk space error with cleanup suggestions
# Verify: Provides specific space requirements
```

### 5. Environment Detection Tests

#### Test 5.1: Conda Environment Active

```bash
# Setup: Activate conda environment
conda activate myproject
gz pm update --manager pip
# Expected: Detects conda, recommends using conda/mamba instead
# Verify: Provides override option with --pip-allow-conda
```

#### Test 5.2: Virtual Environment Active

```bash
# Setup: Python virtual environment active
source venv/bin/activate
gz pm update --manager pip --dry-run
# Expected: Detects virtual env, adjusts pip strategy
# Verify: Uses correct pip executable
```

#### Test 5.3: Multiple Node Version Managers

```bash
# Setup: Both nvm and asdf with node installed
gz pm update --all --check-duplicates
# Expected: Detects node version conflicts
# Verify: Shows duplicate binary warnings
```

### 6. Package Manager Specific Tests

#### Test 6.1: Homebrew Edge Cases

```bash
# Test outdated brew version
# Test formula conflicts
# Test cask updates
# Test cleanup after updates
```

#### Test 6.2: ASDF Plugin Management

```bash
# Test plugin updates
# Test version installation
# Test post-install hooks
# Test compatibility filters
```

#### Test 6.3: NPM Global Packages

```bash
# Test global package updates
# Test npm cache issues
# Test permission problems with global installs
```

#### Test 6.4: Python Package Management

```bash
# Test pip self-update
# Test package dependency conflicts
# Test system vs user packages
```

### 7. Dry Run vs Real Execution

#### Test 7.1: Dry Run Accuracy

```bash
# Run dry-run, capture planned actions
dry_result=$(gz pm update --all --dry-run)
# Execute real update with same conditions
real_result=$(gz pm update --all)
# Verify: Actual actions match dry-run predictions
```

#### Test 7.2: No Changes When Up-to-Date

```bash
# Setup: All packages already latest
gz pm update --all
# Expected: Shows "already latest" messages
# Verify: No actual downloads or installations
```

### 8. Configuration and Strategy Tests

#### Test 8.1: Update Strategy - Latest

```bash
gz pm update --all --strategy latest
# Expected: Updates to absolute latest versions
# Verify: Includes beta/rc versions where applicable
```

#### Test 8.2: Update Strategy - Stable

```bash
gz pm update --all --strategy stable
# Expected: Updates to latest stable versions only
# Verify: Excludes beta/rc versions
```

#### Test 8.3: Update Strategy - Minor

```bash
gz pm update --all --strategy minor
# Expected: Updates to latest patch/minor versions
# Verify: No major version upgrades
```

### 9. Performance and Resource Tests

#### Test 9.1: Large Package Set

```bash
# Setup: System with 100+ packages to update
time gz pm update --all --dry-run
# Expected: Completes within reasonable time
# Verify: Memory usage stays reasonable
```

#### Test 9.2: Parallel Processing

```bash
# Test concurrent manager updates
# Verify no conflicts between managers
# Check resource utilization
```

#### Test 9.3: Progress Indication

```bash
# Test long-running updates show progress
# Verify user can see current operation
# Check time estimates are reasonable
```

### 10. Recovery and Rollback Tests

#### Test 10.1: Interrupted Update

```bash
# Start update, kill process mid-way
gz pm update --all &
sleep 30; kill %1
# Restart update
gz pm update --all
# Expected: Continues from safe state
# Verify: No corrupted package states
```

#### Test 10.2: Failed Package Installation

```bash
# Setup: Package with installation failure
# Expected: Other packages continue updating
# Verify: Clear error reporting for failed package
```

### 11. Integration Tests

#### Test 11.1: Full System Update Workflow

```bash
# Test complete workflow:
# 1. Status check
gz pm status
# 2. Update plan
gz pm update --all --dry-run
# 3. Execute update
gz pm update --all
# 4. Verify results
gz pm status
```

#### Test 11.2: Multi-User Environment

```bash
# Test updates in shared system
# Verify user-specific vs system packages
# Check permission handling
```

### 12. Regression Tests

#### Test 12.1: Version Detection Accuracy

```bash
# Verify version parsing for all supported managers
# Test edge cases: pre-release, custom builds
# Check version comparison logic
```

#### Test 12.2: Configuration File Handling

```bash
# Test various config file formats
# Verify backup and restore mechanisms
# Check compatibility across versions
```

## Test Execution Framework

### Automated Test Runner

```bash
#!/bin/bash
# test-pm-update.sh - Automated test runner for PM update

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_exit_code="${3:-0}"
    
    echo "Running: $test_name"
    eval "$test_cmd"
    local actual_exit_code=$?
    
    if [ $actual_exit_code -eq $expected_exit_code ]; then
        echo "✅ PASS: $test_name"
        return 0
    else
        echo "❌ FAIL: $test_name (exit code: $actual_exit_code, expected: $expected_exit_code)"
        return 1
    fi
}

# Run all test categories
run_basic_tests
run_platform_tests
run_error_handling_tests
run_environment_tests
# ... etc
```

### Test Data Setup

```yaml
# test-fixtures.yml - Test environment configurations
environments:
  macos_homebrew:
    platform: darwin
    managers: [brew, asdf, npm, pip]
    packages:
      brew: [git, node, python]
      asdf: [nodejs, python, golang]
      
  ubuntu_apt:
    platform: linux
    managers: [apt, asdf, npm, pip]
    packages:
      apt: [git, build-essential]
      
  arch_pacman:
    platform: linux
    managers: [pacman, yay, asdf]
    packages:
      pacman: [git, base-devel]
```

### Expected Results Validation

```bash
# validate-results.sh - Validate test results against expected outcomes

validate_output_format() {
    local output="$1"
    local format="$2"
    
    case $format in
        "text")
            grep -q "🔄 Updating" <<< "$output"
            grep -q "📊 Summary" <<< "$output"
            ;;
        "json")
            jq . <<< "$output" >/dev/null
            jq -r '.managers[].name' <<< "$output"
            ;;
    esac
}

validate_manager_detection() {
    local platform="$1"
    local output="$2"
    
    case $platform in
        "darwin")
            grep -q "brew.*✅" <<< "$output"
            grep -q "apt.*🚫" <<< "$output"
            ;;
        "linux-ubuntu")
            grep -q "apt.*✅" <<< "$output"
            ;;
    esac
}
```

## Test Environment Setup

### Docker Test Containers

```dockerfile
# Dockerfile.test-ubuntu - Ubuntu test environment
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    curl git build-essential \
    python3 python3-pip \
    nodejs npm

# Install asdf
RUN git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.14.0

# Install Homebrew (Linuxbrew)
RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

COPY gz /usr/local/bin/gz
RUN chmod +x /usr/local/bin/gz

CMD ["bash"]
```

### CI/CD Integration

```yaml
# .github/workflows/pm-tests.yml
name: PM Update Tests

on: [push, pull_request]

jobs:
  test-pm-update:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup test environment
      run: |
        # Install package managers for testing
        # Setup test fixtures
        
    - name: Run PM update tests
      run: |
        make test-pm-update
        
    - name: Upload test results
      uses: actions/upload-artifact@v3
      with:
        name: pm-test-results-${{ matrix.os }}
        path: test-results/
```

This comprehensive test suite covers all the major scenarios and edge cases identified in the enhanced specification, ensuring robust validation of the PM update functionality.
