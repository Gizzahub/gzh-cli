#!/bin/bash

cd /Users/archmagece/myopen/Gizzahub/gzh-manager-go

echo "=== CURRENT STATE CHECK ==="
echo "Date: $(date)"
echo ""

echo "=== GIT STATUS ==="
git status --porcelain
echo ""

echo "=== RUNNING go fmt ==="
go fmt ./... 2>&1
echo ""

echo "=== RUNNING golangci-lint ==="
golangci-lint run --config .golangci.yml --fix --timeout 10m 2>&1 | head -50
echo ""

echo "=== CHECKING SPECIFIC FILES ==="
echo "-- Checking unused variables in internal/errors/recovery.go --"
grep -n "circuitOpen\|mu.*sync" internal/errors/recovery.go || echo "No unused variables found"
echo ""

echo "-- Checking error handling patterns --"
grep -n "err.*=" internal/errors/recovery.go | head -10
echo ""

echo "=== SUMMARY ==="
echo "Check completed at $(date)"
