#!/bin/bash
# Performance Test Commands

echo "Testing bulk-clone with large org (if available)"
time gz bulk-clone --org kubernetes || echo "Bulk clone test skipped"

echo "Testing repo-config audit performance"
time gz repo-config audit --org test-org || echo "Audit performance test skipped"

echo "Testing dev-env switching performance"
time gz dev-env status || echo "Dev-env performance test skipped"
