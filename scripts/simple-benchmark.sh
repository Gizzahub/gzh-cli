#!/bin/bash

# 스크립트명: simple-benchmark.sh
# 용도: 간단한 성능 벤치마킹
# 사용법: ./scripts/simple-benchmark.sh
# 예시: ./scripts/simple-benchmark.sh

set -e

echo "🚀 Simple Performance Benchmark"
echo "==============================="
echo "Timestamp: $(date)"
echo "Git: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
echo ""

# Binary size
echo "💾 Binary Size:"
binary_size=$(ls -lh ./gz | awk '{print $5}')
echo "  Current: $binary_size"

# Simple startup time test (3 iterations)
echo ""
echo "📊 Startup Performance (3 iterations):"
total_time=0
for i in {1..3}; do
    start_time=$(date +%s.%N)
    ./gz --help >/dev/null 2>&1
    end_time=$(date +%s.%N)
    iteration_time=$(echo "$end_time - $start_time" | bc -l)
    printf "  Iteration %d: %.3fs\n" $i $iteration_time
    total_time=$(echo "$total_time + $iteration_time" | bc -l)
done

avg_time=$(echo "scale=3; $total_time / 3" | bc -l)
echo "  Average: ${avg_time}s"

# Performance threshold check
threshold="0.050"
if (( $(echo "$avg_time > $threshold" | bc -l) )); then
    echo "  ⚠️  WARNING: Average startup time ${avg_time}s exceeds threshold ${threshold}s"
else
    echo "  ✅ Startup time within acceptable range (< ${threshold}s)"
fi

# Memory stats
echo ""
echo "🧠 Memory Usage:"
./gz profile stats 2>/dev/null | head -7

# Test key commands
echo ""
echo "⚡ Command Response Test:"
commands=("--help" "synclone --help" "git --help" "profile --help")

for cmd in "${commands[@]}"; do
    start_time=$(date +%s.%N)
    ./gz $cmd >/dev/null 2>&1
    end_time=$(date +%s.%N)
    cmd_time=$(echo "$end_time - $start_time" | bc -l)
    printf "  %-15s: %.3fs\n" "$cmd" $cmd_time
done

echo ""
echo "🎉 Benchmark completed successfully!"