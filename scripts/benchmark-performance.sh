#!/bin/bash

# 스크립트명: benchmark-performance.sh  
# 용도: 성능 벤치마킹 및 회귀 테스트 자동화
# 사용법: ./scripts/benchmark-performance.sh [--baseline] [--compare baseline.json]
# 예시: ./scripts/benchmark-performance.sh --baseline > baseline.json

set -e

# Default values
BASELINE_FILE=""
CREATE_BASELINE=false
ITERATIONS=5
OUTPUT_FORMAT="json"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --baseline)
            CREATE_BASELINE=true
            shift
            ;;
        --compare)
            BASELINE_FILE="$2"
            shift 2
            ;;
        --iterations)
            ITERATIONS="$2"
            shift 2
            ;;
        --format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --baseline              Create new baseline measurements"
            echo "  --compare FILE          Compare against baseline file"
            echo "  --iterations N          Number of iterations (default: 5)"
            echo "  --format FORMAT         Output format: json, human (default: json)"
            echo "  --help                  Show this help"
            echo ""
            echo "Examples:"
            echo "  $0 --baseline > baseline.json"
            echo "  $0 --compare baseline.json"
            echo "  $0 --format human"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Ensure binary exists
if [[ ! -f "./gz" ]]; then
    echo "Error: gz binary not found. Run 'make build' first." >&2
    exit 1
fi

# Create output directory
mkdir -p tmp/benchmarks

# Function to measure startup time
measure_startup_time() {
    local total_time=0
    local times=()
    
    for ((i=1; i<=ITERATIONS; i++)); do
        # Use time command to measure precisely
        local time_output=$(time -p ./gz --help 2>&1 >/dev/null | grep real | awk '{print $2}')
        times+=("$time_output")
        total_time=$(echo "$total_time + $time_output" | bc -l)
    done
    
    local avg_time=$(echo "scale=6; $total_time / $ITERATIONS" | bc -l)
    local min_time=$(printf '%s\n' "${times[@]}" | sort -n | head -1)
    local max_time=$(printf '%s\n' "${times[@]}" | sort -n | tail -1)
    
    echo "{\"avg\": $avg_time, \"min\": $min_time, \"max\": $max_time, \"iterations\": $ITERATIONS, \"times\": [$(IFS=,; echo "${times[*]}")]},"
}

# Function to measure binary size
measure_binary_size() {
    local size_bytes=$(stat -c%s "./gz")
    local size_mb=$(echo "scale=2; $size_bytes / 1024 / 1024" | bc -l)
    
    echo "{\"bytes\": $size_bytes, \"mb\": $size_mb},"
}

# Function to measure memory usage
measure_memory_usage() {
    local mem_stats=$(./gz profile stats 2>/dev/null | grep -E "(Heap Allocated|Heap System|Heap In Use)" | sed 's/.*: *//g')
    local heap_alloc=$(echo "$mem_stats" | head -1 | sed 's/[^0-9.]//g')
    local heap_sys=$(echo "$mem_stats" | sed -n '2p' | sed 's/[^0-9.]//g')
    local heap_inuse=$(echo "$mem_stats" | tail -1 | sed 's/[^0-9.]//g')
    
    echo "{\"heap_alloc_mb\": ${heap_alloc:-0}, \"heap_sys_mb\": ${heap_sys:-0}, \"heap_inuse_mb\": ${heap_inuse:-0}},"
}

# Function to test key commands performance
measure_command_performance() {
    local commands=("--help" "synclone --help" "git --help" "profile --help")
    local results=()
    
    for cmd in "${commands[@]}"; do
        local total_time=0
        local cmd_times=()
        
        for ((i=1; i<=3; i++)); do
            local time_output=$(time -p ./gz $cmd 2>&1 >/dev/null | grep real | awk '{print $2}')
            cmd_times+=("$time_output")
            total_time=$(echo "$total_time + $time_output" | bc -l)
        done
        
        local avg_time=$(echo "scale=6; $total_time / 3" | bc -l)
        results+=("{\"command\": \"$cmd\", \"avg_time\": $avg_time}")
    done
    
    echo "["
    IFS=,; echo "${results[*]}"
    echo "],"
}

# Function to create JSON output
create_json_output() {
    local timestamp=$(date -Iseconds)
    local git_commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local git_branch=$(git branch --show-current 2>/dev/null || echo "unknown")
    
    echo "{"
    echo "  \"timestamp\": \"$timestamp\","
    echo "  \"git_commit\": \"$git_commit\","
    echo "  \"git_branch\": \"$git_branch\","
    echo "  \"system\": {"
    echo "    \"os\": \"$(uname -s)\","
    echo "    \"arch\": \"$(uname -m)\","
    echo "    \"go_version\": \"$(go version | awk '{print $3}')\""
    echo "  },"
    echo "  \"measurements\": {"
    echo "    \"startup_time\": $(measure_startup_time)"
    echo "    \"binary_size\": $(measure_binary_size)"
    echo "    \"memory_usage\": $(measure_memory_usage)"
    echo "    \"command_performance\": $(measure_command_performance)"
    echo "    \"success\": true"
    echo "  }"
    echo "}"
}

# Function to create human-readable output
create_human_output() {
    echo "🚀 Performance Benchmark Results"
    echo "================================"
    echo "Timestamp: $(date)"
    echo "Git: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown") ($(git branch --show-current 2>/dev/null || echo "unknown"))"
    echo ""
    
    echo "📊 Startup Performance:"
    local startup_data=$(measure_startup_time)
    local avg_time=$(echo "$startup_data" | jq -r '.avg')
    local min_time=$(echo "$startup_data" | jq -r '.min')
    local max_time=$(echo "$startup_data" | jq -r '.max')
    echo "  Average: ${avg_time}s (${ITERATIONS} iterations)"
    echo "  Range: ${min_time}s - ${max_time}s"
    
    if (( $(echo "$avg_time > 0.05" | bc -l) )); then
        echo "  ⚠️  WARNING: Startup time exceeds 50ms threshold"
    else
        echo "  ✅ Startup time within acceptable range"
    fi
    
    echo ""
    echo "💾 Binary Size:"
    local size_data=$(measure_binary_size)
    local size_mb=$(echo "$size_data" | jq -r '.mb')
    echo "  Size: ${size_mb}MB"
    
    echo ""
    echo "🧠 Memory Usage:"
    local mem_data=$(measure_memory_usage)
    local heap_alloc=$(echo "$mem_data" | jq -r '.heap_alloc_mb')
    echo "  Heap Allocated: ${heap_alloc}MB"
    
    echo ""
    echo "⚡ Command Performance:"
    local cmd_data=$(measure_command_performance)
    echo "$cmd_data" | jq -r '.[] | "  \(.command): \(.avg_time)s"'
}

# Function to compare with baseline
compare_with_baseline() {
    if [[ ! -f "$BASELINE_FILE" ]]; then
        echo "Error: Baseline file '$BASELINE_FILE' not found" >&2
        exit 1
    fi
    
    echo "🔍 Performance Comparison"
    echo "========================"
    echo "Baseline: $BASELINE_FILE"
    echo "Current:  $(date)"
    echo ""
    
    # Get current measurements
    local current_startup=$(measure_startup_time | jq '.avg')
    local current_size=$(measure_binary_size | jq '.mb')
    
    # Get baseline measurements
    local baseline_startup=$(jq '.measurements.startup_time.avg' "$BASELINE_FILE")
    local baseline_size=$(jq '.measurements.binary_size.mb' "$BASELINE_FILE")
    
    # Calculate differences
    local startup_diff=$(echo "scale=6; $current_startup - $baseline_startup" | bc -l)
    local size_diff=$(echo "scale=2; $current_size - $baseline_size" | bc -l)
    
    echo "📊 Startup Time:"
    echo "  Baseline: ${baseline_startup}s"
    echo "  Current:  ${current_startup}s"
    echo "  Change:   ${startup_diff}s"
    
    if (( $(echo "$startup_diff > 0.01" | bc -l) )); then
        echo "  🔴 REGRESSION: Startup time increased significantly"
    elif (( $(echo "$startup_diff < -0.01" | bc -l) )); then
        echo "  🟢 IMPROVEMENT: Startup time decreased"
    else
        echo "  ✅ No significant change"
    fi
    
    echo ""
    echo "💾 Binary Size:"  
    echo "  Baseline: ${baseline_size}MB"
    echo "  Current:  ${current_size}MB"
    echo "  Change:   ${size_diff}MB"
    
    if (( $(echo "$size_diff > 1" | bc -l) )); then
        echo "  🔴 REGRESSION: Binary size increased significantly"
    elif (( $(echo "$size_diff < -1" | bc -l) )); then
        echo "  🟢 IMPROVEMENT: Binary size decreased"
    else
        echo "  ✅ No significant change"
    fi
}

# Main execution
main() {
    if [[ "$CREATE_BASELINE" == "true" ]]; then
        if [[ "$OUTPUT_FORMAT" == "human" ]]; then
            create_human_output
        else
            create_json_output
        fi
    elif [[ -n "$BASELINE_FILE" ]]; then
        compare_with_baseline
    else
        if [[ "$OUTPUT_FORMAT" == "human" ]]; then
            create_human_output
        else
            create_json_output
        fi
    fi
}

# Check dependencies
if ! command -v bc &> /dev/null; then
    echo "Error: 'bc' command not found. Please install bc for calculations." >&2
    exit 1
fi

if ! command -v jq &> /dev/null && [[ "$BASELINE_FILE" != "" ]]; then
    echo "Error: 'jq' command not found. Please install jq for JSON processing." >&2
    exit 1
fi

main