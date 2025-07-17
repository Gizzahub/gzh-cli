#!/bin/bash

# Change to the project directory
cd /Users/archmagece/myopen/Gizzahub/gzh-manager-go

# Run golangci-lint with the correct config file
echo "Running golangci-lint..."
golangci-lint run --config .golangci.yml --fix --out-format colored-line-number 2>&1 | tee lint-output.txt
echo "Lint completed. Results saved to lint-output.txt"