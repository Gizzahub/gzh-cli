#!/bin/bash

# Test What Works - Run tests for components that can be tested independently
# This script tests individual packages and components that don't require the full binary

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
RESULTS_FILE="${SCRIPT_DIR}/component-test-results.md"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Initialize results
echo "# Component Test Results" > "$RESULTS_FILE"
echo "Run Date: $(date)" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

cd "$PROJECT_ROOT"

echo -e "${GREEN}Running Component Tests${NC}"
echo ""

# Function to run go test for a package
test_package() {
    local package="$1"
    local name="$2"
    
    echo -e "${BLUE}Testing $name${NC}"
    echo "## $name" >> "$RESULTS_FILE"
    
    if go test -v "./$package" -count=1 2>&1 | tee -a "$RESULTS_FILE"; then
        echo -e "${GREEN}✅ PASSED${NC}"
        echo "Status: ✅ PASSED" >> "$RESULTS_FILE"
    else
        echo -e "${RED}❌ FAILED${NC}"
        echo "Status: ❌ FAILED" >> "$RESULTS_FILE"
    fi
    echo "" >> "$RESULTS_FILE"
}

# Test individual packages that should work
echo -e "${YELLOW}=== Testing Core Packages ===${NC}"

# Config package tests
test_package "pkg/config" "Configuration Package"
test_package "internal/config" "Internal Config"

# Utility packages
test_package "internal/utils" "Utilities"
test_package "internal/testlib" "Test Library"

# Command packages that have tests
test_package "cmd/always-latest" "Always Latest Command"
test_package "cmd/config" "Config Command"
test_package "cmd/ide" "IDE Command"

# Bulk clone package
test_package "pkg/bulk-clone" "Bulk Clone Package"

# Performance and monitoring
test_package "pkg/memory" "Memory Management"
test_package "pkg/cache" "Cache Package"

echo ""
echo -e "${YELLOW}=== Testing Integration Components ===${NC}"

# Test environment detection
echo -e "${BLUE}Testing Environment Detection${NC}"
if go test -v "./internal/testlib" -run TestEnvironmentDetection 2>&1; then
    echo -e "${GREEN}✅ Environment detection working${NC}"
else
    echo -e "${YELLOW}⚠️  Environment detection needs setup${NC}"
fi

echo ""
echo -e "${YELLOW}=== Checking Build Issues ===${NC}"

# Identify which packages have compilation errors
echo "## Compilation Issues" >> "$RESULTS_FILE"
echo -e "${BLUE}Checking compilation errors...${NC}"

PROBLEM_PACKAGES=(
    "pkg/github"
    "cmd/repo-sync" 
    "cmd/net-env"
)

for pkg in "${PROBLEM_PACKAGES[@]}"; do
    echo -e "${YELLOW}Checking $pkg for errors...${NC}"
    if ! go build -o /dev/null "./$pkg" 2>&1; then
        echo -e "${RED}❌ $pkg has compilation errors${NC}"
        echo "- $pkg: Compilation errors" >> "$RESULTS_FILE"
    fi
done

echo "" >> "$RESULTS_FILE"

# Create a fixes needed file
cat > "${SCRIPT_DIR}/fixes-needed.md" << 'EOF'
# Fixes Needed for Full QA Testing

## Compilation Errors to Fix:

### 1. pkg/github package:
- Missing methods on EventProcessor interface
- Unused variable 'cloneStats' 
- Missing fields on RepositoryInfo struct (Topics, Visibility, IsTemplate)
- Undefined strings.Dir function
- Type mismatches with RepositoryOperationResult

### 2. cmd/repo-sync package:
- Missing dependency parser functions
- Undefined variables and imports
- Interface implementation mismatches

### 3. cmd/net-env package:
- Duplicate function declarations
- Method redeclarations

## Quick Fix Commands:

```bash
# Fix unused variables
sed -i 's/cloneStats/_ \/\/ cloneStats/g' pkg/github/cached_client.go

# Add missing imports
# Add to files missing imports:
# import "path/filepath"
# import "sort"

# Fix interface implementations
# Review and update interface methods to match expected signatures
```

## To Run Full Tests After Fixes:

```bash
# 1. Fix compilation errors above
# 2. Build the binary
make build

# 3. Run full automated tests
./tasks/qa/run_automated_tests.sh
```
EOF

echo ""
echo -e "${GREEN}=== Test Summary ===${NC}"
echo "Component test results saved to: $RESULTS_FILE"
echo "Fixes needed documented in: ${SCRIPT_DIR}/fixes-needed.md"

# Count successes
PASSED=$(grep -c "✅ PASSED" "$RESULTS_FILE" || true)
FAILED=$(grep -c "❌ FAILED" "$RESULTS_FILE" || true)

echo -e "Component Tests - Passed: ${GREEN}$PASSED${NC}, Failed: ${RED}$FAILED${NC}"

echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Review ${SCRIPT_DIR}/fixes-needed.md for compilation errors"
echo "2. Fix the compilation errors"
echo "3. Run 'make build' to create the gz binary"
echo "4. Run './tasks/qa/run_automated_tests.sh' for full testing"
echo "5. For manual tests, see /tasks/qa/manual/"