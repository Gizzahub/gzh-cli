#!/bin/bash

# Debug GZH CLI with Delve debugger
# Usage: ./scripts/debug/debug-cli.sh [command] [args...]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEBUG_PORT=2345
DEBUG_HOST="127.0.0.1"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

echo -e "${BLUE}🐛 GZH Manager Go Debugger${NC}"
echo -e "${BLUE}================================${NC}"

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

# Build with debug symbols
echo -e "${YELLOW}🔨 Building with debug symbols...${NC}"
go build -gcflags="-N -l" -o gz-debug main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build successful${NC}"

# Prepare debug command
if [ $# -eq 0 ]; then
    # Default: debug help command
    DEBUG_ARGS=("--help")
    echo -e "${YELLOW}💡 No arguments provided, debugging with --help${NC}"
else
    DEBUG_ARGS=("$@")
    echo -e "${YELLOW}💡 Debugging with arguments: ${DEBUG_ARGS[*]}${NC}"
fi

echo -e "${BLUE}🚀 Starting Delve debugger...${NC}"
echo -e "${BLUE}   Host: $DEBUG_HOST${NC}"
echo -e "${BLUE}   Port: $DEBUG_PORT${NC}"
echo -e "${BLUE}   Args: ${DEBUG_ARGS[*]}${NC}"
echo ""
echo -e "${YELLOW}📋 Debug Commands:${NC}"
echo -e "   ${GREEN}b main.main${NC}     - Set breakpoint at main function"
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

# Start debugging session
trap 'echo -e "\n${YELLOW}🛑 Debugging session ended${NC}"; rm -f gz-debug' EXIT

# Launch delve in debug mode
dlv debug \
    --listen="$DEBUG_HOST:$DEBUG_PORT" \
    --headless \
    --api-version=2 \
    --accept-multiclient \
    --log \
    --log-output="debugger,gdbwire,lldbout,debuglineerr,rpc" \
    -- "${DEBUG_ARGS[@]}"
