#!/bin/bash

# Script to categorize QA tests and prepare them for execution

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MANUAL_DIR="${SCRIPT_DIR}/manual"

# Create manual directory if not exists
mkdir -p "$MANUAL_DIR"

echo "Categorizing QA tests..."

# Process github-organization-management.qa.md (ALL MANUAL)
echo "Processing github-organization-management.qa.md..."
cp "${SCRIPT_DIR}/github-organization-management.qa.md" "${MANUAL_DIR}/"

# Create agent-friendly commands for manual tests
cat > "${MANUAL_DIR}/github-org-management-agent-commands.md" << 'EOF'
# GitHub Organization Management - Agent-Friendly Test Commands

## Test 1: Repository Configuration Diff Tool
```bash
# Prerequisites: Need a GitHub organization with test repositories
# Copy and paste this block to an agent:

# Setup test environment
export GITHUB_ORG="your-test-org"
export GITHUB_TOKEN="your-github-token"

# Test the diff tool (after gz is built and functional)
gz repo-config diff --org $GITHUB_ORG --repos repo1,repo2

# Verify the HTML report
ls -la repo-config-diff-*.html
```

## Test 2: Policy Compliance Audit
```bash
# Prerequisites: Need GitHub org with multiple repositories
# Copy and paste this block to an agent:

# Run audit with policy file
cat > test-policy.yaml << 'POLICY'
policies:
  - name: "Branch Protection Required"
    type: "branch_protection"
    severity: "high"
    rules:
      - main_branch_protected: true
      - require_pr_reviews: true
POLICY

gz repo-config audit --org $GITHUB_ORG --policy test-policy.yaml --output audit-report.html
```

## Test 3: Webhook Management
```bash
# Prerequisites: Need a webhook endpoint (can use webhook.site for testing)
# Copy and paste this block to an agent:

# Create webhook config
cat > webhook-config.yaml << 'CONFIG'
webhooks:
  - url: "https://webhook.site/your-unique-id"
    events: ["push", "pull_request"]
    active: true
CONFIG

gz repo-config webhook setup --org $GITHUB_ORG --config webhook-config.yaml
```

## Test 4: GitHub Actions Permission Policy
```bash
# Prerequisites: GitHub org with Actions enabled
# Copy and paste this block to an agent:

# Create actions policy
cat > actions-policy.yaml << 'POLICY'
github_actions:
  allowed_actions: "selected"
  allowed_actions_list:
    - "actions/checkout@*"
    - "actions/setup-node@*"
  permissions:
    contents: "read"
    pull-requests: "write"
POLICY

gz repo-config actions-policy apply --org $GITHUB_ORG --policy actions-policy.yaml
```

## Test 5: Dependency Management Policy
```bash
# Prerequisites: Repos with package files
# Copy and paste this block to an agent:

# Create Dependabot config
cat > dependabot-policy.yaml << 'POLICY'
dependabot:
  updates:
    - package_ecosystem: "npm"
      directory: "/"
      schedule:
        interval: "weekly"
    - package_ecosystem: "gomod"
      directory: "/"
      schedule:
        interval: "weekly"
POLICY

gz repo-config dependabot apply --org $GITHUB_ORG --policy dependabot-policy.yaml
```
EOF

# Process network-environment-management.qa.md (MIXED)
echo "Processing network-environment-management.qa.md..."

# Extract automated tests
cat > "${SCRIPT_DIR}/network-env-automated.sh" << 'EOF'
#!/bin/bash
# Automated Network Environment Tests

echo "Running AWS Profile Management Test"
gz dev-env aws-profile switch default || echo "AWS profile switching not available"

echo "Running GCP Project Management Test"
gz dev-env gcp-project switch default-project || echo "GCP project switching not available"

echo "Running Azure Subscription Management Test"  
gz dev-env azure-subscription switch default-sub || echo "Azure subscription switching not available"
EOF

# Extract manual tests to manual folder
cat > "${MANUAL_DIR}/network-env-manual-tests.md" << 'EOF'
# Network Environment Manual Tests

## Docker Network Profiles
Manual setup required - Docker must be running with test containers

## Kubernetes Namespace Settings
Manual setup required - Kubernetes cluster access needed

## Multi-VPN Management
Manual setup required - VPN configurations needed

## Network Performance Monitoring
Manual setup required - Network monitoring infrastructure needed
EOF

# Process user-experience-improvements.qa.md
echo "Processing user-experience-improvements.qa.md..."

# Extract automated performance tests
cat > "${SCRIPT_DIR}/performance-automated.sh" << 'EOF'
#!/bin/bash
# Performance Test Commands

echo "Testing bulk-clone with large org (if available)"
time gz bulk-clone --org kubernetes || echo "Bulk clone test skipped"

echo "Testing repo-config audit performance"
time gz repo-config audit --org test-org || echo "Audit performance test skipped"

echo "Testing dev-env switching performance"
time gz dev-env status || echo "Dev-env performance test skipped"
EOF

# Create a summary of all manual tests
cat > "${MANUAL_DIR}/ALL_MANUAL_TESTS_SUMMARY.md" << 'EOF'
# Manual QA Tests Summary

## Files Moved to Manual Testing:

1. **github-organization-management.qa.md** - ALL tests require GitHub org setup
2. **Network Environment Manual Tests** - Docker, K8s, VPN setup required  
3. **UI/UX Manual Verification** - Visual inspection required

## How to Use:

1. Each manual test file contains agent-friendly command blocks
2. Copy the entire command block and paste into an agent session
3. Replace placeholder values (tokens, org names, etc.)
4. Run the commands and verify outputs

## Prerequisites for Manual Testing:

- GitHub organization with admin access
- GitHub personal access token with full repo permissions
- Docker running locally (for Docker tests)
- Kubernetes cluster access (for K8s tests)
- VPN configurations (for VPN tests)
- Cloud provider credentials (AWS/GCP/Azure)
EOF

chmod +x "${SCRIPT_DIR}/network-env-automated.sh"
chmod +x "${SCRIPT_DIR}/performance-automated.sh"

echo "Test categorization complete!"
echo ""
echo "Manual tests moved to: ${MANUAL_DIR}/"
echo "Automated test scripts created in: ${SCRIPT_DIR}/"
echo ""
echo "Next steps:"
echo "1. Fix compilation errors in the project"
echo "2. Run 'make build' to create the gz binary"
echo "3. Run the automated test scripts"
echo "4. For manual tests, see ${MANUAL_DIR}/ALL_MANUAL_TESTS_SUMMARY.md"