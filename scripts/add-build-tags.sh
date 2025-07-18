#!/bin/bash

# Script to add build tags to integration and e2e test files

set -e

echo "Adding build tags to integration tests..."

# Function to add build tags to a file
add_build_tags() {
    local file=$1
    local tag=$2
    
    # Check if file already has build tags
    if head -n 3 "$file" | grep -q "//go:build\|// +build"; then
        echo "Skipping $file - already has build tags"
        return
    fi
    
    # Create temporary file with build tags
    {
        echo "//go:build $tag"
        echo "// +build $tag"
        echo ""
        cat "$file"
    } > "$file.tmp"
    
    # Replace original file
    mv "$file.tmp" "$file"
    echo "Added $tag tag to $file"
}

# Add integration tags to test/integration
find test/integration -name "*_test.go" -type f | while read -r file; do
    add_build_tags "$file" "integration"
done

# Add e2e tags to test/e2e
find test/e2e -name "*_test.go" -type f | while read -r file; do
    add_build_tags "$file" "e2e"
done

echo "Build tags added successfully!"

# Update Makefile to support running tests with tags
echo ""
echo "To run tests with build tags, use:"
echo "  go test -tags=integration ./test/integration/..."
echo "  go test -tags=e2e ./test/e2e/..."
echo ""
echo "Or update your Makefile with:"
echo "  test-unit:"
echo "      go test ./... -short"
echo "  test-integration:"
echo "      go test -tags=integration ./test/integration/..."
echo "  test-e2e:"
echo "      go test -tags=e2e ./test/e2e/..."