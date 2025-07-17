#!/bin/bash

# Format code
echo "Running go fmt..."
go fmt ./...

# Run lint
echo "Running golangci-lint..."
golangci-lint run --config .golangci.yml --fix --out-format colored-line-number > lint-output-round2.txt 2>&1

echo "Formatting and linting completed. Results saved to lint-output-round2.txt"