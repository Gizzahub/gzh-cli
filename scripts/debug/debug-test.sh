#!/bin/bash

# Debug Go tests with Delve debugger
# Usage: ./scripts/debug/debug-test.sh [package] [test_name]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEBUG_PORT=2346
DEBUG_HOST="127.0.0.1"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

echo -e "${BLUE}🧪 GZH Manager Go Test Debugger${NC}"
echo -e "${BLUE}====================================${NC}"

# Check if delve is installed
if ! command -v dlv &> /dev/null; then
    echo -e "${RED}❌ Delve debugger not found. Installing...${NC}"
    go install github.com/go-delve/delve/cmd/dlv@latest
fi

# Change to project directory
cd "$PROJECT_ROOT"

# Set up environment
export GZH_DEV_MODE=true
export GO111MODULE=on
export GORACE="halt_on_error=1"

# Parse arguments
PACKAGE="./..."
TEST_NAME=""

if [ $# -ge 1 ]; then
    PACKAGE="$1"
fi

if [ $# -ge 2 ]; then
    TEST_NAME="$2"
fi

echo -e "${YELLOW}📌 Package: $PACKAGE${NC}"
if [ -n "$TEST_NAME" ]; then
    echo -e "${YELLOW}🎯 Test: $TEST_NAME${NC}"
else
    echo -e "${YELLOW}🎯 Test: All tests${NC}"
fi

# Validate package exists
if [ "$PACKAGE" != "./..." ] && [ ! -d "$PACKAGE" ]; then
    echo -e "${RED}❌ Package directory '$PACKAGE' not found${NC}"
    exit 1
fi

echo -e "${BLUE}🚀 Starting test debugger...${NC}"
echo -e "${BLUE}   Host: $DEBUG_HOST${NC}"
echo -e "${BLUE}   Port: $DEBUG_PORT${NC}"
echo ""
echo -e "${YELLOW}📋 Debug Commands:${NC}"
echo -e "   ${GREEN}b TestMain${NC}      - Set breakpoint at TestMain"
echo -e "   ${GREEN}b Test*${NC}        - Set breakpoint at test functions"
echo -e "   ${GREEN}c${NC}              - Continue execution"
echo -e "   ${GREEN}n${NC}              - Next line"
echo -e "   ${GREEN}s${NC}              - Step into"
echo -e "   ${GREEN}p <var>${NC}        - Print variable"
echo -e "   ${GREEN}l${NC}              - List source code"
echo -e "   ${GREEN}bt${NC}             - Stack trace"
echo -e "   ${GREEN}goroutines${NC}     - List goroutines"
echo -e "   ${GREEN}q${NC}              - Quit debugger"
echo ""
echo -e "${YELLOW}🌐 Web UI: http://$DEBUG_HOST:$DEBUG_PORT${NC}"
echo ""
echo -e "${BLUE}Press Ctrl+C to stop the debugger${NC}"
echo ""

# Prepare test command arguments
TEST_ARGS=()
if [ -n "$TEST_NAME" ]; then
    TEST_ARGS+=("-test.run" "$TEST_NAME")
fi
TEST_ARGS+=("-test.v")

# Start debugging session
trap 'echo -e "\n${YELLOW}🛑 Test debugging session ended${NC}"' EXIT

# Launch delve in test mode
if [ "$PACKAGE" = "./..." ]; then
    echo -e "${YELLOW}⚠️ Cannot debug all packages at once. Please specify a single package.${NC}"
    echo -e "${YELLOW}Example: ./scripts/debug/debug-test.sh ./cmd/bulk-clone${NC}"
    exit 1
else
    dlv test \
        --listen="$DEBUG_HOST:$DEBUG_PORT" \
        --headless \
        --api-version=2 \
        --accept-multiclient \
        --log \
        --log-output="debugger,gdbwire,lldbout,debuglineerr,rpc" \
        "$PACKAGE" \
        -- "${TEST_ARGS[@]}"
fi