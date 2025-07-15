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
