> ⚠️ 이 QA는 자동으로 검증할 수 없습니다.  
> 아래 절차에 따라 수동으로 확인해야 합니다.

### ✅ 수동 테스트 지침
- [ ] 실제 환경에서 테스트 수행
- [ ] 외부 서비스 연동 확인
- [ ] 사용자 시나리오 검증
- [ ] 결과 문서화

---

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
