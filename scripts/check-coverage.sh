#!/bin/bash

# Script to check test coverage and enforce thresholds

set -e

# Default thresholds
TOTAL_THRESHOLD=${COVERAGE_THRESHOLD:-70}
PACKAGE_THRESHOLD=${PACKAGE_COVERAGE_THRESHOLD:-60}

echo "Checking test coverage..."
echo "Total threshold: ${TOTAL_THRESHOLD}%"
echo "Package threshold: ${PACKAGE_THRESHOLD}%"
echo ""

# Run tests with coverage
go test -coverprofile=coverage.out -covermode=atomic ./... > /dev/null 2>&1

# Get total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo "Total coverage: ${TOTAL_COVERAGE}%"

# Check if total coverage meets threshold
if (( $(echo "$TOTAL_COVERAGE < $TOTAL_THRESHOLD" | bc -l) )); then
    echo "❌ Total coverage ${TOTAL_COVERAGE}% is below threshold ${TOTAL_THRESHOLD}%"
    exit 1
else
    echo "✅ Total coverage meets threshold"
fi

echo ""
echo "Checking package coverage..."

# Check individual package coverage
FAILED_PACKAGES=()

while IFS= read -r line; do
    if [[ $line == *"total"* ]]; then
        continue
    fi
    
    PACKAGE=$(echo "$line" | awk '{print $1}')
    COVERAGE=$(echo "$line" | awk '{print $3}' | sed 's/%//')
    
    if [[ -n "$COVERAGE" ]] && (( $(echo "$COVERAGE < $PACKAGE_THRESHOLD" | bc -l) )); then
        FAILED_PACKAGES+=("$PACKAGE: ${COVERAGE}%")
    fi
done < <(go tool cover -func=coverage.out)

if [ ${#FAILED_PACKAGES[@]} -gt 0 ]; then
    echo "❌ The following packages are below the ${PACKAGE_THRESHOLD}% threshold:"
    for pkg in "${FAILED_PACKAGES[@]}"; do
        echo "  - $pkg"
    done
    exit 1
else
    echo "✅ All packages meet the coverage threshold"
fi

echo ""
echo "Coverage check passed! 🎉"

# Generate detailed report if requested
if [ "$1" = "--report" ]; then
    echo ""
    echo "Generating detailed coverage report..."
    make cover-report
fi