#!/bin/bash

# Test script to verify lint exclusions are working properly

echo "=== Testing golangci-lint exclusions ==="
echo "Current directory: $(pwd)"
echo "Config file: .golangci.yml"
echo ""

# Check if the file exists
if [ -f "pkg/bulk-clone/example_test.go" ]; then
    echo "✓ File exists: pkg/bulk-clone/example_test.go"
else
    echo "✗ File not found: pkg/bulk-clone/example_test.go"
    exit 1
fi

# Check config file
if [ -f ".golangci.yml" ]; then
    echo "✓ Config file exists: .golangci.yml"
else
    echo "✗ Config file not found: .golangci.yml"
    exit 1
fi

echo ""
echo "=== Testing exclusion patterns ==="

# Test skip-files patterns
echo "Checking skip-files patterns:"
grep -A 5 "skip-files:" .golangci.yml | grep -E "(example_test|pkg/bulk-clone)"

echo ""
echo "Checking exclude-rules patterns:"
grep -A 5 "pkg/bulk-clone/example_test" .golangci.yml

echo ""
echo "=== Running golangci-lint to test exclusions ==="
echo "This will show if the file is being processed or excluded..."

# Run golangci-lint with verbose output to see what files are being processed
golangci-lint run --config .golangci.yml --verbose 2>&1 | grep -E "(example_test|Processing|Skipping)" || echo "No example_test references found in lint output"

echo ""
echo "=== Test completed ==="
