#!/bin/bash

# Attach Delve debugger to running GZH process
# Usage: ./scripts/debug/debug-attach.sh [pid]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEBUG_PORT=2347
DEBUG_HOST="127.0.0.1"

echo -e "${BLUE}🔗 GZH Manager Go Process Debugger${NC}"
echo -e "${BLUE}====================================${NC}"

# Check if delve is installed
if ! command -v dlv &> /dev/null; then
    echo -e "${RED}❌ Delve debugger not found. Installing...${NC}"
    go install github.com/go-delve/delve/cmd/dlv@latest
fi

# Function to find GZH processes
find_gzh_processes() {
    echo -e "${YELLOW}🔍 Looking for GZH processes...${NC}"
    ps aux | grep -E '(gz|gzh-manager)' | grep -v grep | grep -v debug
}

# Parse arguments
TARGET_PID=""

if [ $# -eq 1 ]; then
    TARGET_PID="$1"
    # Validate PID
    if ! kill -0 "$TARGET_PID" 2>/dev/null; then
        echo -e "${RED}❌ Process $TARGET_PID not found or not accessible${NC}"
        exit 1
    fi
else
    # Find running GZH processes
    GZH_PROCESSES=$(find_gzh_processes)

    if [ -z "$GZH_PROCESSES" ]; then
        echo -e "${RED}❌ No running GZH processes found${NC}"
        echo -e "${YELLOW}💡 Start a GZH process first:${NC}"
        echo -e "   ./gz serve --port 8080 &"
        echo -e "   ./gz monitoring start &"
        echo -e "   ./gz bulk-clone --config samples/bulk-clone-simple.yaml &"
        exit 1
    fi

    echo -e "${GREEN}🎯 Found GZH processes:${NC}"
    echo "$GZH_PROCESSES"
    echo ""

    # Extract PIDs
    PIDS=($(echo "$GZH_PROCESSES" | awk '{print $2}'))

    if [ ${#PIDS[@]} -eq 1 ]; then
        TARGET_PID="${PIDS[0]}"
        echo -e "${GREEN}✅ Auto-selecting PID: $TARGET_PID${NC}"
    else
        echo -e "${YELLOW}🤔 Multiple processes found. Please specify PID:${NC}"
        for i in "${!PIDS[@]}"; do
            PID="${PIDS[$i]}"
            PROC_INFO=$(ps -p "$PID" -o pid,ppid,cmd --no-headers 2>/dev/null || echo "$PID - Process info unavailable")
            echo -e "   ${BLUE}[$((i+1))]${NC} $PROC_INFO"
        done
        echo ""
        read -p "Enter PID or selection number: " SELECTION

        # Check if selection is a number (1-based index)
        if [[ "$SELECTION" =~ ^[0-9]+$ ]] && [ "$SELECTION" -ge 1 ] && [ "$SELECTION" -le ${#PIDS[@]} ]; then
            TARGET_PID="${PIDS[$((SELECTION-1))]}"
        else
            TARGET_PID="$SELECTION"
        fi

        # Validate selected PID
        if ! kill -0 "$TARGET_PID" 2>/dev/null; then
            echo -e "${RED}❌ Invalid PID: $TARGET_PID${NC}"
            exit 1
        fi
    fi
fi

# Get process information
PROC_INFO=$(ps -p "$TARGET_PID" -o pid,ppid,cmd --no-headers 2>/dev/null || echo "Process info unavailable")
echo -e "${BLUE}🎯 Target Process:${NC}"
echo "   PID: $TARGET_PID"
echo "   Info: $PROC_INFO"
echo ""

echo -e "${BLUE}🚀 Attaching Delve debugger...${NC}"
echo -e "${BLUE}   Host: $DEBUG_HOST${NC}"
echo -e "${BLUE}   Port: $DEBUG_PORT${NC}"
echo -e "${BLUE}   PID: $TARGET_PID${NC}"
echo ""
echo -e "${YELLOW}📋 Debug Commands:${NC}"
echo -e "   ${GREEN}bt${NC}             - Stack trace"
echo -e "   ${GREEN}goroutines${NC}     - List goroutines"
echo -e "   ${GREEN}b <func>${NC}       - Set breakpoint"
echo -e "   ${GREEN}c${NC}              - Continue execution"
echo -e "   ${GREEN}n${NC}              - Next line"
echo -e "   ${GREEN}s${NC}              - Step into"
echo -e "   ${GREEN}p <var>${NC}        - Print variable"
echo -e "   ${GREEN}l${NC}              - List source code"
echo -e "   ${GREEN}vars${NC}           - Show variables"
echo -e "   ${GREEN}q${NC}              - Quit debugger (process continues)"
echo ""
echo -e "${YELLOW}🌐 Web UI: http://$DEBUG_HOST:$DEBUG_PORT${NC}"
echo ""
echo -e "${RED}⚠️  Note: Attaching may pause the target process${NC}"
echo -e "${BLUE}Press Ctrl+C to stop the debugger${NC}"
echo ""

# Start debugging session
trap 'echo -e "\n${YELLOW}🛑 Debugging session detached${NC}"' EXIT

# Launch delve in attach mode
dlv attach \
    --listen="$DEBUG_HOST:$DEBUG_PORT" \
    --headless \
    --api-version=2 \
    --accept-multiclient \
    --log \
    --log-output="debugger,gdbwire,lldbout,debuglineerr,rpc" \
    "$TARGET_PID"
